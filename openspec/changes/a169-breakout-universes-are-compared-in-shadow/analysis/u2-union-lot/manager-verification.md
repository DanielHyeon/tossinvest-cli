# Manager 검증 — U2 후보 합집합·순위

후보 `9ebb80d62`, 부모 `789ed34fb`, 2026-10-31. Manager가 신규 생산 함수 전체와 계획·독립 검토·시험 목록·영수증 해석을 대조했다. U0/U1과 점수 함수를 결합한 미연결 순수 로트의 게시를 승인한다.

후보는 두 receipt의 전체 Rows 중 필터 통과 합집합이다. 앞N Symbols로 후보를 미리 자르지 않는다. volume raw가 존재하면 필터 실패와 무관하게 V에 우선 사용하고 invalid V를 amount로 대체하지 않는다. 후보 사유는 V 오류→daily 부재→score 오류이며 missing과 empty를 구별한다. FetchedAt 둘을 검증한 뒤 max를 selected_at으로 사용한다. 양쪽 가격/사유와 receipt 복사본을 유지하며 단일 canonical 가격을 임의로 만들지 않는다.

Rows는 symbol 순으로 고정하고 선택 순서는 ratio 내림·volume 존재/순위·amount 존재/순위·symbol이다. 정책 N20/30 및 short/empty를 유지한다. 기존 함수는 수정하지 않았고 production 직접 caller0다. envelope 검증은 현재 score 오류 계약의 history_lt_20을 재사용하므로 그 계약 변경 시 union의 경계 fixture도 재검토해야 한다.

독립 Go1.25 정확 선택 normal0.008초/tagged race1.058초 PASS. 원문 `/tmp/a169-u2-union-independent-normal.log`, `/tmp/a169-u2-union-independent-tag-race.log`, lo-only/no-route. 작성자 gzip6개 SHA 및 source-copy 유효 변이4종 RED 확인. 변이는 사용 필드 domain stub의 임시 모듈에서 수행했으며 전체 제품 변이 증명은 아니다. 동일 시각에서 increase 정렬은 ratio와 동등하므로 검출로 세지 않는다. 모든 rank tie 뒤 symbol은 유효 입력에서 도달하지 않아 별도 실행 근거를 주장하지 않는다.

impeccable N/A — backend pure 함수. gstack은 수동 범위 checklist, CodeGraph 최신 index 미확인, AST 직접 호출과 source 손읽기를 구분한다. 전체 tests/gate·실제 HTTP/계좌/운영·측정·archive 미실행. 자료 원문 provenance·deadline·cache·원장·귀속·보고·fence·abort는 후속이다. next-stage-status의 과거 gap 사람 결정은 미응답 상태이며 승인된 것으로 간주하지 않는다.

최신 main과 source 합성 상태를 확인하고 정확 범위 검증 및 a169 strict·PM·diff를 통과하면 Terra가 commit/merge/push한다. 현재 main과 동일한 부모라면 변경 없는 시험을 불필요하게 반복하지 않는다. a128 기존 자료를 보존하고 a169 전체2.2/1.3을 체크하거나 archive하지 않는다.
