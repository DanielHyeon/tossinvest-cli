codex
**delta `d31347ab8..7621c5c2e`에서 새 P0–P2 결함은 찾지 못했습니다.** V4c가 잘못된 credit을 삭제하거나 필요한 반박을 빠뜨리는 구체적 반례는 확인되지 않았습니다. 다만 검증·처분 설명에서 P3 두 가지를 구분해야 합니다.

**1. P3 — M21은 행동 전체가 동일하지 않습니다. 해제 결과만 같습니다.**

다음 입력은 현재 구현과 “Refresh 앞에서 무조건 반박” 변이를 구별합니다.

| 주기 | 브로커·장부·체결·수렴 |
|---|---|
| k | S 보유가 양쪽 10. 로컬 SELL 주문이 브로커 열린 주문에 없어 missing-order 차단 삽입·a192 결속. 체결 없음 |
| k+1 | 외부 매도로 브로커 0, 로컬 10. 엔진 체결 없음. 수렴 10→0 커밋으로 조정 credit 생성. 같은 행에 수량 흔적이 남아 결속이 tainted |
| 주기 사이 | SELL 주문 취소 확정. 외부 매수로 브로커 S 보유 1 |
| k+2 | 브로커 1 / 로컬 0 → 외부 포지션. ingest가 1을 접고 수렴 조정은 없음. Refresh 성공 |

현재 코드는 credit이 있으므로 `Observe`의 분쟁→Awaiting 경로를 거친 뒤 credit을 지웁니다. M21은 credit을 먼저 지워 `orderTerminalCandidate`를 호출합니다. 따라서 **추가 장부 조회가 발생하고 보류 사유가 `quantity_dispute_seen`으로 바뀝니다.** 둘 다 차단은 유지합니다.

근거: [mismatch.go:683](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:683), [order_terminal_release.go:374](/home/daniel/tossos-work/a201-impl/internal/reconcile/order_terminal_release.go:374). 현재 엔진 시험은 이 진단 차이를 단언하지 않습니다. [ledger.md:52](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/analysis/mutation/ledger.md:52)의 상세 설명은 이를 인정하지만, “행동 결과 무변”은 **차단·해제 결과에 한정**해야 합니다. 현재 생산 구현 결함은 아닙니다.

**2. P3 — R3 수용은 검증 공백 수용이며, 두 구현이 현재 V2 계약에 모두 맞는다는 뜻은 아닙니다.**

반례 입력:

- k: A·B 모두 수량 불일치. 표식 이후 각각 체결 커밋, 수렴은 둘 다 미반영, Observe 성공.
- credit 계산에서 A watermark 조회 성공 후 B 조회 실패.
- k+1: A·B 모두 새 비교에서 일치하고 추가 체결·조정 없음.

현재 `return nil`은 둘 다 credit을 주지 않습니다. 변이 `return out`은 A에 credit을 주어 k+1에 A를 해제할 수 있습니다. 이는 [design.md:12](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/design.md:12)의 **“하나라도 못 읽으면 그 주기 아무 종목에도 주지 않음”**과 다릅니다.

현재 [a201_fill_credit.go:65](/home/daniel/tossos-work/a201-impl/internal/reconcile/a201_fill_credit.go:65)는 계약대로 동작합니다. 시험을 추가하지 않는 처분은 유지할 수 있지만, [ledger.md:40](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/analysis/mutation/ledger.md:40)의 “둘 다 design V2 안”이라는 근거에는 동의하지 않습니다. 기존 R3 검증 공백의 재확인이며 새 생산 결함은 아닙니다.

**V4c 오류·반환 경로 확인**

- `Refresh` 반환은 세 가지입니다. Journal 없음→성공, `ActiveReconcileStates` 오류→실패, 정상 완료→성공. **오류 반환은 한 곳이고 tracker 변경 전입니다.** `blocks`, permanent/pending 상태, streak, credit, gate를 부분 변경한 뒤 오류를 반환하는 경로는 없습니다. [mismatch.go:1038](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:1038)
- 반박은 `disputed[symbol] && observed.After(credit)`일 때만 삭제합니다. 같은 as-of의 당해 수렴 credit, 다른 종목 credit, 더 미래 credit은 유지됩니다. 빈 값·파싱 불가능한 as-of도 삭제하지 않습니다. 따라서 해당 관측에서는 해제도 못 하지만, **뒤의 유효한 비교에서는 남은 credit을 사용할 수 있습니다.** 이는 기존 시간 비교 계약과 같습니다. [mismatch.go:512](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:512)
- `RunOnce`의 안정화 실패 반환과 로컬 읽기 오류 반환은 모두 `Compare` 전입니다. 반박할 완성된 diff가 없습니다. ingest·convergence 오류는 `cycle.Err`에 남기고 Refresh/Observe까지 진행하므로 반박을 건너뛰지 않습니다. [reconcileloop.go:430](/home/daniel/tossos-work/a201-impl/internal/app/engine/reconcileloop.go:430)
- `Observe`의 반환 세 곳 모두 반박 경로가 있습니다.

| 반환 | 선행 반박 |
|---|---|
| 깨끗한 갈래 authority 오류, :669 | V4b 호출 :668 |
| blocking 갈래 authority 오류, :743 | 호출 :728 |
| 정상 또는 persist 오류, :867 | 끝 credit 루프 :841–856. persist 오류도 이 루프를 통과 |

근거: [mismatch.go:635](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:635).

**차단 유지의 영향은 진입 제한뿐만은 아닙니다.**

Refresh 실패 주기는 기존처럼 편입 판정을 건너뛰고 health 실패 횟수를 올립니다. V4c가 이후 차단을 유지시키면 해당 종목의 편입도 계속 막힙니다. 기존 관리 포지션의 청산은 계속 가능하지만 RECONCILE 확정 하한이 적용되며, 하한 조회 실패 시 제출 수량은 0입니다. 따라서 “운영자 해제까지 진입만 막힘”으로 축약하면 부정확합니다. 이는 이번 반박 정책의 기존 차단 경로상 결과이며, 잘못된 credit 삭제 반례는 아닙니다.

근거: [adoption.go:204](/home/daniel/tossos-work/a201-impl/internal/app/engine/adoption.go:204), [exitwiring.go:192](/home/daniel/tossos-work/a201-impl/internal/app/engine/exitwiring.go:192), [exitloop.go:1766](/home/daniel/tossos-work/a201-impl/internal/app/engine/exitloop.go:1766).

**동시성**

새 메서드는 `t.mu`를 잡은 상태에서 `symbolsInDispute` 계산과 credit 삭제를 모두 수행합니다. Refresh·Observe·credit 부여도 같은 mutex를 사용합니다. Refresh 실패와 반박 호출 사이에는 잠금 공백이 있지만, 새로 들어온 같은/더 늦은 as-of credit을 지우지는 않습니다. diff는 이 RunOnce가 만든 지역 비교 결과이며, 확인한 경로에서 이를 병행 수정하는 호출자는 없습니다. 새 data race나 잘못된 해제 인터리빙은 찾지 못했습니다.

단, 구조 시험은 `Lock`/`Unlock` **호출 개수**를 검사합니다. 잠금이 실제로 반박을 감싸는 순서까지 증명하는 시험은 아닙니다. 현재 순서는 [a201_fill_credit.go:111](/home/daniel/tossos-work/a201-impl/internal/reconcile/a201_fill_credit.go:111)에서 직접 확인했습니다.

**수정 시험이 겨냥한 가드를 실제로 구별하는가**

소스와 기존 변이 로그의 실패 사유를 대조했습니다. 이번 회차에서 재실행하지는 않았습니다.

| 시험 | 확인 결과 |
|---|---|
| R2 | 뒤 주기에 새 조정이 없어 덮어쓰기 은폐가 사라짐. M17은 실제 `Released:1` 때문에 실패 |
| R4 | 표식 바로 다음 문장이 로컬 읽기인지 검사. M18은 그 인접성 위반으로 실패 |
| R5 | 사용 가능한 체결 credit을 남긴 채 `ATTRIBUTION_FAILED` 행으로 교체. M19는 그 운영자 전용 행을 실제 해제하여 실패 |
| G2 | 다른 계좌 체결은 **결속 watermark 읽기 전** 발생. 대상 계좌의 수량 불일치·외부 포지션 흔적은 없고, 취소 후 비교는 깨끗함. 원래 주장 유지. M01은 `ORDER_TERMINAL_CONFIRMED` 대신 `FILL_APPLIED`가 기록되어 실패 |
| G3 | watermark 이동과 수렴 성공이 함께 존재. M03은 `ADJUSTMENT_APPLIED` 대신 `FILL_APPLIED`가 기록되어 실패 |
| 새 Refresh 시험 | M20 삭제 시 체결·조정 두 판 모두 다음 깨끗한 주기에 실제 해제되어 실패. 비분쟁·동일 as-of 대조군은 유지 |

G2의 최종 원인 단언은 [a201_fill_credit_cycles_test.go:583](/home/daniel/tossos-work/a201-impl/internal/app/engine/a201_fill_credit_cycles_test.go:583)에 있습니다. 다만 이 시험 하나가 a192의 정확한 두 비교 타이밍까지 검증하는 것은 아닙니다.

**문서 및 R6 처분**

V4c 행, 새 Scenario, MODIFIED 문장은 현재 호출 위치·엄격한 선후 조건과 맞습니다. R1의 evidence 문자열·상수 주석·설계·Scenario 본문도 “로컬 읽기 직전 표식 뒤”로 정렬됐습니다. 남은 표현상 흠은 `엄격히 나중 credit`이라는 역전된 표현과 V8 네 항목 뒤의 `셋 다`입니다. 정확한 비교식과 규범 문장은 명확하여 별도 P2로 올리지 않았습니다.

R6의 **현재 생산 호출 그래프에서 도달 불가** 처분은 뒤집을 근거를 찾지 못했습니다. 시장 범위 행 작성 조건은 released owner이고, `releaseRiskBucketOwner`의 생산 호출자는 확인되지 않습니다. 직접 행을 주입하면 해제 오류가 반복되는 기제는 남지만 delta가 도입한 결함은 아닙니다. 운영 DB 건수 주장은 이번에 검증하지 않았습니다.

검토 기준 HEAD는 `7621c5c2e8a8dd4271fe02ac10081c020485676f`, 종료 시 작업 트리도 깨끗합니다. 파일 수정·`go test`·빌드·변이 실행은 하지 않았습니다. 기존 시험 로그는 관련 실패 구간만 읽었으며 전체 검증 PASS를 재확인한 것은 아닙니다. ai-memory CLI 조회는 sandbox의 연결 거부로 불가하여 저장소 원문으로 검토했습니다. `impeccable: N/A — 백엔드 읽기 전용 리뷰`.

---

<!-- 리뷰어(Opus) 주: 위는 codex 최종 답 원문 그대로(전사 5617 줄 중 5454-5532). 실행: `codex exec --sandbox read-only -C /home/daniel/tossos-work/a201-impl "<prompt>" </dev/null`,
model gpt-6-astra, session 01a20be7-2244-75a1-b850-7a6c82e9fc11, exit 0. 전사 전체와 프롬프트는 세션 스크래치에만 있음 —
/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/6ba8fcd4-b024-437e-a4c2-ebe621826fe5/scratchpad/codex-r2.txt ·
/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/6ba8fcd4-b024-437e-a4c2-ebe621826fe5/scratchpad/codex-r2-prompt.txt
(프롬프트 요지: 델타 d31347ab8..7621c5c2e 만 · V4c · 초점 1~7 · 「파일 변경 · go test 금지 · 반례 입력을 구체적으로」). -->
