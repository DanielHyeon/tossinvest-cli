package engine_test

// a210 리뷰 F1 — 상자 표지(A210_TEST_HOME_BOX)를 물려받은 자식이 그 표지를 믿기 전에 검증하는지 잼.
// 위조 · 잔류 표지를 그대로 믿으면 경로 env 가 상자를 가리키지 않은 채 시험이 돌아 기본 경로가 실제 HOME 으로 떨어짐.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a210ChildProbeEnv 가 1 일 때만 자식 탐침 시험이 일함(평소 실행에서는 아무것도 안 함).
const a210ChildProbeEnv = "A210_CHILD_PROBE"

// TestA210BoxChildProbe 는 자식 프로세스 안에서만 일하는 탐침 — 기본 데이터 경로에 파일 하나를 씀.
// 부모 시험은 그 파일이 자식의 HOME(부모의 임시 디렉터리) 아래 생겼는지로 「상자 밖으로 떨어졌는지」를 봄.
func TestA210BoxChildProbe(t *testing.T) {
	if os.Getenv(a210ChildProbeEnv) != "1" {
		t.Skip("자식 프로세스 탐침 전용")
	}
	dir, err := journal.DataDir()
	if err != nil {
		t.Fatalf("DataDir: %v", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a210-probe"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
}

// a210ChildCase 는 자식에게 줄 표지 · 추가 env · 작업 디렉터리.
type a210ChildCase struct {
	marker string
	env    []string
	dir    string
}

// TestA210ForgedBoxMarkerIsRefused 는 위조(없는 경로) · 잔류(있지만 경로 env 와 안 맞음) · 상대경로 표지를 받은 자식이
// 시험을 돌리지 않고 a210 거절 문구와 rc ≠ 0 으로 끝나며, 자식 HOME 아래 아무것도 쓰지 않는지 단언함.
// 거절 문구로 단언하는 이유: rc ≠ 0 만 보면 아래 계층(journal.DataDir 의 상대경로 거절 등)이 대신 막은 경우도 통과함.
func TestA210ForgedBoxMarkerIsRefused(t *testing.T) {
	cases := map[string]func(t *testing.T) a210ChildCase{
		"forged-missing-dir": func(t *testing.T) a210ChildCase {
			return a210ChildCase{marker: filepath.Join(t.TempDir(), "no-such-box")}
		},
		"stale-existing-dir": func(t *testing.T) a210ChildCase { return a210ChildCase{marker: t.TempDir()} },
		// 상대경로 표지 — 자식 작업 디렉터리 안에 실제로 있고 경로 env 넷도 그 안을 가리켜 「존재 · 일치」 검사만으로는 통과하는 모양.
		"relative-marker": func(t *testing.T) a210ChildCase {
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, "rel5"), 0o700); err != nil {
				t.Fatal(err)
			}
			return a210ChildCase{marker: "rel5", dir: dir, env: []string{
				"TOSSOS_DATA_DIR=" + filepath.Join("rel5", "tossos"), "XDG_DATA_HOME=" + filepath.Join("rel5", "data"),
				"XDG_CONFIG_HOME=" + filepath.Join("rel5", "config"), "XDG_CACHE_HOME=" + filepath.Join("rel5", "cache"),
			}}
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			c := build(t)
			home := t.TempDir()
			cmd := exec.Command(os.Args[0], "-test.run=^TestA210BoxChildProbe$", "-test.count=1")
			cmd.Env = append(a210ChildEnv(home, c.marker), c.env...)
			cmd.Dir = c.dir
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Errorf("표지 %s 를 받은 자식이 rc 0 으로 시험을 돌림 — 표지를 검증 없이 믿음\n%s", name, out)
			}
			if !strings.Contains(string(out), "a210: 물려받은 상자 표지 거절:") {
				t.Errorf("자식이 a210 표지 검증에서 거절되지 않음(다른 지점에서 멈췄거나 시험을 돌림)\n%s", out)
			}
			var leaked []string
			_ = filepath.WalkDir(home, func(p string, d os.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					leaked = append(leaked, p)
				}
				return nil
			})
			if len(leaked) != 0 {
				t.Errorf("자식이 HOME 아래 파일을 씀(상자 밖으로 떨어짐): %v", leaked)
			}
		})
	}
}

// a210ChildEnv 는 지금 env 에서 경로 env · 표지를 빼고 HOME · 표지 · 탐침 스위치를 새로 준 자식 env.
func a210ChildEnv(home, marker string) []string {
	drop := map[string]bool{"HOME": true, a210ChildProbeEnv: true, "A210_TEST_HOME_BOX": true,
		"TOSSOS_DATA_DIR": true, "XDG_DATA_HOME": true, "XDG_CONFIG_HOME": true, "XDG_CACHE_HOME": true}
	var env []string
	for _, kv := range os.Environ() {
		if name, _, _ := strings.Cut(kv, "="); !drop[name] {
			env = append(env, kv)
		}
	}
	return append(env, "HOME="+home, "A210_TEST_HOME_BOX="+marker, a210ChildProbeEnv+"=1")
}
