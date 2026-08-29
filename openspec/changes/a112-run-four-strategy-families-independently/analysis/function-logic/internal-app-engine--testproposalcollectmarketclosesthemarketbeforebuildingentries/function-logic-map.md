# Function Logic Map: TestProposalCollectMarketClosesTheMarketBeforeBuildingEntries (base 이름, 현재 트리에 없음)

- Source: `internal/app/engine/strategy_proposal_ambiguity_test.go` — **base revision**, `aeeb209e`
- Base source SHA-256: `3d60a03e9e4c85be223883e1885f496dc48eb7eb1a53b66622f0014b95303794`
- Signature: `TestProposalCollectMarketClosesTheMarketBeforeBuildingEntries(params=1, results=0)`
- Source range at base: `26:1`–`109:2`
- AST evidence: `ast.json` (`revision: base`); AST 분기 17개.
- Risk scan: `risk-pattern-report.md`.
- Lot: a112 L5 — 태스크 5.4.

## 이 lot 이 무엇을 했는가 — 이 함수와 파일은 **삭제됐다**

이 테스트는 코드를 실행하지 않았다. `go/parser` 로 `collectMarket` 을 파싱해
"`batch.Ambiguous` 를 부르고 `return` 하는 if 문이, 첫 `entries = append` 보다
**앞에** 있다"는 *구조* 를 확인했다. 태스크 5.4 는 `Ambiguous` 호출을 없애고
보정 중재로 대체했으므로 그 구조는 더 이상 존재하지 않는다.

## 왜 구조 검사가 있었나 — 그때는 행동 검사를 쓸 수 없었다

지키려던 성질은 "한 종목의 fail-closed 가 시장 수준의 fail-open 이 되면 안 된다"였다.
그것을 행동으로 재려면 한 종목이 여러 가족 제안을 낸 배치를 만들어야 하는데,
`strategyproposal` 의 테스트 seam 은 종목당 제안 하나만 담을 수 있었다
(`ProductionBatchAuthorityForTest` 의 map key 가 종목이다). 그래서 구조로 갈음했다.

## 무엇이 대신 들어왔는가 — 같은 성질의 **행동** 검사

`ProductionBatchAuthorityMultiLaneForTest` 가 종목당 여러 제안을 담을 수 있게 되면서
이 성질을 직접 실행해 잴 수 있게 됐다. `a112_arbitration_test.go` 의
TestARefusedArbitrationClosesTheWholeMarketRatherThanReleasingTheOtherSymbol 은
KR 에 종목 두 개를 두고 첫 종목의 중재를 동률로 닫은 뒤,

- 그 시장이 `Ready=false`, `Reason=ARBITRATION_REFUSED`, `ArbitrationRefusal=ARBITRATION_SCORE_TIE` 이고
- `entries` 가 0개이며 `ResultAuthority().kr.ready` 가 거짓이고
- **다른 시장(US)** 은 그대로 진행되는지

를 확인한다. 뮤테이션 M-E1(거절을 `continue` 로 바꿔 그 종목만 빼기)은 이 테스트에
KILLED 됐다 — 즉 이 테스트는 구조가 아니라 결과로 같은 성질을 고정한다.

구조 검사를 함께 남기지 않는다. 같은 판단을 두 곳에 적으면 한쪽은 언젠가
실행되지 않는 죽은 검사가 되고, 죽은 검사는 지켜 주는 것이 없다.

## Inputs and invariants

- 입력은 base signature 그대로다(`*testing.T` 하나, 파일 시스템에서 소스를 파싱).
- 불변식: 대상 성질은 유지된다. 재는 방법만 구조에서 행동으로 바뀌었다.

## Branches and early returns

base 리비전의 17개 분기는 모두 AST 탐색 보조(파싱 실패, 노드 종류 판별, 위치 비교)와
그 결과에 대한 `t.Fatal` 단언이다. 어느 분기도 production 코드를 실행하지 않는다.
전체 열거는 `ast.json` 에 있다.

## Calls and live bindings

base 리비전의 호출은 `go/parser.ParseFile`, `go/ast.Inspect`, `token.NoPos` 비교와
`t.Fatal`/`t.Fatalf` 뿐이다. production 심볼과의 유일한 결합은 파일 이름 문자열
`"strategy_proposal_authority.go"` 과 메서드 이름 `"collectMarket"`, `"Ambiguous"` 이었다.

## State mutations and fallbacks

없음. 소스 파일을 읽기만 하며 디스크·저널·브로커에 쓰지 않는다.

## Safety conclusion

- 테스트 전용 함수다. production 동작을 만들지 않는다.
- 잃은 커버리지는 없다: 같은 성질을 행동으로 재는 테스트가 들어왔고, 그 테스트가
  fail-open 뮤테이션을 실제로 죽인다.
