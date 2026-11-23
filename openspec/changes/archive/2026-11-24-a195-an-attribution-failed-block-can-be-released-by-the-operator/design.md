# a195 Design

## 1. 결정

| # | 결정 | 근거 |
|---|---|---|
| D1 | 해제 가능한 원인 = `QUANTITY_MISMATCH`, 그리고 **종목 범위**(symbol 비어 있지 않음)의 `ATTRIBUTION_FAILED`. 그 밖은 종전 거절. `ATTRIBUTION_FAILED` 는 투영 거절 세 유형(`UNATTRIBUTED_SELL`·`SELL_ON_CLOSED`·`ENTRY_WHILE_CLOSING`, position.go:98-108 → position_projection.go:428-432)을 **모두** 받는다(Manager 결정 2026-11-23) | 이번 사고의 원인이 이것 하나이고, 나머지 원인의 해제 근거는 아무도 정의하지 않았다(모르는 것을 풀지 않는다). 세 유형은 모두 「투영이 거부한 체결」이고 복구 근거가 같다(엔진 잠금 아래 수량 일치 + 사람 확인). 유형은 evidence 문구에만 있고(position_projection.go:373), **문구로 판정을 가르지 않는다**. 유형별 증거가 필요해지면 원장에 유형 칸을 두는 별도 change 다. → r1 Codex P1(R2) 기각 |
| D2 | `ATTRIBUTION_FAILED` 의 해제 근거 = **명령의 기존 판정 그대로** — 엔진 잠금 아래 안정 재조회 비교가 진입을 막는 차이(수량 불일치, 원장에만 있는 주문 = `diff.BlocksEntry()` 참)를 보지 않음 + 운영자 신원·note. 해제 대상 종목에 외부 보유·외부 주문이 있어도 거절하지 않는다. 새 증거 종류·새 플래그·새 확인 문구·새 거절을 만들지 않는다(Manager 결정 2026-11-23) | (1) 수량 불일치 차단의 운영자 해제가 오늘 같은 조건이다 — 원인별로 조건을 갈라 두 판정을 만들지 않는다. (2) 해제는 그 외부 보유를 다음 대사 주기의 정상 경로(외부 포지션 접기 → 편입, reconcileloop.go:438-470)에 다시 연다 — 기존 편입 조건(제외 목록·편입 설정·가격 신선도, adoption.go:163·230-246)과 보호 상태 생성(adoption.go:394)이 성공하면 손절 관리 대상이 된다. 이것이 보호 방향이고, 거절하면 2026-11-23 사고처럼 그 보유가 보호 없이 남는다(r2 정정: 해제가 보호 복원을 보장하지는 않는다 — §4 에서 따로 확인). (3) `Observe` 의 자동 해제가 외부 보유를 분쟁으로 세는 것(`symbolsInDispute` mismatch.go:585-593)은 「조정 없이 우연히 일치」를 막는 자동 규칙이고, 운영자 해제는 사람 확인이 그 자리를 채운다. → r1 Codex P0(R1) 기각 |
| D3 | 원인 판정은 새 파일의 새 함수 하나(예: `reconcileResolveRefusedCause(blocks) error`)로 빼고 validate 클로저는 그것을 부른다. 허용 원인 목록은 그 함수 한 곳. 클로저 본문 변경은 원인 순회(B13~B14)를 그 호출 한 줄로 바꾸는 것뿐이고, 빈 블록 검사(B12)는 그대로 두거나 새 함수로 옮기는 쪽 중 본문 diff 가 작은 쪽을 고른다 | 목록이 한 곳에 있어야 다음 원인을 더할 때 자리를 세지 않는다. 기존 본문 편집을 최소로 줄인다 |
| D4 | 해제 기록 형식 무변경 — `Tracker.Resolve` 를 고치지 않는다 | `ExpectCause: firstNonEmpty(block.Cause, QUANTITY_MISMATCH)` 가 이미 원인별로 맞춘다(proposal 표 4행; AST: analysis/cited-ast/Tracker.Resolve.ast.json B4·B5) |
| D5 | 계좌 전체(symbol 빈) `ATTRIBUTION_FAILED` 는 거절 | 저장소에 그 모양을 쓰는 자리가 없고(HEAD grep: 빈 symbol 로 쓰는 6자리 전부 `IDENTIFIER_CONFLICT` — `analysis/code-context/evidence-reconciliation.md`), 없는 모양에 해제 규칙을 만들지 않는다. 거절 문구는 종전 문구 |
| D6 | a166 census 갱신 — `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 의 키 `cmd/tossctl/engine_reconcile.go||buildReconcileResolveRuntime` **하나만**, 그 함수 편집과 같은 커밋에서 갱신(Manager 결정 2026-11-23) | 판정이 그 클로저 안에 있어 본문 편집이 불가피하다. 저장소에 corpus 생성 도구·`-update` 플래그는 없다(`shapes.json` 을 읽는 곳은 a166_wiring_census_test.go:35-36 뿐). 절차는 선례 a189 `3af104af9`(design D10)·a192 `analysis/census-repair/`(`shape.go.txt`·`enc.go.txt`·`shapes-key-diff.txt`)를 따른다: census 와 같은 추출(parser mode 0 · Doc 제거 · `format.Node`)과 Go `encoding/json` 문자열 인코딩으로 값을 만들고, base 에서 같은 추출이 기존 값을 바이트 그대로 재현함을 먼저 단언한 뒤 바꾼다. 키별 비교로 그 키 하나만 바뀌고 `sites.json` 무변경임을 기록한다. 손으로 값을 고치지 않는다 |
| D7 | 도움말(`newEngineReconcileResolveCmdWithDeps` 의 `Long`, :81-89)과 성공 출력(`runEngineReconcileResolve` :212)을 원인 중립 문장으로(예: 「활성 RECONCILE 차단」) 고친다(Manager 결정 2026-11-23). 성공 출력은 **원인 중립 문장만** 쓰고 해제한 원인 목록은 출력하지 않는다 — 목록을 넘기는 데이터 경로(runtime 구조체·클로저·census 확장)를 만들지 않는다(Manager 결정 2026-11-23, r2 N2) | 「quantity-mismatch」만 말하는 문장은 이 change 뒤 거짓이 된다. 두 함수 모두 a166 census 키가 아니다(shapes.json 의 engine_reconcile.go 키는 `buildReconcileResolveRuntime` 하나). 파일 텍스트 시험(`TestEngineReconcileResolveProductionWiringHasNoBrokerMutationOrHTTPDoor`)의 금지 토큰(`PlaceOrder`·`CancelOrder`·`AmendPendingOrder`·`ModifyOrder` 등, 대소문자 구분)을 쓰지 않고, 성공 문장에 「released」를 남긴다(engine_reconcile_test.go:291·:394 포함 단언, :316 실패 시 부재 단언) |

## 2. 거절하는 정상 입력 (fail-closed 가 무엇을 거부하는지)

- `IDENTIFIER_CONFLICT` 활성 행이 하나라도 있으면 명령 전체가 거절된다 — 같이 있던 `QUANTITY_MISMATCH`·`ATTRIBUTION_FAILED` 도 풀리지 않는다. 종전과 같은 동작이고(원인 하나라도 근거 없으면 전부 거절), 이 change 는 그것을 바꾸지 않는다.
- 재조회 비교에 진입을 막는 차이(수량 불일치, 원장에만 있는 주문)가 남아 있으면(다른 종목의 차이 포함 — 비교는 계좌 전체) 종전대로 거절된다. 이 판정 함수는 건드리지 않는다.

통과시키는 새 입력: 활성 블록이 {`QUANTITY_MISMATCH`, 종목 `ATTRIBUTION_FAILED`(세 유형 모두)} 로만 이루어진 모든 조합. 외부 보유·외부 주문·평단 차이는 종전처럼 거절 사유가 아니다(D2) — 해제 대상 종목의 외부 보유도 포함한다.

## 3. 시험 전략 (RED 먼저, Manager 결정 2026-11-23 으로 보강)

새 시험 파일 `cmd/tossctl/a195_reconcile_resolve_attribution_test.go`. 가짜 수집기·임시 원장·합성 값만.

1. **헬퍼 단위 시험(원인 표 고정)** — 새 원인 판정 함수에 직접: QM 만(통과) / 종목 AF 만(통과) / QM+AF 와 AF+QM(순서 둘 다, 통과) / `IDENTIFIER_CONFLICT` 단독(거절) / AF+`IDENTIFIER_CONFLICT` 와 그 반대 순서(거절, 문구가 `IDENTIFIER_CONFLICT` 를 이름으로 말함) / 계좌 전체 AF(거절) / 빈 목록(종전 문구 거절 — B12 를 옮길 때. 안 옮기면 아래 4 가 덮는다).
2. **원장을 거치는 통합 시험** — 실 원장 임시 경로에 행을 넣고 `Tracker.Restore` → production `runEngineReconcileResolve` 를 가짜 runtime 으로 구동하되 validate 는 **스텁이 아니라 실제 헬퍼를 부르는 클로저**(`tracker.Blocks()` → 헬퍼), resolve 는 `tracker.Resolve`:
   - 2a. 종목 AF 한 행 + 진입을 막는 차이 없음 → 해제: `released_at` 채워짐, `release_cause = OPERATOR`, 근거에 운영자 신원·note, 재-Restore 블록 0.
   - 2b. 같은 행 + 진입을 막는 차이 남음 → 거절, 행 무변(종전 판정이 앞에 선다).
   - 2c. QM + AF → 둘 다 해제.
   - 2d. AF + `IDENTIFIER_CONFLICT` → 거절, 두 행 다 무변.
3. **외부 보유 입력(D2 고정)** — 실제 `reconcile.Comparer` 와 `reconcile.LocalStateFromJournal` 로 만든 diff: AF 종목의 인스턴스가 CLOSED 이고 계좌에 그 종목 양수 보유(→ `ExternalPos`, `BlocksEntry()` 거짓) → 해제가 진행됨.
4. **활성 블록 0** → 종전 문구 거절(회귀).
5. **자동 해제 금지 회귀** — AF 행 + 차이 없는 비교를 `Tracker.Observe` 에 여러 번 → 미해제(기존 동작 고정, spec 「귀속 실패는 스스로 풀리지 않는다」).
6. **문구(D7)** — 성공 출력이 원인 중립이고 「released」를 포함.

변이(전부 CAUGHT 여야 함, 변이마다 닿음 확인, 사본 하네스는 무변이 대조군 GREEN·시작 sha 단언 후): 원인 판정 삭제 · 허용 목록에 `IDENTIFIER_CONFLICT` 추가 · 종목 범위 조건 삭제(= 계좌 전체 AF 허용) · `ATTRIBUTION_FAILED` 를 목록에서 빼기(1·2a·2c 가 죽어야 함) · 순회 중 첫 허용 원인에서 즉시 통과 반환 · **클로저가 헬퍼 결과를 무시** — 이것은 census 가 그 함수 본문을 고정하므로 `TestA166WiringCensusBaseline` 이 죽이는 것으로 보고, 변이 원장에서 census 시험으로 실제로 죽음을 확인한다(구조 검출임을 원장에 적는다).

선택 시험: `./cmd/tossctl -run 'TestA166EngineReconcile|TestEngineReconcileResolve|A195'`, `./internal/accountbudgetruntime`(census), `./internal/reconcile -run 'Resolve|Restore|Observe'`. 전체 스위트는 종결 게이트에서만.

## 4. 운영 절차 (배포 뒤, 사람 승인)

1. 엔진 정지 창에서(두 시장이 닫힌 시간, 또는 엔진이 이미 꺼져 있을 때) `tossctl engine reconcile-resolve --confirm …` 실행 전, 거절 조건을 장부로 예측한다 — 활성 행의 원인 목록과 그 종목의 계좌·장부 수량.
2. 실행 뒤 `reconcile_states` 의 그 행 `released_at`·`release_cause` 와 다음 `reconcile.clean` 의 `blocked` 값을 확인한다. 해제는 진입 차단을 푸는 것이지 편입·손절 복원 자체가 아니다 — 재기동 뒤 그 보유의 편입과 보호 상태를 따로 확인한다 — 대상 종목이 편입 제외 목록에 없는지·편입 설정, 그리고 실제 exit state 가 열렸는지(미편입 포지션은 exit 관측 대상에서 빠진다, exitloop.go:592).

## 5. 후속 후보와 잔여 위험 (이 change 밖, 수리 안 함)

1. 경합 자체의 수리 — 엔진 주문의 매도 체결이 「방금 수렴 조정으로 닫힌 인스턴스」에 닿는 경우를 차단 없이 처리(조사의 β: 싱크 쪽 인정, α: 수렴 보류, γ: 증명 뒤 자동 해제). 같은 경합의 반대 순서(체결 먼저)가 `QUANTITY_MISMATCH` 고착을 만드는지는 추론이고 재현 전이다.
2. 그 경합에서 엔진 손절이 `COMPLETED`·trade outcome 없이 「외부 청산」으로 기록되는 것(조사 인용 F15).
3. 차단 진입 순간의 알림과 차단 때문에 편입되지 못한 보유의 알림 — a194.
4. 잔여 위험: 해제 뒤 같은 매도 주문의 뒤이은 체결 이벤트가 귀속 실패를 다시 걸 수 있다(미측정) — a194 의 차단 보고가 그것을 드러낸다.
5. 잔여 위험: scope_market 이 붙은 `QUANTITY_MISMATCH` 행이 있으면 `Tracker.Resolve` 의 배치가 「not active」로 전부 실패하는 기존 한계(요청에 `ScopeMarket` 없음, reconcile_states.go:463-465 는 빈 시장이면 `scope_market IS NULL` 만 찾음).
6. 기록: `Tracker.Resolve` 의 실제 식은 `firstNonEmpty(block.Cause, QUANTITY_MISMATCH)` 다(1단계에서 정정함).
