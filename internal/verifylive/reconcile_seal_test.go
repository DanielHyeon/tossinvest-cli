package verifylive

// reconcile_seal_test.go — a121 봉인 census 강화(A-RED 리뷰 P1-1).
//
// 앞선 census(reconcile_structure_test.go)는 쓰기 이름·type assertion·Broker/Client **반환**만 봤고, 리뷰어 변이
// M9(ReconcileParams 에 Broker 필드)·M10(`var _ = New` — 실행기 생성자 참조)이 살아남았다. 여기서는 세 겹으로 닫는다.
//
//  1. 이름 금지: verifylive 의 reconcile*.go 에서 식별자 Broker·Client·Runner·New·Options 를 필드·파라미터·참조 어디에도
//     쓰지 않는다(AST — 타입 정보 없이도 보이는 층).
//  2. 입력 타입 핀: ReconcileParams(·Account·Approval)의 필드 타입은 좁은 목록 안이다(reflect).
//  3. 도달 census(go/types): reconcile*.go 가 쓰는(호출·값 참조) 모든 함수·메서드는 reconcile*.go 에 정의됐거나 명시
//     허용 목록 안이고, 쓰는 패키지 수준 변수는 reconcile*.go 의 것이며, import 는 표준 라이브러리·official(Reconcile*
//     타입만)·attest(Mask 만)뿐이다. 이것이 glob 한계(reconcile*.go 밖 헬퍼로의 탈출)를 막는다 — 밖의 헬퍼를 부르면
//     그 이름이 허용 목록에 없어서 실패한다.
//
// cmd 쪽은 좁은 생성자 안에서 *official.Client 를 감싸야 하므로(M13) 이 census 의 대상이 아니다.

import (
	"bufio"
	"bytes"
	"context"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const verifylivePath = "github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"

// reconcileAllowedOutside 는 reconcile*.go 밖에 정의됐지만 대사 경로가 써도 되는 함수·메서드다(FullName).
// 리뷰 허용 목록(LoadEntries·PendingCleanup·M0Unsettled·attest.Mask·Digest)에 기록 통로 둘만 더한다.
//   - 읽기 통로: readRecordRaw(a121 GREEN 판정 2026-10-05 — 엄격 해독·원문 바이트 지문의 입력, os.ReadFile 잎은 record.go
//     소유 — TestNoAutomationBypassExists 의 경계).
//   - 추가 통로: 기록 파일 자체 잠금 lockedRecord 넷(codex CG-2·CG-3 수리 2026-10-05 — 처음의 Recorder 3이름
//     (OpenRecorder·Append·Close)을 대체: 같은 inode 의 다른 대사와 flock 으로 직렬화, 잠근 fd 로 읽은 바이트가 곧 쓸
//     파일, 한 번의 Write·fsync·read-back). Recorder 는 더 이상 허용하지 않는다.
var reconcileAllowedOutside = map[string]bool{
	verifylivePath + ".LoadEntries":                               true,
	verifylivePath + ".PendingCleanup":                            true,
	verifylivePath + ".M0Unsettled":                               true,
	verifylivePath + ".Digest":                                    true,
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest.Mask": true,
	verifylivePath + ".readRecordRaw":                             true,
	verifylivePath + ".lockRecordForAppend":                       true,
	"(*" + verifylivePath + ".lockedRecord).contents":             true,
	"(*" + verifylivePath + ".lockedRecord).appendLine":           true,
	"(*" + verifylivePath + ".lockedRecord).release":              true,
}

// reconcileBannedImports 는 표준 라이브러리여도 reconcile*.go 가 import 하면 안 되는 패키지다(A-RED 재검 P2-a) — 직접
// 네트워크·프로세스·시스템 호출은 ReconcileReader 를 우회하는 문이다. 빈 식별자 import(`import _ "net/http"`)도 센다.
var reconcileBannedImports = map[string]bool{"net": true, "net/http": true, "os/exec": true, "syscall": true}

// reconcileVerifyliveFiles 는 verifylive 의 reconcile*.go(비시험) 파일 이름이다.
func reconcileVerifyliveFiles(t *testing.T) map[string]bool {
	t.Helper()
	matches, err := filepath.Glob("reconcile*.go")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, m := range matches {
		if !strings.HasSuffix(m, "_test.go") {
			out[m] = true
		}
	}
	if len(out) == 0 {
		t.Fatal("no reconcile*.go source — the census would pass on an empty sample")
	}
	return out
}

// TestReconcileVerifyliveFilesNameNoBrokerRunnerOrConstructor 는 1층(이름 금지)이다.
func TestReconcileVerifyliveFilesNameNoBrokerRunnerOrConstructor(t *testing.T) {
	banned := map[string]bool{"Broker": true, "Client": true, "Runner": true, "New": true, "Options": true}
	for path := range reconcileVerifyliveFiles(t) {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && banned[id.Name] {
				t.Errorf("%s: reconcile source names %q — a Broker, a runner or its constructor in scope reopens a mutation path",
					fset.Position(id.Pos()), id.Name)
			}
			return true
		})
	}
}

// TestReconcileParamsCarryOnlyNarrowTypes 는 2층(입력 타입 핀)이다.
func TestReconcileParamsCarryOnlyNarrowTypes(t *testing.T) {
	allowed := map[reflect.Type]bool{
		reflect.TypeOf(""): true,
		reflect.TypeOf((*ReconcileReader)(nil)).Elem():                                true,
		reflect.TypeOf(ReconcileAccount{}):                                            true,
		reflect.TypeOf((*io.Writer)(nil)).Elem():                                      true,
		reflect.TypeOf(func(context.Context, ReconcileApproval) error { return nil }): true,
		reflect.TypeOf(func() time.Time { return time.Time{} }):                       true,
	}
	check := func(typ reflect.Type, ok func(reflect.Type) bool) {
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.Anonymous {
				t.Errorf("%s embeds %v — embedding promotes methods past the census", typ.Name(), f.Type)
			}
			if !ok(f.Type) {
				t.Errorf("%s.%s has type %v, outside the narrow set", typ.Name(), f.Name, f.Type)
			}
		}
	}
	check(reflect.TypeOf(ReconcileParams{}), func(ft reflect.Type) bool { return allowed[ft] })
	scalar := func(ft reflect.Type) bool { return ft.Kind() == reflect.String || ft.Kind() == reflect.Int }
	check(reflect.TypeOf(ReconcileAccount{}), scalar)
	check(reflect.TypeOf(ReconcileApproval{}), scalar)
}

// TestReconcileVerifyliveFilesReachOnlyAllowlistedCode 는 3층(go/types 도달 census)이다.
func TestReconcileVerifyliveFilesReachOnlyAllowlistedCode(t *testing.T) {
	fset, files, info := typeCheckVerifylive(t)
	ours := reconcileVerifyliveFiles(t)
	seen := 0
	for _, file := range files {
		name := filepath.Base(fset.Position(file.Pos()).Filename)
		if !ours[name] {
			continue
		}
		seen++
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			switch {
			case reconcileBannedImports[path]:
				t.Errorf("%s imports %s — the reconcile path reaches the broker only through ReconcileReader", name, path)
			case !strings.Contains(strings.SplitN(path, "/", 2)[0], "."): // 표준 라이브러리
			case path == "github.com/JungHoonGhae/tossinvest-cli/internal/official",
				path == "github.com/JungHoonGhae/tossinvest-cli/internal/attest":
			default:
				t.Errorf("%s imports %s", name, path)
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			obj := info.Uses[id]
			if obj == nil || obj.Pkg() == nil {
				return true
			}
			where := fset.Position(id.Pos())
			declFile := filepath.Base(fset.Position(obj.Pos()).Filename)
			pkg := obj.Pkg().Path()
			switch o := obj.(type) {
			case *types.Func:
				switch {
				case pkg == verifylivePath && ours[declFile]:
				case reconcileAllowedOutside[o.FullName()]:
				case !strings.Contains(strings.SplitN(pkg, "/", 2)[0], "."): // 표준 라이브러리
				default:
					t.Errorf("%s: reconcile source reaches %s, defined outside reconcile*.go and not allowlisted", where, o.FullName())
				}
			case *types.Var:
				if o.Parent() == obj.Pkg().Scope() && pkg == verifylivePath && !ours[declFile] {
					t.Errorf("%s: reconcile source uses package variable %s from %s", where, o.Name(), declFile)
				}
			case *types.TypeName:
				if pkg == "github.com/JungHoonGhae/tossinvest-cli/internal/official" && !strings.HasPrefix(o.Name(), "Reconcile") {
					t.Errorf("%s: reconcile source uses official.%s (only the Reconcile* read types are inside the seal)", where, o.Name())
				}
			}
			return true
		})
	}
	if seen != len(ours) {
		t.Fatalf("type-checked %d of %d reconcile files", seen, len(ours))
	}
}

// typeCheckVerifylive 는 verifylive 파일들을 export data 로 타입 검사한다 — tossos_testseams 태그로 읽어 생산 파일과
// seam 파일(reconcile_testseam.go)을 함께 덮는다(태그는 파일을 더할 뿐 생산 파일을 빼지 않는다).
func typeCheckVerifylive(t *testing.T) (*token.FileSet, []*ast.File, *types.Info) {
	t.Helper()
	cmd := exec.Command("go", "list", "-tags", "tossos_testseams", "-export", "-deps", "-f", "{{.ImportPath}}\t{{.Export}}", ".")
	cmd.Env = append(os.Environ(), "GOFLAGS=")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list -export: %v", err)
	}
	exports := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), "\t", 2)
		if len(parts) == 2 && parts[1] != "" {
			exports[parts[0]] = parts[1]
		}
	}
	ctx := build.Default
	ctx.BuildTags = append(append([]string(nil), ctx.BuildTags...), "tossos_testseams")
	bp, err := ctx.ImportDir(".", 0)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, name := range bp.GoFiles {
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	imp := importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		p, ok := exports[path]
		if !ok {
			return nil, os.ErrNotExist
		}
		return os.Open(p)
	})
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}}
	conf := types.Config{Importer: imp}
	if _, err := conf.Check(verifylivePath, fset, files, info); err != nil {
		t.Fatalf("type-checking verifylive: %v", err)
	}
	return fset, files, info
}
