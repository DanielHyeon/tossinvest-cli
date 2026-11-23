# 변이 원장 — 사본 /home/daniel/tossos-work/a200-mutation HEAD 8560c26223b6c6dc8d320d5038bcadf1d6606fb6

## 무변이 대조군: GREEN
```
$ go test ./internal/journal -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.460s
$ go test ./internal/reconcile -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.412s
```

| id | 변이 | 파일 | 결과 | 시험 꼬리 |
|---|---|---|---|---|
| M1 | (가) 거절 유형 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.720s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.691s |
| M2 | (나) 「0 조정 정확히 하나」 삭제 | a200_adjustment_close_fill_recognition.go | SURVIVED | ok  	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.553s / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.490s |
| M3 | (나) 「마지막 조정 행」 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT(빌드 실패) | FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal [build failed] / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile [build failed] / FAIL |
| M4 | (나) kind 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.370s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.139s |
| M5 | (다) 발의 intent 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.345s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.800s |
| M6 | (라) PLACE · AMEND 유일 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.659s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.870s |
| M7 | (라) PLACE 종류 검사 삭제(AMEND 하나도 통과) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.494s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.800s |
| M8 | (라) 접수 확정 상태 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.625s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.202s |
| M9 | (라) 상한 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.784s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.770s |
| M10 | (라) 상한 비교를 < 로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.448s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.498s / FAIL |
| M11 | (라) 감소분 대신 new_quantity | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.399s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.707s / FAIL |
| M12 | (라) 감소분의 prev 열을 다른 열(prev_avg_price)로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.219s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.559s |
| M13 | 인정 기록(표지) 쓰기 생략 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.396s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.461s |
| M14 | 인정 뒤 수량 다시 빼기(투영 쓰기) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.296s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.713s / FAIL |
| M15 | 인정 뒤 차단도 쓰기 | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.372s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.954s / FAIL |
| M16 | 판정 호출 삭제(오늘 동작으로 되돌림) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.224s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.275s / FAIL |

살아남은 변이 1 / 16. 끝난 뒤 사본 git status 깨끗함.
