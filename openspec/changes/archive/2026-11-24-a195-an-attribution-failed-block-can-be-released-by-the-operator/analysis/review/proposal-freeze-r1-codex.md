**판정: proposal-freeze 보류.** 현재 HEAD `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`와 작업 디렉터리의 계약·소스를 대조했습니다. 아래 첫째·셋째·넷째 항목은 기존 `analysis/code-context/evidence-reconciliation.md`에도 기록돼 있으나, 대상 계약에는 아직 해결되지 않았습니다.

[P0] 외부 보유가 남아 있어도 “재조회 일치”로 간주해 귀속 실패 차단을 해제한다

- **근거:** `design.md:8,16`과 delta `specs/reconciliation/spec.md:5`는 차이가 남으면 거절한다고 정합니다. 실제 명령은 `cmd/tossctl/engine_reconcile.go:197–204`에서 `diff.BlocksEntry()`만 검사합니다. 이는 `Quantities`·`MissingOrders`만 셉니다(`internal/reconcile/compare.go:312–319`).
- **반례:** 해당 종목의 인스턴스가 모두 CLOSED이면 `HeldBySymbol`은 종목을 제외합니다(`compare.go:154–156`). 브로커에 양수 보유가 남으면 비교기는 이를 수량 불일치가 아닌 `ExternalPos`로 분류합니다(`compare.go:410–417`). 제안한 원인 허용만 추가하면 이 상태에서도 해제가 가능합니다.
- **영향:** 귀속되지 않은 실제 보유가 남았는데 차단을 풀 수 있습니다. 기존 자동 해제는 바로 이 경우를 분쟁으로 취급합니다(`internal/reconcile/mismatch.go:585–593,689–699`). 실제 손실 발생은 확인하지 않았으나, 안전 불변식 위반 가능성은 코드 경로로 확인됩니다.
- **제안:** “진입 차단 diff 없음”과 “해제 대상 종목의 불확실성 해소”를 분리하십시오. 대상 종목 `ExternalPos`의 처리 정책을 먼저 결정하고, CLOSED→재매수→미편입 반례를 실제 비교기·원장 경로 시험에 포함해야 합니다.

[P1] 원인 이름 하나가 서로 다른 귀속 실패를 포괄하는데, 모두 같은 증거로 해제한다

- **근거:** `design.md:7–8`은 종목 `ATTRIBUTION_FAILED` 전체를 허용합니다. 그러나 `internal/journal/position_projection.go:428–432`는 OVERSELL 이외 거절을 모두 이 원인으로 매핑합니다. 여기에는 `UNATTRIBUTED_SELL`, `SELL_ON_CLOSED`, `ENTRY_WHILE_CLOSING`이 포함됩니다(`internal/position/position.go:98–108`).
- **영향:** 사고의 X59 복구를 넘어, 매도 귀속 부재와 청산 중 매수까지 해제 범위가 넓어집니다. 비교기는 비-CLOSED 인스턴스의 **종목별 합계**와 미체결 주문 신원을 비교합니다. 거절된 체결의 인스턴스 귀속·전이 복구는 확인하지 않습니다(`compare.go:148–178,193–237,436–490`). 외부 주문도 해제 거절 사유가 아니며, 평균단가는 이 `Compare` 호출 자체에서 비교하지 않습니다(`compare.go:353–358,468–473`).
- **제안:** 거절 유형별 필요한 확인을 계약에 명시하십시오. 수량 합계가 같아도 귀속 복구가 불명확한 경우를 거절할지, 운영자 note가 어떤 확인을 대신하는지 결정해야 합니다. 단순히 `Diff.Clean()`으로 바꾸는 것만으로 귀속 증명이 완성되지는 않습니다.

[P1] 기존 하네스 재사용만으로는 production 원인 가드의 변이 네 개를 검증할 수 없다

- **근거:** `design.md:22–32`는 기존 하네스와 네 변이를 지정합니다. 기존 단위 하네스는 비교·검증을 스텁으로 대체합니다(`cmd/tossctl/engine_reconcile_test.go:48–54`). 통합시험도 `validate`를 직접 복제하고 builder를 대체합니다(`engine_reconcile_test.go:365–387`). a166 TLS 시험은 production builder를 호출하지만 반환된 `validate`를 호출하지 않습니다(`cmd/tossctl/a166_reconcile_safety_test.go:87–119`).
- **영향:** 새 시험이 같은 구조를 복제하면 production 가드를 삭제하거나 허용 목록을 바꿔도 시험이 그대로 통과할 수 있습니다. helper 직접 시험만 추가해도 production 클로저가 helper를 호출하지 않는 변이는 놓칩니다.
- **제안:** 실제 `buildReconcileResolveRuntime`이 반환한 검증 클로저를 거치는 시험을 필수로 명시하십시오. 가짜 수집기를 연결하더라도 production 검증·비교·해제 연결은 유지해야 합니다.

네 변이의 **예상 검출 관계**는 다음과 같습니다. 실행하지 않았으므로 CAUGHT 판정은 아닙니다.

| 변이 | 검출해야 할 시나리오 |
|---|---|
| 원인 판정 삭제 | 3·4·6 |
| `IDENTIFIER_CONFLICT` 허용 | 3·4 |
| 종목 범위 조건 삭제 | 6 |
| `ATTRIBUTION_FAILED` 허용 제거 | 1·5 |

[P1] 시험 전략이 자동 해제 금지와 혼합 원인의 순서 의존 변이를 빠뜨린다

- **근거:** delta `specs/reconciliation/spec.md:19–21`에는 “반복 관측해도 귀속 실패가 해제되지 않음”이 있으나 `design.md:24–32`의 일곱 시험·네 변이에 대응 항목이 없습니다. 현재 자동 해제는 원인과 release policy 양쪽으로 제한됩니다(`mismatch.go:667–672,1408–1419`).
- **영향:** CLI 허용 목록 시험만으로는 자동 해제 금지를 입증하지 못합니다. 혼합 원인 시험도 배치를 하나만 쓰면 “첫 허용 원인을 만나면 즉시 성공 반환” 변이가 살아남을 수 있습니다. `Tracker.Blocks()`는 정렬된 순서로 반환합니다(`mismatch.go:1296–1301`).
- **제안:** 다음을 추가하십시오.
  - Restore 후 반복 Observe 및 조정 크레딧 존재 상태에서도 `ATTRIBUTION_FAILED` 유지.
  - 허용 원인이 거절 원인보다 앞·뒤에 오는 혼합 배치.
  - production helper 호출 제거·결과 무시, 비교 가드 우회 변이.
  - P0의 외부 보유와 외부 주문만 남은 입력을 실제 비교기로 생성하는 시험.

[P1] a166 census 정지 조건이 이미 충족됐지만 변경·검증 범위가 확정되지 않았다

- **근거:** `proposal.md:49`, `tasks.md:11`은 census에 있으면 Manager에게 복귀하도록 정합니다. 실제 `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json:7`에 `buildReconcileResolveRuntime` 본문 전체가 고정돼 있습니다. census는 포맷된 AST와 저장값이 다르면 실패합니다(`internal/accountbudgetruntime/a166_wiring_census_test.go:209–216`).
- **영향:** validate를 helper 호출로 바꾸는 편집도 이 검사를 깨뜨립니다. `design.md:34`의 선택 시험에는 해당 패키지가 없고, `TestA166EngineReconcileOptionsGuard`·`TestA166EngineReconcileLocalTLS`도 `A195|ReconcileResolve` 필터에 포함되지 않습니다.
- **제안:** freeze 전에 census fixture 갱신 범위와 독립 검토 책임을 결정하십시오. 관련 census·a166 안전 시험을 선택 검증에 포함하고, 구조 검사 실패를 의미적 변이 검출로 계산하지 마십시오.

[P2] 도움말·성공 문구와 운영 절차가 새 해제 범위를 반영하지 않는다

- **근거:** 도움말은 `active quantity-mismatch RECONCILE blocks`만 설명하고, 성공 출력도 `quantity-mismatch reconcile blocks released`입니다(`cmd/tossctl/engine_reconcile.go:82–89,212`). `proposal.md:48`의 편집 표면은 builder 중심이며 이 두 함수는 빠져 있습니다. `design.md:38–39`의 운영 절차도 원인·수량·해제 상태만 확인합니다.
- **영향:** 귀속 실패를 해제해도 수량 불일치 해제로 표시됩니다. 운영자는 해제 성공을 귀속 복구나 손절 보호 복원으로 오인할 수 있습니다.
- **제안:** 도움말·출력 수정과 해당 함수의 영향 분석을 범위에 넣으십시오. 운영 문서에는 해제의 증거 한계와, 재기동 후 편입·보호 상태 확인을 구분해 명시해야 합니다.

「확인했으나 문제 없음」

- 현재 production validate는 `QUANTITY_MISMATCH` 이외 원인을 거절합니다. 제안 기능은 아직 구현되지 않았습니다.
- 명령은 `--confirm`·신원·note를 검사하고, 엔진 잠금을 확보한 뒤 runtime 구성·스냅숏 수집을 수행합니다.
- 현재 Observe와 a192 종결 증명 경로는 `ATTRIBUTION_FAILED`를 자동 해제하지 않습니다. 영구 수량 차단도 operator-only로 유지됩니다.
- `Tracker.Resolve`는 실제 원인을 `ExpectCause`로 전달하고 `OPERATOR`·신원·note를 기록합니다. `ReleaseReconciles`는 모든 대상의 활성 상태·원인을 검사한 뒤 단일 트랜잭션으로 해제합니다.
- 현재 귀속 실패 INSERT는 `scope_market`을 쓰지 않습니다. 따라서 이 생산 경로에서는 `Resolve`의 빈 시장 범위와 `activeScopeWhere`의 `scope_market IS NULL`이 일치합니다.
- 공식 조회 인터페이스만 사용하는 구성과 브로커 주문 변경 호출 부재를 확인했습니다.

파일 생성·수정, 빌드·시험·변이·네트워크·tossctl·docker 실행은 **0회**입니다. 기존 AST·검토 기록은 참고했으나 재생성하거나 해시 검증하지 않았습니다. ai-memory 검색 도구는 미제공이며 네트워크 금지에 따라 CLI 조회도 생략했습니다. 판정은 현재 소스의 정적 대조 결과입니다.