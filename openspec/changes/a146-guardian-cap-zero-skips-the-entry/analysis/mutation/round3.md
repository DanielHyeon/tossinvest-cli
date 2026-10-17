# a146 변이 원장 round3 — 구현 리뷰 생존 자리

- 사본: 저장소 밖 연결 워크트리 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/a146-mut3-953210` @ `35db12b7` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 GREEN)
- 시험 묶음(변이마다 같음): `go test -count=1 -tags tossos_testseams -run '^(TestA146.*|TestA137.*|TestAnExistingGuardianCapRefusalIsNotAScopeRefusal|TestTheEntrySkipTypeIsMadeOnlyWhereTheCensusSaysItIs)$' ./internal/app/engine`

| 변이 | 파일 | 결과 | 처음 실패한 시험 |
|---|---|---|---|
| R10 q_final 거절이면 건너뜀 | `internal/app/engine/strategy_market_handoff_delivery.go` | CAUGHT | `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` |
| R12 원인 무관 상한 코드면 건너뜀 | `internal/app/engine/strategy_market_handoff_delivery.go` | CAUGHT | `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` |
| M35 타입 또는 문구 | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestTheEntrySkipTypeIsMadeOnlyWhereTheCensusSaysItIs` |
| M36 문구만 | `internal/app/engine/strategy_market_handoff_delivery.go` | CAUGHT | `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` |
| M31 다른 시장에 귀속 | `internal/app/engine/strategy_entry_supervisor.go` | CAUGHT | `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| P1 다중 오류 거름 제거 | `internal/app/engine/a146_entry_skip.go` | SURVIVED | — |
| P2 트리 전체 errors.As | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` |
| P3 감싸개 문구 대조 제거 | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` |
| P4 FX 권위를 알림에서 뺌 | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146TheProductionNoticeIsOneWarningLine` |

## 분류

- 비동등 **8 / 8 CAUGHT**: 구현 리뷰 둘의 생존 자리 R10 · R12 · M35 · M36 · M31 과 승격 탐침 P2 · P3, 그리고 P4(FX 권위). 이 결과로 Join · 결함 감싸개 모양이 시험으로 고정됨.
- **P1 SURVIVED는 중복 판정이었음.** 다중 오류에는 `errors.Unwrap` 이 nil 을 돌려주므로, 다중 오류 거름이 없어도 그 모양은 스킵이 되지 않음. 그래서 거름을 지웠음(KISS, 이 기록 뒤 커밋).
  지운 뒤에도 같은 모양은 `TestA146OnlyAPureSkipIsSkippedEveryLookalikeStops` 의 「skip joined with a fault」 · 「fault joined with a skip」 행이 잡음. 하네스의 P1 행은 이제 NOT-REACHED.
- 영수증: `make test-race` 통과(rc 0 — RACE_PACKAGES 와, 새로 등록한 a146 두 시험을 포함한 엔진 이름 목록).
