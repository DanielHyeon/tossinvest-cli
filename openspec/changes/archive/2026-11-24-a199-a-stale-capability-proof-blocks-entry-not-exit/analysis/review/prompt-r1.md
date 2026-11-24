너는 TossOS(Go 자동매매, 실제 돈) 저장소의 **적대적 설계 리뷰어**다. 읽기 전용이다 — 파일을 쓰거나 고치지 말고, 네트워크·`tossctl`·`~/.config/tossctl` 에 접근하지 말고, 시험을 돌리지 마라(필요하면 소스를 읽어 판단).
작업 디렉터리는 저장소 루트다. 아래 change 의 **설계(아직 구현 전)** 를 깨려고 하라.

change: `openspec/changes/a199-a-stale-capability-proof-blocks-entry-not-exit/`
- 읽을 것: `proposal.md`, `design.md`, `specs/engine-safety/spec.md`(delta), `analysis/claim-check.md`, `analysis/edit-surface.md`,
  `analysis/verified-census/README.md`, `analysis/d1a-d4-sites.md`, `analysis/evidence-phase0.md`.
- 현재 코드(설계가 바꿀 것): `internal/app/engine/interlock.go`(verifyGate · runInterlock · logGateDecision · AutomationStatus), `internal/app/engine/engine.go`(NewContext),
  `internal/app/engine/exitwiring.go`(Context.ExitObserver), `internal/app/engine/reconcileloop.go`(Context.ReconcileDriver), `internal/app/engine/attestation_watch.go`,
  `cmd/tossctl/engine.go`(runEngineRun · engineRuntime), `internal/attest/attest.go`, `internal/execgw/retry.go`(EntryGate · latchOrder), `internal/execgw/gateway.go`(submit · checkEntry),
  `internal/execgw/protection.go`, `internal/app/engine/gateway.go`(buildGateway), `internal/app/engine/strategy_entry_supervisor.go`.
- 정본: `openspec/specs/engine-safety/spec.md`(특히 「자동화 게이트 기동 인터록」, 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」,
  「재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다」, 「자동 진입은 모든 안전 권한의 교집합이다」), `openspec/specs/risk-management/spec.md`(운영 모드 · EXIT_ONLY 삭제).

요지: 자동화 게이트 ON 엔진이 capability attestation 의 **낡음**(만료 · 필요 endpoint 누락)만 이유로 기동을 거부하던 것을, 「청산 전용 기동」(진입 래치 + 청산 루프 유지)으로 바꾼다.
다른 attestation 결함(없음 · 파싱 불가 · 새 형식 · 불완전 · 계좌 불일치)과 다른 인터록 조항은 계속 거부한다.

찾을 것(각각 구체적인 입력 · 실행 순서 · 파일:줄로):
1. **이 설계가 진입(노출 증가)을 열 수 있는 입력이나 경로** — 래치가 서지 않거나 늦게 서는 순서, 래치를 우회하는 mutation 생산자, `Verified` 대신 새 표지를 읽게 되는 소비자가 진입 쪽으로 새는 곳, 재시작·갱신 경합.
2. **손절을 막거나 늦추는 입력** — 청산 전용 기동이 실제로는 exit 관측을 못 세우는 경우(Guardian · 한도 · 게이트웨이 배선 등), 래치가 매도·취소까지 막는 경우.
3. **닫힌 집합의 경계 오류** — 청산 전용으로 내려가면 안 되는데 내려가는 attestation(예: Verify 순서가 가리는 결함, 빈 필드, 형식), 반대로 내려가야 하는데 거부되는 것.
4. **정본과의 모순** — delta 가 정본의 다른 Requirement 와 충돌하는 문장, delta 안의 자기모순, 「실행 중인 엔진은 바뀌지 않는다」 와의 관계.
5. **빠진 소비자** — 새 표지 · 새 사유 코드 · 새 감사 action 을 알아야 하는데 설계가 안 다루는 자리(감시 문장, 콘솔, 상태 출력, 사유 코드 열거, 시험 핀).
6. claim-check.md 의 판정 중 틀렸다고 보는 것.

출력 형식(한국어): 발견마다 `[P0|P1|P2|P3] 제목` 한 줄, 그 아래 근거(파일:줄 · 입력 · 순서), 영향, 제안. P0 = 진입이 열리거나 손절이 막힘, P1 = 정본 위반 · 안전 논증의 구멍, P2 = 빠진 소비자 · 시험 공백, P3 = 문구. 마지막에 「확인했지만 문제없음」 목록(짧게). 추측은 추측이라고 적어라.
