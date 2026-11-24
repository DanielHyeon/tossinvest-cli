# a197 Tasks

> **High-risk(주문 취소·손절 경로·운영 토글).** 실계좌 주문·취소를 내는 자동 시험 금지. 픽스처는 합성 값만(실제 종목·수량·가격·주문번호·계좌 금지). 브로커는 가짜 판독기·가짜 게이트웨이·httptest 만.
> 토글 기본 false, 꺼짐 = 오늘 동작. 운영 flip 은 사람이 한다. 전체 시험·전체 게이트는 종결 단계에서만. 원장 스키마 변경 없음.
> 리뷰는 최대 2회차, 리뷰가 낸 새 가드·새 정책은 Manager 결정 전 구현 금지.

## 1. 계약과 증거 (production Go·시험 편집 0)

- [x] 1.1 Story `STORY-TOS-a197`(FEAT-TOS-009) 등록 확인, `base-commit.txt` 고정. — 완료 2026-11-23: `064f59fd9`, PM `--check` rc 0.
- [x] 1.2 CodeGraph hard evidence(버전 명기): `ExitObserver.submit`·`record`·`clearTheSymbol`·`alertProposalRefused`·`release`·`isProtective`, `execgw` 의 거절 분류(`classifyMutation`·`ReasonOppositePendingOrder`)·취소 경로(`Gateway.Cancel`, `Issuer.IssueReduction`), 미체결 목록 판독기(`execgw.OfficialOrders`·flatten 의 `ScanOrders`), 원장의 주문 귀속 판독(`LiveOrdersForSymbol` 등) — `analysis/code-context/`. — 완료: CodeGraph 1.6.0 · CodeGraphContext 0.5.1(색인 0) · GBrain(낡음).
- [x] 1.3 `tools/logic-map` AST 를 `analysis/cited-ast/` 에 현재 HEAD 로 뽑고 proposal 표 여섯 줄과 design D1·D2·D6·D7·§2 의 주장을 대조한다: (a) `submit` default 갈래에서 `out.Reason`·`out.State` 를 읽을 수 있는가, (b) 기존 취소 경로를 외부 주문 번호로 부를 수 있고 원장에 무엇이 남는가, (c) 「엔진 귀속」을 주문 번호로 판정할 원장 판독이 무엇인가(`LiveOrdersForSymbol` 이 종결 뒤 주문을 빼면 귀속 판정에 못 쓴다 — 그러면 대체 판독을 찾아 보고), (d) 미체결 목록 판독의 쪽 넘김·실패 표현. 틀리거나 없는 재료가 있으면 문서를 고치지 말고 **Manager 에게 보고**(설계 결정이 바뀐다). — 완료: `analysis/cited-ast/README.md`(AST 28 · probe 1). **Manager 결정 필요 4건**(D6 취소 결정 발행기 · D2 귀속 판독 · D8 토글 audit 자리 · park 매수).
- [x] 1.4 편집 표면 세기: `submit`·`alertProposalRefused` 를 고정하는 구조 시험(`a087_submit_structure_test`, a163 판정 핀 등)과 census, 진행 중 change a087·a167 의 같은 함수 편집 여부(브랜치·worktree 확인), 판독기 주입에 필요한 배선 함수가 a166 census 에 있는지, 기존 exit 토글의 설정 자리·audit 경로(정본 「체결 나이 보고는 기본 꺼진 운영 토글 뒤에 있다」의 구현을 선례로). census·진행 중 change 와 겹치면 **멈추고 Manager 에게 돌아온다**. — 완료: `analysis/edit-surface.md`. 겹침: a166 census `NewContext`(토글 audit 선례를 따를 때만), a087 `landed-commit.txt` 부재.
- [x] 1.5 rate limit 계상: 이 단계가 더하는 호출(미체결 조회·취소)을 retry matrix 문서에 적을 자리와 그룹 확인. — 완료: `analysis/edit-surface.md` §1.5.
- [x] 1.6 편집 대상 기존 함수의 FLM/BTM 번들(pre-edit 판). — 완료: `ExitObserver.submit` · `ExitObserver.alertProposalRefused` · `Context.ExitObserver`(조건부). 토글 자리가 정해지면 `mergeEngine` · `recordGateSettings` 추가 가능. `check_analysis.py --change` rc 0(required 0).
- [x] 1.7 `openspec validate a197-… --strict` 통과, proposal-freeze 적대 리뷰(다른 계열, 최대 2회, 적대적 Eng 관점 필수) — `review.md`. — 완료 2026-11-23: strict valid rc 0, codex r1(P0 0 · P1 7 · P2 4, 전부 Manager 결정 대기), r2 미실시(개정 없음) — `review.md`, `analysis/review/proposal-freeze-r1-codex.md`. **r2 실시 2026-11-25**(개정본, codex, 마지막 회차): P0 0 · P1 3 · P2 5 — 사실 정정 반영, Manager 결정 대기 M-1 · M-2 · M-3(+ O-1). `analysis/review/proposal-freeze-r2-codex.md`.
- [ ] 1.8 [blocked] Manager 1단계 검증·freeze. — 선행: a198(길 치우기의 취소 결정 발행 수리) 착지 + 아래 재설계 결정(design §7, issues.md, Manager 결정 2026-11-23).
  - 2026-11-25 재개 준비: 재개 조건 1(a198 착지) **충족**(origin/main `b93ed133a` 병합, base 재고정 — Manager 승인 2026-11-25, `review.md` 「재개 준비」). 2차(같은 날): 선택지 (E) · SQL 전수 · 시한 구조 — `analysis/resume/decision-2-addendum.md`.
  - 2026-11-25 재개 개정: 조건 2(선행 2 = (E) 종목 범위 park — Manager 결정 2026-11-25) **충족**, 조건 3(proposal · design · spec delta · tasks 개정, `openspec validate --strict`) **충족**, 조건 4(r2 적대 리뷰) — `review.md` 「proposal-freeze 적대 리뷰 r2」. **해제 조건(남음)**: r2 처분의 Manager 결정 대기 M-1(「한 번만」 범위) · M-2(중단 창 복구) · M-3(대기열 재검증 가드)와 design O-1(토글 audit 실패 처리)의 결정 → 결정 반영(문서) → Manager freeze. r2 가 마지막 회차라 3회차 리뷰는 없다.
- [x] 1.9 재개 개정의 pre-edit FLM/BTM 번들 — 새로 편집 대상이 된 `ExitObserver.endExitOrderCycle` · `mergeEngine` 2 개 추가(기존 3 개 `submit` · `alertProposalRefused` · `Context.ExitObserver` 와 합쳐 5). `check_analysis.py --change` required 0.

## 2. Pre-Edit

- [ ] 2.1 Pre-Edit 선언을 `review.md` 에 확정 — 대상 5 함수(`ExitObserver.submit` · `ExitObserver.alertProposalRefused` · `ExitObserver.endExitOrderCycle` · `Context.ExitObserver` · `mergeEngine`)와 새 파일 목록. O-1 결정이 `NewContext` 쪽이면 `NewContext` · `recordGateSettings` 와 a166 census 갱신이 더해진다(그때 번들 추가).

## 3. RED → GREEN

- [ ] 3.1 RED: `internal/app/engine/a197_external_opposite_buy_test.go`(+ 설정 시험) — design §4 의 1~18.
- [ ] 3.2 GREEN: 새 파일(외부 치우기 단계 · 미체결 판독 해석 · 「이미 냈다」 원장 판독 함수 — 새 파일) + 설정 토글(`config.Engine` 칸 · raw · `mergeEngine`) + `submit` B15 대기열 넣기 · `alertProposalRefused` 문장 + `endExitOrderCycle` 첫 문장 호출 + `Context.ExitObserver` 판독기 · 토글 · audit 배선.
- [ ] 3.3 토글 꺼짐 회귀: 기존 a094 · a087 · a129 · a198 시험 무편집 통과(design §4-17).
- [ ] 3.8 요청 수 · 시한 계측(design §4-16) — 결과를 design D4 에 적는다.
- [ ] 3.9 편집 뒤 5 번들 AST · FLM · BTM 재생성. retry matrix 문서 갱신(1.5 — 주기당 OPEN 1 + 사전 단건 조회 1 + 취소 1).

## 4. 검증

- [ ] 4.1 변이 원장(`analysis/mutation/`) — design §4 의 변이 목록(18), 닿음 확인·무변이 대조군 GREEN 뒤.
- [ ] 4.2 선택 시험 원문(design §4), `make lint`, `openspec validate --strict`, `check_analysis.py --change a197-…`. 미실행 범위 기록.
- [ ] 4.3 구현 적대 리뷰(다른 컨텍스트) → gstack review 1회 — `review.md`.

## 5. 종결

- [ ] 5.1 Manager 검증(diff·시험 원문 재실행).
- [ ] 5.2 완료 게이트(격리 worktree) → archive → Story 경로 갱신 → push.
- [ ] 5.3 (사람 승인) 운영 반영(토글은 꺼진 채), 토글 flip 은 별도 사람 결정 — design §5.
