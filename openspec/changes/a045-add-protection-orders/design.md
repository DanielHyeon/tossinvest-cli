## Context

엔진은 local exit를 수행하지만 broker-resident protection은 UNWIRED다. 조건주문 실제 형태는 `verify-execution-capability` attestation이 권위이며 StockOS/KIS 동작을 Toss에 대입하지 않는다.

## Goals / Non-Goals

**Goals:** 체결 수량을 항상 하나의 브로커 보호 청구권으로 덮고 crash/restart에서 복구한다.

**Non-Goals:** paper/shadow/canary 주문, 익절 조건주문, 미검증 capability 추정, 운영자 승인 없는 LIVE flip.

## Decisions

1. 기본 보호는 attested `SINGLE+MARKET` stop 하나이며 익절은 기존 local reduce-only 경로다.
2. saga 상태는 PLANNED→REGISTERING→ACTIVE→REPLACING/TRIGGERED→CLOSED/RECONCILE이며 mutation attempt와 broker IDs를 영속한다.
3. 조건주문 예약을 포함한 sellable ownership을 하나의 수량 ledger로 계산해 일반 매도와 중복 예약하지 않는다.
4. 보호선 상승은 replace 의미가 attested된 경우에만 수행하며 새 ID가 생기면 새 ID를 먼저 회수 가능한 상태로 기록한다.
5. `ProtectionReady`는 market/type/quantity/session capability가 일치할 때만 WIRED다.
6. UI 소유 카테고리는 `exit-protection`이다. 공통 exit 기준선과 브로커 상주 보호를 별도 section으로 보여주되 같은 effective trigger의 관계를 설명한다.
7. attestation 전 descriptor는 enabled OFF, capability unavailable이며 읽기 전용이다. attestation 후 지원 type이 SINGLE+MARKET으로 확정되어도 activation은 OFF를 유지하고 trigger는 a041 effective line에서만 파생한다.
8. 보호 약화 요청은 before/after, 영향 포지션·수량, 보호 공백 가능성, 적용 시점과 3초 지연 확인을 표시한다. 단순 강화는 같은 수준의 위험 확인을 요구하지 않는다.
9. capability attestation은 strict versioned matrix다. account/profile, market, session, conditional type,
   trigger source, quantity/min/max/partial-fill, persistence, reservation, idempotency, replace의 atomicity와
   continuous coverage, tool/build, evidence digest, issued/expiry를 모두 typed claim으로 고정한다. unknown
   field/legacy format/symlink/non-regular file/소유자·권한 불일치는 fail-closed다.
10. atomic 또는 continuous replace가 attested되지 않은 profile은 `ProtectionReady=WIRED`가 될 수 없다.
    지원 profile의 첫 fill 관측 뒤 같은 supervisor cycle이자 1초 안에 registration attempt를 arm하고,
    2초 안에 ACTIVE가 아니면 신규 진입을 latch하고 `PROTECTION_GAP`으로 reconcile/escalate한다.
11. flatten은 조건주문 cancel 확인에 최대 2초를 부여한다. 그 안에 terminal cancel과 broker sellable
    quantity를 권위 있게 확인한 경우에만 기존 reduce-only liquidation을 실행한다. response loss,
    trigger 경합 또는 모호한 상태는 `IN_DOUBT`로 격리하고 신규 exposure를 차단하며 최우선 reconcile과
    사람 emergency action을 요구한다. oversell 가능성이 있는 blind liquidation은 금지한다. 이 2초
    경로와 연속 보호를 증명하지 못하는 capability profile은 활성화 대상이 아니다.

## Risks / Trade-offs

- [등록과 journal commit 사이 crash] → idempotency key와 reconciliation으로 UNKNOWN을 격리한다.
- [부분체결 중 oversell] → broker sellable snapshot과 local reservations의 보수적 최솟값을 사용한다.
- [보호 등록 실패] → exposure-raising automation을 latch하고 즉시 운영 경고·reconcile/attested emergency 정책으로 전환한다.
- [UI가 미검증 capability를 사용 가능한 기본값처럼 보임] → unavailable/read-only 상태와 attestation provenance를 필수 descriptor로 고정한다.

## Migration Plan

schema와 official client를 dormant 상태로 배포한 뒤 전체 gate 완료 시 운영자가 LIVE engine을 승인한다. canary 단계는 없다. older binary는 active broker protection을 감독할 수 없으므로 binary-only rollback은 금지한다. rollback은 신규 진입을 끄고 현재 controller가 보호를 유지한 채 reconciliation을 완료하거나, attested cancel/flatten 절차로 active protection과 position을 0으로 만든 뒤 DB/WAL/SHM migration 전 backup을 복원하는 경우에만 허용한다.

## Open Questions

실제 profile claim과 evidence digest는 사람 attestation 전에는 채울 수 없다. 그 전 구현 허용 범위는 strict schema/parser, dormant domain, fake official gateway와 `UNWIRED/OFF` 경로뿐이며 실제 conditional mutation과 WIRED 전환은 금지한다.
