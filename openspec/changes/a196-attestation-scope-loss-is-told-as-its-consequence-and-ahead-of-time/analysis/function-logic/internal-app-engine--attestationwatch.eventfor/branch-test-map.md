# Branch Test Map: `attestationWatch.eventFor`

- 편집 전(pre-edit). 커버: `../../implementation/base-cover-attestation_watch.out`(`-run 'TestA135'`, HEAD `064f59fd9`, rc 0).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | critical 단계 + 알림 켜짐 → critical 타입 | `TestA135WatchAnnouncesEachStageOnce`(D-2·D-1·만료) · 반대: `TestA135NotificationsOffNeverRaisesCritical` | 해당 없음(기존 분기) | yes — 316.32-318.3 count 1 |
| B2 | 만료 시각 있음 | 날수·만료 시험 | 해당 없음 | yes — 321.27-323.3 count 1 |
| B3 | 단계별 문구 switch | 아래 B4~B9 | 해당 없음 | yes — 324.2-329.17 count 1 |
| B4 | notice·imminent 문구 | `TestA135WatchAnnouncesEachStageOnce` | 해당 없음 | yes — 330.34-332.134 count 1 |
| B5 | expired 문구 | `TestA135InvalidStatesAreRemindedDaily` 등 | 해당 없음 | yes — 333.20-335.111 count 1 |
| B6 | scope_lost 문구(endpoint 이름) | `TestA135RenewalResetsAndScopeLossIsImmediate` | 해당 없음 | yes — 336.22-339.55 count 1 |
| B7 | incomplete 문구 | `TestA135AlertCarriesNoAccount`「incomplete」 | 해당 없음 | yes — 340.23-342.146 count 1 |
| B8 | account_mismatch 문구 | `TestA135AlertCarriesNoAccount`(불일치) | 해당 없음 | yes — 343.28-345.114 count 1 |
| B9 | unreadable 문구 | `TestA135UnreadableAlertCarriesNoAccountOrPath` | 해당 없음 | yes — 346.10-348.109 count 1 |
| B10 | 필드 expires_at | `TestA135AlertCarriesNoAccount` | 해당 없음 | yes — 354.27-356.3 count 1 |
| B11 | 필드 days_left | `TestA135AlertCarriesNoAccount`「notice」「imminent」 | 해당 없음 | yes — 357.56-359.3 count 1 |
| B12 | 필드 missing_endpoints | `TestA135AlertCarriesNoAccount`「scope_lost」 | 해당 없음 | yes — 360.24-362.3 count 1 |
| B13 | 필드 supervised_proof_at | `TestA135AlertCarriesNoAccount`(모든 픽스처가 증거를 실음) | 해당 없음 | yes — 363.25-365.3 count 1 |
| (끝) | Key = `attestation-watch:`+key+`:`+identity | `TestA135WatchAnnouncesEachStageOnce` · `TestA135SameStageOnANewAttestationIsAnnouncedAgain` | 해당 없음 | yes — 367.2-374.3 count 1 |
| (a196) | 범위 상실 제목 = 재시작 거부 단정, 본문 = 누락 endpoint·손절 멈춤·`tossctl verify run` | design §5-7 | 2단계 | 2단계 |
| (a196) | 날수 본문 = 기한 종류·시각·할 일 | design §5-1 | 2단계 | 2단계 |
| (a196) | 거절 Key 에 발급 신원 없음 | design §5-6, 변이 「거절 key 에 발급 신원 되돌리기」 | 2단계 | 2단계 |
