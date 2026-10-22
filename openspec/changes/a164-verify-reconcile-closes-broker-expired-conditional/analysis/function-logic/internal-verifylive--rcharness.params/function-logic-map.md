# Function Logic Map: `rcHarness.params`

- Source: `internal/verifylive/reconcile_fixture_test.go` (370-389)
- Qualified function: `rcHarness.params`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` bb3af371cc0bf0cd…)
- AST evidence: `ast.json` — AST branches 1, 반환 2(줄 371, 386), 호출 이름 5
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 시험 하네스(시험 전용). a164 편집: `Expired: h.expired` 한 줄 — 만료 모드 시험이 같은 하네스로 Reconcile 을 몰게 함. 비례 원칙상 무거운 규율 대상 아님 — 번들은 게이트가 요구해서 둠.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 수신자 `h *rcHarness` | 시험 하네스 | `reconcile_fixture_test.go` | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 383: `if h.onApprove != nil {` | 없음 | `h.onApprove()` | `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` · `TestReconcileAppendsOneReconciledAbsentLineWhenEveryConditionHolds` (시험 파일 함수 — coverprofile 계측 밖, 실행 PASS 가 증거) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (함수 리터럴 Approve) | 승인 기록 | 시험 seam | AST |

AST 호출 이름 전체: `append`, `h.onApprove`, `h.out.String`, `h.reader.listCalls`, `len`.

## State mutations and fallbacks

- `Approve` 리터럴이 하네스의 승인·화면·목록 호출 수를 기록함(기존).

## Safety conclusion

- 생산 동작 무관(시험 파일). `expired` 거짓이면 a121 하네스와 같은 Params.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 1 에서 편집 뒤 1 로. `source_sha256` 1364538003c2… → bb3af371cc0b….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B1 = 새 B1..B1.
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
