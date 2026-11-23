codex
**판정: freeze 불가 — v2에도 해제 증거 누락과 지속 장애 감시 무력화에 설계 결정 수준 P1이 있습니다.**

1. **P1 — 주기 전체를 버리면 다른 종목의 a192 결속 증거가 사라져, 이후 해제 조건이 완화됩니다. 설계 결정에 닿음: 예(D3'·D4'·D6', p2).**

   재현 입력:

   - B에 이 프로세스가 O1의 missing-order로 만든 차단이 있습니다. O1은 무체결 취소로 종결됐고, 종결 증명은 이미 무장됐습니다.
   - 같은 B의 별도 기존 주문 O2는 비종결입니다. 이전에는 브로커 목록과 정상 매칭됐습니다.
   - 이번 안정화 스냅숏에서 O2가 누락되어 `MissingOrders`에 들어갑니다. B의 수량은 일치하고 B 체결은 없습니다.
   - 무관한 A에서 W2가 발생하여 **주기 전체를 보류**합니다.
   - 다음 주기에는 O2가 브로커 목록에 다시 나타나 정상 매칭됩니다. 비교 전체는 깨끗하고 O2는 여전히 비종결입니다.

   기존 경로는 이번 관측에서 O2를 결속에 추가하므로, 이후 O2 비종결을 이유로 차단을 유지합니다. v2는 그 호출을 생략하여 **O1만 확인하고 `ORDER_TERMINAL_CONFIRMED`로 해제할 수 있습니다.** B에는 뒤 체결이 없으므로 C12 워터마크도 막지 못합니다. 해제 tx 재검증 역시 누락된 O2를 복원하지 않습니다.

   근거: [Observe:647](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:647), [결속 주문 추가:326](/home/daniel/tossos-work/a201-impl/internal/reconcile/order_terminal_release.go:326), [비종결 거절:228](/home/daniel/tossos-work/a201-impl/internal/reconcile/order_terminal_release.go:228), [워터마크·무장 판정:258](/home/daniel/tossos-work/a201-impl/internal/reconcile/order_terminal_release.go:258), [tx 재판정:416](/home/daniel/tossos-work/a201-impl/internal/reconcile/order_terminal_release.go:416).

   정본 [spec:381](/home/daniel/tossos-work/a201-impl/openspec/specs/reconciliation/spec.md:381)의 **행 생존 중 관측된 missing-order 전부**라는 조건과 충돌합니다. “Observe 함수 무편집”은 “해제 효과 불변”을 보장하지 않습니다.

   같은 원인은 승격에도 적용됩니다. B의 동일 missing-order가 두 번 관측된 뒤, A 때문에 보류한 주기에는 B가 정상 매칭되고, 다음 주기에 B가 다시 missing이면 기존 streak 2가 남아 **비연속 관측으로 3회 영구 승격**합니다. 기존 경로는 중간 비교에서 B의 streak를 제거합니다. 근거: [advanceStreaks:1089](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:1089), [정본:53](/home/daniel/tossos-work/a201-impl/openspec/specs/reconciliation/spec.md:53).

   **처분:** 보류 대상을 이유로 다른 종목의 결속·실격·연속성 증거까지 버리는 결정은 수정해야 합니다.

2. **P1 — D7'의 성공 처리가 지속 장애의 운영 모드 강화를 계속 막을 수 있습니다. 설계 결정에 닿음: 예(D7').**

   재현 입력:

   - 차단 없는 편입 후보 B의 가격 조회가 매번 실패합니다. 수집·장부 읽기·수렴은 정상입니다.
   - 홀수 주기마다 서로 다른 A 종목에서 W2가 발생합니다. 짝수 주기에는 보류 조건이 없습니다.
   - 기본 열화 임계는 연속 실패 5회입니다.

   | 경로 | 연속 실패 계수 |
   |---|---|
   | 기존 | B 가격 조회를 매번 수행하여 `1 → 2 → 3 → 4 → 5`, 운영 모드 강화 |
   | v2 | 홀수 주기는 편입을 생략하고 성공 처리하여 `0 → 1 → 0 → 1 → 0`, 임계 미도달 |

   **연속 보류 상한 1을 지켜도 장애 은폐 기간에는 상한이 없습니다.** 안정화 실패가 기존에도 성공 처리된다는 사실은, 동일한 안정화 성공 입력에서 발생하는 이 차이를 없애지 않습니다.

   근거: [편입 조회 오류:222](/home/daniel/tossos-work/a201-impl/internal/app/engine/adoption.go:222), [성공 시 계수 초기화:301](/home/daniel/tossos-work/a201-impl/internal/app/engine/reconcileloop.go:301), [건강 임계 판정:420](/home/daniel/tossos-work/a201-impl/internal/app/engine/runtime.go:420), [운영 모드 강화:462](/home/daniel/tossos-work/a201-impl/internal/app/engine/runtime.go:462), [설계의 한 주기 영향 주장:15](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/design.md:15).

   이미 강화된 운영 모드가 자동 완화되지는 않습니다. 문제는 **앞으로 이루어져야 할 강화가 발생하지 않는 것**입니다. 건강 래치는 계수 0에서 해제됩니다.

   **처분:** 보류를 성공으로 취급하는 건강 계약과 교대 보류 입력을 다시 결정해야 합니다.

3. **P2 — Refresh를 먼저 실행하므로 “보류 주기의 차단·연속 횟수 무변”은 성립하지 않습니다. 설계 결정에 닿음: 예(D1'·D4', ADDED).**

   재현 입력:

   - 이전 영구 승격 쓰기가 실패하여 pending account block이 남습니다.
   - 다음 Observe에서 승격을 얻은 dispute가 사라졌지만 authority read가 실패하여, `pendingPermanent=true`, 빈 `pendingPermanentKey`가 남습니다.
   - 이번 Refresh는 성공하고 durable permanent row가 없음을 확인합니다. 이후 A의 W2 조건으로 보류합니다.

   Refresh는 **pending account block을 제외하고 streak를 지운 뒤 gate를 갱신**합니다. 따라서 Observe를 생략해도 이번 주기의 상태는 이미 바뀌었습니다.

   근거: [이전 authority 오류 경로:708](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:708), [Refresh의 proposal 철회:1044](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:1044), [streak 초기화·gate 갱신:1070](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:1070), [ADDED:5](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:5).

   이는 기존 Refresh의 정당한 복구 동작입니다. 이를 금지하기보다, 계약을 **“보류 판정 이후 Observe에 의한 변경 없음”**으로 한정해야 합니다. durable permanent row가 임의 해제된다는 발견은 아닙니다.

4. **P2 — 연속 상한 1은 재기동 및 관측 없는 중간 주기를 포함한 판정 지연 상한이 아닙니다. 설계 결정에 닿음: 예(D5').**

   - **재기동:** 보류 → 드라이버 재생성 → 다른 종목 W2 → 보류를 반복하면 메모리 불리언은 매번 초기화됩니다. 성공적으로 판정한 주기 없이 보류가 이어질 수 있습니다.
   - **중간 조기 반환:** 보류 → 불안정/로컬 읽기 오류/Refresh 오류 → W2에서, “직전 주기”를 문자 그대로 적용하면 다시 보류합니다. 반대로 기존 탐침은 보류 검사 지점까지 도달해야 불리언을 초기화하므로, 이 입력에서는 보류하지 않습니다.

   근거: [D5':13](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/design.md:13), [조기 반환:428](/home/daniel/tossos-work/a201-impl/internal/app/engine/reconcileloop.go:428), [탐침 초기화:188](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/analysis/probe-v2/probe.diff:188).

   **처분:** 상한을 프로세스 내 보류 횟수로 주장할지, 실제 판정을 반드시 한 번 수행한다는 계약으로 잡을지 구분해야 합니다. 현재 탐침은 최종 문구의 중간 주기 의미를 검증하지 않았습니다.

5. **P2 — ADDED가 보류의 예외와 다른 종목의 기존 계약을 충분히 정의하지 않습니다. 설계 결정에 닿음: 예(정본 delta).**

   구체적인 충돌:

   - A의 W2와 B의 독립적인 missing-order가 함께 있으면, B에는 체결도 수량 변화도 없는데 신규 차단을 생략합니다. **“체결이 지나간 비교는 확인된 불일치가 아니다”**만으로 B의 확정된 불일치까지 부정할 수 없습니다. [정본:53](/home/daniel/tossos-work/a201-impl/openspec/specs/reconciliation/spec.md:53)의 즉시 ordinary block 계약에 명시적 예외가 필요합니다.
   - 첫 시나리오는 W2/W3이면 무조건 보류한다고 쓰지만, 직전 보류·판독 실패 예외가 빠졌습니다. [ADDED:8](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:8)와 [:16](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:16)이 같은 입력에 반대 결과를 요구합니다.
   - 마지막 credit 시나리오는 다음 비교가 해당 종목에 일치하면 해제를 보장하지만, 다른 종목의 missing-order로 `diff.BlocksEntry()`가 참이면 현재 Observe는 해제 갈래에 들어가지 않습니다. [ADDED:28](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:28), [Observe:650](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:650).
   - 다른 종목의 실제 disputed 관측까지 버리면 credit 반박도 생략됩니다. 이는 [정본:312](/home/daniel/tossos-work/a201-impl/openspec/specs/reconciliation/spec.md:312)의 심볼별 후속 관측 계약과의 관계를 명시해야 하는 변경입니다.

   **처분:** ADDED 한 문장으로 정본 :53·:303·a192의 관측 의미가 그대로 보존된다고 처리할 수 없습니다.

나머지 요청 항목의 판정:

- **R1-1의 즉시 오해제는 닫혔습니다.** 해당 주기에 Observe 자체를 호출하지 않으므로 이전 credit으로 해제되지 않습니다. 기존 탐침도 이를 보여 줍니다. 다만 이것이 위의 다른 종목 증거 누락까지 닫았다는 뜻은 아닙니다.
- **과잉 보류는 보수 방향만이 아닙니다.** 다른 계좌의 동일 종목 체결, 수렴이 거절한 종목의 무관한 체결도 조건을 만족할 수 있습니다. 이때 차단 없는 정상 편입 후보까지 `judgeHoldings`를 건너뛰므로, 기존에는 붙었을 자동 손절이 늦어집니다. 이미 보호 중인 포지션의 청산 경로와 신규 보호 부착은 구분해야 합니다.
- **누락 조건:** 공시된 backfill·판독 실패·연속 둘째 주기 외에, 읽은 범위에서 새로운 W2/W3 누락 반례는 확정하지 못했습니다. `delta=0`은 `MoveNone`이며 `NewInstance=true` 행들은 `MoveAdds`이므로, 설계가 미확인으로 남긴 **delta 0 새 인스턴스 추측은 현재 전이표로 지지되지 않습니다.** [position.go:293](/home/daniel/tossos-work/a201-impl/internal/position/position.go:293), [table.go:64](/home/daniel/tossos-work/a201-impl/internal/position/table.go:64).
- **탐침의 판독 실패 처리는 D8'과 다릅니다.** A의 워터마크가 조건을 만족하고 B 판독이 실패하면 탐침은 계속 보류합니다. 전 주기 판독 실패 시 보류 금지라는 계약이라면 불일치입니다. [probe.diff:178](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/analysis/probe-v2/probe.diff:178). 기존 PASS를 이 오류 계약의 검증으로 사용할 수 없습니다.

HEAD `c0ca86df5b71a486a654023215be02176392805a`에서 지정 문서·탐침 원문·재현 시험과 관련 소스 구간을 정적으로 검토했습니다. 위 신규 반례는 **미실행 입력 추론**입니다. 기존 AST/FLM은 참조했으며 재추출·그래프 갱신은 하지 않았습니다. 파일 변경·빌드·시험·네트워크·tossctl·docker 실행 0. ai-memory retrieval 스킬을 읽었으나 MCP 부재와 네트워크 금지로 조회하지 않았습니다. `impeccable: N/A — backend 설계 리뷰, UI 변경 없음`.

**「설계 결정에 닿는 P0/P1」: 예 — P1 2건.**
