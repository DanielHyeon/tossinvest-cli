## 1. 계약과 PM 검사

- [x] 1.1 `docs/WORKFLOW.md`와 OpenSpec 템플릿에 `aNNN-kebab-intent ↔ STORY-TOS-aNNN` 규칙과 a040 cutoff를 기록한다.
- [x] 1.2 PM validator에 신규 번호형 ID 형식·번호 일치·번호 중복 검사를 RED 테스트로 추가한다.
- [x] 1.3 최소 validator 구현으로 RED를 GREEN으로 전환하고 legacy Story/change 회귀를 유지한다.

## 2. 포트폴리오와 검증

- [x] 2.1 a040 이후 Story/change를 번호 순서로 등록하고 generated tracker를 갱신한다.
- [x] 2.2 `openspec validate a040-adopt-stockos-openspec-naming --strict`, PM tests, `make sdd-check`를 통과한다.
- [x] 2.3 경량 proposal review를 `review.md`에 기록하고 `make gate CHANGE=a040-adopt-stockos-openspec-naming`을 통과한다.
