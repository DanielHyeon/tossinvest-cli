# a205 review

## Pre-Edit Gate (2026-11-27, 팀메이트 Opus)

```text
Pre-Edit Gate:
- change id / task id: a205-exit-observation-2s / 2.2 · 2.3
- 대상 심볼(패키지.함수):
    생산(상수 · 주석만, 함수 본문 0): engine.DefaultExitObservationInterval(5s→2s) · engine.DefaultExitObservationOutage 주석 ·
      engine.externalBuyListTimeout 주석 · console.lineRefreshInterval(5s→2s) · console.engineRuntimeInterval(5s→2s)
    시험(기존 함수 본문 편집): console.TestTheRedrawCadenceIsTheEnginesAndNotTheCaches · console.TestTheCacheIntervalsAreTheirOwnConstants ·
      (r1 #1 뒤 추가) console.TestTheOverviewSpendsNothingHoweverOftenItReloads
    시험(새 파일): engine.TestA205DefaultExitObservationIntervalIsTwoSeconds · engine.TestA205LowerLimitRefreshBudgetAtTheDefaultInterval
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md(1.6.0, base e2921fc00) — 각 상수의 호출자 1,
    `Interval()` 의 Tracer 호출은 CodeGraph 누락 → AST 로 보충
- CodeGraphContext 후보와 evidence reconciliation: not-applicable(사유 codegraphcontext-context.md) · 미해소 불일치 0(evidence-reconciliation.md)
- 기존 동작 파악 근거: base AST 7(analysis/ast-evidence/claims.md C1~C4) · 생산 배선 cmd/tossctl/engine.go:675-685(Interval 미전달) ·
    사전 탐침(상수만 2s, 엔진 패키지 시험 rc 0) · 콘솔 import 금지 static_test.go:598
- Function Logic Map / Branch Test Map: analysis/function-logic/internal-console--testtheredrawcadenceistheenginesandnotthecaches/ ·
    analysis/function-logic/internal-console--testthecacheintervalsaretheirownconstants/ ·
    analysis/function-logic/internal-console--testtheoverviewspendsnothinghoweveroftenitreloads/(r1 #1 뒤, 편집 전 scaffold · 편집 후 AST)
    (생산 함수 본문 편집 0 — 번들 불요, check_analysis required = 위 시험 함수 3)
- upstream 상속 테스트 영향: no — upstream/main 에 internal/app/engine 없음(design D6), 콘솔 두 상수는 TossOS 전용
- 실패 테스트 선행 작성: yes — analysis/verify/red.txt
- 설정·DB·journal 변경과 rollback: 없음(스키마 · 설정 키 · 토글 0). 런타임 쓰기 빈도만 2.5배(design §5). rollback = 직전 이미지로 엔진 교체
- 안전 불변식 §0 위반 여부 검토: 통과 — §0.3 손절 즉시성: 판정 주기가 빨라지고 하한가 갱신기는 판정 goroutine 밖(C2) ·
    §0.4 공식 API 호출 추가 없음(같은 호출의 빈도 증가 — design §5 에 계상) · §0.9 손절 · 익절 규칙 무변경, 관측만 잦아짐(보수 방향) ·
    §0.2 토글 아님, 엔진 없는 upstream 과 동일 · §0.7 운영 토글 flip 0
```

## proposal-freeze 리뷰

이 change 는 상수 3개 · 주석 2곳 · 시험 3개라 proposal-freeze 리뷰를 구현 리뷰와 **합쳐** 한 번에 받음(사유: 설계 결정이 diff 와 같은 크기이고,
리뷰 회차 상한 2 — Manager 지시). 리뷰어에게 proposal · design · spec delta · diff 를 함께 주고 설계 결정 D1~D6 을 적대적으로 보게 함. 아래 「구현 리뷰」 참조.

## 구현 리뷰 r1 (2026-11-27, Codex codex-cli 0.165.0 — 작성자와 다른 계열, 적대적 Eng · proposal-freeze 겸)

- 구성: Codex 1 보이스, `codex exec -s read-only`, 입력 = proposal · design · spec delta · tasks · claims · diff(`git diff` + 새 시험 파일). 리뷰어 실행 전후 `git status --porcelain` 동일(쓰기 0).
- 판정: **FAIL** — P0/P1(런타임 안전 결함) 0, P2 blocker 5. 리뷰어 원문 요지는 아래 표, 7 항목 판정 ①~⑦ 포함.

| # | 지적 | 처분 | 근거 |
|---|---|---|---|
| 1 | [P2] `TestTheOverviewSpendsNothingHoweverOftenItReloads` 가 2s 에서 반드시 실패(10회 × 2s = 20s ≤ TTL 30s) | **수용** — `redraws` 10 → 16(32s). 조건(창 > TTL)은 그대로 | 첫 GREEN 실행(`analysis/verify/green-attempt1.txt`)이 같은 실패를 냄 — 리뷰가 먼저 짚음. 새 FLM 번들 |
| 2 | [P2] D2 「겹침 0」 논증 불성립(시한은 토큰 수신 뒤 시작), 새 시험의 `prefetch+read ≤ interval` 은 항등식 | **수용** — design D2 를 「예산 합 = 주기」와 「비중첩」으로 분리, (가)·(다)의 「지킴」 철회, 항등식 단언 삭제 · 시험 이름 `…BudgetAtTheDefaultInterval` 로 바꾸고 「비중첩 증명 아님」을 주석에. 비중첩 강제는 새 가드 → **사람 결정**(R1b) | `refresh` :191 · :221, `Run` :434 · :435 |
| 3 | [P2] 「재시도가 잦아 첫 성공까지 기대 시간이 늘지 않음」 · 「약화 아님」은 미입증. 생산 admission=nil | **수용** — 미입증으로 정정, 「오늘의 LIMIT ≠ 하한가 보호」 명시, 생산 dormant(`engine.go:627`) 사실 추가 → 위험은 활성화 뒤 조건부(R1). 예산 재배분 (다)는 **사람 결정** | `engine.go:627`, a087 order-execution delta :34 |
| 4 | [P2] a197 3s 를 단계 전체 지연 상한처럼 설명 | **수용** — design D3 정정(3s 는 목록 판독만, 단계 전체 시한 없음 — 같은 파일 :159). 코드 주석은 「이 단계가 도는 주기만 다음 주기 시작이 최대 이만큼 밀림」→ 단계 체류만큼으로 고침 | `a197_external_opposite_buy.go:159-160` |
| 5 | [P2] 부하 표를 안전 상한으로 읽을 수 있음 | **수용** — §5 에 「평균 추산, 안전 상한 아님」과 표 밖 호출(a197 5쪽 · 취소 · 401 재전송 · 다른 루프 · 다른 프로세스 · 버스트) 명시, K≥4 경계 삭제. 감쇠 · admission 은 **사람 결정**(R2) | `client.go:325`, `a197…:260` |
| ⑦ | D6 「엔진이 꺼져 있으면 변화 없음」 부정확 — 콘솔 재로드는 엔진과 무관하게 바뀜 | **수용** — D6 · proposal 정정(upstream 에 콘솔도 없음 실측 0) | `git ls-tree upstream/main` |

- 새 가드 추가: 0(Manager 지시). 사람 결정으로 넘긴 것: R1(시한 재배분 · `/price-limits` 지연 측정), R1b(비중첩 강제), R2(재제출 감쇠 · admission).
- `analysis/verify/red.txt` 의 `TestA205LowerLimitRefreshStillFitsInsideOneInterval` 는 r1 전 이름(같은 시험 — RED 원문은 고치지 않음).

## 구현 리뷰 r2 (2026-11-27, Codex — 마지막 회차)

- 판정: **PASS** — r1 #1~#5 · ⑦ 처분 전부 확인, 잔여 blocker 0, 새 P0/P1 0. 리뷰어 실행 전후 `git status --porcelain` 동일.
- P3 두 건 **수용**: (1) 새 시험 주석이 「실제로 자는 주기」를 고정한다고 과장 → 「상수와 `Interval()` 반환값」으로 축소, Run · Sleep 미실행 명시.
  (2) proposal 의 「시험 함수 본문 2개」 → 3개.
- 리뷰 회차 상한 2 도달 — 추가 회차 없음. 사람 결정으로 넘긴 것: R1 · R1b · R2(design 「남은 위험」).

## 검증 (영향 범위 — 전체 시험 · `make gate` 는 요청 없음, 안 돌림)

| 명령 | 결과 | 원문 |
|---|---|---|
| `go test -count=1 -run A205 ./internal/app/engine/` + 콘솔(RED, 상수 5s) | rc 1 — 새 시험 2 · 콘솔 5 실패 | `analysis/verify/red.txt` |
| `go test -count=1 ./internal/app/engine/... ./internal/console/... ./cmd/tossctl/`(첫 GREEN) | rc 1 — engine 10m 기본 시한 초과(동시 부하 · 단독 406.7s), console `TestTheOverviewSpendsNothing…` 실패(r1 #1), **cmd/tossctl ok 71.4s** | `analysis/verify/green-attempt1.txt` |
| `go test -timeout 30m -count=1 ./internal/app/engine/... ./internal/console/...`(r1 처분 뒤) | **rc 0** — engine 429.8s · console 75.8s | `analysis/verify/green.txt` |
| 대상 시험 `-v`(A205 2 · 콘솔 3) | rc 0, 5 PASS | `analysis/verify/targeted-v.txt` |
| `make lint`(gofmt · vet · 태그 vet) | rc 0 | `analysis/verify/lint.txt` |
| `make test-race` 둘째 줄(엔진 목록) + `-race -run 'A087|A205|LowerLimit|ExitObserver'` | rc 0 · rc 0, DATA RACE 0 | `analysis/verify/race.txt` |
| `go test -tags tossos_testseams ./internal/app/engine/ ./internal/console/`(test-seams 중 영향 패키지) | rc 0 — 572.4s · 100.6s | `analysis/verify/seams.txt` |
| `check_analysis.py --change` · `openspec validate --strict` · PM `--check` | rc 0 · rc 0 · rc 0(required 3, 증거 완전) | `analysis/verify/sdd-checks.txt` |

- 실행 순서 주의: green.txt 뒤에 바뀐 것은 주석뿐(a197 주석 한 줄 · 새 시험 주석 — r2 P3). seams · race(둘째 명령) · sdd-checks 는 최종 트리에서 돎.
- 안 돌린 것: 저장소 전체 `make test` · `make test-seams` · `make test-race` 첫 줄(전략 7 패키지 — 이 change 와 무관), `make sdd-check`(이 worktree 에 CodeGraph 색인 없음), `make gate`.
  `cmd/tossctl` 은 첫 GREEN 의 통과 결과(생산 상수는 최종과 같음, 그 뒤 생산 코드는 주석만 바뀜).
