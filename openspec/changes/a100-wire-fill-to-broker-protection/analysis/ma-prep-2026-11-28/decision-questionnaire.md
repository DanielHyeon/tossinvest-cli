> **보존 사본(2026-11-28).** Manager 세션 scratch 에서 만든 read-only 설문 원문을 그대로 옮겼다. 사용자가 2026-11-28 「권장대로」로
> Q1~Q8 권장안을 쓰인 그대로 수락했다 — 결정 기록과 그 반영은 `ma-runbook.md` 「2026-11-28 사용자 결정」, 분기 인용의 AST 는
> 같은 디렉터리 `citation-ast/`(HEAD `9b61cc03d` 에서 다시 뽑아 설문의 `source_sha256` 7개와 일치 확인). 이 사본 안의 「scratch `ast/`」는 그 디렉터리를 뜻한다.

# a100 M-A 재개 — 사람 결정 설문 (2026-11-28, read-only 조사)

- 기준: main HEAD `baeddfe31`, prep 브랜치 `docs/a100-ma-prep` = `e01414cd2`(로컬 전용 — `git ls-remote origin` 에 없음, worktree `/home/daniel/tossos-work/a100-ma-prep`).
- 이번 조사에서 한 것: `git log/show/merge-tree`, `sha256sum`·`go version -m`(실행 안 함), `tools/logic-map` AST 7개(아래), 로컬 verify 기록·attestation 메타데이터·config 토글 값만 읽기.
- 안 한 것: 브로커 호출 0, `tossctl` 실행 0, 컨테이너·엔진 조작 0, 편집·커밋 0. `check_analysis.py` 재실행 안 함.
- AST(HEAD 기준, scratch `ast/`): `ReconcileDriver.judgeHoldings`(분기 16), `Runner.stepConditionalTrigger`(6), `Runner.createConditional`(11), `Service.ConditionalPlace`(2), `resolveVerifyRecordFor`(3), `NearStopTrigger`(5), `M0ExactPrerequisites`(3). 아래 분기 번호는 이 산출물 기준이다.

---

## 먼저 고를 것 — 요약

| # | 질문 | 권장 |
|---|---|---|
| Q1 | parent 조건주문은 누가 만드나 | M0 verify 가 만든다(CLI place·토글 flip 단계 삭제) |
| Q2 | attestation 은 M-A 필수 입력인가 | 참고 기록. 단 M-A 뒤 a100 바이너리로 재발급은 필수(0.10 d) |
| Q3 | 표본 종목을 엔진 편입에서 어떻게 떼나 | 매수 **전** exclude 저장 + 닫힌 창 엔진 재기동 |
| Q4 | 어느 바이너리로 재나 | 현재 설치본 `f73e0917b`(clean) + M0 경로 차분 4파일 좁은 재리뷰 |
| Q5 | 어느 verify 기록을 쓰나 | data-dir 기록을 `--record` 로 명시 + 사전 `order-cancel` 재측정 1회 |
| Q6 (신규) | US 발동 관측도 같은 세션에서 하나 | 분리 — KR 완전 통과 뒤 별도 승인 세션 |
| Q7 (신규) | 측정 창 동안 soak 을 세우나 | 세운다(창 10분 + 연결 90초 동안만) |
| Q8 (신규) | prep 브랜치를 어떻게 반영하나 | 답을 덧붙인 뒤 merge(충돌 0), 낡은 AST 3개 재동결 |

---

## Q1. parent 조건주문은 누가 만드나 — CLI `order conditional place` 인가, M0 `verify run` 인가

- **왜 중요한가**: M0 경로만 broker create 전에 pending intent 를 fsync 한다. CLI 로 만든 parent 는 crash 시 소유자 복구가 없고, M0 는 외부 parent 를 받을 입력이 없다. 그리고 두 경로를 둘 다 쓰면 같은 보유에 손절 2개가 붙는다.
- **증거(HEAD, 브랜치 판 이후 무변경)**:
  - `internal/verifylive/steps_trigger.go:91-166` — `stepConditionalTrigger` 가 L144 `r.createConditional` 로 **자기 parent 를 만든다**. 본문 고정: SINGLE·MARKET·SELL·수량 `MinQuantity`(=1, `pricing.go:42`)·만료 `r.expireDate()`(L124-132).
  - `internal/verifylive/mutate.go:503-556` — `createConditional`: B2@508 사람 확인 → B3@515 `pending-create` checkpoint → L532 broker create → B11@545 `parent-created`.
  - `internal/trading/conditional.go:123-131` — CLI 경로는 `ConditionalGate`(B1@124) → create 뿐, verify 기록·checkpoint 없음.
  - `cmd/tossctl/verify.go:185-215` — 외부 parent id 를 받는 flag 없음.
  - **새로 확인**: verify 경로는 `trading.conditional` 토글을 읽지 않는다. `buildVerifyBroker`(`cmd/tossctl/verify.go:875-905`)는 official client 를 바로 만들고 `trading.Service`·`ConditionalGate` 를 거치지 않는다(`rg ConditionalGate` → `internal/trading/conditional.go` 만). 즉 M0 를 쓰면 런북 §4·§8 의 토글 flip 은 필요 없다.
  - trigger 위치는 「현재가 1~2틱 아래」가 아니다. `NearStopTrigger`(`pricing.go:221-256`)는 **최우선 매수호가 바로 위 첫 틱, 그리고 마지막 체결가 이하**다. B3@231 `bid >= last` 이면(마지막 체결이 매수호가에서 났을 때) 오류 → 단계 skip.
- **선택지**
  1. **M0 가 만든다.** 런북 §5(CLI place)와 §4/§8(토글 flip·원복)을 뺀다. `verify run` 한 번으로 등록·관측·정리. tasks 0.2a.6, design D10-A 와 일치. 코드 변경 0.
  2. CLI 로 만든 parent 를 M0 가 관측. M0 에 새 입력이 필요하다(코드 변경 + 0.2a 재리뷰). pending intent 가 없어 crash 복구 계약이 깨진다.
- **권장: 1.** 코드·tasks·design 셋이 이미 그렇게 돼 있고, 운영 토글 flip 단계 하나(사람 승인 대상)가 통째로 사라진다. 2 는 새 구현이다.

## Q2. soak/attestation 은 M-A 실행의 필수 입력인가, 참고 기록인가

- **왜 중요한가**: 필수로 두면 M-A 가 열리지 않는다(아래 이유). 참고로 두면 M-A 는 열리지만, attestation 에 조건주문 mutation 이 없는 상태는 그대로다 — 그것을 고치는 것이 M-A 뒤의 0.10 (d) 다.
- **증거**
  - `measurement-prereq.md:68-72` (6) 필수 vs `ma-runbook.md`(6dbb752f6 판) R0-6 「참고용」 — 모순은 그대로다.
  - verify 경로는 attestation 을 읽지 않는다. HEAD 의 `attest` 언급은 `attest.Mask`(계좌 마스크, `confirm.go:122`, `steps.go:1353`)와 주석뿐이다.
  - **현재 host 상태(새 실측)**: `~/.config/tossctl/capability-attestation.json` — issued 2026-11-27 15:00Z, expires 2026-12-27, `verified_by=tossos-attest.timer`, renewal-status `outcome=issued`(매일 자동 갱신). endpoint 11개 안에 `POST /conditional-orders`·`/modify`·`DELETE` 는 **없다**. supervised 는 `POST /orders`·`/cancel` 둘뿐.
  - tasks 0.10 (`tasks.md:203-205`, `:220-222`): 순서는 「verify DELETE·modify 증거 → a100 바이너리로 `soak attest` → 배포」. 조건주문 mutation 이 attestation 에 실리는 것은 M-A **뒤**다.
- **선택지**
  1. 필수 — attestation 이 측정 바이너리에 결속되고 조건주문 mutation 을 담을 때까지 M-A 를 막는다. 그런데 그것은 M-A 의 산출물이다. 순환이라 영원히 안 열린다.
  2. **참고** — 상태(유효기간·endpoint 목록)만 R0 에 기록한다. `measurement-prereq.md` (6)을 「배포 전 필수, M-A 전 참고」로 정정한다.
- **권장: 2.** 1 은 순환이다. 안전 쪽 손실도 없다: 엔진은 attestation 에 없는 조건주문 mutation 을 자동 경로에서 쓰지 않고, M-A 는 사람이 직접 서는 일회 측정이다.

## Q3. 새 표본 종목을 엔진 편입에서 어떻게 떼어 두나

- **왜 중요한가**: 엔진이 표본을 편입하면 엔진 손절과 M-A 조건주문이 같은 주식을 두고 다툰다. 2026-10-07 에 실제로 엔진이 먼저 팔아 표본이 사라졌다(B3). 거꾸로도 위험하다. M-A 조건주문은 매도가능수량 1주를 예약한다(`steps_trigger.go:102-103` 주석, 측정값 `conditional.reserves_sellable_quantity`). 그래서 편입된 포지션이면 **엔진 자신의 손절 매도가 수량 부족으로 막힐 수 있다**(안전 불변식 4).
- **증거**
  - `internal/app/engine/adoption.go:107-197` `judgeHoldings` — B7@141 `p.ExitEligible()` 이면 `continue`. 이 분기가 B12@166 `Adoption.Excludes` 보다 **앞**에 있다. 그래서 이미 편입된 포지션은 나중에 exclude 에 넣어도 풀리지 않는다. B13@174 `!Enabled && !Included` 는 `enabled=true` 면 항상 거짓이다.
  - **새로 확인**: exclude 변경은 **다음 엔진 기동에만 반영된다.** `internal/app/engine/reconcileloop.go:383` `opts.Adoption = c.Config.Engine.Adoption`(드라이버를 만들 때 값을 복사), 콘솔 문구 `internal/console/templates_settings.go:103` 「실제 편입은 다음 엔진 기동의 엔진 대사 루프가 수행」.
  - 현재 config(값만 읽음): `adoption.enabled=true`, `exclude_symbols` 0개, `include_symbols` 4개.
  - 엔진 재기동은 그동안 손절이 없다는 뜻이다(엔진 정지 = 손절 없음). 그래서 닫힌 창에서만 한다. `docs/operations.md:345` 는 KST 05:00~09:00 이라 적지만, KR NXT 가 08:00 에 열리고 US 정규장이 겨울에 06:00 KST 에 닫히므로 **실효 창은 06:00~08:00 KST**다.
- **선택지**
  1. **매수 전에** 표본 종목을 `exclude_symbols` 에 저장(콘솔, audit 기록 `cmd/tossctl/adoptionsettings.go:62-92`) → 06:00~08:00 KST 에 엔진 재기동(사람 승인) → 엔진 로그나 콘솔에서 새 값이 적용됐는지 확인 → 그 뒤 장중에 매수. 영향은 그 1종목뿐.
  2. 측정 동안 `adoption.enabled=false`(재기동 2회 필요). 미편입 보유 전부가 편입을 못 받는다. 범위가 넓다.
  3. 편입을 허용한다. 엔진 손절과 경쟁하고, 예약 때문에 엔진 손절이 막힐 수 있다. 금지 권장.
- **권장: 1.** 영향 범위가 가장 좁고, B7 순서 때문에 「매수 전 + 재기동」 두 조건이 모두 필요하다. 재기동은 운영 작업이라 별도 승인이 필요하다.

## Q4. 어느 바이너리로 재나

- **왜 중요한가**: M0 진입 판정(`M0ExactPrerequisites`)과 receipt 형식은 바이너리가 정한다. 신원이 결속되지 않은 바이너리의 측정은 증거가 되지 않는다.
- **증거(새 실측 — B1 상태가 바뀌었다)**
  - 설치본 `/home/daniel/.local/bin/tossctl`: SHA-256 `ed904db0e646…b398a`, mtime 2026-11-25 18:46, `vcs.revision=f73e0917b`, **`vcs.modified=false`**, go1.26.5. 11-10 의 dirty `5287fe3c…`(488b370f) 는 더 이상 설치본이 아니다.
  - `git merge-base --is-ancestor 6f06eee8e f73e0917b` = 참. a144 TTL-edge 면제가 들어 있다(`m0_recovery.go:98-108`, B1@99 range·B2@100 trigger/TTL-edge continue·B3@103 미PASS 거부).
  - `f73e0917b..main` 13커밋 중 `internal/verifylive`·`internal/official`·`cmd/tossctl/verify.go`·`internal/trading/conditional.go` 변경 0. 설치본 = HEAD 와 같은 M0 경로다.
  - A-M0 가 수용한 것은 M0 merge `016da6245`(2026-08-15) 판이다. `016da6245..f73e0917b` 에서 M0 비시험 소스 차분: `steps_trigger.go`·`mutate.go` **0줄**, `record.go` +183(a131 봉인·flock), `official/client.go` +15, `cmd/tossctl/verify.go` +24, `m0_recovery.go` +9(TTL-edge). 나머지 추가는 reconcile·읽기 전용 파일이다.
- **선택지**
  1. **현재 설치본 `f73e0917b` 그대로.** clean·소스 결속·TTL-edge 포함. 위 4파일 차분만 다른 계열 리뷰어가 M0 계약(0.2a.3~0.2a.8)에 비춰 좁게 재리뷰한다. 새 설치 없음.
  2. HEAD 로 새 후보를 빌드해 8/15 처럼 원자 설치한다. M0 경로가 설치본과 같아 얻는 것이 없고, 설치 승인 1건만 는다.
  3. `882a0b49` 판을 다시 만든다. TTL-edge 면제가 없어 M0 가 거부한다(TTL-edge 는 두 기록 모두 `skipped`).
- **권장: 1.** 측정 직전 R0 에서 `sha256sum` 과 `go version -m` 으로 위 값을 다시 고정한다. 설치본이 그사이 바뀌면 이 답은 무효다.

## Q5. `--config-dir` 를 붙인 verify 는 어느 기록을 쓰나 — 그리고 그 기록이 M0 를 열 수 있나

- **왜 중요한가**: `--config-dir` 를 빼면 기본 토큰 캐시를 써서 엔진 토큰을 무효화한다(2026-10-19 사고, B7). 붙이면 기록이 config 쪽으로 바뀐다(`resolveVerifyRecordFor` B2@770). M0 는 **고른 기록의 직전 판정**으로 진입을 결정한다.
- **증거(새 실측 — 기록 step 별 최신 판정, 필드만 읽음)**
  - `M0ExactPrerequisites` 는 trigger·TTL-edge 를 뺀 **모든 step 의 마지막 판정이 PASS** 여야 한다(`Passed` = `LastEntry` 판정, `record.go:659-662`). 거부는 broker 를 만들기 전에 로컬에서 난다(`cmd/tossctl/verify.go:410-431`).
  - data-dir `~/.local/share/tossos/capability-verify.jsonl`(63행, mtime 10-19): `order-cancel`=**fail**(2026-10-19, `insufficient-buying-power`). 나머지는 PASS. artifact 는 전부 cancelled.
  - config-dir `~/.config/tossctl/capability-verify.jsonl`(35행, mtime 11-24 — 10-20 때 27행에서 늘었다): `sell-boundary`=**fail**(07-31, NXT 미지원), `conditional-modify`=**fail**·`conditional-cancel`=**fail**(08-13, 404 not-found). `order-cancel` 은 11-23 PASS. 조건주문 artifact 1건은 10-23 `reconcile-expired` 행으로 닫혔다(a164).
  - ⇒ **지금은 어느 기록으로도 M0 가 열리지 않는다**(신규 blocker B14).
  - 콘솔 `/verify` 는 root 의 `--config-dir` 로 기록을 고른다(`cmd/tossctl/console.go:224`). 운영 콘솔이 config 쪽 기록을 쓰는지는 미확인이다.
  - 줄마다 봉인값이 달라 두 기록이 같은 계좌에 결속됐는지는 정적으로 대조할 수 없다(미확인 — 실행 시 runner 가 판정).
- **선택지**
  1. `--config-dir ~/.config/tossctl` 만 붙인다 → config 기록. M0 전에 `sell-boundary`(먼 지정가 매도)와 conditional register→modify→cancel 사슬을 다시 재야 한다. live mutation 이 여럿이고, 정규장(NXT 아님)이어야 한다.
  2. **`--config-dir ~/.config/tossctl --record ~/.local/share/tossos/capability-verify.jsonl`** → data 기록. M0 전에 `order-cancel` 하나만 다시 재면 된다(먼 지정가 매수 1건 → 즉시 취소, 체결 의도 없음). 1주 매수가능금액이 필요하다.
- **권장: 2.** 사전 mutation 이 1건(체결 안 됨)으로 가장 적다. 토큰은 config 쪽을 써서 B7 을 피한다. 남는 것: (a) `order-cancel` 재측정도 실주문이라 별도 사람 승인이 필요하고, (b) 매수가능금액이 부족하면 10-19 처럼 다시 fail 한다. 측정 직전 `tossctl verify status --record <그 경로>`(로컬 전용, 네트워크 0)로 판정·outstanding 을 사람이 재확인한다.

## Q6 (신규). US 발동 관측(verify-observes-the-trigger 3.x)을 같은 세션에 넣나

- **왜 중요한가**: US 는 장 시간(겨울 KST 23:30~06:00)과 기록 파일이 다르다. 그리고 v-o-t 3.3 이 지정한 US 종목은 엔진 **include 목록 안**에 있어, 이미 편입됐다면 exclude 로도 풀 수 없다(Q3 의 B7 순서).
- **증거**: US 기록은 data-dir `capability-verify-us.jsonl` 에만 있다(34행, trigger=deferred·TTL-edge=skipped·나머지 PASS → M0 전제 충족). `~/.config/tossctl/capability-verify-us.jsonl` 은 **없다.** 그래서 `--config-dir` 만 붙이면 빈 기록이 되고 M0 가 거부한다 → `--record` 명시가 필수다. 10-07 기록상 US 보유는 1주 미만이었다 → `sellable < MinQuantity` 이면 skip(`steps_trigger.go:99-108`).
- **선택지**
  1. **KR 완전 통과 뒤 별도 승인 세션.** US 표본은 include 목록 **밖**의 종목으로, Q3 의 1 과 같은 절차로 매수 전에 떼어 둔다.
  2. 같은 날 연속(KR 장중 → 그날 밤 US). 피로가 쌓이고 재기동 창이 겹친다. KR 결과가 4갈래 중 무엇인지 보기 전에 US 를 열게 된다.
  3. US 는 a100 뒤로 미룬다. verify-execution-capability 2.6 의 「미검증 시장 자동 진입 금지」로 US 는 fail-closed 상태로 남는다.
- **권장: 1.** KR 이 부분 통과면 design D10 에 따라 어차피 a100 이 멈추므로 US 를 같이 돌릴 이유가 없다. 3 도 안전하지만, a100 이 KR·US 둘 다 배선(`protection_wiring.go:41-42` 둘 다 `Wired:false`)이라 US 증거가 결국 필요하다.

## Q7 (신규). 측정 창 동안 soak 을 세우나

- **왜 중요한가**: critical window 안의 첫 401/429 는 되돌릴 수 없는 HOLD 다(tasks 0.2a.7). 그러면 1주를 팔고도 증거를 잃는다. 대기 국면은 5초 간격 × 10분 ≈ 240요청이다(`steps_trigger.go:58-88`). soak 는 15분마다 돌고, 직전 cycle 이 11-27 18:34Z 로 살아 있다. account-budget(a166)은 cutover 만 했고 활성화하지 않았다(OFF).
- **선택지**
  1. **세운다** — 측정 시작 직전 정지, 끝나면 재기동(콘솔, 사람). attestation 갱신은 매일 1회라 몇 시간 정지는 문제없다(`MaxRecordAge` 48h, `tasks.md:194-195`).
  2. 그대로 둔다 — 429 가 나면 측정만 낭비된다(안전 손실은 없음).
- **권장: 1** (v-o-t 3.2 와 같다). 엔진은 세우지 않는다 — 다른 보유의 손절이 엔진에 있다.

## Q8 (신규). prep 브랜치를 어떻게 반영하나

- 증거: `git merge-tree --write-tree main docs/a100-ma-prep` → 트리 `23a483a9b`, **충돌 0**(`tasks.md` 자동 병합 — main 쪽 a191 커밋 `5ddd0fbeb` 와 겹침). 브랜치는 origin 에 없고 로컬에만 있다. 재동결 AST 11개 중 3개는 원본이 main 에서 또 바뀌었다: `internal/app/engine/interlock.go`(3커밋), `cmd/tossctl/engine.go`(3), `internal/journal/exit_snapshot.go`(1).
- 선택지: 1) 이 설문 답을 런북 「Manager 결정 필요」 아래에 덧붙인 뒤 merge하고, 낡은 3개를 재동결한다. 2) 브랜치를 버리고 main 에 새로 쓴다.
- **권장: 1.** 충돌이 없고, 3커밋의 원장·정정이 그대로 유효하다(B1 은 아래 표로 갱신).

---

## Blocker 표 (2026-11-28 재대조)

| B# | 내용 | 아직 참? | 증거 |
|---|---|---|---|
| B1 | 측정 바이너리 신원 불일치(dirty 488b370f) | **해소(성격 바뀜)** | 설치본이 `f73e0917b` clean 으로 바뀌었다(SHA `ed904db0…`, 11-25). 남은 것은 Q4 의 좁은 재리뷰 결정 |
| B2 | TTL-edge 면제가 바이너리에 있나 / 기록의 TTL-edge 판정 | **해소** | `6f06eee8e` ⊂ `f73e0917b`. 면제가 있으니 두 KR 기록의 TTL-edge `skipped` 는 M0 를 막지 않는다 |
| B3 | 2026-10-07 표본 소진 | 참 | 새 표본 조달 필요(현재 보유는 브로커 조회라 안 봄) |
| B4 | include 가 범위를 좁힌다는 0.9 가정이 틀림 | 참 | config `enabled=true`, exclude 0, include 4. `judgeHoldings` B12@166·B13@174. **추가: exclude 는 다음 엔진 기동에만 반영**(`reconcileloop.go:383`) |
| B5 | parent 생성 주체 모순 | 참(결정 대기 Q1) | AST `stepConditionalTrigger` L144, `createConditional` B3@515, `ConditionalPlace` B1@124 |
| B6 | attestation 필수/참고 모순 | 참(결정 대기 Q2) | 문서 두 문장 그대로. attestation 자체는 유효(~12-27)지만 조건주문 mutation 이 없다 |
| B7 | 런북 명령의 `--config-dir` 누락 | 브랜치에서 해소, **main 에선 참** | main 의 `ma-runbook.md` 는 브랜치 이전 판. merge 하면 해소 |
| B8 | `--config-dir` 가 기록을 바꾼다 | **실측 완료 → B14 로 승격** | Q5 증거 |
| B9 | base 표류·FLM stale | 참(악화) | `882a0b49..HEAD` 1,712커밋(11-10: 1,378). `check_analysis` 재실행 안 함 |
| B10 | soak/attestation 현재 상태 | **해소(확인됨)** | soak cycle 11-27 18:34Z ok, attestation issued 11-27 → expires 12-27, 자동 갱신 `issued` |
| B11 | 잔여 조건주문 | 미확인 | soak 의 by-id probe 가 ok → 목록(OPEN 또는 CLOSED)에 1건 이상. OPEN 잔여 여부는 브로커 조회가 필요(R0 에서 사람 입회로) |
| B12 | a087 동승 | 해소 | a087 5.1~5.3 `[x]` |
| B13 | 세션 날짜·expire | 참 | 미정. expire 는 코드가 KST 오늘+7일로 정한다(`steps.go:1253-1256`) |
| **B14 (신규)** | 어느 KR 기록도 `M0ExactPrerequisites` 를 통과 못 함 | 참 | data: order-cancel fail(10-19). config: sell-boundary·conditional-modify·conditional-cancel fail. `m0_recovery.go:98-108`, `record.go:659` |
| **B15 (신규)** | exclude 반영에 엔진 재기동 필요 | 참 | Q3. 재기동은 06:00~08:00 KST 창 + 사람 승인 |
| **B16 (신규)** | US 기록이 config-dir 에 없음 | 참 | `~/.config/tossctl/capability-verify-us.jsonl` 없음 → US M0 는 `--record` 필수(Q6) |

## prep 브랜치 merge-tree 결과

```
git merge-tree --write-tree --name-only main docs/a100-ma-prep
→ 23a483a9b81cb26a8812362273f7a87481da7abc   (rc 0, 충돌 0)
  Auto-merging openspec/changes/a100-wire-fill-to-broker-protection/tasks.md
```

3커밋: `e253b558e` 원장 · `c03928b63` 런북 정정·Q1~Q5 · `e01414cd2` AST 재동결(11개, 34파일 +7537/-509). merge 하지 않았다.

---

## M-A 실계좌 측정 — 사람이 승인하는 것 (권장안 Q1·Q3·Q4·Q5·Q7 기준)

각 ⌨ 줄은 실행 직전 그 줄에 대해 따로 승인한다(§0-1·§0-7). 이 문서에 동의하는 것은 그 승인이 아니다. 에이전트는 mutating 명령을 하나도 실행하지 않는다.

### 사전 조건 (전부 참이어야 시작)

1. 설치본 신원: `sha256sum ~/.local/bin/tossctl` = `ed904db0e64644b341ba4af0c756479838495c40e735b4f0f31f513ff18b398a`, `go version -m` → `vcs.revision=f73e0917b…`, `vcs.modified=false`. Q4 재리뷰가 P0=0/P1=0.
2. 표본: KR 유동성 종목 1개. 엔진 관리 종목이 아니고 OPEN 잔여 조건주문과 겹치지 않는다. 매도가능 **최소 1주**(최소 단위). 이미 엔진에 편입된 포지션이면 안 된다.
3. 그 종목이 `exclude_symbols` 에 저장돼 있고, 저장 **뒤** 엔진이 재기동해 그 값을 읽었다(Q3). 매수는 그다음이다.
4. 기록: data-dir 기록의 `order-cancel` 이 PASS 로 바뀌었다(Q5 사전 측정). `verify status --record …` 의 outstanding 이 0 이고, M0 checkpoint 도 미해결 0이다.
5. 매수가능금액: 사전 `order-cancel` 탐침 1건(먼 지정가 매수 1주)을 낼 만큼 있다.
6. soak 정지(Q7). 엔진은 가동 상태를 유지한다.
7. `--trigger-receipt-dir`: 현재 uid 소유의 0700 빈 디렉터리. 심링크면 안 된다.
8. 사람이 콘솔과 터미널(TTY) 앞에 있다 — `--confirm-each` 의 typed confirmation 은 TTY 전용이다.

### 시간

- 엔진 재기동(조건 3): **06:00~08:00 KST**(두 시장이 닫힌 창). 측정 전날 또는 당일 아침.
- 측정: **KRX 정규장 09:00~15:30 KST.** NXT 시간대에는 하지 않는다(07-31 `sell-boundary` 가 NXT 미지원으로 fail). 개장 직후와 동시호가는 피하고, 호가가 `bid < last` 로 1틱 이상 벌어진 순간에 시작한다(아니면 `NearStopTrigger` B3 로 skip).
- 한 번의 소요: 대기 최대 10분(`DefaultTriggerWindow`) + 발동 뒤 연결 관측 90초(`TriggerLinkWindow`).

### 단계

1. 👁 R0 read-only 고정: 바이너리 신원, `verify status --record <data 기록>`, 잔여 조건주문 목록(사람 입회 조회), 표본 매도가능수량, 호가.
2. ⌨ (사전, 별도 승인) `tossctl --config-dir ~/.config/tossctl verify run --market KR --resume --redo order-cancel --record ~/.local/share/tossos/capability-verify.jsonl` — 먼 지정가 매수 1건 → 즉시 취소. 체결 의도는 없다. (이 redo 조합이 일반 경로에서 허용되는지는 아직 검증 안 함 — 실행 전에 `--list` 로 확인.)
3. ⌨ (본측정, 별도 승인) `tossctl --config-dir ~/.config/tossctl verify run --market KR --holding-symbol <표본> --include-trigger --confirm-each --resume --redo conditional-trigger --trigger-receipt-dir <0700 디렉터리> --record ~/.local/share/tossos/capability-verify.jsonl`
   - 도구가 하는 일: 매도가능수량·호가 읽기 → trigger = 매수호가 바로 위 첫 틱(≤ 마지막 체결가) → typed confirmation → pending intent fsync → **SINGLE·MARKET·SELL·1주·만료 오늘+7일 조건주문 1건 등록** → 5초/1초 폴링 → 발동이면 child 주문의 체결을 official by-id 로 확인.
   - `trading.conditional` 토글은 건드리지 않는다(Q1 증거).
4. 👁 기록·4갈래 판정(통과/부분/4b 실패/실패). 완전 통과만 0.2 체크 + 0.11 raw-status 표 동결.
5. ⌨ 정리: soak 재기동. exclude 해제는 별도 결정(해제하면 다음 기동부터 엔진이 표본을 편입한다).

### 취소 경로

- 10분 안에 trigger 에 닿지 않으면: 도구가 **자기가 등록한 조건주문을 같은 단계에서 취소**하고 `skipped(INCONCLUSIVE)` 로 끝낸다. 계좌 잔여물은 0. 취소 직후 경합은 보유 수량으로 판정하고, 그것도 못 읽으면 fail.
- 중간에 끊기면(Ctrl-C·프로세스 사망): 조건주문은 「발동하도록 값 매긴 살아 있는 손절」로 남는다. `tossctl verify abort` 가 기록상 outstanding 만 나열하고 취소한다. 안 되면 토스 앱에서 직접 취소한다. 시간이 지난다고 자동 취소되지 않는다(v-o-t 1.20).
- 발동했는데 child 를 90초 안에 못 보면: `fail`, child 는 붙잡힌 채 보고된다. M0 는 HOLD 이고 자동 취소·재등록하지 않는다(0.2a.6·0.2a.7). 사람이 대사한다.

### 최악의 경우

1. **발동 = 1주가 시장가로 팔린다. 되돌릴 수 없다**(이것이 측정 목적이다). 손실 상한은 1주 가격 × 슬리피지다. 유동성 종목이면 1~수 틱 수준이다.
2. 프로세스가 죽고 아무도 abort 하지 않으면: 그 손절이 **최대 7일**(만료) 살아 있다가 언제든 1주를 시장가로 판다. 그동안 그 1주의 매도가능수량이 예약된다.
3. 조건 3을 건너뛰어 엔진이 표본을 편입하면: 엔진 손절과 경합하거나, 예약 때문에 **엔진 손절 매도가 수량 부족으로 막힐 수 있다.** 그래서 조건 3이 필수다.
4. 429/401 이 critical window 안에서 나면: 영구 HOLD. 1주는 팔리고 증거는 무효다(안전 손실은 없음).
5. 브로커 매도는 보유가 있어야 나가므로 공매도나 수량 초과는 불가능하다. 수량은 코드에 1주로 고정돼 있다(`pricing.go:42`, 시험 `pricing_test.go:165`).

---

## 남은 미확인 (검증하지 않은 것)

- 현재 보유·OPEN 잔여 조건주문(브로커 조회 필요 — 안 함).
- data/config 두 KR 기록이 같은 계좌에 결속됐는지(봉인값이 줄마다 달라 정적 대조 불가).
- 운영 콘솔 `/verify` 가 config 기록을 쓰는지(컨테이너 config-dir 미확인).
- `--resume --redo order-cancel` 이 일반 경로에서 허용되는지(코드 미추적).
- `check_analysis.py` 의 현재 stale 수(재실행 안 함).
- 본문에 적은 「Manager 결정이 왜 필요한가」의 Manager 쪽 근거(M-A 가 v-o-t 3.x 와 v-e-c 2.5/2.6 을 흡수한다는 것)는 저장소 문서에서 찾지 못했다 — 위임 지시로만 받았다.
