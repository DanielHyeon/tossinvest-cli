# 26라운드 보이스 C (시험 적정성 · 변이) — 원문

- 검토 대상: `15cb8540` 읽기 전용 사본(`071e67c1` · `d8769cfb` 이전). 변이는 사본 `/tmp/claude-1000/a092-r26-C-2754562` 에서만(`git rev-parse` 실패 단언, 전용 GOCACHE, `GOFLAGS=-trimpath`, 한 번에 하나, 편집마다 원본과 cmp 로 복원 확인). 무변이 대조군 C00 GREEN(engine · obs · execgw · journal · cmd 의 `TestA092`). 사본 · 캐시 삭제 전 원본과 diff 동일 확인.

**판정: APPROVE** — P0 · 안전 불변식 위반 없음. 저자 주장 7 은 실행으로 거짓, P1 세 건은 착지 전 수리 권고.

## 발견

| # | 등급 | 파일:줄 | 무엇이 틀렸나 | 근거 | 제안 |
|---|---|---|---|---|---|
| 1 | P1 | `internal/obs/normal_relay.go:47-74`, `internal/obs/notifier.go:179-191` | 주장 7 거짓. 버림 경로 넷: (a) 패닉 정지 뒤 잔여 · 넘김 무기록, 버퍼가 찬 뒤 줄은 사유를 「full」로 오기 (b) 배수 뒤 넘김 무기록 (c) 발행 실패는 `engine.alert_undelivered`(K13 이 빌리지 말라 한 타입), 키 없음 (d) Publisher nil 무기록 | E1 패닉 → 세 건 `drop recorded = false`, 줄 0. E2 배수 뒤 → false, 0. E3 → `{"msg":"engine.alert_undelivered",...,"severity":"critical","event":"exit.proposal_capped","severity":"normal","error":"503"}` 키 없음. E4 → 키 기록 false. 변이 Y10 SURVIVED | 발행 실패 · nil Publisher 는 `logNormalDrop`; `Run` 에 defer 배수; 정지 뒤 `Offer` 는 닫힘 표지로 버림 기록; 각 경로 시험 |
| 2 | P1 | `internal/obs/notifier.go:753`, `internal/obs/a099_round4_test.go:~151` | a099 교차 시험의 원래 주제(죽은 발송자의 held 행이 소리 없이 넘어가지 않음)가 빔 — INFO 로 내리며 신호를 `logClaimStolen`(WARN)으로 넘겼다고 주석에 적었으나 그 줄의 **등급**을 못 박는 시험 없음 | Y05(`Warn` → `Event`) obs 전체 SURVIVED; 줄 삭제(Y09)만 `TestContentionLossAndTakeoverAreThreeEvents` 가 잡음 | 인수 줄의 `level == "WARN"` · `claimed_by` 단언 |
| 3 | P1 | `internal/app/engine/mode_control_transport_unix.go:147` | 모드 완화 엔드포인트 토큰 검사(`alertControlAuth`, C12 「다른 힘이면 다른 토큰」)를 지키는 시험 없음. 메서드 제한 · 모르는 필드 거절도 | Z01(토큰 검사 제거) · Z02(GET 허용) · Z03(`DisallowUnknownFields` 제거) 모두 engine `TestA092\|TestA109\|Control\|Endpoint` · cmd `TestA092\|Mode\|Control` 에서 SURVIVED | 토큰 없음 · 알림 제어 토큰 → 401, GET → 405, 모르는 필드 → 400 거절 시험 |
| 4 | P2 | `internal/app/engine/a092_structure_pins_test.go:370` | K3 · M2 핀이 `body[commitAt+1:projectAt]` 만 — 투영 호출 자체를 `go` 로 감싸도 못 봄 | X01 `go p.ProjectOperatingMode(record)` → `TestA092CommitThenProjectInOneCall` SURVIVED(원장 P03 CAUGHT 는 행동 시험 7개 덕) | 구간 `[commitAt+1:projectAt+1]`, 부모가 `GoStmt` 아님 단언 |
| 5 | P2 | `a092_structure_pins_test.go:75-104,116-135,177-204` | 호출 전수 · Flush 핀이 이름 기반 — 메서드 값(복합 리터럴 값 · `return` · 패키지 `var` 초기화식 · 별칭 변수 호출) 누락; 다른 이름의 지역 변수로 부르면 기록자 전수(주장 3 근거)도 누락 | X02(`a092MutFlusher{flush: n.Flush}`) → K19 SURVIVED. X03(`enq := g.journal.EnqueueAlert; enq(ctx, …)`) → K6/K7 SURVIVED | `go/types` 로 `*types.Func` 참조 전부, 또는 `SelectorExpr` 이름 출현 전부 |
| 6 | P2 | `a092_structure_pins_test.go:393-448` | K18 폐포가 호출 이름만 따라가 메서드 값 · 함수 필드 · 다른 패키지 경유에서 끊김; 양성 대조 `len(seen) < 2` 는 금지 대상 도달 가능성을 못 잼 | X04(`esc := g.escalateFor; _ = esc(ctx, …)` 를 IssueReduction 에) SURVIVED | 알려진 도달 경로 하나를 양성 대조로 |
| 7 | P2 | `a092_structure_pins_test.go:452-510`, `internal/app/engine/exitwiring.go:352` | 주장 1 합성의 빈틈 둘 — (a) k3 핀은 `f.official.X` 모양만(별칭 · 헬퍼 누락) (b) `if opts.Floor == nil` 존재 검사(N11 이 Alerts · Announcer 에서 없앤 것); 호출자가 Floor 를 넘기면 공유 Retrier 가 exit goroutine 에. cmd 핀은 `engineRuntime` 리터럴 키만 | X05(Query 밖에서 별칭으로 `SellableQuantity`) SURVIVED | Floor 무조건 덮기 · 거절, 핀은 `official` 참조 전체 |
| 8 | P2 | `internal/obs/normal_relay.go:17` · `:50`, `internal/app/engine/modeops.go:89`, `mode_control_transport_unix.go:163-164` | 못 박히지 않은 결정 넷: 버퍼 64, 「종료가 먼저」 선검사, 타입 있는 nil audit 거절(생산에서는 `c.Audit != nil` 이 먼저 걸러 죽은 갈래로 보임 — 판독), 503 매핑 | Y02 · Y08 · Y06 · Y07 SURVIVED | 의도면 시험으로 고정, 죽은 갈래면 삭제 |
| 9 | P2 | `internal/obs/a092_lock_scope_test.go` `TestA092AReleaseBeforeTheEpochReadRelatches` | 훅이 안 불리면 해제도 없어 「잠김」이 공허하게 참; 단계 도달은 형제 시험(`EachLatchSite`)만 단언 | 코드 판독 | 시험마다 `seen` 에 단계 도달 단언 |

## 저자 주장 판정

1. **참(현 HEAD 한정), 핀은 못 지킴** — 생산 floor 는 exit Retrier 복사본, 브로커 호출 전부 `retrier.Query` 안. 합성의 두 다리가 존재 검사와 모양 핀(X05 SURVIVED)이라 회귀 방지 증거 아님(#7).
2. **참(코드 추적)** — exit 경로 알림 부품 전부 기록 전용 · 복사본, Gateway 알림기 없음, n.mu 를 전송 위에서 쥐는 것은 Flush(생산 호출자 0). 추측: `SetMaxOpenConns(1)` 에서 다른 goroutine 이 tx 를 연 채 네트워크를 기다리는 경로는 전수로 안 봄.
3. **참(설계 범위 안)** — U11 · L14 · W04 CAUGHT, 삽입 입구 셋(`recordAlertTx` 호출자). `parkAlert` 가 셈 뒤 넣은 행이 있어도 전달 실패 사유는 풀림 — 자기 사유(`UNRESOLVED_IN_DOUBT`)가 막음. 전수 핀은 메서드 값 못 봄(X03).
4. **참** — M01~M11 CAUGHT, Y01(해제 때 revision 안 올림) CAUGHT. Seq 0 공허 교차 시험 없음.
5. **참** — 판정은 `TransitionOperatingMode`, 결과 재읽기(M14), 원장만 고친 완화는 산 게이트 못 엶. 엔드포인트 토큰 시험 없음(#3) — 원장 판정 우회는 아님.
6. **참(판독)** — 강등 보고, 이관은 recover 경계 보조 실행자. 패닉 정지 시 기록 없는 유실(#1).
7. **거짓** — #1(E1~E4 실행, Y10 SURVIVED).

## 교차 change 시험 편집

- a096: 「셈~해제 사이 기록 없음」 행동 → 구조 핀. W04 는 타이밍 시험 둘이 잡음.
- a099: 주제 사라짐(#2).
- a124 ×3: 문서 참/거짓 핀으로 갱신, epoch 시험에 Seq 를 붙인 것은 올바른 수리. (b) 전수 walker 도 메서드 값 못 봄(#5 와 같은 계측기).
- a098 ×2 · a109 · help: 주제 보존. 새 표면 인증 없음(#3).

## 변이 전부

| ID | 파일:줄 | 변이 | 결과 | 실패 시험 |
|---|---|---|---|---|
| C00 | — | 무변이 대조군 | GREEN | — |
| X01 | journal/operating_mode.go:491 | `go p.ProjectOperatingMode(record)`, K3 핀만 | SURVIVED | — |
| X02 | obs/normal_relay.go(추가) | 복합 리터럴 안 `n.Flush` 메서드 값 | SURVIVED | — |
| X03 | execgw/replay.go(추가) | `enq := g.journal.EnqueueAlert; enq(...)` | SURVIVED | — |
| X04 | execgw/riskguardian.go IssueReduction | `esc := g.escalateFor; esc(...)` | SURVIVED | — |
| X05 | app/engine/exitwiring.go:231 | Query 밖 별칭 `SellableQuantity` | SURVIVED | — |
| Z01 | mode_control_transport_unix.go:147 | 토큰 검사 제거 | SURVIVED | — |
| Z02 | 같은 파일 :148 | POST 제한 제거 | SURVIVED | — |
| Z03 | 같은 파일 :189 | `DisallowUnknownFields` 제거 | SURVIVED | — |
| Y01 | execgw/modegate.go:57 | 해제 때 `revision++` 제거 | CAUGHT | TestA092TheModeRevisionMovesOnlyWhenPresenceChanges |
| Y02 | obs/normal_relay.go:17 | 용량 64 → 1 | SURVIVED | — |
| Y03 | obs/record_only.go:99 | `logged.Fields = nil` 제거 | CAUGHT | TestA092RecordCriticalKeepsFieldsOutOfTheLog |
| Y04 | obs/record_only.go:66 | AnnounceOperatingMode nil 가드 제거 | CAUGHT | TestA092RecordOnlyNilNotifierIsANoOp |
| Y05 | obs/notifier.go:753 | 인수 줄 WARN → INFO | SURVIVED | — |
| Y06 | app/engine/modeops.go:89 | 타입 있는 nil audit 허용 | SURVIVED | — |
| Y07 | mode_control_transport_unix.go:164 | 503 → 500 | SURVIVED | — |
| Y08 | obs/normal_relay.go:50 | 종료 선검사 제거 | SURVIVED | — |
| Y09 | obs/notifier.go(claimAndDeliver) | `logClaimStolen` 호출 삭제 | CAUGHT | TestContentionLossAndTakeoverAreThreeEvents(/takeover) |
| Y10 | obs/normal_relay.go:58 | 발행 실패 무기록 버림 | SURVIVED | — |
| W04 | obs/notifier.go:971 | Acknowledge 의 Clear 를 goroutine 으로 | CAUGHT | TestRecoveredDeliveryDoesNotReleaseTheGateByItself, TestAcknowledgeWhileStillPendingKeepsTheBlock |

E1~E4 는 변이가 아니라 사본의 임시 관측 시험. journal 전체 판은 안 돌림(`TestA092` 필터만).

Recommendation: APPROVE, 착지 전 #1~#3(P1) 수리 — 불변식 위반 경로는 없으나 델타의 버림 기록 SHALL 이 실행으로 반증됐고, 모드 완화 표면의 토큰 통제와 a099 교차 시험이 옮긴 WARN 신호를 지키는 시험이 없음(Z01 · Y05 SURVIVED).
