package engine

// a092 착지 단위 ⑤ — 구조 핀(23.3 K3 · K5 · K6 · K7 · K18 · K19, 24.3 M2 · M4 · M6 · M7, 25라운드 보이스 B #2 · #3).
// 각 핀은 저장소 비시험 Go 파일을 AST 로 걸어 센다(문자열 · 주석은 세지 않음). 양성 대조: 같은 걸음이 시험 파일에서는 무엇인가를
// 찾아야 함 — 못 찾으면 계측기가 눈먼 것임.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type a092Site struct {
	file string
	fn   string // 둘러싼 함수(메서드면 Recv.Name)
	pos  token.Position
	call *ast.CallExpr
}

// a092Walk 는 저장소 Go 파일을 걷고 파일마다 visit 을 부름.
func a092Walk(t *testing.T, tests bool, visit func(path string, fset *token.FileSet, f *ast.File)) {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", "openspec", ".sdd":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") != tests {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return rerr
		}
		visit(filepath.ToSlash(rel), fset, f)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func a092FuncName(fd *ast.FuncDecl) string {
	if fd.Recv != nil && len(fd.Recv.List) == 1 {
		switch rt := fd.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if id, ok := rt.X.(*ast.Ident); ok {
				return id.Name + "." + fd.Name.Name
			}
		case *ast.Ident:
			return rt.Name + "." + fd.Name.Name
		}
	}
	return fd.Name.Name
}

// a092CallsNamed 는 메서드/함수 이름이 name 인 호출 자리(선택자 · 식별자 둘 다)를 모음. minArgs 로 좁힘.
func a092CallsNamed(t *testing.T, tests bool, name string, minArgs int) []a092Site {
	t.Helper()
	var sites []a092Site
	a092Walk(t, tests, func(path string, fset *token.FileSet, f *ast.File) {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				c, ok := n.(*ast.CallExpr)
				if !ok || len(c.Args) < minArgs {
					return true
				}
				var got string
				switch fn := c.Fun.(type) {
				case *ast.SelectorExpr:
					got = fn.Sel.Name
				case *ast.Ident:
					got = fn.Name
				}
				if got == name {
					sites = append(sites, a092Site{file: path, fn: a092FuncName(fd), pos: fset.Position(c.Pos()), call: c})
				}
				return true
			})
		}
	})
	return sites
}

func a092Fns(sites []a092Site) []string {
	var out []string
	for _, s := range sites {
		out = append(out, s.file+":"+s.fn)
	}
	sort.Strings(out)
	return out
}

// K6 · K7 · M4: critical 행을 쓰는 원장 입구 셋의 비시험 호출자 전수.
func TestA092TheCriticalRecordersAreEnumerated(t *testing.T) {
	for _, c := range []struct {
		name string
		want []string
	}{
		{"RecordAlert", []string{"internal/obs/record_only.go:Notifier.recordCritical"}},
		{"ClaimAlertForDelivery", []string{"internal/obs/notifier.go:Notifier.claimAndDeliver"}},
		// 입구 밖 기록자 — 삽입 전에 자기 사유를 잠금(아래 핀).
		{"EnqueueAlert", []string{"internal/execgw/replay.go:Gateway.parkAlert"}},
	} {
		if len(a092CallsNamed(t, true, c.name, 1)) == 0 {
			t.Fatalf("control: no test caller of %s found — the walk is blind", c.name)
		}
		got := a092Fns(a092CallsNamed(t, false, c.name, 1))
		if strings.Join(got, ",") != strings.Join(c.want, ",") {
			t.Errorf("non-test callers of %s = %v, want %v — a new critical recorder must use the notifier entry "+
				"or latch its own reason before inserting (a092 critical-record rule)", c.name, got, c.want)
		}
	}
}

// K6 금지 형태: 입구 밖 기록자(parkAlert)는 삽입 앞에서 자기 사유를 잠그고, 그 사이에 해제 호출이 없음.
func TestA092TheOutsideRecorderLatchesBeforeItInserts(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join("..", "..", "execgw", "replay.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "parkAlert" {
			continue
		}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if s, ok := c.Fun.(*ast.SelectorExpr); ok {
					switch s.Sel.Name {
					case "Block", "EnqueueAlert", "Clear", "ClearSymbol", "ClearSymbolReason":
						order = append(order, s.Sel.Name)
					}
				}
			}
			return true
		})
	}
	if strings.Join(order, ",") != "Block,EnqueueAlert" {
		t.Errorf("parkAlert call order = %v, want [Block EnqueueAlert] — the latch must come first and nothing may release it in between", order)
	}
}

// K19: Notifier.Flush(ctx) 는 비시험 호출자 0 — 전송 위에서 n.mu 를 쥐는 마지막 경로라 호출자가 생기면 exit 기록이 그 전송을 기다림.
// Flush 는 흔한 이름이라 인자가 하나 이상인 호출과, 호출 아닌 메서드 값 참조를 셈(bufio · http.Flusher 의 Flush() 는 인자 0).
func TestA092TheNotifierFlushHasNoProductionCaller(t *testing.T) {
	if len(a092CallsNamed(t, true, "Flush", 1)) == 0 {
		t.Fatal("control: no test caller of Flush(ctx) found — the walk is blind")
	}
	if got := a092Fns(a092CallsNamed(t, false, "Flush", 1)); len(got) != 0 {
		t.Errorf("Notifier.Flush has a production caller %v — it holds n.mu across the transport (a092 K19); "+
			"move its send outside the lock before wiring it", got)
	}
	var values []string
	a092Walk(t, false, func(path string, fset *token.FileSet, f *ast.File) {
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				// 호출 위치의 선택자는 위에서 셈 — 인자는 계속 걸음.
				for _, a := range x.Args {
					ast.Inspect(a, func(m ast.Node) bool {
						if s, ok := m.(*ast.SelectorExpr); ok && s.Sel.Name == "Flush" {
							values = append(values, fset.Position(s.Pos()).String())
						}
						return true
					})
				}
				return false
			case *ast.AssignStmt:
				for _, r := range x.Rhs {
					if s, ok := r.(*ast.SelectorExpr); ok && s.Sel.Name == "Flush" {
						values = append(values, fset.Position(s.Pos()).String())
					}
				}
			}
			return true
		})
	})
	if len(values) != 0 {
		t.Errorf("Flush is referenced as a value at %v — a method value is a caller the call census cannot see", values)
	}
}

// M6: flatten CLI 조립에는 복구 · 재생 · 알림기가 없음 — 프로세스 밖 기록자(parkAlert)에 도달하지 않음.
// 비시험 obs.Notifier 생성은 newNotifier 하나.
func TestA092TheFlattenAssemblyHasNoOutsideRecorder(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join("..", "..", "..", "cmd", "tossctl", "flatten.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			switch x.Sel.Name {
			case "Recovery", "Replayer", "ReplayInDoubt", "Notifier":
				t.Errorf("flatten.go references %s at %s — the flatten process would reach an outside critical recorder",
					x.Sel.Name, fset.Position(x.Pos()))
			}
		case *ast.KeyValueExpr:
			if k, ok := x.Key.(*ast.Ident); ok && (k.Name == "Recovery" || k.Name == "Replay") {
				t.Errorf("flatten.go wires %s at %s", k.Name, fset.Position(x.Pos()))
			}
		}
		return true
	})
	var builders []string
	a092Walk(t, false, func(path string, fset *token.FileSet, f *ast.File) {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				lit, ok := n.(*ast.CompositeLit)
				if !ok {
					return true
				}
				switch ty := lit.Type.(type) {
				case *ast.SelectorExpr:
					if x, ok := ty.X.(*ast.Ident); ok && x.Name == "obs" && ty.Sel.Name == "Notifier" {
						builders = append(builders, path+":"+a092FuncName(fd))
					}
				case *ast.Ident:
					if ty.Name == "Notifier" && strings.HasSuffix(path, "internal/obs/"+filepath.Base(path)) {
						builders = append(builders, path+":"+a092FuncName(fd))
					}
				}
				return true
			})
		}
	})
	sort.Strings(builders)
	if strings.Join(builders, ",") != "internal/app/engine/exitwiring.go:newNotifier" {
		t.Errorf("non-test obs.Notifier constructions = %v, want only newNotifier", builders)
	}
}

// K5 · M7: execgw.New 비시험 호출자 전수 — 각 Options 에 Entry 가 있고 값은 같은 함수의 NewEntryGate 결과. 엔진 조립에서는 그 게이트가
// newNotifier 의 게이트 인자와 같은 식별자.
func TestA092EveryGatewayGetsItsEntryGate(t *testing.T) {
	type found struct{ fn, entry, gateVar, notifierGate string }
	var sites []found
	a092Walk(t, false, func(path string, fset *token.FileSet, f *ast.File) {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			var gateVars []string
			notifierGate := ""
			var news []*ast.CallExpr
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.AssignStmt:
					if len(x.Rhs) == 1 {
						if c, ok := x.Rhs[0].(*ast.CallExpr); ok {
							if s, ok := c.Fun.(*ast.SelectorExpr); ok && s.Sel.Name == "NewEntryGate" {
								if id, ok := x.Lhs[0].(*ast.Ident); ok {
									gateVars = append(gateVars, id.Name)
								}
							}
						}
					}
				case *ast.CallExpr:
					if s, ok := x.Fun.(*ast.SelectorExpr); ok && s.Sel.Name == "New" {
						if pkg, ok := s.X.(*ast.Ident); ok && pkg.Name == "execgw" {
							news = append(news, x)
						}
					}
					if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "newNotifier" && len(x.Args) >= 2 {
						if g, ok := x.Args[1].(*ast.Ident); ok {
							notifierGate = g.Name
						}
					}
				}
				return true
			})
			for _, c := range news {
				entry := ""
				if len(c.Args) == 1 {
					if lit, ok := c.Args[0].(*ast.CompositeLit); ok {
						for _, el := range lit.Elts {
							if kv, ok := el.(*ast.KeyValueExpr); ok {
								if k, ok := kv.Key.(*ast.Ident); ok && k.Name == "Entry" {
									if v, ok := kv.Value.(*ast.Ident); ok {
										entry = v.Name
									} else {
										entry = "<not an identifier>"
									}
								}
							}
						}
					}
				}
				gv := strings.Join(gateVars, ",")
				sites = append(sites, found{fn: path + ":" + a092FuncName(fd), entry: entry, gateVar: gv, notifierGate: notifierGate})
			}
		}
	})
	if len(sites) != 2 {
		t.Fatalf("non-test execgw.New callers = %+v, want the engine assembly and the flatten CLI — a new caller must be pinned here", sites)
	}
	for _, s := range sites {
		if s.entry == "" || s.entry != s.gateVar {
			t.Errorf("%s: Entry = %q, want the NewEntryGate result of the same function (%q)", s.fn, s.entry, s.gateVar)
		}
		if strings.HasSuffix(s.fn, "internal/app/engine/gateway.go:buildGateway") && s.notifierGate != s.entry {
			t.Errorf("buildGateway: the notifier's gate %q is not the gateway's Entry %q (M7)", s.notifierGate, s.entry)
		}
	}
}

// K3 · M2: 전이 함수의 커밋 성공 경로 — Commit 을 담은 if 다음 문장부터 ProjectOperatingMode 호출까지 반환 · go 문이 없음.
// 투영기 몸체에도 go 문이 없음(「동시에」 = 같은 호출 안의 커밋 뒤 투영).
func TestA092CommitThenProjectInOneCall(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join("..", "..", "journal", "operating_mode.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body []ast.Stmt
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == "TransitionOperatingMode" {
			body = fd.Body.List
		}
	}
	commitAt, projectAt := -1, -1
	for i, st := range body {
		ast.Inspect(st, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if s, ok := c.Fun.(*ast.SelectorExpr); ok {
					if s.Sel.Name == "Commit" && commitAt < 0 {
						commitAt = i
					}
					if s.Sel.Name == "ProjectOperatingMode" && projectAt < 0 {
						projectAt = i
					}
				}
			}
			return true
		})
	}
	if commitAt < 0 || projectAt <= commitAt {
		t.Fatalf("Commit at statement %d, ProjectOperatingMode at %d — want the projection after the commit in the same body", commitAt, projectAt)
	}
	for _, st := range body[commitAt+1 : projectAt] {
		ast.Inspect(st, func(n ast.Node) bool {
			switch n.(type) {
			case *ast.ReturnStmt, *ast.GoStmt:
				t.Errorf("a return or go statement between the commit and the projection at %s", fset.Position(n.Pos()))
			}
			return true
		})
	}
	g, err := parser.ParseFile(fset, filepath.Join("..", "..", "execgw", "modegate.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(g, func(n ast.Node) bool {
		if _, ok := n.(*ast.GoStmt); ok {
			t.Errorf("the projector starts a goroutine at %s — the projection must finish inside the transition call", fset.Position(n.Pos()))
		}
		return true
	})
}

// K18 · 보이스 B #2: exit 관측기에 주입되는 발의자(Guardian)의 IssueReduction 은 모드 승격(announcer 경로)에 닿지 않음 —
// 같은 패키지 안 호출 폐포로 셈. 닿게 되면 exit goroutine 에 동기 통지가 생김.
func TestA092TheExitIssuerDoesNotReachAnAnnouncer(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, filepath.Join("..", "..", "execgw"), func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	calls := map[string][]string{}
	for _, p := range pkgs {
		for _, f := range p.Files {
			for _, d := range f.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok || fd.Body == nil {
					continue
				}
				name := fd.Name.Name
				ast.Inspect(fd.Body, func(n ast.Node) bool {
					switch x := n.(type) {
					case *ast.CallExpr:
						switch fn := x.Fun.(type) {
						case *ast.SelectorExpr:
							calls[name] = append(calls[name], fn.Sel.Name)
						case *ast.Ident:
							calls[name] = append(calls[name], fn.Name)
						}
					case *ast.SelectorExpr:
						if x.Sel.Name == "announcer" || x.Sel.Name == "Announcer" {
							calls[name] = append(calls[name], "<announcer>")
						}
					}
					return true
				})
			}
		}
	}
	seen := map[string]bool{"IssueReduction": true}
	queue := []string{"IssueReduction"}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, callee := range calls[cur] {
			switch callee {
			case "<announcer>", "escalateFor", "EscalateOperatingMode", "AnnounceOperatingMode":
				t.Errorf("IssueReduction reaches %s through %s — the exit goroutine would announce synchronously", callee, cur)
			}
			if _, ok := calls[callee]; ok && !seen[callee] {
				seen[callee] = true
				queue = append(queue, callee)
			}
		}
	}
	if len(seen) < 2 {
		t.Fatal("control: IssueReduction's call closure is empty — the walk is blind")
	}
}

// k3 · 보이스 B #3: 청산 수량 상한 조회의 브로커 호출(f.official.*)은 전부 f.retrier.Query 에 넘긴 함수 리터럴 안에 있음 —
// 그래서 그 조회의 401 이 exit Retrier(기록 전용 통지자)를 탐. 우회 호출이 생기면 이 합성 주장이 깨짐.
func TestA092TheFloorQueriesOnlyThroughTheRetrier(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "exitwiring.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && a092FuncName(fd) == "reconcileFloor.ConfirmedFloor" {
			body = fd.Body
		}
	}
	if body == nil {
		t.Fatal("reconcileFloor.ConfirmedFloor not found")
	}
	inside := map[token.Pos]bool{}
	queries := 0
	ast.Inspect(body, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		s, ok := c.Fun.(*ast.SelectorExpr)
		if !ok || s.Sel.Name != "Query" {
			return true
		}
		if r, ok := s.X.(*ast.SelectorExpr); !ok || r.Sel.Name != "retrier" {
			return true
		}
		queries++
		for _, a := range c.Args {
			if lit, ok := a.(*ast.FuncLit); ok {
				ast.Inspect(lit, func(m ast.Node) bool {
					if m != nil {
						inside[m.Pos()] = true
					}
					return true
				})
			}
		}
		return true
	})
	if queries == 0 {
		t.Fatal("control: no f.retrier.Query call found — the pin measures nothing")
	}
	ast.Inspect(body, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if s, ok := c.Fun.(*ast.SelectorExpr); ok {
			if r, ok := s.X.(*ast.SelectorExpr); ok && r.Sel.Name == "official" && !inside[c.Pos()] {
				t.Errorf("f.official.%s at %s is called outside f.retrier.Query — its 401 would bypass the exit Retrier",
					s.Sel.Name, fset.Position(c.Pos()))
			}
		}
		return true
	})
}
