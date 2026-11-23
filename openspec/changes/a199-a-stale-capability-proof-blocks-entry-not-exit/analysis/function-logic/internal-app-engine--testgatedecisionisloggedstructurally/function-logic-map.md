# Function Logic Map: `TestGateDecisionIsLoggedStructurally`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/interlock_test.go` (771-841)
- AST evidence: `ast.json` — branches **11** / returns **0** / calls **44**, source SHA-256 `23dc2525546a616531f0e55d28ff2d05fc352ad351efff108d47fb150c92cf47`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/interlock_test.go --func TestGateDecisionIsLoggedStructurally`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 수락 · 거절 · 게이트 OFF 의 구조 로그 한 줄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하위 「refused」 픽스처 | 만료 attestation(804) | — | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@780) | `if _, err := openProtectedGateEngineLogging(t, dir, srv, matchedGuardian(), &logs); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@785) | `if line["event"] != string(obs.EventOperatingMode) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@788) | `if line["gate_enabled"] != true \|\| line["gate_verified"] != true {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@791) | `if line["max_total_exposure"] != "5000000" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@794) | `if _, present := line["reason"]; present {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@809) | `if _, err := openGateEngineLogging(t, dir, srv, matchedGuardian(), &logs); err == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@814) | `if line["level"] != "WARN" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@817) | `if line["gate_verified"] != false {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@821) | `if !strings.Contains(reason, "expired") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@833) | `if _, err := openGateEngineLogging(t, dir, srv, nil, &logs); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@837) | `if line["gate_enabled"] != false \|\| line["gate_verified"] != false {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 44개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 임시 디렉터리.

## Safety conclusion

- a199 편집(예정): 「refused」 픽스처 교체 + (조건부) 청산 전용 하위 시험.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: no(시험 함수)
