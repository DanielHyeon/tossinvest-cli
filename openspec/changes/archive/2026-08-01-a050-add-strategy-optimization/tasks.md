## 1. 설정 계약과 RED

- [x] 1.1 a049 evidence read model과 기존 optimization/config/audit 함수의 impact와 Function Logic·Branch Test Map을 작성한다.
- [x] 1.2 registry/snapshot/preview/CAS/apply/history/rollback/evidence 부족/LIVE 무변경,
  StockOS lane-console 탐색/부분 저장, 자유입력 control 0, registry option-only submission과
  `ui-design.md` category/default/help 계약 RED 테스트를 추가한다.

## 2. Optimization lifecycle

- [x] 2.1 a041 `internal/settingmeta` provider를 조합하는 category registry, exact-one-owner coverage validator와 immutable settings snapshot/candidate repository를 구현한다.
- [x] 2.2 preview validation, evidence digest와 insufficient-evidence 상태를 구현한다.
- [x] 2.3 CAS apply/history/rollback과 before/after/actor/reason audit를 구현한다.
- [x] 2.4 console/httpapi가 journal을 writable로 열지 않는 narrow `OptimizationCommander`와 durable idempotent control command service를 구현하고 모든 write adapter가 이를 공유하게 한다. engine만 journal trading state를 쓴다.

## 3. Console과 검증

- [x] 3.1 `/optimization`에 여섯 category, 반응형 탐색, 기본/desired/effective 값, field 설명,
  input-free registry control, sticky changed-subset preview, history와 rollback 표면을 `ui-design.md`대로 구현한다.
- [x] 3.2 active position·lane state·LIVE toggle 자동 변경이 없고 high-risk apply/rollback이 desired는 보존하되 effective entry를 manifest 재승인 전 OFF로 만드는 것을 통합/정적 테스트로 고정한다.
- [x] 3.3 360px/keyboard/44px touch/CSP/loading·error·stale·412·insufficient evidence UI 테스트와 race/full test·vet·validate, security/adversarial review를 통과한다.
- [x] 3.4 `make gate CHANGE=a050-add-strategy-optimization`을 통과한다.
