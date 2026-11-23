# a195 — 귀속 실패(ATTRIBUTION_FAILED) 종목 차단을 운영자가 풀 수 있다

> **상태: 1단계(계약·증거), High-risk(대사 해제).** Story `STORY-TOS-a195`(FEAT-TOS-005). 작성: Manager(2026-11-23).
> 분기 주장의 AST 는 tasks 1.3 에서 base `064f59fd9` 로 다시 뽑아 `analysis/cited-ast/` 에 두었다(2026-11-23). 표의 인용은 그 산출물이다.

## Why

2026-11-20 12:56 KST 운영(Manager 실측 — 운영 장부·엔진 로그, 시각·상태만):

- 엔진 손절 매도(전량)가 12:56:39 접수 확정됐다. 12:56:42 대사가 계좌 0 주를 먼저 보고 수렴 조정(`UNKNOWN`, 31 → 0)으로 포지션을 닫았다.
- 같은 초에 그 주문의 체결이 「이미 닫힌 인스턴스에 대한 매도」(전이 X59 `SELL_ON_CLOSED`)로 거부되어 그 종목에 `ATTRIBUTION_FAILED` 차단이 걸렸다. `released_at` 은 지금까지 비어 있다.
- 11-23 08:09 KST 같은 종목을 다시 샀을 때 엔진은 그 보유를 편입하지 않았다(차단 종목은 편입 판정에서 빠진다). 손절 상태·관측 기록 0 건. 그 포지션은 손절 없이 2 시간 반을 갔다.

코드 근거:

| 사실 | 근거 |
|---|---|
| 운영자 해제 명령은 `QUANTITY_MISMATCH` 가 아닌 활성 원인이 하나라도 있으면 거절한다 | `cmd/tossctl/engine_reconcile.go` `buildReconcileResolveRuntime` 의 validate 클로저 — range@307 → if@308 → return@309, 문구 `active cause … requires cause-specific recovery evidence`(AST: analysis/cited-ast/buildReconcileResolveRuntime.ast.json B13·B14, return@309) |
| 그 거절은 엔진 잠금과 스냅숏 수집 뒤에 난다 — 엔진을 세워도 실패한다 | `runEngineReconcileResolve` if@201(AST: analysis/cited-ast/runEngineReconcileResolve.ast.json B22; 잠금 B7 if@134, 스냅숏 B11~B19, 비교 B21 if@198 이 앞선다) |
| `ATTRIBUTION_FAILED` 를 푸는 다른 자리는 없다 | 자동 해제 둘(`Tracker.Observe` 해제 루프, a192 종결 증명)은 `QUANTITY_MISMATCH` 만 본다(AST: analysis/cited-ast/Tracker.Observe.ast.json B6 range@666 → B7 if@667 continue; analysis/cited-ast/orderTerminalReleaseVerdict.ast.json B3 if@182 — 원인 조건은 그 if 의 줄 184). 해제 쓰기 자리는 이 둘과 운영자 해제(`Tracker.Resolve`)뿐이다(HEAD grep, `analysis/code-context/codegraph-baseline.md`) |
| 해제 기록 자체는 원인을 가리지 않는다 | `Tracker.Resolve` 가 블록마다 `ExpectCause: firstNonEmpty(block.Cause, QUANTITY_MISMATCH)`, `Cause: OPERATOR` 로 요청을 만든다(`ScopeMarket` 는 비움 — `ATTRIBUTION_FAILED` 행은 scope_market NULL 로만 쓰인다, position_projection.go:409). 장부는 활성 행의 원인이 `ExpectCause` 와 같을 때만 푼다(reconcile_states.go:383) (AST: analysis/cited-ast/Tracker.Resolve.ast.json B4 range@1363, B5 if@1364 — 영속되지 않은 빈 종목 수량 블록만 건너뜀) |

즉 **이 차단은 지금 빌드에서 어떤 자동 경로로도, 어떤 운영자 명령으로도 풀리지 않는다.** 정본 reconciliation 은 영구 불일치의 해제를 「운영자 확인」으로 정했는데, 이 원인에는 그 확인을 받는 경로가 없다. 막고 있는 것은 validate 의 원인 목록 한 줄이다.

## What Changes

1. `tossctl engine reconcile-resolve` 가 **종목 범위의 `ATTRIBUTION_FAILED` 차단**도 해제 대상으로 받는다. 받는 조건은 지금 `QUANTITY_MISMATCH` 에 요구하는 것과 같다 — 엔진 잠금, 명령 자신의 안정 재조회가 차이를 보지 않음, 운영자 신원과 확인 내용(note), `--confirm`.
2. 그 밖의 원인(`IDENTIFIER_CONFLICT` 등)은 종전 문구로 거절한다. 거절 대상은 「해제 근거가 정의되지 않은 원인」이고 그 목록이 코드 한 곳에 있다.
3. 해제 기록은 종전 형식 그대로다(release cause `OPERATOR`, 근거 `operator <신원>: <note>`). 원장 스키마·release cause 집합 변경 없음.

## 이 change 가 하지 않는 것

- **차단이 생기지 않게 하는 수리**(엔진 자신의 매도 체결과 수렴 조정의 경합) — 후속 Story 후보(design §5). 이 change 뒤에도 같은 경합은 같은 차단을 만들고, 달라지는 것은 그것을 풀 수 있다는 것이다.
- `ATTRIBUTION_FAILED` 의 자동 해제. 영구 차단은 운영자만 푼다(정본 그대로).
- 엔진을 세우지 않는 해제. 이 명령은 종전대로 엔진 잠금을 잡는다.
- `IDENTIFIER_CONFLICT`·계좌 전체 차단의 해제 근거 정의.
- 차단 때문에 편입되지 못한 보유를 알리는 일 — a194.
- 운영 장부의 해제 실행. 배포와 명령 실행은 별도 사람 승인이다.

## Capabilities

### Modified Capabilities

- `reconciliation`: 운영자 해제가 받는 원인에 종목 범위 `ATTRIBUTION_FAILED` 를 더하고 그 근거(명령 자신의 재조회 일치)를 요구로 적는다.

## Impact

- 기존 함수 본문 편집(gate ⑤ FLM 대상): `cmd/tossctl/engine_reconcile.go` `buildReconcileResolveRuntime`(validate 클로저의 원인 판정을 새 함수 호출로 바꿈). 기존 시험 중 `ATTRIBUTION_FAILED` 거절을 단언하는 것이 있으면 그 시험 함수도 대상이다(tasks 1.4 에서 센다).
- 새 함수는 새 파일에 둔다. `buildReconcileResolveRuntime` 이 a166 census(`internal/accountbudgetruntime` shapes.json)에 고정돼 있으면 **편집 전에** Manager 에게 돌아온다(tasks 1.4).
- 브로커 호출 추가 0. 명령이 이미 하는 재조회만 쓴다.
- 롤백 = 이전 이미지. 이 change 가 쓴 해제 행은 종전 형식의 `OPERATOR` 해제라 옛 이미지가 그대로 읽는다.
