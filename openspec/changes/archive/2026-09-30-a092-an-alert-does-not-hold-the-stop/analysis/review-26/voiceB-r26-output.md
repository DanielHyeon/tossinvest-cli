# 26라운드 보이스 B (스펙 대조) — 원문

- 검토 대상: `15cb8540` 읽기 전용 사본(`071e67c1` · `d8769cfb` 이전). 사본 `/tmp/claude-1000/a092-r26-B-2901749`(`git rev-parse` 실패 단언 뒤) — `-run A092` obs · execgw · journal · app/engine · cmd/tossctl 전부 ok, 프로브 셋 재현, 사본 · GOCACHE 삭제.

**판정: BLOCK** — 사유는 안전 불변식 8 위반 하나. a092 가 새로 만든 표면(기록 전용 입구, `mode-release`)에서 브로커 계좌번호 원문이 네 곳으로 샘: 구조화 로그, 게이트 설명, CLI 출력, 오류 본문. 수리 범위는 좁고 손절 경로는 안 건드림.

`AccountRef` 가 원문 계좌번호라는 근거: `interlock.go:684` `accountRef := strings.TrimSpace(first.DisplayName)`, `official/reads.go:93` `DisplayName: a.AccountNo`.

## 발견

| # | 등급 | 파일:줄 | 무엇이 틀렸나 | 근거 | 제안 |
|---|---|---|---|---|---|
| 1 | P1 (불변식 8) | `internal/obs/record_only.go:50`, `:70` | f48e7865 는 필드를 로그에서 빼는 수리를 `RecordCritical`(`:98-100`)에만 함. 같은 입구의 `Notify` · `AnnounceOperatingMode` 는 `n.logEvent(e, …)` 로 `Fields` 전부를 씀 — 모드 통지의 `FieldAccount: rec.AccountRef`(`:173`), exit 두절 알림의 `obs.FieldAccount: o.opts.AccountRef`(`exitloop.go:836`). `ModeOperations` 는 통지자로 `RecordOnly`(`modeops.go:65`) → `mode-release` 호출마다 원문 계좌번호가 로그에 남음. | 사본 프로브: `{"msg":"engine.operating_mode",…,"account":"99887766554","from_state":"ENTRY_BLOCKED","to_state":"NORMAL",…}`. f48e7865 제목 "기록 전용 입구의 로그 줄에서 필드를 뺌" 인데 셋 중 하나만. | `Notify` · `AnnounceOperatingMode` 에도 `logged.Fields = nil`(또는 계좌 키 제거) + 같은 모양 시험. |
| 2 | P1 (불변식 8) | `record_only.go:134-138`, `modeops.go:106`, `mode_control_transport_unix.go:166`, `journal/record_alert.go:26,33`, `outbox.go:343,349` | 원장 오류 문구에 event key 가 들어감 — 모드 통지 키 `operating_mode:<계좌>:…`(`record_only.go:154`), 두절 알림 키 `exit.observation_outage\|<계좌>`(`exitloop.go:832`). 기록 실패 시 ① 게이트 설명(저자 스스로 `mode_projection_wiring.go:18` · `blockEntryOnDeliveryStop` 에 "원장 오류 원문을 싣지 않음" 규칙) ② `ModeReleaseResult.NotifyError`(CLI 출력 `engine_mode_release.go:108`) ③ Error 로그. 별도로 500 경로는 `err.Error()` 그대로 — `"journal: committing the %s → %s transition for %s"`(`operating_mode.go:400,455,485`)처럼 계좌를 담은 원장 오류가 CLI stderr 에. | 원장을 닫은 뒤 프로브: `GATE critical_alert_undelivered: … journal: recording alert operating_mode:99887766554:NORMAL:t2: sql: database is closed`, 반환 오류에도 같은 문자열. | 게이트 설명 고정 문구(`operatingModeRestoreFailed` 방식), `NotifyError` · 500 본문은 고정 분류만, 원문 오류는 계좌를 뺀 채 로그에만. |
| 3 | P2 | `internal/obs/normal_relay.go:47-62` | 주장 7 「실행자 정지 뒤 남은 것」이 패닉 정지에서 불성립 — `drainOnShutdown` 은 ctx 종료에서만; 패닉이면 버퍼 잔여 · 발행 중 알림 무기록, 뒤의 `Offer` 는 죽은 버퍼를 채우며 조용히 성공. 정상 종료에서도 배수 뒤 `Offer` 무기록. | 프로브(용량 4, 패닉 publisher): 6건 → 발행 0, 버림 1줄(무기록 유실 5). 이 경로를 재는 시험 없음. | recover → 배수 · 기록 → 재panic, 또는 죽음 표지로 `Offer` 가 버림 기록. |
| 4 | P2 | `notifier.go:186-189`, `normal_relay.go:81-84` | 발행 실패가 키 없는 `EventAlertUndelivered` 로 기록 — 주장 7 「모든 버림 경로가 유형 + 키」와 다름. 버림 줄의 `FieldEvent` 가 로거 event 키와 겹침. | 프로브: `"event":"engine.normal_alert_dropped",…,"event":"exit.proposal_capped"`. 키 중복은 기존 `publishBestEffort` 관례. | 발행 실패에도 `logNormalDrop`(키 포함), 필드 이름 `trigger_event`. |
| 5 | P2 | `modeops.go:116-126`, `mode_control_transport_unix.go:157-167` | 커밋 뒤 재읽기 · `PendingAlerts` 실패 시 `result, err` 반환 → 전송 계층이 result 를 버리고 500 → 커밋된 완화가 실패로 보임(델타 근거 문장과 어긋남). 오류 문구에 계좌(#2). | 코드 확인. | 200 + result, `ReadbackError` 칸. |
| 6 | P2 | `modeops.go:122-134`, `engine_mode_release.go:110-113` | `NoticePending=false` 를 "기록됨(이미 전달 처리됨)" 으로 출력 — 실제로는 PENDING 목록에 없다는 뜻일 뿐(승인됐거나 행이 없음). `Notifier.Journal` nil 이면 `recordCritical` 이 nil → `Notified=true`. | 코드 확인. 생산 조립은 Journal 이 채워져 추정 경로. | 통지 행을 id · 키로 다시 읽어 상태를 그대로. |
| 7 | P2 | `exitwiring.go:352-354`, `a092_structure_pins_test.go:452` | Floor 는 `if opts.Floor == nil` 일 때만 — B#4 무조건 덮기 미적용. k3 합성은 `engineRuntime` 리터럴 키만 보는 핀에 기댐. floor AST 핀도 `ConfirmedFloor` 본문만 셈(헬퍼 · 별칭 누락). | 코드 확인. | 무조건 `opts.Floor = exitSideFloor(…)`, 핀은 호출 폐포 단위. |
| 8 | P2 | `a092_structure_pins_test.go:393`, `exitwiring.go:337` | K18 「입구 도달 경로 전수」 핀은 `IssueReduction` 폐포만 — `Submit = c.Gateway` 간선은 핀 없음. 현재 `execgw.Gateway` 엔 announcer · notifier 없음(승격은 `retry.go` · `riskguardian.go` 뿐) → 주장 2 는 성립, tasks 25.10 의 「전수」는 과장. | grep, 코드 확인. | `Gateway` 제출 폐포에도 BFS 핀, 또는 문구를 「Issuer 한정」으로. |
| 9 | P2 | `runtime.go:462-467`, `riskguardian.go:644-648` | B#7 수리(`ErrModeAnnouncementFailed` = 커밋됨)가 호출자 둘에만 — 남은 둘은 통지만 실패해도 "did not reach the operating mode, so a restart would lift the block" 오보(exit 범위 밖, 동기 통지자 경로, claim 실패 때 도달). | 코드 확인. | 같은 `errors.Is` 분기. |
| 10 | P2 | `alertdelivery.go:444-447` | 주석 "생산에서 진입을 막는 것은 위의 게이트 래치임(design D10 — 투영기 미배선)" 이 단위 ④ 이후 거짓 — 전달 실패 승격은 이제 투영으로 진입을 막고 `mode-release` 필요(K16). | 코드 확인. | 주석 수정. |

## 저자 주장 판정

1. **참(조건부)** — 생산 조립에서 성립(`engine.go:652` Floor 미전달, `exitSideFloor` 가 retrier 교체, 두 브로커 조회 모두 `f.retrier.Query` 안, 401 → `escalateCredentialFailure` → `RecordOnly`). 행동 시험 없이 합성으로만 섬, 빈틈 #7.
2. **참** — 코드 열거(Alerts · Announcer RecordOnly, Retrier · Floor 사본, Issuer 승격 없음, Gateway 통지 부품 없음, `parkAlert` 는 로컬 `EnqueueAlert`, n.mu 보유자 로컬 작업만, `Flush` 비시험 호출자 0 — K19 핀 · grep). 빈틈 #8. 실측 아님.
3. **참(코드 확인 범위, 추정 포함)** — outbox 쓰기는 `ClaimAlertForDelivery` · `RecordAlert` 가 n.mu 아래, 잠금 밖은 `EnqueueAlert` 하나(`parkAlert`, 삽입 전 자기 사유 먼저).
4. **참(코드 확인)** — rowid 순서, `modeSeq` 울타리(`<=` 무시), 한 잠금 교체, 복원 Seq. 새 시험은 안 돌림.
5. **참** — 비시험 OPERATOR 전이 호출은 `modeops.go:93` 하나, 승인 · auditor 필수와 audit→commit 은 원장 강제(`operating_mode.go:423-432`, `:476`), 타입 있는 nil 거절, 투영기는 엔진 프로세스에만, 게이트 직접 Clear 경로 없음. 단 #5 · #2.
6. **참** — 강등 보고(`engine.go:355-364`), 핸들러 패닉 net/http, 이관 패닉 `runAuxiliaryBody` recover, 이관에 OnStop 래치 없음, `Offer` 비차단.
7. **거짓** — 패닉 정지(#3, 실측) · 발행 실패(#4).

tasks §25 대조: 25.8 미체크 · 25.9 남김 항목은 정직. 25.10 「K18 전수 핀」은 과장(#8). f48e7865 는 셋 중 하나에만(#1).

Recommendation: BLOCK — 안전 불변식 8 위반을 a092 의 새 생산 경로(`mode-release`)가 매 호출 만들고 실측 재현. #1 · #2(필드 제외, 고정 문구)를 고치면 나머지 P2 는 착지 뒤 후속 가능.
