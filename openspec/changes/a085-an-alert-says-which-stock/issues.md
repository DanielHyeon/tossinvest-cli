# a085 · Issues

## I1 — 콘솔 절반은 a086이다

`operatorview.BuildExitLine`이 `snapshot_quarantined`에서 저장값을 버리는 문제는
a077의 승인된 요구사항과 충돌하므로 이 change에서 뺐다. `proposal.md`의 "분리 결정"
참조. a086은 `operator-console`의 해당 요구사항을 MODIFIED로 고치고 그 수정분에 대한
리뷰를 받아야 한다.

## I2 — 차단 evidence 문자열이 첫 관측에 고정된다 (a083 I4와 동일 건)

포지션 화면의 "the engine believes 2 of 103590, the account says 3"은 차단이 처음
기록된 시각의 값이고, 같은 화면 아래의 원장 수량 4와 다르다. `EnterReconcile`이
멱등이라 첫 evidence를 보존하기 때문이다.

a085는 알림 문구만 다루므로 범위 밖이다. a086이 화면 표현을 다룰 때 함께 판단한다.

## I3 — 알림 본문의 원문 에러는 여전히 영어다

한국어 설명 뒤에 `원인: <원문>`으로 붙인다. 의도된 것이다 — 운영자가 그대로 검색하거나
붙여넣을 수 있어야 하고, 브로커·Go 에러 문자열을 번역하면 원본과 대조할 수 없다.

## I4 — `make gate`가 병행 세션의 편집으로 막혀 있다 (a084 I8과 동일 건)

`.claude/CLAUDE.md`와 `.codex/agents.md`를 다른 세션이 편집 중이라 agent-config sync가
drift를 보고한다. 다른 세션의 진행 중 안전 부트스트랩 편집을 덮어쓰지 않았다.

a085 자체의 증거는 전부 통과했다.

```
go test ./...   8016 passed, 0 failed, 99 packages
go vet ./...    clean
openspec        valid --strict
logic-map       evidence complete (24 함수)
```

## B1 — **표시(blocking)**: 레지스트리가 자기가 고치려던 그 알림에서 경주에 진다

`gateway.go:288`이 빈 `&InstrumentNames{}`를 만들고, 유일한 writer는
`reconcileloop.go:409`의 **안정화된** 주기 안에 있다(2초 이상 떨어진 두 수집이
일치해야 한다 — 움직이는 계좌에서는 몇 분). `runtime.go:279`는 모든 루프를 동시에
시작하고 ExitObserver의 첫 주기는 t≈0, 5초 간격이다.

그런데 알림 latch는 **프로세스당 1회**이고 다시 무장되지 않는다.

```go
exitloop.go:1407   if o.unmanaged[p.ID] { return }
                   o.unmanaged[p.ID] = true
```

즉 한 포지션이 평생 받는 그 한 번의 알림이, 이름을 알기 전에 나갈 확률이 높다 —
**코드만 찍힌다. a085가 고치려던 바로 그 결함이다.** 비결정적이라 테스트에서는 통과하고
시작 시점에 바쁜 계좌에서 실패한다. 재시작마다 창이 다시 열린다.

곁가지: `Label`은 코드를 대문자화하는데 `Fields[symbol]`은 `p.Symbol` 원문이라
소문자 저장 심볼에서 제목과 payload의 대소문자가 갈릴 수 있다.

고칠 방향: 레지스트리를 원장에 저장하거나(재시작 보존), 이름을 알기 전 알림을
지연/재무장하거나, adoption 시점에 보유 응답에서 이름을 직접 읽는다.

## B2 — **§0.8(blocking, 선재 결함)**: 알림 payload가 본문으로 렌더되어 계좌번호가 ntfy에 실린다

`internal/obs/notifier.go:405`

```go
body = strings.TrimSpace(body + "\n" + renderFields(e.Fields))
```

그리고 엔진 알림 다수가 `obs.FieldAccount: d.opts.AccountRef`를 싣는다
(`adoption.go:362,424,463`, `exitloop.go:767`, `exitwiring.go:112,150`).
`resolveAccountRef`는 브로커 실계좌번호를 돌려준다. 따라서 **계좌번호가 ntfy 토픽으로
전송된다.** ntfy 전송부 주석 자신이 그 토픽이 추측 가능한 공개 ntfy.sh 토픽일 수 있다고
적어 두었다.

`internal/obs/`는 이 diff가 건드리지 않았다 — **선재 결함이다.** 하지만 a085가 새로 쓴
`TestTheUnmanagedAlertNamesTheStockInKorean`은 §0.8을 단언한다고 적어 놓고
`Event.Title+Body`만 본다. **발행되는 표면이 아니다.** a085는 확인하지 않은 보장을
주장하고 있다.

두 가지가 필요하다: (1) 테스트를 `obs.notificationFor(...).Body`로 옮겨 주장과 검사를
일치시킨다. (2) 계좌번호 노출 자체는 별도 change로 — 본문에서 걸러내거나 끝 4자리로
마스킹한다.

## I5 — exit 경로 알림은 `Names`가 nil인 상태로만 테스트된다

`exitloop_test.go` 하네스가 `ExitObserverOptions.Names`를 설정하지 않아, `o.label()`
6곳과 `notifierAlerter` 2곳이 전부 nil 레지스트리로만 실행된다. a085의 이름 단언은
전부 reconcile 드라이버의 unmanaged 알림에 있다.

즉 proposal이 인용한 그 실제 원장 행 — `the exit policy could not judge 032820` — 이
`에이치엘비(032820)`로 렌더되는지 **어떤 테스트도 확인하지 않는다.**
`exitwiring.go:338`의 `opts.Names = c.Names`가 사라져도 전 스위트가 green이다.

## I6 — 한글 판정이 14음절 샘플이다

`a085_alert_text_test.go:40`의 `strings.ContainsAny(..., "가나다라마바사아자차카타파하")`는
한글 범위가 아니라 표본을 본다. 그 음절을 안 쓰는 재작성은 "한국어가 아니다"로 실패한다.
`unicode.Is(unicode.Hangul, r)`나 0xAC00–0xD7A3 범위로.
같은 파일 :80의 `strings.Contains(alert.Title, "(")`도 제목의 모든 괄호를 금지한다.

## I7 — 두 배선 경로가 레지스트리 규칙에서 어긋난다

`ReconcileDriver`는 `opts.Names = c.Names`로 **무조건 덮어쓰고**, `ExitObserver`는
nil일 때만 채운다. 한쪽 규칙으로 통일한다.

## I8 — runtime supervisor 알림 2건은 영어로 남았다

의도적 범위 결정인지 누락인지 diff와 a085 문서 어디에도 없다. 종목을 가리키지 않는
루프 생명주기 알림이라 영어로 둔다면 그 경계를 한 줄로 적는다.
