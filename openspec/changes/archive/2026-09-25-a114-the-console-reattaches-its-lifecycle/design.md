# a114 설계 — 콘솔이 자기 lifecycle 에 재부착한다

작성 2026-09-25. base `634cf3c5`(a113 착지 뒤). 분기 주장은 `analysis/function-logic/` 의 AST 번들
4개(`runConsole` · `consolePositionPolicyCommander.quarantineClient` · `Console.handlePositionManagement` ·
`Console.decoratePositionRows`)를 먼저 만든 뒤 그 열거를 근거로 썼다. 원형은 a109 D4 의 httpapi
재부착(`cmd/tossctl/httpapi_strategy_attach.go`, http-api-service spec)이다.

## 0.1 합본 여부 (Manager 결정, 2026-09-25)

**합본하지 않는다.** Story↔change 1:1 계약을 유지한다. 같은 `runConsole` 을 편집하지만 a114 는
포지션 정책 lifecycle dial(AST B33–B37, console.go:395–409), a115 는 전략 projection dial(B38–B42,
:410–421)로 **다른 블록·다른 dial** 이다. 같은 Teammate 가 a114 → a115 순서로 순차 구현하고, a115 의
base 는 a114 착지 커밋 뒤에 잡는다 — 동시 편집이 없으므로 충돌하지 않는다.

## 병의 재확인 (AST 열거 기준)

- `runConsole` B32(`engineDir != ""`) 안에서 B33(:396, descriptor `os.Stat` 성공)일 때만
  `positionpolicyrpc.Dial` 을 **한 번** 부르고(B35 실패면 경고 + nil), 성공한 client 를
  `consolePositionPolicyCommander{lifecycle: client}` 로 굳힌다. B34/B37(:408)은 부재가 아닌 stat
  오류의 경고뿐이다. 다시 dial 하는 경로는 없다.
- 결과 ①: 콘솔 부팅 때 descriptor 가 없으면(엔진 정지·아직 기동 중 — 콘솔 자신이 autostart 한 엔진이
  아직 endpoint 를 발행하지 않은 순간 포함) commander 가 nil → `handlePositionManagement` 는
  `Wired: PositionPolicies != nil`(:225) = false 로 「배선되지 않아 조회만 가능」을 그리고 B5(:236)
  에서 끝난다. 엔진이 나중에 떠도 콘솔 재시작 전까지 그 상태다.
- 결과 ②: 부팅 때 붙은 client 는 엔진이 재시작하면(새 ephemeral 포트·새 토큰) 영구 실패한다.
- `quarantineClient`(B1 :28)는 `c.lifecycle.(exitQuarantineClient)` 로 격리 해제를 발견한다 —
  lifecycle 자리를 감싸는 값이 그 세 메서드를 갖지 않으면 **격리 해제 버튼이 사라진다**(회귀 위험).

## D1 — lifecycle 자리를 감싸는 wrapper (새 파일 `console_lifecycle_attach.go`)

`positionPolicyLifecycleAttachment` 가 `positionPolicyLifecycleClient`(List/Preview/Apply)와
`exitQuarantineClient`(Quarantines/PreviewQuarantineRelease/ReleaseQuarantine)를 둘 다 구현한다.
a109 D4 의 네 규칙을 옮긴다:

1. **모든 상태를 감싼다** — 비어 있음(부팅 때 못 붙음)뿐 아니라 부착 후 endpoint 실패한 live client
   도 재부착 대상이다(a109 freeze P0-1).
2. **요청 경로는 dial 하지 않는다** — `positionpolicyrpc.Dial` 은 health GET(연결 + 최대 5s)을
   품는다. 요청은 지금 자리의 client 로 즉시 답하거나(비어 있으면 즉시 detached 오류), 시도를 깨우는
   것까지만 한다. 시도는 백그라운드 single-flight, 최소 간격 30s(package var, 테스트 주입).
3. **전이 시에만 로그** — 부착·탈착 전이 한 줄씩. 시도 실패는 침묵.
3a. **렌더가 없어도 깨운다**(freeze P1-2). 콘솔에는 httpapi 의 publisher 같은 상시 구동원이 없다 —
   화면이 안 열려 있으면 엔진이 떠도 아무도 시도를 깨우지 않고, 콘솔 자신이 autostart 한 엔진도 그렇다.
   wrapper 는 콘솔 ctx 위에 **간격 주기의 펌프 goroutine** 을 하나 두고, 자리가 시도 대상(`failed`)일 때만
   `wake()` 한다(rate limit·single-flight 는 wake 가 그대로 진다). 비용은 탈착 동안 간격당 health GET 1회
   (descriptor 가 없으면 stat 뿐)이다. 간격이 0 이하(테스트)면 펌프를 띄우지 않는다.
3b. **부팅 해석은 `lastTry` 를 찍지 않는다** — 부팅 직후 첫 wake 가 간격만큼 막히지 않게 한다.
   a081 캐시(position_policy_cache.go)는 실패한 읽기를 그 간격 동안 들고 있으므로, 부착 뒤 첫 캐시
   갱신까지 화면이 늦을 수 있다 — 캐시의 설계대로이며 이 change 가 바꾸지 않는다.
4. **밀려난 client 는 놓아 준다** — `io.Closer` 면 닫는다(가짜 client 로 시험; freeze P2 — 운영 client 에는
   오늘 no-op 임을 숨기지 않는다). `positionpolicyrpc.Client` 는 오늘 Close 가
   없어 no-op 이고, 그 유휴 연결은 엔진 서버 `IdleTimeout: 15s`(internal/app/engine/
   position_policy_transport.go:153) 또는 엔진 사망 시 커널이 닫는다 — 상한 15초. Close 추가는
   파일 표면 밖(issues.md).

추가 규칙(콘솔 고유):

- **명령을 다시 보내지 않는다.** Preview·Apply·격리 해제는 엔진 상태를 바꾼다. 실패한 호출은 그대로
  운영자에게 가고, 재부착은 **다음** 호출을 위한 것이다(엔진이 이미 적용했을 수 있는 명령의 이중 전송
  금지).
- **엔진이 답한 것은 탈착이 아니다**(freeze P1-1 로 개정). 여섯 메서드가 자리 하나를 공유하므로,
  한 메서드의 반복 실패를 탈착으로 읽으면 다른 메서드의 성공과 번갈아 **전이 로그가 깜빡이고 멀쩡한
  client 가 매 렌더 교체된다**(예: List 성공 · Quarantines 가 엔진의 `"internal"` 500 — engine
  position_policy_transport.go:261). 그래서 「답했다」를 넓게 잡는다:
  ① 타입 있는 거절 — positionpolicy 11 · exitquarantine 8 sentinel(`errors.Is`, 목록 한 곳, exported
  `Err*` 선언과 완전성 테스트) ② **코드가 붙은 모든 rpcError** — 두 decoder 가 모르는 코드(`internal`
  등)에 쓰는 `"… control: remote failure"` 문구. 단 ③ **토큰 거절은 예외** — 엔진 auth 의 고정 문구
  `"local bearer token rejected"`(engine position_policy_transport.go:210)는 탈착이다: 재시작한 엔진이 같은
  포트에 다른 토큰으로 뜨면 옛 client 가 받는 답이고, 그것을 「답했다」로 읽으면 옛 토큰에 영구히 묶인다.
  나머지(연결 거부·timeout 의 `*url.Error`, 코드 없는 `HTTP %d`, 응답 해독 실패 — 우리 엔진이 아닌 것이
  그 포트에 있다)는 전부 탈착이다 — **모르는 오류의 기본값은 재부착**.
  문구 둘과 엔진 auth 문구는 **다른 파일의 상수**라서, 그 세 문자열을 go/parser 로 원본에서 찾는 테스트가
  문구 변경을 잡는다. 정석은 positionpolicyrpc 에 타입 있는 `ErrUnauthorized`·원격 실패 타입을 두는 것이고
  그것은 파일 표면 밖이다(issues.md 후속).
- **취소는 판정이 아니다** — `requestCancelled`(httpapi 원형, 같은 패키지 함수 재사용). **늦은 실패는
  옛 자리의 소식이다** — 자리 세대(seat) 비교.
- detached 오류 문구는 엔진 부재를 **단정하지 않는다**(spec): "엔진이 내려갔거나 아직 기동 중이거나
  이 표면 없이 강등 부팅했을 수 있다 … 엔진이 endpoint 를 발행하면 콘솔 재시작 없이 다시 붙는다".
  격리 해제의 detached 는 `exitquarantine.ErrUnwired` 로 감싸지 않는다 — 그 문구는 「이 빌드에
  배선되지 않았다」를 포함하므로 부착 전 상태에는 거짓이다.

## D2 — runConsole 편집은 블록 하나

B33–B37(lifecycle dial 블록)을 `positionPolicyCommander = consolePositionPolicyCommanderFor(ctx,
engineDir, cmd.ErrOrStderr())` 한 줄로 바꾼다. 새 함수(새 파일)는 부팅 1회 해석을 **오늘처럼 동기로**
한 번 하고(엔진이 이미 떠 있으면 즉시 붙는다), 그 결과로 wrapper 를 초기화하며, 부재가 아닌 실패만
경고 한 줄을 남긴다. 반환은 언제나 non-nil `*consolePositionPolicyCommander` 이다(engineDir 가 있을 때만
부른다 — nil 포인터를 인터페이스에 넣는 함정 없음).

화면 영향(편집하지 않는 소비자 — AST 근거):

- `handlePositionManagement`: `Wired` 는 engineDir 가 해석되면 true 가 된다. 부착 전 List 는 detached
  오류 → `LoadErr`(:252) 「불러오기 실패: …」로 **부재를 상태로** 그린다. 「배선되지 않았다」는
  engineDir 해석 실패(진짜 미배선)에만 남는다.
- `decoratePositionRows`: B1(:99) `runtimeAttempted` 가 true 가 되고, 실패한 읽기는 `policyByID` nil →
  행은 「관리 여부 불명」(fail-closed, a081 교리)로 간다. 연결 없는 detached 오류라 엔진 단일 write
  connection 을 쓰지 않는다.
- 콘솔 UI 에 타이핑 확인·추가 승인 마찰 없음(사용자 지시 2026-07-27) — 버튼·폼 무변경.

대안과 기각:

- 소비자 6곳의 `PositionPolicies != nil` 을 presence 신호로 바꾸기(a109 P1-4 모양): lifecycle 에는
  a108 식 dormant(기능 미사용) 개념이 없다 — engineDir 가 있으면 콘솔은 그 엔진의 콘솔이다. 6 함수
  편집·FLM 비용 대비 얻는 것은 「부착 전」을 「미배선」으로 그리는 것뿐이고 그것이 이 병의 원인이다. 기각.
- 요청마다 재해결(`positionPolicyRuntimeDescriptorReader` 선례): Dial 이 health GET 을 품고 List 는
  화면 수만큼 불린다. spec 의 SHALL NOT(요청 경로 dial) 위반. 기각.
- httpapi `strategyRuntimeAttachment` 제네릭화: 기존 a109 함수 편집(표면 밖). 구조를 옮겨 적고
  `requestCancelled` 만 재사용한다.

## 비목표

- `positionPolicyRuntimeDescriptorReader`(runtime unix endpoint, 매 read 재해결) — 별도 endpoint, 무변경.
  commander 가 non-nil 이 되면서 탈착 중에도 렌더마다 불리게 된다(freeze P2): descriptor 가 없으면
  open 실패로 즉시 끝나고, 있으면 unix socket 연결 1회다(a081 캐시가 빈도를 제한). spec 은 이 읽기를
  SHALL NOT 의 대상(lifecycle dial)에서 명시적으로 제외한다.
- `writePositionPolicyError`·`writeQuarantineError` 의 기본 갈래가 transport 실패에도 「아무것도 변경되지
  않았다」를 붙이는 문구(freeze P2): 부착된 자리에서 Apply 가 timeout 나면 거짓일 수 있다. 표면 밖 —
  issues.md.
- 콘솔 전략 projection dial — a115.
- `positionpolicyrpc.Client.Close` 추가 — 표면 밖(issues.md).

## 테스트 전략

진짜 엔진 endpoint(`engine.StartPositionPolicyCommandServer`)로 늦은 기동·재시작 두 시나리오, 가짜
client·손 시계로 기전(요청 비차단·single-flight·rate limit·전이 1회·취소·늦은 실패·답한 거절·명령
미재전송·격리 해제 전달·문구), go/parser 로 `runConsole` 이 lifecycle 을 직접 dial 하지 않음과 답 목록
완전성·문구 원본 핀. 뮤테이션은 a109 T2 원장의 콘솔 적용판(M8b·M9·M10b·M11·M12b·M13·M14·M15·
M29·M30·M35·M36) + 콘솔 고유(답 분류 전부-실패·토큰 거절 답 취급·재전송·격리 전달·detached 의 ErrUnwired
감싸기·부팅 lastTry·펌프 제거·부팅 dial 재도입)를 사본 하네스로 잰다. a109 M31(publisher wake)은 펌프 테스트가
대신하고, M33(빈 자리 sentinel)·M34(무조건 wake — 집계가 읽기를 가리는 경로가 콘솔에는 없다)·M37·M38
(positionpolicyrpc transport — 표면 밖)은 not-applicable 사유를 원장에 적는다.
`go test -race -count=1 ./cmd/tossctl/ -run …` + 콘솔 실측(규칙 13: 격리 config 로 콘솔 기동, 엔진 없이
`/position-management` 본문 확인).
