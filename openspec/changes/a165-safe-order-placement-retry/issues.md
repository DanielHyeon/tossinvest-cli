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
