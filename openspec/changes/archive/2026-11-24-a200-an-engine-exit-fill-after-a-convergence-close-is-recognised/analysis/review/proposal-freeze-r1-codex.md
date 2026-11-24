**P0 확정 없음. P1 2건, P2 3건.** 현재 HEAD `b62dd4cfd63e94467c7d416054fca02cba14d3b1` 기준입니다. 설계 반례와 실제 운영에서 확인된 현상을 구분했습니다.

**[P1] D2a의 delta 합은 실제 누적 체결보다 작아질 수 있다**

- **근거:** [design.md:12](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/design.md:12). `RecordFill`은 float64로 누적 차를 계산하고 작은 delta를 0으로 바꿉니다([fills.go:412](/home/daniel/tossos-work/a200-impl/internal/journal/fills.go:412), [fills.go:1924](/home/daniel/tossos-work/a200-impl/internal/journal/fills.go:1924)). 이후 스냅숏은 전진하지만 `fill_events`는 양의 delta만 저장합니다(:437, :464).
- **재현 방법:** 감소분 `1,000,000,002`로 닫힌 인스턴스에 동일 pending intent의 Working 누적 관측을 `1,000,000,000 → 1,000,000,001 → 1,000,000,003` 순으로 적용합니다. 저장 delta는 `1,000,000,000 → 0 → 2`, 합은 `1,000,000,002`입니다. 마지막 실제 누적은 감소분보다 1 크지만 D2a와 `(라)`를 통과합니다.
- **왜 문제인가:** 0 관측에서 인정 함수를 호출하지 않는다는 사실은 안전 증명이 아닙니다. 그 관측이 다음 delta의 기준을 바꿉니다. decimal 합산은 이미 누락된 양을 복구하지 못하고, `합 ≥ 이번 delta`도 통과합니다. `claim-check.md:26`의 근거가 불충분합니다.
- **제안:** delta 합을 누적의 보수적 상한으로 취급하지 마십시오. 최소한 현재 주문의 원문 누적과 합의 불일치를 거절하고, 주문별 누적 증거와 delta 합의 정합성을 검증하도록 D2a를 보강해야 합니다.
- **확인 수준:** 원문 코드와 동일 산술의 읽기 전용 계산으로 확인. 위 대수량은 합성 경계값이며 운영 발생·브로커 허용 범위는 미확인입니다.

**[P1] D2는 주문 소유를 증명하지만, 그 체결이 해당 조정의 원인임을 증명하지 않는다**

- **근거:** [design.md:10–12](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/design.md:10). `resolveFillOrigin`은 주문과 intent의 소유·범위를 확인합니다([position_projection.go:139](/home/daniel/tossos-work/a200-impl/internal/journal/position_projection.go:139)). `currentInstance`는 최신 **원장** 인스턴스를 선택하고(:271), 조정은 pending intent를 지우지 않습니다([position_adjustments.go:420](/home/daniel/tossos-work/a200-impl/internal/journal/position_adjustments.go:420)).
- **반례 구성:** 엔진 매도 intent가 pending인 상태에서 외부 처분으로 기존 보유 31이 없어져 `UNKNOWN 31→0` 조정이 생깁니다. 이후 사람이 재매수한 31이 아직 원장에 접히기 전에 기존 엔진 주문의 매도 31이 체결됩니다. 원장에는 여전히 옛 CLOSED 인스턴스만 있고, pending 일치·누적 31 이하까지 전부 통과합니다. 그러나 이 매도는 **조정 이후 취득한 보유의 처분**입니다.
- **왜 문제인가:** 문서가 범위 밖으로 분리한 재매수 사례는 새 인스턴스가 **이미 생성된 경우**입니다(:40). 생성 전 창은 인정 판정에 도달합니다. 트랜잭션 직렬화는 조정 커밋이 먼저였다는 것만 증명하며, 브로커 체결이 조정의 계좌 스냅숏에 포함됐다는 증명이 아닙니다.
- **제안:** “조정의 설명으로 입증됨”과 “수량 상한 안의 소유 확인 체결”을 구분해야 합니다. 조정 스냅숏과 실행의 인과관계를 입증할 증거가 없으면 인정 거절하거나, 이 모호성을 허용하는 별도 안전 결정을 명시해야 합니다.
- **확인 수준:** D2가 구분하지 못하는 설계 반례입니다. 실제 브로커에서 해당 외부 매도·재매수·미결 주문 순서가 허용되는지는 미확인입니다. 운영 발생으로 주장하지 않습니다.

**[P2] “조정을 사이에 둔 부분 체결은 거절”은 중간 체결에서 거짓이다**

- **근거:** [design.md:29·38](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/design.md:29), [claim-check.md:32](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/analysis/claim-check.md:32).
- **재현 방법:** 기존 재현의 `31 → 선행 체결 10 → 조정 21→0`까지 유지합니다. 이후 체결을 한 번에 21 보내지 말고 `1, 20`으로 나눕니다. 첫 체결의 누적은 `11 ≤ 21`이므로 인정됩니다. 마지막 체결에서야 `31 > 21`로 거절됩니다.
- **왜 문제인가:** 기존 시험은 마지막 전량 관측만 검증했습니다. 그 결과로 모든 straddling 체결이 거절된다고 일반화했습니다. ADDED 시나리오도 같은 잘못된 단정을 합니다.
- **제안:** 실제 의도가 누적 상한이면 “상한을 넘는 체결부터 거절”로 문서·시나리오를 수정하고 중간 인정 시험을 추가하십시오. 조정 전 체결 존재 자체를 배제하려면 별도 판정 근거가 필요합니다.

**[P2] 재기동 뒤에는 인정되어도 다음 일치에서 QM 차단이 풀리지 않는다**

- **근거:** [reconciliation delta:4](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/specs/reconciliation/spec.md:4)는 다음 일치 해제를 SHALL로 보장합니다. 그러나 `Tracker.Restore`는 조정 크레딧과 종결 증명 상태를 버립니다([mismatch.go:998](/home/daniel/tossos-work/a200-impl/internal/reconcile/mismatch.go:998)). 크레딧 없는 QM은 일치해도 `AwaitingAdjustment`입니다(:675–686).
- **재현 방법:** W6에서 QM 영속화 후 다음 일치 전에 Tracker를 재생성하고 `Restore`합니다. 인정 체결을 적용한 뒤 깨끗한 비교를 반복합니다. CLOSED/0·AF 없음은 유지되지만 QM은 남습니다.
- **왜 문제인가:** `claim-check` C-D6의 “원장 등가”만으로 메모리 크레딧 등가를 증명할 수 없습니다. 진입·편입 차단도 남습니다. a194 역시 이 상황을 이미 설명합니다([reconcile_block_report.go:173](/home/daniel/tossos-work/a200-impl/internal/app/engine/reconcile_block_report.go:173)).
- **제안:** 재기동 없음·유효한 크레딧 보존·비교 시각 조건을 해제 보장에 명시하고, 재기동 사례를 잔여 제한과 시험에 추가하십시오. 기존 복구 정책 변경까지 요구하는 발견은 아닙니다.

**[P2] 인정 실패가 항상 영구 AF 차단을 만든다는 delta는 W6 경로와 충돌한다**

- **근거:** [reconciliation delta:10–12](/home/daniel/tossos-work/a200-impl/openspec/changes/a200-an-engine-exit-fill-after-a-convergence-close-is-recognised/specs/reconciliation/spec.md:10)는 조건 불충족 시 AF 생성·자동 해제 불가를 보장합니다. 실제 `enterReconcileScopeInTx`는 활성 종목 행이 있으면 즉시 반환합니다([position_projection.go:401](/home/daniel/tossos-work/a200-impl/internal/journal/position_projection.go:401)). design D4는 이 예외를 알고 있습니다.
- **재현 방법:** W6처럼 QM을 먼저 저장한 뒤 pending과 다른 intent의 매도 등 D2 불충족 체결을 적용합니다. AF는 생성되지 않습니다. 이후 유효한 크레딧과 일치 비교가 있으면 기존 QM이 해제됩니다.
- **왜 문제인가:** 구현 방침 “현재 동작 유지”와 delta의 SHALL이 양립하지 않습니다. 이 경로에서는 a194의 AF 운영자 전용 보고도 발생하지 않습니다.
- **제안:** “활성 행이 없을 때 AF 생성, 기존 활성 행 처리 규칙 유지”로 계약을 정확히 좁히십시오. 항상 AF를 남기려면 기존 함수 하나만 수정한다는 설계 범위를 다시 결정해야 합니다.

D3의 NULL 표지가 판정 소비자를 오작동시키는 반례는 확인하지 못했습니다. intent·decision·`COMPLETED` 필터는 표지를 제외하고, 평가 증거는 legacy 갈래로 읽힙니다. 콘솔·provenance에는 표시됩니다. 상태기계 MODIFIED 본문에서도 추가한 예외 외의 의도하지 않은 변경은 발견하지 못했습니다.

파일·Git 상태 변경과 네트워크 사용은 없었습니다. **Go 시험은 미실행**이며 기존 `w4.txt`·`cycles.txt`는 과거 실행 증거로만 사용했습니다. ai-memory 서버 조회·그래프 재생성도 수행하지 않았습니다. `impeccable: N/A — 백엔드 설계 읽기 전용 리뷰`.