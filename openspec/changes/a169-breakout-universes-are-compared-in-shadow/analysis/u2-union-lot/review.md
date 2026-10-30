# U2 합집합 순수 로트 검증

기준 게시 main `789ed34fb`, 계획·독립 검토 `analysis/u2-union-plan.md` 및 `u2-union-plan-review.md`. `a169Select` 전체 Rows의 any-pass 합집합으로 후보를 만든다. 양쪽 원본과 이유를 보존하며 volume 항목이 존재하면 필터 실패해도 V 우선, invalid V면 amount로 fallback하지 않는다. V 오류 → daily 부재 → score 오류 순이다. 수신 두 시각을 각각 검증하고 max를 사용한다. 결과 Rows는 symbol순, 선택은 ratio 내림 → volume 존재/rank → amount 존재/rank → symbol순이다. 값과 결과 receipt는 독립 사본이다.

`TestA169(Selection|SecurityType|U2Score|U2Union)` Go1.25 선택 normal 0.010초·tagged race 1.066초 PASS. 31위 이후 각 랭킹 후보 상위, 한쪽 필터 실패, invalid V와 daily 부재 우선순위, 20/30·short·empty, 수신 시간·장외, missing/empty, score 오류, ratio 동률의 volume 및 amount rank, 원본/결과 독립과 map 삽입 순서, 유효 V0를 확인했다. 엄격하게 유일한 ranking rank 및 합집합에서 각 후보가 최소 한 ranking에 존재한다는 전제 아래 모든 rank tie 이후 symbol fallback은 유효 데이터에서 도달 불가능해 별도 도달 증거를 주장하지 않는다. RED는 새 함수 부재 컴파일 실패.

source-copy 변이 any→both, Rows→selected 상위만, volume 필터 실패시 amount fallback, volume 존재 tie 역전은 모두 assertion으로 거절. 독립 임시 모듈에서 실제 source/test 복사본과 사용 필드만 지닌 domain stub을 사용했다. increase_ppm 정렬 변이는 같은 selected_at에서 단조 증가하여 ratio 정렬과 동등하므로 유효 변이로 세지 않았다. 비test cmd/internal/tools Go AST의 `a169U2Union` 직접 호출 0; reflection 등 전체 간접 호출 증명 아님. 원문 gzip -n 및 SHA 보존.

수동 gstack 범위: 새 순수 함수·tests만. SQL, 운영 파일, 네트워크, 공유 동시 상태, LLM, UI, 새 의존성 없음. 모든 error path 전체 결과 오류 또는 명시적 후보 제외를 반환, nil/빈 daily 구분과 입력 공유 방지 검토. 후보수 최대200 및 일봉 최대25는 선행 함수 경계에 의존. 기존 로트의 제한을 확대하지 않는다. 검색 우선: 기존 a169Select/a169U2Score·stdlib sort 재사용, ai-memory 선행 검색 결과 없음, 현재 worktree CodeGraph index 없음; 실제 source/AST 확인. impeccable N/A — 순수 backend 계산. 전체 test/gate·실시장·네트워크·운영, 선정 영수증 provenance, 마감, caller/원장/보고 연결, 전체 freeze/측정은 미실행. 독립 검토·Manager 판정 전 push하지 않는다.
