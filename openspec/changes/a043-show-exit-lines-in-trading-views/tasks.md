## 1. Read model과 RED

- [x] 1.1 a041/a042 snapshot API를 확인하고 positions/orders view builder 영향과 CSP·모바일 branch map을 작성한다.
- [x] 1.2 완전/stale/unknown/1주/미연결 order fixture의 렌더링 RED 테스트를 추가한다.

## 2. 화면 구현

- [x] 2.1 transport-neutral `internal/operatorview.ExitLineView` adapter를 구현해 console/httpapi의 별도 계산을 금지한다.
- [x] 2.2 `/positions`에 현재·다음 기준선, rung, 수량, 1주 설명과 `position-management` 문맥 링크를 추가한다.
- [x] 2.3 `/orders`에 명시적 attempt-intent lineage로 exit decision의 trigger snapshot, 미연결 상태와 `exit-protection` 문맥 링크를 추가하고 symbol/time fuzzy join을 금지한다.

## 3. 검증

- [x] 3.1 CSP, 360px 반응형 overflow, 접근성, `/positions`·`/orders` POST 405와 form/input/textarea/select/button/contenteditable 부재 DOM 테스트 및 console focused/full test를 통과한다.
- [ ] 3.2 UI 경량 독립 리뷰와 `make gate CHANGE=a043-show-exit-lines-in-trading-views`를 통과한다.
