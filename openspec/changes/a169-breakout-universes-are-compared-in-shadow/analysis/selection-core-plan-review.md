# a169 순수 선정 core 계획 독립 검토

대상 selection-core-plan.md 2026-10-31 원문, source d9827f316. 판정: 구현 전 P1 두 건 정정 필요. 손읽기·원문 검색만; AST·시험·운영 미실행. a169 전체 freeze/측정 승인이 아니다.

## P1 — 종목군 N과 A140 top 혼동
plan:11,15는 U0/U1 선정 수를 상수 30으로 고정한다. `design.md:43,57-60,114-115,129,349-353,653`은 A140 재구성과 cfg.Top만 30으로 고정하고 U0/U1의 `per_universe`는 정상 30, D3(a) 승인 시 20인 별도 판이라고 명시한다. `short`도 N 기준이다. 계획대로면 승인된 20 판의 20개 선정과 short를 표현할 수 없다. 최소 수정: 입력으로 정책 N(허용 20 또는 30)을 받아 앞 N과 short를 계산하고, A140 top30은 이 순수 함수 밖의 별도 계약으로 둔다. 20/30 독립 expected fixture를 추가한다. 새 사용자 정책이 아니라 기존 D3 보존이다.

## P1 — securityType enum 두 값 오기
plan:17의 `DR`·`WARRANTS`는 정본 `design.md:280-288` 및 proposal enum의 `DEPOSITARY_RECEIPT`·`STOCK_WARRANTS`와 다르다. 이대로면 실제 해당 값이 기타가 아닌 미분류가 되고, 정의 밖 별칭을 기타로 판정한다. 두 원문 enum으로 정정하고 모든 enum과 unknown/빈값을 literal golden 및 OpenAPI enum 완전성 시험으로 고정하라. 나머지 STOCK/FOREIGN_STOCK, ETF/FOREIGN_ETF, ETN, INFRASTRUCTURE_FUND/REIT 대응은 정본과 같다.

## P2 — receipt 및 rank 신뢰/지원 조건
plan:11의 metadata는 type·caution·received만 있어 정본 `design.md:114-119,475-476`의 market=KR, duration=realtime, requested count=100, rankedAt 원문, 받은 시각·상위 100 원본을 완전 기록할 수 있다는 보장은 없다. `domain.Ranking`은 request로부터 type/market/duration을 echo하고 `RankedAt`은 nullable 원문, `FetchedAt`을 가진다(`internal/official/market_reads.go:349-352,360-402`); caution·count는 도메인 결과에서 인증되지 않는다. 순수 함수가 metadata를 받거나 후속 receipt를 원본으로 보존하는 명시적 책임과 nullable rankedAt을 규정하라. HTTP 수행 증명은 여전히 후속 caller의 몫이다.

plan:13의 Rank 양수·엄격 증가와 중복 symbol 전체 실패는 정본에 명시된 broker 보장이라기보다 새로운 fail-closed 지원 제한이다. adapter는 rank를 그대로 전달할 뿐 검증하지 않는다(`market_reads.go:384-402`). 순수 core에서 보수적으로 거부할 수 있으나 공식 API의 rank 보장 또는 실제 데이터 사실로 단정하지 말고, 값 부재/동률/중복 시 해당 결과를 fail-closed로 취급하는 제한·후속 receipt 검증을 명시하라. 즉시 정책 승인을 받았다는 주장은 금지한다.

## 유지할 경계
필터 순서·1000 및 cap 포함, 빈 cohort short 수용, 유형 불명 종목 제외 금지, 입력/결과 독립, metadata가 HTTP 증명이 아니라는 점은 계획 범위에 부합한다. U2·원장·CLI·profile·shared event/fence와 실제 선정 시각 결합은 후속으로 분리하고 tasks 2.2/전체 freeze·a169 측정 완료를 주장하지 않는다. 실제 브로커·운영·전체 tests/gate 미실행.

## 수정안 재검토

수정된 plan:11–17,27 손읽기. P1 두 건 해소: U0/U1 N=20/30과 별도 A140 top30, 실제 DEPOSITARY_RECEIPT/STOCK_WARRANTS 원문 분류 및 N 경계 시험이 반영됐다. 도메인 영수증 복사·KR/realtime/count/caution 검증, nullable RankedAt 원문과 rank fail-closed 지원 제한도 명시했다. 이 한정된 미연결 U0/U1·유형 순수 로트는 **bounded READY**; a169 전체 freeze/구현/측정 승인이 아니다. 새 P0/P1 없음. source·시험·운영 실행 없음.
