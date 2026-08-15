//go:build unix

package main

// a109_the_lock_outlives_every_endpoint_close_test.go — T2 §2.4 (design D5a).
//
// `runEngineRun` 에는 테스트가 없는 불변식이 하나 있다:
//
//	**endpoint 의 Close 는 journal flock 을 쥔 채로 돈다.**
//
// 그 성질은 defer **등록 순서** 하나에 얹혀 있다. `lock.Release()` 가 가장 먼저
// 등록되므로 LIFO 상 가장 나중에 실행되고, 그래서 이 프로세스의 잔재 제거와 다음
// 엔진의 회수·발행이 겹칠 수 없다. 회수 함수들이 flock 을 자기 1차 방어로 인용하는
// 근거가 정확히 이 순서다(design D2).
//
// # 왜 정적(go/parser) 검사인가
//
// 이 불변식은 **실행으로 잴 수 없다.** 재려면 두 엔진 프로세스를 종료 경합에 놓고
// 특정 인터리빙을 재현해야 하는데, 그 테스트는 통과해도 아무것도 증명하지 못한다
// (a099 R1 이 배운 것: 통과는 증거가 아니다). 반면 순서는 소스에 그대로 있고,
// 그것을 옮기는 편집은 파서가 100% 잡는다.
//
// 이 파일이 잡는 실제 사고 모양: 누군가 강등 처리를 정리하면서 `defer lock.Release()`
// 를 endpoint 기동 **아래**로 옮긴다. 그러면 flock 이 먼저 풀리고, autostart 가 세운
// 다음 엔진의 회수가 이 프로세스의 Close 와 같은 디렉터리에서 겹친다.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// a109EndpointCloseDefers 는 부팅 7단계가 등록하는 endpoint Close defer 들이다.
// 하나라도 사라지면 이 테스트가 먼저 운다 — "순서는 맞는데 대상이 없어졌다"를
// 통과로 세지 않기 위해서다.
var a109EndpointCloseDefers = []string{
	"policyControl.Close", "policyRuntime.Close", "strategyRuntime.Close", "alertControl.Close",
}

// a109LockReleaseDefer 는 부팅 1단계의 배타성 해제다.
const a109LockReleaseDefer = "lock.Release"

// a109DeferTargets 는 함수 본문이 **자기 수준에서** 등록하는 defer 들을 등록 순서대로
// 모은다.
//
// 중첩 함수 리터럴 안으로 들어가지 않는 이유는 그 안의 defer 는 **다른 함수의 defer
// 스택**에 쌓이기 때문이다. 같이 세면 `defer func() { _ = ectx.Close() }()` 같은 줄이
// 순서를 왜곡한다.
func a109DeferTargets(body *ast.BlockStmt) []string {
	var targets []string
	var walk func(node ast.Node)
	walk = func(node ast.Node) {
		ast.Inspect(node, func(n ast.Node) bool {
			switch typed := n.(type) {
			case *ast.FuncLit:
				return false // 다른 defer 스택이다
			case *ast.DeferStmt:
				targets = append(targets, a109CallText(typed.Call.Fun))
				// 인자 안의 함수 리터럴까지 훑을 이유가 없다.
				return false
			}
			return true
		})
	}
	walk(body)
	return targets
}

// a109CallText 는 defer 대상의 이름을 사람이 읽는 모양으로 만든다.
func a109CallText(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return a109CallText(typed.X) + "." + typed.Sel.Name
	case *ast.FuncLit:
		return "func literal"
	}
	return "?"
}

func a109RunEngineRunBody(t *testing.T) *ast.BlockStmt {
	t.Helper()
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, "engine.go", nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("engine.go 파싱: %v", err)
	}
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "runEngineRun" && function.Body != nil {
			return function.Body
		}
	}
	t.Fatal("engine.go 에서 runEngineRun 을 찾지 못했다")
	return nil
}

// TestTheJournalLockIsReleasedAfterEveryEndpointClose 는 D5a 그 자체다.
func TestTheJournalLockIsReleasedAfterEveryEndpointClose(t *testing.T) {
	targets := a109DeferTargets(a109RunEngineRunBody(t))

	lockAt := -1
	for index, target := range targets {
		if target == a109LockReleaseDefer {
			if lockAt >= 0 {
				t.Fatalf("%s defer 가 두 번 등록됐다: %v", a109LockReleaseDefer, targets)
			}
			lockAt = index
		}
	}
	if lockAt < 0 {
		t.Fatalf("%s defer 가 없다 — 배타성이 명령이 끝나기 전에 풀리는지조차 알 수 없다: %v",
			a109LockReleaseDefer, targets)
	}

	for _, endpoint := range a109EndpointCloseDefers {
		closeAt := -1
		for index, target := range targets {
			if target == endpoint {
				closeAt = index
				break
			}
		}
		if closeAt < 0 {
			t.Errorf("%s defer 가 사라졌다 — endpoint 하나가 명령이 끝나도 안 닫힌다: %v",
				endpoint, targets)
			continue
		}
		if lockAt >= closeAt {
			t.Errorf("%s defer(등록 #%d)가 %s(등록 #%d)보다 먼저 등록되지 않았다.\n"+
				"defer 는 LIFO 이므로 flock 이 이 Close 보다 **먼저** 풀린다 — 그러면 "+
				"이 프로세스의 잔재 제거와 다음 엔진의 회수·발행이 같은 디렉터리에서 겹친다.\n"+
				"등록 순서: %v",
				a109LockReleaseDefer, lockAt, endpoint, closeAt, targets)
		}
	}

	// 순서를 옮기지 않고 **조건 안에 숨기는** 편집도 같은 사고를 만든다. lock 해제는
	// 어떤 분기에도 들어가면 안 된다: 이 defer 는 함수 최상위 문장이어야 한다.
	if !a109LockReleaseIsTopLevel(t) {
		t.Error("lock.Release defer 가 조건문 안으로 들어갔다 — 그 분기를 타지 않는 경로에서 " +
			"배타성이 명령보다 오래 산다")
	}
}

// a109LockReleaseIsTopLevel 은 그 defer 가 함수 본문의 **직속** 문장인지 본다.
func a109LockReleaseIsTopLevel(t *testing.T) bool {
	t.Helper()
	for _, statement := range a109RunEngineRunBody(t).List {
		deferred, ok := statement.(*ast.DeferStmt)
		if ok && a109CallText(deferred.Call.Fun) == a109LockReleaseDefer {
			return true
		}
	}
	return false
}

// TestTheDegradationCommentStillCitesTheForbiddenThree 는 문서의 회귀 핀이다.
//
// 금지 3종의 이유는 **코드로 표현할 수 없다** — 등급표에 이름을 올리는 것은 다른
// 패키지의 한 줄이고, 그것을 막는 것은 이 주석뿐이다. 주석이 사라지면 다음 사람은
// 그 한 줄이 실계좌의 신규 진입을 영구 차단한다는 것을 알 길이 없다.
//
// # 인용은 **고유해야** 한다 — a109 §2-fix F4 (A2 P2-4)
//
// 처음 두 단정은 공허했다. `"critical"` 은 바로 윗줄의 `"criticalEvents"` 안에 이미
// 들어 있고, `"flock"` 은 이 파일과 무관한 영문 헤더(engine.go:14)가 만족시킨다.
// 실측으로 확인했다: 금지 2번 줄(engine.go:366)을 지워도, 순서 불변식의 flock 문장을
// 지워도 이 테스트는 **초록이었다**(뮤테이션 원장 M24·M25).
//
// 그래서 각 인용을 그 문장에만 있는 **고유 구절**로 바꾼다. 정적 핀이 재는 것은
// 「글자가 파일 어딘가에 있는가」가 아니라 「그 금지의 이유가 아직 적혀 있는가」다.
func TestTheDegradationCommentStillCitesTheForbiddenThree(t *testing.T) {
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, "engine.go", nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("engine.go 파싱: %v", err)
	}
	var comments strings.Builder
	for _, group := range parsed.Comments {
		comments.WriteString(group.Text())
		comments.WriteString("\n")
	}
	text := comments.String()
	for _, required := range []string{
		"criticalEvents", // 등급표 등재 금지
		"severity 를 critical 로 올리지 마라",      // severity 승격 금지 (그 문장에만 있는 구절)
		"원장 outbox(`Journal.EnqueueAlert`)", // 원장 outbox 적재 금지
		"restoreAlertEntryLatch",            // 왜 그것이 사고인가
		"flock을 쥔 채로",                       // defer 순서 불변식의 근거 (영문 헤더로는 못 만족한다)
	} {
		if !strings.Contains(text, required) {
			t.Errorf("engine.go 주석이 %q 를 더 이상 인용하지 않는다 — "+
				"그 인용이 사라지면 금지의 이유도 사라진다", required)
		}
	}
}
