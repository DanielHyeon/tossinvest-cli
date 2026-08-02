# Change: a054-console-status-shell

## Story Mapping

- Story: `STORY-TOS-a054`
- Feature: `FEAT-TOS-004`
- Story ↔ OpenSpec: 같은 `a054` 번호로 1:1 연결

## Why

사용자 요청(2026-08-02): 콘솔이 **너무 복잡하고 응집력이 없다**. 참조는 StockOS다.

측정하면 불평은 구체적이다.

| 사실 | 측정 |
|---|---|
| "이 숫자가 언제 것인가"를 화면마다 다르게 말한다 | `TakenAt`·`AgeSeconds`·`RefreshedAtText`·`NowText`가 **38곳**에 흩어져 있고 전부 설명 문단 안에 있다 |
| 모바일에서 표가 넘친다 | `.data-table` 없는 bare `<table>` **28개**. 개요 화면은 표 6개 전부가 bare — `.table-scroll` 래퍼도 없다 |
| 무엇이 중요한지 화면이 말하지 않는다 | body 15px / h2 16.8px / h1 21.6px — h2가 본문보다 **12%** 클 뿐이다 |
| 같은 것에 이름이 둘이다 | `.status-pill`([templates.go:66](../../../internal/console/templates.go#L66), 사용 8곳)과 `.state-badge`([templates.go:109](../../../internal/console/templates.go#L109), 사용 2곳)는 같은 프리미티브다 |
| 공용 프리미티브가 있는데 안 쓴다 | `.status-strip` 1회, `.eyebrow` 2회, `.section-kicker` 2회 — 정의는 있고 사용이 없다 |
| `/`의 이름이 셋이다 | 경로는 `/`, nav 라벨은 "검증 콘솔", `<h1>`은 "대시보드"([templates.go:310](../../../internal/console/templates.go#L310)). `/dashboard`의 `<h1>`은 "개요"다 |

StockOS는 이 여섯 중 다섯을 **한 줄의 상시 chrome**으로 푼다: 브랜드(홈) · 장 시계 ·
`DATA 3s` 신선도 칩 · 준비도 pill · 상태 칩. 신선도는 코드 한 곳에서 계산되고
([uiSignals.ts:76-90](../../../../stockos/apps/web/src/lib/uiSignals.ts)) 모든 화면이 같은 칩을
본다.

TossOS는 **그 데이터를 이미 다 갖고 있다.** 없는 것은 표면이다.

### 이 change가 내비게이션을 건드리지 않는 이유

nav 12항목 재편과 설정 재분류는 `a055-console-settings-cadence`다. 두 작업의 위험 등급이
다르다 — 이 change는 표시 계층과 라우트 **2건**을 바꾸고, a053은 nav 계약과 `편입 설정 화면`
요구사항을 바꾼다. 한 change로 묶으면 chrome 회귀와 IA 회귀가 같은 diff에서 섞인다.

### `/`를 옮기는 것이 이 change에 있는 이유

`/`가 검증 콘솔인 한 상태 표시줄은 "지금 어느 화면인가"를 정직하게 말할 수 없다 — nav는
"검증 콘솔"이라 하고 제목은 "대시보드"라 한다. 이름을 하나로 만드는 것이 chrome의 전제다.

## What Changes

- **콘솔 공통 상태 표시줄** — 모든 화면 상단에 같은 strip: 엔진 상태, 시장 세션 advisory,
  데이터 신선도(기록된 시각 + 경과 + 톤), 지금 걸려 있는 재로드 주기. 자동 재로드는
  **유지한다**(사용자 결정 2026-08-02) — 없앨 것은 "언제 것인지 모른다"이지 자동 갱신이 아니다.
- **신선도는 기록이 있을 때만 말한다** — 시각이 기록되지 않은 화면은 신선도 칸을 비우고
  사유를 말한다. 화면이 시각을 만들어내지 않는다는 기존 SHALL NOT을 chrome에서도 지킨다.
- **`/`는 개요로 리다이렉트하고 검증 콘솔은 `/verify/console`을 갖는다** — 재시작 핸드오프
  ([restart.go:147-151](../../../internal/console/restart.go#L147-L151))와 원격 로그인 후
  리다이렉트([remote.go:327](../../../internal/console/remote.go#L327))의 착지 경로를 함께 옮긴다.
- **핸드오프 토큰은 리다이렉트에 실어 나르지 않는다** — `session0`이 핸들러보다 먼저 돌고
  ([console.go:801-841](../../../internal/console/console.go#L801-L841)) `grantSession`이
  `handoff` 파라미터를 **의도적으로 삭제한 뒤** 리다이렉트한다
  ([restart.go:249-251](../../../internal/console/restart.go#L249-L251)). 계약은 결과로 쓴다 —
  토큰 1회 소비, 렌더된 화면 착지, 주소창에 소비된 토큰 잔류 금지.
- **재시작·soak 안내는 검증 콘솔로 돌아간다** — `redirectDashboard`
  ([restart.go:320-326](../../../internal/console/restart.go#L320-L326))가 싣는 것은 재시작
  결과이고 그 컨트롤은 검증 콘솔에 있다. 개요로 보내면 누른 버튼의 결과를 다른 화면에서 읽게 된다.
- **승인 대기 중인 검증 run을 표시줄이 알린다** — 승인 창은 짧고 소진 사고 기록이 있다
  (M11·M18·M22·M23). 표시줄은 모든 화면에 있으므로 이 표시가 승인 창의 발견성을 화면 위치와
  무관하게 만든다. 이것이 `a055`의 검증 콘솔 이동을 승인하는 전제다.
- **쿠키 `Path: "/"`는 바꾸지 않는다** — [remote.go:435](../../../internal/console/remote.go#L435),
  [remote.go:503](../../../internal/console/remote.go#L503),
  [restart.go:259](../../../internal/console/restart.go#L259)의 `"/"`는 라우트가 아니라 쿠키
  스코프다. 경로 일괄 치환이 이 셋을 건드리면 세션이 깨진다.
- **화면 이름 일치** — 각 화면의 `<h1>`과 nav 라벨과 `Nav` 식별자가 같은 것을 가리킨다.
- **표시 프리미티브 통합** — `.status-pill`과 `.state-badge`를 하나로, bare `<table>` 28개를
  `.data-table`로, 타입 스케일을 15/17/20/28로, `dl` 라벨 열을 12rem에서 8rem으로.
- **개요 상단 요약** — 세로 6섹션 앞에 요약 strip. 각 칸은 상세를 소유한 화면으로 간다.

## Capabilities

### Modified Capabilities

- `operator-console`: 모든 화면이 같은 상태 표시줄과 같은 표시 프리미티브를 쓰고, 화면 이름과
  경로와 제목이 한 화면을 가리킨다.

## Impact

- **Affected specs**: `operator-console` — `콘솔 공통 상태 표시줄`,
  `화면 이름·경로·제목은 한 화면을 가리킨다`, `포지션·주문 외 화면의 반응형 표시`,
  `개요는 상단 요약에서 답한다` **ADDED**. MODIFIED 없음(아래 위험 참조).
- **자동 검사에 넣지 않는 것**: 브라우저 레이아웃 실측. 이 저장소에는 `documentElement.scrollWidth`를
  측정하는 테스트 하니스가 없고(`grep 'scrollWidth\|375' internal/console/*_test.go` → 0건),
  기존 반응형 검증은 렌더 결과의 문자열 존재 검사다
  ([strategy_runtime_test.go:195](../../../internal/console/strategy_runtime_test.go#L195)).
  없는 하니스를 전제한 SHALL은 검증되지 않은 채 통과한다 — 실측은 사람이 1회 수행해 증거로 남긴다.
- **Affected code**: `internal/console` — `templates.go`(style/head/nav/foot), 화면 템플릿 8개,
  page struct에 공용 chrome 임베딩, `console.go` 라우트 2건, `restart.go`, `remote.go`,
  `pages.go`.
- **Affected code(무변경)**: `internal/journal`·`internal/official`·`internal/config`·
  `internal/verifylive`. 상태 표시줄이 읽는 값은 전부 기존 접근자다 —
  `enginelock.Read`([engineproc.go:129](../../../internal/console/engineproc.go#L129)),
  `verifylive.SessionAdvisoryFor`, 각 화면이 이미 계산하는 캐시 시각.
- **Data/operations**: 신규 브로커 호출 **0건**. 상태 표시줄은 `os.Stat` 1회와 마커 파일 1회
  읽기로 구성되며 `binstamp.Of`는 해시가 아니라 stat이다
  ([binstamp.go:82-97](../../../internal/binstamp/binstamp.go#L82-L97)).
- **Safety invariants**: 콘솔의 상태변경 행위 목록은 **한 건도 늘지 않는다**. 신규 라우트는
  전부 GET이고 폼이 없다. 손절·익절·사이징·Guardian·원장·인증 경로 무변경.

## Risks

| 위험 | 완화 |
|---|---|
| `operator-console` 스펙에 미아카이브 delta가 쌓여 있다 — `콘솔 안전 불변식`을 MODIFY하는 change가 **6건**, `편입 설정 화면`이 **3건** | 이 change는 **ADDED만** 쓴다. MODIFIED 블록 전체 치환 경쟁에 참여하지 않는다 |
| 경로 일괄 치환이 쿠키 스코프 `Path: "/"`를 건드린다 | 위 3개 위치를 proposal과 design에 명시하고, 세션 쿠키 왕복 테스트를 RED로 먼저 만든다 |
| 핸드오프 토큰이 사라진 경로에 착지한다 | `restartTarget`의 반환값을 새 개요 경로로 옮기고, 토큰 소비 후 착지 화면이 렌더되는 것을 테스트로 고정 |
| 공용 chrome 임베딩이 화면별 `RefreshSeconds()`를 덮는다 | 임베딩은 기본값만 제공하고 화면이 정의한 메서드가 우선함을 **각 화면 주기별 테스트**로 고정. 특히 `/verify`의 "run이 working일 때만 2초"가 유지되는지 |
| 상태 표시줄이 매 렌더마다 파일을 읽어 2초 주기 화면의 비용이 는다 | 읽는 것은 마커 1개와 stat 1회. 브로커 호출은 늘지 않음을 rate budget 테스트로 확인 |
| 세션 advisory가 휴일을 모르는데 표시줄이 "장중"이라 단언한다 | `SessionAdvisory`는 요일·시각만 안다([hours.go:63-64](../../../internal/verifylive/hours.go#L63-L64)). 표시줄은 그 한계를 라벨에 담고 판단 근거로 쓰지 않는다 |
| bare table 28개를 `.data-table`로 바꾸면 `table-layout: fixed`가 기존 열 폭을 깬다 | 화면별로 375px·1280px 두 폭에서 렌더 검사. 열이 깨지는 표는 `.table-scroll` 래퍼로 처리하고 어느 쪽을 썼는지 기록 |
