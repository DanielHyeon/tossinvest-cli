# Branch Test Map: `Runner.RunCycle`

> **측정 방법**: `go test -covermode=set -coverprofile ./internal/soak`.
> 분기의 *조건*이 아니라 **분기 본문 블록의 실행 여부**를 측정했다(같은 줄에서 시작하는
> 커버리지 블록 중 시작 column이 가장 큰 것이 본문이다).
> 측정일 2026-08-11. **편집 전과 편집 후를 각각 측정했다.**
>
> | | 패키지 | `RunCycle` | 분기 수 | source SHA-256 |
> |---|---|---|---|---|
> | 편집 전 | 84.9% | 95.5% | 3 | `ae912bdcdc8829c7…` |
> | 편집 후 | 85.9% | 96.2% | **3 (그대로)** | `84fa92e8716d3a6b…` |

| Branch | Scenario | Test | 본문 실행됨 | 비고 |
|---|---|---|---|---|
| B1 | `:506` 취소된 ctx → 빈 사이클 | — | **no** (506.34 count=0) | 기록이 만들어지지 않는 유일한 경로 |
| B2 | `:522` account 목록 순회 | `internal/soak` 패키지 | **yes** (522.29 count=1) | 정상 경로 |
| B3 | `:523` 첫 비어있지 않은 account 채택 | 동상 | **yes** (523.33 count=1) | `AccountRef` 결정 |

**측정 결과: 편집 전후 모두 3개 중 1개 미실행(B1). 편집이 분기를 만들지 않았다.**

## 이 측정이 0.10 (a)에 갖는 뜻

**분기는 편집 대상이 아니었다.** 셋 중 둘은 계정 해석이고 하나는 취소 확인이며, 추가된
probe는 어느 것도 통과하지 않는다. 실제 편집은 `cycle.Endpoints`에 append하는 **직선 코드**를
`probePrices` 뒤(`:559-562`)에 놓은 것이다. 그래서 회귀가 날 수 있는 곳은 분기가 아니라
**순서와 부작용**이었고, RED는 거기에 붙였다.

## RED → GREEN (`internal/soak/protection_probe_test.go`)

| 고정한 불변식 | 테스트 | 깨지면 무엇이 망가지나 | 상태 |
|---|---|---|---|
| credential 격리 | `TestConditionalReadFailureLeavesTheCredentialStreakIntact` | 조건주문 endpoint 하나가 3일 streak를 무너뜨리고 **a100과 무관한 2026-08-29 재발급까지 막는다** | **GREEN** |
| completeness 격리 | `TestConditionalReadFailureLeavesCompletenessIntact` | 창 안의 completeness 실패가 attestation 전체를 거부시킨다(`attest.go:137-144`) | **GREEN** |
| 순서 | `TestConditionalReadsAreRecordedAfterTheQuoteRead` | 추가 요청이 연 429가 attestation이 이미 의존하는 endpoint를 잡는다(M8) | **GREEN** |
| by-id의 skip | `TestConditionalOrderByIDIsSkippedWhenTheAccountHasNone` | 없는 것을 실패로 적으면 기록이 없는 계좌에서 영원히 실패로 남는다 | **GREEN** |
| by-id의 성공 | `TestConditionalReadsAreProvenWhenTheAccountHasOne` | 목록이 준 식별자를 쓰지 않으면 by-id는 증명되지 않는다 | **GREEN** |
| 전부 GET | `TestConditionalEndpointsAreReads` | non-GET 하나가 `BuildAttestation`을 통째로 거부시킨다(`attest.go:188-194`) | **GREEN** |
| **거부 목록 불변** | `TestConditionalEndpointsStayOutOfTheRefusalCatalogs` | `RequiredEndpoints`/`LiveOnlyEndpoints`를 지금 넓히면 **거부가 증거보다 먼저 온다** | **GREEN** |
| 요청 수 계상 | `TestTheConditionalListRecordsBothGroupRequests` | 두 그룹 호출을 1건으로 적으면 **자기 burst를 절반으로 과소 계상**한다(M8이 쓰라린 전례) | **GREEN** |
| 부분 목록 거부 | `TestAFailedConditionalGroupIsNotAPartialList` | 한 그룹이 실패했는데 다른 그룹의 식별자를 살리면 by-id가 성립하지 않은 read를 증명했다고 주장한다 | **GREEN** |

새 probe 3개는 전부 **100.0%** 측정됐다
(`probeConditionalOrders`, `probeConditionalByID`, `probeSellableQuantity`).

## 어댑터는 stub이 아니라 HTTP로 증명했다

위 테스트는 전부 `soak.Reads` stub을 통과하므로 **답을 어떻게 얻는지에 대해서는 아무것도
증명하지 않는다.** 경로·질의·식별자 전달은 `cmd/tossctl`의 어댑터에만 존재하므로 실제 요청으로
확인했다(`cmd/tossctl/soak_protection_test.go`).

| 확인한 것 | 테스트 |
|---|---|
| 세 endpoint 모두 OK로 기록된다 | `TestSoakRunSurveysTheProtectionReads` |
| 목록을 OPEN·CLOSED 두 그룹으로 읽고, sellable은 조사 종목으로 묻는다 | `TestSoakWalksBothConditionalGroups` |
| by-id가 **목록이 준 식별자**를 쓴다 | `TestSoakReadsTheConditionalTheListReturned` |
| 조건주문 경로가 열려도 GET 외 요청이 없다 | `TestSoakStillIssuesNoMutatingRequestWithTheProtectionReads` |

마지막 것이 특히 필요하다 — 조건주문의 read 경로는 create·modify·cancel과 **같은 경로를
공유**하므로, 그 경로를 답하는 서버를 상대로 read-only를 다시 세워야 한다.

## attestation 적재가 required 목록과 무관하다는 것

`TestAProbedEndpointReachesTheAttestationWithoutBeingRequired`이 이 change 전체가 서 있는
성질을 고정한다: `BuildAttestation`은 `Window.SuccessfulEndpoints()`를 싣고
(`attest.go:183`, `summary.go`), `RequiredEndpoints()`는 `Evaluate`의 거부 기준일 뿐이다
(`attest.go:130`). 이것이 **증거를 먼저 만들고 요구를 나중에 거는** 순서를 가능하게 한다.

- **B1 — a100 범위 밖.** 취소 경로의 도달 조건은 달라지지 않는다.

## 측정으로 답하지 못한 것

- **미보유 종목에 대한 `GET /api/v1/sellable-quantity`의 실제 응답.** soak의 `--symbols`
  기본값은 `005930`이고 이 계좌의 보유 목록 밖이다. 단위 테스트는 stub이 답하므로 이 질문에
  답하지 않는다. **배포 후 첫 사이클의 기록이 답한다**(tasks 0.10c). 실패해도 막히는 것은
  없다 — 세 endpoint 중 어느 것도 거부 목록에 없다.
