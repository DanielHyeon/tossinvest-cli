# Function Logic Map: `TestGateDecisionIsLoggedStructurally`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--testgatedecisionisloggedstructurally.ast.json`.

- Source: `internal/app/engine/interlock_test.go` (749-819)
- AST evidence: `ast.json` — branches **11** / returns **0** / calls **44**, source SHA-256 `d2c0c7aec58bfbb4138e89e776ec80beba0738d2e781dbcdcdc7e23ba2ee0fdb`(파일 단위).
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
| B1 (if@758) | `if _, err := openProtectedGateEngineLogging(t, dir, srv, matchedGuardian(), &logs); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@763) | `if line["event"] != string(obs.EventOperatingMode) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@766) | `if line["gate_enabled"] != true \|\| line["gate_verified"] != true {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@769) | `if line["max_total_exposure"] != "5000000" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@772) | `if _, present := line["reason"]; present {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (if@787) | `if _, err := openGateEngineLogging(t, dir, srv, matchedGuardian(), &logs); err == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@792) | `if line["level"] != "WARN" {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@795) | `if line["gate_verified"] != false {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@799) | `if !strings.Contains(reason, "expired") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@811) | `if _, err := openGateEngineLogging(t, dir, srv, nil, &logs); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@815) | `if line["gate_enabled"] != false \|\| line["gate_verified"] != false {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 44개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 임시 디렉터리.

## Safety conclusion

- a199 편집(적용됨): 「refused」 픽스처 교체 + (조건부) 청산 전용 하위 시험.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: no(시험 함수)
