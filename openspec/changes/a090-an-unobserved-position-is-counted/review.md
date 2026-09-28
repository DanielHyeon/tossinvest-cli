# a090 · review

## 0. 초안 (2026-09-29) — 판정 아님

- 배정: Manager(2026-09-28, 사용자 결정 「a089 처분 추천안 수용」 의 둘째 항목). 정본 소스: a089 2차 리뷰 「a090(신설, 선행)」 원문 · C2 ·
  `exitloop.go:453-462` 의 두 무음 `continue`(base `d3bd1843` 에서 재확인) · a092 `ObserveOnce` FLM 「필요한 RED (a090 후보)」 R1~R6.
- 산출물 순서: AST(8 분기) → 진입 실측(`analysis/harness/observeonce_entry.sh`, commit `b0a202b8`) → FLM·BTM → proposal·design·spec·tasks.
- 사용자 결정 대기: Q1(포지션 단위 두절의 ENTRY_BLOCKED) · Q2(정지·0가격 종목 `/prices` 실측, 선택) · Q3(관측점 — 판정 진입 vs 기록 경계).
- 다음: proposal-freeze 적대 보이스 1(task 0.5) → codex(task 0.6, Manager 슬롯).

## Manager 판정 (2026-09-29) — Q1 · Q2 · Q3 (사용자행 아님)

- **Q1 = (a)** — 정본 준수(`openspec/specs/exit-policy/spec.md:62`·`:65`)이지 신규 정책이 아니다. spec delta 무변경.
- **Q2 = 구현 로트로 이연 + 사전 승인** — 읽기 전용 시세 GET 1회(정지 종목 포함), 장중, 결과로 D6 분기 확정. 문서 표기 [미측정 · 사전 승인된 실측 대기].
- **Q3 = 초안 그대로** — 관측점 판정 진입. 하류 무음 5자리는 명명된 잔여(후속 change 후보).
- 반영: design 「Q — 결정 기록」·D5·D6·D1 · proposal 표 · tasks 0.7·0.8·2.3a.

## 1라운드 — 적대 보이스 1 (Claude, 분리된 컨텍스트, task 0.5) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: 2026-09-29, 읽기 전용 서브에이전트(저장소 쓰기 0 — 끝에 `git status --short <change>` 0줄 확인). 기준 101f1d29 → HEAD 7b9df40f 사이
  `internal/app/engine`·`internal/obs`·`operating_mode.go` 바이트 동일, `exitloop.go` sha256 `522d5d81…` 동일을 리뷰어가 확인.
- 확인된 것: 결함 실재(B6·B7 무음, `:447-448` 형제 리셋, `observe` `:779-783` 한 종목이면 게이트 신선도) · AST 8/5 원문 일치 · 진입 블록이
  하네스 산출과 일치 · `EscalateOperatingMode` 멱등(`operating_mode.go:409-421`) · 키 파싱 소비자 없음(`notifier.go:885-896`) · delta 형식·정본과 무충돌.

| id | 보이스 | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| F1 | P1 | **설계 결함 — 확인(논리)** | 포지션 시계가 **첫 미스**부터라 그 앞의 양보(B1)·전 종목 실패(B4) 주기가 안 세어진다 — 무음 창 최대 ≈2×`outageAfter()`. D3 의 "양보 비대칭이 없다" 는 거짓, 정본 문장("60초 이상 실패")과도 어긋남 | 확인 — B1·B4 는 `states` 순회 전에 돌아간다(`:417-424`·`:441-446`). 고침: 포지션별 **마지막 판정 시각**(없으면 처음 보인 시각/`startedAt`)부터 잰다 |
| F2 | P1 | **설계 결함 — 확인** | 경보(+Q1 강화의 `EventOperatingMode`)가 **루프 안에서 동기로** 나가면 경보당 최악 54초(`obs/alert_lease.go:57` "54s", `n.mu`) × N 이 뒤 포지션의 손절 판정 앞에 선다 — 그 지연이 15초 임대를 태워 **바로 B7 을 만든다**. 파일이 이미 같은 규칙을 둔다: 격리 행을 맨 뒤에(`exitloop.go:608-610` "every valid position—including an emergency breach—is recorded/armed/submitted before alert delivery can wait") | 확인. 고침: 루프 안에서는 기록만, 임계 판정·경보·강화는 **순회 뒤 1회**(+ a094 D−4.6 과 같은 enqueue-only 채택 검토). 차단 스파이 알리미 RED |
| F3 | P1 | **설계 결함** | 임계 아래의 미관측은 여전히 어디에도 안 남는다(메모리 맵뿐) — a089 원문 "세고"·a092 R1 "계수된다"·delta "조용히 건너뛰어서는 안 된다" 미충족 | 동의. 고침: `ExitCycle.Unobserved` + 연속 시작·해제 시 normal 구조화 로그 1줄, R1·R2·R7 이 그것을 단언 |
| F4 | P1 | **설계 결함 — 확인** | "B6·B7 이 유일한 자리" 는 거짓 — `workingSet` 이 보유 포지션을 순회 전에 떨군다(`:537` 무음 · `:531` `:549` `:560` `:596` 은 `cycle.Err` 로그뿐 · `openState` 손절 없는 결정 `:667-671` 매 주기 실패). 정리 기준이 `states` 라 떨군 포지션의 연속이 조용히 지워진다 | `:537` 무음 `continue` 확인. 고침: 정리는 **보유 집합**(`Positions`) 기준. D1 에 workingSet 경로를 명명된 범위 밖(후속)으로 열거하거나 미스로 기록 |
| F5 | P2 | 기록 | a096 재알림 창(1h) 안의 둘째 연속·재시작 재알림은 outbox 에서 ClaimSettled 로 **안 나간다** — "다시 알린다" 가 전송이 아니라 기록 | D2·D4·delta 에 명시 + Notifier 수준 시험 |
| F6 | P2 | 기록 → Q1 비용 | 재시작마다 재강화(R11) — 정지 종목 보유 시 재시작마다 진입 재차단. 모드 이력의 원인이 계정 두절과 구별 안 됨 | Q1 비용 절에 추가, 경보 본문에 포지션 명명 |
| F7 | P2 | 수리 | B3(무보유) 조기 반환이 정리를 건너뛴다 — 같은 id 재채움 시 래치 오염 | B3 에서 두 맵 비우기 |
| F8 | P2 | 증거 | D6 의 장 마감 근거(a112 결정 46)는 단일 종목·L1c 엄격 리더 — 배치·`adaptPrices` 경로가 아니다. 더 강한 증거: a096 `proposal.md:173`(토요일 `last_observed_at`) | 인용 교체 + 혼합 시장 배치는 [미측정] |
| F9 | P2 | 계획 | R9 가 기존 시험 둘만 — Q1(a)면 부분 배치 + `Advance` 60초 넘는 다른 엔진 시험도 바뀐다 | RED 전 전수 검색 task |
| F10 | P2 | 기록 | 격리 포지션은 `judge` 에 닿아(`alertRefused`) 관측으로 친다 — 자기 critical 이 있어 수용 가능, D1 에 명시 | D1 에 명시 |
| F11 | P3 | 편집 | "we chose not to look" 인용 줄은 `:40-41` | 정정 |
| F12 | P3 | 증거 | `observeonce.blocks` 머리줄은 손으로 붙였다 · 하네스가 `ast.json` 을 워킹트리에서 읽는다 | 하네스가 머리줄을 쓰고 `git show <commit>:` 로 읽게 |
| F13 | P3 | 증거 | "무음 continue 2" 는 AST 가 아니라 소스 스캔 | 유도 방식 명시 |
| F14 | P3 | 기록 | tracer 는 비시험 생성자 호출 0 · 단일 종목이라 B6·B7 도달 불가 | 명시 |
| F15 | P3 | 근거 | B7 의 현실 원인: 공식 클라이언트 시한 15초(`internal/official/client.go:20`) = 임대 15초 — 앞 포지션의 멈춘 요청 하나가 뒤 전부를 B7 로 | D1·D6 에 인용 |
| F16 | P3 | 범위 | 형제가 찍는 진입 게이트 신선도는 Q1(a) 로만 다뤄진다 | 명시 |

- **판정: REJECT(P1 4 · P2 6 · P3 6).** 반영하지 않았다. Teammate 권고: codex 전에 2판(F1~F4 + P2/P3 편집)을 쓰면 codex 라운드 하나를 아낀다 —
  F1·F2 는 교차 모델이 확실히 다시 문다.

## 2판 (2026-09-29) — 1라운드 반영 · 판정 아님

Manager 판정(2026-09-29): 2판 먼저 — F1~F4 전부, F2 는 enqueue-only(a094 와 같은 근거·형태), F4 는 범위 안/명명 잔여 표. 정본 `design.md`(2판 전면 개정).

| id | 처분 | 반영 자리 |
|---|---|---|
| F1 (P1) | 기점 = **마지막 판정 시각**(없으면 처음 보유 대상으로 본 시각). B1·B2·B4 주기가 시간에 든다 | design D3 · spec 문장 + 시나리오 「양보와 전 종목 실패도 시간에 든다」 · tasks 2.3b |
| F2 (P1) | 루프 안은 기록만, 임계·알림·강화는 **순회 뒤**, 알림은 **enqueue-only**(`Journal.EnqueueAlert`), 적재 실패는 `Retrier.Gate.BlockUnlessClearedSince` — a094 D−4.6·D−5.3 과 같은 형태 | design D4·D5 · spec 문장 + 시나리오 · tasks 2.3c·2.3d |
| F3 (P1) | `ExitCycle.Unobserved` + 연속 시작·해제 normal 로그 | design D7 · spec 문장 · tasks 2.1·2.4 |
| F4 (P1) | 세는 단위를 **보유 대상 집합**으로(`workingSet` B6 통과 뒤 표시 1자리) — 탈락 다섯 자리(B8·B12·B14·B21 범위 안, B10 명명 잔여) + 미관리 B6(명명 잔여) 표. `workingSet` AST·FLM·BTM(분기 22, 탈락 자리 진입 **전부 0**) 편집 전 산출. B3 조기 반환이 보유 중 전부 탈락을 무보유로 읽는 구멍도 같이 | design D1 표·D2·D8 · `workingSet` 번들 · spec 문장 + 시나리오 「판정 목록에서 빠진 보유 포지션」 · tasks 2.12~2.14 |
| F5 (P2) | **에피소드 key**(`…|<연속 기점>`) — 같은 연속은 한 행, 새 연속은 새 행. a094 D−5.2 교차 인용. 재시작은 a090 에서는 새 에피소드(원장에 믿을 판정 시각 없음 — `exit_snapshot_integrity.go:9-12`) | design D4·D2 · tasks 2.3e·2.11 |
| F6 (P2) | Q1 비용 보충: 재시작마다 재강화, 모드 이력으로는 계정 두절과 구별 불가 → 본문에 포지션 명명 | design D5 · Q1 |
| F7 (P2) | B3 조기 반환 앞에서도 순회-뒤 처리(보유 대상이 있으면 미관측, 없으면 전부 정리) | design D2 · FLM · tasks 2.12 |
| F8 (P2) | 장 마감 근거를 생산 증거(a096 `proposal.md:173` 토요일 `last_observed_at`)로, a112 는 보조, 혼합 시장 배치 [미측정] | design D6 |
| F9 (P2) | RED 전 전수 검색 task(1차 표본 2 — 둘 다 60초 미만) | tasks 2.0 |
| F10 (P2) | 격리 포지션은 관측됨 | design D1 표 · tasks 2.14 |
| F11~F16 (P3) | 인용 `:40-41` 정정 · 하네스가 머리줄을 쓰고 커밋의 `ast.json` 을 읽음(eac13df1) · FLM 에 `continue` 유도 방식 · tracer 도달 불가 · B7 의 현실 원인(클라이언트 시한 15초 = 임대 15초) · 게이트 신선도는 범위 밖 명시 | proposal · 하네스 · FLM · design D1·D6 · proposal Non-goals |

### Manager 판정 (2026-09-29) — base 재고정 시점

- `check_analysis.py --change a090-…` rc 1(깨끗한 worktree, 39dd5d38) — **원인 = 이웃 a066 v35 착지(journal 함수 4개가 base `d3bd1843` 뒤 창에 듦) · 자기 Go 0
  · 재고정은 구현 로트의 첫 행위(WORKFLOW 사람 절차·영수증)**. 설계 단계 rc 1 은 게이트 요건이 아니다 — 지금 고정하면 a066 착지로 곧 낡는다(a125 (나) 원리).

