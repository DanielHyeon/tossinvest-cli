# Branch Test Map: `stageKey`

- 편집 전(pre-edit). 커버: `../../implementation/base-cover-attestation_watch.out`(`-run 'TestA135'`, HEAD `064f59fd9`, rc 0).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 날수 단계 → `stage:days` | `TestA135WatchAnnouncesEachStageOnce`(D-7·D-6·D-3·D-2·D-1 키가 서로 다름) | 해당 없음(기존 분기) | yes — 304.56-306.3 count 1 |
| (끝) | 거절 단계 → `stage:KST날짜` | `TestA135InvalidStatesAreRemindedDaily`(같은 날 1, 다음 날 2) | 해당 없음 | yes — 307.2-307.65 count 1 |
| (a196) | 누락 집합이 키에 들어감: 같은 날 하나 더 빠지면 새 알림 | design §5-6, 변이 「누락 집합을 key 에서 빼기」 | 2단계 | 2단계 |
