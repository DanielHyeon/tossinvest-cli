# a176 design

전제: a175 착지 뒤(base `0e8ba1c7c`). 파일: `templates_overview.go`, `templates.go` 의 `verify` 템플릿(HEAD `4e20c5df5` 에서 :802), `pages.go` 는 편집하지 않는다. file:line 은 정찰(2026-10-25) 값.
**순서 정정(Manager, 2026-10-26)**: "a174·a175 와 병행 가능"을 거둔다. a174 가 같은 `templates.go` 의 `dashboard` define 과 `style` 을 고치고, a175 도 `style` 을 고친다. 그래서 a176 은 **a174 착지 뒤** 착수한다. CSS 는 `style` **끝에** `/* a176 — … */` 블록 하나만 덧붙인다(a174 design §2.5). 착수 시 `base-commit.txt` 를 그때 main HEAD 로 갱신하고, 작업 위치는 `/mnt/D/Axipient/workspace/TossOS.worktrees/a176-overview-summary-and-verify-steps`(사람 결정 2026-10-26).

## 1. 지형 (Manager 재확인, HEAD `0e8ba1c7c` 2026-10-27 — 정의본 대비 바뀐 곳은 ★)
| 대상 | 위치 | 비고 |
|---|---|---|
| 개요 | `/dashboard`(`chrome.go:18` `pathOverview`) → `templates_overview.go` `overview` define(:47~). ★ 요약 `dl.console-summary`(define 안 :27, 시장별 `summary-line`)가 **이미 있다**. 상세 절: 보유 종목·엔진·계좌 보유·오늘·살아 있는 주문·안전·Guardian 한도·오늘 소진분·최근 exit | a173 이 `causes`·`unmeasured`·기술 정보를 넣었다. spec 704~718·1372~1390 |
| ★ 구조적 미측정 설명 | 소진분 절(define 안 :238-257)이 **이미** 첫 설명만 `p.aside` 로, 뒤의 같은 사유는 "구조적으로 미측정 — 위와 같은 …" 한 줄로 그린다(`$prevWhy`). a173 시험 `TestStructuralUnmeasuredSentenceIsSaidOnce`(`a173_common_cleanup_test.go:320`)가 고정 | 정의본의 "같은 글 3회" 는 a173 이전 측정. 1.1 에서 렌더로 다시 센다 — 이미 1회면 2.2 는 회귀 가드로 바뀐다 |
| 개요 뷰 모델 | `overview.go` `overviewPage`(:1149), `handleOverview`(:1182), 패널 함수들 | 핸들러·패널 함수 본문 편집 0 목표 |
| ★ 검증 화면 | `templates.go` `verify` define(:838~). 단계 목록 절 :990-994 `<pre>{{.Steps}}</pre>`. 승인 프롬프트 `<pre>{{$.Prompt}}</pre>`(:875), 진행 줄 `<pre>`(:894) 는 **건드리지 않는다** | spec 1326·1344~1370 |
| ★ 단계 데이터 | **문자열이다.** `pages.go:180-188` `renderVerify` 가 `verifylive.WriteSteps(&steps, false, false)` 출력 전체를 `verifyPage.Steps string`(:137)에 담는다. `WriteSteps`(`internal/verifylive/report.go:416`)는 머리말(승인·그 외 없음·노출) 뒤에 `verifylive.Steps()` 순서로 단계마다 `"\n%-22s %s%s\n"`(ID·라벨·`[mutating, 보유 필요, 옵트인 …, deferred]` 태그) 머리 줄 + task·증명·절차·전송 줄을 쓴다. CLI `cmd/tossctl/verify.go:280` 도 같은 함수 | 구조체 `verifylive.Step`(`verifylive.go:118`): ID·Title·Proves·Tasks·Mutates·Mutations·NeedsHolding·OptIn·Deferred … |
| 시험 | `overview_test.go`, `screen_paths_test.go`, a173 계약 시험, verify 관련 시험(확인 문자열·CSRF·시장 선택) | |

## 2. 설계
### 2.1 개요 요약 세 묶음 (여섯 사실 유지)
- 상단 요약을 세 카드로: **엔진·보호**(① 실행 여부+기동 거부 사유, ⑤ 중 안전 상태) / **오늘·한도**(③ 시장별 실현손익·왕복·승패, ⑤ Guardian 한도 중 읽을 수 있는 값, 소진분 한 축) / **예외**(② 보유 종목 수·관리 외, ④ 살아 있는 주문, ⑥ 최근 exit, ⑤ 잔여물). 각 칸은 사실 1 + 상세 링크(spec 704). 재계산 0 — 기존 패널 값 그대로.
- 1280 × 900 에서 요약 세 카드가 첫 화면 안(≤ 900px). 상세 섹션은 그대로 아래에(spec 704 "제거되지 않는다").
- "구조적으로 미측정 — …" 설명: ★ a173 이 이미 1회로 줄였다(§1). 1.1 렌더 재측정에서 1회가 확인되면 손대지 않고 2.2 를 회귀 가드로 쓴다. 남은 반복이 있으면 기존 `$prevWhy` 방식을 그대로 쓰고 새 사유 코드는 만들지 않는다.
- 검증 기록 경로는 a173 의 `details.tech`.
### 2.2 `/verify` 단계 목록 구조화 (★ Manager 결정 2026-10-27)
- **verifylive 패키지와 `renderVerify`·핸들러는 편집하지 않는다.** 실계좌 검증 패키지라 위험 범위를 넓히지 않는다. 태그(mutating 등)를 콘솔에서 다시 계산하지도 않는다 — CLI 와 어긋날 자리를 만들지 않으려는 것이다.
- 새 파일의 새 메서드 `(p verifyPage) StepItems() []verifyStepItem{ID, Header, Body}` 가 `p.Steps` 문자열을 **알려진 단계 ID 경계에서 나눈다**: `verifylive.Steps()` 순서대로 각 단계의 머리 줄 시작(`"\n" + fmt.Sprintf("%-22s ", s.ID)`)을 찾아 자른다. 첫 조각은 머리말(Preamble)이다. `Header` 는 단계 조각의 첫 줄(ID·라벨·태그 그대로), `Body` 는 나머지.
- 불변식(시험으로 고정): ① 머리말 + 모든 조각을 이으면 `p.Steps` 와 **바이트 단위로 같다**, ② 조각 수 == `len(verifylive.Steps())`, ③ 각 조각은 자기 ID 로 시작한다, ④ 경계를 하나라도 못 찾으면(형식이 바뀌면) `StepItems` 는 nil 을 돌려주고 템플릿은 **지금처럼 `<pre>{{.Steps}}</pre>` 전체**를 그린다(fail-safe).
- 렌더: 머리말은 `<pre>` 그대로. 단계마다 `<details class="step"><summary><code>Header</code></summary><pre>Body</pre></details>`. mutating 표시는 Header 안의 `[mutating …]` 글자가 이미 하고, 필요하면 Header 에 `mutating` 이 있을 때만 `status-pill` 배지를 덧붙인다(판정이 아니라 같은 글자의 강조).
- 시작·승인·중단 폼·CSRF·확인 문자열 절차·시장 선택 링크·승인 프롬프트 `<pre>` 는 바이트 불변(렌더 비교 시험).
- 접힌 상태에서도 14 단계 머리 줄이 한 화면에서 훑어진다. 새 `details` 는 모바일 `summary { display: list-item }` 복원을 a176 CSS 블록에 넣는다.

## 3. 측정
a173 하네스. 기준선: `/dashboard` 4458px(a173 after), `/verify` 5188px. 목표: 개요 요약 ≤ 900px(요약 영역 하단 좌표), `/verify` 접힌 상태 ≤ 1800px, 설명 중복 0, 폼 바이트 동일.

## 4. 안전
`/verify` 는 승인 창이 있는 화면: 폼·버튼·확인 문자열·CSRF 는 건드리지 않는다. 개요는 폼 0(spec 1374). 브로커 호출·새 상수 0.
