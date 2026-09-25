# Branch Test Map: `EntryGate.Clear`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 래치 있음 → 지우고 `revision` 증가 / 없음 → 무변화 | execgw 게이트 시험 (1.4 에서 이름 확정) | no(회귀 핀) | 미측정 |
| **a124 RED** | 해제 세대가 래치 유무와 무관하게 해제 요청마다 1 증가, 다른 사유 세대 불변, `revision` 은 기존대로 | tasks 2.13 | 예정 | 예정 |
