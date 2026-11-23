# proposal-freeze 리뷰 2회차(마지막) — codex (원문)

- 날짜: 2026-11-23. 보이스: codex-cli 0.165.0(작성자와 다른 계열), `codex exec --sandbox read-only --ephemeral -C <worktree> -o … "<프롬프트>" </dev/null`, rc=0.
- 대상: Manager 결정(2026-11-23) 반영 뒤 미커밋 문서(HEAD 636eb76f5).
- 프롬프트 원문:

```text
TossOS change a194 proposal-freeze 리뷰 2회차(마지막)다. 대상: openspec/changes/a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported/ (proposal.md, design.md, specs/exit-policy/spec.md, specs/reconciliation/spec.md, specs/engine-safety/spec.md(새 MODIFIED), tasks.md, review.md; MODIFIED 생성 근거 analysis/spec-delta/). 실제 돈을 다루는 Go 자동매매 엔진의 대사·편입 경로 알림 추가다. 적대적 Eng 관점으로 본다.

1회차 발견 8건과 Manager 처분(2026-11-23, design.md 결정 표·§4·§5·§6 에 반영):
- R1-1(P1) 알림 꺼짐 일반 등급 경로 없음 → 보유 보고는 꺼짐이면 exit.position_unmanaged + 새 사실 값 + (포지션,사실) 메모리 래치, 차단 보고는 꺼짐이면 critical 없이 로그 한 줄(그날 한 번). 꺼진 엔진에서 outbox critical·진입 래치·모드 승격 0 을 시험으로 고정.
- R1-2(P1) D5 하루 한 번이 Notify(창 1시간)로 불가 → 새 배선 없음(RecordCritical 안 끌어옴, Context.ReconcileDriver 무편집). 드라이버 메모리 래치(key=타입·범위·종목·원인·KST 날짜, outbox key 동일)를 Notify 가 nil 을 돌려준 뒤에만 세움. 재시작 시 그날 한 번 더(공시).
- R1-3(P1) D6 현재형 → 관측 시각 + 과거형, blocked_stale 에는 운영자 해제 안내 없음.
- R1-4(P1) ADDED 만으로 정본 무알림 Scenario 를 못 좁힘 → exit-policy·engine-safety MODIFIED 둘(스크립트로 정본을 떠서 정확한 치환, analysis/spec-delta/).
- R1-5(P2) 실패·미수렴 주기 침묵 → 범위 밖으로 공시(proposal 「하지 않는 것」, design §6).
- R1-6(P2) D1 판정 재료 → Tracker.Blocks() 의 Covers ∧ Release==ReleaseOperatorOnly, 원인·시각은 그 블록의 Cause·Since(가장 이른 Since). blocked() 참인데 Blocks() 에 덮는 블록이 없으면 스스로 풀리는 차단과 같은 길(10분 뒤 원인 「미상」).
- R1-7(P2) 시험 공백 → 전부 수용(Notify 호출 수·전달 시도 수로 단언, 기록 실패→재시도 RED, KST/UTC 자정, 실 Notifier+outbox 하네스, 정규식 보강, 변이 3개 추가).
- R1-8(P2) ExitEligible 과장 → design §3 정정, ADDED 괄호 「이미 exit 관리 대상이다」.
그 밖: cycle.Unmanaged 는 올리지 않음(adoption_include_test.go:133 무편집 통과). 편집하는 기존 함수 본문은 judgeHoldings 하나(구조체 새 필드, 맵은 새 함수 안에서 생성).

할 일: 현재 소스(internal/app/engine/adoption.go, reconcileloop.go, internal/reconcile/mismatch.go, internal/obs/notifier.go·event.go, internal/app/engine/alertdelivery.go, 관련 시험)와 정본 openspec/specs/{exit-policy,reconciliation,engine-safety}/spec.md 로 직접 확인해 (1) 처분이 틀렸으면 근거(파일:줄)와 함께 반박하고, (2) 새 발견은 P0/P1/P2 로 적어라. 특히: 처분이 만든 새 침묵·폭주 경로, Notify 가 critical 을 동기 전송하는 동안의 대사 루프 영향, MODIFIED 블록이 정본과 의도한 편집 외에 다른지, 결정끼리 모순, spec delta 와 design 의 불일치. 각 발견이 사실 정정인지 결정 변경인지 적어라.
파일을 수정하지 말 것. 읽기 전용 명령만. go test·make 금지.
```

## 최종 메시지 원문

**freeze 보류 권고 — P0 0, P1 2, P2 3.** 현재 HEAD `636eb76f5a0632aed3cebb11735bf6f6c8a0be72`와 미커밋 문서를 기준으로 검토했습니다. 아래는 구현 결함 판정이 아니라 계약·시험 계획의 발견입니다.

1. **R2-1 · P1 — D5의 기록 성공 여부를 지정된 호출 경로로는 알 수 없다.**  
   [design.md:13](/home/daniel/tossos-work/a194-impl/openspec/changes/a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported/design.md:13)은 `d.alert → Alerts.Notify`를 유지하면서 `Notify == nil` 뒤에만 래치를 세우라고 합니다. 그러나 [reconcileloop.go:568](/home/daniel/tossos-work/a194-impl/internal/app/engine/reconcileloop.go:568)의 `alert`는 **반환값이 없고 오류를 로그로 삼킵니다**. 기존 함수 본문 편집은 `judgeHoldings` 하나로 제한되어 있습니다(design:20).

   따라서 지정 경로 뒤에 무조건 래치를 세우면 기록 실패 이후 그날의 차단 보고가 사라집니다.  
   **분류: 사실 정정·호출 경로 명확화.** 새 보고 함수에서 `opts.Alerts.Notify` 반환값을 직접 확인하면 기존 함수 편집 범위를 늘리지 않고 해결할 수 있습니다. `Alerts == nil`을 기록 성공으로 취급해서도 안 됩니다.

2. **R2-2 · P1 — 동기 전송 비용이 보유 수만큼 누적된다. 미전달 행에는 1시간 창도 적용되지 않는다.**  
   [notifier.go:354](/home/daniel/tossos-work/a194-impl/internal/obs/notifier.go:354)는 `deliver`를 동기로 호출합니다. 기본값은 전송당 10초, 3회 시도, 사이 대기 2초입니다. 전송이 모두 시간 초과하면 **보고 하나당 네트워크·대기만 34초**입니다. 원장 쓰기 대기를 포함한 기존 전달 예산은 [alert_lease.go:57](/home/daniel/tossos-work/a194-impl/internal/obs/alert_lease.go:57)에 54초로 명시되어 있습니다.

   계좌 전체 차단 아래 보유 10개라면 첫 주기에 보유 보고 10개와 차단 보고 1개가 직렬 전송되어, 네트워크·대기만 약 **6분 14초** 걸릴 수 있습니다. 뒤쪽 보고는 그동안 durable 기록조차 시작하지 못합니다. [Run:405](/home/daniel/tossos-work/a194-impl/internal/app/engine/reconcileloop.go:405)는 이 작업이 끝난 뒤 60초를 쉬므로 다음 대사·편입도 늦어집니다.

   더구나 [outbox.go:379](/home/daniel/tossos-work/a194-impl/internal/journal/outbox.go:379)는 `PENDING`이면 항상 전달 대상으로 판정합니다. 1시간 창은 `DELIVERED/ACKNOWLEDGED`에 적용됩니다. 따라서 D3의 매 관측 `Notify`는 장애 중 매 주기 전송 예산을 다시 소비합니다. “보유당 1시간마다 1건”(proposal:58)은 **성공적으로 정산된 경우로 한정해야** 합니다.

   **분류: 사실 정정 + Manager 결정 필요.** 동기 지연을 수용할지, 기록·전송 분리를 허용할지 확정해야 합니다. 유지한다면 새 보고와 기존 `adopt`의 호출 순서, 다중 보유·느린 전송·PENDING 반복 시험이 필요합니다. 기존 exit 루프 전체가 함께 멈춘다는 근거는 없으며, 여기서 확인한 영향은 대사 진행과 후속 보고·편입 지연입니다.

3. **R2-3 · P2 — MODIFIED 복사는 정확하지만, 완성된 계약에는 세 가지 모순이 남는다.**  
   현재 정본과 실제 delta를 직접 비교했습니다. 의도한 **E1·E2·S1·S2 외 본문 차이는 없습니다.** 문제는 유지한 문장의 적용 범위입니다.

   - [engine-safety delta:25](/home/daniel/tossos-work/a194-impl/openspec/changes/a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported/specs/engine-safety/spec.md:25)는 여전히 모든 critical 무관리 보고에 **“편입 시도가 실패한 시각”**을 요구합니다. 새 :5·:133은 시도하지 않은 차단 보고를 포함하므로, 실패 시각과 차단 관측 시각을 분기해야 합니다.
   - [exit-policy delta:19](/home/daniel/tossos-work/a194-impl/openspec/changes/a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported/specs/exit-policy/spec.md:19)의 두 Scenario는 알림 ON 전제 없이 critical을 요구합니다. 같은 파일 :17·:27의 OFF 일반 등급 규칙과 충돌합니다. reconciliation delta:7도 같은 문제가 있습니다.
   - engine-safety delta:5는 **“10분 넘게”**, exit-policy delta:5·:24와 design:37은 **“10분 이상”, 정확히 10분에 보고**입니다. `>`와 `>=`가 갈립니다.

   **분류: 승인된 결정에 맞추는 사실·계약 정정.** 새 정책은 필요하지 않습니다. 다만 “다른 문장·Scenario는 바꾸지 않는다”(design:60)는 편집 제한은 위 정정에 맞춰 풀어야 합니다.

4. **R2-4 · P2 — 추가한 변이를 죽일 시험 입력이 아직 빠져 있다.**  
   [design.md:36](/home/daniel/tossos-work/a194-impl/openspec/changes/a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported/design.md:36)의 시험 목록과 :51의 변이 목록이 완전히 대응하지 않습니다.

   - `Blocks().Release` 대신 `EntryAllowed` 사유로 판정하는 변이를 잡으려면 **SNAPSHOT_STALE/UNAVAILABLE 복원 블록**처럼 두 판정이 갈리는 입력이 필요합니다. 현재 열거된 귀속 실패·수량 불일치·계좌 영구 차단으로는 이 차이를 입증하지 못합니다. 실제 차이는 [mismatch.go:1398](/home/daniel/tossos-work/a194-impl/internal/reconcile/mismatch.go:1398)와 [symbolgate.go:157](/home/daniel/tossos-work/a194-impl/internal/execgw/symbolgate.go:157)에 있습니다.
   - 시험 9는 **D5 차단 보고 래치**의 실패→재시도입니다. D3 보유 보고 앞에 래치를 넣는 변이는 별도로 **보유 보고 기록 실패→같은 사실 재관측→기록 성공**을 검증해야 합니다. 시험 1의 “창 안에서 전달 시도가 늘지 않음”만으로는 잘못된 메모리 래치도 통과합니다.
   - D3는 전달 완료 후 **1시간 경과 뒤 재전송**도 명시적으로 검증해야 합니다.

   **분류: 시험 계획 보완.** R1-7을 수용했다는 상태와 별개로, 위 변이를 `CAUGHT`로 만들 구체적 입력·단언이 필요합니다.

5. **R2-5 · P2 — “대사 주기가 실패하면 두 보고 모두 없다”는 공시가 코드보다 넓다.**  
   [design.md:68](/home/daniel/tossos-work/a194-impl/openspec/changes/a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported/design.md:68)와 proposal:39는 실패 주기 전체를 무보고로 설명합니다. 그러나 [RunOnce:438](/home/daniel/tossos-work/a194-impl/internal/app/engine/reconcileloop.go:438)은 ingest·converge 오류를 `cycle.Err`에 남긴 뒤 계속 진행합니다. 이후 Refresh·Observe가 성공하면 :470의 `judgeHoldings`가 실행됩니다.

   **분류: 사실 정정.** 무보고 범위는 안정화 미성공·로컬 상태 조회 실패·Refresh 실패·Observe 실패 등 **실제로 조기 반환하는 경로**로 적어야 합니다. 모든 `cycle.Err != nil` 주기를 건너뛰도록 구현하면, 승인한 기존 경로 보존을 벗어나 새 침묵을 만듭니다.

R1-1의 OFF 타입 분리, R1-6의 `Covers ∧ ReleaseOperatorOnly`, `cycle.Unmanaged` 비증가 처분은 현재 코드·관련 시험과 부합합니다. 알림 ON의 critical 미전달에 따른 계좌 진입 차단·모드 승격도 이미 공시된 기존 동작이므로 별도 새 발견으로 세지 않았습니다.

파일 수정·`go test`·`make`·운영 접근은 하지 않았습니다. MODIFIED 생성기는 실행하지 않고 읽기 전용 비교만 수행했습니다. ai-memory 조회는 샌드박스의 로컬 서버 연결 거부로 실패했으며, 결론은 현재 원문에 근거합니다. 시험·변이·gstack·graph 재실행 결과는 주장하지 않습니다. `impeccable: N/A — backend/documentation-only`, 화면 변경 없음.