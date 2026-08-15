# a109 설계 — 형제 endpoint들도 회복한다

작성 2026-08-15. base `016da624`. 모든 분기 주장은 `analysis/function-logic/`의 AST 산출물
13개(발행 2 + 회수 5 + Close 3 + 기동 1 + 소비자 1 + command 발행 1)를 먼저 만든 뒤 그
열거를 근거로 썼다(FLM-before-claiming). a108의 확정 계약(transport_unix.go, D1-2·D2-2·
D3-2·D4-2)이 이식 원형이다.

## 병의 재확인 (AST 열거 기준)

| endpoint | 발행 | 회수 | 산-주인 | fatal 지점 |
| --- | --- | --- | --- | --- |
| position policy command (loopback TCP) | descriptor stage+rename ✓ | 없음 — staging 잔재 무한 축적 | 해당 없음(TCP ephemeral port) | engine.go:272 `return err` |
| position policy runtime (unix socket) | socket **최종 경로 bind→chmod** (`StartPositionPolicyRuntimeServer` :72–85), descriptor stage+rename ✓ | `PrepareRuntimeSocket` 정확-0600만 (runtime_unix.go:62) | probe 없이 무조건 unlink (:68) → 탈취 | engine.go:277 `return err` |
| alert control (unix socket) | socket **최종 경로 bind→chmod** (`StartAlertControlServer` :99–112), descriptor stage+rename ✓ | `PreparePrivateSocket`→`statPrivateSocket` 정확-0600만 (private_endpoint.go:90–92) | probe 없이 무조건 unlink (:75) → 탈취 (A1 F4 실측) | engine.go:313 `return err` |

- pre-chmod 잔재: `net.Listen`이 umask(컨테이너 실측 077)로 만든 0700 socket이
  listen→chmod 사이의 죽음으로 최종 이름에 남으면, 두 socket endpoint의 회수가 이를
  "not an exact 0600 Unix socket"으로 **매 부팅 영구 거부**한다(A1 F3 실측 3/3). a108
  이전의 strategy projection과 동일 기전이다.
- 산-주인 탈취: 두 socket 회수 함수 모두 **생존 probe 없이** 유효한 0600 socket을
  unlink한다. 살아 있는 주인의 socket 위에 두 번째 서버가 올라선다. proposal은 alert
  control만 실측을 인용하지만 AST상 두 endpoint가 같은 기전을 공유한다.
- descriptor에는 병이 없다: 세 endpoint 모두 이미 stage+rename이고, 기동은 이전
  descriptor의 내용을 검증하지 않는다(rename이 덮는다). 0바이트·잘린 descriptor 잔재는
  현재도 관용된다 — 이 관용은 핀으로 고정한다(D5).
- staging 잔재 축적: `os.CreateTemp` 임시 파일(`.position-policy-control-*`,
  `.position-policy-runtime-*`, `.endpoint-*`)은 crash 시 아무도 치우지 않고, 어떤
  검증도 디렉터리 내용을 열거하지 않으므로 조용히 쌓인다.

## D1 — socket 발행은 a108의 의례를 쓴다: 임시 이름 bind → 0600 → rename

두 socket endpoint의 발행을 `listenPrivateSocket`(strategyprojectionrpc) 의례로 바꾼다:
staging 이름에 bind → `SetUnlinkOnClose(false)` → chmod 0600 → 최종 이름으로 rename.
최종 이름은 완성된 socket에만 붙는다 — pre-chmod 상태가 최종 이름을 가지는 순간이
사라진다(원천 소거).

**D1a — staging basename은 11자다.** a108의 `.s-`+종류1+hex8=12자는 `runtime.sock`(12자)
이하지만 **`alerts.sock`은 11자**라 그대로 쓰면 "staging ≤ 최종" 계약이 깨진다 — 최종
경로는 bind 가능한데 staging만 sun_path 상한(이식 가능 경계 103, a108 Codex 재판정)을
넘는 배포가 생긴다. 형제 공용 staging은 `.s-`+종류1+hex7=**11자**로 고정하고, staging
basename ≤ 각 endpoint 최종 socket basename을 테스트로 고정한다(D5).
strategyprojectionrpc의 12자는 a108 소유이므로 건드리지 않는다.

**D1b — 공용 기계는 `internal/positionpolicyrpc`에 둔다.** private_endpoint.go의 자기
문서가 이유다: 보안 원시(symlink 거부·소유·hard-link 검증·no-follow open)는 이 패키지의
비공개 함수이고 "복사한 검사는 어긋나기 시작한 검사다"(a098 D7.1). 이름-독립 형태의
staged listen·회수·probe를 이 패키지에 추가하고 두 engine transport가 함께 쓴다.
strategyprojectionrpc를 이 기계로 접는 것(descriptor 발행 3벌 fold 포함)은 **하지
않는다** — 동작하는 a108 확정 코드의 무근거 재작성이다. 선언된 생략으로 남긴다.

## D2 — 회수의 전체성: 자기 수명주기가 만드는 모든 부분 상태 + 사망 증명

두 socket endpoint의 control 디렉터리 회수를 a108 `reclaimStaleControlDirectory` 모형으로
바꾼다:

- **열거**: 디렉터리 엔트리를 전부 읽고, 아는 이름만 회수한다 — 최종 2개(descriptor·
  socket), 신규 staging(`.s-*`), 그리고 **legacy staging**(`.position-policy-control-*`,
  `.position-policy-runtime-*`, `.endpoint-*` — 구버전 crash가 이미 남겼을 수 있는 우리
  잔재. 정규 파일 모양만). 그 밖의 낯선 엔트리는 아무것도 건드리지 않고 거부한다.
- **모양 검증**: 잔재 socket은 a108의 좁은 완화를 그대로 — `perm&0o077 == 0`(group/other
  비트 없음)이면 pre-chmod 0700도 회수 대상. 소유 uid·비symlink·nlink 1은 그대로 요구.
- **사망 증명**: 제거 전 connect probe(`projectionSocketAccepts` 의례: ECONNREFUSED·
  ENOENT·owner-write-비트-없음 = 사망, 그 외 = 생존으로 간주하고 거부). PID는 판정에
  쓰지 않는다(a102 D4b-2 — 컨테이너 재생성은 PID를 재배정한다). **수락 중인 socket은
  절대 unlink하지 않는다** — 산-주인 탈취가 거부로 바뀐다.
- **1차 방어는 여전히 journal flock이다**: `runEngineRun`은 endpoint 기동 전에 flock을
  쥔다(부팅 1단계). probe는 flock이 막지 못하는 것(엔진 아닌 프로세스의 경로 점유,
  다른 journal 디렉터리의 엔진)에 대한 심층 방어이고, 코드 주석이 flock을 근거로
  명시 인용한다 — proposal의 "flock 명문화" 요구를 여기서 충족한다.

**D2a — command endpoint(TCP)의 회수는 staging 위생뿐이다.** socket이 없으므로 probe할
대상도, pre-chmod 병도 없다. descriptor는 rename이 덮는다. 이 endpoint의 회수는 자기
staging 잔재(legacy 포함) 제거로 한정한다 — 없는 병을 고치지 않는다(YAGNI).

**D2b — Close는 a108 계약을 따른다**: listener를 Close가 **직접** 닫고(late-unlink 경합
A1 F5 차단 — `AlertControlServer`는 현재 listener 필드 자체가 없다: 추가한다),
`SetUnlinkOnClose(false)`이므로 경로 제거는 Close의 제거 루프 하나뿐이다. 제거 순서
(descriptor→socket→dir)와 ErrNotExist 관용은 현행 유지.

## D3 — 세 fatal은 전부 강등한다 (proposal 0.3 판정)

`cmd/tossctl/engine.go`의 세 `return err`(:272/:277/:313)를 a108 D3의 강등 의례로 바꾼다.
endpoint별 판정 근거:

| endpoint | 잃는 것 | 보수성 논증 |
| --- | --- | --- |
| policy runtime | 콘솔·읽기 화면 | a108 D3 원문 그대로 — 조회 전용, 잃는 것은 화면이지 판정이 아니다 |
| alert control | 운영자 ack 표면 | ack 불가 → 미전달 critical의 entry latch **유지** → 신규 진입 차단 지속 = 더 보수적. `AlertOperations`는 gateway에 닿지 않는다(transport 자기 문서 :24–28) — 손절·청산 경로 무관 |
| policy command | Preview/Apply·격리 해제 표면 | 표면 부재 = 아무것도 느슨해지지 않는다. 격리는 유지되고 정책은 변하지 않는다. 회복 수단은 사람의 엔진 재시작 |

공통 근거: ① 엔진 싱글턴은 journal flock이 강제하고 endpoint는 루프의 입력이 아니라
표면이다(engine.go a108 주석 :288). ② fatal 유지가 만드는 것은 사고 모양 그대로다 —
잔재·이물 → autostart 영구 기동 루프 → **장중 손절 없음**. ③ 세 소비자
(console.go·engine_alerts_client_unix.go·position_policy_commander.go)는 이미 "descriptor
없음 = 엔진이 없다"를 다루므로 강등의 소비자측 화면 값은 기존 부재 경로다.

**D3a — 보고는 a108 D3-2 교리를 그대로 일반화한다.** `reportStrategyProjectionDegraded`를
endpoint 이름을 받는 형태로 일반화(stderr 안내 + **의도적으로 등급표 미등재** obs Normal
이벤트 + 콘솔 휴면 표면). 금지 3종 불변: obs criticalEvents 등재 금지·severity critical
금지·원장 outbox 금지(미전달 PENDING 행은 다음 부팅의 진입 게이트를 잠근다 —
`engineStrategyProjectionDegradedEvent` 주석 :330–359가 정본). Notify는
`context.WithoutCancel` + goroutine(부팅 비차단, a108 T3-fix).

**D3b — 강등해도 회수·발행 실패의 원인은 지운다.** 강등은 마지막 줄이지 수리가 아니다.
D1·D2가 자기 잔재 병을 원천 소거하므로, 강등이 실제로 발동하는 것은 이물·환경 이상뿐
이어야 한다. 그 상태로 부팅을 거듭 나는 것은 D3a의 보고가 매 부팅 보이게 한다.

## D4 — httpapi는 엔진이 돌아오면 다시 붙는다 (2b.1)

감독 순서 고정(compose 의존)이 아니라 **소비자측 lazy 재-dial**을 택한다. 근거: ①
감독 순서는 "가동 중 엔진 재시작 후 복귀"를 표현할 수 없다 — 재-dial은 냉부팅 순서와
가동 중 재시작을 한 기전으로 덮는다. ② 배포 파일이 아니라 코드가 계약을 진다.
③ 사람이 누르는 seam의 재사용이 아니다(boot-is-not-a-button).

설계: `strategyRuntimeReaderFor`의 판정(부재=nil / 잔재=sentinel)을 유지하되, nil·
sentinel 상태의 reader를 **재부착 가능한 wrapper**로 감싼다 — Snapshot 경로에서 rate
limit(최소 간격, 30s 기본) 아래 stat+Dial **1회**를 시도하고, 성공하면 live client로
원자 교체(single-flight, 실패해도 요청은 기존 강등 값으로 즉시 응답 — 요청 경로를
막지 않는다). 엔진측 in-process 재시도 금지(a108 D2-2)는 그대로다 — 이것은 소비자다.

## D5 — 사고급 핀 (a108 2.5의 실패-불가 핀 교체)

a108 2.5 핀(`a108_every_endpoint_survives_a_leftover_test.go`)은 현재 코드가 이미
관용하는 모양만 깔아 **실패할 수 없었다**(a108 review §1 F4). a109가 사고급 모양을 깐다
— RED에서 실제로 실패해야 하는 것: ① pre-chmod 0700 socket 잔재에서 기동(두 socket
endpoint), ② 수락 중인 socket 위 두 번째 기동 거부(탈취 금지), ③ staging 잔재
(legacy·신규)에서 기동 + 회수 후 잔재 0, ④ 낯선 엔트리 거부. 이미 관용되는 모양
(0바이트 descriptor 등)은 관용을 핀으로 고정한다. staging basename ≤ 최종 basename
경계(103 이식 경계 포함)를 상수 테스트로 고정한다. 기존 a108 2.5 핀은 유지한다 —
관용의 핀이며 여전히 참이다.

**D5a — defer 순서 핀 (2b.2).** runEngineRun의 불변식 "endpoint Close는 journal flock을
쥔 채로 돈다"는 지금 defer 등록 순서 하나에 얹혀 있고 테스트가 없다. go/parser 기반
정적 테스트로 고정한다: runEngineRun 본문에서 `lock.Release()` defer가 모든 endpoint
`Close()` defer보다 **먼저 등록**된다(LIFO상 가장 나중 실행).

**D5b — 롤백 절차 일반화 (2b.3)는 측정으로 쓴다.** a108 D5-3의 일반화가 필요한지부터
측정한다: 구버전(=현재 main) 회수는 디렉터리를 열거하지 않으므로 신버전의 `.s-*` 잔재를
무시할 가능성이 높다 — 그러면 a109를 가로지르는 롤백에 사전 wipe가 필요 없다는 것이
결론이고, 그 **측정 결과**를 배포 절차 prose에 기록한다. 주장 먼저 쓰고 측정을
생략하는 것을 금지한다(generated-evidence-must-be-measured).

## 작업 분할 (파일 표면 비겹침)

| | 표면 | 내용 |
| --- | --- | --- |
| T1 | `internal/positionpolicyrpc/*`, `internal/app/engine/{position_policy_runtime_transport_unix.go, alert_control_transport_unix.go, position_policy_transport.go}` + 그 패키지의 a109 테스트 | D1·D2·D5(①~④·경계) + 뮤테이션 원장 |
| T2 | `cmd/tossctl/{engine.go, httpapi.go}` + cmd 패키지 a109 테스트 + 배포 절차 prose | D3·D4·D5a·D5b + 뮤테이션 원장 |

T2는 세 Start의 실패를 시뮬레이트하는 seam(cli_testseams.go 관례)을 도입해 T1의 완료를
기다리지 않는다. 각 Teammate는 자기 표면의 FLM 마크다운 맵(scaffold 완료, AST 존재)을
**편집 전에** 완성하고 구현 후 재최신화한다. High-risk(엔진 기동 경로) — Pre-Edit 선언
필수.

## 선언된 생략 (not-applicable)

- descriptor 발행 3벌의 fold(publishPrivateDescriptor 주석의 초대): 병이 없는 표면의
  High-risk 리팩터링 — 하지 않는다.
- strategyprojectionrpc의 staging 12자→11자 통일: a108 소유, `runtime.sock` 12자에
  적법 — 건드리지 않는다.
- 콘솔의 형제 endpoint 화면 강화: 소비자는 이미 부재를 다룬다. 화면 신설은 scope 밖.
