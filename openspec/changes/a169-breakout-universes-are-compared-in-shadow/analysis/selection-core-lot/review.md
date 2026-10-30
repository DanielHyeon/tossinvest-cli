# a169 순수 선정 로트 구현 검토

Manager 제한 승인 plan `selection-core-plan.md`, 독립 설계 review 및 사전 분석 `pre-edit.md`. 기준 09477dddda4c3101ea105fedf713d1d8ab456e15; plan 별도 commit 91868a516. 기존 함수·production caller/flag/HTTP/profile/ledger/report 미수정, U2·귀속·판정·실측과 a169 전체 freeze는 미완이다. UI N/A backend-only, 기존 domain과 Go stdlib 사용, 새 의존성 없음.

컴파일 RED 원문 `/tmp/a169-selection-red.log`; exact snapshot Go1.25 정상 `/tmp/a169-selection-snapshot-normal.log`, tossos_testseams tagged race `/tmp/a169-selection-snapshot-race.log`. 지정 정규식 `^(TestA169(Selection.*|SecurityType.*)|TestEffectivePolicyIsCanonicalAndExcludesRuntimeSource)$`, 단일 package `./tools/a140-breakout-signal-shadow`. test fixture만 사용하고 네트워크·실행 도구·운영은 호출하지 않았다. 기존 canonical policy test는 비어 있지 않은 digest·runtime source 제외를 확인하며 기존 digest 바이트 비교 시험으로 과장하지 않는다.

사유 순서, 1000/cap 포함, N20/30 및 100개 전부·short/empty, 잘못된 metadata·중복/rank, NaN/Inf 및 symbol, 정확 enum과 OpenAPI StockInfo enum golden, result/input slice 독립을 확인. floor·cap 경계·N·숫자 검사·FOREIGN_ETF 제거 overlay 다섯 변이 각각 RED. nil/빈 입력 slice의 모양 보존은 구현 검토에서 보완한 추가 GREEN이며 별도 선행 RED로 주장하지 않는다. 검색 원문 `/tmp/a169-selection-caller-census.txt`는 새 파일 함수 정의와 새 함수 내부 호출만 나타내고 외부 production caller가 없다. base~main merge의 기존 a140/domain 변경0도 확인했다.

수동 gstack 관련 체크: 입력 검증과 정렬/중복, finite 수치, 사유 우선순위, 반환 slice 소유권 및 OFF 비연결. SQL/LLM/네트워크/운영 변경 없음. skill의 전체 자동 specialist/mixed-model 및 telemetry/upgrade는 실행하지 않았고 독립 검증은 P3로 남는다. 초기 RED·GREEN·변이·선택 normal/race 원문은 gzip 보존, 변이 내용은 /tmp overlay를 사용해 제품 파일을 수정하지 않았다. 전체 tests/gate·실제 broker·측정·활성화·archive 미실행.
