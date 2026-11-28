codex
**판정: 조건부 — 설계 수준 P0/P1은 발견하지 못했으나, 명세 정합성 P2 2건과 실행계획 P3 1건을 정리해야 합니다.**

1. **P2 — V4b 시나리오가 같은 시각·과거 비교까지 credit 소멸을 요구합니다.**

   [delta spec:36](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:36)은 외부 포지션 관측과 권위 확인 실패만으로 credit 소멸을 단정합니다. **비교가 credit보다 엄격히 나중이라는 조건이 없습니다.**

   재현 입력: `adjusted[A]=T`, `diff.AsOf=T`, `ExternalPos={A}`, blocking 항목 없음, pending permanent 권위 조회 실패. V4b가 호출할 [refuteUsableAdjustmentCredits:525](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:525)는 `creditUsableBy(T,T)=false`라 credit을 보존합니다. 반면 새 시나리오는 소멸을 요구합니다. 이는 같은 비교의 소멸을 금지하는 [delta spec:230](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:230)과도 충돌합니다.

   **처분:** WHEN에 엄격히 나중인 유효 as-of 조건을 추가하십시오. MODIFIED :51의 추가 문장에도 동일 조건을 명시하면 해석이 닫힙니다. 설계 결정 변경은 불필요합니다.

2. **P2 — 수용한 V5b 원인 이름 예외가 규범 시나리오에는 반영되지 않았습니다.**

   [design:73](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/design.md:73)은 동일 as-of 덮어쓰기에서 실제 조정 뒤에도 `FILL_APPLIED`가 기록될 수 있음을 수용합니다. 그러나 [delta spec:81](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:81)과 [:256](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/specs/reconciliation/spec.md:256)은 조정 뒤 일치의 원인을 여전히 `ADJUSTMENT_APPLIED`로 단정합니다.

   재현 입력: `adjusted[A]=filled[A]=T` → 실제 조정 후 `AdjustmentApplied(T,A)` → `T'>T`의 깨끗한 비교. [AdjustmentApplied:495](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:495)는 `adjusted`만 덮어쓰므로 V5b는 `FILL_APPLIED`를 선택하지만 두 시나리오는 다른 원인을 요구합니다.

   **처분:** Manager가 수용한 등식 기반 원인 선택과 동일 문자열 예외를 delta에도 반영하십시오. **1회차 구현 대안을 재요구하는 발견이 아니라, 확정한 결정과 규범의 불일치입니다.** 해제 자체의 신규 안전 결함은 확인하지 못했습니다.

3. **P3 — tasks와 편집 표면이 1회차 반영 전 상태입니다.**

   [tasks:23](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/tasks.md:23)은 RED를 §3의 1~11로 한정하고, GREEN은 `Observe` 1문장만 적으며, 변이는 §3-12를 가리킵니다. 현재 design은 **12=V4b, 13=V5, 14=V5b, 15=변이**입니다. [edit-surface:10](/home/daniel/tossos-work/a201-impl/openspec/changes/a201-a-quantity-comparison-outrun-by-a-fill-does-not-block/analysis/edit-surface-v3.md:10)에도 V4b 호출이 빠져 있습니다.

   재현 가능한 누락 경로: 작업자가 tasks의 명시 범위만 수행하면 이번 수정의 핵심 회귀시험과 V4b 구현을 빠뜨릴 수 있습니다.

   **처분:** RED 1~14, 변이 15, Observe 두 편집 자리로 정렬하십시오. 설계 결정 변경은 불필요합니다.

나머지 검토 결과:

- **V4b:** `disputed`는 원래 비교의 수량·missing-order·외부 포지션 집합이며, 현재 반박 규칙과 일치합니다. 다른 종목·같거나 이른 시각의 credit은 삭제하지 않습니다. 다른 authority 오류 반환은 기존 :716 반박을 거치고, persist 오류는 Observe 끝 회계 루프를 건너뛰지 않습니다. 추가적인 반박 누락이나 과잉 반박 반례는 확인하지 못했습니다.
- **V5b:** 현재 생산 코드의 `adjusted` 쓰기·삭제·초기화 자리를 대조했습니다. 공시한 동일 문자열 재사용 외의 원인 오표지는 발견하지 못했습니다. `Restore`·`Resolve` 뒤 남은 `filled`만으로는 해제되지 않습니다.
- **a192·이전 반례:** 부여 주기의 수량 관측이 결속을 먼저 taint합니다. V4b 예외와 이전 체결 credit의 후속 해제 효과를 제외하면, 신규 credit은 성공한 Observe 뒤에 생깁니다. R1-1·R1-2·R2-1·R2-2의 관측 삭제·생략 문제는 재개방되지 않았습니다.

HEAD `d3e56fbd37e2fd232d8742b3b90e7713c91540e7` 기준 정적 리뷰입니다. 위 입력은 미실행 추론이며, 기존 탐침에는 V4b가 없어 그 PASS를 수정 후 검증으로 인정하지 않았습니다. 파일 변경·빌드·시험·네트워크·tossctl·docker 실행 0. 그래프/AST 재추출·freshness 검증과 ai-memory 서버 조회는 미실행입니다. `impeccable: N/A — backend/documentation-only 설계 리뷰, UI 변경 없음`.

**「설계 결정에 닿는 P0/P1」: 아니오.**
