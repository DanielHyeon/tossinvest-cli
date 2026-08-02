## 1. Contract and evidence

- [x] 1.1 `STORY-TOS-a055`과 `a055-console-settings-cadence`의 번호·범위가 1:1인지 검증한다.
- [x] 1.2 `a054-console-status-shell`이 완료·아카이브됐는지 확인한다. 겹치는 파일은 `templates.go`이고 겹치는 블록은 nav 하나다.
- [x] 1.2b **안전 선행 확인**: `a054`의 "승인 대기 중인 검증 run 표시 + 직접 링크"가 구현·검증됐는지 확인한다. 없으면 검증 콘솔을 도구 탭 진입점으로 옮기는 작업(3.x)을 착수하지 않는다.
- [x] 1.3 `python3 tools/sdd/capture_change_base.py --change a055-console-settings-cadence`로 구현 전 commit을 고정한다.
- [x] 1.4 `make sdd-sync` 후 설정 렌더 경로(`handleSettings` 및 view 조립 함수)의 definition/callers/impact를 확인한다.
- [x] 1.5 design §3 사유 표의 7개 게이트가 현재 HEAD에 전부 실재하는지 재확인한다. 없는 게이트를 spec에 쓰지 않는다.
- [x] 1.6 `편입 설정 화면`을 MODIFY하는 미아카이브 change 3건과 이 delta의 base 차이를 `issues.md`에 기록한다.
- [x] 1.7 기존 함수 편집 대상에 Function Logic Map과 Branch Test Map을 작성한다.

## 2. RED → GREEN → REFACTOR — navigation

- [x] 2.1 RED: nav가 6항목이고 각 항목이 라벨과 설명을 함께 렌더하며 현재 화면에 `aria-current`가 붙는다.
- [x] 2.2 RED: 어떤 화면도 다른 화면 내부에서만 도달 가능하지 않다 — `/strategy-runtime`, `/strategy-runtime/market-schedule`, `/performance-history`가 nav 경로로 도달된다(정적 검사).
- [x] 2.3 RED: 기존 경로가 전부 살아 있다. nav에서 빠진 화면도 직접 URL로 렌더된다.
- [x] 2.4 GREEN: nav 블록과 진입점 링크를 고친다.

## 3. RED → GREEN → REFACTOR — 설정 4탭

- [x] 3.1 RED: `/settings/standing|daily|strategy|tools` 넷이 렌더되고 전부 GET이며 라우트 표 정적 검사가 넷을 본다.
- [x] 3.2 RED: `/settings`가 `/settings/daily`로, `/settings#adoption`이 `/settings/standing#adoption`으로 리다이렉트한다.
- [x] 3.3 RED: design §2 배치대로 각 설정 **컨트롤**이 정확히 한 탭에만 나타난다. 다른 탭 값의 **읽기 전용 표시와 링크는 유지**된다 — 당일 탭이 게이트 상태를 보여주고 상시 탭으로 링크하되 스위치는 복제하지 않는다.
- [x] 3.4 RED: 각 탭 헤더가 제목·한 줄 설명·현재 운영 상태(게이트·엔진·한도 설정 여부)를 표시한다.
- [x] 3.5 RED: 전략·도구 탭의 각 진입점 링크가 현재 desired/effective 요약 한 줄을 함께 표시하고, 그 값을 재계산하지 않는다.
- [x] 3.6 RED: POST 경로 8건의 계약이 무변경이다 — 같은 action, 같은 필드명, 같은 CSRF 게이트, 같은 audit.
- [x] 3.7 GREEN: 라우트 4건과 템플릿 분할을 구현한다.

## 4. RED → GREEN → REFACTOR — 카드 표준

- [x] 4.1 RED: 각 설정 카드 헤더가 `현재 → 변경`을 표시한다. 변경이 없으면 현재값만 표시한다.
- [x] 4.2 RED: 적용 후 미리보기가 바뀌는 값과 반영 시점을 표시한다.
- [x] 4.3 RED: design §3의 7개 사유가 각각 발생하는 상태를 만들고, 저장 표면이 있어야 할 자리에 그 사유가 이름으로 표시되는지 확인한다. **저장 표면이 비활성인 경우와 아예 렌더되지 않는 경우 모두**를 덮는다 — 이 콘솔은 seam이 없으면 폼을 비활성화하지 않고 렌더하지 않으므로, `disabled` 속성만 찾는 검사는 0건을 찾고 통과한다.
- [x] 4.4 RED: 저장 결과가 해당 폼 옆에 표시된다 — 어느 폼이 저장됐는지 식별 가능하다.
- [x] 4.5 RED: **같은 통화 안에서** Guardian 한도의 완화와 강화가 다른 표시를 받고, 완화가 **차단되지는 않는다**.
- [x] 4.5b RED: 한도 **통화 변경**은 강화도 완화도 아닌 별개 축으로 표시되고, 어느 시장의 자동 진입이 닫히는지를 함께 말한다. 숫자 대소를 방향으로 해석하지 않는다(KRW 500,000 → USD 3,000이 "강화"로 표시되면 FAIL).
- [x] 4.6 RED: 렌더 결과에 `on[a-z]+=` 인라인 핸들러가 없다 — 프리셋 폼의 죽은 `onsubmit="return confirm(...)"` 제거를 포함한다.
- [x] 4.7 RED: 타이핑 확인·자유 문구 입력·필수 사유 입력이 없다.
- [x] 4.8 GREEN: 카드 표준을 설정 폼 전체에 적용한다.

## 5. RED → GREEN → REFACTOR — 산문과 접힘

- [x] 5.1 GREEN 선행: design §5 표의 6·7·8번(현재 `.muted`)을 `.notice`로 승격한다.
- [x] 5.2 RED: **`.notice`/`.danger`를 가진 요소가 `<details>` 안에 있으면 검사가 실패한다.** 문구 매칭이 아니라 클래스 위치로 판정한다 — 문구는 한 글자만 바뀌어도 침묵으로 통과한다.
- [x] 5.3 RED: 자동 재로드가 걸리는 화면에서 `?explain=<id>`가 해당 disclosure만 열고, 재로드 후에도 열린 채 렌더된다. 그 화면에는 URL을 바꾸지 않는 여닫기 수단이 없다.
- [x] 5.4 RED: **자동 재로드가 없는 화면(설정 4탭 등)은 native `<details>`로 렌더되고 `explain` 파라미터를 요구하지 않는다.**
- [x] 5.5 RED: 알 수 없는 `explain` 값은 무시되며 오류가 아니다. 파라미터가 저장·판정·audit 어디에도 도달하지 않는다.
- [x] 5.6 RED: 콘솔 전 화면의 문체가 하나다 — 존댓말과 해라체가 섞이지 않는다.
- [x] 5.7 GREEN: 8개 템플릿의 산문을 정리한다. 설정 화면 설명문 비율을 **정리 전과 같은 측정 방법으로** 재측정하고 결과를 `review.md`에 기록한다(비율 자체는 SHALL이 아니다 — 수치 목표는 안전 문구를 지우는 방향으로 게이밍된다).

## 6. Verification and completion

- [x] 6.1 변이 검증: `.notice` 요소 하나를 `<details>` 안으로 옮겨 5.2가 FAIL하는지, 사유 하나를 지워 4.3이 FAIL하는지, 컨트롤 하나를 두 탭에 넣어 3.3이 FAIL하는지, 통화 변경을 대소 비교로 판정하게 만들어 4.5b가 FAIL하는지 확인하고 전부 되돌린다.
- [x] 6.2 설정 4탭이 a052가 도입한 반응형 렌더 검사를 통과한다(레이아웃 실측은 자동 검사가 아니다 — a054 design §5b).
- [x] 6.3 CSP 회귀 검사와 응답 헤더 무변경을 확인한다.
- [x] 6.4 `openspec validate a055-console-settings-cadence --strict --no-interactive`를 통과한다.
- [x] 6.5 upstream 상속 테스트 650개 green 유지를 확인한다.
- [x] 6.6 `make sdd-check`와 독립 리뷰를 통과하고 `review.md`를 남긴다. `편입 설정 화면` MODIFIED가 있으므로 requirement 변경 리뷰 게이트를 적용한다.
- [x] 6.7 `make gate CHANGE=a055-console-settings-cadence`를 통과하고 PM Story를 동기화한다.
