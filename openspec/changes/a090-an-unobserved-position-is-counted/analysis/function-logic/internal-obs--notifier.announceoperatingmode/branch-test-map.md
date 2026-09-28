# Branch Test Map: `Notifier.AnnounceOperatingMode`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:50` nil Notifier 무동작 | 구현 로트가 `internal/obs` 시험에서 확인해 채운다(설계 단계 미측정) | no | no |
| B2 | `:54` 완화 방향 표기 | 같음 | no | no |

> 설계 단계 산출물 — 진입 실측은 구현 로트(3판 tasks 1.1)가 한다. 이 change 가 이 함수에 하는 편집은 동작 무변화 추출뿐이다.
