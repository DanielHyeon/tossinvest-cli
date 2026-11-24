codex
**판정: freeze 불가 — 조건 (다)의 안전성 전제와 D5의 추적 종료 판정에 P1 반례가 있습니다.**

1. **P1 — 과거 계좌 수량과의 우연한 일치를 현재 일치로 인정해 기존 차단을 해제합니다. 설계 결정에 닿음: 예(D3·D4·D6, 결정 1).**

   재현 가능한 입력 순서:

   - 이전 주기의 조정으로 종목 S의 로컬·계좌가 20이며, QM 차단과 사용 가능한 이전 credit이 남아 있습니다.
   - 사람이 5주를 매도합니다. 이번 안정화 스냅숏은 계좌 15, 로컬 읽기는 20입니다.
   - **스냅숏 수집 후** 엔진의 별도 매도 5주가 실제 체결되고, 로컬 읽기 뒤 조정 tx 전에 커밋됩니다. 이제 **실제 계좌 10, 로컬 15**입니다.
   - 조정은 `expected 20 / actual 15`로 stale 반환합니다. 종목 워터마크는 증가했고 `Converged`에는 없습니다.
   - 재판정은 **새 로컬 15와 옛 계좌 15**를 같다고 판단합니다. 주문은 처음 스냅숏에서 정상 매칭된 열린 주문으로 두면 missing-order도 필요 없습니다.

   세 조건이 모두 참이지만 실제 5주 차이는 그대로입니다. 삭제된 S는 `disputed`에서 빠지고, 이전 credit으로 **S 자신의 QM이 `ADJUSTMENT_APPLIED` 해제**됩니다. 다른 종목의 credit 해제도 같은 깨끗한 갈래에서 진행됩니다. `cycle.Err`의 stale 오류는 이 해제를 막지 않습니다.

   근거: [design.md:11](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/design.md:11), [RunOnce:451](/home/daniel/tossos-work/a201-impl/internal/app/engine/reconcileloop.go:451), [Observe:650](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:650), [credit 해제:675](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:675), [credit 반박:841](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:841).

   **조건 (다)는 이 반례를 막지 못합니다.** 워터마크는 체결의 *장부 커밋*이 늦었다는 증거이지, 그 체결이 기존 계좌 스냅숏에 이미 반영됐다는 증거가 아닙니다. 정본 [spec.md:77](/home/daniel/tossos-work/a201-impl/openspec/specs/reconciliation/spec.md:77)의 재조회 일치와 [:79](/home/daniel/tossos-work/a201-impl/openspec/specs/reconciliation/spec.md:79)의 재수집을 로컬 재읽기만으로 충족했다는 해석은 안전성 근거가 되지 않습니다.

2. **P1 — D5는 정정 후에도 살아 있는 주문의 missing-order를 삭제할 수 있습니다. 설계 결정에 닿음: 예(D5).**

   입력: 같은 계좌·시장·거래일·종목·side의 주문 `a-parent`가 최초 로컬 읽기에서 추적되고, 브로커 목록에는 없어 missing으로 잡힙니다. 이후 `a-parent → z-child` 정정이 확정되고 자식 주문은 여전히 비종결입니다. 별도 주문의 체결로 해당 종목 수량은 D3을 충족하게 만듭니다.

   `TrackedFillOrders`는 부모·자식을 반환할 수 있지만, `LocalStateFromJournal`은 둘을 **lineage 해석 후 같은 current identity**로 map에 넣습니다. `ORDER BY 1` 순서에서 뒤의 `z-child`가 덮어쓰면 최종 `OriginalOrderID`는 `z-child`입니다. D5는 `a-parent`가 없다는 이유로 기존 missing을 삭제합니다. 그러나 새 로컬을 같은 스냅숏과 실제로 비교하면 **살아 있는 `z-child`가 missing**입니다. 설계에는 그 새 missing을 추가하는 단계가 없습니다.

   근거: [design.md:13](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/design.md:13), [TrackedFillOrders:1639](/home/daniel/tossos-work/a201-impl/internal/journal/fills.go:1639), [정렬:1652](/home/daniel/tossos-work/a201-impl/internal/journal/fills.go:1652), [원 번호와 map 저장:225](/home/daniel/tossos-work/a201-impl/internal/reconcile/compare.go:225), [missing 생성:476](/home/daniel/tossos-work/a201-impl/internal/reconcile/compare.go:476).

   여섯 scope를 맞춰도 **원 번호 소실과 주문 종결은 동치가 아닙니다.** 기존 missing의 승격 키도 사라집니다. a192는 lineage 이동을 명시적으로 거절하지만([order_terminal_release.go:219](/home/daniel/tossos-work/a201-impl/internal/reconcile/order_terminal_release.go:219)), D5는 그 판정에 도달하기 전에 증거를 제거합니다.

   탐침은 `LocalStateFromJournal` 대신 `TrackedFillOrders`를 직접 읽고 원 번호만 map에 넣습니다([probe.diff:115](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/analysis/probe-d5/probe.diff:115)). 따라서 기존 D5-on 결과는 **최종 설계의 이 반례를 검증하지 않았습니다.**

3. **P2 — “차단 지연 상한 한 주기”는 성립하지 않습니다. 설계 결정에 닿음: 예(§2 위험 수용 근거).**

   1번을 반복하면 됩니다. 최초 외부 매도 1주로 차이를 만든 뒤, 매 주기 스냅숏 두 번이 안정화된 **다음** 엔진 부분매도 1주를 커밋합니다.

   | 주기 | 최초 로컬 / 스냅숏 | 체결 후 로컬 / 실제 계좌 | 재판정 |
   |---|---|---|---|
   | 1 | 100 / 99 | 99 / 98 | 삭제 |
   | 2 | 99 / 98 | 98 / 97 | 삭제 |
   | 3 | 98 / 97 | 97 / 96 | 삭제 |

   안정화 실패 없이 실제 1주 차이가 여러 주기 지속됩니다. 재판정에서 매번 삭제되므로 수량 차단도 생성되지 않습니다. 별도 런타임 장애 감시가 개입하는지는 다른 계약이며, D3이 보장하는 한 주기 상한은 아닙니다. 또한 실제 루프는 작업 완료 뒤 period를 기다리므로 60초 자체도 전체 주기 시간의 상한이 아닙니다.

   근거: [design.md:21](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/design.md:21), [Run:412](/home/daniel/tossos-work/a201-impl/internal/app/engine/reconcileloop.go:412), [clearStreaks:655](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:655). 1번과 같은 원인이므로 독립 P1로 중복 계산하지 않았습니다.

나머지 요청 항목의 판정:

- **조건 (나):** 정상 반환한 부분 커밋, `Applied=false`, 알림 실패에서는 `Converged`가 유지됩니다. `ApplyPositionAdjustment`에는 **성공한 Commit 뒤 오류를 반환하는 후속 연산이 없습니다**([position_adjustments.go:366](/home/daniel/tossos-work/a201-impl/internal/journal/position_adjustments.go:366)). Commit 오류의 실제 durable 결과가 모호한 경우까지 이번 읽기로 입증하지는 못했으며, 확정 결함으로 세지 않았습니다.
- **오류 처리:** 앞선 Ingest/수렴 오류는 Observe 실행을 막지 않습니다. Refresh 실패는 재판정 자리 전에 반환합니다. Observe의 일부 해제가 커밋된 뒤 오류가 나도 이미 해제된 행은 유지됩니다. 따라서 `cycle.Err != nil`을 1번 반례의 안전장치로 볼 수 없습니다.
- **D5 추가 경계:** `terminal=true, fail_closed=true`도 `TrackedFillOrders`의 `terminal=0` 조건에서 제외됩니다([fills.go:1604](/home/daniel/tossos-work/a201-impl/internal/journal/fills.go:1604)). 추적 부재가 깨끗한 종결 증명이라는 주장도 성립하지 않습니다. 다만 별도 fail-closed 차단까지 사라진다고 확인하지 않았으므로 별도 P1로 세지 않았습니다.
- **a192:** 기존 결속이 있는 *같은 종목*은 뒤 체결 워터마크 검사로 해제가 거절됩니다([order_terminal_release.go:258](/home/daniel/tossos-work/a201-impl/internal/reconcile/order_terminal_release.go:258)). 이를 무시한 “수량 항목 삭제만으로 모든 a192 차단 해제” 주장은 하지 않습니다. 반면 이전 credit 경로는 이 검사를 거치지 않습니다.
- **D2:** 현재 SQLite의 AUTOINCREMENT와 직렬 쓰기에서 id 커밋 순서 반례는 찾지 못했습니다. 문제는 id 순서가 아니라 **계좌 스냅숏과 체결의 인과관계를 증명하지 못한다는 점**입니다.

HEAD `4c9d53138ab1ac81ba0a4d099dcd7cafb3c757da`의 작업 파일과 지정 증거를 정적으로 검토했습니다. 위 입력은 **실행하지 않은 반례**이며 기존 로그의 PASS는 과거 탐침 결과입니다. 파일 변경·빌드·시험·네트워크·tossctl·docker 실행 0. ai-memory MCP 부재, CLI 조회 미실행. `impeccable: N/A — backend 설계 리뷰, UI 변경 없음`.

**「설계 결정에 닿는 P0/P1」: 예 — P1 2건.**
hook: Stop
hook: Stop
