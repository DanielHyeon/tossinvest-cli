# Function Logic Map: `attestationWatch.prepare`

> **편집 후 지도(post-edit) — 2단계 GREEN 뒤, 2026-11-24, base `603046a23`.** 분기 표는 AST 열거(`../../harness/build_bundles.py --post`). 편집 전 원본 AST 는 `../../pre-edit-ast/internal-app-engine--attestationwatch.prepare.ast.json`.

- Source: `internal/app/engine/attestation_watch.go` (354-407)
- AST evidence: `ast.json` — branches **11** / returns **3** / calls **16**, source SHA-256 `85663d710d3b71b71b99efc3a3e60849154d430713ce898d154af5cf0f0e7510`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func attestationWatch.prepare`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 읽기 · 판정 · 중복 확인 · 문구 — healthy 면 알림 없이 반환(B8@345).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `w.load`·`w.clock`·`w.accountRef` | newAttestationWatch | 판정 입력 | 아래 분기 표 |
| `w.sent`·`w.episode` | 감시 메모리 | 하루 한 번 · 에피소드 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@356) | `if rec := recover(); rec != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@358) | `if w.log != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@373) | `if identity != w.identity \|\| w.sent == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (range@378) | `for k := range w.sent {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (if@379) | `if strings.HasPrefix(k, string(stageHealthy)+":") {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (switch@389) | `switch v.Stage {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (case@390) | `case stageHealthy, stageNotice, stageImminent:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (case@391) | `default:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@394) | `if episode != w.episode {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B10 (if@398) | `if v.Stage == stageHealthy && !w.exitOnly {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B11 (if@403) | `if w.sent[key] {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 16개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- `w.identity`·`w.sent`·`w.episode`·`w.episodeAt`.

## Safety conclusion

- a199 편집(적용됨): 결정 7: 기동 표지가 청산 전용이면 healthy 여도 하루 한 번 알림 — 조건을 파일이 아니라 기동 표지에.
- Safe edit boundary: 편집 전 AST(`../../pre-edit-ast/`)와 비교해 위 편집 칸 밖의 분기는 무편집. 행동 고정은 a199 시험 · 변이 원장(`../../mutation/`).
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
