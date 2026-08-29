# a082 · Design

> **개정 2026-08-05 (독립 리뷰 반영).** 초안은 수정을 둘로 나눴다 — 채택-우선(D1)과
> 캐시 파일 mtime 확인(D2). **D2는 철회했다.** 적대적 리뷰가 그것이 degraded
> 상태에서 base보다 나쁘다는 것을 실측으로 보였고, 그 근거가 D3(락 거부)의 근거와
> 같은 문장이었다. 아래는 개정 후 설계이며 철회 기록은 D2에 남긴다.

## D1 — 채택-우선이 핑퐁을 끊는 지점

핑퐁의 엔진은 `refresh()`의 무조건 교환이다.

```go
// base f1aae509, token.go:89-93
func (m *tokenManager) refresh(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exchange(ctx)
}
```

401이 말하는 것은 **"내가 제시한 토큰이 낡았다"**이지 **"새 토큰을 발급받아야
한다"**가 아니다. 둘을 같게 취급한 것이 결함이다. 다른 보유자가 이미 새 토큰을
받아 파일에 써 뒀다면 그것을 받아 쓰는 것이 맞고, 그때 새로 교환하면 그 보유자의
토큰을 죽인다.

수정 후 `refresh(ctx, refused)`는 이 순서다.

1. 파일을 읽는다.
2. 유효하고 **`refused`와 다르면** 채택하고 `adopted=true`로 반환한다.
3. 아니면 교환하고 `adopted=false`로 반환한다.

### `refused`를 인자로 받는 이유

초안은 `m.cache.AccessToken`을 "방금 실패한 토큰"으로 **추론**했다. 리뷰가 그
추론을 깼다 — 엔진은 supervised loop마다 goroutine을 띄워 client 하나를 공유하므로
(`runtime.go:277-283`), 형제 goroutine이 요청 생성과 거부 도착 사이에 `m.cache`를
바꿀 수 있다. 그 창에서 추론하면 이 goroutine은 교환해 버리고, 방금 형제가 얻은
토큰을 죽인다. **한 프로세스 안에서 일어나는 같은 결함이다.**

호출자는 자기가 거부당한 토큰을 알고 있다. 추론을 없애고 사실을 넘긴다.

### 수렴

A가 T1(파일 현재값)을 쥐고 B·C가 낡은 T0을 쥔 상태.

| | 사건 | 교환 |
|---|---|---|
| 1 | B 요청 → 401 → refresh(T0) → 파일 T1 ≠ T0 → 채택 | **0** |
| 2 | B 재시도 성공 | 0 |
| 3 | C도 같은 경로 | **0** |
| 4 | 셋이 T1로 수렴, 토큰 수명 동안 조용 | 0 |
| 5 | 만료 → 셋이 각자 교환 (동시에 만료 판정을 하므로) | 3 |
| 6 | 마지막에 발급된 것으로 다음 거부에서 수렴 | 0 |

만료 시점의 3회는 초안의 "수명당 1회"보다 많다. 초안이 틀렸다 — 세 프로세스의
`isStillValid`(60초 skew)가 같은 순간에 뒤집히므로 셋 다 교환한다. 그래도
**1분에 7회에서 하루에 3회**다.

## D2 — 철회: 캐시 파일 mtime 확인

초안은 `token()`이 메모리 캐시를 믿기 전에 파일 mtime을 보게 했다. 회전마다
프로세스당 401 한 번을 아끼는 것이 목적이었다. **철회한다.**

적대적 리뷰가 실측한 것 셋이다.

**① stat 실패를 "바뀌었다"로 읽으면 요청마다 교환한다.** 파일을 stat할 수 없으면
(ENOSPC, 권한 드리프트, bind mount 부재로 read-only root에 떨어짐) 메모리 갈래를
건너뛰고, `loadCache`가 없는 파일에 `(nil, nil)`을 주고, `isStillValid(nil)`이
false라 **매 호출 교환**한다. 유효한 24시간 토큰을 손에 쥐고서. 측정: 같은
조건에서 base 1회, 초안 10회. 그리고 그 교환은 `m.mu`를 쥔 채 도는 네트워크
호출이라 exit 루프의 가격 읽기가 매 사이클 그 뒤에 줄 선다.

**② 읽고 나서 stamp하면 감지기가 영구히 죽는다.** `loadCache()` → `stampCacheFile()`
사이에 다른 보유자가 쓰면, **옛 바이트에 새 mtime**이 찍힌다. 그 뒤로
`cacheFileChanged()`는 영원히 "안 바뀜"이라 401이 강제할 때까지 회전을 못 본다 —
D2가 없애겠다던 바로 그 401이다.

**③ D3의 근거를 D2가 어긴다.** stat은 `m.mu`를 쥔 채 돌고, `context`를 받지 않아
취소도 데드라인도 걸 수 없다. 측정: ext4 55ns → 775ns, FUSE 87µs. NFS가 응답하지
않으면 mount timeout까지 uninterruptible sleep이다. **락을 거부한 이유가 이
경로에 무한정 기다릴 수 있는 것을 놓지 않기 위해서였는데, stat이 정확히 그것이다.**

철회의 대가는 회전마다 프로세스당 401 한 번이고, 그것은 `send()`의 재시도가
흡수한다. 하루 3회다. 그 값에 위 셋을 사지 않는다.

**형제 goroutine 전용 갈래도 같은 이유로 넣지 않았다.** 메모리 캐시를 먼저 보는
갈래를 따로 두면 파일 읽기를 아끼지만, 변이 검증에서 그것을 지워도 아무 테스트가
깨지지 않았다 — `exchange`가 반환 전에 파일을 쓰므로 형제의 토큰은 파일 갈래가
이미 찾아낸다. 근거 없는 코드는 남기지 않는다.

## D3 — 파일 락을 쓰지 않는다

이 저장소에는 `internal/config/adoption_flock_unix.go`와 `internal/enginelock`이라는
flock 선례가 둘 있다. 여기서는 쓰지 않는다.

- D1이 수렴한다. 락이 막는 것은 만료 시점의 동시 교환뿐이고, 그때도 각자 유효한
  토큰을 받는다. 오류가 아니라 낭비다.
- 토큰 획득은 **엔진의 모든 읽기가 지나는 경로**이고 손절 판정도 그 위에 있다.
  여기에 블로킹 락을 놓으면 락을 쥔 채 느려진 프로세스가 나머지의 손절 판정을
  지연시킨다. 안전 불변식 3이 금지하는 방향이다.
- D2 철회가 이 논거를 더 강하게 만든다. 락을 거부하면서 취소 불가능한 stat을
  같은 자리에 놓는 것은 앞뒤가 안 맞았다.

## D4 — 상태 코드를 버리지 않는다

```go
// base f1aae509, errors.go:41-45
case code == 401 || code == 403:
	if reIPWord.Match(bytes.ToLower(body)) {
		return ErrIPNotAllowed
	}
	return ErrAuth
```

401과 403이 같은 맨 sentinel이 된다. 이 결함이 사흘 동안 `authentication failed`
한 줄로만 보인 이유다.

`fmt.Errorf("%w (HTTP %d)", ErrAuth, code)`로 감싼다. 판정 규칙은 한 글자도 안
바뀐다 — 401/403 구분도, `ip` 낱말 검사도, 반환하는 sentinel도 그대로다.

`errors.Is(err, ErrAuth)`가 계속 성립해야 한다. 리뷰가 저장소 전체를 독립적으로
훑어 확인했다: production 소비자 7곳(`execgw/retry.go:60`, `classify.go:111`,
`failclosed.go:210`, `cmd/tossctl/soak.go:609,613`, `openapi.go:50,56`,
`errors.go:66-75`) 전부 `errors.Is`/`errors.As`이고, `orders_raw_test.go` 수정 뒤
저장소에 남은 sentinel `==` 비교는 **0건**이다.

`failclosed.go:190-198`의 `refusalBody`가 `err.Error()` fallback을 명시적으로
거부하고 `APIError.Body`만 읽으므로, `(HTTP 401)` 접미사는 운영자 분기 분류에
닿지 않는다.

**본문은 싣지 않는다.** 응답 본문에 계좌 식별자가 들어올 수 있고 이 에러는 로그로
나간다. `*APIError` 갈래는 원래대로 본문을 계속 싣는다 — 그쪽은 passthrough다.

## D5 — 채택한 토큰은 검증된 토큰이 아니다

리뷰가 찾은 것 중 가장 무거운 것이다.

채택한 토큰의 생존은 `ExpiresAt`에서 **추론**한 것이고, `send()`에는 재시도가
하나뿐이다. 회전 경계에서 **마지막에 쓴 보유자와 마지막에 발급받은 보유자가 다를
수 있다** — `saveCache`가 POST 반환 뒤에 돌므로 쓰기 지연 몇 ms면 갈린다. 그러면
파일에 브로커가 이미 버린 토큰이 있고, 채택은 유일한 재시도를 그것에 쓴다.

측정: 그 시나리오에서 base는 `err = <nil>`(갱신이 새로 발급 → 재시도 성공),
초안은 `official: authentication failed (HTTP 401)`.

그 거부의 대가는 작지 않다.

```
errors.go → execgw/classify.go:111 ClassAuthFatal
          → execgw/retry.go:334 Gate.Block(ReasonBrokerAuthRejected)
          + escalateCredentialFailure — 재시작이 못 풀도록 영속화 (retry.go:363-371)
          → EntryGate.Clear는 운영자 전용, "nothing automatic calls it" (retry.go:507-509)
```

그리고 그것을 올린 exit 루프 사이클은 `observe()`가 중단되어 **손절 판정을 아예
하지 않는다**(`exitloop.go:699-708`).

그래서 `refresh`가 **채택했는지 발급했는지**를 반환하고, `send()`는 채택한 토큰이
또 거부당했을 때만 한 번 더 돈다. 두 번째 회차는 채택할 수 없다 — `refresh`는
자기가 거부당한 토큰을 절대 반환하지 않으므로 추측이 자기 자신의 대체에서
배제되고, 두 번째 pass는 교환한다. 상한은 갱신 2회다.

## D6 — 쓰기는 원자적이어야 한다

`os.WriteFile`은 truncate 후 write다. 그 사이에 도착한 읽기는 빈 파일을 파싱하고,
토큰이 없다고 판단하고, 하나 사고, 방금 쓴 보유자의 토큰을 죽인다.

**a082가 그 창을 최대화한다.** 채택하는 읽기는 **다른 보유자가 방금 썼을 때**
정확히 일어나기 때문이다.

임시 파일 + rename으로 바꾼다. 이 저장소가 `internal/config/adoption_io.go`와
`position_policy_transport.go:290`에서 이미 쓰는 형태다.

측정: 400회 동시 읽기/쓰기에서 plain write는 **244회**가 빈 파일이나 파싱 실패를
봤고, rename은 0회다.

## D7 — 무엇을 관측으로 판정할 것인가

단위 테스트는 manager 둘이 한 파일을 공유하는 것으로 프로세스 둘을 흉내낸다.
결함이 **프로세스 경계가 아니라 캐시 계층**에 있으므로 충분하다.

배포 후 실측은 `openapi-token.json`의 mtime을 세는 것으로 한다. 지금 1분에 7회이고,
수정 후에는 **관측 창(수십 분) 안에 0회**여야 한다 — 토큰 수명이 24시간이고 만료
시점에만 교환하므로.
