# 목소리 B 출력(리뷰 대상 80ae96a5) — 하위 에이전트 최종 보고 원문 전사

**판정: APPROVE** (P0·P1 없음. 아래 P2는 착지를 막지 않음)

금지 사항 위반 없음. `~/.codex`, `~/.config/tossctl`, 자격 증명은 읽지도 검색하지도 않았다. 실제 저장소는 읽기만 했고 `git status`는 시작 때와 같다(미추적 3개 그대로).
사본은 하나(`/tmp/claude-1000/a112-rev5222-ywEg`, `.git` 없음)를 만들어 모든 셸 블록에서 다시 썼다. 블록마다 cwd 가드를 걸었다. 실험이 끝날 때마다 `internal/`이 리뷰 트리와 같은지 `diff -rq`로 확인했고(IDENTICAL), 끝에 사본을 지웠다(남은 rev5222 사본 0개).
변이 재현 드라이버는 세션 스크래치패드(`…/scratchpad/rev_mut.py`)에 있다.

## 발견 표

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | P2 | M20이 살아남는 이유는 오늘 코드에서는 참이다. 그런데 그 이유를 시험이 고정하지 않았고, 짧은 시험 하나로 M20을 죽일 수 있다. | 세대는 `manifest.Generation`에서 온다(`production_snapshot_authority.go:172`). 매니페스트 바이트는 적재기마다 고정된 시장 digest 하나로 묶인다(`:154-156`). 같은 위험 쌍이 1차 레그와 dispatch 양쪽에 배선된다(`strategy_entry_supervisor.go:328-337`). 사본 실험: KR 시장의 `.bundle` 칸에만 세대 2 번들을 넣고, 범위 번들은 세대 1로 두었다. 원본은 PASS, M20 변이는 FAIL(`lease risk_policy_generation=2, want 1`). 덧붙여 journal은 lease의 위험 세대를 admission 번들과 대조하지 않고 그대로 받는다(편집 전부터 있던 동작). | 이 시험(`strategy_dispatch_market_authorities.risk_policy_generation` == 범위 번들 세대)을 커밋해 M20을 CAUGHT로 바꾼다. |
| 2 | P2 | review가 자기 RED 영수증을 잘못 서술한다. | review는 「편집 전 4 FAIL — 개수 관문 · 오늘-동등성 핀 · Done · census」라고 쓴다. 실제 `red-5.2.2.2.log`의 4 FAIL은 Done · SecondLeg · AScopeWithoutAccount · Currency다. 또 RED 문구 `paired production authority is incomplete for market`는 B2 문구와 똑같아서(편집 전 `:222`과 `:232`) 개수 관문 때문이라고 특정할 수 없다. | review 문장을 고친다. |
| 3 | P2 | worker BTM B3의 「편집 전 활성화 두 범위 시장 dormant(FAIL 관측)」에는 영수증이 없다. | RED 로그에 `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope`가 없다. 근거가 될 수 있는 것은 같은 동작을 되돌린 변이 X16(CAUGHT, worker `Effective:false`)뿐이다. | 「X16 등가 변이로 관측」으로 고치거나 RED 로그를 첨부한다. |
| 4 | P2 | 전수표 (a)의 「범위 번들 digest 둘」은 「각 범위가 자기 번들로 발급됐다」를 재지 않는다. | `a112_owner_scope_trading_test.go:327-335`는 적재기 출력에서 준비된 번들 digest의 개수만 센다. 그래서 X08에서도 2가 나온다. 실제로 막는 것은 하류의 `production risk authority scope changed` 대조다(X08이 이 대조로 CAUGHT). | 주석(326)과 전수표 문구를 정정한다. 발급된 레그와 번들 digest를 묶어 단언하는 방식도 있다. |
| 5 | P2 | (b) 시험은 이름이 「OfOneCycle」인데 실제로는 둘째 파도에서 잰다. 둘째 파도로 간 이유도 측정과 어긋나고, 단언도 약하다. | 사본 측정, 상한 1200: 첫 파도(한 주기)에서 이미 `OPEN_EXPOSURE would reach 1602 … (snapshot 0, already held 801, new 801)`로 거절된다. 합산 검사가 stale CAS보다 먼저 돈다. 대조: 1200 상한에서 레그 하나(801)만 내면 들어간다. 그러니 성질 자체는 진짜다. 단언(369줄)은 문자열에 "exposure"만 있으면 통과해서 Guardian의 `OPEN_EXPOSURE_EXCEEDED`도 받아 준다. | 첫 파도 오류를 단언하고, `already held`를 확인한다. |
| 6 | (T) | J2 실측의 「reads」는 적재 호출 수가 아니라 `len(scopes)`다. | 핀을 35로 바꾼 실행에서 위험 적재가 전부 실패했는데도 `risk=2 + account=2`가 로그에 찍혔다. | 「범위 항목 수」로 표기한다. |
| 7 | (T) | X06은 4건 중 3건이 nil 포인터를 `%w`로 감싸 문구가 `<nil>`로 바뀐 부작용으로 빨개졌다. | 의도한 성질로 잡은 것은 `TestAForgedScope`(placed=[000660]) 1건이다. | nil이 아닌 값을 감싸는 변이를 추가한다. |
| 8 | (T) | X19 시험 주석과 실측 상태가 다르다. | 주석(453-454줄)은 「EvidenceStale」, 실측은 status `UNKNOWN`. | 주석을 고친다. |
| 9 | (T) | `isolated-verify.log`는 rtk 요약이고 대상도 스크래치 커밋 eb99c48a다. check-analysis 델타 서술도 파일 내용과 다르다. | 80ae96a5 트리로 다시 쟀다. 태그 1190 pass / 2 fail / 1 skip, 무태그 1025 / 2 / 1. 실패 2건(`TestA111ObserverUses…`, `TestA111FallbackSequence…`)은 `-trimpath` 때문이고, 플래그 없이 돌리면 `ok`다. 델타 파일에 실제로 있는 것은 required 176→182와 commit 518→519뿐이다. | 원 출력을 영수증으로 남기고 델타 서술을 정정한다. |
| 10 | (T) | 하네스 대조군은 pass 사건이 0보다 크기만 보고, 기대 개수와 같은지는 보지 않는다. JSON 종료 코드 검사는 `run_tests`(0이 아니면 RED)와 겹친다. | `a112_lot_mutate.py:449-460` | 기대 pass 수를 고정한다. |

## 필수 항목별 판정

**1. 거울 다리 — 통과.** 행을 지어내지 않는다(INSERT 원천은 원장에서 읽은 행뿐, `:151-161`). 복사기 단언은 stub 유도 열 전체로 ORDER BY 한 다중집합 동일성(NULL 은 nil 끼리, 친화성 강제 변환은 Scan 타입 차이로 FAIL) — 같아 보이는 다른 행의 위양성 없음; 단 왕복만 재고 적재기가 읽는 것을 덮는지는 재지 않음. 적재기가 읽는 표 7개 · 열(`production_snapshot_authority.go:371, 461-476`)은 전부 stub 스키마 안. 사본에서 핀을 35로 바꾸고 적재기를 실원장에 직접 붙여 다리 없이 돌린 결과 Done · SecondLeg · AScopeWithout · AForged 4개 전부 PASS — 다리는 원장이 허락하지 않는 발급을 허락하지 않는다. 트립와이어는 핀 35 에서 실제로 FAIL(`real-journal load err=<nil>`), 다리 쓰는 시험 5개도 함께 FAIL.

**2. M20 구조 동등 — 논증은 참, 시험 공백 있음.** 시장 하나는 한 번의 collect 안에서 digest 고정 매니페스트 하나만 쓰고 같은 쌍을 1차 레그 B5 가 먼저 걸러 준다 — 오늘 생산에서 동등. 고정하는 시험이 없고 사본 실험 시험으로 닫힌다(#1).

**3. 변이 원장 — 진실.** 80ae96a5 트리에서 22개 전부 재실행: 21 CAUGHT · M20 SURVIVED, 실패 시험 이름 집합도 원장과 같음(9cbc7560..7ab8cd12 사이 `internal/` diff 비어 있음). 원인별: X01·X02·X03·X04·X05·X07·X10·X11·X15·X16·X17·X18·X21·X22 는 겨냥한 성질로, X08 은 하류 범위 대조와 거절 타입 단언으로, X09·X14 는 fixture allowlist(SYMBOL_NOT_ALLOWED)로, X12·X13 은 범위 수 단언으로 — 모두 정당. X06 은 #7. 대조군 재현: 종료 0, pass 사건 54 / 22. X16·X19 는 공유 슬라이스 오염 없음.

**4. 문서 · 번들 — 대체로 정직, 부정확한 곳 있음(#2~#6, #9).** 표본 번들 6개 Source SHA-256 일치. admit 편집 전 번들의 base 복사 정당(016da624..e8d56d49 admit 파일 변경 0). 「진입 0」(위험 B5 · 계좌 B5 · 1차 레그 B7)은 seam 으로 못 만듦 확인. census 4 → 2 와 사유 정확(X11 · X14 가 3을 만들어 RED). 범위 거절 census 는 함수 단위 개수만 세서 같은 함수 안 자리 바꿔치기는 못 봄 — X22 행동 시험이 막음.

Recommendation: APPROVE. 착지 전이나 다음 로트에서 P2 #1(M20을 죽이는 시험 커밋), #4·#5(단언 강화와 문구 정정), #2·#3·#9(review · BTM 서술 정정)를 처리할 것을 권고한다. 생산 동작 변화나 안전 경계가 뚫린 곳은 찾지 못했다.

사본: `/tmp/claude-1000/a112-rev5222-ywEg` — 만들었고, 지웠다.
