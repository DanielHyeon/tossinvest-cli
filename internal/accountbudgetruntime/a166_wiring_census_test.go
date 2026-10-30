package accountbudgetruntime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type wiringSite struct {
	Site   string `json:"site"`
	Class  string `json:"class"`
	Reason string `json:"reason"`
}
type wiringCorpus struct {
	shapes map[string]string
	sites  []wiringSite
}

func loadWiringCorpus(t *testing.T) wiringCorpus {
	t.Helper()
	var c wiringCorpus
	for _, x := range []struct {
		name string
		dst  any
	}{{"shapes.json", &c.shapes}, {"sites.json", &c.sites}} {
		b, e := os.ReadFile(filepath.Join("testdata", "a166-wiring", x.name))
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(b, x.dst); e != nil {
			t.Fatal(e)
		}
	}
	if len(c.shapes) == 0 || len(c.sites) == 0 {
		t.Fatal("empty wiring corpus")
	}
	for _, s := range c.sites {
		if s.Site == "" || s.Reason == "" || s.Class == "" || s.Class == "UNKNOWN" {
			t.Fatalf("unclassified site %+v", s)
		}
	}
	return c
}
func wiringSource(root, path string, overrides map[string][]byte) ([]byte, error) {
	if b, ok := overrides[path]; ok {
		return b, nil
	}
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
}
func wiringRecv(f *ast.FuncDecl) string {
	if f.Recv == nil {
		return ""
	}
	switch x := f.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if id, ok := x.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return x.Name
	}
	return "?"
}
func wiringFormat(fset *token.FileSet, node any) (string, error) {
	var b bytes.Buffer
	if err := format.Node(&b, fset, node); err != nil {
		return "", err
	}
	return b.String(), nil
}
func wiringImportant(sel *ast.SelectorExpr, aliases map[string]string) bool {
	name := sel.Sel.Name
	if id, ok := sel.X.(*ast.Ident); ok && strings.HasSuffix(aliases[id.Name], "/accountbudget") && name == "Budget" {
		return false
	}
	if name == "Do" || name == "RoundTrip" || name == "NewClient" || name == "Budget" {
		return true
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	p := aliases[id.Name]
	return strings.HasSuffix(p, "/official") && (name == "New" || name == "NewWithAccountBudget") || strings.HasSuffix(p, "/accountbudget") && (name == "Open" || name == "OpenContext" || name == "OpenObserved" || name == "NewTransport") || strings.HasSuffix(p, "/accountbudgetruntime") && name == "Open" || p == "net/http" && (name == "Get" || name == "Post" || name == "PostForm" || name == "Head")
}
func wiringInspect(root string, c wiringCorpus, overrides map[string][]byte) error {
	seenShapes := map[string]bool{}
	var sites []string
	for _, dir := range []string{"cmd", "internal", "tools"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(full string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("source symlink: %s", full)
			}
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
				return nil
			}
			rel, e := filepath.Rel(root, full)
			if e != nil {
				return e
			}
			rel = filepath.ToSlash(rel)
			source, e := wiringSource(root, rel, overrides)
			if e != nil {
				return e
			}
			fset := token.NewFileSet()
			f, e := parser.ParseFile(fset, rel, source, 0)
			if e != nil {
				return e
			}
			aliases := map[string]string{}
			for _, imp := range f.Imports {
				path, e := strconv.Unquote(imp.Path.Value)
				if e != nil {
					return e
				}
				if strings.HasSuffix(path, "/official") || strings.HasSuffix(path, "/accountbudget") || strings.HasSuffix(path, "/accountbudgetruntime") || path == "net/http" {
					alias := filepath.Base(path)
					if imp.Name != nil {
						alias = imp.Name.Name
					}
					if alias == "." {
						return fmt.Errorf("unresolved dot import in %s", rel)
					}
					if alias != "_" {
						aliases[alias] = path
					}
				}
			}
			for _, decl := range f.Decls {
				if _, ok := decl.(*ast.FuncDecl); ok {
					continue
				}
				var outside error
				ast.Inspect(decl, func(n ast.Node) bool {
					sel, ok := n.(*ast.SelectorExpr)
					if ok && wiringImportant(sel, aliases) {
						outside = fmt.Errorf("unclassified package-level selector in %s: %s", rel, sel.Sel.Name)
						return false
					}
					return true
				})
				if outside != nil {
					return outside
				}
			}
			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				var shadow error
				checkName := func(id *ast.Ident) {
					if id != nil && aliases[id.Name] != "" {
						if id.Name == "official" && ((rel == "internal/candidatesrc/candidatesrc.go" && fn.Name.Name == "Panel") || (rel == "internal/hybrid/client.go" && fn.Name.Name == "route") || (rel == "internal/mcp/server.go" && fn.Name.Name == "NewServer")) {
							return
						}
						shadow = fmt.Errorf("import alias shadow in %s/%s: %s", rel, fn.Name.Name, id.Name)
					}
				}
				ast.Inspect(fn, func(n ast.Node) bool {
					switch x := n.(type) {
					case *ast.Field:
						for _, name := range x.Names {
							checkName(name)
						}
					case *ast.ValueSpec:
						for _, name := range x.Names {
							checkName(name)
						}
					case *ast.AssignStmt:
						if x.Tok == token.DEFINE {
							for _, lhs := range x.Lhs {
								if id, ok := lhs.(*ast.Ident); ok {
									checkName(id)
								}
							}
						}
					case *ast.RangeStmt:
						if x.Tok == token.DEFINE {
							if id, ok := x.Key.(*ast.Ident); ok {
								checkName(id)
							}
							if id, ok := x.Value.(*ast.Ident); ok {
								checkName(id)
							}
						}
					}
					return shadow == nil
				})
				if shadow != nil {
					return shadow
				}
				recv := wiringRecv(fn)
				key := rel + "|" + recv + "|" + fn.Name.Name
				if expected, ok := c.shapes[key]; ok {
					fn.Doc = nil
					actual, e := wiringFormat(fset, fn)
					if e != nil {
						return e
					}
					if actual != expected {
						return fmt.Errorf("reviewed function shape changed: %s", key)
					}
					seenShapes[key] = true
				}
				called := map[*ast.SelectorExpr]bool{}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					sel, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					called[sel] = true
					if wiringImportant(sel, aliases) {
						formatted, e := wiringFormat(fset, sel)
						if e != nil {
							walkErr = e
							return false
						}
						prefix := recv
						if prefix != "" {
							prefix += "."
						}
						sites = append(sites, rel+"|"+prefix+fn.Name.Name+"|"+formatted)
					}
					return true
				})
				if walkErr != nil {
					return walkErr
				}
				var unresolved error
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					sel, ok := n.(*ast.SelectorExpr)
					if !ok || !wiringImportant(sel, aliases) || called[sel] {
						return true
					}
					if sel.Sel.Name == "Budget" || sel.Sel.Name == "NewClient" || sel.Sel.Name == "RoundTrip" {
						formatted, e := wiringFormat(fset, sel)
						if e != nil {
							unresolved = e
							return false
						}
						prefix := recv
						if prefix != "" {
							prefix += "."
						}
						sites = append(sites, rel+"|"+prefix+fn.Name.Name+"|VALUE:"+formatted)
						return true
					}
					unresolved = fmt.Errorf("unresolved function value: %s %s", rel, sel.Sel.Name)
					return false
				})
				if unresolved != nil {
					return unresolved
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	for key := range c.shapes {
		if !seenShapes[key] {
			return fmt.Errorf("reviewed function missing: %s", key)
		}
	}
	sort.Strings(sites)
	expected := make([]string, len(c.sites))
	for i, s := range c.sites {
		expected[i] = s.Site
	}
	sort.Strings(expected)
	if len(sites) != len(expected) {
		return fmt.Errorf("callsite inventory count changed: got %d expected %d", len(sites), len(expected))
	}
	for i := range sites {
		if sites[i] != expected[i] {
			return fmt.Errorf("unclassified callsite: got %s expected %s", sites[i], expected[i])
		}
	}
	return nil
}
func TestA166WiringCensusBaseline(t *testing.T) {
	if err := wiringInspect("../..", loadWiringCorpus(t), nil); err != nil {
		t.Fatal(err)
	}
}

// Each mutant is written to an isolated source copy and checked without executing it.
func TestA166WiringCensusRejectsBypasses(t *testing.T) {
	corpus := loadWiringCorpus(t)
	root := "../.."
	cases := []struct{ name, path, old, new, expect string }{
		{"root-error-to-off", "cmd/tossctl/root_budget.go", "factory, err := root.budgetOwner.get(ctx)\n\tif err != nil {\n\t\treturn nil, err\n\t}", "factory, err := root.budgetOwner.get(ctx)\n\tif err != nil {\n\t\treturn official.New(creds, tokenFile, opts...), nil\n\t}", "reviewed function shape changed"},
		{"required-raw-constructor", "internal/accountbudgetruntime/runtime_linux.go", "return official.NewWithAccountBudget(creds, cache, f.budget, f.cleanup, opts...)", "return official.New(creds, cache, opts...), nil", "reviewed function shape changed"},
		{"token-default-http", "internal/official/token.go", "resp, err := m.hc.Do(req)", "resp, err := http.DefaultClient.Do(req)", "reviewed function shape changed"},
		{"replay-raw-template", "internal/execgw/replay_transport.go", "copyClient.Transport = guard", "copyClient.Transport = template", "reviewed function shape changed"},
		{"probe-raw-send", "tools/a087-market-sell-probe/main.go", "sendClient, err = newBudgetedSendClient(d.sendTimeout, d.budget, d.budgetCleanup, d.budgetTemplate)", "sendClient = newSendClient(d.sendTimeout); err = nil", "reviewed function shape changed"},
		{"dead-guard-raw-return", "internal/accountbudgetruntime/runtime_linux.go", "return official.NewWithAccountBudget(creds, cache, f.budget, f.cleanup, opts...)", "if false { _, _ = official.NewWithAccountBudget(creds, cache, f.budget, f.cleanup, opts...) }; return official.New(creds, cache, opts...), nil", "reviewed function shape changed"},
		{"new-unclassified-http", "internal/official/token.go", "\nfunc (m *tokenManager) exchange(", "\nfunc a166Unexpected(c *http.Client, r *http.Request) { _, _ = c.Do(r) }\nfunc (m *tokenManager) exchange(", "callsite inventory count changed"},
		{"escaped-http-method", "internal/official/token.go", "\nfunc (m *tokenManager) exchange(", "\nfunc a166Escape(m *tokenManager) { f := m.hc.Do; _ = f }\nfunc (m *tokenManager) exchange(", "unresolved function value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			original, err := os.ReadFile(filepath.Join(root, tc.path))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(original), tc.old) != 1 {
				t.Fatalf("mutation anchor is not unique: %s", tc.name)
			}
			dir := t.TempDir()
			copyPath := filepath.Join(dir, "source.go")
			mutated := strings.Replace(string(original), tc.old, tc.new, 1)
			if err = os.WriteFile(copyPath, []byte(mutated), 0600); err != nil {
				t.Fatal(err)
			}
			copyBytes, err := os.ReadFile(copyPath)
			if err != nil {
				t.Fatal(err)
			}
			err = wiringInspect(root, corpus, map[string][]byte{tc.path: copyBytes})
			if err == nil || !strings.Contains(err.Error(), tc.expect) {
				t.Fatalf("mutant %s not caught by %q: %v", tc.name, tc.expect, err)
			}
			t.Logf("source-copy mutation rejected: %v", err)
		})
	}
}

func TestA166WiringCensusRejectsDirectorySymlink(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"cmd", "internal", "tools"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "cmd", "linked-directory")); err != nil {
		t.Fatal(err)
	}
	err := wiringInspect(root, wiringCorpus{}, nil)
	if err == nil || !strings.Contains(err.Error(), "source symlink:") {
		t.Fatalf("directory symlink was not rejected: %v", err)
	}
}
