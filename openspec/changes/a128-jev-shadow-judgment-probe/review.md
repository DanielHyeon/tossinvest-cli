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

## s4 집계·보고(2026-11-25) — tasks 5.5 · 5.6 · 5.7

읽기 전용 재집계. probe 실행·외부 API 호출·타이머·운영 config 변경 0. 원장은 동결 사본(저장소 밖, 커밋 안 함)에서만 셈. 보고서
`analysis/calibration-report-s4-2026-11-25.md`, 원문 `…-tool-output.txt`(s4·before) · `…-compare-tool-output.txt` · `…-aux-analysis.txt`(보조 스크립트 소스 포함).

실행 명령과 결과(원문 요지):

- `cp -p` 두 원장 → 사본, 복사 시각 2026-11-24T16:47:22Z. `sha256sum` 원본 = 사본: legacy `726ccb9f80da72fcbeb53d7d034e25df95c82a4fcfa33c6838fab588f0f93d86`
  (133,268,273 B · 157,578 줄) · compare `aec3442e920a69d2e097628e7080e395aa7aa057aac15eff8aabf7fa3ee509ab`(362,906,881 B · 380,175 줄), 둘 다 끝 바이트 `\n`.
- 필터(일회성 스크립트, 바이트 그대로): before `071b61feaca7…`(s3 보고서 민감도 해시와 동일) · s4 `a570249a4106…`(126,728 줄).
- `go run ./tools/a128-jev-shadow-probe report -ledger <사본>` × 3(before · s4 · compare), HEAD `75d3af9e3`, go1.26.5 → 셋 다 rc 0 · stderr 0 바이트 · 「알 수 없음」 0 행 ·
  p ≥ 0.90 권고 전부 「보류」(6 · 12 · 48 개), 예비 채택/기각 0.
- `python3 -B -m unittest discover -s tools/a145-jev-offline-analysis -p 'test_*.py'` → `Ran 16 tests … OK` rc 0; a147 동일 → `Ran 22 tests … OK` rc 0.
- `git show -s 56b73e2a1` · `git reflog --date=iso HEAD` · `git reflog refs/remotes/origin/main` → 반영 시각(tasks 5.5). `journalctl --user` 는 2026-11-11 부터만 보존.
- `systemctl --user list-timers --all` → a128 은 KR 타이머 하나(다음 2026-11-25 08:50), `is-enabled tossos-a128-us.timer` → `disabled`. 단위는 읽기만 함.

실행하지 않은 범위: 전체 시험 · `make gate` · archive(4.3 은 5.8 뒤) · Go 도구 시험(도구 코드 무변경) · bootstrap 구간(보조 수치는 점추정) ·
같은 state 를 J3·J4 유무로 묻는 대조(원장에 없음 — J1·J2 전후 비교는 기간 대조일 뿐).

- Function Logic Map: not-applicable — Go 함수 편집 0(이 로트의 diff 는 문서·분석 텍스트뿐).
- 변이 원장 · 다중 적대 리뷰: not-applicable — 비례 원칙(측정 도구의 오프라인 보고, High-risk 경로 무접촉 · 게이트 판정 변경 0).

남은 위험·미측정:

- J3·J4 의 「크기 순위」 관측은 같은 기간 안 점추정이며 독립 표본이 아님(10분마다 같은 종목 · 60분 창 겹침). 복제는 다른 기간(11-02~11-24)의 같은 원장 계열.
- s4 측정이 없는 평일(원인 미확인): KR 10-24 · 10-31 · 11-09 · 11-10, US 10-23 · 10-24 · 10-31 · 11-09 이후. US 타이머 중지 시점 미확인.
- 도구 머리말의 「2일 측정」 문장은 `report.go` 고정 문구라 s4 표본을 기술하지 않음(보고서에 적음, 코드는 안 고침).

### 6.7 일별 영수증 — 현황(실행 안 함)

- 커밋된 영수증: `analysis/d2-receipts/minute-bars/KR-2026-10-23.json` 하나(종목 30 · 호출 60 · fetched_at 2026-10-23T13:41:32Z)와
  `analysis/d2-receipts/security-types.json`(runs = KR 2026-10-23 하나, 종목 30) — 둘 다 `4e7426624`(2026-10-25). US 영수증 0.
- 창의 끝(사전등록 8절): KR 세션 날짜(KST) 2026-10-23 ~ 2026-12-24, US 세션 날짜(UTC = 개장일) 2026-10-23 ~ 2026-12-24. 판정은 US 12-24 13:00 ET + 60분 청산 뒤
  한 번, 그 전은 중간 집계.
- 빠진 날짜(원장에 D2 목록이 있는 날, 2026-11-24 까지): KR 10-26 · 10-27 · 10-28 · 10-29 · 10-30 · 11-02 · 11-03 · 11-04 · 11-05 · 11-06 · 11-11 · 11-12 · 11-13 ·
  11-16 · 11-17 · 11-18 · 11-19 · 11-20 · 11-23 · 11-24(20 일); US 10-26 · 10-27 · 10-28 · 10-29 · 10-30 · 11-02 · 11-03 · 11-04 · 11-05 · 11-06(10 일).
  원장 자체가 없는 날(규칙상 결측, 영수증 불가): KR 10-24 · 10-31 · 11-09 · 11-10, US 10-23 · 10-24 · 10-31 · 11-09 이후.
- **발견(사람 판단 필요)**: `tools/a128-d2-receipts` 의 `--ledger` 기본값과 `prereg_eval.py` 의 `LEDGER` 가 둘 다 legacy `analysis/shadow-ledger/ledger.jsonl`
  (마지막 쓰기 2026-10-31)을 가리킨다. 11-02 부터 probe 는 `shadow-ledger-compare/` 에 쓰고(기본 `--input-profile compare`), 그 원장은 같은 시각에 current ·
  initial 두 입력의 J2 를 갖고 KR 은 11-13 부터 풀 45 다. 기본값 그대로면 11-02 이후 D2 는 목록 없음 → 전부 결측으로 셀 것이고, compare 원장을 쓰려면 어느 입력의
  J2 가 「목록」인지 사전등록에 없는 선택이 필요하다(코드는 안 고침).
- 분봉 소급 한도: 6.4 실측은 「최소 9일 전」까지만 확인 — 10-26 은 이제 30일 전이라 받을 수 있는지 미측정.

### 5.8 사람 결정 — 선택지 초안(결정 아님)

1. **측정 종료 · 90% 임계 기각 기록 → 4.3(gate → archive)** — p ≥ 0.90 실현 n 이 J1 0 · J2 1 · J3 0 · J4 0(s4 질문당 15,840 판단); J2 p ≥ 0.90 이 current 입력
   약 4 만 판단에 1 건이라 n 30 에 약 120 만 판단이 필요. 단 KR 타이머는 D2 사전등록 창(~12-24)·a184 입력 비교와 공유 — 끄면 그 둘도 끝남(별도 결정).
2. **측정 계속(현 상태: KR 타이머만)** — 새로 얻는 것은 J3·J4 크기 순위의 국면 재현성뿐이고 90% 임계 표본은 사실상 늘지 않음; KR 하루 2,970 요청(45 종목 ×
   33 사이클 × 입력 2).
3. **후속 change A — 수동 보조(읽기 전용 표시)로 J3·J4 를 「60분 변동 크기」 순위로 재정의** — 사이클 안 AUC 자기 사건 0.65~0.79 · 반대 사건 0.65~0.77,
   ρ(p,|수익률|) +0.37~+0.49, compare current 에서 재현. 임계 후보는 p 절대값이 아니라 사이클 안 순위(예: J4 상위 k) 또는 J4 p ≥ 0.50(주 표본 실현 0.345 ·
   0.483 vs 전체 0.145) — 착수 전에 표본 밖 기간의 새 사전등록 필요, 방향 신호가 아니므로 매수 보조로 쓰지 않는다는 문장 포함.
4. **후속 change B — 엔진 veto(High-risk)** — 현재 근거 없음: J2 의 60분 방향 관계는 s3 역방향(KR 0.463 → 0.296) · s4 평평(KR)/순방향(US 0.484 → 0.658)으로
   국면마다 부호가 바뀌고, J3·J4 는 방향 정보가 없음(사이클 안 ρ(p, 수익률) −0.10~+0.02). 착수하면 FLM · 사람 토글 · audit 전 절차이며 임계 후보가 없음.
5. **US 측정 재개 여부(운영 토글 — 사람 승인)** — US s4 는 이미 13 세션(주 8 + 복제 5)으로 5.6 충족; 재개 이유는 D2 US 창(11-09 이후 데이터 0)뿐.
