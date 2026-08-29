# Branch Test Map: `httpAPIReader.Snapshot`

- Source: `cmd/tossctl/httpapi_reader.go` (450-520)
- 이 change 가 편집한 분기는 **B8·B9·B10** 이다(Fix 라운드 6.8②, design D4-2).
  기준 판은 분기 9개·return 9개였고(AST 실측: 편집 전 소스로 재추출), 편집이 return
  하나를 없애고 `else` 가지 하나를 만들어 분기 10개·return 8개가 됐다.
- RED/GREEN 열은 **이 change 에서 관측했는가**를 뜻한다. B1~B7 은 이 change 가
  건드리지 않은 기존 fail-closed 행이므로 `no/no` 가 정상이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `r.Engine` 실패 | — (미고정) | no | no |
| B2 | `r.Positions` 실패 | — (미고정) | no | no |
| B3 | `r.Orders` 실패 | — (미고정) | no | no |
| B4 | `r.Candidates` 실패 | — (미고정) | no | no |
| B5 | `r.Performance` 실패 | — (미고정) | no | no |
| B6 | `r.Settings` 실패 | — (미고정) | no | no |
| B7 | `r.Optimization` 실패 | — (미고정) | no | no |
| B8 | **reader 가 있다 / 없다** | `TestAnAbsentStrategyReaderStaysDormantRatherThanUnavailable` `a108_the_aggregate_snapshot_outlives_the_strategy_read_test.go:208` (거짓 가지) · `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` `:160` (참 가지) | no (기존 동작) | yes |
| B9 | **`Read` 실패 → 집계는 살고 전략만 unavailable** | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` `:160` | **yes** | yes |
| B10 | **`Read` 성공 → 읽은 값이 그대로 실린다** | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` 의 대조군 절반(살아 있는 projection 이 `EVIDENCE_STALE` 로 통과) | no | yes |

## B8~B10 이 지는 네 가지

| 성질 | 어디서 재는가 | 죽이는 뮤테이션 |
|---|---|---|
| 전략 읽기 실패가 집계를 죽이지 않는다 | positions/orders/engine 이 살아 있는지 본다 | M10 |
| 읽기 실패는 `RUNTIME_UNAVAILABLE` 이다 | KR·US 두 시장의 refusal code | M11 |
| reader 부재는 `NOT_CONFIGURED` 다 | 두 번째 테스트 전체 | M11 |
| 살아 있는 값은 접히지 않고 통과한다 | 대조군의 `EVIDENCE_STALE` · `strategy.calls == 2` | M12 |

## Fix 라운드(6.8②)에서 RED 을 관측한 순서 — **커밋으로 남았다**

테스트만 담은 커밋 `d8b27021` 이 GREEN 커밋 `aecc03e0` 보다 먼저 있다.

1. `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` 가
   「집계 스냅샷이 실패했다: strategy projection runtime: socket is invalid」로 죽었다.
   **대조군 절반(살아 있는 projection)은 그 전에 통과**했다 — 일곱 조회가 전부 성공하는
   fixture 라는 증거이고, 실패가 여덟 번째 때문이라는 증거다.
2. 같은 커밋에서 `TestAnAbsentStrategyReaderStaysDormantRatherThanUnavailable` 는
   **통과**했다: reader 부재 경로는 이 change 이전에도 dormant 였다. 이 대조군이 없으면
   「항상 dormant 를 준다」는 구현이 1번 테스트를 통과한다.
3. 흡수를 구현하니 둘 다 GREEN.

## 미고정으로 남긴 분기 (사유)

B1~B7 은 이 change 가 편집하지 않은 기존 fail-closed 경로다. 일곱 개 전부 「하위 read
실패 → `return nil, err`」의 같은 모양이고, 고정하려면 일곱 개의 어댑터를 각각
실패시키는 harness 가 필요하다. a108 의 범위는 B8~B10 이므로 **선언된 생략**이다.

다만 이 생략에는 조건이 하나 붙는다: 새 테스트 fixture 는 B1~B7 을 **전부 성공**시켜야
한다. 하나라도 실패하면 「집계가 살아남았다」를 잴 수 없고 테스트가 잘못된 이유로
통과한다. `a108SnapshotReader` 가 일곱 어댑터를 모두 채우는 이유이고,
대조군 읽기(`a108ReadAggregate` 첫 호출)가 그것을 매 실행 확인한다.
