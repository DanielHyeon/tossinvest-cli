# Review: a055-console-settings-cadence

- Date: 2026-08-02
- Scope: navigation 재편, 설정 4탭 분류, 설정 카드 표준, 산문·접힘 규칙
- Review class: UI change + **requirement 변경**(`편입 설정 화면` MODIFIED) — 경량 리뷰에
  requirement 변경 게이트를 더한다
- Voices: Manager 셀프리뷰, 적대적 Eng, DX(검사 가능성), QA/안전
- Status: **accepted with six corrections applied below**

## Findings

### F9. 접지 않는 문구 검사를 문자열 매칭으로 설계했다 — 수정

초안은 여덟 종류의 문구를 산문으로 열거하고 "그 문구가 `<details>` 안에 있으면 검사가
실패한다"고 요구했다. 한국어 문장 조각 매칭은 문구가 한 글자만 바뀌어도 침묵으로 통과한다 —
검사가 죽어도 알 수 없는 종류의 검사다.

코드에는 이미 정답이 있다. 이 여덟 종류는 대부분 `<p class="notice">` 또는
`<p class="danger">`로 렌더되고 있다.

- **수정**: 규칙을 클래스 기반으로 바꿨다 — 접지 않는 항목은 `.notice`/`.danger`를 달고,
  **그 두 클래스는 `<details>` 안에 나타나지 않는다**. 기계적으로 검사 가능하고 문구 변경에
  견딘다. 현재 `.muted`로 렌더되는 "반영 시점"은 `.notice`로 승격한다.

### F10. `?explain=` 요구가 정작 필요 없는 화면에 걸려 있었다 — 범위 축소

설정 화면에는 `Refresh` 필드도 `RefreshSeconds()`도 **없다** — 자동 재로드가 걸리지 않는다.
그런데 콘솔 산문의 79%가 설정 화면에 있다. 즉 접힘 상태 유실 문제는 **정작 접기가 가장
필요한 화면에는 존재하지 않는다.**

게다가 초안 설계에는 모순이 있었다. 접힘 상태를 URL로 표현하면 운영자가 `<details>`의 native
삼각형을 클릭했을 때 URL이 바뀌지 않아 다음 재로드에 닫힌다. **열리는 것처럼 보이다가
닫히는 것은 지금보다 나쁘다.**

- **수정**: URL 접힘 요구를 **자동 재로드가 걸리는 화면에만** 적용한다. 설정 화면은 native
  `<details>`로 충분하고 모순도 없다. 자동 재로드 화면에서는 disclosure를 링크로 여닫는
  형태로 한정해 native toggle과 URL 상태가 어긋나지 않게 한다.

### F11. 존재하지 않는 UI 패턴을 겨냥한 요구 — 수정

초안은 "사유 없이 **비활성인** 컨트롤이 있어서는 안 된다"고 썼다. TossOS 설정 화면에는
비활성 컨트롤이 거의 없다 — seam이 없으면 `{{if not .Wired}}…{{else}}<form>`으로 **폼을 아예
렌더하지 않는다**([templates_settings.go:25](../../../internal/console/templates_settings.go#L25),
[:102](../../../internal/console/templates_settings.go#L102),
[:161](../../../internal/console/templates_settings.go#L161)).

`disabled` 속성을 찾는 검사는 0건을 찾고 통과한다. 도달 불가능한 요구사항이다.

- **수정**: "저장 표면이 **없거나** 비활성이면 그 자리에 이름 붙은 사유가 렌더된다"로 바꿨다.
  부재를 포함해야 실제 패턴을 덮는다.

### F12. 완화/강화 판정이 통화 변경에서 무너진다 — 수정

초안은 "방향은 제출값과 현재값의 대소 비교뿐"이라고 못박았다. Guardian 한도에는
`limit_currency`가 있고 KRW→USD로 바꾸면 `500,000 → 3,000`처럼 숫자가 **작아지면서 실제로는
전혀 다른 축의 변경**이 된다. 대소 비교는 이것을 "강화"로 표시한다 — 거짓이다.

게다가 현재 템플릿이 이미 경고하는 실제 귀결이 있다: "USD 한도를 기록하면 국내 자동 진입이,
KRW 한도를 기록하면 미국 자동 진입이 닫힌다"([templates_settings.go:142-144](../../../internal/console/templates_settings.go#L142-L144)).

- **수정**: 통화 변경은 강화도 완화도 아닌 **제3의 축**으로 표시하고, 어느 시장의 진입이
  닫히는지를 함께 말하도록 요구사항에 넣었다.

### F13. "복제 금지"가 상태 표시까지 금지하는 것으로 읽힌다 — 수정

초안의 "한 설정 항목은 정확히 하나의 하위 화면에만 나타나야 한다"는 현재의 유용한 상호참조를
금지하는 것으로 읽힌다 — Guardian 한도 섹션이 게이트 ON/OFF **상태**를 보여주고 스위치는
운영 섹션에 있다고 안내하는 패턴([templates_settings.go:98-100](../../../internal/console/templates_settings.go#L98-L100)).
그 안내는 좋은 것이고 없애면 안 된다.

- **수정**: 금지 대상을 **컨트롤 복제**로 한정하고, 다른 탭 값의 **읽기 전용 표시와 링크는
  오히려 요구**하도록 문장을 나눴다.

### F14. 검증 콘솔을 도구 탭으로 옮기는 것의 승인 창 위험 — a052에 보상 요구 추가

검증 승인 창은 짧고 소진 사고 기록이 있다(M11·M18·M22·M23). 발견성을 2클릭 깊게 만드는 것은
UI 정리가 아니라 사고 확률 증가다.

- **수정**: `a054`에 "상태 표시줄이 승인 대기 중인 run을 표시하고 직접 링크한다" 요구사항을
  추가했고, 이 change의 **선행 조건**으로 명시했다. 표시줄은 모든 화면에 있으므로 순 발견성은
  이동 전보다 좋아진다. 이 보상 없이 이동만 하는 것은 승인하지 않는다.

## 수용하지 않은 지적

- *"산문 79% → 25% 같은 수치 목표를 SHALL로 박아라"* — 거부. 비율 SHALL은 안전 문구를 지우는
  방향으로 게이밍된다. 대신 접지 않는 항목을 클래스로 고정(F9)하고 비율은 tasks의 측정·기록
  항목으로 남긴다.
- *"필수 사유 입력을 넣어 audit을 강화하자"* — 거부. 설정 저장은 이미 서버에서
  audit된다([settings.go:223](../../../internal/console/settings.go#L223)). 안전 불변식 §5의
  추적성은 충족돼 있고, 입력 강제는 마찰만 늘린다. typed confirmation도 같은 이유로 거부.
- *"`/optimization` 계열을 설정 하위 경로로 옮기자"* — 거부. a050이 고정한 canonical category
  deep link를 참조하는 요구사항이 넷이고, 미아카이브 delta 스택 위에서 그것들을 MODIFY하는
  것은 이 change가 감당할 위험이 아니다(design §2).

## 미결 — 리뷰가 판단할 것 (issues.md I1)

`편입 설정 화면`을 MODIFY하는 미아카이브 change 3건 중 **어느 것도 현재 본문의
"외부 종목 자동관리" 문장을 담고 있지 않다.** 지금 그것들을 아카이브하면 스펙이 되돌아간다.
같은 부채가 `콘솔 안전 불변식`에도 있다(미아카이브 MODIFY **6건**).

이 change는 base를 현재 승인 본문으로 명시하고 `콘솔 안전 불변식`은 건드리지 않는 것으로
자기 노출을 최소화했다. **아카이브 순서 결정과 미아카이브 console 계열 change 정리는 이
change의 구현 착수를 막지 않지만, 아카이브 전에 결론이 필요하다.**

## Verification evidence

- `openspec validate a055-console-settings-cadence --strict --no-interactive` → valid.
- design §3의 차단 사유 7종이 전부 현재 HEAD에 실재함을 파일·행 단위로 확인.
- 설정 화면에 `Refresh`/`RefreshSeconds` 부재 확인(F10의 근거).
- `templates_settings.go:111`의 `onsubmit` 인라인 handler가 배포 CSP에서 실행되지 않음을
  확인 — `default-src 'none'`이고 `script-src` 지시가 없다(issues.md I2).
- 구현 착수 전이므로 테스트 실행 결과 없음. proposal-freeze 리뷰다.

## Function Logic Map

기존 함수 내부 편집이 있다 — `handleSettings`와 설정 view 조립 함수, nav 렌더 경로.
**not-applicable 아님.** tasks 1.7이 구현 전 작성을 요구한다.

## Verdict

여섯 개 수정을 반영한 상태로 proposal을 freeze한다. 선행 조건은 `a054-console-status-shell`의
완료이며, 그중 **승인 대기 표시 요구사항(F7/F14)은 이 change의 검증 콘솔 이동을 승인하는
전제**다. 저장 계약·게이트 권한 경계·LIVE 승인 규칙은 범위 밖이고 변경하지 않는다.

---

# 구현 후 추기 (2026-08-02)

구현·검증을 마친 뒤 계약과 실제 코드가 어긋난 지점, 측정 결과, 상속 테스트 조정 내역을
기록한다. 위 proposal-freeze 절은 그대로 두고 여기에만 덧붙인다.

## 결과 요약

| 항목 | 값 |
|---|---|
| `go test ./...` | **5906 passed / 78 packages** (착수 시점 5870, 신규 +36, 회귀 0) |
| `go vet ./...` | 이슈 없음 |
| `openspec validate a055 --strict` | valid |
| 변이 검증 | 4종 전부 RED → 되돌린 뒤 GREEN |
| 375px 실측 | 19화면 × 2폭 = **38측정, 넘침 0** |
| 설정 화면 설명문 비율 | **88.8% → 65.7%** (같은 방법으로 양쪽 측정, 아래 참조) |

## I1. 통화 변경을 대소 비교에서 제외하는 것은 코드에도 반영해야 했다

계약(§4.5b)은 "통화가 바뀌면 대소 비교가 무너진다"를 화면 표시 규칙으로 썼다. 구현에서
그것은 `previewAgainst`의 `comparable` 한 줄이 됐고, **변이 검증 M4**가 그 줄을 지웠을 때
`일일 손실 한도(비율)`이 "변경 없음"으로 표시됐다 — 비율은 통화와 무관하니 우연히 같았고,
그래서 "변경 없음"이라는 **거짓이 아닌 것처럼 보이는 라벨**이 붙었다. 통화가 바뀌는 순간
다섯 줄 전부에서 방향 라벨을 떼는 편이 옳다. 테스트는 다섯 줄 모두를 검사한다.

## I2. `previewAgainst`는 `config.AutomationGate`를 받으면 안 됐다

`engineproc_test.go:TestTheConsoleDecidesNothingAboutTheGate`는 `AutomationGate`라는 철자를
`settings.go`·`settings_limits.go`·`templates_settings.go` 셋에서만 허용한다. 신설
`settings_card.go`가 그 타입을 파라미터로 받자 가드가 걸렸다. **가드에 파일을 추가하지 않고**
`config.GuardianLimits`와 bool 둘로 쪼개 받았다 — 이 파일은 스위치와 무관해야 하고, 가드는
정확히 그 사실을 지키고 있었다.

## I3. 손절폭 grid 불일치는 **차단**이 아니라 **주의**다

design §3의 표는 이 사유의 저장 칸을 "불가"로 적었다. 그대로 `Blocks`에 넣자
`TestTheStopFractionIsASlider`의 legacy off-grid 케이스가 깨졌다 — 폼을 렌더하지 않으면
**운영자가 그 값을 고칠 수단이 화면에서 사라진다.** 표의 "불가"는 *현재 값을 그대로 저장할
수 없다*는 서버 판정이지 저장 표면의 부재가 아니다. `Cautions`로 옮겼고, 서버의 거부는
그대로다(침묵 반올림 없음).

## I4. 행 단위 증거는 `<details>`로 남겼다 — §5의 대상이 아니다

design §6의 표는 자동 재로드 화면의 disclosure를 전부 `?explain=` 링크로 바꾸라고 읽힌다.
`/positions`·`/orders`의 **행 상세**(`.row-details`)까지 바꿔 보고 되돌렸다. 근거 셋:

1. 요구사항의 주어는 **설명문**이다("콘솔 화면의 설명문은 두 종류로 나뉘고"). 행 상세는
   그 행의 현재 증거이지 설명이 아니다.
2. 파라미터가 하나이므로 B행을 열면 A행이 닫힌다. 두 주문을 비교하려는 운영자에게는
   재로드마다 닫히는 `<details>`보다 나쁘다.
3. 열지 않으면 증거가 **문서에서 사라진다** — Ctrl-F로 찾을 수 없고 스크린 리더의 읽기
   순서에도 없다. 상속 테스트 8건이 그 사실을 즉시 드러냈다.

바꾼 것은 설명 성격의 `.explain` 넷뿐이다(`/orders` 2, `/positions` 2). 테스트도 그 범위로
쓴다: "자동 재로드 화면에 `<details class="explain">`이 없다".

## I5. `관리 상태와 데이터 기준` 문단은 두 종류가 섞여 있었다

접으려 하자 상속 테스트 둘이 라벨 개수를 세다 실패했다. 읽어 보니 첫 문장은
"관리 외 = 손절·익절이 걸려 있지 않다"는 **지금 참인 사실**이고 — 계약이 *항상 보이라*고
분류한 쪽 — 나머지가 출처·경계다. 문단을 둘로 갈라 앞은 남기고 뒤만 접었다.

## I6. `/settings` 진입 앵커는 서버가 볼 수 없다

`/settings#adoption`의 fragment는 요청에 실리지 않는다. 리다이렉트는 fragment를 읽지 못하고
브라우저가 그것을 그대로 `/settings/daily#adoption`으로 가져간다 — 그 섹션이 없는 탭이다.
두 가지로 답한다: 서버가 볼 수 있는 `?section=adoption`은 상시 탭으로 303하고, 당일 탭에는
`id="adoption"` 앵커를 두어 상시 탭으로 보낸다. 둘 다 테스트한다.

## 산문 비율 재측정 (task 5.7)

proposal의 79%는 측정 방법이 기록되어 있지 않아 재현할 수 없다. **양쪽을 같은 방법으로
다시 쟀다** — base commit `b331f664`의 clean worktree와 현재 worktree에 같은 측정 코드를
넣고, 렌더된 `<main>` 안의 태그를 제거한 글자 수를 분모로, `<p>` 요소 안의 글자 수를 분자로
계산했다.

| | 화면 | 설명문 | 전체 | 비율 |
|---|---|---|---|---|
| before | `/settings` | 2,963 | 3,336 | **88.8%** |
| after | `/settings/standing` | 1,259 | 1,703 | 73.9% |
| after | `/settings/daily` | 1,524 | 2,366 | 64.4% |
| after | `/settings/strategy` | 123 | 363 | 33.9% |
| after | `/settings/tools` | 391 | 584 | 67.0% |
| after | 합계 | 3,297 | 5,016 | **65.7%** |

**비율보다 중요한 것은 한 번에 읽는 양이다.** 가장 긴 탭이 2,366자로, 이전 한 화면
3,336자보다 적다. 총량이 는 것(3,336 → 5,016)은 카드마다 적용 후 미리보기와 이름 붙은
사유가 새로 생겼기 때문이고, 그것이 이 change가 산문을 대체하려던 구조다.

비율 자체는 SHALL이 아니다 — 목표 수치를 세우면 안전 문구를 지우는 방향으로 게이밍된다.
접지 않는 여덟 종류는 클래스 위치 검사로 고정했다.

## 변이 검증 (task 6.1)

| 변이 | 기대 FAIL | 결과 |
|---|---|---|
| `.notice` 하나를 `<details>` 안으로 이동 | `TestNoWarningIsHiddenInsideADisclosure` | FAIL ✓ |
| `cardblocked`가 아무것도 렌더하지 않게 | `TestEveryCardEitherSavesOrSaysWhyNot` | FAIL ✓ |
| 게이트 스위치를 당일 탭에 복제 | `TestEachSettingControlAppearsOnExactlyOneTab` | FAIL ✓ |
| 통화 변경을 대소 비교로 판정 | `TestACurrencyChangeIsNeitherTighteningNorLoosening` | FAIL ✓ |

넷 다 되돌린 뒤 전체 수트 GREEN을 재확인했다.

중첩 인식 스캐너 자체도 측정한다(`TestTheNestingAwareScanFindsAWarningInsideAnInnerDisclosure`):
안쪽 `</details>`를 바깥의 끝으로 취급하는 depth 버그는 **이 콘솔이 실제로 쓰는 마크업에서
정확히 clean을 보고**하므로, 스캐너를 믿지 않고 다섯 케이스로 고정했다.

## 375px 실측 (task 6.2)

a054 design §5b가 "자동 검사로 대체 불가"로 남긴 항목. 19개 렌더 산출물(설정 4탭, 저장 결과가
붙은 당일 탭, `?explain=`이 열린 주문 화면 포함)을 loopback으로 띄우고 375px·1280px iframe에서
`documentElement.scrollWidth` vs `clientWidth`를 쟀다. **38측정 전부 `ok`, 넘침 0.**

## 상속 테스트 조정

| 테스트 | 조정 | 근거 |
|---|---|---|
| `settings_limits_test.go` 12곳 | `/settings` → `/settings/daily` | 한도는 당일 탭 소유 |
| `settings_test.go` 4곳, `settings_autostart_test.go` 3곳 | → `/settings/standing` | 편입·자동 시작은 상시 |
| `system_update_test.go`, `signed_release_test.go` | → `/settings/tools` | 업데이트는 도구 |
| `settings_operating_test.go` 7곳 | 거래 정책 → 당일, 게이트·사전 판정 → 상시 | 컨트롤 소유 탭 |
| `remote_csp_test.go` | → `/settings/standing` | adoption control을 검사한다 |
| `cmd/tossctl/adoptionsettings_stop_percent_test.go` | → `/settings/standing` | 같은 이유 |
| `TestThePresetControlsAskForNoTyping` | `onsubmit` **존재** 단언 → **부재** + 미리보기 단언 | issues.md I2 |
| `TestExternalPositionAutomaticManagementHasADiscoverableMenu` | 옛 nav 라벨 → 새 계약 | 이 change가 MODIFY한 요구사항 |
| `currentNavLabel` (a054 헬퍼) | 주 nav를 aria-label로 찾고 `<small>` 설명 제거 | 설정 탭 바도 `aria-current="page"`를 갖는다 — 문서 순서 규칙이었으면 탭 라벨과 제목을 비교했을 것 |
| `optimization_*_test.go` 3곳 | 합쇼체 문구 → 해라체 | §5.6 문체 통일 |
| `display_primitives_test.go` | 인라인 handler 예외 삭제 | 셋 다 제거됐으므로 예외가 없어졌다 |

## 아직 남은 것

`issues.md` I1의 아카이브 순서 결정은 그대로 미결이다. 이 change의 delta는 `편입 설정 화면`을
MODIFY하므로 **아카이브 전에** 미아카이브 console 계열 3건과의 순서를 정해야 한다. 구현은
막히지 않았다.
