//go:build unix

package strategyprojectionrpc

// a113 — projection probe 도 사망을 **증명**한다.
//
// a109 §1-fix F1 은 형제 endpoint 의 회수에서 「owner 쓰기 비트가 없으면 죽었다」는 추정을
// chmod-then-probe 로 바꿨다. 그 추정의 원형인 이 패키지(a108)의 probe 에는 같은 절이
// 남아 있었다(a109 issues I1). 여기 있는 테스트는 그 절이 만드는 사고 모양 — **쓰기 비트가
// 깎인 산 socket 을 지우고 그 위에 두 번째 서버를 세운다** — 을 디스크에서 재현한다.
//
// root 는 DAC 를 우회해 EACCES 가 나지 않으므로 이 모양 자체가 성립하지 않는다(a108 관례대로 Skip).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTheReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped 는 a109
// `TestReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped` 의 원형판이다.
//
// 재는 것은 셋이다 — ① 기동이 **거부**되는가, ② 거부하면서 주인의 socket·descriptor 를
// **그대로 두는가**(거부했지만 이미 지웠다가 통과하면 안 된다), ③ 그 자리의 socket 이
// **같은 파일**이고 여전히 수락하는가(지우고 새로 세운 것과 구별한다).
func TestTheReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root 는 DAC 를 우회하므로 쓰기 비트가 깎인 socket 도 EACCES 를 만들지 않는다")
	}
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	listener := a108LiveSocket(t, dir)
	a108WriteLeftoverDescriptor(t, dir, 1<<30)
	if err := os.Chmod(SocketPath(dir), 0o400); err != nil {
		t.Fatal(err)
	}
	before, err := os.Lstat(SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}

	server, err := Start(dir, a108Reader())
	if err == nil {
		_ = server.Close()
		t.Fatal("쓰기 비트가 깎인 산 socket 위에서 기동이 받아들여졌다 — 산 주인의 endpoint 를 탈취했다")
	}
	// 거부 사유까지 본다(freeze P2-3). 다른 이유(모양 검사 등)로 거부했다면 이 테스트는
	// 사망 판정을 재지 않은 것이다.
	if !strings.Contains(err.Error(), "still alive") {
		t.Fatalf("거부 사유가 주인의 생존이 아니다: %v", err)
	}
	for _, path := range []string{SocketPath(dir), DescriptorPath(dir)} {
		if _, statErr := os.Lstat(path); statErr != nil {
			t.Fatalf("거부하면서 %s 를 치웠다: %v", path, statErr)
		}
	}

	after, err := os.Lstat(SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) {
		t.Fatal("거부했다지만 그 자리의 socket 이 다른 파일이다 — 산 주인의 socket 을 지우고 새로 세웠다")
	}
	// probe 는 권한을 **정확히 0600** 으로 되돌린다(freeze P1-3). 0666 같은 모드면 산 주인의
	// socket 이 group/other 쓰기로 남는다 — 이 단정을 테스트 자신의 chmod 보다 먼저 둔다.
	if got := after.Mode().Perm(); got != 0o600 {
		t.Fatalf("probe 뒤 산 socket 의 권한 = %#o, want 0600 (발행 계약)", got)
	}
	conn, err := net.Dial("unix", SocketPath(dir))
	if err != nil {
		t.Fatalf("주인의 socket 이 더는 수락하지 않는다: %v", err)
	}
	_ = conn.Close()
	accepted, err := listener.Accept()
	if err != nil {
		t.Fatalf("연결이 주인의 listener 에 도착하지 않았다: %v", err)
	}
	_ = accepted.Close()
}

// a113ProbeInfo 는 회수가 넘기는 것과 같은 값 — 검증 시점의 FileInfo — 을 만든다.
func a113ProbeInfo(t *testing.T, path string) os.FileInfo {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

// TestTheStaleProbeAsksInsteadOfGuessing 은 회수 전용 probe 의 판정 표다(a109
// `TestThePrivateSocketProbeReadsOnlyTwoThingsAsDead` 원형판).
//
// 사망으로 읽는 것은 **둘뿐**이다 — 연결 거부와 파일 부재. 쓰기 비트가 깎인 잔재는 chmod
// 0600 뒤에 물으므로 EACCES 를 만나지 않고 거부/수락으로 결정적으로 갈린다.
func TestTheStaleProbeAsksInsteadOfGuessing(t *testing.T) {
	unprivileged := func(t *testing.T) {
		t.Helper()
		if os.Geteuid() == 0 {
			t.Skip("root 는 DAC 를 우회하므로 EACCES 를 만들 수 없다")
		}
	}
	for _, test := range []struct {
		name  string
		build func(t *testing.T, dir string) (string, os.FileInfo)
		want  bool
		// wantMode 가 0 이 아니면 probe 뒤 그 경로의 권한이 정확히 이 값이어야 한다.
		wantMode os.FileMode
	}{
		// 검증한 뒤 사라졌다 — 주인이 자기 Close 로 지운 순간과 같다. chmod 가 ErrNotExist 로
		// 갈라내고, 부재는 그 자체가 사망 판정이다.
		{"검증 뒤 사라진 파일은 사망", func(t *testing.T, dir string) (string, os.FileInfo) {
			a108MakeControlDir(t, dir)
			a108DeadSocketWithMode(t, dir, 0o600)
			info := a113ProbeInfo(t, SocketPath(dir))
			if err := os.Remove(SocketPath(dir)); err != nil {
				t.Fatal(err)
			}
			return SocketPath(dir), info
		}, false, 0},
		{"연결 거부는 사망", func(t *testing.T, dir string) (string, os.FileInfo) {
			a108MakeControlDir(t, dir)
			a108DeadSocketWithMode(t, dir, 0o600)
			return SocketPath(dir), a113ProbeInfo(t, SocketPath(dir))
		}, false, 0},
		{"수락 중이면 생존", func(t *testing.T, dir string) (string, os.FileInfo) {
			a108MakeControlDir(t, dir)
			a108LiveSocket(t, dir)
			return SocketPath(dir), a113ProbeInfo(t, SocketPath(dir))
		}, true, 0},
		// UMask=0277 배포의 pre-chmod 잔재. 추정이 사라진 뒤에도 영구 거부가 되면 안 된다.
		{"쓰기 비트가 깎인 죽은 socket 도 사망", func(t *testing.T, dir string) (string, os.FileInfo) {
			unprivileged(t)
			a108MakeControlDir(t, dir)
			a108DeadSocketWithMode(t, dir, 0o500)
			return SocketPath(dir), a113ProbeInfo(t, SocketPath(dir))
		}, false, 0},
		// ⛔ 위 잔재와 디스크 권한이 같고(쓰기 비트 없음) 결과만 반대다.
		{"쓰기 비트가 깎여도 수락 중이면 생존", func(t *testing.T, dir string) (string, os.FileInfo) {
			unprivileged(t)
			a108MakeControlDir(t, dir)
			a108LiveSocket(t, dir)
			if err := os.Chmod(SocketPath(dir), 0o400); err != nil {
				t.Fatal(err)
			}
			return SocketPath(dir), a113ProbeInfo(t, SocketPath(dir))
		}, true, 0o600},
		// 검증 정보가 없으면 「검증한 그 파일인가」를 물을 수 없다 — 답을 얻지 못한 것은 생존.
		{"검증 정보가 없으면 생존", func(t *testing.T, dir string) (string, os.FileInfo) {
			a108MakeControlDir(t, dir)
			a108DeadSocketWithMode(t, dir, 0o600)
			return SocketPath(dir), nil
		}, true, 0},
		// 이름을 다시 볼 수 없는 이유가 부재가 아니면(ENOTDIR) 물어보지 못한 것이다.
		{"재확인이 부재 아닌 이유로 실패하면 생존", func(t *testing.T, dir string) (string, os.FileInfo) {
			blocker := filepath.Join(dir, "regular")
			if err := os.WriteFile(blocker, []byte("디렉터리가 아니다"), 0o600); err != nil {
				t.Fatal(err)
			}
			return filepath.Join(blocker, SocketFileName), a113ProbeInfo(t, blocker)
		}, true, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := shortRuntimeDir(t)
			path, before := test.build(t, dir)
			if got := staleProjectionSocketAccepts(path, before); got != test.want {
				t.Errorf("staleProjectionSocketAccepts = %v, want %v", got, test.want)
			}
			if test.wantMode != 0 {
				info, err := os.Lstat(path)
				if err != nil {
					t.Fatal(err)
				}
				if got := info.Mode().Perm(); got != test.wantMode {
					t.Errorf("probe 뒤 권한 = %#o, want %#o — 권한 복원은 발행 계약(0600)뿐이다", got, test.wantMode)
				}
			}
		})
	}
}

// TestTheStaleProbeRefusesASocketThatChangedUnderIt 은 a109 §2b.3 G7
// `TestTheProbeRefusesASocketThatChangedUnderIt` 의 원형판이다.
//
// chmod 는 검증한 **inode** 가 아니라 **이름**에 걸린다. 그 사이에 같은 uid 의 다른 프로세스가
// 이름을 갈아끼우면 우리는 다른 파일의 권한을 바꾸고 그 파일에 연결한다. 그래서 chmod 뒤 다시
// Lstat 해서 검증한 파일이 아니면 답을 얻지 못한 것이고, 그것은 생존으로 읽는다(보수 방향).
func TestTheStaleProbeRefusesASocketThatChangedUnderIt(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o600)
	before := a113ProbeInfo(t, SocketPath(dir))
	// 대조군: 그 자리의 그 파일이면 죽었다고 답한다.
	if staleProjectionSocketAccepts(SocketPath(dir), before) {
		t.Fatal("아무도 수락하지 않는 socket 을 살아 있다고 읽었다 — 대조군이 틀렸다")
	}

	// 같은 이름, 다른 파일. 후계자를 원본이 아직 있는 동안 만든 뒤 rename 으로 덮는다 —
	// 지우고 다시 만들면 파일시스템이 같은 inode 를 재배정할 수 있다.
	successor := filepath.Join(ControlDirectory(dir), stagingPrefix+"s0000001")
	listener, err := net.Listen("unix", successor)
	if err != nil {
		t.Fatal(err)
	}
	listener.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(successor, SocketPath(dir)); err != nil {
		t.Fatal(err)
	}
	if !staleProjectionSocketAccepts(SocketPath(dir), before) {
		t.Error("검증한 파일과 다른 파일을 죽었다고 읽었다 — 회수는 그것을 지운다. " +
			"물어보지 못한 것을 죽었다고 읽지 않는다")
	}
}

// TestTheStaleProbeLeavesAnUnverifiedNameAlone 은 검증 정보 없이 불렸을 때 **권한을 바꾸지
// 않는다**는 핀이다. chmod 는 검증을 통과한 inode 를 우리 발행 계약으로 되돌리는 것이지,
// 검증하지 않은 이름의 권한을 바꾸는 것이 아니다.
func TestTheStaleProbeLeavesAnUnverifiedNameAlone(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o500)
	if !staleProjectionSocketAccepts(SocketPath(dir), nil) {
		t.Fatal("검증 정보 없이 사망을 판정했다 — 답을 얻지 못한 것은 생존이다")
	}
	info, err := os.Lstat(SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o500 {
		t.Fatalf("검증하지 않은 이름의 권한을 바꿨다: %#o, want 0500", got)
	}
}

// TestTheReclaimHandsTheProbeTheInodeItVerified 는 freeze P2-2 의 AST 핀이다.
//
// probe 의 SameFile 재확인은 **검증한 그 inode** 와 비교할 때만 뜻이 있다. 회수가 verify 의
// 반환값 대신 새 Lstat 을 넘겨도 행동 테스트는 전부 통과한다(그 사이에 이름이 바뀌는 경합을
// 결정적으로 만들 수 없다). 그래서 배선을 구문으로 못 박는다: `reclaimStaleControlDirectory`
// 안에서 `staleProjectionSocketAccepts` 의 둘째 인자는 `verifyStaleSocketShape` 호출이
// 묶은 식별자여야 하고, 그 식별자는 한 번만 대입된다.
func TestTheReclaimHandsTheProbeTheInodeItVerified(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "transport_unix.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var reclaim *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "reclaimStaleControlDirectory" {
			reclaim = fn
		}
	}
	if reclaim == nil {
		t.Fatal("reclaimStaleControlDirectory 를 찾지 못했다")
	}
	calledName := func(call *ast.CallExpr) string {
		if ident, ok := call.Fun.(*ast.Ident); ok {
			return ident.Name
		}
		return ""
	}
	verified := ""
	assignments := map[string]int{}
	var probeArgs [][]ast.Expr
	ast.Inspect(reclaim.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					assignments[ident.Name]++
				}
			}
			if len(n.Rhs) == 1 && len(n.Lhs) == 2 {
				if call, ok := n.Rhs[0].(*ast.CallExpr); ok && calledName(call) == "verifyStaleSocketShape" {
					if ident, ok := n.Lhs[0].(*ast.Ident); ok {
						verified = ident.Name
					}
				}
			}
		case *ast.CallExpr:
			if calledName(n) == "staleProjectionSocketAccepts" {
				probeArgs = append(probeArgs, n.Args)
			}
		}
		return true
	})
	if verified == "" || verified == "_" {
		t.Fatal("회수가 verifyStaleSocketShape 의 FileInfo 를 받지 않는다")
	}
	if len(probeArgs) != 1 || len(probeArgs[0]) != 2 {
		t.Fatalf("회수 안의 staleProjectionSocketAccepts 호출 = %d개, want 정확히 1개(인자 2)", len(probeArgs))
	}
	ident, ok := probeArgs[0][1].(*ast.Ident)
	if !ok || ident.Name != verified {
		t.Fatalf("probe 의 둘째 인자가 verify 가 검증한 FileInfo(%s)가 아니다", verified)
	}
	if assignments[verified] != 1 {
		t.Fatalf("%s 가 회수 안에서 %d번 대입된다 — 검증한 inode 가 아닌 값으로 덮일 수 있다",
			verified, assignments[verified])
	}
}

// TestTheStaleProbeChecksTheNameOnBothSidesOfTheChmod 는 뮤테이션 N3 이 살아남은 자리의 구조 핀이다.
//
// chmod **뒤**의 재확인은 이름이 chmod 와 connect 사이에 바뀌는 경합에서만 일을 한다. 그
// 경합은 결정적으로 만들 수 없고, 바뀐 파일 테스트(`TestTheStaleProbeRefusesASocketThatChangedUnderIt`)는
// chmod **앞**의 재확인에서 먼저 걸린다 — 앞의 확인이 뒤의 확인을 가린다. 그래서 행동으로는
// 뒤의 확인을 지워도 초록이다(원장 N3). 순서를 구문으로 못 박는다:
// nil 검사 → 재확인 → chmod → 재확인 → 묻기.
func TestTheStaleProbeChecksTheNameOnBothSidesOfTheChmod(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "transport_probe_unix.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var probe *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "staleProjectionSocketAccepts" {
			probe = fn
		}
	}
	if probe == nil {
		t.Fatal("staleProjectionSocketAccepts 를 찾지 못했다")
	}
	var order []string
	ast.Inspect(probe.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			if fun.Name == "sameVerifiedSocket" || fun.Name == "projectionSocketAccepts" {
				order = append(order, fun.Name)
			}
		case *ast.SelectorExpr:
			if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "os" && fun.Sel.Name == "Chmod" {
				order = append(order, "os.Chmod")
			}
		}
		return true
	})
	want := []string{"sameVerifiedSocket", "os.Chmod", "sameVerifiedSocket", "projectionSocketAccepts"}
	if strings.Join(order, " → ") != strings.Join(want, " → ") {
		t.Fatalf("probe 의 순서 = %v, want %v", order, want)
	}
}
