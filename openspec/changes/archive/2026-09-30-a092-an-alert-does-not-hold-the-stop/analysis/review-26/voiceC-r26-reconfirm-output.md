# 26라운드 보이스 C 재확인 (55b435ea)

판정: #2 · #3 닫힘. 새 시험에 돌린 변이 12개 중 11개 CAUGHT. 새 발견 P1 1 · P2 1.

사본 `/tmp/claude-1000/a092-r26-C2-3290024` 에서 돌렸음.
- `git -C` rev-parse 실패를 단언했음.
- 전용 GOCACHE · `GOFLAGS=-trimpath` 사용.
- 변이는 한 번에 하나씩, cmp 로 복원을 확인했음.
- 무변이 대조군 Q00 은 GREEN.
- 끝난 뒤 삭제했음.

## 판정

1. **#2 WARN 핀 — 닫힘.** Q01(=Y05) CAUGHT. 줄이 없으면 Fatalf, level · claimed_by 도 단언함.
2. **#3 엔드포인트 거절 — 닫힘.**
   - Q02~Q04(=Z01~Z03) CAUGHT.
   - 양성 대조는 실질(200 + NORMAL). 거절마다 모드 · audit 0 을 확인함.
   - 500 본문 변이 Q09 CAUGHT.
   - 빈틈 N2(P2): "alert-control-token" 케이스는 토큰 길이 19 대 18 이라 길이 검사만으로 거절됨.
3. **새 시험 표본 11/12 CAUGHT.**
   - SURVIVED 는 Q06 하나(`logFailure` 의 계좌 가림 제거).
   - 원인: 픽스처 알림기에 Log 가 없어 `logFailure` 가 한 번도 실행되지 않음.
4. **이월 #4~#9** 는 판정하지 않음.

참고: #1(`071e67c1`)은 E1~E4 로 다시 쟀고, 네 경로 모두 키와 함께 기록됨.

## 변이

| ID | 자리 | 변이 | 결과 |
|---|---|---|---|
| Q00 | — | 대조군 | GREEN |
| Q01 | notifier.go:761 | 인수 줄 Warn→Event | CAUGHT |
| Q02 | mode 라우트 | 토큰 검사 제거 | CAUGHT |
| Q03 | 〃 | GET 허용 | CAUGHT |
| Q04 | 〃 | DisallowUnknownFields 제거 | CAUGHT |
| Q05 | alert_control_transport_unix.go:273 | 길이만 비교 | CAUGHT(공유 함수 시험만 — 모드 시험 통과) |
| Q06 | modeops.go:175 | logFailure 무가림 | **SURVIVED** |
| Q07 | modeops.go:184 | detachedAnnouncer 요청 ctx | CAUGHT |
| Q08 | modeops.go:129 | `after := ctx` | CAUGHT |
| Q09 | transport:168 | 500 본문 원문 | CAUGHT |
| Q10 | alertdelivery.go:365 | 반납 선점을 AlreadySettled 로만 | CAUGHT |
| Q11 | engine_mode_release.go:145 | 「이미 전달 처리됨」 | CAUGHT |
| Q12 | riskguardian.go | `if false` | CAUGHT |
| Q13 | runtime.go:463 | `if false && …` | CAUGHT |

## 새 발견

| # | 등급 | 자리 | 무엇 | 제안 |
|---|---|---|---|---|
| N1 | P1 | modeops.go:170-178 | `logFailure` 가림을 지키는 시험 없음(Q06) | Log 를 붙인 픽스처로 단언 |
| N2 | P2 | a092_mode_endpoint_rejects_test.go:180-181 | 틀린 토큰이 길이 차이로 거절됨 | 같은 길이 토큰 |

Recommendation: #2 · #3 닫힘으로 기록하고 착지를 진행하되, N1 은 착지 전에 시험 하나로 막을 것.
