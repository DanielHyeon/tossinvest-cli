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

## codex 1라운드 (교차 모델, task 0.6) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: session **`01a0e940-56d4-7d60-b426-1e43ad4548f4`**, 2026-09-29 03:21:42~03:25:35 KST, rc 0, tokens 144,693, 401 없음. 트리 = `git archive 356309f9`
  (17452 = ls-tree), 인용 Go 드리프트 0. 프롬프트 `analysis/freeze-review/codex-r1-prompt.md`(ecd67a44).
- 보이스 1 발견 판정: RESOLVED 11 · PARTIAL 5(F2 · F3 · F4 · F8 · F14). AST 해시 두 번들 일치, 분기 좌표 8/22 · return 5/3 일치. `Retrier.Gate` 는 생산에서 non-nil
  (`internal/app/engine/gateway.go:249`·`:324` → `exitwiring.go:50`)이나 생성은 nil 을 허용(`execgw/retry.go:309`).

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| N1 | P0 | **설계 결함 — 확인(논리)** | B10(완료 정책)은 B6 통과 **뒤**라 표시되고 판정되지 않아 세어진다 — R14(세지 않음)와 모순. 표시를 B10 뒤로 옮기면 B8 실패를 잃는다 | 확인 — `workingSet` 순서 B6 `:512` → B7 `:525` → B8 `:527` → B10 `:533`. 표시 1자리로는 둘 다 못 가른다. B10 에 명시적 제외(표시 해제)가 필요 |
| N2 | P0 | **설계 결함 — 확인** | 모드 강화의 공지가 생산에서 **동기 critical `Notify`** — 순회 뒤라도 **다음 주기**의 손절 관측을 늦춘다. D5 가 스스로 "다음 주기 시작을 늦출 수 있다" 고 적었다 | 확인 — 생산 `Announcer: ectx.Notifier`(`cmd/tossctl/engine.go:639`) → `Notifier.AnnounceOperatingMode` → `n.Notify(EventOperatingMode …)`(`internal/obs/mode.go:57`). 방향: 이 강화의 공지만 enqueue-only |
| N3 | P1 | **설계 결함 — 확인** | D7 의 로그는 생산에서 **안 남는다** — 생산 관측자에 `Log` 가 없고(`engine.go:634-640`), `Run` 은 성공 주기의 `ExitCycle` 을 버린다. 또 "로그는 등급 없음" 은 거짓 — 로거가 `SeverityOf(type)` 를 싣는다(`internal/obs/log.go:197`) | 확인(`engine.go` 옵션에 `Log` 없음, `log.go:197`). 방향: 최소 로거 배선을 범위에 넣고 배선 수준 시험, 로그 이벤트 타입 재고 |
| N4 | P1 | 설계 | 벽시계 역행이 창을 늘리고 에피소드 key 를 재사용할 수 있다 — `now − anchor` 가 벽시계 | 미확인. 방향: 경과는 단조 시계(`clock` 임대 헬퍼)로, 표시·key 는 따로 |
| N5 | P1 | 설계 | 실패 상태 전이 미정의 — 적재 성공·모드 커밋 실패, 커밋됐으나 공지 실패, 운영자 해제와의 경합 | 미확인. 방향: 적재 성공 · 강화 커밋 성공 · 실패 재시도 상태를 가르기 |
| N6 | P1 | 문서 | delta 의 "한 번의 미응답은 경보가 아니다" 는 마지막 판정 기점과 모순(그 한 주기가 판정 뒤 60초 이상이면 경보 — R3b 가 바로 그것) · B4 주기의 계수·정리 미정의 | 확인(논리). 방향: 시나리오를 경과 시간으로 한정, 계수/로그와 경보를 가르고 B4 정의 |
| N7 | P2 | 기록 | 임계 전 재시작 반복·지속 B2 는 탐지를 무기한 막는다 — 무조건 60초 상한 주장 금지 | 조건부 보장으로 명시 |
| N8 | P2 | 기록 | 장 마감 결론이 증거를 넘는다(한 사례) | "이 사례에서 관측됨" 으로 |
| N9 | P2 | 증거 | BTM B8 귀속 시험이 틀림(`TestAFailedObservationHoldsTheJudgement` 는 B4) | 정정 |
| N10 | P2 | 정의 | "관측됨" = 판정 진입 도달 — 선택자 스탬프 실패 등 즉시 오류도 관측됨으로 친다 | 지표 정의를 "판정 진입 도달성" 으로 명시, 즉시 오류 열거 |

- **판정: REJECT(P0 2 · P1 4 · P2 4). 반영하지 않았다.** Manager 결정 사항: N1(B10 제외 방식 — 표시 해제 자리 추가 = `workingSet` 편집 2자리), N2(강화 공지를 enqueue-only 로 — a094 R6-3 과 같은 우회 문제), N3(로거 배선 범위 편입).

## 3판 (2026-09-29) — codex 1라운드 반영 · 판정 아님

Manager 처분(2026-09-29)대로 썼다. 정본 `design.md`(3판).

| id | 처분 | 반영 자리 |
|---|---|---|
| N1 (P0) | B10 진입 첫 문장에 표시 해제 — `workingSet` 편집 2자리. 표시를 B10 뒤로 옮기는 안은 B8(범위 안 탈락)을 잃어 버렸다. 쌍을 구조 핀으로(해제 = B10 첫 문장, 표시~해제 사이 새 `continue`/`return` 이면 빨강, B8 은 이름으로 허용) | design D1 · D8 · spec 시나리오 「완료된 exit 정책」 · tasks 2.14·2.15 |
| N2 (P0) | 강화 **공지만** enqueue-only(모드 커밋은 동기). 공지 내용은 `AnnounceOperatingMode` 에서 순수 함수로 추출해 공유(편집 전 AST·FLM 산출 — 분기 2). **key 에 전이 id** — 창 0 적재는 id 없는 key 를 두 번째 강화부터 흡수한다(a094 R6-2 와 같은 기전). a094 D−5.2·D−5.3 교차 인용 | design D5 · spec 문장 + 시나리오 · tasks 2.3a·2.3c·3.4 |
| N3 (P1) | 최소 로거 배선 편입(`engineRuntime` 옵션에 `Log: logger` 한 줄) — 로그 주장의 성립 조건. "로그는 등급 없음" 정정(`log.go:197`) → 새 **normal** 타입 `exit.position_unobserved`. 기존 `o.log` 줄도 생산에 나가기 시작함을 기록 | design D7 · D8 · spec 문장 · tasks 2.1·2.17·3.4 |
| N4 (P1) | 경과는 단조 앵커(`clock.LeaseAnchor`/`LeaseElapsed`), 표시는 벽시계 UTC 따로, key 는 연속 시작 때 만든 **연속 id**(`opts.NewID`) | design D2·D3·D4 · spec 문장 · tasks 2.3f |
| N5 (P1) | 실패 전이 네 상태(`enqueued` · `enqueue_failed` 재시도 · `tightened` · `tighten_failed` 재시도), `ErrModeAnnouncementFailed` 는 커밋됨 + 잠금, 운영자 완화 뒤 같은 연속 재강화 없음 | design D10 · spec 문장 · tasks 2.3g |
| N6 (P1) | 시나리오를 「임계 아래」 로 한정(R3b 와 정합), B4 주기는 계수·경보 안 하고 기록 유지 | design D11-3 · spec 문장·시나리오 |
| N7 (P2) | 조건부 보장 문언(재시작 없음 · 처리 주기 존재), 지속 B2 는 명명 구멍 | design D11 · D9 |
| N8 (P2) | 장 마감: "이 사례에서 관측됨" · 빠지면 보수 쪽 거짓 양성 | design D6 |
| N9 (P2) | BTM B8 귀속 정정(틀린 시험 → 귀속 미측정) | ObserveOnce BTM |
| N10 (P2) | 「관측됨」 = 판정 진입 도달성, 즉시 오류 셋 열거·고정 | design D9 · tasks 2.16 |

## codex 2라운드 (교차 모델, task 0.6b) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: session **`01a0e969-e9b7-7740-a033-1b81b05e8d84`**, 2026-09-29 04:07:06~04:11:27 KST, rc 0, tokens 145,811, 401 없음. 트리 = `git archive f478ddc6`
  (17466 = ls-tree), Go 드리프트 `engine.go` +2(인용은 이미 이 트리 좌표 — 프롬프트 정정 0611a8b4).
- 1라운드 판정: RESOLVED 7(N1 · N2 · N4 · N6 · N7 · N8 · N9) · PARTIAL 3(N3 · N5 · N10). **두 P0(N1·N2) 해소 인정.** 모드 공지 추출은 동작 보존 가능, 키 모양 파싱 소비자 없음.

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| R2-1 | P0 | **안전 불변식 §8 — 확인(기존 관행 포함)** | 새 경보·로그가 계좌 정보를 싣는다 — `AccountRef` 는 **실제 계좌번호**(`internal/app/engine/interlock.go:680-686` "DisplayName carries accountNo")인데 D4 가 `account` 필드를 넣으면서 "계좌번호 없음" 이라 적었다. 로거 배선은 관측자의 기존 줄(계좌 ref 를 싣는 오류 줄 `exitloop.go:1730`, 기준선·위험 가격, 수량)도 생산에 내보낸다 | 확인 — `interlock.go:680-686` 인용 그대로. **기존 관행이기도 하다**: `FieldAccount: AccountRef` 가 비시험 코드 23 자리(계정 두절 경보 `exitloop.go:838` · 모드 공지 `obs/mode.go:67` · adoption · exitwiring)에 이미 있고, obs 에 가림(redact) 함수는 0. a090 은 그 관행을 **넓힌다**(새 경보 · 새 배선). 방향 후보: a090 새 경보·로그는 계좌 필드를 싣지 않고, 로거 배선은 새 normal 줄만 내보내도록 좁힌다 — 기존 관행 정리는 별도(사용자/Manager 결정) |
| R2-2 | P1 | 설계 | 모드 커밋 + 공지 적재 실패 뒤 **공지 복구 경로 없음** — `tightened` 로 가면 다시 강화하지 않고, 이후 no-op 강화는 공지를 안 한다 | 확인(논리 — `operating_mode.go:411-416` no-op 은 공지 없음). 방향: 공지 적재를 별도 상태로, 같은 전이 id key 로 재시도 |
| R2-3 | P1 | 설계 | R15 구조 핀이 약속한 "새 탈락 보호" 를 못 준다 — B5~B6 사이 새 조기 탈락은 표시 앞이라 안 잡힌다. "B8 이름 허용" 은 정확한 노드 정의가 없으면 B8 블록 안 새 `return` 을 통과시킨다. 저장된 AST 에는 순서 있는 문장 트리·`continue` 목록이 없다 | 방향: `go/parser` 로 현재 소스를 직접 파싱해 재귀 이탈 목록, B8 의 원래 `continue` 노드만 허용 |
| R2-4 | P2 | 기록 | 해제 보호는 시도 단위지 연속 단위가 아니다 — 실패→해제→재시도→실패는 새 세대로 다시 잠근다(새 증거일 수 있음). `ClearEpoch` 는 모드 완화를 막지 않는다 | 구분을 기록 + 시험 |
| R2-5 | P2 | 기록 | 임대 헬퍼는 주입 시계를 단조로 만들지 않는다(`clock.Fake` 폴백은 `Now/Since`) — 시험 픽스처 계약 명시, 계정 `checkOutage` 는 여전히 벽시계 | 명시 |
| R2-6 | P2 | 편집 | D1 의 옛 문장(B2 는 계정 사다리가 덮음 · tracer B7 도달 불가 · 하류 만료 조건 과장) · workingSet FLM 「Safety conclusion」 이 탈락 자리 편집 금지(B10 해제와 모순) | 정리 |

- **판정: REJECT(P0 1 · P1 2 · P2 3). 반영하지 않았다.** R2-1 은 a090 을 넘는 **기존 관행**(계좌번호가 critical 경보 필드·로그에 실림)을 드러냈다 — a090 범위 처분과 기존 관행 처분을 나눠 Manager/사용자 결정으로 올린다.

## 4판 (2026-09-29) — codex 2라운드 + a092 교차 반영 · 판정 아님

Manager 처분(2026-09-29)대로 썼다. 정본 `design.md`(4판).

| id | 처분 | 반영 자리 |
|---|---|---|
| R2-1 (a) | 새 경보·공지·로그에서 계좌 ref 제거(필드·key 모두). D4 의 거짓 서술("계좌번호 없음") 정정. 관측자 전체 `Log` 배선 철회 → **전용 로거 `UnobservedLog`**(새 normal 줄만, 기존 줄 생산 출력 무변화). 카나리 시험 | design D4 · D5 · D7 · D8 · D12 · spec 문장 + 시나리오 · tasks 2.17 · 3.4 |
| R2-1 (b) | 사용자 큐(Manager). 사실 요약 — 아래 「계좌 정보 사실 고정」 | design D12 |
| R2-2 | 커밋된 강화의 공지 적재 실패 → 미적재 공지 대기열, 같은 전이 id 로 재시도, 연속 종료 뒤에도, 재시작 손실은 이름 붙인 잔여 | design D10 · spec 문장 + 시나리오 · tasks 2.3g |
| R2-3 | `go/parser` 이탈 전수 핀 — 얼린 목록(감싼 `if` 조건 원문 + 블록 내 순번), B8 은 **원래 `continue` 노드 하나**만, 변이 다섯 | tasks 2.15 |
| R2-4 | 해제 보호는 시도 단위 — 해제 뒤 새 실패는 새 증거 | design D10 · tasks 2.3g ⑥ |
| R2-5 | 주입 시계 계약(`LeaseAnchor` 폴백은 `Now/Since`) — 분리 픽스처 사용, 역행 전후 앵커, 계정 사다리는 벽시계 | design D3 · tasks 2.3f |
| R2-6 | D1 의 B2 행(어느 사다리도 안 잼) · tracer B7 도달 가능 · 하류 과장 정정 · workingSet FLM Safety conclusion 정정 | design D1 · workingSet FLM |
| a092 K6 | 새 알림·공지는 a092 **단일 입구**(창 0)로. 직접 `EnqueueAlert` 금지(census 핀). 적재 실패 잠금은 입구의 생산자 래치. 입구 밖 형태면 먼저-잠금 | design D4 · D5 · spec 문장 · proposal 선후 관계 |
| 세대 읽기 시점 | 3판의 "적재 **전** 세대 읽기" 는 정본(`engine-safety/spec.md:1468-1470`) 위반 — a094 D−6.2 와 같이 정정(입구 밖 형태일 때 오류 반환 직후) | design D4 · spec 문장 · tasks 2.3g ⑤ |

### 계좌 정보 사실 고정 (R2-1 (b) — 사용자 결정 대상, 2026-09-29 측정)

- `AccountRef` 는 **실제 계좌번호**다 — `internal/app/engine/interlock.go:680-686`("DisplayName carries accountNo" · `accountRef := strings.TrimSpace(first.DisplayName)`).
- `obs.FieldAccount` 의 비시험 사용은 **20 자리**(정의 `internal/obs/log.go:45` 제외; `grep -rn FieldAccount internal cmd --include=*.go`, `_test.go` 제외,
  2026-09-29 측정 — 3판 대화 보고의 "23" 은 시험 포함 수라 정정). 그중 **19 자리가 원문 `AccountRef`**, **1 자리만 가린 값**이다(`interlock.go:441`
  `status.MaskedAccount()` → `attest.Mask`, `:273`). 파일별: adoption 4 · exitloop 2(계정 두절 경보 `:838` · 관측자 오류 로그 `:1730`) · exitwiring 4 ·
  reconcileloop 3 · runtime 3 · notifier 2(`:387`·`:394`) · mode 1(`:67`).
- 그중 **critical 경보는 ntfy 외부 전송**을 탄다(알림기 전달 경로).
- `internal/obs` 에 가림 함수는 **0** 이다. 저장소에는 **선례가 하나** 있다 — `attest.Mask`(`interlock.go:273` `MaskedAccount`) — 가림 설계의 출발점이 될 수 있다.
- 기존 관행이라 a090 범위 밖. 외부 전송 여부·가림 설계는 **사용자 결정**(Manager 가 사용자 보고에 올림). a090 은 그 관행을 넓히지 않는다.
- **5판 추가(codex 3라운드 R3-3)**: 공유 실패 경로도 계좌·원문 오류를 싣는다 — `internal/obs/notifier.go:274-280`(기록 실패 로그: 원문 `err`) ·
  `:385-396`(승격 로그: `FieldAccount, n.AccountRef`). 같은 사용자 큐 항목에 귀속한다.

## codex 3라운드 (교차 모델, task 0.6d) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: session **`01a0e9bc-63aa-7de0-bfd5-619b1c72f8c7`**, 2026-09-29 05:37:11~05:40:38 KST, rc 0, tokens 154,484, 401 없음. 트리 = `git archive 9fa0bb90`
  (17502 = ls-tree), 드리프트 engine.go +2(인용은 이 트리 좌표). 프롬프트 `codex-r3-prompt.md`(4213587c).
- 2라운드 판정: RESOLVED 3(R2-2 · R2-3 · R2-4) · PARTIAL 3(R2-1 · R2-5 · R2-6). **P0 없음.** 기존 관측자 줄의 생산 출력 무변화 확인, 세대 읽기 시점 정본 일치 확인.

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| R3-1 | P1 | **문서 결함 — 확인** | 활성 RED 가 여전히 계좌를 담은 key 를 요구한다(R3 `type\|account\|position\|…`, R3a `operating_mode:<acct>:…`) — R17 카나리와 동시에 통과 불가. GREEN 도 직접 `EnqueueAlert` 를 적는다 | 확인 — 4판에서 design·spec 은 고쳤으나 tasks 2.3·2.3a·3.1 을 안 고쳤다(내 누락) |
| R3-2 | P1 | 설계 | a092 의 공지자를 재사용하면 계좌 필드가 따라온다 — a092 는 기존 Notifier 와 같은 이벤트 필드를 요구한다 | 방향: a090 전용 정화 설정(계좌 없는 key·필드, 전이 id, 창 0)을 입구 위에. 재사용은 그 계약이 지원될 때만 |
| R3-3 | P1 | 설계 | 전용 로거는 **공유 실패 경로의 로그**를 못 덮는다 — 기록 실패를 a092 생산자 경로에 맡기면 그 경로의 오류·승격 로그가 원문 오류·계좌 ref 를 싣는다(`notifier.go:274-280`·`:385-396`) | 방향: a090 이 촉발한 기록·승격 실패의 로그를 경계에서 정화, R17 에 실패 주입 카나리 |
| R3-4 | P1 | 설계 | "a092 입구 전이면 입구 밖 + 먼저-잠금" 대안은 잠금 사유·해제 소유·성공 뒤 동작·a092 census 편입이 미정 | 방향 둘: a092 입구를 **구현 하드 의존**으로, 또는 대안의 전 수명주기를 freeze 전에 설계 |
| R3-5 | P2 | 기록 | 공지 재시도는 순회-뒤 처리 자리에서만 돈다 — B1/B2/B4 가 이어지면 재시도 안 됨 | "처리 주기에서" 로 한정 + 시험 |
| R3-6 | P2 | 핀 | 이탈 좌표가 유일하지 않다 — B12 와 B21 이 모두 `if qerr != nil` 안의 같은 순번 | 조상 경로 또는 조건 출현 순번 포함 |
| R3-7 | P2 | 픽스처 | a111 분리 시계 픽스처는 역행 **뒤** 만든 앵커의 경과를 틀리게 낸다(`wallOffset=-120s` 뒤 새 앵커는 즉시 120초) | 앵커 인식 픽스처로 |
| R3-8 | P3 | 편집 | tasks·ObserveOnce FLM 의 "tracer B7 도달 불가" · code-context 의 "여덟째 `o.alert` 호출자" 잔존 | 정리 |

- **판정: REJECT(P1 4 · P2 3 · P3 1). 반영하지 않았다.** Manager 결정: R3-4(a092 입구를 구현 하드 의존으로 올릴지 — R5-4/K6 판단과의 관계), R3-2·R3-3(정화를 a090 에 둘지 a092 입구 계약에 요구할지).

## 5판 (2026-09-29) — codex 3라운드 반영 · 판정 아님

Manager 처분(2026-09-29). **설계 freeze 는 a092 와 독립, 구현은 a092 `RecordAlert` 착지 뒤(하드 조건)** — 두 층을 가른다.

| id | 처분 | 반영 자리 |
|---|---|---|
| R3-1 | 자기 누락 수리 — tasks 2.3·2.3a 의 계좌 포함 key 기대값 → 무계좌 key, 3.1·3.4 의 직접 `EnqueueAlert`·게이트 잠금 → 입구 | tasks 2.3 · 2.3a · 3.1 · 3.4 · 1.0 |
| R3-2 | a092 announcer 재사용 안 함 — **a090 전용 정화 어댑터**(무계좌 key·필드, 전이 id, 창 0)가 입구 위에서 기록. a092 재개방 없음 | design D5 · tasks 3.4 |
| R3-3 | a090 데이터는 구성상 무계좌 → 실패 주입 카나리 셋. 공유 경로의 자기 계좌·원문 오류(`notifier.go:274-280`·`:385-396`)는 사용자 큐 항목에 귀속(좌표 추가) | design D13 · D12 · tasks 2.17 · review 「계좌 정보 사실 고정」 |
| R3-4 | a092 입구 = **구현 하드 의존**, 입구 밖 대안 문안 삭제(비례 원칙) | design D4 · D8 · spec 문장 · tasks 0.9 · proposal |
| R3-5 | 공지 재시도는 처리 주기에서만 — 문언·시험 | design D10 · tasks 2.3g ⑧ |
| R3-6 | 이탈 좌표 = 조상 경로 + 같은 조건 출현 순번 + 종류, 다중집합 | tasks 2.15 |
| R3-7 | 앵커 인식 픽스처 신규 | design D3 · tasks 2.3f |
| R3-8 | tracer B7 도달 가능 · "여덟째 `o.alert` 호출자" 정정 | tasks 1.0 · ObserveOnce FLM · codegraph-baseline |
| 부수 | 세대 읽은 뒤·적용 전 해제 RED 추가 | tasks 2.3g ⑦ |

### a094 7라운드 R7-4 교차 (Manager 2026-09-29)

- a092 재무장 요구의 대상 = a092 D0.3h 의 창 기반 exit 기록자, a090 의 에피소드 key · 창 0 기록자는 대상 밖(정합 해석, a092 델타 `:44` · D0.3h 4 표 인용). → design D14.
  a090 codex 4라운드(좁은 확인)는 이 문언이 a090·a094 두 change 에 착지한 뒤 돈다.

## codex 4라운드 (좁은 확인, task 0.6f) — **PASS**

- 실행: session **`01a0e9da-a230-7903-b6ec-54b94e113f8c`**, 2026-09-29 06:10:14~06:12:02 KST, rc 0, tokens 110,307, 401 없음. 트리 `git archive 0b5ea307`(17529). 프롬프트 485933e8.
- 3라운드 판정: RESOLVED 6(R3-2 · R3-4 · R3-5 · R3-6 · R3-7 · R3-8) · PARTIAL 2(R3-1 · R3-3 — 아래 P2 로 강등). **P0·P1 없음.**
- 확인된 것: a090 이 입구에 넘기는 알림·공지 입력은 생산 구성에서 **구성상 무계좌**(포지션 id·전이 id 는 해시 — `internal/journal/position_projection.go:441-446`,
  `operating_mode.go:720-726`; 자동 공지의 actor·cause 고정 `:504-509`). 이탈 전수 좌표는 열 개 모두 유일(B12 `result.Corruption != nil` · B21 `identityErr != nil` 조상으로 갈림).
  D14 는 새 런타임 결함 없음.

| id | codex | 분류 | 요지 | 처리 |
|---|---|---|---|---|
| N1 | P2 | 편집 | 5판 경로 정리 미완 — 직접 적재·삭제된 입구 밖 대안·계좌 필드 언급이 활성 문구에 남음(design `:122` · proposal `:57` · tasks `:53`·`:61`·`:134`) | freeze 전 정리 제안 |
| N2 | P2 | 시험 | 실패 카나리가 좁다 — 모드 커밋 실패 시 공지가 입구에 안 가 "입구 인자 무계좌" 가 **공허하게** 참. a090 소유 실패 로그 검사·계좌 섞인 오류 주입 요구 없음 | freeze 전 tasks 2.17 보강 제안(주입 실패 도달 단언 · a090 소유 로그·전체 이벤트 검사 · 계좌/오류 카나리) |
| N3 | P2 | 교차 | D14 는 정합 해석이나 a092 델타 문언은 여전히 문자 그대로 모든 exit goroutine 기록을 덮는다 — erratum 은 미적용 | Manager 가 a092 구현 로트의 명확화 후보로 등록됨 — 기록 유지 |

- **판정: PASS**(P2 3 기록). freeze 절차 대상.

## freeze 선언 (2026-09-29) — Manager 확인 대기

- freeze 전 P2 반영(Manager 판정): **N1** 활성 문구의 직접 `EnqueueAlert`·삭제된 입구 밖 대안·계좌 필드 언급 정리(design D5·D4·D10 표, proposal 표, tasks 2.3d·2.3g⑤·안전표 §8) ·
  **N2** tasks 2.17 에 ⓐ 주입 도달 단언(모드 기록 실패는 공지가 입구에 안 가므로 대기열 부재·재시도 상태로) ⓑ `UnobservedLog` 로 나간 줄 전체 검사 ⓒ 계좌 섞인 오류
  주입 — a090 소유 로그는 오류 **종류**만 싣는다. N3 는 a092 구현 로트의 명확화 후보(Manager 등록)로 기록 유지.
- 리뷰 이력: 적대 보이스 1 REJECT(1라운드) → codex 1 REJECT → 2 REJECT → 3 REJECT → **4(좁은 확인) PASS**. 결정 기록: Q1~Q3 · N1~N10 · R2-1~R2-6 · R3-1~R3-8 처분 · a092 교차(K6 · 세대 시점 · D14).
- **freeze 게이트 줄**: tasks 0.1~0.3 · 0.5 · 0.5a · 0.6~0.6h 체크. 남은 선행(구현 전): 0.4 base 재고정(구현 로트 첫 행위) · 0.8 Q2 장중 실측(사전 승인) · **0.9 a092 `RecordAlert` 착지(하드)**.


## 구현 로트 (2026-09-30) — base 재고정 · 좌표 대조 · Pre-Edit

### 선행 게이트 상태

- **0.4 base 재고정**: `1ffe2295`(단독 커밋) — `d3bd1843` → `2f698db6`. 귀속 실측: 옛 base 이후 a090 디렉터리를 만진 비병합 커밋 20 개, `.go` 편집 0.
  옛 base 의 required 47 은 형제 착지 몫. 승인 참조: 사용자 상임 지시(WORKFLOW 「사람 승인 base 재고정」) + 2026-09-30 "남은것도 처리" 재개(Manager 지시).
- **0.9 a092 입구 착지**: 작업 브랜치 `feat/a112-four-family-runtime` 에 착지·아카이브(`c6e2e3ac` 입구 · `75d138b5` archive). **로컬 `main` 참조(`62b35779`)에는
  아직 없다** — 이 저장소의 착지 대상은 작업 브랜치이며 Manager 가 a092 정본 위에서 구현하라고 지시했다(2026-09-30). 입구의 실제 모양:
  `Notifier.RecordCritical(ctx, e, remindAfter)`(`internal/obs/record_only.go:86-101`, 공개) → `recordCritical` 이 `n.mu` 아래 `Journal.RecordAlert`,
  실패 시 `Gate.Block(ReasonAlertUndelivered)` + `n.escalate`(생산자 래치, `:136-158`).
- **0.8 Q2 장중 실측**: **사람 항목으로 표기만** — 이 로트는 실행하지 않는다(Manager 2026-09-30 지시: 라이브 실측 실행 금지). 구현은 막히지 않는다(design Q2).

### 코드 좌표 대조 (frozen 문서 ↔ base `2f698db6`) — 문서는 frozen, 정정은 여기에만

`ObserveOnce`(413-470)·`workingSet`(493-612)의 AST 는 옛 base 판과 **필드 단위로 같다**(재추출 비교) — 분기 번호·줄 좌표 전부 유효.
낡은 좌표:

| frozen 문서의 좌표 | base `2f698db6` | 비고 |
|---|---|---|
| `exitloop.go:1076` `LeaseElapsed` | `:1081` | a092 편집으로 5줄 밀림 |
| `exitloop.go:1730` 관측자 `logErr` 의 `FieldAccount` | `:1735` | 같음 |
| 하류 임대 재검사 `:859` `:956` `:1027` `:1050` `:1180` | `:864` `:961` `:1032` `:1055` `:1185` | 같음(Q3 명명 잔여) |
| `cmd/tossctl/engine.go:639` 생산 `Announcer` = Notifier | `:652-658` 는 `Announcer`·`Alerts` 를 넘기지 않음 — `Context.ExitObserver` 가 `obs.RecordOnly` 로 **덮는다**(`exitwiring.go:348-351`) | a092 C1 |
| `internal/obs/mode.go:49-74` `AnnounceOperatingMode`(2 분기) · `:67` `FieldAccount` | `mode.go:48-54`(1 분기) · 사건 구성은 `operatingModeEvent`(`record_only.go:169-191`, `FieldAccount` `:184`) | **a092 K1 이 추출을 이미 했다** — a090 tasks 3.4 의 추출은 공개 래퍼 한 줄로 끝나고 이 함수는 편집 0 |
| mode 공지 key `operating_mode:<account>:<mode>`(id 뺌) | `OperatingModeEventKey` = `operating_mode:<account>:<mode>:<전이 id>`(`record_only.go:164-166`) | a092 K1. a090 key(`operating_mode:<mode>:<전이 id>`)는 그대로 — 계좌만 뺀다 |
| `internal/obs/notifier.go:262-280` 생산자 래치 | `record_only.go:136-158`(`recordCritical`) | a092 |
| `notifier.go:385-396` 승격 로그(`:387`·`:394` 계좌) | `notifier.go:425-437`(`escalate`, `FieldAccount` `:434`) | D13 귀속 그대로(사용자 큐) |
| `event.go:332` · `:348` | `:343`(표) · `:358`(`SeverityOf`) | |
| `obs/alert_lease.go:57` "54s" | `:58` | |
| `operating_mode.go:409-421` 무변화 전이 | `:415-421` | |

### Pre-Edit Gate (tasks 1.0)

```text
Pre-Edit Gate:
- change id / task id: a090-an-unobserved-position-is-counted / 1.0 (구현 2.x · 3.x)
- 대상 심볼(패키지.함수):
    engine.ExitObserver.ObserveOnce (기존) — B6·B7 원인 기록 2 · 판정 진입 표시 1 · B3 앞 처리 1 · 순회 뒤 처리 1
    engine.ExitObserver.workingSet (기존) — B6 통과 뒤 표시 1 · B10 첫 문장 해제 1
    engine.ExitCycle · ExitObserver · ExitObserverOptions (구조체 필드: Unobserved int · 비공개 주기 집계 · 기록 맵 · UnobservedLog)
    cmd/tossctl.engineRuntime (기존) — `UnobservedLog: logger` 한 줄
    새 파일 internal/app/engine/exit_unobserved.go (새 함수만)
    internal/obs: 새 공개 래퍼 OperatingModeEvent(새 함수) · 상수 EventExitPositionUnobserved
    ※ Notifier.AnnounceOperatingMode 는 편집 0(a092 가 추출 완료 — 위 좌표 표)
- CodeGraph (1.6.0, base 2f698db6 sync 뒤): callers ObserveOnce = Run(exitloop.go:354) · tracer Run(tracer.go:273) + 시험 6;
    callers workingSet = ObserveOnce 하나; callers engineRuntime = cmd 시험 4 + engine.go(engineRuntimeFactory)
- CodeGraphContext: advisory — update 가 300s 타임아웃(sdd-sync rc 2, 상습). 하드 증거는 CodeGraph + 현재 HEAD + AST. 설계 단계
    evidence-reconciliation.md 의 결론(호출 사슬)은 AST 동일성으로 유지
- 기존 동작 파악 근거: ObserveOnce·workingSet FLM/BTM(편집 전, AST 옛 base 와 동일) · engineRuntime FLM/BTM(1.1, a092 번들과 같은 sha) ·
    a092 정본 「등급화된 알림」(exit 관측 goroutine critical 은 기록까지 동기 · 발송 임차 없음 · 세울 사유 없는 기록자는 알림기 기록 입구)
- Function Logic Map / Branch Test Map: analysis/function-logic/ 네 번들
- upstream 상속 테스트 영향: no — 새 필드·새 파일. 기존 시험 하네스의 Alerts(fakeAlerts)는 기록 입구를 갖지 않으므로 a090 알림·강화는 기존 시험에서
    구조적으로 일어나지 않는다(2.0 전수 결과와 함께 R9 로 확인). 연속 id 가 opts.NewID 를 쓰므로 미관측 연속이 생기는 기존 시험의 intent id 순번이
    밀릴 수 있다 — 전체 스위트로 확인
- 실패 테스트 선행 작성: yes (2.1~2.17 RED 먼저)
- 설정·DB·journal 변경과 rollback: 없음 — 스키마 0 · 토글 0 · 새 브로커 호출 0. 롤백 = 커밋 되돌림
- 안전 불변식 §0 위반 여부 검토: 통과 — 판정·발의·주문 경로 편집 0(judge 이하 무편집), 새 알림은 순회 뒤·기록 전용(창 0),
    강화는 조이기만, 새 로그·알림·공지에 계좌 없음. 직접 원장 기록자 없음(알림기 기록 입구 RecordCritical 만)
```

### 2.0 전수 검색 (F9)

정적 전수(`internal/app/engine/*_test.go`, `h.entry`/`adopt*` 2개 이상 + `Advance`): **2** — `TestA111QuoteEvidenceUsesOnePostBatchClockAndNeverFallsBackFromBadOfficialTime`
(`Advance(2s)`) · `TestA111SlowFirstPositionExpiresLaterQuoteWithoutAbandoningStartedProtection`(`16s`). 둘 다 60초 미만 — 경보·모드 단언이 바뀔 시험 0.
보강: 기존 하네스의 `Alerts` 는 `fakeAlerts`(Notify 만)라 a090 의 기록 입구 해석이 nil 을 돌려준다 — 기존 시험에서 a090 알림·강화는 **임계와 무관하게** 0 이다.

### RED → GREEN → 변이 (구현 로트, 격리 연결 워크트리)

공유 트리에 이웃 미커밋 편집(a094 journal 시그니처)이 있어 컴파일이 깨졌으므로 Go 작업은 detached 연결 워크트리(`3260f4eb` 기준)에서 했다.

- **RED-0**(컴파일): 새 API 없음 — `analysis/red/red-0-compile.log`.
- **RED-1**(단언, API 스텁만): 실패 22 / 통과 7 — `analysis/red/red-1-assert.log`. 통과 7(R5 · R6 · R8 · R10 · R14 · R16 · R17 배선)은 「무변화 · 범위 밖」
  핀이라 편집 전에도 참인 것이 맞다(그 핀이 무는지는 변이로 잼). 주의: RED-1 의 두 트리거 픽스처는 이벤트 타입 문자열이 틀렸었다
  (`operating_mode` → 실제 `engine.operating_mode`) — GREEN 단계에서 고쳤고, 그 둘은 RED-1 에서 다른 단언으로 이미 빨강이었다.
- **GREEN**: `go test ./internal/app/engine/ ./internal/obs/ ./cmd/tossctl/ -count=1` 전부 ok(engine 349.8s — **R9: 기존 엔진 시험 전부 무변화 통과**) ·
  `make lint` rc 0 · `go test -race -run 'TestA090|TestA092|TestA111|TestExitObserverConcurrent' ./internal/app/engine/` ok(559.7s).
  GREEN 도중 기존 시험 1 이 빨강 → 설계 수정: `TestAnnouncingAQuarantineDoesNotChangeTheWorkingSet` 은 `ExitCycle` 을 `==` 로 비교한다 — 주기 집계를
  `ExitCycle` 의 비공개 포인터로 싣던 1차 구현이 그것을 깼다. 집계를 관측자 필드(`cycleUnobserved`, 주기 신원 = 그 주기의 `ExitCycle` 변수 주소)로 옮겼다.
  `ExitCycle` 에 더한 것은 `Unobserved int` 하나(비교 가능 유지).
- **변이**(`analysis/mutation/a090_mutate.py`, 사본 워크트리 · 무변이 대조군 GREEN 선행 · sha 복원 단언 · 한 판씩): 1판 **24/27** —
  생존 3(M22 주기 신원 무시 · M23 같은 연속 재적재 · M24 정리 삭제). 「동등 변이」로 넘기지 않고 시험을 보강했다:
  M22 — 옛 주기 집계는 **표시가 0 인 주기(B3)** 에서만 새어 나온다(표시가 있으면 새 집계가 시작됨) → 두 포지션을 다 없앤 뒤 B3 주기로 잰다;
  M23 — 원장 key 중복 제거가 행 수를 가려 주던 우연 안전 → 입구 호출 수를 연속당 1 로 직접 셈(R3e);
  M24 — 관측 불가능한 메모리 성장 → 좁은 시험 창 `UnobservedRecordsForTest`(export_test)로 기록 수를 셈(R8 · B4 누수 시험).
  2판 **27/27 CAUGHT** — `analysis/mutation/round1-24of27.log` · `round2-27of27.log`. 축: 기록 삭제 · 수리 되돌림(표시·해제·판정 표시·순회 뒤·B3) ·
  무조건 실행(B1·B4 에서 처리) · 순서(순회 앞 처리) · 경계 · 기점 · 에피소드 key · 재강화 래치 · 계좌 필드 · 공지 key · 동기 Notify 경로 · 대기열 ·
  벽시계 경과 · 알림 없는 조임 · 원문 오류 로그 · 생산 배선 삭제/전체 Log.
- **tasks 2.3g ⑦ not-applicable**: 착지한 입구(`recordCritical`)는 기록 오류가 돌아온 뒤 `Gate.Block` 을 **무조건** 건다 — 해제 세대를 읽는 창이 없다.
  「세대를 읽은 뒤·적용 전 해제」 경우가 존재하지 않는다. ⑤(기록 중 해제가 끼어도 잠금)는 그 무조건 잠금으로 성립하고 ⑥ 을 시험이 잰다.
- **tasks 2.17 배선의 모양**: cmd 시험은 엔진 패키지의 `OptionsForTest` 에 닿지 못하고 `Runtime` 은 관측자를 드러내지 않는다 — 생산 조립 쪽은 `go/parser`
  키 전수 핀(M26 · M27 CAUGHT), `Context.ExitObserver` 통과는 엔진 쪽 행동 시험이 잰다. 실제 로그 출력 단언은 엔진 쪽 픽스처(전용 로거 버퍼)에서 한다.
- **-race 목록 편입 not-applicable**: a090 은 새 goroutine·공유 상태를 만들지 않는다(관측자 필드는 한 goroutine 소유 — 기존 `unmanaged`·`refused` 맵과 같은 계약).

### a094 착지(`b74875e7`) 위 3-way 재적용 (Manager 창 2026-09-30)

- `git apply -3`(engine.go · cmd 시험 제외 — engine.go 창은 따로 열림) 충돌 0. a094 가 `ObserveOnce`·`workingSet` 본문을 바꾸지 않았음을 확인(base `2f698db6` 과
  바이트 동일) — 편집 뒤 좌표만 +10(a094 의 구조체 필드). FLM/BTM 「편집 뒤」 좌표를 그에 맞춤.
- **기록 입구를 a094 의 명시 주입으로 통일**: a094 가 `ExitObserverOptions.Critical`(`CriticalRecorder`, `Context.ExitObserver` 가 생산에서 항상
  `c.Notifier` 로 덮음 — `exitwiring.go:352`)을 착지했다. a090 은 처음에 `opts.Alerts` 타입 분기로 알림기를 찾았으나 같은 단일 입구를 두 방식으로
  찾을 이유가 없으므로 `opts.Critical` 로 바꿨다(새 배선 0 — a094 의 배선을 공유).
- 재적용 뒤 전판: `go test ./internal/app/engine/ ./internal/obs/ -count=1` ok(engine 393.0s) · `make lint` rc 0.

## 리뷰 라운드 1 — 분리 컨텍스트 적대 리뷰(Claude code-reviewer, 2026-09-30) — **APPROVE-WITH-FIXES** (P0 0 · P1 0 · P2 6 · P3 6)

리뷰어는 자기 overlay 변이로 저자의 27/27 목록 밖을 쟀다. 안전(손절 지연·판정 경로·계좌) 발견 없음.

| id | 분류 | 요지 | 처분 |
|---|---|---|---|
| P2-1 | 시험 | 앵커를 `o.clk.Now()` 로 바꾼 변이 생존 — 주입 시계는 lease 확장이 없어 행동 시험이 단조성을 못 봄 | **수용** — `a090_monotonic_pin_test.go` 구조 핀(LeaseAnchor 1 · LeaseElapsed ≥2 · `o.clk.Now`/`time.Now` 0). 변이 M29 |
| P2-2 | 시험 | `ErrModeAnnouncementFailed` 를 커밋 실패로 다루는 변이 생존 — 기존 시험은 입구의 자기 승격이 가렸음 | **수용** — `TestA090R3gARecoveredNoticeDoesNotRetightenAfterTheOperatorRelaxes`(복구 → 완화 → 재강화 0). 변이 M30 |
| P2-3 | 시험 | 적재된 공지가 대기열에 남는 변이 생존(창 0 이 행 수를 가림) | **수용** — 같은 시험이 대기열 0 · 입구 호출 불변을 셈. 변이 M31 |
| P2-4 | 시험 | B2 경로에서 settle 을 부르는 변이 생존(기점 삭제 → 탐지 60초 지연) | **수용** — `TestA090R3bAWorkingSetErrorKeepsTheBase`. 변이 M28 |
| P2-5 | 시험 | R15 census 가 `break`·`goto` 를 못 봄 | **수용** — census 가 모든 BranchStmt(순회 수준) · `panic` 을 이탈로 셈. 변이 6·7(break · panic 표시 앞) 추가 |
| P2-6 | 설계 | settle 의 루프 체류 몫(K 포지션 × 기록 트랜잭션 + P 공지 재시도, 실패 시 입구의 승격 트랜잭션)이 이름·예산 없이 늘어남(engine-safety 「루프에 남는 몫은 이름을 갖고 편성」) | **Manager 판정 요청** — frozen design 에 없는 항. 사실: 전부 로컬 원장 트랜잭션, 순회 **뒤**(그 주기의 손절 제출 뒤), 다음 주기 판정 앞. K 는 연속당 적재 1회로 유계(재시도는 실패 상태에서만), P 는 전이당 1. 저장소가 계속 실패하면 주기당 2K+2P 트랜잭션 — 상한 캡은 설계 변경이라 이 로트에서 넣지 않음 |
| P3-1 | 정책 | 이미 ENTRY_BLOCKED(다른 트리거)면 `changed=false` 로도 tightened — 운영자가 다른 원인을 풀면 이 연속은 다시 막지 않음 | design D10 문자 그대로(「이미 그 모드면 조인 것으로 본다」). Manager 에 정책 질문으로 올림 |
| P3-2 | 계좌 | `position_id` 는 계좌번호의 무염 해시 — 저엔트로피 계좌는 오프라인 역산 가능 | 기존 관행(원장 내부 키 허용)과 같음 — D12 사용자 큐 항목에 추가 제안 |
| P3-3 | 계좌 | a090 기록 실패가 입구의 `escalate` 로그(`FieldAccount` 원문, `notifier.go:433-444`)를 주기마다 새로 부를 수 있음 | D13 귀속(공유 경로 · 사용자 큐) — a090 이 새 경로를 하나 더한다는 사실을 그 항목에 추가 |
| P3-4 | 시험 | 엔진 조립 시험은 넣은 값을 되읽는 것에 가까움 | 인정 — 생산 조립 쪽 핀은 cmd 의 AST 키 전수(M26·M27) |
| P3-5 | 시험 | 연속 id 가 intent id 생성기를 공유 — 결정적 NewID 시험에서 intent id 가 밀림 | design D4 문자 그대로. 기존 전판 무회귀로 확인 |
| P3-6 | 잠재 | 관측자를 동시에 돌리면 맵 동시 쓰기 | 기존 맵(`unmanaged`·`refused`)과 같은 계약(한 goroutine 소유), 생산 호출자 없음 |

### 착지 준비 (a094 `d485a45f` 위, 2026-09-30)

- `engineRuntime` 편집을 `d485a45f`(a094 는 `engineRecoverySequence` 만 편집 · engineRuntime 무접촉) 위에 다시 적용. 파일 앞쪽 a094 편집으로 좌표 +11 —
  engineRuntime 번들 「편집 뒤」 좌표 갱신.
- 전판(격리 워크트리): `go test ./internal/app/engine/ ./internal/obs/ ./cmd/tossctl/ -count=1` ok(engine 445.0s · obs 59.4s · cmd 72.8s) ·
  `make lint` rc 0 · 변이 **31/31 CAUGHT**(`analysis/mutation/round4-d485a45f-31of31.log` — 저자 27 + 리뷰어 생존 4, 무변이 대조군 GREEN 선행).
  리뷰어 P2-5 의 census 변이(break · panic)는 `TestA090R15TheCensusCatchesTheFiveMutations` 안에서 빨강을 단언한다.
- `check_analysis`(격리 워크트리): a090 번들 오류 0. 남은 것은 base 뒤 형제 착지 함수(a094 `ResolveExitProposal` 등)뿐 — 착지 뒤 `--record-landing` 으로 창을 좁힐 몫.

### Manager 판정 (2026-09-30) — 리뷰 라운드 1 처분

- 설계 변경 1(기록 입구 = a094 `opts.Critical`) **승인** — 입구 단일화 방향.
- **P2-6 승인 — 캡 미도입 · 이름 붙인 잔여.** 정본 engine-safety 「등급화된 알림」: "루프에 남는 몫은 이름을 갖고 편성되어야 하며(SHALL — 이름 없는 여유는
  다음 편집이 말없이 먹는다), 그 몫이 관측 주기보다 작아야 한다". a090 이 exit 관측 goroutine 에 남기는 몫:

  | 이름 | 크기(주기당) | 위치 | 종류 |
  |---|---|---|---|
  | `a090.settle.alert` — 임계 넘은 연속의 알림 기록 | 연속당 1 회(성공 뒤 래치). 기록 실패 상태면 매 처리 주기 1 회 + 입구의 승격 트랜잭션 1 → 최악 **2K** (K = 기록 실패 중인 연속 수) | 순회 **뒤** — 그 주기의 모든 손절 판정·제출 뒤, 다음 주기 판정 앞 | 로컬 원장 트랜잭션(`RecordAlert` · `EscalateOperatingMode`) |
  | `a090.settle.tighten` — 강화 커밋 | 연속당 1 회(성공 뒤 래치). 실패 상태면 매 처리 주기 1 | 같음 | 로컬 원장 트랜잭션 |
  | `a090.settle.notice` — 미적재 공지 재시도 | 전이당 대기 1 — 실패 중이면 매 처리 주기 1 + 입구 승격 1 → 최악 **2P** | 같음 | 로컬 원장 트랜잭션 |

  정상 상태(저장소 건강)의 정상 몫은 연속 시작·해제 때의 기록 몇 건뿐이다. 최악(저장소 지속 실패)은 주기당 **2K+2P** 로컬 트랜잭션 — 원격 왕복 0.
  「원장이 멈추면 관측도 멈춘다」 독트린 안이며, 상한 캡은 frozen 범위 밖(후속 후보)이다. 이 몫이 관측 주기(5초)보다 작다는 것은 **측정하지 않았다**
  (로컬 SQLite 트랜잭션 수 ms 가정 — 실측 의무는 남는다).
- **P3-1 확인** — no-op 전이도 tightened(이미 차단된 상태에 차단 트리거가 겹치는 것은 완화가 아님, D10 · 「자동은 조이기만」).
- **P3-2 · P3-3 사용자 큐 추가 승인.** P3-3 귀속 판정: 계좌 원문을 싣는 줄은 `Notifier.escalate`(`internal/obs/notifier.go:425-437`, `FieldAccount, n.AccountRef`)
  — **base 자리**다(a092 이전부터 있던 승격 로그, a090 은 그 함수를 편집하지 않았다). a090 은 기록 실패 시 그 경로를 부르는 호출자를 하나 더할 뿐이므로
  a090 신설 표면이 아니다 → 사용자 큐(D13 항목에 「a090 기록 실패도 이 경로를 부른다」 추가). a090 이 만든 로그 표면(`UnobservedLog` 줄)은 계좌 0 — R17 카나리.
