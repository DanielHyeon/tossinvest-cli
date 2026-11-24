# Function Logic Map: `TestA200CurrentBehaviourPartialStopFillsAfterAConvergenceClose` (지운 base 판)

- Source: `internal/journal/a200_convergence_close_exit_fill_repro_test.go` (301-333), base 28680af3d, source sha256 `f123aa623dde…`.
- AST evidence: `ast.json`(revision: base, 분기 6).
- Risk scan: `risk-pattern-report.md`.
- 처분: 조정 뒤 부분 체결 둘의 편집 전 동작(X58 행 1 · 둘째 무기록) — 새 계약으로 뒤집혀 지움. 뒤를 잇는 시험 `TestA200PartialStopFillsAfterAConvergenceCloseAreEachRecognised`(internal/journal/a200_recognition_test.go 또는 같은 파일).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 임시 원장 · 운영 훅 리터럴 · 합성 픽스처 | `t.TempDir()` | `a200Journal` | 단언 실패면 시험 실패 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 if@310 | 단언 조건(원장 관측값) | 없음 | 참이면 t.Fatalf / t.Logf | 시험 자신 |
| B2 if@315 | 단언 조건(원장 관측값) | 없음 | 참이면 t.Fatalf / t.Logf | 시험 자신 |
| B3 if@319 | 단언 조건(원장 관측값) | 없음 | 참이면 t.Fatalf / t.Logf | 시험 자신 |
| B4 if@324 | 단언 조건(원장 관측값) | 없음 | 참이면 t.Fatalf / t.Logf | 시험 자신 |
| B5 if@327 | 단언 조건(원장 관측값) | 없음 | 참이면 t.Fatalf / t.Logf | 시험 자신 |
| B6 if@330 | 단언 조건(원장 관측값) | 없음 | 참이면 t.Fatalf / t.Logf | 시험 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 1단계 하네스(a200Journal · a200ArmStop · a200Converge · readA200Ledger) | 원장 모양 만들기 · 읽기 | 오류면 t.Fatal | AST |

## State mutations and fallbacks

- 시험 함수 — 생산 상태 변경 없음(임시 원장만).

## Safety conclusion

- Safe edit boundary: 시험 전용. 생산 코드 영향 없음.
- High-risk impact: no(시험 함수 — 다만 지운 기대는 새 계약 시험이 이어받음).
