//go:build unix

package engine

// a109_a_degraded_command_boot_advertises_nothing_test.go — a109 §2b.3 G4.
//
// command endpoint 의 descriptor 는 **loopback 주소와 bearer 토큰**이다. 그것을 읽는
// 소비자(콘솔의 격리 해제·Preview/Apply)는 거기 적힌 포트로 연결해 토큰을 보낸다.
//
// 죽은 run 이 남긴 descriptor 가 그대로 있으면 그 포트는 **재활용된다** — ephemeral
// 포트는 커널이 다시 나눠 주고, 컨테이너를 다시 만들면 PID 도 같은 자리에 재배정된다
// (a102 D4b-2). 그러면 소비자는 무관한 프로세스에 자기 토큰을 보낸다.
//
// a109 D3 이후 이 경로는 실재한다: Start 가 실패해도 엔진은 **강등**으로 계속 돌고,
// 그 부팅은 descriptor 를 새로 쓰지 않는다. 그래서 지우는 자리는 Start 의 **초입**이다 —
// 그 뒤의 어떤 실패도 죽은 run 의 광고를 남기지 않도록.
//
// # 지우기 전에 무엇을 확인하는가
//
// 기존 private 검사(`ValidatePrivateFile`: 정규 파일 · 정확-0600 · 우리 uid · 링크 1)를
// 그대로 쓴다. 통과하지 못하는 것은 우리가 만들 수 있는 모양이 아니므로 **건드리지
// 않는다** — 이 endpoint 의 규칙은 D2a 그대로다(낯선 것은 무시하고, 위생이 표면을
// 없애지 않는다).
//
// # 전제: journal flock
//
// 살아 있는 주인의 광고를 지우는 것 아닌가 — 아니다. `runEngineRun` 은 endpoint 기동
// **전에** journal flock 을 쥐므로(부팅 1단계) 같은 journal 을 쓰는 두 번째 엔진은 존재할
// 수 없다. 이 자리에서 보이는 descriptor 는 정의상 **죽은 run 의 것**이다.

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// a109DeadRunDescriptor 는 죽은 run 이 남긴 광고 하나다(재활용 가능한 포트 + 토큰).
func a109DeadRunDescriptor(t *testing.T, path string) {
	t.Helper()
	body := fmt.Sprintf(`{"address":"127.0.0.1:34567","token":%q,"pid":2}`,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestTheCommandBootDropsADeadRunsDescriptor 는 G4 의 동작이다.
func TestTheCommandBootDropsADeadRunsDescriptor(t *testing.T) {
	dir := privateEngineTestDir(t)
	controlDir := positionpolicyrpc.ControlDirectory(dir)
	a109ControlDir(t, controlDir)
	descriptor := positionpolicyrpc.DescriptorPath(dir)
	a109DeadRunDescriptor(t, descriptor)

	dropStalePositionPolicyDescriptor(descriptor)

	if _, err := os.Lstat(descriptor); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("죽은 run 의 descriptor 가 남았다(err=%v) — 강등 부팅은 그것을 새로 "+
			"쓰지 않으므로, 소비자는 재활용된 포트에 토큰을 보낸다", err)
	}
}

// TestTheCommandBootLeavesWhatItCannotVouchFor 는 그 지우기의 **경계**다.
//
// 이 endpoint 의 규칙은 D2a 다: 낯선 것은 건드리지 않는다. 위생이 격리 해제 표면을
// 없애서는 안 되므로, 확신할 수 없는 모양은 그대로 두고 오류도 내지 않는다.
func TestTheCommandBootLeavesWhatItCannotVouchFor(t *testing.T) {
	cases := []struct {
		name string
		make func(t *testing.T, path string)
	}{
		{
			name: "남에게 열린 파일",
			make: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "디렉터리",
			make: func(t *testing.T, path string) {
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "symlink",
			make: func(t *testing.T, path string) {
				target := filepath.Join(filepath.Dir(path), ".position-policy-control-9999")
				if err := os.WriteFile(target, []byte("{}"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			controlDir := positionpolicyrpc.ControlDirectory(dir)
			a109ControlDir(t, controlDir)
			descriptor := positionpolicyrpc.DescriptorPath(dir)
			testCase.make(t, descriptor)

			dropStalePositionPolicyDescriptor(descriptor)

			if _, err := os.Lstat(descriptor); err != nil {
				t.Fatalf("우리가 만들 수 없는 모양을 지웠다: %v", err)
			}
		})
	}
}

// TestTheCommandStartDropsTheDescriptorBeforeAnythingCanFail 은 **자리**를 잰다.
//
// 동작만 재면 「지운다」는 참이면서 「강등 부팅이 광고를 남긴다」도 참일 수 있다 —
// 지우는 줄이 실패 가능한 줄들 **뒤에** 있으면 그렇다. 그 순서는 동작 테스트로 잴 수
// 없다: 이 Start 의 남은 실패 지점(`net.Listen`·descriptor 발행)은 테스트가 결정적으로
// 만들 수 있는 seam 이 아니고, `crypto/rand.Read` 는 Go 1.24 이후 오류 대신
// **프로세스를 죽인다**(GOROOT crypto/rand: "crashes the program irrecoverably").
// 그래서 순서를 소스에서 읽는다 — a109 가 이미 두 곳에서 쓰는 정적 핀 방식이다.
func TestTheCommandStartDropsTheDescriptorBeforeAnythingCanFail(t *testing.T) {
	const source = "position_policy_transport.go"
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, source, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	ast.Inspect(parsed, func(node ast.Node) bool {
		function, ok := node.(*ast.FuncDecl)
		if ok && function.Name.Name == "StartPositionPolicyCommandServer" {
			body = function.Body
		}
		return body == nil
	})
	if body == nil {
		t.Fatal("StartPositionPolicyCommandServer 를 소스에서 찾지 못했다")
	}

	positions := map[string]int{}
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name := ""
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			name = fun.Name
		case *ast.SelectorExpr:
			if pkg, ok := fun.X.(*ast.Ident); ok {
				name = pkg.Name + "." + fun.Sel.Name
			}
		}
		if _, seen := positions[name]; !seen && name != "" {
			positions[name] = fileSet.Position(call.Pos()).Line
		}
		return true
	})

	drop, dropped := positions["dropStalePositionPolicyDescriptor"]
	if !dropped {
		t.Fatal("Start 가 죽은 run 의 descriptor 를 지우지 않는다 — 강등 부팅은 그 광고를 " +
			"그대로 두고, 소비자는 재활용된 포트에 토큰을 보낸다")
	}
	validate, validated := positions["positionpolicyrpc.ValidateControlDirectory"]
	if !validated {
		t.Fatal("Start 가 control 디렉터리를 검증하지 않는다 — 이 파일의 전제가 깨졌다")
	}
	if drop < validate {
		t.Errorf("descriptor 제거(%d줄)가 control 디렉터리 검증(%d줄)보다 **앞**이다 — "+
			"소유를 확인하기 전에 지우는 것이다", drop, validate)
	}
	listen, listened := positions["net.Listen"]
	if !listened {
		t.Fatal("Start 가 listen 하지 않는다 — 이 파일의 전제가 깨졌다")
	}
	if drop > listen {
		t.Errorf("descriptor 제거(%d줄)가 listen(%d줄)보다 **뒤**다 — 그 사이의 실패는 "+
			"죽은 run 의 광고를 그대로 남긴 채 강등된다", drop, listen)
	}
}
