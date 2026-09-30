# a126 review

## 0.1 base 재고정 승인 (2026-09-30)

- 승인 참조: 사용자 2026-09-28 결정 ③(사용량 수명주기를 독립 change 로 분리 신설) + 2026-09-30 "남은것도 처리" 재개 지시
  (Manager 전달). WORKFLOW 「사람 승인 base 재고정」 조건 2 의 기록이 이 절이다.
- 조건 1 귀속 실측: a126 디렉터리를 만진 비병합 커밋 1건(`32e6ffeb`), `.go` 수정 0 → 자기 Go 커밋 0. 옛 이름 없음.
- 조건 3: 단독 커밋 `d8ed3de1`(`ed9d3685` → `989ab031`). 사이 커밋 4건의 `.go` 변경 0 → required 0, 형제 착지 몫 0.
- 독립 확인: freeze 리뷰 R3 이 조건 1·3 을 git 으로 재확인했다(아래 R3 "무발견 확인").

## 0.5.1 proposal-freeze 독립 적대 리뷰 (2026-09-30, 대상 `e2876d48`)

보이스 셋, 전부 읽기 전용(저장소 쓰기 0, 운영 DB·브로커 무접촉). 작성자와 다른 컨텍스트.

- **R1** — `code-security-auditor`: fail-open 사냥(감소가 원장 사실 없이 일어나는 경로 · late fill 되돌림 · 완화).
- **R2** — `code-reviewer`: 원장 의미론 · latch 순서 · a066 D5/D6/D8 정합.
- **R3** — `general-purpose`: 증거 규율(AST 인용 대조, 추출기 재실행) · 비목표 · 기준 커밋 재고정 · 시험 가능성.

판정: R1 · R2 · R3 모두 **APPROVE-WITH-FIXES**. **P0 0건.** 세 보이스 모두 "영수증 없이 사용량이 줄어드는 경로"(Q1 fail-open
축)를 찾지 못했다(R2: 모든 작성자 열거 아래 유지).

### 발견과 처분 (Manager 판정 2026-09-30 반영)

| # | 발견 | 보이스 | 심각도 | 처분 |
|---|---|---|---|---|
| 1 | D4 "오늘과 정확히 같은 합" 거짓 — 영수증 커밋~late fill 창에 공유 bucket 여유가 풀리고, 그 사이 진입이 소비한 여유는 ORPHAN_FILL 되돌림이 취소 못 함. 해제는 ACTIVE 주문 부재만 요구(`pending_entry`) | R1 #1 | P1 (배선 시 P0) | **수용.** D4 재서술, 창을 잔여 R3(a066 #8 가족)으로 명명, owner 해제 배선 로트의 착수 조건으로 tasks 3.1. ~~"R3 해소 또는 사람 수용"~~ → **면제 불가 의존**(0.5.2 C3 재판정이 대체). 해제 경로에 조건을 더하는 안은 D5(편집 금지) 위반이라 기각 |
| 2 | late fill 경로 중 ORPHAN_FILL 이 서는 것은 하나뿐 — (a) 소유 모호 `fills.go:540–543` (b) 증분 판독 불가 binding B2/B3 (c) terminal ReplayMismatch (d) 미등록 order id. 셋 다 `released_at IS NULL` 필터 또는 조회 공집합. 스펙 시나리오 과잉 주장 | R1 #2 · R3 #1 · R2 #4 | P1 | **수용.** 스펙 시나리오를 "증분을 읽을 수 있는 등록 주문의 late BUY 가 ORPHAN_FILL 을 세우면 되돌린다" 로 좁힘. (a)–(e) 를 D4·R3 에 명명. 되돌림 사실 확장은 freeze 범위 초과. M3b 잔여 핀(서지 않음을 단언) 추가 |
| 3 | D3(한도 모집단이 떠남을 따름)은 완화 — 작은 한도를 선언한 행이 모두 떠나면 큰 선언 한도의 진입이 자기 한도로만 cap, 전부 떠나면 B10 `!found` → 기록 한도 검사 소멸. a066 #3(a) 방어 약화, §0.9 | R1 #3 · R2 #1 · R3 #10 | P1 | **수용(R1 첫 안).** 떠난 행은 한도 모집단에 **남긴다** — 빠지는 것은 사용량 합뿐. `smallestRecordedBucketLimit` 불변. 비용 = 활성성 잔여 R5. 사람 항목 불요(fail-closed 선택). M4 를 "선언 한도 확대가 여전히 거절" 로 교체, 스펙 시나리오 추가 |
| 4 | D1 join 키 미명세 — 예약 행 owner 키 사본(`r.*`)과 decision 사본(`d.*`) 사이 FK·CHECK 없음. 갈라지면 활성 owner 행이 떠난 것으로 보여 total<own → 모든 체결 ReplayMismatch | R1 #4 · R2 #5 | P2 | **수용.** D1 조건 4: `r.*` 로 잇고 `d.*` 와 같아야 함, 다르면 `ErrJournalUsageInvalid`. M6b 시험 |
| 5 | Q3 latch 열거 불완전 — UNKNOWN_ACTUAL_RISK 는 운영자 해제가 없고 `clearResolvedUnknownLatches` 가 actual 완성 시 자동으로 푼다. "자동 경로는 latch 를 풀지 않음" 거짓, 스펙 순서 문장 틀림, H2 논증에 영향 | R2 #2 · R3 #7 | P2 | **수용.** Q3 에 latch 작성자 전수(AST), 순서를 "latch 해소(RISK_OVERAGE 운영자 · UNKNOWN 자동)" 로 정정, 델타 요구 2 문장 정정. D7/H2 는 이것을 **강화 근거**로 인용(a066 이 이미 원장 사실 결속 자동 해제를 둘 한다) |
| 6 | 기준 커밋 재고정 조건 2 — 승인 기록이 인용된 review.md 가 없었음 | R3 #2 | P2 | **수용.** 이 파일 「0.1」 절 |
| 7 | D6 `RevalidateQFinalAdmission` 동작 주장에 AST 인용 없음 | R3 #3 | P2 | **수용.** 추출 대상 추가, D6 이 B25–B27 인용 |
| 8 | 스펙 델타에 한도 모집단 시나리오 없음 | R3 #4 | P2 | **수용.** #3 처분과 함께 "떠난 reservation 의 한도는 계속 cap 한다" 시나리오 |
| 9 | RowDigest 에 `departed` 추가 — 변이·RED 없음, 배포 순간 모든 digest·snapshotVersion 이 `v1` 접두 그대로 바뀜 | R3 #5 · R2 #7 | P2 | **변경 철회.** snapshot digest 가 이미 파생 filled·held 를 담으므로 형식 변경 불필요(D2). 변이 불요 |
| 10 | ORPHAN_FILL 은 Campaign hook 결선(`apply_hook.go:280`) 의존인데 D4 가 침묵 | R1 #8 | P2 | **수용.** D4 경로 (e) 로 명시, RED 시험 이름에 결선 전제 명시 |
| 11 | D2 의 새 거절이 공유 bucket 의 무관한 활성 owner 로 파급(ReplayMismatch → scope latch, 체결은 커밋) | R2 #6 | P3 | **수용.** D2 에 의도된 파급 범위로 명시, M6 하위 사례 |
| 12 | 떠난 행의 손상 플래그(owner 0·행 1)는 복구 경로 없음 | R2 #3 | P3 | **수용.** 잔여 R6, M7 은 기능이 아니라 보수적 막힘으로 기록 |
| 13 | owner 플래그는 되돌림 트리거가 아님 | R1 #5 | P3 | **수용.** D4 말미 한 줄 |
| 14 | D1-2 released_at 동등은 단일 작성자라 항진 | R1 #7 | P3 | **수용.** 부패 검출로 재서술 — 불일치는 "안 떠남" 이 아니라 거절(M6b) |
| 15 | fill 시 overage 한도는 셋째 모집단(owner 자기 snapshot 최소) | R2 #8 | P3 | **수용.** Q3 한 줄(a066 기존, 불변) |
| 16 | `readProductionRiskUsage` 가 census 에 없음 | R2 #9 | P3 | **수용.** 추출 대상 추가, 1.0.2 번들 목록에 포함 |
| 17 | `smallestRecordedBucketLimit` 삭제·호출자 재작성 하위 task 없음 | R3 #6 | P3 | **소멸.** #3 처분으로 삭제하지 않음. 주석 정정만 1.2 에 |
| 18 | `riskMinorMonotoneDelta` 내부 동작 주장에 AST 없음 | R3 #8 | P3 | **수용.** 추출 대상 추가, 현재 동작 절이 그 B3 인용 |
| 19 | R4 에 `f48e7865` 누락 | R3 #9 | P3 | **수용.** R4·1.0.1 에 둘 다 |

**무발견 확인 (보이스가 대조하고 문제없음을 적은 것).**
- R3: design 의 AST 인용 14건 스팟 체크 일치. 추출기 stdout 재실행 3 함수(`recomputeOverageLatches`·`refuseStaleBucketUsage`·
  `latchReleasedOwnerLateFillInTx`)가 census 와 sha·줄·id·kind 까지 일치. base 재고정 조건 1·3 git 재확인. 델타 두 요구 모두
  첫 물리 줄에 SHALL/MUST, strict validate 통과. 비목표 침범 없음(D3 는 #3 처분 뒤 해소). 운영 원장 실측 문장은 값·순간·
  모집단을 갖추고 H1 은 "표기만" 으로 바르게 틀지어짐.
- R2: Q3 순서는 owner 플래그로 강제되고 행 플래그는 모든 작성자가 같은 트랜잭션에서 함께 쓴다. 건강한 원장에서 total<own
  불가. 어떤 봉인도 계좌 전체 사용량을 봉인하지 않음. 스펙의 replay 결정성 문장과 정합(파생이므로). Q4 논증 정확.
- R1: 영수증·released_at 작성자는 `releaseRiskBucketOwner` 하나, 재사용 owner 키는 PK 로 불가.

**반영 산출물.** design.md 2판(D1 조건 4 · D2 파급/RowDigest 철회 · D3 불변 · D4 경로 열거와 정정 · Q3 latch 전수),
spec 델타 2판(요구 1 첫 줄 "한도 모집단에 남아야" · 시나리오 좁힘/추가 · 요구 2 순서 정정), tasks.md(M 목록 · 3.1 착수 조건),
`analysis/freeze-ast/`(24 함수로 확장). 수정 뒤 strict validate: valid.

## 0.5.2 codex 교차 모델 리뷰 (2026-09-30)

- 실행: `codex exec -s read-only` (codex-cli 0.154.0, model `gpt-6-astra`, challenge 모드). 대상 = `e2876d48` + 위 0.5.1 초안 +
  design 2판·델타 2판(워킹트리). 필수 공격 4항목(fail-open · latch 순서 · a066 D5/D6/D8·정본 스펙 정합 · 0.5.1 처분의 적정성).
- **판정: BLOCK.** P0 0 · P1 3 · P2 2.
- 작성자 확인: #1 의 order id 충돌 검사가 owner generation 안만 본다는 것(`risk_bucket_fill.go:107–113`,
  `strategy_dispatch_runtime.go:1144–1149`, UNIQUE `(decision_id, order_id)` `risk_bucket.go:770`)과 #2 의 제출 재검증이
  latch 만 본다는 것(`RevalidateQFinalAdmission` B25–B27)을 코드로 재확인했다.

| # | 발견 | 심각도 | 처분 |
|---|---|---|---|
| C1 | broker order id 가 세대 간 재사용되면 뒤 generation 의 활성 owner 가 옛 generation 의 late 체결을 가로채고, 되돌림이 "서는 경로" 에서도 옛 owner 는 떠난 채 남는다 | P1 | **수용(발견).** D4 경로 (f) 로 명명, 잔여 R3 에 편입, 두 generation 재사용 시험 핀(1.1). 체결 계보(거래일·decision 식별)로 결속하는 수리는 체결 경로 편집이라 D5 범위 밖 → 배선 로트 조건(3.1) |
| C2 | 되돌림이 서도 구멍: 한도 100, A 60 떠남 → 다른 종목 B 80 발급 → A late 체결로 60 복원(합 140). B 에는 latch·재계산 없음, 제출 재검증은 latch 만 봄 → B 제출됨. D6 "동작 불변" 결론 반박 | P1 | **수용(발견).** D4 「되돌림이 서도 남는 구멍」, D6 정정("코드 불변이나 영향 없음 아님"), 잔여 R3 편입, 해제→발급→late 체결→제출 시험 핀(1.1). 떠남이 새로 여는 순서임을 명시 |
| C3 | 0.5.1 #1·#2·#10 은 해소가 아니라 이연이다. 스펙 시나리오를 ORPHAN_FILL 이 선 경우로 좁히면 요구가 빠진 보호에 조건부가 된다. 배선 조건이 "사람 수용" 으로 면제 가능하면 안 된다 — 면제 불가 활성화 의존이거나, 그 경로들이 떠남을 보수적으로 무효화할 때까지 떠남을 끔 | P1 | **Manager 재판정 대기.** Manager 판정(2026-09-30, R1 #1 (a))은 "R3 해소 **또는** 사람 수용" 이었다. codex 는 "또는 사람 수용" 삭제를 요구한다. 작성자 의견: 생산 효과 0(R1) 동안 실재하지 않으므로 freeze 를 막을 사유는 아니나, 배선 조건을 **면제 불가**로 바꾸는 것은 fail-closed 쪽이며 비용은 배선 로트의 범위 증가뿐이다 — 채택 권고 |
| C4 | D2 는 **떠난** 행만 HELD/held≠0 을 거절하고 델타는 영수증 있는 행 전부를 거절 — 영수증 + scope latch(되돌림) 행이 손상 검사를 우회 | P2 | **수용·반영.** D2 를 "영수증 있는 행은 떠남 여부와 무관하게, scope latch 판정 앞에서" 로 정정. 1.1 에 영수증×scope latch×HELD 조합 |
| C5 | replay 결정성이 수락 시험 없이 주장됨 — 영수증·scope latch 가 저장값 변경 없이 사용량을 바꾸는데 재시작 replay·해제 대 체결 교차 순서 시험이 없음 | P2 | **수용·반영.** 1.1 에 replay 시험(해제 전·후·되돌림 뒤·교차 체결, 사용량·digest·latch 동일, `total<own` 부재) 추가. 0.5.1 무발견 확인의 "파생이므로 정합" 은 논증이지 증명이 아님을 여기 적는다 |

**C3 Manager 재판정 (2026-09-30) — codex 안 채택.** "또는 사람 수용" 을 지우고 면제 불가 의존으로 바꾼다. 이 판정이
0.5.1 #1 처분 (a) 의 그 구절을 대체한다. 근거(Manager): (a) 판정 시점의 위험은 영수증~late fill 창 하나였으나, C1(세대 간
order id 재사용 — 떠남 영구 미복원)·C2(되돌림 전 발급 주문의 한도 초과 제출)로 가족이 커졌고, 커진 가족 전체를 "사람 수용"
한 줄로 우회 가능하게 두는 것은 fail-open 초대다. 사용자의 수용 권한은 tasks 문구에서 나오지 않으므로 문구를 지워도 권한은
줄지 않는다 — 필요해지면 명시적 재결정으로 연다. 반영: tasks.md 3.1, design.md 잔여 R3. 허용되는 대안: 떠남 판정을 기본
OFF(원장 래치)로 싣고 R3 해소 뒤 켜기.

**freeze 상태: 종료.** 0.5.1 발견 19건과 codex C1–C5 전부 처분(Manager 승인 2026-09-30). 수정 뒤 strict validate: valid.

## 0.7 H2

Manager 판정(2026-09-30) — design.md D7 「H2 처분」. 사용자 최종 보고에 해석을 명시해 거부권을 남긴다.

## 1. 구현 로트 (2026-10-01~, Opus 팀메이트)

### 1.0.1 base 재고정 — `b30318d6`

`989ab031` → `a189e74f`(base-commit.txt 단독). 조건 ①: 옛 base 뒤 96 커밋 중 a126 디렉터리를 만진 비병합 5, `.go` 0. 옛 base 판정 required 93 은
전부 형제 착지 몫이고 `internal/riskbucket` · `journal/risk_bucket*` · `production_snapshot_authority` 는 창 안 diff 0. 조건 ②: tasks 1.0.1 +
Manager 배정(2026-10-01) + 사용자 상임 지시.

### 1.0.2 편집 전 FLM — `467322df`

`readProductionRiskUsage` · `aggregateProductionRiskUsage` · `refuseStaleBucketUsage`(주석) · `Journal.releaseRiskBucketOwner` · `loadRiskBucketFillTransition`.
앞의 넷은 freeze census(`989ab031`)와 sha · 분기 일치. 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 편집 전 커버리지(`analysis/impl/coverage-pre-edit.out`,
`-coverpkg=./internal/riskbucket,./internal/journal`)로 생성.

### 1.0.3 a066 결함 수리 (스코프 정정 — Manager 판정 (가) 2026-10-01)

**발견.** 체결이 한 번이라도 있었던 owner 는 영원히 해제되지 않았다. `releaseRiskBucketOwner` 검사표 `unresolved_fill`(`risk_bucket_owner.go:932`)은
`f.actual_known=0 OR NOT EXISTS(evidence)` 였는데, `risk_bucket_fills` 의 생산 INSERT 는 한 자리(`risk_bucket_fill.go:1063`)이고 `actual_known` 을 리터럴 0 으로
쓰며, actual 보완(`completeRiskBucketFillActual`)은 evidence 행만 더한다(fills 는 append-only). 해소 정의의 정본인 `loadRiskBucketFillTransition`(`:885`)은
`actual_known=1 OR EXISTS evidence` 다 — 해제 검사만 반대로 적혀 해소된 체결도 미해소로 셌다.

**실측.** 실제 체결 → actual 보완(`ActualEvidenceCompleted=true`) → 수명주기 종결 → release = `blocked by unresolved_fill`. 그 조건만 임시로 `AND` 로 바꾸면 release
성공 · a126 RED 가 기대대로 실패(5 bucket 전부 filled=50). 임시 편집은 즉시 되돌렸다(diff 0 확인). a066 시험이 못 본 이유: `closeRiskBucketOwnerLifecycle` 이
filled_minor='0' 으로 체결 없이 FILLED 를 만들었다(그 주석이 적은 갭과 같은 뿌리).

**판정 원문(Manager, 2026-10-01).** *"이것은 해제 「완화」가 아니라 같은 규칙의 두 철자가 갈라진 결함이다 — 해소 정의의 정본은 `loadRiskBucketFillTransition`(:885)
이고, 해제 검사만 반대로 적혀 서로를 가렸다 … 방향 안전: 해제 생산 호출자 0(빈 효과) + 배선은 a126 freeze C3 의 면제 불가 조건 뒤 — 「자동은 조이기만」 침해 없음.
이 논거와 함께 사용자 거부권 항목으로 보고한다(해제 경로 접촉이므로). (나) 기각 — 생산이 쓰지 않는 상태를 딛는 픽스처는 시험을 원장 밖에 세운다."*

**수리.** 해소 조각을 공유 상수 `riskBucketFillActualResolvedSQL`(`risk_bucket_fill.go`)로 두고 두 자리가 쓴다 — 재구성은 `CASE WHEN <조각>`, 해제 검사는
`AND NOT <조각>`. 재구성 쪽은 **문자열 동일**(동작 무변). 분기 · 검사 순서 · 쓰기 무변. 결속 시험(Manager 조건 ②):
`TestA126ReleaseAndTransitionShareOneFillResolutionRule`(해소 3상 — known=1 만 · evidence 만 · 둘 다 없음 — 에서 조각의 해소/차단이 정반대) ·
`TestA126ReleaseAcceptsAnActualCompletedFillAndRefusesAnUnresolvedOne`(생산 경로의 두 상태). 이 수리는 **사용자 거부권 항목**이다(해제 경로 접촉).

### 1.0.4 Pre-Edit 선언

```text
Pre-Edit Gate:
- change / task: a126-filled-exposure-leaves-the-bucket / 1.2 GREEN + 1.0.3 a066 결함 수리 + 1.4 픽스처
- 대상 심볼:
    riskbucket.readProductionRiskUsage     — SQL 에 영수증 · owner released_at · scope latch · 결정 사본 조인(행마다 떠남 사실), 분기 무변
    riskbucket.aggregateProductionRiskUsage — D1 · D2 규칙의 유일한 적용 자리(손상 · 불일치 거절 → 떠남 → 합에서만 뺌), RowDigest 형식 불변
    riskbucket.productionRiskUsageRow(타입) — 사실 열 6
    journal.refuseStaleBucketUsage          — 주석 한 줄(D3), 동작 무변
    journal.Journal.releaseRiskBucketOwner  — unresolved_fill SQL 한 곳(공유 조각), 분기 무변
    journal.loadRiskBucketFillTransition    — 해소 조각을 상수로(문자열 동일), 동작 무변
- CodeGraph (1.6.0): ReadJournalBucketUsage 생산 소비자 4(refuseStaleBucketUsage · riskBucketSharedUsage · RevalidateQFinalAdmission ·
  loadProductionRiskEntries) · releaseRiskBucketOwner callers 9 전부 _test.go(생산 0) · loadRiskBucketFillTransition ← ApplyFill 경로
- 기존 동작 근거: HEAD b30318d6(대상 파일 base 이래 무변), freeze census, 기존 시험(a066 · 체결 · 해제 · admission)
- FLM/BTM: analysis/function-logic/ 5 번들(467322df, 편집 전)
- upstream 영향: 없음(TossOS 코드)
- 실패 시험 선행: yes — analysis/impl/red-*.log
- 설정 · DB · journal: 스키마 · 저장값 무변(Q2 파생). 시험 픽스처 2 곳(v27식 축소 스키마)에 빈 표 · 열 추가 — 단언 무변(아래 영수증)
- §0 검토: 통과 — 주문 없음 · 손절 경로 무접촉 · 감소는 영수증에만(§0.9 보수) · 한도 모집단 불변(D3) · 해제 배선 없음(tasks 3.1) ·
  생산 효과 0(해제 생산 호출자 0 · 운영 원장 영수증 0 — design Q1 · Q2)
```

**픽스처 파급 영수증(Manager 조건).** `internal/riskbucket/production_snapshot_authority_test.go` 의 `createProductionRiskDB` 와
`internal/app/engine/strategy_risk_authority_test.go` 의 `createStrategyRiskLoaderJournal` — diff 는 `CREATE TABLE` 문에 열 추가
(reservations: decision_id · market · symbol · owner_prospective_generation, scope_latches: prospective_generation)와 빈 표 셋(receipts · owners ·
final_decisions) 추가뿐이고 **어떤 단언 · INSERT · 기대값도 바뀌지 않았다**(두 파일의 다른 줄 diff 0).

**`loadProductionRiskEntries` 스키마 27 고정 대조(Manager 요청).** `productionRiskJournalSchema = 27`(`production_snapshot_authority.go:33`), `:361` 에서
`PRAGMA user_version` 이 27 이 아니면 거절. 생산 원장은 v32(design 운영 원장 실측) → 이 reader 는 오늘 생산에서 이미 거절하며(a066 잔여 #7), 떠남 판정이 이 경로의
결과를 바꿀 수 없다. v27 원장에도 영수증(v24) · owners · final_decisions(v22) · scope_latches(v22)가 실재하므로 새 조인은 실제 v27 원장에서 성립한다 —
빈 표를 더한 것은 축소 스키마 **픽스처**뿐이다.

### 1.1~1.4 RED → GREEN → 변이 (격리 연결 워크트리)

**RED**(`analysis/impl/red-journal.log` · `red-riskbucket.log`): riskbucket 행 규칙 시험은 새 사실 열이 없어 컴파일 실패. journal 통합 시험은 a066 수리 뒤
13 FAIL · 5 PASS — PASS 다섯은 현 동작을 고정하는 대조(활성 owner 불변 · 영수증 없는 해제 표식 · 떠남 없으면 stale · 부분 매도 불변 · 결속 행렬)다. FAIL 은
전부 기대한 이유(떠남 없음 → filled 50 · 손상 거절 없음 → err nil 등).

**GREEN**: `readProductionRiskUsage` SQL(사실 열 6) · `aggregateProductionRiskUsage`(B3 사본 불일치 · B4 영수증 행 손상/불일치 거절 → B5 떠남) · 행 타입 ·
`refuseStaleBucketUsage` 주석(D3) · a066 수리(공유 조각). `RowDigest` 형식 불변(시험 `TestA126AggregateRowDigestFormatIsUnchanged`).

**D1 조건 4 의 적용 범위(구현 노트).** 사본 불일치 거절은 **영수증이 r.* 또는 d.* 로 있을 때만** 선다. 영수증이 어디에도 없는 행의 불일치는 오늘처럼
셈에 든다 — 떠남 판정과 소유 판정이 갈라질 수 있는 것은 영수증이 있는 경우뿐이고(design D1 조건 4의 목적), 이렇게 두어야 해제 생산 호출자 0 인 오늘의
동작이 손상 원장에서도 바뀌지 않는다(생산 효과 0 논거).

**1.1 목록 대응.** 떠남 양성 `TestA126AReleasedOwnersFilledLeavesEveryBucket` · 활성 불변 `TestA126AnActiveOwnersFilledStaysInEveryBucket` · 영수증 없는 해제
표식 `TestA126AReleaseMarkWithoutAReceiptDoesNotLeave` · late BUY 되돌림 두 경로 `TestA126ALateBuyViaRecordFillWithCampaignHookRevertsTheDeparture` ·
`TestA126ALateBuyViaStrategySettlementWithoutCampaignHookRevertsTheDeparture`(결선 전제를 이름에) · 서지 않는 경로 (a)(b) 잔여 핀
`TestA126ResidualAmbiguousAndUnreadableLateFillsDoNotRevert` · D3 `TestA126ADepartedRowsLimitStillCaps`(+ 반대편 `TestA126WithoutTheDepartureTheSameEntryIsStale`) ·
손상 · 불일치 × scope latch 조합(codex #4) `TestA126CorruptReceiptedRowsAreUnreadable` · D2 파급 `TestA126ACorruptDepartedRowLatchesAnUnrelatedActiveOwnerButKeepsItsFill` ·
latch 집계 유지 `TestA126ADepartedRowsLatchFlagStillBlocksEntry` · 다른 owner latch 불해제 `TestA126ADepartureDoesNotReleaseAnotherOwnersLatch` · 부분 매도
`TestA126APartialSellLeavesUsageUnchanged` · replay 결정성(codex #5) `TestA126UsageIsReplayDeterministicAcrossReleaseRevertAndSharedFills`(해제 전 · 뒤 · 공유
owner 체결 둘 · 되돌림 뒤에 원장을 새로 열어 같은 사용량 · RowDigest · latch, 건강한 활성 owner 체결에 scope latch 0) · 잔여 핀 (f)
`TestA126ResidualReusedOrderIDAcrossGenerationsKeepsTheOldOwnerDeparted` · 잔여 핀 codex #2 `TestA126ResidualRevertAfterAnotherOwnersReservationLeavesItUnlatched`.
riskbucket 행 규칙 7(`a126_departed_rows_test.go`).

**codex #2 핀의 모양(구현 노트).** frozen 문언은 「해제 → 다른 종목 발급 → late 체결 되돌림 → 제출 재검증 통과」다. 이 시험은 발급(`RecordQFinalDecisionAndReserve`)
대신 admission 예약으로 B 를 세우고, 되돌림 뒤 공유 합(50 + 60)이 기록 한도 100 을 넘는데 B 에 bucket latch · owner latch · scope latch 가 **하나도 없음**을
단언한다. 제출 재검증은 latch 만 소비하므로(freeze census `RevalidateQFinalAdmission` B25–B27) latch 부재가 곧 「재검증 통과」의 원인이다. 발급 경로 전체를
세우는 fixture 는 다른 계좌 · 결정 체계(`qFinalIssueFixture`)라 이 로트에서 만들지 않았다 — 리뷰가 부족하다고 보면 보강한다.

**1.4** `closeRiskBucketOwnerLifecycle` 은 주문 없는 owner 에 생산 작성자 경로로 전량 체결을 쌓고(`fillRiskBucketOwnerInFull`), SQL 로 FILLED · held 0 을 만들던
줄과 가림 주석을 지웠다. a066 해제 시험 전부 이 fixture 위에서 통과.

**회귀**(`analysis/impl/regress-1.log`): riskbucket · journal(674s) · app/engine(387s) · execgw · cmd/tossctl 전부 ok. `-race`(a126 시험) ok(`race.log`).
`make lint` rc 0(`lint.log`).

**변이**(하네스 `analysis/impl/mutate.py`, 대조군 GREEN · 시작 sha 단언 · 판마다 복원):

| id | 변이 | 결과 |
|---|---|---|
| M1 · M1b | 떠남 항상 참 · 떠남 미적용(수리 되돌림) | CAUGHT · CAUGHT |
| M2 · M10 | 영수증 대신 owner released_at(판정 · SQL 두 자리) | CAUGHT · CAUGHT |
| M3 · M9 | scope latch 되돌림 제거(판정 · SQL) | CAUGHT · CAUGHT |
| M4 | 떠난 행을 한도 모집단에서 뺌(D3) | CAUGHT |
| M6 · M6b-1~3 · M6c | 손상 수용 · released_at 불일치 수용 · 사본 불일치 수용 · 결정 쪽 영수증 무시 · 손상 검사를 scope latch 뒤로 | 전부 CAUGHT |
| M7 | 떠난 행 latch 를 집계에서 뺌 | CAUGHT |
| P1 · P2 | 전략 정산 경로 · RecordFill 경로의 binding 제거(경로별) | CAUGHT · CAUGHT |
| F1 · F2 · F3 | a066 수리 되돌림 · 해제 검사 fail-open · 공통 조각 OR→AND | 전부 CAUGHT(상수 이동 뒤 재실행 `mutation-3-const-moved.log`) |

18 + 3 = **21/21 CAUGHT, 생존 0**(`mutation-1.log` · `mutation-2-paths.log` · `mutation-3-const-moved.log`). M8(부분 매도 감소)은 감소 코드가 없어 변이할 자리가 없다 —
`TestA126APartialSellLeavesUsageUnchanged` 가 행동으로 고정한다.

**편집 뒤 FLM**: 5 대상 번들 재추출(편집 뒤 커버리지 `coverage-post-edit.out`) + 시험 fixture 셋의 번들(비례 원칙 — 시험 전용, 게이트가 수정 함수로 셈).
공유 상수는 파일 머리(import 뒤)에 두었다 — 함수 사이에 두면 게이트가 앞 함수의 편집으로 셌다(실측: `riskBucketSharedUsage` ·
`recordReleasedRiskBucketOrderInTx` 가 차례로 오탐). `check_analysis`(로컬 커밋 기준) rc 0 evidence complete, required 9.

**생산 효과 0 (동등성 논거, Manager 지시).** 떠남은 영수증에만 걸리고 영수증 작성자 `releaseRiskBucketOwner` 의 생산 호출자는 0(CodeGraph callers 9 전부 시험),
운영 원장 영수증 0(design Q2 실측). 영수증이 없으면 새 SQL 열은 전부 0/""이고 `aggregateProductionRiskUsage` 는 편집 전과 같은 합 · 같은 latch · 같은 RowDigest 를
낸다(B3 · B4 는 영수증이 있을 때만 서고 B5 는 영수증 없으면 거짓). a066 수리도 해제 생산 호출자 0 이라 빈 효과다. 배선은 tasks 3.1 의 면제 불가 의존 뒤.

## 1.5 구현 리뷰 합본 (2026-10-01, 대상 `9cbc7560`, 3 보이스 + codex)

**형식(Manager 지시).** 별도 문맥 Claude 보이스 셋(① 적대 fail-open · ② 원장 정합 · ③ 증거 · 시험 품질) + codex(read-only · non-ephemeral ·
`~/.codex` 금지 + 머리말 신고, 세션 `01a0f3f6-e3b6-7333-adb9-17040a750f96`). **gstack `/review` 는 대체** — 공유 작업 트리에 병행 로트(a090 · a094 · a112)의
미커밋 diff 가 섞여 있어 diff 기반 리뷰가 남의 변경을 이 로트의 것으로 읽는다. 보이스는 전부 `9cbc7560` 분리 워크트리에서 돌았고 끝나고 지웠다.

| 보이스 | 판정 | P0/P1 |
|---|---|---|
| ① 적대 fail-open | APPROVE | 0 (P2 1 · P3 2) |
| ② 원장 정합 | APPROVE | 0 (P3 4) |
| ③ 증거 · 시험 품질 | APPROVE-WITH-FIXES | 0 (P2 4 · P3 여럿) |
| codex | **FAIL** | 0 — P2 셋, 전부 시험 공백. "No new P0/P1 implementation defect" · 머리말 신고 있음 |

**생산 코드 결함 0 은 네 목소리 일치.** ① 은 조인 키가 전부 PK 라 행 복제가 없고, 떠남은 r.* 로 찾은 영수증 + r.*=d.* 에서만 서며, 대소문자 · 공백 불일치는
BINARY 비교로 영수증을 못 찾아 계상 쪽(안전 방향)으로 떨어짐을 공격으로 확인. D4 되돌림은 latch 종류 필터 없는 EXISTS 라 네 CHECK 값 전부에서 섬. a066 수리
철자는 두 자리(`risk_bucket_owner.go:932` · `risk_bucket_fill.go:891`)뿐이고 다른 `actual_known` 읽기(`risk_bucket.go:572`)는 digest 입력. 생산 효과 0 재확인
(영수증 작성자 `:1029` · owner released_at 작성자 `:1017` 모두 `releaseRiskBucketOwner` 안, 생산 호출자 0).

### 지적 목록과 처분안 (Manager 결정 전 — 제안)

| # | 출처 | 등급 | 내용 | 처분안 |
|---|---|---|---|---|
| R1 | codex 1 | P2 | 「떠남은 다른 owner latch 를 풀지 않는다」 뒷절 미시험 — B latch 를 운영자 경로로 푼 뒤 B 의 다음 실측 체결이 A 제외 합으로 overage 를 계산해야 함(delta :60-61) | **시험 보강**: A 포함이면 overage · A 제외면 아님이 되는 값으로 연장 |
| R2 | codex 2 · ③ · ① P3 | P2 | codex #2 잔여 핀이 발급 대신 admission, `RevalidateQFinalAdmission` 미호출 — 재검증이 latch 아닌 재계산으로 구멍을 닫아도 핀이 초록으로 남음 | **Manager 결정**: (가) 발급 + 재검증 실호출로 재작성 (나) 현 모양 유지 + R3 에 한쪽성 명기. 권고 (가) |
| R3 | codex 3 | P2 | replay fingerprint 에 snapshot digest 없음(tasks 1.1 문언) | **논거 + 보강**: snapshot digest = H(manifest · dim · value · limit · filled · held · RowDigest · asOf)(`production_snapshot_authority.go:403-404`) — 고정 입력 아래 fingerprint 필드의 함수이고 Latched 면 생성 자체가 거절(`:392`). 문언 이행을 위해 생산 snapshot 생성기로 digest 를 실측해 fingerprint 에 넣는 보강 제안 |
| R4 | ③ X3 | P2 | `State=="HELD"` 절 삭제 변이 생존 — HELD 시험 전부 held=5 동반 | **시험 보강**: HELD · held=0 사례를 두 손상 시험에 |
| R5 | ③ X7 | P2 | `OwnerKeyMatches` 에서 `d.symbol=r.symbol` 삭제 생존 — 사본 불일치는 generation 으로만 고정 | **시험 보강**: symbol 불일치 사례(journal 실 원장) |
| R6 | ③ X1/X2/X9 | P2 | 영수증 조인 rc 의 generation · symbol · account 삭제 생존 — 오늘은 released_at 불일치 역지지로 우연히 fail-closed | **시험 보강**: 같은 종목 옛 세대 해제 + 새 세대 활성 체결에서 새 세대 사용량이 읽히고 계상됨 |
| R7 | ① | P2 | tasks 3.1 면제 불가 의존에 코드 tripwire 없음 — 생산 호출 census(`risk_bucket_fill_test.go:343-347`)에 `.releaseRiskBucketOwner(` 없음. 호출 한 줄이면 R3 구멍이 열린 채 떠남이 켜지고 어떤 시험도 안 깨짐 | **시험 보강**: census 에 추가, 메시지에 tasks 3.1 · H1 인용 |
| R8 | ② P3-4(a) · ③ B3 | P3 | 영수증 있는 손상 행의 admission 거절을 a126 시험이 고정하지 않음 — BTM B3 인용 시험은 admission 을 안 돌림(B3 오류 반환을 끄면 a066 `TestA066StorageErrorExitsFailClosed` 만 잡음) | **시험 보강** + BTM B3 인용 정정 |
| R9 | ② P3-1 | P3 | 「문자열 동일」 거짓 — 공유 조각에 바깥 괄호 추가 | 정정: 「의미 동일, 바깥 괄호만 추가」(이 절이 정정 기록) |
| R10 | ② P3-2 · P3-3 | P3 | design D1 조건 4 의 영수증 한정 적용 · D5 「`releaseRiskBucketOwner` 편집 없음」이 구현과 어긋남 | design errata 두 줄(D1 조건 4 · D5 에 1.0.3 수리 인용) |
| R11 | ③ | P3 | `red-journal.log` 가 이전 시험 판본에서 나옴(줄 번호 불일치 · 한 FAIL 은 fixture 오류). 확정 시험으로 재도출하면 a066 수리 + 옛 reader = 12 FAIL / 6 PASS(기록 13/5) — `ReleaseAccepts…` 는 떠남이 아니라 a066 수리 시험 | RED 로그 재생성(확정 시험 · 두 상태: 편집 전 전체 · a066 수리만), 1.1~1.4 절의 13/5 를 이 절에서 정정 |
| R12 | ③ | P3 | 전략 정산 시험(`:218`)은 `backfillConfirmedStrategyFillTx` 직접 호출 · attempt 상태 위조 — 경로 끝의 binding 호출만 잰다(P1 CAUGHT) | 시험 주석에 도달성 미증명 명기 |
| R13 | ③ | P3 | `TestA126ReleaseAndTransitionShareOneFillResolutionRule` 은 상수 진리표 — 두 자리 공유는 F1 · F3 가 강제 | 기록만 |
| R14 | ③ · ① | P3 | 하네스가 시험 파일 sha 를 단언하지 않음 · mutation-1 은 옛 fill.go(`d154c684`) 위 — ③ 이 확정 트리에서 M1 · M2 · M3 · M6c · M7 재실행 5/5 CAUGHT | 하네스에 시험 파일 sha 추가, 보강 뒤 전수 재실행 |
| R15 | ③ X8 · X10 · X11 | P3 | 생존하나 fail-closed(결정 쪽 영수증 generation · `OwnerReleasedAt==""` · scope latch 세대 무시) | 기록만(보수 방향) |
| R16 | ① | P3 | reader 가 못 보는 손상: 활성 owner 체결을 해제 owner 예약에 배분하는 allocation 행(작성자 없음) · 공유 조각 셋째 철자 방지는 행동 시험뿐 | 기록만(3.1 배선 로트의 입력) |

③ 의 fixture 비약화 점검: 축소 스키마 diff 는 열 · 빈 표 추가뿐, `closeRiskBucketOwnerLifecycle` 호출자의 단언 변경 0, a066 `BlockingField` 단언
(`risk_bucket_owner_test.go` :70/114/138/155/193/254/484/636 · `a066_relaxation_test.go:370`) 전부 통과, `risk_bucket_owner.go` 덮인 블록 199→200(잃은 블록 0).
FLM 좌표 현행(③ 대조).

**판정: 수리 로트 필요(시험 · 문서만, 생산 코드 변경 0).** R1 · R3 · R4~R8 은 시험 보강, R2 는 Manager 결정, R9~R14 는 문서 · 하네스. 수리 뒤 전수 변이 재실행 +
codex resume 재리뷰 → 격리 게이트(2.1) → 아카이브 승인(2.2).

### 1.5.2 수리 로트 (2026-10-01, Manager 판정: R2=(가) · 개시 · Go 시험 착지 창)

**Manager 판정 인용.** 「R2 = (가) — 핀은 인접 대리(admission)가 아니라 실제 발급 경로+RevalidateQFinalAdmission 을 재야 한다. …
fixture 비용 수용」 · 「R1·R3~R14 처분안 그대로 승인 … R15·R16 기록 잔여 승인」.

**생산 코드 변경 0.** 바뀐 것은 시험 네 파일 · 하네스 · 문서뿐 — `git diff 7ab8cd12 -- '*.go'` 의 비시험 파일 0.

| # | 처분 | 자리 |
|---|---|---|
| R1 | 「떠남은 다른 owner latch 를 풀지 않는다」를 **생산 재계산**으로 다시 씀: A filled 50 · B 4 주 실가격 6(24 + held 30) → A 포함 104 > 100 으로 B 에 RISK_OVERAGE(SQL 주입 제거) → A 해제 뒤 latch 유지 · 신규 진입 차단 → `ReleaseRiskOverageLatch`(운영자) → B 5 주 누적(30 + 25) — A 포함이면 105 > 100 인 값에서 latch 0 · 사용량 55 | `TestA126ADepartureDoesNotReleaseAnotherOwnersLatch` |
| R2 | 잔여 핀을 발급(`RecordQFinalDecisionAndReserve`, q 12) + 제출 재검증(`RevalidateQFinalAdmission`) 실호출로 재작성 — 되돌림 전 재검증 통과(양성) · 되돌림 뒤 합 110 > 100 에서도 통과 · B latch 0 을 단언. 옛 이름 `…ReservationLeavesItUnlatched` 는 `…IssuancePassesSubmitRevalidation` 으로 바뀜 | 변이 C2close(재검증이 재계산으로 닫음) CAUGHT |
| R3 | snapshot digest 를 **생산 생성기**로 실측: 생성기는 v27 원장 전용이라 journal 실 원장에서 못 돎 → riskbucket 축소 v27 원장에 해제 전 · 해제 뒤 · 공유 owner 체결 · 되돌림 단계를 쌓고 단계마다 생성기를 두 번(새 읽기 전용 연결 = 재시작 replay) 불러 bundle digest · dimension 별 filled/held/snapshot digest 동일, 떠남이 공유 dimension 의 snapshot digest 를 움직임, 되돌림 digest = 영수증 · released_at · scope latch 를 지운 "떠난 적 없는" 원장 digest | `TestA126SnapshotDigestIsReplayDeterministicAcrossReleaseSharedFillAndRevert`(`tossos_testseams`) |
| R4 | HELD · held 0 사례를 두 손상 시험에 | 변이 X3 CAUGHT |
| R5 | 사본 불일치를 symbol · market 축으로(journal 실 원장). account 는 사용량 조회 조건이라 행이 빠져 대상 아님 | X7 · X7m CAUGHT |
| R6 | 해제 A 와 키 셋을 공유하는 활성 owner 셋(새 generation · 같은 generation 다른 종목 · 다른 계좌)의 예약이 판독되고 셈에 듦 | `TestA126AReceiptBindsOnlyItsOwnOwnerKey` — X1 · X2 · X9 CAUGHT |
| R7 | 생산 호출 census 에 `.releaseRiskBucketOwner(` 추가, 같은 줄 주석에 tasks 3.1 · H1 | `risk_bucket_fill_test.go` 한 줄 — 변이 T1(생산 호출 한 줄 추가) CAUGHT |
| R8 | 손상 변형 전부에서 admission 이 `ErrRiskBucketSnapshotMismatch` + "ledger usage unreadable" 로 거절(문구로 갈래를 가름 — 반환을 지우면 빈 사용량이 "not an amount" 로 같은 종류를 냄). snapshot 은 50 을 주장 — 되돌림 변형에서 손상 없는 앞 bucket 이 stale 로 먼저 거절해 갈래를 가리지 않게 | 변이 B3 CAUGHT · BTM B3 인용 정정 |
| R9 | 정정: 공유 조각은 옛 해제 철자와 **의미 동일, 바깥 괄호만 추가**(1.0.3 · 커밋 노트의 「문자열 동일」은 거짓) | 이 절 |
| R10 | design D1 조건 4 · D5 errata | `design.md` |
| R11 | RED 재도출(확정 시험, 시험 19): (i) 편집 전 생산(reader · 해제 검사 = `9cbc7560^`) **14 FAIL / 5 PASS** — FAIL 14 전부 `a126Release` 의 "blocked by unresolved_fill"(a066 결함의 실측 증거) · (ii) a066 수리만 **14 FAIL / 5 PASS** — 떠남 없음(filled 50) · 손상 거절 없음(err nil) · stale 등 기대한 이유, `ReleaseAccepts…` 는 여기서 PASS(a066 수리 시험). 1.1~1.4 절의 13/5 와 증거 보이스의 12/6 은 각각 그 시점 시험 판본의 수 — 확정본 기준은 위 | `red-journal.log`(ii) · `red-journal-pre-edit.log`(i) · riskbucket 은 여전히 컴파일 실패(M1b 가 행동 RED) |
| R12 | 전략 정산 시험 주석에 범위(직접 backfill · attempt 상태 위조 · 도달성 미증명) | 시험 주석 |
| R13 | 기록만 — 공유 조각 진리표 시험, 두 자리 공유는 F1 · F3 가 강제 | — |
| R14 | 하네스가 시험 파일 여섯의 sha 를 시작에 기록하고 판마다 재단언. 전수 재실행 | `mutate.py` |
| R15 · R16 | 기록 잔여(Manager 승인): X8 · X10 · X11 생존하나 fail-closed · allocation 행 손상 모양 · 공유 조각 셋째 철자 방지 | — |

**변이 전수**(`analysis/impl/mutation-4-fixlot.log`, 시작 트리 `7ab8cd12` + 시험 diff, 대조군 GREEN, 시험 sha 판마다 불변): 기존 21 + 새
X1 · X2 · X3 · X7 · X7m · X9 · B3 · C2close · T1 = **27/27 CAUGHT, 생존 0**. RUNS 에 `tossos_testseams` riskbucket · census 시험 추가.

**회귀**(`analysis/impl/regress-2.log`): riskbucket · journal 무태그(522s) · `tossos_testseams`(643s) ok, `-race` a126 ok, `make lint` rc 0.

**FLM.** 생산 함수 편집 0 → 생산 번들 재추출 없음. BTM B3 인용만 정정. 시험 함수 `TestRiskBucketUnsafeEvidenceAndReleaseMethodsAreNotExported`
는 목록에 원소 하나를 더함(분기 무변, 비례 원칙 — 게이트가 요구하면 경량 번들).
