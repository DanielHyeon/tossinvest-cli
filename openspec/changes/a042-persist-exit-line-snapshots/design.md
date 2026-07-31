## Context

a041 snapshot은 계산 정본이지만 crash recovery의 정본은 SQLite journal이어야 한다. 현재 exit state 필드와 새 설명 필드 사이의 원자성·migration 경계를 정한다.

## Goals / Non-Goals

**Goals:** snapshot 핵심 상태를 additive schema로 저장하고 재시작 후 단조 보호를 복구한다.

**Non-Goals:** 과거 row 일괄 backfill, UI 자체 구현, 정책 재선택, 주문 제출.

## Decisions

1. 기존 exit state에 nullable policy version, next target/protection, observation source/at을 additive migration으로 추가한다. baseline/high-water/rung은 기존 권위 필드를 유지한다.
2. 판정 결과와 exit event는 한 SQLite transaction에서 기록한다. 주문 intent는 기존 arm-before-submit 계약을 유지한다.
3. legacy NULL은 현재 필드로 재구성 가능한 값만 계산하고 불가능한 값은 unknown으로 남긴다. 과거 보호선을 낮추는 default를 넣지 않는다.
4. recovery는 saved와 recomputed를 각각 하나의 coherent snapshot 후보로 검증한 뒤 더 안전한 후보
   전체를 선택한다. field별 max를 섞어 존재하지 않는 policy/rung/target tuple을 합성하지 않는다.
5. console read model은 `saved`, `recomputed`, `effective`를 분리하고 각 값의 version/source/observed-at과 stale/unknown reason을 제공한다. 불명 상태를 화면 기본값으로 대체하지 않는다.
6. protection/high-water만 같은 policy digest 안에서 단조 비교할 수 있다. rung은 같은 immutable policy digest에서만 비교하고 next target/protection은 선택된 policy snapshot에서 다시 파생한다. policy digest가 다르거나 후보 우열을 결정할 수 없으면 해당 포지션 entry/자동 판정을 격리한다.
7. snapshot ID, decision ID, observation ID와 policy ID/version/digest는 exit event와 같은 transaction에 position generation별로 저장한다.

## Risks / Trade-offs

- [migration 중 crash] → schema version transaction과 reopen fixture로 검증한다.
- [추가 write 비용] → 한 evaluation transaction 안에서 갱신하고 별도 시계열 row 남발을 피한다.

## Migration Plan

백업 후 additive migration, reopen/crash test, read compatibility 순으로 적용한다. 현재 이전 binary는
더 높은 `user_version`을 거부하므로 binary-only rollback을 주장하지 않는다. rollback은 engine을
중지하고 DB/WAL/SHM 세트를 보존한 뒤 migration 전 backup을 원자적으로 복원하고 broker open order를
reconcile하는 절차다. destructive down migration은 하지 않는다.

## Open Questions

없음.
