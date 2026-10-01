# a126 · issues

## I1. a066 해제 검사의 `unresolved_fill` 철자 결함 — 이 로트에서 수리 (Manager 판정 (가) 2026-10-01)

- **무엇**: `releaseRiskBucketOwner` 검사표 `unresolved_fill`(편집 전 `internal/journal/risk_bucket_owner.go:932`)이 `f.actual_known=0 OR NOT EXISTS(evidence)` 로
  적혀, 해소 정의의 정본 `loadRiskBucketFillTransition`(편집 전 `risk_bucket_fill.go:885`, `actual_known=1 OR EXISTS evidence`)과 반대였다. fills 의 생산 INSERT
  (`risk_bucket_fill.go:1063`, 편집 전)는 `actual_known` 을 늘 0 으로 쓰므로 체결이 한 번이라도 있던 owner 는 해제될 수 없었다.
- **a066 아카이브 교차**: 사용량 수명주기 갭의 기록 `openspec/changes/archive/2026-09-28-a066-add-multi-horizon-risk-buckets/review.md:1124`(#2 "filled_minor is
  never released") · `:1161`("`filled_minor='0'` hides the gap"), 그리고 그 fixture 주석(편집 전 `internal/journal/risk_bucket_owner_test.go:802–804`). 같은 fixture 가
  이 결함도 가렸다 — 체결 없이 FILLED 를 만들었으므로 `unresolved_fill` 에 닿는 체결 행이 없었다.
- **수리**: 해소 조각을 공유 상수 `riskBucketFillActualResolvedSQL` 로 두고 두 자리가 씀(review 1.0.3). 결속 시험 · 변이 F1~F3 CAUGHT.
- **방향 · 거부권**: 해제를 여는 쪽이나 해제 생산 호출자 0(빈 효과), 배선은 tasks 3.1 뒤. **사용자 거부권 항목**으로 보고한다(해제 경로 접촉).
