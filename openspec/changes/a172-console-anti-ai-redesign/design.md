# a172 design

## 토큰 계획
`:root` 의 기존 색 변수를 유지하고 반경·간격·그림자 토큰을 추가한다(2단계). 라이트/다크 둘 다 정의. 팔레트는 Tailwind 기본값 직결을 피하고 차분한 대비로 조정한다.

## 타입 스케일
본문·제목은 `clamp(min, preferred, max)` 유동 스케일. 단, 표 밀도(`.positions-table` 12px/18px, a058 사용자 결정·spec 요구)는 고정 유지.

## 반경 계획
균일 4/6px 를 버리고 요소 역할별로 다른(비대칭 포함) 반경 토큰을 둔다. 1단계에서는 최소로 중립화만 한다.

## 모션 계획
4단계. 스프링 느낌 `cubic-bezier` 오버슈트는 `transform`/`opacity` 만, `prefers-reduced-motion: reduce` 에서 제거. 주문·토글·승인·청산 버튼에는 지연 전환을 걸지 않는다(안전 불변식 4).

## 검증 계획
`gofmt -l`, `go vet`, `go test ./internal/console/ ./cmd/tossctl/`, 로컬 콘솔 렌더(1280/375px, 라이트/다크) 스크린샷·콘솔 에러 확인. 주문·토글·승인 버튼은 누르지 않는다.

## 모션 주의 (Impeccable 3단계 리뷰, 4단계 입력)
- 전환 0 유지 대상: 주문·승인·청산·토글 `button`/`.button-link`/`form`, `.sticky-save`(승인 바), `.console-approval`. 상태 규칙(`:hover`/`:active`/`:disabled`)은 즉시 반영이고 `transition` 을 걸지 않는다(안전 불변식 4).
- `.detail-cell .row-details[open]` 는 위치 지정에 `transform: translate(-50%, -50%)` 를 이미 쓴다. 열림 애니메이션은 `opacity` 또는 개별 `translate`/`scale` 속성으로만 하고 `transform` 을 애니메이션하지 않는다.
- 데이터 갱신(`meta refresh` 5초)으로 바뀌는 상태 변화에는 애니메이션을 걸지 않는다. 모션 후보는 탭·내비 `a.on` 전환과 `details` 열림뿐이다.
- `prefers-reduced-motion: reduce` 에서 모든 모션을 끈다. 컴포지터 속성(`transform`/`opacity`)만 쓰고 `will-change` 는 활성 애니메이션 중에만 둔다.

## 4단계 구현 결과 (위 「모션 주의」를 따름)
- 토큰 `--dur-1/2/3`(120/220/360ms), `--ease-spring`·`--ease-settle`(`linear()`, `@supports` 바깥 기본 `ease-out`), 움직임 토큰 `--nudge/--press/--pop-from/--tile-on`. reduced-motion 은 토큰만 리셋(`!important` 없음).
- 제외는 선택자로 지정: 일반 버튼 전환 규칙이 `:not(:where(form:not([method=get]) *, .sticky-save *, .approval-actions *, .confirm-check *, .console-approval *))` 로 뺀 뒤, 같은 컨테이너에 `transition: none` 을 한 번 더 박음. POST 폼 선택자를 `form:not([method=get])` 로 쓴 이유: 읽기 전용 화면 시험이 페이지 어디에도 `method="post"`/`method=post` 문자열이 없을 것을 고정하기 때문(CSS 포함). 모든 `<form>` 이 POST 이므로 폼 안 버튼은 전부 제외되고, 전환이 걸리는 것은 폼 밖 `.button-link` 링크 버튼뿐이다.

## 4단계 리뷰 반영으로 바뀐 수치
- `--dur-3` 360 → 300ms. 모달 열림은 전용 `--ease-pop`(12점, 정점 약 1.11~1.13, 오버슈트 약 12%), nav·눌림은 `--ease-spring`(정점 1.287) 그대로. `.choice-tile` scale 은 정적 상태에 전환이 일어날 수 없어 삭제(`--tile-on` 토큰 삭제).
- 제외 선택자에 `[formmethod]`, `[formaction]`, `form button`, `form input[type=submit]` 추가. 서버가 `data-mutating` 표식을 붙이는 방식은 Go 변경이라 이번 change 범위 밖(후속 후보).
- 모달 스크림: 열린 모달이 있는 `.detail-cell::before`(`position: fixed; inset: 0; z-index: 20`), `opacity` 만 애니메이션, `pointer-events: none` 이라 뒤의 어떤 버튼도 막지 않음. 모달 너비는 `100vw` 대신 `100%`(스크롤바 폭 어긋남 제거).

## 6단계 — CSS-only 앰비언트 톤 배경·hover 조명
- 앰비언트: `body > header`·`.console-status` 의 `::before`/`::after` 가 정적 저채도 radial-gradient 덩어리(투명 가장자리, blur 없음)다. 움직임은 `translate`·`rotate`·`opacity` 만, 20~40초 `alternate` 루프. 배경·그라디언트 자체는 애니메이션하지 않는다. `mix-blend-mode`·`filter` 0.
- 톤이 정보를 나른다: `body:has(...)` 가 `--tone` 을 정함 — 기본 `--accent`, 엔진 running + 데이터 ok 면 `--ok`, 데이터 warn 또는 엔진 stopped 면 `--warn`, 데이터 stale 이면 `--bad`. 알파 10% 이하(`color-mix`). 톤은 이미 텍스트 단어로도 렌더되므로 색만이 전달 수단이 아니다.
- 새로고침 화면: 상태 스트립이 이미 `data-reload="on"` 을 렌더하므로 마크업 변경 없이 `body:has([data-reload=on])` 에서 `animation: none`(2초마다 재시작해 튀는 것을 막음). 움직이는 것은 새로고침이 없는 화면뿐.
- hover 조명: `.optimization-nav a`·`.settings-tabs a`·`.choice-tile`·`.tier-card` 의 `::after` radial 하이라이트를 `opacity` 로만 페이드(`--dur-2`, `--ease-settle`), 포인터 추적 없음. 주문·승인 영역(`.sticky-save`·`.approval-actions`·`.confirm-check`·`.console-approval`) 안에는 적용하지 않는다.
- 방어: `prefers-reduced-motion` 에서 `animation: none`, `prefers-reduced-transparency` 에서 덩어리 제거. `will-change` 총량 ≤2 유지.

## 6.1 정찰: 검토했으나 채택 안 함 (JS·WebGL 안)
- 후보(esbuild 0.25 로 최소 사용 트리셰이킹, min / gz): raw WebGL2 290 B / 240 B · twgl.js 7.0.0(MIT) 32.6 KB / 11.1 KB(전체판 81.7 / 24.7) · OGL 1.0.11(Unlicense) 44.4 KB / 12.7 KB(전체판 132.7 / 38.9) · regl 2.1.1(MIT) 전체판 86.7 / 28.4 · three 0.186.1(MIT) 529 KB / 132 KB · glslCanvas 0.2.6(xhr 의존). 전부 의존성 0(glslCanvas 제외). 1순위는 raw WebGL2.
- "Radiance": 웹 셰이더 라이브러리가 아님 — npm `radiance@0.0.1`(2022, 무관), `@radiance/core` 404, zbanks/radiance 는 Qt/C++ VJ 앱.
- 이유: 콘솔 CSP 는 `default-src 'none'` 에 `script-src` 가 없고, 시험이 `<script`·`script-src` 부재를 여러 곳에서 고정(`display_primitives_test.go:296,304`, `market_schedule_test.go:142`, `optimization_ui_contract_test.go:107`, `performance_history_test.go:98` 등). 최적화 미리보기는 `consoleHTMLCSP + "; script-src 'sha256-…'"` 이어 붙이기라 `script-src 'self'` 를 넣으면 지시어가 중복돼 깨진다. 스펙이 script·CSP 완화를 SHALL NOT 으로 금지.
- 자동 새로고침은 대시보드·검증 2초(주문·포지션도 주기 reload): WebGL 컨텍스트가 페이지마다 새로 생겨 SwiftShader 첫 컨텍스트 95~127 ms, 이후 2.7~5.9 ms.
