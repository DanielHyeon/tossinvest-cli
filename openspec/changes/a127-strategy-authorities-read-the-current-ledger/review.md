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
