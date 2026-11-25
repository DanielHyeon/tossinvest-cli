# a168 design — 검증 대사 읽기가 모호한 브로커 응답과 200 아닌 성공 상태를 거절한다

> 상태: **동결(2026-11-25, Manager 결정 §10 1~4 확정)** — 동결 후보 작성 — base `e57d06cd6`(`base-commit.txt`), worktree `~/tossos-work/a168-impl`.
> 사람 결정(proposal 「사람 결정」): Q1 = (a) 대사 리더만 HTTP 200 · Q2 = R-a·R-b 범위 밖 · Q3 = 충족.
> 근거 산출물: AST `analysis/ast/`(30개, `generate.py` → `enumerate-output.md`, STALE 0), 실행 측정
> `analysis/measure/decode_ambiguity.go` → `decode_ambiguity-output.txt`(httptest, 실 브로커 호출 0), 정상 입력 census
> `analysis/measure/normal_input_census.py` → `normal_input_census-output.txt`, CodeGraph 1.6.0 `analysis/code-context/`.
> 아래에서 함수 내부 분기를 근거로 쓰는 문장은 전부 `enumerate-output.md` 의 분기 id(Bn)·줄을 인용한다.

## §0 한 줄 요지

대사 읽기 네 개(`Client.ReconcileConditionalOrdersPage` · `ReconcileOpenOrdersPage` · `ReconcileInstrument` ·
`ReconcileConditionalOrder`)가 **기존 해독이 끝난 뒤**, 그 결과를 정한 마지막 HTTP 시도의 (상태 코드, 본문 원문)을 관측자로
받아 ① 상태 코드가 정확히 200 이고 ② 본문의 어느 객체에도 `encoding/json` 이 한 필드로 접을 수 있는 두 키가 없을 때만 결과를
돌려준다. 공용 `Client.send` · `unwrapAndDecode` · `token.go` 는 한 줄도 바꾸지 않는다.

## §1 proposal 주장 대조 (AST · 실행 측정)

| # | proposal 주장 | 판정 | 근거 |
| --- | --- | --- | --- |
| P1 | `reconcileObject` 는 `reconcile_detail_reads.go:96` | 맞음 | AST `reconcileObject` 96–127 |
| P2 | `DecodeReconcileConditionalPage` `:136`, 행 해독 `json.Unmarshal(raw,&o)` `:144` | 맞음 | AST B3 줄 144 |
| P3 | `apiConditionalOrder` `conditional_reads.go:28` | 맞음 | HEAD 28 (모든 스칼라 필드 `string`, `Second *apiConditionalCondition`) |
| P4 | `"first":{"triggeredOrderId":"CHILD","triggeredOrderId":null}` 는 **미발동으로 읽힌다** | **틀림** | 측정 W1: `triggered="ORD-9"` 로 남음. `encoding/json` 은 string 필드에 대한 `null` 을 무시(덮지 않음)한다. 숨김은 **뒤 값이 빈 문자열**일 때 생긴다 — W1b(`"ORD-9"` 다음 `""`) → `triggered=""`. 같은 성질: W2(접힌 키 + null)는 숨기지 못하고 W2b(접힌 키 + `""`)는 숨긴다 |
| P5 | 다른 id 의 발동 흔적을 중복 키로 숨기면 A5 셋째와 부재 모드 발동 흔적 판정이 못 본다(두 모드 공통) | **만료 모드만 맞음**(리뷰 1회차 Claude P2) | 해독 층: W1b·W2b·W3·W4 ACCEPT — `checkReconcileNoFiredTrace`(`reconcile_check.go:417`)는 해독된 행만 봄. 그러나 **부재 모드는 CLOSED 행이 하나라도 있으면 거절**한다: `reconcileClosedStatusAllowlist` 가 비어 있고(`reconcile.go:418`) `checkReconcileClosedStatus` B2(줄 429, `!allowlist[status]`)가 부재 갈래 마지막(`Reconcile` 줄 312)에 돈다. 그래서 부재 모드의 실제 구멍은 행 **통째** 숨김 — 봉투 `result` 중복(E1·E2) · 컬렉션 중복(R1·R4)으로 OPEN 행이나 CLOSED 행을 지우는 모양이다(`checkReconcileAbsence` B1 줄 385 은 OPEN 행 수만 봄). 만료 모드는 `checkReconcileExpiry` B2(줄 83, 발동 흔적)와 B14(줄 121, `EXPIRED` 어휘)만 CLOSED 행을 보므로 다른 id 의 `EXPIRED` 행에 W1b 모양 발동 흔적 숨김이 열려 있다. 측정 W1b 의 행은 `CANCELED` 라 두 모드 모두 이미 `closed-status` 로 거절된다 |
| P6 | `decodeReconcileEnvelope` `:78`, map 해독 `:83` | 대체로 맞음 | 함수 78–132, map **선언** 83 · 해독 호출은 B2 줄 84. R1(컬렉션 키 중복)·R2(`hasNext` 중복) ACCEPT — 확정 결함 |
| P7 | `unwrapAndDecode` `client.go:213` | **줄 낡음** | 현재 228–243(a166 착지로 15줄 밀림). 봉투 중복 `result` 는 확정 결함: E1·E2(`Result` 접힘) ACCEPT rows=0, by-id 도 D1·D2 ACCEPT |
| P8 | `Client.send` `client.go:320`, 2xx 판정 `:362`, 401 루프 `:344` | **줄 낡음** | 현재 send 335–381, 2xx 판정 B9 줄 377, 401 루프 B4 줄 359. 성질은 맞음 — S1~S7(201·202·203·206·299) 전부 ACCEPT, S8(204 빈 본문)은 해독 실패로 ErrServer |
| P9 | 「trailing data after a list page」 를 거절해야 한다(spec 초안 둘째 시나리오) | **이미 거절됨** | E3·E3b·E3c(객체·`}`·`]` 꼬리) → ErrServer. `unwrapAndDecode` B2 의 `json.Unmarshal(body,&env)` 가 입력 전체를 값 하나로 요구한다. result 이하 층은 `json.RawMessage` 하나라 꼬리가 존재할 수 없다. 시나리오는 계약으로 남기되 RED 가 아니라 **특성 고정 시험**이다 |
| P10 | (암묵) `reconcileObject` 의 뒤따르는 데이터 검사가 꼬리를 막는다 | **부분만** | AST B8 `if dec.More()` 는 다음 바이트가 `}`·`]` 이면 false — 측정 T-direct: 꼬리 `}`·`]` ACCEPT, 꼬리 값 REFUSE. 생산 경로에서는 P9 때문에 도달 불가. 이 change 는 고치지 않음(§9 R3) |
| P11 | R-a 줄 `token.go:108` · `verify_reconcile.go:226`·`:238` | **줄 낡음** | `refresh` 110, `ensureIdentity` 237, `if r.verified` 238. 범위 밖(Q2) — 기록만 |
| P12 | R-b 줄 `reconcile.go:343`·`:359`·`:407`, 시험 `reconcile_codex_test.go:185` | 맞음 | HEAD 그대로. 범위 밖(Q2) |
| P13 | 영향 「후보 파일: reconcile_reads.go · reconcile_detail_reads.go」 | 맞음 + 새 파일 하나 | §7 |
| P14 | 겹침: 진행 중 a165 가 `send` 를 다룸 | 해소 | a165 archive `2026-10-29` |

추가 발견(proposal 에 없던 것, 측정):

- **접힌 키(대소문자 변형)가 기여한다.** 구조체 해독은 정확히 같은 키를 먼저 찾고 없으면 대소문자를 무시해 맞춘다(Go `encoding/json`
  필드 매칭 — `bytes.EqualFold` 와 같은 접기). 그래서 계약에 없는 키가 계약 키의 값을 덮는다: W7 `"symbol":"005930","Symbol":"000660"` →
  `symbol=000660`, E2 봉투 `Result`. map 해독 층(`decodeReconcileEnvelope` 의 result 객체, by-id 의 result·first)은 정확한 키만 읽으므로
  접힌 변형을 **무시**한다(R3·D4 ACCEPT 이나 값은 계약 키의 것).
- **by-id(a164)도 봉투 층은 열려 있다.** D1·D2 — `reconcileObject` 는 `unwrapAndDecode` 가 고른 result 만 본다.
- **종목 양성 대조**도 같은 성질: I1 `[{"symbol":"000660","symbol":"005930"}]` → echo `005930`.
- 행 안 `first` 중복(W5)은 구조체에 **병합**된다 — 앞 `first` 의 `triggeredOrderId` 와 뒤 `first` 의 `status` 가 섞인다.

## §2 Q1 (a) 는 `send` 를 건드리지 않고 된다

- 대사 리더는 지금 상태 코드를 못 본다: 네 리더는 `getAcct`/`get` → `getWithHeaders` → `send` 를 부르고(AST `Client.get` · `getAcct` ·
  `getWithHeaders`), `send` 는 2xx 면 `unwrapAndDecode` 결과만 돌려준다(AST send B9 줄 377 → 반환 380).
- **기존 통로가 있다 — 시도 관측자.** `doRequest`(203–222)는 성공 시 줄 220 에서 `observeAttempt(req.Context(), AttemptTrace{StatusCode,
  Body: 복사본})` 를 부른다(AST doRequest 호출 3개: 207·217·220). `send` 의 401 루프(B4)는 시도마다 `doRequest` 를 다시 부르고 마지막
  시도의 (code, body) 로 판정·해독하므로(줄 369 재대입 → 377·380), **관측된 마지막 시도 = 결과를 정한 시도**다. `send` 가 nil 을 돌려줬다면
  `doRequest` 가 적어도 한 번 줄 220 으로 관측을 냈다(B1~B3·B5~B7 의 오류 반환은 모두 non-nil).
- 선례 셋: `m0_reads.go:24-36`(관측자 부착, 체인 없음) · `strict_quote_reads.go:210`(`strictQuoteFetch`, 바깥 관측자 체인) ·
  `strict_minute_candles.go:247`(`strictMinuteFinalAttempt` — 「마지막 시도」 를 고르는 이유 주석).
- 그래서 리더 쪽에서 `WithAttemptObserver` 로 시도를 모으고 마지막 시도의 `StatusCode == 200` 을 요구하면 된다. **`send`·`doRequest`·
  `unwrapAndDecode`·`token.go` 편집 0.** 주문·정정·취소 경로는 같은 `send` 를 그대로 지나므로 동작 불변.
- 계약 근거: `docs/migration/openapi.latest.json` 의 네 GET 성공 응답은 **200 하나뿐**이다 — `/api/v1/conditional-orders` {200,400,422},
  `/{conditionalOrderId}` {200,400,404}, `/api/v1/orders` {200,400,401,404,429,500}, `/api/v1/stocks` {200,400,401,403,429,500}.
  정본 spec `order-execution` 「records a broker-confirmed expiry」 도 by-id statement 를 「HTTP 200 passing schema validation」 으로 적는다
  (a164 D3 와 같은 문장 — 이 change 뒤 코드가 그 문장과 같아진다).

## §3 결정

### D1. 성공 상태 계약 = 정확히 200 (대사 읽기 네 개)

- 판정 입력: 관측된 시도 목록의 **마지막** 원소. `StatusCode != 200` 이면 거절. 관측 0 이면 거절(도달 불가 방어 — §2).
- 거절 오류: `ErrReconcileSchema` 를 감싼 오류(새 sentinel 없음 — YAGNI). 문구는 상태 숫자만 담음(본문·계좌 없음).
- 404 silent 분류(by-id)는 `getAcct` 가 오류를 돌려주는 갈래라 이 판정에 닿지 않는다 — a164 분류 불변.

### D2. 모호한 본문 = 어느 깊이든 한 객체 안의 「접히는 두 키」

- 정의: 본문 원문(봉투 포함)의 모든 JSON 객체에서, 두 키가 `strings.EqualFold` 로 같으면(완전히 같은 키 포함) 거절. Go 의
  `encoding/json` 필드 매칭 접기가 `bytes.EqualFold` 와 같다는 것이 근거 — go1.26.5 `src/encoding/json/fold.go:15`
  「foldName … is identical to bytes.EqualFold(x, y)」, 그리고 §1 추가 발견 W7·E2 측정이 그 동작을 보임.
- 왜 모든 깊이·모든 키인가: 판정 입력(봉투 `result` · 목록 result 의 컬렉션·`hasNext`·`nextCursor` · 행의 id·symbol·market·status·
  `first`·`second` · `first.status`·`triggeredOrderId` · 주문 행 · 종목 항목 · by-id result·first)이 다섯 층에 흩어져 있고, 층별로 고르면
  「볼 곳을 고르는」 가드가 된다. 한 규칙을 한 자리(본문 전체)에 둔다. 비용: 판정에 안 쓰는 깊은 필드의 중복도 거절(W8) — §6 census 로 센다.
- 뒤따르는 데이터: D2 가 따로 보지 않는다. 본문은 이 판정에 오기 전에 `unwrapAndDecode` B2 가 값 하나임을 이미 요구했다(P9, E3*).
  같은 규칙을 두 번 두지 않는다(기억 「판정이 둘이면 반증이 죽는다」). 그 의존은 시험 R-T(§8)로 고정한다.
- 구현은 **새 해독기가 아니다**: a164 `reconcileObject`(토큰 단위 · 정확한 중복 거절)를 객체마다 그대로 부르고, 돌려받은 키 집합에
  `EqualFold` 짝 검사를 더하고, 값이 객체·배열이면 재귀한다. 배열은 `json.Unmarshal` 로 `[]json.RawMessage` 로만 쪼갠다.
  `reconcileObject` 본문은 바꾸지 않는다. 키 비교는 **해독된 토큰 문자열**로 한다(`\u0074riggeredOrderId` 는 해독하면 `triggeredOrderId` 와 같은 키라 `reconcileObject` 줄 115 가 이미 정확한 중복으로 거절하고, `\u0054riggeredOrderId` 는 접힘 짝이다).
- 왜 `strictMinuteCheckJSON`/`strictMinuteWalk`(`strict_minute_candles.go:489,503` — 한 번 훑기, 정확한 중복·깊이·꼬리)를 쓰지 않나:
  (1) 정확한 중복만 보고 접힘을 안 본다 — 측정된 덮어쓰기 W7·E2·W2b·W3b 가 남는다. (2) 깊이 상한 `strictMinuteMaxDepth` 는 a112 봉 계약의
  상수다 — 대사 판정이 남의 계약 상수에 묶인다. (3) 그 함수를 고치면 a112 리더의 판정이 바뀐다. 사용자 지시(「`reconcileObject` 의 토큰
  단위 규칙 재사용」)와도 맞다. 대가는 비용 — §9 R5.

### D3. 위치 — 기존 해독 **뒤**, 결과 반환 **앞**

각 리더의 순서: (1) 기존 입력 검사(빈 status·빈 id) → (2) 전송(`getAcct`/`get`, 관측자 부착) — 오류면 그대로 반환(by-id 는 silent 분류
먼저) → (3) **기존 해독**(`DecodeReconcile*` · 종목 항목 수) → (4) **새 판정 D1 → D2** → (5) 결과 반환.

- 새 거절을 기존 해독 뒤에 두는 이유: 기존 스키마 거절 시험의 픽스처가 새 가드에 먼저 걸리면 그 시험이 남의 가드를 잰다(기억 「새 가드는
  뒤의 가드를 못에서 뽑는다」). 해독은 부수 효과가 없어 뒤에서 거절해도 쓰기 0 이다.
- D1 을 D2 앞에 두는 이유: 상태 코드는 본문 해석보다 앞선 사실이다. 둘 다 같은 오류 계열이라 판정 결과는 같다.
- verifylive 쪽 가드 순서(a121 「가드 순서 계약」, `archive/2026-10-06-a121-…/design.md:420-428`)에는 **새 단계가 없다**. 새 거절은 이미 있는
  「종목 조회」·「두 번 읽기」·(만료 모드) by-id ①② 읽기의 **읽기 실패**로 나타나며, 그 자리의 기존 거절 코드를 그대로 받는다(§4).

### D4. 거절 코드 — 새 코드 없음

| 읽기 | 호출 자리(AST) | 오류 → 거절 코드 |
| --- | --- | --- |
| 조건주문 목록 | `readReconcileConditionalGroup` 줄 262 → 264 | `read-error`(`RefuseReadError`, `reconcile.go:125`) |
| 일반 주문 목록 | `readReconcilePlainGroup` 줄 290 → 292 | `read-error` |
| 종목 양성 대조 | `checkReconcileInstrument` 줄 159 → 161 | `instrument-control`(`RefuseInstrumentControl`, `reconcile.go:123`) |
| by-id 상세 | `readReconcileByID` 줄 57 → 59 | `expiry-read-error`(`RefuseExpiryReadError`, `reconcile.go:150`) |

리더 오류는 무엇이든 이 표의 코드 하나로 접힌다(문자열 상수는 HEAD 에서 읽음). 지금 E3(ErrServer)과 스키마 거절(ErrReconcileSchema)이
같은 코드를 받는 것과 같다.

### D6. 계약 키의 대소문자 변형 거절 (Manager 결정 4 = (나), 2026-11-25)

- 정의: 대사 읽기의 HTTP 200 본문 어느 깊이의 어느 객체에서든, 키가 **계약 키 목록**의 한 키와 `strings.EqualFold` 로 같으면서 바이트가
  다르면 거절. 계약 키 목록 = 대사 리더가 응답에서 읽는 키 전부(봉투 `apiEnvelope` · `decodeReconcileEnvelope` 의 키 리터럴과 컬렉션 인자 ·
  `apiConditionalOrder` · `apiConditionalCondition` 태그 · 주문 행·종목 항목의 익명 구조체 태그 · by-id 해독의 키 리터럴) — 23개.
- 왜: by-id 해독은 정확한 키만 읽고(`reconcile_detail_reads.go:76,80`, 측정 D4 ACCEPT) 목록 구조체 해독은 같은 바이트를 접어 읽는다(E2b).
  두 해독기가 같은 응답을 다르게 읽는 것이 「모호한 응답」 이다. 이 규칙 뒤에는 두 해독기가 같은 바이트만 읽는다(계약 키의 변형이 본문에
  있을 수 없으므로 구조체 접힘이 고를 대상이 정확한 키 하나뿐).
- 모양 선택: Manager 선호대로 **평평한 규칙 하나**(경로·층 구분 없음). 경로별 표는 층마다 다른 키 집합을 들고 다녀야 해서 더 크다 —
  더 작거나 안전하다는 근거를 못 찾음. 대가: 층이 다른 계약 키의 변형도 거절(예: 행 안의 `Orders`) — 오거절 방향.
- 목록은 한 곳(`reconcile_response.go` `reconcileContractKeys`)에만 두고, 시험 하나(`TestReconcileContractKeysAreTheKeysTheReadersRead`)가
  리더의 구조체 태그(`reflect` — 이름 있는 셋)와 리더 함수 본문의 키 자리(go/ast — 인덱스 리터럴 · `reconcileRequiredString`/
  `reconcileOptionalString` 둘째 인자 · `decodeReconcileEnvelope` 둘째 인자 · map 리터럴 키 · 함수 안 익명 구조체 태그)를 모아 목록과
  같아야 통과한다. 그 밖의 새 구조 가드는 두지 않는다(Manager 조건).
- D2 와의 관계: 계약 키에 대해서는 D6 이 짝(`status`+`Status`)도 먼저 잡는다. D2 의 짝 규칙은 계약 밖 키(`datetime`/`dateTime` 모양)를
  위해 남는다. 같은 객체 안 판정 순서: 정확한 중복(`reconcileObject`) → D6 → D2 짝.
- 재census(Manager 조건): `normal_input_census-output.txt` 의 `variant` 열.
  - 응답 키 계약인 openapi 스키마 속성 이름 423 중 변형 **0**(B4).
  - openapi example 563(객체·배열 240) **0**, testdata **0**, 시험 리터럴 553 **0**(클라이언트 경로 픽스처는 이 안에 있음).
  - 문서 전체 검사(B2·D)의 2건은 응답 키가 아니다: `openapi.latest.json` `components.parameters.Symbol`(매개변수 정의 이름)과
    `docs/api/openapi-v1.json` `components.{responses,schemas}.Orders`(TossOS 자체 운영자 API 문서 — 브로커 아님). 리더가 받는 본문에 나올
    수 없는 문서 구조 키라 「정상 입력 0」 으로 판단해 구현을 진행함(판단 근거를 보고에 명시).

### D5. 관측자는 체인하지 않는다

대사 파일(`internal/official/reconcile*.go`)은 a121 봉인 census `TestReconcileFilesCallNoWriteMethodAndAssertNoType`
(`internal/verifylive/reconcile_structure_test.go`)가 type assertion 을 금한다. 바깥 관측자를 꺼내 체인하려면
`ctx.Value(...).(AttemptObserver)` assertion 이 필요하다. 오늘 대사 경로 ctx 에 바깥 관측자는 없다(`WithAttemptObserver` 생산 호출
5곳 중 verifylive 는 `runner.go:212` M0 읽기 전용, cmd 0곳). 그래서 `m0_reads.go` 처럼 체인 없이 붙인다 — 대사 읽기 호출 동안만 바깥
관측자가 가려진다(§9 R4 에 기록).

## §4 호출 자리와 모드 (AST)

호출자 수가 아니라 자리를 센다(기억 「호출자 수는 고칠 자리 수가 아니다」).

- `verifylive.Reconcile`(AST 46분기): 읽기 자리 **6** — B16 줄 249 종목 ①, B17(`if p.Expired`) 안 B18 줄 258 by-id ①, 줄 262 목록 집합 ①
  (B19), 줄 266 목록 집합 ②(B20), B21(`if p.Expired`) 안 B22 줄 271 by-id ②, B23 줄 277 종목 ②.
- `readReconcileSet`: 자리 3 — 줄 198 조건 OPEN · 201 일반 OPEN · 204 조건 CLOSED. 그룹 함수는 페이지 루프 안 자리 1개씩(262 · 290).
- cmd 위임(`reconcileReads`): 자리 4 — 목록 둘은 `ensureIdentity` 뒤(254→257 · 261→264), by-id 도 뒤(15→18), 종목은 재확인 없이(268).
- 생산 호출자는 이것뿐이다(CodeGraph 1.6.0 callers + HEAD grep: 네 `Client.Reconcile*` 의 생산 호출자는 cmd 위임 하나씩, 위임의 생산
  호출자는 verifylive 하나씩). 엔진·대사 해제(`engine reconcile-resolve`)·주문 경로는 이 메서드를 부르지 않는다.
- **모드별 도달**: 부재 모드 — 종목 2 + 목록 집합 2 × 그룹 3 × 페이지 ≥1 = GET ≥ 8 이 새 판정을 지난다. 만료 모드 — 그 위에 by-id 2.
  두 모드 모두 같은 목록 판정을 지나므로 P5 의 「두 모드 공통」 결함이 같은 자리에서 닫힌다.

## §5 이 가드가 통과시키는 것 / 거부하는 정상 입력

- 통과: HTTP 200 이고, 본문의 어느 객체에도 접히는 두 키가 없는 응답 — 오늘 받아들여지던 응답 중 그런 것은 **해독 결과가 바이트 단위로
  같다**(새 판정은 값을 바꾸지 않고 통과/거절만 정함).
- 거부하는 정상 입력(실제로 브로커가 정상인데 거절되는 경우) — 의도된 비용:
  1. 브로커가 계약에 없는 200 아닌 2xx(예: 203 프록시 변형)를 돌려줌. 계약상 정상이 아니다.
  2. 응답 어딘가(판정에 안 쓰는 깊은 필드 포함)에 대소문자만 다른 두 키가 있음. **선례 1건이 저장소에 있다**:
     `fixtures/responses/auth-sanitized/asset-sections-v2.json` 의 한 객체가 `datetime` 과 `dateTime` 을 함께 담는다 — WTS(비공식) 응답이라
     대사 endpoint 가 아니지만 토스 백엔드가 그런 응답을 낸 적이 있다는 증거다. 대사 endpoint 모양 응답에서는 0건(§6). 거절되면 사람이
     대사를 못 할 뿐 주문·엔진에는 영향이 없다(대사는 사람이 돌리는 정리 명령, `mutating: true`).
  3. 정확히 같은 키가 두 번(RFC 8259 는 이름이 유일해야 한다고 SHOULD 로 적음). 0건(§6).

## §6 거부될 정상 입력 census (모집단·방법·수)

`normal_input_census-output.txt`(파이썬 `object_pairs_hook` 로 원문 키 순서를 받아 `lower()` 와 `casefold()` 둘 다로 접힘 판정, `raw_decode`
로 꼬리 판정):

| 모집단 | 수 | 접히는 키 | 꼬리 | 평가 못 함 |
| --- | --- | --- | --- | --- |
| A. official·verifylive·cmd/tossctl `*_test.go` 225 파일의 JSON 후보 리터럴 | 553 | 7 — **전부 a112 엄격 봉·호가 리더의 의도된 거절 픽스처**(`a112_mbus_read_test.go:148-149` · `strict_minute_candles_test.go:247-249,286` · `strict_quote_reads_test.go:237`), 대사 endpoint 아님 | 4 — 전부 `a169_stocks_security_type_test.go:33-36` 의 `+` 로 잇는 조각(이어 붙이면 정상 JSON — 확인함) | 108 — fmt 서식(`%q`)·조각 리터럴. 대사 관련 19건은 조각·서식(목록은 실행 로그) — **정적으로는 못 봤음** |
| B1. 세 패키지 testdata | 1 | 0 | 0 | 0 |
| B2. `openapi.latest.json` 문서 전체 | 1 | 0 | 0 | 0 |
| B3. openapi example 값 | 563(객체·배열 240) | 0 | 0 | 0 |
| C. 세 패키지 시험의 200 아닌 2xx 상태 리터럴 | — | — | — | **0 자리** |
| D. 저장소 `*.json`·`*.jsonl` 중 브로커 응답 모양 키를 담은 파일 | 28 | 1 — 위 WTS `asset-sections-v2.json` | 0 | 0 |

| E. 위에서 평가한 객체의 비 ASCII 키 | **0** | — | — | — |

술어 동치(리뷰 1회차 Codex P2): 파이썬 `lower()`/`casefold()` 는 Go `bytes.EqualFold`(단순 접기)와 비 ASCII 에서 갈리고(`ss`~`ß` 과대),
짝 없는 대리쌍 처리도 다르다(Go 는 U+FFFD 로 대체 — 과소). 평가한 모집단의 키가 **전부 ASCII**(E = 0)이므로 이 모집단에서는 두 술어가 같다.

결론: 대사 endpoint 모양의 정상 입력 중 새 판정이 거부하는 것 **0건(평가한 범위)**. 「평가 못 함」 108건은 정적으로 못 봤다. 2단계의
동적 확인은 **클라이언트 경로를 지나는 시험만** 센다(리뷰 1회차 Codex P2 — `DecodeReconcile*` 를 직접 부르는 시험은 새 판정을 지나지
않는다): official 클라이언트 시험(`reconcileTestServer`·`newDetailServer`), verifylive HTTP 시험(`newRCHTTPServer`), cmd/tossctl HTTP
시험이다. 직접 해독 시험의 수락 픽스처는 새 시험 C6(§8)이 새 판정 함수에 직접 넣어 센다.

## §7 편집 목록

| 파일 | 함수 | 종류 | FLM |
| --- | --- | --- | --- |
| `internal/official/reconcile_reads.go` | `Client.ReconcileConditionalOrdersPage` | 기존 함수 본문 수정(관측자 부착 · 해독 뒤 판정) | 번들 필요 |
| 〃 | `Client.ReconcileOpenOrdersPage` | 〃 | 번들 필요 |
| 〃 | `Client.ReconcileInstrument` | 〃 | 번들 필요 |
| `internal/official/reconcile_detail_reads.go` | `Client.ReconcileConditionalOrder` | 〃(silent 분류는 그대로 앞) | 번들 필요 |
| **새** `internal/official/reconcile_response.go` | 시도 수집 · D1 · D2 함수 셋 | 새 파일·새 함수 | 불필요(새 leaf) |
| **새** `internal/official/reconcile_ambiguity_test.go` | 클라이언트 경로 RED·대조(편집 전 컴파일 — 기존 API 만 씀) | 새 시험 파일 | 불필요 |
| **새** `internal/official/reconcile_response_test.go` | 판정 함수 직접 시험 · C6 · 계약 키 목록 핀(GREEN 로트) | 새 시험 파일 | 불필요 |
| **새** `internal/verifylive/reconcile_ambiguity_http_test.go` | 두 모드 통합 RED | 새 시험 파일 | 불필요 |

- 바꾸지 않는 것: `decodeReconcileEnvelope` · `DecodeReconcileConditionalPage` · `DecodeReconcileOrderPage` · `DecodeReconcileConditionalDetail` ·
  `reconcileObject` · `send` · `doRequest` · `unwrapAndDecode` · `token.go` · verifylive·cmd 생산 코드 · 기존 시험 함수(gate ⑤ 대상 0 목표).
- 새 파일 이름이 `reconcile*.go` 라 a121 봉인 census 대상이다 — type assertion·type switch·쓰기 메서드 이름 0(D5).
- a166 배선 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`)에 위 파일·함수 없음(HEAD grep — 기억
  「a166 census 는 배선 함수 본문을 고정한다」 확인).
- 의존·설정·DB·journal 변경 0.

## §8 2단계 시험 계획

선택 시험 패키지: `internal/official` · `internal/verifylive` · `cmd/tossctl` — 무태그와 `-tags tossos_testseams` 둘 다, `-count=1`. 그리고
`make lint`. 근거: 편집 파일이 `internal/official` 에만 있고 생산 호출 사슬이 cmd 위임 → verifylive 뿐(§4). `codegraph affected --filter
'*_test.go'`(1.6.0) 는 시험 파일 817개를 내지만 이는 `internal/official` 패키지 전체의 전이 의존이고 바뀐 심볼의 영향은 위 셋이다
(`analysis/code-context/codegraph-affected-tests.txt`). 미실행: `make test` · `make test-seams` 전체 · `make test-race` · `make gate`(사용자 명시
요청 시에만).

RED(편집 전 실패 확인 → GREEN):

| id | 층 | 입력 | 기대 |
| --- | --- | --- | --- |
| R1 | official 클라이언트 | 조건주문 목록 W1b(`"ORD-9"` 다음 `""`) · W2b · W3 · W3b · W4 · W5, 그리고 비 ASCII 접힘 `"second":{…},"ſecond":null`(U+017F) · Kelvin `K`(U+212A) 키, 이스케이프·접힘 키 `"triggeredOrderId":"ORD-9","\u0054riggeredOrderId":""`(해독하면 `TriggeredOrderId` — 원문 바이트도 해독 문자열도 다르고 접기로만 같음, 리뷰 2회차 Codex P2) | ErrReconcileSchema |
| R2 | 〃 | 봉투 E1 · E2, result R1 · R2 | ErrReconcileSchema |
| R3 | 〃 | 일반 주문 R4 · W6 · W7 | ErrReconcileSchema |
| R4 | 〃 | 종목 I1 | 오류 |
| R5 | 〃 | by-id D1 · D2 · D4, `ſ`·이스케이프 변형 | 오류(silent 아님) |
| R6 | 〃 | 네 리더 × 201 · 202 · 203 · 206 · 299 | 오류 |
| R7 | 〃 | 401 → 201(재시도 뒤 마지막 시도가 200 아님) | 오류 — 재시도 끝 상태를 봄(M2 는 못 잡음 — C2 가 잡음) |
| R8a | verifylive HTTP · 만료 모드 | CLOSED 의 **다른 id `EXPIRED` 행**에 W1b 모양 발동 흔적 숨김(편집 전 줄을 쓴다는 것은 AST 추론 — `checkReconcileExpiry` B2·B14 — 이고 2단계 RED 에서 실측) | `read-error`, 기록 줄 0 |
| R8b | verifylive HTTP · 부재 모드 | 조건 OPEN 페이지 `"conditionalOrders":[살아 있는 행],"conditionalOrders":[]`(R1 모양) 과 봉투 `result` 중복(E1 모양) | `read-error`, 기록 줄 0 |
| R8c | verifylive HTTP · 만료 모드 | R8b 와 같은 OPEN 행 숨김 | `read-error`, 줄 0 |
| R9 | verifylive HTTP | 만료 모드 by-id 201 | `expiry-read-error`, 줄 0 |
| R10 · R10b | official 클라이언트 | 판정에 안 쓰는 필드의 중복과 계약 밖 접힘 짝(`datetime`/`dateTime` · `size`/`ſize`) — W8(`first.condition` 안) · `{"orders":[],…,"metadata":{"x":1,"x":2}}`(리뷰 1회차 Codex P2) | ErrReconcileSchema — 「모든 깊이·모든 키」 고정 |

| R11 | verifylive HTTP · 만료 모드 | by-id 두 번 모두 200 statement 인데 `first.TriggeredOrderId` 단독 · `Second` 단독 · `Status`+`status`(D4) — 대상은 CLOSED 목록에 없음 | `expiry-read-error`, 줄 0. 편집 전 「줄 씀」 실측 |
| R12 | official 클라이언트 · verifylive HTTP | 목록의 단독 변형 — 봉투 `Result` 단독(E2b, 빈 페이지 → 편집 전 줄 씀), 행 `TriggeredOrderId` 단독 | `read-error` |

R8a·R8b·R8c 는 편집 전 실행에서 **줄이 써짐**을 먼저 확인한다(RED 가 기록 0 이 아니라 「줄 씀」 을 보여야 함 — W1b `CANCELED` 행처럼
이미 다른 가드로 거절되는 픽스처는 RED 가 아니다).

대조군·특성 고정(편집 전에도 GREEN 이어야 함 — 처음부터 GREEN 인 것은 RED 로 세지 않는다):

| id | 내용 |
| --- | --- |
| C1 | 정상 200 응답 네 리더 — 해독 결과가 편집 전과 같음(기존 시험 + 새 표 시험) |
| C2 | 401(본문에 중복 키) → 200 정상 — 수락. 판정이 **마지막** 시도 본문만 본다는 양성 대조 |
| C3 | W1(`"ORD-9"` 다음 `null`) — 편집 전 ACCEPT 이고 값 `ORD-9`(P4 정정의 고정). 편집 뒤에는 D2 로 거절 |
| R-T | E3 · E3b · E3c 꼬리 — 편집 전후 모두 거절(`unwrapAndDecode` 의존 고정, spec 둘째 시나리오) |
| C4 | by-id 404 `conditional-order-not-found` — 편집 뒤에도 silent(a164 분류 불변) |
| C5 | 순서 핀: 스키마 구멍 + 중복 키를 함께 가진 본문 → 기존 스키마 문구(예: `hasNext is missing`)가 나옴 — 새 판정이 기존 해독 **뒤**임을 고정 |
| C6 | 직접 해독 시험의 수락 픽스처(`TestDecodeReconcilePagesAcceptWellFormedPages` · `…KeepsTheSecondLegAndTheTrigger` · by-id golden)를 봉투로 감싸 새 판정 함수에 직접 넣음 — 거절 0 |

변이 원장(3.1, 각 변이는 사본에서·시작 sha 단언·무변이 대조군 GREEN 확인 뒤):

| id | 변이 | 잡을 시험 |
| --- | --- | --- |
| M1 | 200 판정을 2xx 판정으로 | R6 |
| M2 | 마지막 대신 첫 시도 | C2(401 → 200 수락이 거절로 바뀜). R7 은 못 잡음(첫 시도 401 도 200 아님) |
| M3 | D2 짝 비교를 정확 비교로(`EqualFold` → `==`) | R10b 계약 밖 접힘 짝 `{"datetime":1,"dateTime":2}`(D6 이후 계약 키 짝 W3b·E2·W7 은 D6 이 먼저 잡으므로 이 변이의 증인이 못 됨) |
| M3b | D2 짝 비교를 `ToLower(a)==ToLower(b)` 로 | R10b `{"size":1,"ſize":2}`(U+017F, 계약 밖) |
| M3d | D6 비교를 `ToLower` 로 | R1 `"second":…,"ſecond":null` · by-id `ſecond` 단독(계약 키 `second` 의 접힘 변형). Kelvin 키는 `ToLower` 로도 `k` 라 증인이 못 됨 |
| (M3c 삭제) | 원문 키 바이트 비교 | 구현이 키를 `reconcileObject` 가 해독한 문자열로만 받으므로 원문 바이트에 닿는 변이 자리가 없음 — 이스케이프 키 `\u0054riggeredOrderId` 는 행동 시험(R1·R5)으로만 둠 |
| M4 | 배열 재귀 생략 | R1 · R3 |
| M5 | 객체 재귀 생략 | R1 W1b · W5 |
| M9 | 알려진 키 아래로만 재귀(모르는 필드 생략) | R10 |
| M7b~d | 일반 주문 목록 · by-id · 종목 판정을 해독 앞으로(리뷰 1회차 Testing 추가) | C5 네 리더 |
| M12 | 정확한 중복 판정(`reconcileObject`)을 map 해독으로(리뷰 1회차 Testing 추가) | R1 · R2 · R4 · W1 · W3~W6 |
| M13 | 거절 경로의 키 인용(`strconv.Quote`) 삭제(리뷰 1회차 보안 수리의 핀) | `TestReconcileUnambiguousNamesTheRuleThatRefused` 제어 문자 키 |
| M10 | 계약 키 목록에서 `triggeredOrderId` 삭제 | R11 · R12 · 목록 핀 |
| M11a | D6 술어를 바이트 같음으로(`k != c && EqualFold` → `k != c && k == c`, 규칙 꺼짐) | R11 · R12 |
| M11b | D6 의 `k != c` 삭제(계약 키 자신도 거절) | C1 양성 대조 |
| M6a~d | 네 리더에서 판정 호출 하나씩 삭제 | 리더별 R1~R6 |
| M7 | 판정을 기존 해독 앞으로 이동(리더 하나) | C5 |
| M8 | 관측 0 방어 삭제(빈 목록을 통과) | 판정 함수 직접 시험(빈 시도 → 거절) |

## §9 남은 위험 · 범위 밖

- **R1 (Q2 범위 밖)** R-a 401 갱신 뒤 신원 재확인 없음 · R-b 직렬화 전 시각 — proposal 「관련 잔여」 그대로.
- **R2 접힘 규칙의 오거절**: §5-2 의 WTS 선례. 대사 endpoint 에서 생기면 사람이 대사를 못 한다(안전 방향, 주문 영향 0). 범위를
  「구조체로 해독하는 층의 키」 로 좁히는 안은 §10 Manager 결정 1.
- **R3 `reconcileObject` 꼬리 검사의 사각**(P10): `}`·`]` 꼬리를 못 봄. 생산에서는 도달 불가(P9). 고치려면 a164 함수 편집 — 이 change 에서 안 함.
- **R4 관측자 가림**(D5): 대사 읽기 동안 바깥 관측자가 있으면 그 시도를 못 본다. 오늘 0곳.
- **R5 비용**: 해독 비용은 객체마다 `reconcileObject` 를 다시 불러 하위 트리를 다시 토큰화하므로 O(본문 × 깊이)다(리뷰 1회차 Claude P3). 그와 별도로 접힘 짝 비교는 객체마다 O(키 수²)다 — 얕고 넓은 객체에서도 생김(리뷰 2회차 Codex P3). 깊이는
  `encoding/json` 의 중첩 상한(10000)까지 갈 수 있고, 본문 크기 상한은 기존과 같이 없다(`doRequest` 의 `io.ReadAll`). 대사는 사람이 돌리는
  정리 명령이고 입력은 공식 브로커 응답이라 수용 — 측정은 2단계에서 정상 응답 크기로 한 번.
- **R5 실측(2단계)**: 정상 모양 조건주문 페이지에서 `reconcileUnambiguous` 한 번 — 1행 601 B ≈ 0.05 ms · 20행 10.7 KB ≈ 1 ms ·
  100행 53 KB ≈ 5 ms(`analysis/measure/r5_cost-output.txt`, i9-12900H, go1.26.5). 실행당 GET ≥ 8~10 이라 15 s 창에 비해 무시 가능.
  병적 입력(리뷰 1회차 실측): 객체 중첩 깊이 4,000 에 566 ms · 649 MB 할당, 한 객체 키 20,000 개에 0.66 s — 깊이 10,000 이면 수 초·수 GB 로
  외삽(미측정). 호출자는 `tossctl verify reconcile` CLI 하나(엔진 아님). 깊이·크기 상한은 새 가드라 §10 Manager 결정 대기 5.
- **R6 동적 census 는 2단계에서만 닫힌다**(§6 「평가 못 함」 108, 클라이언트 경로 시험 + C6).
- **R7 404 silent 본문의 중복 키**(리뷰 1회차 Codex P2): `{"error":{"code":"permission-denied","code":"conditional-order-not-found"}}` 404 는
  `reconcileDetailSilent`(`reconcile_detail_reads.go:157-172`)가 마지막 값으로 silent 로 가른다. 새 판정은 200 본문에만 걸리므로(D1·D3)
  이 갈래는 바뀌지 않는다 — spec 도 「성공한 읽기의 본문」 으로 좁혔다. 줄을 쓰는 길은 없다: 만료 모드는 silent 를 근거로 세지 않고
  (`reconcile_expired.go:88-95`) 둘 다 silent 면 거절, 하나만이면 `expiry-reads-differ`. a164 분류(사용자 범위 「silent 불변」)라 고치지 않음.
- **R8 `GOEXPERIMENT=jsonv2` 빌드는 안 봤다**: 접힘 동치 근거(`fold.go`)는 기본 빌드(`//go:build !goexperiment.jsonv2`) 것이다. `go.mod`
  는 go 1.25.0, 측정 툴체인은 1.26.5 — `fold.go` 의 `bytes.EqualFold` 동치 문장은 1.21 부터 같다(1.25 소스는 안 열어 봤음).

## §10b Manager 결정 대기 — 2단계 리뷰에서 나온 것 (구현 안 함)

5. **병적 본문 비용 상한**(R5) — 깊이·본문 크기 상한 또는 한 번 훑기 O(n) 로 바꾸기. 권장: 지금은 잔여로 둠(정상 응답 5 ms, CLI 전용), 상한이 필요하면
   별도 change.
6. **[리뷰 2단계 1회차 — Claude 적대 · Codex 둘 다] 만료 모드에서 대상 아닌 CLOSED 행의 `first.status` · `type` 미검사** — 다른 id 의
   `{"status":"EXPIRED","first":{"status":"WATCHING"}}` · `"type":"OCO"` 행이 top-level `EXPIRED` 만으로 통과(`reconcile_expired.go:120-124`,
   leg 검사는 대상 행만 `:106-109`·`:145-150`). 해독 모호함이 아니라 판정 의미의 기존 공백(a164) — 이 change 범위 밖. 권장: 별도 후속
   change(목록 행에 `Type` 노출 + 모든 CLOSED 행 leg 어휘 검사), 브로커가 그런 조합을 내는지 영수증 확인 선행.

## §10 Manager 결정 (2026-11-25 확정 — 아래 「대기」 문안은 기록으로 남김)

- 1 = (가) 모든 깊이·모든 객체의 접히는 두 키 거절. 2 = `ReconcileInstrument` 포함. 3 = D5 체인 없음 수락(R4 잔여, 새 구조 시험 없음).
  4 = (나) 이 change 에 포함 — D6.

### (기록) 결정 대기 문안

1. **D2 범위** — (가) 모든 깊이·모든 객체(권장, 이 문서) / (나) 구조체 해독 층(봉투·행·`first`·`second`·주문 행·종목 항목)만 / (다) 정확한
   중복만(접힘 제외). 권장 (가): 한 규칙·한 자리이고, 측정된 덮어쓰기(W7·E2·W2b·W3b)를 다 닫으며, 대사 endpoint 모양 정상 입력 0건.
   (다)는 W7·E2·W2b·W3b 를 열어 둔다.
2. **종목 양성 대조(`ReconcileInstrument`)를 포함** — 권장 포함. 「대사 읽기의 성공 계약은 하나」(spec)이고 I1 이 같은 성질.
3. **D5 체인 없음** — 권장 수락. 체인하려면 type assertion 을 봉인 census 밖 새 파일에 둬야 하는데, 오늘 바깥 관측자가 0이라 YAGNI.
4. **홀로 선 대소문자 변형 키(리뷰 1회차 P1 — Claude·Codex 둘 다, Codex 는 OUT-OF-SCOPE 표기)** — by-id 200 본문
   `{"result":{…,"status":"EXPIRED",…,"first":{"status":"EXPIRED","TriggeredOrderId":"ORD-9"},"Second":{"status":"WATCHING"}}}` 는 짝이 없어
   D2 를 통과하고, by-id 해독은 정확한 키만 읽어(`reconcile_detail_reads.go:76,80`) `TriggeredOrderID=""`·`HasSecond=false` 로 본다.
   목록 행 구조체 해독은 같은 바이트를 `ORD-9`·둘째 다리로 읽는다(측정 E2b 와 같은 접힘). 대상이 CLOSED 목록에서 빠져 있으면 만료 줄이
   써질 수 있다 — **a164 부터 있던 해석 차이**이고 이 change 의 「중복」 정의 밖이다. 선택지: (가) 이 change 에 넣지 않고 잔여로 기록(이 문서의
   기본 — 사용자 지시 「리뷰가 제안한 새 가드는 계획에 넣지 말 것」) / (나) D2 에 「정확한 키로 읽는 객체(by-id result·first, 목록 result)에서
   읽는 키와 접히지만 바이트가 다른 키」 거절을 더함(같은 walker 안 몇 줄, 거부할 정상 입력은 §6 모집단에서 0 — 비 ASCII·대소문자 변형 키 0).
   권장 **(나)**: 같은 메커니즘(접힘)이 by-id 에서 반대 방향으로 증거를 지우는 유일한 남은 append 경로이고, 브로커 계약 위반 바이트가
   필요하지만 비용이 작다. 사람 결정이 필요한 범위 확장이므로 Manager 가 정한다.
