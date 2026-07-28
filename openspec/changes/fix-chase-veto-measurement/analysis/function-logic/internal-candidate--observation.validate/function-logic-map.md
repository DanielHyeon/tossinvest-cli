# Function Logic Map: `Observation.validate`

- Source: `internal/candidate/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L269–310, 분기 9개)
- Risk scan: `risk-pattern-report.md`

관측 하나가 나중에 무엇과도 조인될 수 없는 행인지를 경계에서 거른다. 이 change가 **case를
하나** 더했다(B7): `o.Reported.RankRequested < 0`.

절대값이 아니라 부호가 문제다. 부재는 여기서 0이고, 음수는 더 작은 부재가 아니라 **아무것도
만들 수 없는 수**다. 그리고 `positive()`가 저장 시 비양수를 NULL로 접고 `truncationOf`가
`requested <= 0`을 `unknown`으로 접으므로, 음수가 통과하면 그 행은 "아무도 재지 않았다"로
**위장한 채** 저장된다. 음수 rank를 거부하는 것과 같은 이유로 같은 자리에서 거부한다.

새 case의 **위치**도 우연이 아니다. 음수 rank 검사(B6) 바로 뒤, rank/total 정합성 검사(B8·B9)
앞이다 — 세 검사가 전부 "이 숫자들이 존재할 수 있는가"이고 나머지는 "이 숫자들이 서로
맞는가"다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.Market`/`o.Symbol` | trim 후 비어 있지 않아야 | 소스 | 빈 값은 오류 |
| `o.Source` | trim 후 비어 있지 않아야 | 소스 | 빈 값은 오류 |
| `o.ObservedAt` | non-zero | 스캔 | zero는 오류 — 시간축이 이 패키지의 기능이다 |
| `o.Reported.Rank`/`RankTotal` | 0 이상, `Rank <= RankTotal`, `Rank>0`이면 `RankTotal>0` | 소스 | 위반은 오류 |
| `o.Reported.RankRequested` | **0 이상** | 소스 어댑터 | 음수는 오류(신규) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 검증 switch | `out`의 market/symbol만 정규화 | — | `TestARankWithoutItsListLengthIsRefused` |
| B2 | `out.Market == ""` | 없음 | 오류 | **커버 없음** |
| B3 | `out.Symbol == ""` | 없음 | 오류 | **커버 없음** |
| B4 | source가 공백뿐 | 없음 | 오류 | **커버 없음** |
| B5 | `o.ObservedAt.IsZero()` | 없음 | 오류 | **커버 없음** |
| B6 | `Rank < 0 || RankTotal < 0` | 없음 | 오류 | **커버 없음** |
| B7 | `RankRequested < 0` **(신규)** | 없음 | 요청 수를 실은 오류 | `TestANegativeRequestedCountIsRefusedByTheObservationBoundary` |
| B8 | `Rank > 0 && RankTotal == 0` | 없음 | 오류 — rank/0은 +Inf라 모든 임계를 통과한다 | `TestARankWithoutItsListLengthIsRefused` |
| B9 | `RankTotal > 0 && Rank > RankTotal` | 없음 | 오류 | **커버 없음** |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.ToUpper` / `strings.TrimSpace` | market·symbol 정규화 | 순수 | ast.json calls |
| `o.ObservedAt.UTC()` | instant 정규화 | 순수 | ast.json calls |
| `fmt.Errorf` | 각 거부 | — | ast.json calls |

## State mutations and fallbacks

- 값 복사 `out`만 바꾼다 — 수신자는 값이고 호출자의 원본은 그대로다.
- 정규화는 market·symbol·instant 셋뿐이다. **값 필드는 손대지 않는다** — 소스가 어떤 수치를 비워 두는 것은 이 패키지가 기록하는 일상 상태다.
- fallback 없음. 거부는 zero `Observation`과 오류다.

## Safety conclusion

- Safe edit boundary: switch case 1개 가산. 기존 8개 case의 조건·순서·문구 무변경.
- High-risk impact: no (경계 검증). 방향은 **더 거부하는 쪽**이며, 이 case가 없으면 음수가 NULL로 접혀 미측정으로 위장한다.
