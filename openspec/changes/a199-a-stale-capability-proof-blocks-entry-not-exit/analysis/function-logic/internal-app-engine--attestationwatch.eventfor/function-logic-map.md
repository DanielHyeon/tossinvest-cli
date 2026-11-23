# Function Logic Map: `attestationWatch.eventFor`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--attestationwatch.eventfor.ast.json`.

- Source: `internal/app/engine/attestation_watch.go` (402-524)
- AST evidence: `ast.json` — branches **27** / returns **1** / calls **29**, source SHA-256 `be8b45222d70996e0008e4fc2fe04ba983dbdd2132342a7f6c99e4577836bab2`(파일 단위).
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
| B1 (if@404) | `if v.critical() && w.critical {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@409) | `if !v.ExpiresAt.IsZero() {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (switch@429) | `switch v.Stage {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (case@430) | `case stageNotice, stageImminent:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@431) | `if v.DeadlineKind == deadlineProof {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (else@435) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (case@440) | `case stageProofOverdue:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (case@444) | `case stageExpired:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (case@447) | `case stageScopeLost:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (range@451) | `for _, e := range v.Missing {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@452) | `if mutationEndpoint(e) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B12 (if@456) | `if mutations == len(v.Missing) {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B13 (else@460) | `} else {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B14 (if@463) | `if mutations > 0 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B15 (case@468) | `case stageIncomplete:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B16 (case@471) | `case stageAccountMismatch:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B17 (case@475) | `case stageHealthy:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B18 (case@481) | `default:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B19 (if@487) | `if w.exitOnly && v.Stage != stageHealthy {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B20 (switch@488) | `switch v.Stage {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B21 (case@489) | `case stageNotice, stageImminent, stageProofOverdue:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B22 (case@491) | `default:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B23 (if@499) | `if !v.ExpiresAt.IsZero() {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B24 (if@502) | `if v.Stage == stageNotice \|\| v.Stage == stageImminent {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B25 (if@505) | `if len(v.Missing) > 0 {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B26 (if@508) | `if !v.ProofAt.IsZero() {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B27 (if@511) | `if !v.Deadline.IsZero() {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 29개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 없음(문구 조립).

## Safety conclusion

- a199 편집(적용됨): expired · scope_lost 의 귀결 문장(「청산 전용으로 뜬다」), 청산 전용으로 돌고 있으면 현재 사실(D7). 날수 · 지남 단계 문장은 Q.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
