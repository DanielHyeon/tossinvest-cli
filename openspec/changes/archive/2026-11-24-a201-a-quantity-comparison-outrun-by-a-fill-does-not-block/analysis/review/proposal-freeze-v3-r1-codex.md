codex
**판정: freeze 불가 — credit 반박을 건너뛰는 P1 1건과 원인 표지·명세 충돌 P2 2건이 있습니다.**

1. **P1 — 외부 포지션 관측에서 authority 조회가 실패하면, 반박됐어야 할 체결 credit이 살아남습니다. 설계 결정에 닿음: 예(V4·반박 분기 무편집).**

   재현 가능한 입력 순서:

   - A에서 정상 W2가 발생합니다. A의 QM 차단과 체결 credit이 생기고, 실제 계좌·로컬은 모두 0입니다.
   - 다른 종목 B의 동일 missing-order가 연속 임계에 도달하지만 영구 승격 쓰기가 실패합니다. `pendingPermanent=true`가 남습니다. 이 비교에서 A는 일치하므로 A credit은 보존됩니다.
   - 다음 비교는 B 주문이 정상 매칭되고, **A의 외부 보유 5주만** 보고합니다. `diff.AsOf`는 A credit보다 엄격히 나중입니다. Ingest는 A를 외부 포지션으로 접습니다.
   - RunOnce의 `Refresh`는 성공시키고, **Observe 내부의 다음 `ActiveReconcileStates` 호출만 실패**시킵니다.
   - 이후 `Refresh`가 durable permanent 행 부재를 확인합니다. 다음 비교는 접힌 A의 로컬 5와 계좌 5가 일치합니다. 새 체결 credit·수렴 credit은 없습니다.

   `ExternalPos`는 `disputed[A]=true`지만 `BlocksEntry()`에는 포함되지 않습니다. 따라서 Observe는 clean 갈래로 들어가 **credit 반박 전에 authority 오류로 반환**합니다. 마지막 회계 루프도 실행되지 않습니다. 이후 Refresh는 credit을 지우지 않으므로 A가 `FILL_APPLIED`로 해제됩니다.

   근거: [외부 포지션의 분쟁 판정:576](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:576), [BlocksEntry:312](/home/daniel/tossos-work/a201-impl/internal/reconcile/compare.go:312), [조기 반환:659](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:659), [blocking 갈래에만 있는 조기 반박:716](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:716), [끝 반박 루프:826](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:826), [Refresh:1044](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:1044).

   **V8의 외부 거래 상쇄와 다릅니다.** credit 부여 당시에는 정상 W2였고, 이후 명시적으로 관측된 A 분쟁이 credit을 소멸시켜야 합니다. 외부 거래로 상쇄하지 않아도 Ingest 이후 일치로 해제됩니다.

   기존 조정 credit에도 잠재한 구멍입니다. 그러나 “기존 코드 공유로 체결 credit의 반박이 보장된다”는 [V4:14](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/design.md:14)는 성립하지 않습니다. **clean 갈래의 오류 반환 전 반박을 보장하도록 편집 표면·검증 계획을 수정해야 합니다.** Observe 오류 때 *새 credit을 부여하지 않는 것*은 기존 credit의 생존을 막지 못합니다.

2. **P2 — as-of 일치만으로 원인을 구분하면, 같은 as-of로 덮어쓴 조정 credit이 체결 원인으로 기록됩니다. 설계 결정에 닿음: 예(V4·V5).**

   재현 입력:

   - A에 `adjusted[A]=T`, `filled[A]=T`인 체결 credit이 있습니다.
   - 이후 실제 조정이 적용되면서 `AdjustmentApplied(T, A)`가 호출됩니다.
   - `T`보다 나중의 깨끗한 비교를 Observe에 전달합니다.

   `AdjustmentApplied`는 `adjusted`만 덮어씁니다. 따라서 두 값은 계속 같고, 제안된 표지는 **실제 최신 근거가 조정인데도 `FILL_APPLIED`**를 선택합니다. 부여 메서드의 정리도 값이 같으므로 제거하지 않습니다.

   근거: [AdjustmentApplied:484](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:484), [곁 맵 정리·대입:227](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/analysis/v3-probe/probe.diff:227), [V5의 최신 조정 우선 주장:15](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/design.md:15).

   이는 정상 주기 간격에서 매번 발생한다는 주장이 아닙니다. **동일 as-of 입력 재사용·벽시계 역행 경계**의 반례입니다. 현재 comparer는 시각을 초 단위로 기록하며, credit API는 단조 증가나 고유성을 요구하지 않습니다([compare.go:361](/home/daniel/tossos-work/a201-impl/internal/reconcile/compare.go:361)).

   결과는 감사 원인 오표지이며, broker-zero에서도 원래 조정 증명으로 심사할 행을 원인 허용 목록에서 거절하게 됩니다. **credit 작성 종류를 시각 동등성과 분리하거나, 동일 as-of 덮어쓰기의 우선순위를 명시적으로 처리해야 합니다.**

3. **P2 — MODIFIED에 남은 시나리오가 v3의 정상 해제를 금지합니다. 설계 결정에 닿음: 예(정본 delta).**

   입력은 설계가 이미 채택한 **“앞 조정 credit 위 W2”**입니다.

   - 이전 조정 credit이 있습니다.
   - 나중 W2 비교가 같은 종목의 불일치를 관측해 옛 credit을 반박합니다.
   - Observe 뒤 체결 credit을 부여합니다.
   - 다음 깨끗한 비교에서 새 조정 없이 `FILL_APPLIED`로 해제합니다.

   ADDED는 이 결과를 요구하지만, MODIFIED의 「조정 이후의 재조회가 여전히 불일치한다」는 여전히 **“해제는 새 조정과 그 뒤의 일치를 다시 요구한다”**고 규정합니다.

   근거: [ADDED:11](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:11), [충돌하는 MODIFIED:250](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:250), [정본 원문:335](/home/daniel/tossos-work/a201-impl/openspec/specs/reconciliation/spec.md:335).

   `generate.py`의 역치환·시나리오 이름 보존 단언은 **치환 정확성**만 확인합니다. 새 규칙과 보존한 THEN의 의미 충돌은 검증하지 않습니다. 해당 THEN을 조정 경로로 한정하거나 새 체결 credit 경로를 포함해야 합니다.

나머지 요청 항목의 판정:

- **R1-1 원형·R1-2 lineage 삭제·R2-1 관측 생략·R2-2 교대 성공 처리는 v3 구조에서 재개방되지 않습니다.** diff를 삭제하지 않고 Observe·편입 판정을 실행합니다. 이는 정적 판정이며 해당 입력을 재실행하지 않았습니다.
- **새 credit의 부여 주기에는 그 credit이 Observe 결과를 바꾸지 않습니다.** 성공 반환 뒤 부여하기 때문입니다. 다만 이전 주기의 체결 credit까지 포함해 “모든 Observe 결과 무변”이라고 일반화할 수는 없습니다.
- **V5의 a192 순서 논증은 정상 결속 수명에서 성립합니다.** 부여 조건의 수량 불일치가 결속을 먼저 `tainted`로 만듭니다. `fail_closed` 결속 행을 체결 경로가 푸는 것은 설계가 명시적으로 확장한 동작이며, 기존 a192 종결 증명이 통과한 것으로 보고하면 안 됩니다.
- 종목이 비교에서 사라진 경우·외부 주문만 있는 경우는 현재 분쟁 집합에 들지 않습니다. 다른 종목 때문에 해제를 기다리는 동안 credit이 남는 것은 채택한 조정 credit 규칙과 같습니다. 이를 독립 결함으로 세지 않았습니다.

HEAD `7793590963af7c67b23b97fb7e0a9acec4e51596`의 지정 문서·증거와 관련 소스를 정적으로 검토했습니다. 기존 AST/FLM을 참조했으며 재추출·freshness 검증은 하지 않았습니다. 위 반례는 **미실행 입력 추론**입니다. 과거 탐침 PASS를 이번 실행 증거로 사용하지 않았습니다. 파일 변경·빌드·시험·네트워크·tossctl·docker 실행 0. ai-memory MCP 부재로 추가 조회 미실행. `impeccable: N/A — backend 설계 리뷰, UI 변경 없음`.

**「설계 결정에 닿는 P0/P1」: 예 — P1 1건.**
