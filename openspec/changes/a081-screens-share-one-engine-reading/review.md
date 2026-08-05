# a081 · Review

> **읽는 순서 주의.** 1·2차는 작성자 자신의 리뷰이고, **3차 독립 리뷰가 그 둘의
> 결론 일부를 뒤집었다.** 아래 두 절은 당시 판단의 기록으로 남기며, 어긋나는
> 문장은 3차가 정본이다. 특히 1차의 "캐시가 실어 나르는 값이 렌더 단위로 안
> 움직인다"와 "오늘보다 느려지는 화면이 없다", 2차의 "변이 6.2가 RED"는 전부
> 3차에서 반증되거나 좁혀졌다.

## 1차 — proposal freeze (2026-08-05)

### 이 change가 성립하는 근거를 다시 확인했다

| 주장 | 확인 |
|---|---|
| 표시 경로가 렌더마다 엔진에 2회 닿는다 | `portfolio_pages.go` B1 안의 `Runtime`·`List`. 캐시 없음 |
| `Runtime`도 로컬 호출이 아니다 | `cmd/tossctl/position_policy_commander.go:35-41` — descriptor 재해석 + 매 호출 `DialRuntime` |
| `List`가 엔진의 쓰기 커넥션을 탄다 | `position_policy_command.go:138-142` → `s.j.PositionPolicies` → `journal.go:151` `SetMaxOpenConns(1)` |
| 뺏은 시간이 판정 간격에 더해진다 | `exitloop.go` `Run`은 ticker가 아니라 `작업 후 sleep` |
| ~~캐시가 실어 나르는 값이 렌더 단위로 안 움직인다~~ | **3차에서 반증.** adoption effective는 생성자 고정이 맞지만 blocks는 **살아 있는** tracker 상태다 |
| 보호선의 값은 이 경로에서 오지 않는다 | 워터마크·기준선·저장 exit 근거는 `c.positions()`의 journal 읽기 |

### 스펙 충돌 검토

셋을 대조했다.

- **`rate budget 보호`** — 브로커 예산에 대한 것이다. a081은 브로커에 닿지 않고
  `holdingsTTL`을 건드리지 않는다. 충돌 없음.
- **"편입 보조 상태는 candidate와 reconcile 차단을 함께 설명한다"** — "runtime
  unavailable인 non-managed 행은 desired를 effective로 위장하지 않고 `UNKNOWN`"이
  SHALL이다. **이것이 캐시 설계를 결정했다** — 마지막 성공을 되살리는 캐시였다면
  이 SHALL을 시간 축에서 깨뜨린다. 마지막 결과를 서빙하는 형태만 통과한다
  (design D3).
- **"포지션 관리는 실제 adoption desired/effective를 구분한다"** — `/position-management`
  전용 SHALL이다. 그 화면은 직접 읽기를 유지하므로 판정 근거가 바뀌지 않는다
  (design D6).

그래서 delta는 **ADDED 1건**이다. a080과 달리 MODIFIED가 필요 없다.

### 이 change가 스스로에게 물은 것

**"오늘보다 느려지는 화면이 있는가."** ~~없다.~~ **3차에서 좁혀졌다** — 자동
재로드에 대해서는 참이지만 **수동 새로고침은 오늘 언제나 새 읽기를 가져온다.**
그 창은 D2의 간격 분리로 줄었고, ADDED 요구사항이 나이의 상한을 명시한다.

**"a080의 목적을 해치는가."** 아니다. a080이 빨리 보이게 하려는 값(워터마크,
기준선, rung 발의)은 journal 읽기에서 오고 이 change는 그 경로를 손대지 않는다.

**"엔진을 건드리는가."** 아니다. 엔진 파일을 한 줄도 열지 않았다.

## 2차 — 구현 후 자기 검토 (2026-08-05)

### 관측한 것

| 항목 | 결과 |
|---|---|
| RED (캐시 없는 HEAD) | 렌더 6회 = 엔진 읽기 6벌, 두 화면 3렌더 = 3벌, 실패 후 4렌더 = 4회 재연결 |
| GREEN | `internal/console` 703 passed, 0 failed, 1 skipped |
| 기존 테스트 수정 | **0건.** 5.4 충족 |
| 변이 6.1 (간격 gate 제거) | RED — 6회/3회/6회로 되돌아감 |
| 변이 6.2 (실패 시 직전 성공 부활) | RED — **단, `Runtime` 절반만.** 3차가 lifecycle 절반의 같은 변이를 703건 통과시켰다 |
| 변이 6.3 (무효화 제거) | RED — Apply·격리 해제 뒤 `runtime 4→4, list 4→4` |
| `make vet` | clean |
| `make validate` | exit 0 (`✓ change/a081-screens-share-one-engine-reading`) |
| `make sdd-check` | exit 0 — CodeGraph hard evidence fresh. codegraphcontext·GBrain은 advisory WARN |
| logic-map | 4 target 전부 통과 |

### 구현 중 바뀐 판단 1건 — B1 조건은 건드리지 않는다

처음에는 B1을 `c.enginePolicy != nil`로 바꿨다. `a053_exit_line_reference_test.go`의
corrupt 케이스가 깨졌다 — 그 테스트는 하네스 구성 **후** `h.opts.PositionPolicies = nil`로
seam을 떼서 "lifecycle 미배선" 상태를 만든다. 캐시는 생성자에서 이미 만들어졌으므로
배선된 시절의 읽기를 계속 보게 된다.

테스트를 고치는 대신 코드를 되돌렸다. B1이 `c.opts.PositionPolicies`에 대고
"배선되었는가"를 계속 묻는 것이 Function Logic Map이 안전 경계로 적어 둔 것과
같고("B1은 커맨더가 배선되었는가를 계속 묻는다"), 런타임에 seam을 떼는 호출자가
캐시된 과거를 보지 않게 하는 것이 옳은 방향이다. 결과적으로 기존 테스트 수정
0건으로 끝났다.

### 남은 위험

- **간격 안의 staleness.** 엔진이 죽은 직후 최대 한 간격(30초) 동안 화면이 마지막
  성공을 보여준다. 오늘의 30초 재로드가 이미 갖는 창과 같은 크기이고, 그 뒤에는
  반드시 unknown으로 수렴한다(6.2가 고정). 재로드 주기를 5초로 내리는 a080이
  land하면 이 창은 5초에서 30초로 **넓어진다** — 다만 넓어지는 대상은 adoption
  effective·lifecycle이고, 보호선의 값과 엔진 생존 신호(`protectionLiveness`)는
  아니다. a080 재개 시 이 문장을 다시 읽을 것.
- **엔진 쪽 변화의 지연.** 콘솔을 거치지 않은 adoption·대사 변화는 최대 한 간격
  늦게 보인다. 콘솔이 관측할 수 없는 사건이고 오늘의 재로드 지연과 같은 급이다.
- **미실측.** 실제 컨테이너에서 엔진 사이클 시간이 나아지는지는 task 6.6이고
  사람 승인이 필요하다. 코드상 근거는 있지만 숫자는 아직 없다.

### 미완

- **task 6.6** — 컨테이너 실측. 운영 환경 조작이므로 사람이 직접 승인한다.
- **task 7.4** — 별도 컨텍스트의 독립 리뷰. a080의 F1이 정확히 이 단계에서
  나왔으므로 생략하지 않는다.

## 3차 — 별도 컨텍스트 독립 리뷰 3인 (2026-08-05, task 7.4)

WORKFLOW.md §역할 분리에 따라 작성자와 분리된 컨텍스트 셋에게 렌즈를 나눠 맡겼다.
**안전**(캐시가 운영자에게 보여주는 것이 틀릴 수 있는가), **테스트**(각 테스트가
이름값을 하는가 — 구현을 변이시켜 확인), **스펙·SDD**(기존 SHALL 충돌과 산출물
정합성).

P0는 없었다. **P1 7건이 나왔고 그중 셋이 설계를 바꿨다.**

### 설계를 바꾼 것

| 발견 | 어디서 | 조치 |
|---|---|---|
| **취소된 요청이 공유 reading을 오염시킨다** — 브라우저가 렌더를 버리면 그 실패가 캐시되고, 건강한 엔진인데도 두 화면의 보호선이 한 간격 동안 사라진다. 재로드로 회복 불가 | 안전·테스트 **둘이 각각 재현** | `context.WithoutCancel` + 자체 타임아웃. design **D4b 신설**, spec에 SHALL NOT 추가, `TestAnAbandonedRenderDoesNotPoisonTheSharedReading` |
| **"이 값들은 렌더 단위로 안 움직인다"가 거짓** — 대사 차단은 살아 있는 tracker 상태다. 30초를 붙들면 차단된 보유를 "편입 예약됨"으로 표시한다(낙관적으로 틀림). 그리고 **`Runtime`은 DB를 아예 안 건드린다** — 안전 논거는 `List`에만 적용된다 | 안전 P1-2 | 간격 분리: `List` 30초 / `Runtime` 5초. design **D2 전면 개정**, `TestAReconcileBlockIsNotHeldForTheLifecycleInterval` |
| **"한 시점의 짝" 전제가 애초에 성립하지 않았다** — `c.positions()`가 렌더마다 journal을 읽으므로 시점이 셋이다 | 안전 P1-2(b) | D4 **철회**. 보존하는 것을 동시성에서 **불일치의 방향**(fail-closed)으로 옮기고, 그 비용을 명시. `TestAStaleLifecycleListCanOnlyWithholdAVerdictNeverInventOne` |

### 테스트를 바꾼 것

**가장 아픈 것**: `TestAFailedReadingIsNotMaskedByThePreviousSuccess`가 **절반만
증명하고 있었다.** 리뷰어가 stale lifecycle 목록을 되살리는 변이를 넣자 **703건
전부 통과**했다. 원인은 픽스처 — `managedPolicyState()`의 `PositionID`가 `p-1`인데
journal 픽스처는 `pos-managed`라서 `policyByID` 조회가 한 번도 맞지 않았고,
lifecycle 절반이 모든 화면 테스트에서 죽어 있었다.

그리고 그쪽이 **더 위험한 절반**이다. `Managed()`는 캐시된 lifecycle만 읽고
`ProjectManagement`는 `Managed`를 `EffectiveKnown`보다 먼저 보므로, stale 목록이
되살아나면 행이 죽은 엔진의 증거로 `엔진 관리`를 주장한다.

a080의 F1과 **같은 결함 종류**가 그 F1을 고치는 change 안에서 재발한 것이다.

| 발견 | 조치 |
|---|---|
| lifecycle 절반이 미검증 | `PositionID`를 `pos-managed`로 정합. `TestAFailedLifecycleReadingIsNotMaskedByThePreviousSuccess` 신설 — 변이 M3가 이제 RED |
| 동시성 SHALL에 테스트가 없다 (goroutine 0개) | `TestConcurrentRendersCostOneReading`. 읽기 창을 넓히는 hook을 fake에 추가해 "락을 놓고 늦게 stamp" 변이가 RED가 되게 함 |
| `if redraws <= 1 { t.Fatal }`이 **죽은 코드** (`const`라 컴파일 타임 결정) | 간격에서 유도해 계산 — 간격을 바꾸면 자기 검사도 따라 움직인다 |
| 거부 테스트가 부정 단정뿐 (404여도 통과) | 거부 문구를 **긍정으로** 단정 |
| 두 화면 공유 테스트가 "/dashboard가 아예 안 읽음"과 구분 못 함 | 두 body 모두에 판정 렌더를 단정 |
| `Runtime.Blocks` 슬라이스 aliasing 미복사 | `snapshotLocked`에서 복사 + 테스트 |
| 테스트 파일 주석이 사실과 다름("모든 기존 테스트가 미배선") | 정정 — a052·a053·a077·a079는 배선한다 |

### 스펙을 바꾼 것

| 발견 | 조치 |
|---|---|
| 간격 staleness가 runtime-unavailable SHALL을 **시간 축에서** 약화하는데 상한이 없다 | ADDED에 "캐시된 읽기의 나이는 그 간격을 넘지 않는다" SHALL 추가. D2의 간격 분리가 그 창을 실제로 줄인다 |
| 면제 기준 "명령을 발행하는 화면"이 코드와 첫날부터 불일치 — `positionPolicySummary`(설정 탭)는 명령을 발행하지 않으면서 렌더마다 `List`를 부른다 | 기준을 **"자동 재로드가 없는 화면"** 으로 교체. D6 개정, issues I2에 기재 |
| 절대 상한 SHALL과 무효화 SHALL이 서로 양보하지 않음 | 무효화를 상한의 명시적 예외로 문구화 |
| 인용 줄번호 4건 오류 (a080에서 지적한 rot이 여기서 재발) | 전부 정정 |
| 엔진 시작·정지가 무효화 집합에 없다 | **무효화를 넓히지 않고** `Runtime` 간격을 5초로 줄여 해소. 같은 렌더에서 `protectionLiveness`가 엔진 정지 사실을 지연 없이 표시한다. 판단 근거를 D5에 기록 |
| 격리 해제 무효화가 **존재하지 않는 data flow를 주장** — 해제는 `exit_snapshot_quarantines`만 쓰고 lifecycle SELECT는 그것을 join하지 않는다 | 호출과 그 테스트를 **제거**하고 이유를 코드 주석으로 남김 |

### 확인된 것 (반증 없음)

- 핵심 전제 전부: 단일 쓰기 커넥션, `List`가 그 위의 SELECT, exit 루프의
  `작업 후 sleep` 구조, `Runtime`이 매 호출 dial. 인용 코드는 그 4건 외 전부 verbatim.
- Logic Map 4건의 AST 해시·분기 번호·safe edit boundary 전부 일치. task 참조 rot 없음.
- 다른 스펙 4종(`position-exit-policy-management`·`exit-policy`·`engine-safety`·
  `console-request-origin`) 충돌 없음. 새 라우트·POST·origin 표면 0건.
- 브로커·주문·journal 쓰기 도달 경로 **없음**. `positionPolicyReader`가 두 read만
  명명하므로 `Preview`/`Apply`는 타입상 도달 불가.
- 데드락 없음, `-race` clean. 무효화 커버리지는 모든 mutating route 열거로 확인.
- STORY 승인 기준 ↔ 구현 일치 (a080의 실패 모드 미재발).
- 기존 테스트 수정 0건.

### 프로세스에서 배운 것

리뷰어 셋에게 **같은 워킹 트리**를 줬다. 한 명이 변이 검증으로 소스를 바꾸는
동안 다른 한 명이 테스트를 돌려 6건 실패를 봤고, 그것을 "환경 위험"으로 보고했다.
결과에는 영향이 없었지만(격리 사본에서 재확인) 다음부터는 리뷰어마다 worktree를
분리해야 한다. `issues.md` I6에 남긴다.

### 결론

**설계 3건·테스트 7건·스펙 7건을 고쳤다.** 변이 검증을 7종으로 늘렸고 전부 RED를
관측했다. 리뷰가 없었다면 이 change는 자기가 고치겠다던 결함 종류를 그대로 안고
land할 뻔했다.
