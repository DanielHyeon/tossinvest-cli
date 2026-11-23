# Function Logic Map: `judgeAttestation`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--judgeattestation.ast.json`.

- Source: `internal/app/engine/attestation_watch.go` (108-173)
- AST evidence: `ast.json` — branches **17** / returns **5** / calls **12**, source SHA-256 `85663d710d3b71b71b99efc3a3e60849154d430713ce898d154af5cf0f0e7510`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func judgeAttestation`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 감시의 판정 — Load 오류 → unreadable, Verify 첫 실패 → 단계, 통과면 기한까지 날수.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `att`·`loadErr` | attest.Load | B1@101 | 아래 분기 표 |
| `accountRef` | 엔진 자격 증명의 계좌 | Verify | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@112) | `if len(proofs) > 0 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@115) | `if loadErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@123) | `if errors.As(err, &unproven) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (if@126) | `if stale != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@127) | `if stale.expired {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (else@129) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@135) | `if err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (switch@136) | `switch {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (case@137) | `case errors.Is(err, attest.ErrAccountMismatch):` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (case@139) | `case errors.Is(err, attest.ErrIncomplete):` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (case@141) | `default:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B12 (if@151) | `if remaining <= 0 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B13 (switch@160) | `switch {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B14 (case@161) | `case v.DaysLeft > attestationNoticeDays:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B15 (case@163) | `case v.DaysLeft > attestationCriticalDays:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B16 (case@165) | `default:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B17 (if@168) | `if v.Stage != stageHealthy {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 12개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음(순수).

## Safety conclusion

- a199 편집(적용됨): 어느 결함이 청산 전용 / 거부인지를 인터록과 같은 함수로(D7) — 지금은 만료 + 다른 계좌를 expired 로 판정(B4@108), 인터록(D1a)은 거부.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
