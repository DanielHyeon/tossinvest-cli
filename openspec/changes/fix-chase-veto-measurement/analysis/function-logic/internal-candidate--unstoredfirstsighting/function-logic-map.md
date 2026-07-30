# Function Logic Map: `unstoredFirstSighting`

- Source: `internal/candidate/veto.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L764–811, 분기 9개)
- Risk scan: `risk-pattern-report.md`

저장된 최초 순위가 없을 때 그 '없음'이 넷 중 무엇인지 이름 붙이고, 그렇게 말하는 읽기를
함께 보고한다. 이 change가 **한 줄**을 더했다:

```
out.Truncation = truncationOf(o.Reported.RankRequested, o.Reported.RankTotal)
```

거부된 sighting도 절단 사실을 실어야 화면이 "왜 못 쟀는지"와 "무엇을 읽었는지"를 같이
보여줄 수 있다. 이 함수는 **아무것도 결정하지 않는다** — 넷 다 이미 미측정이고, 절단 사실은
표시용으로만 실린다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `observations` | 한 후보의 관측 | `Assess`의 10분 창 | 비면 `NO_OBSERVATIONS` |
| `c.FirstSeenAt` | non-zero(호출자가 확인) | `candidates` | — |
| `o.Reported.Rank`/`RankTotal` | 둘 다 양수여야 '순위 있는 행' | 관측 | 아니면 후보에서 제외 |
| `o.Reported.RankRequested` | 0 또는 양수 | 관측 | 0이면 `Truncation`이 `unknown` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for i, o := range observations` | `earliest`/`candidateRow` 갱신 | — | `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps` |
| B2 | `Rank <= 0 || RankTotal <= 0` | skip | continue | `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps`('nothing ranked') |
| B3 | `earliest < 0 || o.ObservedAt.Before(...)` | `earliest` 갱신 | — | 동상 |
| B4 | `!nearFirstSighting(o.ObservedAt, c.FirstSeenAt)` | 후보 행에서 제외 | continue | `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps`('a ranked row with no stored column') |
| B5 | `candidateRow < 0 || Before(...)` | `candidateRow` 갱신 | — | 동상 |
| B6 | 이유 switch | — | — | `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps` |
| B7 | `len(observations) == 0` | 없음 | `NO_OBSERVATIONS` | `TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps`('nothing recorded at all') |
| B8 | `earliest < 0` | 없음 | `NOT_RANKED` | 동상('nothing ranked') |
| B9 | `candidateRow >= 0` | `row`와 사유를 `NO_FIRST_RANK`로 | — | `TestASessionStartDoesNotStampThePanelAsSeenLate`(첫 tick) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `nearFirstSighting(o.ObservedAt, c.FirstSeenAt)` | ±TTL 창 | 순수 | ast.json calls |
| `truncationOf(o.Reported.RankRequested, o.Reported.RankTotal)` **(신규)** | 보고할 절단 사실 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — `out`(값 복사)에 필드를 채워 돌려준다.
- fallback 없음. 넷 다 미측정이고 어느 것도 통과가 아니다.

## Safety conclusion

- Safe edit boundary: 본문 1줄 가산(표시용 필드). 네 사유의 판정 로직 무변경.
- High-risk impact: no — 이 함수의 모든 경로가 이미 미측정이다. 실린 절단 사실은 화면용이며 어떤 판정에도 쓰이지 않는다.
