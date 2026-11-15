# a128 리뷰 기록

## 1.8 (state_rev a128-s3 — WTS 조회 입력) — 경량 리뷰, 2026-10-15

- 범위: `tools/a128-jev-shadow-probe/` 만(새 `wts.go`·`wts_test.go` + 배선). `internal/` 무접촉, High-risk 경로 0, 게이트 판정 변경 0
  → 비례 원칙상 Function Logic Map·변이 원장·다중 적대 리뷰는 `not-applicable`. 시험 + lint + 경량 리뷰 1 회.
- 리뷰어: 독립 code-reviewer 서브에이전트 1(읽기 전용, offline `go vet`·`go test -race` 재실행 — 94 PASS). 판정 **APPROVE-WITH-FIXES**,
  blocker 0.
- 확인된 것(리뷰어): 표면 = 조회 셋(AST 고정) · 피드 범위 계좌 맥락 없음 · `AuthError`/`StatusError` 메시지에 본문·쿠키·헤더 없음 ·
  WTS 계수가 `broker_reads` 와 분리 · 거절 상태는 사이클별(시장 간 공유 없음) · 보관소 `newsMu` 보호 · `created > at` 선행 차단.

| # | 지적 (non-blocker) | 처리 |
|---|---|---|
| 1 | 시간대 없는 createdAt = KST 가정이 틀리면 선행 차단이 그만큼 약해짐 | 부분 수리: 받은 순간 + 5분보다 늦은 기사는 버리고 세며 전부면 결손. 반대 방향(실제 UTC 를 KST 로 읽어 9h 이르게 보임)은 기계적으로 못 가림 — 새는 폭은 `at`~피드 수신 사이(수 초, 피드는 사이클 초반 한 번)에 만들어진 기사뿐. **첫 s3 사이클 로그 `newest raw` 를 벽시계와 대조해 시간대를 확정할 것(사람/Manager).** |
| 2 | 0 건 피드가 news=present·빈 목록 | 수리 — 쓸 수 있는 기사 0 = 결손(`TestEmptyFeedAndEmptyFlowsAreGaps`) |
| 3 | 직전 거래일 수급 0 행이 하루 캐시되어 「0 행」으로 나감 | 수리 — 오류로 돌려 이름·재시도(같은 시험) |
| 4 | 값 0 인 지수를 이름 없이 버림 | 수리 — `major_indices: <이름> has no positive value` |
| 5 | US 종목 기사 이름 일치가 영문/한글 불일치면 언제나 빈 목록 | 미수리·기록: s2 state 실측상 공식 `/stocks` US 이름은 한글(예 AAPL=애플) → 일치 가능. 일치율은 첫 US s3 세션 state 로 잴 것 |
| 6 | 5xx·429 에 종목마다 실패 요청 반복 | 수리 — 서버 쪽 연속 3 실패면 그 사이클 수급 멈춤(`TestConsecutiveServerFailuresStopTheCycleFlowReads`·`TestFlowFailureCountIsConsecutive`) |
| 7 | 24h 창이 정리 시장의 at 기준이라 다른 시장에선 조금 넘을 수 있음 | 수리 — 고르는 쪽에서도 이 판단의 at 으로 창(`TestHeadlinesApplyTheWindowAtTheirOwnJudgment`) |
| 8 | `day_cache_hits` 에 WTS 수급 캐시 적중이 섞임 | 미수리·기록(표시용 수, 예산 판단은 `broker_reads`·`wts_requests` 로 함) |

- 자기 리뷰에서 찾은 것(리뷰 전): dry run 에서 피드 53/53 이 RFC3339 해석 실패 → news=present·빈 목록 → 시간대 없는 모양 해석 +
  결손 처리로 수리(`TestFeedWithNoParseableTimeIsAGapNotEmptyNews`).
- 남은 위험: 뉴스 시간대 가정(1번) · 뉴스 포함 입력 토큰 미실측 · US 기사 일치율 미실측 · 영숫자 KR 코드는 수급·공식 읽기 모두 결손(후속 change).

## 5.x (s4 — J3·J4 질문 추가) — 경량 리뷰, 2026-10-20

- 전문: `analysis/gstack/terra-s4-review-2026-10-20.md`(gstack `/review` preamble 이 격리 worktree 가드에 막혀 checklist 범주로 자기 리뷰).
- 판정: 이 change 범위 blocker 0, 고친 것 1(InputTokens 주석). **착지 선행 조건 1**: a145·a147 오프라인 분석기가 「한 state = J1·J2 두 행」을
  가정 — s4 행이 원장에 들어가면 a145 는 `ValueError('unpaired questions')` 로 멈추고 a147 은 전부 `invalid_pair` 로 셈(tasks 5.5).
  같은 날 이 로트에서 수리(J3·J4 행을 짝 판정 전에 빼고 수를 남김 — s4 이전 출력 불변, 변이 5/5, tasks 5.5 영수증).

## a187 반영 경계 — 기록, 2026-11-15

a187(`a187-strict-quote-readers-accept-kr-alphanumeric-symbols`)이 두 엄격 호가 리더의 KR 종목 문법을 `^[A-Z0-9]{6}$` 로 넓힘. 프로브는 systemd 사용자 타이머가 이 공유 트리에서 `go run` 하므로(`run-probe.sh`) 경계는 「코드가 이 체크아웃에 들어온 시점 뒤 첫 기동」임. 읽기 전용으로만 잼(단위 조회·journal·원장 읽기). 이 절은 tasks 체크박스를 바꾸지 않음.

- 반영 커밋: 운영 코드를 바꾼 GREEN 커밋 `305733da1`(커밋 시각 2026-11-12T20:18:10+09:00). 이 커밋은 a187 worktree(`~/tossos-work/a187-quote-symbols`)에서 만들어졌고, 이 공유 체크아웃에는 **2026-11-13 00:12:56 +0900** 에 `be90ea3e7` 로의 fast-forward 로 들어옴(공유 체크아웃 `git reflog`: `be90ea3e7 HEAD@{2026-11-13 00:12:56 +0900}: merge fix/a187-strict-quote-kr-alphanumeric: Fast-forward`, 직전 항목은 2026-11-09 21:12:38 의 `6dbb752f6`). 그 사이 공유 체크아웃 HEAD 는 `6dbb752f6` 이었음.
- 단위: `tossos-a128-kr.timer`(enabled, `OnCalendar=Mon..Fri 08:50`, `Persistent=false`) → `tossos-a128-kr.service`(`WorkingDirectory=/mnt/D/Axipient/workspace/TossOS`, `ExecStart` 가 `run-probe.sh KR … --pool 45`). `tossos-a128-us.timer` 는 disabled.
- 반영 뒤 첫 기동: **2026-11-13T08:50:42+09:00**. 근거 `journalctl --user -u tossos-a128-kr.service -u tossos-a128-us.service --since "2026-11-12 20:18:10" -o short-iso` 원문 두 줄(그 뒤 기동 없음 — 11-14·15 는 주말, 다음 예정 2026-11-16 08:50):
  - `2026-11-13T08:50:42+09:00 rog systemd[1575]: Started tossos-a128-kr.service - TossOS a128 Jev KR shadow measurement.`
  - `2026-11-13T15:20:15+09:00 rog systemd[1575]: tossos-a128-kr.service: Consumed 26.553s CPU time, 451.4M memory peak, 0B memory swap peak.`
  - 직전 기동은 `2026-11-12T08:50:12+09:00`(반영 전). 그날 기동 로그 파일 `/tmp/a128-kr-1113.log` 는 재부팅 뒤 남아 있지 않아 컴파일 시점 HEAD 를 로그로는 확인하지 못함 — 기동 시각이 반영 시각 뒤라는 것만이 근거임.
- 행 교차 확인(원장 `analysis/shadow-ledger-compare/ledger.jsonl`, 마지막 수정 2026-11-13 15:20:15, KR `judgment` 행만, 영문이 섞인 6자리 종목 = `^(?=.*[A-Z])[A-Z0-9]{6}$`; 가격·판단값은 읽지 않음):
  - 경계(2026-11-13 00:12:56) 전: 영문 섞인 종목 행 5,176 개 **전부** `collection_gaps` 에 `top_of_book: official strict quote: SYMBOL_INVALID …` 가 있음(`current`·`initial` 각 2,588). 그 문구가 있는 마지막 회차 시각 **2026-11-12T14:20:00+09:00**.
  - 경계 뒤: 영문 섞인 종목 행 792 개(종목 15개, 회차 32개) 중 그 문구 0개, `top_of_book` 결손 자체도 0개. 그 문구가 없는 첫 회차 시각 **2026-11-13T09:10:00+09:00**(그날 첫 KR 회차는 09:00 이고 그 회차에는 영문 섞인 종목 행이 없었음).
  - 경계 뒤 KR 의 `top_of_book` 결손은 숫자 종목의 `EMPTY_SIDE` 40개뿐(a187 과 무관한 응답 쪽 결손).
- 바뀌는 것: 영문 섞인 KR 종목의 호가 입력뿐임(`current`·`initial` 두 입력군 모두). 입력 판본은 그대로 — `current` 는 경계 전후 모두 `state_rev` `a128-s3`, `initial` 은 전후 모두 `initial-fields-v1`. 수급 결손(`investor_flows_per_stock`)은 a187 이 바꾸지 않음.
- 같은 경계에 겹친 변화: 같은 밤 단위에 `--pool 45` 가 붙음(단위 파일 mtime 2026-11-13 00:14:07, a184 2판 전환). 그래서 2026-11-13 부터 KR 판단 행 수도 하루 7,920 → 11,880 으로 바뀜. 두 변화의 경계가 같은 기동(2026-11-13 08:50:42)이라 기동 시각만으로는 둘을 가를 수 없음 — 호가 입력 변화는 영문 섞인 종목 행에만 있고, 판단 수 변화는 모든 행에 걸침.
