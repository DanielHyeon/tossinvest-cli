**BLOCK — #1·#2 처분·#4·(T)는 CLOSED. #3은 생산에서 접근 가능한 제네릭 주조 우회가 남아 OPEN.**

이번 재확인에서는 `~/.codex`에 접근하지 않았습니다. 저장소 시험·변이는 재실행하지 않았습니다. 제출 원장과 코드를 대조했고, 아래 반례의 축약 모델만 `go tool compile -o /dev/null /dev/stdin`으로 컴파일했습니다(종료 0, 실행하지 않음).

경로 약칭: `E` = `internal/app/engine/`, `H` = `internal/strategyhandoff/`, `M` = `openspec/changes/a112-run-four-strategy-families-independently/analysis/measurements/lot-5.6.2-5.2.2/`.

| 항목 | 판정 | 근거 |
|---|---|---|
| **#1 생산 사이클 행동 시험** | **CLOSED** | `E/a112_market_cycle_delivery_test.go:50–54`는 캐시만 주입하고 `runProductionStrategyMarketCycle`을 직접 호출한다. 캐시 반환은 `E/strategy_refresh_wave.go:73–74`, `E/strategy_entry_supervisor.go:563–565`이며 이후 레인 평가와 전달 호출부(`:519–552`)는 그대로 실행된다. 관측 2·주문 1·둘째 범위 identity 오류를 함께 단언(`a112_market_cycle_delivery_test.go:57–66`). H01/H02 검출 기록은 `M/mutation-5.2.2.1-fix3.tsv:3–4`. |
| **#3 기존 네 반례** | **CLOSED** | 직접 별칭 반환은 타입 동일성(`H/mint_census_test.go:84–85`), 반환값 없는 out-param은 `:128–140`, 함수 값 var는 `:190–191`, 직접 init 재대입은 `:197–215`에서 검출한다. H04~H07 기록은 원장 `:7–10`. |
| **#3 같은 부류의 주조** | **OPEN** | 일반 비공개 구조체 반환과 인터페이스 메서드는 `H/mint_census_test.go:100–118`의 재귀 검사로 검출한다. 그러나 제네릭 `T`의 주조와 포인터를 통한 공개 var 교체를 결합하면 공개 표면·census를 피하면서 엔진에서 접근할 수 있다. 아래 발견 #1. |
| **#2 처분의 정직성** | **CLOSED — 동결된 2차 방어라는 범위** | `E/a112_market_delivery_structure_test.go:19–24`는 포인터·필드·맵·송신·미해소 외부 식별자·원소 변경·다른 호출로의 전달을 명시한다. 실제 검사 한계와 맞는다. H03에서 구조 시험도 실패한 것은 포인터 추적의 증명이 아니라 추가된 본문 문장 때문이며, 행동 시험 검출은 별도로 기록돼 있다(`M/…fix3.tsv:5`). 재수리 요구 없음. |
| **#4 `var` 리팩터** | **CLOSED** | `E/a112_market_delivery_structure_test.go:132–142`가 `ValueSpec`을 처리한다. N02 GREEN 기록은 `M/…fix3.tsv:6`. 이번 세션에서 GREEN 재실행은 하지 않았다. |
| **(T) 유일한 방어 문구** | **CLOSED** | 무조건적인 주장은 수정됐다. `E/strategy_dispatch_handoff.go:34–38`은 fixture 순서에만 한정하며, `E/a112_owner_scope_handoff_test.go:189–191`은 생산 재수집의 추가 거절을 명시한다. |
| **새 생산 동작 변경** | **CLOSED — 발견 없음** | 이전 `99ad897c` 사본과 현재 `internal/`·`cmd/`의 모든 비시험 Go 파일을 직접 비교했다. 차이는 `E/strategy_dispatch_handoff.go:34–38`의 주석 한 곳뿐이다. 이 수리에서 새 OFF·주문·안전 루프 동작 변경은 없다. |

#1의 관측값은 엄밀히 **보호 관측 지점까지 도달한 dispatch 횟수**입니다(`E/strategy_dispatch_cycle.go:75–96`, `strategy_dispatch_cycle_test.go:34–37`). 이 fixture에서는 두 범위가 해당 지점에 도달하므로 호출부에서 하나를 누락하면 실패합니다. 다른 헬퍼에서 제거해도 같고, 순서를 뒤집으면 identity 거절이 먼저 발생해 관측 수가 1이 됩니다. 따라서 요청한 호출부 누락 부류는 검출합니다. 모든 입력·모든 변형에 대한 보편 증명으로 확대할 수는 없습니다.

새 발견은 한 건입니다.

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | **P1** | 제네릭 주조와 간접 재대입을 결합하면 기존 공개 오류 값을 통해 새 `Delivered`를 얻을 수 있다. | `H/mint_census_test.go:114–121`은 제약 인터페이스의 메서드만 순회하고 제네릭 인스턴스의 구체 타입을 검사하지 않는다. `:237–252`는 타입이 `T`인 리터럴을 `Delivered`로 세지 않는다. `:197–215`는 대입 좌변에 패키지 var 식별자가 있어야 하므로 매개변수 포인터를 통한 교체를 놓친다. `H/escape_test.go:80–87`은 비공개 수신자 메서드를 제외한다. | 아래 결합 반례를 회귀 변이로 추가. 제네릭 인스턴스에서 경계 값이 생성되는 자리와 공개 var 주소가 전달되어 변경되는 경로를 검증할 것. |

기존 `ErrNoDelivery = errors.New(...)` 선언을 유지하고, 같은 패키지에 다음을 추가하는 반례입니다.

```go
type mintingError struct{ error }

func makeSeam[T ~struct{ result strategyflow.Result }](
    r strategyflow.Result,
) T {
    return T{result: r}
}

func (mintingError) Mint(r strategyflow.Result) any {
    return makeSeam[Delivered](r)
}

func replaceError(dst *error) {
    *dst = mintingError{*dst}
}

func init() {
    replaceError(&ErrNoDelivery)
}
```

엔진에서는 새 공개 이름 없이 접근할 수 있습니다.

```go
strategyhandoff.ErrNoDelivery.(interface {
    Mint(strategyflow.Result) any
}).Mint(r).(strategyhandoff.Delivered)
```

`makeSeam`의 반환형은 `T`, `Mint`의 반환형은 `any`, 변경 대상은 지역 매개변수 `dst`여서 현재 census가 모두 놓칩니다. 패키지 초기화가 기존 공개 오류 값에 구현체를 연결하므로 **도달 불가능한 비공개 헬퍼만 추가하는 반례가 아닙니다.** 기존 두 문을 거치지 않고 값이 채워진 봉투를 생성합니다.

이는 현재 생산에 해당 주조 코드가 있다는 주장이 아닙니다. **수리된 검사가 도달 가능한 둘째 주조 문을 허용한다는 정적 판정**이며, 실제 저장소에 변이를 적용한 통과 여부는 실행 검증하지 않았습니다.

Recommendation: BLOCK 유지 — #3의 제네릭 주조·간접 재대입 결합 반례를 검출한 뒤 협대역 재확인.
