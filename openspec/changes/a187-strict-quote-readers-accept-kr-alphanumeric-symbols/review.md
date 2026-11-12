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
- 3.1 변이(`analysis/mutation/mutation.py`, 원문 `mutation-3.1.txt`): 사본 `/tmp/a187-mut-<pid>-…`, `GOFLAGS=-trimpath`, 전용 GOCACHE. 무변이 대조군 GREEN. 6개 모두 CAUGHT — 숫자 전용 되돌림(`TestA187StrictQuoteReadersAcceptKRAlphanumeric`·분봉 수락 시험), 길이 고정 제거(`TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest/kr_symbol_too_short`·`TestA187…StillRefuseOtherKRForms/0101N`·`0101N00`·분봉 거절 시험), 앞 고정 제거(`…StillRefuseOtherKRForms/0101N00`·`_0101N0`), 뒤 고정 제거, 소문자 허용(`…StillRefuseOtherKRForms/0101n0`·분봉 `kr_symbol_lowercase`), 래퍼가 검사를 건너뜀(`TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` 11 갈래).
  - 대조군 1차는 FAIL 이었음(`history-1-control-trimpath.txt`): 정적 가드 3개. 둘은 `-trimpath` 에서 `runtime.Caller` 경로를 못 여는 것, 하나는 아래 base 기존 결함. 셋을 `-skip` 으로 빼고 다시 돌림.
- 3.2(원문 `analysis/verify/`): `go test -count=1 -v` 6개 패키지 범위 — 최상위 PASS 507, 서브 포함 PASS 919, **FAIL 5**, SKIP 0. `go vet` rc 0, gofmt 0줄, `openspec validate --strict` valid.
  - FAIL 5개는 모두 **base 6dbb752f6 에서도 같은 문구로 실패**함(`analysis/verify/base-failures.txt`, base 사본에서 같은 시험을 돌림). 이 change 와 무관하며 고치지 않음:
    1. `internal/official` `TestA112MBUSStaticNoProductCallerOrForbiddenPath`: `forbidden resolved A112 M-B0 references: []string{"internal/accountbudgetruntime/v2_token_test.go: selector official.A112MBUSCandle"}`
    2. `tools/a128-jev-shadow-probe` `TestImportsStayInsideTheAllowlist`: `internal/accountbudget`·`internal/accountbudgetruntime` import 가 허용 목록 밖
    3. `tools/a128-jev-shadow-probe` `TestNoMutatingOrAccountSurfaceIsNamed`: `source uses official.Option, outside the read-only allowlist`
    4. `tools/a140-breakout-signal-shadow` `TestImportsStayInsideTheAllowlist`: 같은 두 import
    5. `tools/a140-breakout-signal-shadow` `TestNoMutatingAccountOrWriteSurfaceIsNamed`: `source uses official.Option, outside the allowlist`
  - 숫자 전용을 고정한 시험은 위 두 사례 말고 6개 패키지에서 더 나오지 않음(실패 5개 모두 무관). 주문(`internal/orderintent`)·a087 쪽 시험은 통과.
- 하지 않은 것: 3.3(다른 계열 리뷰·gstack), 3.4, 4.x(실측·반영). 전체 gate·전체 시험. 공식 API 호출.
- 남은 위험: design Risks 그대로(공식이 영숫자 코드를 호가·현재가에서 받는지 미실측, 프로브 입력의 조용한 변화, 주문 쪽 시장 추론이 영숫자 KR 을 `us` 로 봄). 그리고 main 의 정적 가드 5개가 base 에서 이미 깨져 있음 — 별도 처리가 필요함.
