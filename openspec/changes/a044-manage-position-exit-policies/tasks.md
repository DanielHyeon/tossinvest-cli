## 1. 수명주기 계약과 RED

- [ ] 1.1 a042 완료를 확인하고 adoption/exit/config/console write 함수의 Function Logic·Branch Test Map을 작성한다.
- [ ] 1.2 preview/CAS/audit/release/re-adopt/generation/rebind 금지 RED 테스트를 추가한다.

## 2. Journal과 domain

- [ ] 2.1 additive override/generation schema와 repository transaction을 구현한다.
- [ ] 2.2 preview/apply/release/re-adopt command와 active protection/exit 충돌 guard를 구현한다.
- [ ] 2.3 console read-only 불변을 유지하는 narrow `PositionPolicyCommander`와 serialized journal command service를 구현한다.

## 3. `position-management` 운영 표면

- [ ] 3.1 a050 descriptor 계약에 맞춰 종목별 정책과 외부 매수 자동편입을 별도 section으로 구성하고 label/help/default/desired/effective/range/apply timing/provenance를 구현한다.
- [ ] 3.2 enabled OFF, 합성 손절 5%(2~20%, step 0.5%), 빈 include/exclude, exclude 우선, 공통 정책 상속과 1주 동작을 fixture로 고정한다.
- [ ] 3.3 session+CSRF+If-Match console routes가 commander만 호출하게 하고, 영향 종목/generation preview와 명확한 before/after/restart 설명을 구현한다.
- [ ] 3.4 release/re-adopt를 일반 저장과 분리된 danger action으로 만들고 active exit 충돌·412·stale·unknown 상태를 테스트한다.
- [ ] 3.5 LIVE toggle·기존 position snapshot·broker call 무변경 정적/통합 테스트를 통과한다.

## 4. 검증

- [ ] 4.1 migration/crash/race/full test·vet·validate와 적대적 Eng 리뷰를 통과한다.
- [ ] 4.2 `make gate CHANGE=a044-manage-position-exit-policies`을 통과한다.
