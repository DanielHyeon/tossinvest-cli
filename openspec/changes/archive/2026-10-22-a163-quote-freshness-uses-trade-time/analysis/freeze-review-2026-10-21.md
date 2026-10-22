# a163 proposal-freeze 독립 적대 리뷰 — 2026-10-21

- 리뷰어: 독립 적대 리뷰어(저자 아님). 읽기 전용. production 편집 0 · 브로커 호출 0 · 커밋 0.
- 대상: 커밋 `60033240a`(change 문서) · base `f7d80c348`. `git diff f7d80c348 HEAD -- internal` 은 비어 있다(대상 production 파일 무편집).
- 도구: `tools/logic-map`(HEAD 에서 `GOFLAGS=-trimpath` 로 다시 빌드해 직접 열거) · `go1.26.5`(go.mod `go 1.25.0`) · `grep`.
  codegraph CLI 는 이번 리뷰에서 돌리지 않았다 — 호출자 수는 `grep` 전수로 대신 쟀다(아래 「검증한 주장」).
- 방향(사용자 결정 A, H1~H7)은 재론하지 않는다. 이 리뷰는 **그 방향이 쓰인 대로 강제되는가**만 본다.

## 판정: **APPROVE-WITH-CHANGES**

방향 자체(판정은 받은 시각 그대로, 브로커 시각은 보고에만)는 안전 4 와 양립하고, 판정 경로에 새 분기를 두지 않는다는 설계는 맞다.
P0 은 없다. 그러나 P1 넷이 **편집 전에** 고쳐져야 한다: 하나는 JSON 계약이 「모름」을 「서기 1년」으로 내보내는 결함이고, 하나는
「판정 함수 diff 0」 강제 수단이 원리상 성립하지 않는 것이며, 하나는 이 change 의 유일한 산출물(나이 로그)이 생산에서 0 줄이 되는
배선 공백이고, 하나는 청산 goroutine 에 새 코드를 넣으면서 패닉 격리를 정하지 않은 것이다.

---

## P1 — 편집 전에 고쳐야 함

### P1-1. `json:"source_at,omitempty"` 는 0 값 `time.Time` 을 생략하지 않는다 — 「모름」이 `0001-01-01T00:00:00Z` 로 나간다

- 근거: design.md:30 · design.md:76·90(H7 (a) `source_at,omitempty`) · proposal.md:46. `domain.Quote`(internal/domain/models.go:399-421)의
  유일한 시각 필드 `FetchedAt` 도 `omitempty` 없이 `json:"fetched_at"` 이다.
- 실측(스크래치 Go 프로그램, go1.26.5): `struct{A time.Time \`json:"a,omitempty"\`}{}` → `{"a":"0001-01-01T00:00:00Z"}`.
  `omitempty` 는 struct 타입에 효과가 없다. `omitzero` 는 생략한다(`{}`).
- 실패 시나리오:
  1. `tossctl quote --json`(cmd/tossctl/quote.go:43 → `hybrid.Client.GetQuote` internal/hybrid/client.go:117-130)의 **WTS 갈래**
     (`internal/client/quote.go:115`)와 `/stocks` 어댑터 `adaptStocks`(internal/official/stock_reads.go:81-97)는 `SourceAt` 을 채우지
     않는다 → 출력에 `"source_at":"0001-01-01T00:00:00Z"` 가 생긴다. WTS 갈래는 upstream 동작이므로 안전 3(토글 OFF = upstream 동일)
     의 출력 계약이 바뀐다.
  2. 공식 `/prices` 의 null·키 없음·해독 실패도 같은 문자열이 된다. 소비자가 나이를 계산하면 약 2025 년 묵은 시세 — spec delta
     official-read-compatibility.md:7 「zero value … MUST NOT be filled from a local clock」 의 취지(「모름」을 값으로 위장하지 않음)를
     JSON 층에서 어긴다.
- 놓친 생산자: `quote-time-evidence.md` §3 은 시각의 **소비자**만 열거했고 `domain.Quote` 의 **생산자**(`adaptStocks` · WTS
  `client.GetQuote`)를 열거하지 않았다. JSON 계약 판단은 생산자 전수에서 해야 한다.
- 요구 수정(둘 중 하나, design D2·H7 기록에 명시):
  - (a) `json:"source_at,omitzero"`(go.mod `go 1.25.0` 이라 사용 가능). 시험: 0 값 Quote 의 JSON 에 `source_at` 키가 **없다**,
    WTS 갈래 Quote 의 JSON 바이트가 편집 전과 같다.
  - (b) 저장소 선례(`adaptTrades` market_reads.go:311 `Time: t.Timestamp`, `adaptMarketIndicatorPrices` :436 `Timestamp: p.Timestamp`)처럼
    **원문 문자열**(`SourceTimestamp string \`json:"source_timestamp,omitempty"\``)을 보존하고 해독은 보고 경로에서만. 어댑터에 해독
    분기가 0 이 되고 null 과 해독 실패를 보고 쪽에서 가를 수 있다(P3-3). ponytail 사다리상 (b)가 더 짧다.

### P1-2. 「판정 함수 일곱 diff 0」 은 쓰인 수단(`source_sha256`)으로 성립하지 않는다 — 그리고 헬퍼 우회를 막지 못한다

- 근거: tasks.md:31-32 「편집 뒤 source_sha256 이 base 와 같아야 함」 · tasks.md:53 「1.2 의 일곱 함수 source 해시 불변 확인」 ·
  design.md:46-47.
- 실측: logic-map 의 `source_sha256` 은 **함수가 아니라 파일 전체**의 해시다. `quoteUsable`·`observe`·`ObserveOnce`·`stableObservationID`·
  `snapshotContext` 번들이 모두 `ba3051547658…` = `sha256(internal/app/engine/exitloop.go)`. `adaptPrices` 번들 `b356524b48df…` =
  `sha256(internal/official/market_reads.go)`. `tools/logic-map/check_analysis.py:2963-2967` 도 이 값을 「수정된 함수의 리비전」 대조에만
  쓰고, **수정되지 않았음**을 단언하는 경로는 없다(게이트는 diff 훅이 닿은 함수에 증거를 **요구**할 뿐 금지하지 않는다).
- 실패 시나리오: 이 change 는 `observedQuote`(exitloop.go:829-834) · `observe` · `ObserveOnce` 를 같은 파일에서 편집한다 → 일곱 번들의
  `source_sha256` 은 **반드시** 바뀐다 → 3.4 는 충족 불가 → 구현자가 조용히 기준을 바꾸거나(전칭 판정이 표본 0 으로 통과) 항목을 건너뛴다.
- 우회: 일곱 함수의 본문이 그대로여도 판정은 바뀔 수 있다 — 그들이 부르는 `snapshotContext`(exitloop.go:1190, `stableObservationID` 의
  유일한 생산 호출자, :1192) · `requiresFullJudgement` · `noteHeldProposal` · `decimalOf` · `clock.LeaseElapsed` ·
  `execgw.QueryPriceEvidenceDuration`, 그리고 **편집 대상인** `ObserveOnce` 루프(B6·B7 :503·:511) 와 `observe` 필터(B4·B7 :864·:871).
  예: `snapshotContext` 가 `observedQuote{…, FetchedAt: quote.SourceAt}` 를 넘기면 일곱 함수 본문은 바이트 동일한 채 관측 identity 가 바뀐다.
- 요구 수정:
  1. 함수 단위 고정: 커밋된 Go 시험이 `go/parser` 로 base 블롭(`git show <base>:<file>` 이 아니라 시험 안에 굳힌 기대 해시)과 현재 파일의
     `FuncDecl` 을 `go/format` 렌더해 **함수별** sha256 을 비교. 대상에 `snapshotContext`·`requiresFullJudgement`·`noteHeldProposal` 을 더할 것.
  2. 역할 고정(존재 검사가 아닌 금지 census): `engine`·`official` 패키지 전 파일에서 `SourceAt`(또는 P1-1(b)의 원문 필드) **식별자의 모든
     출현**(selector · 복합 리터럴 키 · method value)을 세어, 허용 자리 = {`adaptPrices` 의 대입 1 · `observe` 의 `observedQuote{…}` 리터럴 1 ·
     `exit_source_age.go`} 밖에서 0 임을 단언. 이것이 「헬퍼가 판정에 브로커 시각을 흘리는」 경로 전부를 닫는 유일한 범위다.
  3. 두 시험 모두 변이로 반증(아래 P2-2 ⑨⑩).

### P1-3. 나이 로그의 생산 싱크가 정해지지 않았다 — 지금 설계대로면 생산 로그는 0 줄이고 H6 (b) 분포는 수집되지 않는다

- 근거: design.md:54 「로그 필드에는 …」, proposal.md:31-32, spec delta exit-policy/spec.md:11-13(SHALL 구조화 로그). 싱크 이름이 어디에도 없다.
- 실측:
  - `ExitObserverOptions.Log` 는 생산에서 **nil** 이다(exitloop.go:190-191 주석 「Log 의 기존 줄은 계좌 원문을 싣고(logErr) 생산에서 nil」;
    cmd/tossctl/engine.go:671-681 은 `UnobservedLog`·`ExitOrderLog` 만 넣는다; internal/app/engine/exitwiring.go:357 은 `ZeroFloorLog` 만).
    `o.log`(exitloop.go:1932-1941)는 `opts.Log == nil` 이면 무동작.
  - 알림 경로로 싣는 경우: 생산 `opts.Alerts` 는 `obs.RecordOnly`(exitwiring.go:333·349)이고 `RecordOnly.Notify` 는
    `n.logEvent(withoutFields(e), …)`(internal/obs/record_only.go:51, :106-108 `e.Fields = nil`)로 **필드를 지운 뒤** 로그한다 → 나이·사유·
    포지션 id 필드가 로그에 남지 않는다.
  - 같은 결함의 선례: exitloop.go:217-218 주석 「2026-10-07 010170 손절이 이 로그에 0 줄이었음」(a129) — 시험은 로거를 주입하고 생산은 nil.
- 실패 시나리오: 시험(로거 주입)은 전부 초록, 배포 뒤 `source_*` 줄 0, H6 (b) 「운영 로그가 곧 표본」 결정이 표본 0 으로 남고 H1 후속 결정이 불가.
- 요구 수정: 전용 로거 옵션(예: `SourceAgeLog *obs.Logger`, a090 `UnobservedLog` 선례) 또는 기존 `UnobservedLog` 재사용을 design D4 에 명시하고,
  `cmd/tossctl/engine.go` 배선을 proposal 「영향」(proposal.md:40-42)에 추가. 생산 배선 시험: `engine.go` 가 그 필드를 채우는지(구조 시험 또는
  `Context.ExitObserver` 를 거친 조립에서 non-nil 단언) — 「시험이 매번 조종석에 앉으면 생산 절반이 안 돈다」.
  `source_missing` 알림(H4)도 로그 줄에 필드가 지워진다는 사실을 design 에 적고, 필드가 필요한 기록은 전용 로거로 낼 것.

### P1-4. 청산 goroutine 에 들어가는 새 코드의 패닉 격리가 정해지지 않았다(안전 4)

- 근거: design D4(design.md:48-53) — 보고를 `ObserveOnce` 안 「순회 뒤 한 번」에 둔다. `ObserveOnce` 는 `Run`(exitloop.go:400) 의 매 주기다.
- 실측: 감독 루프에는 recover 가 없다(internal/app/engine/runtime.go:26-31 은 패닉을 「루프가 사라짐」으로 다룰 뿐; `recover()` 는
  auxiliary.go:132 · exit_order_log.go:264 등 개별 경계에만). goroutine 패닉은 프로세스를 죽이고, protection 이 UNWIRED 인 지금 엔진 정지 =
  손절 없음이다. 선례: a129 `endExitOrderCycle` 의 로그 경로는 `defer func() { _ = recover() }()`(exit_order_log.go:264)로 격리돼 있다.
- 실패 시나리오: 래치 맵을 지연 초기화하지 않은 채 `o.sourceLatch[id] = …`(nil map 쓰기) → 첫 `source_missing` 주기에 패닉 → 엔진 종료 →
  이후 모든 손절 판정 없음. 보고 전용 코드가 손절을 멈추는 가장 짧은 경로다.
- 요구 수정: `exit_source_age.go` 의 진입 함수를 a129 와 같은 recover 경계로 감싸고(패닉은 버림 + 계좌 없는 한 줄), 시험: 보고기 내부 패닉
  주입 → 같은 `ObserveOnce` 가 정상 반환하고 다음 주기의 제출 수가 같다. 단 **concurrent map write 는 recover 불가 fatal** 이므로 래치 맵은
  `o.unobserved` 와 같은 단일 goroutine 소유임을 design 에 적고 `make test-race` 의 엔진 이름 목록에 새 시험을 넣을 것(tasks 4.2 에 명시).

---

## P2 — 구현 중 반드시 반영

### P2-1. §2 「편집 전 GREEN 고정 시험」 2.1~2.3 은 편집 전에 컴파일되지 않는다

- 근거: tasks.md:35-39. 편집 전 `domain.Quote` 에 브로커 시각 필드가 없으므로(models.go:421 까지) 가짜 `Prices` 로는 「9h30m 묵은 브로커 시각」을
  표현할 수 없다.
- 요구 수정: 2.1~2.3 은 **공식 어댑터 + httptest 본문**(`"timestamp":"2026-08-27T19:59:59.000+09:00"` · `null` · 키 없음 · `"garbage"` · 미래)으로
  `official.Client` 를 `ExitObserver.opts.Prices` 에 물려 구동할 것. 그래야 편집 전에 GREEN 이고(필드는 무시됨) 편집 뒤에도 같은 시험이
  어댑터~판정 전 구간을 잰다. 이 e2e 가 없으면 「어댑터 해독 실패가 행을 버려 `unobservedNoQuote` 로 빠진다」(P2-2 ⑦)를 루프 시험이 못 본다.

### P2-2. 변이 목록(tasks.md:57-58)이 필요한 축을 다 덮지 않는다

현재 ①~⑥ 에 더해 최소:
- ⑦ 어댑터가 해독 실패 행을 `continue`(행 탈락) — 그 종목이 `quotes` 에서 빠져 `ObserveOnce` B6 :503 `unobservedNoQuote` 로 판정 0. P2-1 의 e2e 로만 잡힌다.
- ⑧ 어댑터가 해독 실패 시 `Prices` 에 오류 반환 — `observe` B1 :849 → 주기 전체 판정 0.
- ⑨ `observe` 필터(B7 :871)의 `effectiveAt` 을 `SourceAt` 기준으로(④ 「FetchedAt 에 SourceAt 대입」과 다른 자리 — ④ 가 어댑터인지 `observe`
  리터럴인지 문구가 모호하므로 둘로 나눌 것).
- ⑩ `snapshotContext` 가 `stableObservationID` 에 `SourceAt` 을 넘김(P1-2 census 가 잡아야 함).
- ⑪ 보고기 패닉 경계 제거(P1-4).
- ⑫ 생산 싱크 배선 제거(`engine.go` 의 새 로거 필드 삭제) — P1-3 배선 시험이 잡아야 함.
- ⑬ JSON 태그 `omitzero`→`omitempty`(P1-1(a) 를 고른 경우).
무변이 대조군 GREEN 을 먼저 남길 것(tasks 에 이미 있음).

### P2-3. 「판정에 들어간 관측마다」 매 주기 로그는 5 초마다 포지션 수만큼 영원히 쌓인다

- 근거: spec delta exit-policy/spec.md:11-13(SHALL, 관측마다) · proposal.md:31-32. `DefaultExitObservationInterval = 5 * time.Second`(exitloop.go:98) →
  포지션 하나당 17,280 줄/일. 같은 파일의 정책은 반대다: exitloop.go:427-428 「Successful cycles say nothing. One line every five seconds forever
  is not observability.」 a090 의 미관측 줄도 **전이 때만** 낸다(exit_unobserved.go:180-195).
- 요구 수정: H6 분포 수집에 필요한 최소 빈도를 design 에 정하고 spec 문구를 그에 맞출 것 — 예: 포지션별 나이 구간(bucket)이 바뀔 때만 + 상한
  주기(예 1 분) 한 줄, 또는 주기 요약 한 줄. 판정 순서와 무관한 결정이라 방향 A 를 건드리지 않는다.

### P2-4. 래치 수명·해제 조건이 판정 집합과 묶이지 않았다

- 근거: design.md:52-53. 해제 조건 「임계 안의 관측 · 시각 생김」만 있고 (a) 그 주기에 판정되지 못한 포지션(`quote_expired`·`no_quote`)의 래치를
  어떻게 하는지, (b) 보유가 끝난 포지션의 래치를 언제 지우는지가 없다. a090 은 `settleUnobserved` 가 이번 주기 보유 표시 밖의 기록을 지운다
  (exit_unobserved.go:152-157).
- 요구 수정: 래치 정리를 a090 과 같은 기준(이번 주기 `held` 표시)으로 하고, 판정 안 된 포지션은 래치 상태를 바꾸지 않는다고 명시. 시험 2.6 에
  「포지션 청산 뒤 같은 id 재등장 없음 → 맵 크기 0」 단언 추가.

### P2-5. 공유 일반 등급 버퍼(64)를 다른 exit 일반 알림과 나눠 쓴다

- 근거: `obs.DefaultNormalRelayCapacity = 64`(internal/obs/normal_relay.go:18), exit 의 모든 일반 등급이 같은 `c.NormalAlertRelay()`
  (exitwiring.go:333). 가득 차면 버림(normal_relay.go:55-56).
- 위험: 재시작 직후(래치는 재시작에 잃는다 — design.md:53) 보유 전 종목의 `source_missing` 이 한 주기에 몰리면 같은 주기의 다른 일반 알림
  (`EventExitProposalCapped` 등)이 버려질 수 있다. 포지션 수가 작으면 무해하지만 design 에 상한(주기당 N 건 + 요약 1 건)을 정할 것.

---

## P3 — 정리

- P3-1. proposal.md:23 「체결이 오래된 관측을 기록·알린다」 는 H1(design.md:84 「`source_old` 알림을 내지 않는다」)과 어긋난다. 「기록하고, 시각 부재만 알린다」로.
- P3-2. proposal.md:66 「보고 자체도 끌 수 있게 할지는 H2」 — H2 기록(design.md:85)은 등급만 정했고 끄기 여부가 없다. 미결 항목을 결정 기록에서 닫을 것(안전 3 은 판정 경로 기준으로 이미 충족).
- P3-3. null 과 해독 실패를 한 사유 `source_missing` 으로 합치면 계약 형식 변화(예 `+0900`, 공백 구분 — 실측: 둘 다 `time.RFC3339Nano` 해독 실패)를 null 과 구별할 수 없다. 로그 필드에 `source_unparseable` 을 따로 둘 것(알림 사유는 H4 대로 하나여도 됨).
- P3-4. 나이 정의 「`FetchedAt − SourceAt`」(design.md:51)의 `FetchedAt` 은 `observedQuote.FetchedAt`(= 0 이면 `validatedAt` 대체, exitloop.go:867-870)로 명시. `domain.Quote.FetchedAt` 이 0 인 대체 경로에서 음수 거대값이 `source_ahead` 로 잘못 분류되지 않게.
- P3-5. tasks 0.5(tasks.md:19)는 결과를 `review.md` 에 두라고 하고 WORKFLOW.md:91·:290 은 `review.md` 존재가 gate ④ 다. 이 파일은 `analysis/` 에 있으므로 저자가 `review.md` 에 판정과 처분을 옮길 것.
- P3-6. tasks 번호 2.9 가 2.8 앞(tasks.md:45-46).
- P3-7(기존 위험, 이 change 가 키우지 않음): `apiPrice.Timestamp` 는 `string` 으로 **이미** 디코드된다(market_reads.go:137). 브로커가 이 필드 타입을 바꾸면(숫자 epoch 등) 오늘도 `Prices` 전체가 디코드 오류 → `observe` B1 → 판정 0 이다. a163 은 이 표면을 넓히지 않지만 issues.md 잔여(5.1)에 적을 가치가 있다.

---

## 「이 diff 의 어느 줄이 판정 자격을 바꿀 수 있나」(tasks 0.5 필수 질문) — 답

설계가 계획한 줄 중 판정 자격에 닿을 수 있는 것은 셋이다.
1. `adaptPrices` 의 해독 줄 — 오류 반환·행 탈락으로 바뀌면 판정 0(P2-2 ⑦⑧). 설계 문구(D3 「오류를 내지 않는다」)는 맞고, 강제 시험이 P2-1 e2e 로 보강돼야 한다.
2. `observe` 의 `observedQuote{…}` 리터럴 — 필드 혼동(④·⑨)이면 B7 필터·identity·원장 시각이 바뀐다. P1-2 census 가 닫는다.
3. `ObserveOnce` 의 보고 호출 줄 — 순회 안으로 들어가거나(②) 패닉하면(P1-4) 뒤 포지션·다음 주기의 손절이 늦거나 사라진다.
그 밖의 판정 경로 변화는 설계상 없다. 단, 1.2 의 diff 0 강제가 쓰인 대로는 성립하지 않으므로(P1-2) 「없다」는 지금 **주장**이지 **고정**이 아니다.

---

## 검증한 주장 (직접 재측정)

- `adaptPrices`(market_reads.go:168-180): 분기 1 · 반환 1 · 호출 `make`·`len`·`append`·`parseDecimal`·`time.Now`·`UTC` — `Timestamp` 읽기 0. logic-map 재생성으로 확인.
- `apiPrice.Timestamp string \`json:"timestamp"\``(market_reads.go:137) — null·키 없음 모두 `""` 로 디코드(구별 불가) 사실 확인.
- `observe` 분기 15·반환 6, `ObserveOnce` 분기 8·반환 5·defer 2, `quoteUsable` 분기 1·반환 2, `stableObservationID` 분기 4·반환 2, `settleUnobserved` 분기 12 — 저자 열거와 일치.
- `quoteUsable` 호출 자리 6(exitloop.go:511·963·1067·1138·1161·1291), `adaptPrices` 비시험 호출자 1(:149), `stableObservationID` 비시험 호출자 1(`snapshotContext` :1192) — grep 전수.
- `ObserveOnce` B7 `continue`(:511-515), 나머지 다섯 `return nil` — AST 반환 좌표로 확인.
- `settleUnobserved` 가 순회 뒤 한 번(:523) — 원문 확인. 생산 exit 알림기는 `RecordOnly`(exitwiring.go:333·349), 일반 등급은 유계 relay `Offer`(record_only.go:52-58, normal_relay.go:43-58) — 사이클이 원격 전송을 기다리지 않는다는 설계 전제는 맞다.
- 새 `EventType` 은 `criticalEvents`(internal/obs/event.go:357-392)에 넣지 않는 한 `SeverityOf` 가 normal(:395-400) — H2 (a) 가 기본값으로 성립. 변이 ⑥ 은 표 등재로 반증 가능.
- `time.RFC3339Nano` 해독(go1.26.5 실측): `.000+09:00` · 무소수 · 9 자리 소수 · `Z` 모두 성공, `+09:00` 은 같은 절대 시각으로 UTC 정규화. `+0900`·공백 구분·오프셋 없음은 실패. `0001-01-01T09:00:00+09:00` 은 `IsZero()==true`(「모름」으로 섞임 — 무해).
- `omitempty` 는 0 `time.Time` 을 생략하지 않음, `omitzero` 는 생략 — 실측(P1-1).
- PM: `STORY-TOS-a163.yaml` 추적됨, `FEAT-TOS-009.yaml:35` · `_registry.yaml:195` 연결 확인.
- 참조한 정본 요구 「관측 사이클은 알림의 원격 전송을 기다려서는 안 된다」(openspec/specs/exit-policy/spec.md:64) 존재 확인.

## 검증하지 못한 주장

- codegraph 1.6.0 의 `callers Prices`(비시험 18) 수치 — CLI 를 돌리지 않았다(grep 으로 핵심 함수만 대조).
- a112 결정 46 실측 4건의 원문(`archive/…a112…/review.md:774-835`)과 OpenAPI `PriceResponse.timestamp` nullable 원문 — 저자 인용을 재열람하지 않았다.
- C7~C11·C13~C16 각 행의 AST 좌표 — 이 change 의 동작 변화 0 범위라 재생성하지 않았다.
- `exit_e2e_test.go` 가 공식 어댑터를 httptest 로 물리는 기존 하네스를 갖는지 — 보지 않았다(P2-1 수리 비용 판단에 필요).
- tracer 가 `Run` 과 동시에 `ObserveOnce` 를 부를 수 있는지 — 확인하지 않았다(P1-4 의 단일 goroutine 소유 전제).

## 손으로 읽은 증거 표시

- `quote-time-evidence.md` §1 의 「`adaptTrades`·`adaptMarketIndicatorPrices`·`adaptMarketIndicatorCandles` 는 브로커 시각을 보존한다」 — AST 없음(저자도 「대조 사실」로 한정). 실제로는 **해독 없이 원문 문자열** 보존이다(market_reads.go:311·:436). P1-1(b) 의 선례로 쓸 수 있다.
- §3 C15·C16 — 저자가 AST 없음으로 명시, 분기 주장 없음. 문제 없음.
- design D4 의 「`noteJudged` 처럼 주기 집계에 적기만」 · 「`settleUnobserved` 와 같은 자리」 — 위치 주장은 원문으로 맞음(exitloop.go:518·523). 분기 주장은 아님.
- 그 밖의 분기·반환 주장은 전부 `enumerate-output.md` 좌표에서 왔고 재측정과 일치했다.
