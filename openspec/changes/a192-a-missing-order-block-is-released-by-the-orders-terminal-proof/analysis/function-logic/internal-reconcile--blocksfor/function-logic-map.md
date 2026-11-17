# Function Logic Map: `blocksFor` (reconcile)

- Source: `internal/reconcile/mismatch.go` (1439-1472)
- 편집 전: `../../pre-edit-ast/reconcile--blocksfor.ast.json`(1402~1433, sha f72bcbc7d288, 분기 2).
- AST evidence: `ast.json`(편집 뒤 sha 9895f31184e5, 분기 2, return: return@1471). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/reconcile/mismatch.go --func blocksFor`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 재번호: `renumber.txt`(`../renumber.py` — 옛/새 분기를 (종류, 그 줄의 소스 문장)으로 정렬. 손 재번호 안 함).

## 재번호

```
B1:range@1404 -> B1:range@1441
B2:range@1418 -> B2:range@1455
```

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `diff.MissingOrders[i]` | `LocalOrder` | `Compare` | — |
| 불변식 | 블록 집합·키·Detail 무변경 | design D8 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (range@1441) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B2 (range@1455) | missing-order 마다 — 편집: 리터럴에 `missing` 비공개 필드 1 | 블록 append(키·값·개수 불변) | — | `TestMissingOrderRaisesASymbolBlock`, `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `missing.Identity`(call@1468) | 결속 신원(정규화) | 순수 | AST |

## State mutations and fallbacks

- 없음(값 생성). 새 필드는 장부에 쓰이지 않음.

## Safety conclusion

- Safe edit boundary: 복합 리터럴 필드 1. 분기·호출 순서 불변.
- High-risk impact: yes(블록 생성) — 블록 키·Reason·Release·Detail 그대로.
