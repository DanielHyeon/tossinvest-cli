# a163 review

## 0.5 proposal-freeze 리뷰 — 2026-10-21

- 리뷰: 독립 적대 리뷰어(저자 아님), 읽기 전용. 원문 전체는 `analysis/freeze-review-2026-10-21.md`(정본 — 이 절은 판정과 처분만 옮김).
- 대상: 커밋 `60033240a`(change 문서) · base `f7d80c348`. `internal/` 은 base 와 diff 0.
- **판정: APPROVE-WITH-CHANGES** — P0 0 · P1 4 · P2 5 · P3 7. 방향(사용자 결정 A, H1~H7)은 재론하지 않음.
- 처분 로트: 2026-10-21 문서 로트(production Go 편집 0). 각 P1 은 반영 전에 현재 HEAD 코드로 재검증함(아래 「재검증」).

### 필수 질문 — 「이 diff 의 어느 줄이 판정 자격을 바꿀 수 있나」

리뷰 답(원문 그대로 요약): 셋 — ① `adaptPrices` 의 시각 줄(행 탈락·오류 반환이면 판정 0) ② `observe` 의 `observedQuote{…}` 리터럴
(필드 혼동이면 B7 필터·identity·원장 시각이 바뀜) ③ `ObserveOnce` 의 보고 호출 줄(순회 안 이동·패닉이면 뒤 포지션·다음 주기 손절이
늦거나 사라짐). 처분 뒤: ① 은 해독 없는 대입 한 줄로 줄어 행 탈락·오류 분기가 구조적으로 없고(design D3) e2e 2.1~2.3 + 변이 ⑦⑧ 이 지킴,
② 는 렌더 해시·`SourceTimestamp` 금지 census(design D7) + 변이 ④b⑨⑩, ③ 은 순회 밖 한 줄 + recover 경계 + 패닉 주입 시험(2.10) + 변이 ②⑪.

### 처분표

| # | 요지 | 처분 | 반영 위치 |
| --- | --- | --- | --- |
| P1-1 | `time.Time`+`omitempty` 는 0 값을 `0001-01-01T00:00:00Z` 로 내보냄 → WTS 갈래 `tossctl quote --json` 이 바뀜(안전 3) | **반영** — 원문 문자열 `SourceTimestamp string` · `source_timestamp,omitempty`(리뷰 안 (b)). (a) `omitzero` 대신 (b) 를 고른 이유: 어댑터 새 분기 0 · 해독이 recover 경계 안 · null/해독 실패 구별 · 같은 파일 선례 | design D2·D3 · proposal 「무엇을」1·「영향」 · spec official-read-compatibility · tasks 2.0·2.5·3.1 · 변이 ⑬ |
| P1-2 | `source_sha256` 은 파일 해시 → 「diff 0」 충족 불가, 헬퍼 우회 못 막음 | **반영** — 함수 열 개 렌더 해시 시험(기대값 `analysis/render-hash/render-hash-base.txt`) + `SourceTimestamp` 금지 census + e2e | design D7 · tasks 1.2·3.5 · 변이 ⑨⑩ |
| P1-3 | 나이 로그의 생산 싱크 없음(`opts.Log` 생산 nil, `RecordOnly` 는 필드 지움) | **반영** — `ExitObserverOptions.SourceAgeLog` + `engineRuntime` 배선 + 구조·행동 배선 시험 | design D4 「생산 싱크」 · proposal 「영향」 · tasks 1.1·2.11·3.4 · 변이 ⑫ |
| P1-4 | 청산 goroutine 의 새 코드에 패닉 격리 없음(감독자 recover 없음) | **반영** — 보고 진입 함수 recover 경계, 순회 안 새 코드 0, 래치 맵 단일 goroutine 소유, `-race` 목록 등재 | design D4 「패닉 격리」·「단일 goroutine 소유」 · D6 · spec exit-policy 시나리오 · tasks 2.10 · 변이 ⑪ |
| P2-1 | 2.1~2.3 이 편집 전에 컴파일 안 됨 | **반영** — 공식 어댑터 + httptest(`exit_e2e_test.go` 기존 하네스) e2e 로 편집 전 GREEN | tasks 2.1~2.3 |
| P2-2 | 변이 목록 부족 | **반영** — ⑦~⑯ 추가, ④ 를 ④a(어댑터)·④b(`observe` 리터럴)로 분리 | tasks 4.1 |
| P2-3 | 매 주기·포지션마다 로그(17,280 줄/일/포지션) | **반영** — 구간 전이 때만 한 줄(a090 전이 줄 모양), spec 문구도 전이 기준으로 | design D4 「로그 빈도」 · spec exit-policy · tasks 2.12 · 변이 ⑭ |
| P2-4 | 래치 수명·정리 기준 없음 | **반영** — a090 `held` 기준 정리, 판정 안 된 포지션은 래치 불변, 맵 크기 0 단언 | design D4 「래치와 정리」 · tasks 2.6 · 변이 ⑮ |
| P2-5 | 일반 등급 버퍼(64) 공유, 재시작 직후 몰림 | **반영** — 주기당 8건 상한(넘는 것은 다음 주기), 버퍼 가득 참 시 버림·로그 줄은 남음을 명시 | design D4 「주기 상한」 · spec exit-policy · tasks 2.6 · 변이 ⑯ |
| P3-1 | proposal:23 「오래된 관측을 기록·알린다」 ↔ H1 | **반영** — 「나이를 기록, 시각 부재만 알림」 | proposal 「왜」 끝 |
| P3-2 | H2 의 보고 끄기 여부 미결 | ~~잔여(사용자 확인 대기) — 「토글 없음」~~ → **닫힘 2026-10-22**: 사용자 결정 「토글 필요」, a163 안 · 기본 OFF(아래 「0.8 설계 동결」) | design D8 · 「사람 결정 기록 — 사용자 결정 2026-10-22」 · spec exit-policy 토글 요구사항 · proposal 「무엇을」6·안전 3·§0.5 행 · tasks 0.8·2.13~2.16·3.6·변이 ⑰~㉒ |
| P3-3 | null 과 해독 실패를 한 사유로 합침 | **반영** — 로그 구간 `missing`/`unparseable` 분리(알림 사유는 H4 대로 하나) | design D4 「분류」 · spec 시나리오 · tasks 2.12 |
| P3-4 | 나이의 `FetchedAt` 이 어느 값인지 | **반영** — `observedQuote.FetchedAt`(대체 경로 `exitloop.go:867-870` 포함) | design D4 「나이」 · tasks 2.12 |
| P3-5 | 결과를 `review.md` 에 | **반영** — 이 파일 | `review.md` · tasks 0.5 |
| P3-6 | tasks 2.9 가 2.8 앞 | **반영** | tasks §2 |
| P3-7 | 기존 위험: `apiPrice.Timestamp` 디코드 타입 변경 시 `Prices` 전체 오류 | **잔여(기록)** — a163 이 넓히지 않음 | design D3 끝 · tasks 5.1 |

### 재검증 (처분 로트, 2026-10-21, HEAD `60033240a`, go1.26.5)

- P1-1: `internal/domain/models.go:421` 의 `FetchedAt`(`time.Time`, 태그 `fetched_at`, `omitempty` 없음)가 유일한 시각 필드. 스크래치 실측: `time.Time`+`omitempty`
  → `{"a":"0001-01-01T00:00:00Z"}`, `omitzero` → `{}`, `string`+`omitempty` → `{}`. `tossctl quote --json` 경로 `cmd/tossctl/quote.go` →
  `hybrid.Client.GetQuote`(`internal/hybrid/client.go:117-130`, WTS 갈래 `c.Client.GetQuote`) → `output.WriteQuote` → `writeJSON(quote)`
  (`internal/output/quote.go:27-28`). 생산자 `internal/client/quote.go:115`(`quote := domain.Quote{…}`)·`adaptStocks`
  (`internal/official/stock_reads.go:81-97`) 둘 다 새 필드를 채우지 않음 — 리뷰 주장 확인. 선례: `internal/verifylive/record.go:254`(omitzero),
  `internal/official/market_reads.go:311`·`:436`, `internal/domain/models.go:721`(원문 문자열).
- P1-2: Manager 확인(번들 9개 `source_sha256` = `sha256(exitloop.go)`) 인용. 처분 로트에서 렌더 해시를 실측해 `analysis/render-hash/` 에
  커밋(생성기 `render_hash.go.txt`) — 그 과정에서 `judge`(`alertdelivery.go`)·`record`(`strategy_lane_runtime.go`) 동명 함수를 찾음(design D7 에 반영).
- P1-3: Manager 확인 인용 + 재확인 — `exitloop.go:188-191`(Log 생산 nil 주석, `UnobservedLog` 선례), `cmd/tossctl/engine.go:671-681`
  (`UnobservedLog`·`ExitOrderLog` 만), `internal/obs/record_only.go:51`·`:106-108`(`withoutFields`), `exitwiring.go:333·349`. 배선 시험 선례
  `cmd/tossctl/a129_engine_runtime_exit_order_log_test.go`.
- P1-4: AST 로 확인(`analysis/ast/enumerate-output.md`, 이 로트에서 `engine--runtime.run` · `engine--exitobserver.run` ·
  `engine--exitobserver.writeexitorder` 등 7개 추가 생성): `Runtime.Run`(`runtime.go:288-379`) go 3 · defer 4(`cancel`·`wg.Done`×2·
  `supervisorWG.Done`) · `recover` 호출 0 — 루프 goroutine `:306-309` 는 `defer wg.Done()` 뿐. `ExitObserver.Run`(`exitloop.go:395-405`)
  defer 0 · `recover` 0. 선례 `writeExitOrder`(`exit_order_log.go:259-270`) 의 `recover` 호출 `:264`(AST `RECOVER` 행).
- 리뷰가 검증하지 못한 항목 중 둘을 닫음: ① tracer 동시 호출 — `tracer.go:297` 이 `ObserveOnce` 를 부르나 `NewTracer`(`tracer.go:239`)
  비시험 호출자 0(grep 전수) → 생산 유일 호출자 `exitloop.go:400`, 단일 goroutine 소유 전제 성립. ② e2e 하네스 — `exit_e2e_test.go`
  `newE2EBroker`·`servePrices`·`buildE2EStack`(`official.WithBaseURL(srv.URL)`)가 있고 `e2eClockDomainPrices` 는 `FetchedAt` 만 덮음.
  `servePrices` 는 `map[string]string` 이라 null·키 없음 본문은 변형이 필요(tasks 2.1 에 적음).

### 검증하지 못한 것 (이 로트)

- codegraph 1.6.0 `callers Prices`(비시험 18) 수치 — 리뷰와 같이 돌리지 않음.
- a112 결정 46 원문·OpenAPI `PriceResponse.timestamp` nullable 원문 — 재열람 안 함(저자 인용 그대로).
- 렌더 해시 시험·census 시험·배선 시험 자체 — 문서 로트라 아직 없음(tasks 1.2·2.11 이 만든다). `analysis/render-hash/render-hash-base.txt` 는
  기대값의 원천일 뿐 시험이 아니다.

### 사용자 확인 대기 잔여 — 2026-10-22 둘 다 닫힘

1. ~~H2 보충 — 보고 끄기 토글을 두지 않는다는 설계 판단~~ → 사용자 결정 2026-10-22 「토글 필요」로 뒤집힘(아래 처분 1).
2. ~~design D2 의 필드 형태 변경(`source_at` time → `source_timestamp` string)~~ → 사용자 결정 2026-10-22 승인(아래 처분 2).

## 0.8 설계 동결 — 사용자 결정 2026-10-22

- 경로: Manager 경유 사용자 회신(2026-10-22). 범위 판단(토글을 a163 안에 둘지)은 Manager 결정.
- 성격: 사람 결정 기록 + 문서 로트(production Go 편집 0). freeze 리뷰 재론 아님 — 리뷰가 남긴 사용자 확인 대기 둘을 닫고 수치를 승인받음.

| # | 사용자 결정 | 처분 | 반영 위치 |
| --- | --- | --- | --- |
| 1 | H2 보충: exit-loop 체결 시각 보고 토글 **필요**(「토글 필요」). Manager 범위: a163 안(별도 Story 아님), **기본 OFF** — OFF = 오늘 동작(새 로그 줄 0 · `source_missing` 알림 0). 토글은 보고·로그·알림 경로만 막고 어댑터의 `source_timestamp` 원문 보존은 막지 않음. 생산 ON flip 은 사람의 행동 | **반영** — 기존 세 패턴 재사용(새 설정 메커니즘 0): ① 저장 = engine 블록 bool 키 `engine.exit_source_age_report`(선례 `Engine.Autostart` `internal/config/engine.go:48·265·328-330`, raw `*bool`, omitempty) ② 읽기 = 기동 1회, `Context.ExitObserver` 가 옵션을 덮어씀(선례 `exitwiring.go:355` `NotificationsEnabled`) ③ audit = `recordGateSettings` 끝 행 + 새 action `exit.source_age_report`(선례 a136·a138·a139 끝 행 `interlock.go:795-803`). 판단 자리 = 보고 진입 함수의 recover 경계 바로 뒤 — `ObserveOnce` 분기 수·판정 함수 렌더 해시 불변, census 가 `SourceAgeReport` 출현을 셈. 콘솔 스위치·config 쓰기 코드 0 | design D8 · D4 머리 주석 · D7 2 · 위험 · spec exit-policy 「체결 나이 보고는 기본 꺼진 운영 토글 뒤에 있다」 · proposal 「무엇을」6·「영향」·안전 표 · tasks 0.8·1.1·1.2·2.13~2.16·3.6·4.1 ⑰~㉒·4.2·5.1 |
| 2 | H7 형태 변경 승인: `SourceTimestamp string` · json `source_timestamp,omitempty` 가 `source_at` time 을 대체 | **반영** — D2 형태가 정본, 2026-10-21 H7 행 문자열은 대체됨으로 표시 | design 「사람 결정 기록 — 사용자 결정 2026-10-22」 |
| 3 | 수치 승인: 로그 구간 경계 5분 · 1시간, 한 주기 알림 최대 8건 | **반영** — 사람 승인 값으로 표시(바꾸려면 새 사람 결정) | design 결정표 · D4 「분류」·「주기 상한」(값 불변) |

### D8 근거 (AST · grep, 2026-10-22, HEAD `c8d5db19b`)

- AST 셋 추가: `analysis/ast/config--mergeengine.ast.json`(분기 5 · 반환 2 — B4 `:342` `raw.AutomationGate == nil` 조기 반환) ·
  `engine--recordgatesettings.ast.json`(분기 6 · 반환 2 — B5 range · B6 `RecordChange` 오류 → `return err`) ·
  `engine--context.exitobserver.ast.json`(분기 6 · 반환 4 — `:342`·`:355` 덮어쓰기는 어느 분기 안에도 없음). `enumerate-output.md` 재생성: AST 32 · STALE 0.
- 주기 재읽기 선례 없음: `internal/app/engine` 비시험 파일의 설정 로드는 `engine.go:370` 한 줄(grep `config.NewService`·`.Load(`·`LoadRaw`).
  `LoadRaw*` 호출자는 전부 콘솔 저장 어댑터(`cmd/tossctl/*settings.go`·`engineautostart.go`·`soakautostart.go`).
- 공개 스키마 engine 블록 `additionalProperties: false`(`schemas/config.schema.json:95-98`) → 속성 추가 필요. a136 이 키 추가에 `schema_version` 5 를 유지한 선례.
- `recordGateSettings` 호출자: 생산 1(`engine.go:472`) · 시험 4(`a074_notification_audit_test.go:55·84·105·124`). a074 시험은 앞부분 행만 단언(`:146` `len(entries) < len(want)`) — 끝 덧붙임과 양립.

### 검증하지 못한 것 (0.8 로트)

- 토글 시험·census 확장·audit 행 시험 자체 — 문서 로트라 아직 없음(tasks 2.13~2.16 이 만든다).
- 공식 갈래 `tossctl quote --json` 의 `source_timestamp` 키는 토글과 무관하게 생긴다 — H7(a) 사용자 승인과 Manager 범위(어댑터 보존은 토글 밖)에 따른
  것이고, 이 키가 upstream 공식 갈래 출력 계약에 주는 영향은 이 로트에서 다시 재지 않았다(2026-10-21 D2 판단 그대로).

## 1.3 Pre-Edit 선언 — 2026-10-22 구현 로트(격리 worktree `/tmp/tossos-a163-impl-20261022`, 브랜치 `a163-impl`, HEAD `1eab9b931`)

```text
Pre-Edit Gate:
- change id / task id: a163-quote-freshness-uses-trade-time / 3.1~3.6 (RED 2.0~2.16 선행)
- 대상 심볼(패키지.함수):
  · official.adaptPrices (리터럴 대입 한 줄)
  · engine.ExitObserver.observe (observedQuote 리터럴 필드 하나) · engine.ExitObserver.ObserveOnce (순회 뒤 호출 한 줄)
  · cmd/tossctl.engineRuntime (SourceAgeLog 옵션 키 하나)
  · config.mergeEngine (B4 앞 if 하나) · engine.recordGateSettings (끝 인자 + 끝 행) · engine.Context.ExitObserver (덮어쓰기 한 줄)
  · engine.NewContext (recordGateSettings 호출 인자 하나) · a074 시험 함수 넷(호출 인자 false) — issues.md I3
  · 새 코드(새 파일): internal/app/engine/exit_source_age.go · 새 타입 필드(domain.Quote.SourceTimestamp · observedQuote.SourceTimestamp ·
    ExitObserverOptions.SourceAgeLog/SourceAgeReport · config.Engine.ExitSourceAgeReport · rawEngine) · 새 상수(audit.ActionExitSourceAgeReport ·
    obs 이벤트 둘) · schemas/config.schema.json 속성 · Makefile RACE 목록
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-callers-0.7-20261022.txt (codegraph 1.6.0, make sdd-sync 직후)
- CodeGraphContext 후보와 evidence reconciliation: evidence-reconciliation.md 「tasks 0.7 재질의」 — 불일치 2(issues.md I1·I2, 시험 범위만)
- 기존 동작 파악 근거: HEAD 원문 + AST 32(analysis/ast) + FLM 번들 12(analysis/function-logic) + 기존 시험(exitloop_test · exit_e2e_test ·
  a090 · a111 · a129 · a074 · config engine_test · official market_reads_test)
- Function Logic Map / Branch Test Map: analysis/function-logic/ 12 번들(편집 전 revision current = base 소스) — 편집 뒤 최신화(tasks 3.5)
- upstream 상속 테스트 영향: yes — domain.Quote JSON(tossctl quote --json). 회귀 방지: 2.0 골든 둘(WTS · adaptStocks, 편집 전 GREEN)
- 실패 테스트 선행 작성: yes — 고정 시험(1.2 · 2.0~2.4)은 편집 전 GREEN 실측, RED(2.5~2.16)는 GREEN 전 실패 실측을 analysis/red-green/ 에 기록
- 설정·DB·journal 변경과 rollback: 설정 키 하나 추가(engine.exit_source_age_report, 기본 false, omitempty). journal schema 변경 0.
  rollback = 이 브랜치 커밋 되돌림. 키가 남은 config 파일도 옛 바이너리가 읽음 — internal/config 의 로더는 encoding/json 기본 디코드이고
  DisallowUnknownFields 사용 0(grep 2026-10-22), 그래서 모르는 키는 버려짐(= 꺼짐과 같음)
- 안전 불변식 §0 위반 여부 검토: 통과 — 판정 경로 새 분기 0(렌더 해시 · census · ObserveOnce 분기 수), 손절 즉시성 불변(e2e 2.1~2.3),
  OFF = 오늘 동작(2.13 · 2.14), 공식 API 호출 0 추가, 토글 쓰기 코드 0(사람 flip), 계좌 ref 미기재(2.8)
```

## 4.4 구현 리뷰 — gstack `/review` · 2026-10-22 (구현 저자와 다른 컨텍스트 셋 + Codex)

- 대상: `1eab9b931..01d1fc34d`(편집 커밋 셋), 뒤이은 처분 커밋 `01d1fc34d` · `ff232b3cf`.
- 보이스: ① 전문 리뷰(시험 · 유지보수 · 보안 렌즈, Claude 하위 에이전트, 읽기 전용) ② 적대 리뷰(Claude 하위 에이전트, 시험은 요약 모드) ③ Codex 적대
  리뷰(`codex exec -s read-only`, 정적 — 시험 미실행) ④ 저자 critical pass(SQL · 동시성 · LLM · 셸 · enum 범주 — 해당 0: SQL · LLM · 셸 없음,
  새 상수 셋(`obs.EventExitSourceAge` · `obs.EventExitSourceMissing` · `audit.ActionExitSourceAgeReport`)의 소비자는 `criticalEvents` 표 ·
  이벤트 유일성 시험 · audit action 소비자 0(grep) — 빠진 갈래 없음).
- **판정: P0 0 · P1 1(Codex) · P2 3 · P3 다수. 손절 판정 · 제출을 바꾸는 결함은 셋 모두 못 찾음.** 아래 처분 뒤 남은 차단 사유 없음(Manager 확인 요청 I5~I8).

| # | 출처 | 요지 | 처분 |
| --- | --- | --- | --- |
| R1 | 전문(시험) | 거둔 패닉의 고정 문구 줄을 아무 시험도 안 셈 — 호출 삭제가 살아남음 | **수리** — 패닉 시험이 줄 수를 셈(로거 패닉 0 · 알림기 패닉 1), 변이 ㉓ CAUGHT |
| R2 | 전문(유지) · 적대 P3-3 | 결정적 패닉이면 매 주기 Warn 한 줄 영구 | **수리** — 패닉 연속의 첫 주기에만(`sourceAgePanicLogged`, 패닉 없는 보고가 풂), 변이 ㉔ CAUGHT |
| R3 | 전문(보안) · 적대 P3-2 | 원문 64 바이트 바이트 자르기가 글자 중간을 자르고 개행 · 제어문자를 알림 본문(renderFields → ntfy)에 통과 | **수리** — 글자 경계 자르기 + 인쇄 불가 글자 `?`(issues I4), `TestA163TheRawBrokerTimeIsBoundedPrintableAndValidUTF8`, 변이 ㉖ CAUGHT |
| R4 | 전문(시험) | ahead 구간 무알림을 아무 시험도 안 셈 | **수리** — 분류 시험이 구간마다 알림 수 단언, 변이 ㉕ CAUGHT |
| R5 | 전문(시험) | missing → unparseable 래치 동작 미고정 | **수리** — `TestA163MissingThenUnparseableIsOneReason`(spec 「같은 사유는 한 번만」 — 한 알림, 구간 줄은 남음) |
| R6 | 전문(유지) | `le_15s` 이름 · 상한 8(= 64 의 1/8)이 다른 패키지 상수에 묶임 | **수리** — `TestA163TheBucketNamesStillMatchTheirSourceConstants` |
| R7 | 전문(시험) · 적대 P3-7 | 공식 /prices 갈래 JSON 에 새 키 — 고정 시험 없음 | **수리** — `TestA163TheOfficialPricesQuoteJSONCarriesTheBrokerTime`. 키 생김 자체는 H7 사용자 승인 범위(토글 무관) |
| R8 | 전문(시험) · 적대 P3-6 | 금지 census 가 엔진 · 공식 어댑터만 — flatten · 진입 · 전략의 미래 읽기 못 막음 | **수리** — census 를 `cmd` · `internal` 모듈 전체 비시험 파일로(정확한 자리 9), 변이 ㉗(flatten 이 브로커 시각을 읽음) CAUGHT |
| R9 | 전문(시험) · 적대 P3-5 | Makefile 주석이 `-race` 가 단일 소유를 증명한다고 과장 | **수리** — 주석 정정(근거는 호출 구조) |
| R10 | Codex P2 | 거대 원문을 매 주기 끝까지 해독(파서가 소수부 자릿수 무제한) | **수리** — 64 바이트 초과는 해독 없이 unparseable(RFC3339Nano 최장 35 — 정상 입력 거부 0), 변이 ㉘ |
| R11 | Codex **P1** | 보고 줄 쓰기는 동기 — 로그 싱크가 멈추면 청산 goroutine 이 멈춰 다음 주기 손절 판정이 없음 | **잔여(issues I6, Manager 판단 요청)** — 같은 goroutine 의 a090 `UnobservedLog` · a129 `ExitOrderLog` · a091 `ZeroFloorLog` 가 이미 같은 엔진 로거에 같은 방식으로 씀(새 위험 종류 아님). 그 주기의 판정 · 제출은 끝난 뒤이고 토글 OFF(기본)면 줄 0. design D4 「생산 싱크」 가 엔진 로거를 정했으므로 비동기 싱크는 설계 변경 — 이 change 에서 하지 않음 |
| R12 | 적대 **P2-1** | 시각 부재 ↔ 복귀 진동마다 재무장 → 무한정 알림(Relay 는 Key 중복 제거 없음), 주기당 8 상한은 주기 간 누적을 못 막음 | **잔여(issues I5)** — 재무장 규칙은 design D4 「래치와 정리」(사람 승인 설계). 토글 ON 전 사람 결정 필요(쿨다운 · 히스테리시스). 상수 주석을 「한 주기 안에서」로 정정 |
| R13 | 적대 P3-1 | 15 s 경계 히스테리시스 없음 — 중간 유동 종목이 le_15s ↔ le_5m 를 오가며 줄 다수 | **잔여(issues I5)** — 구간 경계는 사람 승인 값(2026-10-22). H6 분포 수집 목적상 전이 줄이 곧 표본 |
| R14 | Codex P2 | 쓰기 실패(ENOSPC 등)여도 래치가 전진해 그 구간 줄이 사라짐 | **잔여(issues I7)** — `obs.Logger` 는 쓰기 결과를 돌려주지 않음(기존 a090 · a129 줄 동일). 판정 무관 |
| R15 | Codex P3 | 마지막 포지션이 닫히면(무보유 경로) 래치 맵이 다음 보고까지 남음 | **수용(설계대로)** — design D4 「무보유 경로에서는 보고를 부르지 않는다」. 크기는 직전 보유 포지션 수로 유계, 다음 보유 주기 또는 재시작에 정리 |
| R16 | 적대 P3-4 | OFF 여도 업그레이드 뒤 첫 기동에 audit 행 `"" → "false"` 하나 | **수용** — design D8 「audit」 의 `RecordChange` 선례(a136 · a138 · a139 와 같음). OFF 의 정의(줄 · 알림 · 래치 0)는 audit 파일을 포함하지 않음(issues I8) |
| R17 | 적대 · Codex | 출력 `source_timestamp` 키는 토글과 무관 | **수용** — H7 · D8 사용자 승인 |

- 적대 리뷰가 확인한 비결함(요약): recover 경계가 해독 · 래치 · 로그 · 알림 전부를 감쌈 · 뮤텍스는 defer Unlock 이라 거둔 패닉이 잠금을 남기지 않음 ·
  생산 `ObserveOnce` 호출자 `Run` 하나 · 보고는 판정 순회 뒤 · `stableObservationID` 는 필드를 명시 나열해 새 필드가 identity 를 안 바꿈 ·
  Key 충돌 없음 · 계좌 ref 없음(position_id 는 sha256 접두이고 로거가 가림) · `Time.Sub` 포화로 age_seconds 넘침 없음 · 윤초는 unparseable.
- 리뷰 원문: 하위 에이전트 보고는 이 세션의 메시지(파일 아님), Codex 출력은 세션 스크래치. 결정만 이 표로 옮김.
