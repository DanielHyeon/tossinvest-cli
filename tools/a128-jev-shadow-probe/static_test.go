package main

// static_test.go 는 "브로커에는 읽기만" 을 약속이 아니라 도달 가능성으로 고정함
// (선례: tools/a112-l1c-quote-probe 의 TestProbeReachesOnlyTheTwoStrictReaders).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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
		// 1.8: WTS 조회 셋(wts.go — 표면은 아래 시험이 고정)과 그 세션 저장소 읽기.
		"github.com/JungHoonGhae/tossinvest-cli/internal/client":  true,
		"github.com/JungHoonGhae/tossinvest-cli/internal/session": true,
	}
	for _, file := range nonTestSources(t) {
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			if strings.Contains(path, ".") && !allowedModule[path] {
				t.Errorf("import %q is outside the allowlist (stdlib + official/domain/app-paths/config/client/session only)", path)
			}
		}
	}
}

// TestWTSSourceIsTheThreeReads 는 WTS 표면(인터페이스)의 메서드 집합을 조회 셋으로 고정하고, WTS 클라이언트의 나머지
// 공개 메서드(주문·계좌·보유·관심종목 쓰기 등 — internal/client 소스에서 AST 로 뽑음)가 어떤 selector 로도 나오지 않으며,
// tossclient·wtssession 패키지에서 쓰는 이름이 허용 목록 안에 있음을 봄. 새 WTS 메서드가 생겨도 이 목록이 자동으로 막음.
func TestWTSSourceIsTheThreeReads(t *testing.T) {
	t.Parallel()
	allowedMethods := map[string]bool{"GetMarketNews": true, "GetMarketIndices": true, "GetTradingFlows": true}
	var methods []string
	for _, file := range nonTestSources(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			spec, ok := node.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "wtsSource" {
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
	if want := []string{"GetMarketIndices", "GetMarketNews", "GetTradingFlows"}; !reflect.DeepEqual(methods, want) {
		t.Fatalf("wtsSource methods = %v, want exactly the read-only trio %v", methods, want)
	}

	// internal/client 의 *Client 공개 메서드 전부 — 허용 셋 밖은 이름조차 나오면 안 됨.
	_, thisFile, _, _ := runtime.Caller(0)
	clientDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "client")
	paths, err := filepath.Glob(filepath.Join(clientDir, "*.go"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("cannot list internal/client sources: %v", err)
	}
	forbidden := map[string]bool{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 || !fn.Name.IsExported() || allowedMethods[fn.Name.Name] {
				continue
			}
			// *Client 의 메서드만 — 오류 타입의 Error 같은 메서드는 WTS 표면이 아님.
			star, isStar := fn.Recv.List[0].Type.(*ast.StarExpr)
			if !isStar {
				continue
			}
			if ident, isIdent := star.X.(*ast.Ident); !isIdent || ident.Name != "Client" {
				continue
			}
			forbidden[fn.Name.Name] = true
		}
	}
	for _, name := range []string{"PlacePendingOrder", "CancelPendingOrder", "AmendPendingOrder", "ListPositions", "GetAccountSummary"} {
		if !forbidden[name] {
			t.Fatalf("the WTS method census missed %s — the guard would measure nothing", name)
		}
	}
	allowedPackage := map[string]map[string]bool{
		"tossclient": {"New": true, "Config": true, "Client": true, "AuthError": true, "StatusError": true, "MaxNewsLimit": true},
		"wtssession": {"NewFileStore": true, "ErrNoSession": true, "Session": true},
	}
	saw := map[string]bool{}
	for _, file := range nonTestSources(t) {
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if forbidden[selector.Sel.Name] {
				t.Errorf("source names WTS method %s; the probe reaches only the read-only trio", selector.Sel.Name)
			}
			if pkg, ok := selector.X.(*ast.Ident); ok && allowedPackage[pkg.Name] != nil {
				saw[pkg.Name] = true
				if !allowedPackage[pkg.Name][selector.Sel.Name] {
					t.Errorf("source uses %s.%s, outside the read-only allowlist", pkg.Name, selector.Sel.Name)
				}
			}
			return true
		})
	}
	if !saw["tossclient"] || !saw["wtssession"] {
		t.Fatalf("the guard saw tossclient %v / wtssession %v — a side it saw nothing of is measured by nothing", saw["tossclient"], saw["wtssession"])
	}
}

// TestNoPersonalizedNewsScopeInSource 는 계좌 맥락 뉴스 범위(PERSONALIZED·PERSONALIZE_WATCH·PERSONALIZE_HOLD — 보유·관심
// 종목을 드러냄)와 별칭이 소스 어디에도 없음을 봄. 피드 범위 목록은 계좌 맥락이 없는 범위만.
func TestNoPersonalizedNewsScopeInSource(t *testing.T) {
	t.Parallel()
	_, thisFile, _, _ := runtime.Caller(0)
	paths, _ := filepath.Glob(filepath.Join(filepath.Dir(thisFile), "*.go"))
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, word := range []string{"PERSONALIZ", "GetNewsBriefing", `"holdings"`, `"watchlist"`, `"recommended"`} {
			if strings.Contains(string(raw), word) {
				t.Errorf("%s mentions %s — account-context news never feeds the state", filepath.Base(path), word)
			}
		}
	}
	for _, scope := range newsFeedScopes {
		if scope != "ALL_HIGHLIGHT" && scope != "SOARING_STOCK" && scope != "HOT" {
			t.Errorf("news scope %q is outside the account-free set", scope)
		}
	}
}
