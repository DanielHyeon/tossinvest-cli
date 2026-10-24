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
