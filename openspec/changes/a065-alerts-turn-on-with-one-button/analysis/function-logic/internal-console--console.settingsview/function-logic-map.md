# Function Logic Map: `Console.settingsView`

- Source: `internal/console/settings.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Base commit: `openspec/changes/a065-alerts-turn-on-with-one-button/base-commit.txt`
- 위험 등급: Normal — 화면 하나를 렌더하기 위한 읽기다. 계좌·원장·브로커에 닿지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.opts.Settings` | nil 가능 | 주입 | `Wired=false`, 카드가 사유를 렌더 |
| `c.opts.Limits` | nil 가능 | 주입 | `LimitsWired=false` |
| `c.opts.TradingPolicy` | nil 가능 | 주입 | `TradingWired=false` |
| `c.opts.EngineBoot` | nil 가능 | 주입 | `AutostartWired=false` |
| **`c.opts.Notifications`** | nil 가능 | 주입 | `NotificationsWired=false` |
| `c.opts.SystemUpdater` | nil 가능 | 주입 | `UpdateWired=false` |
| `r.URL.Query()` | 임의 | 브라우저 | 문자열 그대로 |

**불변식 (유지)**: 각 seam은 **독립적으로** 읽힌다. 함수 주석이 이유를 적는다 — 네 탭이
같은 것을 읽으므로 읽기를 탭별로 쪼개면 새 seam이 생길 때 맞춰야 할 자리가 넷이 된다.
그리고 한 seam의 실패가 다른 seam의 섹션을 죽여서는 안 된다: 각 블록은 자기
`…LoadErr`에만 기록하고 조기 반환하지 않는다.

**불변식 (유지)**: 이 함수는 **쓰지 않는다.** 읽기만 하고 `settingsPage`를 만든다.
`TestTheConsoleWritesNothingButTheEvidenceItsRunnerWrites`가 패키지 전체에 대해
그것을 강제한다.

**a065가 바꾸는 것**: `c.opts.Notifications != nil` 블록 하나를 `EngineBoot` 블록과
`engineNoteNow()` 호출 사이에 더한다. 그 블록은 기존 넷과 **같은 모양**이다 — wired
플래그, `Load()`, 실패 시 `…LoadErr`, 값 대입.

**a065가 바꾸지 않는 것**: 기존 열 분기의 조건·순서·대입, `engineNoteNow()`의 위치,
`ReleaseDownloadWired`의 계산.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (181) | `opts.Settings != nil` | `Wired`, `Block`, `Verdict` | — | 기존 |
| B2 (184) | 편입 읽기 실패 | `LoadErr` | — | 기존 |
| B3 (189) | `opts.Limits != nil` | `LimitsWired`, `Gate` | — | 기존 |
| B4 (192) | 한도 읽기 실패 | `LimitsLoadErr` | — | 기존 |
| B5 (197) | `opts.TradingPolicy != nil` | `TradingWired`, `Trading` | — | 기존 |
| B6 (200) | 거래 정책 읽기 실패 | `TradingLoadErr` | — | 기존 |
| B7 (209) | `opts.EngineBoot != nil` | `AutostartWired`, `Autostart` | — | 기존 |
| B8 (212) | 자동 시작 읽기 실패 | `AutostartLoadErr` | — | 기존 |
| **B9 (217)** | **`opts.Notifications != nil`** | `NotificationsWired`, `Notifications` | — | **4.1–4.3, 7.10** |
| **B10 (220)** | **알림 읽기 실패** | `NotificationsLoadErr` | — | **7.10** |
| B11 (226) | `opts.SystemUpdater != nil` | `UpdateWired`, `Update`, 릴리스 영수증 | — | 기존 |

조기 반환은 없다. 함수 전체가 하나의 읽기 시퀀스이고 마지막에 `page`를 반환한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `c.chromeOnRequest` | 공통 상태 표시줄 | 없음 | AST |
| `c.engineRunning` | 반영 시점 문장 | 마커 파일 stat, 실패=false | AST |
| `Settings.Load` | 편입 블록 | error → `LoadErr` | AST |
| `Limits.Load` | 한도 블록 | error → `LimitsLoadErr` | AST |
| `TradingPolicy.Load` | 거래 정책 | error → `TradingLoadErr` | AST |
| `EngineBoot.Load` | 자동 시작 | error → `AutostartLoadErr` | AST |
| **`Notifications.Load`** (신규) | 알림 블록 | error → `NotificationsLoadErr` | 신규 |
| `c.engineNoteNow` | 마지막 기동 결과 | 없음 | AST |
| `SystemUpdater.Inspect` | 업데이트 상태 | 없음 | AST |

**네트워크 호출이 없다.** 새 seam의 `Load`는 로컬 설정 파일 파싱이다 — `Test`만
네트워크를 쓰고 그것은 이 함수가 부르지 않는다.

## State mutations and fallbacks

- 지역 `page` 구조체만 쓴다. 파일 쓰기도, 원장 쓰기도, 주문도 없다.
- 새 필드의 zero value: `NotificationsWired=false`, `Notifications=config.Notifications{}`
  (= `Enabled:false`), `NotificationsLoadErr=""`. seam이 없는 빌드는 편집 전과 같은
  화면을 렌더한다 (§0.2).
- 알림 읽기 실패는 `NotificationsLoadErr`에만 기록되고 다른 섹션을 건드리지 않는다.
  `TestAnUnreadableAlertConfigDoesNotTakeTheTabDown`이 그것을 고정한다.

## Safety conclusion

- Safe edit boundary: 새 `if` 블록 하나(B9·B10)와 `settingsPage` 필드 셋.
- High-risk impact: **no** — 주문·손절·사이징·Guardian·원장·인증·체결 경로에 닿지 않는다.
- §0.2: seam이 주입되지 않은 빌드에서 이 함수의 관측 가능한 결과가 편집 전과 같다.
- §0.8: 이 함수는 채널 식별자를 **읽어서 `page`에 담는다.** 그것이 응답 본문으로만
  나가고 리다이렉트 URL로 나가지 않는 것은 handler의 계약이며
  `TestTheNoticeNeverCarriesTheChannel`이 고정한다.
