# a102 — 부팅이 엔진을 굶기지 않는다

## Why

2026-08-13 02:03:30.545Z, 배포 재기동 직후 엔진이 죽었다.

```
Error: reconcile: restart recovery did not complete:
       reconcile: snapshot discarded after a partial read:
       walking the open-order list: official: rate limited
```

그리고 **아무도 다시 세우지 않았다.** 콘솔은 healthy였고, 포지션 3건은
`trading.conditional=False`라 브로커에 거치 손절이 없었으며, KRX는 장중이었다.
`engine.operating_mode`가 죽기 2.3초 전에 그 상태를 정확히 적어 두었다 —
`"protection":"UNWIRED"`, *"보호는 이 프로세스가 살아 있는 동안만 유효하다"*.

### 부팅 타임라인 (컨테이너 로그 · engine.log, 밀리초 실측)

| UTC | 사건 |
|---|---|
| 02:03:26.559 | `startEngine`이 detached spawn. `engineStartProbe = 3s` 시작 |
| 02:03:28.236 | 엔진 배너 `engine.operating_mode` · protection UNWIRED |
| **02:03:29.559** | probe 3초 만료 → 콘솔 "엔진 자동 시작: 엔진을 시작했다" |
| **02:03:29.561** | **2ms 뒤** soak 자동시작 → `orders 2363 on 26 page(s)` 전량 순회 |
| 02:03:30.545 | 엔진 사망 — restart recovery가 429를 받고 fail-closed |

### 세 겹의 원인

1. **probe의 전제가 이 경로에서 거짓이다.** `cmd/tossctl/engineproc.go:214-216`은
   *"모든 거부는 루프가 시작되기 전에 일어나므로, probe 뒤에도 살아 있는 엔진은
   그 전부를 통과한 엔진이다"*라고 적었다. 그런데 restart recovery는 브로커 계좌를
   읽고, 그 읽기가 probe **뒤에** 실패할 수 있다. 986ms 차이로 놓쳤다.

2. **엔진→서베이 순서가 실제로 직렬화하지 못한다.** `cmd/tossctl/console.go:349-351`은
   순서의 이유를 정확히 적었다 — *"둘은 한 계좌의 rate 예산을 공유한다"*. 의도는 맞다.
   그러나 그 순서가 주는 것은 **3초 머리 시작뿐**이고, 26페이지·2363건 순회는 3초로
   끝나지 않는다. 서베이는 probe 만료 2ms 뒤에 같은 `/api/v1/orders`를 때렸다.

3. **시작 뒤를 아무도 보지 않는다.** `runConfiguredEngineAutostart`
   (`cmd/tossctl/engineautostart.go:57-87`)는 `start()`를 한 번 호출하고 문자열을
   돌려주고 끝난다. 재시도도 watchdog도 없다. `docs/operations.md:104-106`도 그렇게
   적혀 있다 — 자동시작은 supervisor가 아니라 **부팅 시 버튼 한 번 누르기**다.

### 비대칭이 거꾸로다

같은 429가 **정상 운영 중에는 살아남는다.** 2026-08-12T12:52:37.661Z:

```
reconcile.mismatch  error: ...walking the open-order list: official: rate limited
                    detail: the reconciliation cycle failed; it will be retried next period
```

engine.log에 `official: rate limited`는 **15회** 나온다. 주기 실패는 다음 주기가
받아 준다. **기동만 한 번에 포기한다.** 그런데 기동 중이 바로 보호가 아직 없는
시점이다. 끈질겨야 할 쪽이 덜 끈질기다.

read-only 서베이는 15초 백오프로 두 번 재시도한다(soak.log 02:03:30 · 02:03:46 ·
02:04:02). **손절을 든 쪽이 조회만 하는 쪽보다 쉽게 포기한다.**

## What Changes

두 겹으로 고친다. 하나는 안전망, 하나는 예방이다. 둘 중 하나만으로는 부족하다.

### 겹 1 — 엔진이 rate limit에 죽지 않는다 (안전망)

`Recovery.stableSnapshot`(`internal/reconcile/recovery.go:333-359`)이 rate limit을
영구 실패와 구분한다. **재시도 루프는 이미 있다** — AST 실측 branches **5** /
returns **4** / calls **7**(`analysis/function-logic/internal-reconcile--recovery.stablesnapshot/ast.json`).
`MaxAttempts` 루프와 `Interval` sleep이 이미 그 자리에 있고, `Collect` 오류만
무조건 즉시 이탈한다.

- rate limit이면 **stabilisation attempt를 소모하지 않는다.** 읽기가 없었으므로
  스냅샷도 없고, 안정화 판정의 입력이 아니다.
- 별도의 rate-limit 예산과 백오프로 재시도한다. 그 예산이 소진될 때만 기존처럼
  `ErrRecoveryIncomplete`로 fail-closed한다.
- 백오프는 서베이의 규율(15초) **이상**이다. 보호를 든 쪽이 조회만 하는 쪽보다
  먼저 포기하지 않는다.

방향은 보수적이다. 이 변경은 엔진이 **더 오래 게이트를 닫고 기다리게** 만들 뿐,
잘못 읽은 장부로 매매를 시작하게 하지 않는다.

### 겹 2 — 부팅이 경합을 만들지 않는다 (예방)

서베이 기동이 **3초 타이머가 아니라 엔진의 준비 완료 신호**를 기다린다.

오늘 그 신호는 **없다.** 마커(`engine-run.json`)는 recovery보다 **먼저** 쓰인다 —
배너가 `active marker …`를 찍은 뒤에 recovery가 실패했다. 마커는 "누가 lock을
잡았다"이지 "계좌를 다 읽었다"가 아니다. 엔진은 상태를 이미 갖고 있다
(`Recovery.Complete()`, `recovery.go:299`). 그것을 관측 가능하게 만든다.

- 대기에는 **상한**이 있고, 상한을 넘으면 서베이는 그냥 시작한다. 서베이는 read-only
  선택 기계장치이고, 그것이 없다고 운영자 화면이 없어서는 안 된다
  (`runConfiguredSoakAutostart`의 기존 계약, soakautostart.go:83-86).
- 부팅 노트가 **어느 쪽이었는지 말한다**: "엔진 준비 확인 후 시작" / "상한 초과 —
  엔진 준비 전에 시작". 조용한 상한 초과는 금지다.
- `bootSurvey`의 기존 규칙(이미 도는 서베이를 죽이지 않는다, a101)은 그대로다.

### 왜 둘 다인가

- **겹 2만**: 부팅 창만 막는다. 08-12 12:52처럼 장중에 겹치는 경우는 그대로 남는다.
- **겹 1만**: 매 부팅마다 429를 만드는 구조가 남고, 엔진 기동이 백오프만큼 느려진다.
  보호가 없는 시간이 길어진다.
- 겹 1은 **어디서 겹치든 안 죽게** 하고, 겹 2는 **부팅에서 겹치지 않게** 한다.

## 범위 밖

- `engineStartProbe`의 결과 기반 전환(원래 옵션 1). 겹 2가 서베이를 준비 신호에
  묶으면 probe는 더 이상 서베이 기동의 방아쇠가 아니다. probe 주석의 거짓 전제는
  design에 기록하고 별도로 다룬다.
- 자동시작의 supervisor화(죽으면 다시 세우기). 재기동 정책은 별도 결정이고,
  이 change는 **죽지 않게** 하는 쪽이다.

## Impact

- `internal/reconcile` — High-risk. 대사 경로. 방향은 보수(더 기다린다).
- `cmd/tossctl` 부팅 배선 — 기동 순서. 주문 경로 아님.
- `internal/enginelock` — 마커 확장 가능성. design에서 결정한다.
