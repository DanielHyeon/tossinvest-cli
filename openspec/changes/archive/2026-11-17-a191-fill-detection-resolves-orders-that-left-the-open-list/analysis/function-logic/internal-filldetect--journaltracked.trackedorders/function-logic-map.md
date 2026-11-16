# Function Logic Map: `JournalTracked.TrackedOrders` (internal/filldetect)

- Source: `internal/filldetect/ledger.go` 125~150 (편집 뒤). 편집 전: `../../pre-edit-ast/filldetect--journaltracked.trackedorders.ast.json`(125~149, sha 75966bf1507d).
- AST evidence: `ast.json`(편집 뒤 sha b5bc0e3023e9). 분기 B1(if@126)·B2(if@129)·B3(if@133)·B4(range@137) — 편집 전과 같음. return 4(마지막 return 148→149).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t.Journal` | 비nil | 운영 배선 | nil → 오류(B1) |
| `t.AccountRef` | 비어 있지 않음 | 운영 배선 | 빈 값 → 오류(B2) |
| `row.IntentID` | 원장의 intent id | `journal.TrackedFillOrder`(fills.go:297) | 그대로 옮김 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@126) | 저널 없음 | — | return@127 | 기존(ledger_test) |
| B2 (if@129) | 계좌 없음 | — | return@130 | 기존 |
| B3 (if@133) | 원장 조회 오류 | — | return@134 | 기존 |
| B4 (range@137) | 행마다 `TrackedOrder` 만듦 — **편집: `IntentID: row.IntentID` 필드 1줄** | out append | — | `TestJournalTrackedCarriesTheIntentID`, `TestJournalTrackedPreservesCanonicalOrderScope` |
| return@149 | 끝 | — | 목록 | — |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Journal.TrackedFillOrders`(@132) | 추적 원장 | 오류 그대로 | AST |

## State mutations and fallbacks

- 없음(값 복사). 편집은 복합 리터럴 필드 1개 — 분기·호출 변화 0.

## Safety conclusion

- 경보 식별자를 옮길 뿐 추적 집합의 구성(어떤 주문이 들어오나)은 바꾸지 않음.
