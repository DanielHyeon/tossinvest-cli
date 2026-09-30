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
