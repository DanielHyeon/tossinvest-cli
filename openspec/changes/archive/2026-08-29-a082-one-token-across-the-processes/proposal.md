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

### 거부당한 토큰이 무엇인지는 호출자가 알려준다

초안은 그것을 캐시 상태에서 **추론**했다. 적대적 리뷰가 그 추론을 깼다 — 엔진은
supervised loop마다 goroutine을 띄워 client 하나를 공유하므로, 형제가 요청 생성과
거부 도착 사이에 캐시를 바꿀 수 있고 그 창의 추론은 같은 결함을 프로세스 안에서
되풀이한다. 호출자는 자기가 거부당한 토큰을 알고 있다. 그것을 넘긴다.

### 채택한 토큰이 또 거부당하면 발급해서 다시 시도한다

채택한 토큰의 생존은 만료 시각에서 **추론**한 것이고 `send()`에는 재시도가 하나뿐이다.
회전 경계에서 마지막에 쓴 보유자와 마지막에 발급받은 보유자가 다를 수 있으므로,
파일에 죽은 토큰이 있을 수 있다. 그것에 유일한 재시도를 쓰고 인증 실패를 올리면
**엔트리 게이트가 잠기고 그 잠금은 재시작으로 안 풀린다.** 그때만 한 번 더 돈다
(design D5).

### 캐시 파일 쓰기를 원자적으로 바꾼다

`os.WriteFile`은 truncate 후 write다. 그 사이에 읽으면 빈 파일을 보고 토큰이
없다고 판단해 하나 사고, 방금 쓴 보유자의 토큰을 죽인다. **이 change가 그 창을
최대화한다** — 채택하는 읽기는 다른 보유자가 방금 썼을 때 정확히 일어나기
때문이다. 임시 파일 + rename으로 바꾼다 (design D6).

### 캐시 파일 mtime 확인은 넣지 않는다

초안에 있었고 **철회했다.** stat 실패를 "바뀌었다"로 읽으면 파일을 못 읽는 동안
요청마다 교환하고, 읽은 뒤 stamp하면 감지기가 영구히 죽고, 무엇보다 취소할 수
없는 파일시스템 호출을 손절 판정이 지나는 경로의 뮤텍스 아래 놓는다 — 락을
거부한 것과 같은 이유로 이것도 거부한다 (design D2·D3).

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
