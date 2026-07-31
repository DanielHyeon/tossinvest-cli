## 1. Contract and evidence

- [ ] 1.1 `STORY-TOS-aNNN`과 `aNNN-<kebab-case-intent>`의 번호·범위가 1:1인지 검증한다.
- [ ] 1.2 관련 코드·테스트·운영 설정의 hard evidence를 수집한다.
- [ ] 1.3 기존 함수 편집 시 Function Logic Map과 Branch Test Map을 작성한다.

## 2. RED → GREEN → REFACTOR

- [ ] 2.1 요구사항별 실패 테스트를 먼저 추가하고 RED를 기록한다.
- [ ] 2.2 최소 구현으로 GREEN을 만든다.
- [ ] 2.3 리팩터링 후 대상 테스트를 다시 검증한다.

## 3. Verification and completion

- [ ] 3.1 `openspec validate aNNN-<kebab-case-intent> --strict --no-interactive`를 통과한다.
- [ ] 3.2 독립 리뷰와 `make sdd-check`를 통과한다.
- [ ] 3.3 `make gate CHANGE=aNNN-<kebab-case-intent>`를 통과하고 PM Story를 동기화한다.
