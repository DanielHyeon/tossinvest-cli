# 8.5 voice 3 (evidence quality) — SubagentHandback 원문 (agent a2f009eee4efbb58e, 저장 2026-10-04)

# a112 8.5 독립 적대 리뷰 — 보이스 ③ (증거 · 시험 품질)

~/.codex 아래는 아무것도 읽지 않았다 (사본의 저장소-추적 `.codex/` 디렉터리는 이름만 보였고 내용을 열지 않음). 실제 저장소는 읽기 전용으로만 썼다 — 모든 변이는 `git -C <repo> archive <commit> | tar -x` 사본(scratchpad/voice3/{A2,B,B2,B3,P})에서, 복원은 sha256 단언으로 확인.

## 판정: HOLD

생산 코드 자체의 결함은 못 찾았다(수락 집합 1,129행 전수 차분 pre/post 동일). HOLD 사유는 이 로트가 파는 것 — "sentinel 보존 · 결속 완전성을 시험이 못 박는다" — 에 구멍 둘(P1)이 실측으로 뚫렸기 때문이다. 둘 다 시험 한 줄/한 모양으로 닫힌다.

## 내가 돌린 변이 표 (총 72건; 대조군 전부 GREEN 확인 후 판정)

대조군(무변이): A2 breakoutlane `ok 0.22s` · B2 strategyrouter `ok 1.56s` · B 엔진 full(testseams) `ok 506s, 1363 pass` · B cmd/tossctl `ok 71.9s`. GOFLAGS 비움(−trimpath 는 runtime.Caller 기반 guard_test·a111 을 깨서 가짜 FAIL — 실측 후 제외), 전용 GOCACHE.

### 대상 B1 — internal/strategyrouter/production_family_activation.go (사본 B2, 스위트 = strategyrouter 전체)
| 변이 | 결과 | 잡은 시험 |
|---|---|---|
| T451~T459, T544~T554, T564, T567, T568, T570, T597~T601 (fieldCheck 항별 "영원히 안 보고" 26종) | 전부 CAUGHT | a112 error-fields 표 + 기존 결속/수명/서술자 시험들 |
| R5/R6/R7 (desired·family·lane_version 항 삭제), R16(horizon), R9/R12/R13 (복합 메시지에 Undeclared 가시 누출), R14(읽기 결함 누출), B2/B3 동형 재현 | CAUGHT | 원장 13/13과 정합 — 저자 원장 재현 성공 |
| **R4 = T602: `effective` 유효성 항 삭제/무력화** | **SURVIVED (패키지 전체)** | 없음 — 프로브 시험이 `Effective:"MAYBE", Desired:ON` 매니페스트의 **verified=true 수락**을 실측 (`log-R4.txt`: err=nil, verified=true) |
| **L229,L234,L438,L471*,L510,L517,L520,L524,L556,L572,L604,L609,L613,L618 + R10,R11,R15c: `%.0w` 로 Undeclared 를 메시지 불변·보이지 않게 사슬에 추가** | **router 스위트에서 L386,L461,L468,L482 넷만 CAUGHT**(TestOnlyAnEmptyPin…), 나머지 13곳 SURVIVED | — |
| L471(digest 불일치), L556(몸통 결속), L438(ctx nil) → **엔진 full(testseams)+cmd/tossctl 로 승격 재판** | L471 CAUGHT(TestADeclaredActivationThatLapses…/핀_+_바뀐_바이트) · **L556 SURVIVED** · L438 SURVIVED(단, 생산 배선은 nil ctx 를 안 넘김) | — |
| T565/T566(issued·expires 파싱 항), R3=T569(issued<expires 항), T450/R8b(owner_uid), R17(validMarket 중복) | SURVIVED — 전부 **수락-동등** 분석 완료(아래 P2) | — |

`%w%.0w` 기법 검증: 별도 프로브로 `errors.Is` 양쪽 true · 메시지 불변 실측("A: x", is(a)=is(b)=true) — 변이가 코드에 닿았음은 CAUGHT 사례(L461 등)가 양성 대조군.

### 대상 B2 — internal/app/engine collectMarket (사본 B3; 집중 스위트 후 생존자는 패키지 full 재판)
| 변이 | 결과 | 잡은 시험 |
|---|---|---|
| E4: `gate=` 대입을 클로저로 감싸 **옛 시점**(조정 직전)에 호출 — census 회피 철자 | 행동층 CAUGHT(19 fail) / **census 단독(E4census)은 PASS** | TestEveryReachableProposalClosure… |
| E10: 조정이 관문을 **다시 읽음**(두 번째 familyGateFor) | CAUGHT | loads-수 단언(적재 1회 핀) — 좋은 시험이다 |
| E8: B13(중재 거절) 닫힘이 활성화를 버림 | CAUGHT | TestAClosedMarketStillCarriesTheGatesActivation |
| **E9: B10(FAMILY_GATE_CLOSED) 닫힘이 활성화를 버림** | **SURVIVED (full)** | 없음 — 가지는 도달됨(커버리지: TestAGatedFamily…가 393행 몸통 2회) |
| **E7: B12(큐 넘침) 닫힘이 활성화를 버림** | **SURVIVED (full)** | 없음 — 도달됨(TestAMarketWithMoreScopes…가 409행 1회) |
| E5/E6: B11(충돌)·B14(미해결) 닫힘이 활성화를 버림 | SURVIVED (full) — 단 **도달 불가 공간의 변이**로 판단: lineage identity 가 LaneID·LaneVersion 을 해시(strategyflow/types.go:386~)하고 중복 종목은 B7 이 먼저 닫음 → census-only 라는 저자 판단을 **반박이 아니라 지지** | — |

### 대상 A — internal/breakoutlane/machine.go (사본 A2, 스위트 = breakoutlane 전체)
| 변이 | 결과 | 잡은 시험 |
|---|---|---|
| A3(wick `<=`→`<`), A6(B7 비점착 철자) | CAUGHT | 1.2 반사실 경계 시험 — CF 원장과 정합 |
| A1(입장 경로 1_200_000→1_500_000), A2(→0), A7(B7 에 `<2.0M` 추가) | SURVIVED — **셋 다 수락-동등(동등 변이)**: 입장 봉은 RVOL≥1.5M 이므로 입장 경로의 `≥1.2M` 은 항상 참(1.5M 이하 어떤 값도 동등), B7 도달 봉은 RVOL<1.5M 이므로 `<2.0M` 항상 참. 1.5M 초과 방향은 CF-10/기존 시험이 핀(실측) | — |
| **A4(B7 의 close buffer 를 `close>resistance` 로 약화)** | 커밋된 스위트 **SURVIVED** → 프로브(ATR=50, buffer=5, close=103)가 CAUGHT | 픽스처가 ATR=10 → buffer=1틱이라 두 식이 겹침 |
| **A5(B7 이 첫 post-range 봉만 봄)** | 커밋된 스위트 **SURVIVED** → 프로브(둘째 봉 1.2-only)가 CAUGHT | 커밋된 반사실 표 다섯 경우 전부 Bars[15]=첫 post-range 봉만 변이 |

## P0 / P1 / P2

**P0: 없음.** (현재 코드의 판정·수락은 전/후 동일 — 차분 전수 1,129행 pre/post byte-identical, `verified=true` 는 baseline 1건뿐.)

| 등급 | 지적 | 재현 | 결과 | 제안 |
|---|---|---|---|---|
| **P1-1** | 서술자 `effective` 유효성 항을 아무 시험도 안 진다 — 항 하나가 빠지면 **수락 집합이 넓어진다**(verified 매니페스트에 열거 밖 effective). 로트의 "결속 판정은 앞 판과 같다" 주장이 이 축에서 미핀 | R4/T602 변이 + 프로브: `Effective:"MAYBE", Desired:ON` → **err=nil, verified=true**, strategyrouter 패키지 전체 GREEN | SURVIVED | a112 표에 한 모양 추가: `descriptor: effective not in the enum`(Desired=ON, Effective="MAYBE") — 기존 시험은 Desired 무효(T601 CAUGHT)와 eff-ON-des-OFF 만 있다 |
| **P1-2** | "sentinel 동일성 보존 — 엔진 판별이 기댄다" 가 시험으로는 **배타성이 아니라 포함성만** 핀됨: 선언-but-고장 거절 ~17곳 중 Undeclared 를 몰래 같이 감는 변이를 4곳(빈핀계·읽기결함·폐기·만료 + 엔진층의 digest 불일치)만 잡고 **13곳 생존**. 대표 L556(몸통 결속)은 strategyrouter·엔진 full(testseams)·cmd/tossctl 세 스위트 전부 생존 — 그 변이가 살면 선언된 시장의 몸통 결속 불일치가 **기존 경로(무관문)** 로 열린다(strategy_family_activation.go:132 가 Undeclared 하나로 가름) | `fmt.Errorf("%w%.0w: …", Unavailable, Undeclared, …)` 변이 17종; `log-L556.txt`: engine ok 569s, tossctl ok 71.9s | SURVIVED 13/17 | `TestEveryActivationRefusalNamesItsFieldAndKeepsItsSentinel` 에 한 줄: `errors.Is(err, Undeclared) == (s.want == Undeclared)` — 40개 모양 전부가 즉시 배타성 핀이 된다 |
| P2-1 | collectMarket 13 닫힘의 "활성화를 싣는다" 가 도달 가능한 가지 중 **B10·B12 에서 미핀**: 두 가지 모두 시험이 지나가지만(393행 2회·409행 1회) carriage 단언이 없다. 영향은 닫힌 주기의 관측/승격 표시(항목 0 — 주문 없음)라 P2 | E9·E7 변이(가지 안에서 `gate.activation = zero`) 엔진 full SURVIVED | SURVIVED | TestAGatedFamily…·TestAMarketWithMoreScopes… 에 `familyActivation().Verified()` 단언 한 줄씩. B11/B14 는 census-only 유지 타당(identity 가 LaneID 해시) — 단 E5/E6 류 선택자-대입 변이는 census 도 행동도 못 본다는 한계를 BTM 에 적을 것 |
| P2-2 | 수명 항 3곳의 생존은 전부 **파생 거절**이 대신 막는 수락-동등(issued/expires 파싱 실패·역순은 영값 비교·만료 판정이 잡음) — 거절은 유지되나 **종류/필드명이 표류**(Unavailable↔Expired, issued_at↔"issued_at before approved_at"). "두 판정이 서로 가려 준다" 패턴 | T565/T566/T569(R3) SURVIVED; 차분 표의 life.expires.eq.issued 등 전 행 U=true 로 양 커밋 동일 | SURVIVED(동등) | 표에 `issued_at not canonical`·`expires_at not canonical`·`expires_at not after issued_at` 세 모양 추가하면 항이 핀된다 |
| P2-3 | 2.3 커밋된 반사실 픽스처의 두 맹점: buffer 가 1틱으로 접혀 close-buffer 항이 `>resistance` 와 구별 안 되고(A4), 다섯 경우 전부 첫 post-range 봉만 변이(A5) | A4/A5 스위트 생존 → ATR=50 프로브·둘째 봉 프로브가 각각 격추 | SURVIVED→프로브 CAUGHT | ATR 큰 반사실 경우 1 + 비첫봉 경우 1 추가 |
| P2-4 | `1_200_000` 리터럴 두 자리 질문(브리프 A-3)의 답: **B7 자리는 변이 핀이 충분**(CF-03/04/05 + 내 A3/A6 — 경계 정확·배타 핀). **입장 경로 자리는 1.5M 이하 전 구간이 동등 변이**(입장이 RVOL≥1.5M 을 함의)라 골든 결속이 지킬 행동이 없다 — 위쪽(>1.5M)은 CF-10·기존 시험이 핀. 골든 `rvol_counterfactual_ppm` 배열 자체는 여전히 어떤 시험도 코드 리터럴과 대조하지 않음(guard_test 는 rvol_min_ppm·wick 만) | A1/A2 SURVIVED + 동등성 논증(위) | — | 골든 배열↔리터럴 대조를 원하면 B7 자리 하나만 묶으면 된다; 입장 경로는 `= true` 로 단순화하거나 주석으로 동등성을 적을 것 |
| P2-5 | BTM 인용 부정확 1건: Load B10(끝의 ctx 재확인, 488행)의 GREEN 란 "도달 — 측정은 기존 BTM" — **현 스위트 전수 커버리지에서 488행 몸통 count=0**이고, 인용된 pre-edit 번들에는 BTM 자체가 없다(ast.json+FLM 만). 인용 시험(TestACancelledContext…)은 **앞의** ctx 검사(440행, count=1)만 지난다. 나머지 BTM 행은 전수 대조에서 정확(B7 신규 가지 커버리지 2/1/1 실측, decode B1 도달불가 명시는 count=0 과 정합) | cov.py: full-suite 488 count=0, 440 count=1 | — | 그 행을 "도달 불가(결정적 입력 없음) — census/검토로 닫음" 으로 정정 |

## 시도했으나 막힌 공격 (안 본 것과 구별)
- **수락 집합 변경**: 47개 축 단독+쌍 조합 1,129행을 `0b441270^` 와 `0b441270` 격리 사본에서 동일 하네스로 전수 비교 → **byte-identical**. `known &&` 묶음, 수명 항 순서(영값 시각), B5/B6 분리, 설정 결속 market 항 모두 수락 불변.
- **census 회피(E4)**: 클로저 철자로 census 는 통과시켰지만 행동층(활성화 적재 횟수+carriage 단언)이 19건으로 격추 — 2층 설계가 의도대로 동작.
- **관문 이중 읽기(E10)**: loads-수 단언이 격추.
- **failedFields 항별 무력화 34종**: 29 CAUGHT — 복합 결속의 항 대부분은 기존+신규 시험이 실제로 지고 있다.
- **2.3 결정 오염**: decisionSeal 은 provenance RVOL 플래그를 해시하지 않음(machine.go:194 parts 열거로 확인), B7 은 전이·수량에 안 닿고, 시험의 쌍둥이 비교+seal 재계산 단언이 CF-09(전이 추가) 를 격추 — 기록 전용 주장 성립. 패키지 밖 소비자 0 실측(grep 전수: RVOLAt*·Provenance() 외부 사용 0).
- **골격 증거 검증**: RED 로그는 편집 전 두 경계 경우만 실패(진짜 RED), renumber.txt 는 difflib 정렬 형식 정합, 저자 변이 원장 13/13·10/10 중 내가 재현한 동형은 전부 같은 판정.

## 저장소 무변경 실측
시작: `?? .reticle-setup-crash.log / ?? …review-8.5-2026-10/ / ?? w4.log`, HEAD `556c3c1f6cf9…`.
종료: 동일 3개 untracked, HEAD `556c3c1f6cf9…`, `git diff` 빈 출력(rc=0). 변경 0.

비고: 측정 중 한 번 사본 B 에서 변이 실행과 커버리지 실행이 겹칠 뻔해 커버리지 쪽을 중단·폐기하고 유휴 사본에서 다시 쟀다(변이 판정은 영향 없음 — 커버리지는 읽기 전용). 전용 GOCACHE 사용, go test 외 명령 0, 네트워크/브로커/토글 접근 0.
