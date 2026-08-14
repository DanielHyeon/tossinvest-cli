# a110 — 같은 미해결 분쟁만 영구 차단이 된다

## Why

2026-08-14 운영 화면에서 자동 편입의 저장값과 실행값이 모두 ON인데도 세 보유분이 6일
23시간째 계좌 전체 영구 대사 차단 아래 남아 기준선·익절·손절을 만들지 못했다. 원장과 로그를
대조하니 2026-08-07의 세 blocking 관측은 같은 분쟁의 반복이 아니라 서로 다른 종목의 수량
차이였고, 앞선 두 symbol block은 곧 `ADJUSTMENT_APPLIED`로 정상 해제됐다. 그런데 identity가
없는 단일 연속-failure count가 세 관측을 합쳐 「다시 봐도 해결되지 않은 같은 문제」로 판정했다.

## What Changes

- 영구 승격 streak를 계좌의 모든 dirty comparison이 공유하지 않고, **동일한 canonical
  blocking dispute**만 이어받게 한다.
- 수량 불일치는 정규화한 symbol과 **promotion 전용 exact finite-decimal** local/broker
  quantity tuple, missing order는 기존
  canonical account·market·trading-day·symbol·side·opaque-order identity로 streak를 구분한다.
- 다른 분쟁으로 바뀌거나 exact tuple이 바뀌면 이전 streak는 승격 근거가 되지 않는다. 다만
  ordinary symbol block은 첫 관측부터 지금처럼 즉시·durable-first로 차단한다.
- 같은 분쟁이 threshold만큼 실제로 반복되면 기존 account-wide
  `reconciliation_mismatch_permanent`와 operator-only release를 그대로 만든다.
- 기존 영구 차단은 자동 해제하지 않는다. 고친 빌드에서도 엔진 정지, 세 번의 안정된 공식
  snapshot, clean blocking diff, operator identity·note를 요구하는 기존
  `engine reconcile-resolve`만 해제할 수 있다.
- 영구 승격의 journal write가 실패하면 그 pending 승격은 원래 dispute가 바로 다음 blocking
  comparison에도 있을 때만 재시도한다. clean 또는 다른 dispute로 바뀌면 stale account-wide
  pending 승격을 철회하되 ordinary block의 fail-closed retry는 유지한다.
- 2026-08-07 사고 순서를 개인정보 없는 fixture로 고정하고, release 후 자동 편입이 exit state를
  `SEED`로 만든 뒤 별도 exit-observer 평가가 canonical snapshot을 기록한 다음에만 actionable
  익절·손절선이 나타나는 기존 계약을 회귀 검증한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `reconciliation`: 영구 승격의 「연속 실패」를 동일 canonical blocking dispute의 연속 관측으로
  한정한다. ordinary block, durable ordering, operator-only permanent release는 유지한다.

## Impact

- Code: `internal/reconcile/mismatch.go`의 promotion streak와 그 helper.
- Tests: `internal/reconcile`의 단위·복원·뮤테이션 회귀, `internal/app/engine`의 incident-shape
  adoption 회귀.
- Schema/API/dependency: 없음. streak는 현재 failure count처럼 process-local이며 restart는
  durable permanent block만 복원한다.
- Operations: 현재 운영 block 해제는 이 change의 코드 권한이 아니다. 배포 뒤에도 별도의 사람
  승인과 기존 audited recovery command가 필요하다.
- Safety: LIVE 주문·order preview·토글 변경 없음. 일반 대사 차단과 exit immediacy는 그대로다.
