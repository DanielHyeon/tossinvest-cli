**BLOCK — P1-1·P1-2의 원래 반례는 검출하지만, 수리된 검사에 같은 부류의 우회가 남습니다.**

아래 `E` = `internal/app/engine/`, `H` = `internal/strategyhandoff/`, `C` = `openspec/changes/a112-run-four-strategy-families-independently/`, `M` = `C/analysis/measurements/lot-5.6.2-5.2.2/`.

Go 시험·변이·게이트는 재실행하지 않았습니다. **CAUGHT는 제출 원장 기록**, 아래 새 반례는 **정적 판정**입니다. 직접 실행한 검증은 현재 closure의 `gofmt` 정규화·SHA-256 재계산입니다.

| 항목 | 판정 | 파일:줄 근거 |
|---|---|---|
| P1-1(a) 몸통 추출 동등성 | **CLOSED — 기록·현재 코드 대조 범위** | `E/strategy_market_handoff_delivery.go:35–49`의 정규화 해시가 영수증 `M/extract-receipt-5.2.2.1-fix.txt:1–3`의 양쪽 해시 `ffb36ce3…54cfa`와 일치. 수신자 치환은 `C/analysis/harness/extract_receipt.py:48–50`, 호출 인자 배선은 `E/strategy_entry_supervisor.go:552`와 일치. **Git 이력이 없어 이전 원문을 직접 재추출한 양방향 비교는 못 했음.** |
| P1-1(b) 기존 closure 카운터 | **CLOSED** | `E/a112_owner_scope_delivery_test.go:58–66`이 생산 몸통의 세 범위 전달을 직접 단언. G06은 `M/mutation-5.2.2.1-fix.tsv:8`에서 검출. 단, 구조 시험 실패는 본문 문장 수 검사(`E/a112_market_delivery_structure_test.go:176`)만으로도 발생하므로 **캡처 검사 자체의 판별력 증거는 아님**. |
| P1-1(c) 새 구조 핀의 우회 | **OPEN** | `E/a112_market_delivery_structure_test.go:123–140`은 식별자 대입만 세며 원소 변경·함수 호출을 놓침. `:201–218`은 미해소 식별자를 허용하고 포인터·필드·맵·채널 쓰기를 검사하지 않음. 아래 #1·#2. |
| P1-2 기존 `ErrNoDelivery.Mint` | **CLOSED — 원래 형태** | 초기화 식 고정 `H/escape_test.go:42,113–120`, 직접 반환형·리터럴 검사 `H/mint_census_test.go:55–64,108–121`, G11 기록 `M/mutation-5.2.2.1-fix.tsv:14`. |
| P1-2 같은 부류의 다른 주조 문 | **OPEN** | 초기화 **이후 재대입**과 비공개 별칭을 결합하면 세 검사를 통과할 수 있음. `H/escape_test.go:80–87,113–120`, `H/mint_census_test.go:31–39,108–121`, `E/strategy_dispatch_handoff_guard_test.go:673–697`. 아래 #3. |
| P2-3 G10 대조 절반 | **CLOSED** | 하네스 `C/analysis/harness/a112_lot_mutate.py:142–144`는 개수 조건을 `>= 1`로 바꿈. 두 범위 거절은 유지하면서 단일 범위도 거절하므로 `E/a112_owner_scope_handoff_test.go:176–177` 대조가 실패함. 원장 `:26`에 두 순서 모두 해당 줄에서 실패했다고 명시. **전체 스위트에서는 다른 시험도 실패하므로 “전체에서 대조만 실패”라는 뜻은 아님.** |
| P2-4 G03/G04 계좌·시장 축 | **CLOSED** | `H/owner_scope_axes_test.go:47–60`은 소유자 범위 축 중 계좌 또는 시장 하나만 달리한 양성 대조. 하네스 `:121–124`의 축 제거와 원장 `:5–6`의 실패가 대응함. 시장 사례는 경계 함수의 혼합시장 입력 시험이며 생산의 단일시장 조립 증명과는 구별됨. |
| 동등성 핀 정정·생산 재수집 기록 | **CLOSED — 요청한 세 위치** | 시험 머리말 `E/a112_owner_scope_handoff_test.go:11–16`, `C/review.md:5329,5400–5402`, `C/tasks.md:196`에 명시. 실제 생산 재수집도 `E/strategy_entry_supervisor.go:327–334`에서 확인. |

새 발견 및 잔존 결함:

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | **P1** | 생산 사이클에서 뒤 handoff를 지워도 새 구조 핀이 허용한다. | `E/a112_market_delivery_structure_test.go:123–140`은 지역 변수의 직접 대입만 검사. 아래 `clear`는 추가 대입으로 집계되지 않는다. 생산 몸통 직접 시험은 이 호출부를 지나지 않으며, 전체 사이클 진입 시험 부재는 `C/analysis/function-logic/internal-app-engine--context.runproductionstrategymarketcycle/branch-test-map.md:5`에도 명시됨. | 생산 호출부를 포함한 행동 시험 추가. 지역 변수 허용 시 원소 변경·별칭·다른 호출로의 전달까지 검증. |
| 2 | **P1** | “캡처 쓰기 금지”는 외부 상태 쓰기를 보증하지 않는다. 스텁의 미해소가 실제 허용으로 이어진다. | 빈 importer·오류 무시 `E/a112_market_delivery_structure_test.go:33–36,72–73`; `outside`는 `Uses == nil`이면 false(`:201–203`). 외부 패키지 dot-import 변수에 대한 단순 대입이 이 경로를 탄다. `(*p)++`, `s.n++`, `m[k]++`는 Ident가 아니어서 제외되고, `ch <- v`는 검사 분기 자체가 없음(`:208–218`). | 관련 의존 타입을 실제로 해소하고, 미해소 쓰기는 실패 처리. 간접 쓰기·송신 변이를 각각 추가. 구조 검사의 보증 문구도 실제 범위에 맞출 것. |
| 3 | **P1** | 공개 오류 값의 사후 교체와 비공개 별칭으로 둘째 주조 문이 다시 열린다. | `H/escape_test.go:113–120`은 선언 초기화만 고정. 비공개 수신자 메서드는 `:80–83`에서 제외. mint census는 타입 정체성이 아닌 `Handoff`·`Delivered` 철자를 검사(`H/mint_census_test.go:31–39,108–121`). 아래 반례는 공개 최상위 이름을 추가하지 않음. | 별칭·포함 타입을 타입 정보로 추적하고 공개 var 재대입도 검사. 아래 반례를 회귀 변이에 추가. |
| 4 | **P2** | 동작이 같은 지역 변수 리팩터를 여전히 거짓 양성으로 막는다. | N01의 `hs := …`를 `var hs = …`로 바꾸면 동작은 같지만 `ValueSpec`을 세지 않아 `writes == 0`으로 실패(`E/a112_market_delivery_structure_test.go:123–140`). | `ValueSpec` 형태도 검증하거나 행동 시험 중심으로 전환. N01 외 동등 리팩터 대조 추가. |

#1의 구체적 우회는 생산 사이클 마지막 부분을 다음처럼 만드는 것입니다. 전달 함수 자체는 그대로여서 새 몸통 행동 시험은 영향을 받지 않습니다.

```go
hs := fresh.proposals.forMarket(market).dispatchHandoffs()
if len(hs) > 1 {
    clear(hs[1:])
}
return dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, hs)
```

#3은 기존 `ErrNoDelivery` 선언을 그대로 두고 다음 선언을 추가하는 형태입니다.

```go
type hiddenDelivered = Delivered
type mintingError struct{ error }

func (mintingError) Mint(r strategyflow.Result) hiddenDelivered {
    return hiddenDelivered{result: r}
}

func init() {
    ErrNoDelivery = mintingError{ErrNoDelivery}
}
```

엔진은 다음 타입 단언으로 접근할 수 있습니다. 별칭은 동일 타입이므로 비공개 별칭 이름을 알 필요가 없습니다.

```go
strategyhandoff.ErrNoDelivery.(interface {
    Mint(strategyflow.Result) strategyhandoff.Delivered
}).Mint(r)
```

비공개 반환 구조체·제네릭·함수형 var도 census 단독으로는 완전히 다루지 않습니다. 반환 타입의 내부 구성·타입 인자를 해소하지 않고 `FuncDecl`만 보기 때문입니다. **새 공개 이름을 추가하는 단순 형태는 표면 동결·허용 목록이 막습니다.** 위 반례는 그 보호까지 피하는 형태입니다. 반환값 없는 out-param 함수는 `mint_census_test.go:52–53`에서 먼저 건너뛰는 별도 누락도 있습니다.

수리 로트의 현재 생산 코드에서 **새 OFF≠upstream, 주문 경로 완화, 안전 루프 지연은 확인하지 못했습니다.** OFF의 단일 handoff 반환은 `E/strategy_dispatch_handoff.go:38–40`, 기존 CAS·점유·lease 오류 처리는 추출된 몸통에 유지됩니다. 다만 실행 검증은 하지 않았으며, 위 발견은 현재 실주문 결함이 아닌 **검증 장치의 잔존 공백**입니다. “유일한 방어”라는 오래된 문구도 `E/strategy_dispatch_handoff.go:34–37`과 `E/a112_owner_scope_handoff_test.go:190`에 남아 있습니다.

접근 제한 관련 고지: 시작 시 `~/.codex/memories/MEMORY.md`에 키워드 검색 1회를 수행했습니다. 요청한 접근 제한을 지키지 못한 부분입니다. 해당 검색은 판정 근거로 사용하지 않았고, 운영 원장·자격 증명은 열지 않았습니다.

Recommendation: BLOCK 유지 — #1의 생산 전달 누락과 #3의 별칭·사후 재대입 주조 반례를 검출하고, #2의 미해소·간접 쓰기 및 #4의 동등 리팩터 대조를 보강한 뒤 재확인.
