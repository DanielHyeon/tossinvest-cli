# Function Logic Map: `Notifier.logClaimHeld`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :742–752, 분기 1 · 반환 1 · 호출 4, source_sha256 `d705f78d68c1…`, 추출 커밋 `55963f29`(25라운드 수리 뒤 재추출). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존. 26라운드 수리 두 로트(`d8769cfb` · `b910173a`) 뒤 재추출(마지막 `15b64676`) — 이 함수 본문 · 분기 수 불변(같은 파일의 다른 함수 편집으로 줄 이동 · 파일 해시만 바뀜, 분기 좌표는 `ast.json` 이 정본).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ③ — `fbc6df5f`): Manager 판정 (나): `n.Log.Warn` → `n.Log.Event`(INFO) + 주석 정정(a098 의 산 발송자 · 단위 ③ 의 동시 동기 관측 → 정상 경로, 죽은 발송자 신호는 `logClaimStolen` WARN).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `claim` | `ClaimHeldElsewhere` 결과 | `claimAndDeliver` B6 · `Flush` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `n.Log == nil` (:717) | — | 반환 | (미실행) |
| 종단 | — | `n.Log.Event(EventAlertClaimHeld, 보유자 · 나이 · 만료)` | — | `TestAHeldRowIsNotWhispered`(INFO · 실재 · 보유자) — 변이 L11(줄 삭제) · L12(WARN 복귀) CAUGHT |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Log.Event` | 한 줄 | — | AST |

## State mutations and fallbacks

- 로그 한 줄만.

## Safety conclusion

- Safe edit boundary: 등급 · 주석.
- High-risk impact: 낮음.
