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

## 1. 구현 로트 (2026-10-01~)

### 1.0.2 편집 전 FLM 번들 + Pre-Edit 선언

격리 워크트리(`scratchpad/wt127`, base `de3b4f65` = 새 base `3403be28` + 문서)에서 편집 대상 8 함수의 편집 전 번들(`analysis/function-logic/`) —
AST(`tools/logic-map`) · 위험 패턴 · 분기 표(`analysis/harness/branch_table.py`, 편집 전 커버리지 `analysis/impl/coverage-pre-edit.out` =
riskbucket · strategyrouter `tossos_testseams` 전체 + app/engine `-run 'Risk|Route|A112'`) · 산문(`analysis/harness/bundle_prose.json` →
`write_bundles.py`). `check_analysis` evidence complete(편집 0 상태).

**편집 전 실측이 결함을 그대로 보인다**: `loadProductionRiskEntries` B5(버전 비교)와 `openProductionRouteSnapshot` B4 · `LoadProductionRouteAuthorityBatch`
B7(opener 실패)의 진입 실측이 **아니오** — 시험 전체가 핀과 같은 `user_version=27` 픽스처라 거절 갈래를 한 번도 지나지 않았다(가린 픽스처).

**Pre-Edit 선언**(High-risk — 위험 · route 권한):
- 편집 함수: `LoadProductionRiskSnapshotAuthority`(주입 가드 — 정책 결속 앞) · `loadProductionRiskEntries`(B5 비교 · 문구, 읽기 tx 하나, prepare 선행) ·
  `readProductionRiskUsage`(SQL 상수 이동, 바이트 동일) · `LoadProductionRouteAuthorityBatch`(주입 가드 · opener 인자 · B7 원인 보존) ·
  `openProductionRouteSnapshot`(B4 비교 · 문구, prepare) · `loadProductionRouteOwnersFrom`(SQL 상수 이동, 바이트 동일) · engine 두 `collectMarket`(config
  필드 하나). 타입 `ProductionRiskSnapshotConfig` · `ProductionRouteConfig` 에 필드 하나, 리터럴 상수 둘 삭제.
- 보수 방향 논거: 판정을 넓히는 편집 없음 — 받는 원장은 「이 프로세스가 연 현재 원장」 하나뿐이고(더 옛 · 더 새 · 주입 누락 거절), 판독 SQL ·
  사용량 판정 · owner 재구성은 바이트 · 분기 불변, 추가되는 것은 거절 갈래(주입 가드 · prepare 실패)와 트랜잭션 묶음뿐.
- 손절 · 청산 경로 무관(진입 권한만). LIVE 주문 · 토글 변경 없음, 시험은 원장 픽스처만.

### 1.1~1.4 RED → GREEN → 변이 (격리 워크트리, 2026-10-01)

**RED**(`analysis/impl/red.log`, 시험만 — 생산 판정 무변, config 필드만 먼저): riskbucket · strategyrouter · engine 의 a127 시험이 전부 기대한 이유로 실패
(핀 27 거절 · 주입 가드 없음 · BeginTx 없음 · Batch 가 원인 지움 · 엔진 자리 주입 없음). 축소 픽스처를 주입 값(1)으로 바꾼 기존 시험도 핀 27 에 걸려 실패 —
**같은 결함을 이번엔 픽스처가 드러냄**. a112 트립와이어를 뒤집은 `TestTheRiskLoaderReadsTheRealJournal` 도 실패(실 원장 거절).

**GREEN**(D1 · D2 · D3 · D7):
- riskbucket: `LoadProductionRiskSnapshotAuthority` 에 주입 가드(정책 결속 앞, `journal schema version not injected`), `loadProductionRiskEntries` 가
  읽기 전용 tx 하나에서 버전 확인(정확 일치 · 방향 문구) → 두 SQL 상수 prepare → scope latch → 다섯 사용량. 리터럴 `productionRiskJournalSchema` 삭제,
  SQL 은 상수(`productionRiskScopeLatchSQL` · `productionRiskUsageSQL`, 바이트 동일)로.
- strategyrouter: Batch 주입 가드(매니페스트 · 원장 열기 앞), opener 가 주입 값과 정확 일치 · 방향 문구(`%w` Unavailable) · 두 SQL 상수 prepare, Batch 의
  `:352` 감싸기가 원인을 `%w` 로 보존. 리터럴 `productionRouteJournalV` 삭제, SQL 상수(`productionRouteOwnersSQL` · `productionRouteCampaignSQL`).
- engine: 두 config 리터럴에 `JournalSchemaVersion: journal.SchemaVersion`.
- 시험 기반(D4): `tossos_testseams` seam `SignedProductionRouteConfigForTest`(+ 내부 픽스처와 본문 일치 시험) · 외부 시험 패키지 `strategyrouter_test` 의
  `journal.Open` 실 원장 양성. a112 거래 픽스처의 위험 적재기를 실 원장으로 단일화(다리 · 헬퍼 셋 · 트립와이어 삭제 → 양성 시험), 손상 주입 시험 둘은
  `useRiskStub` 으로 stub 을 명시 선택(실 원장의 트리거가 그 모양을 막음).

**편집 전 · 뒤 진입 실측**(커버리지): 버전 거절 갈래가 편집 전 **아니오**(가린 픽스처) → 편집 뒤 risk B7 · B8 · route B5 · B6 **예**.

**변이**(`analysis/impl/mutation-1.log`, 하네스 `mutate.py` — 무변이 대조군 GREEN · 시작 sha · 시험 파일 sha 판마다 불변 · 빠른 구조 시험 선행):
S1 · S2 · S3 · S4 · S5a · S5b · S6a~d · S7 · S8 · S9a · S9b · S10a · S10b · S11 · S12a · S12b · S13a~c · S14 = **23/23 CAUGHT, 생존 0**. Manager 필수 축: 버전 확인 생략
(S3 · S4) · prepare 생략(S8 · S14) · tx 분리(S13a~c) · 0 수락(S6a · S6c, 음수 S6b · S6d) 전부 포함.

**회귀**(`analysis/impl/regress-1.log`): riskbucket · strategyrouter · strategyflow · strategyproposal · strategyarbiter · strategyworker · app/engine ·
execgw · cmd/tossctl 무태그 · `tossos_testseams` 모두 ok. journal 은 **a066 census 시험 하나**만 실패 — `TestA066StorageErrorExitsFailClosed` 는 저장 출구 ·
tx 여는 함수 수를 얼린 구조 시험이고 a127 이 `loadProductionRiskEntries` 에 출구 셋(BeginTx · user_version 판독 · prepare)과 tx 여는 함수 하나를 더함(P1 · P2
는 그 시험이 통과로 검사). main 의 a112 `0b6cb9b1`(336→337, 함수별 이름 표 도입) 위로 재기준한 뒤 340 · 13 · 함수별 6 으로 갱신, 단독 재실행 ok.
`-race`(a127 시험 · 관련 패키지) ok, `make lint` rc 0(`analysis/impl/race-lint.log`).

**재기준**: 작업 트리를 main `f0f7d668` 위로 옮김(그 사이 대상 파일 중 바뀐 것은 a066 census 시험 하나 — 위). 착지 창 범위의 형제 커밋 때문에 착지 전
base 재고정이 필요(자기 Go 커밋 0 — 조건 ① 첫째 갈래).

**편집 뒤 FLM**: 8 번들 재추출(편집 뒤 커버리지 `analysis/impl/coverage-post-edit.out`) + 시험 함수 경량 번들 12(현재 8 · 삭제된 4 는 `revision: base`).

### 1.6 착지 창 · base 재고정 (2026-10-01)

- Manager 인용: 「**착지 창 + base 재고정 승인(첫째 갈래) — 순서 1~4 진행.** … 다리 제거 · 실원장 단일화는 **그 다리에 적어 둔 제거 조건의 계획된 이행**이다.
  census 340 재기준(함수별 표 갱신+사유) 방식 맞다. a112 소유 시험 편집 · 동결 증거 재기준 두 자리는 내가 통지한다.」
- base 재고정 `3403be28` → `f0f7d668`: (1) 귀속 실측 — 옛 base 이후 a127 자기 Go 커밋 **0**. (2) 승인 — 위 인용. (3) 단독 커밋 — 다음 커밋.
  사유: 형제 착지(a112 `0b6cb9b1` 의 a066 census 시험 함수 본문 변경 등)가 옛 창에 들어옴.

### 1.6.1 구현 리뷰 — 독립 보이스(적대 fail-open + 수락 진정성 + 증거, 대상 `d8a1c312`)

판정 **APPROVE-WITH-FIXES** — P0 0 · P1 0 · P2 1 · P3 8. fail-open 경로 없음(주입 가드 · 정확 일치 · 단일 읽기 tx · 오류 신원 · 엔진 두 자리 · SQL 상수
네 개가 `f0f7d668` 과 바이트 동일 · 실 원장 수락 진정 · a066 새 출구 정확히 셋). RED 진정성 재현(편집 전 생산 파일 + 새 시험), 변이 재실행 S6b · S8 · S13a ·
S14 CAUGHT, 추가 변이 X1 · X2 · X4 · X5 · X7 · X8 CAUGHT, X3(`TxOptions{ReadOnly:true}` → nil) 생존 — DSN 이 이미 `mode=ro` · `query_only` 이고 드라이버가
`_txlock` 없이 ReadOnly 를 무시하므로 등가(정보).

| # | 등급 | 지적 | 처분 |
|---|---|---|---|
| V1 | P2 | tasks 1.3 이 체크됐는데 `a126_snapshot_digest_replay_test.go:6` 주석(「v27 원장만 읽으므로(productionRiskJournalSchema)」) 미갱신 | 주석 갱신(축소 원장인 이유 = 패키지 순환, 실 원장 수락은 engine 시험) |
| V2 | P3 | route S8 시험이 거절 사유를 고정하지 않음 — 편집 전 코드(핀 27)에서도 통과 | `journal read set unavailable` 문구 단언 추가 |
| V3 | P3 | `len(bundle.Entries()) != 5` 는 항상 거짓(길이 5 배열), 범위 루프는 빈 경우에도 통과 | `bundle.Digest() != ""` + 준비된 범위 정확히 2 |
| V4 | P3 | route 실 원장 양성은 빈 원장만 — active owner 재구성 · campaign 질의를 실 STRICT 타입에서 돌지 않음 | 잔여 기록(불일치는 거절 쪽 — fail-closed). 생산 첫 실 owner 판독은 활성화 전 사람 항목 H1 과 같은 창에서 관측 |
| V5 | P3 | S7/S10 구조 시험은 go/types 가 아니라 이름 기반 — 문서는 go/types 라 적음 | 문서를 구현에 맞춤(import 해석 AST 선택자, 지역 식별자 가림은 범위 밖) |
| V6 | P3 | 편집 뒤 FLM 의 호출 산문이 편집 전(db · ReadJournalBucketUsage(db)) | 편집 뒤 호출 산문 분리(`calls_post`) |
| V7 | P3 | route BTM B4(user_version 판독 실패 — 편집된 갈래)를 「기존」으로 표기 | 편집된 갈래 · 시험 없음 not-applicable(결함 주입 seam 없음, fail-closed) 로 정정 — risk B6 도 같은 사유 명시 |
| V8 | P3 | Batch 경계 문구에 sentinel 이 두 번 | 기록(외관) |
| V9 | P3 | route 실 원장 시험의 `os.Chmod(0o600)` 가 journal.Open 의 권한 강화 회귀를 가림 | 삭제 |

### 1.6.2 구현 리뷰 — codex(Manager 슬롯, 17:03:49~17:05:31, read-only · 머리말 신고, 대상 `d8a1c312`)

**FAIL** — P0 0 · P1 0 · P2 2 · 생산 fail-open 0(`analysis/review-impl/codex-i1-*`). codex 가 참으로 확인: 주입 가드 · 양방향 거절이 데이터 판독 앞 · 단일 tx ·
오류 신원 · 엔진 두 자리 · SQL 상수 네 개 바이트 동일(스크립트) · 두 수락이 `journal.Open` · a112 단언 불약화 · a066 +3/+1 일치. (§1.6.1 은 `d8a1c312`
뒤에 쓴 것이라 codex 가 보지 못함 — 중복 제외는 수동으로 함.)

| # | 지적 | 처분 |
|---|---|---|
| K1 | P2 — S13 구조 시험은 수신자 **철자**만 봄: 같은 이름으로 tx 를 닫고 다시 열어도 통과, BeginTx 수 · ReadOnly 옵션 미검사(spec 「같은 읽기 전용 트랜잭션」 · design 「수신자 동일성과 tx 수명」) | risk 구조 시험에 수명 단언 추가 — BeginTx 정확히 하나 · `&sql.TxOptions{ReadOnly: true}` · tx 대입 하나 · defer 된 Rollback 외 Rollback/Commit 0. 변이 S13d(ReadOnly 제거 — 보이스 X3 생존분)· S13e(닫고 다시 엶) 추가 |
| K2 | P2 — 열 삭제 시험은 「prepare 를 latch 질의 뒤로」 같은 **순서** 위반을 못 가름(spec 「첫 원장 데이터 질의 전에 prepare」) | 순서 구조 단언 — risk: 데이터 질의(PRAGMA 제외 Query*Context · ReadJournalBucketUsage)는 전부 마지막 PrepareContext 뒤. route: opener 는 읽기 전용 BeginTx 하나 · 데이터 질의 0 · 모든 prepare 뒤에만 성공 반환, Batch 는 opener 를 owner 재구성보다 먼저. 변이 S15(risk 데이터 질의를 prepare 앞) · S16(opener 가 prepare 앞에서 데이터 판독) 추가 |

**1.6 수리 묶음 변이**(`analysis/impl/mutation-2.log`, 같은 규율): 기존 23 + S13d · S13e · S15 · S16 = **27/27 CAUGHT, 생존 0** — 보이스 X3(ReadOnly 제거)도 이제 잡힘.
수리 묶음은 시험 · 하네스 · 문서만(생산 코드 변경 0).

**재리뷰 생략 근거**(Manager 판정 2026-10-01): 생산 코드 무변(수리 묶음은 시험 · 하네스 · 문서) · 남은 지적이 P2 시험 판별력 계급이고 수리가 구조 단언(수명 · 순서)으로 종결형 ·
변이 27/27 CAUGHT — a091 i2 와 같은 처분(P2 꼬리는 수리 기록으로 닫고 게이트행). 수리 묶음 착지 `1e25b3a3`.
