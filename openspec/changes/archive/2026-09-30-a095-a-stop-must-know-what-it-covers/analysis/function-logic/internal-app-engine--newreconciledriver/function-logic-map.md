# Function Logic Map: `NewReconcileDriver`

- Source: `internal/app/engine/reconcileloop.go` (`304`–`345`)
- Qualified: `NewReconcileDriver`
- AST evidence: `ast.json` (`source_sha256` 9ca090f75ee95da0…) — **편집 뒤**(구현 로트 10판, 2026-09-30 재추출)
- Risk scan: `risk-pattern-report.md`
- 분기 10 · return 8

**역할.** 대사 루프의 배선을 검증하고 드라이버를 만든다. a095가 여기서 바꾸는 것은 **메모리 래치 두 map의 모양**뿐이다 —
분기는 바뀌지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Journal` · `Collector` · `Tracker` · `Ingest` · `Converge` | 모두 non-nil | 조립부 | B2~B5 — `ErrReconcileDriverUnavailable` |
| `opts.AccountRef` | 공백 아님 | 조립부 | B6 — 거절 |
| 편입 켜짐 · include 지정 ∧ `Prices == nil` | 불가 | 설정 · 조립부 | B7 — 거절(합성 t0는 관측이어야 함) |
| `opts.CommonPolicy` | 빈 값 또는 등록된 id | 설정 | B8 · B9 — 모르는 id 거절 |
| `opts.Clock` | nil 허용 | 조립부 | B10 — 시스템 시계 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:305` `switch {` | — | — | 기존 |
| B2 | `:306` `case opts.Journal == nil:` | — | `:307` 오류 | 기존 |
| B3 | `:308` `case opts.Collector == nil:` | — | `:309` 오류 | 기존 |
| B4 | `:310` `case opts.Tracker == nil:` | — | `:311` 오류 | 기존 |
| B5 | `:312` `case opts.Ingest == nil \|\| opts.Converge == nil:` | — | `:313` 오류 | 기존 |
| B6 | `:314` `case strings.TrimSpace(opts.AccountRef) == "":` | — | `:315` 오류 | 기존 |
| B7 | `:317` `case (opts.Adoption.Enabled \|\| len(opts.Adoption.IncludeSymbols) > 0) && opts.Prices == nil:` | — | `:318` 오류 | 기존 |
| B8 | `:321` `if id := strings.TrimSpace(opts.CommonPolicy); id != "" {` | `opts.CommonPolicy = id` | — | 기존 |
| B9 | `:322` `if _, ok := exitpolicy.CommonPolicyByID(id); !ok {` | — | `:323` 오류 | 기존 |
| B10 | `:334` `if d.clk == nil {` | `d.clk = clock.System()` | — | 기존 |

## Calls and live bindings

`exitpolicy.CommonPolicyByID`(B9) · `clock.System`(B10) · `fmt.Errorf`. 브로커 · 원장 호출 없음. 오류는 조립 거절이며 되던진다.

## State mutations and fallbacks

- 드라이버 구조체 생성: `unmanaged` · `grown` 래치 map 초기화(`:331-332`), `stabiliser`, `ingest` 복사 뒤 `d.ingest.Alert = nil`(`:342-343`).

## Safety conclusion

- **Safe edit boundary (a095, 편집 뒤)**: `:331` `unmanaged: map[string]bool{}` → (포지션 → 조건 집합) map, `:332` `grown: map[string]bool{}` →
  (포지션 → 보고한 최대 수량) map. **분기 B1~B10 무변화** — 거절 조건 · 순서 · 오류는 그대로다. `NotificationsEnabled`는 검증하지 않는다
  (bool, 영값 = 꺼짐 = 결정 (2)의 안전 방향).
- **High-risk impact**: yes(대사 루프의 조립) — 편집은 초기값의 타입뿐.
