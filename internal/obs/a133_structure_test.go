package obs_test

// a133_structure_test.go — 경계를 우회하는 전송 수단이 생기면 실패하는 구조 시험(go/parser, cmd/ · internal/ 비시험 Go 전부, design D2).
//
// 금지를 이름으로 적음:
//   ① `Publish(context.Context, Notification) error` 꼴 메서드를 가진 생산 타입은 obs.Ntfy 하나뿐 — 새 전송 수단은 여기서 실패하고,
//      그 수단도 redactForTransport 를 부르게 고친 뒤 이 목록을 늘림.
//   ② Ntfy.Publish 의 **첫 문장**이 msg = redactForTransport(msg) — 요청 구성(http.NewRequestWithContext)보다 앞.
//   ③ renderFields 가 redact.IsAccountKey 를 부름(필드 이름은 펼치기 전에만 앎).
//   ④ redactForTransport 가 redact.String 을 부름(규칙 사본 금지 — 규칙은 internal/redact 하나).
//   ⑤ internal/obs 안에서 net/http 를 import 하는 파일은 ntfy.go 하나.
// 이 핀이 증명하는 범위: obs.Publisher 를 채우는 생산 타입(별칭 import · 타입 별칭 포함)과 obs 패키지 안의 전송. obs 밖에서
// 알림 문구를 http 로 직접 보내는 코드는 이 핀 밖이다(spec 시나리오는 이 범위로 좁힘 — 리뷰 F-2).

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

func a133RepoRoot(t *testing.T) string {
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

type a133File struct {
	rel  string
	file *ast.File
}

func a133ProductionFiles(t *testing.T) []a133File {
	t.Helper()
	root := a133RepoRoot(t)
	var out []a133File
	for _, top := range []string{"cmd", "internal"} {
		err := filepath.Walk(filepath.Join(root, top), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			f, perr := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
			if perr != nil {
				return perr
			}
			rel, _ := filepath.Rel(root, path)
			out = append(out, a133File{rel: filepath.ToSlash(rel), file: f})
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

// a133TypeName 은 타입 식의 끝 이름(`Notification` · `obs.Notification` · `context.Context` → 마지막 식별자, 패키지 포함)을 돌려줌.
func a133TypeName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		if id, ok := x.X.(*ast.Ident); ok {
			return id.Name + "." + x.Sel.Name
		}
	case *ast.StarExpr:
		return a133TypeName(x.X)
	}
	return ""
}

func a133RecvName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	return a133TypeName(fd.Recv.List[0].Type)
}

// a133IsPublisherMethod 는 obs.Publisher 를 채우는 메서드 모양인지 봄 — 매개변수 둘의 둘째가 Notification(obs 안) 또는 obs.Notification.
func a133IsPublisherMethod(f a133File, fd *ast.FuncDecl) bool {
	if fd.Recv == nil || fd.Name.Name != "Publish" || fd.Type.Params == nil {
		return false
	}
	var params []ast.Expr
	for _, p := range fd.Type.Params.List {
		n := len(p.Names)
		if n == 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			params = append(params, p.Type)
		}
	}
	if len(params) != 2 || a133TypeName(params[0]) != "context.Context" {
		return false
	}
	return a133IsNotificationType(f, params[1])
}

// a133ObsAliases 는 파일이 internal/obs 를 어떤 이름으로 import 하는지 돌려줌(별칭 포함 — 리뷰 F-2).
func a133ObsAliases(f a133File) map[string]bool {
	names := map[string]bool{}
	for _, imp := range f.file.Imports {
		if strings.Trim(imp.Path.Value, `"`) != "github.com/JungHoonGhae/tossinvest-cli/internal/obs" {
			continue
		}
		name := "obs"
		if imp.Name != nil {
			name = imp.Name.Name
		}
		names[name] = true
	}
	return names
}

// a133IsNotificationType 는 타입 식이 obs.Notification 인지 봄 — obs 안의 `Notification`, 별칭 import 의 `<별칭>.Notification`.
func a133IsNotificationType(f a133File, e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name == "Notification" && f.file.Name.Name == "obs"
	case *ast.SelectorExpr:
		id, ok := x.X.(*ast.Ident)
		return ok && x.Sel.Name == "Notification" && a133ObsAliases(f)[id.Name]
	}
	return false
}

func a133Func(t *testing.T, files []a133File, rel, recv, name string) *ast.FuncDecl {
	t.Helper()
	for _, f := range files {
		if f.rel != rel {
			continue
		}
		for _, d := range f.file.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name && a133RecvName(fd) == recv {
				return fd
			}
		}
	}
	t.Fatalf("%s: function %s.%s not found", rel, recv, name)
	return nil
}

func a133Calls(n ast.Node) []string {
	var calls []string
	ast.Inspect(n, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			switch fn := c.Fun.(type) {
			case *ast.Ident:
				calls = append(calls, fn.Name)
			case *ast.SelectorExpr:
				if id, ok := fn.X.(*ast.Ident); ok {
					calls = append(calls, id.Name+"."+fn.Sel.Name)
				}
			}
		}
		return true
	})
	return calls
}

func a133Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestA133TheOnlyTransportPassesTheBoundary(t *testing.T) {
	files := a133ProductionFiles(t)

	// ① 생산 발행자 전수.
	var publishers []string
	for _, f := range files {
		for _, d := range f.file.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && a133IsPublisherMethod(f, fd) {
				publishers = append(publishers, f.rel+":"+a133RecvName(fd))
			}
		}
	}
	// 같은 모양을 다른 이름으로 만드는 길(타입 별칭 · 정의 타입)도 막음 — 그 이름으로 Publish 를 만들면 ① 이 못 셈(리뷰 F-2).
	for _, f := range files {
		ast.Inspect(f.file, func(n ast.Node) bool {
			if ts, ok := n.(*ast.TypeSpec); ok && a133IsNotificationType(f, ts.Type) {
				t.Errorf("%s: type %s renames obs.Notification — a transport under that name escapes the census", f.rel, ts.Name.Name)
			}
			return true
		})
	}
	sort.Strings(publishers)
	if strings.Join(publishers, " ") != "internal/obs/ntfy.go:Ntfy" {
		t.Fatalf("production notification transports = %v, want only internal/obs/ntfy.go:Ntfy — a new transport must call redactForTransport before any byte leaves and be added here", publishers)
	}

	// ② Ntfy.Publish 의 첫 문장이 치환 대입이고, 요청 구성이 그 뒤.
	pub := a133Func(t, files, "internal/obs/ntfy.go", "Ntfy", "Publish")
	if len(pub.Body.List) == 0 {
		t.Fatal("Ntfy.Publish has no body")
	}
	first, ok := pub.Body.List[0].(*ast.AssignStmt)
	if !ok || len(first.Lhs) != 1 || len(first.Rhs) != 1 {
		t.Fatalf("Ntfy.Publish's first statement is not `msg = redactForTransport(msg)`")
	}
	lhs, _ := first.Lhs[0].(*ast.Ident)
	call, _ := first.Rhs[0].(*ast.CallExpr)
	param := pub.Type.Params.List[len(pub.Type.Params.List)-1].Names[0].Name
	if lhs == nil || lhs.Name != param || first.Tok != token.ASSIGN || call == nil {
		t.Fatalf("Ntfy.Publish's first statement does not reassign its notification parameter %q", param)
	}
	if fn, ok := call.Fun.(*ast.Ident); !ok || fn.Name != "redactForTransport" || len(call.Args) != 1 {
		t.Fatalf("Ntfy.Publish's first statement does not call redactForTransport")
	}
	if arg, ok := call.Args[0].(*ast.Ident); !ok || arg.Name != param {
		t.Fatalf("redactForTransport is not given the notification parameter %q", param)
	}
	if !a133Contains(a133Calls(pub), "http.NewRequestWithContext") {
		t.Fatal("Ntfy.Publish no longer builds its request with http.NewRequestWithContext — re-check the boundary")
	}

	// ③ renderFields 가 계좌 키를 판정함.
	if !a133Contains(a133Calls(a133Func(t, files, "internal/obs/notifier.go", "", "renderFields")), "redact.IsAccountKey") {
		t.Fatal("renderFields does not call redact.IsAccountKey — account-key fields reach the notification body")
	}

	// ⑤ obs 패키지 안에서 네트워크로 나가는 파일은 ntfy.go 하나뿐(net/http import) — obs 안의 새 전송 수단은 여기서 실패(리뷰 F-2).
	var httpFiles []string
	for _, f := range files {
		if !strings.HasPrefix(f.rel, "internal/obs/") {
			continue
		}
		for _, imp := range f.file.Imports {
			if strings.Trim(imp.Path.Value, `"`) == "net/http" {
				httpFiles = append(httpFiles, f.rel)
			}
		}
	}
	if strings.Join(httpFiles, " ") != "internal/obs/ntfy.go" {
		t.Fatalf("internal/obs files importing net/http = %v, want only internal/obs/ntfy.go", httpFiles)
	}

	// ④ 치환 함수는 a131 규칙을 부름.
	if !a133Contains(a133Calls(a133Func(t, files, "internal/obs/a133_transport_redact.go", "", "redactForTransport")), "redact.String") {
		t.Fatal("redactForTransport does not call redact.String")
	}
}
