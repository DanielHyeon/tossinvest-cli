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
