# a164 proposal-freeze 적대 리뷰 — 2026-10-22

- 리뷰어: 독립 적대 Eng(설계 비작성자). 범위: `636be0080` + `f4dc1559d` 의 proposal·design·spec delta·tasks·receipt.
- 기준 HEAD `f4dc1559d`. base `4cd7af19d` 이후 `internal/`·`cmd/` 변경은 `internal/filldetect/unknown_status_test.go` 하나뿐이라 change 의 AST 29 는 신선하다(`git diff --stat 4cd7af19d HEAD -- internal cmd`).
- 읽기 전용. tossctl·계좌 읽기·주문 0. 저장소 편집은 이 파일 하나.
- 필수 질문 「이 diff 의 어느 줄이 살아 있는 조건주문을 닫힘으로 적을 수 있나」에 대한 답은 아래 「핵심 판단」.

## 판정: **APPROVE-WITH-CHANGES**

P0 0 · P1 1 · P2 6 · P3 6. P1 은 구현 착수(1.1) 전에 design·spec·tasks 를 고쳐야 한다. P2 는 RED 로트(2.x·4.x) 안에서 닫아도 된다.

## 핵심 판단 — 살아 있는 조건주문을 놓치는 경로

생존 판정을 실제로 짊어지는 것은 증거 규칙 A2~A4 가 아니라 **A5** 다.

- 같은 계좌·같은 심볼의 OPEN 조건주문 그룹 전체가 비어 있으면, 그 심볼에 살아 있는 조건주문은 없다. OpenAPI `getConditionalOrders`
  는 OPEN = `{WATCHING, PAUSED, ORDERING, ORDERED}` 이고 「다른 채널(토스증권 앱 등)에서 등록한 조건주문도 함께 반환」이다
  (`docs/migration/openapi.latest.json` paths `/api/v1/conditional-orders` get).
- 정정은 symbol 을 바꾸지 못한다(`ConditionalOrderModifyRequest` 에 `symbol` 없음, 같은 파일 components). 그래서 정정 후속은 같은 심볼
  OPEN 에 보이고(→ `open-conditional`), 발동한 후속은 CLOSED `COMPLETED`/`triggeredOrderId`(→ `closed-fired-trace`) 또는 OPEN 일반
  주문(→ `open-plain-order`)으로 보인다.
- 이 셋은 a121 의 기존 읽기·판정을 그대로 쓴다: `readReconcileSet`(`internal/verifylive/reconcile_check.go:184-197`)이 OPEN 조건 →
  일반 OPEN(`ReconcileOpenOrdersPage`, `internal/official/reconcile_reads.go:204-210`) → CLOSED 를 페이지 끝까지 읽고(`reconcilePager`
  반복 커서·빈 커서·10 페이지 상한 `steps.go:61`), 오류는 `read-error` 로 거절(`reconcile_check.go:252-254`, `:280-282`).
  일반 OPEN 은 계약상 `limit`·`cursor` 무시 전량 반환이라, 서버가 hasNext 를 잘못 켜도 같은 페이지 반복 → `duplicate-row`/`repeated-cursor` 로 닫힌다.

결론: A5 가 그대로 서 있는 한, by-id·CLOSED 증거가 어떤 조합이든 **살아 있는** 조건주문을 닫힘으로 적는 입력 순서는 찾지 못했다.
A2~A4 가 막는 것은 「만료」라는 **이름의 정확도**(발동·정정을 만료로 부르지 않기)다. 남는 생존 위험은 a121 과 같은 잔여(심볼 필터를
무시한 빈 목록 — 양성 대조가 종목 endpoint 만 잰다)뿐이다. 이 구도 때문에 아래 P1·P2 는 「A5 를 약하게 만들 수 있는 구현 자유도」와
「정확도 가드의 구멍」을 겨눈다.

## P1

### P1-1. by-id 가 404(silent)인데 CLOSED 가 대상을 이름하면 **수락**한다 — 정정 서명과 같은 모순을 근거로 쓴다

- 근거: design D2 A3 「by-id 가 404 여도 CLOSED 목록이 이름하면 통과」(design.md D2 「이 가드가 통과시키는 것」), spec delta
  Scenario 「only the CLOSED list names the artifact」 → `exactly one reconciled expired line is appended`.
- 왜 틀렸나: 계약은 상세 GET 이 「진행 중 + 종료된 조건주문을 모두 조회」한다(`getConditionalOrder` description). 그러니 목록이 이름하는
  id 를 상세가 404 로 답하는 것은 브로커 자기모순이다. 그리고 「옛 id 상세 404」는 정확히 M19·M40 이 잰 **정정**의 서명이다
  (`record.go` Artifact.ChainID 주석 「A conditional modify issues a new identifier and invalidates the old one immediately」).
  design D7 의 안전 논증(「옛 id 가 `EXPIRED` 로 읽힌다는 것은 그 id 가 정정으로 대체되지 않았다는 측정과 맞는다」)은 by-id statement
  경로에만 성립하고, 이 경로에서는 거꾸로 정정 쪽 증거다. 지금은 A5 가 후속을 잡아 생존 사고는 아니지만(위 「핵심 판단」), 「모순이면
  거절」이라는 fail-closed 원칙을 설계가 스스로 깨고 그 근거를 D7 로 잘못 댄다.
- 비용: 0. 0.6 영수증 ① 은 by-id 200 `EXPIRED`(`analysis/receipts/expiry-probe-2026-10-22.json` `get_by_id`)라 333430 은 고친 규칙으로도 닫힌다.
- 고침:
  1. A3 를 「by-id 두 응답이 **둘 다 statement**」로 바꾼다. CLOSED 행은 있으면 A4 일관성 검사만 받고, 없어도(축출 — Q1 미측정) 무방.
  2. by-id silent ×2 + CLOSED 가 대상을 이름 → 새 거절 `expiry-sources-disagree`. by-id silent ×2 + CLOSED 부재 → `expiry-evidence-absent`(그대로).
  3. spec delta 의 Scenario 「only the CLOSED list names the artifact」를 거절 시나리오로 뒤집고, D2 「통과시키는 것」·D7 문장을 고친다.
  4. tasks 2.2 의 「CLOSED 만(by-id silent)」 수락 갈래를 거절 갈래로 옮기고, 4.x 에 「silent+CLOSED 수락으로 되돌림」 변이를 더한다.

## P2

### P2-1. spec 이 by-id 읽기를 승인·계좌 결속 **앞**에서 금지하지 않는다

- ADDED requirement: 「human approval before any order or conditional-order **list** read」. by-id 는 목록이 아니라 계좌 범위 상세 GET 이다
  (`ConditionalOrder` → `getAcct`, `internal/official/conditional_reads.go` `ConditionalOrder`). 설계 D4 는 승인 뒤에 두지만 spec 문장대로면
  승인 전 by-id 구현도 적합하다. 기존 「binds account, profile and market before reading」(`openspec/specs/order-execution/spec.md:799-806`)의
  괄호(「account-list read and the instrument read … are the only official reads permitted earlier」)가 결속 쪽은 막지만 승인 쪽은 안 막는다.
- 고침: ADDED 문장을 「before any account-scoped order or conditional-order read, **including the by-id reads**」로. tasks 2.4 에
  「승인 거절 시 by-id 호출 0」 단언, 4.x 에 「by-id ① 을 승인 앞으로」 변이.

### P2-2. ADDED requirement 에 a121 「audit and attestation boundaries」의 짝이 없다

- 정본 `spec.md:860-883` 은 `reconciled-absent` 에 대해 Calls 미탑재(「so the reads it made cannot become successful-endpoint evidence」)·
  멱등·원 실패 보존·soak/attestation/engine-interlock 불산입·Q2 식별자 규칙을 요구한다. a164 ADDED 는 「no account number in clear and no
  raw response body」만 적는다. design D5 는 「Calls 없음(a121 과 같음)」이라 쓰지만 spec·tasks 에 시험이 없다.
- 왜 중요한가: by-id 200 은 `GET /conditional-orders/{id}` 성공 응답이다. 구현이 Runner 관례대로 Call 을 남기면 `SucceededEndpoints`
  (`cmd/tossctl/soak.go:429-437` 소비)가 그것을 endpoint 증거로 셀 수 있다 — 엔진 interlock 경계.
- 고침: 정본 requirement 를 MODIFIED 로 넓혀 두 종결(`reconciled-absent`·`reconciled expired`)을 함께 덮거나, ADDED 에 같은 문장과
  Scenario 「attestation remains unchanged」를 복제. tasks 2.5 에 「줄의 `calls` 비어 있음 · `SucceededEndpoints` 결과 불변 · 같은 artifact
  재실행 = `no-candidate`」.

### P2-3. CLOSED 행에는 leg status 검사가 없다(A4 비대칭)

- A4 는 `first.status` 만료 어휘를 **by-id statement 에만** 요구한다. 목록 행 타입 `ReconcileConditionalRow` 에는 `FirstStatus` 가 없다
  (`internal/official/reconcile_reads.go:32-39`; 해독 `:145-148` 은 `First.TriggeredOrderID` 만 옮김).
- 결과: top-level `EXPIRED` · `first.status` `ORDERED`/`COMPLETED` · `triggeredOrderId` 키 없음(C2 정정으로 「없음」) 행이 CLOSED 진술로
  통과한다(design R1 의 감사 정확도 오류). P1-1 을 고치면 by-id statement 가 필수라 영향이 줄지만, CLOSED 행은 여전히 S 의 원소다.
- 고침: 만료 모드 전용 새 해독(새 파일 — a121 해독 불편집)으로 CLOSED 행의 `first.status` 를 노출하고, **대상 id 행**에도 A4 leg 어휘를 요구.
  거절 코드 `expiry-target-status`.

### P2-4. 새 by-id 리더 메서드가 `ensureIdentity` 를 거쳐야 함이 어디에도 없다

- cmd 래퍼는 계좌 범위 GET 둘만 `ensureIdentity` 로 감싼다(`cmd/tossctl/verify_reconcile.go:228-240`; 종목 읽기 `:242-244` 는 계좌 범위가 아님).
  D4 순서에서 만료 모드의 **첫** 계좌 범위 읽기는 by-id ① 이다. 새 메서드가 래퍼를 건너뛰면 신원 재확인(codex CG-6)이 by-id ① 뒤로 밀린다.
  (지금 구조에서는 by-id ①·② 분류 불일치로 결국 거절되지만, 그 우연에 기대는 것이다.)
- 고침: tasks 3.x 에 「`reconcileReads` 의 by-id 메서드는 `ensureIdentity` 뒤에만 위임」, 2.6 봉인 시험에 「by-id 가 첫 계좌 범위 호출일 때
  Accounts 재조회가 먼저」 단언.

### P2-5. 변이 목록이 거절 이유 하나당 변이 하나를 채우지 못한다

4.1~4.10 에 없는 것(각각 시험이 잡아야 함):

1. by-id `first.status` 어휘 검사 삭제(A4 leg).
2. silent 분류를 「404 면 전부」로 넓힘(본문 `conditional-order-not-found` 비교 삭제) · 본문 해독 불능 404 를 silent 로.
3. 429·5xx·`ErrAuth` 를 silent 로 접음(`internal/official/errors.go` `classifyStatus` 의 sentinel 들).
4. 두 by-id statement 비교의 **필드별** 삭제(status · first.status · triggeredOrderId · second) — 4.4 는 비교 통째 삭제만 본다.
5. 만료 모드가 `readReconcileSet` 대신 일반 OPEN 을 빼먹은 읽기 집합을 씀(판정이 아니라 **읽기** 생략 — 4.5 판정 삭제로는 안 죽는다).
   시험은 리더가 받은 `ReconcileOpenOrdersPage` 호출이 두 번임을 단언.
6. 만료 모드에서 A6 CLOSED 어휘 검사 자체 삭제(4.10 은 기본값 확장만 본다).
7. `ReconciledArtifacts` 필터(`internal/verifylive/projection_reconciled.go:26` `a.ReconciledAbsent`)에 넷째 종결 미반영 — 닫힌 artifact 가
   outstanding 에서도 reconciled 목록에서도 사라져 status 에서 안 보임.
8. by-id ① 을 승인 앞으로(P2-1), by-id 메서드의 `ensureIdentity` 생략(P2-4), silent+CLOSED 수락(P1-1).

### P2-6. design 이 A5 를 「안전 등뼈」로 적지 않는다

- D7 은 by-id 진술을 주 근거로, A5·A7 을 「그래도 유지」로 적는다. 위 「핵심 판단」대로 실제 생존 가드는 A5 이고 A2~A4 는 정확도 가드다.
  이 위계가 문서에 없으면 이후 리뷰·리팩터가 「증거가 강하니 A5 는 비용」으로 완화할 여지가 생긴다(H4 가 이미 그 비용을 사람 결정으로 올렸다).
- 고침: D2 머리에 「A5 는 생존 판정, A2~A4 는 종결 이름 판정 — A5 는 어떤 증거 조합에서도 생략·완화 불가」 한 문단. tasks 2.3 에
  「by-id 200 `EXPIRED` ×2 + CLOSED `EXPIRED` 인데 OPEN 조건 1건 → `open-conditional`」 명시 사례.

## P3

- **P3-1. 신선도 창 시작점.** D4 순서상 by-id ① 은 첫 목록 읽기 **앞**인데 `start := now()` 는 첫 목록 직전이다(`reconcile.go:231`).
  by-id ① 이 창 밖이다. 고침: 만료 모드는 by-id ① 직전을 `start` 로(A1 「첫 읽기 시작」 문장과 일치).
- **P3-2. H2 서술의 조건 누락.** 내가 만든 AST(`validateM0TriggerMode`, `source_sha256` 포함, 스크래치 — 아래)에서 B5(`:417`
  `ok && Kind != "pending-create"`)·B7(`:419` `!ok`) 둘 다 거짓인 경우, 즉 **pending-create 체크포인트가 있으면** B8
  `M0ExactPrerequisites`(`:424`)를 건너뛰고 B9 `Outstanding`(`:428`)이 관문이다. proposal 「비목표」의 「cmd 진입은 전제 판정에서 먼저
  멈추고」는 무조건형이라 과장이다(design D9 는 「B7(`!ok`) 안에서」로 정확). 고침: proposal 문장에 「pending-create 체크포인트가 없을 때」.
  결론(「a164 만으로 `--include-trigger` 는 안 풀림」)은 감사가 보고한 전제 fail 셋이 사실이면 유지된다 — 기록 미열람이라 미검증.
- **P3-3. by-id `type` 미검사.** OCO/OTO 는 `second` 로 거르지만 `type == "SINGLE"` 은 한 줄 비용이다(계약 enum `SINGLE/OCO/OTO`).
  `second` 없는 OTO 같은 계약 이탈을 이름으로 거절.
- **P3-4. RED 의 컴파일 전략.** 새 필드 `ReconciledExpired`·새 거절 코드·새 리더 메서드를 부르는 RED 시험은 편집 전 컴파일되지 않는다.
  a121 처럼 골격 커밋(`reconcile_reads.go` 머리 주석 「골격」)으로 심볼을 먼저 세우고 RED 가 단언에서 실패하게 할지 tasks 2 머리에 적을 것.
- **P3-5. 거절 코드 우선순위.** A5 보조 함수를 부재 모드와 공유하면, 발동 흔적 루프(`reconcile_check.go:390-394`)는 대상 행을 빼지 않는다
  (부재 모드에서는 대상 행이 `:380-389` 에서 먼저 반환하므로 문제 없었다). 만료 모드에서 대상 행이 `COMPLETED` 면 `closed-fired-trace` 와
  `expiry-target-fired` 중 어느 것이 나는지 design 에 고정하고 2.3 이 그 코드를 단언할 것(D2 A5 「다른 id」를 코드로).
- **P3-6. 승인 화면 문구.** `WriteText` 머리 「공식 목록 조회 전용」(`reconcile_check.go:415`)과 질문 「reconciled absent 로 기록할까?」
  (`verify_reconcile.go:260`)는 만료 모드에서 by-id 읽기를 말하지 않는다. tasks 2.7 이 모드 표시를 요구하므로 by-id 2회도 화면에 넣을 것.

## 질문별 확인

1. **증거 규칙(D2/D3).** 생존: A5 로 막힘(핵심 판단). 정정 후 옛 id: 상세 200 `EXPIRED` + 후속 생존이면 OPEN 에 보여 거절. 체인: A7 +
   후속이 기록에 있으면 후보 2 → `multiple-candidates`(`reconcile_check.go:50-52`). OCO: `HasSecond` → `oco-row`. 같은 id 재사용: 단조 규칙
   (`record.go:712` 「종결 뒤 비-종결 줄은 되살리지 않음」)이 재사용 id 의 새 생존을 숨길 수 있으나 브로커 id 는 opaque 고유 — a121 과 같은
   잔여, 비차단. status 어휘 drift: 미지 → `expiry-target-status`/`closed-status`. 페이지: a121 pager. 목록-상세 불일치: P1-1. 시계: `time.Now`
   단조 성분으로 Sub — 창 시작만 P3-1. 부분 읽기·429·5xx: 자동 재시도 없음 → 거절(D6, C5).
2. **C2(키 없음 = null).** 계약 일치: `ConditionalOrderCondition.required = [type, status, orderSide]`, `triggeredOrderId` 형 `[string, null]`;
   `ConditionalOrderDetailResponse` 의 `second` 는 `nullable` 이고 required 아님(`docs/migration/openapi.latest.json` components). 도메인 어댑터는
   status 를 그대로 옮김(`conditional_reads.go:226` leg · `:240` 최상위), 도메인은 `TriggeredOrderID` `omitempty`·`Second` 포인터 `omitempty`
   (`internal/domain/models.go:784` · `:798`) — C1·C2 서술 정확. 구멍: 「발동했는데 키가 빠진」 응답은 top-level 이 `ORDERED`/`COMPLETED` 라
   A4 가 거절하고, child 가 걸려 있으면 A5 일반 OPEN 이 거절한다. 남는 것은 CLOSED 행 leg 미검사(P2-3) — 정확도 문제, 생존 문제 아님.
3. **H4/C4 일반 주문.** 계획은 a121 `readReconcileSet` 고정 순서를 그대로 쓴다(D4) — 그 함수는 `ReconcileOpenOrdersPage` 를 끝까지 읽고 오류에
   `read-error` 로 거절한다(위 줄). 단 design 이 「`readReconcileSet` 재사용」을 이름으로 못 박지 않았고 읽기 생략 변이가 없다 → P2-5 ⑤.
   C4(0.6 의 `orders list` 는 WTS pending 이라 공식 판정이 아님)는 코드와 맞는 정직한 서술이다.
4. **소비자.** `terminal()` 한 술어(`record.go:736`)를 `outstandingLines`(`:712`·`:722`)·`ReconciledArtifacts`(`projection_reconciled.go:18`)·
   `runner.go:1052` 가 쓰고, `Outstanding` 호출자(CodeGraph 1.6.0 `codegraph callers Outstanding` 20건: `AbortTargets`·`Abort`·`sweepStep`·시험 등)와
   cmd `validateM0TriggerMode`·`verifylive.New`(`runner.go:346`)·console `readVerify`(`internal/console/data.go:369-384`)·`subjectLost`
   (`redo.go`)·`liveConditional`(`steps.go`)이 그 투영을 거친다 — 넷째 종결은 술어 한 줄로 전부에 닿는다. 술어를 우회하는 것: `m0ManualReconcileIDs`
   (`m0_manual.go:11` B4 `!Cancelled && !Filled`, AST 스크래치) — 트리거 단계 줄의 artifact 를 수동 보유로 **더할** 뿐이라 안전 방향(닫힘을
   만들지 않음), 비차단. `ReconciledArtifacts` 필터는 넷째 종결을 넣어야 함(P2-5 ⑦). scrub-account-records 는 `map[string]any` 로 줄을 다시 써서
   (`internal/redact/record.go:116-159`) 미지 키를 보존한다 — 새 키 유실 없음. 구 바이너리: `decodeEntry`(`record.go:597-607`) unknown 키 버림 →
   비-종결 → outstanding(안전 방향, 기존 시험 `TestAnOlderBinaryStillSeesTheReconciledArtifactAsOutstandingAndHeld` 가 본보기). 형식 버전 불변이라
   `errUnknownFormat` 경로 없음.
5. **mutating.** `--expired` 는 같은 cobra 명령의 플래그라 명령 단위 `Annotations{"mutating": "true"}`(`verify_reconcile.go:60`)를 상속하고,
   `runVerifyReconcile` 한 경로(`:73-110`: preflight → 실행 flock → rate lease → 좁은 생성자 → `Reconcile`(승인 포함))를 지난다. 별도 하위 명령을
   만들지 말 것을 tasks 2.7 에 명시 권고.
6. **spec·tasks 품질.** MODIFIED requirement 는 정본과 Q6 문장 하나 + 시나리오 하나만 다르다(스크래치 diff — 문장 외 누락 0). 요구는 시험 가능.
   빈틈: P1-1(시나리오 뒤집기)·P2-1·P2-2(spec 문장)·P2-5(변이)·P3-4(RED 컴파일).
7. **H2.** 확인 — cmd 는 B8(`:424`) 이 B9(`:428`)보다 앞, `New` 는 B9(`runner.go:346`)가 `M0ExactPrerequisites`(`:358` 부근)보다 앞. proposal
   「비목표」와 design D9·R4 가 「a164 만으로 안 풀림」을 정직하게 적는다. 조건 누락만 P3-2.

## 리뷰가 쓴 증거

- change 산출물: `analysis/ast/enumerate-output.md`(B 번호), `analysis/receipts/expiry-probe-2026-10-22.json`, `analysis/codegraph-affected-narrow.txt`.
- 이 리뷰가 새로 뽑은 AST(커밋 안 함, 스크래치 `…/scratchpad/a164-review/`): `validateM0TriggerMode`(B1~B9, B8 `:424`·B9 `:428`),
  `m0ManualReconcileIDs`(B1~B8, B4 `:11`), `ReconciledArtifacts`(B1~B6, B6 `:26`) — `go build ./tools/logic-map` 후 `-file … -func …`.
- Function Logic Map: not-applicable — 리뷰 문서는 편집을 하지 않는다. 함수 분기를 근거로 쓴 곳은 위 AST 와 change AST 의 B 번호에 줄을 대조했다.
  구현 로트 1.1 FLM/BTM 의무는 그대로다.
- CodeGraph: 1.6.0, `codegraph callers Outstanding`(20) · `codegraph callers ReconciledArtifacts`(3: `BuildReport`·`BuildProgress`·시험).
  CodeGraphContext 는 change R5 대로 이 체크아웃을 못 봄 — 안 봤음.

## 검증하지 못한 것

- 브로커 실제 동작: 정정된 옛 id 가 시간이 지나 상세/CLOSED 에 `EXPIRED` 로 다시 보이는지(R2), CLOSED 보존(Q1), by-id 404 본문 code 문자열의
  현재 값 — 계좌 읽기 금지라 미측정.
- 실기록(`capability-verify.jsonl`)의 333430 artifact 의 `ChainID`·`HeldUntil`·M0 체크포인트, 감사가 말한 전제 fail 셋 — 파일 미열람.
- 시험·게이트·`openspec validate` 미실행(문서 리뷰 범위). 0.6 영수증의 「OPEN 일반 주문 0」은 공식 읽기가 아니다(C4) — 6.2 실행이 처음 판정.
