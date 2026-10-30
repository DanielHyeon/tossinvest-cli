# Branch Test Map

| 분기 | 확인 | 상태 |
|---|---|---|
| worker done과 finishedFlag 불일치 | 교체·relaunch 거절 | 선택 fixture 확인; 전체 분기 미검증 |
| shutdown 또는 worker timeout | 오류 보존, exec 0 | 선택 fixture 확인; 전체 분기 미검증 |
| 성공 종료 | 실제 완료와 owner Close 뒤 relaunch | 선택 fixture 확인; 전체 분기 미검증 |
