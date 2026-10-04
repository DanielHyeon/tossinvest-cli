# a112 6.2 봉인 로트 적대 리뷰 — 목소리 A(다섯 축 바깥의 위조), 대상 `686b94e4`

`~/.codex` 아래 파일은 열지도 검색하지도 않았다. 운영 원장·자격 증명·엔진·토글도 건드리지 않았다.

**판정: BLOCK**

봉인은 1차 레그 권한의 기준(`loader.proposals`)을 dispatch 쪽 목록(`fresh.proposals`)과 **같은 backing array** 로 들고 있다. 그래서 dispatch 쪽 목록에 제자리 쓰기를 하면 기준도 같이 바뀌고, 이때 1차 레그 판정은 자기 자신과 비교하게 된다. 로트가 스스로 정한 위협 모형(「이 파도의 중재가 고르지 않은 진짜 봉인 제안을 dispatch」, review.md)에서, 같은 범위 패자가 Guardian precheck 까지 통과했다. 원인은 편집 전에도 있었다. 하지만 이 로트는 "봉인이 서 있다"고 선언하고 5.2.2.2 착수 조건 (1)을 충족으로 처리한다. 이 주장은 착지 상태로 둘 수 없다.

## 발견 표

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | **P1**(로트의 위협 모형 기준으로는 P0 문구 「위조가 1차 레그 발급까지」에 해당. 생산 코드에 그런 쓰기가 0건이고 편집 전부터 있던 구조라 P1로 매김) | **신뢰 뿌리 별칭.** `newProductionStrategyFirstLegAuthorityLoader`(`strategy_account_first_leg_authority.go:202-208`)가 `proposals` 를 값으로만 복사한다. 그래서 `entries` 슬라이스가 조립의 `proposalAuthority`, assembly 의 `proposals`(`strategy_entry_supervisor.go:329-336`), dispatch 가 읽는 `fresh.proposals.forMarket(market)`(`:552`)과 backing array 를 공유한다. 이 상태에서 `fresh.proposals.forMarket(m).entries[0] = <같은 범위 패자>` 처럼 dispatch 쪽에 제자리 쓰기를 하면 loader 필드를 건드리지 않고도 봉인 기준이 바뀐다. 이것은 HANDOFF §4 의 자기 참조 함정이 선택 경로가 아니라 **별칭 경로**로 재현된 것이다. | 사본 실험 `TestRev62AAliasedTrustRoot`: 제자리 쓰기 전 `loser err=production proposal identity changed`. 쓰기 후 `loser err=<nil> issuedCampaign="campaign-rev62-loser"`, `validateStrategyFirstLegAuthority=<nil>`, `guardian precheck err=<nil>`. 비활성 시장 `dispatchHandoffs` 는 `n=1 first=""`(Admitted)로 패자를 그대로 건넨다. 활성 시장에서만 A-lite 가 `HANDOFF_MARKET_CLOSED` 로 막는다. dispatch 몸통(`strategy_market_handoff_delivery.go:32-49`)과 `strategy_dispatch_cycle.go` 는 proposals 에서 `familyActivation` 만 읽고 다른 대조는 하지 않는다. | 생성자에서 `entries` 를 복제한다(`proposals.kr.entries = append([]strategyProposalEntryAuthority(nil), proposals.kr.entries...)`, us 도 같게). 사본에서 이 두 줄을 넣자 같은 실험이 `production proposal identity changed` 로 거절됐고 로트의 봉인 시험 전부가 초록이었다(원복 후 sha256 OK). 별칭 행동 시험 하나와 「제자리 쓰기 뒤에도 거절」 변이를 원장에 추가할 것. |
| 2 | **P1** | **A-lite 와 봉인이 둘 다 못 보는 모양.** A-lite digest(`strategy_proposal_set_digest.go:11-16`, 조립 `strategy_proposal_authority.go:420-423`)는 (종목, `Lineage.Identity`)만 담고 `ExecutionTerms.Identity()` 는 담지 않는다. 그래서 같은 계보에 손절/목표를 재작성한 항목(로트 축 ⑤)을 제자리로 쓰면 활성 시장에서 A-lite 가 통과하고, #1 의 별칭 때문에 봉인도 통과한다. `dispatchHandoffs` 주석의 「권한 값을 통째로 위조하면 … 봉인은 1차 레그 권한 … 이 진다」(`strategy_dispatch_handoff.go:50-52`)와 review.md 의 같은 문장은 **제자리 쓰기에서는 거짓**이다. 참인 것은 슬라이스를 통째로 재할당하는 경우뿐이다. | `TestRev62AAliasedTermsRewrite`: `same.lineage==winner.lineage? true; rewritten.lineage==same.lineage? true; terms differ? true`. 제자리 쓰기 후 `aliased terms rewrite: err=<nil>`, 활성 시장 A-lite `n=1 first=""`(Admitted). 즉 손절·목표가 바뀐 1차 레그가 두 방어를 모두 통과한다. #1 수리 후에는 `err=production proposal identity changed`. | #1 수리로 봉인 쪽이 닫힌다. 주석과 review 문장은 「재할당 위조에 대해서만」으로 정정할 것. A-lite 를 terms 까지 넓힐지는 선택이다(2차 방어이므로 필수 아님). |
| 3 | P2 | **봉투 부가 필드 미봉인.** `collectStrategyFirstLegAuthority` 는 발급값의 `Entry.Currency` 를 `accepted.currency`(`:268`)에서 그대로 가져오고 대조하지 않는다. `validateStrategyFirstLegAuthority` 도 같은 봉투 값끼리 비교한다(`strategy_first_leg_admission.go` `entry.Currency != accepted.currency`). 이 값을 막는 것은 Guardian 하나다. | `TestRev62ACurrencyEnvelopeField`: `collect err=<nil> Entry.Currency="USD"`(KR), `validateStrategyFirstLegAuthority=<nil>`, `guardian precheck err=… CURRENCY_UNRESOLVED (market_currency): KR requires KRW`. 생산에서 accepted 를 만드는 곳은 `validateStrategyFirstLegResult`(`:105`) 하나뿐이다. | 1차 레그 권한 안에서 currency 를 `result.Lineage.Market` 으로 다시 유도한다. 봉인의 「봉투를 믿지 않는다」 원칙에 맞추는 것이고, 지금은 Guardian 이 fail-closed 로 막고 있다. |
| 4 | P2(T) | **범위 선택의 오늘 효력은 0.** `Lineage.Valid()` 가 모든 필드를 해시하므로(`strategyflow/types.go:371-419`) identity 가 같으면 NewOwnerKey 도 같다. 개수 관문이 1인 동안 B3 는 identity 가드가 이미 거절하는 입력만 거절하고, **거절 문구만 바뀐다**. RED 셋도 문구 차이일 뿐 행동 차이가 아니다. | 사본 A/B: 편집 전 함수(`686b94e4^`)로 로트 시험과 내 실험을 돌리면 수락 집합이 같다(대조군만 발급, 나머지는 전부 거절). 실패는 문구 단언 셋뿐이고 `red-6.2-seal.log` 와 일치한다. | 거짓은 아니다(review 가 「판정만 더 엄격」이라 씀). 다만 5.2.2.2 착수 근거로 인용할 때 「봉인의 보호 효력은 개수 관문을 걷는 순간부터 생긴다」를 명시할 것. #1 이 닫히기 전에는 착수 조건 (1) 을 충족으로 보지 말 것. |
| 5 | P2(T) | **5.2.2.2 로 넘어갈 위험: 계좌 권한이 범위에 묶이지 않음.** 계좌 권한은 `entries[0]` 의 종목으로 적재되고(`strategy_account_first_leg_authority.go:153-164`, `len!=1` 이면 거절), `collect` 는 선택된 범위와 계좌 권한의 종목을 대조하지 않는다(`strategyaccount.Authority` 에 Symbol 접근자가 없다). 오늘은 개수 관문 두 개가 가린다. | 코드 읽기(추측 아님: 대조하는 줄이 없다). 실행은 개수 관문 때문에 불가. | 5.2.2.2 설계에 「계좌 권한도 소유자 범위 단위로 다시 유도」를 명시할 것. |
| 6 | P2(T) | 제안 집합 digest 식이 **두 곳에 복사**돼 있다(`strategy_proposal_authority.go:420-423` 인라인, `strategy_proposal_set_digest.go`). 등식 시험은 한 픽스처로만 잰다. 갈라지면 fail-closed(활성 시장이 전부 닫힘)라 안전 방향이긴 하다. | `TestTheProposalSetDigestMatchesWhatTheAssemblyRecords` PASS(대조군 실행). | 조립이 `strategyProposalSetDigest(entries)` 를 부르게 해서 한 곳으로 만든다. |

## 필수 항목 판정

1. **새 위조 축.** 모든 결과는 사본에서 실행한 것이다. 대조군(승자)은 `err=<nil>` 로 발급됐다.
   - **통과(err=nil)한 축:** 신뢰 뿌리 별칭(#1), 별칭을 통한 조건 재작성(#2), currency 봉투 필드(#3, Guardian 이 거절).
   - **거절된 축:**
     - **포지션 세대:** 세대는 identity 해시와 OwnerKey 에 둘 다 들어 있다. 조립 뒤에 세대가 바뀌면 CAS 줄(`:234-237`)이 거절한다.
     - **계좌 표기·종목 대소문자:** 정규화 전후가 갈려도 identity 가 원시 값을 해시하므로 거절된다. 검증을 우회해 identity 문자열만 복사한 경우 collect 는 nil 을 주지만, 발급값은 조립 항목 그대로다(`issuedCampaign="campaign-risk-loader-kr" issuedQty=8`). 이어서 `validateStrategyFirstLegAuthority` 가 `sealed strategy result does not match the accepted result` 로 거절한다.
     - **시장 조합:** KR 결과를 US 자리에 넣으면 선택 실패로 거절된다. 소문자 시장은 `incomplete for market` 으로 거절된다.
     - **다른 파도:** loader 의 쌍과 dispatch 의 쌍은 같은 조립의 같은 변수다(`strategy_entry_supervisor.go:329-336,552`). 파도 사이 경합은 없고, 오히려 너무 같아서 생긴 문제가 #1 이다.
   - **위험·계좌 쌍이 다른 범위를 가리키는 경우:** 위험 쪽은 `:227-229` 가 대조한다. 계좌 쪽은 #5 를 보라.
   - **검증 우회:** 생산에서 accepted 를 만드는 곳은 `validateStrategyFirstLegResult` 하나다(`strategy_dispatch_cycle.go:151` 에서 `admit`).
2. **선택 → 대조 순서를 공허하게 만드는 조합:** 없음. 선택이 accepted 에서 읽는 값을 위조해 다른 항목을 고르게 해도, identity 가 전체 필드 해시이므로 그 항목의 identity 와 같아질 수 없다. NewOwnerKey 는 Trim/ToUpper 로 정규화하고 identity 는 원시 값을 해시한다. 이 어긋남은 선택을 넓힐 수만 있고 대조를 통과시키지는 못한다. 공허해지는 길은 선택 순서가 아니라 **기준의 별칭**(#1)이다.
3. **A-lite 와 봉인 사이:** 둘 다 못 보는 모양이 있다. 제자리 쓰기로 같은 계보의 조건을 재작성하는 경우다(#2). A-lite 는 생산을 바꾸지 않는다. 활성화가 없는 시장은 `dispatchHandoff()` 단일 경로이고, 사본 실험에서도 비활성 시장은 digest 를 보지 않았다(`unactivated: n=1 first=""`).
4. **토글 OFF = upstream:** 성립한다. 편집 전 함수와 A/B 로 비교했을 때 새로 통과하는 입력은 0이고, 수락 집합이 같다. 거절 문구(Detail)만 바뀐다. 새로 거절될 수 있는 입력은 NewOwnerKey 정규화에 실패하는 조립 항목(빈 계좌·종목, 256B 초과, 세대 0)뿐이다. 이들은 원래 위험 범위 대조나 CAS 에서 이미 거절되던 입력이다.

## 실행 기록

- 사본 `/tmp/claude-1000/a112-rev62-TwMO`(리뷰 트리 `cp -a`, `.git` 없음, `GOFLAGS=-trimpath`, 전용 GOCACHE). 무변이 대조군 먼저: `TestTheFirstLegSeal*`, `TestAnActivatedMarket…`, `TestTheProposalSetDigest…`, `TestTheFirstLegBackstop…` 전부 PASS.
- 실험 파일 `internal/app/engine/zz_rev62a_test.go`(사본에만 둠). 실험은 한 번에 하나씩 했다.
  1. 무수정 판정.
  2. 생성자 복제 수리 가설. 원복 후 `sha256sum -c` OK.
  3. `686b94e4^` 함수로 A/B. 원복 후 `sha256sum -c` OK.
- 사본과 GOCACHE 는 삭제했다. `/tmp/claude-1000/a112-rev62-yp9I` 는 내 사본이 아니라 그대로 두었다.
- 참고: 리뷰 도중 실제 저장소의 HEAD 가 `16cb1a1a` 로 바뀌었고 작업 트리 diff 가 있었다. 둘 다 내가 만든 것이 아니다. 나는 `git show|log|diff` 만 썼다.

Recommendation: BLOCK. #1 을 고쳐야 착지할 수 있다. 수리는 생성자에서 `entries` 를 복제하는 것이고, 별칭 행동 시험과 변이를 함께 추가한다. 이어서 #2 의 주석·review 문장을 정정하고 이 로트를 재리뷰한다. 그 전에는 5.2.2.2 착수 조건 (1) 을 충족으로 보지 않는다. #3~#6 은 이월해도 된다.
