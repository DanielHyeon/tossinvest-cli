# a124 리뷰 기록

## §0 proposal-freeze 1판 (2026-09-26) — 판정 **REJECT** (design 개정 후 재freeze)

- 대상: proposal · design · tasks · `specs/engine-safety/spec.md` (개설 커밋 `2fbdcd78`), base `4798d399`.
  AST 번들은 `463cc895` 추출이고 두 커밋 사이 Go 소스 diff 는 0 이다(`git diff --stat 463cc895 4798d399 -- '*.go'` 빈 출력).
- 위험 등급: **High-risk** (진입 게이트 · 운영 모드 · 원장 outbox).
- 보이스
  1. **적대적 Eng** — Teammate(Opus, 작성자 Manager Fable 과 분리된 컨텍스트). HEAD 파일을 직접 읽고 AST 번들의
     열거만 분기 근거로 썼다.
  2. **교차 모델** — codex-cli 0.154.0, `codex exec -s read-only --ephemeral`, model `gpt-6-astra`, reasoning medium.
     프롬프트는 Teammate 의 발견을 주지 않은 독립 지시다. 원문: `analysis/freeze-review/codex-r1-prompt.md`,
     출력: `analysis/freeze-review/codex-r1-output.md` (실행 전후 `git status` 에서 이 change 밖 변경 0 — codex 는 쓰지 않았다).
  3. `autoplan` 대화형 4관점은 자율 세션에서 돌릴 수 없어 위 두 독립 패스로 대체했다(a114 §0 선례). CEO/보안/QA 관점은 아래 셀프 요약.
- 셀프 요약: **범위** — 사용자 결정 20-1 ⓒ 그대로(래치·승격·굶주림). **보안** — 새 게이트 detail·로그에 행 본문·계좌·토큰이
  들어갈 자리가 생긴다(F8). **QA** — RED 2.1~2.6 은 F1·F2·F3 의 경합·실패 경로를 아직 담지 않는다.

### 손절 즉시성 (규칙 7)

- `go test ./internal/app/engine/ -run 'TestTheExitCycleDoesNotLengthenWhileTheSenderIsStuckInTheTransport|TestAStopAlertDoesNotWaitBehindTheBacklog' -count=1 -v`
  → **rc 0** (exit 사이클 체류 5.91 ms 기준 / 5.54 ms 발송자 갇힘 · 정지 알림 체류 5.20 ms(10행) / 5.38 ms(1000행)).
  이 수치는 **base 의 기준선**이며 a124 편집 뒤의 증거가 아니다.
- 구조 근거: 진입 게이트는 노출을 늘리는 변이에만 묻는다(`execgw/gateway.go:855-859`), 트리거는 `ENTRY_BLOCKED` 로만 간다
  (`journal/operating_mode.go:537-545`). 단, 원장 연결이 하나(`journal/journal.go:174`)라서 실행자가 더하는 쓰기는 exit 루프와
  같은 연결을 기다린다 — 「다른 goroutine 이니 무관」은 증거가 아니다(F12).

### 발견 표

출처: T = Teammate 적대 Eng, C = codex. 심각도는 두 보이스 중 높은 쪽을 적되 다르면 병기한다.

| id | 출처 | 심각도 | 발견 | 증거 | 판정 → 요구 |
|---|---|---|---|---|---|
| F1 | T·C | **P0** | D1 의 판정 입력 「`MarkAlertAttemptFailed` 가 돌려주는 갱신된 행의 attempts」가 **API 에 없다**. `SettleResult` 는 `Outcome·ClaimedBy·ClaimedAt·ExpiresAt` 뿐이고, 구현이 실제로 쓸 수 있는 `alert.Attempts` 는 **나열 시점 스냅숏**이다 — 배치 뒤쪽 행은 나열 뒤 최대 ~10×10 s 에 임차되고 그 사이 다른 발송자가 올리거나 재무장이 0 으로 되돌린다 | `alert_claim.go:140-145` · `outbox.go:467-474` · `alertdelivery.go:150·168·222` · 재무장 `outbox.go:337` | 수용. design D1 은 **같은 트랜잭션에서 커밋된 증가 후 값**(토큰·PENDING CAS 아래)을 판정 입력으로 명시하고, 그 journal API(예: `SettleResult` 에 attempts, `RETURNING attempts`)를 Impact·tasks 3.x 에 넣는다. RED 에 「나열 뒤 값이 바뀐 행」 추가. modernc sqlite v1.54.0 의 `RETURNING` 지원은 1.2 에서 실측 |
| F2 | T·C | **P0** | 승인과 래치·승격이 원자적이지 않다. (a) 발송 중 승인 → 실패 기록이 `SettleAlreadySettled` — 여기서 잠그면 안 된다. (b) 실패 기록 `SettleApplied` 직후 운영자가 승인·게이트 해제(·모드 완화)하고 그 뒤 실행자가 `Block`·`Escalate` → **빈 backlog 에 잠긴 게이트와 되살아난 모드**. 실행자는 뮤텍스를 안 잡고(`alertdelivery.go:19-20`), `Acknowledge` 는 `n.mu` 로 동기 경로만 배제한다 | `notifier.go:851-876` · `alert_claim.go:330-334` · `operating_mode.go:410-421` | 수용. 이것은 새 취향이 아니라 **이미 사람이 정한 규칙**이다: a092 델타 「운영자의 승인은 … 전송의 성공 여부는 그 사유를 되살리지 않는다(SHALL NOT)」(사용자 결정 2026-08-10 결정 2 = 안 1, `a092…/specs/engine-safety/spec.md:66-68`). a124 가 래치의 주인이 되므로 그 SHALL 도 a124 가 진다 → spec delta 에 시나리오로, design 에 **배제 수단**으로. 수단 선택은 아래 「Manager 결정 M1」 |
| F3 | T·C | P0(C) / P1(T) | FLM `deliverOne` B10(실패 기록 자체 오류)에 정책이 없다. 기록이 계속 실패하면 attempts 가 안 올라 한도 판정이 영원히 안 선다 — a092 뒤 동기 경로가 없으면 진입이 계속 열린다. 오늘도 같은 분기지만 오늘은 동기 경로가 가린다 | FLM deliverOne B10 · `alertdelivery.go:221-226` · 동기 대응 B13·B15 는 로그 후 계속, 소진 시 잠금(`notifier.go:494-497·565-571`) | 수용. design 에 B10 정책을 적는다. 「Manager 결정 M2」 |
| F4 | T·C | P1 | D6 의 최악 래치 식이 **틀렸다**. 행 하나의 재시도 사이에 같은 배치의 나머지 행 서비스가 끼는 것을 빠뜨렸다. 반례: 10 행 전부 10 s timeout → 사이클 = 10×10+2 = 102 s, 첫 행의 셋째 실패 = 2×102+10 = **214 s** (D6 식 = 100+3×2+3×10 = 136 s) | `alertdelivery.go:122-136·154-162` · `obs/alert_lease.go:17` · `obs/ntfy.go:95-100` | 수용. 식을 `(L−1)×(B×T_pub + I) + T_pub` (배치 포화) 와 큐 대기 항으로 다시 쓰고, 원장·임차·반납 대기는 **전제 조건**으로 밝힌다. 무한한 backlog·원장 정지에는 유한 상한이 없다고 적는다(a092 델타 「보장이 아니다」와 같은 결). a092 델타 식의 「사이클 주기」는 **배치 서비스 시간을 포함한 주기**라고 a092 21판에 전한다 |
| F5 | C (T 부분 동의) | P1(C) / P2(T) | R2 가 막는 굶주림은 한 형태뿐이다: ① 한도 아래 `HeldElsewhere` 행 10 개가 선택을 채우면 시도 없이 한 사이클이 빈다 — **임차 만료(81 s)까지로 유계**(`DefaultAlertLease`, 만료 뒤 탈취) ② 한도 층 안에서 오래된 10 행이 영구 실패(독 행)하면 그 뒤 한도 행은 재시도되지 않는다 — 게이트는 이미 잠겨 있으므로 안전 결과는 같고 전달만 늦다 ③ 한도 아래 행이 계속 들어오면 한도 행은 잔여 자리를 못 얻는다 — 설계 의도 | `outbox.go:518-521` · `alertdelivery.go:178-192` · `alert_claim.go:13-34·162-165` | 부분 수용(P2). spec 문구는 유지하되 design D4 에 세 경우와 유계·비유계를 적는다. 임차 가능 여부를 정렬에 넣을지(`claim_token = ''` 먼저)는 선택 — YAGNI 로 기록만 |
| F6 | C | P1(C) / P2(T) | 배포 첫 사이클: 이미 `attempts ≥ 3` 인 PENDING 행은 즉시 한도 층으로 내려가고, 다음 실패에서 durable 승격한다. 기동 복원은 메모리 래치만 복원한다 | `gateway.go:153-167·269` | 수용(P2). design Risks 에 「배포 직후 기대 동작」 — 게이트는 기동 복원으로 이미 잠김, 첫 실패 사이클에서 `ENTRY_BLOCKED` 기록 1 행 — 을 적는다. 운영 원장 조회는 이 로트에서 하지 않았다(사람 몫) |
| F7 | T·C | P1(C) / P2(T) | design 「안전 불변식 대조」의 *"알림 게이트 OFF 면 Notifier 자체가 없다"* 는 **틀렸다**. Notifier 는 엔진에서 무조건 생성된다. 실제 OFF 경계는 `engine run` 이 gate OFF 에서 기동을 거부하는 것이다 | `gateway.go:323` · `engine.go:528-530` · `TestAGateOffEngineRefusesWithoutEnumeratingClauses` **rc 0** (2026-09-26) | 수용. 문장을 교정하고 근거를 이 시험으로 바꾼다 |
| F8 | C | P1 | D5 가 게이트 detail · 로그 · 승격 실패 처리를 정하지 않았다. 동기 경로의 detail 형식을 베끼면 원문 오류가 들어가고(`notifier.go:563-571`), 모드 전이 오류는 계좌 참조를 담을 수 있다(`operating_mode.go:394·449`) | `auxiliary.go:206-211`(고정 문구 선례) · `gateway.go:161-166`(수만 쓰는 선례) | 수용. detail 은 고정 문구 또는 수만(불변식 8), 승격 실패는 error 로그(행 내용·토큰 없이)이고 메모리 래치는 남는다(`Notifier.escalate` 계약과 같음) |
| F9 | T·C | P3 | D1 「`Block` 재호출은 같은 사유의 갱신」 — 실제는 **없을 때만 삽입**, detail 은 처음 것이 남는다 | `execgw/retry.go:526-533` | 수용. 문구 교정 |
| F10 | T·C | P2 | `SettleOutcome` 의 영값이 `SettleApplied` 다. 오류 반환의 `SettleResult{}` 가 「적용됨」으로 읽힌다 — 판정은 `err` 를 먼저 봐야 한다 | `alert_claim.go:107-109·322-324` | 수용. Pre-Edit 불변식 + RED(기록 오류 시 잠금 판정이 「적용됨」으로 새지 않는다) |
| F11 | T·C | P2 | proposal 열린 결정의 *"같은 3 이면 약 6 s … 동기 경로는 약 34 s"* 는 **서로 다른 실패 양상을 비교**했다. 같은 양상에서는 둘이 같다(아래 D2 표) | 아래 D2 | 수용. proposal 문장 교정 |
| F12 | T·C | P2 | 원장 연결이 하나라 D3(publisher 없음도 기록)은 사이클마다 행당 쓰기 1 회를 더한다 — 10 행이면 2 s 마다 쓰기 10 회(정산 1 회 5.584 ms 실측값 기준 ≈ 56 ms). 보존 시험은 「transport 에 갇힘」만 재고 「publisher 없음」 은 안 잰다 | `journal.go:174` · `alertdelivery.go:67-69` | 수용. tasks 2.6 에 publisher 없음 변형의 exit 체류 측정을 더한다 |
| F13 | T·C | P2 | D3 을 받아도 **지연은 같지 않다**: 동기 경로는 publisher 가 없으면 첫 반복에서 `break` 후 곧장 잠근다(≈0 s), 실행자는 한도 3 이면 ≈4~6 s | `notifier.go:429-431·565-571` · `alertdelivery.go:205-212` | 기록. 결과(잠금·승격)는 같고 지연만 다르다. 규칙을 하나로 두기 위해 「사이클당 1 시도」를 유지하는 쪽을 권고, 즉시 소진을 원하면 Manager 가 바꾼다 |

### 열린 결정의 답

**D2 — 한도 값: ㄱ `attempts ≥ 3` (증거로 답함, 사람 결정 불필요).**

계산값(측정 아님). 기본값: `DefaultCriticalAttempts = 3` (`notifier.go:45`) · `DefaultRetryDelay = 2 s` (`:48`) ·
`DefaultPublishTimeout = 10 s` (`alert_lease.go:17`) · `alertDeliveryInterval = 2 s` · 배치 10 (`alertdelivery.go:63·74`).
실행자는 사이클 먼저, 대기 뒤(`:122-136`). 원장 쓰기·반납 대기는 뺐다(동기 경로 문서상 상한은 그것까지 넣어 54 s, `alert_claim.go:21`).

| 상황 (한도 3) | 동기 경로 | 실행자 |
|---|---:|---:|
| 행 하나, 즉시 실패 | 4 s (0·2·4) | 4 s (0·2·4) |
| 행 하나, 매번 10 s timeout | 34 s | 34 s (0–10 · 12–22 · 24–34) |
| 배치 10 행 전부 timeout | — | 첫 행 214 s · 끝 행 304 s |

- 같은 양상에서 **한도 3 = 동기 경로와 같은 시간**이다. ㄴ(「등가 시간」)은 한 값으로 정의되지 않는다 — 34 s ÷ 2 s = 17 회로 잡으면
  즉시 실패 32 s 지만 timeout 이면 12×17−2 = 202 s 가 되어, 동기 경로보다 **덜 보수적**이다. ㄷ 도 같은 식을 따른다.
- a092 델타가 이미 「전송 시도 횟수는 다시 계약이다 — 줄이지 않는다(SHALL NOT)」 고 적는다(`a092…/spec.md:74`).
- 대가: `Notify` 를 거친 행에서는 오늘과 같다(동기 경로가 이미 4~34 s 에 잠그고 승격한다). **새로 생기는 대가**는 `Notify` 없이
  기록되는 행(`execgw/replay.go:551` `parkAlert`)과 a092 뒤의 모든 행이 durable 승격을 만든다는 것 — 보수 방향이다.
- 사람 결정이 필요한 경우는 ㄴ·ㄷ 처럼 한도를 **늘리는** 쪽을 고를 때뿐이다(진입 차단 완화).

**D3 — publisher 부재를 실패 시도로 센다 (증거로 답함, 사람 결정 불필요).**

- `newNotifier` 문서가 이미 *"nil publisher … outbox row is written and stays PENDING, the entry gate latches, and sustained failure
  escalates to ENTRY_BLOCKED. That is the specified direction"* 라고 적는다(`exitwiring.go:60-70`).
- 동기 경로: `deliver` B3(`notifier.go:429-431`) → B26·B27 잠금(`:565-571`), `notifyCritical` B4 승격(`:223-228`).
- a092 델타 「전송 수단이 구성되지 않은 배선에서도 시도는 실패로 세어져야 한다(SHALL)」 + 시나리오 「전송 수단이 없는 배선」
  (`a092…/spec.md:76·108-110`).
- 세지 않는 안은 무설정 엔진이 영구히 진입을 열어 두는 구멍이다. 지연 차이는 F13.

### Manager 결정 (재freeze 전에 design 에 적을 것)

**M1 — F2 의 배제 수단.** 사람의 규칙(승인이 이긴다)은 이미 있다. 남은 것은 수단이다.

| 안 | 내용 | 완결성 | 표면 | 비용 |
|---|---|---|---|---|
| A | 시도 기록 + 한도 판정 + 모드 승격을 **원장 트랜잭션 하나**로(새 journal 메서드). 게이트 `Block` 은 커밋 뒤, 직전에 행 상태 재확인 | 모드: 완결(승인 tx 와 SQLite 가 직렬화) · 게이트: 재확인~`Block` 사이 잔여 창 | journal 1 메서드 | 잔여 창의 결과는 보수 방향(빈 backlog 에 잠긴 게이트 — 빈 목록 `Acknowledge` 로 풀림, `notifier.go:854-876`) |
| B | `EntryGate` 에 사유별 해제 세대 — 「그 뒤 해제되지 않았으면 잠금」 | 게이트 완결 | **execgw(High-risk) 표면 추가** — Non-goal 밖 | 모드는 A 가 따로 필요 |
| C | 실행자의 정산~판정~`Block`~`Escalate` 를 `Notifier` 의 배제 아래(원격 전송은 밖) — a092 델타가 허용하는 형태(「잠금은 고를 때와 정산할 때만」, `a092…/spec.md:60`) | 완결(`Acknowledge` 가 같은 배제 아래 세고-푼다) | `obs` 에 배제 진입점 1 개(동기 경로 함수 무편집) | 원장 로컬 연산 두 개(~10 ms)만큼 동기 `Notify` 가 기다릴 수 있다 → 2.6 보존 시험으로 잰다. 역으로 a092 착지 전에는 동기 `deliver` 가 원격 전송 동안 `n.mu` 를 쥐므로(`notifier.go:251-255`) 실행자가 최대 한 동기 예산(54 s)만큼 기다린다 — 손절 무관, 배달 지연. `escalate` 가 `n.mu` 밖인 이유(announcer 재진입)는 실행자가 announcer nil 이라 해당 없음 |

Teammate 권고는 **C**(잔여 창 없음, Non-goal 「동기 경로 편집 없음」 유지, `Acknowledge` 조건 불변). A 는 잔여 창을 「보수 방향이라 수용」으로
적을 때만. **어느 안도 `Acknowledge` 의 해제 조건(승인 + 미전달 0)을 바꾸지 않는다** — 바꾸는 안이 필요해지면 그때 사람 결정으로 올린다.

**M2 — F3(B10) 정책.** 기록 실패는 attempts 를 못 올리므로 원장 밖 판정이 필요하다.

| 안 | 내용 | 방향 |
|---|---|---|
| (i) | B10 즉시 `Block`(고정 detail) + 승격 시도 | 가장 보수 — 일시적 busy 한 번이 운영자 승인을 부른다 |
| (ii) | 행별 **연속 B10** 을 메모리로 세어 같은 한도(3)에서 `Block` + 승격 시도, 성공 기록이면 0 으로 | 보수 — 재시작하면 계수는 지워지나 기동 복원이 PENDING 으로 잠근다 |
| (iii) | 오늘과 같이 로그만 | a092 뒤 구멍 — **권고하지 않음** |

Teammate 권고는 (ii). 둘 다 보수 방향이라 사람 결정 대상은 아니다(Manager 선택). spec 은 「기록 실패도 지속 실패다」 한 문장을 요구에 더한다.

### 재freeze 조건

F1·F2·F3·F4·F8 을 design(과 spec 시나리오: 발송 중 승인 · 승인 직후 늦은 실패 · 기록 실패 지속)에 반영하고, F5·F6·F7·F9·F10·F11·F12·F13 을
기록·교정한 뒤 같은 두 보이스로 재리뷰한다. 그때까지 tasks 0.4 는 열어 두고 1.2 이후는 착수하지 않는다.

## §0.2 proposal-freeze 2판 (2026-09-26) — 판정 **REJECT** → design 3판

- 입력: Manager 결정 **M1 = C · M2 = (ii)** (2026-09-26, coordinator 메시지). design 2판 = D1 판정 입력 · D2/D3 확정 · D6 재유도 · D7 배제 ·
  D8 기록 실패 · D9 정제 + 1판 발견 처분표. spec 에 「승인이 이긴다」 · 「기록 실패도 지속 실패」 · 시나리오 3, tasks 2.7~2.11 · 3.4~3.6.
- 2판이 새로 기대는 함수의 AST 번들을 **문서보다 먼저** 추출: `Journal.settleUnderClaim`(편집 대상) · `Notifier.Acknowledge` ·
  `Journal.TransitionOperatingMode`(대조), base `4798d399`. `check_analysis.py` 에서 이 셋의 형식 오류 0.
- 보이스: codex 2회차(`analysis/freeze-review/codex-r2-prompt.md` · `codex-r2-output.md`, 같은 설정 · 독립 프롬프트 — 1판 발견은 「검증할 주장」으로만
  줬다) + Teammate 적대 Eng.
- codex 2회차: 1판 발견 F1·F7·F9·F10·F11·F12·F13 **RESOLVED**, F2·F3·F4·F5·F6·F8 **PARTIAL**. 새 발견:

| id | 심각도 | 발견 | 판정 → design 3판 |
|---|---|---|---|
| N1 | P1 | 배제 구간에 승격 트랜잭션·로그까지 넣어 exit `Notify` 가 그만큼 기다린다; 「연결 대기가 옮겨 갈 뿐」은 거짓 | 수용 → 배제 = 정산 트랜잭션 하나, 승격·로그 밖; 승격을 밖에 둬도 최종 상태가 같다는 논증(승인은 모드를 풀지 않는다); exit 몫은 결정적 시험 2.6 |
| N2 | P1 | 승인 뒤 기록이 원장 **오류**로 끝나면 결과를 몰라 D8 이 승인 뒤 재잠금 | 수용 → 한도 도달 시 배제 안 울타리 읽기(`UndeliveredCount`) — 0 이면 잠그지 않음, 읽기 오류면 잠금(원장 불가독 = 감시 실패, a092 :66 이 겨냥한 사실과 다름을 기록) |
| N3 | P1 | 발행 성공 + 전달 정산 실패(`deliverOne` B11)가 판정 밖 — 동기 경로는 즉시 잠근다(`notifier.go:465-491`) | 수용 → 전달 정산 표: 오류·`NotFound`·모르는 값 → 즉시 잠금 + 승격, 임차 유지. Teammate 추가: 오류는 발행~정산 사이 승인을 가리므로 **같은 울타리** |
| N4 | P1 | 행별 계수의 수명(재무장 같은 id · 가득 찬 배치 · 반납 성공 초기화) 미정 | 수용 → 계수를 **실행자 단위 하나**로; 초기화는 정산 쓰기 성공만. 한계(간헐 실패 · 한 행만 영구 실패)는 기록 — 기동 복원이 덮는다 |
| N5 | P1 | D6 가 조건부 예시를 일반 상한처럼 썼다 | 수용 → 일반 상한 없음 명시, 전제 H(동질 두절) 조건부 식, 이질 실패의 `K` 항, 무계 경우 명명, 781 s 철회 |
| N6 | P2 | 실패 기록 `NotFound` 의 「승격 parity」는 틀렸다(`lost` → `owed=false`) | 수용 → 잠금만 |
| N7 | P1 | spec/tasks 가 새 실패 정책을 다 싣지 않았다 | 수용 → spec: 나열 오류 · 정산 성공만 초기화 · 울타리 · 전달 기록 실패 · 배제 범위 · 로그 정제 + 시나리오 2; tasks 2.6·2.8~2.12, 뮤테이션 목록 확장 |
| N8 | P2 | 배포 첫 사이클 승격은 조건부 | 수용 → Risks 문장 교정 |

- F8 부분 해소 지적(로그의 `err.Error()` 에 계좌): 2판의 「계좌 필드 없이」는 기존 로그 관례(`Notifier.escalate` 가 `FieldAccount` 를 쓴다)와
  어긋나 **철회**했다. 게이트 detail 정제(불변식 8의 노출 자리)는 유지. 로그 규약 자체의 강화는 이 change 범위 밖.
- Teammate 적대 Eng 추가 발견(3판 반영): 전달 정산 오류의 승인 가림(위 N3 행) · a092 착지 전 `M ≤ 54 s` 가 81 s 임차 안에 드는지(D6 주석) ·
  역순 잠금(연결을 쥔 채 `n.mu`) 부재를 `Notify` 호출 7 자리와 journal 콜백 메서드 0 으로 확인(D7).

## §0.3 proposal-freeze 3판 (2026-09-26) — 판정 **REJECT** → design 4판

- 보이스: codex 3회차(`codex-r3-prompt.md` · `codex-r3-output.md`, 같은 설정 · 독립 프롬프트, 이전 발견은 「검증할 주장」) + Teammate 적대 Eng.
  3회차 실행 중 Teammate 가 design D4 의 한 칸(`HeldElsewhere` 유계 조건)을 고쳤다 — codex 가 어느 판을 읽었는지 확정할 수 없어 그 칸은 4회차가 다시 본다.
- codex 3회차: F1·F4·F6·F7·F9·F10·F11·F12·F13·N6·N8 **RESOLVED**, F2·F3·F5·N1·N2·N3·N4·N5·N7 **PARTIAL**, F8 **NOT RESOLVED**. 새 발견:

| id | 심각도 | 발견 | 판정 → design 4판 |
|---|---|---|---|
| R1 | **P0** | 3판은 정산 트랜잭션을 `n.mu` 안에 넣었다 — exit `Notify`(`exitloop.go:1710`)와 비상 청산(`flatten.go:694`, `context.Background()`)이 잡는 뮤텍스를 실행자가 원장 대기 동안 쥔다. `sync.Mutex` 는 ctx 를 보지 않아 **취소 불가 대기**. 「트랜잭션 하나」는 시간을 묶지 않는다 | **수용** → 배제를 **메모리 전용**으로: `Notifier.ackGen`(원자) + `AckGeneration()` + `LatchUnlessAcknowledgedSince(gen, Block)`. 원장·승격·로그·반납 전부 배제 밖. 대가: `Acknowledge` 에 세대 증가 **한 줄**(해제 조건 불변) |
| R2 | P1 | D9 가 기존 관례를 들어 새 로그에 계좌·원문 오류를 허용 — 불변식 8 위반 | **수용** → 새 줄은 허용 목록 필드만, 승격 오류는 `err` 대신 고정 분류, sentinel 시험. 3판 문장 철회 |
| R3 | P1 | 실행자 단위 계수는 `LeaseLost`/`AlreadySettled`(0 행 쓰기)와 다른 행 성공이 지운다 → 한 행 영구 실패를 놓친다 | **수용** → 행별 계수로 되돌리고 그 행의 `Applied` 만 지움 |
| R4 | P1 | 전역 `UndeliveredCount` 울타리는 오류 난 에피소드가 남았는지 모른다(승인 → 새 행 B → A 오류 → B 때문에 재잠금) | **수용** → 원장 읽기 울타리 폐기, 승인 세대 울타리. 승인은 계수 맵 전체를 비운다 |
| R5 | P1 | spec 이 승격을 배제 안과 밖에 동시에 요구, tasks 3.5 와 4.1 이 모순 | **수용** → 배제 계약 하나(메모리 전용)로 spec·tasks·처분표 정렬 |
| R6 | P1 | 「임차 유지 → 다음 사이클 재발행 없음」은 거짓 — 배치가 81 s 를 넘기면 만료 후 재발행 | **수용** → 보장을 「임차가 살아 있고 교체되지 않은 동안」으로, 동기 경로도 같다(`notifier.go:486-490`), 만료 시험 추가 |
| R7 | P2 | D6 의 `M ≤ 54 s` 는 근거 없음, 만료는 `LeaseLost` 를 만들지 않는다(토큰 교체만) | **수용** → 철회, 문장 교정. 4판의 `M` 은 판정만 늦춘다 |
| R8 | P2 | 실행자 단위 계수는 한 배치 안에서 3 을 채워 「최소 ≈ 4 s」가 거짓 | **수용** → 행별 계수로 최소 ≈ 2 사이클 복원 |

- codex 지적 교정 두 가지: gate OFF 경계의 인용을 `cmd/tossctl/engine.go:220-221`(`errEngineGateOff`)로; 「journal 에 콜백 없음」은 틀렸다(`TransitionOperatingMode`
  가 트랜잭션 안에서 `Auditor.RecordAction` — FLM B23). 4판은 원장을 배제 밖에만 두므로 교착 논증이 그것에 기대지 않는다.
- **Manager 확인 필요**: 4판의 R1 해법은 M1 = C 의 「Notifier 배제」 안이지만 `Acknowledge` 에 **한 줄**(세대 증가)을 넣는다. 해제 조건(승인 + 미전달 0)은
  바뀌지 않는다. 비목표 「`Acknowledge` 의 해제 조건을 바꾸지 않는다」와 충돌하지 않는다고 판단했고 proposal Impact 에 적었다.
- Teammate 적대 Eng 추가(4판 반영): 반납을 울타리 **전**으로(실행자가 `n.mu` 를 기다리는 동안 임차를 쥐지 않게), 세대 비교 시점(사이클 시작 · 울타리)을 명시.

## §0.4 proposal-freeze 4판 (2026-09-26) — 판정 **REJECT** · **Manager 재결정 필요 (M1)**

- 보이스: codex 4회차(`codex-r4-prompt.md` · `codex-r4-output.md`, 같은 설정 · 독립 프롬프트) + Teammate 적대 Eng(아래 재확인).
- codex 4회차: F1·F4·F6~F13 · N6 · N8 · R2 · R3 · R6 · R7 · R8 **RESOLVED**; F2 · F3 · F5 · N1~N5 · N7 · R1 · R4 · R5 **PARTIAL**. 새 발견:

| id | 심각도 | 발견 | Teammate 재확인 |
|---|---|---|---|
| Q1 | **P0** | 4판의 「메모리 전용」 배제도 손절 경로에 원격 대기를 들인다: 실행자가 `n.mu` 를 쥔 채 `Gate.Block` → `g.mu` 를 기다리는데, 전략 진입 dispatch 는 `g.mu` 를 **브로커 전송 콜백 동안** 쥔다 | **확인** — `execgw/strategy_entry_gate_authority.go:60-72` (`g.entry.mu.Lock(); defer Unlock(); … return fn()`), `fn` 은 `gateway.go:692-708` 에서 `plan.call(tctx, …)` 로 전송. exit `Notify`(`exitloop.go:1710`)·비상 청산(`flatten.go:694`)이 `n.mu` 를 기다린다 |
| Q2 | P1 | 세대를 `Acknowledge` **시작**에 올리면, 진행 중인 승인 안에서 스냅숏한 실행자가 같은 세대로 울타리를 통과해 해제 뒤 재잠금 | 확인 — 증가 지점과 `Clear` 지점(:875) 사이에 원장 연산이 있다 |
| Q3 | P1 | 무관한(또는 실패한) 승인이 세대를 바꿔 필요한 판정을 버리고, 그 행이 다음에 성공하면 **영영 잠기지 않는다**; 반복 승인이 D8 계수를 계속 지운다 | 확인 — 4판의 「다음 실패가 다시 판정」은 보장이 아니다(전달 성공 시 PENDING 을 떠남, `outbox.go:453-456`) |
| Q4 | P2 | 정산 `Applied` 뒤 · 울타리 전의 전달+재무장은 옛 판정을 새 에피소드에 적용(보수 방향) | 수용(기록 예정) |
| Q5 | P2 | 남이 정산한 행의 계수가 승인·재시작까지 남는다 | 수용(기록 예정) |
| Q6 | P2 | D6 전제 H 가 공식에 불충분, a092 식에 `C` 를 넣으면 발행 시간을 이중 계산(`Q + L·T + (L−1)·C`) — a092 식은 재정의가 아니라 교체해야 | 수용(기록 예정) |
| Q7 | P3 | 처분표 F2 행이 D7 과 모순 | 반영(표 교정) |

### 기존 결함 (a124 범위 밖, 보고)

Q1 의 경로는 **HEAD 동기 경로에 이미 있다**: `Notifier.deliver` 와 `claimAndDeliver` 가 `n.mu` 를 쥔 채 `n.Gate.Block`(`notifier.go:280·484·520·571`)을 부른다 →
전략 dispatch 가 `g.mu` 를 전송 동안 쥐면, 전달 실패를 겪은 동기 `Notify`(exit 루프 포함)와 그 뒤의 모든 `n.mu` 대기자가 브로커 전송을 기다린다.
전략 진입은 이 빌드에서 휴면이라(`cmd/tossctl/engine_strategy_entry_dormant_test.go`) 오늘 발현하지 않지만, **전략 진입을 켜기 전에** 닫혀야 한다.
`EntryGate` 의 `g.mu` 를 전송 콜백 동안 쥐는 설계(a061 계열 ABA 봉인)의 문제이며 a124 가 고칠 표면이 아니다.

### M1 재결정 요청 — 선택지

C(`Notifier` 배제)는 두 형태 모두 떨어졌다: 3판(원장 대기를 `n.mu` 안에, R1) · 4판(`g.mu` 대기를 `n.mu` 안에, Q1). **`n.mu` 안에서 `Gate.Block` 을
부르는 모든 형태는 Q1 을 물려받는다** — 동기 경로의 기존 결함과 같은 모양이다.

| 안 | 내용 | Q1 | Q2 | Q3 | 표면 |
|---|---|---|---|---|---|
| **B′ (권고)** | `EntryGate` 에 사유별 **해제 세대**: `Clear(reason)` 가 실제로 지웠을 때만 그 사유의 세대를 올린다(`g.mu` 아래, 이미 `revision++` 하는 자리 — `retry.go:537-543`). 새 메서드 `ClearEpoch(reason)` · `BlockUnlessClearedSince(reason, epoch, detail)`(둘 다 `g.mu` 만). 실행자는 정산 **전**에 세대를 읽고, 판정 뒤 `BlockUnlessClearedSince`. `n.mu` 를 잡지 않는다. `Acknowledge` 무편집 | 해소 — 실행자는 손절 경로가 기다리는 뮤텍스를 쥐지 않는다(`g.mu` 대기는 실행자 자신의 것) | 해소 — 세대가 **해제 순간**에 오른다 | 해소 — 해제는 미전달 0 일 때만 일어나므로(FLM acknowledge B10), 세대가 바뀐 판정은 그 행이 이미 승인·전달된 것. 실패한·부분 승인은 세대를 안 바꾼다 | **execgw(High-risk)** 필드 1 · 메서드 2, `Clear` 한 줄 |
| C″ | C 유지, `Block` 을 `n.mu` 밖으로 빼고 `n.mu` 안에서는 「해제 예약」만 | Q1 해소 | 해제~`Block` 사이 창이 다시 열린다(F2 재발) | — | obs |
| A | 원장 트랜잭션으로 판정 + 게이트를 원장에서 유도 | 해소 | 해소 | 해소 | 게이트 래치를 원장 파생으로 — 기동 복원·`Acknowledge` 해제 경로 변경 = **사람 결정** 대상 |

Teammate 권고는 **B′**. 1판 M1 표에서 B 를 「execgw 표면 추가 — Non-goal 밖」이라 낮게 봤으나, C 가 Q1 을 피할 수 없다는 것이 드러났다.
B′ 는 `Acknowledge` 의 해제 조건과 `Notifier` 를 건드리지 않는다. 결정되면 design 5판(D7 · D8 · spec 「승인이 이긴다」 · tasks 2.9 · 2.10 · 3.5 ·
proposal Non-goals/Impact)을 쓰고 Q4~Q6 을 반영해 5회차 재freeze 한다. **tasks 0.4 는 체크하지 않는다.**

## §0.5 proposal-freeze 5회차 (2026-09-26) — design 5판 1쇄 **REJECT** → 5판 2쇄

- 입력: Manager 재결정 **M1 = B′**(2026-09-26 — 근거 세 자리를 Manager 가 실측 재확인). design 5판 1쇄 = `EntryGate` 사유별 해제 세대(필드 1 · `ClearEpoch` ·
  `BlockUnlessClearedSince` · `Clear` 한 줄), 실행자 `n.mu` 무관, `Acknowledge` 무편집, spec 새 요구 「진입 게이트는 사유별 해제 세대를 가진다」, Q4~Q6 반영,
  proposal Follow-ups 「전략 진입 활성화 전 필수 선행」.
- 편집 전 AST 번들 추가: `EntryGate.Clear`(편집 대상) · `EntryGate.Block`(대조), base `4798d399`. `Notifier.Acknowledge` 번들은 「무편집」으로 되돌림.
- **Manager 지시와 다른 점 (확인 요청)**: 해제 세대를 「실제로 지웠을 때만」이 아니라 「해제 요청마다」 올린다. 반례 ㉣(래치가 서기 전 전체 승인 →
  지울 래치 없음 → 세대 불변 → 빈 backlog 위 재잠금 + durable 승격). codex 5회차가 독립으로 동의: *"The author's deviation from M1 is correct and
  necessary"*. `revision` 은 지시대로 실제 삭제 때만.
- Teammate 적대 Eng(1쇄 작성 중): 세대를 임차 **전**에만 읽으면 「읽기 → 해제 → 같은 id 재무장 → 임차」에서 새 에피소드의 판정이 버려진다 → 1쇄는 임차 **후**로 옮겼다.
- codex 5회차(`codex-r5-prompt.md` · `codex-r5-output.md`): Q1 · Q2 · Q4 · Q5 · Q7 · N1 · R1 등 다수 RESOLVED. 새 발견:

| id | 심각도 | 발견 | 판정 → 5판 2쇄 |
|---|---|---|---|
| V1 | P1 | 임차 **후** 한 번 읽기도 깨진다: 임차 → 승인·해제 → 새 세대 읽기 → 발행 → 전달 정산 오류 → 잠금 + 승격(승인된 에피소드가 새 세대를 물려받음) | **수용** → 임차 **전후 두 번** 읽고 다르면 그 행을 이번 사이클에 보내지 않음(㉨). 임차 전 읽기의 재무장 반례도 함께 막힌다 |
| V2 | P1 | (P) 의 「해제 순간 미전달 0」은 거짓 — `Acknowledge` 는 셈(:870) **뒤** 해제(:875)하고, 그 사이 `replay.go:551` `EnqueueAlert` 가 `n.mu` 없이 새 행을 넣는다. 그 행의 필수 판정이 버려진다 | **수용** → 울타리 실패면 버리기 전에 새 세대를 읽고 **원장에서 행 상태 확인**: PENDING 아니면 버림, PENDING·읽기 오류면 새 세대로 재시도, 세 번 연속 경합이면 무조건 잠금(㉩). 셈~해제 경합 자체는 HEAD `Acknowledge` 의 성질이라 a092 소유로 Follow-ups |
| V3 | P1 | proposal 은 「임차 전」, design/spec 은 「임차 후」 | **수용** → 전후 두 번으로 세 문서 정렬 |
| V4 | P1 | spec 「손절 경로가 기다리는 잠금을 잡지 않는다」가 `g.mu`(비상 청산 `Gate.Block` 도 잡음)와 모순 | **수용** → 「실행자가 잡는 잠금은 게이트 잠금뿐, 그 안에서 게이트 상태만 — 손절 경로의 대기는 그 구간 하나로 경계」 |
| V5 | P2 | 같은 id 에피소드 carryover · 나열 계수 초기화 순서 | **수용** → carryover 는 보수 방향으로 받아들이고 시험 기대값으로, 나열 계수는 증가 전에 세대 비교 |
| V6 | P2 | D6 에 마지막 나열 비용 누락, tasks 5.2 에 옛 지시 | **수용** → `Q + (L−1)·C + I_list + (T + S + M)`, 「조건부 상한」, 5.2 는 식 교체 |


## §0.6 proposal-freeze 6회차 (2026-09-26) — design 5판 2쇄 **REJECT** → 5판 3쇄

- codex 6회차(`codex-r6-prompt.md` · `codex-r6-output.md`): 세대 괄호(임차 전후 두 번)는 **sound**, 해제 요청마다 올리는 편차는 **necessary**(재확인),
  `revision` 보존 옳음, `Clear(ReasonAlertUndelivered)` 의 다른 경로 없음, 실행자 잠금 순서에 원격 대기 전파 없음, D8 행별 계수 「substantially repaired」,
  D1 가산 읽기 구현 가능(단 **`tx` 로 읽을 것** — 연결 하나). 새 발견:

| id | 심각도 | 발견 | 판정 → 5판 3쇄 |
|---|---|---|---|
| W1 | **P0** | 원장 확인이 「PENDING 아님」이면 버리는데, DELIVERED 는 사람 확인이 아니다. 셈~해제 사이에 들어온 B 가 한도에 닿고 → 옛 해제가 세대를 올리고 → 동기 경로가 B 를 전달하면 B 는 **승인 없이** 판정을 잃는다(정본 :183 「수동 확인으로 해제」 위반) | **수용** → 버리는 근거는 **ACKNOWLEDGED 뿐**, DELIVERED·PENDING·읽기 오류는 잠금(㉪) |
| W2 | P1 | 나열 계수는 행이 없어 원장 확인을 할 수 없다 | **수용** → 울타리 실패면 계수를 0 으로 되돌려 재계수. 근거: 해제는 `Acknowledge` B10 에서만 오고 거기 닿으려면 원장 읽기가 성공했어야 한다 — 연속의 전제가 반증됨 |
| W3 | P1 | 세 번 경합 뒤 무조건 잠금은 「승인이 이긴다」와 모순(마지막 해제가 실제 승인일 수 있다) | **수용** → 무조건 잠금 폐기, **다음 사이클로 미룸**(㉫). 재시작 시 소실은 기존 메모리 래치 한계와 같음을 기록(기동 복원은 PENDING 만 센다) |
| W4 | P2 | 재시도 헬퍼가 D1 의 「잠금만」 판정(실패 기록 `NotFound`·모르는 값)을 잃는다 | **수용** → `latchConfirmed(row, e, escalate)` 로 운반 |
| W5 | P3 | D8 · spec 해제 세대 요구 · tasks 3.5 · 불변식 문장에 옛 (P) 잔재 | **수용** → 정렬 |

## §0.7 proposal-freeze 7회차 (2026-09-26) — design 5판 3쇄 **REJECT** · Manager 확인 요청 (판정 원칙)

- codex 7회차(`codex-r7-prompt.md` · `codex-r7-output.md`): 세대 기계 자체는 **sound**, 해제 요청마다 올리는 편차 **correct and necessary**(세 번째 재확인),
  잠금 순서 문제 없음, D1 · D2 · D3 · D6 수치 확인. 새 발견:

| id | 심각도 | 발견 | Teammate 판정 |
|---|---|---|---|
| X1 | P1 | 미룬 판정을 새 세대로 `latchConfirmed` 에 넣으면 울타리가 먼저 통과해 원장 확인 없이 잠근다 — 사이클 사이의 승인을 못 본다 | 확인 |
| X2 | P1 | **ACKNOWLEDGED 만 근거로 삼으면 「전달 뒤 수동 해제」를 표현할 수 없다.** `AcknowledgeAlert` 는 PENDING 행만 바꾸므로(`outbox.go:494-503`) 전달된 행은 결코 ACKNOWLEDGED 가 되지 않는다. 판정 대기 → 남이 전달 → 운영자 승인·해제 → 늦은 판정이 DELIVERED 를 보고 재잠금 + 승격. W1(DELIVERED 로 버리면 승인 없이 판정 소실)과 정면으로 부딪힌다 | 확인 — 3쇄의 근거 규칙은 틀렸다 |
| X3 | P2 | 미룬 판정 맵의 작업 예산 · 삭제 · 병합 규칙, D8 거름 예외, 반납 `Applied` 제외 명시 | 확인 |
| X4 | P2 | 「exit 체류 불변」은 과대 — 실행자의 원장 트랜잭션이 연결 하나를 점유한다 | 확인 — 「원격 대기 전파 없음」과 「원장 몫은 측정」으로 가른다 |
| X5 | P3 | 처분표 V2 행에 폐기된 「무조건 잠금」, deliverOne FLM 의 B8 요약(「attempts+1」)이 틀림 | 확인 |

### 왜 수렴하지 않는가 — 원칙이 없었다

2~3쇄는 「해제가 있었으면 → 원장에 무엇이 적혀 있으면 버린다」를 사례마다 고쳐 왔다(PENDING 아님 → ACKNOWLEDGED 만). W1 과 X2 는 **같은 원장 상태
(DELIVERED)** 에서 반대 답을 요구한다 — 원장 상태로는 가를 수 없다. 가르는 것은 **시간 순서**다: 해제가 판정 근거보다 **앞**이었나 **뒤**였나.

### 제안 — 원칙 E 「늦은 적용은 제때 적용과 같아야 한다」 (6판 방향, Manager 확인 요청)

> 실행자가 판정을 늦게 적용한 결과(게이트 · 모드)는, 그 판정을 **근거가 커밋된 순간** 원자적으로 적용했을 때와 같아야 한다. 그 순간 **뒤**의 해제는
> 제때 적용된 래치를 지웠을 것이므로 판정을 버린다. 그 순간 **앞**의 해제는 지우지 못했을 것이므로 적용한다.

구현은 3쇄보다 **작아진다**:

- 해제 세대를 **정산이 돌아온 직후**(`Applied` 면 커밋 뒤) 한 번 읽는다. 임차 전후 괄호 · 원장 재확인 루프 · 미룬 판정 맵이 전부 필요 없다.
- `Applied` 판정: 세대가 바뀌었으면(= 커밋 뒤 해제) **버린다**, 같으면 잠근다. 커밋과 세대 읽기 사이의 해제를 놓치는 창은 잠그는 쪽으로만 틀린다.
- 오류 판정(전달 정산 오류 · D8 기록 실패 계수 · 나열 계수): 근거의 순간 = 오류가 돌아온 순간. 그 뒤 세대를 읽고, 원장에서 행이 ACKNOWLEDGED 면 버린다
  (오류가 승인을 가렸을 수 있다 — V1). 그 밖은 같은 세대 규칙.
- 사례 대조: X2(판정 → 전달 → 수동 해제) → 해제가 근거 뒤 → **버림**(정본 「전달 복구 후 수동 확인으로 해제」와 같다). W1(B 의 셋째 실패 커밋 →
  옛 승인의 늦은 해제) → 해제가 근거 뒤 → **버림** — 제때 잠갔어도 그 해제가 지웠을 것이다. B 가 잃는 것은 이 change 의 판정이 아니라 **`Acknowledge` 의
  셈~해제 경합**(Follow-ups, a092 소유)이다. ㉣(래치 전 전체 승인) → 해제가 근거 뒤 → 버림. V1(임차 → 승인·해제 → 정산 오류) → 원장 ACKNOWLEDGED → 버림.
- spec 「승인이 이긴다」를 원칙 E 문장으로 바꾼다. W1 을 P0 로 본 codex 판정은 「제때 적용해도 같은 손실」이라는 논증으로 **거절** 대상이 된다 —
  이것이 Manager 확인이 필요한 이유다(리뷰 판정을 뒤집는 논증).

**요청 두 가지 (Manager):** ① 원칙 E 채택 여부 ② 해제 세대를 「해제 요청마다」 올리는 편차(5·6·7회차 codex 가 세 번 「necessary」로 확인) 승인.
채택되면 6판을 쓰고 8회차 재freeze 를 돈다. **tasks 0.4 는 체크하지 않는다.**

## §0.8 Manager 확인 (2026-09-26) → design 6판

- **원칙 E 채택** — 「늦은 적용은 제때 적용과 같아야 한다」. Manager 가 영수증 세 곳(`outbox.go` PENDING 전용 승인 UPDATE · `notifier.go` 셈 :870 → 해제 :875 ·
  `replay.go` 무잠금 `EnqueueAlert`)을 직접 재확인했다.
- **6회차 W1 P0 를 논증으로 거절 — Manager 승인 (근거: a092 사람 결정).** W1 은 「옛 승인의 늦은 해제 뒤 남이 전달하면 B 의 판정이 승인 없이 사라진다」였다.
  원칙 E 아래에서 그 판정은 근거 확정(B 의 한도 커밋) **뒤**의 해제로 버려지며, 이는 제때 적용했어도 그 해제가 지웠을 결과와 같다. 이 거절은 새 완화가
  아니라 기존 사람 결정 「운영자의 승인은 … 전송의 성공 여부는 그 사유를 되살리지 않는다」(a092 델타 `specs/engine-safety/spec.md:66-68`, 사용자 결정
  2026-08-10 결정 2 = 안 1)의 적용이다. B 가 실제로 잃는 것은 `Acknowledge` 셈~해제 경합(proposal Follow-ups, a092 소유)이다.
- **해제 세대 편차 승인** — 「해제 요청마다 증가」. Manager: 해제 요청은 미전달 0 확인 뒤에만 일어나므로(:874–875) 요청 자체가 사람 승인의 표식이고
  삭제 유무는 우연이다. ㉣(빈 backlog 재잠금 + durable 승격 잔존)을 RED 시험으로(tasks 2.9).
- design 6판: D7 을 원칙 E 로 재작성(정산 직후 세대 한 번, 오류 판정만 원장 승인 **시각**, 괄호 · 재확인 루프 · 미룬 판정 맵 제거), D8 동일 규칙,
  spec 「늦은 적용은 제때 적용과 같아야 한다」 + 두 갈래 시나리오(앞 → 잠금 · 뒤 → 버림, 제때 적용 등가성) + 오류가 가린 승인, tasks 2.9 · 2.10 · 4.1 재작성.

## §0.9 proposal-freeze 8회차 (2026-09-26) — design 6판 1쇄 **REJECT** → 6판 2쇄

- codex 8회차(`codex-r8-prompt.md` · `codex-r8-output.md`): 원칙 E 와 해제 요청마다의 세대 증가는 방향이 맞다고 보았고, W1 거절은 「게이트 래치에 한해」 성립,
  운영 모드에는 성립하지 않는다고 판정(PARTIAL). 새 발견:

| id | 심각도 | 발견 | 판정 → 6판 2쇄 |
|---|---|---|---|
| Y1 | **P0** | 원칙 E 의 「게이트와 모드 등가성」이 1쇄에서 틀렸다 — 제때 적용은 차단 **과 승격**을 하고 뒤의 `Acknowledge` 는 **차단만** 지운다(모드는 사람의 완화로만). 1쇄는 「뒤」 해제에서 승격까지 버려 제때라면 남았을 durable `ENTRY_BLOCKED` 를 잃는다 | **수용** → 「뒤」 해제면 차단만 버리고 **승격은 적용**. 이것으로 W1 거절도 완결된다 — B 의 사건은 운영 모드로 남는다 |
| Y2 | P1 | 오류 판정의 `acknowledged_at < t_j` 비교는 승인 완료 순서가 아니다(승인 트랜잭션이 연결을 얻기 전에 찍힌 벽시계, 되감김) — 부분 승인이 판정을 부당하게 버린다 | **수용** → 시각 비교 폐기. 오류 판정은 원장 상태로 버리지 않는다(보수 — 빈 목록 승인이 푼다) |
| Y3 | P1 | D8 의 거름 예외(「이미 정산됨」 · 완전 나열 부재 · 나열 성공)가 spec 에 없다 | **수용** → spec 에 전부 |
| Y4 | P1 | tasks 2.6 의 「exit 체류 불변」이 원장 연결 하나(`journal.go:174`)와 모순 | **수용** → (a) 잠금·원격 대기 격리 단언 (b) 트랜잭션 안 지연 주입으로 원장 경합 측정, 수락선 = 기준선 + 트랜잭션 하나 |
| Y5 | P2 | 세대 읽기가 `g.mu` 를 무한정 기다릴 수 있고 그동안 임차를 쥔다 | **수용** → 실패 기록 경로는 반납 뒤에 읽음, 전달 정산 오류 경로(임차 유지)의 만료 재발행은 R6 과 같은 성질로 기록 |
| Y6 | P3 | proposal 의 B8 기준선(「attempts+1」) · Follow-ups 의 시간 순서 문장이 뒤집힘 | **수용** → 교정 |

- **Manager 문구 교정 (보고)**: §0.8 의 「㉣(빈 backlog 재잠금 + durable 승격 잔존)을 RED」 중 **「durable 승격 잔존」은 원칙 E 아래에서 결함이 아니다** — ㉣ 을 제때
  적용하면 차단 + 승격 → 해제가 차단만 지우므로 승격은 남는다. 그래서 RED ㉣ 은 「빈 backlog 재잠금 없음 + 승격은 제때와 같이 남음」을 단언한다(tasks 2.9).
  「실제 삭제 때만」 변형이 만드는 결함은 **빈 backlog 재잠금**이다.

## §0.10 proposal-freeze 9회차 (2026-09-26) — design 6판 2쇄 **REJECT** → 6판 3쇄

- codex 9회차(`codex-r9-prompt.md` · `codex-r9-output.md`). P0 없음. 새 발견:

| id | 심각도 | 발견 | 판정 → 6판 3쇄 |
|---|---|---|---|
| Z1 | P1 | 계수 (2, e0) → 한도째 오류 → 해제 → 늦은 세대 읽기 → 리셋 → 판정 없음. 제때라면 승격이 남았다 | **수용** → 한도째에서는 세대로 리셋하지 않음, 승격 항상, 차단은 직전 증가 때 세대로 |
| Z2 | P1 | 차단이 원칙 E 로 거절되고 승격 쓰기도 실패하면 아무것도 남지 않는다(행은 이미 전달돼 기동 복원도 못 덮음) | **수용** → 그 조합이면 무조건 차단 |
| Z3 | P1 | proposal(시각 비교 잔재) · D1(승격이 울타리에 묶임) · spec(실패 기록 `NotFound` 에도 승격) 불일치 | **수용** → D7 「판정 → 행동 표」 하나에 모두 정렬 |
| Z4 | P2 | 가산 읽기 실패의 원자성 시험 의무가 없다 | **수용** → 2.7 |
| Z5 | P2 | 2.6 수락선이 유도되지 않았다 | **수용** → exit 사이클 원장 연산 수 × 실행자 트랜잭션 최대 길이 |

## §0.11 proposal-freeze 10회차 (2026-09-26) — design 6판 3쇄 **REJECT** → 6판 4쇄

- codex 10회차(`codex-r10-prompt.md` · `codex-r10-output.md`). P0 없음. 새 발견:

| id | 심각도 | 발견 | 판정 → 6판 4쇄 |
|---|---|---|---|
| AA1 | P1 | 조건부 차단 성공 → 해제 → 승격 쓰기 실패면 차단도 모드도 남지 않는다(Z2 의 대체가 `!blocked` 에만 걸려 있었다) | **수용** → 승격 쓰기 실패면 조건부 차단 결과와 무관하게 무조건 차단 |
| AA2 | P1 | D8 의 한도 판정과 리셋 규칙 · 시험 2.10 기대값이 서로 모순 | **수용** → 「오류 하나의 전이」 표(한도 판정이 리셋보다 먼저), 기대값 세 줄로 정정 |
| AA3 | P1 | 2.6 수락선이 실행자 트랜잭션이 느려질수록 커진다(자기 조정) — a098 시험은 바로 그 이유로 고정 여유를 쓴다(`a098_…_test.go:61-72`) | **수용** → 고정 여유 `a098ExitCycleDwellMargin` 인용 + 여유를 넘는 주입에서 빨강 확인 |
| AA4 | P2 | `PendingAlerts` 에 인자를 더하면 obs 호출자(`notifier.go:737·855`)를 고쳐야 한다 — 「obs 무편집」과 충돌 | **수용** → 새 메서드 `PendingAlertsForDelivery`, `PendingAlerts` 불변 |

## §0.12 proposal-freeze 11회차 (2026-09-26) — design 6판 4쇄 **REJECT** → 6판 5쇄

- codex 11회차(`codex-r11-prompt.md` · `codex-r11-output.md`). P0 없음, AA1~AA4 해소 확인. 새 발견:

| id | 심각도 | 발견 | 판정 → 6판 5쇄 |
|---|---|---|---|
| AB1 | P1 | 식 정렬 `(attempts >= ?), id` 는 `idx_outbox_state(state, id)` 로 못 풀어 PENDING 전부를 정렬한다 — 연결 점유가 배치가 아니라 backlog 크기에 비례, 손절 경로와 연결 공유. 기존 시험이 못 잡는다 | **수용** → D4 비용 모형(O(P log P)), 2.6 (c) P = 10 · 1 000 · 10 000 측정(고정 여유), 넘으면 가산 색인 = 스키마 변경이라 멈추고 Manager 에게. 색인은 미리 넣지 않는다(측정 먼저) |
| AB2 | P2 | 원칙 E 는 보수적 근사다 — 시나리오가 허용 예외(세대 읽기 창 · 승격 실패 무조건 차단)를 적지 않았다 | **수용** → D7 「보수적 근사」, 시나리오 세 곳에 예외, 2.9 에 두 순서 시험 |

## §0.13 proposal-freeze 12회차 (2026-09-26) — **실행 불가 (교차 모델 인증 실패)** · 판정 없음

- codex 12회차(`codex-r12-prompt.md`, 6판 5쇄 대상)를 두 번 실행했고 두 번 모두 `401 Unauthorized: Incorrect API key provided`(codex-cli 0.154.0, 재연결 5회 ×2 뒤)로
  끝났다. 출력 없음. 자격 증명은 사람 몫이라 건드리지 않았다(안전 불변식 8 · 시크릿).
- Teammate 적대 Eng 패스(6판 5쇄): 새 P0/P1 없음. 남는 것은 전부 측정 의무로 적혀 있다 — 2.6 (b) 원장 경합(고정 여유), 2.6 (c) backlog 크기(넘으면 스키마 변경
  = Manager), 4.4 래치 시간. 11회차까지의 발견(F~AB)은 처분표와 이 기록에 모두 있다.
- **판정: 없음.** 두 보이스 중 교차 모델이 없으므로 PASS 로 적지 않는다. tasks 0.4 는 체크하지 않는다. codex 인증이 복구되면 `codex-r12-prompt.md` 그대로 재실행.
