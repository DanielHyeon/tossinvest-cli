# a169 U2 합집합 독립 검토

대상 9ebb80d62b691baaaeb7f5e020cbfa4f92ff4de7, parent 789ed34fb. 한정 순수 함수 PASS, P0/P1 발견 없음. 손읽기·선택 시험·원문 대조, 소스 변경 없음.

`a169_u2_union.go:28-123`은 전체 ranking Rows 중 Reason 없는 행의 ANY-pass로 후보를 만들며 앞N Symbols로 자르지 않는다. 두 원본 및 행 사유 보존, 필터 실패해도 volume 원본이 있으면 raw V 우선, invalid V→daily 누락→score 순서, missing vs 빈 일봉, 각 수신 시간 검증 및 max, 순위·결과 symbol 정렬, input receipt 복사 계약을 확인했다. 단일 last_price projection·실제 provenance/ledger/마감은 구현하지 않는다.

Go1.25 정확 작성자 선택 `^TestA169(Selection|SecurityType|U2Score|U2Union)` normal 0.008초, tagged race 1.058초 PASS. 원문 `/tmp/a169-u2-union-independent-normal.log` 및 `...-tag-race.log`, 각각 lo-only/no-route. 6개 gzip 압축해제 SHA가 목록과 일치. source-copy stub 모듈의 ANY→BOTH, Rows→selected, volume fallback, tie 역전 네 변이 모두 RED 원문 확인; 제품 전체 증거가 아니다. increase 키는 모든 후보 공통 양수 시간비에서 ratio의 단조 변환이므로 동등변이로 제외했으며, symbol tie fallback은 유일 rank/각 후보 랭킹 존재 전제에서 도달하지 않는다고 기록한다. 직접 비test caller 0 source 검색 및 작성자 AST 증거는 동적 reflection 전체 증명이 아니다.

수동 gstack 제한 점검: receipt 소유권, 빈 입력/오류, 정렬 및 숫자 경계, 공유 상태/네트워크 미존재. 전체 자동 pipeline은 아님. 실제 broker·운영·전체 tests/gate·a169 전체 2.2/freeze/측정/archive 미완. Manager 승인 및 게시 별도.
