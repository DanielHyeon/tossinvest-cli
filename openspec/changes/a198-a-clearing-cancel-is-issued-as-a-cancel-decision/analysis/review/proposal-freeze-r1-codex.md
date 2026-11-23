# proposal-freeze r1 — codex (다른 계열), 2026-11-23

- 명령: `codex exec -s read-only -C ~/tossos-work/a198-impl -o <out> "<prompt>" </dev/null` (rc 0). HEAD e1eaf720c + 이 브랜치 미커밋 analysis. 프롬프트 원문은 아래 「프롬프트」.

## 프롬프트

```text
You are an adversarial code reviewer. READ-ONLY: do NOT modify, create or delete any file, do not run git commands that change state, do not run tossctl, make, systemctl, or touch ~/.config/tossctl. Report in Korean, at most 50 lines.

Repo: current directory (Go). Change under review: commit range 064f59fd9..HEAD, production file internal/app/engine/attestation_watch.go (functions judgeAttestation, restartDeadline (new), attestationWatch.prepare, stageKey, attestationWatch.eventFor; new stage proof_overdue). Tests: internal/app/engine/a196_attestation_deadline_internal_test.go, internal/soak/a196_supervised_deadline_test.go, edits in internal/app/engine/a135_attestation_watch_internal_test.go. Contract: openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/{design.md,specs/engine-safety/spec.md}.
Use `git diff 064f59fd9 -- internal` to see the diff.

Context: the watcher re-reads the capability attestation hourly and alerts before an engine restart would be refused. Deadline = min(ExpiresAt, ProofAt + (ExpiresAt - IssuedAt)), ProofAt = binding(SupervisedBy). Issuer: internal/soak/attest.go BuildAttestation/acceptSupervised (age >= validity excluded) and cmd/tossctl/soak.go supervisedProofs. Interlock: internal/app/engine/interlock.go uses attest.Attestation.Verify. Outbox: internal/obs/notifier.go (critical alerts deduped by Event.Key in journal outbox, DefaultRemindAfter 1h re-arms delivered rows).

Find defects, assuming the implementation is wrong:
1. Paths where an alert is swallowed (sent map / identity reset / outbox Key) across stage transitions healthy->notice->imminent->proof_overdue->scope_lost->healthy and among expired/account_mismatch/incomplete/unreadable; KST date rollover; same-day A->B->A; checkOnce marks sent.
2. restartDeadline edge cases (ProofAt zero, ExpiresAt<=IssuedAt, ProofAt>IssuedAt, equal deadlines, remaining exactly 0/negative).
3. Whether the computed deadline equals when the issuer actually stops carrying the proof; whether the soak test truly pins that.
4. Any divergence between watcher verdict and interlock Verify.
5. Alert text: false or misleading sentences (e.g. scope_lost body claims POST supervised evidence was missing even when a GET endpoint is missing; imminent text), wrong time zone, account/path leakage.
6. Notifications-disabled downgrade for new/changed critical stages.
7. Test gaps: mutations that would survive.
Give each finding: severity (P0 alert permanently swallowed/wrong verdict, P1 must fix, P2 advisory), file:line, evidence.
```

## 원문(마지막 메시지)

**판정: 현 문서 그대로 freeze 불가. P1 4건, P2 2건. P0는 입증하지 못했습니다.** 취소 결정의 종류·방향 수정은 타당하지만, 처음 열리는 실제 취소 경로의 복구·관측·호출 계약이 덜 닫혔습니다.

Git 명령 없이 HEAD 참조 파일에서 `e1eaf720c5fa1622321b98e73a792d2e5762b3d2`를 확인하고 현재 체크아웃 소스를 직접 읽었습니다. 기존 AST 36개의 소스 해시가 현재 파일과 일치했습니다. Git blob과 작업 파일의 동일성은 확인하지 않았습니다. 제공된 probe는 기존 실행 증거로 읽었으며 재실행하지 않았습니다.

1. **[P1] D4대로 구현하면 운영에서는 취소 실패 원인이 계속 버려집니다.**

   [design.md:10](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:10)은 `warnEpisode`로 실패를 기록하도록 지정합니다. 그러나 [exit_held_proposal.go:336](/home/daniel/tossos-work/a198-impl/internal/app/engine/exit_held_proposal.go:336)의 `warnEpisode`는 `o.log`를 호출하고, [exitloop.go:1962](/home/daniel/tossos-work/a198-impl/internal/app/engine/exitloop.go:1962)는 `opts.Log == nil`이면 버립니다.

   실제 [engineRuntime:671](/home/daniel/tossos-work/a198-impl/cmd/tossctl/engine.go:671)은 계좌 정보 노출 때문에 관측자 전체 `Log`를 공급하지 않습니다. [Context.ExitObserver:336](/home/daniel/tossos-work/a198-impl/internal/app/engine/exitwiring.go:336)도 이를 연결하지 않습니다.

   **확인된 사실:** 운영 배선을 유지하며 D4만 구현하면 새 원인 로그는 0건입니다. 기존 지연·연속 실패 경보는 남지만 원인 문구는 고정 문자열입니다. [delta:29](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/specs/exit-policy/spec.md:29)의 “실패와 원인을 기록” 요구를 충족하지 못합니다.

   **필요 처분:** 운영에 연결된 계좌 비노출 기록 경로를 지정하고, 운영 조립으로 발행 실패·제출 실패 기록을 검증해야 합니다. 시험에서만 `Log`를 넣으면 이 결함을 가립니다. 원시 `err.Error()`에는 계좌가 포함될 수도 있으므로 전체 `Log` 연결만으로 해결해서도 안 됩니다([durability.go:243](/home/daniel/tossos-work/a198-impl/internal/journal/durability.go:243)).

2. **[P1] 취소 ACKED 뒤 crash/최종 기록 실패는 재시작으로도 풀리지 않는 손절 차단을 만듭니다.**

   [dispatch.go:181](/home/daniel/tossos-work/a198-impl/internal/journal/dispatch.go:181)은 ACKED를 먼저 기록하고, :199에서 별도로 CONFIRMED를 기록합니다. 그 사이 crash 또는 최종 쓰기 실패면 CANCEL이 ACKED에 남습니다.

   이후 경로는 다음과 같습니다.

   - [RecoverPending:81](/home/daniel/tossos-work/a198-impl/internal/journal/recovery.go:81): ACKED 보존.
   - [confirmAcked:51](/home/daniel/tossos-work/a198-impl/internal/reconcile/acked_boot.go:51): CANCEL/AMEND는 해소하지 않고 경보만 기록.
   - [세션 중 해소:175](/home/daniel/tossos-work/a198-impl/internal/app/engine/in_doubt_settle.go:175): IN_DOUBT PLACE만 대상.
   - [OperatorResolve:143](/home/daniel/tossos-work/a198-impl/internal/journal/resolution.go:143): `UNRESOLVED_IN_DOUBT`에서만 전이 허용.
   - [clearTheSymbol:1659](/home/daniel/tossos-work/a198-impl/internal/app/engine/exitloop.go:1659): 같은 종목 ACKED 때문에 계속 치움 미완료.

   **확인된 사실:** 이 상태 전이에서는 재시작·일반 resolver·해동 API 모두 ACKED 취소를 닫지 못합니다. 대상 주문의 종결 체결 기록만 생겨도 해당 CANCEL attempt는 그대로 남습니다. ACKED만 진행 중이면 연속 실패 계수에서도 제외됩니다([exit_held_proposal.go:70](/home/daniel/tossos-work/a198-impl/internal/app/engine/exit_held_proposal.go:70)); 지연 경보는 남습니다.

   **조건부 영향:** 해당 쓰기 창에서 장애가 발생하면 그 종목의 재취소·손절·flatten이 계속 막힙니다. 장애 발생 빈도나 운영 발생은 미입증입니다.

   **필요 처분:** [design §2:20](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:20)의 IN_DOUBT 공시와 별도로 ACKED 복구 계약을 정해야 합니다. §4의 5xx 시험에는 이 crash window가 없습니다.

3. **[P1] 전일 주문을 오늘 취소하면 “확정 취소 재전송 금지”가 깨집니다.**

   [clearTheSymbol:1670](/home/daniel/tossos-work/a198-impl/internal/app/engine/exitloop.go:1670)은 **원 주문의 거래일**로 `ConfirmedCancelOf`를 조회합니다. 그 함수는 CANCEL intent의 거래일과 정확히 같아야 찾습니다([exit_proposal_release.go:155](/home/daniel/tossos-work/a198-impl/internal/journal/exit_proposal_release.go:155)).

   반면 [Gateway.prepareRequest:1020](/home/daniel/tossos-work/a198-impl/internal/execgw/gateway.go:1020)은 **취소 실행 시점의 거래일**을 계산해 :1048에 기록합니다.

   **확인된 사실:** 주문 거래일 D, 취소 거래일 D+1이면 성공한 취소를 다음 조회가 찾지 못합니다. 종결 기록이 아직 없고 `withPending=true`라면 다음 주기에 다시 취소합니다. `withPending=false`면 :1680의 분기로 그 매도를 건너뜁니다.

   이는 [정본 exit-policy:579](/home/daniel/tossos-work/a198-impl/openspec/specs/exit-policy/spec.md:579)의 확정 취소 재전송 금지와 맞지 않습니다. `withPending=false` 경로에서는 종결 대기 방벽도 빠집니다. 다만 **두 번째 매도의 실제 접수까지 입증한 것은 아닙니다**. 발의·게이트웨이의 추가 조건이 있습니다.

   **필요 처분:** 취소 실행일과 대상 주문 신원을 구분하는 결속 계약, 거래일 경계 및 번호 재사용 시험이 필요합니다. 거래일 조건을 단순 삭제하면 과거 취소가 새 주문에 붙는 반대 결함이 생깁니다.

4. **[P1] 실제 경로의 사전 GET이 정본 “조회 0건”과 충돌하며, 취소 1회만 센 예산으로는 부족합니다.**

   [Service.Cancel:218](/home/daniel/tossos-work/a198-impl/internal/trading/service.go:218)은 매번 `GetOrderAvailableActions`를 호출합니다. 운영 구현은 [broker.go:57](/home/daniel/tossos-work/a198-impl/internal/app/engine/broker.go:57) → [precheck.go:98](/home/daniel/tossos-work/a198-impl/internal/app/engine/precheck.go:98)의 `OrderRawByID`입니다. 취소 앞에 GET 1회, 조회 시한 최대 2초가 붙습니다.

   **확인된 사실:** 수리가 열어 주는 정상 경로는 최소 `GET → CANCEL POST`입니다. [정본:573](/home/daniel/tossos-work/a198-impl/openspec/specs/exit-policy/spec.md:573)과 [시나리오:624](/home/daniel/tossos-work/a198-impl/openspec/specs/exit-policy/spec.md:624)의 “새 브로커 조회 0건”과 문언상 충돌합니다. [proposal:47](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/proposal.md:47)의 취소 1회 설명만으로 실제 비용을 나타낼 수 없습니다.

   **조건부 영향:** 여러 주문을 순차 취소하면 사전 조회 지연이 누적됩니다. 매수는 확정 취소 중복 방지가 없어, 손절 제출이 실패하고 원 주문의 종결 기록이 늦으면 다음 주기에 GET+POST를 반복합니다. 429·전송 불명은 다시 IN_DOUBT 차단으로 연결될 수 있습니다. **실제 rate limit 초과는 미측정**이며 기존 공통 admission이 있다는 사실과 구분해야 합니다.

   **필요 처분:** 정본 해석·예외를 먼저 확정하고 GET, POST, 다음 주기 재취소, 여러 주문의 지연을 계상해야 합니다. 가짜 브로커만 쓰는 §4는 운영 `officialBroker`의 GET을 측정하지 못합니다.

5. **[P2] D3의 새 조립 오류는 취소 기능 하나가 아니라 전체 엔진 런타임을 막습니다.**

   [design D3:9](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:9)은 새 인터페이스 단언 실패를 기동 오류로 만듭니다. [engineRuntime:684](/home/daniel/tossos-work/a198-impl/cmd/tossctl/engine.go:684)는 이 오류를 즉시 반환하므로 이후 Recovery 조립과 Runtime 생성에도 도달하지 않습니다.

   **확인된 사실:** 조립 누락 시 충돌 주문 없는 포지션의 정상 손절까지 시작하지 못합니다. [design §3:26](/home/daniel/tossos-work/a198-impl/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/design.md:26)의 “새 거절 없음”에는 이 새 기동 실패 조건이 빠져 있습니다.

   **한계:** 운영 `RiskGuardian`에 메서드를 정확히 추가하면 이 실패가 발생한다고 단정할 근거는 없습니다. fail-fast 자체를 결함으로 본 것도 아닙니다.

   **필요 처분:** 전체 보호 중단이라는 영향과 선택 이유를 명시하고, 실제 운영 Guardian의 인터페이스 충족 및 runtime 조립 성공을 검증해야 합니다. “미구현 Guardian이면 오류” 시험만으로 정상 기동 보존은 증명되지 않습니다.

6. **[P2] §4의 “결정 종류 CANCEL을 원장으로 단언”은 현재 저장 모델로 직접 검증할 수 없습니다.**

   [journal.Decision:219](/home/daniel/tossos-work/a198-impl/internal/journal/decision.go:219)에 종류 필드는 없습니다. [DecisionRequest.build:372](/home/daniel/tossos-work/a198-impl/internal/journal/decision.go:372)는 PLACE만 키를 만들며 CANCEL과 AMEND는 모두 키 없는 결정으로 저장합니다. [checkIdempotencyKey:397](/home/daniel/tossos-work/a198-impl/internal/execgw/guardian.go:397)도 비PLACE에서는 키 부재만 확인합니다.

   **확인된 사실:** 발행기의 `KindCancel`을 `KindAmend`로 바꿔도 같은 reduction preimage라면 저장된 결정 모양은 같습니다. 취소 attempt의 `kind=CANCEL`은 [Gateway.Cancel:410](/home/daniel/tossos-work/a198-impl/internal/execgw/gateway.go:410)이 정하므로 그것을 단언해도 발행 종류 변이는 잡히지 않습니다.

   또한 reduction preimage 검사는 계좌·시장·종목·방향·수량 상한을 확인하지만 **대상 주문번호는 결속하지 않습니다**([guardian.go:358](/home/daniel/tossos-work/a198-impl/internal/execgw/guardian.go:358)). 잘못된 주문번호 전달을 결정 검사가 막아 준다고 가정할 수 없습니다.

   **필요 처분:** 발행 요청의 종류를 검사하는 증거와 실제 전달된 대상 주문번호의 증거를 구분해야 합니다. 이것만으로 잘못된 주문 취소가 현재 발생한다고 단정하지는 않습니다.

§4에 추가해야 할 핵심 반증은 다음과 같습니다. **미실행 시험 제안이며, 발견 건수에 중복 합산하지 않았습니다.**

| 빠진 변이·상황 | 필요한 단언 |
|---|---|
| 같은 종목의 다른 주문번호로 취소 대상 교체; 외부 주문·다른 계좌/시장·모호한 소유·lineage successor | 취소한 정확한 주문번호 집합, 비대상 취소 0 |
| 원 수량 10·부분 체결 3; 결정만 잔량 7로 변경 | 현재 요청 수량 10과의 불일치를 탐지. `LiveOrder.Quantity`는 잔량이 아닌 intent 수량([fills.go:1864](/home/daniel/tossos-work/a198-impl/internal/journal/fills.go:1864)) |
| 매도 일부 체결 후 취소 종결; 종결 직전 추가 체결 | 종결 전 추가 SELL 0, 실제 체결 반영 후 잔여량만 제출. 단순 terminal 표지만 주입하지 않음 |
| ACKED 기록 후 crash/CONFIRMED 쓰기 실패 | 재시작 후 차단 해소 계약, 무근거 재취소·재매도 0 |
| 전일 주문의 당일 취소·주문번호 재사용 | 확정 취소 재전송 0, 과거 취소 오결속 0 |
| 409 already-processing·다른 409·429·timeout | 엔진 정책의 동일 주기 재전송 0, 상태 분류·진입 latch·손절 보류 정확성 |
| 운영 로그 싱크 누락·계좌를 포함한 오류·`err=nil` 비확정 결과 | 원인 기록 존재, 계좌 노출 0, panic 0 |
| IN_DOUBT를 진행 중 취소로 잘못 분류 | 연속 실패 경보와 시간 경보 각각 발생. §4-6의 시간 경보만으로는 계수 변이를 못 잡음 |
| 가격 없는 시장가·복수 주문 중 하나 실패 | 시장가 취소 가능, 하나라도 미확정이면 SELL 0 |

반대로 **부분 체결만으로 현재 취소 preimage가 불일치한다는 주장은 성립하지 않습니다.** 현재는 결정 최대 수량과 요청 수량 모두 원 주문 수량을 사용합니다. 제공된 probe도 `요청 10 / 상한 7` 거절, `요청 7 / 상한 10` 허용을 보여 줍니다. 또한 CANCEL의 키 제거는 replay 계약 위반이 아닙니다. [replay.go:399](/home/daniel/tossos-work/a198-impl/internal/execgw/replay.go:399)는 CANCEL 재생을 명시적으로 거절합니다. 문제는 그 이후 복구 상태의 완결성입니다.

파일 수정·Git 명령·빌드·시험·캐시 생성은 하지 않았습니다. ai-memory CLI 조회는 네트워크 권한 오류로 실패했습니다. CodeGraph/AST 재생성도 실행하지 않았으며, 기존 AST 해시 대조와 직접 소스 추적으로 검토했습니다. `impeccable: N/A — backend/documentation-only`, UI 변경 없음. 위 판정은 정적 설계 리뷰이며 운영 재현·전체 감사·배포 승인을 뜻하지 않습니다.


## 정정 (2026-11-23, r2 P2-5 지적)

위 「프롬프트」 블록은 **a198 이 아니다** — 같은 세션 스크래치 경로(`codex-prompt.txt`)를 병행 작업(a196 구현 리뷰)이 실행 뒤 덮어쓴 것을 이 파일로 옮겼다. r1 실행에 실제로 넘긴 프롬프트(이 세션의 기록에서 옮김)는 아래이며, 「원문(마지막 메시지)」은 a198 설계 리뷰 결과 그대로다(본문 내용이 a198 문서·소스를 가리키는 것으로 확인).

```text
TossOS change a198 의 proposal·design·spec delta·tasks 를 적대적 Eng 관점으로 리뷰하라.
대상 문서: openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/{proposal.md,design.md,tasks.md,specs/exit-policy/spec.md}
정본: openspec/specs/exit-policy/spec.md 의 「보호 청산의 길 치우기는 엔진 귀속 주문만 다룬다」, openspec/specs/order-execution/spec.md.
관련 소스(현재 HEAD 로 직접 확인할 것): internal/app/engine/exitloop.go(ExitObserver.clearTheSymbol·record·noteDelay·NewExitObserver), internal/app/engine/exit_held_proposal.go, internal/app/engine/exitwiring.go(Context.ExitObserver), internal/execgw/{riskguardian.go,issue.go,gateway.go,guardian.go,cancel_refusal.go,a094_shared_judgements.go}, internal/journal/{durability.go,decision.go,fills.go,exit_proposal_release.go}, internal/flatten/flatten.go, internal/trading/service.go, internal/app/engine/{broker.go,precheck.go}, cmd/tossctl/engine.go(engineRuntime).
실측: analysis/probe/probe-output.txt, probe-ext-output.txt.

맥락: 실제 돈을 다루는 손절 경로에서, 지금까지 원장이 거절해 한 번도 나가지 않던 **엔진 주문 취소**가 처음으로 실제로 나가게 하는 수리다.
현재 소스로 직접 확인하고 다음을 찾아라:
1. 취소가 처음 나가면서 새로 열리는 위험 — 잘못된 주문 취소, 살아 있는 매도 위에 두 번째 매도, 부분 체결 수량 불일치, 모호한 취소가 손절을 막는 경로, rate limit.
2. 결정 종류·방향을 바꾸는 것이 다른 검사(preimage·멱등 키·replay·복구)와 어긋나는 곳.
3. 조립 실패가 기동을 막아 손절 전체를 멈추는 부작용.
4. 정본 spec 과의 모순.
5. 시험 전략(design §4)이 놓치는 변이.
각 발견을 P0/P1/P2 로 매기고, 파일:줄 근거와 함께 적어라. 사실과 추측을 구분하라.
파일을 수정하지 말 것. 명령은 읽기 전용만 실행할 것(git·go 빌드 캐시 쓰기·시험 실행 금지).
```
