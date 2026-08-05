# a082 · Review

## 2026-08-05 · proposal-freeze (Manager 셀프리뷰)

WORKFLOW.md 리뷰 게이트: 첫 구현 task 착수 전 proposal/design/spec 델타 리뷰 1회.
**인증 경로는 High-risk이므로 적대적 Eng 관점을 포함한다** (위험 등급 가중).

### 근거가 실제로 코드에 있는가

[[tossos-manager-design-from-docs]]의 교훈 — OpenSpec을 쓰기 전에 그 결정이 읽을
값·경로가 코드에 있는지 확인한다. 확인한 것은 넷이다.

| 주장 | 확인 |
|---|---|
| `saveCache`의 호출자는 `exchange()` 하나 | `rg` 재검증: token.go:139 한 곳 |
| `refresh()`의 production 호출자는 `send()`의 401 갈래 하나 | client.go:331 한 곳 |
| `token()`이 반환 직전 `m.cache`를 반환값으로 맞춘다 | 세 갈래 전부. `refresh()`의 채택 비교가 여기 기댄다 |
| `ErrAuth` 판정이 전부 `errors.Is`/`errors.As` | production 5곳. **테스트 1곳이 `==`였다 → issues I5** |

**CodeGraph가 틀렸다.** `classifyStatus`의 호출자를 `send` 하나로 보고했는데 HEAD
재검증에서 `token.go:123`(OAuth 교환 자체)이 하나 더 있었다. CLAUDE.md가
"CodeGraph는 hard evidence이되 현재 HEAD로 재검증한다"고 적은 이유가 이것이다.
그 두 번째 호출자 덕에 **교환 엔드포인트의 인증 실패도** 상태 코드를 갖게 된다.

### 적대적 관점 — 이 설계가 틀릴 수 있는 곳

**Q. 토큰 획득 경로에 stat을 넣는 것이 손절 판정을 늦추지 않는가?**
로컬 파일 stat 한 번이고, 같은 경로가 이미 실패 시 파일 **읽기**를 한다. 더 중요한
것은 무엇을 **넣지 않았는가**다 — 프로세스 간 블로킹 락은 거절했다(design D3).
이 경로는 exit 루프의 모든 읽기가 지나므로 여기서 다른 프로세스를 기다리면 그
시간이 손절 판정 간격에 더해진다(안전 불변식 3). 스펙에 SHALL NOT으로 적었다.

**Q. 채택이 죽은 토큰을 물어올 수 있지 않은가?**
채택 조건이 "**다르면**"이라 방금 거부당한 토큰은 절대 채택하지 않는다. 다른
토큰이 그 사이 죽었다면 재시도가 401을 받고, 그 다음 `refresh()`가 이번엔 같은
토큰을 보고 교환한다. 무한 루프가 없는 이유는 "다르다"가 매번 좁아지기 때문이다.
변이 M4가 이 조건을 "항상"으로 넓히자 **기존 `TestTokenRefresh`가 깨졌다**.

**Q. 오류를 감싸면 분류가 바뀌지 않는가?**
`%w`라 `errors.Is`가 그대로다. 그리고 문자열로 판정하는 곳이 없다는 것을
`TestNothingDecidesAnAuthRefusalByReadingItsMessage`가 고정한다 — 같은 문구의
무관한 오류가 `errors.Is`를 만족하지 **않는** 것까지 단언한다.

**Q. 응답 본문을 로그에 흘리지 않는가?**
코드만 싣는다. `TestAnAuthRefusalDoesNotCarryTheResponseBody`가 계좌 번호를 닮은
문자열로 고정한다. `*APIError` 갈래는 원래대로 본문을 계속 싣는다 — 그쪽은
passthrough이고 이 change가 건드리지 않는다.

### 스코프 판정

`send()`의 403 재시도와 다른 공유 파일들은 **뺐다**. 근거는 issues I1·I2.
403은 브로커가 실제로 무엇을 주는지 모르는 상태에서 재시도 의미를 바꾸는 것이라
근거가 없고, 이 change의 상태 코드 수정이 그 근거를 만든다.

**판정: 착수 가능.**

## 2026-08-05 · 구현과 검증

### 변이 검증 4종

| 변이 | 무엇을 지웠나 | 결과 |
|---|---|---|
| M1 | `refresh()`의 채택 갈래 | **RED** — `TestARotationThatLandsMidRequestCostsNoToken` |
| M2 | `token()`의 파일 변경 감지 | **RED** — `TestTokenPrefersTheCacheFileWhenAnotherProcessRewroteIt` |
| M3 | 상태 코드를 다시 버림 | **RED** — `TestAnAuthRefusalCarriesItsStatusCode` (4행 전부) |
| M4 | 채택 조건을 "다르면"→"항상" | **RED** — `TestARefusedProcessWithNothingToAdoptStillExchanges` **와 기존 `TestTokenRefresh`** |

M4가 가장 중요하다. **손대지 않은 기존 테스트가 깨진다**는 것이 이 change의 좁음을
증명한다 — 단일 프로세스 의미론은 글자 그대로 보존된다.

### M1이 처음에 RED가 아니었다

첫 시도에서 채택 갈래를 지웠는데 헤드라인 테스트가 **통과했다**. 파일 변경
감지(D2)가 그 시나리오를 이미 흡수하고 있었다.

그대로 뒀으면 근거 없는 코드가 남는다. 그래서 D1이 실제로 무엇을 사는지 다시
따졌고, D2가 볼 수 없는 창이 하나 있었다 — **토큰을 건네준 뒤, 브로커가 답하기
전에 일어난 회전**. `TestARotationThatLandsMidRequestCostsNoToken`이 그 창을 열고,
M1은 그 테스트에 대해 RED다.

**이것이 production에서 실제로 일어난 창이기도 하다.** 재시도가 흡수하지 못한
거부 10건이 전부 그 창에서 나왔다 — 다른 창이었다면 재시도가 살렸을 것이다.

비중은 설계 초안이 암시한 것과 다르다: D2가 무거운 일을 하고 D1은 남은 경주를
닫는다. issues I6에 적었다.

### 관측

| | |
|---|---|
| RED 관측 | 24 요청 → **23 교환** (수정 전, 헤드라인 테스트) |
| `internal/official` | **208 passed**, 0 failed |
| `-race` | clean |
| `make test` 전 패키지 | 0 failed |
| 기존 테스트 수정 | **1건** — issues I5에 이유 |
| Function Logic Map | 5 target `evidence complete` |

logic-map target이 3에서 5로 늘었다. `exchange`(stamp 한 줄)와 위 기존 테스트가
수정된 기존 함수로 잡혔다 — [[tossos-logic-map-scope-creep]]의 반복이다.

## 2026-08-05 · Requirement 변경 리뷰 (WORKFLOW.md §142, task 7.2)

`persistent credential and token lifecycle`을 MODIFY하므로 요구된다. 기존 절과
scenario 6건은 글자 그대로 보존했고, 절 셋과 scenario 넷을 더했다.

### 기존 문장과 충돌하는가

기존: "Short-lived access-token expiry SHALL be handled by the **existing official
client renewal behavior**".

더한 절은 그 renewal behavior가 **공유 파일 위에서 무엇을 해야 하는지**를 말한다.
기존 문장을 부정하지 않고 그 안을 채운다 — 여전히 official client가 처리하고,
여전히 운영자가 키를 다시 넣지 않는다. 새 요구사항을 따로 세우지 않은 이유가
이것이다: 별도 ADDED로 세우면 "renewal은 client가 알아서" 하는 SHALL과 "renewal은
이렇게 수렴해야" 하는 SHALL이 둘이 되어 어느 쪽이 이기는지 문서가 말하지 않는다.

기존: "The setup path SHALL be offered when credentials are missing or
**authentication-rejected**, not merely because the cached access token expired."

이 change는 그 판정에 닿지 않는다. 인증 거부로 분류되는 오류 집합은 한 글자도
안 바뀌고(`classifyStatus`의 조건식 보존), 바뀌는 것은 메시지뿐이다. 다만 이
change가 없으면 **토큰 경합이 authentication-rejected로 보이므로** setup path가
잘못 제안될 수 있었다. 그 위험을 줄이는 방향이다.

### `credential ingress is secret-safe`와 충돌하는가

그 요구사항: "The console SHALL NOT place the key or secret in HTML responses,
redirect URLs, **error text**, application logs, audit details, retained memory,
or test output."

새 SHALL은 오류 텍스트에 **HTTP 상태 코드**를 넣는다. 키도 시크릿도 아니다.
그리고 같은 절이 **응답 본문은 넣지 않는다**를 SHALL NOT으로 못박는다 — 본문에
계좌 식별자가 들어올 수 있기 때문이다. 두 요구사항이 같은 방향이고,
`TestAnAuthRefusalDoesNotCarryTheResponseBody`가 그것을 고정한다.

### 안전 불변식과 충돌하는가

새 SHALL NOT("이 수렴을 위해 프로세스 간 블로킹 대기를 도입하지 않는다")은 안전
불변식 3(손절·비상 청산의 즉시성을 약화·지연하지 않는다)을 스펙 문장으로 옮긴
것이다. 토큰 획득은 exit 루프의 모든 읽기가 지나므로, 여기 락을 놓으면 한
프로세스의 지연이 다른 프로세스의 판정 간격에 더해진다. 이 change가 flock 선례
둘(`internal/enginelock`, `internal/config/adoption_flock_unix.go`)을 따르지 않은
이유이고, design D3에 남겼다.

### 판정: 수용

- 요구사항을 넓히지 않는다. 더한 것은 전부 **기존 renewal behavior의 계약**이고,
  그중 둘은 SHALL NOT(무엇을 하지 말 것)이다.
- scenario 넷은 전부 코드로 관측 가능하고 테스트가 하나씩 붙어 있다.
- 다른 스펙과의 충돌 없음. `engine-safety`·`reconciliation`은 이 경로의 오류
  **분류**에 기대는데 분류는 안 바뀐다.

## 미결 — 배포 뒤에 답이 나오는 질문

`send()`는 401에만 재시도한다. 브로커가 낡은 토큰에 403을 준다면 이 change의 채택
갈래에 **닿지 않는다**. 상태 코드 수정이 그 답을 만들고, task 6.7의 실측이 읽는다
(issues I1).

`make lint`는 이 change 이전부터 red다 — `internal/httpapi/performance_attribution_test.go`의
gofmt 드리프트이고 base `f1aae509`에서 이미 그렇다. 이 change의 파일은 전부
gofmt clean이고 `make gate`는 lint를 돌리지 않는다.

## 2026-08-05 · 독립 리뷰 1 (적대적 안전 렌즈, 별도 worktree) 과 그 반영

**결과: P0 2건, P1 3건, P2 2건.** 초안 설계의 절반을 철회했다.

### P0-1 — stat 실패가 요청마다 교환을 만든다 → **D2 철회**

`cacheFileChanged()`가 stat 오류를 "바뀌었다"로 읽었다. 파일을 stat할 수 없으면
(ENOSPC, 권한 드리프트, bind mount 부재로 read-only root에 떨어짐) 유효한 24시간
토큰을 손에 쥐고도 **매 호출 교환**한다. 리뷰어 측정: 같은 조건에서 base 1회,
초안 **10회**. 그리고 그 교환은 `m.mu`를 쥔 채 도는 네트워크 호출이다.

### P0-2 — 채택한 토큰은 검증된 토큰이 아닌데 유일한 재시도를 거기 쓴다

회전 경계에서 **마지막에 쓴 보유자와 마지막에 발급받은 보유자가 다를 수 있다.**
그러면 파일에 죽은 토큰이 있고 채택이 재시도를 거기 쓴다. 리뷰어 측정: base
`err = <nil>`, 초안 `authentication failed (HTTP 401)`.

그 거부의 대가를 리뷰어가 끝까지 추적했다 — `ClassAuthFatal` → `Gate.Block` +
`escalateCredentialFailure`(재시작이 못 푸는 영속화) → `EntryGate.Clear`는 운영자
전용. 그리고 그것을 올린 exit 사이클은 손절 판정을 하지 않는다. **내가 코드로
재확인했다** (`retry.go:334-338,507-509`, `exitloop.go:699-708`).

### P1-3 · P1-4 · P1-5 · P2-6 · P2-7

- **P1-3** 읽고 나서 stamp하면 옛 바이트에 새 mtime이 찍혀 감지기가 영구히 죽는다.
- **P1-4** `saveCache`가 truncate 후 write라 torn read가 난다. → D6로 채택.
- **P1-5** design D1의 "형제 goroutine" 주장이 거짓. 형제도 파일을 쓰므로 토큰이
  같아 보여 교환한다. 엔진은 loop마다 goroutine을 띄운다(`runtime.go:277-283`).
  → `refresh(ctx, refused)`로 추론 자체를 없앴다.
- **P2-6** mtime 동등 비교는 같은 tick 안의 두 쓰기를 구분 못 한다.
- **P2-7** stat이 `m.mu` 아래 hot path에 있고 취소할 수 없다. 측정 ext4 775ns,
  FUSE **87µs**, NFS 무응답 시 mount timeout까지 블록. **락을 거부한 근거(D3)를
  stat이 그대로 어긴다.**

### 처분 — 넷을 한 번에 없앴다

P0-1·P1-3·P2-6·P2-7은 **전부 mtime 확인 하나**를 가리킨다. 그것을 철회했다.
대가는 회전마다 보유자당 401 한 번이고 `send()`의 재시도가 흡수한다 — 하루 3회.
그 값에 위 넷을 사지 않는다.

P0-2는 `refresh`가 채택/발급을 반환하고 `send`가 채택 뒤에만 한 번 더 도는 것으로
고쳤다 (design D5). P1-4는 임시 파일 + rename (D6). P1-5는 `refused` 인자 (D1).

### 형제 goroutine 전용 갈래도 뺐다

P1-5 반영으로 메모리 캐시를 먼저 보는 갈래를 넣었다가, **변이 검증에서 그것을
지워도 아무 테스트가 안 깨졌다**. `exchange`가 반환 전에 파일을 쓰므로 형제의
토큰은 파일 갈래가 이미 찾아낸다. 근거 없는 코드는 남기지 않는다 — 뺐다.

### 리뷰어가 확인해 준 것

- `ErrAuth` wrap이 저장소 어디의 분류도 바꾸지 않는다. 소비자 7곳 전부
  `errors.Is`/`errors.As`이고, `orders_raw_test.go` 수정 뒤 sentinel `==` 비교는
  저장소에 **0건**이다. `failclosed.go:190-198`이 `err.Error()` fallback을 명시적으로
  거부하므로 `(HTTP 401)` 접미사가 운영자 분기 분류에 닿지 않는다.
- 본문은 안 실린다. 새 시크릿·PII 표면 없음. `saveCache`는 0600 유지.
- 채택 경로에 nil 역참조 없음. 손상된 캐시 파일이 panic을 내지 않는다.
- `saveCache`의 유일한 호출자, `refresh`의 유일한 production 호출자 — 둘 다 확인.
- 단일 프로세스 의미론 보존, `-race`·`go vet` clean.

### 개정 후 변이 검증 6종

| 변이 | 결과 |
|---|---|
| M1 파일 채택 갈래 제거 | **RED** — 헤드라인 + 중간 회전 + 채택 3건 |
| M2 형제 전용 갈래 제거 | **통과 → 그 갈래를 삭제했다** |
| M3 상태 코드 다시 버림 | **RED** |
| M4 채택에서 `!= refused` 제거 | **RED** — 5건, 기존 `TestTokenRefresh` 포함 |
| M5 `send`의 상한을 2→1 | **RED** — 채택-후-거부 테스트 |
| M6 `saveCache`를 plain write로 | **RED** — 400회 중 **244회** torn read |

### 관측 (개정 후)

| | |
|---|---|
| `internal/official` | **210 passed**, 0 failed |
| `-race` | clean |
| `go vet` / gofmt | clean |
| 기존 테스트 수정 | **2건** — `orders_raw_test.go`(I5), `token_test.go`(I7) |
| Function Logic Map | **6 target** `evidence complete` |

### 운영 사고 하나 — 리뷰어 worktree가 저장소 안에 생긴다

`make test`가 갑자기 7건 실패했다. 원인은 내 코드가 아니라
`.claude/worktrees/agent-*/`였다 — 저장소를 걸어다니는 정적 가드
(`internal/candidate/bandscale_test.go:462`)가 리뷰어 worktree 안의 `band.go`를
저장소 파일로 셌다. **리뷰어가 도는 동안에는 `make test`와 `make gate`를 믿을 수
없다.** a081 issues I6("리뷰어에게 각자 worktree를 줘라")의 대가가 이것이고, 다음에는
worktree를 저장소 밖에 두어야 한다.

## 2026-08-05 · 독립 리뷰 2 (테스트 렌즈, 별도 worktree) 과 그 반영

**결과: P0 0건, P1 6건, P2 5건.** 리뷰어가 변이 17종을 돌려 **8종이 살아남는 것**을
보였다. 다만 이 리뷰는 `56e85c68`(리뷰 1 반영 전)을 봤으므로, 절반은 mtime 확인과
함께 이미 사라졌다.

### 이미 사라진 것 — P1-3 (survivors 6종)

N2·N3·N9·N10·N11·N15는 전부 `cacheFileChanged`/`stampCacheFile`을 겨눈다. 리뷰 1
반영으로 그 메커니즘을 통째로 철회했으므로 **대상이 없다.** 리뷰어가 "N2/N3에서
매 요청 `os.ReadFile`+`Unmarshal`을 한다"고 지적한 것도 같은 이유로 사라졌다.

리뷰 둘이 독립적으로 같은 코드를 겨눴다는 것 자체가 신호였다.

### P1-1 — 테스트가 이름값을 못 했다 → **개정으로 해소, 실측 확인**

`TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot`가 `refresh()`를 **한
번도 부르지 않았다**. `token()`의 디스크 갈래가 요청을 만들기도 전에 픽스처를
덮어써서 401이 나지 않았다.

개정 후 `token()`은 base 그대로(메모리 우선)라 낡은 토큰이 실제로 제시되고 401을
받는다. `refresh` 첫 줄에 panic을 심어 확인했다.

```
TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot -> PROBE HIT
TestARefusedProcessWithNothingToAdoptStillExchanges       -> PROBE HIT
TestARotationThatLandsMidRequestCostsNoToken              -> PROBE HIT
TestAnAdoptedTokenThatIsAlsoRefusedStillEndsOnAMintedOne  -> PROBE HIT
```

D2 철회의 부수 효과다 — 테스트가 비로소 이름대로 동작한다.

### P1-4 · P1-5 — 채택 조건이 양방향으로 안 묶여 있었다 → 테스트 2건 추가

- **N5** 만료된 파일 토큰을 채택한다. `send()`의 유일한 재시도를 확실한 거부에
  쓴다. → `TestAdoptionRefusesAnExpiredFileToken`. **RED 확인.**
- **N6** 채택 조건을 토큰 동일성이 아니라 만료 시각으로 키잉한다. "다르면 채택"이
  좁음의 논거 전부인데 아무것도 그것을 토큰에 묶지 않았다. →
  `TestAdoptionKeysOnTheTokenItself`(같은 토큰 + 더 먼 만료 ⇒ 교환해야 한다).
  **RED 확인** (3건 동시).

### P1-6 — 감싼 `ErrAuth`로 소비자를 시험한 곳이 없었다 → 가장 중요한 지적

a082가 감싸는 근거는 "소비자 전부 `errors.Is`"였다. 그런데 **fixture는 전부 맨몸
sentinel**이라, 소비자가 unwrap을 멈춰도 아무도 몰랐다. 리뷰어가 소비자별로
`errors.Is`→`==` 변이를 돌려 확인했다.

| 소비자 | 변이 결과 (반영 전) | 반영 후 |
|---|---|---|
| `execgw/retry.go:60` | KILLED (다른 테스트가 감싼 fixture를 쓴다) | 유지 |
| `execgw/failclosed.go:210` | **SURVIVED** | `TestBrokerBranchesMapToStableReasonCodes`에 감싼 행 2개 → **KILLED** |
| `execgw/classify.go:111` (`statusOf`) | **SURVIVED** | `TestStatusOfReadsWrappedSentinels` 신규 → **KILLED** |

`failclosed.go`는 `ReasonBrokerAuthRejected`를 정하는 fail-closed 분류기이고
`statusOf`는 "확정 거부냐 모호하냐"를 정해 **종목 차단 여부**를 가른다. 둘 다
High-risk다.

`statusOf`는 unexported라 `export_status_test.go`로 shim을 냈다. 그 파일을 새로
만든 이유는 기존 `export_test.go`에 함수를 붙이면 바로 위 `init()`이 증거 요구
대상으로 끌려오기 때문이다 — [[tossos-logic-map-scope-creep]]가 또 맞았다.

### P2-1 — `==`→`errors.Is` 확대가 두 행에서는 **약화**였다 → 되돌렸다

issues I5가 "감싸지 않은 sentinel에는 `errors.Is(x,x) == (x==x)`"라고 썼다. 고정된
값에는 참이지만 **단정은 코드가 만드는 무엇에나 적용된다** — `errors.Is`는 sentinel을
감싼 모든 오류를 받아들인다. 리뷰어가 429/5xx에 본문을 실어 감싸는 변이(N17)로
보였다: 확대한 단정은 통과하고, 원래 `==`는 잡는다.

**auth 행만 넓히면 됐다.** 나머지 둘은 `==`로 되돌렸다 — 거기서 동일성은 "아무도
본문을 붙이기 시작하지 않았다"는 tripwire이기도 하고, 이 change가 401/403에 대해
`TestAnAuthRefusalDoesNotCarryTheResponseBody`로 지키려던 것과 같은 부류다.

### P2-4 — 자기를 끄는 단정

`codeIn`이 기본값으로 `""`를 돌려주고 `strings.Contains(s, "")`는 항상 참이다.
{401,403} 밖의 행을 더하는 순간 코드 단정이 조용히 통과한다. `t.Fatalf`로 바꿨다.

### 해소되었거나 남긴 것

- **P1-2**(branch-test-map이 근거를 못 주는 테스트를 인용) — 개정으로 맵을 다시
  썼고, M1이 실제로 깨는 3건만 인용한다. `TestARotationThatLandsMidRequestCostsNoToken`도
  이제 맵에 있다.
- **P2-2**(헤드라인 테스트를 단독 변이가 못 깬다) — 개정 후 M1 **단독으로 깨진다**
  (mtime 확인이 없어졌으므로). self-check가 느슨하다는 지적은 유효하나, 변이 증거가
  그 자리를 대신한다.
- **P2-3**(`token()`이 `m.cache`를 반환값과 맞춘다는 불변식이 안 묶임) — 개정으로
  **무의미해졌다.** `refused`가 인자라 `refresh`가 더 이상 `m.cache`에 기대지 않는다.
- **P2-5**(`make test`가 `internal/journal` 600초 타임아웃으로 붉어질 수 있다) —
  이 change와 무관한 기존 취약성이고, 부하에 따라 갈린다. issues I10.

### 리뷰어가 확인해 준 것

- 초안의 변이 4종이 전부 문서대로 재현된다. M4가 손대지 않은 `TestTokenRefresh`를
  깨는 것도 확인.
- `TestARotationThatLandsMidRequestCostsNoToken`은 주장하는 창을 실제로 연다 —
  회전이 핸들러 안에서, `token()`이 T1을 건넨 뒤 응답 전에 발생하고, 재시도는
  `token()`이 아니라 `refresh()`를 지난다. self-check도 진짜다.
- `classifyStatus`의 판정 보존: 항상 401 보고(N12), 본문 재삽입(N13) 둘 다 죽는다.
- 저장소 어디에도 `== Err*` sentinel 비교가 남아 있지 않다.

### 최종 관측

| | |
|---|---|
| `internal/official` + `internal/execgw` | **550 passed**, 0 failed |
| `-race` (`internal/official`) | **212 passed**, clean |
| `go vet` (전 패키지) / gofmt | clean |
| 기존 테스트 수정 | **5건** — 전부 이유를 issues에 (I5·I7·I11) |
| Function Logic Map | **9 target** `evidence complete` |
