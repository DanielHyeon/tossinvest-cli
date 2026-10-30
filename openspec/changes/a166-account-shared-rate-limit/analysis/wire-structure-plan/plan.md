# a166 2.4 생산 송신 구조검사 — 제안 / 구현 전

기준 main 6e0968222993756a41065044996f674b35e20a27. engine poll은 다른 writer가 수정 중이므로 그 결과가 고정·통합된 SHA에서 최종 검사 목록과 증거를 다시 만든다. 현재 검사는 계획과 source 손읽기이며 전체 AST/값 흐름 검증을 통과했다고 주장하지 않는다. 사전 재사용 자료: runtime-v2-token-lot/final-wire-census.md는 ec8+v2 결합 손읽기였고 이번에는 offline tool 추가와 현재 source를 대조했다. `source-sites.txt`는 비test Go의 New/NewClient/Budget/Open/NewTransport/Do/RoundTrip 검색 원문 79행(동명 비HTTP Do 포함), 파일 SHA는 아래 receipts. 완전 자동 call graph가 아니다.

## source 경로 분류

| 분류 | 대상과 필수 확인 |
|---|---|
| required root / official | cmd/tossctl/main.go의 단일 root owner → root_budget.go get → runtime.Open required v2 OpenObserved/VerifyIdentity → Factory.NewClient → official.NewWithAccountBudget → 같은 Budget NewTransport → client.hc 및 tokenManager.hc 결속. official.client.doRequest와 token.exchange의 각 Do 및 401 send retry를 실제 receiver별로 확인한다. cmd root.go → app.New provider → hybrid client 사용도 확인한다. |
| required engine / flatten / replay | engine_assembly.go owner.get → Options.BudgetFactory → engine.NewOrderPath/Context의 NewClient 및 Budget → gateway의 NewBudgetedHTTPReplay. flatten의 owner.get → NewOrderPath도 따로 검사. replay의 인증 official.UsesAccountBudget와 direct guarded HTTP의 Budget이 같은 값이며 nil/이상시 전송 0. |
| required a112 clone | a112_mbus_read.go에서 sealed client.budget과 authority clone → NewTransport(same budget) → hc.Do, nil branch는 OFF만. token은 재교환하지 않고 캐시만 쓴다. |
| required 별도 probe | a087 main execute에서 production prepareProbeBudget → factory Budget/client → UsesAccountBudget → newBudgetedSendClient(same budget) 및 source.AuthHeaders → sendOnce. 조건이 required인데 raw openOfficial/newSendClient로 가지 않음을 분기 구조와 동작으로 검사한다. |
| required 도구·CLI | a112-l1c, a112-mb-us, a128-d2, a128-jev, a130, a140 각 env parser → runtime.Open → Factory.NewClient; dry-run/빈 lots는 송신 0. verify/console/openapi/MCP/reconcile/candidate/soak의 root-bound builder 호출자와 실패 분기를 함수별로 pin. 생성자 이름만 있는 죽은 참조로 통과시키지 않는다. |
| OFF / 저수준 | root_budget.go의 nil owner raw New, app.New의 nil provider, engine.NewOrderPath의 nil factory, runtime ModeOff의 raw New, 공개 HTTPReplay 및 a087 openOfficial/newSendClient는 OFF/명시 저수준 경로. production required 제어 흐름에서 도달하면 실패; 외부 임의 Go 프로그램까지 자동 강제라고 주장하지 않는다. unsupported 플랫폼 required는 오류, OFF는 기존 동작. |
| offline | accountbudgetruntime/offline_linux.go의 OpenContext는 빈 새 root를 명시 provision할 때 임시 Budget을 여는 경로이고 client/HTTP가 없다. 관측 enforcement용 OpenObserved와 혼동하지 않는다. |
| WTS / judge / 비공식 | internal/client·monitor WTS, a128 wts.go, 외부 TypeSafe judge, release/update·ntfy·push·로컬 RPC·전략증거 SEC/DART는 각각 URL/authority·호출 receiver 근거를 확인해 공식 Open API의 감시 대상과 분리한다. monitor의 공식 URL registry는 contract-only이며 WTS runner가 호출하지 않는다. unknown direct HTTP는 이유 없는 제외 대신 실패·수동 분류 대상으로 둔다. |
| 테스트 support | `_test.go`, internal/testenv 등 test-only support는 production 목록과 별도로 고정한다. build tag별 non-test 파일(예: unsupported stub)과 generator 산출물은 무심코 빼지 않는다. sync.Once.Do는 HTTP가 아니며 receiver/type 근거로 분류한다. |

## 최소 구현 계획 — Manager 확인 후

1. `go/parser`/`go/ast`와 import alias 해석을 사용해 프로젝트 cmd/internal/tools 비test source의 공식 constructor·budget factory/open·HTTP request/Do/RoundTrip 호출과 함수값·interface escape를 식별한다. 모든 발견 위치를 파일·함수·receiver·분류·근거의 고정 목록과 대조하고 새로운 미분류 호출은 실패한다. 동명 `sync.Once.Do` 등은 명시 type/수신자 근거로 식별하며 `official` alias, dot import, 함수값 전달, shadowing, 미해석 dynamic receiver를 조용히 통과시키지 않는다. scan/목록은 발견 장치일 뿐 required 값 흐름 판정이 아니다. 전체 빌드태그 후보 분류와 alias를 명시하고 모듈 밖 프로그램은 범위 밖이다.
2. 실제 소수 경계 함수에 한정해 AST 분기·대입·호출·return 결속을 검사한다: root owner 생성과 mode에 따른 error 반환, factory Required→OpenObserved 및 identity 성공→같은 Budget NewWithAccountBudget, sealed client/token 동일 hc, engine/flatten factory 전달, a112/replay/probe의 같은 Budget과 guarded Do. 필요하면 현재 구조를 거부하는 strict shape 검사로 unknown 구문·재대입·함수값 escape는 fail closed; 의미를 이해하지 못하는 범용 interprocedural analyzer를 만들지 않는다. 기존 probe static_test.go의 Go parser/file traversal과 execgw/a166_action_cases_test.go의 실제 switch case/typed value 검사를 참고하되 다른 표현을 단순 이름 수로 치환하지 않는다.
3. 현 local TLS·two-process·failure fixture와 직접 연결해 OFF/required missing/closed·wire0·token/401·clone/replay/probe를 선택한다. 구조 AST 결과와 런타임 fixture 증거를 분리해 보고한다. poll writer 통합 후 engine 동일 client/계좌 및 Token/Poll/Other context 흐름까지 별도 재감사한다. 직접 HTTP 외 알려지지 않은 wrapper/dynamic dispatch는 검증자가 명시 분류할 때까지 완료를 거부한다.
4. 검사 함수가 받는 source copy/AST를 시험 내에서 변형하여 실제 파일·브로커 접근 없이 RED를 고정한다: (a) required factory 실패를 official.New/OFF로 바꿈, (b) required Factory.NewClient가 raw New를 반환, (c) token Do를 http.DefaultClient로 바꿈, (d) replay 또는 probe direct POST를 raw client로 바꿈, (e) 다른 Budget이나 unused guarded call만 남기고 실제 전송을 raw로 바꿈, (f) 새 import alias/dynamic Do 또는 새 callsite 미분류. 최소 각 경계의 잘못된 경로가 실패해야 하며 test helper 자체의 판정과 실제 wiring 행동 fixture를 모두 확인한다. 어떠한 변이도 실제 broker 주소로 요청하지 않는다.
5. 이전 a087 `TestImportsStayInsideTheAllowlist`는 accountbudgetruntime이 기존 main.go import인데 누락돼 실제 정확 테스트 RED였다. 별도 41d7eb4ec은 해당 import만 추가해 GREEN; 이 수정은 전체 호출·값 흐름 검사의 증거가 아니다. `TestNoProductionOrderSurfaceIsNamed`의 selector 이름 금지 목록은 concrete client의 새로운 비금지 read를 놓칠 수 있어 2.4 완료 근거로 쓰지 않는다. 개선 시 직접 수신자·호출자·허용 호출 의미를 고정하고 unused reference·selector 별칭 변이로 검출한다.

## Pre-Edit·검증 조건

구현 대상 테스트 파일·재사용 helper를 Manager가 결정한 뒤 영향 함수 source·AST·FLM/BTM 및 RED를 수정 전에 기록한다. 새 검사에는 미분류 항목 거절·OFF/legacy 예외의 실제 조건·nil/error 전파·test-only 제외·명시된 build tag와 우회 변이를 포함한다. Go1.25 정확 선택 normal/tagged race만 실행하며 네트워크 fixture는 fake credentials, exact local TLS DialContext, loopback-only/no-route. 전체 tests/gate·실 브로커·운영·설정·archive 금지. 실제 모든 운영 참여 namespace/UID/inode·사용 중인 이전 바이너리·process와 동적 외부 프로그램은 tasks 4.2 사람 검증으로 남는다. a087 admission nil, engine poll 미완을 그대로 유지한다.

search-first: 저장소 parser/AST tests와 stdlib를 확장하는 최소안; 새로운 dependency·범용 분석기·외부 vendor 조사 없음(사용자 범위 제한). gstack은 구현 후 범위 수동 리뷰. grill-me/grilling의 요구·예외·실패 경계는 기존 승인 사항으로 정리했으며 미결정 구현 scope는 Manager 검토 대상이다. impeccable N/A — 정적 backend/test-only. 현재 시험 미실행(별도 whitelist delta의 정확 정적시험만 실행), 이 계획은 2.4 완료 판정이 아니다.
