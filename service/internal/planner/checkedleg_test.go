package planner

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestOnlyCheckedLegReachesTheRoutingProvider is #255 AC4 held to the source:
// "every route the planner obtains is checked before it is used, whatever path
// through the planner obtained it". The behavioural tests in access_test.go
// cover the paths Plan has today; this one covers the path a later change adds.
// Outside New, which stores it, the routing field is read in one function, and
// that function is checkedLeg, which hands every route to the check before it
// returns. A second method that called p.routing.Route — an alternative route
// for #248, a leg rebuilt after avoidance — fails here until it goes through
// checkedLeg.
//
// It reads the package's non-test files and refuses to pass having read none,
// or having found no use at all, so a rename cannot make it vacuous.
func TestOnlyCheckedLegReachesTheRoutingProvider(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	files, uses := 0, 0

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}

		files++

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "routing" {
					return true
				}

				if _, isIdent := sel.X.(*ast.Ident); !isIdent {
					return true
				}

				uses++

				if fn.Name.Name != "checkedLeg" {
					t.Errorf("%s: %s reads the routing provider; only checkedLeg may, so every route is checked before it is used",
						fset.Position(sel.Pos()), fn.Name.Name)
				}

				return true
			})
		}
	}

	if files == 0 || uses == 0 {
		t.Fatalf("read %d files and found %d reads of the routing provider, want at least one of each", files, uses)
	}
}
