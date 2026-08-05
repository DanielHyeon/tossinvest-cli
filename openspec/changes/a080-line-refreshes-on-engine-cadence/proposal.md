# a080 · 라인은 엔진의 박자로 갱신된다

- **Feature**: `FEAT-TOS-004` — Operator console controls and visibility
- **Story**: `STORY-TOS-a080`
- **Spec**: `operator-console`

## Why

a067은 `/positions`와 `/dashboard`의 `라인` 열이 **무엇을 보여주는가**를 고쳤다.
다섯 값은 이제 엔진이 살아 있는 한 나이로 닫히지 않는다.

이 change는 남은 절반을 고친다 — **얼마나 자주 보여주는가**.

엔진은 5초마다 판정한다.

```go
// internal/app/engine/exitloop.go
const DefaultExitObservationInterval = 5 * time.Second
```

화면은 30초마다 다시 그린다. 두 화면이 각자 같은 식으로.

```go
// internal/console/portfolio_pages.go
func (positionsPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }

// internal/console/overview.go
func (overviewPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }
```

```html
<!-- internal/console/templates.go -->
{{if .Refresh}}<meta http-equiv="refresh" content="{{.RefreshSeconds}}">{{end}}
```

**화면 갱신 주기가 브로커 캐시 TTL에서 파생된다.** 그 30초는 `holdingsTTL`이고,
`holdingsTTL`은 "브로커에 얼마나 자주 물어봐도 되는가"라는 rate budget 상수다
(`console-operator-overview`: "one refresh is one call"). 보호선을 얼마나 자주
다시 그려야 하는가와는 아무 관계가 없다. a067이 신선도 판정에서 걷어낸 것과
**정확히 같은 혼동**이 갱신 주기에 한 번 더 남아 있는 것이다.

결과: 워터마크가 오르고, 기준선이 승격되고, rung 발의가 pending으로 들어가도
운영자는 그것을 **최대 30초 늦게** 본다. 판정은 이미 5초 전에 끝났고 그 값은
이미 journal에 있는데, 화면이 그것을 읽으러 가지 않는다.

### 파생을 지키는 근거는 캐시 코드와 모순된다

`positionsPage.RefreshSeconds` 위의 주석이 그 파생을 이렇게 정당화한다.

```go
// RefreshSeconds is the reload period: exactly the holdings cache TTL, derived
// from it so the two cannot drift apart — a period under the TTL would be a
// reload that costs broker calls faster than the budget the spec fixes.
```

**그 문장은 현재 코드에서 참이 아니다.** `get`은 `tried` 기준으로 엄격히
TTL-gate된다.

```go
// internal/console/holdings.go
if !hold && (!c.attempted || now.Sub(c.tried) >= c.ttl) {
	c.refreshLocked(ctx, now)
}
```

리로드 주기가 5초든 30초든 브로커 호출은 **TTL당 최대 1회**다. 주기를 TTL
아래로 내리는 것이 호출 수를 늘리지 않는다. 늘어나는 것은 캐시 조회와 템플릿
렌더뿐이고, 그 둘은 rate budget 항목이 아니다.

즉 30초는 예산이 요구한 값이 아니라 **예산이 요구한다고 잘못 적힌 값**이다.
a080은 그 파생을 끊고 두 상수가 각자의 이유로 서 있게 한다. 이 사실은
`issues.md` I1로 기록한다.

### 그렇다고 `RefreshSeconds`만 5로 바꾸면 안 된다

예산이 막지 않는다는 것이 그 한 줄 수정으로 충분하다는 뜻은 아니다. 남는
문제가 둘 있다.

1. **조작 중인 화면이 날아간다.** 5초마다 전체 리로드는 스크롤 위치와 입력
   중인 폼을 지운다. 설정 화면에 meta refresh가 없는 것도 같은 이유이고
   (`console_prose_test.go`의 `reloadingScreens`가 그 경계를 고정한다),
   30초에 한 번이라 견딜 만하던 것이 5초에 한 번이면 화면을 못 쓰게 만든다.
   사용자가 부분 갱신을 요구한 이유다.
2. **한 상수가 두 결정을 계속 대표한다.** 값만 바꾸고 파생을 남기면 다음에
   `holdingsTTL`을 건드리는 사람이 자기도 모르게 화면 주기를 바꾸고, 그
   반대도 같다. I1의 잘못된 주석이 생긴 경로가 정확히 그것이다.

## What Changes

### 두 상수를 분리한다

`holdingsTTL`(브로커 호출 주기)과 화면 갱신 주기는 별개의 값이 된다. 전자는
30초 그대로다 — 이 change는 rate budget을 **한 호출도** 늘리지 않는다. 후자는
엔진 관측 주기와 같은 5초가 된다.

### 기전은 그대로 두고 주기만 옮긴다

두 화면은 이미 meta refresh로 스스로 다시 열린다. a080은 그 주기의 출처만 바꾼다.
새 라우트·새 템플릿·새 클라이언트 코드·새 주입 seam이 전부 0건이다.

> **초안은 여기서 달랐다.** 처음 설계는 fragment endpoint와 hash-pinned 스크립트로
> 라인 셀만 교체하는 것이었다. 그것은 콘솔의 무-JavaScript 아키텍처를 위반하고,
> 근본 원인 조사에서 스크립트를 버려도 두 번째 스펙 제약이 남는다는 것이 드러났다
> (issues.md I3·I4). 아래는 그 조사 뒤의 설계다.

### 상한을 지키는 주체를 스펙에 바로 쓴다

`rate budget 보호`는 재로드 주기가 **캐시 TTL 이상**이어야 한다고 SHALL로 정하고,
그 근거를 괄호에 이렇게 적는다 — "열린 탭 하나의 비용 상한은 holdings 1콜/TTL".

그 상한을 실제로 지키는 것은 주기가 아니라 캐시다.

```go
// internal/console/holdings.go
if !hold && (!c.attempted || now.Sub(c.tried) >= c.ttl) {
	c.refreshLocked(ctx, now)
}
```

주기를 6배로 올리면 **캐시** 도달이 6배가 되고 브로커 도달은 그대로다. 측정으로
확인했다 — TTL 4주기 동안 5초마다 24회 재로드해도 브로커 호출은 5회 이내이고,
이 gate를 제거하면 즉시 24회가 된다.

그래서 MODIFY는 요구사항을 약화하지 않는다. **상한을 주기 조건으로 대리 표현하던
것을 상한 자체로 바꾸고**, 그 상한을 누가 지키는지 명시한다. 서버측 폴러 부재,
갱신 1회 = holdings 1콜, TTL 15초 하한, 검증 중 보류와 그 우회 금지는 글자 그대로
유지된다.

### 전체 재로드의 비용은 이 두 화면에서 이미 지불되어 있다

5초 전체 재로드가 보통 잃는 것들을 하나씩 확인했다.

| 잃을 것 | 실제 | 근거 |
|---|---|---|
| 열린 접힘 | 안 잃는다 | 재로드 화면의 접힘 상태는 URL에 있다 (a055 §6) |
| 편집 중인 폼 | 없다 | 두 화면에 form·input·button이 아예 없다 (a057이 옮겼고 테스트가 고정한다) |
| 브로커 예산 | 안 쓴다 | 위 |
| 스크롤 위치 | 브라우저 복원에 의존 | **미검증** — 유일하게 열린 위험, task 6.5 실측 |

## Non-Goals

- **`holdingsTTL`을 바꾸지 않는다.** 브로커 호출 주기는 rate budget 사안이고
  이 change의 주장은 "표시를 빠르게"이지 "브로커에 자주 묻자"가 아니다.
- **엔진 관측 주기 5초를 바꾸지 않는다.** exit 경로는 High-risk이고, 표본
  간격이 허용 오차를 정의한다는 것은 `exit-policy` spec이 이미 SHALL로
  못박은 사실이다. 화면을 빠르게 하는 것은 그 오차를 줄이지 않는다.
- **5초 관측 공백과 엔진 정지 중 무보호를 해결하지 않는다.** 그것은 브로커측
  조건주문(로드맵 2c `add-protection-orders`)의 몫이고, `ProtectiveCapability`
  산출은 `verify-observes-the-trigger` 3.x 실측 뒤에만 열린다. a080은 그
  기다림 동안 **운영자가 더 빨리 보게** 만들 뿐이며, 보호 자체는 1초도
  앞당기지 않는다. 이 구분을 흐리면 안 된다.
- **client-side script를 도입하지 않고 CSP를 완화하지 않는다.** 콘솔의 무-JS는
  이 change가 지켜야 할 아키텍처다 (issues.md I3·I4).
- 다른 화면의 주기는 건드리지 않는다. `/orders`·`/signals`는 자기 데이터 소스의
  주기를 그대로 쓰고, 설정·이력 화면은 계속 자동 재로드가 없다.
- `/position-management`와 `cmd/tossctl/httpapi_reader.go`(별도 daemon, 별도
  포트)는 대상이 아니다.

## Impact

- `internal/console/line_cadence.go` — **신규 파일**. `lineRefreshInterval` 상수와
  그 근거(왜 엔진 주기인가, 왜 예산이 막지 않는가, 왜 스크립트가 아닌가).
- `internal/console/portfolio_pages.go`·`overview.go` — 두 `RefreshSeconds`가
  새 상수를 읽고, 파생을 정당화하던 잘못된 주석을 정정한다. 각각 분기 0개의
  순수 함수이고 Function Logic Map을 편집 전에 작성했다.
- 테스트 4건 갱신 — 전부 옛 결합(주기 = `holdingsTTL`)을 명문화하던 것들이다.
  갱신은 약화가 아니라 **판정 근거의 이동**이며, 상한 판정은 새로 추가한
  `TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling`이 직접 한다.
- `openspec/specs/operator-console` — **MODIFIED 2건** + ADDED 1건(무-스크립트
  자세의 성문화). MODIFIED는 `rate budget 보호`의 한 절과, `콘솔 공통 상태 표시줄`의
  scenario `화면별 재로드 주기 보존` 하나다. 두 번째는 독립 리뷰가 찾았다
  (review.md F3 — 초안은 그 scenario와 정면으로 모순된 채 통과할 뻔했다).
  MODIFIED는 a055 `issues.md` I1의 부채를 늘리지만, 기존 SHALL과 모순되는 ADDED를
  얹는 것보다 낫다.
- **선행 의존: a081.** 재로드를 6배로 올리는 것이 성립하려면 이 두 화면의 엔진
  프로세스 읽기가 렌더 횟수에서 분리되어 있어야 한다. 그 분리를 만드는 것은
  a081이며, a080은 a081이 land한 뒤에만 land한다 (review.md F1).
- rate budget: **새 계상 항목 0건.** 새로운 공식 API 호출이 없고, 보유 조회 rate는
  `holdingsTTL`이 정한 상한 안에 머문다 — 그 상한을 지키는 것이 주기가 아니라
  캐시라는 것이 이 change의 핵심 근거이고, 측정과 변이 검증으로 확인했다.
