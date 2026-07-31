## 1. 시간·예산 RED

- [ ] 1.1 a047 runtime, clock/calendar, candidate cadence와 API budget 영향 지도를 작성한다.
- [ ] 1.2 KR/US regular, holiday, early close, DST, clock jump, calendar 6시간 freshness/refresh 실패, 50%·최소 5-call reserve와 restart approval RED 테스트를 추가한다.
- [ ] 1.3 activation manifest scheduler/calendar mismatch·expiry·revocation 시 auto-resume 거부 RED 테스트를 추가한다.

## 2. Scheduler 구현

- [ ] 2.1 typed scheduler decision과 exchange-calendar adapter를 구현한다.
- [ ] 2.2 entry/candidate cadence를 reserved safety budget 뒤에 배치한다.
- [ ] 2.3 desired state version/actor/approval을 저장하고 유효한 상태만 재시작 복원한다.

## 3. 통합·검증

- [ ] 3.1 `strategy-runtime > 시장·일정` descriptor와 화면에 label/help/default/desired/effective/market/session/apply timing/calendar provenance/reason을 구현한다.
- [ ] 3.2 scheduler OFF, auto-start OFF, market none, regular only와 calendar read-only를 fixture로 고정하고 모바일·접근성·CSP를 검증한다.
- [ ] 3.3 closed market에서도 reconcile/exit/filldetect가 지속됨을 supervisor 테스트로 고정한다.
- [ ] 3.4 시간 경계/race/full test·vet·validate와 적대적 Eng 리뷰를 통과한다.
- [ ] 3.5 `make gate CHANGE=a048-add-market-aware-scheduler`을 통과한다.
