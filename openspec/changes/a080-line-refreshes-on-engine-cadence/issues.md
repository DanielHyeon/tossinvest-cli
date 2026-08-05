# a080 · Issues

## I1 — `RefreshSeconds`의 파생을 지키는 주석이 캐시 코드와 모순된다

**발견**: task 1.1 (근거 고정), 2026-08-04. 현재 HEAD `64f7afdb`.

`positionsPage.RefreshSeconds` 위의 주석은 파생을 이렇게 정당화한다.

```go
// internal/console/portfolio_pages.go:52-54
// RefreshSeconds is the reload period: exactly the holdings cache TTL, derived
// from it so the two cannot drift apart — a period under the TTL would be a
// reload that costs broker calls faster than the budget the spec fixes.
```

`holdingsCache.get`은 마지막 **시도 시각** 기준으로 TTL-gate한다.

```go
// internal/console/holdings.go:179
if !hold && (!c.attempted || now.Sub(c.tried) >= c.ttl) {
	c.refreshLocked(ctx, now)
}
```

`tried`는 `refreshLocked` 진입에서 갱신되므로(`c.tried, c.attempted = now, true`)
브로커 호출은 **TTL당 최대 1회**이며 리로드 주기와 무관하다. 리로드를 5초로
내려도 호출 rate는 그대로다. 늘어나는 것은 캐시 mutex 획득과 템플릿 렌더이고
둘 다 §0.4 계상 항목이 아니다.

**즉 주석이 서술하는 위험은 현재 코드에 존재하지 않는다.** 30초는 rate budget이
요구한 값이 아니라 요구한다고 잘못 적힌 값이다.

**영향**: 이 주석은 a080 이전까지 "화면을 더 자주 갱신하자"는 제안을 예산
문제로 반려하는 근거로 읽혔을 수 있다. a080은 파생을 끊으면서 이 주석도
사실에 맞게 고친다 (task 4.3).

**한계**: 주석이 처음부터 틀렸는지, 아니면 `get`이 무조건 refresh하던 시절에
맞았다가 캐시가 TTL-gate로 바뀌면서 뒤처졌는지는 확인하지 않았다. 결론이
같으므로(지금은 틀렸다) a080의 판단에는 영향이 없다.

## I4 — 근본 원인: 진짜 제약은 두 개였고 둘 다 스펙에 있다 · **해소됨**

I3을 받아 근본 원인을 조사한 결과, 이 change가 부딪힌 제약은 하나가 아니라
둘이었다. 초안은 그중 하나도 정확히 보지 못했다.

### 제약 1 — 콘솔은 무-JavaScript 표면이다 (I3)

`archive/2026-07-31-streamline-trading-views/design.md`의 **Non-Goals**:

> JavaScript, CSP nonce, `unsafe-inline` script, 외부 CSS/폰트 도입.

같은 문서 Context: "TossOS 콘솔은 외부 자산 없이 `html/template`와 inline CSS만
사용하고, 배포 CSP는 `default-src 'none'`과 `form-action 'self'`를 강제한다."

그 change가 **해결한 문제**가 정확히 이것이었다 — 편입·제외 체크박스가 CSP가
차단하는 `onchange`로 submit해서 **동작하지 않았다**. 고친 방식은 CSP 완화가
아니라 JS 의존 제거였다. "CSP 회귀 검사" scenario는 그 상태로의 회귀를 막는다.

design D5가 선례로 삼은 `optimizationPreviewScript`는 반례가 아니다. 그것은
POST로만 렌더되는 preview 서브페이지에 있고, 무-스크립트 테스트가 도는
`consoleScreens` 목록에 **들어 있지 않다**. `/optimization` 본화면은 그 목록에
있고 무-스크립트로 통과한다.

### 제약 2 — 재로드 주기가 캐시 TTL 이상이어야 한다 (스크립트를 버려도 남는다)

스크립트를 포기하고 meta refresh 주기만 5초로 내려도 걸린다.
`operator-console` — Requirement "rate budget 보호":

> 포지션 화면은 브라우저 재로드 지시(meta refresh)를 포함할 수 있으며 **그 주기는
> 캐시 TTL 이상이어야 한다(SHALL** — 각 재로드는 요청 시 lazy 갱신을 그대로 타므로
> 열린 탭 하나의 **비용 상한은 holdings 1콜/TTL**이다)

즉 두 경로 모두 스펙에 막혀 있었다. 초안이 "스펙 delta는 ADDED만"으로 갈 수
있다고 본 것 자체가 틀렸다.

### 결정적 관찰

위 SHALL의 **괄호 안 근거가 SHALL 본문보다 약하다.** 근거는 "비용 상한은
holdings 1콜/TTL"이라고 말하는데, 그 상한은 캐시가 지키는 것이지 주기가 지키는
것이 아니다(I1이 코드로, task 2.1이 측정으로 보인 그대로). 주기를 TTL 이상으로
묶는 것은 상한을 지키기 위해 **필요하지 않다**.

### 정석 수정

- 스크립트·fragment endpoint·CSP 완화·`LinesURL`·swap 마커를 **전부 철회**한다.
  제약 1은 완화 대상이 아니라 지켜야 할 아키텍처다.
- meta refresh 주기를 엔진 관측 주기(5초)로 내리고, `rate budget 보호`의 그 한
  절만 MODIFY한다 — "주기 ≥ TTL"에서 "상한은 캐시가 지키고 주기는 TTL과 독립"으로.
- 전체 재로드의 비용이 이 두 화면에서 이미 지불되어 있음을 확인했다: 접힘은
  URL이 보존하고(a055 §6), 두 화면에 form·input·button이 없으며(a057이
  `/position-management`로 옮겼고 테스트가 고정한다), 브로커 상한은 캐시가 지킨다.

### 결과

`internal/console` **696 passed, 0 failed**. 무-스크립트 가드 3건은 손대지 않고
통과한다. 갱신된 테스트 4건은 전부 옛 결합(주기 = `holdingsTTL`)을 명문화하던
것들이며, 그 결합을 끊는 것이 이 change다.

## I3 — 콘솔의 무-스크립트 자세는 스펙 요구사항이고 **초안 설계**는 그것을 위반했다

**발견**: task 4.x GREEN 도중, 2026-08-04. 구현 중단(WORKFLOW.md §예외 경로 ①).

design D5는 `optimization_view.go`의 sha256 pin을 선례로 삼아 두 화면에
`script-src`를 더하는 것을 "선례를 따르는 일"로 판단했다. **그 판단이 틀렸다.**
그것은 한 화면에 개별 부여된 예외이고, 콘솔의 기본 자세는 스펙 수준에서
무-스크립트다.

### 위반하는 것

`openspec/specs/operator-console/spec.md` — Requirement "CSP 안전한 포지션 관리 조작"

```
#### Scenario: CSP 회귀 검사
- **WHEN** 포지션 페이지의 렌더 결과와 응답 CSP를 검사하면
- **THEN** 렌더 결과 전체에 `on[a-z]+=` inline handler, `<script>`, `javascript:` URL이
  없고 응답 CSP의 `default-src 'none'`과 `form-action 'self'`가 유지된다
```

요구사항 본문의 SHALL NOT은 **조작이 script에 의존하지 않을 것**을 말하므로 이
change는 그 절을 만족한다(스크립트는 가산적이고, 없어도 화면은 오늘과 같이
동작한다 — `TestTheFallbackReloadStaysInTheMarkup`이 증명한다). 그러나 위 scenario는
의존이 아니라 **존재**를 금지하며 범위가 "렌더 결과 전체"다. 이 change는 어긴다.

### 이것을 지키는 테스트 3건

세 개의 서로 다른 change가 각자 이 불변식을 고정하고 있다.

| 테스트 | 판정 |
|---|---|
| `TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript` | `/positions`에 `<script>` 금지 |
| `TestA057HoldingsViewsStayInputFreeAccessibleAndResponsive` | 두 화면에 input/script 표면 금지 |
| `TestNoScreenSmugglesInAScript` | 모든 화면 CSP가 `consoleHTMLCSP`와 **정확히 일치**, `script-src` 포함 금지 |

세 번째가 근거를 명시한다.

```go
t.Errorf("%s permits scripts; the dead handlers above would stop being dead", screen.path)
```

즉 `script-src`를 비워 두는 것은 스크립트를 안 쓰기 위해서가 아니라, 템플릿에
inline handler가 섞여 들어와도 **반드시 죽어 있게** 만들기 위한 장치다.

### 기술적 반론과 그 한계

sha256 pin은 그 해시의 스크립트만 허용하므로 inline event handler는 여전히
실행되지 않는다(`unsafe-hashes` 없이는). 따라서 세 번째 테스트의 명시 근거는
hash-pinned 정책에 대해서는 과보수적이다.

**그것으로 진행 근거를 삼지 않는다.** 세 change가 각각 고정하고 스펙 scenario가
뒷받침하는 자세를, 그 자세를 세운 적 없는 컨텍스트가 기술적 세부로 뒤집는 것은
WORKFLOW.md §권위 경계가 막는 일이다. 실거래 화면의 deny-by-default 완화는
사람이 결정한다.

### 상태

- `internal/console` 실패 3건 — 전부 이 충돌 하나다. (라우트 정적 가드 위반은
  별건이었고 리터럴 경로로 수정 완료.)
- 스펙 위반 없이 구현된 것: fragment endpoint 2개, 상수 분리, `RowKey`, 템플릿,
  폴백 계약.
- 결정 대기: scenario MODIFY 여부. Requirement 수준 수정이므로 WORKFLOW.md §142에
  따라 gstack 리뷰 재실행이 필요하다.

## I2 — `RefreshSeconds`는 하나가 아니라 둘이다

`positionsPage`와 `overviewPage`가 각각 같은 식으로 `holdingsTTL`에서 파생한다
(`portfolio_pages.go:57`, `overview.go:1164`). 제안 초안은 후자만 알고 있었다.
두 화면 모두 대상이며 tasks 4.3이 둘 다 덮는다.

브로커 관계는 서로 다르다 — `/positions`는 `get`, `/dashboard`는 `peek`
(design D2). 갱신 주기는 같아져도 이 차이는 유지되어야 한다.
