BLOCK

읽기 전용 코드 리뷰입니다. 시험·변이는 실행하지 않았으며, 아래 “생존”은 코드상 판정입니다. 전용 사본은 생성하지 않았고 삭제 대상도 없습니다. 금지된 다른 리뷰 파일과 운영 데이터는 열지 않았습니다.

아래 경로 약칭: `E` = `internal/app/engine/`, `H` = `internal/strategyhandoff/`, `C` = `openspec/changes/a112-run-four-strategy-families-independently/`, `M` = `C/analysis/measurements/lot-5.6.2-5.2.2/mutation-5.2.2.1.tsv`.

| # | 등급 | 주장 | 증거(파일:줄 · 실행 여부) | 권고 |
|---|---|---|---|---|
| 1 | **P1** | 생산 전달 구조 핀은 “모든 범위 전달”을 보증하지 않는다. 호출을 남긴 채 두 번째 범위부터 버리는 변이가 통과한다. | `E/a112_owner_scope_handoff_test.go:201–223`은 호출식 이름·개수만 검사. `E/strategy_entry_supervisor.go:552–566`의 callback 실행 횟수·데이터 흐름은 검사하지 않음. 해당 함수 행동 시험 부재도 `C/analysis/function-logic/internal-app-engine--context.runproductionstrategymarketcycle/branch-test-map.md:5`에 명시. **정적 판정**. | 실제 생산 사이클을 통해 범위별 전달 순서·횟수·오류 전파를 검증하고 아래 우회 변이를 추가. |
| 2 | **P1** | census와 공개 표면 동결을 동시에 우회하는 새 주조 문을 만들 수 있다. | `E/strategy_dispatch_handoff_guard_test.go:677–682`는 수신자 있는 함수를 제외. `H/escape_test.go:77–80`은 비공개 타입 메서드를 제외하고, `:104–108`은 공개 변수를 타입 없이 `"var"`로만 기록. **정적 판정**. | 공개 변수의 실제 타입과 그 타입을 통해 접근 가능한 메서드까지 고정. 아래 반례를 회귀 시험으로 추가. |
| 3 | **P2** | RED 생략 시험 중 **오늘-동등성 대조 부분**에는 독립적인 변이 대체 증거가 없다. | `M:16`의 F14는 시험 전체 실패만 기록. `E/a112_owner_scope_handoff_test.go:107–114`의 본군 단언만으로도 F14를 검출할 수 있어 `:118–141` 대조의 판별력을 입증하지 못함. | 본군의 기존 거절은 유지하면서 단일 범위 성공을 깨는 변이를 추가하고 실패 단언을 기록. |
| 4 | **P2** | 소유자 범위 네 축 중 계좌·시장 구분 자체는 이번 시험들이 검증하지 않는다. | `H/handoff.go:262–266`; `H/owner_scope_test.go:20–75`는 다른 종목·다른 세대·동일 계좌의 표기 차이만 시험. 계좌를 상수로 바꾸는 변이는 F07과 다르며 해당 시험들을 통과할 수 있음. | 종목·세대가 같고 계좌만 다른 양성 대조 추가. 시장 축은 함수의 “단일 시장 입력” 전제와 맞춰 검증하거나 전제를 명시. |

**1. 오늘-동등성 핀: fixture 안에서는 B2 개수 조건 귀속이 코드상 타당하지만, 생산 전체의 유일한 방어라는 증명은 아니다.**

두 실행이 문자 그대로 동일한 조립은 아닙니다. fixture를 다시 만들므로 임시 원장·Guardian·스파이·서명 키가 달라집니다(`E/strategy_dispatch_cycle_test.go:241–290`, `E/strategy_risk_authority_test.go:130–138`). 그러나 고정 시각과 같은 입력으로 위험·환율·계좌·일정 준비 상태를 구성하고, 두 범위 본군은 **제안만** 교체합니다(`a112_owner_scope_handoff_test.go:83–92`). 따라서 이 차이가 B2의 다른 조건을 거짓으로 만든다는 근거는 없습니다.

또한 원래 제안을 첫 항목으로 보존합니다(`:41–42`). B2 개수 조건을 제거하면 첫 범위는 기존 제안과 identity가 일치하고, 대조와 같은 하류 경로를 진행합니다. 두 번째 범위는 `strategy_account_first_leg_authority.go:223–224`의 identity 검사에 걸립니다. **fixture에서 주문 0이 다른 관문으로 반드시 유지된다고 볼 수 없습니다. 첫 범위 주문이 가능해지는 것이 코드상 예상 결과입니다.** 다만 F14 원장에는 실패 단언이 없어 실행 증명은 아닙니다.

반면 실제 생산 조립은 다릅니다. 두 제안에서 `ResultAuthority()`가 준비되지 않은 결과를 만들고(`strategy_proposal_authority.go:189–191`), 위험 권한도 준비되지 않으며(`strategy_risk_authority.go:176–177`), 계좌 권한 역시 개수로 거절합니다(`strategy_account_first_leg_authority.go:155–156`). 이 권한들을 실제로 재수집합니다(`strategy_entry_supervisor.go:329–336`). **생산에서는 개수 조건만 지워도 B2의 위험·계좌 조건으로 주문 0이 유지됩니다.** fixture는 이 재수집을 의도적으로 우회한 국소 시험입니다.

**2. RED 생략 시험의 변이 대체: 대부분 존재하지만, 대조 부분과 구조적 우회는 남는다.**

| RED 생략 항목 | 원장 대응 | 판정 |
|---|---|---|
| 모든 범위 전달 | F09, `M:11` | 해당 시험만 실패 |
| 첫 오류 뒤 중단 | F10, `M:12` | 해당 시험만 실패 |
| 생산 전달 구조 핀 | F11, `M:13` | 해당 시험만 실패하나 아래 우회는 미검출 |
| 종목 표기 정규화 | F06, `M:8` | 대응 있음. 원장은 상위 시험 이름만 기록 |
| 계좌 표기 정규화 | F07, `M:9` | 대응 있음. 원장은 상위 시험 이름만 기록 |
| 활성화됐지만 닫힌 시장 | F13, `M:15` | 해당 시험만 실패 |
| 오늘-동등성의 단일 범위 대조 | F14, `M:16` | **대조 부분을 검증했다는 증거 없음** |

원장 밖 구조 우회 예시는 다음과 같습니다. 기존 callback 본문은 그대로 두고 앞에 횟수 제한만 추가합니다.

```go
seen := 0
return deliverEachStrategyHandoff(
    fresh.proposals.forMarket(market).dispatchHandoffs(),
    func(delivered strategyhandoff.Delivered) error {
        seen++
        if seen > 1 {
            return nil
        }
        // 기존 callback 본문
    },
)
```

구조 핀의 `delivers == 1`, `singular == 0`이 그대로이고, dispatch 호출 census도 한 자리 그대로입니다(`strategy_dispatch_handoff_guard_test.go:789–812`). 헬퍼 직접 시험은 생산 callback을 실행하지 않으므로 이 변이를 잡지 못합니다. **현재 구현에 이 결함이 있다는 주장이 아니라, 유일한 생산 배선 핀이 깨진 구현을 허용한다는 P1입니다.**

**3. census 소스 유도: 현재 두 함수는 찾지만, 모든 경계 생성 문을 찾지는 못한다.**

새 공개 함수가 `Delivered`, `any`, 별칭 타입을 반환하거나 출력 포인터·callback으로 값을 내보내면 census가 놓칠 수 있습니다. 새 공개 메서드나 함수 변수도 마찬가지입니다. 다만 **새 공개 이름·서명 변경**은 대체로 `escape_test.go:119–136`의 표면 동결이 먼저 거절하므로, 이것만으로 두 검사 공통 우회라고 판정해서는 안 됩니다.

두 검사 모두 놓치는 구체적 모양은 **기존 공개 변수 + 비공개 구체 타입의 공개 메서드**입니다.

```go
type deliveryError struct{ error }

func (deliveryError) Mint(r strategyflow.Result) Handoff {
    return Handoff{selected: []strategyflow.Result{r}, pending: 1}
}

var ErrNoDelivery = deliveryError{
    errors.New("strategyhandoff: delivery body is nil"),
}
```

엔진은 `strategyhandoff.ErrNoDelivery.Mint(result)`를 호출할 수 있습니다.

- 기존 공개 변수 이름과 `"var"` 표면은 유지됩니다.
- 비공개 수신자 `deliveryError`의 `Mint`는 동결 검사에서 제외됩니다.
- census도 메서드를 제외하므로 `Mint`를 문으로 등록하지 않습니다.
- 기존 import만 사용하므로 import 허용 목록도 그대로입니다(`H/dependency_closure_test.go:25–31`).

이는 **경계 주조 검사 우회**이며, 하류 identity 방어까지 무력화하거나 실제 주문을 보장한다는 뜻은 아닙니다.

**4. 추가 P0: 검토한 변경에서 확인하지 못했다.**

OFF 경로는 기존 `dispatchHandoff()` 하나를 반환하고, 전달 헬퍼는 동일한 `Deliver(body)` 오류를 그대로 반환합니다(`strategy_dispatch_handoff.go:38–40,58–64`). 생산 생성자는 실제 `c.Entry`를 요구·전달하고(`strategy_entry_supervisor.go:394–409`), B13은 게이트를 닫은 뒤 계속 실행합니다(`:944–948,997–1005`). 노출을 늘리지 않는 주문 변경은 EntryGate를 건너뜁니다(`internal/execgw/gateway.go:851–855`). 승인된 이월 범위 B14를 새 결함으로 재분류하지 않았습니다. **실행에 의한 안전 루프 생존 검증은 이번 리뷰에서 하지 못했습니다.**

Recommendation: BLOCK 유지 후 생산 전달 우회·공개 변수 경유 주조 반례를 검출하고 대조 변이를 보강할 것 because 현재 착지 증거에는 깨진 구현을 통과시키는 P1 검증 공백이 있다.
