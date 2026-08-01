## 1. 분석과 RED

- [x] 1.1 a041 base commit, CodeGraph impact와 `proposalQuantity`, ladder/ratchet/exitloop Function Logic·Branch Test Map을 작성한다.
- [x] 1.2 snapshot 단조성·breach 우선순위·1주 중간익절 생략·최종 1주 전량·0수량 금지 RED 테스트를 추가한다.
- [x] 1.3 policy ID/version/digest, snapshot/decision ID 결정성과 중복 소비자 race RED 테스트를 추가한다.

## 2. GREEN과 통합

- [x] 2.1 immutable `ExitLineSnapshot`과 세 preset의 label/help/default/unit/1주 projection descriptor 계약을 최소 구현한다.
- [x] 2.2 transport-neutral `internal/settingmeta` 최소 계약과 finite option/control/provenance validator를 구현한다.
- [x] 2.3 ladder/ratchet과 exitloop가 같은 snapshot을 소비하도록 연결하고 기존 정책 수치를 보존한다.
- [x] 2.4 자체 진입·외부 편입·pending proposal·부분/최종 청산과 a050 descriptor 소비 회귀 테스트를 통과한다.

## 3. 검증

- [x] 3.1 Function Logic Map을 구현 후 AST와 맞추고 focused/race/full test·vet·validate를 통과한다.
- [x] 3.2 적대적 Eng 독립 리뷰와 `make gate CHANGE=a041-complete-exit-line-contract`을 통과한다.
