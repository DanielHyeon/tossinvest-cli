package main

// static_test.go 는 "브로커에는 읽기만" 을 약속이 아니라 도달 가능성으로 고정함
// (선례: tools/a112-l1c-quote-probe 의 TestProbeReachesOnlyTheTwoStrictReaders).

import (
	"go/ast"
	"go/parser"
	"go/token"
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

// TestMarketSourceIsTheReadOnlyGets 는 브로커 표면(인터페이스)의 메서드 집합을 고정함.
// ExchangeRate 는 1.5 표본 선택의 US 가격 → 상한 통화 환산용 공개 GET(/api/v1/exchange-rate)으로 추가됨.
func TestMarketSourceIsTheReadOnlyGets(t *testing.T) {
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
	// 1.7: 봉(1m 엄격·1d)·유의·가격제한폭·지수 시세·지수 봉·시장 투자자 매매 — 전부 공개 시장 데이터 GET.
	want := []string{"Candles", "ExchangeRate", "MarketIndicatorCandles", "MarketIndicatorPrices", "MarketInvestorTrading",
		"PriceLimits", "Prices", "Rankings", "Stocks", "StrictMinuteCandles", "StrictOrderbookTop", "TypedMarketCalendar", "Warnings"}
	if !reflect.DeepEqual(methods, want) {
		t.Fatalf("marketSource methods = %v, want exactly the public-market GETs %v", methods, want)
	}
}

// TestNoMutatingOrAccountSurfaceIsNamed 는 주문·계좌 메서드 이름이 어떤 selector 로도 나오지 않고,
// official 패키지에서 쓰는 이름이 허용 목록 안에 있음을 봄. config 패키지도 같은 방식으로 묶음 —
// 주문 상한은 파일 읽기(NewService → Status)로만 얻고, 쓰기 메서드(Init·Save*)는 이름조차 못 나옴.
func TestNoMutatingOrAccountSurfaceIsNamed(t *testing.T) {
	t.Parallel()
	forbidden := map[string]bool{
		"PlaceOrder": true, "CancelOrder": true, "ModifyOrder": true,
		"CreateConditionalOrder": true, "CancelConditionalOrder": true, "ModifyConditionalOrder": true,
		"ModifyConditionalOrderRef": true, "Accounts": true, "BuyingPower": true, "Holdings": true,
		"HoldingsRaw": true, "SellableQuantity": true, "SellableQuantityRaw": true, "Orders": true,
		"OrderByID": true, "OrdersRaw": true, "OrdersPageRaw": true, "OrderRawByID": true, "AuthHeaders": true,
		"VerifyAuthoritativeAccountIdentity": true, "SelectedAccountSeq": true, "WithAccountSeq": true,
		"ConditionalOrders": true, "ConditionalOrder": true, "SaveCredentials": true, "DeleteCredentials": true,
		// config 쓰기 표면 — 프로브는 config 를 읽기만 함.
		"Init": true, "SaveEngineAdoption": true, "SaveEngineExitPolicy": true, "SaveEngineGateLimits": true,
		"SaveNotifications": true, "SaveNotificationsEnabled": true, "SaveTradingPolicy": true,
		"SaveEngineGateEnabled": true, "SaveEngineAutostart": true, "SaveSoakAutostart": true,
	}
	allowedOfficial := map[string]bool{
		"New": true, "LoadCredentials": true, "Client": true, "Credentials": true,
		"StrictTopOfBook": true, "MarketCalendarResponse": true, "ErrRateLimited": true, "StrictMinutePage": true,
	}
	allowedConfig := map[string]bool{"NewService": true}
	sawOfficial, sawConfig := false, false
	for _, file := range nonTestSources(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if forbidden[selector.Sel.Name] {
				t.Errorf("source names %s; the probe never reaches order or account surfaces", selector.Sel.Name)
			}
			if pkg, ok := selector.X.(*ast.Ident); ok && pkg.Name == "official" {
				sawOfficial = true
				if !allowedOfficial[selector.Sel.Name] {
					t.Errorf("source uses official.%s, outside the read-only allowlist", selector.Sel.Name)
				}
			}
			if pkg, ok := selector.X.(*ast.Ident); ok && pkg.Name == "tossconfig" {
				sawConfig = true
				if !allowedConfig[selector.Sel.Name] {
					t.Errorf("source uses tossconfig.%s, outside the read-only allowlist", selector.Sel.Name)
				}
			}
			return true
		})
	}
	if !sawOfficial || !sawConfig {
		t.Fatalf("the guard saw official.* %v / tossconfig.* %v — a side it saw nothing of is measured by nothing", sawOfficial, sawConfig)
	}
}

// TestOutboundBytesPassTheStateGuard 는 judge 와 storeState 가 직렬화 바이트를 가드에 통과시키는지
// 고정함. 지금은 PublicState 타입이 계좌 필드를 못 담아 행동 시험으로는 이 호출 삭제가 안 잡힘 —
// 둘째 겹이 조용히 사라지지 않게 구조로 못 박음.
func TestOutboundBytesPassTheStateGuard(t *testing.T) {
	t.Parallel()
	want := map[string]bool{"judge": false, "storeState": false}
	for _, file := range nonTestSources(t) {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if _, tracked := want[fn.Name.Name]; !tracked {
				continue
			}
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "assertPublicJSON" {
						want[fn.Name.Name] = true
					}
				}
				return true
			})
		}
	}
	for name, called := range want {
		if !called {
			t.Errorf("%s no longer passes its outbound bytes through assertPublicJSON", name)
		}
	}
}

// TestImportsStayInsideTheAllowlist 는 생산 패키지 import 를 읽기 표면 넷으로 묶고 새 외부 의존을 막음.
// internal/config 는 1.5 의 주문 상한 읽기용(config 파일 읽기만 — 쓰기 이름은 위 시험이 금지).
func TestImportsStayInsideTheAllowlist(t *testing.T) {
	t.Parallel()
	allowedModule := map[string]bool{
		"github.com/JungHoonGhae/tossinvest-cli/internal/official":  true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/domain":    true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/app/paths": true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/config":    true,
	}
	for _, file := range nonTestSources(t) {
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			if strings.Contains(path, ".") && !allowedModule[path] {
				t.Errorf("import %q is outside the allowlist (stdlib + official/domain/app-paths/config only)", path)
			}
		}
	}
}
