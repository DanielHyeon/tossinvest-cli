# a080 · Review

## 2026-08-04 · proposal-freeze

- **등급**: Normal (WORKFLOW.md 위험도 분류). operator-console 표시 경로이며
  주문·손절·사이징·Guardian·원장·reconciliation·인증·체결 감지 어디에도 닿지
  않는다. WORKFLOW.md §리뷰 게이트의 "UI·문서·도구 change는 경량 리뷰로
  충분하다"에 해당한다.
- **보이스**: Manager 셀프리뷰 + 적대적 보안 관점 (design 말미가 요구).
- **한계 명시**: 이 리뷰는 proposal을 쓴 것과 **같은 컨텍스트**에서 수행했다.
  WORKFLOW.md §역할 분리는 작성자와 검증자의 분리를 요구하며, 이 셀프리뷰는
  그것을 대체하지 않는다. 구현 완료 후 별도 컨텍스트의 독립 리뷰가 task 7.1로
  남아 있고, `make gate` 전에 반드시 수행한다.
- **입력**: `openspec validate --strict` 통과, `check_analysis.py` 통과,
  `internal/console` 베이스라인 699 passed.

### R1 — "한 호출도 늘리지 않는다"는 과한 주장이다 · **수용, 문구·테스트 수정**

proposal 초안이 rate budget 영향을 "**한 호출도** 늘리지 않는다"로 썼다.
적대적으로 세어 보면 그것보다 약하다.

- meta refresh는 페이지 **로드 완료 후** 30초에 발화한다. 따라서 오늘의 실제
  브로커 호출 간격은 30초 + 로드·렌더 시간으로 30초를 **약간 넘는다**.
- 5초 폴링은 TTL 만료 직후 첫 tick에 refresh를 잡으므로 간격이 30~35초로
  수렴한다.

즉 호출 rate가 미세하게 **오를 수 있다**(대략 30초당 1회를 향해 조여진다).
늘지 않는 것이 아니라 **TTL이 정한 상한을 넘지 않는** 것이다. TTL이 곧 예산
자체이므로 §0.4 위반은 아니지만, 주장은 사실에 맞춰야 한다.

**조치**: proposal Impact 문구를 "상한을 넘지 않는다"로 고친다. task 2.1의
판정도 "오늘과 호출 수가 같다"가 아니라 "창 길이 ÷ TTL의 올림을 넘지 않는다"로
쓴다 — 후자가 실제로 지켜야 하는 성질이고, 전자는 타이밍에 따라 flaky하다.

### R2 — 열린 탭 하나가 journal read를 6배로 만든다 · **수용, 측정 항목 추가**

부분 갱신은 5초마다 `livePositions` + `decoratePositionRows`를 돈다. 탭 하나를
하루 열어 두면 오늘 2,880회이던 것이 17,280회가 된다. 같은 기계에서 엔진이
5초마다 journal에 쓴다.

완화 근거는 있다 — journal은 WAL이고 `readonly.go:104`가 "A reader in WAL mode
is not normally blocked by a [writer]"를 명시한다. 따라서 쓰기 차단은 설계상
발생하지 않는다.

그래도 **읽기 볼륨 자체는 실측 대상**이다. 차단자는 아니지만 "아마 괜찮다"로
넘길 항목도 아니다.

**조치**: task 6.8에 열린 탭 1개 기준 journal read 지연과 엔진 사이클 시간을
함께 보는 항목을 더한다.

### R3 — a055 §6의 fold 계약이 부분 갱신 경계와 충돌할 수 있다 · **수용, task 추가**

`portfolio_pages.go:66-68`은 "a reloading screen cannot keep a native fold open
(change a055 §6)"을 이유로 URL 기반 `Explain` 상태를 쓴다. 스크립트가 meta
태그를 제거하면 전체 리로드가 멈추므로 그 전제가 바뀐다.

- URL 기반 상태는 계속 정확하다 → **회귀는 없다**.
- 그러나 fold가 **교체 구간 안에** 있으면 5초마다 닫힌다. 오늘 30초마다 닫히던
  것이 5초마다 닫히는 것이고, 이는 개선이 아니라 **악화**다.

**조치**: fragment 경계는 fold를 포함하지 않는다. task 3.11로 고정한다.

### R4 — fragment를 직접 열면 CSP 없는 HTML이 뜬다 · **수용, 설계 제약으로 고정**

fragment는 `text/html`이다. 운영자가 주소를 직접 열거나 리다이렉트로 도달하면
브라우저가 그것을 페이지로 렌더한다. 전용 writer를 새로 만들면 그 응답에
CSP·`Cache-Control`이 빠진다.

이미 있는 경로가 그 셋을 한꺼번에 해결한다 — `renderHTML`이
`Content-Type`, `Cache-Control: no-store`(`pages.go:395`), CSP를 모두 설정한다.

**조치**: fragment는 반드시 `renderHTML`을 경유한다(전용 `w.Write` 금지).
`no-store`는 부수효과가 아니라 **필수**다 — 5초마다 폴링하는 응답이 캐시되면
화면이 얼어붙고, 그것이 이 change의 목적을 조용히 무효화한다. task 2.8로 고정.

### R5 — CSP 확대 범위 · **설계대로 수용**

콘솔에 처음으로 상시 실행 스크립트와 `connect-src`가 들어온다. 세 가지를 확인했다.

1. 범위가 두 화면을 넘지 않는가 → `consoleHTMLCSP` 자체는 무변경이고 두 페이지
   핸들러만 확장 문자열을 고른다. task 2.5가 다른 화면의 헤더 동일성을 고정한다.
2. 해시가 상수에서 파생되는가 → `optimization_view.go:140-143` 선례를 그대로
   따른다. task 2.7이 손 관리 목록이 아님을 고정한다.
3. fragment가 세션 게이트 안인가 → task 2.3.

`connect-src 'self'`는 fetch 대상이 같은 origin 하나뿐이므로 최소 범위다.
`default-src 'none'`이 유지되므로 이미지·폰트·프레임은 여전히 전면 차단이다.
**수용.**

### 거절한 대안

- **JSON fragment + 클라이언트 렌더**: `operatorview`가 단독 소유한 fail-closed
  표기 규칙이 JavaScript에 복제된다. a067이 한곳에 모은 것을 다시 가른다. 거절.
- **`holdingsTTL`을 5초로 축소**: 브로커 호출을 6배로 만든다. 이 change의 주장은
  "표시를 빠르게"이지 "브로커에 자주 묻자"가 아니다. 거절.
- **meta 태그를 템플릿에서 빼고 스크립트가 삽입**: 스크립트가 막힌 클라이언트가
  영원히 갱신되지 않는다 — 오늘보다 나쁘다. 거절 (design D4).
- **스트립에 "5초 (스크립트 없으면 30초)" 조건부 문구**: 절반의 경우 항상 틀린
  문장이며 `chrome.go:77-79`가 막으려던 형태 그 자체다. 거절 (design D4b).

### 결론

**proposal-freeze 통과.** R1은 문구·판정 기준을 고치고, R2·R3·R4는 task로
추가한 뒤 구현에 착수한다. R5는 설계대로 진행한다.

구현 완료 후 **별도 컨텍스트의 독립 리뷰가 남아 있다**(task 7.1). 이 셀프리뷰는
그것을 대체하지 않는다.

---

## 2026-08-04 · Requirement 변경 리뷰 (WORKFLOW.md §142)

`rate budget 보호`를 MODIFY하므로 재실행한다. 1차 리뷰는 fragment + 스크립트
설계를 대상으로 했고 그 설계는 철회되었다(issues.md I3·I4).

- **등급**: Normal 유지. 주문·위험·원장 경로 무접촉, 새 공식 API 호출 0건, 새
  클라이언트 코드 0줄.
- **한계 명시**: 1차와 같이 작성 컨텍스트의 셀프리뷰다. 별도 컨텍스트의 독립
  리뷰는 task 7.5로 남아 있으며 `make gate` 전에 수행한다.

### Q1 — 상한이 여전히 지켜지는가 · **예, 측정과 변이로 확인**

`TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling`: TTL 4주기(2분) 동안 5초마다
24회 재로드 → 브로커 호출 5회 이내. 테스트 자신이 "재로드 횟수 ≤ 허용 호출 수"이면
`t.Fatal`로 자기를 무효 선언하므로 무의미하게 통과할 수 없다.

변이 6.2에서 `holdingsCache.get`의 TTL gate를 제거하자 **5회 허용에 24회 관측**으로
RED. 이 테스트는 실제로 load-bearing이다.

### Q2 — 상한을 지키는 주체가 실제로 캐시인가 · **예**

`tried`는 `refreshLocked` 진입에서 갱신되고 gate는 `now.Sub(c.tried) >= c.ttl`이다.
주기는 이 조건에 등장하지 않는다. Q1의 변이가 그 인과를 직접 보인다 — gate를
없애자 주기가 곧바로 호출 수가 되었다.

### Q3 — 나머지 절이 글자 그대로 보존되었는가 · **예**

MODIFIED 본문에서 유지된 것: 서버측 폴러 부재(SHALL NOT), 갱신 1회 = holdings 1콜,
심볼별 fan-out 금지, TTL 15초 하한, TTL 내 재요청·다중 탭의 무호출, 캐시 시각 표시,
검증 중 보류의 두 판정 경로(in-process 신호 / runlock mtime 5분 상한), 자동 재로드가
그 보류를 우회하지 못한다는 SHALL.

바뀐 것은 한 절뿐이다 — "주기는 캐시 TTL 이상"에서 "상한은 캐시가 지키고 주기는
TTL과 독립"으로. scenario는 3건이 늘었고(비용 상한, peek 화면, 주기의 출처) 기존
2건은 유지된다.

### Q4 — MODIFIED를 쓰는 것이 옳은가 · **예, 대안이 더 나쁘다**

a055 `issues.md` I1의 미아카이브 MODIFY 부채를 늘린다는 비용은 실재한다. 그러나
기존 SHALL이 "주기 ≥ TTL"을 명시하는 한, ADDED로 "주기는 TTL과 독립"을 더하면
**스펙이 자기모순**에 빠진다. 모순되는 ADDED가 MODIFIED 하나보다 나쁘다.

### R6 (신규) — 스크롤 위치는 유일한 미검증 전제 · **수용, 실측으로 이관**

전체 재로드가 이 두 화면에서 감당 가능한 근거 넷 중 셋은 코드·테스트로 확인했다
(접힘은 URL, 폼 없음, 예산은 캐시). 네 번째인 스크롤 복원은 브라우저 동작이고
검증하지 않았다.

**조치**: task 6.5의 실측 항목으로 명시했다. 나쁘게 나오면 그때가 스크립트 예외를
증거와 함께 제안할 자리이지 지금이 아니다 — 그 순서를 뒤집은 것이 초안의 오류였다.

### R2 재확인 — journal read 볼륨

읽기 볼륨 6배는 부분 갱신이든 전체 재로드든 동일하게 발생한다. WAL이라 쓰기
차단은 설계상 없고(`readonly.go:104`), 실측 항목으로 task 6.6에 남아 있다.

### 1차 리뷰 항목의 처분

| 1차 | 처분 |
|---|---|
| R1 상한 문구 과장 | 반영됨 — proposal·spec·test 전부 "상한" 표현으로 통일 |
| R2 journal 볼륨 | 유지 (task 6.6) |
| R3 fold 경계 | **소멸** — fragment가 없어졌고 접힘은 URL이 보존한다 |
| R4 fragment CSP/no-store | **소멸** — fragment 없음 |
| R5 CSP 확대 범위 | **철회** — 확대 자체를 하지 않는다 (I3) |

### 결론

**Requirement 변경 통과.** R6을 실측 항목으로 두고 진행한다. 별도 컨텍스트의
독립 리뷰(task 7.5)는 여전히 남아 있다.

---

## 2026-08-04 · 독립 리뷰 (별도 컨텍스트 3인, task 7.5)

작성 컨텍스트가 아닌 별도 컨텍스트 3개가 각각 spec 계약 / 테스트 커버리지 /
적대적 실패 관점으로 검토했다. **결과: 이 change는 현 상태로 배포 가능하지 않다.**

### F1 — **BLOCKING (P0)** · 재로드는 엔진 프로세스로 들어가는 무캐시 RPC를 6배로 만든다

`decoratePositionRows`는 렌더마다 엔진에 **동기 RPC 2회**를 한다. TTL도 캐시도 없다.

```go
// internal/console/portfolio_pages.go:100-101
runtime, _ = c.opts.PositionPolicies.Runtime(ctx)
if states, err := c.opts.PositionPolicies.List(ctx); err == nil {
```

그 `List`는 엔진 쪽에서 이렇게 처리된다.

```go
// internal/app/engine/position_policy_command.go:138-142
func (s *PositionPolicyCommandService) List(ctx context.Context) ([]positionpolicy.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.j.PositionPolicies(ctx)
}
```

`s.j`는 **엔진이 소유한 쓰기 핸들**이고 그 핸들은 커넥션이 하나다.

```go
// internal/journal/journal.go:151
db.SetMaxOpenConns(1)
```

즉 콘솔의 `List()`는 exit 루프의 `RecordExitJudgementResult`(BEGIN IMMEDIATE +
`synchronous=full`)와 **같은 커넥션·같은 뮤텍스에서 직렬화**된다. 그리고 exit
루프는 ticker가 아니라 `작업 후 5초 sleep`이므로(`exitloop.go:350-357`), 그
커넥션에서 뺏은 시간은 판정 간격에 **그대로 더해진다**.

`.claude/CLAUDE.md` 안전 불변식 4 — "손절·비상 청산의 즉시성을 약화하거나
지연하지 않는다" — 에 정면으로 걸린다.

**왜 내 테스트가 못 잡았나**: `newDashboardHarness`가 `o.PositionPolicies`를
배선하지 않는다(`portfolio_test.go:246-251`에 없다). 그래서
`TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling`은 이 경로를 **한 번도
실행하지 않는다**. "6배로 늘어나는 것은 캐시 mutex와 템플릿 렌더뿐"이라는
issues.md I1의 서술은 틀렸다 — 실제로는 엔진 RPC 2회, SQLite read-only 오픈
(스키마 재검증 ~40 statement), config 파일 파싱이 함께 6배가 된다.

부수 효과: `List` 실패 시 `policyByID`가 nil로 남아
`exit_line_reference.go:99`가 모든 관리 행을 "관리 세대 확인 불가"로 닫는다.
그 RPC의 클라이언트 timeout이 정확히 5초 — 새 재로드 주기와 같다.

### F2 — (P2) `consoleScreens`를 5로 바꾸면서 스트립 검증이 눈이 멀었다

`"5초마다"`는 `"15초마다"`의 **부분 문자열**이고, 표의 나머지 비-0 항목이 전부
15초(`ordersTTL`, `candidate.DefaultWatchInterval`)다.

```go
// status_strip_test.go:203
if on && !strings.Contains(cell, strconv.Itoa(screen.reload)+"초마다") {
```

리뷰어가 변이를 양방향으로 실행했다: 템플릿의 `{{.RefreshSeconds}}초마다`를
`15초마다` 고정으로 바꾸면 **HEAD에서는 FAIL, a080 적용 후에는 696 passed**.
tasks.md 3.3의 "green 유지 = 성립"이라는 판정 근거가 틀렸다 — green 유지는
성질이 아니고, 그 테스트의 일은 출처가 하나가 아니게 된 순간을 잡는 것이다.

### F3 — (P1) a054의 승인된 scenario와 충돌하는데 MODIFY하지 않았다

`operator-console` Requirement `콘솔 공통 상태 표시줄`:

> **THEN** 검증 화면은 run이 작업 중일 때만 짧은 주기를 쓰고, **캐시 기반 화면은
> 각자의 캐시 TTL을 그대로 쓴다**

`/positions`·`/dashboard` 둘 다 캐시 기반이고 둘 다 더 이상 TTL을 쓰지 않는다.
a080 delta는 `rate budget 보호` 하나만 MODIFY한다. 두 번째 MODIFIED가 필요하다.

### F4 — (P1) MODIFIED가 승인된 scenario 2건을 조용히 떨어뜨렸다

`검증 실행 중 — 캐시 있음`과 `— 콜드 캐시`가 delta에 없다. delta 머리말은
"검증 중 보류와 그 우회 금지는 전부 글자 그대로 유지된다"고 적었는데, 산문은
맞지만 scenario는 아니다. 특히 콜드 캐시의 `검증 중 — 데이터 없음` 문구와
journal 측 표시 보장이 사라진다.

### F5 — (P1) Function Logic Map이 **철회된 설계**를 서술하고 있다

두 `RefreshSeconds`의 `branch-test-map.md`가 아직 "폴백 상수 30", "스크립트
미설치 시 스트립 30초마다", "스크립트 설치 시 meta 태그 없음"을 적고 있고,
존재하지 않는 task 3.8·3.9·3.10을 참조한다. AST는 재생성했지만 map 본문은
갱신하지 않았다. gate는 산출물 **존재**를 검사하지 실질을 검사하지 않으므로
통과했다 — 즉 gate가 잡아야 할 종류의 stale 증거다.

### F6 — (P2) 파일 머리말 주석 2건이 a080이 없앤 결합을 아직 주장한다

```go
// portfolio_pages.go:12-16
// The positions screen reloads itself at the holdings cache TTL — no faster
// holdings.go:25-29
//            or by the positions screen re-opening itself at the TTL
```
issues.md I1이 지목한 결함(코드와 모순되는 주석이 몇 달간 예산 거부 근거로
읽혔다)과 **똑같은 것을 두 개 더 남긴다**. 하나는 §0.4 rate budget 회계
블록 안이다.

### F7 — (P2) `TestTheOverviewSpendsNothingHoweverOftenItReloads`는 이름값을 못 한다

`peek`은 `now`를 갱신 판단에 쓰지 않으므로 clock advance가 장식이다. 시계를
멈춰도, 1회만 돌려도 0이 나온다. 기존 `TestTheOverviewMakesNoBrokerCall`이
캐시를 데운 뒤 `2*holdingsTTL`을 세 번 진행시키는 더 강한 테스트로 이미 있다.

### F8 — (P2) `STORY-TOS-a080`의 승인 기준이 철회된 설계를 아직 서술한다

`docs/pm/portfolio/stories/STORY-TOS-a080.yaml`의 승인 기준 5·6·7이 부분 갱신
스크립트 설계의 것이다.

```text
"A client where the partial refresh cannot run keeps today's full-page reload…"
"A failed partial refresh keeps the last displayed values…"
"The relaxed script-src and connect-src apply only to the two screens…"
```

issues.md I3·I4로 철회된 설계이고 `design.md` D5·D-old가 그 철회를 기록한다.
PM 계층만 옛 설계를 승인 기준으로 들고 있다. F2~F7과 같은 batch에서 정리한다.

### 확인된 것 (반증 없음)

- 브로커 상한 주장 자체는 참이다. `tried`가 실패 호출에도 찍히고, 콜드 캐시는
  1회만 발화하며, 뮤텍스가 fetch를 감싸고, `hold`는 크레딧을 쌓지 않는다.
- `/dashboard`는 진짜 peek 전용이다.
- 무-스크립트 가드 3건은 손대지 않고 통과한다.

### 결론

**BLOCKED.** F1은 이 change의 전제("6배로 늘어도 비용은 그대로")를 무효화하고
안전 불변식 4에 닿는다. F2~F8은 F1의 처분이 정해진 뒤에 정리한다.

### F1 처분 (2026-08-04, 사용자 결정)

**선행 change로 분리한다.** `a081-screens-share-one-engine-reading`이 콘솔의
엔진 정책 읽기를 TTL 캐시 뒤로 옮겨, 엔진 도달 횟수를 **렌더 횟수에서
분리**한다. a081이 land하면 F1의 기전 자체가 사라지고 — 재로드를 6배로 해도
엔진 읽기는 오늘과 같거나 적다 — a080은 상수 하나를 바꾸는 change로 줄어든다.

거부한 대안 둘:

- **15초 절충.** 원인을 고치지 않고 배수를 6에서 2로 줄일 뿐이다. "얼마면
  괜찮은가"라는 답 없는 질문을 남기고, 다음에 주기를 다시 만지는 사람이 같은
  계산을 처음부터 해야 한다.
- **a080 철회.** F1은 a080이 만든 결함이 아니라 a080이 **드러낸** 기존 결함이다.
  오늘도 두 탭이 열려 있으면 30초마다 엔진 읽기 2벌이 들어간다. 되돌려도 그
  비용은 남는다.

a080은 a081이 land하고 그 예산 테스트가 이 경로를 실제로 실행한 것이 확인된
뒤에 재개한다. F2~F8은 그 재개 batch에서 정리한다.

### F1~F8 처리 현황 (2026-08-05)

| # | 상태 | 무엇을 했나 |
|---|---|---|
| F1 | **해소 대기** | 선행 change `a081-screens-share-one-engine-reading` 구현 완료. a081의 독립 리뷰가 끝나고 land한 뒤 a080 재개 |
| F2 | 미착수 (코드) | `status_strip_test.go`의 부분 문자열 판정. a080 코드 재적용과 같은 batch |
| F3 | **완료** | `콘솔 공통 상태 표시줄`을 두 번째 MODIFIED로 추가하고 scenario `화면별 재로드 주기 보존`의 THEN을 고쳤다. 나머지 절과 scenario 9건은 글자 그대로 보존. `openspec validate --strict` 통과 |
| F4 | **완료** | 떨어졌던 `검증 실행 중 — 캐시 있음`·`— 콜드 캐시` 두 scenario를 delta에 복원 |
| F5 | **완료** | 두 `RefreshSeconds`의 branch-test-map을 다시 썼다. 없는 task 3.8·3.9·3.10 참조 제거, "폴백 상수·값 30 유지" 서술을 정석 수정 내용으로 교체, 엔진 부하 계약(C7)을 새로 세움. function-logic-map 본문의 같은 서술도 갱신 |
| F6 | 미착수 (코드) | `portfolio_pages.go`·`holdings.go` 파일 머리말 주석. 코드 재적용과 같은 batch |
| F7 | 미착수 (테스트) | `TestTheOverviewSpendsNothingHoweverOftenItReloads`의 장식용 clock advance. 코드 재적용과 같은 batch |
| F8 | **완료** | `STORY-TOS-a080.yaml` 승인 기준 5·6·7을 철회된 스크립트 설계에서 현재 설계로 교체 |

F2·F6·F7을 지금 하지 않는 이유는 하나다. a080의 코드 변경은 되돌려 둔 상태이고
(a081이 같은 파일을 고치는 동안 두 change의 diff가 섞이면 logic-map gate가 서로를
상대 change의 편집으로 잡는다), 그 셋은 전부 되돌린 코드 위의 수정이다. a081이
land하면 a080 코드를 재적용하면서 같은 batch에서 처리한다.

**재개 시 반드시 할 것**: a080의 `base-commit.txt`를 a081 land 후 커밋으로 다시
고정하고 (`capture_change_base.py`는 덮어쓰기를 거부하므로 파일을 지우고 재실행),
네 target의 AST를 재생성한다.

## 2026-08-05 · 재개 batch와 두 번째 MODIFIED의 Requirement 변경 리뷰

a081이 `30d8bb93`으로 land·배포됐고 `[blocked]`를 해제했다. 이 절은 8장의 실행
기록과, 두 번째 MODIFIED(`콘솔 공통 상태 표시줄`)에 대한 Requirement 변경 리뷰다.

### 코드 재적용 (8.5)

보관해 둔 패치를 되돌렸다. 새 파일 둘(`line_cadence.go`,
`line_cadence_test.go`)과 기존 다섯 파일의 편집이며, a081이 같은 패키지를 고친
뒤에도 충돌 없이 들어갔다 — a081은 `decoratePositionRows`를, a080은
`RefreshSeconds`를 고치므로 같은 파일의 다른 자리다. 패키지 **711건 통과**
(a081 시점 706 + a080의 5건).

### F2 — 부분 문자열에 눈이 먼 단정 (8.6)

`status_strip_test.go`가 재로드 셀을 이렇게 검사하고 있었다.

```go
if on && !strings.Contains(cell, strconv.Itoa(screen.reload)+"초마다") {
```

**`strings.Contains("15초마다", "5초마다")`는 참이다.** a080이 두 라인 화면을
`5초마다`로 바꾼 순간, `/signals`의 `15초마다`를 그린 셀이 `screen.reload == 5`를
만족하게 됐다. 리뷰어가 템플릿을 `15초마다`로 고정하는 변이를 돌렸고 통과했다.

셀이 말하는 숫자를 읽어 비교하도록 고쳤다(`reloadPeriodIn`). 같은 변이를 재현해
**RED를 실측**했다 — `/dashboard`·`/positions` 두 화면에서
`the strip says 15s and the meta tag uses 5s`. 기록은
`analysis/function-logic/internal-console--testthereloadcellandthemetatagareonefact/branch-test-map.md`.

a080 이전에는 라인 화면이 `30초마다`였고 테이블의 어떤 값도 다른 값의 접미사가
아니었다. **주기를 바꾼 것이 이 결함을 도달 가능하게 만들었다.**

`reloadPeriodIn`은 `cellOf` 옆이 아니라 `line_cadence_test.go`에 뒀다. 기존 파일에
새 함수를 넣으면 logic-map이 바로 위 함수(`cellOf`)까지 증거 대상으로 잡는데,
`cellOf`는 이 change와 무관한 표시줄 테스트 전부가 쓰는 헬퍼다. 실제로 처음
`status_strip_test.go`에 넣었을 때 checker가 `cellOf`를 요구했고, 옮기자 사라졌다.

### F6 — 없어진 결합을 계속 말하던 주석 (8.7)

`portfolio_pages.go`와 `holdings.go`의 파일 머리말이 "reloads itself at the
holdings cache TTL — no faster"를 근거로 예산을 설명하고 있었다. a080이 그 근거를
옮겼는데 주석은 그대로였다 — issues.md I1이 지목한 것과 같은 종류의 거짓이다.
둘 다 "상한은 캐시가 쥐고 있고 재로드 주기는 그것을 결정하지 않는다"로 고쳤다.

### F7 — 장식용 clock advance (8.8)

`TestTheOverviewSpendsNothingHoweverOftenItReloads`의 `h.clock.advance`가 판정에
기여하지 않았다. 걷어내는 대신 **하중을 받게** 만들었다: 재로드 창이 캐시 TTL보다
길지 않으면 `t.Fatalf`(그보다 짧으면 get 기반 화면도 갱신하지 않아 peek과 구별할 수
없다), 그리고 마지막에 get으로 읽는 `/positions`를 한 번 렌더해 **같은 카운터가
실제로 움직이는지** 확인한다. 그 자체 검사가 없으면 아무 데도 연결되지 않은
카운터가 0을 반환해도 통과한다.

### Requirement 변경 리뷰 — `콘솔 공통 상태 표시줄` (8.11)

base(`30d8bb93`) 대비 실제 수정분은 두 곳이다.

1. 톤 임계 문단에 한 문장 추가 — "톤의 근거인 캐시 TTL과 그 화면의 재로드 주기는
   별개의 값이다(SHALL — 재로드가 잦아지는 것은 캐시 도달 빈도만 바꾸고 톤 임계를
   바꾸지 않는다)."
2. scenario `화면별 재로드 주기 보존`의 THEN — "캐시 기반 화면은 각자의 캐시 TTL을
   그대로 쓴다" → "나머지 화면은 각자 자기 주기의 출처를 그대로 유지한다 — 표시줄은
   그 주기를 말할 뿐 정하지 않는다".

나머지 절과 scenario 9건은 글자 그대로 보존했다.

**검토 1 — 표시줄이 주기를 정하지 않고 말하기만 한다는 성질이 보존되는가.**
보존된다. 그 SHALL은 base에 이미 있고("표시줄은 지금 걸려 있는 주기를 말할
뿐이며") a080은 건드리지 않았다. 코드에서도 `templates.go:321`의
`content="{{.RefreshSeconds}}"`와 표시줄 셀이 **같은 값**을 읽으며, 그것을
`TestTheReloadCellAndTheMetaTagAreOneFact`가 고정한다 — F2 수정으로 그 테스트가
비로소 실제로 고정한다. scenario THEN의 수정은 이 성질을 약화한 것이 아니라,
"캐시 기반 화면 = 캐시 TTL"이라는 **더 이상 참이 아닌 등식**을 지우고 원래 성질만
남긴 것이다.

**검토 2 — 톤 임계의 근거인 캐시 TTL이 재로드 주기와 분리되는가.**
구조적으로 분리되어 있다. `freshness.tone()`(chrome.go:150)은 `f.TTL`만 읽고
`RefreshSeconds`에 도달하지 않는다. `f.TTL`을 채우는 곳은
`holdingsSnapshot.freshness()`(chrome.go:268)의 `TTL: holdingsTTL` 한 곳이며 a080이
바꾸지 않았다. 즉 추가한 SHALL은 새 제약이 아니라 **이미 참인 성질을 명문화**한
것이고, a080이 두 값을 다르게 만든 지금이 그것을 적을 자리다.

**함께 확인한 것.** scenario `표시줄이 브로커 비용을 늘리지 않는다`("TTL 안에서
여러 번 재로드해도 holdings 1콜 상한")는 수정하지 않았다. a080 이전에는 자동
재로드 주기가 TTL과 같아 이 scenario를 만족시키려면 사람이 손으로 새로고침해야
했다. 이제 자동 재로드가 TTL 안에 여섯 번 들어가므로 **이 scenario가 기본 동작으로
실행된다**. `TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling`이 그것을 직접
단언한다.

**기록해 둘 귀결 하나 (결함 아님).** 개요 화면은 peek으로 읽어 스스로 갱신하지
않으므로, `/dashboard`만 열어 둔 운영자에게는 경과가 계속 자란다. a080 이후 그
사실이 5초마다 갱신되어 보이고, 이전에는 최대 30초 늦게 보였다. 톤 임계는 그대로다
— 바뀐 것은 **표시가 진실을 따라잡는 속도**이며 방향은 정직해지는 쪽이다.

**판정: 수용.** 두 수정 모두 요구사항을 넓히지 않고, 하나는 이미 참인 성질의
명문화이고 다른 하나는 더 이상 참이 아닌 등식의 제거다. 새 SHALL·SHALL NOT은
추가되지 않았다(추가된 문장은 기존 SHALL NOT 문단 안의 부연이다).

### F1~F8 최종 처리 현황

| 발견 | 처리 | 근거 |
|---|---|---|
| F1 | **완료** | 선행 change a081이 결합을 제거하고 land·배포됨 (`30d8bb93`) |
| F2 | **완료** | 8.6 — 단정을 값 비교로, 변이 M-F2 RED 실측 |
| F3 | 완료 | 두 번째 MODIFIED 추가 (`15d25f80`) |
| F4 | 완료 | 떨어진 scenario 2건 복원 (`15d25f80`) |
| F5 | 완료 | logic-map에서 철회된 스크립트 설계 서술 제거 (`15d25f80`) |
| F6 | **완료** | 8.7 — 두 파일 머리말 주석 |
| F7 | **완료** | 8.8 — clock advance를 하중 받게 + 카운터 자체 검사 |
| F8 | 완료 | STORY 승인 기준 교체 (`15d25f80`) |

### 증거 재기준화

`base-commit.txt`를 `840b3377` → `30d8bb93`으로 재고정했다. logic-map target은
넷에서 **다섯**으로 늘었다 — F2 수정이
`status_strip_test.go:TestTheReloadCellAndTheMetaTagAreOneFact`를 수정된 기존
함수로 만들었다. `check_analysis.py` `evidence complete`.

## 2026-08-05 · 배포와 컨테이너 실측 (task 6.5·6.6)

사람이 승인한 배포다. `aaa7638d`가 `main`에 land했고 그 코드로 만든
`tossos@sha256:fafe53a3a15fecf32e44bb3c055deb0e22251292bc455e8476864eafcae2f16a`
를 `httpapi` → `tossos` 순서로 하나씩 교체했다. 두 service 모두 healthy, 엔진은
네 루프를 모두 올렸다. journal schema는 전후 모두 v29.

이 change에는 배포 동일성을 확인할 심볼이 없다 — `lineRefreshInterval`은 `const`라
인라인되고 `lineRefreshSeconds`는 인라인되며 `reloadPeriodIn`은 테스트 전용이다.
대신 **렌더된 값**이 판정 근거다.

### 주기가 실제로 갈라졌다

| 화면 | 배포 전 | 배포 후 | 출처 |
|---|---|---|---|
| `/positions` | `content="30"` | **`content="5"`** | 엔진 관측 주기 |
| `/dashboard` | `content="30"` | **`content="5"`** | 엔진 관측 주기 |
| `/orders` | `content="15"` | `content="15"` | `ordersTTL` — 무변동 |
| `/signals` | `content="15"` | `content="15"` | `DefaultWatchInterval` — 무변동 |

표시줄 셀도 같은 값을 말한다: `/positions`·`/dashboard`가 `5초마다`,
`/orders`가 `15초마다`. **이 둘이 바로 F2가 혼동하던 쌍이다** — 배포된 화면에서
그 둘이 나란히 다른 값을 말하고 있고, 이제 그것을 테스트가 값으로 판정한다.

### 6.6 — 탭 하나를 열어 둔 상태 (읽기 볼륨 6배)

`/positions`를 5초 간격으로 3분간 36회 렌더했다. 배포 전 같은 창에서는 6회였다.

| | 값 |
|---|---|
| 렌더 | 36회 / 180초 |
| 최소 · 중앙값 | 8.6ms · 12.7ms |
| p90 · 최대 | 332.8ms · 455.7ms |
| **100ms 초과** | **정확히 6회** |

180초 ÷ 30초 TTL = **6개의 TTL 창**이고, 비싼 렌더가 정확히 6회다. 나머지 30회는
캐시에서 서빙됐다. **재로드 빈도가 6배가 되고 브로커 비용은 그대로**라는 것이 이
change가 성립하는 유일한 근거였고, 그것이 배포된 시스템에서 관측됐다.

엔진 `reconcile` 루프의 완주 간격이다. 부하 창은 02:00:20~02:03:20이다.

```
01:56:00  01:57:03  01:58:06  01:59:08 │ 01:59:42* │ 02:00:45  02:01:46  02:02:48  02:03:51
    62.6s     62.5s     62.9s     62.7s │           │    62.8s     60.5s     62.6s     62.5s
                                        └ 엔진 재시작 └────────── 부하 창 ──────────┘
```

부하 구간(62.8·60.5·62.6·62.5초)이 조용한 구간(62.5~62.9초)과 구분되지 않는다.
60.5초는 부하 때문이 아니라 그 사이클이 브로커 인증 실패로 **일찍 끝나서**다
(아래 참조). WAL이라 쓰기 차단이 없다는 설계상의 예상과 실측이 일치한다.

### 6.5 — 개요 화면의 브로커 호출 0

`/dashboard`만 6초 간격으로 10회, 54초에 걸쳐 열었다. 표시줄의 데이터 시각 칸이다.

```
02:03:07Z (55초 전) tone=warn
02:03:07Z (61초 전) tone=stale
...
02:03:07Z (109초 전) tone=stale
```

**캐시 시각이 `02:03:07Z`에 고정된 채 경과만 자란다.** 개요가 브로커를 한 번이라도
불렀으면 시각이 갱신됐을 자리다. peek 계약이 5초 주기에서도 지켜진다.

**그리고 톤 임계가 재로드 주기를 따라가지 않았다.** 55초에서 `warn`, 61초부터
`stale` — 경계는 30초(TTL)와 60초(2×TTL)이고 새 재로드 주기 5초가 아니다. 이것이
두 번째 MODIFIED가 명문화한 성질(`톤의 근거인 캐시 TTL과 그 화면의 재로드 주기는
별개의 값이다`)의 **실측 확인**이다.

### 6.5 — 열린 상세가 재로드를 넘어 유지된다

접힘 상태는 URL에 있다(a055 §6). `/positions?explain=holdings-basis`는 접힌
렌더보다 294바이트 크고 해당 패널의 링크가 닫기 링크(`/positions`)로 바뀐다. 그
URL의 meta refresh도 `content="5"`이고, meta refresh는 **현재 URL을 그대로 다시
연다**. 따라서 열린 상세는 재로드를 넘어 구조적으로 유지된다.

### 6.5 — 아직 확인하지 못한 것 하나

**스크롤 위치가 재로드를 넘어 유지되는지는 이 실측이 답하지 못한다.** design D4가
"미검증 전제"라고 적어 둔 바로 그 항목이고, 브라우저의 스크롤 복원 동작이라
`curl`로는 관측할 수 없다. 사람이 `/positions`를 열어 아래로 스크롤한 뒤 5초를
기다리는 것으로 끝나는 확인이며, **그 전에는 6.5를 완료로 표시하지 않는다.**

화면에 그려진 값 자체(보유 종목의 보호선·관리 배지)는 계좌 정보라 이 기록에 넣지
않는다.

### 함께 관측된 것 — a080과 무관한 기존 문제

부하 창 안의 `02:01:46` 사이클이 `reconcile.mismatch`였다.

```
error: reconcile: snapshot discarded after a partial read:
       walking the open-order list: official: authentication failed
```

브로커 인증 실패다. **a080이 만든 것이 아니다** — 같은 오류가 2026-08-02부터
사흘에 걸쳐 10회 기록되어 있고, a080은 브로커 경로에 닿지 않는다. 엔진은
fail-closed로 스냅샷을 버리고 다음 주기에 재시도했으며 그 다음 사이클은 clean이다.
별도 조사 대상으로 남긴다.
