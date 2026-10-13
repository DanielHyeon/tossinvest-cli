package redact_test

// a131_structure_test.go — 새 싱크가 규칙을 우회하면 실패하는 구조 시험(go/parser, 저장소 전체 비시험 Go).
//
// 금지를 이름으로 적음(존재가 아니라 역할):
//   - slog 핸들러·로거를 obs/redact 밖에서 만들지 않음 — 만들면 그 로거는 계좌 키 버림·치환을 안 거침.
//   - 표준 log 패키지·slog 패키지 수준 함수(slog.Default 등)를 쓰지 않음 — 허용 목록의 자리만(이유와 함께).
//   - O_APPEND 로 파일을 여는 자리는 허용 목록뿐이고, 기록·감사 쓰기 파일은 redact 규칙을 부름.
//   - 등록 입구(accountsLocked)·obs 로거(NewLogger)·엔진 출력(newEngineRunCmd)·마지막 오류 줄(main)이 규칙을 부름.

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

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root not found at %s: %v", root, err)
	}
	return root
}

type goFile struct {
	rel  string
	file *ast.File
	fset *token.FileSet
}

// productionFiles 는 cmd/ · internal/ 아래 비시험 Go 파일 전부를 읽음(tools/ · openspec/ 제외 — 운영 프로세스가 아님).
func productionFiles(t *testing.T) []goFile {
	t.Helper()
	root := repoRoot(t)
	var out []goFile
	for _, top := range []string{"cmd", "internal"} {
		err := filepath.Walk(filepath.Join(root, top), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			fset := token.NewFileSet()
			f, perr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
			if perr != nil {
				return perr
			}
			rel, _ := filepath.Rel(root, path)
			out = append(out, goFile{rel: filepath.ToSlash(rel), file: f, fset: fset})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(out) < 500 {
		t.Fatalf("census read only %d files — the walker is blind", len(out))
	}
	return out
}

func selectorCalls(f *ast.File) []string {
	var calls []string
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if s, ok := c.Fun.(*ast.SelectorExpr); ok {
				if id, ok := s.X.(*ast.Ident); ok {
					calls = append(calls, id.Name+"."+s.Sel.Name)
				}
			}
		}
		return true
	})
	return calls
}

func funcBody(t *testing.T, files []goFile, rel, recv, name string) *ast.FuncDecl {
	t.Helper()
	for _, g := range files {
		if g.rel != rel {
			continue
		}
		for _, d := range g.file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Name.Name != name {
				continue
			}
			r := ""
			if fd.Recv != nil && len(fd.Recv.List) > 0 {
				switch x := fd.Recv.List[0].Type.(type) {
				case *ast.StarExpr:
					if id, ok := x.X.(*ast.Ident); ok {
						r = id.Name
					}
				case *ast.Ident:
					r = x.Name
				}
			}
			if r == recv {
				return fd
			}
		}
	}
	t.Fatalf("%s: function %s.%s not found", rel, recv, name)
	return nil
}

func calls(n ast.Node) map[string]int {
	out := map[string]int{}
	ast.Inspect(n, func(x ast.Node) bool {
		if c, ok := x.(*ast.CallExpr); ok {
			if s, ok := c.Fun.(*ast.SelectorExpr); ok {
				if id, ok := s.X.(*ast.Ident); ok {
					out[id.Name+"."+s.Sel.Name]++
				}
			}
		}
		return true
	})
	return out
}

func TestA131NoLoggerOrAppendWriterBypassesRedaction(t *testing.T) {
	files := productionFiles(t)

	// slog 패키지 수준 기본 로거 — 계좌가 없는 측정 도구 하나만(엔진·soak·verify 경로 밖).
	slogDefaultAllowed := map[string]string{
		"internal/measure/degrade/degrade.go": "측정 도구의 nil 로거 대체 — 엔진·soak·verify 의존 그래프 밖",
	}
	appendAllowed := map[string]string{
		"internal/soak/record.go":              "soak 기록 — redact.SealRecordLine",
		"internal/verifylive/record.go":        "검증 기록 — redact.SealRecordLine",
		"internal/audit/audit.go":              "감사 기록 — RedactEntry(redact.String)",
		"cmd/tossctl/engineproc.go":            "엔진 자식 stdout/stderr — 자식 안에서 치환(newEngineRunCmd · main · obs)",
		"cmd/tossctl/soakproc.go":              "soak 자식 stdout/stderr — 자식 안에서 치환(main)",
		"cmd/tossctl/scrub_account_records.go": "a131 정리 명령의 교체 창 꼬리 옮기기 — 같은 규칙",
		"internal/soak/account_seal.go":        "교체된 soak 기록 다시 열기 — 쓰기는 Recorder.Append(SealRecordLine)",
		"internal/verifylive/account_seal.go":  "교체된 검증 기록 다시 열기 — 쓰기는 Recorder.Append(SealRecordLine)",
	}
	var problems []string
	for _, g := range files {
		for _, imp := range g.file.Imports {
			if imp.Path.Value == `"log"` {
				problems = append(problems, g.rel+": imports the standard log package (bypasses redaction)")
			}
		}
		inObsOrRedact := strings.HasPrefix(g.rel, "internal/obs/") || strings.HasPrefix(g.rel, "internal/redact/")
		for _, c := range selectorCalls(g.file) {
			switch c {
			case "slog.New", "slog.NewJSONHandler", "slog.NewTextHandler":
				if !inObsOrRedact {
					problems = append(problems, g.rel+": builds a slog logger/handler outside obs/redact ("+c+")")
				}
			case "slog.Default", "slog.SetDefault", "slog.Info", "slog.Warn", "slog.Error", "slog.Debug", "slog.Log":
				if _, ok := slogDefaultAllowed[g.rel]; !ok {
					problems = append(problems, g.rel+": uses the package-level slog logger ("+c+")")
				}
			}
		}
		ast.Inspect(g.file, func(n ast.Node) bool {
			s, ok := n.(*ast.SelectorExpr)
			if !ok || s.Sel.Name != "O_APPEND" {
				return true
			}
			if id, ok := s.X.(*ast.Ident); ok && id.Name == "os" {
				if _, ok := appendAllowed[g.rel]; !ok {
					problems = append(problems, g.rel+": opens a file O_APPEND outside the classified writers — classify it in a131's census")
				}
			}
			return true
		})
	}

	// 기록·감사 쓰기 파일이 규칙을 부름.
	want := map[string]string{
		"internal/soak/record.go":              "redact.SealRecordLine",
		"internal/verifylive/record.go":        "redact.SealRecordLine",
		"internal/audit/redact.go":             "redact.String",
		"internal/official/account_redact.go":  "redact.Register",
		"cmd/tossctl/scrub_account_records.go": "redact.SealRecordLine",
	}
	for _, g := range files {
		if call, ok := want[g.rel]; ok {
			if calls(g.file)[call] == 0 {
				problems = append(problems, g.rel+": does not call "+call)
			}
			delete(want, g.rel)
		}
	}
	for rel := range want {
		problems = append(problems, rel+": file not found")
	}
	sort.Strings(problems)
	if len(problems) != 0 {
		t.Fatalf("redaction bypasses:\n  %s", strings.Join(problems, "\n  "))
	}
}

func TestA131TheSinksCallTheRule(t *testing.T) {
	files := productionFiles(t)
	cases := []struct {
		rel, recv, name, call string
		min                   int
	}{
		{"internal/obs/log.go", "", "NewLogger", "redact.Handler", 1},
		{"internal/official/reads.go", "Client", "accountsLocked", "registerAccountNumbers", 1},
		{"cmd/tossctl/engine.go", "", "newEngineRunCmd", "redact.Writer", 2},
		{"cmd/tossctl/main.go", "", "main", "redact.Writer", 2},
		{"internal/audit/audit.go", "Log", "Record", "RedactEntry", 0},
		{"internal/soak/record.go", "Recorder", "Append", "redact.SealRecordLine", 1},
		{"internal/verifylive/record.go", "Recorder", "Append", "redact.SealRecordLine", 1},
		{"internal/verifylive/record.go", "lockedRecord", "appendLine", "redact.SealRecordLine", 1},
		{"internal/soak/record.go", "", "OpenRecorder", "redact.PrepareSealKey", 1},
		{"internal/verifylive/record.go", "", "OpenRecorder", "redact.PrepareSealKey", 1},
		{"cmd/tossctl/main.go", "", "main", "root.SetErr", 1},
	}
	for _, c := range cases {
		fd := funcBody(t, files, c.rel, c.recv, c.name)
		got := calls(fd)
		if !strings.Contains(c.call, ".") {
			found := 0
			ast.Inspect(fd, func(n ast.Node) bool {
				if ce, ok := n.(*ast.CallExpr); ok {
					if id, ok := ce.Fun.(*ast.Ident); ok && id.Name == c.call {
						found++
					}
				}
				return true
			})
			if found < c.min || (c.min == 0 && found == 0) {
				t.Errorf("%s %s.%s calls %s %d time(s)", c.rel, c.recv, c.name, c.call, found)
			}
			continue
		}
		if c.call == "RedactEntry" {
			found := false
			ast.Inspect(fd, func(n ast.Node) bool {
				if ce, ok := n.(*ast.CallExpr); ok {
					if id, ok := ce.Fun.(*ast.Ident); ok && id.Name == "RedactEntry" {
						found = true
					}
				}
				return true
			})
			if !found {
				t.Errorf("%s %s.%s does not call RedactEntry", c.rel, c.recv, c.name)
			}
			continue
		}
		if got[c.call] < c.min {
			t.Errorf("%s %s.%s calls %s %d time(s), want >= %d", c.rel, c.recv, c.name, c.call, got[c.call], c.min)
		}
	}
	// main 의 마지막 오류 줄은 감싼 쓰기 대상으로만 — os.Stderr 에 직접 쓰는 Fprint 가 남으면 우회.
	fd := funcBody(t, files, "cmd/tossctl/main.go", "", "main")
	ast.Inspect(fd, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if s, ok := ce.Fun.(*ast.SelectorExpr); ok && strings.HasPrefix(s.Sel.Name, "Fprint") && len(ce.Args) > 0 {
			if a, ok := ce.Args[0].(*ast.SelectorExpr); ok && a.Sel.Name == "Stderr" {
				t.Errorf("main writes to os.Stderr directly at %v", ce.Pos())
			}
		}
		return true
	})
}

// TestA131ReadersOpenTheSealOnEveryDecodedLine 은 읽는 쪽 셋이 해독한 줄마다 봉인을 여는지 셈(LoadCycles · LoadEntries 는
// flush 와 마지막 줄 두 자리, 대사 엄격 해독은 한 자리) — 한 자리를 빠뜨리면 그 줄의 계좌가 빈 값이 됨.
func TestA131ReadersOpenTheSealOnEveryDecodedLine(t *testing.T) {
	files := productionFiles(t)
	cases := []struct {
		rel, name, call string
		want            int
	}{
		{"internal/soak/record.go", "LoadCycles", "openCycleAccount", 2},
		{"internal/verifylive/record.go", "LoadEntries", "openEntryAccount", 2},
		{"internal/verifylive/reconcile_record.go", "readRecordStrictNoTail", "openRecordEntries", 1},
		{"internal/verifylive/record.go", "openRecordEntries", "openEntryAccount", 1},
	}
	for _, c := range cases {
		fd := funcBody(t, files, c.rel, "", c.name)
		got := 0
		ast.Inspect(fd, func(n ast.Node) bool {
			if ce, ok := n.(*ast.CallExpr); ok {
				if id, ok := ce.Fun.(*ast.Ident); ok && id.Name == c.call {
					got++
				}
			}
			return true
		})
		if got != c.want {
			t.Errorf("%s %s calls %s %d time(s), want %d", c.rel, c.name, c.call, got, c.want)
		}
	}
}
