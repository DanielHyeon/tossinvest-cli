# a081 · Issues

작업 중 확인한 사실 중 이 change가 고치지 않는 것, 또는 다른 곳에 남는 것.

## I1 — `Runtime`은 매 호출 descriptor를 다시 읽고 소켓을 새로 연다

```go
// cmd/tossctl/position_policy_commander.go:35-41
func (r positionPolicyRuntimeDescriptorReader) Runtime(ctx context.Context) (...) {
	client, err := positionpolicyrpc.DialRuntime(ctx, r.descriptorPath)
	...
}
```

주석이 그 이유를 밝힌다 — 사이드카가 엔진보다 먼저 뜰 수 있고 엔진 재시작이
소켓과 bearer token을 바꾸므로, 기동 시점 client를 붙들면 실효 정책이 영영
unknown이 된다. 타당하다. 다만 그 결과 이 읽기는 "RPC 1회"가 아니라 "파일 읽기 +
dial + RPC"이고, a081 이전에는 **렌더마다** 그것을 했다.

a081은 그 횟수를 간격당 1회로 묶는다. dial 자체를 재사용하지는 않는다 — 위
주석의 이유가 그대로 유효하고, 간격당 1회면 비용이 문제되지 않는다.

## I2 — 같은 형태의 결합이 남아 있는 곳

`decoratePositionRows`만 고쳤다. 확인한 나머지는 이렇다.

| 경로 | 렌더마다 하는 일 | 판단 |
|---|---|---|
| `c.opts.Settings.Load()` (`portfolio_pages.go`) | 콘솔 프로세스 안의 로컬 config 파일 읽기·파싱 | **엔진에 닿지 않는다.** 단일 쓰기 커넥션과 무관하므로 a081의 근거가 적용되지 않는다. 재로드 주기를 올릴 때 다시 볼 값이다 |
| `positionPolicySummary` (`settings_tabs.go`) | 렌더마다 `List` — **엔진에 닿는다** | 설정 탭은 자동 재로드가 없어 사람이 여는 만큼만 든다. ADDED 요구사항의 면제 기준을 "명령을 발행하는가"에서 "자동 재로드가 없는가"로 바꾼 이유가 이 호출이다 (design D6) |
| `c.opts.StrategyRuntime` (`console.go:247`) | 다른 화면의 경로 | 이 change의 두 화면이 부르지 않는다. 같은 형태인지 별도 확인 필요 |
| `/position-management`·설정 탭의 `List`/`Runtime` | 엔진 읽기 | **의도적으로 직접 읽기 유지** (design D6). 자동 재로드가 없어 사람이 여는 만큼만 든다 |
| `c.positions()`의 journal 읽기 | 콘솔의 read-only 핸들 | WAL이라 엔진 쓰기를 막지 않는다. 보호선의 신선도가 여기서 나오므로 캐시하면 a080의 목적이 사라진다 |

`StrategyRuntime`은 a081의 범위가 아니지만, 재로드 주기를 만지기 전에 같은
질문을 받아야 한다.

## I3 — `make lint`는 이 change 이전부터 red다

```text
gofmt: 아래 파일이 포맷되지 않았습니다 — `make fmt` 를 실행하세요:
  internal/httpapi/performance_attribution_test.go
```

base commit `840b3377`에서 이미 그렇다. 작업 중 `make fmt`가 이 파일을 고쳤고,
a081과 무관한 파일이므로 되돌렸다. 되돌린 이유는 둘이다 — 이 change의 diff에
관계없는 패키지를 넣지 않기 위해서, 그리고 그 한 줄이 logic-map gate에서 "수정된
기존 함수"로 잡혀 공백 수정에 Function Logic Map을 요구하기 때문이다.

`make gate`는 `sdd-check`·`test`·`vet`·`validate`만 돌리므로 gate는 영향받지
않는다. 별도 정리 대상이다.

## I4 — 캐시 실패가 `unknown`으로 수렴하는 것은 zero value에 얹혀 있다

`decoratePositionRows`는 `Runtime`의 에러를 검사하지 않는다.

```go
runtime, _ = c.opts.PositionPolicies.Runtime(ctx)   // 편집 전
```

실패 시 `EffectiveKnown = false`가 되는 것은 에러 처리 때문이 아니라 **반환된
zero value 때문**이다. `operator-console`의 "runtime unavailable인 행은 desired를
effective로 위장하지 않는다"는 SHALL이 그 위에 서 있다.

a081의 캐시가 마지막 성공이 아니라 마지막 **결과**를 서빙하는 이유가 이것이다
(design D3). 이 성질은 눈에 잘 띄지 않으므로 Function Logic Map의 불변식 1로
적었고 `TestAFailedRuntimeReadingIsNotMaskedByThePreviousSuccess`가 고정한다.

**같은 것이 lifecycle 쪽에도 있고 그쪽이 더 위험하다.** `positionRow.Managed`는
캐시된 lifecycle만 읽고 `ProjectManagement`는 `Managed`를 `EffectiveKnown`보다
먼저 본다. 즉 stale 목록이 되살아나면 runtime이 정직해도 행이 `엔진 관리`를
주장한다. 초안은 이 절반을 논증도 테스트도 하지 않았다 —
`TestAFailedLifecycleReadingIsNotMaskedByThePreviousSuccess`가 지금 고정한다.

## I5 — 테스트 하네스가 이 경로를 실행하지 않았다

`newDashboardHarness`는 `Options.PositionPolicies`를 배선하지 않는다. 그래서
a080의 예산 테스트를 포함한 대부분의 화면 테스트가 `decoratePositionRows`의
**엔진 미배선 갈래**만 실행했고, 엔진 읽기 횟수를 세는 테스트가 하나도 없었다.

a081은 하네스를 바꾸지 않고 `countingPolicyHarness`를 따로 만들었다 — 기존 703건의
판정 근거를 건드리지 않기 위해서다. 다만 "기본 하네스가 엔진을 배선하지 않는다"는
사실 자체는 다음에도 같은 종류의 사각지대를 만들 수 있다.

## I6 — 리뷰어 셋에게 같은 워킹 트리를 줬다

독립 리뷰 3인이 같은 체크아웃에서 동시에 돌았다. 한 명이 변이 검증으로 소스를
바꾸는 동안 다른 한 명이 테스트를 돌려 6건 실패를 관측했고, 그것을 "환경 위험"으로
보고했다. 결과에는 영향이 없었다 — 격리 사본에서 재확인했고 최종 트리도 온전하다.

그래도 이것은 운이 좋았던 것이지 설계가 아니다. 다음부터 리뷰어는 각자 worktree를
받아야 한다. 특히 **변이 검증을 허용하는 브리프**를 주는 경우에는 필수다.

## I7 — 자동 재로드 없는 화면의 엔진 읽기는 상한이 없다

`positionPolicySummary`(설정 탭)와 `/position-management`는 렌더마다 `List`를
직접 부른다. 의도적이다 (design D6) — 사람이 여는 만큼만 들기 때문이다.

다만 "사람이 여는 만큼"에 기술적 상한은 없다. 새로고침을 연타하는 운영자는
엔진의 단일 쓰기 커넥션에 그만큼 닿는다. 브라우저 하나가 손절 판정 간격에
영향을 줄 수 있다는 점에서 a081이 고친 결합의 작은 판본이 남아 있는 셈이다.

이 change에서 고치지 않는 이유는 둘이다. 명령 화면의 목록은 capability 발행의
근거라 신선해야 하고, 사람의 연타는 자동 재로드와 달리 **의도적이며 유한하다.**
자동 재로드가 있는 화면과 성격이 다르므로 같은 해법을 쓰지 않는다. 다음에
이 화면들에 자동 갱신을 붙이려는 시도가 있으면 그때 이 항목을 먼저 읽어야 한다.
