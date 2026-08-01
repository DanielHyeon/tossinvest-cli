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
12. protection attestation은 `TossOS/protection-capability-attestation/ed25519/v1` domain의
    Ed25519 signed envelope로만 수용한다. envelope version, domain, algorithm, key ID와 canonical
    payload bytes를 길이 구분된 signing input에 모두 결박하고 payload/signature/public key는
    unpadded base64url의 단 하나의 canonical 표현만 허용한다. payload는 strict JSON을 다시
    canonical marshal한 bytes와 byte-for-byte 같아야 하므로 unknown/duplicate field, field order,
    whitespace, 숫자·시간·base64의 비정본 표현은 거절한다.
13. trust root는 attestation과 분리된 디렉터리에 운영자가 미리 배포한 exact basename의 read-only
    canonical JSON keyset이다. immutable startup policy가 absolute path, expected owner UID와 canonical
    file SHA-256을 모두 pin하며 attestation path/owner에서 어느 값도 추론하지 않는다. 런타임 writer,
    TOFU, 자동 갱신, fallback key, embedded production key는 없다. direct parent/file의 symlink,
    non-regular file, hard link, owner/mode, pinned digest 또는 post-read identity 불일치는 fail-closed다.
    attestation parent/file은 runtime UID 소유 `0700/0600`이다. trust parent는 pinned owner 소유이고
    group/other write가 없으며 non-owner traverse가 가능해야 하고, trust file은 pinned owner 소유
    `0444`다. 이 분리는 root와 non-root provisioning UID 모두 지원하며 같은 UID가 파일을 치환해도
    immutable digest mismatch로 거절한다.
14. 각 trust key는 unique key ID, exact `PROTECTION_ATTESTATION_SIGNER` role, Ed25519 algorithm,
    32-byte public key, not-before/not-after와 explicit `ACTIVE`/`REVOKED` status를 가진다. `REVOKED`는
    revoked-at/reason이 모두 있어야 하며 과거 발행분을 포함해 해당 key의 모든 서명을 즉시
    fail-closed하는 hard revocation이다. `ACTIVE`에는 revocation metadata가 있으면 안 된다.
    rotation은 old/new ACTIVE key의 validity overlap으로 수행하며 attestation의 issued/expiry 전체와
    현재 검증 시간이 선택된 key window 안에 있어야 한다.
15. cryptographic signature 성공만으로 WIRED가 되지 않는다. matrix time, exact runtime account/
    profile/market/session/type/trigger/quantity, 두 verifier tool version/build와 외부 evidence bytes의
    digest를 모두 다시 검증한 `VerifiedProtectionCapability`만 후속 wiring의 입력 후보가 된다.
16. verifier는 호출자가 path, owner UID, digest, key 또는 시간을 주입하는 API를 제공하지 않는다.
    고정된 canonical policy source와 clock을 소유하는 sealed verifier만 parse와 최종 authorization을
    수행한다. 현재 production에는 이 verifier를 provision하는 constructor가 없으므로 zero value는
    fail-closed하며 계속 `UNWIRED/OFF`다. policy는 positive generation을 갖고 attestation/root 경로,
    root owner와 digest를 pin한다. verifier는 한 번 본 최고 generation/digest를 mutex로 기억하고 낮은
    generation 또는 같은 generation의 다른 digest를 이후 호출에서도 거절한다. parse 이후 caller가
    제공한 scope/evidence의 비교와 해시를 먼저 끝내고, 최종 authorization 직전에 policy/root/key/
    status/window/signature와 verifier clock을 다시 읽고 검증한다.
17. signed matrix의 시간은 exact UTC만 허용하고 evidence, capability와 trust key 배열은 identity key의
    strict ascending order여야 하며 중복을 거절한다. account reference는 separator 없이 8-14 ASCII
    digits만 허용한다. authorization 결과는 전체 matrix가 아니라 exact matched scope와 capability row
    하나만 보유하며 tool map도 deep copy한다.
18. repository insert는 revision-1 `PLANNED`만 허용한다. 이후 persistence API는 이벤트별 메서드만
    공개하고 저장된 row를 `Transition`에 넣어 attempt/broker lineage를 검증한 다음 단일 SQL
    `WHERE revision=?` CAS로 갱신한다. 마지막 event kind와 전체 canonical event fingerprint를 함께
    영속하며 CAS 패자는 이 둘과 timestamp/event-specific 결과가 완전히 같은 revision+1 row일 때만
    idempotent success다. schema v1 migration에서 빈 event identity는 정확한 revision-1 `PLANNED`에만
    허용하고 그 외 legacy row는 시작 시 fail-closed한다. flatten의 public API는 decision time을 내부 clock에서
    얻으며 ALLOWED enum만으로 실행할 수 없다. exact scope/broker/quantity와 2초 deadline에 결박되고
    copy 간 atomic state를 공유하는 opaque one-shot authorization을 한 번 consume해야 한다.

## Risks / Trade-offs

- [등록과 journal commit 사이 crash] → idempotency key와 reconciliation으로 UNKNOWN을 격리한다.
- [부분체결 중 oversell] → broker sellable snapshot과 local reservations의 보수적 최솟값을 사용한다.
- [보호 등록 실패] → exposure-raising automation을 latch하고 즉시 운영 경고·reconcile/attested emergency 정책으로 전환한다.
- [UI가 미검증 capability를 사용 가능한 기본값처럼 보임] → unavailable/read-only 상태와 attestation provenance를 필수 descriptor로 고정한다.

## Migration Plan

schema와 official client를 dormant 상태로 배포한 뒤 전체 gate 완료 시 운영자가 LIVE engine을 승인한다. canary 단계는 없다. older binary는 active broker protection을 감독할 수 없으므로 binary-only rollback은 금지한다. rollback은 신규 진입을 끄고 현재 controller가 보호를 유지한 채 reconciliation을 완료하거나, attested cancel/flatten 절차로 active protection과 position을 0으로 만든 뒤 DB/WAL/SHM migration 전 backup을 복원하는 경우에만 허용한다.

## Open Questions

실제 profile claim과 evidence digest는 사람 attestation 전에는 채울 수 없다. 그 전 구현 허용 범위는 strict schema/parser, dormant domain, fake official gateway와 `UNWIRED/OFF` 경로뿐이며 실제 conditional mutation과 WIRED 전환은 금지한다.

signer/signature/trust-root 형식과 회전·폐기 semantics는 Decisions 12-15로 동결했다. 다만 실제
production keyset과 사람 검증 evidence는 아직 배포되지 않았다. trust root 또는 서명 envelope가
없으면 parser는 실패하고 시스템은 계속 `UNWIRED/OFF`다. 이 구현과 사용자의 포괄 승인만으로
`ProtectionReady=WIRED`, 운영 토글 또는 주문 권한을 만들 수 없다.
