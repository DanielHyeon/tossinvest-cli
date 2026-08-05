# a080 · Design

> **개정 2026-08-04.** 초안은 fragment endpoint + hash-pinned 스크립트로 부분
> 갱신하는 설계였다. 그것은 콘솔의 무-JavaScript 아키텍처를 위반하며, 근본 원인
> 조사 결과 스크립트를 버려도 두 번째 스펙 제약이 남는다는 것이 드러났다
> (issues.md I3·I4). 아래는 그 조사 뒤의 설계다. 철회된 결정은 D-old에 남긴다.

## D1 — 바꾸는 것은 기전이 아니라 주기다

두 화면은 이미 meta refresh로 스스로 다시 열린다. a080은 그 **주기의 출처**만
바꾼다.

```text
before   RefreshSeconds() = holdingsTTL / time.Second         (30초)
after    RefreshSeconds() = lineRefreshInterval / time.Second  (5초)
```

새 기전이 없다. 새 라우트, 새 템플릿, 새 클라이언트 코드, 새 주입 seam이 전부
0건이다. 이 change가 추가하는 것은 상수 하나와 그 근거다.

## D2 — 상한을 지키는 것은 주기가 아니라 캐시다

`rate budget 보호`가 재로드 주기를 캐시 TTL 이상으로 묶는 이유는 "열린 탭 하나의
비용 상한은 holdings 1콜/TTL"이다. 그 상한을 실제로 지키는 것은 캐시다.

```go
// internal/console/holdings.go
if !hold && (!c.attempted || now.Sub(c.tried) >= c.ttl) {
	c.refreshLocked(ctx, now)
}
```

`tried`는 `refreshLocked` 진입에서 갱신되므로 브로커 도달은 TTL당 최대 1회이며
재로드 주기와 무관하다. 주기를 6배로 올리면 **캐시** 도달이 6배가 되고 브로커
도달은 그대로다.

따라서 MODIFY는 요구사항을 약화하지 않는다. 상한을 **주기 조건으로 대리 표현하던
것을 상한 자체로 바꾸고**, 그 상한을 누가 지키는지 명시한다. 나머지 절은 글자
그대로 유지된다.

`/dashboard`는 더 강하다 — `peek`만 쓰므로 어떤 주기에서도 0콜이다
(`console-operator-overview` D4, 화면에 문구로 쓰여 있다).

## D3 — 5초는 콘솔의 상수이고, 엔진 패키지를 import하지 않는다

주기의 근거는 `engine.DefaultExitObservationInterval`이지만 콘솔이 그 상수를
import하지는 않는다. 콘솔은 엔진과 별도 프로세스이고, 엔진 런타임 패키지에 대한
컴파일 의존은 "콘솔은 브로커·journal writer·엔진 컨트롤러를 받지 않는다"는
경계를 넓히는 첫 걸음이다.

두 값이 갈라지면 화면이 엔진보다 느려질 뿐 틀리지는 않는다 — fail-safe 방향이다.

## D4 — 전체 재로드의 비용은 이 두 화면에서 이미 지불되어 있다

초안이 부분 갱신을 택한 이유는 "5초 전체 리로드가 스크롤과 조작 상태를 날린다"
였다. 이 두 화면에 한해 그 전제를 하나씩 확인했다.

| 잃을 것 | 실제 | 근거 |
|---|---|---|
| 열린 접힘 | 안 잃는다 | 재로드 화면의 접힘 상태는 URL에 있다 (a055 §6) |
| 편집 중인 폼 | 없다 | 두 화면에 form·input·button이 아예 없다 — a057이 `/position-management`로 옮겼고 `TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript`가 고정한다 |
| 브로커 예산 | 안 쓴다 | D2 |
| 스크롤 위치 | 브라우저 복원에 의존 | **미검증** — task 6.8 실측 항목 |

앞의 셋은 코드와 테스트로 확인했다. 네 번째만 브라우저 동작이라 실측이 필요하고,
그것이 이 설계에서 유일하게 열린 위험이다. 나쁘게 나오면 그때가 스크립트 예외를
**증거와 함께** 제안할 자리이지, 지금이 아니다.

이 세 전제는 스펙에 ADDED로 적는다 — 하나라도 무너지면 주기를 올린 결정이 함께
재검토되어야 하기 때문이다.

## D5 — 무-스크립트는 이 change가 지켜야 할 아키텍처다

`archive/2026-07-31-streamline-trading-views/design.md`의 Non-Goals가 "JavaScript,
CSP nonce, `unsafe-inline` script, 외부 CSS/폰트 도입"을 명시하고, 세 change가
테스트로 고정하며, `operator-console` spec이 CSP 회귀 scenario로 뒷받침한다.

`script-src`를 비워 두는 것은 스크립트를 안 쓰기 위해서가 아니라 **템플릿에
inline handler가 섞여 들어와도 반드시 죽어 있게** 만들기 위한 장치다. 그 장치는
이 change가 스크립트를 쓰지 않는다는 사실과 별개로 유지되어야 한다.

a080은 그 자세를 성문화하는 ADDED 요구사항과 `TestTheLineScreensStillShipNoScript`를
남긴다. 다음에 이 화면을 빠르게 만들려는 시도가 세 개의 실패 테스트에서 이유를
역추적하지 않도록.

## D-old — 철회된 결정

| 결정 | 내용 | 철회 사유 |
|---|---|---|
| D1-old | fragment는 서버 렌더 HTML | fragment 자체가 없어졌다 |
| D2-old | `/positions`는 `get`, `/dashboard`는 `peek` | 유효하지만 fragment 없이 기존 핸들러가 이미 그렇다 |
| D4b-old | 스크립트가 스트립 reload 셀을 다시 쓴다 | 스트립은 `RefreshSeconds`를 그대로 읽고 값이 하나뿐이라 불일치가 생기지 않는다 |
| D5-old | 두 화면에 sha256 `script-src` | issues.md I3 — 아키텍처 위반 |
| D7-old | 연속 실패 시 전체 리로드 폴백 | 폴백이 곧 본체가 되었다 |

## Function Logic Map

| 함수 | 편집 | Logic Map |
|---|---|---|
| `positionsPage.RefreshSeconds` | 반환식의 출처 상수 교체 + 잘못된 주석 정정 | 작성 완료 |
| `overviewPage.RefreshSeconds` | 반환식의 출처 상수 교체 + 주석 갱신 | 작성 완료 |

초안 설계에서 잡혔던 `Console.routes`·`handlePositions`·`handleOverview`는 그
편집이 철회되어 대상에서 빠진다. 두 함수 모두 High-risk 경로가 아니다.

## 위험도

**Normal.** 주문·위험·원장 경로를 건드리지 않고, 새 공식 API 호출이 0건이며,
새 클라이언트 코드가 0줄이다.

`rate budget 보호`가 Requirement 수준으로 수정되므로 WORKFLOW.md §142에 따라
리뷰를 재실행한다. 검토 대상은 셋이다 — 상한이 여전히 지켜지는가, 그 상한을
지키는 주체가 실제로 캐시인가, 나머지 절이 글자 그대로 보존되었는가.
