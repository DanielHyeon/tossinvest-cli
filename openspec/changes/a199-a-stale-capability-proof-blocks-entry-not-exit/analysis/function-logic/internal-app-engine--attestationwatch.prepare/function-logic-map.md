# Function Logic Map: `attestationWatch.prepare`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/attestation_watch.go` (309-353)
- AST evidence: `ast.json` — branches **9** / returns **3** / calls **14**, source SHA-256 `9b71a387860c809bc6ee8f80528cd5cc1446fdd8dd080eac14164d7548eae1f6`(파일 단위).
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
| B1 (if@311) | `if rec := recover(); rec != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@313) | `if w.log != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@328) | `if identity != w.identity \|\| w.sent == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B4 (switch@336) | `switch v.Stage {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B5 (case@337) | `case stageHealthy, stageNotice, stageImminent:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B6 (case@338) | `default:` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B7 (if@341) | `if episode != w.episode {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B8 (if@345) | `if v.Stage == stageHealthy {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B9 (if@349) | `if w.sent[key] {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 14개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- `w.identity`·`w.sent`·`w.episode`·`w.episodeAt`.

## Safety conclusion

- a199 편집(예정): 결정 7: 기동 표지가 청산 전용이면 healthy 여도 하루 한 번 알림 — 조건을 파일이 아니라 기동 표지에.
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
