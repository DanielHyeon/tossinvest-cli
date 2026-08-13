//go:build unix

package engine

// a108 tasks 2.5 — 부팅 순서의 **네** endpoint 전부에 crash-shape 핀을 깐다.
//
// 8/13 사고에서 엔진을 세운 것은 strategy projection 하나였고, 나머지 셋은 같은
// 재부팅 잔재를 실제로 통과했다. 하지만 그것은 **관측이지 핀이 아니다** — 관용은
// 설계이므로 지금 통과한다는 사실이 다음 판에서도 통과한다는 보장을 주지 않는다.
//
// 여기서 재는 것은 하나다: 지난 기동이 남길 수 있는 **부분 잔재**를 디스크에 만들어
// 놓고 기동이 이어지는가. 잔재의 모양은 각 endpoint의 자기 수명주기가 만든다.
//
//	빈 디렉터리    control dir을 만들고 그 다음 단계 전에 죽었다
//	descriptor만   socket이 unlink된 뒤(또는 TCP라 socket이 아예 없다) 죽었다
//	socket만       descriptor를 지우고 socket 제거 전에 죽었다
//	둘 다          SIGKILL·전원 단절
//
// production 코드는 건드리지 않는다. 핀이 결함을 드러내면 거기서 멈추고 보고한다.

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// a108LeftoverDir는 control 디렉터리를 지난 기동과 같은 0700으로 만든다.
func a108LeftoverDir(t *testing.T, controlDir string) {
	t.Helper()
	if err := os.Mkdir(controlDir, 0o700); err != nil {
		t.Fatal(err)
	}
}

// a108LeftoverDescriptor는 지난 기동의 endpoint.json을 남긴다. 내용은 지금
// 쓸모없는 값(죽은 주소·못 쓰는 토큰)이다 — 잔재란 원래 그런 것이다.
func a108LeftoverDescriptor(t *testing.T, path string, body []byte) {
	t.Helper()
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
}

// a108LeftoverDeadSocket은 아무도 수락하지 않는 socket 파일을 남긴다.
func a108LeftoverDeadSocket(t *testing.T, path string) {
	t.Helper()
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	listener.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
}

// a108SocketAccepts는 endpoint가 말로만 기동한 것이 아니라 실제로 수락하는지 본다.
func a108SocketAccepts(t *testing.T, path string) {
	t.Helper()
	conn, err := net.DialTimeout("unix", path, time.Second)
	if err != nil {
		t.Fatalf("기동했다는 endpoint가 수락하지 않는다(%s): %v", path, err)
	}
	_ = conn.Close()
}

// TestPositionPolicyCommandServerStartsOverALeftover — 로컬 TCP endpoint.
// socket 파일이 없으므로 이 endpoint의 잔재는 디렉터리와 descriptor뿐이다.
func TestPositionPolicyCommandServerStartsOverALeftover(t *testing.T) {
	for _, test := range []struct {
		name  string
		build func(t *testing.T, dir string)
	}{
		{"빈 디렉터리", func(t *testing.T, dir string) {
			a108LeftoverDir(t, positionpolicyrpc.ControlDirectory(dir))
		}},
		{"descriptor만 남았다", func(t *testing.T, dir string) {
			a108LeftoverDir(t, positionpolicyrpc.ControlDirectory(dir))
			a108LeftoverDescriptor(t, positionpolicyrpc.DescriptorPath(dir),
				[]byte(`{"address":"127.0.0.1:1","token":"dead-token-from-the-last-boot","pid":16}`))
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			test.build(t, dir)
			server, err := StartPositionPolicyCommandServer(dir, &fakePolicyCommands{})
			if err != nil {
				t.Fatalf("잔재에서 position policy command 기동이 거부됐다: %v", err)
			}
			t.Cleanup(func() { _ = server.Close() })
			// 잔재 descriptor가 이번 기동의 것으로 갈아끼워졌는지 본다.
			body, err := os.ReadFile(positionpolicyrpc.DescriptorPath(dir))
			if err != nil {
				t.Fatal(err)
			}
			var descriptor positionpolicyrpc.Descriptor
			if err := json.Unmarshal(body, &descriptor); err != nil {
				t.Fatal(err)
			}
			if descriptor.PID != os.Getpid() || descriptor.Address == "127.0.0.1:1" {
				t.Fatalf("잔재 descriptor가 그대로 남았다: %+v", descriptor)
			}
		})
	}
}

// TestPositionPolicyRuntimeServerStartsOverALeftover — Unix socket endpoint.
func TestPositionPolicyRuntimeServerStartsOverALeftover(t *testing.T) {
	for _, test := range []struct {
		name  string
		build func(t *testing.T, dir string)
	}{
		{"빈 디렉터리", func(t *testing.T, dir string) {
			a108LeftoverDir(t, positionpolicyrpc.RuntimeControlDirectory(dir))
		}},
		{"descriptor만 남았다", func(t *testing.T, dir string) {
			a108LeftoverDir(t, positionpolicyrpc.RuntimeControlDirectory(dir))
			a108LeftoverDescriptor(t, positionpolicyrpc.RuntimeDescriptorPath(dir),
				[]byte(`{"socket":"runtime.sock","token":"dead-token-from-the-last-boot","pid":16}`))
		}},
		{"socket만 남았다", func(t *testing.T, dir string) {
			a108LeftoverDir(t, positionpolicyrpc.RuntimeControlDirectory(dir))
			a108LeftoverDeadSocket(t, positionpolicyrpc.RuntimeSocketPath(dir))
		}},
		{"둘 다 남았다", func(t *testing.T, dir string) {
			a108LeftoverDir(t, positionpolicyrpc.RuntimeControlDirectory(dir))
			a108LeftoverDeadSocket(t, positionpolicyrpc.RuntimeSocketPath(dir))
			a108LeftoverDescriptor(t, positionpolicyrpc.RuntimeDescriptorPath(dir),
				[]byte(`{"socket":"runtime.sock","token":"dead-token-from-the-last-boot","pid":16}`))
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			test.build(t, dir)
			commands := &fakePolicyCommands{runtime: positionpolicy.ManagementRuntime{EffectiveKnown: true}}
			server, err := StartPositionPolicyRuntimeServer(dir, commands)
			if err != nil {
				t.Fatalf("잔재에서 position policy runtime 기동이 거부됐다: %v", err)
			}
			t.Cleanup(func() { _ = server.Close() })
			reader, err := positionpolicyrpc.DialRuntime(context.Background(),
				positionpolicyrpc.RuntimeDescriptorPath(dir))
			if err != nil {
				t.Fatalf("회수 후 descriptor를 못 읽는다: %v", err)
			}
			if _, err := reader.Runtime(context.Background()); err != nil {
				t.Fatalf("회수 후 endpoint가 응답하지 않는다: %v", err)
			}
		})
	}
}

// TestAlertControlServerStartsOverALeftover — 운영자 alert socket.
func TestAlertControlServerStartsOverALeftover(t *testing.T) {
	for _, test := range []struct {
		name  string
		build func(t *testing.T, dir string)
	}{
		{"빈 디렉터리", func(t *testing.T, dir string) {
			a108LeftoverDir(t, AlertControlDirectory(dir))
		}},
		{"descriptor만 남았다", func(t *testing.T, dir string) {
			a108LeftoverDir(t, AlertControlDirectory(dir))
			a108LeftoverDescriptor(t, AlertControlDescriptorPath(dir),
				[]byte(`{"socket":"alerts.sock","token":"dead-token-from-the-last-boot","pid":16}`))
		}},
		{"socket만 남았다", func(t *testing.T, dir string) {
			a108LeftoverDir(t, AlertControlDirectory(dir))
			a108LeftoverDeadSocket(t, AlertControlSocketPath(dir))
		}},
		{"둘 다 남았다", func(t *testing.T, dir string) {
			a108LeftoverDir(t, AlertControlDirectory(dir))
			a108LeftoverDeadSocket(t, AlertControlSocketPath(dir))
			a108LeftoverDescriptor(t, AlertControlDescriptorPath(dir),
				[]byte(`{"socket":"alerts.sock","token":"dead-token-from-the-last-boot","pid":16}`))
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			test.build(t, dir)
			// 이 핀이 재는 것은 기동이지 경로 동작이 아니므로 ops는 최소로 둔다.
			server, err := StartAlertControlServer(dir, &AlertOperations{})
			if err != nil {
				t.Fatalf("잔재에서 alert control 기동이 거부됐다: %v", err)
			}
			t.Cleanup(func() { _ = server.Close() })
			a108SocketAccepts(t, AlertControlSocketPath(dir))
			body, err := os.ReadFile(AlertControlDescriptorPath(dir))
			if err != nil {
				t.Fatal(err)
			}
			var descriptor AlertControlDescriptor
			if err := json.Unmarshal(body, &descriptor); err != nil {
				t.Fatal(err)
			}
			if descriptor.PID != os.Getpid() || descriptor.Token == "dead-token-from-the-last-boot" {
				t.Fatalf("잔재 descriptor가 그대로 남았다: %+v", descriptor)
			}
		})
	}
}
