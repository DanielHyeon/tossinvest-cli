## Why

`/positions`와 `/orders`는 필요한 사실을 갖고 있지만 동일한 시각 무게의 긴 설명과 넓은 표에 묻혀
핵심 상태를 빠르게 읽기 어렵다. 또한 포지션 편입·제외 체크박스는 deny-by-default CSP가 차단하는
inline `onchange`에 의존해, 특히 “제외” 클릭이 화면상 동작하지 않는다.

## What Changes

- 포지션 행의 1차 정보 계층을 관리 상태·손익·진입가·현재 보호선·익절 진행으로 재구성한다.
- 기존 journal 값을 재계산하지 않고 최초 손절과 현재 기준선을 명확히 구분해 표시한다.
- 편입·제외를 JavaScript 없는 명시적 POST 버튼으로 바꾸고 현재/변경 상태를 문구로 확인시킨다.
- 주문 화면의 미체결 요약, 필터, 목록을 한 흐름으로 정돈하고 낮은 우선순위의 근거 설명은 접는다.
- 주문 목록을 핵심 열 중심으로 줄이고 세부 값은 행 내부 상세 영역으로 이동해 작은 화면에서도
  수평으로 깨지지 않게 한다.
- CSP 자체를 완화하지 않고 inline event handler를 제거한다.

## Capabilities

### New Capabilities

- 없음.

### Modified Capabilities

- `operator-console`: 포지션·주문 화면의 핵심 정보 계층, 반응형 표시, CSP 안전한 편입·제외 조작을
  운영 계약으로 고정한다.

## Impact

- `internal/console`의 서버 렌더 템플릿, 공용 스타일, 표시 전용 row helper와 fixture 기반 테스트.
- 기존 `/settings/include`, `/settings/exclude` POST/CSRF 계약과 브로커·journal 읽기 seam은 유지한다.
- 주문 제출·취소·정정, exit 정책 계산, journal 스키마, 엔진·soak·운영 토글에는 영향이 없다.
