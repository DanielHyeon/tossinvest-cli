## 1. 근거와 RED

- [x] 1.1 candidate shadow observation의 시장·세션·누락률·5/15/30분 markout 보고서를 생성하고 review 입력으로 고정한다.
- [x] 1.2 incomplete/wrong-market/unapproved/approved threshold set, legacy-unapproved와 5/15/30분 +60초 tolerance/not_measured markout RED 테스트를 추가한다.

## 2. 승인 계약

- [x] 2.1 immutable threshold set schema, evidence digest와 fail-closed loader를 구현한다.
- [x] 2.2 verdict에 threshold version을 연결하고 order/RiskIntent dependency가 없음을 정적 테스트로 고정한다.
- [x] 2.3 transport-neutral `internal/markout`을 기존 관측만 소비하도록 구현하고 a049 재사용 golden fixture를 고정한다.
- [x] 2.4 change 완료 상태를 `unapproved/passed=0`으로 기록하고, 최초 수치 승인은 구현 gate 밖의 별도 사람 evidence activation record로 남긴다. (activation record 없음 — 의도된 dormant 상태)

## 3. 검증

- [x] 3.1 `candidate-filters` descriptor와 화면에 label/help/default state/desired/effective/unit/range/direction/sample/evidence/apply timing을 구현한다.
- [x] 3.2 unapproved를 숫자 0으로 표시하지 않는 read-only 상태, 시장·세션 전환, incomplete evidence, preview/CAS와 모바일·접근성 테스트를 통과한다.
- [x] 3.3 candidate focused/full test·vet·validate와 독립 리뷰를 통과한다.
- [x] 3.4 `make gate CHANGE=a046-approve-candidate-veto-thresholds`을 통과한다.

## 4. a047 handoff fail-closed 보강

- [x] 4.1 `AssessApprovedCandidate` 변경 전 Function Logic Map과 Branch Test Map을 고정한다.
- [x] 4.2 dangerous/unmeasured/wrong-market/pass와 candidate-life/provenance RED 테스트를 추가한다.
- [x] 4.3 measured-and-clear 전용 typed approval과 immutable threshold/candidate-life provenance를 구현한다.
- [x] 4.4 focused/race/full test·vet·validate·SDD 검증을 통과한다.
- [ ] 4.5 구현 컨텍스트와 분리된 독립 재리뷰를 통과한다.
- [ ] 4.6 `make gate CHANGE=a046-approve-candidate-veto-thresholds`을 다시 통과한다.
