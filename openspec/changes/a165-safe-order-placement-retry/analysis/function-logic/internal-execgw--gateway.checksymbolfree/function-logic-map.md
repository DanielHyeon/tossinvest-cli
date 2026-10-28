# Function Logic Map: `Gateway.checkSymbolFree`

- Source: `internal/execgw/gateway.go` (802-838)
- Qualified function: `Gateway.checkSymbolFree`
- Revision: `current` (a165 base `34e12967e`, `source_sha256` c304e32a62e42759…; GREEN 뒤 AST 재추출 2026-10-29)
- AST evidence: `ast.json` — AST branches 10, 반환 9, 호출 10(`go run ./tools/logic-map`, 2026-10-29). 1.3의 기존 분기 7개 뒤에 latch 읽기/거절 분기 3개가 추가됨.
- Risk scan: `risk-pattern-report.md`
- 실행 칸 측정: 2.1 Pre-Edit에서 기준선을 기록했고, A165 GREEN은 B3·B8~B10의 경로를 실행함. 기존 오류 주입 경로는 이번 범위에서 미측정으로 남김.

## 역할 (a165)

같은 종목의 변경 가능 여부를 원장으로 판정한다. 앞 절반(B1·B2)은 **모든 변경**(취소 포함)을 막는 미종결 판정이고 판정 함수는 exit 청소와 같은 `unsettledFor` 다(a094 D−4.7).
B3(814) `!plan.raisesExposure` 에서 취소·청산 매도·노출을 줄이는 정정은 빠져 나간다. 뒤 절반(B4~B10)은 **노출을 늘리는 변경(진입)만** 막는 판정이다.

a165 는 뒤 절반의 끝(B5 range 가 끝난 뒤, 마지막 `return nil, nil`(837) 앞)에 **진입 전용 latch**(design D11) 호출을 더함:
`g.cancelRefusalLatch(ctx, plan)`(새 함수, 새 파일). 앞 절반·B3·`unsettledFor` 는 바꾸지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan.market·plan.symbol | 정규화된 시장·종목 | Gateway.place·Cancel·Amend | — |
| plan.raisesExposure | place 매수(377) · amend 증량(447) 만 참, 취소는 거짓(416) | mutationPlan 구성(손으로 읽음) | B3 에서 거짓이면 뒤 절반 미실행 |
| journal | 미종결·미해소 attempt | `PendingAttempts`·`UnresolvedAttempts` | 읽기 오류는 `error` 로 반환(submit B8 → 전송 없음) |

## Branches and early returns

AST 열거 순서 그대로.

Exact AST return positions: 806:3, 810:3, 815:3, 819:3, 824:4, 827:4, 833:3, 835:3, 837:2

| Branch | Coordinate | Condition | Mutation/side effect | Return/error | Executed (baseline) |
|---|---|---|---|---|---|
| B1 | `if at 805:2` | if err != nil {  (unsettledFor 오류) | 없음 | `nil, err` (806) | 미측정 |
| B2 | `if at 808:2` | if len(unsettled) > 0 { | 없음 | `ReasonSymbolInFlight` 거절 (810) — 모든 변경 | 미측정 |
| B3 | `if at 814:2` | if !plan.raisesExposure { | 없음 | `nil, nil` (815) — **취소·청산·노출 감소는 여기서 끝남. a165 latch 에 닿지 않음** | A165 GREEN |
| B4 | `if at 818:2` | if err != nil {  (UnresolvedAttempts 오류) | 없음 | `nil, err` (819) | 미측정 |
| B5 | `range at 821:2` | for _, rec := range unresolved { | 없음 | 반복 | 미측정 |
| B6 | `if at 823:3` | if err != nil {  (attemptTargets 오류) | 없음 | `nil, err` (824) | 미측정 |
| B7 | `if at 826:3` | if same { | 없음 | `ReasonUnresolvedInDoubt` 거절 (827) — 진입만 | 미측정 |
| B8 | `if at 832:2` | if rejected, err := g.cancelRefusalLatch(ctx, plan); err != nil { | 읽기 오류 | `nil, err` (833), fail-closed | 오류 주입 seam 없음 |
| B9 | `else at 834:9` | latch 오류가 아닌 갈래 | 없음 | B10으로 진행 | A165 GREEN |
| B10 | `if at 834:9` | rejected != nil | latch 거절 | `rejected, nil` (835) | A165 GREEN |

## Calls and live bindings

`g.unsettledFor`(B1·B2 의 입력, exit 청소와 flatten 과 공유) · `g.journal.UnresolvedAttempts`(B4) · `g.attemptTargets`(B6) · `g.cancelRefusalLatch`(B8) · `reject`.

AST 호출 이름 전체: `fmt.Errorf`, `g.attemptTargets`, `g.cancelRefusalLatch`, `g.journal.UnresolvedAttempts`, `g.unsettledFor`, `len`, `reject`.

## State mutations and fallbacks

- 없음(읽기 전용 판정). 오류는 호출자(submit B8, gateway.go:525)가 전송 없이 거절로 끝낸다(fail-closed).

## a165 편집 계획 (design D11)

- 삽입 위치: B5 range 의 닫는 괄호 뒤, `return nil, nil`(837) 앞. 즉 B3 이 거짓을 돌려준 뒤에만 닿는다 — **AST 상 취소·청산 매도는 B3(815)에서 반환하므로 latch 를 읽지 않는다.**
- 삽입 내용: 오류와 거절을 분리한 `if/else if` 두 단계. 거절 사유는 기존 어휘 `ReasonSymbolInFlight`(a165 이전에 같은 상황 — 취소 IN_DOUBT — 이 진입에 준 사유와 같음, B2(810)).
- 위치를 B7 뒤(가장 뒤)로 두는 이유: 앞에 두면 B4~B7 의 기존 거절 시험이 latch 에 먼저 걸려 UNRESOLVED_IN_DOUBT 가드를 더는 못 박는다(「새 가드는 뒤의 가드를 못에서 뽑는다」).
- 새 분기 3(B8~B10)가 생김. GREEN 뒤 이 번들의 `ast.json` 을 다시 뽑아 좌표와 수를 확인함.

## Safety conclusion

- latch 는 B3 뒤에만 있으므로 exit·flatten·reconcile·보호 경로(전부 `raisesExposure=false`)는 이 함수에서 latch 를 보지 않는다. `unsettledFor`(exit·flatten 과 공유되는 판정)는 바꾸지 않는다(1.3 2차 리뷰 P1). High-risk: yes.
