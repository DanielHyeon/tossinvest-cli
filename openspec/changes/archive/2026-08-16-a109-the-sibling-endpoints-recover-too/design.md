# a109 설계 — 형제 endpoint들도 회복한다

작성 2026-08-15, freeze 판결 반영 개정(같은 날 — review.md §0). base `016da624`. 모든
분기 주장은 `analysis/function-logic/`의 AST 산출물 **20개**(freeze 적대 리뷰 P1-6이
잡은 descriptor 발행 3·보고 1·디렉터리 검증 3 추가분 포함)를 먼저 만든 뒤 그 열거를
근거로 썼다(FLM-before-claiming). a108의 확정 계약(transport_unix.go, D1-2·D2-2·D3-2·
D4-2)이 이식 원형이다.

## 병의 재확인 (AST 열거 기준)

| endpoint | 발행 | 회수 | 산-주인 | fatal (`return err`) |
| --- | --- | --- | --- | --- |
| position policy command (loopback TCP) | descriptor stage+rename ✓ | 없음 — staging 잔재 무한 축적 | 해당 없음(TCP ephemeral port) | engine.go:274 |
| position policy runtime (unix socket) | socket **최종 경로 bind→chmod** (`StartPositionPolicyRuntimeServer` :72–85), descriptor stage+rename ✓ | `PrepareRuntimeSocket` 정확-0600만 (runtime_unix.go:62) | probe 없이 무조건 unlink (:68) → 탈취 | engine.go:279 |
| alert control (unix socket) | socket **최종 경로 bind→chmod** (`StartAlertControlServer` :99–112), descriptor stage+rename ✓ | `PreparePrivateSocket`→`statPrivateSocket` 정확-0600만 (private_endpoint.go:90–92) | probe 없이 무조건 unlink (:75) → 탈취 (A1 F4 실측) | engine.go:315 |

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
  검증도 디렉터리 내용을 열거하지 않으므로 조용히 쌓인다. **이 이름들은 legacy가
  아니라 현행이다**(freeze P1-1) — descriptor fold는 선언된 생략이므로 a109 바이너리도
  같은 이름을 계속 만든다. 회수·핀은 "현행·구버전 공통 CreateTemp staging"으로 다룬다.
- 회수 밖에 남는 상태(freeze P1-7 열거, 전부 D3 강등+보고가 받는다): ① control 디렉터리
  자체의 perm 변형(owner 비트를 깎는 umask 등) — 환경 이상으로 분류, 회수·완화하지
  않는다(a108의 정확-0700 디렉터리 검사와 일관). ② descriptor 최종 이름 자리의
  디렉터리·이물(rename EISDIR) — 이물 간섭, 회수하지 않는다. ③ Close의 `sync.Once`
  부분 실패 잔재 — 재시도하지 않으며 다음 부팅의 회수가 덮는다. ④ alert control 판
  `publishPrivateDescriptor`는 rename 직전 디렉터리 재검증이 없는 기존 비대칭(P2-6) —
  descriptor fold와 함께 선언된 생략.

## D1 — socket 발행은 a108의 의례를 쓴다: 임시 이름 bind → 0600 → rename

두 socket endpoint의 발행을 `listenPrivateSocket`(strategyprojectionrpc) 의례로 바꾼다:
staging 이름에 bind → `SetUnlinkOnClose(false)` → chmod 0600 → 최종 이름으로 rename.
최종 이름은 완성된 socket에만 붙는다 — pre-chmod 상태가 최종 이름을 가지는 순간이
사라진다(원천 소거).

**D1a — staging basename은 11자다.** a108의 `.s-`+종류1+hex8=12자는 `runtime.sock`(12자)
이하지만 **`alerts.sock`은 11자**라 그대로 쓰면 "staging ≤ 최종" 계약이 깨진다 — 최종
경로는 bind 가능한데 staging만 sun_path 상한을 넘는 배포가 생긴다(staging과 최종은 같은
디렉터리이므로 basename 비교가 곧 전체 경로 비교다 — freeze 생존 판정 1). 형제 공용
staging은 `.s-`+종류1+hex7=**11자**로 고정한다. **hex7은 `hex.EncodeToString`으로 바로
나오지 않는다**(짝수 길이만 낸다, freeze P1-8 측정) — 4바이트 난수의 hex를 `[:7]`로
절단한다(엔트로피 28비트; 충돌은 회수가 방금 비운 디렉터리라 무해 — a108 `stagingPath`
주석과 동일 논거). 상수 테스트는 staging basename **길이 11을 직접 세고**, 각 endpoint
최종 socket basename과의 ≤ 관계식을 고정한다(D5). Linux 실측 bind 상한은 107자다
(freeze P2-1 측정) — 103은 이식성 각주로만 두고 절대 경로 103 요구를 테스트에 넣지
않는다. strategyprojectionrpc의 12자는 a108 소유이므로 건드리지 않는다.

**D1b — 공용 기계는 `internal/positionpolicyrpc`에 둔다.** private_endpoint.go의 자기
문서가 이유다: 보안 원시(symlink 거부·소유·hard-link 검증·no-follow open)는 이 패키지의
비공개 함수이고 "복사한 검사는 어긋나기 시작한 검사다"(a098 D7.1). 이름-독립 형태의
staged listen·회수·probe를 이 패키지에 추가하고 두 engine transport가 함께 쓴다.
이름-독립이므로 a108과 달리 **아는-이름 집합을 호출자가 넘긴다**(최종 descriptor·socket
이름 + 현행 CreateTemp 접두 + 신규 `.s-` 접두). 하나라도 빠지면 낯선-엔트리 거부가 우리
자신의 잔재를 거부하므로, **집합의 완전성을 테스트가 고정한다**: 각 transport의 발행
경로가 실제로 만드는 모든 이름이 그 transport가 넘기는 집합에 속함을 검사한다(freeze
P2-5). strategyprojectionrpc를 이 기계로 접는 것(descriptor 발행 3벌 fold 포함)은
**하지 않는다** — 동작하는 a108 확정 코드의 무근거 재작성이다. 선언된 생략으로 남긴다.

## D2 — 회수의 전체성: 자기 수명주기가 만드는 모든 부분 상태 + 사망 증명

두 socket endpoint의 control 디렉터리 회수를 a108 `reclaimStaleControlDirectory` 모형으로
바꾼다:

- **열거**: 디렉터리 엔트리를 전부 읽고, 아는 이름만 회수한다 — 최종 2개(descriptor·
  socket), 신규 staging(`.s-*`), 그리고 **현행·구버전 공통 CreateTemp staging**
  (`.position-policy-control-*`, `.position-policy-runtime-*`, `.endpoint-*` — descriptor
  fold를 생략했으므로 a109 바이너리 자신의 crash도 계속 남기는 우리 잔재다, P1-1).
  staging 엔트리는 모양(정규 파일 또는 socket)에 더해 **소유 uid도 검증한다** — a108은
  자기 접두 하나만 다뤘지만 a109는 일반적 접두(`.endpoint-*`)까지 넓히므로 검사 부재의
  폭을 함께 넓히지 않는다(P1-7③). 그 밖의 낯선 엔트리는 아무것도 건드리지 않고
  거부한다.
- **모양 검증**: 잔재 socket은 a108의 좁은 완화를 그대로 — `perm&0o077 == 0`(group/other
  비트 없음)이면 pre-chmod 0700도 회수 대상. 소유 uid·비symlink·nlink 1은 그대로 요구.
  **완화는 회수 전용 함수에만 존재한다** — 발행 후 최종 이름은 항상 0600이므로
  클라이언트 검증(`ValidatePrivateSocket`/`ValidateRuntimeSocket`)은 정확-0600을
  유지한다. 공유 helper(`statPrivateSocket`)의 조건을 완화하는 손쉬운 구현은 클라이언트
  경계를 넓힌다 — 뮤테이션 원장이 "클라이언트 검증을 완화하면 테스트가 죽는가"를
  확인한다(freeze P1-3).
  control **디렉터리** perm에는 대응 완화를 두지 않는다 — owner 비트를 깎는 umask는
  환경 이상으로 분류하고 그 endpoint는 강등·보고된다(P1-7①, a108의 정확-0700 검사와
  일관).
- **사망 증명**: 제거 전 connect probe. **A1 P1-A 판결로 a108 의례에서 한 가지를
  바꾼다** — owner-write-비트-없음=사망 추정은 수락 중인 0400 socket을 죽었다고 읽고
  지운다(A1 실측 재현). 이 시점의 socket은 이미 우리 uid 소유·비symlink·0700 디렉터리
  검증을 통과했으므로 **chmod 0600 후 dial**하면 생사가 결정적으로 갈린다(산 socket은
  수락→거부·보존, 죽은 socket은 ECONNREFUSED→회수; chmod 실패는 생존 간주). PID는
  판정에 쓰지 않는다(a102 D4b-2). **수락 중인 socket은 최종 이름이든 staging 이름이든
  절대 unlink하지 않는다**(A1 P2-D — socket 모양의 staging 엔트리에도 같은 probe).
  거부 오류는 위반 엔트리의 이름을 지목한다(A1 P2-E — D3 "원인 제거 후 재시작"의 전제).
- **1차 방어는 여전히 journal flock이다**: `runEngineRun`은 endpoint 기동 전에 flock을
  쥔다(부팅 1단계). probe는 flock이 막지 못하는 것(엔진 아닌 프로세스의 경로 점유,
  다른 journal 디렉터리의 엔진)에 대한 심층 방어이고, 코드 주석이 flock을 근거로
  명시 인용한다 — proposal의 "flock 명문화" 요구를 여기서 충족한다.

**D2a — command endpoint(TCP)의 회수는 staging 위생뿐이다.** socket이 없으므로 probe할
대상도, pre-chmod 병도 없다. descriptor는 rename이 덮는다(0바이트·잘린 잔재 포함 —
기동은 이전 descriptor를 읽지 않는다, freeze 생존 판정 5). 이 endpoint의 회수는 자기
staging 잔재(현행 CreateTemp 이름 포함) 제거로 한정하고, **낯선 엔트리는 오늘처럼
무시한다** — 열거+거부를 넣으면 이물 하나가 격리 해제 표면을 매 부팅 지우는 새 실패
경로가 생긴다(freeze P1-2). 낯선-엔트리 거부 시나리오는 socket 발행 endpoint에만
적용된다(spec delta도 그렇게 좁혔다).

**D2b — Close는 a108 계약을 따른다**: listener를 Close가 **직접** 닫고(late-unlink 경합
A1 F5 차단 — `AlertControlServer`는 현재 listener 필드 자체가 없다: 추가한다),
`SetUnlinkOnClose(false)`이므로 경로 제거는 Close의 제거 루프 하나뿐이다. 제거 순서
(descriptor→socket→dir)와 ErrNotExist 관용은 현행 유지.

## D3 — 세 fatal은 전부 강등한다 (proposal 0.3 판정)

`cmd/tossctl/engine.go`의 세 `return err`(:274/:279/:315)를 a108 D3의 강등 의례로 바꾼다.
endpoint별 판정 근거(freeze P0-4 반영 — **모든 논증의 기준선은 fatal이 만드는 상태다**:
autostart 영구 기동 루프 = 모든 포지션이 손절을 잃는다):

| endpoint | 잃는 것 | 판정 근거 |
| --- | --- | --- |
| policy runtime | 콘솔·읽기 화면 | a108 D3 원문 그대로 — 조회 전용, 잃는 것은 화면이지 판정이 아니다 |
| alert control | 운영자 ack 표면 | ack 불가 → 미전달 critical의 entry latch **유지** → 신규 진입 차단 지속(보수 방향). `AlertOperations`는 gateway에 닿지 않는다(transport 자기 문서 :24–28) — 손절·청산 경로 무관 |
| policy command | Preview/Apply·**격리 해제** 표면 | 격리 해제는 격리된 포지션의 **손절 포함 미판정 상태**를 푸는 유일한 장중 경로다(exitloop.go:560–568, exitpolicy/recovery.go:33–46). 강등은 그 경로를 잃는다 — "아무것도 느슨해지지 않는다"가 아니라 **"격리된 포지션의 무보호가 유지된다"**. 그래도 fatal(전 포지션 무보호)보다 엄격히 낫다 — 논증은 이 비교로만 성립한다 |

공통 근거: ① 엔진 싱글턴은 journal flock이 강제하고 endpoint는 루프의 입력이 아니라
표면이다(engine.go a108 주석 :288). ② fatal 유지가 만드는 것은 사고 모양 그대로다 —
잔재·이물 → autostart 영구 기동 루프 → **장중 손절 없음**. ③ D1·D2가 자기 잔재를 원천
소거한 뒤 강등의 잔여 원인은 이물·환경 이상이고 **결정적**이므로, 회복 수단은 "엔진
재시작"이 아니라 **"보고된 원인을 제거한 뒤 재시작"**이다(P0-4 — 재시작만으로는 같은
강등이 재현된다. 그래서 D3a의 원인 가시화가 회복 경로의 전제다).

**D3a — 보고는 a108 D3-2 교리를 그대로 일반화한다.** `reportStrategyProjectionDegraded`를
endpoint 이름을 받는 형태로 일반화(stderr 안내 + **의도적으로 등급표 미등재** obs Normal
이벤트). 금지 3종 불변: obs criticalEvents 등재 금지·severity critical 금지·원장 outbox
금지(미전달 PENDING 행은 다음 부팅의 진입 게이트를 잠근다 —
`engineStrategyProjectionDegradedEvent` 주석 :330–359가 정본). Notify는
`context.WithoutCancel` + goroutine(부팅 비차단, a108 T3-fix).

**D3a-2 — 소비자가 강등을 엔진 부재로 단정하지 않게 한다(freeze P0-3).** a108의 교리는
"콘솔·httpapi 전략 화면이 dormant로 뜬다"는 세 번째 표면을 전제했는데, 형제에는 그
표면이 없고 기존 소비자 메시지는 강등을 **다른 원인으로 단정한다** — alerts CLI는
"엔진이 없다, engine run을 살려라"(engine_alerts_client_unix.go:33–35, 엔진은 돌고
있는데), 격리 해제는 "control plane이 제공하지 않는다"(빌드가 낡았다는 뜻,
internal/console/exit_quarantine.go:227–229). 운영자는 그 말을 따라 엔진을 재시작하고,
원인이 결정적이므로 같은 강등이 재현된다. 수정은 **문구의 정직화**다(새 디스크 상태·
마커·화면 신설 없음): 두 메시지를 "엔진이 없거나, 엔진이 이 표면 없이 강등 부팅했다 —
엔진 로그의 강등 보고를 확인하라"로 바꾸고 메시지 텍스트를 테스트로 핀한다. 강등을
엔진 부재로 **단정하는** 소비자 메시지가 없어야 한다는 것이 spec 요구다.
`templates_position_policy.go`의 "배선되지 않아 조회만 가능"은 단정이 아니므로 유지
(선언된 생략).

**D3b — 강등해도 회수·발행 실패의 원인은 지운다.** 강등은 마지막 줄이지 수리가 아니다.
D1·D2가 자기 잔재 병을 원천 소거하므로, 강등이 실제로 발동하는 것은 이물·환경 이상뿐
이어야 한다. 그 상태로 부팅을 거듭 나는 것은 D3a의 보고가 매 부팅 보이게 한다.

## D4 — httpapi는 엔진이 돌아오면 다시 붙는다 (2b.1, freeze P0-1·P0-2·P1-4 반영 개정)

감독 순서 고정(compose 의존)이 아니라 **소비자측 lazy 재-dial**을 택한다. 근거: ①
감독 순서는 "가동 중 엔진 재시작 후 복귀"를 표현할 수 없다 — 재-dial은 냉부팅 순서와
가동 중 재시작을 한 기전으로 덮는다. ② 배포 파일이 아니라 코드가 계약을 진다.
③ 사람이 누르는 seam의 재사용이 아니다(boot-is-not-a-button).

설계(개정):

- **wrapper는 모든 상태를 감싼다 — live client 포함**(P0-1). 원 설계는 nil·sentinel만
  감쌌는데, 그러면 엔진이 살아 있을 때 뜬 httpapi가 잡은 live client는 엔진 재시작
  (새 socket·새 토큰) 후 영구 실패하며 재부착 대상이 아니다 — spec 시나리오 2가
  기전 없이 SHALL이 된다. 재부착 트리거는 "reader가 부재이거나 **직전 Read가
  실패했다**"이다.
- **요청 goroutine에서 Dial하지 않는다**(P0-2). `strategyprojectionrpc.Dial`은 본문에
  200ms connect probe를 포함한다(transport_unix.go:415, :29) — Snapshot 경로에서
  동기로 부르면 "요청 경로 비차단" SHALL NOT이 구현 시점에 거짓이 된다. 시도는
  **백그라운드 single-flight**로 돌고(rate limit 최소 간격, 30s 기본, 테스트 주입
  가능), 요청은 언제나 현재 값으로 즉시 응답한다. 요청 경로가 하는 일은 "시도를
  깨우는 것"까지다.
- **부재/unavailable 화면 구분을 wrapper가 보존한다**(P1-4). 소비자측 nil 검사가 두 곳
  있다 — `cmd/tossctl/httpapi_reader.go:565–571`(dormant 판정)과
  `internal/httpapi/strategy_runtime.go:18`(SSE 경로). wrapper는 정의상 non-nil이므로
  그대로 감싸면 화면이 dormant→unavailable로 회귀한다(a108 D4-2가 금지한 접힘).
  wrapper는 부재 상태를 별도로 노출하고, 두 nil 검사를 그 신호로 교체하며,
  dormant/unavailable 화면 값을 테스트로 핀한다. 이 두 파일은 T2 표면에 명시한다.
- **보고는 상태 전이 시 1회다**(P2-4). 30초마다 실패를 stderr에 찍지 않는다 — 시도
  실패는 침묵, 부착·탈착 전이만 로그한다.
- **상시 구동원은 publisher 루프다**(§2-fix F3 / A2 P2-3). 요청이 없어도 도는 것은
  `publishHTTPAPISnapshots` 하나뿐인데, 깨우기가 집계(`Refresh`) **성공** 안에 있으면
  전략과 무관한 조회 하나의 고장이 재부착을 영원히 잠근다 — `httpAPIReader.Snapshot`
  은 앞의 일곱 조회 중 하나만 실패해도 전략 블록 앞에서 끝나고 루프는 `continue`한다.
  깨우기는 `Refresh` **밖에서** 부른다. 새 ticker는 두지 않는다(주기는 이미 맞고,
  시도에는 자기 rate limit이 있다).
- **취소·늦은 실패는 endpoint 판정이 아니다**(§2-fix F1·F2 / A2 P2-1·P2-2). REST 경로가
  `request.Context()`를 그대로 넘기므로 요청 취소가 `context.Canceled`로 도착하고,
  `Read`는 잠금 밖에서 읽으므로 새 부착 뒤에 옛 실패가 도착할 수 있다. 전자는 호출자
  ctx가 이미 끝났을 때 상태 무변경, 후자는 읽기가 쓴 **자리의 세대**가 현재일 때만
  전이한다.
- **저장소 내 선례 대비**(P2-2): `positionPolicyRuntimeDescriptorReader`
  (position_policy_commander.go:27–41)는 같은 문제를 "rate limit 없이 매 read 재해결"로
  푼다. httpapi 전략 read는 화면 폴링·SSE로 고빈도이고 Dial에 200ms probe가 있으므로,
  엔진 다운 동안 매 read 재해결은 read마다 probe 비용을 낸다 — rate-limit 백그라운드
  시도를 고른 이유다.

엔진측 in-process 재시도 금지(a108 D2-2)는 그대로다 — 이것은 소비자다.

## D5 — 사고급 핀 (a108 2.5의 실패-불가 핀 교체)

a108 2.5 핀(`a108_every_endpoint_survives_a_leftover_test.go`)은 현재 코드가 이미
관용하는 모양만 깔아 **실패할 수 없었다**(a108 review §1 F4). a109가 사고급 모양을 깐다
— RED에서 실제로 실패해야 하는 것: ① pre-chmod 0700 socket 잔재에서 기동(두 socket
endpoint — 잔재는 umask 조작이 아니라 **명시적 chmod 0700**으로 만든다: umask는
프로세스 전역이라 병렬 테스트를 오염시킨다), ② 수락 중인 socket 위 두 번째 기동
거부(탈취 금지), ③ staging 잔재(현행 CreateTemp·신규 `.s-`)에서 기동 + 회수 후 잔재 0,
④ 낯선 엔트리 거부(socket endpoint 한정 — D2a). 이미 관용되는 모양(0바이트 descriptor
등)은 관용을 핀으로 고정한다. staging basename **길이 11 직접 측정** + 최종 basename과의
≤ 관계식을 상수 테스트로 고정한다(절대 경로 103 요구는 넣지 않는다 — D1a). 기존 a108
2.5 핀은 유지한다 — 관용의 핀이며 여전히 참이다.

**D5a — defer 순서 핀 (2b.2).** runEngineRun의 불변식 "endpoint Close는 journal flock을
쥔 채로 돈다"는 지금 defer 등록 순서 하나에 얹혀 있고 테스트가 없다. go/parser 기반
정적 테스트로 고정한다: runEngineRun 본문에서 `lock.Release()` defer가 모든 endpoint
`Close()` defer보다 **먼저 등록**된다(LIFO상 가장 나중 실행).

**D5b — 롤백 절차 일반화 (2b.3)는 코드 근거와 측정 둘 다로 쓴다(P2-3).** 코드 근거:
구버전(=현재 main) 회수 함수들(`PrepareRuntimeSocket`·`PreparePrivateSocket`·
`writePositionPolicyDescriptor`)은 어느 것도 `os.ReadDir`를 부르지 않으므로 신버전의
`.s-*` 잔재를 보지 못한다 — 사전 wipe 불요가 가설이다. 여기에 실측(구버전 바이너리에
신버전 잔재 모양별 기동)을 더해 배포 절차 prose에 **둘 다** 기록한다. 측정만 남기면
다음 사람이 왜 참인지 모르고, 주장만 남기면 미검증이다.

## 작업 분할 (파일 표면 비겹침)

| | 표면 | 내용 |
| --- | --- | --- |
| T1 | `internal/positionpolicyrpc/*`, `internal/app/engine/{position_policy_runtime_transport_unix.go, alert_control_transport_unix.go, position_policy_transport.go}` + 그 패키지의 a109 테스트 | D1·D2·D5(①~④·경계) + 뮤테이션 원장 |
| T2 | `cmd/tossctl/{engine.go, httpapi.go, httpapi_reader.go, engine_alerts_client_unix.go}`, `internal/httpapi/strategy_runtime.go`, `internal/console/exit_quarantine.go`(문구만) + cmd 패키지 a109 테스트 + 배포 절차 prose | D3·D3a-2·D4·D5a·D5b + 뮤테이션 원장 |

T2는 세 Start의 실패를 **`engineStrategyProjectionStart`(engine.go:416) 패키지 var seam
관례**로 시뮬레이트해 T1의 완료를 기다리지 않는다(freeze P1-5 — `cli_testseams.go`는
`internal/app/engine` 패키지, 즉 T1 표면이라 그 관례를 인용하면 T2가 T1 파일을 건드리게
된다). 각 Teammate는 자기 표면의 FLM 마크다운 맵(scaffold 완료, AST 존재)을 **편집
전에** 완성하고 구현 후 재최신화한다. High-risk(엔진 기동 경로) — Pre-Edit 선언 필수.

## 선언된 생략 (not-applicable)

- descriptor 발행 3벌의 fold(publishPrivateDescriptor 주석의 초대): 병이 없는 표면의
  High-risk 리팩터링 — 하지 않는다. alert control 판의 rename 직전 재검증 비대칭
  (P2-6)도 같은 표면이므로 함께 남긴다.
- strategyprojectionrpc의 staging 12자→11자 통일: a108 소유, `runtime.sock` 12자에
  적법 — 건드리지 않는다.
- 콘솔의 형제 endpoint 화면 **신설**: scope 밖. 단 기존 메시지의 오귀속 정직화는
  D3a-2가 한다(문구만).
- 콘솔 lifecycle client의 재부착(P2-7): `console.go:397`의 부팅 1회 dial은 httpapi와
  같은 병이지만 이 change는 httpapi만 고친다 — D4의 "재부착은 소비자의 것" 논지가
  콘솔에는 아직 적용되지 않았음을 기록한다. 후속 change 등록 후보.
- 콘솔의 전략 projection dial(`console.go:411–421`, A2 P1-1): 부팅 1회 dial 실패를
  nil로 접어 콘솔 화면이 NOT_CONFIGURED로 남는다 — 같은 후속 change 후보. 콘솔 page는
  구분하지만 콘솔 boot가 접는다.
- a108 `projectionSocketAccepts`의 owner-write 사망 추정(A1 P1-A의 a108 대응물,
  issues I1): a108 소유 코드라 이 change는 형제 기계만 고쳤다. 후속 change 등록 후보.
- `templates_position_policy.go`의 "배선되지 않아 조회만 가능" 문구: 원인 단정이
  아니므로 유지(D3a-2).
