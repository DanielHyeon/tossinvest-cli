# 변이 원장 — 사본 /home/daniel/tossos-work/a200-mutation HEAD 8d0655294d7e29a5a2f7d5a64549cd5c8a9220de

## 무변이 대조군: GREEN
```
$ go test ./internal/journal -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.439s
$ go test ./internal/reconcile -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.830s
```

| id | 변이 | 파일 | 결과 | 시험 꼬리 |
|---|---|---|---|---|
| M1 | (가) 거절 유형 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.213s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.895s |
| M2 | (나) 「0 조정 정확히 하나」 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.891s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.004s |
| M3 | (나) 「마지막 조정 행」 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.386s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.851s |
| M4 | (나) kind 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.564s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.525s |
| M5 | (다) 발의 intent 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.531s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.611s |
| M6 | (라) PLACE · AMEND 유일 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.135s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.552s |
| M7 | (라) PLACE 종류 검사 삭제(AMEND 하나도 통과) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.267s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.576s |
| M8 | (라) 접수 확정 상태 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.873s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.766s |
| M9 | (라) 상한 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.583s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.502s |
| M10 | (라) 상한 비교를 < 로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.775s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.041s / FAIL |
| M11 | (라) 감소분 대신 new_quantity | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.369s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.550s / FAIL |
| M12 | (라) 감소분의 prev 열을 다른 열(prev_avg_price)로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.933s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.113s |
| M13 | 인정 기록(표지) 쓰기 생략 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.757s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.070s |
| M14 | 인정 뒤 수량 다시 빼기(투영 쓰기) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.115s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.933s / FAIL |
| M15 | 인정 뒤 차단도 쓰기 | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.327s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.964s / FAIL |
| M16 | 판정 호출 삭제(오늘 동작으로 되돌림) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.580s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.242s / FAIL |

살아남은 변이 0 / 16. 끝난 뒤 사본 git status 깨끗함.
