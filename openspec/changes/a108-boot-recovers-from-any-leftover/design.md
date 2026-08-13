# a108 design — 부팅은 어떤 잔재에서도 회복한다

근거 산출물: `analysis/function-logic/` 7건 (start 15분기 · reclaimstalecontroldirectory
13분기 · server.close 3분기 · dial 2분기 · processalive 1분기 · runenginerun 19분기 ·
runhttpapi 21분기). 아래 분기 주장은 전부 이 AST 열거에서 나온다.

## D1. 회수는 자기 수명주기가 만드는 모든 상태를 다룬다

`Start`·`Close`·`reclaimStaleControlDirectory` 자신이 만들 수 있는 디스크 상태를 열거하면:

| 상태 | 만들어지는 경로 | 현재 동작 | 새 동작 |
| --- | --- | --- | --- |
| S0 빈 디렉터리 | `Start`가 Mkdir 후 listen 전에 죽음; reclaim이 두 파일 제거 후 rmdir 전에 죽음 | **영구 거부** (line 139) | 검증할 것이 없다 — dir 제거 후 진행 |
| S1 descriptor만 | **이번 사고.** graceful shutdown이 socket을 unlink(Go listener 기본), `Close`의 제거 루프 전에 프로세스 사망 | **영구 거부** (line 139) | socket이 없다 = listener가 없다 = endpoint는 죽었다. descriptor·dir 제거 후 진행 |
| S2 socket만 | `Close`/reclaim이 descriptor를 먼저 지우고(제거 순서: descriptor→socket→dir) socket 제거 전에 사망 | **영구 거부** (line 139) | connect probe: 수락하면 주인 생존 → 이번 시도 거부(다음 시도가 정리된 상태를 본다). `ECONNREFUSED`면 사망 → socket·dir 제거 후 진행 |
| S3 둘 다 | SIGKILL·전원 단절 (unlink 없이 사망) | PID 생존 검사 — **재사용된 PID가 영구 거부를 만들 수 있다** (D2) | connect probe로 판정. 소유권·perm·symlink 검사는 유지 |
| S4 낯선 엔트리 | 우리 수명주기 밖 (침입·오배치) | 거부 | **거부 유지** — 우리가 만들지 않은 것은 치우지 않는다 |

전체성이 핵심 성질이다: **위 표의 어떤 상태도 사람 손을 요구하지 않는다.** 예방(종료 순서
조정)은 채택하지 않는다 — 순서를 어떻게 바꿔도 "그 사이에서 죽는" 상태는 존재하고,
회수가 전체적이면 순서는 무관해진다.

보안 속성 보존: 기존 검사(디렉터리 0700·소유 uid·symlink 금지·socket perm 0600·nlink 1)는
회수 가능 상태에서도 전부 수행한다. 검사 실패는 지금처럼 거부다. 이 change는 회수의
**커버리지**를 넓히는 것이지 검증을 완화하는 것이 아니다.

## D2. 생존 판정은 PID가 아니라 socket connect다

현재 S3의 판정은 `processAlive(descriptor.PID)`(kill-0)다. a102 D4b-2가 실측한 대로
컨테이너 재생성 후 PID 배정은 근사-결정적이라, 재생성 뒤 descriptor의 PID 자리에 **무관한
프로세스**가 앉는다(이번 사고의 descriptor도 PID 16 — 새 컨테이너에서 거의 확실히 점유됨).
그 경우 kill-0은 성공하고 reclaim은 "주인이 살아 있다"며 **둘-다-있는 정상 잔재조차 영구
거부**한다. 사고 형태가 하나 더 숨어 있던 것이다.

connect probe가 더 강한 진실을 준다: **이 socket 경로에서 수락하는 자가 있는가.**

- 수락함 ⇒ 그 경로의 주인이 존재한다(그게 누구든 이 디렉터리를 치우면 안 된다) → 거부.
- `ECONNREFUSED` ⇒ listener 없음. `Start`는 listen이 성공한 뒤에만 descriptor를 쓰므로
  (AST start: listen → chmod → token → descriptor 순) 수락하지 않는 socket 파일은 죽은
  주인의 것이다 → 회수.

`processAlive`는 판정에서 제거한다(로그 참고값으로도 남기지 않는다 — 남기면 다음 독자가
다시 판정에 쓴다). a102가 enginelock에 쓴 proc_instance 토큰 방식은 여기선 불채택 —
connect probe가 같은 질문에 더 직접적으로 답하고 descriptor 스키마 변경이 필요 없다.

경합 창의 안전성: S1에서 주인이 아직 `Close` 도중일 수 있다(마이크로초 창). 그 경우 양쪽이
같은 경로를 제거하는데, `Close`는 `ErrNotExist`를 용인하므로(AST server.close) 경합은
양성이다. 문서화하고 수용한다.

## D3. 엔진은 projection 실패로 죽지 않는다

`runEngineRun`의 `strategyprojectionrpc.Start` 오류는 `return err`(exit 1)에서 **강등 +
critical 알림**으로 바뀐다: projection endpoint 없이 루프를 돌리고, 사유를 담은 critical을
발행한다(a074 알림 배관 사용).

이것이 안전한 이유 — 이중 writer 논증: 엔진 싱글턴은 부팅 **1단계**의 journal flock이
강제한다(`engine.go:14` "flock on the journal directory FIRST"). projection 디렉터리는
싱글턴 권위가 아니므로, projection 실패를 무시하고 진행해도 두 번째 엔진이 생길 수 없다.
projection은 `strategyprojection.Reader` 내보내기 전용이라(주입 지점
`strategy_runtime_projection.go:14`) 루프의 입력도 아니다 — 강등은 화면을 잃는 것이지
판정을 잃는 것이 아니다.

반대 방향(관측을 위해 보호를 세우는 것)은 이 제품이 a102에서 이미 지운 비대칭이다.
운영자 표면: 콘솔은 이미 "전략 화면은 dormant로 뜬다"를 갖고 있고(사고 당일 부팅 로그
실측), critical 알림이 "projection이 죽은 채 도는 엔진"을 조용히 두지 않는다.

**강등의 상한**: 알림은 기동 시 1회가 아니라 미해소 상태로 유지한다(운영자가 재기동으로
해소할 때까지 대시보드에 남는 형태 — 구현은 기존 alert 배관의 관행을 따른다).

## D4. httpapi는 부재와 실패를 같은 강등으로 처리한다

현재(AST runhttpapi): `os.Stat(descriptor)` 부재 → `strategyRuntime = nil`로 기동(설계된
강등), 존재+`Dial` 실패 → **fatal**. 반쪽 잔재·죽은 descriptor·깨진 socket 전부 후자로
떨어져 docker restart loop가 된다.

새 동작: dial 실패도 `strategyRuntime = nil` 강등 + 경고 로그. fatal로 남는 것은 descriptor
**조사 자체가 불가능한 경우**(권한 등 `os.Stat`의 비-NotExist 오류)뿐이다 — 그것은 잔재가
아니라 환경 이상이다. 재-dial(lazy 재연결)은 범위 밖 — 지금도 descriptor-부재로 기동한
httpapi는 엔진이 나중에 떠도 전략 없이 남으며, 그 동작과의 대칭을 유지한다(운영 절차는
엔진 기동 후 httpapi 재시작 — docker restart가 이미 그 역할을 한다).

## D5. 배포가 곧 복구이고, 검증은 사고 상태의 재현이다

겹1 바이너리는 이번 사고의 디스크 상태(S1)를 부팅에서 스스로 회수한다. RED 테스트는
tmpdir에 사고 상태 네 가지(S0~S3-사망)를 그대로 만들어 `Start`가 성공함을, S2/S3-생존과
S4에서 거부함을 고정한다. **PID 재사용 회귀 핀**: descriptor PID를 테스트 자신의 PID(확실히
생존)로 쓰고 socket은 죽은 상태로 두면 — 옛 코드는 거부하고 새 코드는 회수한다. 이 핀이
D2 결정을 뮤테이션으로 증명한다(`processAlive`로 되돌리면 실패).

네 endpoint 공통 crash-shape 핀: strategy projection·policy command·policy runtime·alert
control 각각에 대해 "부분 잔재 후 기동 성공"을 고정한다. 나머지 셋은 오늘 실전을
통과했지만(관용 설계) 그것은 관측이지 핀이 아니다 — [[passing-test-is-not-evidence]]의
"재현은 고정이 아니다"를 반복하지 않는다.

## 범위 밖 (선언된 생략)

- 종료 순서 예방(D1에 근거), 콘솔(무변경 — 실측 강등 확인), httpapi lazy 재-dial(D4에 근거).
- 호스트 재부팅 원인 조사(운영 항목), performance.db 부재 경고(별개 기존 사안).
- soak autostart는 a101이 이미 해결 — 이번 재부팅 후 서베이는 스스로 살아났다(로그 실측
  "soak 자동 시작: … 새로 시작했다"). a101의 5.6 검증이 이 사고로 완료된 셈이다.

## 리뷰 개정 (A1·A2 적대 리뷰, 2026-08-14)

두 리뷰 모두 FIX-FIRST였고, 각자 뮤테이션 2건을 독립 재현해 원장의 정직성은 확증했다.
아래 개정이 Fix 라운드(tasks §6)의 계약이다.

### D1-2. 발행도 회수만큼 전체적이어야 한다 — stage+rename 의례 (A1 P1×2)

A1이 실측한 잔여 영구-거부 2종은 둘 다 **생산자**가 만든다:

- pre-chmod socket: `net.Listen` → `Chmod` 사이의 죽음이 umask 077(컨테이너 실측)에서
  0700 socket을 남기고, 회수의 정확-0600 검사가 그것을 "unsafe"로 영구 거부한다.
  **버그가 보안 핀으로 봉인돼 있었다**(M7 해석 정정 대상).
- 부분 descriptor: `O_EXCL`→write→sync는 원자적이지 않고, 내용 파싱 실패는 영구 거부다.
  D2 이후 descriptor 내용은 **어떤 판정에도 쓰이지 않으므로** 이 게이트는 거부만 생산한다.

수정: **두 산출물 모두 stage+rename으로 발행한다** — 같은 저장소의 다른 세 endpoint가
이미 쓰는 의례다. socket은 control dir 안 숨김 임시 이름에 bind → chmod 0600 → rename
(Linux에서 unix socket의 rename은 유효하고 bind는 inode에 붙는다). 파생 규칙:

- 임시 이름(.staging 접두) 잔재는 낯선 엔트리가 아니라 자기 잔재다 — 회수 목록에 추가.
- 검증-사망 잔재의 socket perm 검사는 정확-0600에서 **group/other 비트 없음
  (perm&0o077==0)**으로 좁게 완화한다(자기 uid·0700 dir·비symlink·nlink1·사망 입증
  하에서만) — 구버전 바이너리가 남긴 pre-chmod 잔재의 회수 경로.
- descriptor **내용** 파싱 실패는 사망 입증 시 회수로 읽는다. **형식** 검사(0600·정규
  파일·no-follow)는 유지한다.

### D2-2. 경로를 unlink할 권한은 우리 손에만 (A1 P2)

A1이 300라운드 중 7라운드 안 3회로 재현: `Shutdown`이 `Serve`의 listener 등록을
앞지르면 unlink가 Serve goroutine의 defer로 밀리고, **경로 기준** unlink가 후계자의
새 socket을 지운다. 오늘 도달을 막는 것은 flock이 아니라 "예정된 goroutine이
프로세스와 함께 죽는다"는 사실뿐이다.

수정: `SetUnlinkOnClose(false)` + `Close`가 `s.listener`를 명시적으로 닫고 최종 경로
제거는 `Close`의 `os.Remove`만 수행한다. stage+rename(D1-2)과 결합하면 listener가
기억하는 경로는 임시 이름이라 어느 쪽 unlink도 최종 경로를 건드리지 못한다.
**겹2는 강등 후 in-process projection 재시도를 하지 않는다**(명문) — 재시도는 이 창을
같은 프로세스 안에서 다시 연다. probe와 remove 사이 새-주인 경합의 실제 방어는
journal flock임을 여기 명시한다(코드 주석도 flock을 인용할 것).

### D3-2. outbox critical 철회 — 강등 보고는 게이트에 연결되지 않는다 (A2 P1×2, 결정 반전)

원 D3의 "미해소 critical 알림"을 **철회한다**. A2가 실측·추적으로 보인 두 사실:

1. `UndeliveredCount`는 Type 무필터라(`outbox.go:533`), publisher 미설정 배포에서 강등
   행이 영원히 PENDING으로 남아 **다음 부팅의 entry gate를 잠근다**
   (`restoreAlertEntryLatch` — 해제는 운영자 ack뿐). obs 교리와 정면 충돌: "측정 실패는
   critical이 아니다 … 화면의 오탈자가 실계좌 매매를 멈출 수 있어서는 안 된다"
   (`obs/event.go:266-278`). T2가 인용한 execgw 선례는 등급표 등재 + **의도적** entry.Block
   이라 선례가 아니다(오독 정정).
2. transport가 살아 있으면 행은 ~2초 뒤 DELIVERED로 사라진다 — "미해소 유지"는 애초
   미구현이었고, 그것을 확인한 테스트는 전달 루프 없는 harness에서 PENDING을 본 것이다.

새 계약: 강등 보고는 ① stderr 기동 경고, ② **obs Normal 이벤트 로그**(등급표 미등재가
의도다 — Notifier로 흘러도 best-effort Normal), ③ 콘솔 전략 화면의 기존 dormant 표면.
**강등 기동은 outbox 행을 만들지 않고 다음 부팅의 진입 상태를 바꾸지 않는다** — 둘 다
핀으로 고정한다. proc-instance dedup 토큰 기계는 outbox 제거와 함께 불필요해진다.

### D4-2. httpapi는 projection 때문에 죽지 않는다, 마침표 (A2 P2×2)

- `Dial`이 connect probe로 생존을 확인한다(회수와 같은 원시) — S3(socket 파일이 남은
  사망, **전원 단절의 기본 모양**)에서 lazy client 대신 즉시 오류 → 소비자 강등.
  A2 실측: 지금은 Dial이 성공하고 첫 Read가 죽는다.
- httpapi reader는 strategy Read 실패를 dormant로 흡수한다 — 지금은 집계 스냅샷
  전체(engine·positions·orders)가 함께 죽는다(`httpapi_reader.go:479-483`).
- 비-NotExist stat 오류도 콘솔 패리티로 경고+강등(원 D4의 fatal 결정 철회) — ENOTDIR
  볼륨 오배치가 다시 crash loop를 만든다는 A2 반증을 수용.

### D5-2. 범위 재단 — 형제 endpoint 셋은 a109로 (A1 P1)

A1 실측: policy runtime·alert control도 pre-chmod socket 잔재를 영구 거부하고
(`engine.go:272/277`는 여전히 fatal), alert control은 **살아 있는 주인의 socket 위에
두 번째 서버가 올라선다**. tasks 2.5의 핀은 관용이 자명한 모양만 만들어 실패할 수
없었다(정지 조건 무력화 — 기록). 이것은 a108과 같은 병이지만 표면이 세 배다:
**engine-safety delta의 일반 요구를 strategy projection으로 좁히고, 일반 불변식은
a109 등록으로 옮긴다.** 거짓 SHALL을 정본에 넣지 않기 위한 좁힘이며(A1 P1-3),
좁힘의 근거와 측정은 a109 proposal이 승계한다.

### D5-3. 롤백은 a108을 거꾸로 가로지른다 (gstack 3라운드 적대 발견)

새 바이너리가 **합법적으로** 남기는 잔재(.staging-*, S0/S1/S2)를 구 바이너리의 회수는
영구 거부한다. 롤백은 새 바이너리가 방금 죽었을 때 하는 것이므로, 잔재가 있을 확률이
가장 높은 순간에 구 바이너리가 2026-08-13 사고를 재생산한다. 코드로 못 고친다 — 절차
한 줄이다: **a108을 가로질러 롤백하기 전에 `.strategy-runtime-read/`를 지운다**(전부
일회용 산출물이다). tasks 5.2에 반영, a109 tasks에도 승계.
