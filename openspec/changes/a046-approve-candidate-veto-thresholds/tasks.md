## 1. 근거와 RED

- [ ] 1.1 candidate shadow observation의 시장·세션·누락률·5/15/30분 markout 보고서를 생성하고 review 입력으로 고정한다.
- [ ] 1.2 incomplete/wrong-market/unapproved/approved threshold set, legacy-unapproved와 5/15/30분 +60초 tolerance/not_measured markout RED 테스트를 추가한다.

## 2. 승인 계약

- [ ] 2.1 immutable threshold set schema, evidence digest와 fail-closed loader를 구현한다.
- [ ] 2.2 verdict에 threshold version을 연결하고 order/RiskIntent dependency가 없음을 정적 테스트로 고정한다.
- [ ] 2.3 transport-neutral `internal/markout`을 기존 관측만 소비하도록 구현하고 a049 재사용 golden fixture를 고정한다.
- [ ] 2.4 change 완료 상태를 `unapproved/passed=0`으로 기록하고, 최초 수치 승인은 구현 gate 밖의 별도 사람 evidence activation record로 남긴다.

## 3. 검증

- [ ] 3.1 `candidate-filters` descriptor와 화면에 label/help/default state/desired/effective/unit/range/direction/sample/evidence/apply timing을 구현한다.
- [ ] 3.2 unapproved를 숫자 0으로 표시하지 않는 read-only 상태, 시장·세션 전환, incomplete evidence, preview/CAS와 모바일·접근성 테스트를 통과한다.
- [ ] 3.3 candidate focused/full test·vet·validate와 독립 리뷰를 통과한다.
- [ ] 3.4 `make gate CHANGE=a046-approve-candidate-veto-thresholds`을 통과한다.
