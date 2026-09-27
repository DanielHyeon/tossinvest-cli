# engine-safety — a095 delta (3판)

> **ADDED만 쓴다.** 기존 「등급화된 알림」과 「같은 조건의 critical 알림은 재알림 창 안에서 한 번만
> 전송한다」의 본문은 그대로다.
>
> **3판은 결정된 것만 싣는다**(사용자 결정 2026-09-25, 원문은 `proposal.md` §0). 결정이 덮지 않아
> 비워 둔 것 — 설정 거부 · include 지정 시도 실패 · 연기된 후보의 등급(Q2), 엔진 개설 포지션의 수량 증가
> 검사의 비교 기준(Q3), 수량 증가 사실의 종류 · 등급 · 키(Q4) — 은 답이 온 뒤 이 파일에 더한다.
> 2판의 「편입 기록이 없다는 이유로 수량 증가 검사를 건너뛰어서는 안 된다 … 보호 상태가 없는 경우로
> 보고」 SHALL과 그 시나리오는 결정 (3)에 따라 삭제했다.

## ADDED Requirements

### Requirement: 무관리 보유 보고의 등급은 사실이 정한다

대사(reconcile) 루프가 발견한 무관리 보유 가운데 **편입이 켜져 있어(`adoption.enabled`) 엔진이 보호하기로 했는데 이번 판정에서 편입되지 않은 보유**는 critical 등급으로 보고해야 한다(SHALL).

> **미결(Q2(c))**: 시세 읽기 오류 · 종목 관측 없음 · 관측 묵음으로 **연기되어** 편입되지 않은 후보도 오늘의
> 코드에서는 같은 사유로 모이므로 이 문장에 든다. 답이 연기분을 빼면 이 문장을 좁힌다.

운영자가 고른 상태의 무관리 보고는 critical이어서는 안 된다(SHALL NOT). 운영자가 고른 상태는
`adoption.exclude_symbols`에 있는 종목, 그리고 `adoption.enabled`가 거짓이고 `adoption.include_symbols`에
없는 종목이다. 이 둘의 보고는 이 요구사항 이전과 같은 등급으로 남아야 한다(SHALL — 정본 exit-policy의
`adoption.enabled` false 동등성과 안전 불변식 3).

알림이 꺼진 엔진(`notifications.enabled`가 거짓)에서는 이 요구사항의 어떤 보고도 진입 게이트 래치나
운영 모드 승격(ENTRY_BLOCKED)을 불러서는 안 된다(SHALL NOT — 알림을 끈 것은 운영자가 고른 상태다).

exit 관측 루프가 손절 판정 전에 내는 무관리 보고는 critical이어서는 안 된다(SHALL NOT — 그 발신은
같은 사이클의 모든 손절 판정 앞에 서므로, critical의 배달 대기가 손절 판정을 늦춘다).

서로 다른 발신 자리의 무관리 보고는 서로 다른 event key를 써야 한다(SHALL — 같은 key는 outbox의 한 행으로
합쳐져 먼저 온 자리의 사유만 남는다).

이 요구사항은 손절가 · 최초 위험 · 진입 기준의 값을 바꾸지 않는다(SHALL NOT — 그 값들은 이미 보고된
모든 R의 분모다).

**근거**: `alert_outbox`에 `exit.position_unmanaged` 행은 2026-08-07(13행)과 2026-09-25(16행) 두 측정에서 모두
0이다 — 무관리 보고는 원장에 남은 적이 없다. 동시에, 같은 이벤트 종류를 통째로 critical로 올리면 알림을 끈
기본 설정 엔진이 손으로 산 보유 하나로 진입을 멈추고, exit 관측 루프의 손절 판정 앞에 배달 대기가 선다
(a095 2라운드 리뷰, 사용자 결정 2026-09-25).

#### Scenario: 편입이 켜진 엔진의 편입 실패
- **WHEN** `adoption.enabled`가 참인 엔진의 대사 루프가 편입 후보를 이번 판정에서 편입하지 못하면
- **THEN** 그 무관리 보고는 critical 등급으로 durable outbox에 기록된다

#### Scenario: 의도적으로 제외한 종목
- **WHEN** `adoption.exclude_symbols`에 있는 종목의 보유가 발견되면
- **THEN** 무관리 보고는 critical이 아니고 outbox 행 · 진입 게이트 래치 · 운영 모드 승격을 만들지 않는다

#### Scenario: 편입을 켜지 않은 기본 설정
- **WHEN** `adoption.enabled`가 거짓이고 include 지정도 없는 엔진이 보유를 발견하면
- **THEN** 무관리 보고의 등급과 결과는 이 요구사항 이전과 같다

#### Scenario: 알림이 꺼진 엔진
- **WHEN** `notifications.enabled`가 거짓인 엔진에서 편입 실패가 발생하면
- **THEN** 진입 게이트 래치도 운영 모드 승격도 일어나지 않는다

#### Scenario: exit 관측 루프의 무관리 보고
- **WHEN** exit 관측 루프가 판정 대상이 아닌 보유를 발견하면
- **THEN** 그 보고는 critical이 아니고, 그 사이클의 손절 판정 앞에서 critical 배달을 기다리지 않는다

#### Scenario: 두 발신 자리
- **WHEN** 같은 포지션에 대해 exit 관측 루프와 대사 루프가 각각 무관리 보고를 낸다
- **THEN** 두 보고의 event key가 다르다

#### Scenario: 보고는 기준을 바꾸지 않는다
- **WHEN** 이 요구사항이 적용된다
- **THEN** 진입가 · 최초 손절 · 최초 위험은 종전 값 그대로다
