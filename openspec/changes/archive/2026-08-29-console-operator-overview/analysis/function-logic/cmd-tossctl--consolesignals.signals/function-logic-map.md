# Function Logic Map: `consoleSignals.Signals`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L860–876, 분기 2개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

한 순간(instant)을 기준으로 계약상의 모든 시장을 평가해 값으로 돌려준다. 저장소는 이 함수 안에서 열리고 이 함수 안에서 닫힌다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 요청 context | internal/console 핸들러 | Open의 마이그레이션·질의가 요청과 함께 취소된다 |
| `s.open` | non-nil | `consoleSignalsSeam` 또는 테스트 fixture | 에러는 문맥을 붙여 반환 |
| `consoleSignalsMarkets` | [KR, US] | design.md 결정된 계약값 | 시장 누락은 화면에서 시장 부재로 읽힌다 — 그래서 고정 목록 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 저장소 열기 실패 | 없음 (release 미등록) | `console.SignalsReading{}, fmt.Errorf(...%w)` | `TestTheSignalsSeamOpensTheStoreUnderTheCallersContext` |
| B2 | `range consoleSignalsMarkets` | `out.Markets` append | 시장 수만큼 항목 | `TestTheSignalsSeamReadsTheStoreAndCallsNoSource` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.open` | 저장소 + release 확보 | release는 defer로 반드시 실행 — WAL 청소를 살린다 | `TestTheSignalsSeamDoesNotHoldTheStoreOpenAcrossReads` |
| `store.Now()` | 모든 시장이 **같은 순간**을 쓰게 한다 | 저장소 시계 — 이 프로세스 시계가 아니다 | L803; 두 시장을 두 순간에 읽으면 같은 질문에 다른 답이 나온다 |
| `consoleSignalsMarket` | 시장 1개 평가 | 실패는 해당 시장의 `Why`로 흡수 | L806 |
| `release` (defer) | 저장소 반납 | 실패해도 반드시 실행 | ast.json defers |

## State mutations and fallbacks

- 저장소에 쓰지 않는다. `Assess`는 읽기이고 promote/cool/prune은 호출되지 않는다.
- 한 시장이 실패해도 나머지 시장은 살아남는다(`consoleSignalsMarket`이 흡수) — 함수 전체 에러는 저장소를 못 연 경우뿐.

## Safety conclusion

- Safe edit boundary: 열기/닫기 쌍과 단일 instant. `defer release()`를 잃으면 원장 파일시스템의 WAL이 회수되지 않는다.
- High-risk impact: no (발굴 읽기 경로). 주문·손절·사이징·Guardian·인증 어디에도 닿지 않는다.
