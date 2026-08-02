# Change: a055-console-settings-cadence

## Story Mapping

- Story: `STORY-TOS-a055`
- Feature: `FEAT-TOS-004`
- Story ↔ OpenSpec: 같은 `a055` 번호로 1:1 연결

## Why

사용자 요청(2026-08-02): **설정 화면이 불편하다.** StockOS는 같은 일을 훨씬 쉽게 한다.

두 제품을 직접 대조했다. StockOS 설정이 편한 이유는 화면이 예뻐서가 아니라 **분류 축이
다르기 때문**이다.

| 축 | StockOS | TossOS |
|---|---|---|
| 설정 분류 | **가역성 × 빈도** 4탭 — 상시(비가역·주 1회 미만) / 당일(가역·일 단위) / 전략 / 도구 | 기능명 1페이지 4섹션 |
| 탭 진입 | 각 탭 헤더에 eyebrow + 제목 + **한 줄** 설명 + 현재 운영 상태 칩 | 화면 첫 문단이 10줄 |
| 카드 헤더 | `현재값 → 변경값` | 없음(한 곳만 `(현재 X)` 인라인) |
| 저장 전 | 적용 후 값 미리보기 `<dl>` | 없음 |
| 저장 불가 | **비활성 사유 칩 8종**(코드화·테스트됨) | 없음 — 문단으로 설명 |
| 저장 결과 | 카드별 배너 | 전역 `<p class="notice">` **한 줄** |

측정값도 같은 방향이다.

| 사실 | 측정 |
|---|---|
| 설정 화면 설명문 비율 | **79%** (4,045 / 5,113자) — 8개 화면 중 최악 |
| 콘솔 전체 설명문 | 14,736자 |
| 한 페이지의 폼 컨트롤 | **31개**, 저장 버튼 4개, 결과 표시 1줄 |
| nav 항목 | 평면 **12개**, 설명 0개 |
| 내부 화면에서만 도달 가능한 화면 | **3개** — `/strategy-runtime`, `/strategy-runtime/market-schedule`, `/performance-history` |

그리고 설정 화면 자신이 문제를 자백하고 있다:

> "이 편입 섹션의 저장으로는 automation gate의 ON/OFF와 kill switch를 **콘솔에서 편집할 수
> 없다** — 아래 운영 설정은 별도 승인 표면이다."
> ([templates_settings.go:20-21](../../../internal/console/templates_settings.go#L20-L21))

권한이 다른 세 종류의 변경이 한 페이지에 있어서 **문장으로 경계를 긋고 있다.** 구조로 풀
문제를 산문으로 때우는 것이고, 그 산문이 79%다.

nav 라벨 "외부 종목 자동관리"도 같은 결함의 표면이다 — `/settings#adoption`을 가리키는데
그 페이지에는 Guardian 한도·운영·시스템 업데이트가 더 있다. 라벨이 한 섹션 이름이라
나머지 셋은 nav에서 보이지 않는다.

## What Changes

- **nav 12항목 → 운영 파이프라인 6항목**, 각 항목에 한 줄 설명. `개요 · 신호 · 주문 ·
  포지션 · 이력 · 설정`. 고아 화면 3개가 설정 하위에서 정규 진입점을 얻는다.
- **설정 = 가역성 × 빈도 4탭.** `/settings/standing`(비가역·주 1회 미만),
  `/settings/daily`(가역·일 단위, 기본 진입), `/settings/strategy`, `/settings/tools`.
- **전략·도구 탭은 화면을 흡수하지 않고 진입점만 제공한다** — `/optimization?category=…`,
  `/position-management`, `/strategy-runtime`, `/verify`, `/report`는 자기 경로에 그대로 있다.
  a050이 정한 canonical category deep link를 깨지 않기 위한 결정이다.
- **설정 카드 표준**: 헤더에 `현재 → 변경`, 적용 후 미리보기, 저장 불가 시 **이름 붙은 사유**,
  폼별 결과 표시, Guardian 한도의 **완화/강화 방향 구분**.
- **사유 입력을 필수로 만들지 않는다.** 설정 저장은 이미 서버에서 audit된다
  ([settings.go:223](../../../internal/console/settings.go#L223)) — 추적성은 충족돼 있고,
  타이핑 마찰만 늘어난다. StockOS의 typed-confirmation도 가져오지 않는다.
- **설명은 상태와 분리해 접는다.** 지금 무엇이 참인가와 누르면 무엇이 나가는가는 항상 보이고,
  왜 이렇게 설계했는가와 출처·전례는 접힌다. `<details class="explain">`는 CSS에 이미 있고
  실사용은 1회다.
- **접힘 상태의 URL 표현은 자동 재로드가 걸리는 화면에만 적용한다** — 설정 화면에는
  `Refresh`도 `RefreshSeconds()`도 **없다.** 산문의 79%가 있는 화면에는 애초에 문제가 없으므로
  native `<details>`로 충분하다. 자동 재로드 화면에서는 native toggle과 URL 상태가 어긋나지
  않도록 링크로만 여닫는다.
- **접지 않는 것을 클래스로 고정한다** — 여덟 항목을 문구로 열거하고 문자열 매칭하는 대신
  `.notice`/`.danger`를 달고 **그 클래스가 `<details>` 안에 없을 것**을 검사한다. 문구 매칭은
  한 글자만 바뀌어도 침묵으로 통과한다.
- **문체 통일** — 최적화 화면만 존댓말 30회, 나머지 7화면은 해라체.

## Capabilities

### Modified Capabilities

- `operator-console`: 설정은 기능명이 아니라 가역성과 빈도로 분류되고, navigation은 운영
  흐름을 따르며, 모든 설정 폼이 변경 전후와 차단 사유를 같은 형식으로 표시한다.

## Impact

- **Affected specs**: `operator-console` — `편입 설정 화면` **MODIFIED**(nav 문장 1개);
  `콘솔 내비게이션은 운영 흐름을 따른다`, `설정은 가역성과 빈도로 분류된다`,
  `설정 폼은 변경 전후와 차단 사유를 표시한다`, `설명은 상태와 분리해 접는다` **ADDED**.
- **Affected code**: `internal/console` — `templates.go`(nav), `templates_settings.go`,
  `settings.go`·`settings_limits.go`·`settings_operating.go`·`settings_update.go`의 렌더 경로,
  `console.go` 라우트(설정 하위 4건 신설, 기존 POST 경로 무변경), 나머지 화면 템플릿의 산문.
- **Affected code(무변경)**: 설정 저장 핸들러의 검증·기록 로직. 이 change는 **표시와 배치**를
  바꾸고 저장 의미를 바꾸지 않는다. `/settings/save`·`/settings/limits`·`/settings/trading`·
  `/settings/gate`·`/settings/autostart`·`/settings/system-update/*`의 POST 계약은 그대로다.
- **Data/operations**: 신규 브로커 호출 0건. 신규 config 키 0개. audit 기록 형식 무변경.
- **Safety invariants**: 콘솔의 상태변경 행위 목록은 **한 건도 늘지 않는다** — 신설 라우트 4건은
  전부 GET이다. automation gate·Guardian 한도·kill switch의 편집 권한 경계는 무변경. 한도
  완화는 **표시로 구분할 뿐 차단하지 않는다**(기존 승인 경로 유지). 손절·익절·사이징 로직 무변경.
- **선행 조건(안전)**: `a054`의 "승인 대기 중인 검증 run을 상태 표시줄이 표시하고 직접
  링크한다" 요구사항이 **이 change가 검증 콘솔을 도구 탭 진입점으로 옮기는 것의 전제**다.
  승인 창은 짧고 소진 사고 기록이 있다(M11·M18·M22·M23) — 그 보상 없이 발견성만 2클릭 깊게
  하는 것은 UI 정리가 아니라 사고 확률 증가다.

## Risks

| 위험 | 완화 |
|---|---|
| `편입 설정 화면`을 MODIFY하는 미아카이브 change가 3건인데 **그 셋 중 어느 것도 현재 본문의 nav 문장을 담고 있지 않다** — 지금 아카이브하면 본문이 되돌아간다 | 이 delta의 base는 **현재 승인된 본문**임을 delta 머리에 명시하고, 세 change의 아카이브 순서 충돌을 `issues.md`에 기록해 리뷰가 판단하게 한다 |
| 설정을 4탭으로 쪼개면 한 화면에서 보이던 상호작용(편입 규칙 ↔ Guardian 한도)이 안 보인다 | 각 탭 헤더에 현재 운영 상태 칩(게이트·엔진·kill switch)을 고정 표시하고, 다른 탭의 값에 의존하는 설정은 그 값과 링크를 함께 표시 |
| 전략·도구 탭을 "진입점"으로만 두면 탭이 링크 목록이 되어 또 하나의 빈 화면이 된다 | 각 링크에 현재 desired/effective 요약 한 줄을 붙인다. 요약은 해당 화면이 이미 계산한 값을 옮기며 재계산하지 않는다 |
| 산문 79% → 25%를 목표로 잡으면 안전 문구가 잘린다 | 접지 않는 문구를 spec에 **목록으로 고정**하고, 그 문구들이 접힌 요소 안에 들어가지 않음을 정적 검사로 확인 |
| `?explain=` 쿼리가 늘어나 캐시·북마크·핸드오프 URL과 섞인다 | `explain`은 표시 전용 파라미터로 한정하고 서버 판정에 쓰지 않는다. 알 수 없는 값은 무시하고 접힌 상태로 렌더 |
| nav 12→6이 기존 북마크와 문서 링크를 깬다 | 기존 경로는 전부 살아 있다 — nav에서 빠지는 것이지 라우트가 사라지는 것이 아니다. `/settings#adoption`도 새 탭 앵커로 리다이렉트 |
| 완화/강화 방향 표시가 판정을 새로 계산해 화면이 정책을 재구현한다 | 방향은 제출값과 현재값의 대소 비교뿐이다. 허용 여부 판정은 서버의 기존 검증이 소유하며 화면은 그 결과를 옮긴다 |
| 이 change와 `a054`가 같은 템플릿 파일을 만진다 | `a054` 완료 후 착수한다. `a054`는 chrome·프리미티브·경로, 이 change는 nav·설정·산문 — 두 change의 표면이 겹치는 파일은 `templates.go`이며 겹치는 블록은 nav 하나다 |
