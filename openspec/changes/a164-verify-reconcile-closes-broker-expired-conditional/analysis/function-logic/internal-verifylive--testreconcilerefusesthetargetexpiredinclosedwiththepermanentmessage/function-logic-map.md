# Function Logic Map: `TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage`

- Source: `internal/verifylive/reconcile_g1_test.go` (141-149)
- Qualified function: `TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 9bee0396eed204f1…)
- AST evidence: `ast.json` — AST branches 1, 반환 0(줄 없음), 호출 이름 9
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: Q6 시험(시험 전용). a164 편집: 기대 문구를 D8 로 바꾸고 「영구히」 부재를 단언 — 이름은 FLM 연속성을 위해 a121 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t *testing.T` | — | go test | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 146: `if !strings.Contains(r.Error(), want) \|\| strings.Contains(r.Error(), "영구히") {` | 없음 | `t.Fatalf("Q6 refusal must say %q and no longer \"영구히\"; got %q", want, r.Error())` | `TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage` (시험 파일 함수 — coverprofile 계측 밖, 실행 PASS 가 증거) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `rcRefuse` | 거절·기록 불변 단언 | 시험 | AST |

AST 호출 이름 전체: `h.reader.set`, `newRCHarness`, `r.Error`, `rcA063Entries`, `rcCondPage`, `rcCondRow`, `rcRefuse`, `strings.Contains`, `t.Fatalf`.

## State mutations and fallbacks

- 없음.

## Safety conclusion

- 생산 동작 무관. 부재 모드의 Q6 거절 코드·위치는 그대로임을 이 시험이 계속 잼.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 1 에서 편집 뒤 1 로. `source_sha256` bd4fe45a1338… → 9bee0396eed2….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: replace 옛 B1..B1(1) → 새 B1..B1(1).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
