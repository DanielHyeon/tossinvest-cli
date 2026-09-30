# Function Logic Map: `Context.ReconcileDriver`

- Source: `internal/app/engine/reconcileloop.go` (`353`–`381`)
- Qualified: `Context.ReconcileDriver`
- AST evidence: `ast.json` (`source_sha256` 9ca090f75ee95da0…) — **편집 뒤**(구현 로트 10판, 2026-09-30 재추출)
- Risk scan: `risk-pattern-report.md`
- 분기 5 · return 3

**역할.** 생산 조립 — 엔진 `Context`의 부품과 **로드된 설정**으로 대사 드라이버 옵션을 채운다. 생산의 유일한 대사 드라이버 조립 자리다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` | non-nil | 엔진 조립 | B1 — `ErrReconcileDriverUnavailable` |
| `c.Automation.Verified` | 참 | 자동화 게이트 | B2 — 거절 |
| `c.Config.Engine.Adoption` | 로드된 값(거부 블록은 0 + Rejected) | 설정 로더(`mergeAdoption`) | 무조건 복사(`:367`) |
| `c.Config.Engine.Notifications.Enabled` | 로드된 값(거부 블록은 거짓) | 설정 로더(`mergeNotifications` B3) | **a095가 새로 복사** |
| `opts.Prices` · `opts.Alerts` · `opts.Log` | 호출자가 주면 유지 | 호출자 | B3~B5 — 비어 있으면 Context 값 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:354` `if c == nil {` | — | `:355` 오류 | 기존 |
| B2 | `:357` `if !c.Automation.Verified {` | — | `:358` 오류 | 기존 |
| B3 | `:371` `if opts.Prices == nil {` | `opts.Prices = c.Official` | — | 기존 |
| B4 | `:374` `if opts.Alerts == nil && c.Notifier != nil {` | `opts.Alerts = c.Notifier` | — | 기존 |
| B5 | `:377` `if opts.Log == nil {` | `opts.Log = c.Log` | — | 기존 |

## Calls and live bindings

`NewReconcileDriver`(`:380`, 결과를 그대로 반환). live binding: `c.Config.Engine.Adoption` · `c.Config.Engine.ExitPolicy.CommonPolicy` —
a095 뒤에는 `c.Config.Engine.Notifications.Enabled`도.

## State mutations and fallbacks

- `opts`(값 복사)의 필드를 Context 값으로 **무조건** 덮는 줄(`:360-369`)과 비어 있을 때만 채우는 B3~B5.

## Safety conclusion

- **Safe edit boundary (a095)**: 무조건 복사 줄 묶음(`:360-369`)에 `opts.NotificationsEnabled = c.Config.Engine.Notifications.Enabled` 한 줄을
  더한다 — 분기 없음. 무조건인 이유: 생산 조립이 「알림 켜짐」을 **항상** 로드된 값으로 넘겨야 critical 판정이 전송기 유무가 아니라
  설정으로 선다(델타 「알림 켜짐은 로드된 설정의 값으로 판정」). 호출자가 준 값을 남기면 시험 아닌 호출자가 켜짐을 지어낼 수 있다.
- **High-risk impact**: yes — 이 값이 대사 쪽 critical(→ 진입 차단)의 전제다.
