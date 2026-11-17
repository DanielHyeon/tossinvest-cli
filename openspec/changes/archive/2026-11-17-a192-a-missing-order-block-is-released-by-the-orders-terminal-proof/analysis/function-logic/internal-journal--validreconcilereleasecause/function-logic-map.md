# Function Logic Map: `ValidReconcileReleaseCause` (journal)

- Source: `internal/journal/reconcile_states.go` (111-119)
- 편집 전: `../../pre-edit-ast/journal--validreconcilereleasecause.ast.json`(111~119, sha 0943dee43984, 분기 3).
- AST evidence: `ast.json`(편집 뒤 sha 1591e0153868, 분기 3, return: return@115 return@117). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/journal/reconcile_states.go --func ValidReconcileReleaseCause`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 재번호: `renumber.txt`(`../renumber.py` — 옛/새 분기를 (종류, 그 줄의 소스 문장)으로 정렬. 손 재번호 안 함).

## 재번호

```
B1:switch@112 -> B1:switch@112
B2:case@113 -> B2:case@113
B3:case@116 -> B3:case@116
```

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cause` | 문자열 | 호출자 | 목록 밖 → false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch@112) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B2 (case@113) | 이 판본이 쓰는 원인 — 편집: case 목록에 `ORDER_TERMINAL_CONFIRMED` 추가 | — | true | `TestUnknownReleaseCauseIsRefused`, `TestA192OrderTerminalReleaseCauseIsOneThisBuildWrites` |
| B3 (case@116) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | AST calls 0 |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: case 값 1. 다른 원인 판정 불변(`TestUnknownReleaseCauseIsRefused`).
- High-risk impact: yes(원장 쓰기 제약) — 새 원인 하나만 허용 목록에 더함.
