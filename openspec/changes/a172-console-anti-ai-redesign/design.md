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
