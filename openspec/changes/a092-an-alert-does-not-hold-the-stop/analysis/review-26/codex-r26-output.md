BLOCK

| # | 등급(P0/P1/P2) | 파일:줄 | 무엇이 틀렸나 | 근거(코드 인용 또는 실행 결과) | 제안 |
|---|---|---|---|---|---|
| 1 | **P0** | [alertdelivery.go:605](/tmp/claude-1000/a092-r26-tree/internal/app/engine/alertdelivery.go:605) | **배달 실행자는 반납 결과를 버린다. 25라운드 수리가 이 경로에는 없다.** | `if _, err := ...ReleaseAlertClaim(...)`이 `Outcome`을 버린다. 실패 기록이 `Applied, Attempts=1`이고 반납에서 `NotFound,nil`이면, `recordFailedAttempt`는 365행에서 반환한다. 행 유실을 발견했는데 로그·차단·후속 재시도 근거가 모두 사라진다. 실제 원장도 행 없음에 `NotFound,nil`을 반환한다(`alert_claim.go:372`). | 반납 결과를 반환·분류하고, 행 없음·미지 결과에 원칙 E의 조건부 차단 적용. 이미 확정된 시도 한도 판정은 별도로 보존. |
| 2 | P1 | [normal_relay.go:34](/tmp/claude-1000/a092-r26-tree/internal/obs/normal_relay.go:34) | **실행자 종료·패닉 뒤 유실 기록이 빠진다.** | `Offer`에는 정지 상태 검사가 없다. 종료 배수가 빈 큐를 보고 반환한 뒤에도 알림이 들어간다. 패닉이면 `Run`에 배수 `defer`도 없어 현재 알림과 큐가 남는다. 이후 최대 64건은 아무도 보내거나 버림으로 기록하지 않는다. `StopEvent`는 실행자 정지만 기록한다. | 넘김과 종료를 직렬화하는 짧은 로컬 임계구역 도입. 종료 후 넘김·잔여 큐·패닉 당시 처리 중 사건을 각각 버림으로 기록. 전송은 잠금 밖 유지. |
| 3 | P1 | [normal_relay.go:58](/tmp/claude-1000/a092-r26-tree/internal/obs/normal_relay.go:58), [notifier.go:179](/tmp/claude-1000/a092-r26-tree/internal/obs/notifier.go:179) | **전송기 부재는 무기록 유실이고, 발행 실패에는 키가 없다.** | Relay가 호출하는 `publishBestEffort`는 `Publisher == nil`이면 즉시 반환한다. 발행 오류 로그도 유형·등급·오류만 기록하며 `alert_key`가 없다. 큐에서 꺼낸 뒤 재시도도 없다. 사건 발생 로그는 버림 기록이 아니다. | Relay가 발행 결과를 받아 전송기 부재·발행 실패를 유형+키로 기록하도록 변경. |
| 4 | P1 | [alertdelivery.go:369](/tmp/claude-1000/a092-r26-tree/internal/app/engine/alertdelivery.go:369) | **실패한 전송이 승인·임차 상실로 선점됐다는 기록이 없다.** | `SettleAlreadySettled, SettleLeaseLost` 갈래가 빈 몸체다. 앞서 호출한 `release`도 결과를 버린다. 따라서 운영자가 전송 중 승인하고 전송이 실패하면, 해당 선점 기록 없이 다음 행으로 넘어간다. 델타 62행의 “선점 사실 기록 SHALL” 위반. | 두 선점 결과를 구별해 행 ID·결과를 기록하고 배치는 계속 처리. |
| 5 | P1 | [modeops.go:116](/tmp/claude-1000/a092-r26-tree/internal/app/engine/modeops.go:116), [mode_control_transport_unix.go:157](/tmp/claude-1000/a092-r26-tree/internal/app/engine/mode_control_transport_unix.go:157) | **커밋된 완화가 후속 조회 오류 때문에 일반 실패로 보고된다.** | 통지 기록 실패를 `NotifyError`로 보존한 뒤에도 현재 모드·통지 목록 재조회가 실패하면 오류를 반환한다. HTTP 경로는 그때 `result`를 버리고 500만 보낸다. “완화 성공 + 통지 기록 실패” 정보가 사라진다. | 커밋 여부·전이 ID·통지 실패를 응답에 보존하고, 재조회 실패는 별도 상태로 표시. 읽지 못한 현재 상태를 추정하지 말 것. |
| 6 | P2 | [notifier.go:552](/tmp/claude-1000/a092-r26-tree/internal/obs/notifier.go:552), [notifier.go:903](/tmp/claude-1000/a092-r26-tree/internal/obs/notifier.go:903) | **동기 시도 기록의 미지 결과와 `Flush`에는 분류 불일치가 남는다.** | `deliver`의 시도 기록은 `Applied` 이외를 모두 `lost=true`로 반환하고, `NotFound`에만 차단한다. 미지 값은 선점처럼 승격·차단을 건너뛴다. `Flush`도 반납·전달 정산의 `NotFound`를 로그만 남기고 계속한다. | 명시적 결과 분류로 통일. 단, 현재 원장은 미지 enum을 반환하지 않고 `Flush`의 생산 호출자도 없으므로 현재 생산 장애로 과장하지 않는다. |

위 표는 **정적 코드 반례**다. #1은 원장 행 유실이라는 결함 조건을 전제로 하며 정상 운용에서 발생했다고 주장하지 않는다. 시험은 실행하지 않았다. 환경의 읽기 전용 권한 때문에 요구된 시험 사본을 만들 수 없었다. 파일·git 상태·운영 원장·자격 증명은 변경하거나 열지 않았다.

`55963f29`의 **해당 자리 수리는 유효**하다. `deliver` 598–630행은 반납 오류, 적용, 행 없음·미지 결과, 명명된 선점을 구분한다. 행 없음에는 조건부 차단하고 승격하지 않는다. 그러나 같은 규칙이 배달 실행자와 시도 기록·`Flush`까지 닫히지는 않았다.

| 저자 주장 | 판정 | 코드 근거 |
|---|---|---|
| 1. k3 합성 | **참 — 현재 생산 배선 기준** | `ConfirmedFloor`의 Holdings·SellableQuantity 호출 모두 `f.retrier.Query` 안에 있으며(`exitwiring.go:207,231`), 생산 조립은 기록 전용 announcer를 가진 복사본을 넘긴다(`:333–353`). 현재 코드의 우회는 찾지 못했다. |
| 2. exit 알림 원격 전송·전송 잠금 대기 없음 | **참 — 현재 생산 경로 기준** | 알림·모드 통지는 `RecordOnly`, 일반 등급은 비차단 `Offer`; 동기 `deliver`는 `n.mu.Unlock()` 뒤 호출된다(`notifier.go:347–350`). 잠금을 전송까지 쥐는 `Flush`는 생산 호출자가 없다. |
| 3. Acknowledge 셈~해제 오개방 없음 | **참 — 델타가 정한 범위** | 기록과 승인 셈~해제가 같은 `n.mu`를 사용한다. 입구 밖 기록자는 `parkAlert` 하나이며 삽입 전에 별도 사유를 잠근다(`replay.go:537–553`). 이는 #1의 행 유실 문제와 별개다. |
| 4. 어느 순간에도 최신 커밋보다 덜 보수적이지 않음 | **거짓 — 주장 문구가 스펙보다 강함** | `Commit`과 투영 사이에는 창이 있다(`operating_mode.go:483–491`). 델타는 그 창에서 직전 모드를 보는 것을 허용한다. **반환 이후 보장**, 역순 도착 방지, 원자 교체, rowid 기반 복원은 코드에 있다. |
| 5. mode-release 승인 우회 없음·재읽기·산 게이트 경계 | **참 — 권한·전이 경로 기준** | 원장 API가 방향·승인·commit 전 audit를 판정하고, 엔진 안에서 투영한다. 성공 응답은 재조회 값이다. 다만 재조회 오류 시 커밋 사실을 잃는 #5가 남는다. |
| 6. 보조 실행자·모드 엔드포인트 실패가 엔진을 중단하지 않음 | **참 — 확인한 오류·보조 실행자 Run 패닉 경로** | Auxiliary는 감독 `stops` 채널에 쓰지 않고 패닉을 회수한다(`runtime.go:322–329`, `auxiliary.go:130–136`). 모드 엔드포인트 기동 오류는 강등 보고 후 `rt.Run`으로 진행한다(`engine.go:355–375`). |
| 7. C8 모든 버림 경로 기록 | **거짓** | 가득 참·Relay 미배선·배수 중 꺼낸 사건은 기록한다. 종료 뒤 유입·패닉 잔여·전송기 부재·발행 실패의 사건 키는 #2·#3처럼 빠진다. |

나머지 주요 SHALL 대조에서는 임차 없는 기록·재무장·기존 임차 보존, 기록 실패의 동기 차단과 무통지 승격, `rec.ID` 통지 신원, 원칙 E와 승격 실패 시 무조건 차단, 모드 원자 교체·순서 울타리, 필수 운영자 입력·mutating 표지를 코드로 확인했다. 시간 예산·경합·재시작 시험의 실제 통과 여부는 이번 리뷰에서 검증하지 않았다.

Recommendation: Fix the unchecked release outcomes and missing loss records, then rerun isolated regression tests because the current executor can ignore a vanished critical row and silently lose normal alerts.
