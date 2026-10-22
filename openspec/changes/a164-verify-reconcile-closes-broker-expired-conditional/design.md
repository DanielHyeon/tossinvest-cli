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

**위계 (freeze 리뷰 P2-6, 2026-10-22).** A5 가 **생존 판정**이다 — 같은 심볼에 살아 있는 조건주문·호가에 걸린 child·발동한 후속이
없음을 말하는 것은 A5 뿐이다(OPEN 은 다른 채널 등록분까지 돌려준다 — OpenAPI `getConditionalOrders`; 정정은 심볼을 못 바꾼다 —
`ConditionalOrderModifyRequest` 에 `symbol` 없음). A2~A4·A6·A7 은 **종결 이름 판정**(발동·정정을 「만료」라 부르지 않기)이다.
A5 는 어떤 증거 조합에서도 생략·완화하지 않는다. 증거가 강하다는 이유로 A5 를 비용으로 다루는 변경은 이 change 의 계약 위반이다.

**닫는 조건(전부 참일 때만, 하나라도 판정 불능이면 아무것도 쓰지 않음):**

- A1. a121 공통 전제 전부: 엄격 해독 · 후보 정확히 하나(`selectReconcileCandidate` — `PendingCleanup` 의 조건주문 ∖ M0 부모) ·
  계좌 결속(mixed → mismatch) · Q3 상수 확정 · 승인 · 종목 양성 대조 ×2 · 목록 두 번 읽기 multiset 일치 · 두 읽기 각각 행 검사
  (OCO → 필드 → 심볼 → 시장) · **첫 계좌 범위 읽기(by-id ①) 직전**부터 추가 직전까지 15 s(`reconcileFreshnessBoundValue`, P3-1) ·
  잠금 fd 재해독·지문 일치. 목록 읽기는 a121 `readReconcileSet`(`reconcile_check.go:184-197`)을 **그대로** 부른다 — 세 그룹 중 하나라도
  빠진 읽기 집합은 이 change 의 구현이 아니다(P2-5 ⑤).
- A2. **by-id 상세 GET 두 번** — 첫 목록 읽기 직전 한 번, 둘째 목록 읽기 직후 한 번. 응답마다 셋 중 하나로 분류한다(D3):
  `statement`(200 + 엄격 스키마) · `silent`(404 + 본문 code `conditional-order-not-found`) · 그 밖 전부 거절(`expiry-read-error`).
  두 분류가 다르면 `expiry-reads-differ`. 둘 다 statement 면 (status, first.status, first.triggeredOrderId, second 존재) 가 **필드마다** 같아야 한다.
- A3. **by-id 두 응답이 둘 다 statement 여야 한다**(freeze 리뷰 P1-1). 둘 다 silent 이면:
  CLOSED 목록이 대상 id 를 이름하면 `expiry-sources-disagree`(계약상 상세 GET 은 종료 주문도 돌려주므로 목록-상세 모순이고, 「옛 id 상세 404」는
  M19·M40 이 잰 **정정** 서명이다), 이름하지 않으면 `expiry-evidence-absent`(404 와 부재는 종결 근거가 아님).
  CLOSED 의 대상 행은 필수가 아니다(축출 가능 — Q1 미측정). 있으면 A4 일관성 검사를 받는다.
  진술 집합 S = by-id statement 둘 ∪ 두 읽기의 CLOSED 행 중 `id == target.ID`.
- A4. S 의 **모든** 원소가: top-level status `EXPIRED` · `first.status` 가 만료 leg 어휘 안 · `triggeredOrderId` 비어 있음 · second 없음.
  CLOSED 행의 `first.status` 는 행 타입에 `FirstStatus` 를 더해 읽는다(P2-3 — 해독은 이미 `apiConditionalOrder` 전체를 푼다,
  `internal/official/reconcile_reads.go:141`; 옮기는 대입 한 줄 `:145-148` 에 필드 하나). by-id statement 는 추가로
  `conditionalOrderId == target.ID`(바이트 일치) · `symbol == target.Symbol` · `market == 기록 시장` · `type == "SINGLE"`(P3-3, 계약 enum
  `SINGLE/OCO/OTO`). 하나라도 다르면 거절 — `COMPLETED` 는 `expiry-target-fired`, OPEN 계열(`WATCHING/PAUSED/ORDERING/ORDERED`)은
  `expiry-target-live`, 그 밖(빈 값·미지 status·leg 어휘 밖·type 불일치·id/심볼/시장 불일치)은 `expiry-target-status`.
  만료 leg 어휘·top-level 어휘는 H5 확정대로 OpenAPI enum + 0.6 영수증(top-level `EXPIRED` · `first.status` `EXPIRED`) 둘 다로 고정한다.
- A5. **생존 판정(위계 참조).** 같은 심볼의 OPEN 조건주문 0건 · OPEN 일반 주문 0건 · CLOSED 의 **모든** 행(대상 포함)에 발동 흔적
  (`triggeredOrderId` 또는 `COMPLETED`) 0건 — a121 G1-1·2·3 과 같은 판정·같은 거절 코드(`open-conditional` · `open-plain-order` ·
  `closed-fired-trace`). 한 규칙은 한 자리에 — 두 모드가 같은 보조 함수를 부른다(D4). 공유 발동 흔적 루프(지금 `reconcile_check.go:390-394`,
  AST `checkReconcileAbsence` B6·B7)는 대상 행을 빼지 않으므로 **만료 모드에서 대상 CLOSED 행이 `COMPLETED` 면 `closed-fired-trace` 가 난다**
  (P3-5 — 시험 2.3 이 이 코드를 단언). `expiry-target-fired` 는 by-id 가 `COMPLETED` 를 말하고 목록에 발동 흔적이 없을 때만 난다.
- A6. CLOSED 행 status 허용 목록(expiry 모드 전용 `reconcileExpiryClosedVocabulary` = H5 고정 어휘). 어휘 밖은 `closed-status`.
  부재 모드의 `reconcileClosedStatusAllowlist`(비어 있음)는 건드리지 않는다.
- A7. 대상 `ChainID` 가 비어 있지 않으면, 기록에서 같은 체인의 다른 artifact 가 전부 종결이어야 한다(`expiry-chain-open`).
  체인은 이 도구가 만든 정정 후속을 묶는다(`record.go:240-246`, M19·M40) — 대상의 만료는 후속에 대해 말하지 않는다.

**이 가드가 통과시키는 것:** 이 기록이 소유한, 같은 계좌·프로필·시장의 SINGLE 조건주문 하나로서, 이번 실행의 by-id 상세 GET 두 번이
모두 그 id 를 `EXPIRED`·미발동으로 이름하고, CLOSED 목록이 이름한다면 같은 말을 하며, 같은 심볼에 살아 있는 조건주문·호가에 걸린 일반
주문·발동 흔적이 없고, 15 s 안에 두 번 읽은 결과가 같은 것. 목록이 이미 축출했어도 by-id 가 이름하면 통과한다. **by-id 가 404 면 목록이
무엇을 말하든 통과하지 않는다.**

**사람 결정과의 관계.** P1-1 은 H1~H5(2026-10-22 확정)보다 **엄격한 쪽**으로만 움직인다 — 수락 갈래 하나(by-id silent + CLOSED)를 거절로
바꿀 뿐 새 수락 갈래를 만들지 않는다. 그래서 새 사람 결정을 만들지 않는다(안전 불변식 6 의 「보수 방향만」과 같은 원리). 333430 은 영향 없음 —
0.6 영수증 ① 이 by-id 200 `EXPIRED` 다.

**이 가드가 거부하는 정상 입력(실제로 만료됐는데도 닫지 못하는 경우) — 의도된 비용:**

1. by-id 가 그 id 를 404 로 답함(목록이 이름하든 안 하든). 0.6 시점의 333430 은 해당 없음(by-id 200). H3.
2. 같은 심볼에 다른 OPEN 조건주문이 있음 — 엔진 브로커 보호 조건주문(a100 배선 뒤)·앱에서 건 수동 조건주문 포함(Q4 와 같은 비용).
   333430 은 엔진 adoption 종목이다(`ma-runbook.md:32`). H4.
3. 같은 심볼에 OPEN 일반 주문이 있음(엔진 청산 지정가 등). H4.
4. OCO·OTO artifact, 계좌가 둘 이상인 자격(a121 Q5), by-id 429·5xx·전송 오류(사람이 나중에 다시 돌림).
5. CLOSED 행 중 하나라도 어휘 밖 status(같은 심볼의 다른 종료 조건주문 포함) — `closed-status`.

## D3. 거부한 가드 입력 — 근거가 아닌 것

| 입력 | 거부 이유 |
| --- | --- |
| 정리 DELETE 404 (몇 번이든) | 같은 404 가 정정(살아 있는 후속 — M19·M40 `conditional-order-not-found`), 취소(M20·M41 「식별자가 더 이상 읽히지 않는다」), 축출을 구분 없이 말한다. a121 spec 「DELETE 404 remains nonterminal」 유지 |
| by-id GET 404 (단독이든 CLOSED 행과 함께든) | 위와 같은 이유. `silent` 로만 분류하고 진술로 세지 않는다. CLOSED 가 대상을 이름하는데 상세가 404 면 모순 → `expiry-sources-disagree`(P1-1) |
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

사전 검사(기록 엄격 해독 → 후보 → 계좌) → 정적 거절(Q3 상수 · **만료 어휘 확정** — Q1 은 부르지 않음) → 승인(모드·by-id 2회 표시) →
종목 대조 ① → **창 시작** → by-id ① → 목록 두 번(`readReconcileSet` — 조건 OPEN → 일반 OPEN → CLOSED, a121 고정 순서) → by-id ② →
종목 대조 ② → 신선도 → 목록 multiset · by-id 분류·필드 일치 → 행 검사 ×2 → **A5**(OPEN 조건 → 일반 OPEN → 모든 CLOSED 행 발동 흔적) →
A3·A4 진술 판정 → A7 체인 → A6 CLOSED 허용 목록(마지막) → 잠금·재해독·지문 → 줄 직렬화 → 쓰기 직전 신선도 → 한 번의 쓰기·fsync·read-back.

- by-id 는 계좌 범위 GET 이다(`ConditionalOrder` → `getAcct`). 그래서 **승인과 계좌 결속 뒤**에만 돈다(P2-1 — spec ADDED 문장이 by-id 를
  이름으로 포함). 승인 거절이면 by-id 호출 0.
- cmd 래퍼 `reconcileReads` 의 by-id 메서드는 기존 두 계좌 범위 메서드처럼 `ensureIdentity` 뒤에만 위임한다(P2-4,
  `cmd/tossctl/verify_reconcile.go:228-240` 와 같은 모양). 만료 모드에서는 by-id ① 이 첫 계좌 범위 호출이므로 Accounts 재조회가 그보다 먼저다.
- `--expired` 는 같은 cobra 명령의 플래그다 — 별도 하위 명령을 만들지 않는다(`mutating: "true"` 상속, `runVerifyReconcile` 한 경로).

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
- **감사·증명 경계(P2-2 — a121 「reconciled absence preserves audit and attestation boundaries」의 짝):** 줄에 `calls` 를 싣지 않는다 —
  by-id 200 은 `GET /conditional-orders/{id}` 성공 응답이므로 Call 로 남기면 `SucceededEndpoints`(`cmd/tossctl/soak.go:438` 소비)가
  endpoint 증거로 셀 수 있다. 근거는 버전·도메인 태그 지문으로만 싣는다. 멱등 — 이미 닫힌 artifact 는 후보에서 빠져 재실행은 `no-candidate`.
  단계 판정·취소·체결·성공 endpoint·soak 증명·capability attestation·엔진 interlock 충족으로 세지 않는다. 식별자는 Q2 결정 (a) 대로
  대상 artifact 줄이 이미 실은 키만 다시 쓴다 — 관측에 다른 브로커 id(다른 CLOSED 행의 id 등)를 싣지 않는다.

## D6. rate budget (안전 불변식 §0-4)

`--expired` 실행 1회 = a121 읽기(Accounts 2 · 종목 2 · 조건 목록 OPEN/CLOSED × 페이지 × 2 · 일반 OPEN × 페이지 × 2) + **by-id 2**.
by-id 는 목록과 같은 `CONDITIONAL_ORDER_HISTORY` 그룹(OpenAPI `getConditionalOrder`). 실행은 기존 `acquireVerifyRateBudget` lease 안에서만
돈다. 자동 재시도 없음(429 는 거절).

## D7. Q1 을 만료 모드에 적용하지 않는 이유

Q1 은 「발동한 조건주문이 CLOSED 에서 축출돼 부재가 발동을 숨길 수 있다」를 막는다(a121 G1-4). 만료 모드는 부재에서 추론하지 않는다 —
**by-id 상세 두 번이 대상 자신의 종결을 양성으로 읽어야 한다**(A3, P1-1 수리 뒤). 정정은 옛 id 를 **즉시 404** 로 만든다(M19 US · M40 KR),
그러니 by-id 200 `EXPIRED` 는 그 id 가 정정으로 대체되지 않았다는 측정과 맞는다. 이 논증은 by-id statement 를 요구하기 때문에 수락 갈래
전부에 성립한다(첫 판은 by-id silent + CLOSED 수락을 허용해서 이 논증이 그 갈래에서는 거꾸로 정정 쪽 증거였다 — freeze 리뷰 P1-1).
생존 판정은 Q1 이 아니라 A5 가 맡는다(D2 위계). 잔여는 R2.

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
  freeze 리뷰 P2-3 수리로 by-id 와 CLOSED 행 **둘 다** `first.status` 를 만료 leg 어휘로 검사하므로, leg 가 `ORDERED`/`COMPLETED` 인 경우는 거절된다 —
  남는 것은 top-level·leg 가 모두 `EXPIRED` 인데 child 가 체결된 계약 밖 응답뿐.
- R2. 정정된 옛 id 가 장기적으로 `EXPIRED` 로 다시 보일 가능성 — M19·M40 은 「즉시 404」만 쟀다. A5·A7 이 부분적으로 덮는다.
- R3. 성공 응답 requestId — **1.3 확정(2026-10-22, 구현 로트)**: 노출하지 않는다. `Client.send` → `unwrapAndDecode`(`internal/official/client.go:213-227`)가
  봉투에서 `result` 만 풀고 나머지(요청 id 포함)를 버린다. 그래서 by-id statement 는 `result` 원문 바이트의 지문(`sha256:` + 16바이트 절단, `DigestBytes` 와
  같은 규칙)을 관측에 적는다. silent(404) 는 이 change 의 증거 규칙(A3)상 줄을 쓰지 않으므로 본문 requestId 는 거절 문구에만 남긴다.
- R4. a164 는 `--include-trigger` 를 풀지 않는다(D9).
- R5. CodeGraphContext 색인이 이 체크아웃을 못 본다(`analysis/code-context/cgc-find.txt` — a121 심볼 0건, 옛 경로) — advisory 미반영.
- R6. 감사 사실(08-13 EXPIRED, 404 네 번)은 이 로트에서 재현하지 않았다(proposal 「미검증」).

## FLM/BTM 대상(구현 로트 1절, High-risk 인접 — 면제 없음)

편집할 기존 함수: `verifylive.Reconcile` · `checkReconcileAbsence`(보조 함수 추출) · `official.DecodeReconcileConditionalPage`(행에 `FirstStatus`
대입 — P2-3; 부재 모드는 그 필드를 읽지 않음) · cmd `reconcileReads`(새 by-id 메서드 — 새 함수지만 `ensureIdentity` 경유 P2-4) · `Artifact.terminal` · `Artifact.MarshalJSON` ·
`encodeReconcileLine` · `ReconciledArtifacts` · `writeReconciled`(report.go) · `ReconcileApproval.WriteText` · cmd `newVerifyReconcileCmd` ·
`runVerifyReconcile` · `verifyReconcileApprove`(변수 함수 리터럴 — logic-map 이 못 잡으면 not-applicable 사유 + 행동 시험). 새 함수는 새 파일
(`internal/verifylive/reconcile_expired.go`, `internal/official/reconcile_detail_reads.go`).
