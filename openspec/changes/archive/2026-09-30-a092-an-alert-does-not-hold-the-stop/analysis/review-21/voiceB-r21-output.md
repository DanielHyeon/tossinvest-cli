<!-- 21라운드 보이스 B(산출물 무기화 렌즈) — Claude code-inspector-tester 서브에이전트(독립 컨텍스트, 같은 하네스), 세션 재시작 뒤 재실행.
     프롬프트: prompt-common.md + 호출 렌즈(보이스 B). 대상 8b38de8e(a092 문서·인용 코드는 HEAD 7b9df40f 와 바이트 동일). 원문 그대로. -->

**판정: BLOCK**

21판 설계대로 구현해도 exit 관측 goroutine은 원격 전송을 동기로 기다리는 입구 둘을 그대로 가진다. 운영 모드가 바뀔 때 알리는 경로(ModeAnnouncer)이고, 21판 책임 표·어댑터 설계·RED 목록 어디에도 없다. 이 change의 제목 목표를 이루지 못하므로 P0다.

## 발견 표

| # | 등급 | 주장 | 증거(파일:줄) | 권고 |
|---|---|---|---|---|
| 1 | **P0** | 21판 기록 전용 어댑터는 `Alerts` 입구(`exitloop.go:1710`) 하나만 막는다. exit goroutine은 모드 전이 알림 경로로 `Notifier.Notify` → `notifyCritical` → `claimAndDeliver` → `deliver`(동기 원격, 3회 시도)에 여전히 닿는다. 입구는 둘이다. ① 관측 두절 승격 `exitloop.go:846-847`: `Escalate…(…, o.opts.Announcer)`이고 Announcer는 `cmd/tossctl/engine.go:639`의 `ectx.Notifier`다. ② 가격 읽기 `exitloop.go:747`의 `Retrier.Query`가 401/403을 받으면 `retry.go:365` → `:409-414`로 가고, 이 Retrier의 Announcer는 `exitwiring.go:54`의 notifier다(`exitwiring.go:333`이 공유 `c.Retrier`를 넘긴다). 원장은 커밋 뒤에 announcer를 동기로 부르고(`operating_mode.go:478-479`, AST B27/B28, call@479), 그 다음은 `obs/mode.go:57`의 `n.Notify`다. `EventOperatingMode`는 critical이다(`event.go:328`). D0.3e 4의 *"ExitObserver.alert의 호출은 한 자리다"*는 Alerts에 대해서만 참이다. 20판 design D5(`design.md:1989-2003`)는 이 입구를 알고 있었다(옛 좌표 `exitloop.go:796`). 운영 선례도 있다: 운영 원장의 현재 모드가 `AUTO`/`BROKER_AUTH_REJECTED`이고 `AnnounceOperatingMode`가 2026-07-31에 1회 발동했다(tasks 2.7). 21판이 스스로 추가하는 완화 경로로 NORMAL이 되면 다음 두절이나 401에서 이 경로가 다시 열린다 | 위 | exit observer의 Announcer에도 기록 전용 어댑터를 배선한다. Retrier는 reconcile과 공유하므로(`reconcileloop.go:360`) exit 전용 인스턴스로 나눌지, announce를 기록 전용으로 바꿀지 설계 결정이 필요하다. D0.3e 3번 표에 행을 추가하고, 21.3에 RED(k) 두 건을 더한다: 막힌 publisher로 두절 승격이 일어나도, 가격 읽기가 401을 받아도 exit 사이클이 publish 없이 반환한다 |
| 2 | P1 | D0.3e 5번은 `claimAndDeliver`의 `deliver`를 잠금 밖으로 뺀다. 그러면 동기 발송자가 운영자 해제를 뒤집을 수 있다. 순서: `MarkAlertAttemptFailed`가 Applied(`notifier.go:494`) → `ReleaseAlertClaim`이 Applied(`:541`,`:548`) → 운영자 `Acknowledge`가 셈 0으로 Clear(`:851-875`) → `deliver`가 `n.Gate.Block`(`:571`) → `notifyCritical`이 `n.escalate`(`:223-228`). 21판에서는 투영이 배선되므로 이 승격이 실제 진입 차단이 된다. 21판은 20판 :66의 *"전송의 성공 여부는 그 사유를 되살리지 않는다(SHALL NOT)"*를 a124 소유라며 지웠다(델타 :60). 그러나 정본 「늦은 적용」(`spec.md:1456-1458`)의 주어는 배달 실행자뿐이다. a124 design D10은 *"운영자의 해제를 기계가 무시하는 방향은 … 사람 승인이 지배한다는 불변식에는 역행"*이라고 적는다. 진입 쪽으로는 보수 방향이라 P0는 아니다 | `head-ast-20/internal-obs--notifier.deliver.json`(B21~B27), a124 `design.md:489-495` | 20판 SHALL NOT을 「모든 발송자」로 일반화해 되살리거나, 동기 `deliver`의 래치를 a124처럼 claim 시점 해제 세대에 건 조건부 잠금(`BlockUnlessClearedSince`)으로 바꾼다. 21.2(a) 겹침 표에 이 순서를 넣고 RED를 추가한다 |
| 3 | P1 | 일반 등급 SHALL(델타 :42, Scenario :105-107, exit-policy :27·:59-61)을 유지했지만 21.x에 작업이 0개다. 어댑터 설계(D0.3e 4)도 일반 등급에 대해 말하지 않는다. 그러면 구현은 둘 중 하나가 된다. `publishBestEffort`의 동기 원격(`notifier.go:134-135`·`:165`)을 남기면 델타 위반이다. 일반 알림을 전부 버리면 같은 MODIFIED 본문이 보존하는 정본 *"캡 발생은 알림된다(SHALL)"*(exit-policy 델타 :25)와 어긋난다 | 위 | 21.3/21.4에 일반 등급 이관(D0.8의 유계 버퍼)과 RED를 추가한다. 아니면 사용자 결정으로 두 델타의 문장을 바꾼다 |
| 4 | P1 (T) | 델타 :32의 SHALL NOT *"exit 관측 goroutine이 남긴 critical 행에는 발송 임차가 남아 있어서는 안 된다"*가 무조건이다. 그런데 design D0.3e 4(`:444-446`)는 반납 원장 오류 때 임차가 81초(`alert_claim.go:34`) 남는 것을 허용한다. claim 시점에 남의 임차가 있던 경우(21.3(i))도 행에 임차가 있다. 그러면 Scenario :84-87 *"남의 임차로 보지 않고 집을 수 있다"*가 거짓이 된다 | 델타 :32·:84-87, design :444-446 | 예외 둘(반납 오류 → 임차 만료까지, 남의 임차)을 요구 문장에 명시한다 |
| 5 | P1 (일부 추측) | D0.3f 4 안 (가) 세대 울타리의 정의가 비어 있어 fail-open이 될 수 있다. 순번은 삽입 **전**의 `modeRowCount`다(`operating_mode.go:436`, INSERT `:444`). *"기동 복원은 최신 행의 순번을 싣는다"*에서 그 값이 행 수로 구현되면, 기동 뒤 첫 전이의 순번이 last와 같아 버려진다. 그 전이가 AUTO 강화면 진입이 안 막힌다(추측 — 구현 선택에 달림). 또 복원은 `created_at DESC`로 최신 행을 고르는데(`:626`) 울타리는 행 수 순서라 순서 기준이 둘이다 | design `:572-576` | 순번을 rowid나 삽입 순번 하나로 정의하고, 복원도 같은 행의 순번을 싣는다. RED 두 건 추가: 「기동 복원 직후 첫 자동 강화가 투영된다」, 「벽시계 되감김」 |
| 6 | P2 (T) | `MANIFEST.txt`는 HEAD를 `1a0027d2`로 적는데 문서는 `c1e34dc4`라 한다(D0.3e :387, D0.3f :512, tasks :6·:16). tasks 21.0은 「추출 18개」인데 실제는 20개다. 내용은 같다: 파일 13개 sha가 c1e34dc4·1a0027d2·8b38de8e·HEAD에서 같고, 격리 사본 재추출 결과가 20/20 바이트 동일이다 | `analysis/head-ast-21/MANIFEST.txt:1` | 표기를 통일한다 |
| 7 | P2 (T) | head-ast-21은 문서가 기대는 곳 일부를 추출하지 않았다: `AnnounceOperatingMode`, `Retrier.escalateCredentialFailure`, `alertDeliverer.escalate`(announcer nil, `alertdelivery.go:451-452`), `EntryGate.Clear`/`BlockUnlessClearedSince`(세대 주장), `CurrentOperatingMode`/`modeLatestOrder`, `settleUnderClaim`/`recordAlertTx`/`claimOwed`(재무장 주장). head-ast-20의 `claimowed.json`은 sha가 낡았다(fb28… ≠ HEAD ad74…) | 위 | 추출을 추가한다 |
| 8 | P2 (T) | 21판은 스스로 「사본 둘 방지」 규칙으로 지웠으면서 정본 사본을 남겼다: :44(보조 실행자 의무, 정본 :203), :46과 Scenario 「배달 실행자가 죽는다」 :121-126. 이 Scenario는 정본 :1039와 제목이 같고 조항이 다르며, 정본 :1085-1094가 바로 이 경우를 경고한다. ADDED :140·:150은 risk-management `:133`의 사본이다 | 위 | 정본을 가리키는 문장으로 줄이거나 귀속을 명시한다 |
| 9 | P2 (T) | archive하면 거짓이 될 HEAD 사실이 정본 규범에 들어간다: ADDED :140 *"HEAD 생산 조립에서 투영기 묶기 호출자는 0이다"*, 주석 :50 *"HEAD에서 … :309"*. 정본 :1032-1034의 근거 ①은 처분이 미결이다(21.7(e)) | 위 | HEAD 사실은 design으로 옮긴다. 21.7(e)는 archive 전에 결정한다 |
| 10 | P2 | `Flush`는 `n.mu`(`notifier.go:734-735`)를 쥔 채 `Publish`(`:779`)를 한다. 델타 :48 「모든 보유자」와 exit-policy :39를 문자 그대로 위반한다. 생산 호출자는 0이다 | 위 | 삭제하거나 범위 문장으로 제외한다 |
| 11 | P2 (T) | 안 가는 알림 하나에 트랜잭션 둘이다(`outbox.go:241` + `alert_claim.go:320`). exit-policy Scenario :67-69는 *"outbox 트랜잭션 3건"*이라 적고, engine-safety :68은 쓰기 둘, exit-policy :31은 하나라 적는다. exit-policy 머리 :4-5는 21판 편집을 축소해 적는다 | 위 | 열거를 맞춘다 |
| 12 | P2 (T) | design :465 *"a092가 이 식에 더하는 것은 없다"*는 틀렸다. 어댑터의 순간 claim이 전제 H ③(임차 경합 없음, a124 `design.md:155-156`)을 깨고, 반납 실패는 81초 항을 더한다 | 위 | 전제 위반 경로로 적는다 |
| 13 | P2 | 완화 명령 설계가 빠뜨린 것 셋. Announcer가 명시되지 않았다(`obs/mode.go:13-15`는 완화도 알리라고 한다). `ErrModeAnnouncementFailed`는 커밋 뒤 `(record,true,err)`로 돌아오므로(`operating_mode.go:479-481`) 명령이 「완화됨」으로 보고해야 한다. 엔드포인트는 코드 선례(`alert_control.go:13-15`, 「다른 힘」)대로 분리하는 것이 맞다. a066 작업 트리(미커밋, 참고용)는 `engine entry-lock-release`/`risk-latch-release` 모양이라(`cmd/tossctl/engine_risk_relaxation.go:136,194`) D0.3f 2의 「`engine <대상> relax`」와 다르다 | 위 | 21.7(c)에 반영한다 |
| 14 | P2 | 작업 목록에 없는 요구(판정 f): mutating 표지·콘솔 부재·타이핑 확인 없음 RED, 되살린 선점 문단 :52, Scenario 셋(상한을 읽지 않는 transport, 예산을 줄여도 기록은 그대로다, 재강화는 새 게이트에 투영되지 않는다), 기동 복원 실패 시 기동 거부. 옛 §6·§8 처분은 21.4로 미뤄져 있다 | tasks :24-55 | 추가한다 |

## 필수 판정 1~6

**1. A-1 = B-1 — 주인 문제는 닫혔고, D0.3e 3번 표는 불완전하다.**
- 닫힌 근거: 정본 `spec.md:1432-1496`, 판정 `judge` `alertdelivery.go:429`/`:437`, publisher 부재를 실패 시도로 셈 `:314-336`, 한도 = `DefaultCriticalAttempts` `:98`.
- 표가 빠뜨린 것: Announcer 입구(#1), `notifyCritical`의 journal nil 갈래 B1 `:177` → `:186`(잠재적인 동기 원격), `checkAlertLease` `:260`을 어댑터가 유지할지.

**2. A-2(안 가) — Alerts 입구에 대해서만 제목을 달성하고, 제목 전체로는 미달(#1).**
- 잠금 범위 변경의 위험은 #2다.
- `deliver`의 PRECONDITION 주석(`notifier.go:406-410`)이 거짓이 된다.
- 이중 발송 배제는 a099 임차가 진다.

**3. archive하면 정본에 들어갈 문제.**
- 지킬 수 없는 SHALL: #4.
- 정본 exit-policy와의 긴장: #3.
- 사본과 거짓 사실: #8, #9.
- 두 델타 사이는 트랜잭션 개수만 어긋난다(#11).
- ADDED와 a124 정본은 정합하다. `:1444-1446`은 조건문이라 참으로 남는다. AC2는 상태 세대 증가(`retry.go:537`)를 지켜야 하고, 이는 D0.3f 4가 시험 대상으로 적었다.

**4. Q3(D0.3f).**
- 원장 API는 충분하다: B6 `:371`~B10, B14 `:409`, B16 `:417`, B18 `:424`, B19 `:428`, B24 `:461` < `:468` 모두 확인. 다만 울타리는 원장 편집이다.
- 배선 순서: `:249`~`:269`가 루프 시작보다 앞일 개연성이 높다(루프는 `Run` 안 `go` `@306`·`@324`·`@335`, `Recover@294` 뒤). 증명은 21.7(a)가 진다.
- 원자 교체와 울타리는 #5를 고쳐야 한다.
- 전개 절(기존 `ENTRY_BLOCKED` 행의 처분은 사람 몫)은 적절하다.

**5. Q4·Q6 권고.**
- **Q4 = (c)** 잔여로 적고 구조 핀을 더한다. `parkAlert`는 `EnqueueAlert` `:551`보다 먼저 `ReasonUnresolvedInDoubt`를 잠그고(`replay.go:535-537`), `Acknowledge`는 `ReasonAlertUndelivered`만 푼다(`notifier.go:875`). 그러므로 덮이지 않는 창이 혼자 진입을 열지 못한다. (a)는 `replay.go:101-107`의 이유와 충돌하고, (b)는 High-risk 원장 편집인데 a124 D10 (i)도 닫지 못한다.
- **Q6 = 기록만.** 실행자의 임차당 예산은 시도 1회이므로 정본 「예산을 다 쓰면 임차를 놓는다」(`spec.md:1394-1398`)가 적용된다. 근거는 `alertdelivery.go:342-351`(Y5).

**6. 좌표 전수 — 문서 좌표 약 75개를 대조했고 내용 불일치는 0이다.**
- 출처 표기 오류 2(#6).
- 오도하는 문장 1: D0.3e 4의 「한 자리」.
- brief가 가리킨 `8d9731c1`은 20판이 아니라 19판 델타다. 20판 최종 델타는 `d736e0cb^`에 있고, 이번 삭제 대조(e)는 그것으로 했다.

## (e)·(g) 보조 판정

**(e) 21판이 지운 20판 문단의 소유.**
- 굶주림 :56 → 정본 :1584. 최악 시간 :58 → 정본 :1590. 전송기 부재 :76 → 정본 :1449. 같은 행 이중 발송 → 정본 :1297. 전부 정본이 소유한다.
- :66만 소유가 서지 않는다(#2).
- :86은 소유 문제가 아니라 Q1 범위 결정으로 지운 것이다.

**(g) `check_values.py` — rc 1, FAIL 60.** 21판이 쓴 줄에는 실패가 없다. 이 검사기가 보지 못하는 것:
- 체류 열거가 두 항뿐이다(`:861-867`). 잠금 대기 항과 반납 쓰기 항은 검사하지 않는다.
- 배달 쪽 열거는 명시적으로 검사 밖이다(`:869-877`).
- 좌표·AST ID, 정본과의 중복, 기호로만 된 지연 식은 보지 못한다.

## git 상태

`git -C … status --short -- openspec/changes/a092-an-alert-does-not-hold-the-stop`의 출력은 **빈 줄**이다. 예상했던 untracked `analysis/review-21/`가 나오지 않은 것은 그 디렉터리가 이미 HEAD에 커밋돼 있기 때문이다. 리뷰 도중 병행 커밋으로 HEAD가 `9fbd8e66`에서 `7b9df40f`로 옮겨졌다. 8b38de8e 대비 차이는 review-21 파일 4개뿐이고, a092 문서와 인용 코드는 바이트 동일하다. review-21 안의 파일은 prompt-common.md 말고는 열지 않았다. 재추출은 저장소 밖 `/tmp/claude-1000/a092-r21/voiceB-copy`에서만 했다.

Recommendation: BLOCK because exit goroutine이 모드 전이 알림 경로(`exitloop.go:846-847`, `:747` → `retry.go:409-414` → `operating_mode.go:479` → `obs/mode.go:57`)로 여전히 원격 전송을 동기로 기다리고, 21판의 책임 표·어댑터·RED 어디에도 그 입구가 없어 이 change의 제목과 델타의 SHALL NOT을 이루지 못한다.
