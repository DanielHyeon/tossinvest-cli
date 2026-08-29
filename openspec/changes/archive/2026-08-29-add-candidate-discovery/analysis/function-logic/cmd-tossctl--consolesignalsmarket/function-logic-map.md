# Function Logic Map: `consoleSignalsMarket`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L884–916, 분기 1개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `583772c4` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

시장 하나를 평가한다. **읽지 못한 시장도 목록에서 빠지지 않고 `Why`와 함께 남는다** — 화면에서 사라진 시장은 아무것도 없는 시장과 구별되지 않고, 그 혼동을 없애는 것이 이 화면의 존재 이유다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `store` | non-nil | `Signals` | — |
| `at` | `store.Now()`의 단일 instant | 저장소 시계 | 시장마다 다른 시각을 쓰면 나이 계산이 서로 어긋난다 |
| `Thresholds.NearHighDistancePct` | 저장소에 승인된 유일한 임계 | `candidate.DefaultNearHighThresholdPct` | seen_late·extended는 임계가 없어 `THRESHOLD_ABSENT`(미측정)로 돌아온다 — 통과가 아니다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `candidate.Assess` 실패 | `out.Why = err.Error()` | verdict 없는 시장 항목 | `TestTheSignalsSeamSaysWhereTheMissingSourceNamesAreRatherThanClaimingNone` 계열 + 화면의 미측정 렌더 |
| (else) | 평가 성공 | `out.Verdicts`, `out.Vetoes/Crossings/Bands` | 완전한 시장 항목 | `TestTheSignalsSeamTalliesThroughTheSamePathAsAScan` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `candidate.Assess` | 저장된 관측만으로 veto 판정 | 원천 호출 없음 — 계좌 rate 예산 0 | `TestTheSignalsSeamReadsTheStoreAndCallsNoSource` |
| `candidate.TallyVerdicts` | 스캔 출력과 **같은 reducer** | 손으로 세는 복제본 금지 | `TestTheConsoleDoesNotBuildTheDiscoveryTalliesItself`, `TestTheSignalsSeamTalliesThroughTheSamePathAsAScan` (§5 리뷰 P2 ④) |

## State mutations and fallbacks

- 저장소에 쓰지 않는다.
- `Panel.Why`는 상수 문장이다: 콘솔 프로세스는 스캔을 돌지 않으므로 빠진 원천의 **이름**을 가질 방법이 없고, 지어내는 것보다 없다고 말하는 편이 낫다(tasks 5.7, issues.md 9).

## Safety conclusion

- Safe edit boundary: 임계 공급과 tally 경로. 임계를 지어넣는 것은 D6가 금지하는 편집이고, 화면은 그것을 측정값으로 렌더한다.
- High-risk impact: no (발굴 읽기 경로). 주문 경로 무접촉.
