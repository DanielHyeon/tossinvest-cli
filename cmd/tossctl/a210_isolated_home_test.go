package main

// a210_isolated_home_test.go — 이 패키지 시험 바이너리 전체를 임시 상자 안에 가둠.
//
// 엔진을 조립하는 시험(예: TestA166EngineDecoratorCannotRemoveRequiredBudget)은 --config-dir 만 임시로 주고
// AuditFile 을 비워 둠. 그러면 engine.openAuditLog 가 journal.DataDir() 기본값 — 실제
// ~/.local/share/tossos/audit.log, 즉 운영 컨테이너가 붙여 쓰는 audit — 에 기동 설정 행을 씀(a210 측정).
// 시험 하나하나를 고치는 대신 기본 경로가 읽는 env 를 TestMain 에서 상자로 돌려,
// 지금 시험과 앞으로 더해질 시험이 함께 갇히게 함. internal/app/engine 의 같은 이름 파일과 같은 내용.

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

// a210BoxMarker 는 이 시험 바이너리가 이미 상자 안이라는 표지(값은 상자 경로). 자식 프로세스가 물려받음.
const a210BoxMarker = "A210_TEST_HOME_BOX"

// a210BoxedPaths 는 testenv.Env 중 경로 env 와 상자 안 하위 디렉터리 이름.
// journal.DataDir · candidate.DataDir 는 TOSSOS_DATA_DIR → XDG_DATA_HOME,
// config.DefaultPaths 는 os.UserConfigDir(XDG_CONFIG_HOME) · os.UserCacheDir(XDG_CACHE_HOME)를 읽음.
// testenv.Env 의 나머지 이름(자격 증명 · 언어)은 빈 값으로 둠 — testenv.Isolate 와 같은 규칙.
// shortcut: testenv.Env 에 경로 env 가 새로 더해지면 여기 없으면 빈 값(=기본 경로)이 됨, 그때 이 표에도 더함.
var a210BoxedPaths = map[string]string{
	"TOSSOS_DATA_DIR": "tossos",
	"XDG_DATA_HOME":   "data",
	"XDG_CONFIG_HOME": "config",
	"XDG_CACHE_HOME":  "cache",
}

// a210RunBoxed 는 상자를 만들고 env 를 돌린 뒤 시험을 돌리고 상자를 지움.
// t.Setenv 로 자기 경로를 주는 시험은 그 시험 동안만 이 값 위에 덮고, 끝나면 이 값으로 돌아옴.
func a210RunBoxed(m *testing.M) int {
	// 시험이 자기 바이너리를 자식으로 다시 띄우면(a135 · a166 · mcp 시험) 자식은 부모가 돌려 둔 env 를 물려받음.
	// 자식이 상자를 새로 만들면 panic · os.Exit 로 끝나는 자식은 defer 가 안 돌아 빈 상자가 임시 디렉터리에 남으므로
	// 부모 상자를 그대로 씀. 단 표지를 믿기 전에 검증함 — 위조 · 잔류 표지면 경로 env 가 상자를 안 가리켜
	// 기본 경로가 실제 HOME 으로 떨어지므로 시험을 돌리지 않고 거절함(fail-closed).
	if marker := os.Getenv(a210BoxMarker); marker != "" {
		if err := a210CheckInheritedBox(marker); err != nil {
			fmt.Fprintln(os.Stderr, "a210: 물려받은 상자 표지 거절:", err)
			return 1
		}
		return m.Run()
	}
	root, err := os.MkdirTemp("", "a210-tossctl-home-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "a210: 시험 상자 생성 실패:", err)
		return 1
	}
	defer os.RemoveAll(root)

	// XDG_* 를 돌리기 전에 go 하위 프로세스(go list · go env 를 부르는 시험)의 빌드 캐시 · go env 파일 위치를
	// 지금 값으로 고정함 — 안 그러면 go 가 상자 안에 빈 캐시를 새로 만들어 느려지고 사용자 go env 설정을 못 읽음.
	if err := a210PinGoDirs(); err != nil {
		fmt.Fprintln(os.Stderr, "a210: go 캐시 위치 고정 실패:", err)
		return 1
	}
	if err := os.Setenv(a210BoxMarker, root); err != nil {
		fmt.Fprintln(os.Stderr, "a210: env 설정 실패:", a210BoxMarker, err)
		return 1
	}
	for _, name := range testenv.Env {
		value := ""
		if sub, ok := a210BoxedPaths[name]; ok {
			value = filepath.Join(root, sub)
		}
		if err := os.Setenv(name, value); err != nil {
			fmt.Fprintln(os.Stderr, "a210: env 설정 실패:", name, err)
			return 1
		}
	}
	return m.Run()
}

// a210CheckInheritedBox 는 물려받은 표지가 있는 디렉터리를 가리키고, 경로 env 넷이 전부 그 상자 안
// 정해진 자리인지 확인함. 하나라도 어긋나면 오류 반환. 상대경로 표지는 작업 디렉터리에 따라 뜻이 바뀌므로 거절함.
// shortcut: 표지가 진짜 임시 상자인지(심링크 · os.TempDir 아래)는 안 봄 — env 다섯을 맞춘 의도적 위조까지 막아야 하면 Lstat + TempDir 검사 추가.
func a210CheckInheritedBox(marker string) error {
	if !filepath.IsAbs(marker) {
		return fmt.Errorf("%q 는 절대경로가 아님", marker)
	}
	info, err := os.Stat(marker)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 는 디렉터리가 아님", marker)
	}
	for name, sub := range a210BoxedPaths {
		if got, want := os.Getenv(name), filepath.Join(marker, sub); got != want {
			return fmt.Errorf("%s=%q 가 상자 안 %q 가 아님", name, got, want)
		}
	}
	return nil
}

// a210PinGoDirs 는 GOCACHE · GOENV 가 비어 있으면 go 의 기본 계산식(UserCacheDir/go-build,
// UserConfigDir/go/env)으로 지금 값을 채움. 이미 정해져 있으면 건드리지 않음.
// 상자로 가는 것은 데이터 · 설정 · tossctl 캐시 기본 경로뿐이고, 자식 go 는 이 고정값(실제 go-build 캐시)에 씀.
// shortcut: GOENV 파일의 `go env -w GOCACHE` 설정은 무시함(env 가 이김), 그런 설정이 생기면 `go env GOCACHE` 로 읽도록 바꿈.
func a210PinGoDirs() error {
	if os.Getenv("GOCACHE") == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			return err
		}
		if err := os.Setenv("GOCACHE", filepath.Join(dir, "go-build")); err != nil {
			return err
		}
	}
	if os.Getenv("GOENV") == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		if err := os.Setenv("GOENV", filepath.Join(dir, "go", "env")); err != nil {
			return err
		}
	}
	return nil
}
