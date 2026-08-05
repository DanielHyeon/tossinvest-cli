# a081 · 화면들은 엔진을 한 번만 읽는다

- **Feature**: `FEAT-TOS-004` — Operator console controls and visibility
- **Story**: `STORY-TOS-a081`
- **Spec**: `operator-console`

## Why

`/positions`와 `/dashboard`는 렌더할 때마다 엔진 프로세스로 **동기 읽기 2회**를
한다. TTL도 캐시도 없다.

```go
// internal/console/portfolio_pages.go:93-94 (base 840b3377)
runtime, _ = c.opts.PositionPolicies.Runtime(ctx)
if states, err := c.opts.PositionPolicies.List(ctx); err == nil {
```

둘 다 엔진 프로세스에 닿는다. `Runtime`은 descriptor를 매번 다시 읽고 Unix
소켓을 **매 호출 새로 dial**한다.

```go
// cmd/tossctl/position_policy_commander.go:35-41
func (r positionPolicyRuntimeDescriptorReader) Runtime(ctx context.Context) (...) {
	client, err := positionpolicyrpc.DialRuntime(ctx, r.descriptorPath)
	...
}
```

`List`는 더 무겁다. 엔진 쪽에서 이렇게 처리된다.

```go
// internal/app/engine/position_policy_command.go:138-142
func (s *PositionPolicyCommandService) List(ctx context.Context) ([]positionpolicy.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.j.PositionPolicies(ctx)
}
```

`s.j`는 **엔진이 소유한 쓰기 핸들**이고 그 핸들은 커넥션이 하나다.

```go
// internal/journal/journal.go:151
db.SetMaxOpenConns(1)
```

그래서 콘솔의 `List()`는 exit 루프의 판정 트랜잭션과 **같은 커넥션에서
직렬화**된다. 그리고 exit 루프는 ticker가 아니라 `작업 후 sleep`이다.

```go
// internal/app/engine/exitloop.go:351-359
o.reportCycle(o.ObserveOnce(ctx))
if err := o.clk.Sleep(ctx, o.Interval()); err != nil {
```

그 커넥션에서 뺏은 시간은 다음 판정까지의 간격에 **그대로 더해진다**.

### 이것은 오늘 이미 지불 중인 비용이다

탭 하나가 열려 있으면 30초마다 읽기 2벌, 두 화면을 함께 열어 두면 30초마다
4벌이다. 화면 수와 재로드 주기가 그대로 엔진 부하가 된다 — 콘솔이 화면을 몇 개
띄웠는지가 손절 판정 간격에 영향을 준다는 뜻이고, 그것은 어떤 재로드 주기에서도
성립해서는 안 되는 결합이다.

이 결합은 a080(재로드 주기를 5초로) 독립 리뷰에서 발견되었다
(`a080/review.md` F1). a080이 만든 결함이 아니라 a080이 **드러낸** 기존 결함이며,
a080을 철회해도 남는다. 그래서 별도 change로 먼저 고친다.

### 왜 캐시가 정답인가

읽기 횟수가 **화면 수와 재로드 주기에 묶여 있는 것**이 결함이고, 캐시는 그 묶음을
끊는다. 다만 초안이 여기 적었던 "이 값들은 렌더 단위로 움직이지 않는다"는 한
문장은 **거짓이었고**, 독립 리뷰가 그것을 깼다. 실제는 이렇다.

| 값 | 무엇이 움직이나 | 엔진 쓰기 커넥션 | 간격 |
|---|---|---|---|
| `Runtime().Effective` (adoption 설정) | 엔진 인스턴스 기동값 — 생성자에서 고정 | 안 닿는다 | 5초 |
| `Runtime().Blocks` (대사 차단) | **살아 있는 reconcile tracker** | 안 닿는다 | 5초 |
| `List()` (lifecycle status/generation) | 운영자 Apply, 엔진 adoption | **닿는다** | 30초 |

안전 논거가 적용되는 것은 세 번째 줄뿐이다. 두 번째 줄은 오래 붙들면 화면이
낙관적으로 틀린다. 그래서 두 읽기가 각자의 간격을 갖는다 (design D2).

보호선의 값 자체(워터마크, 기준선, 저장 exit 근거)는 `c.positions()`의 journal
읽기가 가져오며 이 change는 그 경로를 건드리지 않는다.

## What Changes

### 엔진 도달 횟수를 렌더 횟수에서 분리한다

콘솔에 `positionPolicyCache`를 두고, 두 라인 화면의 표시 경로
(`decoratePositionRows`)가 그것을 통해 읽는다. `holdingsCache`가 브로커에 대해
하는 일과 같은 일을 엔진에 대해 한다 — 같은 파일 옆자리, 같은 형태, 같은 이유.

- 간격당 각 읽기 1회가 상한이다.
- 뮤텍스를 읽기 전체에 걸어 동시 렌더 여러 건이 읽기 1벌만 만든다.
- 탭 수와 재로드 주기가 엔진 도달 횟수를 바꾸지 못한다.

### 두 읽기의 간격은 서로 다르다

`List`는 30초, `Runtime`은 5초다. 초안은 30초 하나였고 독립 리뷰가 그 전제를
깼다 — `Runtime`은 **DB를 아예 건드리지 않으므로** 안전 상한이 필요 없고, 그것이
실어 나르는 대사 차단은 살아 있는 tracker 상태라 30초를 붙들면 화면이 이미 편입이
멈춘 보유를 "편입 예약됨"으로 표시한다. 자세한 근거는 design D2.

### 마지막 성공이 아니라 마지막 **결과**를 캐시한다

`holdingsCache`는 실패 시 직전 성공을 나이와 함께 계속 보여준다. 이 캐시는
그렇게 하지 않는다. 한 시도의 결과가 성공이면 성공을, 실패면 실패를 간격 동안
서빙한다.

스펙이 그렇게 요구하기 때문이다.

> runtime unavailable인 non-managed 행은 desired를 effective로 위장하지 않고
> `UNKNOWN`과 runtime unavailable 이유를 표시해야 한다 (SHALL)
> — `operator-console` "편입 보조 상태는 candidate와 reconcile 차단을 함께 설명한다"

죽은 엔진의 마지막 성공을 계속 서빙하면 화면은 아무도 유지하지 않는 보호 설정을
`EffectiveKnown`으로 주장하게 된다. 시도를 간격으로 묶되 결과를 있는 그대로
캐시하는 것이 이 SHALL을 지키는 유일한 형태다.

시도를 묶는 것 자체도 필요하다. 엔진이 답을 못 하는 순간은 렌더마다 소켓을 다시
여는 것이 가장 나쁜 순간이다 (`holdings.go:149-152`가 브로커 429에 대해 같은
판단을 이미 적어 두었다).

### 갱신은 요청의 컨텍스트로 돌지 않는다

reading은 공유되고 그것을 마침 취한 요청은 공유되지 않는다. 브라우저가 렌더를
버렸을 때 그 취소를 캐시에 기록하면 건강한 엔진인데도 두 화면의 보호선이 한 간격
동안 사라지고 재로드로 회복되지 않는다. 독립 리뷰 둘이 재현했다. 갱신은
`context.WithoutCancel` 위에서 자기 타임아웃을 걸고 돈다 (design D4b).

### 콘솔 자신의 정책 변경은 즉시 보인다

운영자가 `/position-management`에서 정책을 Apply하면 콘솔이 캐시를 버린다. 자기가
방금 한 일이 안 보이는 화면은 캐시가 아니라 버그다. 격리 해제는 이 캐시를 지나지
않으므로 무효화하지 않는다 (design D5).

### 자동 재로드가 없는 화면은 계속 직접 읽는다

`/position-management`와 설정 탭은 캐시를 거치지 않는다. 상한이 필요한 이유는
화면이 사람 없이 스스로 반복하기 때문이고, 이 화면들에는 그 이유가 없다. 더해서
capability 발행의 근거가 되는 목록은 그 행동을 위해 방금 취한 읽기여야 한다.

## Non-Goals

- **엔진 코드를 바꾸지 않는다.** `PositionPolicyCommandService`, journal 핸들,
  exit 루프 어느 것도 건드리지 않는다. 부하를 만드는 쪽은 콘솔이고, 고칠 곳도
  콘솔이다.
- **`holdingsTTL`과 브로커 경로를 바꾸지 않는다.** 이 change는 공식 API 호출을
  한 건도 늘리거나 줄이지 않는다.
- **읽기가 신선할 때 화면이 보여주는 값을 바꾸지 않는다.** 그 단서 없이는 참이
  아니다 — 수동 새로고침은 오늘 언제나 새 읽기를 가져오지만 캐시 뒤에서는 간격
  안의 값을 받을 수 있다 (design D8).
- **재로드 주기를 바꾸지 않는다.** 그것이 a080이고, a080은 이 change가 land한
  뒤에 재개한다.
- **`c.positions()`의 journal 읽기를 캐시하지 않는다.** 그것은 콘솔의 read-only
  핸들이고 WAL이라 엔진 쓰기를 막지 않는다. 보호선의 신선도가 거기서 나오므로
  캐시하면 a080의 목적 자체가 사라진다.
- **`Settings.Load()`와 `StrategyRuntime`은 대상이 아니다.** 전자는 콘솔 프로세스
  안의 로컬 파일 읽기이고 후자는 다른 화면의 경로다. 같은 형태의 결합이 있는지는
  별도로 본다 (`issues.md`).

## Impact

- `internal/console/position_policy_cache.go` — **신규 파일**. 캐시와 그 근거.
- `internal/console/portfolio_pages.go` — `decoratePositionRows`가 커맨더 대신
  캐시를 읽는다. 기존 함수의 내부 로직 변경이므로 Function Logic Map과 Branch
  Test Map을 **편집 전에** 작성한다.
- `internal/console/console.go` — 캐시 필드와 생성.
- `internal/console/position_policy.go` — 성공한 Apply 뒤 무효화 1줄.
- `internal/console/exit_quarantine.go` — 무효화를 **넣지 않는** 이유를 주석으로
  남긴다. 격리 상태는 이 캐시를 지나지 않는다.
- `openspec/specs/operator-console` — **ADDED 1건.** 기존 요구사항과 충돌하지
  않는다. `rate budget 보호`는 브로커 예산이고 이 change는 브로커에 닿지 않는다.
- rate budget: **변동 0.** 공식 API 호출 수가 그대로다. 줄어드는 것은 콘솔이
  엔진 프로세스에 거는 로컬 IPC 횟수다.
