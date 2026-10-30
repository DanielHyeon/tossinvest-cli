# 사전 범위·시험 지도

계획 commit 91868a516, main 병합 후 기준 09477dddda4c3101ea105fedf713d1d8ab456e15. Manager 승인 selection-core-plan.md 및 독립 review 그대로. source 검색: domain.Ranking/RankingItem 재사용, 기존 selectUniverse는 HTTP와 a140 정책에 결속돼 새 순수 함수에 재사용하지 않음. Go stdlib만 사용, UI N/A. 새 private 함수와 새 test만 추가하므로 기존 함수 FLM/BTM 면제; caller/flag/run/ledger/HTTP 수정0, 전체 a169 freeze 미완.

신규 selector 입력: cohort, domain.Ranking, requested count, caution, cap, N. 전체 metadata·순서·중복 fail closed 후 전 행 순서대로 reason과 선택 여부, 원본 deep copy 및 short. 유형 분류는 raw 보존·정본 enum만 exact 비교. BTM: U0/U1 타입·KR/realtime/count/caution/time, N 20/30·empty/short, 100개 전체 보존·순위/중복 거부, 각 사유 우선순위 및 floor/cap 포함·NaN/Inf, 유형 alias 제외·원본 결과 copy. RED compile→GREEN, 유효 floor/cap/N/symbol/type 변이, 새 함수의 비test caller0 source census, 선택 normal/tagged race 및 기존 digest 선택 확인. 브로커/운영/전체 tests/gate 없음.
