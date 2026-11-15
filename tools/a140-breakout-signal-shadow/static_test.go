package main

// static_test.go 는 「브로커에는 읽기만 · 설정은 읽기만 · 원장(journal)은 경로만」을 약속이 아니라 도달 가능성으로 고정함
// (선례: tools/a128-jev-shadow-probe/static_test.go).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func nonTestSources(t *testing.T) []*ast.File {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	paths, err := filepath.Glob(filepath.Join(filepath.Dir(thisFile), "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	var files []*ast.File
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		files = append(files, parsed)
	}
	if len(files) == 0 {
		t.Fatal("no source files found")
	}
	return files
}

func TestMarketSourceIsTheFiveReadOnlyGets(t *testing.T) {
	t.Parallel()
	var methods []string
	for _, file := range nonTestSources(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			spec, ok := node.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "marketSource" {
				return true
			}
			for _, field := range spec.Type.(*ast.InterfaceType).Methods.List {
				for _, name := range field.Names {
					methods = append(methods, name.Name)
				}
			}
			return false
		})
	}
	sort.Strings(methods)
	want := []string{"Rankings", "StrictLastPrice", "StrictMinuteCandles", "StrictOrderbookTop", "TypedMarketCalendar"}
	if !reflect.DeepEqual(methods, want) {
		t.Fatalf("marketSource methods = %v, want exactly %v", methods, want)
	}
}

func TestNoMutatingAccountOrWriteSurfaceIsNamed(t *testing.T) {
	t.Parallel()
	forbidden := map[string]bool{
		"PlaceOrder": true, "CancelOrder": true, "ModifyOrder": true, "CreateConditionalOrder": true, "CancelConditionalOrder": true,
		"ModifyConditionalOrder": true, "ModifyConditionalOrderRef": true, "Accounts": true, "BuyingPower": true, "Holdings": true,
		"HoldingsRaw": true, "SellableQuantity": true, "SellableQuantityRaw": true, "Orders": true, "OrderByID": true, "OrdersRaw": true,
		"OrdersPageRaw": true, "OrderRawByID": true, "AuthHeaders": true, "SelectedAccountSeq": true, "WithAccountSeq": true,
		"ConditionalOrders": true, "ConditionalOrder": true, "SaveCredentials": true, "DeleteCredentials": true,
		// config 쓰기 표면.
		"Init": true, "SaveEngineAdoption": true, "SaveEngineExitPolicy": true, "SaveEngineGateLimits": true, "SaveNotifications": true,
		"SaveNotificationsEnabled": true, "SaveTradingPolicy": true, "SaveEngineGateEnabled": true, "SaveEngineAutostart": true,
		"SaveSoakAutostart": true,
		// a166: 예산 핸들 메서드 — 도구는 예산을 시작하고 보조 등급만 다므로 핸들을 직접 잡거나 엔진 관측 상태를 읽지 않음.
		"Budget": true, "Acquire": true, "EnginePollKey": true, "NewObservationFence": true, "ObservationSnapshot": true,
		"PollRemaining": true, "VerifyIdentity": true,
	}
	allowed := map[string]map[string]bool{
		"official": {"New": true, "LoadCredentials": true, "Client": true, "StrictTopOfBook": true, "StrictLastPrice": true,
			"StrictMinutePage": true, "MarketCalendarResponse": true, "ErrRateLimited": true,
			// a166: 예산에 묶인 클라이언트 생성(factory.NewClient)에 넘기는 옵션 타입뿐임 — With… 생성자는 여전히 목록 밖.
			"Option": true},
		"tossconfig": {"NewService": true},
		// 원장 패키지는 생산 evidence.db 의 자리를 알아내는 데만 씀 — 열기 · 쓰기 이름은 못 나옴.
		"journal": {"DefaultPath": true, "DBFileName": true},
		// a166: 계좌 공유 요청 예산 연결 — 예산 시작과 보조 우선순위 표시 이름만 허용함.
		"accountbudget":        {"WithPriority": true, "Auxiliary": true},
		"accountbudgetruntime": {"ParseLaunchConfig": true, "Open": true},
	}
	saw := map[string]bool{}
	// a166: 공유 예산에서 Critical 은 엔진 보호 주문 몫임 — WithPriority 는 직접 호출로만, 둘째 인자는 선택자 accountbudget.Auxiliary 로만 허용함.
	priorityRefs, priorityCalls := 0, 0
	for _, file := range nonTestSources(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok && types.ExprString(call.Fun) == "accountbudget.WithPriority" {
				priorityCalls++
				if len(call.Args) != 2 || types.ExprString(call.Args[1]) != "accountbudget.Auxiliary" {
					t.Errorf("source calls %s; the tool takes only accountbudget.Auxiliary on the shared budget", types.ExprString(call))
				}
			}
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if forbidden[selector.Sel.Name] {
				t.Errorf("source names %s; the tool never reaches order, account or config-write surfaces", selector.Sel.Name)
			}
			if types.ExprString(selector) == "accountbudget.WithPriority" {
				priorityRefs++
			}
			if pkg, ok := selector.X.(*ast.Ident); ok {
				if list, tracked := allowed[pkg.Name]; tracked {
					saw[pkg.Name] = true
					if !list[selector.Sel.Name] {
						t.Errorf("source uses %s.%s, outside the allowlist", pkg.Name, selector.Sel.Name)
					}
				}
			}
			return true
		})
	}
	for pkg := range allowed {
		if !saw[pkg] {
			t.Errorf("the guard saw nothing of %s.* — a side it saw nothing of is measured by nothing", pkg)
		}
	}
	// 호출 자리가 0 이면 위 고정이 아무것도 안 잰 것이고, 직접 호출이 아닌 참조(함수 값)는 인자 검사를 비켜 감.
	if priorityCalls == 0 || priorityRefs != priorityCalls {
		t.Fatalf("accountbudget.WithPriority direct calls %d / references %d — every reference must be a checked direct call, and there must be one", priorityCalls, priorityRefs)
	}
}

func TestImportsStayInsideTheAllowlist(t *testing.T) {
	t.Parallel()
	const module = "github.com/JungHoonGhae/tossinvest-cli/internal/"
	allowed := map[string]bool{
		"official": true, "domain": true, "app/paths": true, "config": true, "clock": true, "scheduler": true,
		"strategyevidence": true, "officialbars": true, "breakoutsignal": true, "breakoutlane": true, "costs": true, "journal": true,
		// a166: 계좌 공유 요청 예산 연결(쓰는 이름은 위 시험이 고정).
		"accountbudget": true, "accountbudgetruntime": true,
	}
	for _, file := range nonTestSources(t) {
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			if !strings.Contains(path, ".") {
				continue
			}
			if !strings.HasPrefix(path, module) || !allowed[strings.TrimPrefix(path, module)] {
				t.Errorf("import %q is outside the allowlist", path)
			}
		}
	}
}
