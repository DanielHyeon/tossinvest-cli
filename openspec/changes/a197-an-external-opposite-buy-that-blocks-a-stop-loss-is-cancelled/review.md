# a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled Review

## 계약 작성 (2026-11-23)
- 작성: Manager(Fable). 근거 조사: 읽기 전용 조사원(Opus) — 분기 주장의 AST 는 세션 스크래치 산출물 인용이며 change 안 사본·HEAD 재확인은 tasks 1.3.
- `openspec validate a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled --strict`: valid (rc 0, 2026-11-23 HEAD 1ea261100).
- proposal-freeze 리뷰: 미실시(tasks 1.7~1.8). → 아래 「proposal-freeze 적대 리뷰 r1」.
- Function Logic Map: 편집 대상 기존 함수의 번들은 tasks 1.6 에서 만든다.

## 1단계 계약과 증거 — **상태: 보류**(Manager 결정 2026-11-23, 선행 a198 · design §7 · issues.md) (2026-11-23, 팀메이트 Opus, worktree `~/tossos-work/a197-impl`, 브랜치 `feat/a197-external-opposite-buy-cancel`)

- base `064f59fd99c11ea3a19adcdb804bbcdfed2470cb` 고정(`capture_change_base.py` rc 0), PM `generate_master_tracker.py --check` rc 0.
- 산출물: `analysis/code-context/`(CodeGraph 1.6.0 · CGC 0.5.1 색인 0 · GBrain 낡음) · `analysis/cited-ast/`(AST 28 + `README.md` 주장 대조) · `analysis/probe/`(일회성 probe, 버리는 worktree 에서 실행 — 이 브랜치 Go 편집 0) · `analysis/edit-surface.md`(1.4 · 1.5) · `analysis/function-logic/`(pre-edit 번들 3: `ExitObserver.submit` · `ExitObserver.alertProposalRefused` · `Context.ExitObserver`). `check_analysis.py --change` rc 0(required 0).
- 문서 정정(결정 무변경, 표시 「AST: …」): proposal 표 1~5행 좌표·근거(`alertProposalRefused` `1901-1918` · 호출 자리 2, default `1602-1615`), design D4 의 R7 인용(OPEN 목록은 R7 에 없음 — 값 3s 는 그대로).
- **Manager 결정 필요(문서 무편집)**: ① D6 — exit 루프의 결정 발행기(`RiskGuardian.IssueReduction`, Side SELL · Kind PLACE)로는 취소를 낼 수 없음, probe 실측 BUY·SELL 모두 `Prepare` 거절 · 브로커 0(기존 a094 길 치우기도 같은 결함 — 생산에서 엔진 주문 취소 0). ② D2 — 「접수 확정 귀속」 판독 없음, 있는 것은 `EngineBrokerOrderIDs`(상태 무관 · 계좌 전체). ③ 번호 없는 park(UNRESOLVED_IN_DOUBT) 엔진 매수는 `PendingAttempts` 에도 귀속 집합에도 없어 「외부」로 분류됨. ④ D8 — a163 토글 선례의 audit 는 a166 census 고정 함수 `NewContext` 편집을 요구.
- `openspec validate a197-… --strict`: valid(rc 0, 정정 뒤).

## proposal-freeze 적대 리뷰 r1 (2026-11-23)

- 보이스: codex(다른 계열), `codex exec -s read-only`, 적대적 Eng 관점. 원문 `analysis/review/proposal-freeze-r1-codex.md`. 판정: 「freeze·구현 진행 승인 어려움 — P0 0, P1 7, P2 4」.
- 처분 규칙: 사실 정정은 반영, 새 가드·정책·범위·결정 변경은 반영하지 않고 Manager 결정 대기.

| # | 등급 | 발견 요지 | 작성자 확인 | 처분 |
|---|---|---|---|---|
| 1 | P1 | D6 의 기존 취소 경로가 생산 배선에서 성립 안 함(RiskGuardian 결정 = PLACE 키 · SELL) | probe 로 독립 실측 일치 | Manager 결정: 보류 — 선행 a198 로 분리(design §7 선행 1) |
| 2 | P1 | 부정 증거로 외부 판정 → park 된 번호 없는 엔진 매수가 취소 대상 | AST 일치(`recovery.go:30-33`) | Manager 결정: 보류 — 재개 시 반영(§7: 미종결·park attempt 있으면 열지 않음, 상태 무관 귀속) |
| 3 | P1 | 새 취소가 IN_DOUBT 면 다음 주기 `clearTheSymbol` 이 손절을 막고, CANCEL 은 세션 중 해소 안 됨(order-execution `:118`) → D7 · 시험 10 의 「다음 거절 주기 재시도」 불성립, 사람이 매수를 치운 뒤에도 손절 지연 | 작성자도 같은 위험 기록(README D7) | Manager 결정: 보류 — 선행 2(세션 중 CANCEL 해소 선행 또는 모호 시 단계 폐쇄 + critical, 재개 전 결정) |
| 4 | P1 | 3s 는 조회만의 상한 — 다수 취소 · 15s HTTP 기본 시한 · 예산 대기 누적 시 뒤 포지션이 `quoteUsable` 만료로 그 주기 판정을 건너뜀 | `official/client.go:22` 15s, `exitloop.go:531` 확인 | Manager 결정: 보류 — 재개 시 반영(§7: 단계 전체 시한 + 주기당 취소 건수 상한) |
| 5 | P1 | `ScanOrders` 는 `HasNext=true ∧ NextCursor=""` 를 성공으로 반환 → D3 「끝까지 못 읽으면 0」 과 반대 | `indoubt.go:754` 확인 — README (d) 의 「부분 목록 반환 없음」 서술이 이 경우를 놓쳤음(README 정정) | Manager 결정: 보류 — 재개 시 반영(§7: 「다음 쪽 있음 ∧ 커서 없음」 = 읽지 못함) |
| 6 | P1 | 행 → 대상 → 취소 요청의 식별 보존 계약 공백(종목 필터 응답 미검증 · 취소 HTTP 는 번호만 · `parseTarget` trim) | 확인 | Manager 결정: 보류 — 재개 시 반영(§7: 주문 번호 바이트 보존, 종목·방향 행 값 재대조) |
| 7 | P1 | 취소 접수 → critical 기록 사이 crash 시 알림 영구 누락, 기동 해소 CONFIRMED 경로 미정 | 미측정(논증) | Manager 결정: 보류 — 재개 시 반영(§7: 취소 전에 「취소를 시도한다」 기록) |
| 8 | P2 | ADDED 만으로 기존 「길 치우기는 엔진 귀속만」 · 「손절 경로에 브로커 조회 안 더함」 SHALL NOT 과의 경계가 닫히지 않음 → MODIFIED 제안 | — | Manager 결정: 보류 — 재개 시 반영(§7: ADDED + 기존 요구 MODIFIED) |
| 9 | P2 | 「OFF = 한 글자도 다르지 않다」 vs D9(OFF 에서도 문장 변경), 「즉시 롤백」 vs 기동 때 읽는 토글 | 확인(`exitwiring.go:358` 기동 고정) | Manager 결정: 보류 — 재개 시 반영(§7: 꺼짐 = 조회 0·취소 0, 알림 문장은 토글 무관, 롤백 = 토글 끄고 재시작) |
| 10 | P2 | audit 배선이 `NewContext`(a166 census) 편집 요구 | 작성자 1.4 와 같음 | Manager 결정: 보류 — 재개 시 반영(§7: `NewContext` 무편집 자리 먼저, 없으면 census 갱신 Manager 승인) |
| 11 | P2 | 알림 key(타입 · 포지션 · 주문 번호)의 「한 번」 단위 불명 — 거래일 다른 번호 재사용 시 창 0 행이 새 알림을 묻음 | 미측정 | Manager 결정: 보류 — 재개 시 반영(§7: 단위 = (포지션, 주문 번호)) |

- 시험 전략 지적(가짜 게이트웨이가 생산 계약 실패를 가림 · 추가 변이 9종)은 원문 표 그대로 Manager 에게 넘김.
- 2회차: **미실시** — r1 의 처분이 전부 Manager 결정 대기라 계약 문서가 바뀌지 않았다(재리뷰할 개정이 없음). 결정 반영 뒤 r2 를 권함.

## Manager 결정 (2026-11-23)

- **보류.** freeze 하지 않는다. probe 가 잰 「exit 루프의 취소 결정 발행이 주문 경로에서 거절된다」는 a197 과 무관한 운영 결함이므로 별도 change a198 로 먼저 고치고, a197 은 그 착지 뒤 재개한다.
- r1 11건의 처분은 위 표대로(재개 시 반영할 결정은 design §7). r2 는 돌리지 않는다 — 재개 때 개정본으로 다시 리뷰한다.
- tasks 1.8 `[blocked]`, 막힌 이유·재개 조건은 `issues.md`.

## 재개 준비 (2026-11-25, 팀메이트 Opus, worktree `~/tossos-work/a197-impl`) — production Go · 시험 편집 0, 계약 본문(D1~D10) 개정 0

- **병합**: `git merge origin/main`(rebase 아님) — origin/main `b93ed133a` 을 `dcf33f5a8` 위로, 병합 커밋 `e7e3f22b4`. 충돌 0. `generate_master_tracker.py --check` rc 0(재생성 불필요).
- **base 재고정(사람 승인 절차 — Manager 승인 2026-11-25)**: `064f59fd9` → `b93ed133a`(병합의 main 쪽 부모 = origin/main). 귀속 실측: 옛 base 이후 이 change 디렉터리를 만진 비병합 커밋 2(`d6fe7be4b` · `dcf33f5a8`) 모두 `.go` 0 → 자기 Go 커밋 0. 옛 base 에서 `check_analysis --change` rc 1 · required 37(창의 착지 커밋 164 — 전부 형제 change 몫). 단독 커밋 `f18de89ef`(`base-commit.txt` 한 파일, 영수증 메시지). 재고정 뒤 required 0.
- **증거 재추출**(`analysis/cited-ast/README.md` §0): 인용 AST 28 중 무변 19 · 재추출 9(본문변경 2 — `Context.ExitObserver`(a199) · `ExitObserver.clearTheSymbol`(a198); 본문동일 7). a198 계열 신규 AST 16. pre-edit FLM 번들 3(`submit` · `alertProposalRefused` · `Context.ExitObserver`) `ast.json` 재추출 + 문서 좌표 이동(+3 / +11, 4 자리 원문 대조) + `Context.ExitObserver` B2 행 갱신 + risk report 재생성.
- **좌표만 고친 문서**(「좌표 재추출 2026-11-25」 표시): proposal 표 4 행(`:1650/1665` → `:1653/1668`, `1602-1615` → `1605-1618`, `else@1406 → clearDelay@1407` → `@1409 → @1410`, `1901-1918 · :1908 · :1530 · :1607` → `1912-1929 · :1919 · :1533 · :1610`). design · spec delta 무편집(좌표 없음).
- **편집 표면 갱신**(`analysis/edit-surface.md` §0): 구조 시험 3 무변, 새 구조 핀 1(`a198_clear_cancel_log_order_test.go:140`), a166 census 90 그대로(`NewContext` 고정 유지, a199 가 갱신한 선례), a087 · a167 상태 무변, a199 청산 전용 기동에서도 관측기 동작, rate limit 정정(대상당 사전 단건 조회 1 추가).
- **probe**(버리는 worktree `~/tossos-work/a197-resume-probe`, 1 회 실행 뒤 `git worktree remove`): `analysis/probe/resume-probe-output.txt`, 원본 `resume-zz_a197_resume_probe_test.go.txt`. 외부 BUY 취소 = CONFIRMED · 브로커 1; 5xx → IN_DOUBT → 손절 SELL · flatten 취소가 `symbol_mutation_in_flight`, D10 `closed=false`, 기동 해소(취소됨) → CONFIRMED 뒤 손절 접수, 판독 불가 → park + 진입 게이트 `unresolved_in_doubt`; 409 `already-canceled` → FAILED_CONFIRMED, `already-modified` → IN_DOUBT.
- **Manager 결정용 메모**: `analysis/resume/decision-2-indoubt.md` — Q1~Q5. 요지: 선행 2 는 a198 로 풀리지 않음(실측), 권고 (C)(조건 4), 대안 (A).
- 실행 안 한 범위: 전체 시험 · `make test` · `make gate` · `make lint` · `make sdd-sync`/`sdd-check`(CodeGraph fingerprint 미갱신) · 적대 리뷰 r2(개정본 없음) · CodeGraph/CodeGraphContext/GBrain 재질의(이번 회차는 AST · rg · probe 만). `mutation_attempts` 판독 SQL 전수(139 줄) 미대조. 엔진 API 취소의 KR 단건 판독 동작 미측정.
- 검증(2026-11-25, HEAD `f18de89ef` + 이 회차 문서 편집 워킹트리): `openspec validate a197-… --strict` → 「Change … is valid」 rc 0 · `python3 tools/logic-map/check_analysis.py --change a197-…` → base `b93ed133afde` → working tree, **required 0**, 「evidence complete or diff-proven exempt」 rc 0 · `python3 tools/pm/generate_master_tracker.py --check` → 「hierarchy and generated trackers are current」 rc 0.

## 재개 준비 2차 (2026-11-25, Manager 지시 — 선택지 (E) · SQL 전수 · 시한 구조) — production Go · 시험 편집 0, 계약 본문 개정 0

- Manager 1차 검증: 보고 수용, 결정 보류. 2차 지시: (E) 「모호해지면 즉시 park」 조사, (C) 조건 3(SQL 전수) 실측, 손으로 읽은 4 곳 AST 화, r1 #4 시한 구조 확인.
- 산출물: `analysis/resume/decision-2-addendum.md`. AST 17 추가(`analysis/cited-ast/README.md` §0.4, 합계 61). probe 2(버리는 worktree `~/tossos-work/a197-resume2-probe`, 1 회씩 실행 뒤 `git worktree remove`): `analysis/probe/resume2-engine-probe-output.txt`(엔진 조립 위 관측 주기) · `resume2-execgw-probe-output.txt`(게이트웨이 범위 — 1·2 회째는 probe 가짜 브로커의 번호 중복 결함이었고 고친 3 회째 원문), 원본 `resume2-zz_*.txt`.
- 요지: (E) 성립(park 뒤 다음 주기 손절 접수 · flatten 통과 · 대사/편입 무관), 대가는 그 종목 진입(해동까지)과 해소기 park 를 쓰면 계정 진입(재시작까지). 권고 (C) → (E)(래치 없는 park) 변경, 근거는 메모 E5. SQL 전수 104 자리: 바꿈 1 · 결과 바뀜/판정 무관 1 · 무관 102 · 미확인 0. 주기 끝 실행은 가능하나 단계 시간(유계 21.05 s + 토큰 잠금 무한)만큼 다음 주기 시작이 밀림.
- 실행 안 한 범위: 대사 루프 · 편입 루프 · 계정 래치 park 의 엔진 조립 위 구동, 401 재전송 · 예산 경합 · 토큰 대기 실측, 전체 시험 · make · sdd-sync/check · 적대 리뷰 · CodeGraph 재질의.
- 검증(2차 편집 워킹트리, HEAD `224fa5ca0`): `openspec validate … --strict` 「is valid」 rc 0 · `check_analysis.py --change` required 0 「evidence complete or diff-proven exempt」 rc 0 · tracker `--check` 「current」 rc 0.

## Manager 결정 (2026-11-25)

> 원문(Manager 메시지, 2026-11-25) 그대로 옮김.

Manager 결정(2026-11-25) — 2차 보고 수용(정본 `order-execution :44 :46 :125 :131` 과 design D4·D5·D7 원문을 직접 읽고 판정).

1. **선행 2 = (E), 종목 범위 park.** 이 단계가 낸 외부 매수 취소가 모호(IN_DOUBT)로 돌아오면 그 자리에서 `Attempt.ResolveUnresolved` 로 park 한다. 계정 래치(`Gate.Block`)는 걸지 않는다. (C)·(A)·(B)는 기각.
   - 근거: (가) 손절·청산 칸에서 오늘보다 늦어지는 칸 0(resume2 엔진 probe), 판정 본체 `unsettledFor` 와 `:46` SHALL 무변. (나) 계정 래치를 안 거는 이유 — 오늘(토글 꺼짐)도 같은 불확실성(외부 매수가 살아 있는지 모름)이 래치 없이 존재하고, 계정 래치는 엔진 프로세스 종료로만 풀리는데 엔진 정지는 손절 부재 창을 만든다(보호 UNWIRED). 정본 `:44` 의 park 기본 의미(「해당 심볼 신규 진입 차단, 운영자 해소만」)와 같다. (다) 거부하게 되는 정상 입력 = 운영자 해동 전까지 **그 종목의 신규 진입**. 이것을 proposal·design·spec 에 명시한다.
   - 정본 delta: `:44`(park 에 이르는 경로 추가)·`:125`·`:131`(이 단계의 park 는 계정 래치를 걸지 않는다)을 MODIFIED 로. `:46` 은 건드리지 않는다.
   - critical 알림 본문에 운영자 절차를 적는다: 앱에서 그 매수 확인 → `tossctl engine attempt-resolve`. 그 명령 도움말이 발주 기준이라 취소 attempt 에서 뜻이 달라지는 점은 운영 문서(design §5)에 번역해 적는다. 명령 코드 수정은 범위 밖.
   - park 불가 예외(원장 쓰기 실패로 DISPATCH_STARTED·ACKED 에 남은 취소 → 재시작까지 막힘)는 잔여 위험으로 design 에 공시한다. 새 가드를 만들지 않는다.
2. **D7 개정 — 같은 (포지션, 주문 번호)에는 취소를 한 번만 낸다.** 「거절·모호면 다음 거절 주기가 다시 연다」는 삭제(현재 코드에서 거짓이고, 거절이 반복되면 주기마다 조회+취소가 더해져 다른 포지션과 호출 예산을 깎는다). 결과가 접수·확정 거절·모호(park) 어느 것이든 그 번호에 다시 내지 않고, 치워지지 않았으면 critical 이 사람에게 넘긴다. 「이미 냈다」 표시는 새 상태·스키마 없이 원장의 그 단계 CANCEL attempt 행(target = 그 번호)으로 읽는 것이 가능한지 AST 로 확인해 설계에 적고, 불가능하면 대안과 함께 열린 항목으로 올려라. §4 시험 10 도 이에 맞춰 고친다.
3. **시한(r1 #4) — D5 유지(인라인, 별도 실행자 없음) + 두 가지.** ① 단계는 포지션 판정 경로 안이 아니라 **관측 주기의 끝**(`endExitOrderCycle` 계열)에서 실행한다 — 같은 주기의 다른 포지션이 시세 만료로 판정을 건너뛰는 일을 없앤다. ② 주기당 외부 취소 **1건** 상한. 「단계 전체 hard 시한」은 걸 수 없음을 D4 에 사실대로 적고, 네가 합산한 최악 추가 시간(유계 21.05s, 토큰 잠금 무상한, 다음 주기 시작이 그만큼 밀림)을 「비차단의 한계」로 공시한다. 보조 실행자 안은 기각 사유와 함께 design 에 남긴다(게이트웨이 종목 직렬화·순서 재설계 필요, 기존 길 치우기도 같은 성질).
   - 미체결 목록 조회의 반복 빈도(거절 주기마다 1 GET 인지, 이미 공표된 OPEN 스냅숏을 읽을 수 있는지)는 네가 코드 증거로 가장 작은 안을 골라 D4 에 적고 rate limit 계상(1.5)을 맞춰라. 고른 근거를 r2 가 보게 한다.
4. **§7 나머지(r1 #2·#5~#11)는 결정대로 반영**하되 네 Q5 의 전제 변경을 적용한다: #2 는 park 쪽 가드만 새로 필요(미종결 절반은 기존 코드가 보장 — 인용), #7 은 a198 D4 핀(판정 경로 안 쓰기 금지)과 충돌하지 않는 자리, #8 은 인접 Requirement 3개(`:885/:905/:917`) 포함, #9 는 청산 전용 기동에서도 단계가 돈다는 사실 공시(토글 꺼짐 = 조회 0·취소 0 은 그대로), #10 은 `NewContext` 무편집 자리 우선.
5. 토글 기본 false, 꺼짐 = 오늘 동작(D9 알림 문장만 토글 무관). 운영 flip 은 사람. 변하지 않는다.

## 재개 개정 (2026-11-25, 팀메이트 Opus) — 조건 3

- 개정: proposal(What Changes · 하지 않는 것 · Capabilities · Impact) · design(D1~D10 개정 + D11 신설, §2~§6, §7 「재개 반영 완료」, 열린 항목 O-1) · spec delta(exit-policy ADDED 7 · MODIFIED 4, order-execution MODIFIED 1 — 「IN_DOUBT 해소」) · tasks(1.8 해제 조건, 1.9, 2~4 단계) · issues(I1 재개 조건 진행).
- 결정 2 의 확인: 「이미 냈다」 표시는 원장 CANCEL attempt 행으로 읽을 수 있다(열 `kind`·`state`·`target_order_id` + intents `source`·`trading_day`, 스키마 무변). 기존 판독 함수 중 맞는 것이 없어(`ConfirmedCancelOf` CONFIRMED 만 · `FailedConfirmedCancelsForSymbol` FAILED 만 · `CloseInDoubtCancelOnTargetClose` id 한정) 새 판독 함수 1 개(새 파일)가 필요 — design D7.
- 결정 3 의 조회 빈도: 공표된 OPEN 스냅숏 없음(`filldetect/detect.go:376-380` 은 개수만) → 거절 주기마다 처리 포지션 1 에 대해 종목 필터 OPEN 1 GET(전량 1 요청, M7) — design D4 (가), 근거와 비용(계좌 예산의 약 7%) 공시.
- 편집 대상 기존 함수 5(FLM 번들 5 — 새로 2: `endExitOrderCycle` · `mergeEngine`). 열린 항목 O-1(토글 audit 실패 처리)은 Manager 결정 대기.

## proposal-freeze 적대 리뷰 r2 (2026-11-25, 마지막 회차)

- 보이스: codex(다른 계열), `codex exec -s read-only`, 적대적 Eng 관점, 기준 HEAD `6a3306f8c`(재개 개정 커밋). 원문 `analysis/review/proposal-freeze-r2-codex.md`. 판정 원문: 「proposal-freeze 보류 — P0 확인 없음, P1 3건·P2 5건. 기존 O-1도 미결이며, 현재 개정본의 재시작·한 번만·운영자 인계 계약은 닫히지 않았습니다.」
- 처분 규칙(r1 과 같음): 사실 정정은 반영, 새 가드·새 정책·범위 변경은 반영하지 않고 Manager 결정 대기.
- 리뷰가 결함으로 **확인하지 않은** 것(원문 끝 절): 직접 park 된 CANCEL 이 손절 · 익절 · 길 치우기 · flatten 을 막는 추가 경로, 정규화 오차단, 단일 호출자의 D7 경합, B1 앞 호출, MODIFIED 5 개의 정본 복사 누락.

| # | 등급 | 발견 요지 | 작성자 확인 | 처분 |
|---|---|---|---|---|
| 1 | P1 | IN_DOUBT 기록 → park 전 중단이면 기동 해소기의 `Resolver.park` 가 계정 래치(`indoubt.go:376-380`); park → 두 번째 critical 전 중단이면 결과 알림 누락 | 코드 일치(`recovery.go:266,300`, `PendingAttempts` 에 park 없음) | **사실 정정 반영**: design D11 「중단 창」 공시. 복구 정책은 **Manager 결정 대기(M-2)** |
| 2 | P1 | D7 표시의 거래일 = 취소 시각의 거래일(`gateway.go:1020`) → 「취소한 거래일당 한 번」, 날짜 경계를 넘는 주문은 다시 대상, 알림 key 에 거래일 없음 | 코드 일치 | **사실 정정 반영**: design D7 「거래일의 귀결」. 범위(생애당/거래일당) · 알림 key 는 **Manager 결정 대기(M-1)** |
| 3 | P1 | 주기 끝 실행 시 대기열 항목의 근거 포지션(존재 · 세대 · 보호 필요성)을 다시 확인하지 않음 | 일치(설계에 재검증 없음, `flushStrategyProtectionAlerts` defer 가 먼저 — `defer@485`) | **사실 반영**: design D5 에 순서 · 낡음 공시. 재검증 가드는 새 가드 → **Manager 결정 대기(M-3)** |
| 4 | P2 | 5xx 뒤 「다음 손절 접수」 단정은 park 가 보장할 수 없음(매수가 남으면 다시 거절) | 일치 | **사실 정정 반영**: spec 시나리오 「취소가 5xx 로 끝난다」 → 「park 때문에 막히지 않는다」, design §4-12 를 두 갈래로 |
| 5 | P2 | 운영자 절차 「살아 있다 → FAILED_CONFIRMED」가 넓음 — 처리 중 · REPLACED · 미지 · 부분 체결 뒤 취소 구분 필요 | `resolveCancel` 갈래와 일치(B7 · B9 · B11 · B13 · B15 · B19 · default) | **사실 정정 반영**: design §5 에 상태별 표(해소기 판정 표와 같게). 운영 문서 번역이라 새 정책 아님 |
| 6 | P2 | critical 기록 실패 시 계정 `ReasonAlertUndelivered` 래치 · 모드 승격(`obs/record_only.go:137-157`) — 「다른 종목 무관」의 예외, 해동으로 안 풀림 | 코드 일치(기존 안전 장치) | **사실 정정 반영**: design D6 · §3 공시. 래치 제거 제안 아님 |
| 7 | P2 | `EngineBrokerOrderIDs` 가 계좌 전체 · 기간 무제한 → 번호 재사용 시 정상 외부 매수를 영구 제외 | 일치(`daily_buy.go:350-357`) | **사실 정정 반영**: design §3 거부 입력에 추가(보수 방향 유지). 귀속 범위 축소는 새 정책 → 제안으로만 기록(결정 요청 아님) |
| 8 | P2 | D4 의 21.05 s · 51.05 s 는 근거 부족 — 토큰 교환 HTTP · 예산 대기 · 단계 ctx 미포함, 350 ms 는 최소 간격, 7% 는 GET 몫만 | 일치(`client.go:336`, `token.go:132`, 401 루프 최대 2 회 — 리뷰가 코드 확인) | **사실 정정 반영**: design D4(라) · D5 를 조건부 추정으로, 「단계 시간에 코드상 상한 없음」 명시, 7% 를 「조회 몫만」으로 |

- O-1(토글 audit 실패 처리)은 리뷰도 미결로 지적(새 발견 수에는 넣지 않음) — Manager 결정 대기 그대로.
- 3회차 없음(r2 가 마지막). 개정 반영 뒤 `openspec validate --strict` valid.

### Manager 결정이 필요한 항목만

1. **M-1** 「한 번만」 범위 — 주문 생애당(거래일 제외, 번호 재사용 새 매수는 영구 제외) vs 취소한 거래일당(지금 글, 날짜 경계를 넘는 주문은 다음 날 다시 1 회). 알림 key 범위 포함.
2. **M-2** 중단 창 복구 — ① IN_DOUBT → park 전 중단(기동 해소기가 계정 래치를 걸 수 있음) ② park → 결과 알림 전 중단(알림 누락). 공시로 둘지, 기동 처리 · 누락 알림 복구 정책을 둘지.
3. **M-3** 대기열 항목의 실행 직전 재검증 가드(포지션 존재 · 세대 · 보호 필요성) — 넣을지.
4. **O-1** 토글 audit 실패 처리 — (가) 그 기동에서 토글 꺼짐 + critical / (나) `NewContext` + census 갱신 승인.

## Manager 결정 (2026-11-25) — r2 처분

> 원문(Manager 메시지, 2026-11-25) 그대로 옮김.

Manager 결정(2026-11-25, r2 처분) — r2 원문(P1 3건·P2)과 design O-1 을 직접 읽고 정한다. review.md 「Manager 결정 (2026-11-25) — r2 처분」 절에 원문 그대로 옮기고 문서에 반영해라. 리뷰 3회차는 없다(결정 반영은 리뷰 면제). Go·시험 커밋 0, push 금지 그대로.

1. **M-1 = 주문 생애당 한 번.** 「이미 냈다」 표시의 키에서 거래일을 뺀다: (계좌 · 시장 · 종목 · 대상 번호 바이트, kind CANCEL, source engine, 상태 무관). 알림 key (포지션, 번호)와 범위가 맞는다.
   - 근거: 사람 주문에 대한 엔진의 취소는 적을수록 보수 방향이다. 거래일 키는 자정을 넘긴 같은 주문을 다시 취소하게 하고 그때 알림은 생략된다(r2 #2).
   - 거부하게 되는 정상 입력: 다른 날 같은 번호로 다시 나온 새 외부 매수(브로커 번호의 전역 유일성 미측정) — 그 경우는 오늘 동작(D9 알림 → 사람이 치움)으로 떨어진다. §3 에 적는다. spec `:67` 의 일자 제한과 `:73-75` 재시작 보장을 이 범위로 맞춘다. 시험: 양 시장 자정 넘김 · 자정 넘긴 재시작 · 번호 재사용.
2. **M-2 = 공시만, 복구 정책 신설 없음.**
   - ① IN_DOUBT 기록 뒤 park 전 중단 → 기동 해소기가 일반 경로로 처리(단건 판독으로 닫히거나, 못 읽으면 `Resolver.park` = 계정 진입 래치). D11 의 「종목 범위」는 **세션 안에서 park 까지 마친 경우**의 보장이라고 한정해 적고, 이 창의 귀결(계정 진입 래치, 손절·청산은 막지 않음, 해제 = 운영자 해동 + 프로세스 재시작)을 §5 운영 절차에 적는다.
   - ② park 뒤 결과 알림 전 중단 → 결과 알림 누락을 잔여로 공시. 대신 **취소 전에 내는 첫 critical 본문**에 「결과가 불명으로 남으면 이 종목 진입이 막히고 운영자 해동이 필요하다 — 절차는 운영 문서」 한 문장을 넣어 첫 알림만으로도 인계가 되게 한다. park 된 attempt 를 운영자가 찾는 기존 표면(콘솔·CLI 목록 등)이 코드에 있으면 사실로 인용하고, 없으면 「없음」이라고 적는다(새로 만들지 않는다).
   - 근거: 두 창 모두 원장 기록 두 건 사이의 프로세스 사망이고, 귀결은 진입 쪽 보수 방향이며 손절·청산을 막지 않는다. 기동 처리 신설은 YAGNI.
3. **M-3 = 최소 재검증 가드 수용.** 주기 끝 실행 직전에 「그 포지션이 여전히 열려 있고 대기열에 넣을 때와 **같은 세대**인가」만 다시 확인하고, 아니면 그 항목을 버린다(취소 0, 알림 0 — 다음 주기의 거절이 다시 넣는다). **보호 필요성 재판정은 넣지 않는다**(판정을 둘로 두지 않는다). 판독은 관측기가 이미 쓰는 기존 함수를 재사용한다 — 새 원장 질의를 만들지 않는다. 기존 함수가 없으면 그 사실과 가장 작은 대안을 「열린 항목」으로 올려라.
   - 근거: 사람 주문을 취소할 권한의 근거는 「이 포지션의 손절이 지금 거절되고 있다」뿐이다. 주기 끝까지의 창(다른 포지션 제출이 느리면 수십 초)에 포지션이 닫히면 근거 없이 사람 매수를 취소하게 된다(r2 #3).
   - 거부하게 되는 정상 입력: 그 창에서 세대가 바뀐 포지션 — 다음 주기가 다시 판정하므로 한 주기 지연. §3 에 적는다. 시험: 적재 뒤 청산 · 세대 변경 · ctx 취소, 변이: 가드 삭제.
4. **O-1 = (나).** 토글 audit 은 a163 선례와 같은 자리(`NewContext` 의 `recordGateSettings`)에 넣고, a166 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`) 갱신을 **Manager 가 승인한다**(a199 선례).
   - 근거: 「runtime config 변경은 audit 로 추적 가능해야 한다」(Non-negotiable)를 기존 장치·기존 실패 의미(audit 실패 = 기동 거절) 그대로 충족한다. (가)는 「audit 실패 → 토글 꺼짐 + critical」이라는 새 정책 갈래를 만든다. `recordGateSettings` 는 이미 매 기동에 도므로 실패 양식이 늘지 않는다.
   - 귀결: 편집 대상에 `NewContext`·`recordGateSettings` 추가(pre-edit FLM/BTM 번들), `Context.ExitObserver` 의 audit 편집은 뺀다(토글 값 덮어쓰기 편집만 남는지 확인해 FLM 표를 맞춘다). 구현 단계 선택 시험에 `./internal/accountbudgetruntime/...` 를 반드시 넣는다고 tasks 에 적는다(census 는 선택 시험에 안 잡혀 a189·a192 에서 두 번 늦게 터졌다).
5. r2 P2 4~8 의 네 처분(사실 정정 반영)은 수용한다. r2 #7 의 「귀속 범위 축소」 제안은 기각(넓은 귀속 = 엔진이 사람 주문으로 오인해 취소하지 않는 쪽, 보수 방향) — 기록만.

## r2 처분 결정 반영 (2026-11-25, 팀메이트 Opus) — 리뷰 면제(결정 반영)

| 결정 | 반영 자리 |
|---|---|
| M-1 생애당 한 번 | design D7(키에서 거래일 제거 · 근거), §3 「다른 날 같은 번호의 새 외부 매수」, §4-10(KR · US 자정 · 자정 넘긴 재시작 · 번호 재사용), 변이 「키에 거래일 추가」; spec ADDED 「같은 주문 번호에 취소를 한 번만」 본문 · 「재시작 뒤」 · 새 시나리오 「날짜가 바뀐 뒤」; proposal What Changes |
| M-2 공시만 | design D11(종목 범위 = 세션 안 park 완료 한정, 중단 창 ①②), D6 첫 critical 문장, §5-5(① 의 귀결 · 해제 = 해동 + 재시작), §5-6(park 된 attempt 를 나열하는 표면 **없음** — `UnresolvedAttempts` 소비자는 `checkSymbolFree` 하나, `alertops.go:186-199` 는 사유 코드만), §6; spec 새 ADDED 「세션 안에서 park하지 못한 … 재시작 복구가 해소한다」, 「기록과 알림」 의 첫 알림 문장, §4-19 |
| M-3 최소 재검증 | design D5(기존 `Journal.Positions` 재사용 — `workingSet` `exitloop.go:577`, 판정식은 `:592` 와 같게: ID · `InstanceSeq` · 닫힘 아님 · 수량 0 아님; 보호 필요성 재판정 없음), §3, §4-15a, 변이 2; spec 「주기 끝」 요구 본문 · 시나리오 「판정 뒤 주기 끝 전에 포지션이 닫힌다」. 기존 함수가 있어 열린 항목 없음 |
| O-1 (나) | design D8 · §7, FLM 번들 `internal-app-engine--newcontext` · `internal-app-engine--recordgatesettings`(새로 2), `Context.ExitObserver` FLM 에서 audit 편집 제거; spec 「토글」 요구에 audit 실패 = 기동 거절; tasks 2.1 · 3.2 · 3.3(`./internal/accountbudgetruntime/...` 필수) |
| 5 r2 #4~#8 수용, #7 축소 기각 | design §3(기각 기록) |

- **O-1 반영 중 확인한 사실(Manager 확인 요청 — 결정 범위 안의 구현 자리)**: 결정 원문은 「`NewContext` 의 `recordGateSettings`」 이고 귀결로 두 함수를 편집 대상에 넣으라고 했다. 그런데 `recordGateSettings` 의 changes 목록 **끝**에 행을 덧붙이면 a163 핀 `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten`(`a163_source_age_audit_test.go:16` — 마지막 행이 `engine.exit_source_age_report`, 그 앞이 `per_trade_risk_budget`)이 깨지고, 중간에 넣어도 같은 핀이 깨진다. 인자를 늘리면 시험 함수 5 개(a074 4 · a163 1)의 호출을 고쳐야 한다(gate ⑤ 번들 5). 그래서 design D8 은 **`NewContext` 안, `recordGateSettings` 호출 바로 뒤에 같은 실패 의미(기동 거절)의 `RecordChange` 한 줄**로 적었다 — 자리(`NewContext` 의 기동 audit 단계) · 실패 의미 · census 갱신은 결정 그대로이고 `recordGateSettings` · a163 핀 · 시험 함수는 무편집이다. `recordGateSettings` 번들은 지시대로 만들었고 편집 대상 목록에서만 뺐다. `recordGateSettings` 안에 꼭 넣어야 한다면 a163 핀 편집과 시험 함수 5 개 번들이 함께 필요하다.
- 편집 대상 기존 함수(최종) **6**: `ExitObserver.submit` · `ExitObserver.alertProposalRefused` · `ExitObserver.endExitOrderCycle` · `Context.ExitObserver` · `mergeEngine` · `NewContext`. pre-edit 번들 7(+ `recordGateSettings` 무편집).
- 열린 항목: **0**(design §7). 위 O-1 자리 사실은 결정 범위 안의 구현 자리 선택이라 열린 항목으로 세지 않았다 — Manager 가 다르게 보면 그때 바꾼다.
- spec delta: exit-policy ADDED 9 · MODIFIED 4, order-execution MODIFIED 1.

## Manager 1단계 검증 · freeze (2026-11-25)

> 원문(Manager 메시지, 2026-11-25) 그대로 옮김.

- Manager 가 직접 읽은 것: 개정 design §0~§5 전문(커밋 `486d11965`), r2 codex 원문의 P1 3건·P2 4·5, probe 원문 `resume-probe-output.txt`, `internal/execgw/gateway.go` `checkSymbolFree`·`attemptTargets` 원문, 정본 order-execution `:44 :46 :125 :131`. spec delta 본문 전문과 AST 61개는 읽지 않았다(검증 명령 결과와 작성자 대조표에 기댐).
- 판정: r2 의 P1 3건은 Manager 결정 M-1(생애당 한 번)·M-2(공시)·M-3(최소 재검증)으로 처분, P2 는 사실 정정 반영. 열린 항목 0. 리뷰는 r1·r2 로 끝(결정 반영은 리뷰 면제).
- **O-1 자리 승인**: `NewContext` 안, `recordGateSettings` 호출 바로 뒤에 같은 실패 의미(기동 거절)의 `RecordChange` 한 줄. `recordGateSettings`·a163 핀·a074/a163 시험 함수는 무편집. 결정의 취지(기동 audit 단계 · 기존 실패 의미 · census 갱신 승인)를 지키면서 핀을 안 건드리는 자리라 받는다.
- **freeze 한다.**

## Pre-Edit 선언 (tasks 2.1, 2026-11-25 — High-risk, 면제 없음)

```text
Pre-Edit Gate:
- change id / task id: a197-an-external-opposite-buy-that-blocks-a-stop-loss-is-cancelled / 2.1 → 3.1~3.9
- 대상 심볼(패키지.함수) — 기존 함수 본문 편집 6(승인된 목록, 그 밖은 멈춤):
  engine.ExitObserver.submit(B15 열 조건 대기열 넣기 + alertProposalRefused 인자, B3 인자)
  engine.ExitObserver.alertProposalRefused(결과 상태·사유로 문장 선택 — D9)
  engine.ExitObserver.endExitOrderCycle(첫 문장에 단계 호출 — D5, B1 앞)
  engine.Context.ExitObserver(판독기 · 토글 덮어쓰기 두 줄 — D8)
  config.mergeEngine(B4 뒤 · B5 앞 토글 복사 — D8)
  engine.NewContext(recordGateSettings 호출 바로 뒤 RecordChange 한 줄 — O-1 승인 자리, a166 census 갱신 Manager 승인)
  타입 선언(함수 아님): ExitObserverOptions(ExternalBuyCancel · OpenOrders), ExitObserver(대기열 · 처리 시각), config.Engine/rawEngine(새 칸),
  audit 동작 상수 1. 새 obs 종류는 만들지 않음 — a091 수 핀(`a091_stop_sold_nothing_test.go:54` critical 20 종)이 새 critical 종류를 막으므로
  기존 critical 종류 `exit.liquidation_delayed` 에 key 접두로 가름(design D6 의 「다른 타입」은 사실 정정 — 같은 종류 · 다른 key).
- 새 파일: internal/app/engine/a197_external_opposite_buy.go(단계 · 행 해석 · 알림),
  internal/journal/a197_external_cancel_issued.go(「이미 냈다」 판독 — 거래일 무관),
  시험: internal/app/engine/a197_external_opposite_buy_test.go, internal/app/engine/a197_startup_audit_test.go,
  internal/config/a197_external_buy_toggle_test.go, internal/journal/a197_external_cancel_issued_test.go.
- CodeGraph definition/callers/callees/impact: 재개 준비 · 개정에서 AST 로 대체(analysis/cited-ast 61, function-logic 7). CodeGraph 재질의는 하지 않음(미실행 —
  sdd-sync 미실행으로 fingerprint 낡음).
- CodeGraphContext 후보와 evidence reconciliation: 미실행(이번 단계 not-applicable — AST · probe 로 대체, 1단계 code-context 산출물 유지).
- 기존 동작 파악 근거: 현재 HEAD AST(cited-ast · function-logic), probe resume*/resume2*, 기존 시험 a094 · a087 · a111 · a129 · a163 · a198 · a074 · a166 census.
- Function Logic Map / Branch Test Map: analysis/function-logic/ 의 7 번들(편집 6 + recordGateSettings 무편집).
- upstream 상속 테스트 영향: yes — 구조 핀(a087 submit · a163 렌더 · a094 한 판정 · a111 · a198 flush)과 a166 census(NewContext). 회귀 방지 = 무편집 통과 확인,
  census 는 a199 방식으로 shapes.json 갱신(Manager 승인).
- 실패 테스트 선행 작성: yes(3.1 RED, analysis/tdd/red-*.txt).
- 설정·DB·journal 변경과 rollback: 설정 칸 1(기본 false) · audit 행 1(값 바뀐 기동만). 원장 스키마 · 새 전이 0(park 는 기존 ResolveUnresolved).
  rollback = 토글 끄고 재시작, 또는 이전 이미지.
- 안전 불변식 §0 위반 여부 검토: 통과 — 토글 꺼짐 = 조회 0 · 취소 0 · park 0(LIVE 주문 0), mutating 명령 실행 0, 손절 · 익절 판정식 · unsettledFor ·
  정본 :46 · 손절 제출 순서 무변. 손절 즉시성: 단계는 주기 끝(같은 주기 판정 뒤) — 다음 주기 시작 지연은 D4 · D5 공시.
- 선택 시험: ./internal/app/engine -run 'A197|A094|A087|A111|A129|A198|A074|A163|ExitClear|ProposalRefused', ./internal/execgw -run 'Refusal|Cancel',
  ./internal/config -run 'A163|A197', ./internal/journal -run 'A197|A198', ./internal/accountbudgetruntime/...(필수), ./internal/obs -run A091,
  -race 는 ./internal/app/engine -run A197.
- 미실행 범위: 전체 시험 · make test · make gate(종결 단계 몫), sdd-sync/sdd-check, CodeGraph 재질의.
```

## 구현 적대 리뷰 r1 (2026-11-25, 1회차) — codex, 다른 계열 · 다른 컨텍스트

- 원문 `analysis/review/impl-r1-codex.md`(기준 HEAD `fb5a4a0bc`, diff 는 `git diff origin/main...HEAD -- '*.go'` 파일). 판정 원문: 「수정 필요 — P1 2건·P2 2건. P0 및 확정적인 오취소 경로는 이번 정적 리뷰에서 발견하지 못했습니다.」
- 처분 규칙: 사실 결함은 고침, 새 가드 · 새 정책 · 범위 확대는 구현하지 않고 Manager 결정 대기.

| # | 등급 | 발견 요지 | 작성자 확인 | 처분 |
|---|---|---|---|---|
| 1 | P1 | 단계 시한(25 s)이 취소 전송 도중 끝나면 게이트웨이의 결과 기록(`MarkInDoubt`)도 같은 ctx 로 실패 → DISPATCH_STARTED 고착 → 그 종목 손절이 재시작까지 막힘 | 코드 일치(`Submit.Cancel(stageCtx…)`, 게이트웨이는 전송 · 기록에 같은 ctx) — 이 단계가 정상 원장에서 스스로 만드는 고착 | **수리**: 단계 전체 ctx 시한 제거, 결정 발행 · 취소 제출은 주기 ctx(판독은 자기 3 s, 사전 조회 · 취소는 어댑터 2 s · 15 s). design D4 정정. 시한 만료 결합 시험은 없음(잔여) |
| 2 | P1 | 첫 critical 기록 실패에도 취소가 진행되고, 실패한 알림을 다시 기록하는 길이 없음(D7 표시 때문에 다음 주기 재시도 없음) | 일치(`recordEpisodeCritical` 반환값 무시). 생산 알림기는 실패 시 계정 래치를 거는 기존 장치 있음 | **Manager 결정 대기(I-1)** — 「첫 critical 성공을 취소의 선행 조건」 · 「결과 알림 재기록」은 새 가드 · 정책 |
| 3 | P2 | attempt 가 없는 실패(결정 발행 실패 · 게이트웨이 기록 전 거절)에도 「다시 취소하지 않는다」고 알리고 같은 key 로 래치 | 일치(D7 · M-1 의 「기록 전 실패는 낸 것이 아님」과 반대 문장) | **수리**: attempt 없음이면 문장 「다음 거절 주기에 다시 시도할 수 있다」 + key 접미 `|not-sent`. 이 갈래 시험 없음(잔여 — 기록 전 거절을 결합 시험에서 만들 하네스 없음) |
| 4 | P2 | 변이 1판 M22 가 실패 시험을 선택하지 않음, M05 는 컴파일 실패, M20 은 1판 생존 | 일치 | **수리 완료(리뷰 전 진행분 포함)**: M05 변이를 컴파일되게 고침(2판 CAUGHT), M20 은 주기 끝 전 청산 시험 추가로 2판 CAUGHT, M22 는 시험 선택 · 거절 지점 단언을 고쳐 3판 CAUGHT — `analysis/mutation/` |

## gstack review r2 (2026-11-25, 2회차 · 마지막) — Claude 전문가 5 + 적대 서브에이전트, 다른 컨텍스트

- 기준 HEAD `8b6ef490a`(merge-base `b93ed133a`), 대상 diff 는 `*.go` · `schemas/*.json`. 전문가: Testing · Maintainability · Security · Performance · State Machines. **Simplification · Red Team 전문가는 보내지 않음**(적대 서브에이전트가 Red Team 역할을 대신 — 적용 범위 축소로 기록). 적대 서브에이전트는 시험 파일을 요약 모드로만 봤다고 스스로 밝힘(677줄 엔진 시험 미열람).
- Gate edits: listed 6 · eligible 1 · inspected 1 · unread 0 — `RH-3?` 1건 = `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 의 `NewContext` 키 갱신. Manager 승인(2026-11-25, O-1 배치 승인 · census 갱신 승인)대로이며 a199 방식 수리 영수증은 `analysis/census-repair/`. 처분: **승인된 gate 편집, 추가 조치 없음**.
- 적대 서브에이전트 권고 원문: 「Recommendation: fix finding 1 before merge because when the ledger fails after Prepare, the critical tells the operator "cancel the buy in the app and the stop goes" while the engine's own pending CANCEL attempt keeps that symbol's stop blocked until a restart.」
- Security 전문가: `NO FINDINGS`.

| # | 출처 | 발견 요지 | 작성자 확인 | 처분 |
|---|---|---|---|---|
| 1 | State Machines · 적대 #1(수렴) | 원장 쓰기가 Prepare 뒤 실패해 취소 attempt 가 RECORDED · DISPATCH_STARTED · ACKED 로 남으면 게이트웨이 종목 검사가 그 종목 손절까지 막는데, default 갈래 알림은 「앱에서 그 매수를 취소해야 손절이 나간다」고 말함. park 실패 갈래도 같은 `next` 문장을 붙여 앞뒤가 모순 | 코드 일치(`gateway.go` submit 의 `out.AttemptID` 는 Prepare 직후 채워지고 `IsTerminal()` 은 CONFIRMED · NOT_DISPATCHED · FAILED_CONFIRMED · UNRESOLVED 만) | **수리 `e8b09fe3f`**: 순수 판정 `externalBuyCancelHeld`(attempt 있음 ∧ 비종결) 갈래 추가, `recordExternalBuyUncleared` 에 `held` 인자 — 참이면 「재시작 복구가 그 시도를 해소할 때까지 막힌다 · 앱에서 취소해도 손절은 나가지 않는다」. park 실패 갈래도 `held=true`. 시험: `TestA197OnlyAnUnfinishedRecordedCancelHoldsTheSymbol` · `TestA197TheUnclearedAlertSaysWhatReleasesTheStop`. **잔여**: handleExternalBuy 의 그 갈래 배선을 결합으로 재는 시험 없음(원장 쓰기 실패 주입 하네스 없음 — 변이 R2a 로 공시) |
| 2 | 적대 #2 | 단계의 `recover()` 가 취소 attempt 생성 뒤 panic 을 warn 한 줄로 삼켜, 종결 전 attempt 가 그 종목 손절을 재시작까지 막는데 critical 이 없음(변경 전엔 프로세스가 죽고 재시작 복구가 해소) | 일치 | **Manager 결정 대기(I-2)** — recover 에서 critical 을 쓰거나 recover 를 없애는 것은 새 정책 |
| 3 | 적대 #3 | `endExitOrderCycle` 은 `ObserveOnce` 의 defer 라 judge panic 의 되감기 중에도 단계가 돌아 실제 취소를 낼 수 있음 | 일치(`ObserveOnce` 는 승인된 편집 함수 6 밖) | **Manager 결정 대기(I-3)** — 정상 반환에서만 돌리는 표식은 승인 밖 함수 편집 · 새 가드 |
| 4 | 적대 #4 | 「외부」= 이 원장에 없는 번호 전부 — 같은 계좌를 쓰는 다른 자동매매(StockOS 등)나 원장 초기화 뒤 옛 엔진 주문도 취소 대상 | 일치(design D2 의 귀속 정의) | **Manager · 사람 결정 대기(I-4)** — 토글 flip 전 운영 전제(계좌 단독 사용) 확인 |
| 5 | 적대 #5 | OPEN 목록 행에 `execution` · `filledQuantity` 가 없으면 행 하나가 판독 전체를 무효로 만들어 기능이 조용히 꺼짐(warn 만, critical 없음). 기존 파서(`brokerstate.ParseOfficialOrder` · manualbuy)는 없음을 0 으로 읽음 | 일치 — OPEN 목록 행 모양은 미측정(M25 는 단건 GET) | **Manager 결정 대기(I-5)** — design D3(보수 해석) 유지 여부 · OPEN 행 실측 |
| 6 | 적대 #6 | 토글 꺼짐에서도 D9 본문 · 시동 audit 행(쓰기 실패 시 시동 거절)이 upstream 과 다름 | 일치 | 설계 수용(D9 · O-1, Manager 승인) — 기록만 |
| 7 | 적대 #7 · Performance ×2 | 막힘이 지속되면 5 s 주기마다 OPEN GET + 계정 전 이력 `EngineBrokerOrderIDs` 적재 + 색인 없는 `mutation_attempts(kind, target_order_id)` 스캔. `externalBuyTried` 는 줄지 않는 map | 일치(색인은 state · intent_id · broker_order_id 뿐). 기존 질의도 같은 무색인 모양(`daily_buy.go` · `exit_proposal_release.go`) | **Manager 결정 대기(I-6)** — 색인 마이그레이션 · 점 조회 전환은 schema 변경 · 범위 확대. D4 에 공시된 비용 |
| 8 | 적대 #8 | 토글 켜짐에서도 D9 본문이 「앱에서 취소해야」라고 말하고, 사람이 먼저 취소하면 엔진 취소가 확정 거절 → 두 번째 「치워지지 않았다」 critical | 일치 | **Manager 결정 대기(I-7)** — 문장 정책 |
| 9 | Maintainability | 거절 알림 제목 「청산 주문이 제출되지 않았다」가 D9 로 고친 본문(「브로커가 거절했다」)과 모순 | 일치 — 제목은 a197 이전부터 있던 문장, D9 범위는 본문 | **Manager 결정 대기(I-8)** — 제목 변경은 D9 범위 확대 |
| 10 | Testing | `nilOpenOrder` 옵션이 쓰이지 않음(판독기 미조립 갈래 시험 0) | 일치 | **수리**: `TestA197NoOpenOrderReaderCancelsNothing` |
| 11 | Testing | `|not-sent` key · 문장 시험 0 | 일치 | **부분 수리**: 문장 · key 는 `TestA197TheUnclearedAlertSaysWhatReleasesTheStop` 가 직접 잼. 결합 갈래(결정 발행 실패 · 기록 전 거절)는 하네스 없어 잔여 |
| 12 | Testing | D9 의 그 밖의 브로커 거절 본문(B3) 시험 0 | 일치 | **수리**: `TestA197TheRefusalBodyNamesWhoRefused`(BTM B3 갱신) |
| 13 | Testing | `TestA197OnePositionPerCycle` 의 「취소 ≤ 1」이 공허 — 가짜 판독기가 종목 필터를 무시해 다른 종목 행이 섞이고 판독이 늘 무효라 취소가 늘 0 | 일치 | **수리**: 가짜 판독기에 `bySymbol`(옵트인 — §4-5 의 「다른 종목 행」 시험은 그대로), 첫 주기 정확히 1 · 둘째 주기 다른 포지션 1 단언 |
| 14 | Testing | 계측 시험에 단언 없음 | 일치 | **수리**: reads == 1 ∧ cancels == 1 단언 |
| 15 | Testing | §4-12 주석의 「같은 종목 진입만 막힘」을 단언하지 않음 | 일치 | 잔여 — 기존 게이트웨이 동작(UNRESOLVED 종목 차단)이라 a197 시험으로 다시 세우지 않음 |
| 16 | Testing | 반복 커서 · 전량 체결 행 · 통화 없음 · park 쓰기 실패 · panic 회복 갈래 시험 0 | 일치 | **부분 수리**: 반복 커서(§4-6 표) · 전량 체결 · 통화 없음(§4-5 표) 추가. park 쓰기 실패 · panic 회복은 주입 하네스 없어 잔여 |

- 2회차 상한에 닿았으므로 위 수리 뒤 추가 리뷰 회차 없음. 수리 커밋 `e8b09fe3f` 는 변이 최종판(run5)과 선택 시험으로만 확인.

## 4장 검증 마감 (2026-11-25, HEAD `e8b09fe3f`)

- 변이 원장 최종판 run5(`analysis/mutation/ledger.md`): design §4 의 22 중 **21 CAUGHT · 생존 0 · M12 NOT_RUN**(만들 수 없음 — 관측기에 진입 게이트 핸들 없음). 추가 R2a SURVIVED(종결 전 attempt 갈래 배선 — 원장 쓰기 실패 주입 하네스 없음) · R2b · R2c CAUGHT.
- 선택 시험 `analysis/verify/selected-tests-r2.txt`: engine(`A094|A087|A111|A129|A198|A074|A163|A197|ExitClear|ProposalRefused`) · execgw(`Refusal|Cancel`) · config(`A163|A197`) · journal(`A197|A198`) · obs(`A091`) · `accountbudgetruntime/...` 전부 ok · rc 0. `-race -run A197` ok(136.8 s). `make lint` rc 0. `openspec validate --strict` valid, `check_analysis --change` required 6 → evidence complete(`validate-check-r2.txt`).
- **미실행 범위**: 전체 시험 · `make test` · `make gate` · `make sdd-check`(지시상 금지), 위 패턴 밖 엔진 시험, `-race` 의 A197 밖, `-trimpath` 엔진 실행(구조 핀 거짓 실패 — 영수증 `selected-tests.txt`), 실브로커 OPEN 행 모양(미측정 — r2 #5).

## Manager 검증 · 종결 (2026-11-25)

### Manager 5.1 검증 — **구현 수용**
- Manager 실측(HEAD `908b1641c`, worktree `~/tossos-work/a197-impl`, 작업 트리 깨끗): 새 파일 `a197_external_opposite_buy.go`(449줄) · `journal/a197_external_cancel_issued.go` 전문과 기존 7파일 편집 diff 전문을 읽음. 편집된 기존 함수는 승인된 6개뿐임을 diff 로 확인. 선택 시험 독립 재실행(`-trimpath` 없이, `-count=1`): `./internal/app/engine -run 'A094|A087|A111|A129|A198|A074|A163|A197|ExitClear|ProposalRefused'` ok 122.791s · `./internal/execgw -run 'Refusal|Cancel'` ok 40.634s · `./internal/config ./internal/journal -run 'A163|A197|A198'` ok · `./internal/accountbudgetruntime/... ./internal/obs` ok(24.650s · 25.461s) · A197 PASS 줄 41 · `go vet`(engine · journal · config) 출력 없음. 읽지 않은 것: 시험 파일 5개 본문, 변이 원장 원문, 번들 AST.

### Manager 결정 2026-11-25 (gstack r2 · r1 대기 항목 처분)

| 항목 | 결정 | 반영 |
|---|---|---|
| D6 이탈(새 critical 종류 대신 `EventExitLiquidationDelayed` + key 접두) | **승인** — 같은 종류를 key 접두 · 접미로 가르는 것이 이 코드의 관례(`thaw-release:` · `nointent:` · `cancel:` · `streak:` · `catchup-list:`, Manager rg 확인) | design D6 문장을 구현에 맞춤 |
| I-1 첫 critical 기록 실패 | **선행 조건으로 하지 않음 · 재기록 없음** — 알림 저장 실패로 손절의 길 치우기를 멈추면 불변식 4 에 반함. 기록 실패는 기존 장치(계정 진입 래치 + 모드 승격)가 받고 취소 사실은 원장의 결정 · intent · attempt 에 남음 | 잔여 |
| I-2 `recover()` 뒤 종결 전 attempt | **새 critical 없음** — 기존 지연 경보(30초) · 연속 실패 경보가 받음(design D11 끝 문장) | 잔여 |
| I-3 되감기 중 단계 실행 | **잔여 공시, `ObserveOnce` 편집 없음** — 같은 권한(이 주기에 손절이 반대 미체결로 확정 거절 · M-3 재검증 통과 · 토글 켜짐) 위의 같은 행위이고 동기 실행이라 결과가 원장에 기록됨. `endExitOrderCycle` 의 기존 줄 쓰기도 같은 성질 | 잔여 |
| I-4 「외부」의 범위 | **pre-flip 사람 확인** | design §5-1 · issues I2 · Story limits |
| I-5 OPEN 행 모양 | **D3 유지 + 영수증으로 못 박기** | 아래 |
| I-6 무색인 판독 비용 | **색인 · 스키마 변경 없음, 한 번 잼 — 두 판독 합이 100 ms 를 넘으면 멈춤** → 측정 합 ≈ 138.5 ms 로 멈춤 → **Manager 결정 2026-11-25 = 선택지 1(측정 수용, 잔여 기록)** | design §6 잔여 · §5-1 · issues I2 · Story limits(원장 행 수 확인) |
| I-7 · I-8 | **무변, 잔여** — D9 는 토글 무관 · 제목 · key 무변이 결정(알림 수 불변) | 잔여 |

이름 붙인 잔여: R2a(종결 전 attempt 갈래 배선의 결합 시험 없음), gstack r2 INCOMPLETE(Simplification · Red Team 전문가 미실행), 시험 공백 넷(not-sent 결합 경로 · park 쓰기 실패 · panic 회복 · §4-12 같은 종목 진입 차단 단언). 리뷰는 2회차로 끝.

### I-5 영수증과 시험
- 공식 스펙 토스증권 Open API 1.2.24 `docs/migration/openapi.latest.json`(a143 `926d67a2f` 로 들어옴): `GET /api/v1/orders` 의 `Order` 필수 = `orderId · symbol · side · orderType · timeInForce · status · quantity · currency · orderedAt · execution`, `price` 는 `string | null`, `OrderExecution` 필수 `filledQuantity`(string; 「체결 내역이 없으면 filledQuantity=0」), 나머지 칸 null 가능. 예시 `pendingMixed` 첫 행이 PENDING BUY 모양.
- 운영 판독기도 같은 칸을 문자열로 읽음: `internal/official/orders_reads.go:18,52,54,58`, `internal/filldetect/payload.go:37-52`.
- **실측은 없음**: `measurements.md` M31 — CLOSED+OPEN 197 건에서 관측 status 는 CANCELED · FILLED · REJECTED 뿐(PENDING 미관측, 필드 이름 목록은 영수증에 없음). M25 · M32 는 단건 GET. → **pre-flip 사람 실측(읽기 전용 OPEN 1회)** 으로 design §5-1 · issues I2 · Story limits 에 올림.
- `parseExternalBuyRow` 와 스펙은 어긋나지 않음 — 시험 `TestA197TheSpecShapedOpenRowIsRead`(스펙 모양 · 합성 값, 지정가 · price null 두 경우) PASS. production 무변.

### I-6 측정 — **멈춤(합 > 100 ms)**
- 하네스 `analysis/perf/a197-perf-harness.go.txt`(일회성, `internal/journal` 에 임시 복사해 실행 뒤 삭제), 원문 `analysis/perf/a197-perf-result.txt`.
- 잰 시각 2026-11-25T04:15:58+09:00, HEAD `908b1641c`, x86_64 20코어, 원장 파일은 /tmp 디스크, 측정 직후 load average 8.08(다른 세션 부하 있음).
- 모집단: 합성 원장 intents 100,000 · mutation_attempts 100,000(PLACE 60 % · CANCEL 30 % · AMEND 10 %, 한 계좌), `EngineBrokerOrderIDs` 가 싣는 번호 70,000.
- 값(5회): `EngineBrokerOrderIDs` first 131.6 ms · median 128.1 ms · max 132.0 ms. `ExternalCancelIssued` 미적중 median 10.4 ms · 적중 median 10.7 ms.
- **합 ≈ 138.5 ms(median) > 100 ms** → Manager 지시대로 여기서 멈춤. 종결(5.2) 미착수.
- 해석(판정 아님): 지배 항은 `EngineBrokerOrderIDs`(계좌 전 이력 적재, 이 단계가 만든 판독 아님 — a136 의 기존 함수)이고, `ExternalCancelIssued` 는 매수 행마다 1 회. 실제 원장 크기는 재지 않았음(운영 원장 열람 금지).
- **Manager 결정 2026-11-25 — I-6 = 선택지 1(측정 수용, 잔여 기록)**. 판정 근거: (가) 시간의 대부분(median 128.1 ms / 138.5 ms)은 a197 이 만든 판독이 아니라 기존 귀속 판독 `EngineBrokerOrderIDs`(a136)다. (나) 이 비용은 토글이 켜져 있고 보호 손절이 반대 미체결로 확정 거절된 주기에만, 주기 끝에서 한 번 든다 — 다음 주기 시작을 약 0.14 s(5 s 간격의 약 2.8%) 늦추며, 같은 단계의 브로커 왕복(조회 · 사전 조회 · 취소)과 같은 자릿수이고 이미 공시한 최악 지연(요청 시한 합 약 21 s)보다 두 자릿수 작다. (다) 선택지 2(행마다 점 조회)는 「엔진 귀속」의 정의를 두 곳에 두게 된다 — 두 정의가 갈리면 엔진 주문을 외부로 오인해 취소하는 쪽으로 틀린다. 귀속 정의는 한 곳에 둔다. 선택지 3 은 스키마 변경이라 범위 밖. 100 ms 는 Manager 가 들여다보려고 둔 임계였고 계약 값이 아니다.
- 잔여 · 후속: 비용은 원장 이력에 선형(합성 10만 행 · load 8 에서 1회 측정). design §6 에 측정값 · 조건, flip 전 사람 확인에 「운영 원장의 `mutation_attempts` 행 수 확인 — 10만 행당 거절 주기마다 약 0.13 s」(design §5-1 · issues I2 · Story limits). 귀속 판독의 점 조회 전환은 필요해지면 별도 change(정의 한 곳 유지).
- I-5 처분 확인(Manager): 스펙 · 운영 판독기와 `parseExternalBuyRow` 가 맞고, 실측 OPEN(PENDING) 목록 행이 없다는 사실과 flip 전 사람 실측 항목을 받음.
