# 변이 원장 — 사본 /home/daniel/tossos-work/a200-mutation HEAD d939e8795e763fafb9ccd186ae46c060c588158c

## 무변이 대조군: GREEN
```
$ go test ./internal/journal -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.425s
$ go test ./internal/reconcile -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.831s
```

| id | 변이 | 파일 | 결과 | 시험 꼬리 |
|---|---|---|---|---|
| M1 | (가) 거절 유형 검사 삭제(순수 판정) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.334s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.643s |
| M17 | 원장 읽기 앞 거절 유형 조기 반환 삭제(impl-r1 P2-1) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.340s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.515s |
| M18 | (라) 「주문 수량 > 0」 삭제(impl-r1 P2-2) | a200_adjustment_close_fill_recognition.go | CAUGHT(빌드 실패) | FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal [build failed] / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile [build failed] / FAIL |
| M2 | (나) 「0 조정 정확히 하나」 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.309s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.573s |
| M3 | (나) 「마지막 조정 행」 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.333s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.314s |
| M4 | (나) kind 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.294s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.502s |
| M5 | (다) 발의 intent 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.717s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.294s |
| M6 | (라) PLACE · AMEND 유일 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.783s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.092s |
| M7 | (라) PLACE 종류 검사 삭제(AMEND 하나도 통과) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.300s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.377s |
| M8 | (라) 접수 확정 상태 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.725s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.444s |
| M9 | (라) 상한 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.964s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.577s |
| M10 | (라) 상한 비교를 < 로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.669s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.840s / FAIL |
| M11 | (라) 감소분 대신 new_quantity | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.347s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.994s / FAIL |
| M12 | (라) 감소분의 prev 열을 다른 열(prev_avg_price)로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.000s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.453s |
| M13 | 인정 기록(표지) 쓰기 생략 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.132s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	8.102s |
| M14 | 인정 뒤 수량 다시 빼기(투영 쓰기) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.262s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.971s / FAIL |
| M15 | 인정 뒤 차단도 쓰기 | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	11.711s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.314s / FAIL |
| M16 | 판정 호출 삭제(오늘 동작으로 되돌림) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.649s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	6.622s / FAIL |

살아남은 변이 0 / 18. 끝난 뒤 사본 git status 깨끗함.
