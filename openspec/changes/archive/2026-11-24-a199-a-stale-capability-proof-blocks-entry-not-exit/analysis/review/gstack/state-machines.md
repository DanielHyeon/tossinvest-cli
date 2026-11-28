# a199 gstack review — state-machines 패스 (HEAD 6c30b91b3, 603046a23..HEAD)

state-machines 전문가 서브에이전트의 산출(읽기 전용, 시험 미실행). 근거 줄 번호는 6c30b91b3 기준.

## 1. 기동 결과 표 (게이트 ON, 따로 적지 않으면)

반환된 모든 Context 에서 불변식 셋이 성립. ExitObserver · ReconcileDriver 는 `RunsExitSide()` 일 때만(exit_only_start.go:101, exitwiring.go:323, reconcileloop.go:368). Guardian 유지 = engine.go:593-595.

| 입력 | 결과 | Verified | ExitOnly | EntryPermitted | 래치 | Guardian | 감사 action | 청산 루프 | 결정 갈래 |
|---|---|---|---|---|---|---|---|---|---|
| 게이트 OFF | 수락(게이트 off — `engine run` 은 `errEngineGateOff`, cmd/tossctl/engine.go:233) | F | F | F | 없음 | 없음 | 설정 기록만 | 없음 | interlock.go:405-409 |
| 유효 | 수락 | T | F | Protection==Wired(이 빌드 F) | 없음 | 있음 | accepted | 있음 | exit_only_start.go:70-74 → interlock.go:422-436 |
| 만료만 | 청산 전용(expired) | F | T | F | 있음 | 있음 | exit_only | 있음 | attest.go:222 → exit_only_start.go:75,92,96 → interlock.go:603-608 → 417-419 → exit_only_start.go:108-129 |
| 범위 상실만 | 청산 전용(scope_lost) | F | T | F | 있음 | 있음 | exit_only | 있음 | attest.go:233 → exit_only_start.go:76,96 |
| 만료 + 범위 상실 | 청산 전용(expired,scope_lost) | F | T | F | 있음 | 있음 | exit_only | 있음 | exit_only_start.go:56-65,96 |
| 파일 없음 | 거절(ErrMissing) | F | F | F | 없음 | (Context 없음) | refused | — | attest.go:165 → interlock.go:541-546, 412-415 |
| 못 읽음 · 깨짐 | 거절(ErrMalformed) | F | F | F | 없음 | — | refused | — | attest.go:167/172 |
| 불완전 | 거절(ErrIncomplete) | F | F | F | 없음 | — | refused | — | attest.go:213-219 → exit_only_start.go:76-78 |
| 빈 endpoint 항목 | 낡은 파일: 거절(ErrIncomplete) / Verify 통과 파일: 수락(Q-B) | F/T | F | F/보호 | 없음 | —/있음 | refused/accepted | —/있음 | exit_only_start.go:85-89 / 71-73 |
| format_version < 1 | 낡은 파일: 거절(ErrMalformed) / Verify 통과: 수락 | F/T | F | F | 없음 | —/있음 | refused/accepted | —/있음 | exit_only_start.go:81-84 / 71-73 |
| 형식 초과 | 거절(ErrFormatTooNew, Load 에서 — 판정 함수 상한은 도달 불가) | F | F | F | 없음 | — | refused | — | attest.go:174-177 |
| 계좌 불일치(만료 아님) | 거절 | F | F | F | 없음 | — | refused | — | attest.go:228 |
| 만료 + 계좌 불일치 | 거절(ErrAccountMismatch, 사슬에 ErrExpired 없음) | F | F | F | 없음 | — | refused | — | exit_only_start.go:92-95 |
| 낡음 + 다른 조항 실패(게이트웨이 · 키 · 한도 · 7b · 7c) | 그 조항으로만 거절 | F | F(603 미도달) | F | 없음 | — | refused | — | interlock.go:568-586 |
| 낡음 + 계좌 미해석 · Guardian 생성 실패 | 더 앞에서 거절 | F | F | F | 없음 | — | refused | — | engine.go:500-503 · 568-571 |
| 낡음 + nil 진입 게이트 | 거절(ErrAutomationGateRefused+ErrGatewayRequired) | F | **반환 status 에 T**(Context 없음) | F | 없음 | — | refused | — | exit_only_start.go:110-117(생산 도달 불가) |
| 낡음 + 감사 기록 실패 | 거절 | F | status 에 T(Context 없음) | F | 버려지는 게이트에 세워짐 | — | 없음(기록 실패 — 수락 경로와 같음) | — | exit_only_start.go:119-127 → engine.go:589-591 |

## 2. AutomationStatus 불변식

비시험 대입 자리: interlock.go:422(Verified=true) · 425(EntryPermitted) · 605(ExitOnly=true) · exit_only_start.go:108-109.

| 불변식 | 결과 | 근거 |
|---|---|---|
| ¬(Verified ∧ ExitOnly) | 성립 | ExitOnly 는 verifyGate 성공 꼬리 605 에서만. 그때 runInterlock 은 417-419 에서 422 전에 반환, exitOnlyStart 가 108 에서 Verified=false |
| ExitOnly → ¬EntryPermitted | 성립 | exitOnlyStart 의 세 출구(116 · 126 · 129) 모두 109 뒤 |
| ExitOnly → 반환된 Context 의 게이트웨이 EntryGate 에 래치 | 성립 | 성공 출구 129 는 118 `entry.Block` 뒤. `facts.entry` = `wiring.entry`(engine.go:586) = 게이트웨이 Options.Entry(gateway.go:265, 328) = Context.Entry(engine.go:610). 오류 출구는 Context 없음. 이 사유를 지우는 호출 0 |

## 3. 감시 단계 × 기동 표지

등급: healthy · notice 밖은 critical(attestation_watch.go:87-94). 신원 = 기한(337), 신원이 바뀌면 `healthy:` 키만 남기고 비움(338-349).

| 단계 | exitOnly=false | exitOnly=true | 덧붙는 문장(exitOnly=true) | 등급 | 키 |
|---|---|---|---|---|---|
| healthy | 알림 없음 | 회복 안내(exitOnlyRecovered) | 없음(본문이 이미 안내) | normal | `healthy:<KST 날짜>:<episodeAt>` — 신원 변경에도 유지, 하루 한 번 |
| notice | 「그 뒤 재시작하면 청산 전용」(+ proof 갈래) | 같음 | 「지금 청산 전용으로 돌고 있다 — 증명은 회복됐다 …」 | normal | `notice:<days>:<deadline>` |
| imminent | 같음 | 같음 | 회복 문장 | critical | `imminent:<days>:<deadline>` |
| proof_overdue | 「재시작하면 청산 전용」 + 재검증 | 같음 | 회복 문장 | critical | `proof_overdue:<deadline>:<day>:<episodeAt>` |
| expired | exitOnlyOnRestart + 만료 | 같음 | 「지금 청산 전용으로 돌고 있다 — 손절 · 익절 · 대사는 돌고 새 진입은 막혀 있다」 | critical | `expired:<day>:<episodeAt>` |
| scope_lost | exitOnlyOnRestart + POST/혼합 갈래 | 같음 | 청산 전용 문장 | critical | `scope_lost:<missing>:<day>:<episodeAt>` |
| incomplete(낡은 파일의 빈 항목 포함) | 「재시작하면 기동이 거부된다」 | 같음 | 청산 전용 문장 | critical | `incomplete:<day>:<episodeAt>` |
| account_mismatch(만료 + 다른 계좌 포함) | 기동 거부 | 같음 | 청산 전용 문장 | critical | `account_mismatch:<day>:<episodeAt>` |
| unreadable(Load 오류 · 형식 < 1 의 낡은 파일 · 미상 오류) | 기동 거부 | 같음 | 청산 전용 문장 | critical | `unreadable:<day>:<episodeAt>` |

## 발견

- [INFORMATIONAL, 6/10] exit_only_start.go:115 — nil 진입 게이트 거절 갈래가 ExitOnly=true 인 status 로 로그를 남겨 거절 줄에 보호 수명 문장이 실림(생산 도달 불가). maintainability 와 같은 지적.
