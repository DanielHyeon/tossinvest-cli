# Design: a055-console-settings-cadence

## 1. 분류 축 — 왜 기능명이 아니라 가역성·빈도인가

기능명으로 나누면 운영자는 매번 "이건 어디 있더라"를 다시 푼다. 가역성과 빈도로 나누면
질문이 "이걸 되돌릴 수 있나 / 오늘 만지는 건가"로 바뀌고, 그 답은 운영자가 이미 알고 있다.

StockOS는 이 축을 5-rule classifier로 명시하고 상시 탭의 접근 부담을 **의도적으로 높여**
오클릭을 막는다([TradingRuntimeSettingsPage.tsx:54-71](../../../../stockos/apps/web/src/components/settings/TradingRuntimeSettingsPage.tsx)).
TossOS에도 같은 축이 이미 암묵적으로 있다 — 게이트·자동시작은 사람 승인이 필요하고
(§0.7), 한도는 프리셋으로 하루 단위 조정이 전제다.

## 2. 무엇이 어느 탭에 가는가 — 실제 경로 기준

현재 등록된 라우트를 직접 확인하고 배치했다([console.go:710-754](../../../internal/console/console.go#L710-L754)).

| 탭 | 성격 | 내용 | 현재 위치 |
|---|---|---|---|
| **상시** `/settings/standing` | 비가역 · 주 1회 미만 · 사람 승인 | 자동화 게이트 · 자동 시작 · 외부 종목 자동편입 규칙 · 시장·일정 | `#operating`의 게이트/autostart, `#adoption`, `/strategy-runtime/market-schedule` |
| **당일** `/settings/daily` (기본 진입) | 가역 · 일 단위 | Guardian 한도 + 프리셋 · 거래 정책 4토글 | `#guardian-limits`, `#operating`의 거래 정책 |
| **전략** `/settings/strategy` | 규칙 자체 | 최적화 · 포지션 정책 · 전략 lane **진입점** | `/optimization`, `/position-management`, `/strategy-runtime` |
| **도구** `/settings/tools` | 진단 · 드물게 | 검증 콘솔 · 검증 · 리포트 · 시스템 업데이트 **진입점** | 검증 콘솔 경로, `/verify`, `/report`, `#system-update` |

**전략·도구는 흡수가 아니라 진입점이다.** `/optimization?category=exit-protection` 같은
canonical deep link는 `거래 화면과 설정 화면의 역할은 분리된다`(스펙 250행)와
`종목별 정책 쓰기…`(265행)가 a050 계약으로 고정하고 있다. 그 화면들을 설정 하위로 **옮기면**
그 요구사항 전부를 MODIFY해야 하고, 미아카이브 delta 스택 위에서 그렇게 하는 것은
이 change가 감당할 위험이 아니다.

대신 각 진입점 링크에 **현재 desired/effective 요약 한 줄**을 붙인다. 요약은 그 화면이 이미
계산한 값을 옮기며 재계산하지 않는다 — 링크만 있는 탭은 또 하나의 빈 화면이다.

기존 POST 경로 8개(`/settings/save`, `/settings/include`, `/settings/exclude`,
`/settings/limits`, `/settings/limits/preset`, `/settings/trading`, `/settings/gate`,
`/settings/autostart`, `/settings/system-update/*`)는 **그대로 둔다.** 이 change는 폼이
어느 화면에서 렌더되는지를 바꾸고 폼이 어디로 제출되는지는 바꾸지 않는다.

`/settings`는 `/settings/daily`로 리다이렉트하고, `/settings#adoption`은
`/settings/standing#adoption`으로 리다이렉트한다 — 기존 nav·문서·북마크가 살아 있다.

## 3. 차단 사유는 발명하지 않는다 — 이미 코드에 있다

StockOS의 8코드를 옮겨오지 않는다. TossOS의 설정 템플릿은 **자기 게이트를 이미 전부 알고
있고, 그것을 문단으로 쓰고 있다.** 할 일은 그 문단을 구조로 바꾸는 것이다.

| 사유 | 현재 근거 | 저장 |
|---|---|---|
| 저장 seam 미배선 | `not .Wired`([:25](../../../internal/console/templates_settings.go#L25)) · `not .LimitsWired`([:102](../../../internal/console/templates_settings.go#L102)) · `not .TradingWired`([:161](../../../internal/console/templates_settings.go#L161)) | 불가 |
| 설정 파일 읽기 실패 | `.LoadErr`([:29](../../../internal/console/templates_settings.go#L29)) · `.LimitsLoadErr`([:74](../../../internal/console/templates_settings.go#L74)) · `.TradingLoadErr`([:159](../../../internal/console/templates_settings.go#L159)) | 불가 |
| 엔진이 거부할 블록 | `.Verdict`([:30](../../../internal/console/templates_settings.go#L30)) · `.LimitVerdict`([:89](../../../internal/console/templates_settings.go#L89)) | 가능하나 경고 |
| 한도 미설정 / 부분 설정 | `.LimitsUnset`([:83](../../../internal/console/templates_settings.go#L83)) · `.LimitsPartlyConfigured`([:87](../../../internal/console/templates_settings.go#L87)) | 가능하나 경고 |
| 손절폭 단위 grid 불일치 | `.NeedsStopPctCorrection`([:46](../../../internal/console/templates_settings.go#L46)) | 불가 |
| 지금 켜면 기동 거부 | `.GateBlockers`([:196](../../../internal/console/templates_settings.go#L196)) | 가능하나 경고 |
| 엔진 실행 중 — 다음 기동부터 반영 | `.EngineRunning`([:58](../../../internal/console/templates_settings.go#L58), [:145](../../../internal/console/templates_settings.go#L145)) | 가능, 주의 |

새 판정을 만들지 않는다. **표시 형식만 통일한다.**

## 4. 카드 표준

```
┌──────────────────────────────────────────────────────┐
│ 일일 손실 한도                     500,000 → 300,000  │  헤더: 현재 → 변경
├──────────────────────────────────────────────────────┤
│  금액 [ 300,000 ] 원                                  │
│                                                       │
│  적용 후                                               │  미리보기 dl
│    일일 손실 한도   500,000원 → 300,000원 (강화)        │
│    반영 시점        다음 엔진 기동                       │
│                                                       │
│  [ 한도 저장 ]  ⓘ 엔진 실행 중 — 기동 시 반영            │  사유 칩
│  ▸ 이 값이 무엇을 막는가                                │  details, 접힘
└──────────────────────────────────────────────────────┘
```

씨앗은 이미 있다 — `(현재 {{.StopPctPercent}})`([templates_settings.go:44](../../../internal/console/templates_settings.go#L44))가
헤더 패턴의 원형이고, `.LimitRows` 표([:76-81](../../../internal/console/templates_settings.go#L76-L81))가
미리보기의 원형이다. 전 카드로 확대하는 것이다.

**완화/강화 방향**은 같은 통화 안에서의 대소 비교다. 허용 여부 판정은 서버의 기존 검증이
소유한다 — "등록된 티어 위로는 올릴 수 없다"([:108](../../../internal/console/templates_settings.go#L108))는
서버가 거부하는 것이고 화면이 재구현할 것이 아니다. 화면은 **방향을 말하고 차단하지
않는다**. 안전 불변식 §9(보수 방향만)를 UI가 눈으로 돕되 승인 경로를 바꾸지 않는다.

**통화가 바뀌면 대소 비교가 무너진다.** 한도에는 `limit_currency`가 있고 KRW→USD로 바꾸면
`500,000 → 3,000`처럼 숫자가 작아지면서 실제로는 전혀 다른 축의 변경이 된다. 대소 비교는
이것을 "강화"로 표시한다 — 거짓이다.

통화 변경은 **제3의 축**으로 표시하고, 현재 템플릿이 이미 말하는 실제 귀결을 함께
낸다: "USD 한도를 기록하면 국내 자동 진입이, KRW 한도를 기록하면 미국 자동 진입이 닫힌다"
([templates_settings.go:142-144](../../../internal/console/templates_settings.go#L142-L144)).

### 죽어 있는 확인 대화 하나

[templates_settings.go:111](../../../internal/console/templates_settings.go#L111)의 프리셋 폼은
`onsubmit="return confirm(...)"`을 갖고 있다. 배포 CSP는 `default-src 'none'`이고 `script-src`가
없으므로 **인라인 핸들러는 실행되지 않는다** — 이 confirm은 한 번도 뜬 적이 없다. 미리보기
`<dl>`이 그 자리를 대신한다. 인라인 핸들러는 제거한다(기존 `편입 설정 화면` 요구사항이
"범위 밖"으로 남겨둔 legacy handler가 이것이다).

## 5. 산문 규칙과 접지 않는 목록

**규칙**: *지금 무엇이 참인가* / *누르면 무엇이 나가는가* → 항상 보인다.
*왜 이렇게 설계했는가* / *출처·전례·경계 설명* → `<details class="explain">`.

`.explain`은 스타일시트에 이미 있고([templates.go:72-73](../../../internal/console/templates.go#L72-L73))
실사용은 `templates_portfolio.go` 한 곳뿐이다.

**절대 접지 않는다**:

| # | 항목 | 현재 클래스 |
|---|---|---|
| 1 | 실계좌에 요청이 나간다는 경고 | `.danger` |
| 2 | 저장 seam 미배선 경고 | `.notice` |
| 3 | 설정 파일 읽기 실패 | `.danger` |
| 4 | 엔진이 지금 거부한다는 판정과 미충족 항목 | `.danger` |
| 5 | 잔여물·캐시 stale·갱신 보류 안내 | `.notice` |
| 6 | 반영 시점("다음 엔진 기동부터") | `.muted` → **`.notice`로 승격** |
| 7 | 한도 통화가 반대편 시장을 닫는다는 귀결 | `.muted` → **`.notice`로 승격** |
| 8 | 사전 판정 통과가 기동을 보장하지 않는다는 경계 | `.muted` → **`.notice`로 승격** |

**검사는 문구가 아니라 클래스 위치로 한다.** 초안은 여덟 문구를 산문으로 열거하고 문자열
매칭을 전제했다. 한국어 문장 조각 매칭은 문구가 한 글자만 바뀌어도 침묵으로 통과한다 —
검사가 죽어도 알 수 없는 종류의 검사다.

규칙: **`.notice`와 `.danger`를 가진 요소는 `<details>` 안에 나타나지 않는다.** 기계적이고,
문구 변경에 견디고, 새 경고가 추가돼도 자동으로 보호된다.

## 6. 접힘 상태와 자동 재로드 — 필요한 화면에만

자동 재로드는 유지한다(사용자 결정). 그런데 **설정 화면에는 `Refresh`도 `RefreshSeconds()`도
없다** — 자동 재로드가 걸리지 않는다. 콘솔 산문의 79%가 설정 화면에 있으므로, 접힘 상태
유실 문제는 **정작 접기가 가장 필요한 화면에는 존재하지 않는다.**

초안 설계에는 모순도 있었다. 접힘 상태를 URL로 표현하면 운영자가 `<details>`의 native
삼각형을 클릭했을 때 URL이 바뀌지 않아 다음 재로드에 닫힌다. **열리는 것처럼 보이다가
닫히는 것은 접기가 없는 것보다 나쁘다.**

따라서 화면을 둘로 나눈다.

| 화면 | 자동 재로드 | 접기 방식 |
|---|---|---|
| 설정 4탭 · 거래 이력 · 리포트 | 없음 | native `<details>`. URL 상태 불필요 |
| 개요 · 검증 콘솔 · 검증 · 포지션 · 주문 · 신호 | 있음 | `?explain=<id>` 링크로만 여닫는다 — native toggle을 제공하지 않아 URL과 화면이 어긋나지 않는다 |

`explain` 제약 셋:

- **표시 전용**이다. 서버 판정·저장·audit에 쓰지 않는다.
- 알 수 없는 값은 무시하고 전부 접힌 상태로 렌더한다(오류 아님).
- 스크롤 위치 보존은 브라우저 기본 동작에 맡긴다 — 이 change가 보장하지 않으며 실기기
  확인 항목으로 남긴다.

## 7. nav 6항목

| 라벨 | 설명 | 경로 |
|---|---|---|
| 개요 | 지금 무엇이 참인가 | 개요 경로 |
| 신호 | 후보 종목과 차단 사유 | `/signals` |
| 주문 | 미체결·종결·브로커 응답 | `/orders` |
| 포지션 | 보유와 보호 상태 | `/positions` |
| 이력 | 거래 이력과 성과 이력 | `/history` |
| 설정 | 상시 · 당일 · 전략 · 도구 | `/settings/daily` |

순서는 엔진의 실제 흐름이다 — 발굴 → 발주 → 보유 → 종결. 학습 가능한 순서가 곧 기억할
필요가 없는 순서다.

`/performance-history`는 이력 화면의 두 번째 탭이 되고, `/strategy-runtime`과
`/strategy-runtime/market-schedule`은 설정 하위에서 진입점을 얻는다. 세 화면 모두 **라우트는
그대로**다 — nav에서 빠지는 것이 아니라 nav에 처음으로 들어오는 것이다.

## 8. 무엇을 하지 않는가

- **필수 사유 입력을 넣지 않는다.** 설정 저장은 이미 서버에서 audit된다
  ([settings.go:223](../../../internal/console/settings.go#L223)) — 안전 불변식 §5의 추적성은
  충족돼 있고, 입력을 강제하면 마찰만 는다.
- **typed-confirmation을 넣지 않는다.** 콘솔 UI에 타이핑 확인·추가 승인 마찰을 넣지 않는다는
  사용자 지시가 있고, `docs/stockos-inventory.md`의 이식 판정도 "이식 안 함"이다.
  `종목 관리 설정은 기본값과 동작을 가까이 설명한다`(스펙 296행)의 기존 SHALL과도 일치한다.
- **저장 의미를 바꾸지 않는다.** POST 경로·검증·audit 형식 전부 무변경.
- **게이트·한도·kill switch의 편집 권한 경계를 바꾸지 않는다.**
- **한도 완화를 차단하지 않는다.** 방향을 표시할 뿐이다.
- **`/optimization` 계열을 설정 하위로 옮기지 않는다.** §2 참조.

## 9. 스펙 부채 — 이 delta의 base 선언

`편입 설정 화면`(스펙 155행)을 MODIFY하는 미아카이브 change가 3건 있고, **그 셋 중 어느
것도 현재 본문의 "외부 종목 자동관리" nav 문장을 담고 있지 않다.** MODIFIED가 블록 전체를
치환하므로, 그 셋을 지금 아카이브하면 본문이 되돌아간다.

이 delta의 base는 **현재 승인된 `openspec/specs/operator-console/spec.md`의 본문**이다
(WORKFLOW 권위 경계: 의도된 동작의 권위는 `openspec/specs/` + 승인된 change). 세 change와의
아카이브 순서 충돌은 `issues.md`에 기록하고 리뷰가 판단한다 — 이 change가 단독으로 결정할
문제가 아니다.
