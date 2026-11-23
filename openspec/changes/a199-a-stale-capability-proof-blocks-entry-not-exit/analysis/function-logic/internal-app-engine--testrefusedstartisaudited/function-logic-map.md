# Function Logic Map: `TestRefusedStartIsAudited`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/interlock_test.go` (649-668)
- AST evidence: `ast.json` — branches **3** / returns **0** / calls **16**, source SHA-256 `23dc2525546a616531f0e55d28ff2d05fc352ad351efff108d47fb150c92cf47`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/interlock_test.go --func TestRefusedStartIsAudited`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 거절된 기동의 감사 1줄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 픽스처 | 만료 attestation(653) | — | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@656) | `if _, err := openGateEngine(t, dir, srv, matchedGuardian()); err == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@662) | `if refused == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@665) | `if !strings.Contains(refused.Detail, "expired") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 16개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 임시 디렉터리.

## Safety conclusion

- a199 편집(예정): 픽스처를 거절이 유지되는 결함으로 교체(만료는 청산 전용이 됨).
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: no(시험 함수)
