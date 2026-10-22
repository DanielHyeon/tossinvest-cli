# a140 review

## 1. 사람 결정 (2026-10-15)

Manager 전달, 사용자 「권장대로」(2026-10-15):

- **D1~D15 · D17~D20**: design §9 권고 그대로.
- **D16 변경**: 「no default veto threshold. When the S3 seam is first turned on it runs RECORD-ONLY (Jev judges, logs to ledger, never
  blocks). The veto threshold is set later by a human, per market (KR and US separately), from the a128 report after ≥5 sessions per market.」
  근거(Manager 실측, a128-s2 원장 시장마다 J2 990 건): p ≥ 0.5 면 KR 16.0% · US 20.2% 거부, KR 거부군이 60 분 +0.14% 대 −0.39% 로 더 나았음 —
  하루 · 상관 표본이라 결론 아님, 고정 0.5 의 근거가 없다는 증거.
- **D17 후속**: 「a128 is moving to state rev a128-s3 (1.8: WTS news, KR per-stock investor flows, US indices) — v1 Jev state for S3 =
  whatever a128 state rev the report calibrates (s3), not s2.」

반영: design §9 D16 · D17 행, §8 S3 행(RECORD-ONLY 먼저), §8 S3 개요(켜는 순서), §9 「사람 결정 기록」.

### 승인 뒤 구현 정밀화 (Manager 확인 대상 — 승인 범위 안이라고 판단, 이견 있으면 되돌림)

- D3: ATR 창을 opening range 15 봉으로 고정(매 분 전체 봉 ATR 은 이미 판정된 setup 을 소급 재해석 — design §3 「구현 정밀화」).
- D4: opening range 봉은 RVOL 이 정의되지 않음(요청 시 거절, lane 입력 칸 0 — 평가기는 읽지 않음, evaluateFresh B4@52 `bars[15:]`).
- 입력 봉의 연속성(개장 분부터 빈 분 없음) 요구 추가.
- **D21(새 결정 — 기본값으로 진행, 확인 요청)**: breakout v1 설정 중 생산 값이 없는 넷 — `RetestTolerancePPM = 250,000` · `MaxQuoteAgeMS = 5,000` ·
  `MaxSpreadPPM = 3,000` · `MaxEntryDriftPPM = 5,000`. 뒤의 셋은 실측값을 원장에 함께 적어 재계산 가능.

## 2. proposal-freeze 경량 리뷰 (2026-10-15)

- 보이스: 독립 리뷰어 1(별도 컨텍스트 · read-only, HEAD `044b7b94`). 비례 원칙상 경량(Low — 새 파일만). 저자 컨텍스트와 분리.
- 판정: **APPROVE-WITH-FIXES**. 코드 사실 주장 전부 참(production.go:439~440 · strategy_proposal_authority.go:353 · 370~371 · production.go:321~322 ·
  328~329 · 생산 호출 0 셋 · machine.go:111~117 · :162 · :52~70 · :101 재확인). 안전 경계 위반 0, a138 · a139 파일 겹침 0.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| 1 | P1 | ARMED 뒤 거절이 정의 안 됨 — 평가기가 ARMED 에서 끝나 setup 이 세션 내내 ARMED, 첫 시도만 보면 엔진보다 PROPOSED 가 적게 잡힘 | **수용** — design §2.1-1 · spec(R2 문장 + 시나리오): 최대 5 분 분마다 재판정, 시도별 실측값 기록 |
| 2 | P1 | `TickMinor` · `FinalCap` · FX `AsOf/FreshUntil` 미정 | **수용** — §2.1-2: 틱 = opening range 저항의 KRX 틱(세션 고정), FinalCap = 최대, FX = 개장~마감 |
| 3 | P1 | 빈 분 하나가 세션 전체를 거절(VI 단일가 미실측) | **수용** — §2.1-3: 개장부터 이어진 앞부분만 평가, 3 분 넘게 안 채워지면 gap 행, 세션 끝 미충족 = 퍼널 `GAP` |
| 4 | P2 | 폴 하한이 개장이라 200 분 뒤 종목당 2 GET(≈1 GET/s) | **수용** — §2.1-4 하한 = 가장 새 저장 봉 − 5 분, §6 예산 문장에 전제로 명기 |
| 5 | P2 | ARMED 봉 위치를 알 수 없음 | **수용** — §2.1-5 앞부분 길이를 늘려 처음 ARMED 인 n 을 찾음 |
| 6 | P2 | 재시작 결정성은 ARMED 전까지만 | **수용** — §2.1-6 원장 setupID 로 중복 방지, §2 문장 고침 |
| 7 | P2 | 봉 ID 가 판에 의존하면 정정으로 setup 이 둘 | **수용** — §2.1-7 · spec R1 문장 + 시나리오, `LaneBarID` = 세션:여는 ms |
| 8 | P2 | 왕복 비용 정수화 · RR 검사 공허 | **수용** — §2.1-8 ceil 규칙, 보고서에 「RR 거절은 구성상 공허」 |
| 9 | P2 | D3 문구 불일치 | **수용** — D3 행에 정밀화 · 사람 확인 대기 표시(tasks 0.4) |
| 10 | P2 | 경로 가드가 하드링크를 못 막음 · 생산 경로 출처 미명시 | **수용** — §2.1-9 이름 · 이웃 journal.db · `os.SameFile`, 출처 = `journal.DefaultPath()` 디렉터리 · 설정 디렉터리 |
| 11 | P2 | 키 공유 — 호가 GET 이 엔진 보호 경로와 같은 그룹 | **수용** — §2.1-10 MARKET_DATA 429 한 번이면 호가 5 분 정지 |
| 12 | P2 | 쓰기 목록에 토큰 캐시 갱신 누락 · 설정 저장 금지 | **수용** — §2.1-11, AST 시험이 config 저장 이름 금지 |
| 13 | P3 | spec 「자기 전용 저장소에만 써야」가 원장 · 토큰 쓰기와 충돌 | **수용** — 「evidence 쓰기는」 |
| 14 | P3 | 빠진 시나리오(ATR 0 · 첫 봉 · 분모 0 · 올림) | **일부 수용** — ATR 0 · 정정 시나리오 추가. 첫 봉 · 분모 0 · 올림은 단위 시험이 덮음(signal_test.go) |
| 15 | P3 | 마지막 봉 미적재 · 마감가 출처 | **수용** — §2.1-12 |
| 16 | P3 | SealBarSeries 비용 | **수용** — §2.1-13 |
| 17 | P3 | `internal/config` import 가 a139 편집 패키지에 컴파일로 묶임 | **수용(위험 기록)** — 파일 겹침 0, a139 가 config 를 바꾸면 도구 빌드만 다시 확인 |
| 18 | P3 | 원장에 위험 한도 출처 기록 — 커밋 전 확인 | **수용** — 3.3 커밋 전 확인 항목 |

FREEZE: 위 P1 셋을 문서로 닫은 뒤 착수(같은 커밋). 남은 사람 확인: D3 정밀화 · D21 기본값(review §1).

## 3. 코드 경량 리뷰 (2026-10-15)

- 보이스: 독립 리뷰어 1(별도 컨텍스트 · read-only, 커밋 `ac32dc9d`, `-overlay` 로 /tmp 시험만 얹어 실측 — 저장소 무변경). 판정: **BLOCK**.
- 안전 경계 · breakoutsignal 정의 · evaluate 배선(틱 · FX · FinalCap · RR · ArmedIndex · 호가 실측식)은 정상 확인.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| 1 | P1 | ARMED 평가 시각이 호가 읽기 **전** — 실제 시계에서 시도 5 회 모두 QUOTE_STALE, PROPOSED 0 | **수리** — 읽은 뒤 `max(now, received_at)`. 변이 R1 CAUGHT |
| 2 | P1 | 라벨이 ArmedIndex+1 부터 — 제안 전에 닫힌 봉으로 결과를 매김 | **수리** — `label_from_ms` = 제안 분 다음 개장 분, spec R3 문장 추가. 변이 R2 CAUGHT |
| 3 | P1 | 시험 시계가 step 안에서 멈춰 위 둘을 가림 · 목표 봉이 제안 전 | **수리** — 가짜 브로커가 호출마다 50 ms 진행, 목표 봉을 제안 뒤(09:20)로, ask 와 봉 정합 |
| 4 | P2 | 컷오프가 폴 전 시각 — 감지 1 분 지연 | **수리** — 폴 뒤 시각. 변이 R3 CAUGHT |
| 5 | P2 | ARMED 상한이 시도 수뿐 | **수리** — 시간 창 5 분, spec R2 문장 보강. 변이 R4 CAUGHT |
| 6 | P2 | 결손 행이 세션당 한 번 — spec 「그 분의 결손 행」과 어긋남 | **수리** — 종목 · 단계 · 분 단위, 계수는 budget 행에도. 변이 R9 CAUGHT |
| 7 | P2 | 봉 그룹 포기 시 나머지 종목 건너뛰기 미구현 | **수리** — 변이 R5 CAUGHT |
| 8 | P2 | 랭킹 · 달력 포기가 결손 행 없이 종료 | **수리(랭킹)** — 결손 행 뒤 거부, 변이 R7 CAUGHT. 달력 포기는 세션 신원 전이라 표준 오류로 종료(원장 행 없음 — 세션이 시작되지 않음) |
| 9 | P2 | 라벨이 앞부분만 봄 | **수리** — 전체 묶음. 변이 R6 CAUGHT(시험 추가 뒤) |
| 10 | P3 | 마감 뒤에도 새 setup 평가 | **수리** — 마감 뒤 라벨 대기만 |
| 11 | P3 | 재시작 때 판(lot) 미대조 | **수리** — 다른 판이면 거부. 변이 R8 CAUGHT |
| 12 | P3 | 세션별 n≥30 판정은 사실상 영구 보류 | **수리** — 판 합계 판정 추가. 변이 R10 CAUGHT |
| 13 | P3 | 가짜 리더 주석이 「자라는 봉을 내줌」 | **수리** — 주석을 엄격 리더 계약대로, 실측 대조는 첫 장중 실행 |

변이 원장: `analysis/red/mutation-tool.txt`(1 판 14 · 2 판 14, 생존 0 — 생존은 시험 추가 뒤 재실행으로 CAUGHT).

## 4. 사람 결정 — 측정 뒤 (2026-10-22)

사용자 원문(Manager 전달, 2026-10-22):

- 「a140: 중단」
- 「a140 승인 확인: 조회 예산(D11)과 호가 한도 기본값(D21)을 지금대로 둬」

사용자가 본 근거(Manager 가 원장에서 집계, 가상 체결 `worst` · 주당 왕복 비용 `round_trip` ≈0.5%) — `analysis/report-2026-10-22-supplement.txt`
「가상 손익」 절에서 같은 원장(sha256 `0b802a92…740c`)으로 다시 재서 일치 확인: 제안 22 · 순이익 거래 7 · 거래당 평균 순 −0.70% ·
순 합계 −65,780 KRW · 가상 명목 9,267,200 KRW. 세션별 순: 10-16 −55,095 · 10-19 +61,177(상승일) · 10-20 −11,526 · 10-21 −21,338 ·
10-22 −38,998 — 양수 세션은 10-19 하나. n=22<30 이고 세션끼리 상관이라 통계 판정이 아니라 사람 판단의 관측 근거다.

반영:

- **S4(엔진 증거 생산) 착수 안 함.** breakout 가족 측정은 여기서 멈춘다. 지표 판본 `breakoutsignal-v1` 은 고치지 않는다(수정 change 없음).
  `internal/breakoutsignal` · `tools/a140-breakout-signal-shadow` 는 생산 import 0 인 채로 남는다(삭제는 이번 결정 범위 밖).
- **D11 · D21 사람 확인 닫힘.** §1 「남은 사람 확인」 중 D21 기본값, 그리고 3.1 이 적은 「D11 단독 영수증 없음」이 이 원문으로 닫힌다.
  D3 정밀화(ATR 창 고정)는 이번 원문에 없다 — 2.2 Manager 수용 그대로이고, 가족이 멈춰 후속 영향이 없다.
- **다른 가족 검토 여부**는 이번 결정에 없다. 이 change 는 정하지 않는다.
- KR 측정 타이머 `tossos-a140-kr.timer` 는 Manager 가 2026-10-22 ≈19:55 KST 에 껐다(disable). 재확인(팀메이트, 같은 날):
  `systemctl --user is-enabled` = `disabled` · `is-active` = `inactive`.
- 서술형 보고서 `analysis/report-2026-10-22.md` 는 작성 도구가 하위 에이전트의 보고서 파일 쓰기를 막아 만들지 못했다(두 번째 시도도 거부).
  수치 영수증은 위 두 txt 가 정본이고, 결정 기록은 이 절이다.

## 5. 완료 게이트 시도 — 2026-10-22 (FAIL, 5 단계)

- 격리 worktree `/tmp/tossos-a140-close-20261022`, HEAD `697002d7d`(4.2 기록 + 4.3 을 게이트 2 단계 전 체크한 시험 커밋 — 공유 브랜치에는 올리지 않음).
  `ls -d /tmp/.git` → 없음(No such file or directory).
- `make sdd-sync` → rc=0, `[sdd-sync] all indexes current`. GBrain 코드 import 는 `Sync BLOCKED … 12641 file(s) failed to parse`(advisory, worktree-로컬 홈).
- `make gate CHANGE=a140-breakout-signals-are-measured-in-shadow` → **GATE FAIL, rc=2**:
  1 `OK: …/tasks.md` · 2 `OK: 미완료 태스크 0건` · 3 `OK: 짝 선언 없음 — 이 change 는 혼자 배포한다` · 4 `OK: …/review.md` ·
  5 `[logic-map] a140-breakout-signals-are-measured-in-shadow: base 526637460edc → working tree (no landed-commit.txt in HEAD) required 129 function(s) — judged at HEAD 697002d7d31a`
  · `this window also holds 159 commit(s) that landed after the base and every existing function they changed is required here too — `--record-landing` cannot narrow it: no landing recorded — no `revision: current` evidence pins a landing for this change`
  → `GATE FAIL: a140-breakout-signals-are-measured-in-shadow — Function Logic Map 산출물 미완료`. 6~11 단계는 돌지 않았다.
- 원인: a140 은 새 파일만 바꿨지만(FLM not-applicable) base `526637460` 뒤 다른 change 의 커밋 159 개가 창에 들어와 그들이 바꾼 기존 함수 129 개를
  요구한다. 닫으려면 base 재고정(docs/WORKFLOW.md 「사람 승인 base 재고정」)이 필요하고 그것은 사람 승인 사항이다. 4.3 은 열어 둔다.

## 6. base 재고정 — 사람 승인 (2026-10-23)

- **승인**: 사용자 결정(Manager 전달, 2026-10-23) — 「base 재고정 승인」, 그리고 이 change 의 완료 gate
  (`make gate CHANGE=a140-breakout-signals-are-measured-in-shadow`) 실행 승인. docs/WORKFLOW.md 「사람 승인 base 재고정」 2 항의 기록이 이 절이다.
- **귀속 실측**(1 항, 범위 `526637460..34e12967e` = 173 커밋, 2026-10-23):
  - a140 디렉터리를 만진 비병합 커밋 9 중 `.go` 를 고친 것 4 — `e076649d1` · `ac32dc9dc` · `c17d48167`(a140 자신) 과
    `6f06eee8e`(`fix(verify): exempt TTL edge…` — a144 의 Go 커밋이 a140 tasks.md 2.1 체크를 함께 실었다. 그 Go(`internal/verifylive/m0_recovery.go` ·
    `cmd/tossctl/verify_test.go`)는 a144 아카이브 번들 `cmd-tossctl--seedverifyrecord` · `internal-verifylive--m0exactprerequisites` ·
    `internal-verifylive--officialrawforpending` 이 덮는다 — a140 의 작업이 아니다).
  - a140 디렉터리를 만지지 않은 a140 Go 커밋 1 — `c564a89bf`(`policy.go` 신설 등). 경로 필터(`internal/breakoutsignal` ·
    `tools/a140-breakout-signal-shadow`)로 찾았다.
  - a140 Go 커밋 넷(`e076649d1` · `ac32dc9dc` · `c17d48167` · `c564a89bf`)이 만진 파일은 전부 옛 base 에 없던 파일이다
    (`git ls-tree -r 526637460 -- internal/breakoutsignal tools/a140-breakout-signal-shadow` → 0 줄, `git diff --name-status` 17 줄 전부 `A`).
    따라서 a140 이 고친 **기존** 함수는 0 이고, fresh 번들로 덮어야 할 대상도 0 이다.
- 옛 base 에서의 required 129(창 안 형제 커밋 159)는 전부 형제 착지 몫이다(§5).
- 새 base = 이 기록 커밋. `base-commit.txt` 만 담은 다음 단독 커밋의 메시지가 옛→새 sha 영수증이다(3 항).

### 6.1 완료 gate 1 차 — 2026-10-23 (FAIL, 5 단계 — 면제 표식 부재)

- 격리 worktree `/tmp/tossos-a140-close-20261022` 를 `5975af2b9`(4.3 체크 · tracker 재생성)로 옮김. `ls -d /tmp/.git` → 없음.
- `make sdd-sync` → rc=2: CodeGraph 동기화 · fingerprint 기록(head `5975af2b9`)은 됐고, GBrain project sync 가 1800 초 시간 초과(advisory).
- `SDD_PYTHON=/tmp/tossos-a148-sdd-venv/bin/python make gate CHANGE=a140-breakout-signals-are-measured-in-shadow` → **rc=2**: 1~4 OK,
  5 `[logic-map] … base 651de4394ef7 → working tree (no landed-commit.txt in HEAD) required 0 function(s) — judged at HEAD 5975af2b9550`
  · `[logic-map] missing analysis or \`Function Logic Map: not-applicable\` review marker` → `GATE FAIL`.
- 재고정으로 required 는 129 → 0 이 됐다. 남은 실패는 review.md 에 정본 면제 문구(`check_analysis.EXEMPTION`)가 없다는 것뿐이다 — tasks.md 머리에는
  있었지만 게이트는 review.md 를 읽는다(`tools/logic-map/check_analysis.py:2886`). 아래 줄이 그 기록이다.

Function Logic Map: not-applicable — a140 은 옛 base 에 없던 새 파일만 만들었다(`internal/breakoutsignal/` · `tools/a140-breakout-signal-shadow/`, 17 개 전부 `A`,
§6 귀속 실측). 기존 함수 편집 0 · 게이트 판정 변경 0(docs/WORKFLOW.md 「비례 원칙」). 재고정 뒤 5 단계 required 0.
