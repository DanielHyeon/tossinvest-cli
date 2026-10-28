# a165 issues

## 사용자 결정 (2026-10-23, 「1~4번 권장대로 진행」) — proposal 열린 질문

정의 단계(2026-10-22, `e66152465`)에서 남긴 질문에 사용자가 권장안대로 답했다. 코드·실계좌 호출·`mutating: true` 명령은 없었다.
이 기록은 착수(design·구현)가 아니다.

| # | 질문 | 결정 | 남는 것 |
|---|---|---|---|
| 1 | 결정 1 의 「주문 접수」에 조건주문 생성을 넣는가 | **제외** — 이중 조건주문은 이중 체결 위험. 조건주문 생성·정정·취소 재시도 0회 유지 | 없음 |
| 2 | (b) 조회 후 재전송을 8s 창 밖에서도 하는가 | **8s 창 안에서만.** 창 밖은 재전송 없이 IN_DOUBT 해소 절차로 | 창 경계(문서상 10분) 미검증은 그대로 — 8s 는 그 경계에서 멀다(proposal 「2.7 과의 관계」) |
| 3 | 기존 401 재전송(`client.go:344` 루프·:354 `doRequest`)을 「확실히 미접수」로 인정하는가 | **인정.** 근거는 a165 에서 시험으로 고정(FLM 열거 뒤 — 루프 조건·횟수·같은 키) | **401 을 받은 POST 가 미접수라는 실측 근거 없음** — 잔여 |
| 4 | 취소·정정 409 재시도(1.3 결정 2)를 a165 에 두는가 | **a165 범위** | 없음 |

반영: `proposal.md` 「사람 결정 (2026-10-23)」, `specs/order-execution/spec.md`(초안 delta), `tasks.md` 0.1.

## 착수 기록 (2026-10-23, Terra) — 0.2~1.3

- 0.2 base `34e12967e`. 단일 writer = Terra(이 세션), worktree `/tmp/tossos-a165-20261023`(브랜치 `a165-impl` → origin `feat/a112-four-family-runtime`).
- **sdd-sync 가 worktree 를 fast-forward 함**: `make sdd-sync` 의 GBrain 단계 뒤 브랜치 reflog 에 `pull --no-recurse-submodules --ff-only: Fast-forward`
  (34e12967e → 5975af2b9)가 남았다. 끌려온 4 커밋은 a164·a140 문서뿐(`.go` 0). base 는 34e12967e 그대로. 같은 도구를 쓰는 다른 worktree 도 같은 일이 생길 수 있다 —
  측정 전 `git log -1` 로 HEAD 를 확인할 것.
- GBrain sync 는 12,708 파일 parse 실패로 BLOCKED(advisory). CodeGraphContext 색인은 다른 체크아웃을 가리킨다(advisory 로도 못 씀).
- 편집 전 기준선 시험: `go test -count=1 -coverprofile ./internal/execgw/` rc=0(64.0s, 73.2%), `./internal/official/` rc=0(1.0s, 85.5%).
  `./internal/journal/` 전체는 coverage 아래 10분 기본 timeout 초과(rc=1, `TestRefuseClaimedStrategyDispatchPreTransportRejectsCrossScopePairedKRUS` 실행 중 — 느림, 실패 아님으로 보이나 미확인),
  `-run 'Classify|SendTracker|Dispatch'` 는 rc=0(120.7s).
- 1.3 FAIL(blocker 3) — `review.md`. **§2 미착수.** Manager 지시 「재시도가 손절/flatten 지연을 늘리면 멈추고 보고」에 해당(design D7).

## 잔여 (이름 붙임)

- R-a165-1 **401 을 받은 POST 가 실제 미접수라는 실측 없음**(사람 결정 3).
- R-a165-2 **429 를 받은 POST 가 실제 미접수라는 것은 사용자 결정의 전제**(확정표 §2.1).
- R-a165-3 **실패 응답(5xx·429·전송 후 끊김) 뒤 같은 `clientOrderId` 재요청의 멱등 동작 미측정** — R11(M14·M35)은 2xx 직후 재요청만 쟀고 openapi 는 실패 경로를 말하지 않는다(1.3 A1).
- R-a165-4 HTTP/2 transport 내부 재전송(REFUSED_STREAM·GOAWAY) 가능성 — 「attempt 당 1회」는 `plan.call` 기준이다(1.3 A11).
- 별건(a165 밖): `strategyDispatchOwnerTakeoverGrace` 20s 가 「15초 클라이언트 timeout」을 전제(strategy_dispatch_runtime.go:664–667)하나 401 루프·timeout 없는 토큰 교환으로 그 전제가 이미 깨져 있다(1.3 A7). 담당 change 미정.

## 사용자 결정 (2026-10-23 2차, Manager 전달 — review.md 「사람 결정 요청」 1~7 전부 권장안)

| # | 결정 | a165 에 남는 것 |
|---|---|---|
| 1 | 엔진 청산 루프의 청산 전 취소가 `409 already-processing` → **D**: 같은 주기 안 재시도 없음, 「확정 미수락」 분류, 다음 주기 재취소(뒤 포지션 지연 0) | 취소 409 `already-processing`(code 일치) → FAILED_CONFIRMED 분류 |
| 2 | CLI `tossctl flatten` 취소에 409 재시도 배선 — 예 | flatten 배선에만 취소 409 재시도(누적 5s·2회) |
| 3 | 청산 매도 접수 재시도 0회 유지 | 노출 감소 place 재시도 없음 |
| 4 | 진입 재시도에서 「전송 전 실패」 갈래 제외 | (a) 전송 전 실패 갈래 없음 |
| 5 | 429 갈래는 a166 착지 전까지 보류(a165 범위에서 뺌), (b) 조회는 OPEN·CLOSED 각 1페이지 고정 | (b) 만, 조회 1페이지씩 |
| 6 | Amend 범위 제외 | 정정 재시도·분류 변경 없음 |
| 7 | R-a165-3 잔여 수락, 단 A1 수리 필수(정착 전 목록 대조 — 일치 정확히 1건·같은 id 일 때만 CONFIRMED) | 재전송 ACK 뒤 목록 대조 |

남는 범위: 진입 (b) 8s 창 조회-후-재전송(+A1), CLI flatten 409 취소 재시도, 엔진 취소 409 「확정 미수락」 분류, 401 재전송 「확실히 미접수」 시험 고정.
범위 밖이 된 1.3 발견: A3(429 패널티 — 429 갈래 보류로 소멸), A5(Amend 제외), S3 의 전송 전 실패 비관 전이(갈래 제외로 소멸), D7 ②③ 청산 place(0회 유지로 소멸).

## 사용자 결정 (2026-10-23 3차, Manager 전달) — 1.3 재리뷰 F3

- **F3 = 1(권고안)**: R1·R2(진입 결과 불명 재전송·정착 전 목록 대조)를 a165 에서 뺀다. a165 = **R3**(엔진 청산 루프 취소 409 → 같은 주기 재시도 없이 확정 미수락, 다음 주기 재취소) +
  **R4**(CLI flatten 취소 409 재시도, 소진 시 오늘 경로 IN_DOUBT → `resolveStep` — F1 수리) + **R5**(401 재전송 근거 시험).
- 진입 재전송은 전략 claimed 경로와 함께 **a171-entry-place-ambiguous-resend**(정의만, 커밋 `566860cfe` — Story STORY-TOS-a171)로. 미해소 F2·F4·F6·A2·F8 과 잔여
  R-a165-1~4 를 a171 proposal 에 이관 기록했다. a165 에 남는 잔여는 R-a165-1(401 POST 미접수 실측 없음 — R5 가 시험으로만 고정)뿐이다.
- 범위 축소에 따라 a165 의 FLM 번들 10개(조회 경로·place·Amend·strategy 게이트 등)를 지우고, 남는 범위가 기대는 10개(New·buildGateway·buildFlattenWiring·
  resolveStep·planLiquidation·UnsettledCancelSymbols·classifyRefusalCode·PlaceOrder·CancelOrder·postWithHeaders)를 더했다(합계 23).
- `check_analysis.py --change a165` 가 2026-10-23 오후부터 `missing evidence for modified function internal/official/stock_reads.go:adaptStocks` 를 낸다 — a165 base 뒤 착지한
  형제 커밋(a169 `064f67297`, `security_type` 노출)이 고친 함수가 워킹트리 창에 들어온 것이다(a165 자기 편집 아님). GREEN 뒤 `--record-landing` 으로 창을 좁히거나,
  안 되면 WORKFLOW 「사람 승인 base 재고정」 절차가 필요하다.

## 사용자 결정 (2026-10-28, 「권장대로」) — 1.3 3차 리뷰 전 H1·H2

| # | 결정 | a165 에 남는 것 |
|---|---|---|
| H1 | R4(CLI flatten 취소 409 재시도)의 최악 지연 — 막힌 취소 1건마다 오늘 경로 앞에 최대 ≤8s, flatten 전체로는 N×8s 가능 — 를 3차 결정대로 **유지**한다. 3차 리뷰는 불변식 4 판정을 명시해 남긴다 | design D3·D4 그대로. 판정은 review.md 「1.3 3차」 불변식 4 |
| H2 | 형제 a169 `064f67297` 의 `adaptStocks` 가 a165 gate ⑤ 창에 들어온 문제 → GREEN 뒤 `--record-landing` 을 먼저 시도하고, 안 되면 archive 번들 복사로 처리한다. **base 재고정은 하지 않는다** | 아래 「H2 실측」 |

### H2 실측 (1.3 3차, 2026-10-28, HEAD `0c9bfd272`)

- `check_analysis.py --change a165-safe-order-placement-retry` rc=1, `missing evidence` **29건**(2026-10-23 의 1건에서 늘었다): `internal/console/*` 23 · `internal/campaignstateevidence` 2 ·
  `tools/a112-mb-us-source` 3 · `internal/official/stock_reads.go:adaptStocks` 1. 전부 base `34e12967e` 뒤 착지한 형제 커밋의 함수이고 a165 자기 편집은 0 이다.
- `--record-landing` 은 창의 **끝**만 좁힌다(WORKFLOW 「착지 지점」, README). 형제 커밋은 base 와 a165 착지 사이에 있으므로 기록해도 창에 남는다 — **H2 의 1차 경로로는 이 29건이 안 풀릴 것으로 본다**
  (예상, 도구로 확인은 GREEN 뒤). 그러면 2차 경로(아카이브 번들 복사, a170 선례 `10cbb15c0`)이고 건수는 a165 GREEN 까지 형제 착지마다 는다.
  복사본은 HEAD 에서 `ast.json` 을 다시 뽑아 같은지 볼 것(a170·a172 교훈). `adaptStocks` 번들은 아카이브가 아니라 진행 중 change 디렉터리(`a169-breakout-universes-are-compared-in-shadow`)에 있다.
- base 재고정은 하지 않는다(H2). 비용을 줄이는 길은 a165 를 빨리 닫는 것뿐이다.

