## 1. Contract and evidence

- [x] 1.1 `STORY-TOS-a054`와 `a054-console-status-shell`의 번호·범위가 1:1인지 검증한다.
- [x] 1.2 `python3 tools/sdd/capture_change_base.py --change a054-console-status-shell`로 구현 전 commit을 고정한다.
- [x] 1.3 `make sdd-sync` 후 `enginelock.Read`·`verifylive.SessionAdvisoryFor`·`binstamp.Of`·각 화면 `RefreshSeconds`의 definition/callers/impact를 확인한다.
- [x] 1.4 `"/"` 참조 전수를 라우트/쿠키 스코프/URL 정규화 셋으로 분류하고 design §4 표와 일치하는지 확인한다.
- [x] 1.5 기존 함수 편집 대상(`handleDashboard`·`restartTarget`·`handleRemoteLogin` 계열·각 page 생성 handler)에 Function Logic Map과 Branch Test Map을 작성한다.

## 2. RED → GREEN → REFACTOR — 공용 chrome

- [x] 2.1 RED: 화면별 재로드 **두 방향** 테스트. ① 주기 — 검증 콘솔·`/verify`는 run이 working일 때만 2초, 개요·`/positions`는 `holdingsTTL`, `/orders`는 `ordersTTL`, `/signals`는 자기 tick. ② 걸림/안 걸림 — `Refresh`가 chrome으로 옮겨가므로 어떤 화면이 설정을 빠뜨리면 자동 재로드가 조용히 꺼진다. 두 필드 모두 화면별로 고정한다.
- [x] 2.2 RED: 상태 표시줄이 모든 화면에 렌더된다. 엔진 미배선·엔진 정지·엔진 실행 세 상태를 각각 문구로 구분한다.
- [x] 2.3 RED: 신선도 세 상태. 캐시 기반 화면은 기록된 시각과 경과, 요청 시점 읽기 화면은 렌더 시각과 "요청 시 읽음", 읽기 실패는 사유. 기록 없는 시각을 만들지 않음을 고정한다.
- [x] 2.4 RED: 신선도 톤이 화면 자기 TTL을 쓴다. `ok`/`warn`/`stale` 경계에서 판정이 바뀐다.
- [x] 2.5 RED: **알려진 갱신 보류에는 톤이 붙지 않는다.** 검증 run 진행 중 캐시 화면, 그리고 스캔 tick 미도래 상태의 `/signals`에서 경고 톤 대신 사유가 표시된다.
- [x] 2.6 RED: 승인 대기 중인 검증 run이 있으면 모든 화면의 표시줄이 그 사실과 직접 링크를 표시하고, 표시줄에는 승인 폼이 없다.
- [x] 2.7 RED: 엔진 마커 갱신 시각이 화면 데이터 시각 칸에 표시되지 않는다 — 둘은 다른 사실이다.
- [x] 2.8 RED: 브로커 호출 회계. 상태 표시줄 추가 후에도 TTL당 holdings 1콜 상한이 유지된다.
- [x] 2.9 GREEN: `chrome` 구조체를 page struct에 임베딩하고 `head` 템플릿에 status strip을 넣는다.
- [x] 2.10 REFACTOR: 화면별로 흩어진 신선도 문장 중 표시줄이 흡수한 것을 제거한다. 흡수하지 못한 문장(집계 기준·갱신 보류 사유)은 남긴다.

## 3. RED → GREEN → REFACTOR — 경로와 이름

- [x] 3.1 RED: `/`가 개요로 303 리다이렉트한다.
- [x] 3.2 RED: 재시작 핸드오프 왕복. `?handoff=<token>`으로 접속 → 토큰 **정확히 1회** 소비 → 렌더된 화면 착지 → 주소창에 소비된 토큰 잔류 없음. 같은 토큰 재사용은 이동 전과 동일하게 거부된다.
- [x] 3.3 RED: 재시작·soak 재기동 결과 안내가 **검증 콘솔**로 돌아가고 `?notice=`가 보존된다. 개요로 가지 않는다.
- [x] 3.4 RED: 원격 로그인 후 리다이렉트가 존재하는 경로에 착지한다.
- [x] 3.5 RED: 세션 쿠키 스코프가 `/`로 유지되어 `/positions`·`/orders`에서 세션이 살아 있다.
- [x] 3.6 RED: 검증 콘솔이 자기 경로에서 렌더되고 엔진 시작/정지·검증 승인 컨트롤을 그대로 갖는다.
- [x] 3.7 RED: 각 화면을 렌더해 `aria-current`가 붙은 nav 항목의 텍스트와 `<h1>` 텍스트가 일치한다(렌더 비교 — 정적 파싱 아님).
- [x] 3.8 GREEN: 라우트 2건과 design §4의 착지 경로를 옮긴다. 쿠키 `Path` 3곳은 건드리지 않는다.
- [x] 3.9 라우트 표 정적 검사가 새 라우트 2건을 보고 있고 둘 다 GET임을 확인한다.

## 4. RED → GREEN → REFACTOR — 표시 프리미티브

- [x] 4.1 RED: 화면별로 렌더 결과에서 판정 — viewport meta 존재, 좁은 viewport 미디어 쿼리 적용, viewport보다 넓은 고정 px 폭 부재. **레이아웃 실측은 자동 검사에 넣지 않는다**(하니스가 없다 — design §5b).
- [x] 4.2 RED: 렌더 결과에 bare `<table>`이 없다 — 모든 표가 `.data-table`이거나 `.table-scroll` 안에 있다.
- [x] 4.3 RED: `.state-badge`가 렌더 결과와 스타일시트 양쪽에서 사라지고 `.status-pill`만 남는다. `strategy_runtime_test.go:195`가 계속 통과한다.
- [x] 4.4 RED: h1/h2/h3/body 폰트 크기가 서로 구분 가능하고 대비 기준을 유지한다.
- [x] 4.5 GREEN: 스타일시트와 8개 템플릿을 고친다.
- [x] 4.6 CSP 회귀 검사: 렌더 결과에 `on[a-z]+=`·`<script>`·`javascript:`가 없고 응답 CSP가 그대로다.
- [x] 4.7 증거: 브라우저 375px·1280px 실측을 사람이 1회 수행하고 결과를 `review.md`에 기록한다. 자동 검사가 아니다.

## 5. RED → GREEN → REFACTOR — 개요 상단 요약

- [x] 5.1 RED: 개요 상단 요약 strip이 렌더되고 각 칸이 상세를 소유한 화면으로 링크한다.
- [x] 5.2 RED: 미측정 칸은 0이 아니라 사유를 말한다.
- [x] 5.3 GREEN: 요약 strip을 개요 템플릿 상단에 넣는다. 하단 상세 표는 유지한다.

## 6. Verification and completion

- [x] 6.1 변이 검증: 화면 하나에서 `RefreshSeconds()`를 지워 2.1①이 FAIL하는지, `Refresh` 설정을 빠뜨려 2.1②가 FAIL하는지, 표 하나를 bare로 되돌려 4.2가 FAIL하는지, 쿠키 `Path`를 `/dashboard`로 바꿔 3.5가 FAIL하는지, 갱신 보류 사유를 지워 2.5가 FAIL하는지 확인하고 전부 되돌린다.
- [x] 6.2 `openspec validate a054-console-status-shell --strict --no-interactive`를 통과한다.
- [x] 6.3 upstream 상속 테스트 650개 green 유지를 확인한다.
- [x] 6.4 `make sdd-check`와 독립 리뷰(경량 — UI change)를 통과하고 구현 후 발견을 `review.md`에 추기한다.
- [x] 6.5 `make gate CHANGE=a054-console-status-shell`를 통과하고 PM Story를 동기화한다.
