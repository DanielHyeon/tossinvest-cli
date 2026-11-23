# Function Logic Map: `judgeAttestation`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/attestation_watch.go` (100-144)
- AST evidence: `ast.json` — branches **13** / returns **4** / calls **11**, source SHA-256 `9b71a387860c809bc6ee8f80528cd5cc1446fdd8dd080eac14164d7548eae1f6`(파일 단위).
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
| B1 (if@101) | `if loadErr != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@106) | `if err := att.Verify(now, accountRef, required); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (switch@107) | `switch {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (case@108) | `case errors.Is(err, attest.ErrExpired):` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (case@110) | `case errors.Is(err, attest.ErrEndpointNotAttested):` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (case@113) | `case errors.Is(err, attest.ErrAccountMismatch):` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (case@115) | `case errors.Is(err, attest.ErrIncomplete):` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (case@117) | `default:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@127) | `if remaining <= 0 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (switch@135) | `switch {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (case@136) | `case v.DaysLeft > attestationNoticeDays:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B12 (case@138) | `case v.DaysLeft > attestationCriticalDays:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B13 (case@140) | `default:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 11개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음(순수).

## Safety conclusion

- a199 편집(예정): 어느 결함이 청산 전용 / 거부인지를 인터록과 같은 함수로(D7) — 지금은 만료 + 다른 계좌를 expired 로 판정(B4@108), 인터록(D1a)은 거부.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
