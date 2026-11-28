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
)

// a210BoxMarker 는 이 시험 바이너리가 이미 상자 안이라는 표지(값은 상자 경로). 자식 프로세스가 물려받음.
const a210BoxMarker = "A210_TEST_HOME_BOX"

// a210BoxedEnv 는 TossOS 기본 경로가 읽는 env 와 상자 안 하위 디렉터리 이름.
// journal.DataDir · candidate.DataDir 는 TOSSOS_DATA_DIR → XDG_DATA_HOME,
// config.DefaultPaths 는 os.UserConfigDir(XDG_CONFIG_HOME) · os.UserCacheDir(XDG_CACHE_HOME)를 읽음.
var a210BoxedEnv = map[string]string{
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
	// 부모 상자를 그대로 씀.
	if os.Getenv(a210BoxMarker) != "" {
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
	for name, sub := range a210BoxedEnv {
		if err := os.Setenv(name, filepath.Join(root, sub)); err != nil {
			fmt.Fprintln(os.Stderr, "a210: env 설정 실패:", name, err)
			return 1
		}
	}
	return m.Run()
}

// a210PinGoDirs 는 GOCACHE · GOENV 가 비어 있으면 go 의 기본 계산식(UserCacheDir/go-build,
// UserConfigDir/go/env)으로 지금 값을 채움. 이미 정해져 있으면 건드리지 않음.
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
