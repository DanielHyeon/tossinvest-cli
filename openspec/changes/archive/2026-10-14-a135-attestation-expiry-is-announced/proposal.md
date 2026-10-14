# a135 — capability attestation 의 만료는 미리 알려진다 (attestation-expiry-is-announced)

## Why

자동화 게이트가 켜진 엔진은 기동할 때만 capability attestation 을 읽는다(`verifyGate` 4~6, `internal/app/engine/interlock.go`).
attestation 이 만료되었거나 필요한 endpoint 를 잃은 상태에서 엔진이 재시작하면 기동이 거절되고, 그때는 청산(손절) 관측 루프도
돌지 않는다. 보호 주문이 UNWIRED 인 지금 손절을 실행하는 것은 엔진 프로세스 하나뿐이다.

`engine-safety` 「등급화된 알림」은 「자격증명 만료 임박」을 critical 로 정했고 `obs.EventCredentialExpiring` 이 그 자리로 선언되어
있지만, **생산 발행자가 0 이다**(`grep -rn EventCredentialExpiring` 비시험 결과는 선언과 등급표뿐). 콘솔의 attestation 칸(a063)은
남은 시간과 72시간 경고를 보여 주지만 운영자가 콘솔을 열어야만 보인다.

Manager 와 이 change 의 실측(2026-10-14, 읽기 전용 · 만료 · endpoint 메타데이터만 — `analysis/verify/live-attestation.md`):

- 운영 attestation 만료 `2026-10-23T01:39:36Z`(발급 2026-09-23, 30일 유효). 남은 시간 약 9일 4시간.
- endpoint 집합은 엔진의 `RequiredEndpoints()` 여덟을 모두 덮는다(POST orders · POST cancel 포함). POST 두 개의 근거는
  감독하 실주문 검증 기록 `2026-09-09T03:23Z`(KR) 이다.
- 갱신 타이머(`tossos-attest.timer`, 6시간마다)는 돌고 있으나 **매번 거절된다**(soak 판독 완전성 실패 7 cycle). 설치된 단위는
  거절을 성공으로 가린다(a063 4.2 미설치 — systemd 는 매번 `Finished`).
- 갱신이 성공하더라도 **POST 범위를 잃는다**: 갱신은 감독 증거를 유효 기간(30일) 안의 것만 싣고(`acceptSupervised` B6 — 나이가
  기간 이상이면 조용히 건너뜀), 빈 감독 증거를 거절하지 않는다(`BuildAttestation` — 거절 분기 B1 · B3 · B4 뿐). 9월 9일 증거는
  2026-10-09 에 기간을 벗어났다. 그러므로 지금 갱신이 성공하면 GET 만 담은 attestation 이 좋은 파일을 덮어쓰고, 다음 재시작은
  `ErrEndpointNotAttested` 로 거절된다.

어느 쪽이든 운영자가 미리 알지 못하면 재시작이 알려 준다. 그것은 너무 늦다.

## What Changes

1. **읽기 전용 판정** — 엔진이 재시작할 때 읽을 바로 그 파일(`AutomationStatus.AttestationFile`)을 다시 읽어, 인터록과 **같은
   술어**(`attest.Attestation.Verify` + `RequiredEndpoints()` + 엔진 자격 증명의 계좌)로 「지금 재시작하면 무장하는가」를 판정하고,
   통과하면 만료까지 남은 날수(D-N, 올림)를 계산한다. 규칙 사본 없음.
2. **알림** — 기존 알림기(`Context.Notifier`, a133 전송 경계 경유):
   - D-7 ~ D-3: normal(`broker.credential_expiry_notice`, 새 normal 타입) — 하루 한 번.
   - D-2 · D-1: critical(`broker.credential_expiring`, 기존 선언).
   - 만료 · 필요 endpoint 누락(갱신 뒤 POST 범위 상실 포함) · 계좌 불일치 · 파일 없음/읽기 불가: 즉시 critical.
   - 중복 방지: attestation 신원(발급 · 만료 시각) × 판정 단계마다 한 번, 재시작이 거절될 상태는 고쳐질 때까지 하루 한 번.
     신원이 바뀌면(갱신) 다시 셈.
   - 알림이 꺼진 엔진(`notifications.enabled=false`)은 critical 대신 normal 타입으로 기록만 함(a091 · a095 와 같은 규칙 —
     보낼 수 없는 critical 행이 진입을 멈추지 않게).
3. **어디서 도는가** — 엔진 런타임의 **보조 실행자**(`Auxiliary`, 감독 집합 밖) 하나, 기동 직후 한 번 + 1시간마다. 게이트가 꺼진
   엔진에서는 아무것도 읽지 않고 취소를 기다린다. 판정 한 번의 패닉 · 오류는 그 판정에서 끝나고(알림기 호출 안의 패닉은 회수하지
   않음 — 알림기 잠금을 영구히 잡지 않게, design D3), 실행자의 정지는 자기 이벤트 타입
   (`engine.attestation_watch_failed`, normal)으로 로그만 남긴다 — 진입 래치 · 운영 모드 · 다른 루프에 닿지 않는다.
4. 회귀 방지 시험(RED → GREEN): 경계(D-8 없음 · D-7 normal · D-3 normal · D-2/D-1 critical · 만료 critical), 범위 상실 critical,
   중복 방지 · 갱신 뒤 재무장, 판정 실패가 엔진 · 실행자를 멈추지 않음, 알림 문구에 계좌 없음, 게이트 꺼짐 무동작,
   그리고 「기간 밖 감독 증거로 갱신하면 POST 가 빠진 attestation 이 나온다」 특성 시험(`internal/soak`, 생산 무편집).

## What Does NOT Change

- 인터록 · attestation 형식 · 유효 기간 · soak 기준 · 갱신 명령 · systemd 단위(a063 의 범위). 이 change 는 갱신을 실행하지도
  예약하지도 않는다. 실주문 0, 운영 토글 0.
- 실행 중인 엔진의 동작: 판정은 관측일 뿐 `Verified` · `EntryPermitted` · 게이트 · 운영 모드를 바꾸지 않는다. 다만 critical 알림이
  전달 한도를 넘겨 실패하면 기존 규칙(engine-safety 「등급화된 알림」)대로 진입이 잠긴다 — 청산 · 손절에는 닿지 않는다.
- 브로커 호출 0 추가(로컬 파일 하나를 1시간마다 읽음). 감독 루프 집합(넷) 그대로.
- 콘솔: a063 의 attestation 칸이 이미 남은 시간 · 누락 endpoint 를 보여 준다. 콘솔 무편집.

## Impact

- Affected specs: `engine-safety`(요구 1개 추가 — 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」).
- Affected code: 새 `internal/app/engine/attestation_watch.go`(판정 · 실행자), `internal/obs/event.go`(이벤트 상수 2개 추가 —
  등급표 무편집), `cmd/tossctl/engine.go` `engineRuntime` 의 보조 실행자 목록 한 줄,
  `cmd/tossctl/a098_the_engine_registers_a_sender_test.go` 의 기대 집합. 시험: 새 `internal/app/engine/a135_*_test.go` ·
  `internal/soak/a135_*_test.go`.
- 위험 등급: 관측 · 알림 경로. 주문 · 손절 · 사이징 · 인터록 판정을 바꾸지 않는다. 비례 원칙(docs/WORKFLOW.md)상 변이 원장 · 다중
  적대 리뷰는 의무가 아니다. 게이트가 요구하는 수정된 기존 함수의 FLM 은 `engineRuntime` 과 그 시험 함수 하나.
- 배포: 이미지 재빌드 + 엔진 재시작(사람). 지금 attestation 은 10-23 까지 유효하므로 그 전 재시작은 무장한다. D-7 경계는
  `2026-10-16T01:39Z` 이다.
- 운영 후속(이 change 밖, 사람): 10-23 전에 감독하 실주문 검증 재실행 + soak 판독 완전성 회복이 없으면 갱신이 불가능하거나
  POST 를 잃는다(design 「운영 사실」).
