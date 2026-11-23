# Function Logic Map: `attestationWatch.eventFor`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/attestation_watch.go` (383-486)
- AST evidence: `ast.json` — branches **22** / returns **1** / calls **29**, source SHA-256 `9b71a387860c809bc6ee8f80528cd5cc1446fdd8dd080eac14164d7548eae1f6`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func attestationWatch.eventFor`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 판정을 알림 하나로 — 제목 · 본문 한국어 고정 문구, 계좌 · 경로 · 원문 오류 없음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v` | judgeAttestation + prepare | 단계 | 아래 분기 표 |
| `w.critical` | notifications.enabled | B1@385 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@385) | `if v.critical() && w.critical {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@390) | `if !v.ExpiresAt.IsZero() {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (switch@407) | `switch v.Stage {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (case@408) | `case stageNotice, stageImminent:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@409) | `if v.DeadlineKind == deadlineProof {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (else@413) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (case@418) | `case stageProofOverdue:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (case@422) | `case stageExpired:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (case@425) | `case stageScopeLost:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (range@429) | `for _, e := range v.Missing {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@430) | `if mutationEndpoint(e) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B12 (if@434) | `if mutations == len(v.Missing) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B13 (else@438) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B14 (if@441) | `if mutations > 0 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B15 (case@446) | `case stageIncomplete:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B16 (case@449) | `case stageAccountMismatch:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B17 (case@453) | `default:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B18 (if@461) | `if !v.ExpiresAt.IsZero() {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B19 (if@464) | `if v.Stage == stageNotice \|\| v.Stage == stageImminent {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B20 (if@467) | `if len(v.Missing) > 0 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B21 (if@470) | `if !v.ProofAt.IsZero() {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B22 (if@473) | `if !v.Deadline.IsZero() {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 29개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음(문구 조립).

## Safety conclusion

- a199 편집(예정): expired · scope_lost 의 귀결 문장(「청산 전용으로 뜬다」), 청산 전용으로 돌고 있으면 현재 사실(D7). 날수 · 지남 단계 문장은 Q.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
