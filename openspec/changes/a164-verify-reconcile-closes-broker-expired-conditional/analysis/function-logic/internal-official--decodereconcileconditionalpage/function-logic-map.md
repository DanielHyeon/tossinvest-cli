# Function Logic Map: `DecodeReconcileConditionalPage`

- Source: `internal/official/reconcile_reads.go` (134-151)
- Qualified function: `DecodeReconcileConditionalPage`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 3fca84976dedd94e…)
- AST evidence: `ast.json` — AST branches 3, 반환 3(줄 137, 143, 150), 호출 이름 6
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 목록 행 엄격 해독. a164 는 행에 `FirstStatus` 한 필드 대입을 더함(P2-3) — 부재 모드는 그 필드를 읽지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `result []byte` | `result` 원문 | `getAcct` | 스키마 위반 = `ErrReconcileSchema` |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 136: `if err != nil {` | 없음 | `return ReconcileConditionalPage{}, err` | `TestDecodeReconcileConditionalPageRefusesSchemaHoles` · `TestDecodeReconcilePageNamesTheGuardThatRefused` · `TestReconcileClientReadsRefuseANullResult` |
| B2 | range 줄 140: `for i, raw := range env.rows {` | 없음 | `var o apiConditionalOrder` | `TestDecodeReconcileConditionalPageKeepsTheSecondLegAndTheTrigger` |
| B3 | if 줄 142: `if err := json.Unmarshal(raw, &o); err != nil {` | 없음 | `return ReconcileConditionalPage{}, reconcileSchema("conditionalOrders[%d] does not decode: %v", i, err)` | 미덮임(시험별 coverage 측정 0) — 행 객체가 `apiConditionalOrder` 로 안 풀리는 경우(예: 필드 형 불일치) — 기존 공백 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decodeReconcileEnvelope` | 봉투 검증 | 오류 반환 | AST |
| `json.Unmarshal` | 행 해독 | 오류 = 스키마 거절 | AST |

AST 호출 이름 전체: `append`, `decodeReconcileEnvelope`, `json.Unmarshal`, `len`, `make`, `reconcileSchema`.

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- Safe edit boundary: 대입 한 줄에 `FirstStatus: o.First.Status` 만.
- High-risk impact: 인접 — 만료 모드 A4 의 CLOSED 행 leg 검사 입력.
