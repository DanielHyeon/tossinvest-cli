# a127 review

## 0. 신설 (2026-10-01)

- 배정: Manager 2026-10-01 — ROADMAP 「a112 이월」 R1(레인 활성화의 경성 선행), 범위 = 신설 문서 + freeze 준비, 구현은 freeze 승인 뒤.
- 범위 판정(Manager 2026-10-01): 착수 실측이 찾은 둘째 자리(`strategyrouter/production.go:38` · `:609`)를 a127 하나로 묶음 — 「같은 결함 · 같은
  수리 모양 · 같은 진입점(supervisor 의 동일 journalPath), route 가 risk 보다 앞이라 하나만 고치면 활성화가 여전히 0」. a112 소유자에는 Manager 가 통지.
- D1 재판정(Manager 2026-10-01): 「(b) 채택 — 내 (a) 지시를 대체한다」. 근거 넷(생산 불변식 `==` · 선례 상황 구분(교차 바이너리 vs 동일
  프로세스) · 표 · 열 목록은 이중 판정 · 결함의 본질은 동결 리터럴) — design D1 에 사슬 기록. 조건: 방향 구분 문구 · 0/미설정 거절 · 열 삭제 변이.

## 0.5.1 proposal-freeze 리뷰 1라운드 (2026-10-01, 대상 `65fe66e7`)

| 목소리 | 판정 | 지적 |
|---|---|---|
| 독립 적대(fail-open + 증거 규율, 분리 워크트리 `a127-rev1` — 제거됨) | APPROVE-WITH-FIXES | P0 0 · P1 2 · P2 6 · P3 6 |
| codex(ephemeral · `-s read-only` · 머리말 신고, 12:34:10~12:36:15 — **슬롯 요청 없이 띄움, Manager 에 즉시 보고 · 이후 요청제**) | FAIL | P0 0 · P1 1 · P2 3 · P3 1(`analysis/review-freeze/codex-r1-*`) |

두 목소리 모두 핵심 설계(D1 상수 주입 `==` · D2 0 거절 · D3 신원 유지)는 진입을 「이 엔진이 연 원장을 읽을 수 있음」 이상으로 넓히지 않는다고 판정.

| # | 출처 | 등급 | 지적 | 처분(design 2판) |
|---|---|---|---|---|
| F1 | 보이스 P1-1 · codex P2 | P1 | proposal.md 끝 · design.md 끝에 도구 호출 텍스트와 spec 사본이 섞여 커밋됨(작성 사고) | 삭제. 하나뿐인 spec 은 `specs/strategy-runtime/spec.md` |
| F2 | 보이스 P1-2 | P1 | D6 의 「4-가족 관문」은 활성화 없는 시장에 서지 않음 — 그 시장은 단일 범위 handoff | D6 재작성: 오늘 0 인 실제 이유(supervisor `:451-452`)와 a127 뒤 새로 도달 가능해지는 경로를 명시, 사람 승인 그림 정정 |
| F3 | codex P1 | P1 | route 는 active owner 가 없으면 campaign 질의를 prepare 하지 않음 → 열 부재가 그 범위에서 안 드러남 | **실측 확인**(`readset-probe.log` 2판: `entry_blocked` 삭제 원장 — owners 성공 · campaign 실패). D7: 판독 전 prepare(같은 SQL 상수), spec 시나리오 · S8 갱신 |
| F4 | codex P2 · 보이스 P2-2 | P2 | route 방향 문구가 Batch `:350-352` 감싸기에서 지워짐 · 엔진은 오류를 버림 | D3: opener 오류에 방향 · `%w`, `:352` 원인 보존, Batch 경계 단언(S12). 엔진 관측 확장은 잔여 |
| F5 | codex P2 · 보이스 P2-4 | P2 | S8(COALESCE)은 변이가 아님 · 대조 줄은 따옴표 오류 · S6 은 거절만 보면 생존 · 행 누락 | 반증표 S1~S13 재작성(S6 은 존재하지 않는 경로로 「열기 전」 관측, S8 은 prepare 삭제, S9~S13 추가). 측정 2판(실제 열 삭제 대조)으로 교체 |
| F6 | 보이스 P2-1 | P2 | 「v28~v35 가 행 의미를 안 바꿨다」는 거짓 — v35 latch 해제 · v34 정책 레코드 의미, DDL 스캔은 Go 작성자를 못 봄 | 증거 기반 정정(두 갈래로 셈), fail-open 아님 근거 명시 |
| F7 | 보이스 P2-3 | P2 | risk 확인-판독이 비트랜잭션, 「엔진만 마이그레이션」 근거 거짓(`flatten.go:226` 은 engine lock 없이 Open) | **실측 확인**. D7: risk 판독을 읽기 tx 하나로(S13) |
| F8 | 보이스 P2-5 | P2 | route 실원장 양성 시험의 기반이 없음(매니페스트 픽스처는 내부 시험 · 순환) | D4: `tossos_testseams` 작성 seam + 외부 시험 패키지, tasks 1.1 에 예산 |
| F9 | 보이스 P2-6 | P2 | a112 동결 증거(FLM 번들 · 하네스)가 트립와이어를 인용 · `collectMarket` 편집이 a112 번들을 밂 | tasks 1.3: 재기준 자리 명시, a112 소유자 조율 |
| F10 | codex P3 · 보이스 P3 | P3 | 좌표(`:614`→`:612`, `:41-45`→`:38-42`, B10/B6 은 둘 다 risk), census 문언 · 세는 규칙, 「a084 부터 28 이상」 부정확(핀은 SchemaVersion 29 시점에 태어남), 큰 사용량 질의는 측정 밖, a126 replay 주석, 시험 픽스처에 새 리터럴 35 금지 | 전부 반영(proposal · design · tasks). 핀 탄생 이력은 실측(`git show 8022f578:internal/journal/schema.go` = 29) |
| F11 | codex | 정보 | 경로 재열기는 inode 결속이 아님 | 잔여 기록 |

검증된 참(보이스): 핀 좌표 · supervisor 좌표 · migrate 동작 · engine config 리터럴 · 선례 둘 · 세 loader 의 다른 생산 호출자 0(함수 값 사용은
`strategy_route_authority.go:101` 하나) · 비시험 스키마 리터럴 둘 · v27 픽스처 셋 · freeze-ast source sha256 · v33 손실 잠금은 admission 이 강제.

판정: **2판으로 재리뷰 필요**(codex 슬롯은 Manager 요청).

## 0.5.2 proposal-freeze 리뷰 2라운드 (2026-10-01, 대상 `951b3ec3`)

| 목소리 | 판정 | 지적 |
|---|---|---|
| codex(Manager 슬롯 부여, 13:27:43~13:30:16, read-only · 머리말 신고, 반납 보고함) | FAIL | P0 0 · P1 0 · P2 3 · P3 1(`analysis/review-freeze/codex-r2-*`) — 1R P1 은 D7 로 닫힘 확인 |
| 독립 보이스 협대역(처분 자리 F2 · F3 · F5 · F7 표적, 분리 워크트리 `a127-rev2` — 제거됨) | APPROVE-WITH-FIXES | P0 0 · P1 1 · P2 3 · P3 6 |

Manager 지시(2026-10-01): F2 의 사람 승인 문장이 freeze 의 하중 — 재리뷰가 공격. 공격이 의도대로 작동했다(아래 G1).

| # | 출처 | 등급 | 지적 | 처분(design 3판) |
|---|---|---|---|---|
| G1 | 보이스 P1-1 · codex P2-a | P1 | 2판 D6 의 「오늘 0 인 이유」(supervisor `:451-452` 승격)는 **화면 · 승격 경로**다 — 주문은 refresh 사이클이 승격과 무관하게 내보냄(`strategy_entry_supervisor.go:500-502` 자기 주석 · `:1044-1046` · `engine.go:682`). 조건도 불완전(제안 서명 · evidence 결속 · 정확히 하나 · FX · candidate · automation gate · 보호 배선) | **실측 확인**. D6 재작성: 주문 경로 기준 필요 조건 1~12 전수와 각 조건의 사람 · 운영 · 자동 표기, 「승격은 주문 관문이 아님」 명시, 경보 수위 정정(Manager: 「열린다」가 아니라 「핀이라는 우연 차단이 사라지고 설계된 조건 사슬만 남는다」), 단일 범위 인용을 `strategy_dispatch_handoff.go:38-39` 로, 다중 범위 경로도 도달 가능해짐 명시. proposal Why · Impact 동문 정정 |
| G2 | 보이스 P2-1 · codex P2-c | P2 | risk 도 조건부 질의 — latch 가 선 범위는 사용량 질의 전에 ScopeRefused 로 돌아가 사용량 전용 열 부재가 **범위 국소 거절로 재표식**(spec · D3 위반, fail-open 아님). spec 의 「판독 전」이 risk 에서 거짓 | 실측(코드 순서 `:380` latch → `:399` 사용량)이 고름: **prepare 선행** — 두 적재기 모두 버전 확인 직후 · 첫 판독 전에 자기 SQL 상수 전부 prepare(risk 는 latch early return 앞). 오류 우선순위 명시. S14 추가 |
| G3 | codex P2-b · 보이스 P3-1 | P2 | S13 은 판독 일부만 tx 로 옮긴 변이를 놓침 | spec 에 「버전 확인과 모든 판독이 같은 읽기 tx」 SHALL, S13 을 세 판독 수신자 동일성 · tx 수명 단언으로. `SetMaxOpenConns(1)` 의 tx 밖 판독은 멈춤(막는 쪽) 기록 |
| G4 | 보이스 P2-2 | P2 | 「오늘 동작 변화 0」의 생산 설정 영수증이 a127 문서에 없음 | 근거를 a112 8.7.1 기록으로 명시하고 배포 전 재실측을 사람 항목 H1(tasks 2.0)로 |
| G5 | 보이스 P2-3 | P2 | S8 · S3/S4 는 픽스처가 고정되지 않으면 생존 | 반증표 머리에 픽스처 규율(버전만 바꾼 온전한 원장 · 조건부 질의 전용 열 삭제) |
| G6 | 보이스 P3 | P3 | modernc 즉시 prepare 의존 · S6 음수 · S2 태그 스위트 · 불변식 3 인용 · spec 의 Batch 경계(설계 수준이면 수용) | D7 의존 기록, S6 에 음수, 1.4 하네스 태그 스위트, D6 불변식 3 · 7 분리 서술 |
| G7 | codex P3 | P3 | design 끝 `</content>` 잔존 | 제거(F1 의 정리 누락 — 이번에 전수 grep 0) |

확인된 참(2R): D7 route prepare 는 route 에 충분(두 SQL 이 버전 뒤 유일한 질의, owners 도 범위 일치 조건부라 prepare 가 덮음) · risk tx 실현 가능 ·
판정 · 오류 신원 불변 · flatten 주장 정확(엔진 밖 `journal.Open` 은 flatten 과 engine_reconcile 둘, 후자는 engine lock 보유) · S1 · S2 · S5 · S7/S10 ·
S9 · S11 · S12 실현 가능 · S6 순서(위험 경로 검증은 `loadProductionRiskEntries` 안, route 는 opener 안 — 가드는 그 앞) · F6 인용.

ROADMAP R1 행 정정(핀 탄생 이력)은 a127 밖 · a112 행이라 Manager 가 a112 소유자에 전달(2026-10-01).

## 0.5.3 proposal-freeze 리뷰 3라운드 — 협대역 codex (2026-10-01, 대상 `18109568`)

codex(Manager 슬롯 부여 · 「PASS 면 freeze 승인 간주」 조건, 14:04:56~14:07:15, read-only · 머리말 신고): **FAIL** — P0 0 · P1 1 · P2 2 · P3 1
(`analysis/review-freeze/codex-r3-*`). PASS 가 아니므로 사전 승인은 발동하지 않았다 — 처분 뒤 재보고.

| # | 지적 | 처분(design 4판) |
|---|---|---|
| H1 | P1 — D6 「전수 1~12, 자동은 11 · 12 뿐」 거짓: 자동 조건 다수 누락(감독자 accepting · 미잠금 레인 · 캠페인 FLAT/CLOSED · lease/fencing · 충돌 attempt · 매수 여력 등)과 사람 조건 누락(거래 정책 · LIVE 마스터 스위치 · 공식 자격 증명) | D6 을 (A) 사람 · 운영 조건 12(세 라운드가 찾은 전부 — 전수 증명 아님, 거래 정책 · LIVE · 자격 증명 추가)과 (B) 자동 런타임 조건(예시 · 비전수, 좌표 포함)으로 재구성. 「a127 이 없애는 것은 핀 둘뿐」 명시. proposal Impact 동문 |
| H2 | P2 — D7 오류 우선순위는 원장 적재 안에서만 성립: 정책 결속(`bindProductionRiskInputs` — 섹터 매핑 없음 `:311` ScopeRefused)이 원장보다 먼저 | D7 에 적용 범위 명시(정책 · 입력 결속 뒤 원장 적재 안), 정책 범위 거절은 원장을 읽지 않은 정확한 범위 거절이라 유지, 주입 누락만 정책 결속 앞 |
| H3 | P2 — spec 「판독 질의 전부는 첫 판독 전에 prepare」는 `PRAGMA user_version` 까지 포함하는 과잉 · 범위 거절 우선 무조건 서술 과잉 | spec: 「원장 데이터 질의 전부는 버전 확인 뒤 · 첫 원장 데이터 질의 전에 prepare」, 「원장에서 유도되는 범위 국소 거절보다 먼저」, 정책 결속 범위 거절은 대상 아님 명시 |
| H4 | P3 — 좌표 낡음(a112 6.1 이 같은 파일 앞쪽 편집) | design 머리에 좌표 기준(`65fe66e7`)과 `18109568` 오프셋(risk +9 · route +7) 명시, 구현 로트는 그 해시 위 편집 전 AST 로 다시 잡음 |

codex 가 참으로 확인: 「승격은 주문 관문이 아님」, D6 의 `engine.go:226/:682` · proposal/account/handoff 인용, risk prepare 를 버전 확인 뒤 · latch 앞에 두는 것과
route 의 같은 자리(버전 확인 뒤 · tx 반환 전) 실현 가능, 같은 tx SHALL 은 두 적재기 모두 실현 · 시험 가능.

## 0.7 freeze 승인 · base 재고정 (2026-10-01)

- Manager 인용: 「**직접 판정: freeze 승인 — 구현 개시하라**(추가 codex 라운드 불요). 근거: H1 의 본질은 「전수」 주장 자체였고, 4판이 그 주장을
  **철회** … 하는 것으로 종결했다 … 「(A) 밖 비사람 스위치 0, 세 라운드 관측」은 관측으로 표기 유지. H2~H4 처분 그대로.」
- 구현 조건(Manager): a112 6.1 해시(`40ec5aff`) 위 + 편집 전 AST 재취득 · High-risk 전면 규율 · 반증 S1~S14 를 RED 세트로 · 수락은 D4 · 변이에 「버전 확인
  생략」 · 「prepare 생략」 · 「tx 분리」 · 「0 수락」 축 필수 · 착지 창 요청 규격 유지.
- **base 재고정** `f9a25549` → `3403be28`(WORKFLOW 「사람 승인 base 재고정」): (1) 귀속 실측 — 옛 base 이후 a127 자기 Go 커밋 **0**(이 change 디렉터리를
  만지며 `.go` 를 고친 비병합 커밋 없음 — a127 커밋은 전부 문서). (2) 승인 — 위 Manager 판정의 「a112 6.1 해시 위」. (3) 단독 커밋 — 다음 커밋.
  사유: 형제 `40ec5aff`(a112 6.1)가 같은 파일(`production_snapshot_authority.go` 의 `validProductionRiskPolicyContents` 등)을 편집해 옛 창에 들어옴.
