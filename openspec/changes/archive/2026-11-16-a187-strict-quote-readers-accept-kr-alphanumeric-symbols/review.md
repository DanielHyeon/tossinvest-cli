# Review

## 제안 단계 기록 (Manager, 2026-11-12)

- 계기: 2026-11-12 a184 A 안에서 `0158S0` 이 `ask_unavailable`·`parse` 로 탈락함. 사용자 결정 「호가 규칙을 넓히는 별도 체인지 먼저 해」.
- memory recall: ai-memory 는 a143 커밋 줄과 과거 세션 발췌만 돌려줌(결정·함정 페이지 없음). Hindsight recall 은 결과가 한도를 넘어 파일로 떨어졌고 읽지 않음 — 그 채널은 안 봤음.
- 증거 순서: 편집 전 AST 10개(`analysis/pre-edit-ast/`, base `6dbb752f6`)를 먼저 만들고 design Context 의 분기 번호를 거기서 옮김.
- 호출자 전수는 `grep` 으로 함(시험 제외): `.StrictOrderbookTop(`·`.StrictLastPrice(`·`PollQuoteL1(`·`.RawMinuteCandles(`, 호출 아닌 참조는 `officialbars.QuoteReader` 인터페이스 선언 2줄뿐. CodeGraph 재열거는 tasks 1.3.
- 둘째 벽 확인: `officialbars.checkSymbol` 은 빈 값·공백만 거름. 호가 본문 키는 `timestamp, currency, asks, bids` 라 종목 메아리 검사가 없음(현재가 행에는 `symbol` 이 있어 메아리 검사가 있음).
- 공식 명세 대조: `docs/migration/openapi.latest.json` 1.2.24, `/api/v1/orderbook` `symbol` 설명 「KRX: 6자리 종목코드 (숫자 또는 영문·숫자 조합; 예: 005930, 0101N0)」. `/api/v1/prices` `symbols` 설명에는 영문 섞인 예시가 없고 `pattern` 만 영숫자를 허용함 — 현재가 쪽은 명세 문구로 확정되지 않음(실측 4.1 에서 확인).
- 개정하는 정본: `openspec/specs/official-read-compatibility/spec.md` 「Minute symbol compatibility does not widen quote or order policy」. MODIFIED 는 옛 시나리오 이름을 지울 수 없어 REMOVED + ADDED 로 씀. 주문 정책 절반과 a087 문구 시나리오는 그대로 옮김.
- 미실측: 공식 호가·현재가 조회가 영문 섞인 코드를 받는지.
- 남은 위험: 프로브 입력의 조용한 변화, 주문 쪽 시장 추론이 영문 섞인 한국 코드를 `us` 로 봄(design Risks).

## 구현 기록 (Opus 팀메이트, 2026-11-12, tasks 1.1~3.2)

- 1.1: `base-commit.txt`(6dbb752f6)는 Manager 가 이미 만듦 — 체크만.
- 1.2: 편집할 함수 `strictMinuteCheckCandleSymbol` 의 편집 전 AST 를 `go run ./tools/logic-map` 으로 뽑음. `analysis/pre-edit-ast/strictminutecheckcandlesymbol.ast.json` 과 내용이 같음(diff 0). design Context 의 분기·행(`strictMinuteCheckSymbol` B1 221·B2 224, `strictQuoteIdentity` B2 201, `StrictOrderbookTop` B2 143·146, `StrictLastPrice` B2 174·177, `StrictMinuteCandles` B3 140, 래퍼 B1 213·216, `InferMarketFromStockCode` B1 277·B2 279, `PollQuoteL1` 108)을 소스와 AST 로 대조했고 모두 맞음. 시험 함수 `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest` 도 본문(사례 표)을 고치므로 번들을 만듦(gate ⑤ 는 시험 함수도 셈).
- 1.3: CodeGraph(codegraph **1.6.0**, 색인 `/mnt/D/Axipient/workspace/TossOS` @ 6dbb752f6 = base) `callers` 원문은 `analysis/codegraph-callers-1.3.txt`. 시험 제외 호출 자리는 proposal Impact 의 grep 전수와 같음: `officialbars.PollQuoteL1`(quote.go:113·117) ← a140 `fetchQuote`(shadow.go:288), a128 프로브(collect.go:453 — CodeGraph 는 동적 호출을 인터페이스 선언 collect.go:31 로 귀속), a140 source.go:148·154(인터페이스 선언 26·27 과 래퍼 메서드 152 로 귀속), a112 `runWithOptions`(main.go:87·91). a184 main.go:287 은 다른 브랜치라 이 색인에 없음. `codegraph affected internal/official/strict_minute_candles.go internal/official/strict_quote_reads.go --filter '*_test.go'` 는 시험 파일 **1,179개**(전이 의존이 넓어 시험 선택 근거로 쓰지 않음).
- 2.1 RED(48f84887e): 새 시험 `a187_kr_alphanumeric_test.go`(수락 4종목·거절 9모양·숫자 KR·US 불변), `internal/orderintent/a187_order_policy_test.go`(주문 쪽 숫자 전용 불변). 기존 거절 표의 `00593A`·`0101N0` 두 줄은 지우고 수락 시험으로 뜻을 뒤집어 옮김. 실패 원문 `analysis/tdd/red-2.1.txt`: `StrictOrderbookTop(KR, 0101N0): official strict quote: SYMBOL_INVALID: symbol "0101N0" is not canonical for market KR`(00593A·0158S0·18064K 같음).
- 2.2 GREEN(305733da1): design D1·D2 그대로. 공용 `strictMinuteKRSymbol` = `^[A-Z0-9]{6}$`, `strictMinuteCandleKRSymbol` 삭제, 래퍼는 한 줄 위임, 주석 정정. 다른 함수 본문·주문 정책·프로브 무편집. 원문 `analysis/tdd/green-2.2.txt`(PASS 106, FAIL 0).
- 2.3: 편집 뒤 AST·FLM/BTM 재생성, `check_analysis.py` 통과(원문 `analysis/tdd/check-analysis-2.3.txt`: required 2, 「evidence complete or diff-proven exempt」).
- 3.1 변이(`analysis/mutation/mutation.py`, 원문 `mutation-3.1.txt`): 사본 `/tmp/a187-mut-<pid>-…`, `GOFLAGS=-trimpath`, 전용 GOCACHE. 무변이 대조군 GREEN. 6개 모두 CAUGHT — 숫자 전용 되돌림(`TestA187StrictQuoteReadersAcceptKRAlphanumeric`·분봉 수락 시험), 길이 고정 제거(`TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest/kr_symbol_too_short`·`TestA187…StillRefuseOtherKRForms/0101N`·`0101N00`·분봉 거절 시험), 앞 고정 제거(`…StillRefuseOtherKRForms/0101N00`·`_0101N0`), 뒤 고정 제거, 소문자 허용(`…StillRefuseOtherKRForms/0101n0`·분봉 `kr_symbol_lowercase`), 래퍼가 검사를 건너뜀(`TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` 실패 하위 시험 10개 + 부모 1개 — 리뷰 r1 지적으로 「11 갈래」를 정정함).
  - 대조군 1차는 FAIL 이었음(`history-1-control-trimpath.txt`): 정적 가드 3개. 둘은 `-trimpath` 에서 `runtime.Caller` 경로를 못 여는 것, 하나는 아래 base 기존 결함. 셋을 `-skip` 으로 빼고 다시 돌림.
- 3.2(원문 `analysis/verify/`): `go test -count=1 -v` 6개 패키지 범위 — 최상위 PASS 507, 서브 포함 PASS 919, **FAIL 5**, SKIP 0. `go vet` rc 0, gofmt 0줄, `openspec validate --strict` valid.
  - FAIL 5개는 모두 **base 6dbb752f6 에서도 같은 문구로 실패**함(`analysis/verify/base-failures.txt`, base 사본에서 같은 시험을 돌림). 이 change 와 무관하며 고치지 않음:
    1. `internal/official` `TestA112MBUSStaticNoProductCallerOrForbiddenPath`: `forbidden resolved A112 M-B0 references: []string{"internal/accountbudgetruntime/v2_token_test.go: selector official.A112MBUSCandle"}`
    2. `tools/a128-jev-shadow-probe` `TestImportsStayInsideTheAllowlist`: `internal/accountbudget`·`internal/accountbudgetruntime` import 가 허용 목록 밖
    3. `tools/a128-jev-shadow-probe` `TestNoMutatingOrAccountSurfaceIsNamed`: `source uses official.Option, outside the read-only allowlist`
    4. `tools/a140-breakout-signal-shadow` `TestImportsStayInsideTheAllowlist`: 같은 두 import
    5. `tools/a140-breakout-signal-shadow` `TestNoMutatingAccountOrWriteSurfaceIsNamed`: `source uses official.Option, outside the allowlist`
  - 숫자 전용을 고정한 시험은 위 두 사례 말고 6개 패키지에서 더 나오지 않음(실패 5개 모두 무관). 주문(`internal/orderintent`)·a087 쪽 시험은 통과. (리뷰 r1 발견 4 로 정정: 여기서 돈 a087 관련 시험은 `internal/official` 의 가격제한폭 필드 시험 `TestA087ProtectivePriceLimitsPreservesNullableRawFields`·`TestA087ProtectivePriceLimitsKeepsJSONNullDistinctFromZero` 두 개뿐이고 주문 스키마 시험이 아님. `tools/a087-market-sell-probe` 패키지 시험은 3.2 범위에 없어 그때는 미실행이었음 — 아래 r1 처리 절에서 따로 돌림.)
- 하지 않은 것: 3.3(다른 계열 리뷰·gstack), 3.4, 4.x(실측·반영). 전체 gate·전체 시험. 공식 API 호출.
- 남은 위험: design Risks 그대로(공식이 영숫자 코드를 호가·현재가에서 받는지 미실측, 프로브 입력의 조용한 변화, 주문 쪽 시장 추론이 영숫자 KR 을 `us` 로 봄). 그리고 main 의 정적 가드 5개가 base 에서 이미 깨져 있음 — 별도 처리가 필요함.

## 구현 리뷰 r1 (gpt-6-astra) 처리 — 2026-11-12

- 원문: `analysis/impl-review-r1-gpt-6-astra.md`(그대로 복사). 판정 FAIL: blocker 0, major 1, minor 3, note 1. 운영 코드 수정 없음.
- 문서 정정(발견 1·2·5)은 Manager 가 했고 커밋 760fc7c42 에 그대로 넣음(proposal Non-goals 주문 문구·Impact 호출량·프로브 두 입력군, design Context 의 `NormalizePlace` 분기·D4·Risks, spec delta 의 주문 문장, `analysis/pre-edit-ast/normalizeplace.ast.json`).

| # | 분류 | 처리 | 근거 위치 |
|---|---|---|---|
| 1 | major | 문서 정정(Manager): 유지 대상은 「시장 추론·주문 정책 무변경」이고 주문 검증 전체가 숫자 전용인 것은 아님. 시험 정정(이번): `TestA187OrderPolicyStaysNumericOnly` → `TestA187MarketInferenceStaysNumericOnly`(이름·주석을 시장 추론 고정으로), `TestA187NormalizePlaceExplicitKRAlphanumericUnchanged` 추가 — `NormalizePlace{Market:"kr", Symbol:"0101N0", Side:"buy", OrderType:"limit", Quantity:1, Price:1000}` 가 오류 없이 Market "kr"·Symbol "0101N0" 을 돌려주는 현 동작(a187 이전부터)을 고정. 판단 없이 현 동작 고정만 함. 통과 | spec delta·proposal·design(760fc7c42), `internal/orderintent/a187_order_policy_test.go`, `analysis/verify-r1/go-test-official-orderintent.txt` |
| 2 | minor | 문서 정정(Manager): 리더 호출 계수와 실제 HTTP 시도를 구분, 재시도·인증 영향 | proposal Impact(760fc7c42) |
| 3 | minor | 시험 보강: 수락에 `A12345`·`ABCDEF`, 거절에 `"0101N0\n"`·`"\n0101N0"`, KR 현재가 메아리 불일치(`StrictLastPrice(KR, "0158S0")` 에 행 symbol `0101N0` → `SYMBOL_ECHO_MISMATCH`, 요청 1회). 리뷰가 생존 예상으로 짚은 변이 3종을 보강 전·후로 돌림(아래) | `internal/official/a187_kr_alphanumeric_test.go`, `analysis/mutation/mutation-r1-before.txt`·`mutation-r1-after.txt` |
| 4 | minor | 기록 정정: 3.2 에서 돈 a087 관련 시험은 가격제한폭 필드 시험 두 개뿐(위 구현 기록 정정). `tools/a087-market-sell-probe` 는 3.2 범위 밖이라 미실행이었음 → 이번에 `go test -count=1 -v ./tools/a087-market-sell-probe/...` 를 돌림: ok, 최상위 PASS 39·서브 포함 62·FAIL 0·SKIP 0. 그 패키지의 스키마 시험은 숫자 코드(`goodSpec()` 의 `005930`)를 씀 — 영숫자 KR 로 a087 스키마를 돌린 시험은 없음 | `analysis/verify-r1/go-test-a087.txt` |
| 5 | note | 문서 정정(Manager): 프로브 반영 경계를 `current`·`initial/compare` 두 입력군 모두에 적용, digest·토큰 수·선별·수집 시간 변화 가능성 기록 | proposal Impact·design Risks(760fc7c42) |

### 리뷰 r1 변이 3종 (보강 전·후)

- 하네스는 `analysis/mutation/mutation.py` 에 3종을 더함(변이 대상이 둘째 파일이어도 사본에서만 바꿈 — `strict_quote_reads.go` 원본 무편집). 대조군은 둘 다 GREEN.
- 보강 전(기준 HEAD 760fc7c42, `mutation-r1-before.txt`): 리뷰 예상대로 **3종 모두 SURVIVED**.
  - `r1-leading-digits` `^[0-9]{4}[A-Z0-9]{2}$`
  - `r1-multiline` `(?m)^[A-Z0-9]{6}$`
  - `r1-echo-skip-krw` `if echoed != symbol && currency != "KRW"`
- 보강 뒤(기준 HEAD 3355528bb, `mutation-r1-after.txt`, 9종 전체 재실행): **9종 모두 CAUGHT**.
  - `r1-leading-digits` → `TestA187StrictQuoteReadersAcceptKRAlphanumeric/A12345`·`/ABCDEF`(`SYMBOL_INVALID: symbol "A12345" is not canonical for market KR`)
  - `r1-multiline` → `TestA187StrictQuoteReadersStillRefuseOtherKRForms/_0101N0#01`·`/0101N0_#01`(개행 두 사례 — go 가 하위 시험 이름의 개행을 `_` 로 바꾸고 공백 사례와 겹쳐 `#01` 을 붙임; `StrictOrderbookTop accepted`)
  - `r1-echo-skip-krw` → `TestA187StrictLastPriceRefusesKREchoMismatch`(`StrictLastPrice accepted a row answering for another KR symbol`)
  - 앞의 6종도 다시 CAUGHT.

### 검증 (원문 `analysis/verify-r1/`)

- `go test -count=1 -v ./internal/official/ ./internal/orderintent/`: 최상위 PASS 289, 서브 포함 584, FAIL 1 — base 기존 결함 `TestA112MBUSStaticNoProductCallerOrForbiddenPath` 하나뿐(`analysis/verify/base-failures.txt`). 그 밖 FAIL 0.
- `go vet`(official·orderintent·a087 probe) rc 0, gofmt 0줄, `openspec validate --strict` valid, `check_analysis.py` rc 0(「evidence complete or diff-proven exempt」 — 이름을 바꾼 시험 함수는 이 change 에서 새로 만든 것이라 base 함수가 아님).

### 미실행

- gstack 리뷰, Manager 재확인(tasks 3.3 은 체크하지 않음), 3.4, 4.x, 전체 gate, 공식 API.
- `tools/a087-market-sell-probe` 의 변이(`mutants.sh`)는 돌리지 않음.

## gstack 리뷰 처리 — 2026-11-13

- 원문: `analysis/gstack-review.md`(그대로 복사). gstack 1.98.0.0 `review`. blocker 0, major 1, minor 6, note 11. 운영 코드 동작 변경 없음 — 운영 파일 편집은 주석 한 줄뿐(`analysis/verify-gstack/production-diff-vs-d5a2738bf.txt`).
- 문서 정정(Manager, f367081f7): spec delta 「Order policy is not widened」 시나리오(발견 1), proposal 프로브 적용 시점·행으로 가르는 법·수급 결손·실측 결과(발견 5·6), design Open Questions(발견 6), tasks 4.2 문장(발견 5).

| # | 분류 | 처리 | 근거 위치 |
|---|---|---|---|
| 1 | major | 수정(Manager 문서): 시나리오를 실제 시험(시장 추론 고정·`NormalizePlace` 현 동작 고정)으로 고쳐 씀. a087 무편집. `rg -n 0101N0 tools/a087-market-sell-probe/` 일치 0건(rc 1) | spec delta(f367081f7), `analysis/verify-gstack/rg-a087-0101N0.txt` |
| 2 | minor | 수정(시험): `TestA187StrictMinuteCandlesAcceptKRAlphanumericOnce`(분봉 `0158S0`·`A12345`·`18064K`, 요청 1회·정확한 query·종목), `TestA187StrictMinuteCandlesStillRefuseOtherKRForms`(분봉 `0158s0`·`0101N0\n`·`\n0101N0`·`a12345`, 요청 0회·`SYMBOL_INVALID`). 기존 분봉 시험 파일은 고치지 않음. 변이 (d) 로 확인 | `internal/official/a187_kr_alphanumeric_test.go`, `analysis/mutation/mutation-gstack-*.txt` |
| 3 | minor | 수정(주석): `strictMinuteKRSymbol` 위 둘째 줄을 「주문 쪽(internal/orderintent)은 이 값을 읽지 않음 — 그쪽에서 숫자 전용인 것은 시장 추론뿐임」으로. 줄 수 같음, gofmt 0줄, `strictMinuteCheckSymbol`·`strictMinuteCheckCandleSymbol` 의 logic-map AST 가 행까지 편집 전과 같음(source_sha256 제외) | c917ccf55 |
| 4 | minor | 수정(시험 표): `unknown market` 행에 `StrictQuoteReasonMarket`, `kr symbol too short` 행에 `StrictQuoteReasonSymbol` 을 채워 사유 비교 갈래(B4·B7)를 살림. 실행에서 두 리더의 실제 사유와 맞음(통과) | `internal/official/strict_quote_reads_test.go`, `analysis/verify-gstack/go-test-official-orderintent.txt` |
| 5 | minor | 수정(Manager 문서): 프로브 경계는 「main 반영 시각」이 아니라 프로브 기동(`go run`)이고, 반영 전 행에는 `top_of_book: …SYMBOL_INVALID` 결손 문구가 남음 | proposal·tasks 4.2(f367081f7) |
| 6 | minor | 수정(Manager 문서): Open Questions·Impact 를 4.1 실측(장 밖 1회, 종목 1개)과 맞춤 | design·proposal(f367081f7) |
| 7 | minor | 수정(시험): `TestA187NormalizePlaceMarketInferenceUnchanged` — `NormalizePlace` 에 시장 `us` 또는 생략 + `0101N0` 은 `us` 로 남고 `005930` 은 `kr` 로 바뀜(현 동작 고정). 변이 (e) 로 확인 | `internal/orderintent/a187_order_policy_test.go` |
| 8 | note | 기록만: 호가 본문에 종목 필드가 없어 `top.Symbol` 은 요청값임(a143 부터의 한계, design Risks) | gstack 원문 |
| 9 | note | 기록만: KR 문법과 US 문법이 `A12345`·`ABCDEF` 에서 처음 겹침. 시장 인자로 가르고 실재 KR 코드는 영문자가 5·6번째 자리뿐이라 지금은 이론적 | gstack 원문, design D3 |
| 10 | note | 후속: 숫자 전용 시장 분류가 있는 다른 자리 — `internal/verifylive/pricing.go`, `internal/flatten/flatten.go`, `internal/client/trading.go`, `tools/a087-market-sell-probe/order.go`, `internal/official/candle_raw.go` — 를 a184 실거래 전환 change 로 인계 | gstack 원문 |
| 11 | note | 기록만: 「유니버스 불변」 문장과 사실상 후보가 넓어지는 것의 읽기 차이. 유니버스 코드는 안 고침 | gstack 원문 |
| 12 | note | 기록만(사실): FLM 번들은 본문을 고친 래퍼(동작은 안 바뀜)와 거절 시험 함수에 붙고, 동작이 바뀐 `strictMinuteCheckSymbol`(패키지 변수 값만 바뀜)에는 없음 — gate ⑤ 는 함수 본문 편집만 셈. 실질은 시험과 변이 11종이 덮음 | `analysis/function-logic/`, `analysis/verify-gstack/check-analysis.txt` |
| 13 | note | 후속: `tools/a112-l1c-quote-probe/main.go:36` 도움말 「KR six digits」 갱신(diff 밖) | gstack 원문 |
| 14 | note | 기록만: a185 브랜치가 같은 PM 줄을 고침 — 둘째로 반영되는 쪽이 충돌을 풂 | gstack 원문 |
| 15 | note | 처리: a184 가 가져간 사본(d5a2738bf 기준)은 이번 수정 뒤 다시 가져감(a184 브랜치 기록) | a184 review.md |
| 16 | note | 수정: 하네스가 변이 판에서 `-v` 를 끄고 돌려, 통과한 시험의 `t.Log` 가 「실패 원문」에 섞이지 않음. 대조군만 `-v` 로 PASS 줄을 셈 | `analysis/mutation/mutation.py` |
| 17 | note | 후속(advisory): 한 줄 위임 래퍼 제거. a187 에서는 design D2 대로 둠 | gstack 원문 |
| 18 | note | 기록만: `openspec/config.yaml` 은 추적 안 되는 파일이고 건드리지 않음 | `git status` |

### 변이 (d)(e) — 보강 전·후

- (d) `g-wrapper-own-regex`: 래퍼가 `^[A-Z0-9]{6}$|^[a-z0-9]{6}$` 를 다시 들여 공용 검사와 갈라짐.
  - 보강 전(c917ccf55, `mutation-gstack-before.txt`): CAUGHT — 기존 `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest/kr_symbol_lowercase`.
  - 보강 뒤(677ac969b, `mutation-gstack-after.txt`): CAUGHT — 위 시험에 더해 `TestA187StrictMinuteCandlesStillRefuseOtherKRForms/0158s0`·`/a12345`(「StrictMinuteCandles accepted」). gstack 이 생존을 예상한 여러 줄 모드 래퍼 변형은 새 분봉 개행 사례(`0101N0\n`·`\n0101N0`)가 막음.
- (e) `g-order-inference-widened`: `intent.go` 의 `looksLikeKRSymbol(intent.Symbol)` 호출을 `^A?[A-Z0-9]{6}$` 인라인 검사로.
  - 보강 전: **SURVIVED**(gstack 예상과 같음).
  - 보강 뒤: CAUGHT — `TestA187NormalizePlaceMarketInferenceUnchanged`(「NormalizePlace("us", "0101N0") market="kr" … want market "us"」).
- 보강 뒤 11종 전체 재실행: 대조군 GREEN, **11/11 CAUGHT**.

### 검증 (원문 `analysis/verify-gstack/`)

- `go test -count=1 -v ./internal/official/ ./internal/orderintent/`: 최상위 PASS 292, 서브 포함 594, FAIL 1 — base 기존 결함 `TestA112MBUSStaticNoProductCallerOrForbiddenPath` 뿐. 그 밖 FAIL 0.
- `go vet` rc 0, gofmt 0줄, `openspec validate --strict` valid, `check_analysis.py` rc 0, PM check current.

### 미실행

- Manager 재확인(3.4), 4.2·4.3, 전체 gate, 공식 API. 후속 항목 10·13·17 은 이 change 에서 하지 않음.


## Manager 재검증 (3.4) — 2026-11-15

- 원문(Manager 실행, 그대로 복사): `analysis/verify-manager/go-test-7pkg.txt`, `analysis/verify-manager/mutation-11.txt`. 아래 수치는 팀메이트가 복사본에서 다시 셌고 Manager 값과 같음.
- 기준 HEAD `be90ea3e7`, base `6dbb752f6`. base..HEAD 19커밋 전부 a187. openspec 밖 변경 파일 10개, 운영 코드 변경은 `internal/official/strict_minute_candles.go` 하나(정규식 변수 2→1, 래퍼 한 줄 위임, 주석). `strictMinuteCandleKRSymbol` 참조 0건, `strictMinuteCheckSymbol` 호출 자리 2곳(래퍼 214행, `strict_quote_reads.go` 201행).
- 시험(`go test -count=1 -v`, 7개 패키지: `internal/official`, `internal/orderintent`, `internal/officialbars`, `tools/a128-jev-shadow-probe`, `tools/a140-breakout-signal-shadow`, `tools/a112-l1c-quote-probe`, `tools/a087-market-sell-probe`): 최상위 PASS 551, 서브 포함 997, FAIL 5, SKIP 0. `TestA187…` 최상위 9개 모두 PASS. FAIL 5개는 시험 이름과 실패 문구(7줄)가 `analysis/verify/base-failures.txt` 와 같음.
- 변이(`analysis/mutation/mutation.py`, 사본): 대조군 GREEN(PASS 줄 592), 11/11 CAUGHT. rc 0 과 「실행 뒤 작업 트리 변화 없음」은 Manager 관측이며 원문 파일에는 찍혀 있지 않음(팀메이트 기록 시점 `git status` 는 깨끗했음).
- `gofmt -l internal/official internal/orderintent` 0줄, `go vet` 두 패키지 rc 0(팀메이트 재실행도 같음), `openspec validate --strict` valid, PM `--check` current.
- 판정: 구현 완료(tasks 3.4 통과).
- 남은 문제: 완료 게이트 7단계 `make test`(`go test ./...`)는 base 에서도 실패하는 정적 가드 5개 때문에 지금 main 에서 통과할 수 없음. 위반을 들여온 커밋은 `78d777cfe`(feat(a166): shadow 도구 예산 연결)·`4d5884cbc`(둘 다 2026-10-31). 그래서 4.3(archive)은 이 5개가 처리될 때까지 보류이며 a187 범위에서 가드를 넓히지 않음. (Manager 문장 중 「마지막 archive 시점」 부분은 저장소 실측과 달라 이 기록에서 뺐음 — 2026-11-16 확인해 아래 「마지막 archive 시점 — 확인」 에 사실로 적음.)
- 미실행: 전체 gate·전체 시험·공식 API.
- gstack 리뷰: `not-applicable` — 문서·증거 기록뿐인 간단한 작업이라 생략(사용자 지시 2026-11-15).

## 4.2 반영 경계 — 2026-11-15

- 운영 코드(GREEN `305733da1`, 2026-11-12T20:18:10+09:00)는 a187 worktree 에서 커밋됐고, 프로브가 `go run` 하는 공유 체크아웃에는 2026-11-13 00:12:56 +0900 `be90ea3e7` fast-forward 로 들어옴(reflog). 그 뒤 첫 프로브 기동은 `tossos-a128-kr.service` 2026-11-13T08:50:42+09:00(journal).
- 원장 행 교차 확인: 영문 섞인 KR 종목의 `top_of_book: …SYMBOL_INVALID…` 결손이 있는 마지막 회차 2026-11-12T14:20:00+09:00, 없는 첫 회차 2026-11-13T09:10:00+09:00(경계 뒤 그 문구 0건). 입력 판본 불변. 상세는 a128 `review.md` 「a187 반영 경계」.
- a184 브랜치(`feat/a184-kr-jev-warning-live-lane` @ `2a12b2449`)는 a187 의 네 파일(`strict_minute_candles.go`, `a187_kr_alphanumeric_test.go`, `a187_order_policy_test.go`, `strict_quote_reads_test.go`)이 main(`be90ea3e7`)과 내용이 같음(`git diff` 0줄). 알림은 Manager 가 전달함.

## 마지막 archive 시점 — 확인 (종결 팀메이트 Opus, 2026-11-16)

위 「Manager 재검증 (3.4)」 의 「Manager 확인 대기」 를 저장소 실측으로 닫음. 근거는 `openspec/changes/archive/` 목록, `git log`, `git merge-base --is-ancestor`, a179 `review.md`.

- 정적 가드 5개가 빨갛던 동안(a190·a189 archive, 2026-11-16 이전) 마지막 archive 는 `2026-10-31-a179-compare-jev-input-profiles`(커밋 `9abcb631a`, 2026-10-31 01:57:58 +0900)였음.
- 가드를 깬 a166 커밋 `9c4bdf4de`(2026-10-31 04:16:34)·`3dc0c776b`(04:55:21)는 a179 archive 뒤이고 그 커밋의 조상이 아님. 위에 적은 `78d777cfe`(04:55:26)·`4d5884cbc`(05:26:28)도 같음.
- a087 커밋 `5c0aeb2b8`(2026-10-29 20:27:39)는 a179 archive 의 조상임. 다만 a179 는 전체 시험·gate 없이 archive 됨(a179 `review.md` 13·39행 「Full repository tests/gates: not run by explicit user scope」).
- 결론: 가드를 깬 커밋들(a166 `3dc0c776b`·`9c4bdf4de`, a087 `5c0aeb2b8`) 뒤로 **완료 gate 를 거친 archive 가 없었음** — 그래서 그동안 main 전체 시험이 archive 경로에서 돌지 않았음. 그 뒤 첫 archive 는 2026-11-16 의 a190(`ff8282056`)·a189(`36560a816`)이고 둘 다 gate 11/11 을 통과함(a189 가 다섯 가드를 고침).

## 운영 반영 (2026-11-16)

- 2026-11-16 KST 06:15 사용자 승인으로 main `7d1e3163a` 이미지(`tossos:a187-strict-quote-readers-accept-kr-alphanumeric-symbols-7d1e3163a`, digest `sha256:e710978e9f1f…`)로 두 컨테이너를 교체함. 롤백 핀은 a176 이미지(둘 다 journal schema 36).
- 직후 확인 통과: healthy · engine run 1 · user_version 36 · mutation_attempts 증가 0 · reconcile.clean.
- Manager 전달 사실을 옮긴 기록이며 종결 팀메이트는 docker·compose·`make image` 를 실행하지 않았음.

## 종결 — 2026-11-16 (종결 팀메이트 Opus)

권한: 사용자 지시(2026-11-15·16, Manager 전달) — main ff 병합·origin/main push·`make gate`·archive.

- 착지 기록: main `36560a816`(a189 archive 뒤)에서 gate ⑤ 단독 판정은 base `6dbb752f6` → 작업 트리 창에 형제 커밋 66개가 들어 required 9, rc 1(a189 의 시험 함수 7개 — `cmd/tossctl/console_test.go:TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` 외 6). 이것은 쌓인 change 의 창 문제이므로 `check_analysis.py --change a187-… --record-landing` 으로 좁힘 — 도구가 이 change 의 번들로 계산한 착지 `be90ea3e7ed673c4992eda2b584eb65c1d562cf6`(저자가 고른 값 아님).
- gate 격리 worktree 를 세울 「완료 커밋」: `landed-commit.txt` 를 담은 이 종결 커밋. 근거 — 기록은 HEAD 커밋에서 읽으므로 그 커밋이어야 ⑤ 가 a187 자기 창(base→착지)만 보고, ⑥~⑪ 은 저장소 전체 검사라 실제 배포될 현재 main(a189 의 가드 수리 포함) 위에서 돌아야 함. a187 의 옛 코드 커밋(`be90ea3e7`·`7d1e3163a`)에서는 ⑦ 이 다섯 가드로 실패하고 ⑪ 이 a190 이전 판정을 씀.
- 착지 기록 커밋 뒤 gate ⑤ 단독: base `6dbb752f6` → landed-commit `be90ea3e7`, required 2, evidence complete, rc 0.
- 격리 gate worktree `~/tossos-work/a187-close`(ext4, detached `5141f6245`): `make sdd-infra` rc 0 → `sdd_sync.py --no-gbrain` rc 1(advisory codegraphcontext 단계, fingerprint 기록됨) → `codegraph status` 0.44초·0.30초 → `make sdd-check` rc 0.
- `make gate`(11:31:37~12:09:35 KST, 전후 HEAD `5141f6245` 동일·dirty 0): **GATE PASS 11/11, rc 0** — ① OK · ② OK · ③ OK(짝 없음) · ④ OK · ⑤ OK(required 2) · ⑥ OK · ⑦ OK · ⑧ OK · ⑨ OK · ⑩ OK · ⑪ OK(「(가) OK: … strict valid」, 「(나) OK: 저장소 전체(strict 없음) 오류 0」, advisory 「저장소 전체 strict 7 passed, 51 failed (58 items)」). 원문 `analysis/gate/`.
- 그 뒤 archive·Story 경로(a187, 그리고 a189 Story 의 `found_by.record` 가 가리키던 이 change 의 파일)·PM check(같은 커밋).
