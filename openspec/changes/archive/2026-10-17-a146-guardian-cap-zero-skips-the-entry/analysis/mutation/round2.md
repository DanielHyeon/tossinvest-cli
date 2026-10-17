# a146 변이 원장

- 사본: 저장소 밖 연결 워크트리 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/a146-mut2-38256` @ `01ca7ff1` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 GREEN)
- 시험 묶음(변이마다 같음): `go test -count=1 -tags tossos_testseams -run '^(TestA146.*|TestAnExistingGuardianCapRefusalIsNotAScopeRefusal|TestTheEntrySkipTypeIsMadeOnlyWhereTheCensusSaysItIs)$' ./internal/app/engine`

| 변이 | 파일 | 결과 | 처음 실패한 시험 |
|---|---|---|---|
| M1 코드 조건 제거(동등 — 선언) | `internal/app/engine/strategy_first_leg_admission.go` | SURVIVED | — |
| M2 sentinel 조건 제거 | `internal/app/engine/strategy_first_leg_admission.go` | CAUGHT | `TestA146OtherGuardianCapRefusalsStayFaults` |
| M3 다른 sentinel | `internal/app/engine/strategy_first_leg_admission.go` | CAUGHT | `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| M4 sentinel 을 오류 전체에서 찾음 | `internal/app/engine/strategy_first_leg_admission.go` | CAUGHT | `TestA146OtherGuardianCapRefusalsStayFaults` |
| M5 원인이 스킵 타입을 잃음 | `internal/app/engine/strategy_first_leg_admission.go` | CAUGHT | `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| M6 코드를 AUTHORITY_MISMATCH 로 | `internal/app/engine/strategy_first_leg_admission.go` | CAUGHT | `TestA146ACapZeroPrecheckIsATypedSkip` |
| M7 기록 호출 제거 | `internal/app/engine/strategy_first_leg_admission.go` | CAUGHT | `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| M8 스킵을 주기 오류에 모음 | `internal/app/engine/strategy_market_handoff_delivery.go` | CAUGHT | `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| M9 스킵 판정 제거 | `internal/app/engine/strategy_market_handoff_delivery.go` | CAUGHT | `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| M10 스킵에서 주기를 끝냄(뒤 범위 굶음) | `internal/app/engine/strategy_market_handoff_delivery.go` | CAUGHT | `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` |
| M11 감독자 기록자 설치 무력화 | `internal/app/engine/strategy_entry_supervisor.go` | CAUGHT | `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| M12 계수 증가 제거 | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| M13 첫 원인 덮어쓰기 | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| M14 종목 dedup 제거 | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| M15 일 키 바뀜에 기억 비우기 제거 | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| M16 빈 일 키에서 침묵(a137 모양) | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146AnUncountableAccountDayStillAnnouncesTheSkip` |
| M17 생산 알림 배선 제거 | `internal/app/engine/strategy_entry_supervisor.go` | CAUGHT | `TestA146TheProductionSupervisorCarriesTheSkipNotice` |
| M18 감독자가 알림 옵션을 버림 | `internal/app/engine/strategy_entry_supervisor.go` | CAUGHT | `TestA146ASkipIsCountedAndAnnouncedOncePerDayAndSymbol` |
| M19 스냅숏 계수 칸 제거 | `internal/app/engine/strategy_entry_supervisor.go` | CAUGHT | `TestA146ACapZeroEntryNoLongerLatchesAnEffectiveMarket` |
| M20 알림 줄에서 가격 제거 | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146TheProductionNoticeIsOneWarningLine` |
| M21 알림 이벤트 이름 바꿈 | `internal/app/engine/a146_entry_skip.go` | CAUGHT | `TestA146TheProductionNoticeIsOneWarningLine` |

## 분류

- **비동등 20 / 20 CAUGHT**(CAUGHT(build) 0). round1(7395cc02)의 build 둘(M14 · M20)은 변이식을 고쳐 행동으로 잡힘, M16(빈 일 키 침묵)은 일 키 seam +
  `TestA146AnUncountableAccountDayStillAnnouncesTheSkip` 로 CAUGHT.
- **M1 동등(사전 선언)**: 원인 사슬에 `ErrStrategyQuantityZero` 를 싣는 q_final 거절은 `EXISTING_GUARDIAN_CAP`(riskguardian_qfinal.go:148) 하나 — 다른 원인 실은
  `qFinalRefusal` 자리(FX · `decision.Refusal` · 주간 예약 · 투영 · 발급 단계)는 사이징 오류를 싣지 않음(grep 전수). 코드 조건은 방어용.
- design D8 의 「a137 갈래 앞으로」는 동등(입력 서로소)으로 원장에 넣지 않음.
