# a115 설계 — 콘솔 전략 화면도 재부착한다

작성 2026-09-25. base `8688f74f`(a114 착지 뒤). 분기 주장은 `analysis/function-logic/` 의 AST 번들 5개
(`runConsole` · `Console.buildMultiMarketStrategyRuntimePage` · `Console.strategyRuntimeSummary` ·
`strategyRuntimeAttachment.StrategyRuntimeConfigured` · `resolveStrategyRuntimeReader`)를 먼저 만든 뒤 그 열거를
근거로 썼다. 원형은 a109 D4 의 httpapi 재부착(`httpapi_strategy_attach.go`)과 그 소비자 nil 검사 교체(a109 freeze P1-4).

## 0.1 합본 여부 (Manager 결정, 2026-09-25)

**합본하지 않는다** — a114 design 0.1 과 같은 결정·근거(Story↔change 1:1, 같은 `runConsole` 의 다른 블록·다른
dial, 같은 Teammate 순차 구현). a115 의 base 는 a114 착지(`1244be95`)·기록 커밋 뒤의 `8688f74f` 다.

## 병의 재확인 (AST 열거 기준)

- `runConsole`(a114 이후 분기 38) B32 안의 전략 블록 B33–B37: descriptor `os.Stat` 성공(B33)이면
  `strategyprojectionrpc.Dial` 을 **한 번** 부르고, 실패(B35)면 경고 + **nil**(「전략 화면은 dormant로 뜬다」),
  성공(B36)이면 client 를 굳힌다. 비부재 stat 오류(B34/B37)도 nil.
- `buildMultiMarketStrategyRuntimePage` B1(:49) 은 `StrategyRuntime != nil` 로 Read 여부를, :48 의
  `Unwired: StrategyRuntime == nil` 로 dormant 를 가른다. `strategyRuntimeSummary` B1(:291) 도 nil 로 가른다.
- 결과 ①(A2 P1-1): 구성된 runtime(descriptor 있음)에 닿지 못하면 boot 가 nil 로 접어 화면이 NOT_CONFIGURED —
  도달 불가를 미구성으로 오귀속한다. ② 부팅 1회라 엔진이 늦게 뜨거나 재시작하면 콘솔 재시작 전까지 회복하지 않는다.
  콘솔 자신이 autostart 한 엔진이 아직 endpoint 를 발행하지 않은 순간에 부팅하면 ①·② 가 겹친다.

## 「구성」의 뜻 — descriptor 가 경계다 (freeze 리뷰 P1-3)

엔진은 projection endpoint 를 항상 열고(설정 토글 없음), **깨끗이 종료하면 descriptor·socket 을 지운다**
(transport 의 Close). 그래서 「구성됐는데 엔진이 내려감」이 화면에서 도달 불가로 보이는 것은 descriptor 가
**남은** 다운(비정상 종료·행업)뿐이고, 깨끗한 정지 뒤나 autostart 의 발행 전 순간에는 dormant 가 뜬다.
이것은 오귀속이 아니라 **선언된 한계**다: dormant 문구 자체가 「runtime endpoint 미기동」을 말하고 엔진
부재를 단정하지 않으며(a109 D3a-2), 펌프가 endpoint 발행 즉시 다시 붙는다. 과도 창의 상한은 펌프 tick +
재시도 간격이다. 대안 — 양의 신호(엔진 마커 등)로 「구성」을 판정 — 은 기각한다: 판정 두 벌(마커 vs
descriptor)이 갈라질 수 있고, dormant 표기가 이미 사실을 말하므로 기계를 늘릴 이유가 없다. spec 시나리오는
이 경계로 적는다(descriptor 잔존 다운 → 도달 불가 · descriptor 부재 → dormant, 둘 다 재부착으로 회복).

**구현 후 정정(review §1 P2-1)**: 위 「깨끗한 정지 뒤에는 dormant 가 뜬다」는 **정지 뒤 부팅한** 콘솔에만 참이다.
이미 live 로 붙어 있던 콘솔은 깨끗한 정지(descriptor 삭제)에도 dormant 로 내려가지 않는다 — wrapper 는 live 자리를
부재로 격하하지 않으므로(a109 `attempt` B1, 무편집 재사용) 엔진이 돌아올 때까지 도달 불가(「읽지 못했다」)다. 그
문구도 엔진 부재를 단정하지 않으므로 spec SHALL 위반은 아니고(시나리오 3 은 부팅 시점), 같은 디스크가 이력에 따라
두 값으로 보이는 비대칭은 선언된 한계로 둔다 — 핀 `TestAnAttachedConsoleShowsACleanStopAsUnreachable`, issues R6.

비부재 stat 오류(EACCES·ENOTDIR)가 nil → dormant 로 접히는 것도 같은 자리의 **선언된 접힘**이다 — httpapi 의
판정과 동일하게 유지한다(리뷰 P2-2). httpapi spec 의 「반쪽 잔재」 문장과 현 코드(sentinel)의 불일치는 이
change 밖 — issues R5.

## D1 — 콘솔은 httpapi 의 재부착 wrapper 를 **그대로** 쓴다

`strategyRuntimeAttachment`(a109 D4, 같은 패키지)는 부재(nil)·도달 불가(sentinel)·live 세 상태를 감싸고, 요청
경로 비차단·single-flight·rate limit·전이 1회 로그·취소·늦은 실패·밀려난 client Close 를 이미 갖췄다. 옮겨 적지
않고 **재사용**한다(복사한 기계는 어긋나기 시작한 기계다). 새 파일 `console_strategy_attach.go` 가 둘을 더한다:

- `resolveConsoleStrategyRuntime(ctx, engineDir, errOut)` — 콘솔용 해석. 판정은 httpapi 의
  `resolveStrategyRuntimeReader` 와 같다(부재 → nil, 비부재 stat 오류 → 경고 + nil, dial 실패 → **sentinel**
  `unavailableStrategyRuntime`, 성공 → client). 다른 것은 경고 문구뿐이다 — httpapi 판은 「데몬」을 말하고 root 에서
  디렉터리를 다시 푼다. 콘솔은 이미 푼 engineDir 를 쓴다.
- `consoleStrategyRuntimeReaderFor(ctx, engineDir, errOut)` — 부팅 1회 해석을 오늘처럼 동기로 하고 그 결과로
  wrapper 를 초기화한 뒤, 콘솔 ctx 위에 **펌프**를 띄운다: 간격의 절반마다 **무조건** `wake()`. 화면이 안 열려
  있어도 엔진이 뜨면 붙는다(a114 freeze P1-2 와 같은 이유 — 콘솔엔 publisher 가 없다).
  「시도 대상일 때만 깨우기」는 기각한다(freeze 리뷰 P1-2): wrapper 의 G2 주석
  (httpapi_strategy_attach.go:102–112)이 바로 그 `failed` 게이트를 기각했다 — live 로 보이는 죽은 자리(엔진 재시작
  뒤 옛 client)는 그 게이트로 드러나지 않아, 렌더 없는 동안 죽은 채 남고 첫 렌더가 「읽지 못했다」를 그린다.
  무조건 wake 의 비용은 wake 안의 rate limit·single-flight 가 「간격당 시도 1회」로 고정한다(같은 파일 :110–112,
  AST `strategyruntimeattachment.wake` B1). 반환은 **언제나 non-nil** 이다 — nil `*strategyRuntimeAttachment` 가
  interface 에 담기면 presence 질문이 nil mutex 에서 패닉한다(:140–143, 리뷰 P2-4). 펌프 간격은 httpapi 의
  `strategyRuntimeRedialInterval` 을 공유하지 않고 **콘솔 전용 변수**로 둔다(a109 시험이 30s 를 고정하고 있다,
  리뷰 P2-5) — interval≤0 이면 펌프를 띄우지 않는 가드(0 주기 ticker 패닉, a114 선례), 시험은 cancel 과
  inFlight 대기로 goroutine 을 정리한다. 재시도 해석의 경고는 `io.Discard` 로 버린다(httpapi.go:263–267 선례 —
  엔진이 하루 내려가면 2880줄, 리뷰 P2-3). 새 부팅 경고 문구는 「전략 화면은 dormant로 뜬다」를 말하지 않는다 —
  a115 뒤에는 거짓이다(현 console.go:401 문구, 리뷰 P2-3). 둘 다 시험한다.

## D2 — 화면은 nil 이 아니라 **부재 신호**를 묻는다 (소비자 2함수 · 판정은 한 벌)

wrapper 는 정의상 non-nil 이므로 화면의 `== nil` 판정을 그대로 두면 부재가 「읽지 못했다」로 회귀한다(a109 freeze
P1-4 와 같은 병의 반대편).

**판정을 새로 만들지 않는다(freeze 리뷰 P1-1).** 첫 판은 `internal/console` 에 `strategyRuntimeWired` 를 새로
두려 했는데, 같은 판정 `httpapi.StrategyRuntimeAbsent` 가 이미 있고 그 주석이 「⛔ 이 검사를 소비자마다 다시 쓰지
마라」다(internal/httpapi/strategy_runtime.go:45–66, 시험 `TestTheAbsenceSignalIsOneJudgementForEveryConsumer` 가
한 벌임을 고정). 두 벌은 갈라지고, 갈라지면 같은 디스크 상태가 화면마다 다른 값이 된다. 대신:

- 판정(`StrategyRuntimeAbsent`)과 presence 인터페이스(`StrategyRuntimePresence`)를 **`internal/strategyprojection`
  으로 옮긴다** — 두 소비 패키지(httpapi·console)가 이미 공유하는 잎 패키지이고, `strategyprojection.Reader` 가
  같은 메서드 집합으로 이미 있다. `internal/httpapi` 쪽 이름은 type alias + 위임으로 유지해 기존 네 소비처와
  시험이 그대로 선다(httpapi 표면 편집은 이 이동·위임뿐 — issues S1).
- `internal/console` 두 소비자의 nil 판정을 그 한 벌 판정으로 바꾼다(한 번 물어 두 자리에 쓴다).
- `cmd/tossctl` 에 `var _ strategyprojection.StrategyRuntimePresence = (*strategyRuntimeAttachment)(nil)` 컴파일
  결속을 추가한다 — 구조적 인터페이스만으로는 메서드 이름이 바뀌어도 컴파일이 통과하고 부재가 조용히
  「읽지 못했다」로 회귀한다(리뷰 P1-1). 세 상태(nil·sentinel·live)와 신호 없는 reader 에서 httpapi 위임과
  판정이 같음을 시험으로 고정한다.

**구현자 주의(리뷰 P2-10)**: console 의 seam 인터페이스(`MultiMarketStrategyRuntimeReader`)는 정확히 `Read` 하나여야
한다(`internal/console/static_test.go:1010` 이 강제). presence 질문은 seam 에 메서드를 넣지 말고 판정 함수 안
타입 단언으로 남긴다.

**파일 표면 해석**: 등록 문서는 "cmd/tossctl/console.go 부팅 경로 + 테스트"라 적었지만 tasks 0.3 이 「소비 page 의
FLM」을, Manager 지시가 「화면이 dormant/unavailable 구분」을 범위로 적었다. 부팅 경로만 고치는 대안(아래)은 스펙
시나리오를 지키지 못하거나 dormant 안내를 잃는다. 두 함수의 nil 판정 한 줄씩 + 판정·presence 의
strategyprojection 이동(httpapi 는 alias·위임)으로 한정하고 issues.md 에 safe-local 로 기록한다.

화면 문구: 바꾸지 않는다 — dormant 는 「runtime endpoint 미기동 — … dormant … truth」, 도달 불가는 「runtime
projection을 읽지 못했다 …」이고 둘 다 엔진 부재를 단정하지 않는다(a109 D3a-2). 구조·문구 무변경이라
`ui-skills-root` 규칙 선택은 not-applicable.

대안과 기각:

- 부팅만 고쳐 nil/sentinel 을 가르고 재부착은 descriptor 가 있을 때만: 콘솔 부팅 때 descriptor 가 없으면(깨끗한
  종료 뒤·autostart 직후) 영구 dormant — 「엔진이 뒤늦게 뜬 뒤 재부착」 SHALL 위반. 기각.
- wrapper 가 부재일 때 `DormantSnapshot` 을 지어 답한다(화면 무편집): 부재를 스냅샷으로 답하는 것은 a109 가 금지한
  새 접힘이고(httpapi_strategy_attach.go `Read` 주석), dormant 안내 줄이 사라진다. 기각.
- httpapi 의 `resolveStrategyRuntimeReader` 재사용: 판정은 같지만 경고가 「데몬」을 말하고 root 로 디렉터리를 다시 푼다.
  판정 4줄을 옮겨 적는 대신 판정 동치를 테스트로 고정한다.

## 비목표·잔존

- wrapper 의 전이 로그 문구는 「데몬은 그대로 돈다」를 말한다(httpapi_strategy_attach.go `observe` — a109 코드, 표면 밖).
  콘솔 stderr 에서 「데몬」은 이 콘솔 프로세스를 뜻하게 된다 — issues.md.
- 전략 dial 외 콘솔 부팅 경로, httpapi 의 동작: 무변경.

## 테스트 전략

internal/console: 부재 신호 false 인 non-nil reader → dormant(Read 0회·신호 물음), 신호 true + Read 실패 → 도달 불가
(NOT_CONFIGURED 아님), 요약도 같은 판정, 신호 없는 reader 는 오늘처럼 wired. cmd/tossctl: 진짜
`strategyprojectionrpc.Start` 로 ① 구성·엔진 다운(죽은 descriptor+socket) → 도달 불가 → 엔진 기동 후 회복 ② 미구성
(descriptor 없음) → dormant → 늦은 엔진 → 회복 ③ 렌더 없이 펌프로 부착 ④ 재시작 재부착, 그리고 `runConsole` 이
`strategyprojectionrpc.Dial` 을 직접 부르지 않음(AST, base 에서 RED). 뮤테이션: nil 접힘 재도입 · 화면 nil 판정 복귀
(두 자리 각각) · 부재 신호 무시 · 펌프 제거 · 부팅 dial 재도입 · sentinel 대신 nil.
