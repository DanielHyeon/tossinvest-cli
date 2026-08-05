# a082 · 프로세스 셋이 토큰 하나를 나눠 쓴다

- **Feature**: `FEAT-TOS-007` — Secure remote access and signed releases
  (`openapi-onboarding` 스펙의 소유 feature. STORY-TOS-037이 그 스펙을 세웠다)
- **Story**: `STORY-TOS-a082`
- **Spec**: `openapi-onboarding`

## Why

24시간짜리 access token이 **1분에 7번** 재발급되고 있다. 2026-08-05 실측이다.

```
12:06:48  12:07:11  12:07:14  12:07:33  12:07:42  12:07:43  12:07:47  12:07:50
```

8초 간격으로 읽은 토큰 값이 다르다 (`…icdZtw` → `…6X1ChQ`). 매번 진짜 새 교환이다.

`saveCache`를 부르는 곳은 하나뿐이다.

```go
// internal/official/token.go:139 (base 96c621d3) — exchange() 안
_ = m.saveCache(ct) // best-effort; do not fail the call if disk write fails
```

24시간 토큰이 캐시 만료로 재교환될 수는 없다. 그러면 `exchange()`에 닿는 다른
경로는 하나다 — **401 재시도**.

```go
// internal/official/client.go:329-331
if code == http.StatusUnauthorized {
	// Force-refresh the token and retry once.
	tok, err = c.tm.refresh(ctx)
```

### 왜 401이 계속 나는가

**프로세스 셋이 토큰 파일 하나를 공유한다.** 운영 컨테이너에서 확인했다.

| 프로세스 | 위치 | config dir |
|---|---|---|
| `console` (PID 7) | `tossos-tossos-1` | `/var/lib/tossos/config` |
| `engine run` (PID 16) | `tossos-tossos-1` | 같음 |
| `httpapi` (PID 7) | `tossos-httpapi-1` | 같은 bind mount |

셋 다 `official.New(*creds, tokenFile)`로 같은 `openapi-token.json`을 연다.

그리고 `token()`은 **메모리 캐시를 먼저 보고 디스크를 다시 읽지 않는다.**

```go
// internal/official/token.go:66-75
// 1. In-memory cache.
if isStillValid(m.cache, now) {
	return m.cache.AccessToken, nil
}
// 2. Disk cache.
if ct, err := m.loadCache(); err == nil && isStillValid(ct, now) {
```

한 번 교환한 프로세스는 그 토큰을 24시간 붙들고, 다른 프로세스가 디스크 토큰을
바꿔 놓은 것을 **영영 모른다**. 브로커는 `client_credentials` 토큰을 하나만 살려
두므로 — 옛 토큰이 계속 유효하다면 401이 없고, 401이 없으면 `refresh()`가 안 불리고,
그러면 파일이 1분에 7번 다시 쓰일 수 없다 — 한쪽의 교환이 나머지 둘의 토큰을 죽인다.

그래서 핑퐁이 된다. A가 교환 → B·C 사망 → B가 401 → 강제 교환 → A·C 사망 → …

### `ErrAuth`가 새어 나오는 순간

`send()`는 **한 번만** 재시도한다. 보통 그 재시도가 성공해서 아무도 못 본다.
실패가 표면에 나오는 것은 **내 갱신과 내 재시도 사이에 제3의 프로세스가 또 교환할
때**다. 좁은 창이고, 3일에 10회라는 관측 빈도가 정확히 그 모양이다.

```text
reconcile: snapshot discarded after a partial read:
  walking the open-order list: official: authentication failed
```

엔진은 fail-closed로 스냅샷을 버리고 다음 주기에 재시도하므로 지금 손해는 없다.
**문제는 켜는 순간이다.**

```go
// internal/execgw/retry.go:329-336
case ClassAuthFatal:
	// Immediate, no retry: a rejected credential does not improve by
	// being presented again, and a live engine with a dead credential
	// must stop opening positions right now.
	if r.Gate != nil {
		r.Gate.Block(ReasonBrokerAuthRejected, ...)
```

죽은 자격증명에는 맞는 판단이다. 그러나 이 경합에서 **자격증명은 죽지 않았다.
경주에서 졌을 뿐이다.** reconcile은 `ScanOrders`를 직접 불러 Gate를 지나지 않지만
exit 루프(`exitwiring.go:50,77`)와 게이트웨이는 Gate를 단다. 시장을 켜면 이 경합이
거짓 신호로 진입을 막을 수 있다.

### 이것은 fork가 만든 결함이다

토큰 매니저는 upstream 코드이고(`2993784a`, PR #112) 거기서는 CLI 한 번에 프로세스
하나라 메모리 캐시 가정이 옳았다. TossOS가 `3cc2dc77 feat(httpapi): add private
operator daemon`으로 **세 번째 상주 프로세스**를 붙이면서 그 가정이 깨졌다.

## What Changes

### 진 쪽은 새 토큰을 사지 않고 이긴 쪽의 토큰을 받는다

핑퐁을 지탱하는 것은 `refresh()`가 **무조건 교환**하는 것이다. 401은 "내 토큰이
낡았다"는 증거이지 "새 토큰을 발급받아야 한다"는 증거가 아니다.

`refresh()`가 교환하기 전에 디스크를 다시 읽는다. 디스크에 유효하고 **방금 실패한
것과 다른** 토큰이 있으면 그것을 채택하고 끝낸다. 없을 때만 교환한다.

수렴한다. A가 T1을 쥐고 있고 B·C가 낡은 T0을 쥔 상태에서 — B가 401 → 디스크의
T1 채택 → 재시도 성공. **교환 0회, A의 토큰 생존.** C도 같다. 셋이 T1로 모이고
다음 교환은 T1이 실제로 만료되는 24시간 뒤다.

### 디스크가 바뀌었으면 메모리 캐시를 믿지 않는다

위 수정만으로도 수렴하지만, 회전마다 프로세스당 401을 한 번씩 먹는다.
`token()`이 메모리 캐시를 믿기 전에 파일이 바뀌었는지 보고, 바뀌었으면 디스크를
다시 읽는다. 그러면 그 401도 사라진다.

### 로그가 401과 403을 구분한다

```go
// internal/official/errors.go:39-45
case code == 401 || code == 403:
	if reIPWord.Match(bytes.ToLower(body)) {
		return ErrIPNotAllowed
	}
	return ErrAuth
```

상태 코드와 본문을 버리고 맨 sentinel만 돌려준다. 그래서 이 결함이 사흘 동안
"authentication failed" 한 줄로만 보였고, 토큰 경합인지 권한·IP 문제인지 구분할
방법이 없었다. 상태 코드를 메시지에 실어 보낸다. `errors.Is`는 그대로 성립해야
하므로 sentinel을 wrap한다.

**이 수정이 다음 질문에 답한다**: 브로커가 낡은 토큰에 401을 주는가 403을 주는가.
`send()`는 401에만 재시도하므로, 답이 403이면 위 수정이 닿지 않는다 (issues I1).

## Non-Goals

- **파일 락을 도입하지 않는다.** 채택-우선만으로 수렴하고, 만료 시점에 두
  프로세스가 동시에 교환해도 각자 유효한 토큰을 얻으므로 오류가 아니다. 락은
  이 change가 고치는 결함의 원인이 아니다 (design D3).
- **`send()`의 403 재시도를 지금 넣지 않는다.** 브로커가 실제로 403을 주는지
  모르는 상태에서 재시도 의미를 바꾸는 것은 근거 없는 변경이다. 위 로그 수정이
  그 근거를 만든다 (issues I1).
- **자격증명·인증 흐름을 바꾸지 않는다.** 키·시크릿 저장, 온보딩, IP 허용 판정
  어느 것도 건드리지 않는다.
- **`ClassAuthFatal`의 Gate 차단 정책을 바꾸지 않는다.** 진짜 죽은 자격증명에
  대해서는 그 즉시성이 옳다. 이 change는 거짓 신호의 **원인**을 없앤다.
- **공식 API 호출을 늘리지 않는다.** 오히려 줄인다 — 1분 7회 교환이 24시간 1회가
  된다.

## Impact

- `internal/official/token.go` — `token()`과 `refresh()`의 내부 로직 변경.
  기존 함수이고 **인증 경로는 High-risk**이므로 Function Logic Map과 Branch Test
  Map을 편집 전에 작성한다.
- `internal/official/errors.go` — `classifyStatus`가 상태 코드를 실어 보낸다.
  기존 함수 변경이므로 같은 증거를 만든다.
- `openspec/specs/openapi-onboarding` — **MODIFIED 1건**
  (`persistent credential and token lifecycle`). "existing official client
  renewal behavior"가 프로세스 하나를 가정하고 있었다는 것이 이 change의 내용이다.
- rate budget: **감소.** OAuth 교환이 1분 7회에서 토큰 수명당 1회가 된다.
  다른 공식 API 호출 수는 그대로다.
