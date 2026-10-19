# Manager 인수인계 — 2026-10-19 (월)

작성: 2026-10-18 Manager 세션 · 기준 HEAD `9eb3bc4e7` (branch `feat/a112-four-family-runtime`, 푸시됨)

이 문서를 받은 세션은 먼저 `CLAUDE.md` → `.claude/CLAUDE.md` → `docs/WORKFLOW.md` 를 읽는다.
안전 부트스트랩이 이 문서보다 우선한다. 실계좌 주문 · `mutating: true` 명령 · 운영 토글은 에이전트가 실행하지 않는다.

---

## 0. 지금 상태 한눈에

| 항목 | 상태 |
|---|---|
| P0 a134 · a135 | archive 됨 |
| P1 a136 · a137 · a138 · a139 | archive 됨 (하루 매수 한도 · 일일 손실 실측 · 최대 보유 6 · 건당 위험 예산 = 일일 손실/3) |
| a146 Guardian 상한 0 → 그 진입만 건너뜀 | archive 됨 (`9eb3bc4e7`) |
| a128 Jev 그림자 측정 | 진행 중 — 측정 누적 단계(아래 §4) |
| a140 KR 돌파 신호 그림자 측정 (P2 S1) | 진행 중 — KR 1/5 세션(2026-10-16) |
| 배포 | **미배포**. 사용자 지시: "배포는 P3 까지 끝나면 바로". P2 · P3 남음 |
| 인증(attestation) 만료 | **2026-10-23 10:39 KST** — §2 사람 작업 순서대로 |

---

## 1. 작업 A — 전략 레인을 켜기 전 필수 후속 2건 (에이전트가 진행)

출처: `openspec/changes/archive/2026-10-17-a146-guardian-cap-zero-skips-the-entry/tasks.md` 「후속」,
`design.md` F10 · D5, `review.md` 구현 리뷰 1 P3-5. 사람 결정(2026-10-18): **P2 S5/S6 (전략 레인 활성화) 전에 착지해야 함.**

### A1. execgw 결속 판정을 사이징 앞으로 (F10)

- **문제**: q_final 사이징이 상한 0 을 내면 그 뒤의 Guardian 판정(정책 버전 · 주간 예약 결속 · 투영 · 버킷 ·
  체인 · 노출)이 돌지 않는다. a146 뒤로는 그 자리에 동시에 있는 결함이 잠금 대신 **조용한 스킵**(주문 0)이 되고,
  상한이 1 이상이 되는 주기에야 드러난다. 값 결함(FX 단위 · 가격 ×100)도 같은 모양으로 숨는다.
- **할 일**: `PrecheckQFinalCampaignFirstLeg` 의 순수 판정(정책 버전 · 주간 예약 결속 등)을
  `PrecheckQFinalEntry`(사이징) **앞**으로 옮긴다. 판정 내용은 그대로, 순서만.
- **위험도**: High (Guardian · execgw). 풀 절차 — FLM/BTM 먼저(순서 주장은 AST 열거로), 특성 시험으로
  현재 순서를 RED 로 고정 → GREEN → 변이 → 독립 적대 리뷰.
- **주의**: 가드를 앞으로 옮기면 뒤 가드가 못에서 빠진다(기억 「새 가드는 뒤의 가드를 못에서 뽑는다」) —
  옮긴 뒤 각 가드가 여전히 자기 시험으로 죽는지 변이로 확인할 것.

### A2. 조립 감독자 알림 배선 (a146 리뷰 P3-5)

- **문제**: `NewPairedStrategyEntryProductionAssembly` 가 만드는 감독자(`internal/app/engine/strategy_entry_supervisor.go`
  근처, Effective 서술자를 가질 수 있음)에 `DailyLossNotice` · `EntrySkipNotice` 옵션이 없다.
  오늘은 투영에만 쓰이고 Run 되지 않아 실해는 없다. Run 하게 되면 일일 손실 차단 · 상한 0 스킵이 스냅숏 계수로만 보이고 경고 줄이 안 나간다.
- **할 일**: 생산 감독자(`NewRefreshingPairedStrategyEntrySupervisor`, `cmd/tossctl/engine.go` 의 호출 자리)와 같은
  두 옵션을 조립 감독자에도 배선. 시험: 조립 감독자로 스킵 · 일일 손실이 알림까지 가는지.
- **위험도**: Medium (알림 배선, 판정 불변). A1 과 **한 change 로 묶어도 되고** 나눠도 된다 — 나누면 A1 먼저.

### change 번호

`a141`~`a156` 은 이미 쓰였다(다른 세션). **쓰기 전에 `ls openspec/changes openspec/changes/archive | grep a1` 로
빈 번호를 확인**하고 그다음 번호를 쓴다. Story 는 FEAT-TOS-013(전략 신호) 또는 risk 쪽 기존 Feature 아래.

---

## 2. 사람 작업 (에이전트는 준비 · 안내만 — 실행하지 않는다)

순서가 중요하다. 1 → 4 는 **10-23 10:39 KST 전**에 끝나야 한다.

1. **a087 5.1 KR 시장가 매도 1주 재시도** (장중 09:00~15:20 KST, 사람 터미널)
   - 대상 010170 1주. 실행 전 앱에서 매도가능수량 ≥1 · 같은 종목 SELL 조건주문 없음 확인.
   - 도구 소스가 착지 커밋 그대로인지: `git status --porcelain -- tools/a087-market-sell-probe` 빈 출력.
   - 두 명령을 **5분 안에** 이어서(미리보기는 전송 없음, 토큰 5분):
     ```
     go run ./tools/a087-market-sell-probe --config-dir ~/.config/tossctl --symbol 010170 --qty 1
     go run ./tools/a087-market-sell-probe --config-dir ~/.config/tossctl --symbol 010170 --qty 1 --execute --confirm <토큰>
     ```
   - 종료 코드: 0 접수 · 2 브로커 거절 · 3 불명(→ 미리보기 재실행 금지, 먼저 `tossctl orders list`).
   - 이전 시도: 2026-10-15 15:13 · 15:16 둘 다 `401 invalid-token` → 「판정 불가(인증)」(`f6bcc95a`).
     그 뒤 사용자가 OpenAPI 키를 새로 발급해 `tossctl openapi login`(10-15 21:45), WTS 세션도 갱신(22:12).
     토큰 캐시는 10-17 04:52 갱신 — 새 키로 받은 것으로 보이나 **POST 로는 아직 미확인**.
   - 에이전트 몫: 영수증(`openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/market-sell-receipts/`)
     을 읽어 tasks 5.1 처분을 기록 · 커밋. 401 이 또 나오면 이웃 단정 금지(기억 「공유 credential 토큰 전쟁」 갱신분) —
     같은 시각 GET 이 성공했는지, `TOSSCTL_OPENAPI_*` env 덮어쓰기, 키의 주문 권한부터 본다.
2. **감독하 실주문 재검증**(capability verification) — 09-09 의 감독 증명이 10-09 에 만료돼, 지금 갱신하면 POST 범위가 빠진다.
3. **soak 회복** — soak 거절이 unit 의 `|| printf` 에 가려져 있었다(a135 실측).
4. **attestation 갱신**.
5. **스킬 플러그인 정리** (에이전트 실행은 권한 분류기가 막음 — 사용자 터미널):
   ```
   claude plugin marketplace add JuliusBrussee/caveman && claude plugin install caveman@caveman --scope user
   claude plugin marketplace add ayghri/i-have-adhd && claude plugin install i-have-adhd@i-have-adhd --scope user
   claude plugin marketplace add DietrichGebert/ponytail && claude plugin install ponytail@ponytail --scope user
   claude plugin uninstall chisle
   ```
   - 2026-10-18 사전 점검(커밋 caveman 81958ad · i-have-adhd 1ecea19 · ponytail 16352ac): hook 은 지시문 주입 ·
     자기 설정/플래그 파일 쓰기뿐. Pre/PostToolUse 출력 재작성 없음, hook 스크립트에 외부 전송 없음, `settings.json` 쓰기 없음.
   - **caveman CLI(`npx @caveman-ai/cli`)는 설치하지 않는다** — 입력 압축 proxy · 기본 텔레메트리가 규칙 14 와 충돌.
   - 필수 개발 규칙 14~16 은 이미 세 스킬로 바뀌었다(`bec03a2c9`).
6. **배포 전 확인(a138)** — 생산 원장에 해석 불가 preimage 를 가진 HELD 예약이 있으면 교체 직후 모든 진입이
   fail-closed 로 막힌다. 교체 창 전에 읽기 전용 질의로 확인(질의는 에이전트가 준비, 실행은 사람이 원장 사본에서).
7. **엔진 교체는 KST 05:00~09:00 창에서만**(엔진 정지 = 손절 없음).

---

> **진행(2026-10-19 14:21 KST, 실측):** 2 · 3 · 4 완료.
> - 3 soak 은 스스로 회복(soak days 3)했다. 그 결과 갱신 타이머가 00:21 · 06:21 에 attestation 을 **GET 전용으로 덮어썼다**(a135 예고대로).
> - 사람이 KR 장중에 엔진을 정지했다. 그 뒤 `verify run --redo order-cancel`(`--config-dir ~/.config/tossctl`)을 돌려 POST orders 05:15:53Z · POST cancel 05:16:29Z 를 감독 증거로 남겼다.
> - 이어 `tossos-attest.service` 갱신 → **POST 포함 11 endpoint · 감독 증거 2 · 만료 2026-11-18T05:16:59Z**.
> - `docker start tossos-tossos-1` 기동 상태는 10-15 와 같다: gate_enabled · gate_verified True, protection UNWIRED, entry_permitted False.
> - 정지 중 살아 있는 포지션은 US 1 · KR 0 이었다. US 손절 판단은 정규장만이라 공백이 없었다.
> - 남은 사람 작업: 1(a087 매도 — `--config-dir` 필수) · 6 · 7.

> **배포(2026-10-19 20:26 KST, 사용자 결정 「P3 전 지금 배포 · 15:35~21:00 창」):** 488b370f → e537ab23c(이미지 788dd6995faa).
> 직후 gate verified · protection UNWIRED · entry_permitted false(전과 같음) · mutation delta 0 · reconcile clean. 롤백 핀 1825e2bc(schema 35).
> 기록: `~/.local/share/tossos-deploy/release-20261019-{runbook,review}.md` · `preflight-20261019T091821Z/post-replace.txt`. a087 5.1 = accepted(FILLED).

## 3. 사용자 결정 대기 — 동시 진행 change 수

- 규칙: 동시 change 최대 4개(2026-09-26 사용자 지시, 기억 「SDD 작업 방식」).
- 현재 열린 change(archive 제외): a063 · a074 · a087 · a100 · a105 · a107 · a128 · a140 · **a147(미커밋)** ·
  **a148~a156(다른 세션이 2026-10-18 `ed0927f28` 로 정의)** · verify-execution-capability · verify-observes-the-trigger.
  a148~a156 만 9개다.
- 선택지(사용자에게 번호로 묻는다 — AskUserQuestion 도구는 이 환경에서 안 보임):
  1. 그대로 둔다(정의만 된 change 는 「진행 중」으로 치지 않는다고 명시).
  2. 다른 세션에 맞추게 한다 — a148~a156 중 착수 4개만 남기고 나머지는 「대기」로 표시.
  3. 다른 세션 작업을 멈추고 Manager 가 순서를 다시 정한다.
- 결정 전에는 작업 A 를 시작해도 되는지 사용자에게 같이 묻는다(작업 A 는 레인 활성화 선행 조건이라 우선순위가 높다).

---

## 4. 계속 돌려야 하는 측정

> **정정(2026-10-19 09:5x, 실측):** 이 문서의 명령에는 원래 `--config-dir` 가 없었다. 빼면 도구가 기본 토큰 캐시(`~/.cache/tossctl`)로
> 새 토큰을 받아 같은 앱 키를 쓰는 엔진 토큰(`~/.config/tossctl`)을 무효화한다(a082 토큰 전쟁). 10-19 오전 a128 · a140 측정과
> `tossctl verify run` 이 그렇게 돌아 두 토큰 파일이 3~4 분마다 번갈아 다시 쓰였고 엔진 대사가 401 을 한 번 맞았다(09:40:40).
> `--config-dir ~/.config/tossctl` 로 다시 띄운 뒤 9 분간 재발 0. **운영 계정에 닿는 모든 도구 · tossctl 명령에 `--config-dir ~/.config/tossctl` 을 붙인다.**
 (읽기 전용 GET — 에이전트 실행 가능)

> **자동 기동(2026-10-19 21:1x, 사용자 지시):** 아래 명령을 이제 systemd 사용자 타이머가 평일마다 띄운다 —
> `tossos-a140-kr.timer` 08:45 · `tossos-a128-kr.timer` 08:50 · `tossos-a128-us.timer` 22:00 KST(유닛은 `~/.config/systemd/user/`,
> 로그는 `/tmp/a1{28,40}-<kr|us>-<MMDD>.log`). 손으로 같은 측정을 또 띄우지 말 것(같은 원장 중복 쓰기). 확인은
> `systemctl --user list-timers | grep -E 'a128|a140'`, 시장별 5 세션을 채우면 `systemctl --user disable --now <timer>`.
> a128 KR 은 `--session "09:00-15:30"` 로 달력을 우회하므로 **평일 KR 휴장일에는 타이머를 미리 꺼야** 한다(다음 휴장 12-25).

| 측정 | 남은 것 | 실행 |
|---|---|---|
| a128 Jev s3 | 시장별 ≥5 세션(Jev 거부 임계는 그 뒤 사람이 시장별로 — a140 D16). 지금 s3: KR 10-16 1세션, US 10-15 · 10-16(일부) | `setsid nohup openspec/changes/a128-jev-shadow-judgment-probe/run-probe.sh KR /tmp/a128-kr-<MMDD>.log --session "09:00-15:30" --wait-open --config-dir ~/.config/tossctl` · US 는 `run-probe.sh US /tmp/a128-us-<MMDD>.log --wait-open --config-dir ~/.config/tossctl` |
| a140 KR 돌파 | 4 세션 더(10-16 1/5 완료) | `go run ./tools/a140-breakout-signal-shadow run --evidence-db ~/.local/state/tossos-a140/shadow-evidence.sqlite --config-dir ~/.config/tossctl` (그날 08:40 이후 기동 — 달력의 오늘 세션을 읽으므로 전날 밤에 띄우면 「session over」로 끝남) |

- 둘을 같은 날 돌려도 MARKET_DATA 그룹 한도 안(a140 D11).
- 프로브는 시작할 때 키 · WTS 세션을 한 번 읽는다 — 키나 세션을 바꾸면 다시 띄운다.
- 측정이 원장에 쓰는 동안 공유 트리에서 `make gate` 를 돌리면 fingerprint 가 낡는다 → 사이클 직후 sync+gate 연속,
  또는 격리 워크트리에서 게이트(§5).

---

## 5. 이번 세션에서 배운 게이트 운용 (재발 방지)

- **공유 트리에 다른 세션의 미커밋 편집이 있으면 게이트는 격리 워크트리에서**: `git worktree add --detach <scratch>/wt-X <HEAD>` →
  `make sdd-infra` → `python3 tools/sdd/sdd_sync.py --no-gbrain` → `make gate CHANGE=…`.
- **PM tracker 재생성도 격리 워크트리에서**, 결과 blob 을 index 로만 커밋:
  `git hash-object -w <wt>/<f>` + `git update-index --cacheinfo 100644,<blob>,<f>` (먼저 `git diff --cached --name-only` 가 빈지 단언).
- **`/tmp/.git` 빈 디렉터리 함정**: /tmp 아래 워크트리에서 `go list` 가 `error obtaining VCS status: exit status 128` →
  `ls -d /tmp/.git` 확인. Go 는 `.git` 이 디렉터리인 곳만 루트로 봐서 워크트리를 지나 /tmp 까지 올라간다(2026-10-18 제거함).
- **`--record-landing` 은 공유 트리에서 거부**(a128 ledger 수정 중) → 격리 워크트리에서 계산해 `landed-commit.txt` 복사 · `git add` 후 커밋.
- **archive 커밋에 메인 spec 병합이 들었는지 stat 으로 확인** — a139 archive(`e54d4efb`)가 빠뜨려 `befea6ee` 로 보충했다.
- **부하 flake**: `TestA091TheReportFitsItsShare` 는 load 10 이상에서 시간 초과로 실패할 수 있다(단독 3/3 통과 확인). 재시도 전 단독 실행으로 가를 것.
- codegraph `status` 15s 시간 초과 · CodeGraphContext 300s 시간 초과는 advisory — 재시도.

---

## 6. 참고

- 프로젝트 기억: `~/.claude/projects/-mnt-D-Axipient-workspace-TossOS/memory/tossos-project-state.md` 끝부분(2026-10-14~18 항목)
- a146 archive: `openspec/changes/archive/2026-10-17-a146-guardian-cap-zero-skips-the-entry/`
- a140 설계 · 사람 결정 D1~D21: `openspec/changes/a140-breakout-signals-are-measured-in-shadow/design.md` §9, `review.md` §1
- a128 s3 정의: `openspec/changes/a128-jev-shadow-judgment-probe/design.md` s3 절, tasks 1.8
- P2 Story 분할 S1~S6: a140 `design.md` §8 (S3 Jev seam = RECORD-ONLY 로 먼저, 임계는 a128 보고서 뒤 시장별)
