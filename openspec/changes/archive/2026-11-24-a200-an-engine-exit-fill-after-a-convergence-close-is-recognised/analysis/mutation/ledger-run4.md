# 변이 원장 — 사본 /home/daniel/tossos-work/a200-mutation HEAD 25d9aea2d9c7ced910ad4fd9dfe23a8bcbc7f6a5

## 무변이 대조군: GREEN
```
$ go test ./internal/journal -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.081s
$ go test ./internal/reconcile -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.508s
```

| id | 변이 | 파일 | 결과 | 시험 꼬리 |
|---|---|---|---|---|
| M1 | (가) 거절 유형 검사 삭제(순수 판정) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.588s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.445s |
| M17 | 원장 읽기 앞 거절 유형 조기 반환 삭제(impl-r1 P2-1) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.307s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.054s |
| M18 | (라) 「주문 수량 > 0」 삭제(impl-r1 P2-2) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.553s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.036s |
| M19 | 조정 행 정렬 뒤집기(created_at, rowid → DESC) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.343s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.018s |
| M2 | (나) 「0 조정 정확히 하나」 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.452s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.061s |
| M3 | (나) 「마지막 조정 행」 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.999s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.660s |
| M4 | (나) kind 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.077s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.511s |
| M5 | (다) 발의 intent 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.996s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.045s |
| M6 | (라) PLACE · AMEND 유일 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	9.493s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	5.998s |
| M7 | (라) PLACE 종류 검사 삭제(AMEND 하나도 통과) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	10.096s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.787s |
| M8 | (라) 접수 확정 상태 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.032s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.316s |
| M9 | (라) 상한 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.443s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.201s |
| M10 | (라) 상한 비교를 < 로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.645s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.218s / FAIL |
| M11 | (라) 감소분 대신 new_quantity | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	9.039s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	5.176s / FAIL |
| M12 | (라) 감소분의 prev 열을 다른 열(prev_avg_price)로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	8.107s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	5.123s |
| M13 | 인정 기록(표지) 쓰기 생략 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	8.591s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.791s / FAIL |
| M14 | 인정 뒤 수량 다시 빼기(투영 쓰기) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	7.400s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.965s / FAIL |
| M15 | 인정 뒤 차단도 쓰기 | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.089s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.742s / FAIL |
| M16 | 판정 호출 삭제(오늘 동작으로 되돌림) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.831s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.643s / FAIL |

살아남은 변이 0 / 19. 끝난 뒤 사본 git status 깨끗함.
