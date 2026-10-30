# Manager 검증 — U2 순수 점수

후보 `14fc8e2afdf330292a8fa59f03534bfe6dd8b22f`, 2026-10-31. Manager가 생산 함수와 시험 전부, 고정 범위 독립 검토를 읽었다. 신규 pure 함수만이며 기존 production caller0. 개발 로트 게시를 승인한다. 실제 U2 후보 선정·측정 준비 승인이 아니다.

KST 날짜 자정 이전 봉만 복사·정렬하고 최신20개를 사용한다. 오늘/미래·0Time 제외, 전체 prior 날짜 중복 거절, 최신20개 밖 Volume 미검사 범위가 계획과 맞는다. V와 volume은 안전 정수 범위 안에서 int64로 변환하고 big.Int 합·곱·두 번 floor 뒤 결과의 int64 범위를 검사한다. ratio 333333→increase499999와 단일 floor500000을 literal fixture로 구별했다. 실패는 빈 결과와 명시 사유이며 점수0 성공으로 바꾸지 않는다.

독립 Go1.25 정상0.003초/tagged race1.013초 PASS. 원문 `/tmp/a169-u2-score-independent-normal.log`, `/tmp/a169-u2-score-independent-tag-race.log`. 영수증 gzip5개 SHA와 변이5개 RED를 확인했다. 변이는 실제 새 source/test의 임시 모듈 복사 및 Time/Volume만 가진 Candle stub에서 실행했으므로 전체 제품의 변이 실행으로 확대하지 않는다. 정상/tagged race는 실제 package에서 수행했다.

PriorDays는 선택20개의 수가 아니라 시간0/오늘 이후를 제외한 전체 prior 봉 수(최대25)이며 ZeroVolumeDays는 선택20개만 센다. 실제 거래일/시장 timestamp 품질은 인증하지 않는다. 수치·중복 지원 제한, 원문 유실/0 구분 불가능성은 계획대로 남는다.

impeccable N/A — backend 계산만. 기존 함수 FLM 변경·신규 의존성 없음. gstack 범위 수동 checklist, CodeGraph index 미초기화, AST 직접 caller0와 source 손읽기 구분. 전체 tests/gate·실제 네트워크·계좌·운영·archive 미실행. 최신 U0/U1 main 통합 후 정확 선택 검증·scoped strict/PM/diff를 통과하면 Terra가 게시한다. 전체 a169 tasks2.2/1.3 및 측정은 미완 유지한다.

게시 합성: selection main `54f1dc59b`와 통합한 source에서 U2 score source/test 원문 차이 없음. Go1.25 `^TestA169(Selection|SecurityType|U2Score)` 정확 선택 normal/tagged race PASS. 결과 원문은 receipts/merged-*. 전체 test/gate·운영·네트워크 미실행. 2.2 및 전체 freeze 미완 유지.
