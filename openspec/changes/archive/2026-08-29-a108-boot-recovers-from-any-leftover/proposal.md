# a108 — 부팅은 어떤 잔재에서도 회복한다

## Why

**2026-08-13 23:35 KST 사고.**

US 정규장 중 호스트가 재부팅됐고(`uptime -s` 2026-08-13 23:35:19), 그 뒤 **엔진이 영구
기동 루프에 빠졌다.** 매 시도가 같은 줄에서 죽는다:

```text
Error: strategy projection runtime: stale endpoint is incomplete
```

httpapi 컨테이너도 같은 원인으로 crash loop에 들어갔다(`Restarting (1)` 반복). 보호
루프(reconcile·exit·filldetect)는 14:33Z부터 전부 정지 — **관리 포지션(IONQ·TSLA 포함)의
손절 감시가 US 장중에 사라졌다.** a056의 8분, a102의 02:03과 같은 계열이지만 이번에는
백오프 문제가 아니라 **어떤 재시도로도 벗어날 수 없는 디스크 상태**다.

### 인과 사슬 (전부 실측)

1. 종료 시퀀스에서 엔진이 SIGTERM을 받아 strategy projection 서버가 `http.Server.Shutdown`
   → Go의 unix listener는 close 시 socket 파일을 **자동 unlink**한다(`net.UnixListener`
   unlink-on-close 기본값). `runtime.sock`이 사라진다.
2. `(*Server).Close`의 잔재 제거 루프(descriptor → socket → dir 순, AST `server.close`
   분기 3)가 끝나기 전에 호스트 종료가 프로세스를 죽였다. bind mount 위에
   **`.strategy-runtime-read/` 디렉터리와 `endpoint.json`만 남았다**(실측: dir mtime
   14:33Z, 엔트리 1개).
3. 재부팅 후 `Start`(AST 분기 15)는 `os.Mkdir` `ErrExist` →
   `reclaimStaleControlDirectory`(AST 분기 13)로 회수를 시도하는데, 그 함수는
   descriptor와 socket이 **둘 다 있어야만** 회수한다
   (`transport_unix.go:139` — `"stale endpoint is incomplete"`). 하나라도 없으면 거부.
4. 거부 → `runEngineRun`이 `strategyprojectionrpc.Start` 오류를 그대로 반환
   (`cmd/tossctl/engine.go:282-286`, AST `runenginerun` 분기 19) → **엔진 exit 1.**
   콘솔 autostart가 ~1분마다 재시도하지만 디스크 상태는 그대로이므로 영구 루프.
5. httpapi는 descriptor가 **없으면** 우아하게 강등하지만(전략 없이 기동), descriptor가
   **있는데 dial이 실패하면** fatal이다(`cmd/tossctl/httpapi.go:146-153`, AST `runhttpapi`
   분기 21). 반쪽 잔재는 정확히 후자를 때린다 → docker restart crash loop.

즉시 복구는 잔재 디렉터리 수동 이동(운영자 실행, 증거는 `/tmp`에 보존)이고, 이 change는
**그 수동 개입이 다시는 필요 없게** 만든다.

## 이것이 a102와 같은 비대칭인 이유

strategy projection은 `strategyprojection.Reader`를 내보내는 **조회 전용 export 표면**이다
(`internal/app/engine/strategy_runtime_projection.go:14`). 콘솔·httpapi가 화면을 그리려고
읽는 socket 하나 때문에 **손절을 든 프로세스 전체가 기동을 거부한다.** a102가 지운
비대칭("보호를 든 쪽이 먼저 포기한다")이 관측 표면에서 재발한 것이다.

같은 부팅 순서의 다른 세 endpoint는 이미 관용적이다 — position policy command(TCP,
descriptor만)·position policy runtime·alert control은 `ErrExist`를 허용하고
`PreparePrivateSocket`이 낡은 socket을 치운다. **오늘 재부팅 잔재를 실제로 통과했다**
(엔진이 그 단계들을 지나 :282에서 죽었다). 거부-영원 회수는 strategyprojectionrpc 하나뿐이다.

## What Changes

세 겹이다.

- **겹1 (회수의 전체성)**: `reclaimStaleControlDirectory`가 자기 수명주기가 만들 수 있는
  **모든** 부분 상태(descriptor만/socket만/빈 디렉터리/둘 다)를 검증 후 회수한다. 생존
  판정은 PID가 아니라 **socket connect probe**로 한다 — 컨테이너 재생성 후 PID 재사용은
  a102 D4b-2가 실측한 실재 위험이고, 현재 코드의 `processAlive(descriptor.PID)`는 남의
  프로세스를 "주인이 살아 있다"로 오판해 **둘-다-있는 상태조차** 영구 거부로 만들 수 있다.
  소유권·symlink·권한 검사(보안 속성)는 전부 유지한다 — 검증 불가능한 낯선 상태는
  지금처럼 거부한다.
- **겹2 (엔진은 죽지 않는다)**: `runEngineRun`에서 projection `Start` 실패는 fatal이 아니라
  **강등 + critical 알림**이다. 싱글턴 권위는 부팅 1단계의 journal flock이지
  (`engine.go:14`) projection 디렉터리가 아니므로, 강등이 이중 writer를 만들 수 없다.
- **겹3 (httpapi 대칭 강등)**: descriptor-부재와 dial-실패를 같은 강등(전략 화면 없이
  기동)으로 처리한다. 지금의 비대칭이 반쪽 잔재를 crash loop로 바꾼다.

**배포가 곧 복구다**: 겹1을 실은 바이너리는 오늘 밤의 잔재 상태를 부팅 시 스스로 회수한다.

## Non-goals

- `SetUnlinkOnClose(false)` 등 종료 순서 예방 조정 — 회수의 전체성이 순서 민감성을
  제거하므로 불필요(선언된 생략). 예방을 더해도 회수가 전체적이지 않으면 미지의 순서가
  같은 사고를 만든다.
- 콘솔 변경 — 이미 강등한다(사고 당일 부팅 로그 실측: "전략 화면은 dormant로 뜬다").
- 다른 세 endpoint의 재작성 — 관용 설계가 이미 있고 오늘 실전을 통과했다. 단, 네 endpoint
  전부에 crash-shape 회귀 핀을 깔고, 핀이 결함을 드러내면 같은 패턴으로 고친다(tasks 3.4).
- 호스트가 왜 재부팅했는가 — 이 change 밖이다(운영 조사 항목).

## Impact

- Affected specs: `engine-safety` (ADDED 2), `http-api-service` (ADDED 1)
- Affected code: `internal/strategyprojectionrpc`(겹1), `cmd/tossctl/engine.go`(겹2),
  `cmd/tossctl/httpapi.go`(겹3)
- High-risk: 엔진 기동 경로 편집 — FLM 7건 생성 완료(`analysis/function-logic/`),
  Pre-Edit 선언은 각 task 착수 시 `pre-edit-gate.md`에 남긴다.
