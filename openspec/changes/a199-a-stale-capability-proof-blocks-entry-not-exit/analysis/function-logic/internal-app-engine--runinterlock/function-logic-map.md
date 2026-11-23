# Function Logic Map: `runInterlock`

> **편집 전 지도(pre-edit) — 1단계, 2026-11-24, HEAD `e4165fa3d`(코드 = `9738a0c32`).** 분기 표는 AST 열거(`../../harness/build_bundles.py`), 2단계 GREEN 뒤 다시 만든다.

- Source: `internal/app/engine/interlock.go` (390-421)
- AST evidence: `ast.json` — branches **3** / returns **4** / calls **8**, source SHA-256 `d018e5cf62544eef99ad68fadebdbe0de63f4677d0bea51d39a33ccc18f037e1`(파일 단위).
  명령 `go run ./tools/logic-map --file internal/app/engine/interlock.go --func runInterlock`.
- Risk scan: `risk-pattern-report.md`.
- 의도: 게이트 OFF 면 즉시 반환, ON 이면 verifyGate → 거절(감사 · 로그) 또는 수락(Verified · EntryPermitted · 감사 · 로그).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `gate.Enabled` | 설정 | config | 아래 분기 표 |
| `log`·`logger` | NewContext | 감사 · 구조 로그 | 아래 분기 표 |
| 불변식 | 수락 감사 기록 실패면 기동 거절(B3) | §0.5 | 아래 분기 표 |

## Branches and early returns

| Branch | Condition (소스 그 줄) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@393) | `if !gate.Enabled {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B2 (if@400) | `if err := verifyGate(&status, facts); err != nil {` | — | AST 열거 | 2단계 RED 에서 이름 확정 |
| B3 (if@410) | `if err := log.Record(audit.Entry{` | — | AST 열거 | 2단계 RED 에서 이름 확정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST calls 8개(`ast.json` 의 calls) | 위 분기 표의 조건 · 반환이 부르는 것 | 반환 오류는 분기 표대로 | AST |

## State mutations and fallbacks

- `status.Verified`(406) · `status.EntryPermitted`(409) 씀, 감사 1줄.

## Safety conclusion

- a199 편집(예정): 세 번째 결과(청산 전용): Verified=false 유지 · 새 표지 · EntryPermitted=false · 새 감사 action · 진입 래치(D3·D4·D5·D8).
- Safe edit boundary: 위 편집 칸 밖의 분기는 무편집이 목표 — 2단계 post-edit 지도가 옛/새 AST 를 정렬해 확인.
- High-risk impact: yes — 기동 인터록 · 주문 권한 · 손절 경로 가용성(안전 부트스트랩 5)
