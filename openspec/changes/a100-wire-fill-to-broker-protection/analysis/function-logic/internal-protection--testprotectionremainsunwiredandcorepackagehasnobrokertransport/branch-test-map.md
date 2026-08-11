# Branch Test Map: `TestProtectionRemainsUnwiredAndCorePackageHasNoBrokerTransport`

> **측정 방법 — 커버리지는 `not-applicable`.** 이 함수는 테스트 함수이고, Go 커버리지는
> `_test.go`를 계측하지 않는다(`internal/protection` 프로파일에 이 파일의 블록이 0개임을
> 확인했다). 대신 측정한 것은 **이 테스트가 현재 통과한다**는 사실이다 —
> `go test -covermode=set ./internal/protection` exit 0, 패키지 70.8%.
> 측정일 2026-08-11, source SHA-256 `0f715a4e134e7477…`(`ast.json`).
>
> 여기서 「실행됨」은 커버리지가 아니라 **이 가드가 지금 무엇을 실제로 거부하고 있는가**로
> 읽는다. 검사 대상 파일은 현재 worktree에 실재하므로 순회 분기는 모두 돌고, 위반 분기
> (`t.Errorf`/`t.Fatal`)는 위반이 없으므로 전부 **미도달**이다 — 그것이 통과의 의미다.

| Branch | Scenario | 현재 상태 | a100의 영향 |
|---|---|---|---|
| B1 | 프로필이 UNWIRED가 아니다 | 미도달(UNWIRED다) | **없음** — a100은 `Wired`를 생산하지 않는다 |
| B2 | 패키지 디렉터리 읽기 실패 | 미도달 | 없음 |
| B3 | core 파일 순회 | 돈다 | 없음 |
| B4 | 디렉터리·비-go·테스트 파일 skip | 돈다 | 없음 |
| B5 | core 파일 파싱 실패 | 미도달 | 없음 |
| B6 | core import 순회 | 돈다 | 없음 |
| B7 | **core가 transport를 import한다** | 미도달 | **없음** — a100은 core에 transport를 넣지 않는다 |
| B8 | `cmd/`·`internal/app/` 순회 | 돈다 | 없음 |
| B9 | walk 에러 | 미도달 | 없음 |
| B10 | skip 판정 | 돈다 | 없음 |
| B11 | 파싱 실패 | 미도달 | 없음 |
| B12 | import 순회 | 돈다 | 없음 |
| B13 | `internal/protection`를 import한다 | **도달한다** (`gateway.go`) | 워커 파일이 import하면 여기도 도달 |
| B14 | **허용 목록에 없다 → 두 번째 조립 경로** | 미도달 | **워커가 `internal/protection`을 import하면 도달 = 실패** |
| B15 | walk 반환 에러 | 미도달 | 없음 |
| B16 | `gateway.go` 읽기 실패 | 미도달 | 없음 |
| B17 | 필수 문자열 순회 | 돈다 | 없음 |
| B18 | **필수 문자열 부재** | 미도달 | **없음이어야 한다** — a100은 readiness 조립 2개를 유지한다 |
| B19 | 금지 문자열 순회 | 돈다 | 없음 |
| B20 | **금지 문자열 존재** | 미도달 | **`protectionofficial.New`를 넣으면 도달 = 실패** |

## a100이 실제로 마주치는 분기는 둘뿐이다

**B20** — `protectionofficial.New`를 프로덕션 조립에 넣는 순간 실패한다. 이것이 D6이 말하는
「봉인을 연다」의 전부다. 금지 4개 중 **1개만** 목록에서 뺀다.

**B14** — 워커 파일이 `internal/protection`을 import하면 실패한다. **여는 것이 아니라 피한다**:
어댑터를 `gateway.go` 안에서 만들고 워커에는 좁은 인터페이스만 넘기면 허용 목록은 그대로다.

나머지 18개 분기는 a100과 무관하며, **a100 이후에도 지금과 똑같이 거부해야 한다.**

## a100의 RED 대상

- **금지 문자열 분기(L84) — RED 필수.** 목록에서 한 줄을 빼기 전에, 남은 3개
  (`protection.NewSupervisor`, `protection.db`, `GatewayFactory`)가 여전히 거부되는지 고정한다.
- **허용 목록 분기(L62) — RED 필수.** 워커 파일이 `internal/protection`을 import하면 실패하는지
  — 즉 허용 목록이 넓어지지 않았음을 테스트가 증명한다.
- **프로필 분기(L15) — 회귀 고정.** a100 이후에도 `ProfileProtection == UNWIRED`인지.
  a071의 보장이 유지된다는 유일한 기계적 증거다.
- **필수 문자열 분기(L79) — 회귀 고정.** readiness 조립 2개가 남아 있는지.
