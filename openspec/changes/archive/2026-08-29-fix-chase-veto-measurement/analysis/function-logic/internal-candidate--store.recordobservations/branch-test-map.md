# Branch Test Map: `Store.RecordObservations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 빈 배치는 트랜잭션을 열지 않는다 | `TestAnEmptyReadingIsNotEvidenceOfAbsence` | — (기존 동작) | yes |
| B2 | 모든 행을 먼저 검증한다 | `TestANegativeRequestedCountIsRefusedByTheObservationBoundary` | yes (가드 무력화 시 실패 확인) | yes |
| B3 | 음수 요청 행 수를 실은 행이 배치를 거부시킨다 | `TestANegativeRequestedCountIsRefusedByTheObservationBoundary` | yes | yes |
| B4 | BeginTx 실패 | **커버 없음** (I/O) | no | no |
| B5 | Prepare 실패 | **커버 없음** — 다만 칼럼 수 불일치는 전 store 테스트가 즉시 실패시킨다 | no | no |
| B6 | 행마다 두 새 칼럼이 쓰인다 | `TestATruncatedReadingReachesTheVerdictAsTruncated`(Cycle → 저장 → `MeasureFirstSighting` 왕복) · `TestASessionStartDoesNotStampThePanelAsSeenLate` | yes (F3가 두 대입을 mutation으로 확인) | yes |
| B7 | 정규화할 수 없는 source | **도달 불가** — `validate`가 먼저 거부한다 | n/a | n/a |
| B8 | INSERT 실패 | **커버 없음** — 칼럼 CHECK 자체는 `TestAFreshStoreCarriesTheSameFourConstraints`가 raw handle로 시험하지만 이 경로를 지나지 않는다 | no | no |
| B9 | Commit 실패 | **커버 없음** (I/O) | no | no |

**정직한 커버리지 기록**: B4·B5·B8·B9는 I/O·SQL 실패 경로이고 이 저장소에 그것을 주입하는
장치가 없다. B7은 `validate`가 먼저 거부하므로 도달 불가다. 전부 이 change가 만든
분기가 아니다 — 이 change가 더한 것은 INSERT 칼럼 두 개이고 그것은 B6이 잡는다.
