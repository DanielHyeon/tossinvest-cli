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
