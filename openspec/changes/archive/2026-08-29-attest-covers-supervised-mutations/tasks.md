# Tasks: attest-covers-supervised-mutations

> 이 change는 게이트 5절을 **만족 가능하게** 만들 뿐 게이트를 열지 않는다. 9절
> `const profileProtection = ProtectionUnwired`(interlock.go:175)는 어떤 설정으로도
> 만족할 수 없고, 2c `add-protection-orders`가 그 상수를 뒤집기 전까지 엔진은 기동하지
> 않는다. 그리고 2c는 여전히 `verify-execution-capability` 2.5~2.9 완료 전 작성 금지다.

- [x] 1.1 [T] `attest.Proof` — endpoint·성공 시각·출처(기록 파일)·시장을 담는 작은 값 타입.
  `internal/soak`과 `internal/verifylive`가 공유하되 서로를 import하지 않게 하는 자리
  (design.md D4). `Attestation.SupervisedBy []Proof`는 가산·`omitempty`,
  format_version 불변. RED: 옛 attestation 파일이 읽히지 않음 → GREEN: 그대로 읽힌다.
- [x] 1.2 [T] `verifylive.SucceededEndpoints(entries, now, maxAge)` — 오류 없이 성공한
  호출의 endpoint를 시각·계좌와 함께 돌려준다. 실패 호출만 있는 endpoint는 포함하지 않는다.
  창 밖 성공은 포함하지 않는다. RED/GREEN.
- [x] 1.3 [T] `BuildAttestation`이 감독 증거를 받는다 — `LiveOnlyEndpoints()` **밖**이면
  거부(soak 기록의 비-GET 거부와 대칭). RED: 조건주문 endpoint가 실림 → GREEN: 거부.
- [x] 1.4 [T] 계좌 결속 — `attest.Mask(soak.AccountRef) == verify.AccountRef`.
  불일치는 **건너뛰지 않고 발급 거부**. RED/GREEN.
- [x] 1.5 [T] 유효 기간 밖 증거는 싣지 않는다. 경계(`==` 시점)를 명시적으로 고정한다. RED/GREEN.
- [x] 1.6 [T] 감독 검증은 GET을 기여하지 못한다 — 검증 기록이 성공한 읽기를 담고 있어도
  attestation의 읽기 집합은 soak 기록에서만 온다. RED/GREEN.
- [x] 1.7 [T] 출처 기록 — 실린 mutation endpoint마다 무엇이 증명했는지가 attestation에 있다.
  RED/GREEN.
- [x] 1.8 [T] 인터록 통합 — 두 증거원이 모두 채워진 attestation이 5절을 통과하고,
  **9절에서 여전히 거부된다**. 이 테스트가 이 change의 안전 주장 자체다. RED/GREEN.
- [x] 1.9 [T] 부족한 채 발급 — 감독 증거가 없으면 읽기만 담아 발급되고 5절이 거부한다
  (기존 동작 회귀 방지).
- [x] 1.10 `cmd/tossctl soak attest` 배선 — 시장별 검증 기록을 `verify`와 **같은 해석
  경로**로 찾아 읽고 전달. `--verify-record`(반복 가능) 재정의. 출력에 무엇이 무엇으로
  증명됐는지와 아직 빠진 것을 표시.
- [x] 2.1 Function Logic Map + Branch Test Map + risk-pattern-report,
  `python3 tools/logic-map/check_analysis.py --change attest-covers-supervised-mutations`
- [x] 2.2 Pre-Edit 선언을 review.md에 기록 (High-risk: 자동화 게이트 인터록 입력)
- [x] 2.3 적대적 Eng 관점을 포함한 gstack proposal-freeze 리뷰 → review.md
- [x] 2.4 `make sdd-sync && make sdd-check`, `go test ./... -count=1`, `make vet`,
  `make validate`, PM registry allowlist + `tools/pm/test_generate_master_tracker.py` fixture
- [x] 2.5 `make gate CHANGE=attest-covers-supervised-mutations`
- [x] 3.1 실계좌 attestation 재발급으로 8/8 확인 — `tossctl soak attest`가 mutation 2개를
  싣는지, 그리고 게이트가 **여전히** 9절에서 거부하는지 둘 다 확인한다. 계좌 호출 0건.
