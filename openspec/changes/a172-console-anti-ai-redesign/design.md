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

## 6단계 — 셰이더 인터랙션 (6.1 정찰 결과)
- 라이브러리: 후보 실측(esbuild 최소 사용 트리셰이킹 min/gz) raw WebGL2 290B/240B · twgl 32.6KB/11.1KB · OGL 44.4KB/12.7KB · three 529KB/132KB. 전체판 twgl 81.7KB/24.7KB, OGL 132.7KB/38.9KB. 전체화면 삼각형 프래그먼트 셰이더 한 장에 라이브러리 이득이 없어 **raw WebGL2**. "Radiance" 는 npm `radiance@0.0.1`(2022, 무관한 의존성 3개)·zbanks/radiance(Qt/C++ VJ 앱)뿐이고 WebGL 셰이더 라이브러리가 아니라 제외.
- 출처: 동일 출처 `embed` 자산, CDN·네트워크 호출 0.
- CSP: 현재 `default-src 'none'` + script-src 없음 → `script-src 'self'` 필요. 최적화 미리보기는 `consoleHTMLCSP + "; script-src 'sha256-…'"` 를 이어 붙이므로 지시어가 중복되면 첫 번째만 유효해 미리보기 인라인 스크립트가 깨진다 → 구현 시 해시 변형이 `script-src 'self' 'sha256-…'` 가 되도록 조립부를 고친다.
- refresh: 대시보드·검증 2초, 주문·포지션도 주기 reload. 컨텍스트는 페이지마다 새로 만들어짐(웜 약 3~5ms SwiftShader, GPU 프로세스 콜드 첫 컨텍스트 약 95~127ms는 브라우저당 1회). 결정: 모듈 로드 시 동기 1프레임을 그려 reload 공백을 막고, 애니메이션 위상은 `sessionStorage` 로 이어 가며, 초기화는 `requestIdleCallback` 으로 미루지 않는다(미루면 빈 배경이 보임). 6.3 에서 실측해 부적합하면 refresh 화면은 정지 프레임으로 한다.
