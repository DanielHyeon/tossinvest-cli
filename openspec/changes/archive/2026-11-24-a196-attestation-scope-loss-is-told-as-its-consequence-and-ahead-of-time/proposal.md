# a196 — 능력 검증 범위 상실은 그 결과로, 그리고 미리 알려진다

> **상태: 1단계(계약·증거), Normal(관측 전용 보조 실행자·알림).** Story `STORY-TOS-a196`(FEAT-TOS-003). 작성: Manager(2026-11-23).
> 분기 주장의 AST 는 조사 산출물(세션 스크래치 `inv-E/*.ast.json`, HEAD `a9dd53400`)에서 인용했고, 이 change 안 사본과 현재 HEAD 재확인은 tasks 1.3 이다 — 그 전까지 「조사 인용」이다.

## Why

2026-11-18 ~ 11-23 운영(Manager 실측 — attestation 파일·알림 원장·엔진 로그·컨테이너 프로세스, 시각·상태만):

- 주문·취소 endpoint 는 사람이 감독한 실계좌 검증(`tossctl verify run`)의 증거로만 attestation 에 실리고, 그 증거는 30 일 유효하다. 11-18 18:00 KST 의 6 시간 갱신부터 증거가 기간을 넘어 attestation 에서 POST 두 개가 빠졌다(11-23 현재 파일의 endpoint 는 조회 9 개).
- 감시는 그때부터 critical 알림을 냈다 — 5 일 동안 21 건, 전부 전달됨. 그러나 운영자는 이것을 「재시작하면 엔진이 뜨지 않는다」로 읽지 못했다.
- 11-23 19:58 호스트 재부팅 뒤 엔진 자동 시작이 기동 인터록 4 항(capability attestation)으로 거절됐고, 그 뒤 엔진 프로세스가 없었다. 보유 종목의 자동 손절이 전부 멈췄다.
- 사전 통지는 없었다. 증거가 만료되기 전에 「며칠 뒤 감독 검증이 만료된다」고 말한 알림이 0 건이다.

코드 근거:

| 사실 | 근거 |
|---|---|
| 알림의 타입(ntfy 첫 태그)은 `broker.credential_expiring`, 제목은 「자동매매 검증서(attestation)에 필요한 주문 권한이 없다」, 본문은 원인을 「…쓰였을 수 있다」로, 결과를 조건문(「만료되거나 무효인 attestation 으로 엔진이 재시작하면…」)으로 말하고 할 일이 없다 | `internal/app/engine/attestation_watch.go` `eventFor`(:337-339 부근 문자열, 조사 인용) |
| 중복 방지 key 가 발급 신원(발급·만료 시각)을 담아, 6 시간 갱신마다 새 행이 생긴다 — 「하루 한 번」이 하루 네 번이 된다 | `stageKey`(:303-308)·`prepare` if@287·outbox key `"attestation-watch:"+key+":"+identity`(:370, 조사 인용). 정본 engine-safety 「파일이 바뀌면 단계를 새로 센다」가 이 동작을 허용한다 |
| 날수 사전 통지는 attestation 자신의 만료만 센다. 6 시간 갱신이 만료를 매번 30 일 뒤로 밀어 사실상 발화하지 않는다 | `judgeAttestation` 의 `daysLeftCeil(att.ExpiresAt.Sub(now))`(AST: call daysLeftCeil@112 — 종전 표기 :118 은 B12 case@118 이었음, 1.3 정정) |
| 감독 증거의 시각은 이미 계산되지만 판정에 쓰이지 않는다 | `judgeAttestation` 의 `ProofAt: binding(att.SupervisedBy)`(:91), `binding`(:134-154, 직접 읽음) |
| 발급자는 같은 유효 기간 값으로 attestation 만료와 감독 증거 수용을 정한다 | `internal/soak/attest.go` `acceptSupervised(…, c.Validity, …)`(:242)·`ExpiresAt: now.UTC().Add(c.Validity)`(:265, 직접 읽음) |
| a135 는 감독 증거 기한의 사전 통지를 일부러 뺐다 — 이유는 「발급자 수용 규칙의 사본이 된다」 | a135 design D6(조사 인용) |

## What Changes

1. **미리 알림**: 감시가 「재시작이 거절되기 시작할 시각」을 attestation 만료와 **감독 증거의 기한** 중 이른 쪽으로 센다. 감독 증거의 기한 = 가장 먼저 기간을 벗어날 POST 감독 증거의 시각 + 그 attestation 의 유효 기간(발급~만료). 파일에 실린 값만으로 계산하므로 발급자 규칙을 복제하지 않는다(a135 D6 의 우려를 지키는 방식). D-7~D-3 일반 등급, D-2·D-1 critical — 기존 날수 규칙 그대로. 기한이 지났는데 파일이 아직 판정을 통과하면(갱신 전·갱신 실패 중) 새 단계로 하루 한 번 critical(Manager 결정 2026-11-23 Q4). 기한은 파일에 **실린** 증거 기준이라, 다른 시장의 더 새 증거가 검증 기록에 있으면 경보가 이르게 나올 수 있다 — 놓치는 방향이 아니므로 받아들인다(Q1).
2. **결과로 말하기**: 거절 상태 알림의 제목과 본문 첫 문장이 「지금 엔진을 재시작하면 기동이 거부된다」를 사실로 말하고(Q3 — 기기의 제목 표시는 미실측), 본문이 자동 손절이 멈춘다는 것과 할 일(`tossctl verify run` 재실행)을 말한다. 타입은 그대로 둔다.
3. **하루 한 번**: 거절 상태는 파일이 다시 발급되어도 같은 상태면 그날 다시 알리지 않는다. 날수 단계는 파일이 아니라 기한을 신원으로 센다. 보냄 상태는 프로세스 수명 안에서만 유지되어 재시작 뒤 현재 단계를 한 번 더 알릴 수 있음을 정본에 적는다(Q6).

## 이 change 가 하지 않는 것

- 엔진이 기동을 거부한 사실을 운영자에게 닿게 하는 새 경로. 엔진 밖 생존 감시(a134, `tossos-heartbeat.timer`)가 그 일을 한다 — 2026-11-23 이 호스트에 그 타이머가 **설치돼 있지 않았다**(Manager 확인). 설치는 운영 절차다(design §4).
- 갱신 타이머가 POST 를 잃는 파일로 덮어쓰는 동작의 변경, 갱신 결과를 `issued` 로 기록하는 것(a063 영역).
- 새 알림 타입, 콘솔 화면의 감독 증거 표시, 콘솔 자동 시작의 3 초 관찰 창.
- 감독 검증의 자동 실행. 실주문 검증은 사람이 터미널에서 확인 문자열을 쳐야 한다.

## Capabilities

### Modified Capabilities

- `engine-safety`: 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」 — 날수의 기준에 감독 증거의 기한을 더하고, 기한 경과·파일 통과 상태를 하루 한 번 critical 로, 거절 상태 알림의 문장과 중복 방지 신원을 고친다. 「만료 N일」 Scenario 4 개를 「기한」 기준으로 고치고 Scenario 6 개 추가(의도한 편집 16 — `analysis/spec-edit-intent.md`).

## Impact

- 기존 함수 본문 편집(gate ⑤ FLM 대상): production 4 — `internal/app/engine/attestation_watch.go` 의 `judgeAttestation`(기한 계산·새 단계), `attestationWatch.eventFor`(문장·key), `stageKey`·`attestationWatch.prepare`(신원). 시험 2 — `a135_attestation_watch_internal_test.go` 의 `TestA135StageBoundaries`(표를 기한 기준으로)·`TestA135SameStageOnANewAttestationIsAnnouncedAgain`(새 규칙으로 재작성). `attestationVerdict.critical`·`daysLeftCeil`·헬퍼 `a135Attestation` 은 무편집(tasks 1.4, AST). 새 계산은 새 함수로.
- a166 census 밖(1.4 실측 0건). 나머지 a135 시험(endpoint 이름·계좌 누출 고정 등)은 그대로 통과해야 한다.
- 관측 전용: 브로커 호출 0, 진입·청산·운영 모드에 닿지 않는다(정본 그대로). 원장 스키마·attestation 파일 형식 변경 없음 — 옛 이미지가 같은 파일을 그대로 읽는다.
- 알림 수: 범위 상실 하루 4 건 → 1 건. 감독 증거 만료 전 7 일 동안 하루 1 건(마지막 이틀은 critical), 기한이 지났는데 갱신이 실패하는 동안 하루 1 건 critical. 엔진 재시작마다 현재 단계 1 건이 더 나갈 수 있다.
