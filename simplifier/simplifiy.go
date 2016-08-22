package simplify

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"

	"../oracleAPI"
)

type Simplifier struct {
	FSet        *token.FileSet
	File        *ast.File
	Src         string
	Mapv2       map[string]Callgraph2
	Path        string
	PackageName string
}

type Callgraph2 struct {
	Name    string      `json:"name"`
	NrRoots int         `json:"rootCount"`
	Calls   []string    `json:"calls"`
	ChanOps []Operation `json:"ops"`
	AllOps  []Operation `json:"allOps"`
}
type Operation struct {
	Type string
	Name string
	Pos  int
	Op   string
	Ops  []Operation
	Else []Operation
}

func New() *Simplifier {
	return &Simplifier{FSet: token.NewFileSet()}
}

func (s *Simplifier) Parse(filePath string) map[string]Callgraph2 {
	s.Path = filePath
	f, err := parser.ParseFile(s.FSet, filePath, nil, parser.ParseComments)
	s.File = f
	if err != nil {
		panic(err)
	}
	s.Mapv2 = make(map[string]Callgraph2)

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.File:
			s.PackageName = x.Name.Name
		case *ast.FuncDecl:
			tmp2 := s.Mapv2[x.Name.Name]
			if x.Name.Name == "main" {
				tmp2.NrRoots++
			}
			tmp2.Calls = make([]string, 0)

			for _, v := range x.Body.List {
				switch y := v.(type) {
				case *ast.ExprStmt:
					switch z := y.X.(type) {
					case *ast.CallExpr:
						n := s.getValue(z)
						if n == "close" {
							if r, o := s.handleClose(z); r {
								tmp2.ChanOps = append(tmp2.ChanOps, o)
								tmp2.AllOps = append(tmp2.AllOps, o)
							}
						} else {
							tmp2.Calls = append(tmp2.Calls, n)

							var o Operation
							o.Type = "Call"
							o.Name = fmt.Sprintf("%v.%v:%v", s.PackageName, n, 0)
							o.Op = "()"
							tmp2.AllOps = append(tmp2.AllOps, o)
						}

					case *ast.UnaryExpr:
						if r, o := s.handleUnaryExpr(z); r {
							tmp2.ChanOps = append(tmp2.ChanOps, o)
							tmp2.AllOps = append(tmp2.AllOps, o)
						}
					}
				case *ast.AssignStmt:
					switch a := y.Rhs[0].(type) {
					case *ast.UnaryExpr:
						if r, o := s.handleUnaryExpr(a); r {
							tmp2.ChanOps = append(tmp2.ChanOps, o)
							tmp2.AllOps = append(tmp2.AllOps, o)
						}
					}
				case *ast.SendStmt:
					//check if a value from another chan was send
					switch h := y.Value.(type) {
					case *ast.UnaryExpr:
						if r, o := s.handleUnaryExpr(h); r {
							tmp2.ChanOps = append(tmp2.ChanOps, o)
							tmp2.AllOps = append(tmp2.AllOps, o)
						}
					}
					if r, o := s.handleSend(y); r {
						tmp2.ChanOps = append(tmp2.ChanOps, o)
						tmp2.AllOps = append(tmp2.AllOps, o)
					}
				case *ast.SelectStmt:
					var opSelect Operation
					opSelect.Type = "Select"
					res := false
					for _, a := range y.Body.List {
						if r, o := s.handleCommClaus(a.(*ast.CommClause)); r {
							//		fmt.Println("PARSER", o)
							res = true
							opSelect.Ops = append(opSelect.Ops, o)
						}
					}
					if res {
						tmp2.ChanOps = append(tmp2.ChanOps, opSelect)
						tmp2.AllOps = append(tmp2.AllOps, opSelect)
					}
				case *ast.IfStmt:
					if r, o := s.handleIfStmt(y); r {
						tmp2.ChanOps = append(tmp2.ChanOps, o)
						tmp2.AllOps = append(tmp2.AllOps, o)
					}
				case *ast.ForStmt:
					//fmt.Println("FOR")
					if r, o := s.handleForStmt(y); r {
						//fmt.Println(o)
						tmp2.ChanOps = append(tmp2.ChanOps, o)
						tmp2.AllOps = append(tmp2.AllOps, o)
					}
					//fmt.Println("ROF")
				case *ast.GoStmt:
					n := s.getValue(y.Call)

					if n == x.Name.Name {
						tmp2.NrRoots++
					} else {
						tmpY := s.Mapv2[n]
						tmpY.NrRoots++
						s.Mapv2[n] = tmpY
					}
				}
			}
			s.Mapv2[x.Name.Name] = tmp2
		}
		return true
	})
	return s.Mapv2
}

func (s *Simplifier) handleUnaryExpr(n *ast.UnaryExpr) (bool, Operation) {
	var resOps Operation
	if n.Op == token.ARROW {
		pos := s.FSet.Position(n.OpPos).Offset
		oraclePath := "oracle"

		oracle := oracler.New(oraclePath, s.Path, "./"+filepath.Dir(s.Path))

		p := oracle.GetPeers("json", pos)

		res := false
		resOps.Type = "Op"
		for i := range p.PeerEntries.Allocs {
			l, _ := oracle.LineNColumn(p.PeerEntries.Allocs[i])
			res = true
			resOps.Ops = append(resOps.Ops, Operation{Type: "Rcv", Pos: int(l), Name: fmt.Sprintf("%v.%v:%v", s.PackageName, "", l), Op: "?"})
		}

		return res, resOps
	}

	return false, resOps
}

func (s *Simplifier) handleSend(n *ast.SendStmt) (bool, Operation) {
	pos := s.FSet.Position(n.Arrow).Offset
	oraclePath := "oracle"

	oracle := oracler.New(oraclePath, s.Path, "./"+filepath.Dir(s.Path))
	p := oracle.GetPeers("json", pos)
	//	fmt.Println("SEND", pos, p.PeerEntries.Allocs)

	res := false
	var resOps Operation
	resOps.Type = "Op"

	for i := range p.PeerEntries.Allocs {
		l, _ := oracle.LineNColumn(p.PeerEntries.Allocs[i])
		res = true
		resOps.Ops = append(resOps.Ops, Operation{Type: "Snd", Pos: int(l), Name: fmt.Sprintf("%v.%v:%v", s.PackageName, "", l), Op: "!"})
	}
	return res, resOps
}

func (s *Simplifier) handleClose(n *ast.CallExpr) (bool, Operation) {
	pos := s.FSet.Position(n.Lparen).Offset
	oraclePath := "oracle"

	oracle := oracler.New(oraclePath, s.Path, "./"+filepath.Dir(s.Path))
	p := oracle.GetPeers("json", pos+1)

	//	fmt.Println("CLOSE", pos, p.PeerEntries.Allocs)

	res := false
	var resOps Operation
	resOps.Type = "Op"

	for i := range p.PeerEntries.Allocs {
		l, _ := oracle.LineNColumn(p.PeerEntries.Allocs[i])
		res = true
		resOps.Ops = append(resOps.Ops, Operation{Type: "Close", Pos: int(l), Name: fmt.Sprintf("%v.%v:%v", s.PackageName, "", l), Op: "#"})
	}
	return res, resOps
}

func (s *Simplifier) handleCommClaus(n *ast.CommClause) (bool, Operation) {
	if n.Comm != nil {
		switch x := n.Comm.(type) {
		case *ast.ExprStmt:
			switch z := x.X.(type) {
			case *ast.UnaryExpr:
				tmp, r := s.handleUnaryExpr(z)
				//fmt.Println("SELECT1", r)
				for _, bel := range n.Body {
					switch valu := bel.(type) {
					case *ast.SendStmt:
						q, t := s.handleSend(valu)
						if q {
							r.Ops = append(r.Ops, t.Ops[0])
						}
					}
				}
				//	fmt.Println("SELECT2", r)
				// switch valu := n.Body[0].(type) {
				// case *ast.SendStmt:
				// 	q, t := s.handleSend(valu)
				// 	if q {
				// 		r.Ops = append(r.Ops, t)
				// 	}
				// }
				return tmp, r
			}
		case *ast.AssignStmt:
			switch a := x.Rhs[0].(type) {
			case *ast.UnaryExpr:
				r, o := s.handleUnaryExpr(a)
				if len(o.Ops) > 0 {
					switch valu := n.Body[0].(type) {
					case *ast.SendStmt:
						q, t := s.handleSend(valu)
						if q {
							o.Ops[0].Ops = append(o.Ops[0].Ops, t.Ops[0])
						}
					}
					return r, o.Ops[0]
				}
			}
		case *ast.SendStmt:
			r, o := s.handleSend(x)
			return r, o.Ops[0]
		}
	} else {
		for _, bel := range n.Body {
			switch valu := bel.(type) {
			case *ast.SendStmt:
				q, t := s.handleSend(valu)
				if q {
					//fmt.Println("!DEFAULT!", t.Ops[0])
					return true, t.Ops[0]
					//r.Ops = append(r.Ops, t.Ops[0])
				}
			}
		}
	}

	return false, Operation{}
}

func (s *Simplifier) handleIfStmt(n *ast.IfStmt) (bool, Operation) {
	res := false
	var resOps Operation
	resOps.Type = "If"

	for _, a := range n.Body.List {
		switch stmt := a.(type) {
		case *ast.ExprStmt:
			switch xp := stmt.X.(type) {
			case *ast.UnaryExpr:
				r, o := s.handleUnaryExpr(xp)
				if r {
					res = true
					resOps.Ops = append(resOps.Ops, o.Ops[0])
				}
			}
		case *ast.AssignStmt:
			switch xp := stmt.Rhs[0].(type) {
			case *ast.UnaryExpr:
				r, o := s.handleUnaryExpr(xp)
				if r {
					res = true
					resOps.Ops = append(resOps.Ops, o.Ops[0])
				}
			}
		case *ast.SendStmt:
			r, o := s.handleSend(stmt)
			if r {
				res = true
				resOps.Ops = append(resOps.Ops, o.Ops[0])
			}
		}
	}

	if n.Else != nil {
		res = true
		switch el := n.Else.(type) {
		case *ast.IfStmt:
			_, o := s.handleIfStmt(el)
			resOps.Else = append(resOps.Else, o)

		case *ast.BlockStmt:
			for _, a := range el.List {
				switch stmt := a.(type) {
				case *ast.ExprStmt:
					switch xp := stmt.X.(type) {
					case *ast.UnaryExpr:
						r, o := s.handleUnaryExpr(xp)

						if r {
							resOps.Else = append(resOps.Else, o.Ops[0])
						}
					}
				case *ast.AssignStmt:
					switch xp := stmt.Rhs[0].(type) {
					case *ast.UnaryExpr:
						r, o := s.handleUnaryExpr(xp)
						if r {
							resOps.Else = append(resOps.Else, o.Ops[0])
						}
					}
				case *ast.SendStmt:
					r, o := s.handleSend(stmt)
					if r {
						resOps.Else = append(resOps.Else, o.Ops[0])
					}
				}
			}
		}
	} else {
		// else NIL bzw +eps fehlt hier noch !?
	}

	return res, resOps
}

func (s *Simplifier) handleForStmt(n *ast.ForStmt) (bool, Operation) {
	res := false
	var resOps Operation
	resOps.Type = "For"

	for _, a := range n.Body.List {
		switch stmt := a.(type) {
		case *ast.ExprStmt:
			switch xp := stmt.X.(type) {
			case *ast.UnaryExpr:
				if r, o := s.handleUnaryExpr(xp); r {
					res = true
					resOps.Ops = append(resOps.Ops, o.Ops[0])
				}
			}
		case *ast.AssignStmt:
			switch xp := stmt.Rhs[0].(type) {
			case *ast.UnaryExpr:
				if r, o := s.handleUnaryExpr(xp); r {
					res = true
					resOps.Ops = append(resOps.Ops, o.Ops[0])
				}
			}
		case *ast.SendStmt:
			if r, o := s.handleSend(stmt); r {
				res = true
				resOps.Ops = append(resOps.Ops, o.Ops[0])
			}
		case *ast.IfStmt:
			if r, o := s.handleIfStmt(stmt); r {
				res = true
				resOps.Ops = append(resOps.Ops, o)
			}
		case *ast.SelectStmt:
			var opSelect Operation
			opSelect.Type = "Select"
			for _, a := range stmt.Body.List {
				if r, o := s.handleCommClaus(a.(*ast.CommClause)); r {
					res = true
					opSelect.Ops = append(opSelect.Ops, o)
				}
			}
			//fmt.Println("FORSEL", opSelect, len(opSelect.Ops))

			if res {
				resOps.Ops = append(resOps.Ops, opSelect)
			}
			//fmt.Println("FOROPS", resOps)
		default:
			fmt.Println(reflect.TypeOf(stmt))
		}
	}
	return res, resOps
}

func (si *Simplifier) getValue(n ast.Expr) (s string) {
	ast.Inspect(n, func(x ast.Node) bool {
		switch y := x.(type) {
		case *ast.Ident:
			if y != nil {
				s += y.Name
			}
			return false
		case *ast.BasicLit:
			s += y.Value
			return false
		case *ast.ChanType:
			s += "chan "
		case *ast.CallExpr:
			s = si.getValue(y.Fun)
			return false
		case *ast.CompositeLit:
			s += si.getValue(y.Type) + "{"
			for i, v := range y.Elts {
				switch p := v.(type) {
				case *ast.KeyValueExpr:
					s += si.getValue(p.Key) + ": " + si.getValue(p.Value)
				}
				if i < (len(y.Elts) - 1) {
					s += ", "
				}
			}
			s += "}"
			return false
		}

		return true
	})
	return
}
