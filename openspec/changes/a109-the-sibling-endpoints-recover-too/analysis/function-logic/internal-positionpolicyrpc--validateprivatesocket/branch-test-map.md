# Branch Test Map: `ValidatePrivateSocket`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | ① 발행된 socket이 정확-0600이면 통과 — `TestTheAlertSocketIsPrivateToThisUser`가 기동 후 perm 0600을 실측하고 이 함수는 그 기동 경로 안에서 불린다. ② 0700 socket은 거부(완화 누출 금지) — a109 §1.5가 정확-0600을 직접 핀으로 고정하고 뮤테이션 M-C1이 그 핀을 죽여 본다 | 위 칸에 함께 적는다(한 분기, 두 방향) | no | yes(§1.5) |
