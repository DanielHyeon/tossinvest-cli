**P0 확인 없음. P1 2건, P2 2건.** HEAD `b15a1f489304212a4e736bb13877882c48257387`와 현재 작업 트리 원문 기준입니다. 1회차에서 수용한 인과 미증명·범위 밖 사실은 재등록하지 않았습니다.

**[P1] D2a가 읽는 `fill_snapshots`에는 해당 주문의 최신 누적이 없을 수 있다**

- **근거:** [design.md:12](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/design.md:12), [fills.go:730](/home/daniel/tossos-work/a200-impl/internal/journal/fills.go:730), [fills.go:775](/home/daniel/tossos-work/a200-impl/internal/journal/fills.go:775).
- **반례:** 다른 거래일의 주문 번호 P가 호환 표 `fill_snapshots`에 누적 0으로 남아 있습니다. 오늘 같은 번호 P의 엔진 매도 10은 `scoped_fill_snapshots`에 저장되지만, 범위가 달라 호환 행은 갱신되지 않습니다(:795–799). 보유 31에서 부모 체결 10을 반영한 뒤 수렴이 21→0으로 닫고, 같은 intent의 정정 자식 C가 21 체결되면 D2a는 **P의 0 + C의 21 = 21**로 인정합니다. 실제 주문별 누적 합은 **10 + 21 = 31 > 21**입니다. 부모의 종결 관측은 아직 도착하지 않은 순서입니다.
- **왜 문제인가:** `fill_snapshots`는 `ON CONFLICT(order_id)`인 호환 표입니다. “그 번호의 여러 범위 행 중 최댓값”을 이 표에서 구할 수 없습니다. 범위 필터 제거만으로 과소계산이 사라진다는 주장이 틀립니다. [position-ledger delta:20](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/specs/position-ledger/spec.md:20)의 straddling 거절도 위반합니다.
- **제안:** `scoped_fill_snapshots`와 legacy 호환 행을 구분해 출처를 정하고, 모든 관련 행을 **decimal 비교**하여 주문 번호별 상한을 구하십시오. 위 범위 충돌 사례를 시험에 추가해야 합니다.
- **확인/추측:** 저장·미러 조건과 계산 결과는 코드로 확인했습니다. 운영에서 번호가 재사용된 사례를 확인한 것은 아닙니다.

**[P1] 최신 원문 스냅숏도 작은 감소로 덮어써져 과거 누적보다 작아질 수 있다**

- **근거:** [fills.go:682](/home/daniel/tossos-work/a200-impl/internal/journal/fills.go:682)는 `nearlyZero` 범위의 누적 감소를 거절하지 않습니다. [fills.go:413](/home/daniel/tossos-work/a200-impl/internal/journal/fills.go:413)는 delta를 0으로 만들지만, :437의 upsert는 실행되어 :742에서 작은 원문 값으로 덮어씁니다.
- **반례:** 감소분 31인 CLOSED 인스턴스에서 같은 pending intent의 부모 P가 Working 누적 `10`으로 관측된 뒤, `9.999999991`로 재관측됩니다. 감소 `0.000000009`는 허용 오차 `0.00000001` 안입니다. 이어 자식 C의 누적 `21.000000009`를 관측하면 새 출처의 합은 정확히 **31**입니다. 그러나 부모의 이미 관측된 누적 10을 유지하면 **31.000000009 > 31**입니다.
- **왜 문제인가:** 첫 발견의 표 선택을 고쳐도 남습니다. 현재 행들의 최댓값은 **관측 이력의 최댓값**이 아닙니다. 1회차의 “양의 delta 누락”과 다른 경로이며, D2a의 “남는 오차는 더 많이 세는 쪽뿐”을 반증합니다.
- **제안:** 덮어쓸 수 있는 최신 값만을 누적 상한으로 삼지 마십시오. 보존된 원문 누적 이력도 활용하는 보수적 판정 또는 증거가 부족하면 인정하지 않는 규칙이 필요합니다. 감소 재관측 시험을 상승·delta 0 시험과 별도로 추가하십시오.
- **확인/추측:** 코드 조건에 따른 합성 반례입니다. 해당 소수 정밀도의 브로커 허용 여부·운영 발생은 미확인입니다.

**[P2] 정정 계보 정상 시험은 부모의 종결 관측 순서를 빠뜨렸다**

- **근거:** [design.md:52](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/design.md:52)는 부모 10 + 자식 21을 모두 인정한다고 계획합니다. 그러나 `REPLACED`는 [derive.go:94](/home/daniel/tossos-work/a200-impl/internal/brokerstate/derive.go:94)에서 terminal이며, [ApplyExitFill:1009](/home/daniel/tossos-work/a200-impl/internal/journal/apply_hook.go:1009)는 successor 검사 없이 pending intent를 지웁니다.
- **반례:** 조정 31→0 → 부모 Working 10 인정 → 부모 REPLACED 누적 10·delta 0 관측 → 자식 21 체결. 부모의 종결 관측이 pending을 비우므로 자식은 합이 31이어도 D2(다)를 실패합니다.
- **왜 문제인가:** Project→Exit 순서는 **이번 체결 전** pending만 보장합니다. 정정 계보 전체가 끝날 때까지 보장하지 않습니다. `ApplyExitFill` 무편집으로 정정 순서 전반의 정상 인정을 약속할 수 없습니다.
- **제안:** 부모 종결 관측이 자식보다 앞서는 시험을 추가하고, 무편집 방침이면 이를 명시적인 거절 제한으로 기록하십시오. 모든 순서를 인정하려면 편집 범위와 pending 해소 결정을 다시 정해야 합니다.
- **확인/추측:** 코드상 경로는 확인했습니다. M28의 `PENDING_REPLACE` 자체를 terminal로 오인한 주장은 아닙니다. 이후 `REPLACED` 관측을 사용하는 사례이며 실계좌 순서는 미측정입니다.

**[P2] 매도 예약 측정은 보유 조회의 지연된 0 응답을 배제하지 않는다**

- **근거:** [design.md:25](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/design.md:25)는 “0으로 읽힘”에서 실제 예약 수량의 소멸을 추론합니다. 하지만 [snapshot.go:267](/home/daniel/tossos-work/a200-impl/internal/reconcile/snapshot.go:267)의 `AsOf`는 로컬 수집 시작 시각이고, 주문 목록과 보유는 별도 요청입니다(:271, :286). [converge.go:209](/home/daniel/tossos-work/a200-impl/internal/reconcile/converge.go:209)는 비교에서 받은 수량으로 조정합니다.
- **반례 가정:** 매수 체결은 원장에 반영됐지만 보유 조회가 매수 전 0을 반환하고, 주문 서비스는 실제 보유에 대한 엔진 지정가 매도를 접수합니다. 그 0으로 조정한 뒤 매도의 부분 체결이 오면, 예약 규칙을 어기거나 외부 처분·재매수를 하지 않아도 D2를 통과할 수 있습니다.
- **왜 문제인가:** M29·M33은 매도가능수량 예약을 측정했으며, 보유 조회와 주문 서비스 사이의 일관성은 측정하지 않았습니다. “인과를 수용한다”는 결정과 별개로, 그 결정을 뒷받침하는 논증에 추가 가정이 빠졌습니다. 이후 대조가 복구 가능성을 주더라도 현재 수량이 정확하다는 증명은 아닙니다.
- **제안:** 보유 조회의 지연·누락 및 주문 서비스와의 일관성도 미측정 가정으로 공시하십시오. 기존 수용 결정을 재개하라는 요구는 아닙니다.
- **확인/추측:** 수집·조정 구조는 확인했습니다. 브로커가 실제로 이런 지연 응답을 반환한다는 주장은 **추측이며 미측정**입니다.

그 밖에 D3 표지·D6 해제 조건·상태기계 MODIFIED에서 새 충돌은 확인하지 못했습니다. 비decimal 원문은 “0으로 대체해 합산해도 된다”는 근거가 없으며, 파싱 실패 처리를 구현 시험으로 고정해야 합니다.

파일·Git 변경, 네트워크, 시험 실행은 하지 않았습니다. 기존 FLM/BTM과 원문 코드를 대조했으며 그래프·AST는 재생성하지 않았습니다. ai-memory MCP는 없었고 서버 조회는 네트워크 금지에 따라 생략했습니다. `impeccable: N/A — backend/documentation-only 읽기 전용 리뷰`.