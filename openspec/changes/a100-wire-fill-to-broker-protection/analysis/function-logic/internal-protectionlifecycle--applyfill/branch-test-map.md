# Branch Test Map: `applyFill`

> **측정 방법**: `go test -covermode=set -coverprofile`. 분기 *조건*이 아니라 **true 결과
> 본문의 실행 여부**를 측정했다. 담당 테스트는 테스트별 개별 프로파일(`-run ^<Test>`)로 특정했다.
> **재측정 (a100 T1, 2026-10-05)**: go1.26.5, 소스 sha256 `de50441b…`(ast.json 과 동일), 작업 base
> `973dd696`. 본문 블록 좌표는 coverprofile 의 `lifecycle.go:<시작행.열>,<끝행.열>` 그대로다.

| Branch | Scenario | Test | true 결과 실행됨 |
|---|---|---|---|
| B1 | position 취득 실패(비-InvalidObservation) | `TestA100T1ApplyFillB1UnknownPositionRefusedWithoutCreatingState` | **yes** (본문 블록 237.63–239.3) |
| B2 | state seal 무효 | `TestA100T1ApplyFillB2BrokenSealRefusedWithoutReseal` (시나리오), `TestA100T1UnreachableApplyFillB2IsShadowedByB1` (도달 불가 구조) | **NO — 도달 불가** (본문 블록 240.24–242.3 = 0). 무효 봉인은 B1 에서 같은 반환값으로 끝난다 |
| B3 | fill 식별자 무효 / broker order id 불일치 | `TestA100T1ApplyFillB3ForeignOrInvalidFillIdentityRefused` | **yes** (본문 블록 244.154–246.3) |
| B4 | 같은 FillID, 다른 내용 → latch + 거부 | `TestDuplicateAndPartialFillConvergeOnce` | **yes** (L252) |
| B5 | 같은 FillID, 같은 내용 → 멱등 Duplicate | `TestDuplicateAndPartialFillConvergeOnce` | **yes** (L249) |
| B6 | 수량 0 또는 claim 초과 | `TestA100T1ApplyFillB6ZeroOrExcessQuantityRefused`, `TestA100T1ApplyFillB7FullFillClosesTerminalThenRefusesFurtherFills` (종료 후) | **yes** (본문 블록 254.107–256.3) |
| B7 | 잔량 0 → `Terminal` 전이 | `TestA100T1ApplyFillB7FullFillClosesTerminalThenRefusesFurtherFills` | **yes** (본문 블록 264.37–266.3) |
| fall-through | 정상 부분체결 반영 | `TestDuplicateAndPartialFillConvergeOnce` | **yes** (L270) |

## T1 결과 — 미실행 5개 중 4개 해소, 1개(B2)는 도달 불가

- **B1·B3·B6·B7** 은 본문 실행 yes. 각 시험은 거절 **문구**까지 단언해 실패 지점을 가른다.
  가드 사슬 변이(작업 트리 무수정, `go test -overlay`)로 각 가드를 끄면 해당 시험이 **다음 벽의 문구**로
  실패함을 확인했다 — B1 을 끄면 B3 의 `fill identity mismatch`, B3 을 끄면 정상 반영(nil) 또는 B6,
  B6 을 끄면 정상 반영(nil), B7 을 끄면 잔량 0 인데 `Phase=ACTIVE, EntryOpen=true`(보호 0 인 채 진입 재개방).
- **B2 는 어떤 입력으로도 본문이 실행되지 않는다.** B1 의 `mutablePosition` 이 첫 줄에서 같은
  `validState` 를 부르고, applyFill 이 넘기는 zero capability(봉인 0)는 봉인 검사를 건너뛰므로
  `mutablePosition` 이 `RefusalInvalidObservation` 을 낼 길이 없다 — 즉 무효 봉인은 항상 B1 에서
  `invalid_state: state seal invalid` 로 끝나고, B2 본문은 같은 반환값을 낼 뿐이다. B2 조건을 `false` 로
  바꾼 변이는 패키지 전 시험을 통과한다(SHADOWED). 1.1.2 의 **안전 요구**(거부 + 재봉인·복구 없음)는
  B1 경로에서 충족되고 시험으로 고정됐다. 본문 yes 는 core 수정 없이 불가 — 결함 기록
  `analysis/t1-unreachable-branches.md` D1.

**9개 전부 yes 가 아니므로 tasks 1.3.2 의 「2절로 가지 않는다」 조건이 걸린다.** 처분(도달 불가 분기의
제거·재배치 또는 `unsupported` 판정)은 Manager·사람 결정이고 core 수정은 M-A 이후다.
