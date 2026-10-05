# Branch Test Map: `prepareRegister`

> **측정 방법**: `go test -covermode=set -coverprofile`. 분기 *조건*이 아니라 **true 결과
> 본문의 실행 여부**를 측정했다. 담당 테스트는 테스트별 개별 프로파일(`-run ^<Test>`)로 특정했다.
> **재측정 (a100 T1, 2026-10-05)**: go1.26.5, 소스 sha256 `de50441b…`(ast.json 과 동일), 작업 base
> `973dd696`. 본문 블록 좌표는 coverprofile 의 `lifecycle.go:<시작행.열>,<끝행.열>` 그대로다.

| Branch | Scenario | Test | true 결과 실행됨 |
|---|---|---|---|
| B1 | position 취득/capability 실패 | `TestSealedStateAndCapabilityTamperFailClosed` | **yes** (L6) |
| B2 | entry latched / phase 부적합 | `TestA100T1PrepareRegisterB2ClosedEntryOrWrongPhaseRefused` | **yes** (본문 블록 8.134–10.3) |
| B3 | 이미 pending인 operation | `TestA100T1PrepareRegisterB3PendingOperationNeverSubmitsTwice` (시나리오 — B2 가 막음), `TestA100T1UnreachablePrepareRegisterB3B4AreShadowedByStateTruthTable` (도달 불가 구조) | **NO — 도달 불가** (본문 블록 11.25–13.3 = 0) |
| B4 | 보호가 이미 active | `TestA100T1PrepareRegisterB4ActiveProtectionNeverSubmitsSecond` (시나리오 — B2 가 막음), `TestA100T1UnreachablePrepareRegisterB3B4AreShadowedByStateTruthTable` (도달 불가 구조) | **NO — 도달 불가** (본문 블록 14.46–16.3 = 0) |
| B5 | 브로커가 정확한 operation 조회를 못 함 | `TestA100T1PrepareRegisterB5MissingExactOperationLookupRefused` | **yes** (본문 블록 17.38–19.3) |
| B6 | 보호 수량이 보유를 정확히 채우지 못함 | `TestRegisterRequiresFullAvailableCoverage` | **yes** (L21) |
| fall-through | 정상 제출 준비 | `TestRegisterResponseCrashRecoversExactlyOnce` 외 | **yes** (L32) |

## T1 결과 — 미실행 4개 중 2개 해소, 2개(B3·B4)는 도달 불가

- **B2·B5** 는 본문 실행 yes. B2 는 시장 latch 단독·포지션 latch 단독(UNPROTECTED + EntryLatch, 재봉인)·
  latch+phase 사례를 잰다. 시장 latch 절이나 포지션 latch 절을 끄면 해당 단독 사례는 **등록이 성공**한다(nil) —
  각 절이 그 사례의 유일한 방어선이다(하네스 `PR_B2`·`PR_B2_no_poslatch` CAUGHT). 거절 시험은 호출 **전** 스냅숏과
  비교해 거절 경로의 제자리 변경도 잡는다(`ALIAS_PR_B2` CAUGHT). B5 는 봉인이 유효한
  능력-부재 capability 로 재며, 같은 `RefusalInvalidObservation` 코드를 내는 B1(봉인 위조
  `capability seal invalid`)과 문구로 갈랐다.
- **B3·B4 는 어떤 유효 상태로도 본문이 실행되지 않는다.** B2 를 통과하려면 phase 가 UNPROTECTED·TERMINAL
  이어야 하는데, `validState` 의 진리표(`validPositionTruth`)가 두 phase 에서 `HasPending=true` 와
  `Observed.Status=ACTIVE` 를 모두 거부한다 — 재봉인해도 B1(`mutablePosition`)에서 `invalid_state` 로 끝난다.
  실제 pending·active 입력은 B2 의 `entry_latched: entry is closed` 로 막히고 **두 번째 명령은 만들어지지
  않는다**(1.2.2·1.2.3 의 안전 요구는 충족·시험 고정). B3·B4 조건을 `false` 로 바꾼 변이는 패키지 전 시험을
  통과한다(SHADOWED). 자기 문구(`operation already pending` · `protection already active`)를 단언하는 RED 증인은
  현 코드에서 FAIL, B2 의 포지션 절을 뺀 가상 수정에서 PASS 다(`analysis/harness/t1_mutation_harness.sh`).
  결함 기록 `analysis/t1-unreachable-branches.md` D2.

**9개 전부 yes 가 아니므로 tasks 1.3.2 의 「2절로 가지 않는다」 조건이 걸린다.** 처분은 Manager·사람 결정이고
core 수정은 M-A 이후다.
