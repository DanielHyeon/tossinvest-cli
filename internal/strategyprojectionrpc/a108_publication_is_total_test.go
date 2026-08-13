//go:build unix

package strategyprojectionrpc

// a108 Fix 라운드 — **발행도 회수만큼 전체적이어야 한다**(design D1-2).
//
// 겹1이 회수의 커버리지를 넓혔지만, A1 적대 리뷰가 남은 영구 거부 2종을 실측했다.
// 둘 다 **생산자**가 만든다 — 회수가 못 치우는 것이 아니라 발행이 원자적이지 않아서
// 생기는 모양이다.
//
//	pre-chmod socket   net.Listen이 만드는 권한은 umask가 정한다(컨테이너 실측 077 →
//	                   0700). listen과 chmod 사이에서 죽으면 최종 이름 자리에 0600이
//	                   아닌 socket이 남고, 회수의 정확-0600 검사가 영구 거부한다.
//	반쯤 쓰인 descriptor  O_EXCL 생성 → write → sync는 원자적이지 않다. 그 사이에서
//	                   죽으면 0바이트·잘린 JSON이 남고, 내용 파싱 실패는 영구 거부다.
//
// **umask를 테스트에서 바꾸지 않는다.** syscall.Umask는 프로세스 전역이라 병렬로 도는
// 다른 테스트가 만드는 파일 권한까지 오염시킨다. 원하는 권한의 파일을 직접 만드는
// 것으로 같은 디스크 상태를 재현한다(A1 절차).

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// a108DeadSocketWithMode는 죽은 socket을 **원하는 권한 그대로** 남긴다.
// listen이 만든 파일을 chmod로 바꾸므로 프로세스 umask를 건드리지 않는다.
func a108DeadSocketWithMode(t *testing.T, dir string, mode os.FileMode) {
	t.Helper()
	listener, err := net.Listen("unix", SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	listener.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := os.Chmod(SocketPath(dir), mode); err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestStartRecoversFromPreChmodSocketLeftover는 A1 F1이다.
// 구버전 바이너리가 listen과 chmod 사이에서 죽으면 이 모양이 남는다 — 0700 socket.
// 아무도 수락하지 않고, 우리 uid 소유이고, 0700 디렉터리 안이고, hard link도 없다.
// group/other 비트가 하나도 없으므로 회수해야 한다.
func TestStartRecoversFromPreChmodSocketLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o700)
	a108WriteLeftoverDescriptor(t, dir, 1<<30)
	a108StartServes(t, dir)
}

// TestStartRecoversFromEmptyDescriptorLeftover는 A1 F2의 0바이트 모양이다.
// **이번 사고와 같은 S1**(socket 없음 = listener 없음 = 주인 사망)인데 descriptor가
// 반쯤 쓰였다는 이유만으로 영구 거부였다. D2 이후 descriptor 내용은 어떤 판정에도
// 쓰이지 않으므로 그 게이트는 거부만 생산한다.
func TestStartRecoversFromEmptyDescriptorLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	if err := os.WriteFile(DescriptorPath(dir), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	a108StartServes(t, dir)
}

// TestStartRecoversFromTruncatedDescriptorLeftover는 같은 결함의 S3 모양이다.
// socket도 남아 있고(전원 단절) descriptor는 JSON 중간에서 끊겼다.
// 주인의 사망은 socket probe가 증명하고, 그 뒤에 내용 실패는 회수다.
func TestStartRecoversFromTruncatedDescriptorLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o600)
	if err := os.WriteFile(DescriptorPath(dir), []byte(`{"schemaVersion":"tossos.st`), 0o600); err != nil {
		t.Fatal(err)
	}
	a108StartServes(t, dir)
}

// TestStartRecoversFromUnpublishedStagingLeftover는 D1-2가 새로 만드는 잔재다.
// stage+rename 발행은 최종 이름을 갖기 전의 임시 이름을 만든다. 그 이름으로 죽으면
// 남는 것은 **낯선 엔트리가 아니라 자기 잔재**다 — 최종 이름을 가진 적이 없으므로
// 아무도 읽지 않았고, 검증할 것도 없다.
//
// 임시 이름을 코드 상수가 아니라 문자열 그대로 쓴다: 이 테스트가 고정하는 것은
// "회수가 접두사를 안다"가 아니라 **디스크에 남는 모양**이다.
func TestStartRecoversFromUnpublishedStagingLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	for _, name := range []string{".s-e2716057", ".s-s9f3ac18"} {
		if err := os.WriteFile(filepath.Join(ControlDirectory(dir), name), []byte("half"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	a108StartServes(t, dir)
	entries, err := os.ReadDir(ControlDirectory(dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".s-") {
			t.Fatalf("발행 전 임시 이름이 살아남았다: %s", entry.Name())
		}
	}
}

// TestDialRefusesSocketWithNoListener는 D4-2의 T1 몫이다(A2 F3 실측).
// 전원이 끊기면 unlink할 기회가 없으므로 socket 파일이 그대로 남는다(S3). 그 위에서
// Dial이 lazy client를 돌려주면 실패는 **첫 Read**에서야 나타나고, 그때는 소비자가
// 이미 "붙었다"고 판단한 뒤라 강등할 기회를 놓친다 — 집계 스냅샷 전체가 함께 죽었다.
func TestDialRefusesSocketWithNoListener(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o600)
	a108WriteLeftoverDescriptor(t, dir, 1<<30)
	client, err := Dial(context.Background(), DescriptorPath(dir))
	if err == nil {
		t.Fatalf("아무도 수락하지 않는 socket에 Dial이 성공했다 (client=%v)", client != nil)
	}
}

// TestStartPublishesBothArtifactsByRename은 발행 의례 자체의 핀이다.
// 기동이 끝난 디렉터리에는 최종 이름 둘만 있어야 한다 — 임시 이름이 남아 있으면
// 발행이 중간에서 멈춘 것이고, listener가 최종 경로를 기억하고 있으면 socket이
// rename이 아니라 제자리에서 만들어진 것이다.
func TestStartPublishesBothArtifactsByRename(t *testing.T) {
	dir := shortRuntimeDir(t)
	server, err := Start(dir, a108Reader())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	entries, err := os.ReadDir(ControlDirectory(dir))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 2 {
		t.Fatalf("control 디렉터리 엔트리 = %v, want 최종 이름 둘만", names)
	}
	if remembered := server.listener.Addr().String(); remembered == SocketPath(dir) {
		t.Fatalf("listener가 최종 socket 경로를 기억한다 (%s) — 제자리에서 bind했다", remembered)
	}
}

// TestPublishedListenerNeverUnlinksTheNameItRemembers는 D2-2의 절반이다.
//
// Go의 unix listener는 닫힐 때 **자기가 기억하는 경로**를 unlink한다. 발행 뒤 그
// 이름은 이미 rename으로 비었지만, 늦은 정리가 도착하는 시점에 그 자리에 다른
// 파일이 앉아 있을 수 있다. 그 표적을 테스트가 직접 놓고 닫아 본다.
func TestPublishedListenerNeverUnlinksTheNameItRemembers(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	listener, err := listenPrivateSocket(ControlDirectory(dir), SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	remembered := listener.Addr().String()
	if err := os.WriteFile(remembered, []byte("남의 파일"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(remembered); err != nil {
		t.Fatalf("listener가 자기가 기억하는 이름의 파일을 지웠다: %v", err)
	}
}

// TestLateListenerCloseCannotUnlinkSuccessorSocket은 A1 F5의 결정적 회귀 핀이다.
//
// A1은 300라운드 중 3회로 재현했다: `Shutdown`이 `Serve`의 listener 등록을 앞지르면
// 정리가 Serve goroutine의 defer로 밀리고, **경로 기준** unlink가 그 사이에 들어선
// 후계자의 socket을 지운다. 여기서는 경합을 기다리지 않고 그 순서를 직접 만든다 —
// 첫 주인의 listener를 손에 쥔 채 후계자를 세우고, 그 다음에 닫는다.
func TestLateListenerCloseCannotUnlinkSuccessorSocket(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	first, err := listenPrivateSocket(ControlDirectory(dir), SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	// 주인이 파일만 치우고 사라진다(Close가 하는 일). listener는 아직 살아 있다.
	if err := os.Remove(SocketPath(dir)); err != nil {
		t.Fatal(err)
	}
	second, err := listenPrivateSocket(ControlDirectory(dir), SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })
	before, err := os.Lstat(SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	// 늦은 정리가 지금 도착한다.
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	after, err := os.Lstat(SocketPath(dir))
	if err != nil {
		t.Fatalf("늦은 listener 정리가 후계자의 socket을 지웠다: %v", err)
	}
	if !os.SameFile(before, after) {
		t.Fatal("후계자의 socket이 다른 파일로 바뀌었다")
	}
	conn, err := net.DialTimeout("unix", SocketPath(dir), projectionProbeTimeout)
	if err != nil {
		t.Fatalf("후계자가 더 이상 수락하지 않는다: %v", err)
	}
	_ = conn.Close()
}

// TestCloseClosesItsOwnListenerWithoutServe는 D2-2의 나머지 절반이다.
//
// `Serve`를 한 번도 부르지 않은 서버가 곧 "Shutdown이 등록을 앞지른" 그 창이다.
// 그 상태에서 `Close`가 돌아왔을 때 listener는 이미 닫혀 있어야 한다 — 닫는 시점을
// 예정된 goroutine에 맡기면 그 goroutine이 프로세스와 함께 죽는 것이 유일한 방어가
// 된다(A1 F5의 지적).
func TestCloseClosesItsOwnListenerWithoutServe(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	listener, err := listenPrivateSocket(ControlDirectory(dir), SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	server := &Server{server: &http.Server{}, listener: listener,
		descriptor: DescriptorPath(dir), socket: SocketPath(dir), controlDir: ControlDirectory(dir)}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("Close가 listener를 닫지 않았다 (두 번째 Close = %v)", err)
	}
	if _, err := os.Lstat(ControlDirectory(dir)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Close가 control 디렉터리를 치우지 않았다: %v", err)
	}
}

// 아래 셋은 A1 F6의 **절 단위 미핀** 소거다. 각 절은 다른 절이 이미 잡아 주는
// 디스크 상태에 가려 혼자서는 죽어 본 적이 없었다.

// TestControlDirectoryModeClausesEachRefuseOnTheirOwn — symlink 절이 여기서만 잰다.
// 실제 `Lstat`이 돌려주는 symlink의 mode에는 ModeDir 비트가 없으므로 `IsDir()` 절이
// 먼저 걸린다. 즉 디스크 상태로는 symlink 절만 실패시킬 수 없고, 그 절만 지운
// 뮤테이션은 파일 기반 테스트를 전부 통과한다. 모드 값을 직접 넣어서 그 자리를 막는다.
func TestControlDirectoryModeClausesEachRefuseOnTheirOwn(t *testing.T) {
	for _, test := range []struct {
		name string
		mode os.FileMode
		want bool
	}{
		{"0700 디렉터리", os.ModeDir | 0o700, true},
		{"디렉터리가 아니다", 0o700, false},
		{"symlink 비트가 서 있다", os.ModeDir | os.ModeSymlink | 0o700, false},
		{"group 비트가 있다", os.ModeDir | 0o750, false},
		{"other 비트가 있다", os.ModeDir | 0o701, false},
		// 축소 방향도 거부다(gstack). 위 두 행만 있으면 `Perm() == 0o700` 을
		// `Perm()&0o077 == 0` 으로 넓힌 뮤테이션이 살아남는다 — 그 판정은 우리가 쓰지도
		// 들어가지도 못하는 디렉터리를 "안전하다"고 읽고, 그 뒤의 제거는 전부 실패한다.
		{"owner 쓰기 비트가 없다", os.ModeDir | 0o500, false},
		{"owner 실행 비트가 없다", os.ModeDir | 0o600, false},
	} {
		if got := controlDirectoryModeIsSafe(test.mode); got != test.want {
			t.Errorf("%s: controlDirectoryModeIsSafe(%v) = %v, want %v", test.name, test.mode, got, test.want)
		}
	}
}

// TestDescriptorIdentityClausesRejectASwappedFile — SameFile 두 절의 핀이다.
// 실제로 이 절이 걸리려면 검사와 열기 **사이**에 파일이 바뀌어야 하는데, 그
// 인터리빙을 테스트가 결정적으로 만들 방법이 없다(seam을 새로 뚫지 않는 한).
// 판정 자체를 세 개의 실파일로 직접 먹인다.
func TestDescriptorIdentityClausesRejectASwappedFile(t *testing.T) {
	dir := shortRuntimeDir(t)
	first := filepath.Join(dir, "one")
	second := filepath.Join(dir, "two")
	for _, path := range []string{first, second} {
		if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	one, err := os.Lstat(first)
	if err != nil {
		t.Fatal(err)
	}
	two, err := os.Lstat(second)
	if err != nil {
		t.Fatal(err)
	}
	if !sameDescriptorFile(one, one, one) {
		t.Fatal("같은 파일 셋을 다르다고 했다")
	}
	if sameDescriptorFile(one, two, one) {
		t.Fatal("검사한 파일과 다른 파일을 열었는데 같다고 했다")
	}
	// 이 줄이 첫 절만 겨눈다: 연 파일과 연 뒤의 파일은 서로 같고, 검사한 파일만
	// 다르다. 두 절을 `&&`로 묶어 놓으면 다른 줄들은 둘째 절만으로도 통과한다.
	if sameDescriptorFile(one, two, two) {
		t.Fatal("검사하기 전에 바꿔치기당했는데 같다고 했다")
	}
	if sameDescriptorFile(one, one, two) {
		t.Fatal("연 뒤에 자리가 바뀌었는데 같다고 했다")
	}
}

// TestDescriptorPublicationIsAtomic은 descriptor stage+rename의 핀이다.
//
// 발행이 원자적이면 최종 이름은 **완성된 파일에만** 붙는다: 다시 발행하는 동안에도
// 읽는 쪽은 옛 내용이나 새 내용 중 하나를 볼 뿐, 반쯤 쓰인 것을 보지 않는다.
// 제자리에서 만들고 쓰는 옛 발행은 이 성질을 두 방향으로 깬다 — 생성과 write 사이에
// 0바이트가 보이고, `O_EXCL`이면 두 번째 발행 자체가 되지 않는다.
func TestDescriptorPublicationIsAtomic(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	tokens := [2]string{strings.Repeat("a", 43), strings.Repeat("b", 43)}
	publish := func(token string) error {
		return writeDescriptor(DescriptorPath(dir), Descriptor{SchemaVersion: descriptorSchema,
			Socket: SocketFileName, Token: token, PID: os.Getpid()})
	}
	if err := publish(tokens[0]); err != nil {
		t.Fatal(err)
	}
	published := make(chan error, 1)
	go func() {
		defer close(published)
		for round := range 300 {
			if err := publish(tokens[round%2]); err != nil {
				published <- err
				return
			}
		}
	}()
	reads, changed := 0, 0
	for {
		select {
		case err, open := <-published:
			if err != nil {
				t.Fatalf("다시 발행하지 못했다: %v", err)
			}
			if !open {
				if reads == 0 {
					t.Fatal("발행 중에 읽기가 한 번도 성사되지 않았다 — 이 테스트는 아무것도 재지 않았다")
				}
				t.Logf("완성된 읽기 %d회 · rename 경합 %d회", reads, changed)
				return
			}
		default:
		}
		descriptor, err := readDescriptor(DescriptorPath(dir))
		if errors.Is(err, errDescriptorChanged) {
			// rename이 우리 발밑에서 파일을 **다른 완성 파일로** 바꿨다. 이것이
			// 원자성의 증거이지 반증이 아니다 — 반쯤 쓰인 것을 본 적은 없다.
			changed++
			continue
		}
		if err != nil {
			t.Fatalf("발행 중에 완성되지 않은 descriptor가 보였다 (%d번째 읽기): %v", reads, err)
		}
		if descriptor.Token != tokens[0] && descriptor.Token != tokens[1] {
			t.Fatalf("발행한 적 없는 토큰을 읽었다: %q", descriptor.Token)
		}
		reads++
	}
}

// --- gstack Fix 라운드 --------------------------------------------------------
//
// 아래는 gstack 7패스 리뷰가 Fix-First 로 판정한 넷이다. 전부 **발행이 만든** 모양이고,
// 앞의 Fix 라운드와 같은 병의 나머지다: 회수는 전체적인데 발행이 그렇지 않다.

// sunPathMax는 unix socket 경로가 가질 수 있는 최대 길이다.
//
// 커널의 `sun_path`는 108바이트이고 마지막 한 자리는 NUL 이 쓴다(실측: 107은 bind
// 되고 108은 `invalid argument`). bind 하는 이름은 **임시 이름**이므로, 임시 이름이
// 최종 이름보다 길면 최종 경로가 멀쩡한 배포에서도 발행만 실패한다.
const sunPathMax = 107

// TestStagingNamesAreNeverLongerThanTheNamesTheyBecome는 그 관계를 이름 길이로 고정한다.
//
// 경로 길이가 아니라 **basename 길이**를 재는 이유: 두 이름은 같은 디렉터리 안에
// 있으므로, basename 이 길지 않으면 경로도 길지 않다. 디렉터리가 얼마나 깊든 성립한다.
func TestStagingNamesAreNeverLongerThanTheNamesTheyBecome(t *testing.T) {
	for _, test := range []struct {
		kind  string
		final string
	}{
		{stagingSocketKind, SocketFileName},
		{stagingDescriptorKind, DescriptorFileName},
	} {
		staged, err := stagingPath("/tmp", test.kind)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(filepath.Base(staged)), len(test.final); got > want {
			t.Errorf("임시 이름 %q(%d자)이 최종 이름 %q(%d자)보다 길다 — "+
				"최종 경로가 sun_path 상한 직전인 배포에서 bind 만 ENAMETOOLONG 으로 죽는다",
				filepath.Base(staged), got, test.final, want)
		}
	}
}

// TestStartPublishesWhereTheFinalSocketPathIsAtTheLimit은 그 배포를 실제로 만든다.
//
// 최종 socket 경로가 정확히 상한인 디렉터리다. 여기서 기동이 실패하면 그것은 「경로가
// 길다」가 아니라 **「우리가 최종 이름보다 긴 이름으로 bind 한다」**는 뜻이다 — 운영자가
// 볼 수 있는 것은 매 부팅 projection 이 사라지는 것뿐이고, 최종 경로는 상한 안이므로
// 원인이 보이지 않는다.
func TestStartPublishesWhereTheFinalSocketPathIsAtTheLimit(t *testing.T) {
	root := shortRuntimeDir(t)
	// SocketPath(dir) = dir + "/" + ControlDirectoryName + "/" + SocketFileName 이므로
	// 그 길이가 상한이 되는 dir 길이를 역산한다.
	want := sunPathMax - len(SocketPath("")) - 1 // 마지막 -1 은 dir 과 그 아래를 잇는 "/"
	pad := want - len(root) - 1
	if pad < 1 || pad > 200 {
		t.Skipf("이 호스트의 tmp 경로(%s)로는 상한 직전 디렉터리를 만들 수 없다", root)
	}
	dir := filepath.Join(root, strings.Repeat("d", pad))
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if got := len(SocketPath(dir)); got != sunPathMax {
		t.Fatalf("fixture 의 최종 socket 경로 = %d자, want %d — 이 테스트는 상한을 재지 않았다",
			got, sunPathMax)
	}
	a108StartServes(t, dir)
}

// TestProjectionLivenessClausesEachDecideOnTheirOwn은 생존 판정 네 절의 핀이다.
//
// `Start` 를 통해 재면 verifyStaleSocketShape 가 먼저 걸리는 모양이 있어서 절 하나를
// 따로 죽여 볼 수 없다. 판정 함수를 직접 부른다 — 검사를 약하게 만든 것이 아니라
// **죽여 볼 수 있게** 만든 것이고, 호출부의 조건은 이 함수를 부르는 그대로다.
//
// owner 쓰기 절이 이 라운드에 새로 생겼다(gstack). `net.Listen` 이 만드는 권한은
// umask 가 정하므로 `UMask=0277` 로 도는 배포에서는 pre-chmod 잔재가 **0500** 으로
// 남는다. 그 socket 에 connect 하면 커널은 EACCES 를 주는데(실측), 옛 판정은 그것을
// 「답이 안 왔으니 살아 있다」로 읽어 영구 거부를 만들었다. 발행은 항상 chmod 0600
// 뒤에 최종 이름을 주므로, owner 쓰기가 없는 socket 은 **산 endpoint 일 수 없다.**
func TestProjectionLivenessClausesEachDecideOnTheirOwn(t *testing.T) {
	for _, test := range []struct {
		name  string
		build func(t *testing.T, dir string) string
		want  bool
	}{
		{"경로가 없다", func(_ *testing.T, dir string) string {
			return filepath.Join(dir, "absent.sock")
		}, false},
		{"아무도 수락하지 않는다", func(t *testing.T, dir string) string {
			a108MakeControlDir(t, dir)
			a108DeadSocketWithMode(t, dir, 0o600)
			return SocketPath(dir)
		}, false},
		{"수락한다", func(t *testing.T, dir string) string {
			a108MakeControlDir(t, dir)
			a108LiveSocket(t, dir)
			return SocketPath(dir)
		}, true},
		{"owner 쓰기 비트가 없다", func(t *testing.T, dir string) string {
			if os.Geteuid() == 0 {
				t.Skip("root 는 DAC 를 우회하므로 EACCES 를 만들 수 없다")
			}
			a108MakeControlDir(t, dir)
			a108DeadSocketWithMode(t, dir, 0o500)
			return SocketPath(dir)
		}, false},
		// 보수 기본값의 대조군이다. 이것이 없으면 "전부 사망으로 읽는다"는 구현이
		// 위 셋을 통과한다 — 그리고 그 구현은 남의 socket 을 지운다.
		{"죽었다는 증거가 아닌 오류", func(t *testing.T, dir string) string {
			blocker := filepath.Join(dir, "regular")
			if err := os.WriteFile(blocker, []byte("이 자리는 디렉터리가 아니다"), 0o600); err != nil {
				t.Fatal(err)
			}
			return filepath.Join(blocker, SocketFileName) // ENOTDIR
		}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := shortRuntimeDir(t)
			if got := projectionSocketAccepts(test.build(t, dir)); got != test.want {
				t.Errorf("projectionSocketAccepts = %v, want %v", got, test.want)
			}
		})
	}
}

// TestStartRecoversFromUnwritableSocketLeftover는 위 절이 디스크에서 하는 일이다.
// `UMask=0277` 배포의 pre-chmod 잔재 — 매 부팅 영구 거부였다.
func TestStartRecoversFromUnwritableSocketLeftover(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root 는 DAC 를 우회하므로 이 잔재가 EACCES 를 만들지 않는다")
	}
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o500)
	a108WriteLeftoverDescriptor(t, dir, 1<<30)
	a108StartServes(t, dir)
}

// TestStartRefusesStagingLeftoverOfAnUnexpectedShape는 S4 규칙을 임시 이름에도 적용한다.
//
// 임시 이름을 「우리 잔재」로 회수하기로 한 것은 그 이름을 **우리만** 만들기 때문이다.
// 그런데 이름만 보고 지우면, 그 자리에 우리가 만들 수 없는 모양(디렉터리 등)이 있어도
// 지우려 든다. 빈 디렉터리면 **남의 디렉터리를 지우고**, 비어 있지 않으면 ENOTEMPTY 로
// 회수 전체가 매 부팅 wedge 된다. 우리가 임시 이름으로 만드는 것은 정규 파일과 socket
// 둘뿐이므로, 그 밖의 모양은 아무것도 건드리지 않고 거부한다(S4 와 같은 규칙).
func TestStartRefusesStagingLeftoverOfAnUnexpectedShape(t *testing.T) {
	for _, test := range []struct {
		name string
		fill bool
	}{
		{"빈 디렉터리", false},
		{"비어 있지 않은 디렉터리", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := shortRuntimeDir(t)
			a108MakeControlDir(t, dir)
			wedge := filepath.Join(ControlDirectory(dir), stagingPrefix+"wedge")
			if err := os.Mkdir(wedge, 0o700); err != nil {
				t.Fatal(err)
			}
			keep := []string{wedge}
			if test.fill {
				inside := filepath.Join(wedge, "inside")
				if err := os.WriteFile(inside, []byte("남의 것"), 0o600); err != nil {
					t.Fatal(err)
				}
				keep = append(keep, inside)
			}
			// 회수 가능한 잔재를 같이 둔다: 거부가 「아무것도 건드리지 않는다」를 재려면
			// 건드릴 수 있는 것이 실재해야 한다.
			a108DeadSocketWithMode(t, dir, 0o600)
			a108WriteLeftoverDescriptor(t, dir, 1<<30)
			keep = append(keep, SocketPath(dir), DescriptorPath(dir))
			a108StartRefusesAndKeeps(t, dir, keep...)
		})
	}
}

// TestDiscardingAFailedPublicationLeavesNothingBehind는 실패 정리의 전체성이다.
//
// `writeDescriptor` 의 rename 이 성공한 **뒤에도** 실패하는 줄이 있다(발행 확인의
// SameFile, 디렉터리 sync). 그 경로에서 descriptor 는 이미 최종 이름을 갖고 있으므로,
// 정리가 socket 과 디렉터리만 치우면 당회 부팅 내내 S1 잔재가 남고 디렉터리 제거도
// ENOTEMPTY 로 실패한다.
//
// 그 인터리빙 자체는 seam 을 새로 뚫지 않고 결정적으로 만들 수 없다(원장 §C 에 사유).
// 여기서 고정하는 것은 **정리의 커버리지**다 — 세 산출물 중 하나라도 빠지면 디렉터리가
// 남는다.
func TestDiscardingAFailedPublicationLeavesNothingBehind(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o600)
	a108WriteLeftoverDescriptor(t, dir, os.Getpid())
	discardPublication(ControlDirectory(dir), DescriptorPath(dir), SocketPath(dir))
	if _, err := os.Lstat(ControlDirectory(dir)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("발행 취소 뒤에도 control 디렉터리가 남았다 (%v) — 산출물 하나가 정리 목록에 없다", err)
	}
}

// TestOwnershipClauseRefusesAnotherUser — uid 절의 핀이다.
// 이 테스트는 비root로 돌고, 비root는 파일 소유자를 바꿀 수 없다. 그래서 "남이 만든
// 디렉터리"라는 디스크 상태 자체를 만들 수 없고, 판정을 직접 먹이는 것이 이 절을
// 죽여 볼 수 있는 유일한 방법이다.
func TestOwnershipClauseRefusesAnotherUser(t *testing.T) {
	mine := uint32(os.Geteuid())
	if !ownedByEffectiveUser(mine) {
		t.Fatal("우리 uid를 남의 것이라고 했다")
	}
	if ownedByEffectiveUser(mine + 1) {
		t.Fatal("남의 uid를 우리 것이라고 했다")
	}
}
