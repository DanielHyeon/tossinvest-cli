## Why

자동 진입 인터록은 `ProtectionReady`를 요구하지만 production profile은 아직
보호주문 배선을 `UNWIRED`로 고정한다. 보호 saga 코드가 존재한다는 사실만으로는
브로커에 상주하는 손절이 KR·US 각각의 주문 유형, 세션, 수량과 교체 의미에서 실제로
동작한다고 증명할 수 없다. 따라서 두 시장의 진입을 연결하기 전에 현재 build와 정확히
일치하는 attestation으로 보호 능력을 제한하고, 등록부터 재기동 복구까지의 수렴 동작을
production assembly에 배선해야 한다.

## What Changes

- KR과 US 각각에 대해 계좌, profile, 시장, 주문 유형, 세션, 수량, trigger source,
  replace 의미와 broker conditional-order identity/query/dedup capability를 포함한
  서명·버전·만료·build/evidence digest attestation을 검증한다.
- Attestation은 고정 trust root에서 검증하며 key ID, algorithm allowlist, revocation,
  rotation overlap, monotonic serial, 최대 수명과 trusted-time rollback 방지 계약을
  통과해야 한다. 증명되지 않은 capability 또는 key 상태는 해당 시장을 `UNWIRED`로 만든다.
- 현재 runtime scope가 attestation과 정확히 일치하고 보호 supervisor/saga가 배선된
  경우에만 해당 시장의 `ProtectionReady`를 `WIRED`로 계산한다. 누락, 만료, unknown
  field, digest 불일치 또는 검증되지 않은 조합은 fail-closed한다.
- 보호 등록, 수량 증가, 더 안전한 trigger로의 교체, 취소, 재기동 복구와 reconciliation을
  durable state machine으로 연결하고 재시도·중복 이벤트에도 하나의 유효한 보호로
  수렴시킨다.
- submit/cancel 결과가 unknown이면 attested exact broker identity/query로만 확인한다.
  idempotency·dedup을 증명하지 못한 submit은 재제출하지 않고 reconciliation에 남기며,
  orphan 보호는 신규 exposure를 닫은 상태에서 정확한 소유권이 확인될 때만 조정한다.
- 엔진 프로세스가 사라져도 attested broker-side protection이 유지되는 계약을 fixture와
  격리된 broker test로 검증하고, 복구 전 신규 노출 증가는 차단한다.
- KR과 US readiness는 독립적으로 판정한다. 한 시장의 검증 실패가 다른 시장의 이미
  유효한 readiness를 무효화하지 않지만, 실패한 시장의 신규 진입은 허용하지 않는다.
- 이 change는 운영 토글, autostart, lane 또는 LIVE 승인을 켜지 않는다. live 보호주문을
  사용하는 검증은 별도의 사람 승인 절차 없이는 실행하지 않는다.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `protection-orders`: KR·US 시장별 attestation 범위 안에서 broker-resident protection의
  등록·교체·취소·복구·reconciliation이 멱등적으로 수렴하고 프로세스 손실에도 보호를
  유지하는 계약을 확정한다.
- `engine-safety`: production `ProtectionReady`를 고정 상수가 아니라 현재 build와 정확한
  market capability attestation 및 실제 protection wiring의 교집합으로 계산하고,
  불일치 시장의 exposure-raising dispatch를 차단한다.

## Impact

- `internal/app/engine`: protection supervisor 구성, 시장별 readiness publication과 startup/
  dispatch interlock 배선.
- `internal/execgw` 및 보호 영속 저장소: idempotent lifecycle, broker ownership 확인,
  restart/reconciliation 수렴.
- Attestation loader/validator: signed evidence, scope, expiry, build digest와 파일
  ownership/permission, trust-root/key lifecycle과 trusted-time 검증.
- 테스트와 운영: live mutation 없이 fixture/isolated broker로 실패·재시작·중복을 검증한다.
  기본값은 계속 `UNWIRED`/OFF이며 사람 승인 범위는 넓어지지 않는다.
