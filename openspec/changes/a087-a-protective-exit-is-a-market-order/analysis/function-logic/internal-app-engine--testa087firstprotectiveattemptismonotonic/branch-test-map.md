# Branch Test Map: `TestA087FirstProtectiveAttemptIsMonotonic`

| Branch | Scenario | Evidence |
|---|---|---|
| B1 | 시간대 로드 성공 | A087 focused race GREEN |
| B2 | 첫 보호 시도는 하한가 | A087 focused race GREEN |
| B3 | 첫 시도 기록 뒤 관측가 fallback | A087 focused race GREEN |

시간대 로드 실패 분기는 환경 오류 방어이며 별도 주입하지 않았다. fixture nil admission에서 누적되던 휴면 자원 RED는 `analysis/20261030-daily-refresh/dormant-resource-red.log`에 있다.
