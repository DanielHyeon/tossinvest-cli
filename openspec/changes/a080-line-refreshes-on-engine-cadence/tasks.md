# a080 · Tasks

> **[해제·8장 완료] 2026-08-05.** 선행 change a081이 `df4407ed`로 land했고 사람이
> 승인한 배포에서 실측까지 끝났다 — 렌더 20회 묶음에서 엔진에 닿은 렌더는 1회,
> 엔진 `reconcile` 사이클은 62.6~63.7초로 무변동. F1의 근거였던 결합이 사라졌으므로
> 이 change를 재개했다.
>
> 8장을 전부 실행했다: 코드 재적용(711 passed), F2(변이 M-F2 **RED 실측**),
> F6, F7, `base-commit.txt` → `30d8bb93`, logic-map target 4 → **5**
> (`evidence complete`), 두 번째 MODIFIED의 Requirement 변경 리뷰(수용).
> **F1~F8 전부 종결.** 기록은 review.md 4차.
>
> 남은 것은 6.4 gate와 6.5·6.6 컨테이너 실측이며, 실측은 이 코드가 배포된 뒤에만
> 할 수 있으므로 a081과 같은 순서(배포 → 실측 → gate)를 따른다.
>
> 아래 blocked 기록은 그때의 판단 근거로 남긴다.
>
> **[blocked → 선행 대기] 2026-08-05.** 독립 리뷰 F1이 이 change의 전제를
> 무효화했다 — `decoratePositionRows`가 렌더마다 엔진 프로세스로 무캐시 동기
> 읽기 2회를 하고, 그 `List`는 엔진의 **단일 쓰기 커넥션**에서 exit 루프의 판정
> 트랜잭션과 직렬화된다. 재로드를 6배로 만들면 손절 판정 간격에 그만큼 더해진다
> — 안전 불변식 4.
>
> **처분 (사용자 결정)**: 선행 change로 분리한다.
> `a081-screens-share-one-engine-reading`이 그 읽기를 TTL 캐시 뒤로 옮겨 엔진
> 도달 횟수를 렌더 횟수에서 분리한다. **a081은 land·배포 완료**(706 passed,
> 변이 검증 7건 RED, 독립 리뷰 3인 P0 0건).
>
> a080의 코드 변경은 되돌려 둔 상태다 — 두 change가 같은 파일을 고치므로 diff가
> 섞이면 logic-map gate가 서로를 상대 change의 편집으로 잡는다. 패치는
> 보관해 두었고 a081 land 후 재적용한다.
>
> **문서 지적은 먼저 정리했다** (review.md "F1~F8 처리 현황"): F3(두 번째
> MODIFIED), F4(떨어진 scenario 2건 복원), F5(logic-map의 철회된 설계 서술),
> F8(STORY 승인 기준). 남은 F2·F6·F7은 코드 재적용과 같은 batch다.

> **개정 2026-08-04.** 초안의 2~5장은 fragment + 스크립트 설계의 task였다. 그
> 설계는 issues.md I3·I4로 철회되었고 아래는 정석 수정 뒤의 task다. 철회된
> 산출물(`line_fragment.go`, `line_fragment_test.go`, 템플릿·CSP·라우트 편집)은
> 전부 원복했다.

## 1. 근거 고정 (편집 전)

- [x] 1.1 `RefreshSeconds`가 `holdingsTTL`에서 파생된다는 것과, 그 `holdingsTTL`이
      브로커 rate budget 상수라는 것을 현재 HEAD에서 확인해 기록한다.
      **발견**: 파생을 정당화하는 주석이 캐시 코드와 모순된다 → issues.md I1.
      `RefreshSeconds`는 하나가 아니라 둘이다 → issues.md I2.
- [x] 1.2 `/positions`가 `get`을, `/dashboard`가 `peek`를 쓰는 현재 경로를
      확인한다 (`portfolio.go:533`, `overview.go:655`). 두 계약은 a080이
      건드리지 않는다.
- [x] 1.3 두 `RefreshSeconds`의 Function Logic Map과 Branch Test Map을 **편집 전에**
      작성한다. `check_analysis.py` 통과.
- [x] 1.4 무-스크립트 가드 3건과 `consoleScreens` 목록의 현재 상태를 확인한다.
      베이스라인: `internal/console` 699 passed.
- [x] 1.5 `chrome.go:77-79`의 "reload 셀과 meta 태그는 한 사실"을 확인한다.
- [x] 1.6 **근본 원인**: 무-JavaScript가 `2026-07-31-streamline-trading-views`의
      명시적 Non-Goal이고, 재로드 주기 ≥ 캐시 TTL이 `rate budget 보호`의 SHALL임을
      확인한다. 두 경로 모두 스펙에 막혀 있다 → issues.md I4.

## 2. 예산 — 이 change가 성립하는 유일한 근거

- [x] 2.1 `/positions`를 TTL보다 짧은 주기로 반복 재로드해도 브로커 호출이
      `ceil(창 길이 / TTL)`을 넘지 않는다.
      `TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling`.
- [x] 2.2 그 테스트가 무의미하게 통과할 수 없음을 자체 검사한다 — 재로드 횟수가
      허용 호출 수보다 많지 않으면 `t.Fatal`.
- [x] 2.3 `/dashboard`는 어떤 주기에서도 브로커 호출 0회.
      `TestTheOverviewSpendsNothingHoweverOftenItReloads`.

## 3. 주기와 그 출처

- [x] 3.1 재로드 주기가 엔진 관측 주기(5초)이고 `holdingsTTL`에서 파생되지 않는다.
      `TestTheRedrawCadenceIsTheEnginesAndNotTheCaches` — 숫자만이 아니라
      **분리 자체**를 판정한다.
- [x] 3.2 두 화면이 같은 주기를 렌더한다.
      `TestBothLineScreensRenderTheSameCadence`.
- [x] 3.3 스트립의 reload 셀과 meta 태그가 계속 한 값을 말한다
      (`TestTheReloadCellAndTheMetaTagAreOneFact` green 유지 — 값 출처가 하나뿐이라
      추가 배선 없이 성립한다).
- [x] 3.4 `consoleScreens` 표의 두 항목을 새 상수로 갱신한다. 이 표가
      `TestEachScreenKeepsItsOwnReloadPeriod`와 3.3의 단일 출처다.

## 4. 무-스크립트 유지

- [x] 4.1 두 화면이 `<script>`를 싣지 않고 CSP가 `consoleHTMLCSP`와 정확히 같다.
      `TestTheLineScreensStillShipNoScript` — a080 쪽에서 제약을 명시해, 다음에
      이 화면을 빠르게 만들려는 시도가 세 개의 실패 테스트에서 이유를 역추적하지
      않게 한다.
- [x] 4.2 무-스크립트 가드 3건이 **손대지 않고** 통과한다.
- [x] 4.3 `lineRefreshInterval` 상수와 그 근거를 `line_cadence.go`에 둔다
      (엔진 패키지 import 없음, design D3).
- [x] 4.4 두 `RefreshSeconds`가 새 상수를 읽고, 잘못된 주석을 정정한다.

## 5. 갱신된 기존 테스트

옛 결합(주기 = `holdingsTTL`)을 명문화하던 4건이다. 결합을 끊는 것이 이
change이므로 갱신 대상이며, 각 갱신은 약화가 아니라 **판정 근거의 이동**이다.

- [x] 5.1 `TestTheOverviewReloadsAtTheCacheTTL` → `...AtTheEngineCadence`.
      새 상수가 `holdingsTTL`과 같아지면 실패하는 절을 더해, 테스트가 분리를
      실제로 구분하게 만든다.
- [x] 5.2 `TestThePositionsScreenAsksTheBrowserToReloadAtTheCacheTTL` →
      `...AtTheEngineCadence`. 상한 판정은 2.1이 직접 한다.
- [x] 5.3 `consoleScreens` 두 항목 (3.4와 동일).
- [x] 5.4 `portfolio_refresh_test.go` 파일 주석의 "reload bounded" 문단을 새
      근거로 갱신한다.

## 6. VERIFY

- [x] 6.1 변이 검증: `RefreshSeconds`를 `holdingsTTL`로 되돌리면 3.1이 RED가
      되는지 확인하고 되돌린다.
- [x] 6.2 변이 검증: `holdingsCache.get`의 TTL gate를 제거하면 2.1이 RED가
      되는지 확인하고 되돌린다.
- [x] 6.3 `make test`(46 packages ok, exit 0), `make vet`(clean), `make validate`(exit 0),
      `make sdd-sync`, `make sdd-check`(exit 0 — CodeGraph hard evidence fresh; codegraphcontext·
      GBrain은 advisory WARN, WORKFLOW.md가 비차단으로 규정).
- [ ] 6.4 `make gate CHANGE=a080-line-refreshes-on-engine-cadence`.
- [ ] 6.5 사람 승인 후 컨테이너 실측 — 워터마크가 오른 직후 `/positions`가 5초
      안에 그 값을 보이는지, **스크롤 위치가 재로드를 넘어 유지되는지**(design D4의
      유일한 미검증 전제), 열린 상세가 유지되는지, `/dashboard`의 브로커 호출이
      0인지.
      **2026-08-05 배포 후 셋은 확인했다** (review.md 5차): 두 라인 화면의
      `content` 30 → **5**(다른 화면은 무변동), `/dashboard`만 10회 열어도 캐시
      시각이 `02:03:07Z`에 고정된 채 경과만 자라 **브로커 호출 0**, 열린 상세는
      URL에 있어 재로드를 넘어 유지된다. **남은 하나는 스크롤 위치**이며 브라우저
      동작이라 `curl`로 관측할 수 없다 — 사람이 `/positions`를 열어 아래로
      스크롤하고 5초를 기다리면 끝난다. 그 확인 전에는 이 항목을 체크하지 않는다.
- [x] 6.6 실측: 탭 1개를 열어 둔 상태의 journal read 지연과 엔진 사이클 시간
      (review.md R2 — 읽기 볼륨이 6배가 된다. WAL이라 쓰기 차단은 설계상 없다).
      **2026-08-05 완료.** 36회 렌더 / 180초에서 100ms를 넘은 렌더가 **정확히
      6회** — 180초 ÷ 30초 TTL과 같다. 재로드 빈도 6배, 브로커 비용 무변동.
      엔진 `reconcile` 간격은 부하 구간 62.8·60.5·62.6·62.5초로 조용한 구간
      62.5~62.9초와 구분되지 않는다. 기록은 review.md 5차.

## 7. 리뷰와 기록

- [x] 7.1 proposal-freeze 리뷰 (review.md 1차).
- [x] 7.2 **Requirement 변경 리뷰 재실행** — `rate budget 보호`를 MODIFY하므로
      WORKFLOW.md §142가 요구한다. 검토 대상 셋: 상한이 여전히 지켜지는가, 그
      상한을 지키는 주체가 실제로 캐시인가, 나머지 절이 글자 그대로 보존되었는가.
- [x] 7.3 발견 사항을 `issues.md`에 남긴다 (I1~I4).
- [x] 7.4 PM story/tracker 동기화 (`generate_master_tracker.py --check` current).
- [x] 7.5 별도 컨텍스트의 독립 리뷰 (WORKFLOW.md §역할 분리) — 3인, review.md 3차.
      **결과: BLOCKED.** F1~F7.

## 8. 재개 batch (a081 land 후)

독립 리뷰 F1~F8의 나머지와, 코드 재적용에 딸린 재기준화다.

- [x] 8.1 F3 — `콘솔 공통 상태 표시줄`을 두 번째 MODIFIED로 추가하고 scenario
      `화면별 재로드 주기 보존`의 THEN을 고친다. 나머지 절과 scenario 9건은
      글자 그대로 보존. `openspec validate --strict` 통과 확인.
- [x] 8.2 F4 — 떨어진 scenario `검증 실행 중 — 캐시 있음`·`— 콜드 캐시`를 복원한다.
- [x] 8.3 F5 — 두 `RefreshSeconds`의 branch-test-map과 function-logic-map 본문에서
      철회된 스크립트 설계 서술과 없는 task 참조(3.8·3.9·3.10)를 걷어낸다.
      엔진 부하 계약(C7)을 새로 세운다.
- [x] 8.4 F8 — `STORY-TOS-a080.yaml` 승인 기준 5·6·7을 현재 설계로 교체한다.
- [x] 8.5 코드 재적용 — 보관한 패치를 a081 land 후 트리에 되돌린다
      (`overview.go`, `overview_test.go`, `portfolio_pages.go`,
      `portfolio_refresh_test.go`, `status_strip_test.go`, `line_cadence.go`,
      `line_cadence_test.go`).
- [x] 8.6 F2 — `status_strip_test.go`의 `strings.Contains(cell, "5초마다")`가
      `"15초마다"`의 부분 문자열에 걸려 눈이 머는 것을 고친다. 리뷰어가 실행한
      변이(템플릿을 `15초마다` 고정으로)를 재현해 RED를 확인한다.
- [x] 8.7 F6 — `portfolio_pages.go`·`holdings.go` 파일 머리말 주석에서 a080이
      없앤 결합("reloads itself at the holdings cache TTL — no faster")을 걷어낸다.
      issues.md I1이 지목한 결함과 같은 종류다.
- [x] 8.8 F7 — `TestTheOverviewSpendsNothingHoweverOftenItReloads`의 장식용 clock
      advance를 걷어내거나, 기존 `TestTheOverviewMakesNoBrokerCall`과 중복이면
      제거한다.
- [x] 8.9 `base-commit.txt`를 a081 land 후 커밋으로 재고정한다
      (`capture_change_base.py`는 덮어쓰기를 거부하므로 파일을 지우고 재실행).
- [x] 8.10 네 target의 AST를 재생성하고 `check_analysis.py`를 다시 통과시킨다.
- [x] 8.11 Requirement 변경 리뷰를 **두 번째 MODIFIED에 대해서도** 실행한다
      (WORKFLOW.md §142). 검토 대상: 표시줄이 주기를 정하지 않고 말하기만 한다는
      성질이 보존되는가, 톤 임계의 근거인 캐시 TTL이 재로드 주기와 분리되는가.
- [x] 8.12 `make test`·`vet`·`validate`·`sdd-sync`·`sdd-check` 재실행 후 6.4 gate.
      재개 batch 후 실측: `internal/console` **711 passed**, `make test` 0 failed,
      `vet` clean, `validate` exit 0, `check_analysis.py` `evidence complete`
      (target 4 → **5**). gate는 6.4~6.6과 함께 배포 뒤에 돈다.
