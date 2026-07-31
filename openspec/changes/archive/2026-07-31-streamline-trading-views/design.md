## Context

TossOS 콘솔은 외부 자산 없이 `html/template`와 inline CSS만 사용하고, 배포 CSP는
`default-src 'none'`과 `form-action 'self'`를 강제한다. 현재 `/positions`는 핵심 보호 정보를 10열
표 다음의 긴 문장에 배치하고, 편입·제외 체크박스는 CSP가 실행하지 않는 inline `onchange`로
submit한다. `/orders`는 12열 표와 세 개의 같은 무게 섹션 때문에 62rem 컨테이너를 넘는다.

StockOS의 참고 화면에서 가져올 원칙은 한 화면에 하나의 명확한 제목, 상단 요약, 본문 목록,
낮은 우선순위 상세의 순서다. 색상·SPA·운영 제어는 복제하지 않는다.

## Goals / Non-Goals

**Goals:**

- 포지션의 관리 여부와 실제 보호선 및 익절 진행을 한 행의 1차 정보로 읽게 한다.
- 주문 화면을 미체결 상태 → 필터 → 목록 순서의 단일 흐름으로 만든다.
- 375px 모바일에서 문서 전체 수평 overflow 없이 핵심 정보를 읽게 한다.
- deny-by-default CSP를 그대로 유지하면서 편입·제외 POST가 작동하게 한다.
- 원장/브로커 미측정 상태와 기존 rate-budget 사실을 보존한다.

**Non-Goals:**

- 주문 제출·취소·정정 UI 추가.
- exit 정책 계산, 보호선 갱신, 손절/익절 값 재계산.
- journal/브로커/API/schema/캐시 변경.
- JavaScript, CSP nonce, `unsafe-inline` script, 외부 CSS/폰트 도입.

## Decisions

### D1. 테이블을 정보의 1차/2차 층으로 나눈다

포지션 표의 1차 열은 종목, 관리, 손익, 가격, 보호/익절, 관리 동작으로 제한한다. 수량·평가금액·
원장 식별자·래칫 진단은 행 아래 `<details>`에 둔다. 현재 보호선은 `Exit.Baseline`, 최초 손절은
`Exit.InitialStop`, 익절 진행은 `Taken()`과 `Rung()`의 기존 journal 문자열만 사용한다.

주문 표의 1차 열은 시각, 심볼/시장, 방향/구분, 상태/체결, 수량, 가격, 발주 주체로 제한한다.
주문번호와 평균체결가 같은 추적용 값은 행의 `<details>`로 이동한다. `<details>`는 native HTML이며
script가 필요 없다.

대안인 12열 표의 단순 축소는 작은 글자와 수평 스크롤만 만들기 때문에 채택하지 않는다.

### D2. 편입·제외는 상태를 서술하는 명시적 submit 버튼이다

체크박스와 `onchange=confirm(...)`을 제거하고 각 행에 CSRF hidden input과 명확한 submit 버튼을
둔다. 버튼 문구는 현재 상태에서 일어날 변경(`관리 편입 예약`, `편입 예약 해제`, `자동관리 제외`,
`제외 해제`)을 말한다. 설명은 결과 경계를 같은 action 영역에 표시한다.

서버 endpoint와 POST payload는 유지한다. 이 방식은 브라우저 confirm이 없지만 행위가 명시적이며,
설정 파일만 갱신한다는 기존 저위험 계약과 일치한다. 별도 확인 화면은 한 번의 클릭 요구와 충돌해
채택하지 않는다. 포지션 화면의 기존 소개문 “편입 버튼은 없다”는 새 동작과 모순되므로 “계좌·주문은
변경하지 않으며 편입·제외 조작은 설정만 갱신한다”로 교체한다.

### D3. 설명은 제거하지 않고 접는다

오류·미측정·stale 안내는 즉시 노출한다. 정상 상태에서 반복되는 provenance는
`<details class="explain">` 안으로 옮긴다. 운영상 중요한 사실은 숨기지 않되 첫 화면의 시각 소음을
줄인다. 페이지 잘림/하한 건수, 조건주문 발주 주체 불명, 해석 불가능 상태, 파싱되지 않은 원본 시각,
캐시 실패·보류·stale은 구현 설명이 아니라 판단의 불확실성이므로 항상 펼쳐 둔다.

### D4. 반응형은 CSS와 `data-label`만 사용한다

공용 `.data-table`은 desktop 표를 유지하고 720px 이하에서 각 행을 block row로 바꾼다. 각 `td`는
`data-label`을 가지며 CSS pseudo-element가 시각 라벨을 표시한다. 원래 `<thead>`와
`<th scope="col|row">`는 접근성 트리에 남기고 시각적으로만 감춘다. 표는 `<caption>`으로 이름을
갖는다.

같은 breakpoint에서 header/nav/dl도 wrap하고 모든 flex/grid child에 `min-width:0`을 적용한다.
문서 폭을 넘기는 code/id는 `overflow-wrap:anywhere`로 감싼다. 버튼·summary·필터 링크에는
`:focus-visible`을 표시하고 조작 대상은 모바일에서 최소 44px 높이를 갖는다. 현재 nav 링크는
`aria-current="page"`를 갖는다. 주문 화면만의 예외 CSS나 script는 만들지 않는다.

## Risks / Trade-offs

- [기존 테스트가 긴 설명의 정확한 위치를 기대] → 문구를 보존하고 구조 중심 테스트를 추가한다.
- [native `<details>` 안의 진단 정보가 덜 보임] → 보호선·익절·상태는 항상 노출하고 진단값만 접는다.
- [확인 dialog 제거로 오클릭 가능] → checkbox보다 동사형 버튼과 결과 설명을 사용하고 기존 CSRF와
  설정 검증을 유지한다.
- [공용 CSS가 다른 화면에 영향] → opt-in class를 사용하고 기존 bare table 규칙은 유지한다.
- [pseudo-element 라벨이 스크린리더 관계를 만들지 못함] → 원래 caption/thead/scope를 접근성 트리에
  유지하고 `data-label`은 시각 배치에만 사용한다.
- [실계좌 fixture 캡처의 정보 노출] → 시각 QA는 fake broker/journal 데이터만 사용한다.

## Migration Plan

1. fixture/HTML 계약 테스트를 먼저 추가한다.
2. 템플릿과 opt-in 공용 스타일을 배포한다. 데이터·설정 마이그레이션은 없다.
3. `/positions`에서 inline handler 부재와 POST endpoint/CSRF를 확인하고 `/orders` read-only 가드를
   재검증한다.
4. 회귀 시 이전 이미지로 되돌리면 된다. journal과 config 형식은 바뀌지 않는다.

## Open Questions

없음. 세부 숫자와 용어는 현재 journal 필드 및 operator-console spec을 그대로 따른다.
