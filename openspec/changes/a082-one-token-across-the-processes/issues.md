# a082 · Issues

작업 중 확인한 사실 중 이 change가 고치지 않는 것, 또는 다른 곳에 남는 것.

## I1 — `send()`는 401에만 재시도하고 403에는 안 한다

```go
// internal/official/client.go:329
if code == http.StatusUnauthorized {
	tok, err = c.tm.refresh(ctx)
```

그런데 `classifyStatus`는 **401과 403을 같이** 인증 거부로 친다. 브로커가 낡은
토큰에 403을 준다면 재시도가 아예 없고, 이 change의 채택 갈래에 **닿지도 않는다**.

지금 고치지 않는 이유는 브로커가 실제로 무엇을 주는지 모르기 때문이다. 재시도
의미를 근거 없이 넓히는 것은 안전 방향이 아니다 — 403은 "이 자격증명으로는 안
된다"일 수도 있고, 그때 재시도는 무의미한 호출을 늘린다.

**이 change의 상태 코드 수정이 그 근거를 만든다.** 배포 뒤 로그에 401만 나오면
현재 설계로 충분하고, 403이 나오면 후속 change에서 403 재시도를 논증해야 한다.
task 6.7의 실측이 이 질문에 답한다.

## I2 — 세 프로세스가 config 디렉터리 하나를 공유하는 구조 자체

```
tossos-tossos-1   PID 7   console
tossos-tossos-1   PID 16  engine run
tossos-httpapi-1  PID 7   httpapi
```

셋 다 `/var/lib/tossos/config`를 본다. 토큰 캐시는 그중 하나일 뿐이다. 같은
디렉터리에 `openapi-credentials.json`, `engine.lock`, `engine-run.json`,
`config.json`, `performance.db`가 있다.

이 change는 토큰 파일만 고친다. **다른 공유 파일에 같은 형태의 가정이 있는지는
확인하지 않았다.** `config.json`은 `adoption_flock_unix.go`가 flock으로 직렬화하고
`engine.lock`은 `enginelock`이 쓰므로 그 둘은 의식적으로 다뤄져 있다. 나머지는
별도로 볼 일이다.

## I3 — 브로커가 토큰 하나만 살려 둔다는 것은 추론이다

직접 문서로 확인한 사실이 아니다. 근거는 관측의 유일한 정합적 설명이라는 것이다 —
옛 토큰이 계속 유효하면 401이 없고, 401이 없으면 `refresh()`가 안 불리고, 그러면
`saveCache`의 유일한 호출자인 `exchange()`가 안 돌고, 24시간 토큰의 캐시 파일이
1분에 7번 다시 쓰일 수 없다.

이 change의 수정은 그 추론이 틀려도 해롭지 않다. 토큰이 서로를 안 죽인다면 401이
안 나고 채택 갈래는 그냥 안 돌 뿐이다. 다만 **왜 고쳤는지에 대한 설명**은 그때
틀린 것이 되므로, 배포 실측(6.7)에서 교환 횟수가 안 떨어지면 이 항목을 먼저 다시
읽어야 한다.

## I4 — `saveCache`는 best-effort이고 이 change가 그 성질에 기댄다

```go
_ = m.saveCache(ct) // best-effort; do not fail the call if disk write fails
```

쓰기가 실패하면 이 프로세스는 브로커가 아는 토큰을 메모리에만 갖고, 파일에는 옛
토큰이 남는다. 그러면 `stampCacheFile()`이 옛 mtime을 잡고, 다음 `token()`이 디스크를
다시 읽어 **옛 토큰을 쓴다** — 방금 산 토큰을 잃는다.

의도한 방향이다. 옛 토큰은 다른 프로세스가 쓰고 있을 수 있는 유효한 토큰이고,
그것이 죽었으면 401 뒤 교환으로 회복한다. 다만 디스크가 계속 쓰기 불가면 이
프로세스는 요청마다 401→교환을 반복하게 된다. **쓰기 실패를 아무도 보고하지
않는다**는 것이 원래 계약이고 이 change는 그것을 바꾸지 않았다.

## I5 — 기존 테스트 1건을 고쳤다

`orders_raw_test.go:TestRawReadsClassifyErrorsLikeEveryOtherRead`가 판정을
`err == ErrAuth`로 하고 있었다. 이 change가 sentinel을 감싸면서 깨졌다.

`errors.Is`로 바꿨다. 근거는 **그것이 production의 계약**이라는 것이다 — 그 오류를
읽는 다섯 곳(`execgw/retry.go:60`, `classify.go:111`, `failclosed.go:210`,
`cmd/tossctl/soak.go:613`, `openapi.go:56`) 전부 `errors.Is`/`errors.As`를 쓰고,
그 테스트의 주석 자신이 "the retry matrix classifies by those sentinels"라고
적어 두었다. `==`는 계약보다 엄격해서 production이 받아들이는 오류를 거부했다.

같은 표의 429·500 행도 함께 바꿨다. 감싸지 않은 sentinel에 대해
`errors.Is(x, x)`는 `x == x`와 같으므로 판정력은 그대로이고, 한 표 안에 두 판정
방식이 섞이는 것이 읽는 사람에게 더 나쁘다.

약화가 아니라는 증거는 `TestNothingDecidesAnAuthRefusalByReadingItsMessage`가
따로 고정한다 — 같은 문구를 가진 무관한 오류는 `errors.Is`를 만족하지 않는다.

## I6 — 헤드라인 테스트는 채택 갈래 없이도 통과한다

`TestTwoProcessesSharingOneCacheFileStopBuyingTokensFromEachOther`(24 요청에
23 교환 → 3 이하)는 **파일 변경 감지(D2)만으로 GREEN이 된다.** 채택 갈래(D1)를
지워도 통과한다.

변이 검증에서 그것을 확인하고 D1을 전용 테스트로 따로 세웠다 —
`TestARotationThatLandsMidRequestCostsNoToken`. 그 테스트가 재는 것은 D2가 볼 수
없는 창이다: 토큰을 건네준 **뒤**, 브로커가 답하기 **전**에 일어난 회전.

즉 **D2가 무거운 일을 하고 D1은 남은 경주를 닫는다.** 두 수정의 비중이 설계
문서가 처음 암시한 것과 다르므로 여기 적어 둔다. D1을 뺄 수 있느냐고 물으면
답은 "뺄 수 있지만 그러면 회전이 일어날 때마다 교환을 한 번 더 사고, 그 교환이
방금 회전한 프로세스의 토큰을 죽인다"다.
