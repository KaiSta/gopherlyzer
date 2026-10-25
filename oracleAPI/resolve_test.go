package oracler_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"

	oracler "github.com/KaiSta/gopherlyzer/oracleAPI"
)

// TestResolveAllocs pins, for the programs in example/ and test/, which
// make(chan ...) site each channel operation is attributed to. Keys are
// "line:kind" of the operation; values are the expected allocation lines.
func TestResolveAllocs(t *testing.T) {
	cases := []struct {
		file string
		want map[string][]int64
	}{
		{
			// channel passed as a parameter to a goroutine
			file: "example/simple/main.go",
			want: map[string][]int64{
				"5:send":  {9}, // first(x), x from main's make
				"13:recv": {9}, // local x
				"15:send": {9}, // local x
			},
		},
		{
			// two parameters passed positionally: x and y must stay distinct
			file: "example/foobar/footest.go",
			want: map[string][]int64{
				"4:recv":  {9},  // foo's x -> main's x
				"5:send":  {10}, // foo's y -> main's y
				"15:send": {9},
				"17:recv": {10},
				"21:send": {9},
			},
		},
		{
			// channels flowing through two levels of parameters, and the
			// nested operations of "a <- <-b"
			file: "example/newsreader/main.go",
			want: map[string][]int64{
				"4:send":  {30}, // news1(n1)
				"8:send":  {31}, // news2(n2)
				"12:send": {20}, // collecter's a <- collect's ch
				"12:recv": {30}, // collecter's b <- n1
				"16:send": {20}, // collecter2's a <- collect's ch
				"16:recv": {31}, // collecter2's b <- n2
				"25:recv": {20}, // collect's own ch
			},
		},
		{
			file: "example/closeSample/main.go",
			want: map[string][]int64{
				"4:recv":   {9},
				"5:recv":   {9},
				"11:send":  {9},
				"13:close": {9},
			},
		},
		{
			// a is package-level and assigned in main; main's b and foo's b
			// share a name but are different channels; baz is never called so
			// its parameter channel has no allocation site.
			file: "test/simple/main.go",
			want: map[string][]int64{
				"8:recv":  {15},
				"10:send": {15},
				"22:recv": {15},
				"27:recv": {19},
				"28:send": {15},
				"30:recv": {15},
				"36:recv": {15},
				"43:recv": {42},
				"48:recv": {},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.file, func(t *testing.T) {
			path := filepath.Join("..", filepath.FromSlash(tc.file))
			ops := channelOps(t, path)

			for key := range tc.want {
				offset, ok := ops[key]
				if !ok {
					t.Fatalf("no channel operation %q found in %s", key, tc.file)
				}

				oracle := oracler.New(path)
				peers := oracle.GetPeers(offset)

				got := make([]int64, 0, len(peers.PeerEntries.Allocs))
				for _, alloc := range peers.PeerEntries.Allocs {
					line, _ := oracle.LineNColumn(alloc)
					got = append(got, line)
				}
				sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })

				want := tc.want[key]
				sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
				if len(got) == 0 && len(want) == 0 {
					continue
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("%s %s: allocs = %v, want %v", tc.file, key, got, want)
				}
			}
		})
	}
}

// channelOps returns the byte offset of every channel operation in path, keyed
// by "line:kind". The offsets are the ones the simplifier queries with.
func channelOps(t *testing.T, path string) map[string]int {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	ops := make(map[string]int)
	add := func(pos token.Pos, kind string, offset int) {
		p := fset.Position(pos)
		ops[strconv.Itoa(p.Line)+":"+kind] = offset
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SendStmt:
			add(x.Arrow, "send", fset.Position(x.Arrow).Offset)
		case *ast.UnaryExpr:
			if x.Op == token.ARROW {
				add(x.OpPos, "recv", fset.Position(x.OpPos).Offset)
			}
		case *ast.CallExpr:
			if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "close" {
				add(x.Lparen, "close", fset.Position(x.Lparen).Offset+1)
			}
		}
		return true
	})
	return ops
}
