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
