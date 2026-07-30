# Branch Test Map: `statfsProber.Probe`

- Source: `internal/candidate/fsprobe_linux.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

정직한 한계: `statfsProber`는 프로덕션 전용 prober라 이 패키지의 테스트가 **직접** 부르지 않는다. `internal/testenv`의 `FixedFSProber`는 테스트 전용이고, 모든 저장소 테스트는 그것이나 `spaceProber`를 주입한다. 두 분기의 계약은 대역으로 고정되어 있고, syscall 자체의 두 결과는 테스트가 아니라 Linux 프로덕션 경로에서만 실행된다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | statfs가 실패하는 마운트 | 없음 — 대역 `spaceProber.fail`이 같은 계약(probe 에러)을 `TestSpaceItCouldNotMeasureIsNotSpaceItHas`로 고정한다 | 아니오(사후 증거) | 간접만 |
| B2 | `Bsize > 0`이면 잔여 바이트가 측정되고, 아니면 미측정으로 남는다 | `TestDiscoveryStopsWritingBeforeTheLedgerRunsOutOfSpace`(측정됨) / `TestSpaceItCouldNotMeasureIsNotSpaceItHas`(미측정 정지) | 기록됨(§5 D16 — 리뷰가 보존·용량 결론 부재를 지적) | yes |
