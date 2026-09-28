판정: **BLOCK**

21라운드의 모드 통지 동기 전송 경로는 22판 설계로 닫힙니다. 그러나 동기 발송자의 승격 실패 처리와 정본 간 계약 충돌이 남습니다.

아래 `A/`는 `openspec/changes/a092-an-alert-does-not-hold-the-stop/`입니다. 반례는 코드·설계에서 도출했으며 실행 재현은 하지 않았습니다.

| # | 등급 | 주장 | 증거(파일:줄) | 권고 |
|---|---|---|---|---|
| 1 | **P0** | **조건부 차단 뒤 승격 실패 시 fail-open 경로가 생긴다.** 소진·반납 → 세대 읽기 → 운영자 승인·Clear → 조건부 Block 생략 → 승격 쓰기 실패 순서에서 알림 차단과 모드 차단이 모두 없을 수 있다. 기존 `Notifier.escalate`는 실패를 로그로만 처리한다. | `A/design.md:648–655`; `internal/obs/notifier.go:378–398,541–573`; 비교 대상 `internal/app/engine/alertdelivery.go:427–439`; `openspec/specs/engine-safety/spec.md:1463–1464` | a124처럼 **승격 실패 시 조건부 차단 결과와 무관하게 Block**하도록 동기 경로의 설계·반환 계약·RED를 추가한다. |
| 2 | **P0(T)** | **AC2의 보장 시점이 정본과 충돌한다.** 새 델타는 commit→projection 사이에 옛 모드로 진입 점검하는 것을 명시적으로 허용하지만 risk-management는 영속과 동시에 투영한다고 요구한다. rowid 울타리는 역순 적용만 막으며 이 창은 닫지 않는다. | `A/specs/engine-safety/spec.md:295–297`; `openspec/specs/risk-management/spec.md:133`; `internal/journal/operating_mode.go:468–479` | risk-management에도 MODIFIED를 넣어 보장 시점을 일치시키거나, 기존 정본을 만족하는 집행 방식을 정한다. 현재 두 델타만 archive하면 충돌이 남는다. |
| 3 | **P1(T)** | **「모든 발송자」가 a124의 예외를 지운다.** 무조건적인 “승인 뒤 재잠금 금지”는 a124가 허용하는 **근거 확정~세대 읽기 사이 해제의 보수적 재잠금**, 그리고 #1의 **승격 실패 재차단**과 양립하지 않는다. | `A/specs/engine-safety/spec.md:64,101–104`; `A/tasks.md:29–30`; 정본 `engine-safety/spec.md:1463–1473`; `internal/app/engine/alertdelivery.go:409–424` | 일반화 문장과 Scenario에 두 예외를 명시한다. 시험도 근거 반환 전·반환~세대 읽기 사이·읽기 후를 구분한다. |
| 4 | **P1(T)** | **Q4 핀의 금지 형태와 양성 대조군이 충돌한다.** 현재 `parkAlert`는 `entry == nil`이면 Block을 건너뛰고 Enqueue에 도달한다. 설계는 바로 이 형태를 금지하면서 현재 코드는 통과해야 한다고 한다. 또한 Enqueue가 B2 안에 있다는 AST 해석도 틀렸다. | `A/design.md:658–666`; `A/tasks.md:31–32`; `internal/execgw/replay.go:534–558` — B1은 `entry != nil`, B2는 JSON 오류 처리이며 Enqueue는 둘 밖 | 핀 범위를 생산 조립의 non-nil 보장까지 포함한 도달 경로로 정의하거나, nil 경로를 명시적으로 수리한다. 존재 검사로 통과시키면 안 된다. |
| 5 | **P1** | **복원 실패 래치를 완화 명령으로 풀 수 있다는 약속이 성립하지 않는 경우가 있다.** 일시적 읽기 오류로 래치했지만 실제 원장 모드가 NORMAL이면, NORMAL 요청은 무변경 반환하여 투영을 실행하지 않는다. 설계의 “현재보다 덜 보수적” 조건으로도 요청할 목적지가 없다. | `A/design.md:544–547,713–716`; `internal/journal/operating_mode.go:409–415,475–476`; `A/tasks.md:38` | 성공한 원장 재조회에 따른 안전한 재투영 경로를 정의하거나, 이 경우 복구는 재시작뿐이라고 제한한다. NORMAL·행 없음 사례를 시험한다. |
| 6 | **P2(T)** | **현재 증거 좌표와 archive 본문에 낡은 설명이 남는다.** C1 생산 Announcer는 `engine.go:636`이 아니라 `:639`다. 새 MODIFIED가 그대로 복사한 risk-management `:102–108`도 현재 승인 규범 위치가 아니다. 「등급화된 알림」의 “운영 모드는 예약만”은 이번 생산 통지 계약과 맞지 않는다. | `A/design.md:610`; `cmd/tossctl/engine.go:634–640`; `A/specs/engine-safety/spec.md:32,189`; `openspec/specs/risk-management/spec.md:121–137` | 현재 규범은 요구사항 제목으로 연결하고, 통째로 교체되는 본문의 예약 표현을 정리한다. 역사 기록과 현재 계약을 구별한다. |

1. **제목 달성 — 설계상 PASS, 구현 증명은 미완료.** 생산 경로의 원격 알림 대기는 아래처럼 닫힌다.

   | 도달 경로 | 22판의 차단 지점·근거 |
   |---|---|
   | `ExitObserver.alert → Notify → notifyCritical → claimAndDeliver → deliver` | 기록 전용 어댑터와 `RecordAlert`. `exitloop.go:1706–1713`, `notifier.go:130–138,200,309`, `A/design.md:636–646` |
   | 같은 `Notify`의 일반 등급 → `publishBestEffort` | 유계 버퍼·별도 보조 실행자. `notifier.go:134–136,161–165`, `A/design.md:686–693` |
   | `checkOutage → EscalateOperatingMode → AnnounceOperatingMode → Notify` | observer 전용 Announcer 교체. `exitloop.go:843–847`, `operating_mode.go:478–479`, `mode.go:57`, `A/design.md:610,621–624` |
   | 가격 `observe → Retrier.Query → escalateCredentialFailure → AnnounceOperatingMode` | observer용 Retrier 값 복사본. `exitloop.go:747`, `retry.go:357–365,413–414`, `exitwiring.go:333`, `A/design.md:619–620` |
   | 상한 조회 `ConfirmedFloor → Retrier.Query`의 holdings·sellable 두 갈래 → 같은 통지 | floor용 Retrier 값 복사본. `exitwiring.go:207,231`, `gateway.go:350–353` |

   추가 경로도 확인했다. 기록 실패의 `Notifier.escalate`는 Announcer에 **nil**을 전달하므로 원격 재진입하지 않는다(`notifier.go:382–383`). 청산의 `IssueReduction`은 거절 시 반환하고 `escalateFor`를 호출하지 않는다(`riskguardian.go:564–610`). `Journal == nil`의 동기 publish 우회는 task 22.3 C16이 명시적으로 금지한다(`A/tasks.md:39`).

   Retrier는 실제로 mutex·내부 계수 없이 설정·공유 객체 참조를 가지며, 세대·신선도 상태는 Gate에 남는다(`retry.go:305–328,335–388`). **모드 통지 주입 지점 셋이라는 측정은 맞다.** 다만 floor 안의 실제 조회 경로는 둘이다. 공유 Notifier의 잠금을 원격 전송 밖으로 옮기는 조건까지 적용해야 0이 된다(`A/design.md:454–458`).

2. **RecordAlert — 설계상 PASS.** `recordAlertTx` 재사용은 중복 제거·재무장을 보존하고 임차 취득을 추가하지 않는다. PENDING은 재무장하지 않아 기존 임차를 건드리지 않고, 정착 행의 재무장은 내용·승인·시도 상태를 함께 초기화한다(`outbox.go:289–346,378–410`). 같은 Notifier의 `n.mu` 아래 호출하면 `Acknowledge`의 셈~Clear와도 배제된다(`A/design.md:646`; `notifier.go:851–875`). 새 공개 함수의 직접 호출까지 이 보장이 자동 확장되지는 않는다. 어댑터가 **기존 Notifier와 같은 mutex**를 사용하고 `n.remindAfter()`의 양수 기본 정책을 전달하는 것이 구현 조건이다(`notifier.go:349–354`).

3. **원칙 E — FAIL.** 세대 읽기 **후** 해제는 `BlockUnlessClearedSince`가 정확히 막는다(`retry.go:571–583`). 그러나 모든 승인 뒤 재잠금을 닫는 것은 아니다. 근거 반환과 세대 읽기는 별도 단계이며 a124도 그 사이의 보수적 재잠금을 허용한다. 더 중요한 새 위험은 **#1의 승격 실패 시 차단 소실**이다. 조건부 Block 자체는 Clear하지 않지만, 기존 무조건 Block을 대체하면서 실패 시 남아야 할 차단을 생략할 수 있다.

4. **커밋 순서 — rowid 선택 PASS / AC2 정본 정합성 FAIL.** 현재 스키마는 TEXT PRIMARY KEY를 가진 rowid 테이블이고, 전이는 rowid를 지정하지 않는 INSERT를 사용한다(`core_domain.go:183–193`; `operating_mode.go:444–447`). DSN도 실제 `_txlock=immediate`를 설정한다(`journal.go:219–225`). 저장소 Go 소스에서 해당 테이블의 DELETE·UPDATE는 발견하지 못했다. 이 append-only 전이 경로에서는 rowid 순서를 방향 판정·복원·투영 울타리에 공통 적용하는 설계가 성립한다(`A/design.md:674–680`). **전이 반환 시점 보장과 “영속과 동시에”의 충돌은 별개이며 #2로 남는다.**

5. **archive 정합성 — FAIL.** MODIFIED 셋·ADDED 하나만으로는 #2·#3의 규범 충돌이 해소되지 않는다. 새 「배달 실행자의 정지가 …」를 정본과 직접 비교한 결과, **규범·Scenario는 동일하고 근거 ① 교체와 변경 설명 주석 추가만 있었다**(`A/specs/engine-safety/spec.md:150–283`). 따라서 그 MODIFIED 자체의 교체 범위는 적절하다. a124의 “투영이 배선되기 전에는”은 조건문이므로 배선 뒤에도 거짓이 되지 않는다(`openspec/specs/engine-safety/spec.md:1444–1446`). 다만 #6의 낡은 문구·좌표까지 정본에 다시 싣는 문제는 남는다.

6. **Q4 / C8 — Q4 FAIL, C8 설계 PASS.** Q4의 세 금지 형태는 필요한 검사지만 현재 양성 대조군과 양립하지 않는다(#4). C8의 별도 보조 실행자는 critical 배달 사이클에 일반 알림 작업을 직접 추가하지 않으며, 기존 런타임의 배수·패닉 격리·정지 기록 계약을 사용할 수 있다(`A/design.md:686–693`; `runtime.go:312–329,347`; `auxiliary.go:88–129`; 정본 `engine-safety/spec.md:203`). “critical 사이클에 작업을 넣지 않는다”는 성립하지만 공유 자원 경합까지 없다는 보장은 아니다.

7. **a094 교차 — outbox 계약 충돌은 없음, 통합 핀 범위 조정 필요.** 제공된 a094 처분 방향인 **에피소드 키 + EnqueueAlert 계약 유지**는 a092의 **같은 키에 대한 시간 기반 재알림 + RecordAlert**와 역할이 다르다. `EnqueueAlert(..., 0)`은 정착 행을 되살리지 않고(`outbox.go:142–146,382–384`), RecordAlert는 이 계약을 바꾸지 않는다(`A/design.md:636–637`). 따라서 RecordAlert를 a094의 에피소드 신원 해결책으로 대체해서는 안 된다. 현재 `parkAlert`의 키는 attempt ID를 포함한다(`replay.go:552`). a094의 새 Enqueue 호출자들은 a092가 요구하는 **모든 직접 기록자 핀**의 대상이 되므로, 각 경로의 선행 차단 사유를 함께 검증해야 한다(`A/design.md:664–667`; `openspec/changes/a094-a-stop-clears-what-blocks-it/tasks.md:161`).

8. **문서 좌표·AST — 추출물 신선도 PASS, 인용 정확성 부분 FAIL.** `head-ast-21`의 **39개 JSON 모두 소스 SHA-256이 현재 파일과 일치**했고 위치 범위 오류도 없었다. MANIFEST의 HEAD는 `bce793a7…`이며 현재 소스 일치는 확인했지만, git archive 트리이므로 그 SHA만으로 과거 `c1e34dc4`와의 동일성까지 증명할 수는 없다(`A/analysis/head-ast-21/MANIFEST.txt:1–40`; `A/design.md:599–601`). 실질적인 AST 오독은 **parkAlert의 B2가 Enqueue를 감싼다는 주장**이다. B2는 `if err != nil`이고 Enqueue는 그 다음 무조건 실행문이다(`replay.go:548–551`). 나머지 주요 Notify·Announce·세 래치·정산 분기 좌표는 현재 코드와 일치한다.

읽기 전용으로 검토했으며 파일 생성·수정, git 상태 변경, 엔진·LIVE 명령 실행은 하지 않았다. 금지된 리뷰 원문·운영 원장·자격 증명도 열지 않았다.

Recommendation: **BLOCK — 동기 경로의 승격 실패 재차단, 원칙 E의 예외, AC2 정본 계약, Q4 핀과 복원 실패 해제 경로를 수정한 뒤 재리뷰한다, because 현재 설계에는 차단 소실 반례와 함께 만족할 수 없는 규범·검증 조건이 남아 있다.**