package oracler

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

// maxDepth bounds alias chasing so that recursive channel plumbing cannot make
// resolution diverge.
const maxDepth = 16

// index holds one type-checked package plus the lookup tables the resolver
// needs. Tables are built once per loaded package.
type index struct {
	fset  *token.FileSet
	files []*ast.File
	info  *types.Info

	// assigns maps a variable or struct field to the expressions assigned to
	// it: var declarations, assignments and composite literal fields.
	assigns map[types.Object][]ast.Expr
	// calls maps a function to the argument lists of all calls to it in this
	// package, including "go" statements and "defer".
	calls map[types.Object][][]ast.Expr
	// params maps a parameter variable to its owner and position, so that a
	// channel received as an argument can be resolved at the call sites.
	params map[types.Object]paramInfo
	// lits records calls of function literals, which have no types.Object.
	lits map[*ast.FuncLit][]*ast.CallExpr
}

type paramInfo struct {
	owner    ast.Node // *ast.FuncDecl or *ast.FuncLit
	idx      int
	variadic bool
}

var (
	cacheMu sync.Mutex
	cache   = make(map[string]*cacheEntry)
)

type cacheEntry struct {
	idx *index
	err error
}

// indexFor loads and indexes the package containing path, once per file.
func indexFor(path string) (*index, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	if e, ok := cache[abs]; ok {
		return e.idx, e.err
	}

	idx, loadErr := load(abs)
	cache[abs] = &cacheEntry{idx: idx, err: loadErr}
	if loadErr != nil {
		fmt.Fprintf(os.Stderr, "oracler: cannot analyse %s: %v\n", path, loadErr)
	}
	return idx, loadErr
}

// load type-checks the package containing abs. It prefers go/packages, which
// understands modules, and falls back to checking the file on its own for
// files that do not belong to a loadable package.
func load(abs string) (*index, error) {
	if idx, err := loadPackage(abs); err == nil {
		return idx, nil
	}
	return loadFile(abs)
}

func loadPackage(abs string) (*index, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo,
		Dir: filepath.Dir(abs),
	}

	pkgs, err := packages.Load(cfg, "file="+abs)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		if pkg == nil || len(pkg.Syntax) == 0 || pkg.TypesInfo == nil || pkg.Fset == nil {
			continue
		}
		for _, f := range pkg.Syntax {
			if sameFile(pkg.Fset.Position(f.Pos()).Filename, abs) {
				return newIndex(pkg.Fset, pkg.Syntax, pkg.TypesInfo), nil
			}
		}
	}

	return nil, fmt.Errorf("no package loaded for %s", abs)
}

func loadFile(abs string) (*index, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, abs, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}

	conf := types.Config{
		Importer: importer.ForCompiler(fset, "source", nil),
		// Keep whatever type information could be computed: the analysed
		// programs are snippets and are not always fully well-typed.
		Error: func(error) {},
	}
	_, _ = conf.Check(file.Name.Name, fset, []*ast.File{file}, info)

	return newIndex(fset, []*ast.File{file}, info), nil
}

func newIndex(fset *token.FileSet, files []*ast.File, info *types.Info) *index {
	ix := &index{
		fset:    fset,
		files:   files,
		info:    info,
		assigns: make(map[types.Object][]ast.Expr),
		calls:   make(map[types.Object][][]ast.Expr),
		params:  make(map[types.Object]paramInfo),
		lits:    make(map[*ast.FuncLit][]*ast.CallExpr),
	}
	for _, f := range files {
		ix.collect(f)
	}
	return ix
}

// collect records parameters, assignments and call sites of one file.
func (ix *index) collect(f *ast.File) {
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name == nil {
			continue
		}
		if fn, ok := ix.info.Defs[fd.Name].(*types.Func); ok {
			if sig, ok := fn.Type().(*types.Signature); ok {
				ix.collectParams(fd, sig)
			}
		}
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncLit:
			if tv, ok := ix.info.Types[x]; ok {
				if sig, ok := tv.Type.(*types.Signature); ok {
					ix.collectParams(x, sig)
				}
			}

		case *ast.ValueSpec:
			for i, name := range x.Names {
				if i < len(x.Values) {
					ix.addAssign(ix.info.Defs[name], x.Values[i])
				}
			}

		case *ast.AssignStmt:
			// A tuple assignment (a, b = f()) carries no per-variable source.
			if len(x.Lhs) == len(x.Rhs) {
				for i, lhs := range x.Lhs {
					ix.addAssign(ix.object(lhs), x.Rhs[i])
				}
			}

		case *ast.CompositeLit:
			for _, el := range x.Elts {
				kv, ok := el.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if key, ok := kv.Key.(*ast.Ident); ok {
					ix.addAssign(ix.info.Uses[key], kv.Value)
				}
			}

		case *ast.CallExpr:
			if lit, ok := x.Fun.(*ast.FuncLit); ok {
				ix.lits[lit] = append(ix.lits[lit], x)
				return true
			}
			if obj := ix.callee(x.Fun); obj != nil {
				ix.calls[obj] = append(ix.calls[obj], x.Args)
			}
		}
		return true
	})
}

func (ix *index) collectParams(owner ast.Node, sig *types.Signature) {
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		p := params.At(i)
		if p == nil || p.Name() == "_" {
			continue
		}
		ix.params[p] = paramInfo{
			owner:    owner,
			idx:      i,
			variadic: sig.Variadic() && i == params.Len()-1,
		}
	}
}

func (ix *index) addAssign(obj types.Object, rhs ast.Expr) {
	if obj == nil || rhs == nil {
		return
	}
	ix.assigns[obj] = append(ix.assigns[obj], rhs)
}

// object returns the object an assignable expression denotes.
func (ix *index) object(e ast.Expr) types.Object {
	switch v := e.(type) {
	case *ast.Ident:
		return ix.info.ObjectOf(v)
	case *ast.SelectorExpr:
		return ix.info.Uses[v.Sel]
	case *ast.ParenExpr:
		return ix.object(v.X)
	case *ast.StarExpr:
		return ix.object(v.X)
	}
	return nil
}

// callee returns the function object called by fun, if it is known.
func (ix *index) callee(fun ast.Expr) types.Object {
	switch f := fun.(type) {
	case *ast.Ident:
		return ix.info.Uses[f]
	case *ast.SelectorExpr:
		if sel, ok := ix.info.Selections[f]; ok {
			return sel.Obj()
		}
		return ix.info.Uses[f.Sel]
	case *ast.ParenExpr:
		return ix.callee(f.X)
	case *ast.IndexExpr:
		return ix.callee(f.X)
	case *ast.IndexListExpr:
		return ix.callee(f.X)
	}
	return nil
}

// resolve returns the make(chan ...) sites that v may originate from.
//
// It follows the shapes the analysed programs use: local and package-level
// variables, parameters through their call sites, and struct fields. Anything
// without a visible origin (a channel returned from another package, reached
// through a function value, ...) yields no allocation site.
func (ix *index) resolve(v ast.Expr) []string {
	seen := make(map[types.Object]bool)
	found := make(map[token.Pos]bool)

	var walk func(e ast.Expr, depth int)
	walk = func(e ast.Expr, depth int) {
		if e == nil || depth > maxDepth {
			return
		}

		switch x := e.(type) {
		case *ast.ParenExpr:
			walk(x.X, depth+1)

		case *ast.StarExpr:
			walk(x.X, depth+1)

		case *ast.CallExpr:
			if ix.isMakeChan(x) {
				found[x.Pos()] = true
			}

		case *ast.Ident:
			obj := ix.info.ObjectOf(x)
			if obj == nil || seen[obj] {
				return
			}
			seen[obj] = true
			if pi, ok := ix.params[obj]; ok {
				ix.walkParam(pi, walk, depth)
				return
			}
			for _, rhs := range ix.assigns[obj] {
				walk(rhs, depth+1)
			}

		case *ast.SelectorExpr:
			obj := ix.info.Uses[x.Sel]
			if obj == nil || seen[obj] {
				return
			}
			seen[obj] = true
			for _, rhs := range ix.assigns[obj] {
				walk(rhs, depth+1)
			}
		}
	}
	walk(v, 0)

	pos := make([]token.Pos, 0, len(found))
	for p := range found {
		pos = append(pos, p)
	}
	sort.Slice(pos, func(i, j int) bool { return pos[i] < pos[j] })

	allocs := make([]string, 0, len(pos))
	for _, p := range pos {
		allocs = append(allocs, ix.position(p))
	}
	return allocs
}

// walkParam follows a parameter to the arguments of every call site.
func (ix *index) walkParam(pi paramInfo, walk func(ast.Expr, int), depth int) {
	switch owner := pi.owner.(type) {
	case *ast.FuncDecl:
		fn, ok := ix.info.Defs[owner.Name].(*types.Func)
		if !ok {
			return
		}
		for _, args := range ix.calls[fn] {
			ix.walkArgs(args, pi, walk, depth)
		}
	case *ast.FuncLit:
		for _, call := range ix.lits[owner] {
			ix.walkArgs(call.Args, pi, walk, depth)
		}
	}
}

func (ix *index) walkArgs(args []ast.Expr, pi paramInfo, walk func(ast.Expr, int), depth int) {
	if pi.variadic {
		// Type errors are tolerated in the analysed snippets, so a call may
		// carry fewer arguments than the parameter list expects.
		if pi.idx > len(args) {
			return
		}
		for _, a := range args[pi.idx:] {
			walk(a, depth+1)
		}
		return
	}
	if pi.idx < len(args) {
		walk(args[pi.idx], depth+1)
	}
}

// isMakeChan reports whether call is make(chan ...) and not a shadowed name.
func (ix *index) isMakeChan(call *ast.CallExpr) bool {
	id, ok := call.Fun.(*ast.Ident)
	if !ok || id.Name != "make" {
		return false
	}
	if obj := ix.info.Uses[id]; obj != nil && obj != types.Universe.Lookup("make") {
		return false
	}
	if tv, ok := ix.info.Types[call]; ok {
		return isChanType(tv.Type)
	}
	return len(call.Args) > 0 && isChanType(ix.info.Types[call.Args[0]].Type)
}

func isChanType(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Chan)
	return ok
}

// channelOperandAt returns the channel expression of the operation whose
// operator sits at byte offset charpos: the channel of a send, the operand of a
// receive, or the argument of close.
func (ix *index) channelOperandAt(path string, charpos int) (ast.Expr, bool) {
	f := ix.fileFor(path)
	if f == nil {
		return nil, false
	}

	var operand ast.Expr
	ast.Inspect(f, func(n ast.Node) bool {
		if operand != nil {
			return false
		}
		switch x := n.(type) {
		case *ast.SendStmt:
			if ix.offset(x.Arrow) == charpos {
				operand = x.Chan
			}
		case *ast.UnaryExpr:
			if x.Op == token.ARROW && ix.offset(x.OpPos) == charpos {
				operand = x.X
			}
		case *ast.CallExpr:
			if len(x.Args) != 1 || !ix.isClose(x.Fun) {
				return true
			}
			if lo, hi := ix.offset(x.Lparen), ix.offset(x.Rparen); charpos >= lo && charpos <= hi {
				operand = x.Args[0]
			}
		}
		return true
	})
	return operand, operand != nil
}

func (ix *index) isClose(fun ast.Expr) bool {
	id, ok := fun.(*ast.Ident)
	if !ok || id.Name != "close" {
		return false
	}
	if obj := ix.info.Uses[id]; obj != nil {
		return obj == types.Universe.Lookup("close")
	}
	return true
}

func (ix *index) fileFor(path string) *ast.File {
	abs := path
	if a, err := filepath.Abs(path); err == nil {
		abs = a
	}
	for _, f := range ix.files {
		if sameFile(ix.fset.Position(f.Pos()).Filename, abs) {
			return f
		}
	}
	return nil
}

func (ix *index) offset(p token.Pos) int {
	if !p.IsValid() {
		return -1
	}
	return ix.fset.Position(p).Offset
}

// position formats p as "file:line:column", keeping only the base name so the
// result stays parseable by LineNColumn.
func (ix *index) position(p token.Pos) string {
	pp := ix.fset.Position(p)
	return fmt.Sprintf("%s:%d:%d", filepath.Base(pp.Filename), pp.Line, pp.Column)
}

// sameFile reports whether a and b name the same file. Paths coming back from
// the go command may use different separators and drive-letter casing than the
// path given on the command line, so both are normalised first.
func sameFile(a, b string) bool {
	if aa, err := filepath.Abs(a); err == nil {
		a = aa
	}
	if bb, err := filepath.Abs(b); err == nil {
		b = bb
	}
	a = filepath.ToSlash(filepath.Clean(a))
	b = filepath.ToSlash(filepath.Clean(b))
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
