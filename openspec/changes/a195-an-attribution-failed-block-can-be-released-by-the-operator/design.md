# a195 Design

## 1. 결정

| # | 결정 | 근거 |
|---|---|---|
| D1 | 해제 가능한 원인 = `QUANTITY_MISMATCH`, 그리고 **종목 범위**(symbol 비어 있지 않음)의 `ATTRIBUTION_FAILED`. 그 밖은 종전 거절 | 이번 사고의 원인이 이것 하나이고, 나머지 원인의 해제 근거는 아무도 정의하지 않았다(모르는 것을 풀지 않는다) |
| D2 | `ATTRIBUTION_FAILED` 의 해제 근거 = **명령 자신의 안정 재조회가 차이를 보지 않음** + 운영자 신원·note. 새 증거 종류·새 플래그·새 확인 문구를 만들지 않는다 | 이 차단의 목적은 「장부와 계좌가 그 종목에서 어긋났을 수 있다」는 불확실성 동안 진입을 막는 것이다. 명령은 이미 엔진 잠금 아래 재조회·비교를 하고 차이가 남으면 거절한다 — 그 일치가 불확실성의 해소 증거다. 사람 확인은 note 로 장부에 남는다 |
| D3 | 원인 판정은 새 파일의 새 함수 하나(예: `reconcileResolveRefusedCause(blocks) error`)로 빼고 validate 클로저는 그것을 부른다. 허용 원인 목록은 그 함수 한 곳 | 목록이 한 곳에 있어야 다음 원인을 더할 때 자리를 세지 않는다. 기존 본문 편집을 한 줄로 줄인다 |
| D4 | 해제 기록 형식 무변경 — `Tracker.Resolve` 를 고치지 않는다 | `ExpectCause: block.Cause` 가 이미 원인별로 맞춘다(proposal 표 4행). tasks 1.3 에서 AST 로 재확인 |
| D5 | 계좌 전체(symbol 빈) `ATTRIBUTION_FAILED` 는 거절 | 저장소에 그 모양을 쓰는 자리가 없다고 조사됐고(조사 인용 F6: 계좌 전체는 `IDENTIFIER_CONFLICT`), 없는 모양에 해제 규칙을 만들지 않는다. 거절 문구는 종전 문구 |

## 2. 거절하는 정상 입력 (fail-closed 가 무엇을 거부하는지)

- `IDENTIFIER_CONFLICT` 활성 행이 하나라도 있으면 명령 전체가 거절된다 — 같이 있던 `QUANTITY_MISMATCH`·`ATTRIBUTION_FAILED` 도 풀리지 않는다. 종전과 같은 동작이고(원인 하나라도 근거 없으면 전부 거절), 이 change 는 그것을 바꾸지 않는다.
- 재조회에 차이가 남아 있으면(다른 종목의 차이 포함) 종전대로 거절된다. 이 판정 함수는 건드리지 않는다.

통과시키는 새 입력: 활성 블록이 {`QUANTITY_MISMATCH`, 종목 `ATTRIBUTION_FAILED`} 로만 이루어진 모든 조합.

## 3. 시험 전략 (RED 먼저)

새 시험 파일 `cmd/tossctl/a195_reconcile_resolve_attribution_test.go`(실 원장 임시 경로, 가짜 수집기 — 기존 `engine_reconcile_test.go` 의 하네스 재사용):

1. 종목 `ATTRIBUTION_FAILED` 한 행 + 재조회 일치 → 해제됨: `released_at` 채워짐, `release_cause = OPERATOR`, 근거에 운영자 신원·note, 그 뒤 `Tracker.Restore` 가 블록 0.
2. 같은 행 + 재조회 차이 남음 → 거절, 행 무변(종전 판정이 여전히 앞에 선다).
3. `IDENTIFIER_CONFLICT` 한 행 → 종전 문구(`requires cause-specific recovery evidence`)로 거절, 행 무변.
4. `ATTRIBUTION_FAILED` + `IDENTIFIER_CONFLICT` → 거절, **두 행 다 무변**.
5. `QUANTITY_MISMATCH` + `ATTRIBUTION_FAILED` → 둘 다 해제.
6. 계좌 전체 `ATTRIBUTION_FAILED`(합성) → 거절.
7. 활성 블록 0 → 종전 문구 거절(회귀).

변이(전부 CAUGHT 여야 함, 변이마다 닿음 확인): 원인 판정 삭제(전부 통과) · 허용 목록에 `IDENTIFIER_CONFLICT` 추가 · 종목 범위 조건 삭제 · `ATTRIBUTION_FAILED` 를 목록에서 빼기(1·5 가 죽어야 함).

선택 시험: `./cmd/tossctl -run 'A195|ReconcileResolve'`, `./internal/reconcile -run 'Resolve|Restore'`. 전체 스위트는 종결 게이트에서만.

## 4. 운영 절차 (배포 뒤, 사람 승인)

1. 엔진 정지 창에서(두 시장이 닫힌 시간, 또는 엔진이 이미 꺼져 있을 때) `tossctl engine reconcile-resolve --confirm …` 실행 전, 거절 조건을 장부로 예측한다 — 활성 행의 원인 목록과 그 종목의 계좌·장부 수량.
2. 실행 뒤 `reconcile_states` 의 그 행 `released_at`·`release_cause` 와 다음 `reconcile.clean` 의 `blocked` 값을 확인한다.

## 5. 후속 후보 (이 change 밖)

1. 경합 자체의 수리 — 엔진 주문의 매도 체결이 「방금 수렴 조정으로 닫힌 인스턴스」에 닿는 경우를 차단 없이 처리(조사의 β: 싱크 쪽 인정, α: 수렴 보류, γ: 증명 뒤 자동 해제). 같은 경합의 반대 순서(체결 먼저)가 `QUANTITY_MISMATCH` 고착을 만드는지는 추론이고 재현 전이다.
2. 그 경합에서 엔진 손절이 `COMPLETED`·trade outcome 없이 「외부 청산」으로 기록되는 것(조사 인용 F15).
3. 차단 진입 순간의 알림과 차단 때문에 편입되지 못한 보유의 알림 — a194.
