**제약 위반: `~/.codex/memories/MEMORY.md`를 한 번 검색했습니다. 검색 결과는 리뷰 판정에 사용하지 않았습니다.**

**BLOCK — 범위 선택·identity 대조는 유지되지만, 공유 slice를 통해 대조 원본까지 교체할 수 있어 봉인 종결을 인정할 수 없습니다.**

아래 반례는 **코드 추론이며 실행 재현하지 못했습니다.** 지정 사본 생성 명령이 `mktemp: … Read-only file system`으로 실패했습니다. 생성된 사본·삭제 대상 없음. 테스트·변이 실행 없음. 기존 RED·변이 원장은 읽었지만 독립 실행 증거로 취급하지 않았습니다.

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | **P0** | dispatch 쪽 권한 복사본의 항목 교체가 loader의 대조 원본까지 바꾼다. 조립이 선택하지 않은 유효 제안으로 `err=nil`에 도달하는 경로가 남는다. | `entries`는 slice이고 loader 생성자는 그대로 저장한다. 조립은 같은 권한을 loader·dispatch·assembly에 공급한다. [권한 타입:146](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_proposal_authority.go:146), [loader 저장:202](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_account_first_leg_authority.go:202), [공유 배선:333](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_entry_supervisor.go:333) | loader가 보유하는 조립 결과를 외부 가변 목록과 분리하고, **loader 생성 후 외부 복사본 교체**를 거절하는 시험 추가. |
| 2 | **P1** | strategyflow census는 함수 별칭 호출·포인터 경유 봉인 쓰기를 놓친다. 기본 빌드의 추가 주조 경로를 닫지 못한다. | 호출 검사는 `CallExpr.Fun` 이름이 직접 `sealProposalResult`인 경우만, 쓰기는 좌변이 직접 `.proposalSeal`인 경우만 센다. 동결 대상은 세 함수뿐이다. [쓰기·호출 검사:169](/tmp/claude-1000/a112-review-686b94e4/internal/strategyflow/seal_census_test.go:169), [동결 목록:228](/tmp/claude-1000/a112-review-686b94e4/internal/strategyflow/seal_census_test.go:228) | 봉인 함수의 값 취득·필드 주소 취득도 통제하고, 간접 호출/쓰기 변이 추가. |
| 3 | **P2** | 두 범위 시험은 올바른 항목 선택을 증명하지 못한다. 개수 관문이 identity 비교 전에 반환한다. | 시험은 두 항목을 넣고 개수 오류만 요구한다. `entries[0]`을 잘못 골라도 `len(entries)!=1`에서 동일 오류다. [시험:160](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/a112_first_leg_owner_scope_seal_test.go:160), [관문 순서:223](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_account_first_leg_authority.go:223) | 선택 함수를 직접 검증해 반환 identity를 확인. “위치 선택이면 identity 거절”이라는 시험 설명 수정. |
| 4 | **P2** | 실행 증거 하네스가 하위 시험 실행을 증명하지 않으며 변이 하네스의 GREEN 판정에도 연결되지 않았다. | 이름 검증기는 `/`가 있는 사건을 버린다. 다섯 축이 모두 skip되어도 부모 PASS면 통과 가능하다. 변이 하네스는 여전히 종료 코드 중심이다. [이름 검증:37](/tmp/claude-1000/a112-review-686b94e4/openspec/changes/a112-run-four-strategy-families-independently/analysis/harness/verify_named_tests.py:37), [변이 판정:294](/tmp/claude-1000/a112-review-686b94e4/openspec/changes/a112-run-four-strategy-families-independently/analysis/harness/a112_lot_mutate.py:294) | 필수 하위 시험별 PASS를 요구하고 대조군·변이 실행 판정에 연결. |

**#1의 구체적 반례**는 기존 fixture로 다음처럼 구성할 수 있습니다. 아래 코드는 실행하지 않은 재현안입니다.

```go
fixture := newFirstLegIdentityFixture(t)
loader := fixture.loaderWith(fixture.proposals)
winner := fixture.proposals.kr.entries[0].authority.Proposal()

twin := fixture.krProposal(t, winner.Lineage.CampaignID, "100", "80", "120")

dispatchCopy := fixture.proposals.kr
dispatchCopy.entries[0].authority = twin // loader의 backing array도 변경

issuance, err := loader.collectStrategyFirstLegAuthority(
    context.Background(), a112Accepted(t, twin),
)
```

`krProposal`은 계좌·시장·종목·수량·시각을 유지하면서 가격을 바꾸는 기존 도구입니다([시험 helper:131](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_first_leg_identity_backstop_test.go:131)). 교체 뒤에는 다음이 성립합니다.

- 범위 선택이 `twin`을 고르고, identity 비교도 `twin` 대 `twin`이 됩니다.
- 항목 수는 하나이며 위험 범위·포지션 세대는 그대로입니다.
- 이후 검사는 원래 조립이 선택한 가격과 비교하지 않습니다. 반환 가격도 교체된 권한에서 읽습니다([가격 추출:256](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_account_first_leg_authority.go:256), [발급 반환:279](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_account_first_leg_authority.go:279)).

따라서 정상 대조군의 나머지 조건을 유지하면 **교체된 손절 가격으로 `err=nil` 반환이 코드상 도출됩니다.** 실제 주문 체결까지 검증했다는 뜻은 아닙니다. 기존 다섯 축 시험은 새 목록을 만들어 loader에 주입하므로 이 공유 참조 조합을 다루지 않습니다([시험 구성:72](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/a112_first_leg_owner_scope_seal_test.go:72)).

**#2의 간접 호출 반례**는 기존 공개 함수 `Evaluate`의 본문을 다음 형태로 바꾸는 것입니다.

```go
mint := sealProposalResult
return mint(evaluateWith(request, strategyrouter.RouteSet, proposalRegistry()))
```

공개 서명은 그대로이고, 동결된 세 함수도 바뀌지 않으며, 직접 `sealProposalResult(...)` 호출도 추가되지 않습니다. 현재 census의 세 검사가 이 새 주조 경로를 검출하지 못하는 구조입니다. **전체 테스트가 GREEN이라는 주장은 하지 않습니다.** 또한 재고정 검사는 어느 review 문서에 digest 문자열이 존재하는지만 확인하므로 독립 리뷰 완료까지 증명하지 않습니다([review 결속:270](/tmp/claude-1000/a112-review-686b94e4/internal/strategyflow/seal_census_test.go:270)).

필수 항목별 판정:

1. **위조의 발급 도달: BLOCK.** #1의 대조 원본 공유 참조 경로가 남습니다.
2. **범위 선택 → identity 순서: 조건부 적합.** 선택 함수 자체는 identity를 읽지 않지만, 원본이 함께 교체되면 대조가 무력화됩니다([선택:29](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_first_leg_owner_scope.go:29)).
3. **A-lite: 비활성 시장 경로는 보존.** 활성 시장에서도 digest가 종목·계보 identity만 담아 같은 계보의 조건 교체를 보지 못합니다. #1과 결합하면 최종 방어도 사라집니다([비활성 반환:40](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_dispatch_handoff.go:40), [digest:14](/tmp/claude-1000/a112-review-686b94e4/internal/app/engine/strategy_proposal_set_digest.go:14)).
4. **strategyflow 주조 폐쇄: 미충족.** #2의 간접 경로가 남습니다.
5. **실행 증거: 제한적.** 부모 시험 PASS 확인은 개선이지만 하위 축·변이 실행의 보장은 아닙니다.
6. **토글 OFF = upstream: 이번 diff에서 새로 허용되는 입력은 발견하지 못함.** 동일한 고정 입력 기준으로 기존 개수·identity 조건은 남고 범위 거절만 추가됐습니다. #1은 이번 변경이 새로 만든 확대가 아니라 **종결하지 못한 기존 구멍**입니다. 전체 upstream 동등성은 실행 미검증입니다.

Recommendation: BLOCK — 대조 원본의 공유 참조를 차단하고 #1 재현을 거절하는 시험 및 간접 주조 변이를 확인하기 전에는 6.2.0 종결·5.2.2.2 개수 관문 제거를 승인하지 마십시오.
