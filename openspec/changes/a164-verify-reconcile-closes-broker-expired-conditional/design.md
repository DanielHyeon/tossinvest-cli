# a164 design — 브로커 확인 만료 대사

함수 내부 주장의 근거는 이 change 의 AST 산출물이다: `analysis/ast/*.ast.json` 29개와 그 기계 열거
`analysis/ast/enumerate-output.md`(`python3 -I enumerate.py`, 2026-10-22, HEAD `4cd7af19d`, 분기 169 · 반환 138 · STALE 0).
아래 `Bn` 은 그 열거의 분기 id 다. 공식 API 사실은 `docs/migration/openapi.latest.json`(2026-10-16 판)과
`openspec/changes/verify-execution-capability/measurements.md` 의 M 행에서만 인용한다.

## D0. 결정 기록 — Q6 대체

- 원 결정(a121 design 「Q6」, Manager 2026-09-27): 대상이 CLOSED 에 `EXPIRED` 로 있으면 거절, 「이 경로로는 영구히 대사되지 않고,
  그때 사용자 재결정 항목이 된다」. 코드: `checkReconcileAbsence` B5 `reconcile_check.go:384` → `RefuseClosedTargetExpired`,
  문구 `reconcileQ6Message`(`:21`).
- **사용자 결정 2026-10-22 (a)**: 만료된 333430 artifact 를 정리하는 경로를 리뷰된 코드 변경으로 연다.
- 처분: Q6 의 **사실 판단**(만료는 부재가 아니라 종결)은 유지한다 — 부재 모드는 계속 그것을 거절한다. Q6 의 **귀결**(「영구히 대사되지
  않는다」)은 **브로커가 이 실행 안에서 종결을 양성 확인한 artifact 에 한해** 이 change 가 대체한다. 그 밖의 경우(브로커가 더는
  이름하지 않음, 과거 관측만 있음)는 Q6 귀결이 그대로 남는다(H3).

## D1. 모드 — 명시 플래그 `--expired`

`Reconcile` 의 정적 거절(B8 Q1 `:205`)은 사람 승인 **앞**에 있고(a121 A-GREEN P2-a), 대상이 만료인지는 목록을 읽은 **뒤**에야 안다.
자동 판별로는 「Q1 을 건너뛸지」를 승인 전에 정할 수 없다. 그래서 운영자가 승인 전에 모드를 고른다.

- 플래그 없음: 지금 경로 그대로(코드·문구·거절 순서 불변 — 단 D8 의 Q6 문구 한 줄 수정은 예외).
- `--expired`: D4 의 순서. 승인 화면과 y/N 질문이 「reconciled expired 로 기록할까?」로 모드를 이름한다.
- 거부안: 자동 판별(부재·만료를 읽은 결과로 가름) — Q1 정적 거절을 승인 뒤로 옮겨야 해서 a121 P2-a 를 되돌린다. H1.

## D2. 증거 규칙 (핵심 계약)

**닫는 조건(전부 참일 때만, 하나라도 판정 불능이면 아무것도 쓰지 않음):**

- A1. a121 공통 전제 전부: 엄격 해독 · 후보 정확히 하나(`selectReconcileCandidate` — `PendingCleanup` 의 조건주문 ∖ M0 부모) ·
  계좌 결속(mixed → mismatch) · Q3 상수 확정 · 승인 · 종목 양성 대조 ×2 · 목록 두 번 읽기 multiset 일치 · 두 읽기 각각 행 검사
  (OCO → 필드 → 심볼 → 시장) · 첫 읽기 시작부터 추가 직전까지 15 s(`reconcileFreshnessBoundValue`) · 잠금 fd 재해독·지문 일치.
- A2. **by-id 상세 GET 두 번** — 첫 목록 읽기 직전 한 번, 둘째 목록 읽기 직후 한 번. 응답마다 셋 중 하나로 분류한다(D3):
  `statement`(200 + 엄격 스키마) · `silent`(404 + 본문 code `conditional-order-not-found`) · 그 밖 전부 거절. 두 분류가 같아야 하고,
  둘 다 statement 면 (status, first.status, first.triggeredOrderId, second 존재) 가 같아야 한다.
- A3. **진술 집합** S = by-id statement ∪ 두 읽기의 CLOSED 행 중 `id == target.ID`. **|S| ≥ 1** — 대상 id 를 이름하는 현재 공식 응답이
  하나도 없으면 거절(`expiry-evidence-absent`).
- A4. S 의 **모든** 원소가: top-level status `EXPIRED` · `triggeredOrderId` 비어 있음 · second 없음. by-id statement 는 추가로
  `conditionalOrderId == target.ID`(바이트 일치) · `symbol == target.Symbol` · `market == 기록 시장` · `first.status` 가 영수증으로
  고정한 만료 leg 어휘 안(0.6 영수증 전 미확정 = 거절, H5). 하나라도 다르면 거절 — `COMPLETED` 는 `expiry-target-fired`,
  OPEN 계열(`WATCHING/PAUSED/ORDERING/ORDERED`)은 `expiry-target-live`, 그 밖은 `expiry-target-status`.
- A5. 같은 심볼의 OPEN 조건주문 0건 · OPEN 일반 주문 0건 · CLOSED 의 **다른 id** 행에 발동 흔적(`triggeredOrderId` 또는 `COMPLETED`) 0건 —
  a121 G1-1·2·3 과 같은 판정·같은 거절 코드(`open-conditional` · `open-plain-order` · `closed-fired-trace`). 한 규칙은 한 자리에 —
  두 모드가 같은 보조 함수를 부른다(D4).
- A6. CLOSED 행 status 허용 목록(expiry 모드 전용 `reconcileExpiryClosedVocabulary`): 영수증 고정 전 비어 있음 = 전부 거절.
  부재 모드의 `reconcileClosedStatusAllowlist`(비어 있음)는 건드리지 않는다.
- A7. 대상 `ChainID` 가 비어 있지 않으면, 기록에서 같은 체인의 다른 artifact 가 전부 종결이어야 한다(`expiry-chain-open`).
  체인은 이 도구가 만든 정정 후속을 묶는다(`record.go:240-246`, M19·M40) — 대상의 만료는 후속에 대해 말하지 않는다.

**이 가드가 통과시키는 것:** 이 기록이 소유한, 같은 계좌·프로필·시장의 SINGLE 조건주문 하나로서, 이번 실행의 공식 GET 하나 이상이
그 id 를 `EXPIRED`·미발동으로 이름하고, 같은 심볼에 살아 있는 조건주문·호가에 걸린 일반 주문·발동 흔적이 없으며, 15 s 안에 두 번 읽은
결과가 같은 것. by-id 가 404 여도 CLOSED 목록이 이름하면 통과하고, 목록이 이미 축출했어도 by-id 가 이름하면 통과한다.

**이 가드가 거부하는 정상 입력(실제로 만료됐는데도 닫지 못하는 경우) — 의도된 비용:**

1. 브로커가 그 id 를 더는 어디에서도 이름하지 않음(by-id 404 + CLOSED 부재). 333430 이 이 경우일 수 있다 — tasks 0.6 이 먼저 잰다. H3.
2. 같은 심볼에 다른 OPEN 조건주문이 있음 — 엔진 브로커 보호 조건주문(a100 배선 뒤)·앱에서 건 수동 조건주문 포함(Q4 와 같은 비용).
   333430 은 엔진 adoption 종목이다(`ma-runbook.md:32`). H4.
3. 같은 심볼에 OPEN 일반 주문이 있음(엔진 청산 지정가 등). H4.
4. OCO·OTO artifact, 계좌가 둘 이상인 자격(a121 Q5), by-id 429·5xx·전송 오류(사람이 나중에 다시 돌림).
5. 만료 leg 어휘 영수증 전(H5) — 구현은 끝나도 0.6 영수증 전에는 늘 거절.

## D3. 거부한 가드 입력 — 근거가 아닌 것

| 입력 | 거부 이유 |
| --- | --- |
| 정리 DELETE 404 (몇 번이든) | 같은 404 가 정정(살아 있는 후속 — M19·M40 `conditional-order-not-found`), 취소(M20·M41 「식별자가 더 이상 읽히지 않는다」), 축출을 구분 없이 말한다. a121 spec 「DELETE 404 remains nonterminal」 유지 |
| by-id GET 404 단독 | 위와 같은 이유. `silent` 로만 분류하고 진술로 세지 않는다 |
| 목록에서 id 부재 | 부재 모드(Q1)의 질문이다. 만료 모드는 부재에서 아무것도 추론하지 않는다 |
| soak by-id probe 「ok」 | `soakReads.ConditionalOrder` 는 오류만 돌려준다(`cmd/tossctl/soak.go:802`, 본문 버림). 계약상 상세 GET 은 종료 주문도 돌려주므로 ok 는 생존도 만료도 아니다 |
| 과거 관측(08-13 감사, 기록의 옛 줄) | 이번 실행의 계좌 결속·Q3 창·두 번 읽기에 묶이지 않는다. H3 |
| 로컬 시계와 만료일 | 기록 `Artifact` 에 만료일 필드가 없고(`record.go:187-255`), 브로커 증거가 아니다 |
| leg status 단독(top-level 없이) | top-level status 는 「살아있는 leg 의 상태를 대표」(계약). top-level `EXPIRED` 를 요구하고 leg 는 일관성 검사로만 쓴다 |
| 호출자가 준 id·심볼 | a121 G2 그대로 — 후보는 기록이 고르고 조회 심볼은 artifact 줄의 바이트 |

### by-id 응답 분류(A2)

- `statement`: HTTP 200, result 가 객체, `conditionalOrderId`·`status`·`symbol`·`market` 이 non-null 문자열, `first` 가 객체이고
  `first.status` 가 non-null 문자열. 이 다섯은 계약의 required 다(`ConditionalOrderDetailResponse.required` ·
  `ConditionalOrderCondition.required`) — 결측·null·형 불일치는 a121 목록 해독과 같은 이유(F2 — null 이 제로값으로 접힘)로 거절.
  `first.triggeredOrderId` 와 `second` 는 계약상 required 가 **아니다**(nullable) — 키 없음과 null 은 둘 다 「없음」이고, 있으면
  형만 검사한다(`triggeredOrderId` 는 문자열, `second` 는 객체 — 다른 형은 거절). **정정 2026-10-22**: 첫 판은 두 키의 존재를 요구했는데,
  그것은 계약에 맞는 응답을 거절하는 가드였다(아래 「0.6 영수증 대조」 C2).
- `silent`: `*official.APIError` 의 `Code == 404` 이고 본문 error.code 가 `conditional-order-not-found`(M40 실측 문자열). 본문 해독 불능 404 는
  거절(silent 아님).
- 그 밖: 400·401/403(`ErrAuth`/`ErrIPNotAllowed`)·429(`ErrRateLimited`)·5xx(`ErrServer`)·전송·스키마 오류 → `expiry-read-error`
  (`internal/official/errors.go:39-61` `classifyStatus`).

## D4. 가드 순서(`--expired`)

사전 검사(기록 엄격 해독 → 후보 → 계좌) → 정적 거절(Q3 상수 · **만료 어휘 확정** — Q1 은 부르지 않음) → 승인(모드 표시) → 종목 대조 ① →
by-id ① → 목록 두 번(조건 OPEN → 일반 OPEN → CLOSED, a121 고정 순서) → by-id ② → 종목 대조 ② → 신선도 → 목록 multiset · by-id 쌍 일치 →
행 검사 ×2 → A5(OPEN 조건 → 일반 OPEN → 다른 id 발동 흔적) → A3·A4 진술 판정 → A7 체인 → A6 CLOSED 허용 목록(마지막) → 잠금·재해독·지문 →
줄 직렬화 → 쓰기 직전 신선도 → 한 번의 쓰기·fsync·read-back.

부재 모드 순서는 바꾸지 않는다. A5 의 세 판정은 `checkReconcileAbsence` 의 B1·B2·B6/B7 와 **같은 보조 함수**로 옮겨 두 모드가 부르고,
부재 모드 안의 거절 순서(OPEN 조건 → 일반 OPEN → CLOSED target → 발동 흔적)는 유지한다(a121 「가드 순서 계약」 load-bearing).

## D5. 기록 — 추가 한 줄

- 줄: `Kind = "reconcile"`(기존 `KindReconcile`) · 새 `StepID = "reconcile-expired"`(카탈로그 ∪ {cleanup, abort, reconcile-absent} 밖) ·
  Title `reconciled expired` · Calls 없음(a121 과 같음).
- Artifact: 새 필드 `ReconciledExpired bool json:"reconciled_expired,omitempty"` + 기존 `ReconciledAt`. `terminal()` 은
  `Cancelled || Filled || ReconciledAbsent || ReconciledExpired`. 대사 artifact 확정형 직렬화(`Artifact.MarshalJSON`)를 넷째 종결로 넓힌다.
- 관측(Observations): `reconcile.basis`(목록 multiset + by-id 진술 행을 그룹 `conditional-by-id` 로 포함한 지문 — 도메인 태그
  `a164/reconcile-expired-basis/v1`) · `reconcile.expired.by_id`(두 응답의 분류와 status·first.status·triggeredOrderId 원문) ·
  `reconcile.expired.closed_rows`(CLOSED 에서 대상을 이름한 행의 status 원문) · 시각은 줄의 `started_at`(첫 목록 읽기)·`finished_at`(추가 시각).
- 요청 id: 실패 응답 본문의 `requestId` 는 silent 분류 때 함께 적는다(M40 처럼 본문에 있음, 계좌 정보 없음). **성공 응답의 requestId 를
  클라이언트가 노출하는지는 미확인** — 구현 로트 1.3 에서 `getAcct` 를 읽어 확정하고, 없으면 응답 바이트 `Digest`(16바이트 절단)를 대신
  적는다. 원문 응답·계좌 원문은 쓰지 않는다.
- 기존 줄을 다시 쓰지 않는다. 실패한 정리 줄은 그대로 남는다.

## D6. rate budget (안전 불변식 §0-4)

`--expired` 실행 1회 = a121 읽기(Accounts 2 · 종목 2 · 조건 목록 OPEN/CLOSED × 페이지 × 2 · 일반 OPEN × 페이지 × 2) + **by-id 2**.
by-id 는 목록과 같은 `CONDITIONAL_ORDER_HISTORY` 그룹(OpenAPI `getConditionalOrder`). 실행은 기존 `acquireVerifyRateBudget` lease 안에서만
돈다. 자동 재시도 없음(429 는 거절).

## D7. Q1 을 만료 모드에 적용하지 않는 이유

Q1 은 「발동한 조건주문이 CLOSED 에서 축출돼 부재가 발동을 숨길 수 있다」를 막는다(a121 G1-4). 만료 모드는 부재에서 추론하지 않고
대상 자신의 종결을 양성으로 읽는다(A3·A4). 남는 후속 위험은 정정뿐인데, 정정은 옛 id 를 **즉시 404** 로 만든다(M19 US · M40 KR) —
옛 id 가 `EXPIRED` 로 읽힌다는 것은 그 id 가 정정으로 대체되지 않았다는 측정과 맞는다. 그래도 A5(같은 심볼 OPEN 조건 0 · 발동 흔적 0)와
A7(기록 체인)을 유지한다. 잔여는 R2.

## D8. 부재 모드 Q6 문구

`reconcileQ6Message` 를 「만료된 artifact 는 부재 대사로 받지 않는다 — 브로커가 확인한 만료는 `--expired` 로 대사한다」로 바꾼다.
거절 코드·위치·순서는 그대로. spec 의 「cannot be reconciled through this operation at all」은 MODIFIED 로 고친다.

## D9. 범위 — M0 전제 판정(H2)

cmd `validateM0TriggerMode`: B3 `M0Unsettled` → B7(`!ok`) 안에서 B8 `M0ExactPrerequisites`(`verify.go:424`) → B9 `Outstanding`(`:428`).
`verifylive.New`: B9 `Outstanding`(`runner.go:346`) → B10 `M0Unsettled` → B15 `M0ExactPrerequisites`(`:358`). 즉 cmd 진입은 전제 판정에서
먼저 멈추고, `M0ExactPrerequisites`(`m0_recovery.go:98`)는 trigger·TTL-edge 외 모든 단계의 `Passed` 를 요구한다. 감사가 보고한 마지막 판정 fail
셋(sell-boundary · conditional-modify · conditional-cancel)은 실주문 재측정으로만 PASS 가 된다 — 코드가 아니다.
**기본(KISS): a164 는 미결 artifact HOLD 만 다룬다.** a164 착지 뒤에도 `--include-trigger` 는 B8 에서 멈춘다. 사람 결정 H2.

## D10. 호환·되돌리기

- 구 바이너리: `decodeEntry`·`decodeRecordStrict` 는 `json.Unmarshal`(unknown field 허용, `record.go:597-612` · `reconcile_record.go:47-64`)이라
  새 필드를 버리고 그 줄을 비-종결로 읽는다 → artifact 가 다시 outstanding·정리 대상(오늘과 같은 상태, DELETE 404 반복) — 안전 방향.
- 되돌리기: 이전 이미지/바이너리로 교체하면 위 상태로 돌아간다. 추가된 줄은 지우지 않으며 새 바이너리가 다시 존중한다. 스키마·설정·토글 변경 0.

## 사람 결정

| id | 질문 | 기본(권장) | 상태 |
| --- | --- | --- | --- |
| H1 | 모드: 명시 `--expired` vs 읽은 결과로 자동 판별 | 명시 플래그(D1) | **확정 2026-10-22(사용자 「둘 다 예」) — 명시 `--expired` 만** |
| H2 | 범위: 미결 HOLD 만 vs `M0ExactPrerequisites` 전제 fail 셋까지 | 미결 HOLD 만(D9) — 전제는 별도 실측 계획 | **확정 2026-10-22 — 미결 artifact HOLD 만. 마지막 판정 fail 전제 셋은 실주문 재측정으로 남김** |
| H3 | 브로커가 더는 이름하지 않을 때(0.6 결과가 404+부재) 과거 관측(08-13)을 근거로 받을지 | 받지 않음 — 그 경우 a164 는 0.6 에서 멈추고 재결정 | **확정 2026-10-22 — 받지 않음**(0.6 에서 브로커가 현재 이름하므로 이번에는 쓰일 일 없음) |
| H4 | 같은 심볼의 다른 OPEN 조건주문·OPEN 일반 주문을 만료 모드에서도 거절할지 | 거절 유지(Q4 와 같은 비용, D2 거부 2·3) | **확정 2026-10-22 — 거절** |
| H5 | 만료 어휘(`EXPIRED`, leg status) 출처 | OpenAPI enum + 0.6 사람 실측 영수증 골든, 둘 다 있어야 허용 목록에 넣음 | **확정 2026-10-22 — 둘 다 요구.** 영수증: top-level `EXPIRED` · `first.status` `EXPIRED`(`analysis/receipts/expiry-probe-2026-10-22.json`) |

## 0.6 영수증 대조 (2026-10-22)

영수증: `analysis/receipts/expiry-probe-2026-10-22.json` — Manager 가 사용자 승인(2026-10-22 「둘 다 예」) 뒤 엔진 프로필 자격으로 돌린
읽기 전용 GET 셋, 수량 필드 제거. 반입 때 키 전수를 세어 계좌·수량 키가 없음을 확인했다.

- ① by-id: rc 0 · top-level `EXPIRED` · `SINGLE` · 333430 · KR · `first` {STOP, `EXPIRED`} · `second`·발동 id 키 없음 · `expire_date` 2026-08-07.
- ② CLOSED·333430: 행 1 = 그 id, `EXPIRED`, `has_next` false. ③ OPEN 조건·333430: 0 행.
- 판정(tasks 0.6): **계속** — 증거 규칙 A3 의 두 출처가 모두 현재 대상을 `EXPIRED` 로 이름한다.

설계 가정과 어긋난 점:

- **C1. 영수증은 원 응답 바이트가 아니다.** `order conditional get/list` 는 도메인 어댑터(`official.Client.ConditionalOrder` →
  `adaptConditionalOrder`)를 거친 CLI 출력이다 — 키가 snake_case(`triggered_order_id`, `has_next`)이고 원 API 는 camelCase
  (`conditionalOrderId`, `triggeredOrderId`, `hasNext` — OpenAPI). status 문자열은 어댑터가 그대로 옮기므로(`conditional_reads.go:240` 최상위 · `:226` leg — `adaptConditionalOrder`·`adaptCondition`)
  **어휘 값의 영수증으로는 유효**하다(H5). 그러나 tasks 2.1 의 원 바이트 골든으로는 쓸 수 없다 — 원 형태 골든은 OpenAPI 스키마 +
  이 영수증의 값으로 만들고, 원 응답의 키 존재 여부는 미측정으로 둔다.
- **C2. 키 부재를 구별하지 못한다.** 도메인 `TriggeredOrderID` 는 `omitempty`, `Second` 는 포인터 `omitempty`(`internal/domain/models.go:784` · `:798`)라
  CLI 출력에서 null·키 없음·빈 문자열이 같아진다. 동시에 OpenAPI 는 두 키를 required 로 두지 않는다 → 첫 판 D3 의 「키 존재 요구」는
  계약 위반이 아닌 응답을 거절했을 것이다. D3 를 정정했다(키 없음 = null = 없음).
- **C3. 영수증의 `target_profit_rate: 0`·`order_price: 0` 은 브로커 값이 아니다** — 도메인 `parseDecimal` 이 null 을 0 으로 접은 값이다. 판정 입력으로 쓰지 않는다.
- **C4. 「OPEN 일반 주문 0」은 이번 영수증으로 증명되지 않았다.** `tossctl orders list` 는 hybrid 가 WTS 에 맡긴 세션 읽기
  `ListPendingOrders` = `/api/v1/trading/orders/histories/all/pending`(`internal/client/account.go:155-174`, `internal/hybrid/client.go:298-300`)이다 —
  미체결 목록이긴 하나 **비공식 WTS** 출처다. 대사가 실제로 부르는 것은 공식 `GET /api/v1/orders?status=OPEN&symbol=…`
  (`official.Client.ReconcileOpenOrdersPage`)이고 이 GET 을 내는 CLI 하위 명령은 없다(`order list --status` 는 없는 플래그). 그래서 H4 의
  일반 주문 쪽은 실기록 대사 실행(tasks 6.2) 안의 공식 읽기가 처음으로 판정한다 — 비어 있지 않으면 그 실행이 `open-plain-order` 로 거절한다.
- **C5. 첫 목록 시도가 429**(「lazy account-seq resolution: official: rate limited」)였고 5초 뒤 재시도로 통과했다. 대사 명령은 자동 재시도가 없으므로
  (D6) 같은 일이 생기면 `accounts-read`/`read-error` 로 거절되고 사람이 다시 돌린다 — 설계와 맞음, 운영 메모.
- 맞은 가정: by-id 가 종료 주문을 돌려준다(OpenAPI 서술과 일치) · 정리 DELETE 404 와 by-id 200 `EXPIRED` 가 공존한다 — 「404 는 종결 근거가 아니며
  만료 대상은 by-id 로 아직 읽힌다」는 D3 의 전제가 실측으로 섰다. `expire_date` 2026-08-07 은 감사의 08-13 `EXPIRED` 와 맞다(H3 — 근거로는 안 씀).

## 잔여 위험

- R1. 계약에 없는 의미: `ORDERED` 뒤 child 가 남은 채 그룹이 `EXPIRED` 가 되고 `triggeredOrderId` 가 빠지는 경우 — A5 일반 OPEN 0 이 resting
  child 를 잡는다. child 가 이미 체결됐다면 살아 있는 조건주문은 없지만 기록이 「만료」라고 잘못 부른다(감사 정확도 문제, 안전 문제 아님).
- R2. 정정된 옛 id 가 장기적으로 `EXPIRED` 로 다시 보일 가능성 — M19·M40 은 「즉시 404」만 쟀다. A5·A7 이 부분적으로 덮는다.
- R3. 성공 응답 requestId 노출 여부 미확인(D5).
- R4. a164 는 `--include-trigger` 를 풀지 않는다(D9).
- R5. CodeGraphContext 색인이 이 체크아웃을 못 본다(`analysis/code-context/cgc-find.txt` — a121 심볼 0건, 옛 경로) — advisory 미반영.
- R6. 감사 사실(08-13 EXPIRED, 404 네 번)은 이 로트에서 재현하지 않았다(proposal 「미검증」).

## FLM/BTM 대상(구현 로트 1절, High-risk 인접 — 면제 없음)

편집할 기존 함수: `verifylive.Reconcile` · `checkReconcileAbsence`(보조 함수 추출) · `Artifact.terminal` · `Artifact.MarshalJSON` ·
`encodeReconcileLine` · `ReconciledArtifacts` · `writeReconciled`(report.go) · `ReconcileApproval.WriteText` · cmd `newVerifyReconcileCmd` ·
`runVerifyReconcile` · `verifyReconcileApprove`(변수 함수 리터럴 — logic-map 이 못 잡으면 not-applicable 사유 + 행동 시험). 새 함수는 새 파일
(`internal/verifylive/reconcile_expired.go`, `internal/official/reconcile_detail_reads.go`).
