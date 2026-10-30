# a169 U2 scalar score 독립 검토

고정 14fc8e2afdf330292a8fa59f03534bfe6dd8b22f, 계획 650003f72. 한 후보 순수 점수 범위 PASS, P0/P1 없음. source 손읽기와 제한 시험만; 운영/측정/전체 a169 완료가 아니다.

`a169_u2_score.go:19-76`은 유한 비음수 정수 ≤2^53−1, KST 자정·정규장·selected, 전체 prior KST 날짜 중복, 최신20개와 zero count를 검증한다. 입력 복사·정렬, big.Int 합/두 번 floor 및 int64 저장 전 검사 구현이 계획과 일치한다. literal 333333→499999는 단일 합성 floor 500000과 구분한다. selection/HTTP/ledger 연결 없음; 비test 직접 caller source 검색 0은 reflection까지 전수 증명 아님.

정확 선택 `^TestA169U2Score` Go1.25 normal 0.003초/tagged race 1.013초 PASS, `/tmp/a169-u2-score-independent-normal.log`, `/tmp/a169-u2-score-independent-tag-race.log`. 두 실행 lo-only/no-route. 작성자 gzip 5개 압축해제 SHA와 목록 일치, today/19/nofloor/nooverflow/single-floor 5변이 거절 원문 확인. source-copy 변이는 별도 임시 모듈과 Candle stub이라 실제 제품 전체 검증이 아님. 실제 데이터·브로커·전체 tests/gate·U2 union·실측 미검증. gstack은 범위 수동 체크다.
