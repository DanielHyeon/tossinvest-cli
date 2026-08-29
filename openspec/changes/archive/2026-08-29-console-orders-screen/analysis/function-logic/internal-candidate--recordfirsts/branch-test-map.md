# Branch Test Map: `recordFirsts`

- Source: `internal/candidate/scan.go`
- GREEN 관측: 2026-07-28 `go test ./internal/candidate/ ./internal/candidatesrc/` — 258 pass, 0 fail.
- RED 표기 규칙: `기록됨(§N)`은 그 결함을 리뷰가 **실행으로 재현**해 `review.md`/`issues.md`에
  남긴 것이고, `아니오(사후 증거)`는 이 evidence pass가 커밋 이후에 작성되어 RED를 직접 관측하지
  못했다는 뜻이다. 없는 관측을 있다고 적지 않는다.

B2·B4·B7·B11은 직접 테스트가 없다. 앞의 둘은 저장소 실패·다중 시장 경로이고 뒤의 둘은 write 실패 경로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 승격이 하나도 없으면 저장소를 읽지 않는다 | `TestAScanDoesNotInventAFirstRankForASourceThatCarriesNone`(간접) | 아니오(사후 증거) | yes |
| B2 | 요약 읽기 실패 | 없음 | 아니오(사후 증거) | 미커버 |
| B3 | 아직 비어 있는 컬럼을 가진 후보만 고른다 | `TestALaterScanDoesNotMoveTheFirstPriceOrTheFirstRank` | 기록됨(issues.md 4) | yes |
| B4 | 다른 시장의 요약은 무시 | 없음 | 아니오(사후 증거) | 미커버 |
| B5 | 승격된 심볼마다 두 사실을 시도한다 | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` | 기록됨(issues.md 4) | yes |
| B6 | 기준선이 없고 가격이 있으면 기록 | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` | 기록됨(§3 P0-3 — 기준선이 조용히 옮겨가던 결함) | yes |
| B7 | 기준선 write 실패는 이름을 남기고 pass는 계속 | 없음 | 아니오(사후 증거) | 미커버 |
| B8 | 기준선 성공은 `FirstPrices`로 센다 | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` | 아니오(사후 증거) | yes |
| B9 | 순위를 나르지 않는 원천은 순위를 만들지 않는다 | `TestAScanDoesNotInventAFirstRankForASourceThatCarriesNone`, `TestALaterScanDoesNotMoveTheFirstPriceOrTheFirstRank` | 기록됨(§4 P1 — D20) | yes |
| B10 | 순위 write의 세 결과를 가른다 | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw` | 아니오(사후 증거) | yes |
| B11 | 순위 write 실패는 이름을 남기고 계속 | 없음 | 아니오(사후 증거) | 미커버 |
| B12 | 창 안의 순위만 `FirstRanks`로 센다 — 창 밖은 성공도 실패도 아니다 | `TestAScanRecordsTheFirstPriceAndTheFirstRankItSaw`, `TestARankFromOutsideTheIdentityWindowIsNotStored` | 기록됨(§4 P1 — 만료 9분 전 148위 행 재현) | yes |
