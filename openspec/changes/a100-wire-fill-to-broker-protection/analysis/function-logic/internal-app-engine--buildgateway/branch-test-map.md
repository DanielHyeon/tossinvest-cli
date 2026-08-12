# Branch Test Map: `buildGateway`

> **측정 방법**: `go test -covermode=set -coverprofile ./internal/app/engine`(패키지 62.1%)와
> `go test -covermode=set -coverpkg=./internal/app/engine ./cmd/tossctl`(engine 대상 8.9%)의
> **합산**. `buildGateway`는 `cmd/tossctl`의 조립 테스트에서도 불리므로 한쪽만으로는 답이
> 나오지 않는다. 분기의 *조건*이 아니라 **true 결과 본문 행의 실행 여부**를 측정했다.
> 측정일 2026-08-11, source SHA-256 `de44aae53a194183…`(`ast.json`).

| Branch | Scenario | Test | true 결과 실행됨 | 비고 |
|---|---|---|---|---|
| B1 | `checkProjectionWired` 실패 → 조립 중단 | — | **no** (L205) | apply hook 미바인딩 |
| B2 | `tracker.Restore` 실패 → 조립 중단 | — | **no** (L227) | projection 복원 실패 |
| B3 | `NewPairedReadinessAdapter` 실패 → 조립 중단 | — | **no** (L251) | readiness 구성 실패 |
| B4 | `execgw.New` 실패 → 조립 중단 | — | **no** (L275) | 게이트웨이 구성 실패 |

**측정 결과: 분기 4개 전부 미실행. 정상 경로만 실행된다** — 함수 블록 11개 중 7개 실행.

## 이 측정이 a100에 갖는 뜻

**엔진 조립의 실패 경로는 한 번도 실행된 적이 없다.** 성공 경로만 검증되어 있다는 뜻이며,
a100이 여기에 새 구성요소를 추가하면 **새 실패 원인이 검증되지 않은 실패 처리 위에 얹힌다.**

그래서 FLM의 결론(생성만·기동은 호출자·에러는 기존 4형태 중 하나)이 취향이 아니라 측정된
제약이다. 새 `return`을 추가하면 **미실행 분기가 5개가 된다.**

## a100의 RED 대상

- **B1~B4는 a100의 RED 대상이 아니다.** a100은 이 네 분기를 새로 도달 가능하게 만들지 않고,
  기존 조립 실패 처리의 커버리지 개선은 이 change의 범위 밖이다.
- **a100이 만드는 것**: 워커 생성이 실패할 수 있다면 그 실패는 B3·B4와 **같은 형태**여야 하고,
  그 분기에 대한 테스트를 a100이 함께 만든다. 형태가 같으면 새 테스트도 기존 네 개와 같은
  방식으로 쓸 수 있다.
- 워커 생성이 실패할 수 없게 만들면(순수 구조체 조립) **분기가 늘지 않는다.** 그쪽이 낫다.
