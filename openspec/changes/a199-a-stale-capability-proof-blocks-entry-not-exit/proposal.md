# a199 — 낡은 capability 증명은 진입만 막고 청산은 막지 않는다

- change id: `a199-a-stale-capability-proof-blocks-entry-not-exit`
- 위험 등급: **High-risk** (기동 인터록 · 주문 권한 · 손절 경로 가용성 — 안전 부트스트랩 5)
- 사용자 승인: 2026-11-24 「1번 올린다로 진행해줘」 「지금부터 작성하고 구현해줘」
- 사전 증거: `analysis/evidence-phase0.md`(HEAD `9738a0c32`, AST 12, CodeGraph 1.6.0) — 아래 주장의 `파일:줄` 과 AST 분기 번호는 전부 그 문서의 것이고, 1단계에서 현재 HEAD 로 다시 뽑아 대조한다.

## Why

자동화 게이트가 켜진 엔진은 capability attestation 이 만료됐거나 엔진이 쓰는 endpoint 를 덮지 않으면 **기동을 통째로 거절**한다(`verifyGate` B6@536 → `runInterlock` B2@400 → `NewContext` B15@588 — Context 없음, 루프 0). 거절은 진입뿐 아니라 exit 관측(손절·익절) · 대사 · 편입 · 체결 감지까지 끈다.

2026-11-23 운영에서 이 일이 일어났다. 감독 증거(주문 접수·취소 POST 의 실주문 증명, 유효 30일)가 11-18 에 기간을 벗어났고 — 발급자는 그런 증거를 오류 없이 건너뛰므로(`soak/attest.go:331-333`) attestation 은 「POST 둘이 빠진 채 정상 발급」됐다 — 실행 중인 엔진은 계속 돌다가, 11-23 19:59 호스트 재부팅 뒤 기동이 `ErrEndpointNotAttested` 로 거절됐다. 그때부터 보유 전부가 자동 손절 없이 남았고 알림은 없었다(거절은 감사 로그와 stderr 뿐, ntfy·outbox 0).

이 거절이 지키려던 것은 코드 주석 한 문장이다(`interlock.go:32-34`, 커밋 `30563ae53`): *"an engine that trades with an unverified capability is the thing this interlock exists to prevent."* 그런데 실패는 비대칭이다 — 증명 안 된 경로로 **진입**하면 손절 없는 새 포지션이 생기지만, 증명이 낡은 경로로 **청산**을 시도하다 실패하면 잃는 것은 「시도하지 않았을 때와 같은 상태」뿐이다. 정본은 같은 비대칭으로 보호 미배선 조항을 이미 옮겼다(`interlock-gates-entry-not-exit`, 2026-08-29: 「그것을 이유로 청산 루프를 거부하는 것은 같은 결함을 더 넓은 범위에 유지한다」). 그 change 는 「attestation 요구는 전부 불변」으로 이 조항을 명시적으로 남겼다. 이 change 가 그 남은 조항에 같은 원리를 적용한다.

덧붙여 지금 생산 빌드는 보호 조립이 두 시장 모두 `Wired:false`(`protection_wiring.go:41-42`)라 게이트 ON 엔진의 모든 매수가 이미 `broker_protection_not_wired` 로 거절된다 — 즉 검증된 런타임이 실제로 하는 일은 **보유의 관리와 청산뿐**인데, 그 일만 하는 엔진을 「진입 능력의 증명이 낡았다」는 이유로 세우고 있다.

## What Changes

1. **attestation 의 낡음(만료 · 필요 endpoint 누락)은 기동 거절이 아니라 「청산 전용 기동」이 된다.** 조건은 닫힌 집합이다: 파일이 읽히고(없음·파싱 불가·새 형식 아님), 형식 번호가 이 빌드가 아는 유효한 값이고(1 이상), 온전하고(계좌·만료·endpoint 목록이 비지 않고 **목록의 항목이 전부 비지 않은 문자열**), **해석된 계좌를 `Verify` 와 같은 비교로 가리키고**, 실패가 만료와 endpoint 누락뿐일 때(판정은 인터록과 감시가 함께 부르는 공용 함수 하나). 그 밖의 attestation 실패(없음 · 무효 · 불완전 · 빈 endpoint 항목 · 형식 번호 무효 · 새 형식 · 계좌 불일치 — 만료가 계좌 불일치를 가리는 경우 포함)와 **다른 모든 인터록 조항**은 지금처럼 기동을 거절한다.
2. **청산 전용 기동에서 도는 것**: 감독 루프 집합은 그대로(reconcile driver + 편입 · exit observer · 체결 감지 · 휴면 strategy-entry 외곽), 보조 실행자도 그대로. Guardian 은 위험 감소 발급을 위해 존재한다.
3. **청산 전용 기동에서 막히는 것**: 노출을 증가시키는 mutation 전부. 집행은 기존의 한 곳(`Gateway` 의 진입 검사 — `EntryGate` 래치)에 **새 사유 하나**로 건다. 래치는 매 기동마다 attestation 에서 다시 유도되므로 재시작으로 풀리지 않는다. 전략 진입 worker 는 `Verified` 를 계속 요구하므로 만들어지지 않는다. `Verified` 의 뜻(인터록 전 조항 충족)은 바꾸지 않는다 — 청산 쪽 소비자만 새 표지를 읽는다.
4. **운영자에게 알린다**: 청산 전용 기동은 감사 기록(새 action) · 구조 로그 · 기동 출력 한 줄을 남기고, attestation 감시(a135·a196)가 그 상태를 **사실대로** 말한다 — 「지금 재시작하면 기동이 거부된다」는 낡음에 대해 더는 참이 아니므로 「새 진입이 막힌 청산 전용으로 뜬다(돌고 있다)」로 바꾼다 — 날수 · 지남 통지의 「그 뒤 재시작이 거부된다」도 같다. 「지금 청산 전용으로 돌고 있다」 알림은 파일이 아니라 엔진의 기동 표지에 걸려, 파일이 회복된 뒤에도 재시작 전까지 「증명은 회복됐다 — 재시작하면 정상으로 뜬다」(일반 등급)로 하루 한 번 나간다. 거절이 유지되는 상태(계좌 불일치 · 무효 · 없음)의 문장은 그대로다. 「보호가 프로세스 수명에 묶여 있다」는 기동 문장은 청산 전용 기동에서도 나간다.
5. 인터록 파일 머리 주석의 낡은 서술(8항 Protection, "no degraded mode")과 `soak attest` 의 낡은 출력 문구(「gate will refuse to start」류)를 정본에 맞춘다.

## 비목표 (하지 않는 것)

- **실행 중 전환.** 감시와 attestation 재읽기는 기동 표지 · 루프 집합 · 진입 래치를 실행 중에 바꾸지 않는다(정본 「실행 중인 엔진의 동작은 바뀌지 않는다」 유지). 판정은 기동 시점뿐이다. 주문마다의 진입 권한은 정본 「자동 진입은 모든 안전 권한의 교집합이다」가 정하며 이 change 는 그것을 바꾸지 않는다.
- 운영 모드(원장 `operating_modes`)에 새 트리거·새 모드를 더하지 않는다. `EXIT_ONLY` 를 재도입하지 않는다(add-core-domain: 「ENTRY_BLOCKED 와 행동 동일」) — 이 상태는 원장이 아니라 attestation 파일에서 매 기동 유도된다.
- 새 설정 키·토글 없음. 게이트 OFF 경로는 한 글자도 바뀌지 않는다(불변식 3).
- 거절이 유지되는 기동(계좌 불일치 등)에 ntfy 알림을 더하는 것 — 엔진 밖 heartbeat 감시(a134)가 「엔진 down」을 알린다. 사유를 싣는 것은 후속 후보.
- 콘솔 「능력 증명 attestation」 섹션의 문구·판정(`readAttestation` 은 자체 규칙) — 후속 Story 후보. 이 change 뒤 콘솔은 낡은 attestation 에 「게이트가 열리지 않는 이유」를 계속 보여 주며 이는 부정확해진다(Story limits 에 고정).
- 발급자가 낡은 감독 증거를 건너뛰는 규칙, 감독 증거의 유효 기간(30일), `verify run` 절차.
- 보호 배선(`Wired`) · a105.

## 영향

- 정본: engine-safety 「자동화 게이트 기동 인터록」 MODIFIED(조항 2 의 귀결 분리 + Scenario 「attestation 만료 상태 기동」 수정 + Scenario 추가), 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」 MODIFIED(낡음의 귀결 문장). ADDED 없음 — 「낡은 증명」 문단이 「자동화 게이트 기동 인터록」 안에서 strict 를 통과한다(intent C; 1단계 사실 정정 — 초안은 「ADDED 2~3」).
- 코드(1단계 확정 — `analysis/edit-surface.md`): 기존 함수 13 — `verifyGate` · `runInterlock` · `logGateDecision`(interlock.go), `NewContext`(engine.go — **a166 census 키 1**), `runEngineRun`(cmd/tossctl/engine.go), `Context.ExitObserver`(exitwiring.go — a198 착지 뒤), `Context.ReconcileDriver`(reconcileloop.go), `judgeAttestation` · `attestationWatch.eventFor` · `newAttestationWatch`(attestation_watch.go), `AllReasonCodes`(execgw/failclosed.go), `runSoakAttest` · `newSoakAttestCmd`(cmd/tossctl/soak.go 문구). 새 파일: `internal/attest` 계좌 비교 메서드, `internal/app/engine` 공용 낡음 판정. 상수 · 패키지 변수: `internal/audit` action, `internal/execgw` 사유 · `latchOrder` · golden.
- 운영: 배포 뒤 낡은 attestation 으로 재시작하면 엔진이 청산 전용으로 뜬다. 정상 모드로 돌아가려면 지금과 같다 — `verify run` → `soak attest` → 재시작.
