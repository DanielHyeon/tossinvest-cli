# a169 U0/U1 선정 core 독립 검토

고정 후보 b43569c7d86b64011560fbd4f5fc4e26e6f20602, source base 09477dddda4c3101ea105fedf713d1d8ab456e15. 한정 구현 PASS, P0/P1 없음. 소스 손읽기·선택 시험·증거 확인이며 실제 broker/운영 또는 a169 전체 freeze 판정이 아니다.

`a169_selection_core.go:32-106`은 U0/U1별 ranking type·KR/realtime·시각·count/caution·N20/30·cap 검증 후 rank 양수 증가/중복을 fail-closed한다. 전체 100행의 필터 우선순위, 가격경계, 앞 N, short/empty, 원본 receipt slice 복사와 raw type mapping은 승인 계획과 일치한다. metadata는 HTTP 수행 증명이 아니며 rank 제한은 broker 보장 주장이 아니다. OpenAPI StockInfo enum exact golden 및 분류 literal 시험이 해당 enum을 비교한다. caller census는 후보의 새 함수 정의 및 내부 호출 외 production 호출이 없음을 한정 source 검색으로 확인했다; 일반 동적 호출 전체 proof는 아니다.

독립 Go1.25 지정 regex normal 0.009초/tagged race 1.046초 PASS. `/tmp/a169-selection-independent-normal.log`, `/tmp/a169-selection-independent-tag-race.log` 원문에 각 lo-only/no-route namespace를 남겼다. `TestEffectivePolicyIsCanonicalAndExcludesRuntimeSource`는 runtime source 제외·digest 비어있지 않음을 확인할 뿐 archived digest 바이트 golden이 아니다.

작성자 gzip 10개 SHA 일치, floor/cap/N/숫자 symbol/type 5개 변이는 각각 대상 literal fixture의 FAIL을 기록했다; 독립 변이를 재실행하지 않았다. nil/empty 모양 보완은 별도 선행 RED로 주장하지 않는다. 수동 gstack 범위 점검: 입력·오류 분리, 정렬/중복, 수치 경계, 사본, 비연결 및 기존 OFF를 source/fixture 관점에서 확인; 자동 mixed-model pipeline은 미실행.

U2·실제 receipt provenance/selected_at·profile·event fence·ledger/report, a169 2.2 전체·freeze·실측 및 archive는 미완. 실제 네트워크/브로커/계좌/운영·전체 tests/gate는 미실행. 게시·최종 승인 별도.
