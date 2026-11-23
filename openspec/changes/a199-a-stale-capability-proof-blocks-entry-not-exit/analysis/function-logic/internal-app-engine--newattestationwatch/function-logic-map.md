# Function Logic Map: `newAttestationWatch`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/attestation_watch.go` (219-237)
- AST evidence: `ast.json` — branches **3** / returns **2** / calls **1**, source SHA-256 `9b71a387860c809bc6ee8f80528cd5cc1446fdd8dd080eac14164d7548eae1f6`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/attestation_watch.go --func newAttestationWatch`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 엔진 Context 에서 감시를 만듦(게이트 · 파일 · 계좌 · 알림기).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.Automation` | NewContext(기동 시점 값) | 227-228 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@221) | `if c == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@224) | `if w.clock == nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@233) | `if c.Notifier != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 1개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- 감시 구조체 채움.

## Safety conclusion

- a199 편집(예정): 기동 때의 청산 전용 표지를 감시에 옮김 — 「지금 청산 전용으로 돌고 있다」의 출처(D6 로 프로세스 수명 동안 불변).
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
