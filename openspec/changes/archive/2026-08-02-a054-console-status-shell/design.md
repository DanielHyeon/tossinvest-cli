# Design: a054-console-status-shell

## 1. 상태 표시줄이 읽는 값 — 존재 확인

설계가 요구하는 값이 코드에 있는지 먼저 확인했다. 문서가 아니라 현재 HEAD 기준이다.

| 칸 | 출처 | 확인 |
|---|---|---|
| 엔진 상태 | `enginelock.Read(c.opts.EngineMarker, now)` → `engineView.Running`/`Wired`/`PID`/`Stale` | [engineproc.go:117-141](../../../internal/console/engineproc.go#L117-L141) |
| 설치 바이너리(엔진 stale 판정용) | `c.opts.Binary()` → `binstamp.Of(path)` = **`os.Stat`, 해시 아님** | [binstamp.go:82-97](../../../internal/binstamp/binstamp.go#L82-L97) |
| 시장 세션 | `verifylive.SessionAdvisoryFor(market, now)` → `At`/`Outside`/`Label` — 순수 계산 | [hours.go:47-59](../../../internal/verifylive/hours.go#L47-L59) |
| 데이터 신선도 | 화면이 이미 계산하는 `TakenAt`/`AgeSeconds` | 아래 §2 표 |
| 재로드 주기 | 화면의 `RefreshSeconds()` | [pages.go:95](../../../internal/console/pages.go#L95), [pages.go:150](../../../internal/console/pages.go#L150), [overview.go:1147](../../../internal/console/overview.go#L1147), [portfolio_pages.go:37](../../../internal/console/portfolio_pages.go#L37), [orders_page.go:50](../../../internal/console/orders_page.go#L50), [signals.go:1020](../../../internal/console/signals.go#L1020) |

**신규 데이터 소스 0개.** 신규 브로커 호출 0건. 상태 표시줄의 비용은 마커 파일 1회 읽기 +
`os.Stat` 1회다.

`snapshot()` 전체를 부르지 않는다. `snapshot()`은 soak 기록 파싱·attestation·검증 기록까지
읽고([data.go:194-209](../../../internal/console/data.go#L194-L209)) 그중 표시줄이 쓰는 것은
`Engine`과 `Session` 둘뿐이다. 2초 주기 화면에 나머지를 얹을 이유가 없다.

## 2. 신선도는 화면마다 다른 것을 뜻한다 — 세 상태

`RefreshedAtText`는 **엔진 마커의** 갱신 시각이지 화면 데이터의 시각이 아니다
([engineproc.go:106](../../../internal/console/engineproc.go#L106)). 둘을 한 칩에 섞으면
"데이터가 3초 전"이라는 문장이 엔진 하트비트를 가리키게 된다. 표시줄은 셋을 구분한다.

| 상태 | 해당 화면 | 표시 | 근거 |
|---|---|---|---|
| **캐시 기반** | `/positions`, `/orders`, 개요, `/signals` | 기록된 캐시 시각 + 경과 + 톤 | `TakenAt`/`AgeSeconds`가 실재 — [templates_orders.go:48](../../../internal/console/templates_orders.go#L48), [templates_overview.go:97](../../../internal/console/templates_overview.go#L97), [templates_signals.go:48](../../../internal/console/templates_signals.go#L48) |
| **요청 시점 읽기** | 검증 콘솔, `/verify`, `/settings`, `/history`, `/optimization`, `/report`, `/position-management` | 렌더 시각 + "요청 시 읽음". 톤 없음 | 이 화면들은 journal·config·기록 파일을 동기 읽기한다. 렌더 시각이 곧 읽은 시각이라 **만들어낸 값이 아니다** |
| **읽지 못함** | 어느 화면이든 | 사유 + 마지막 성공 시각(있을 때만) | 기존 SHALL NOT: 기록되지 않은 시각을 화면이 만들지 않는다 |

톤 임계는 화면의 자기 TTL을 쓴다 — 상수를 새로 만들지 않는다.
`ok` = 경과 < TTL, `warn` = TTL ≤ 경과 < 2×TTL, `stale` = 2×TTL 이상 또는 읽기 실패.
`/positions`·개요는 `holdingsTTL`, `/orders`는 `ordersTTL`이 이미 그 화면의 주기다.

**단, 갱신하지 않는 것이 정상인 상태에는 톤을 붙이지 않는다.** 이 콘솔에는 그런 상태가 둘 있다.

1. **검증 run 진행 중** — `rate budget 보호` 요구사항이 갱신 보류를 *의무화*한다. 보류 중
   캐시는 필연적으로 늙고, 그 늙음은 규정대로 동작한 결과다.
2. **스캔 주기 미도래** — `/signals`의 `TakenAt`은 discovery tick이 돌 때만 갱신된다.
   장 마감 중 스캔이 없는 것은 고장이 아니다.

경과만 보고 판정하면 정상 동작에 경고를 붙이게 되고, 늘 켜진 경고는 아무도 보지 않는다.
보류 사유가 **알려진** 경우 표시줄은 톤 대신 그 사유를 표시한다. 톤은 사유 없이 늙었을 때만
붙는다.

## 3. 공용 chrome을 어떻게 모든 화면에 넣는가

현재 page struct에는 **공통 베이스가 없다**. 각 화면이 `Nav string`·`Refresh bool`을 스스로
선언하고 `RefreshSeconds()`를 스스로 정의한다([pages.go:78-95](../../../internal/console/pages.go#L78-L95)).
`head` 템플릿은 그 필드들을 duck-typing으로 찾는다([templates.go:203](../../../internal/console/templates.go#L203)).

**선택: Go 구조체 임베딩.** 템플릿은 승격된 필드·메서드를 그대로 찾으므로 기존
`{{.Nav}}`·`{{.Refresh}}`·`{{.RefreshSeconds}}`가 문법 변경 없이 계속 동작한다.

```go
type chrome struct {
    Nav      string
    Refresh  bool
    Status   statusStrip   // 엔진·세션·신선도·주기
}
func (chrome) RefreshSeconds() int { return 0 }   // 기본값. 화면이 정의하면 화면이 이긴다
```

거부한 대안: `render`가 `{Page, Chrome}` 래퍼를 만드는 방식. 8개 템플릿의 모든 `{{.X}}`가
`{{.Page.X}}`가 되어 diff가 표시 변경과 무관한 곳까지 번진다.

**주의 — 임베딩이 조용히 바꿀 수 있는 것.** 바깥 타입의 메서드가 승격된 메서드를 가린다.
`dashboardPage`는 `RefreshSeconds() int { return 2 }`를 값 리시버로 갖고 있으므로 그대로 이긴다.
그러나 어떤 화면에서 그 메서드를 실수로 지우면 **컴파일도 테스트도 실패하지 않고** 주기가
0으로 바뀐다. 화면별 주기 테스트를 RED로 먼저 만든다(§6 T2.1).

## 4. 경로 이동 — 바꿀 곳과 바꾸지 말 곳

`"/"` 문자열은 비테스트 코드에 여러 번 나오고 **두 종류가 섞여 있다**.

| 위치 | 무엇 | 처리 |
|---|---|---|
| [console.go:715](../../../internal/console/console.go#L715) | 라우트 등록 | 리다이렉트 핸들러로 교체 |
| [pages.go:98](../../../internal/console/pages.go#L98) | `handleDashboard`의 경로 가드 | 검증 콘솔의 새 경로로 |
| [restart.go:149](../../../internal/console/restart.go#L149) `restartTarget` | 재시작 인터스티셜의 복귀 경로 | **검증 콘솔** — 재시작 컨트롤이 그 화면에 있다 |
| [restart.go:321](../../../internal/console/restart.go#L321) `redirectDashboard` | 재시작·soak 재기동 **결과 안내** | **검증 콘솔** + `?notice=` 보존 |
| [remote.go:327](../../../internal/console/remote.go#L327) | 로그인 후 리다이렉트 | 개요 경로로 |
| [templates.go:217](../../../internal/console/templates.go#L217) | nav href | 검증 콘솔 새 경로 |
| [templates_overview.go:36](../../../internal/console/templates_overview.go#L36), [:63](../../../internal/console/templates_overview.go#L63) | 본문 링크 | 검증 콘솔 새 경로 |
| [remote.go:435](../../../internal/console/remote.go#L435), [remote.go:503](../../../internal/console/remote.go#L503), [restart.go:259](../../../internal/console/restart.go#L259) | **쿠키 `Path: "/"`** | **바꾸지 않는다** |
| [console.go:474-476](../../../internal/console/console.go#L474-L476) | public URL 정규화 | 바꾸지 않는다 |

쿠키 스코프를 `/dashboard`로 좁히면 `/positions`에서 세션이 사라진다. 경로 일괄 치환이
이 change에서 가장 쉬운 사고다.

`/`는 404가 아니라 **303 리다이렉트**로 남긴다. 북마크와 이미 발행된 링크가 바깥에 있다.

### 핸드오프는 리다이렉트가 나르지 않는다

초안은 "리다이렉트가 `?handoff=`를 보존해야 한다"고 썼다. **코드를 읽으면 반대다.**

1. `session0`([console.go:801-841](../../../internal/console/console.go#L801-L841))이 핸들러보다
   **먼저** 돈다. `acceptHandoff`가 거기서 토큰을 소비하므로 토큰은 `handleDashboard`에
   도달하지 않는다.
2. `grantSession`([restart.go:249-251](../../../internal/console/restart.go#L249-L251))은
   `q.Del("handoff")`로 파라미터를 **의도적으로 지우고** 같은 경로로 리다이렉트한다.

`/`가 리다이렉트가 된 뒤 실제 흐름은 이렇다.

```
GET /?handoff=TOKEN  →  session0: 소비·세션 발급 → 303 /          (handoff 제거됨)
GET /                →  session0: 쿠키 OK → 리다이렉트 핸들러 → 303 개요
GET <개요>            →  렌더
```

리다이렉트가 3회지만 정확하다. 계약은 **결과**로 쓴다 — 토큰 1회 소비, 렌더된 화면 착지,
주소창에 소비된 토큰 잔류 금지. 보존해야 하는 쿼리는 `?notice=`다.

## 5. 표시 프리미티브

| 항목 | 지금 | 뒤 |
|---|---|---|
| 상태 알약 | `.status-pill` **8곳** + `.state-badge` **2곳** | `.status-pill` 하나 |
| 표 | bare `<table>` 28개 / `.data-table` 5곳 | 전부 `.data-table`, 열이 깨지는 것만 `.table-scroll` |
| 타입 | 15 / 16.8 / 21.6 | body 15 · h3 17 · h2 20/700 · h1 28/700 |
| `dl` 라벨 열 | `12rem`([templates.go:40](../../../internal/console/templates.go#L40)) | `8rem` |

**통합 방향은 많이 쓰는 쪽을 남긴다.** 초안은 반대로 적었다. 실제 사용은
`.status-pill`이 `templates_optimization.go`에만 8곳이고
([:306-309](../../../internal/console/templates_optimization.go#L306-L309),
[:425](../../../internal/console/templates_optimization.go#L425),
[:443](../../../internal/console/templates_optimization.go#L443),
[:457-458](../../../internal/console/templates_optimization.go#L457-L458)),
게다가 기존 테스트가 그 존재를 검사한다
([strategy_runtime_test.go:195](../../../internal/console/strategy_runtime_test.go#L195)).
`.state-badge`는 2곳이고 검사가 없다. 적은 쪽을 남기면 8곳을 고치고 기존 테스트를 깨는
작업이 되며, 그 작업은 이 통합이 얻으려는 것과 무관하다.

색은 바꾸지 않는다. 현재 `.ok #1a6b2a`는 흰 배경 대비 6.7:1로 WCAG AA를 통과하고,
참조로 검토한 팔레트의 `#16A34A`는 3.48:1로 **통과하지 못한다**. 대비가 나은 값을 예쁘다는
이유로 버리지 않는다.

## 5b. 375px는 무엇으로 검증하는가

이 저장소에 레이아웃을 실제로 측정하는 테스트는 **없다**
(`grep 'scrollWidth\|375' internal/console/*_test.go` → 0건). 기존 반응형 검증은 렌더 결과에
`name="viewport"`·`@media (max-width: 720px)`·`detail-grid`가 문자열로 있는지를 본다
([strategy_runtime_test.go:195](../../../internal/console/strategy_runtime_test.go#L195)).

기존 스펙 문장이 `documentElement.scrollWidth`를 말한다는 이유로 그 문장을 복사하면
**검증할 수 없는 SHALL을 하나 더 만드는 것**이다. 자동 검사는 렌더 결과에서 판정 가능한 넷만
본다: viewport meta 존재, 모든 표가 반응형 클래스 또는 스크롤 래퍼 보유, viewport보다 넓은
고정 px 폭 부재, 좁은 viewport 미디어 쿼리 적용.

브라우저 실측은 tasks의 **1회 사람 확인 증거**로 분리한다. 하니스를 만드는 것은 이 change의
범위가 아니고, 만들지 않은 채 있는 척하는 것이 더 나쁘다.

## 6. 무엇을 하지 않는가

- **자동 재로드를 없애지 않는다.** 사용자 결정(2026-08-02): 자동 갱신은 좋고 갱신 시각을
  표시하라. 주기 정책 자체는 무변경 — `/verify`의 "run이 working일 때만"도 그대로다.
- **nav 12항목을 건드리지 않는다.** `a055`이 소유한다. 이 change는 `/`의 href 한 칸만 바꾼다.
- **StockOS의 typed-confirmation을 가져오지 않는다.** 콘솔 UI에 타이핑 확인·추가 승인 마찰을
  넣지 않는다는 사용자 지시가 있고, `docs/stockos-inventory.md`의 이식 판정도 "이식 안 함"이다.
- **⌘K 팔레트·drawer를 가져오지 않는다.** JS가 필요하고 CSP `default-src 'none'`이 이 콘솔의
  계약이다.
- **`콘솔 안전 불변식`·`편입 설정 화면`을 MODIFY하지 않는다.** 두 요구사항에는 미아카이브
  delta가 각각 6건·3건 쌓여 있다. MODIFIED는 블록 전체를 치환하므로 이 change가 끼어들면
  아카이브 순서가 스펙 본문을 결정하게 된다.

## 7. 확인된 스펙 부채 (이 change 범위 밖, 기록만)

`openspec/specs/operator-console/spec.md`의 `편입 설정 화면`(155행)은 "외부 종목 자동관리" nav를
요구하지만, 그 요구사항을 MODIFY하는 미아카이브 change 3건(`console-excludes-in-one-click`,
`console-sets-guardian-limits`, `console-operator-overview` 계열) 중 **어느 것도 그 문장을 담고
있지 않다**. 지금 그 change들을 아카이브하면 본문이 **되돌아간다**. `a055`이 이 요구사항을
MODIFY할 때 어느 텍스트를 base로 삼을지 명시해야 하며, 그 base는 현재 승인된 본문이다
(WORKFLOW 권위 경계: 의도된 동작의 권위는 `openspec/specs/` + 승인된 change).
