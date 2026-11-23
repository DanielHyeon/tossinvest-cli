# Branch Test Map: `attestationWatch.prepare`

- 편집 전(pre-edit). 커버: `../../implementation/base-cover-attestation_watch.out`(`-run 'TestA135'`, HEAD `064f59fd9`, rc 0).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 판정 패닉 회수 | `TestA135CheckPanicDoesNotStopTheWatch` | 해당 없음(기존 분기) | yes — 271.35-273.20 count 1 |
| B2 | 회수 + 로거 있음 | 없음 | 해당 없음 | **no — 273.20-277.5 count 0**(하네스 log nil; a196 범위 밖) |
| B3 | 신원 변경·첫 호출 → 「보냄」 재설정 | `TestA135RenewalResetsAndScopeLossIsImmediate` · `TestA135SameStageOnANewAttestationIsAnnouncedAgain` · 모든 첫 checkOnce | 해당 없음 | yes — 287.45-290.3 count 1 |
| B4 | healthy → 알릴 것 없음 | `TestA135WatchAnnouncesEachStageOnce`(D-8) | 해당 없음 | yes — 291.29-293.3 count 1 |
| B5 | 이미 보낸 키 | `TestA135WatchAnnouncesEachStageOnce`(같은 D-7) · `TestA135InvalidStatesAreRemindedDaily`(같은 날 만료) | 해당 없음 | yes — 295.17-297.3 count 1 |
| (끝) | 새 단계 → 사건 | 다수 | 해당 없음 | yes — 298.2-298.38 count 1 |
| (a196) | 날수 신원 = 기한 시각: 같은 감독 증거로 재발급 → 재설정 없음 | design §5-2 | 2단계 | 2단계 |
| (a196) | 거절 신원 = 단계+누락 집합+KST 날짜: 같은 날 재발급 2회 → 0, 다음 날 1, 누락 집합 변경 → 1 | design §5-6 | 2단계 | 2단계 |

뒤집히는 기존 단언(편집 뒤): `TestA135SameStageOnANewAttestationIsAnnouncedAgain` 전반(만료된 A → 만료된 B, 같은 날 — 2건·다른 Key 기대 → D5 로 1건).
