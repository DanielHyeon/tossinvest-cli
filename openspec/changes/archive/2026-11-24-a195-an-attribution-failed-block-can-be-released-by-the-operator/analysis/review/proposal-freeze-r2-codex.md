판정: **P1 1건·P2 1건이 남습니다.** r1의 외부 보유 허용 자체를 P0로 다시 제기하지 않습니다. 다만 그 결정을 뒷받침하는 “다음 대사에서 손절이 걸린다”는 주장은 조건부로 수정해야 합니다.

기준: HEAD `e0b8b461a3fe97405132b00f0d04118a3a5e89bd`와 작업 디렉터리의 수정 문서. 정적 검토이며 시험·변이 실행 결과는 아닙니다.

## r1 처분 재검토

| 항목 | 판정 | 근거 |
|---|---|---|
| **R1** | **기각 취지 동의, 근거 (2)는 부분 반박** | 기존 QM 해제도 `BlocksEntry()`만 검사한다는 (1)은 맞습니다(`cmd/tossctl/engine_reconcile.go:197–204`, `internal/reconcile/compare.go:312–319`). 외부 보유 접기→편입 판정 경로도 실제 존재합니다(`internal/app/engine/reconcileloop.go:438–470`). 그러나 편입·보호 성공은 보장되지 않습니다. 아래 새 발견 1 참조. (3)은 자동 해제와 운영자 해제의 정책을 구분한 결정으로 받아들입니다. |
| **R2** | **기각 동의** | 세 거절 유형은 `internal/position/position.go:98–108`에 정의돼 있고, OVERSELL 외에는 AF로 매핑됩니다(`internal/journal/position_projection.go:428–432`). 세 유형을 모두 받는다는 범위가 이제 명시됐습니다. 비교가 체결 귀속을 복구한다는 뜻은 아니지만, 문구 파싱 없이 사람 확인으로 해제한다는 정책 결정 자체를 코드 오류로 반박할 근거는 없습니다. |
| **R3** | **수용 방식 동의, 증명 범위 제한** | 실제 헬퍼 시험과 production `runEngineReconcileResolve` 구동은 기존 스텁 문제를 줄입니다. 다만 production **builder가 반환한 validate**를 실행하는 시험은 여전히 아닙니다. 기존 TLS 시험도 builder 구성 후 validate를 호출하지 않습니다(`cmd/tossctl/a166_reconcile_safety_test.go:91–119`). 이 연결부를 D6의 구조 검사로 보완한다는 처분으로 이해해야 합니다. |
| **R4** | **수용 동의** | 수정 design §3에 순서 양방향, 실제 비교기로 만든 외부 보유 입력, 반복 Observe 회귀가 추가됐습니다. 현재 Observe는 원인과 release policy를 모두 검사합니다(`internal/reconcile/mismatch.go:667–672`). 여섯 변이의 검출 계획은 성립하지만 **CAUGHT는 아직 미확인**입니다. |
| **R5** | **수용 동의** | 해당 builder 본문은 `shapes.json:7`의 고정 대상이며, census는 AST 문자열이 다르면 실패합니다(`internal/accountbudgetruntime/a166_wiring_census_test.go:209–216`). 따라서 기준 corpus를 고정한 상태의 “헬퍼 결과 무시” 변이는 구조적으로 검출됩니다. 이것을 의미적 실행 검증과 구분해 기록한다는 처분도 맞습니다. |
| **R6** | **수용 취지 동의, 구현 범위 모순 잔존** | 도움말·성공 문구 수정은 필요합니다(`cmd/tossctl/engine_reconcile.go:82–89,212`). §4의 편입·보호 별도 확인도 적절합니다. 다만 “실제로 해제한 원인 표시”와 “출력 문자열만 변경”은 현재 인터페이스에서 양립하지 않습니다. 아래 새 발견 2 참조. |

## 새 발견

### [P1] D2가 조건부 편입 경로를 다음 주기 손절 보장으로 서술한다

- **계약 근거:** [design.md:8](/home/daniel/tossos-work/a195-impl/openspec/changes/a195-an-attribution-failed-block-can-be-released-by-the-operator/design.md:8)의 근거 (2).
- **코드로 맞는 부분:** `RunOnce`는 외부 보유를 접고, tracker 갱신·관측 후 편입을 판단합니다(`internal/app/engine/reconcileloop.go:438–470`). 편입 성공 경로는 synthetic stop을 만들고 `AdoptPosition`·`OpenAdoptedExitState`를 호출합니다(`internal/app/engine/adoption.go:368–394`).
- **반례:** 제외 종목이거나 편입 OFF이면서 include 대상도 아니면 미편입으로 남습니다([adoption.go:163](/home/daniel/tossos-work/a195-impl/internal/app/engine/adoption.go:163)). 가격 누락·노후화는 편입을 연기합니다(`adoption.go:230–246`). 편입 기록이 성공해도 exit state 생성은 실패할 수 있습니다(`adoption.go:394–400`). 미편입 포지션은 exit observer가 보호 대상에서 제외합니다([exitloop.go:592](/home/daniel/tossos-work/a195-impl/internal/app/engine/exitloop.go:592)).
- **영향:** **차단 해제 성공·`blocked=0`과 손절 보호 복원 사이에 공백이 존재합니다.** §4:44는 이미 이를 인정하므로 D2의 단정과 모순됩니다. 이번 사고 환경에서 편입 조건이 실제 충족됐는지는 운영 자료를 읽지 않아 확인하지 않았습니다.
- **제안:** 근거를 “해제는 정상 편입 경로를 다시 열며, 기존 편입 조건과 보호 상태 생성이 성공하면 손절 관리 대상이 된다”로 수정하십시오. §4 확인 항목에 대상 종목의 편입 허용·제외 여부와 실제 exit state를 명시하면 됩니다. 새 해제 가드나 토글 변경을 요구하는 발견은 아닙니다.

### [P2] 실제 해제 원인을 출력할 데이터 경로가 편집 범위에서 빠졌다

- **계약 근거:** [design.md:35](/home/daniel/tossos-work/a195-impl/openspec/changes/a195-an-attribution-failed-block-can-be-released-by-the-operator/design.md:35)는 성공 출력이 “해제한 원인을 말함”을 요구합니다. 반면 proposal:49와 기존 FLM은 `runEngineReconcileResolve`의 출력 문장만 변경하도록 제한합니다.
- **코드 근거:** [engine_reconcile.go:43](/home/daniel/tossos-work/a195-impl/cmd/tossctl/engine_reconcile.go:43)의 runtime에는 블록·원인 목록이 없고, `validate`와 `resolve` 모두 `error`만 반환합니다. tracker는 builder 내부 지역 변수입니다(`engine_reconcile.go:277`). 해제 뒤 tracker의 블록도 비워집니다(`internal/reconcile/mismatch.go:1391`).
- **영향:** 출력 문자열만 바꾸면 실제로 해제한 원인을 알 수 없습니다. 허용 원인 두 개를 고정 출력하면 QM만 해제한 경우에도 AF를 해제했다고 오인시킬 수 있습니다. 요구대로 구현하려면 현재 선언보다 편집 범위가 넓어집니다.
- **제안:** 둘 중 하나로 계약을 맞추십시오. 원인 중립 성공 문장만 요구하거나, 해제 전 원인 목록을 성공 출력까지 전달하는 최소 데이터 경로를 편집 범위·FLM·시험에 포함하십시오. 후자를 택하면 QM만·AF만·혼합 출력의 정확성을 각각 확인해야 합니다.

## 확인했으나 문제 없음

- 외부 보유·외부 주문을 해제 거절 사유에서 제외한다는 수정 spec은 현재 `BlocksEntry()` 의미와 일치합니다. 평균단가는 `Comparer.Compare`의 비교 대상 자체가 아닙니다.
- 현재 자동 해제 경로는 AF를 풀지 않습니다. Observe의 원인 검사와 a192 종결 증명의 QM 제한을 확인했습니다(`mismatch.go:667–672`, `order_terminal_release.go:182–186`).
- `Tracker.Resolve`는 실제 원인을 `ExpectCause`로 전달합니다. `ReleaseReconciles`는 모든 대상의 활성 상태·원인을 확인한 뒤 단일 트랜잭션으로 해제합니다(`mismatch.go:1371–1380`, `reconcile_states.go:372–415`).
- §5의 `ScopeMarket` 누락 한계는 실제 코드와 맞습니다. 이번 변경의 새 결함으로 중복 집계하지 않았습니다.
- 해제는 체결 귀속 재처리나 기존 손절 기준선 복원이 아닙니다. §4에 이를 구분한 문장은 타당합니다.

파일 생성·수정, 빌드·시험·변이·네트워크·tossctl·docker 실행 **0회**. 기존 AST/FLM은 일부 참고했으며 재생성·해시 검증하지 않았습니다. ai-memory MCP 미제공, CLI 조회는 네트워크 금지에 따라 생략했습니다. 운영 사고 사실·실제 보호 복원·변이 CAUGHT·전체 게이트 PASS는 이번 검토로 입증하지 않았습니다.