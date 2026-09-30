# Branch Test Map: `ExitObserver.workingSet`

AST 분기 22 · return 3. 진입 실측 `analysis/harness/observeonce.blocks`(commit `eac13df1`, 깨끗한 detached worktree).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:495` 포지션 읽기 오류 | **없음 — 진입 0** | no | no |
| B2 | `:499` exit state 읽기 오류 | **없음 — 진입 0** | no | no |
| B3 | `:503` 상태 색인 | 하네스 전반(`exitloop_test.go` `newExitHarness`) | no | yes |
| B4 | `:508` 포지션 순회 | 같음 | no | yes |
| B5 | `:509` 닫힘·0수량 건너뜀 | 하네스 전반(청산 뒤 주기) | no | yes |
| B6 | `:512` exit 대상 아님 → 미관리 경보 | `TestTheUnmanagedAlertNamesTheStockInKorean`(`a085_alert_text_test.go:22`) · `TestTheUnmanagedAlertSurvivesAdoptionBeingOff`(`reconcileloop_test.go:494`) | no | yes |
| B7 | `:525` 상태 없음 → 열기 | `TestTheLoopOpensTheExitStateOfANewlyHeldPosition`(`exitloop_test.go:447`) | no | yes |
| B8 | `:527` 열기 실패 → 로그·탈락 | **없음 — 진입 0** · **a090 RED R12** | no | no |
| B9 | `:528` | 없음 | no | no |
| B10 | `:533` 완료된 상태 → 무음 탈락 | **없음 — 진입 0** · a090 명명 잔여(design D1 표) | no | no |
| B11 | `:542` 손상 스냅숏 → 격리 | `TestACorruptSnapshotQuarantineIsAnnouncedWhenItIsCreated`(`a074_quarantine_announcement_test.go:86`) | no | yes |
| B12 | `:545` 격리 쓰기 실패 → 탈락 | **없음 — 진입 0** · a090 RED R13 | no | no |
| B13 | `:546` | 없음 | no | no |
| B14 | `:556` 격리 읽기 실패 → 탈락 | **없음 — 진입 0** · a090 RED R13 | no | no |
| B15 | `:561` else | `TestAnAlreadyActiveQuarantineIsNotAnnouncedAgain`(`a074…:203`) | no | yes |
| B16 | `:557` | 없음 | no | no |
| B17 | `:561` 활성 격리 → refused | 같음(`a074…:203`) | no | yes |
| B18 | `:565` else | `TestASupersededQuarantineIsReJudgedAndReleased`(`a084_quarantine_rejudge_test.go:54`) | no | yes |
| B19 | `:565` 재판정 한 번 | 같음 | no | yes |
| B20 | `:589` 정책 신원 불명 → 격리 | `TestALegacyIdentityQuarantineIsAnnouncedWhenItIsCreated`(`a074…:58`) | no | yes |
| B21 | `:592` 격리 쓰기 실패 → 탈락 | **없음 — 진입 0** · a090 RED R13 | no | no |
| B22 | `:593` | 없음 | no | no |

> 표의 시험 이름은 그 분기 블록이 진입했다는 실측과 시험 이름·주석의 대응으로 붙였다 — **분기별 단일 시험 귀속은 측정하지 않았다**(패키지 전체 커버리지).

## 편집 뒤 (a090)

| 이탈 | 편집 뒤 좌표 | a090 에서 | a090 시험 |
|---|---|---|---|
| 표시 앞 B5(보유 아님) · B6(미관리) | `:541` · `:544` | 세지 않음(범위 밖) | `TestA090R8…`(보유 종료 → 정리) · `TestA090R14…`(미관리) |
| 표시 | `:554` `o.markHeld(cycle, p)` | 보유·대상 표시 | `TestA090R15…`(위치 핀) |
| B8 열기 실패 | `:561` | 미관측(`not_in_working_set`) | `TestA090R12APositionWhoseStateCannotOpenIsCountedAndAlerted`(형제 있음 · 단독 B3) |
| B10 완료 | `:567` 첫 문장 `o.unmarkHeld` | 표시 해제 — 세지 않음 | `TestA090R14UnmanagedAndCompletedPositionsAreNotCounted` |
| B12 격리 쓰기 실패 | `:581` | 미관측 | `TestA090R13…/quarantine_write_B12` |
| B14 격리 읽기 실패 | `:592` | 미관측 | `TestA090R13…/active_quarantine_read_B14` |
| B11 · B17 · B20 refused | — | 판정 진입(alertRefused) → 관측됨 | `TestA090R16…/quarantined_refused` |
| B21 격리 쓰기 실패 | `:628` | 미관측(B12 와 같은 규칙 — 집합 계수라 자리별 시험 없음) | 구조 핀 `TestA090R15…` |
