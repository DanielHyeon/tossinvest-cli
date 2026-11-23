# 변이 원장 — 사본 /home/daniel/tossos-work/a200-mutation HEAD 4f6337a1e81ba7f60bc13b4035678578f3a74e03

## 무변이 대조군: GREEN
```
$ go test ./internal/journal -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.565s
$ go test ./internal/reconcile -run TestA200 -count=1
ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.824s
```

| id | 변이 | 파일 | 결과 | 시험 꼬리 |
|---|---|---|---|---|
| M1 | (가) 거절 유형 검사 삭제(순수 판정) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.337s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.898s |
| M17 | 원장 읽기 앞 거절 유형 조기 반환 삭제(impl-r1 P2-1) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.550s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.789s |
| M18 | (라) 「주문 수량 > 0」 삭제(impl-r1 P2-2) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	11.978s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	8.701s |
| M19 | 조정 행 정렬 뒤집기(created_at, rowid → DESC) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	13.106s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.746s |
| M2 | (나) 「0 조정 정확히 하나」 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.891s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.751s |
| M3 | (나) 「마지막 조정 행」 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	12.717s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	8.734s |
| M4 | (나) kind 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	13.113s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	7.898s |
| M5 | (다) 발의 intent 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	11.420s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	6.492s |
| M6 | (라) PLACE · AMEND 유일 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	8.136s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.098s / FAIL |
| M7 | (라) PLACE 종류 검사 삭제(AMEND 하나도 통과) | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	9.618s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	5.172s |
| M8 | (라) 접수 확정 상태 검사 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	8.359s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.787s |
| M9 | (라) 상한 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	6.406s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.173s |
| M10 | (라) 상한 비교를 < 로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.371s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.339s / FAIL |
| M20 | (라) 체결 누적 상한 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	8.149s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	4.615s |
| M21 | (라) 체결 누적 상한을 < 로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.256s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.090s / FAIL |
| M22 | (라) 체결 주문 수량 상한 비교 삭제 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.390s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.915s |
| M11 | (라) 감소분 대신 new_quantity | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.823s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.886s / FAIL |
| M12 | (라) 감소분의 prev 열을 다른 열(prev_avg_price)로 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	5.223s / FAIL / ok  	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.977s |
| M13 | 인정 기록(표지) 쓰기 생략 | a200_adjustment_close_fill_recognition.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.784s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.868s / FAIL |
| M14 | 인정 뒤 수량 다시 빼기(투영 쓰기) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.661s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.025s / FAIL |
| M15 | 인정 뒤 차단도 쓰기 | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.919s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	3.417s / FAIL |
| M16 | 판정 호출 삭제(오늘 동작으로 되돌림) | position_projection.go | CAUGHT | FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/journal	4.826s / FAIL / FAIL / FAIL	github.com/JungHoonGhae/tossinvest-cli/internal/reconcile	2.816s / FAIL |

살아남은 변이 0 / 22. 끝난 뒤 사본 git status 깨끗함.
