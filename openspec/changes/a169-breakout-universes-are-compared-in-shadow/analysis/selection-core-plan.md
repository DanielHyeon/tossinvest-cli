# a169 순수 선정 첫 로트 — U0·U1과 유형

2026-10-31 Manager 제한 freeze 승인. 전체 a169 proposal-freeze 1.3은 여전히 미완료다. 사용자의 남은 개발 진행 승인 아래, 네트워크·원장·프로필·CLI 배선과 분리된 순수 계산 로트만 검토한다. UI 영향 없음: impeccable N/A — backend 계산 함수와 시험만.

## 요구와 결정

Q1–Q12를 변경하지 않는다. 이번 범위는 Q6의 a140 도구 내부 새 비공개 함수, Q7의 U0·U1 필터, Q9 유형 분류, Q10 미달 수 기록이다. a140 archive 선행 조건은 `2026-10-22-a140-breakout-signals-are-measured-in-shadow`로 충족됐다. 운영 `run`, `selectUniverse`, flag, client, evaluator, ledger, report는 수정하지 않는다. 호출자는 새 시험뿐이며 a140 기본 실행·digest가 바뀌는 배선은 없다.

새 파일은 `tools/a140-breakout-signal-shadow/a169_selection_core.go` 및 대응 시험이다. 외부 package·공개 API·새 의존성·interface는 만들지 않는다. `domain.RankingItem`을 입력으로 재사용한다. 기존 selectUniverse는 HTTP와 기존 a140 정책에 결합돼 직접 재사용하지 않고, 공통 필터를 작은 새 순수 함수로 구현한다. 추상 framework나 미래 U2 저장소는 만들지 않는다.

입력은 cohort U0 또는 U1, 해당 amount/volume snapshot의 ordered items(최대100), 수신 시각과 요청 종류, `excludeInvestmentCaution=true`를 나타내는 receipt metadata, 가격 상한이다. domain.Ranking의 Type/MarketCountry/Duration/RankedAt/Items/FetchedAt을 그대로 복사해 보존하고, domain 타입에 없는 요청 count=100와 caution=true를 별도 receipt로 받는다. MarketCountry=KR, Duration=realtime, 요청 count=100을 검증한다. RankedAt은 비어 있을 수 있는 원문이며 임의 파싱/대체하지 않는다. 입력 metadata는 caller의 주장일 뿐 HTTP 수행 증명이 아니다. 후속 caller가 실제 요청 영수증과 결속해야 하며 이번 로트는 이를 인증하지 않는다. snapshot 종류/주의 제외 표시/수신 시각 누락·잘못된 cohort·상한의 NaN/Inf/1000 미만·100개 초과는 결과 전체 오류다. 선정 수 N은 기존 D3 정책의 20 또는30만 허용하며 다른 값은 오류다. A140 재구성 cfg.Top=30은 별도 계약으로 이 함수가 바꾸지 않는다. 새로운 `--top` 처리나 flag는 추가하지 않는다.

순서: 입력의 API 순서와 Rank가 일치해야 한다. Rank는 양수·엄격 증가, 동일 snapshot 내 중복 symbol은 전체 오류다. 이는 broker가 rank를 보장한다는 주장이 아닌 이 core의 fail-closed 지원 제한이다. 실제 중복/누락/동률 발생 여부는 미측정이며 후속 receipt 검증에서 확인한다. 재정렬이나 조용한 중복 제거로 영수증을 정상화하지 않는다. 최대100개 전체를 검사하고 모든 행의 판정 근거를 반환한다. 원본 slice/items는 변경하지 않고 결과 저장소도 공유하지 않는다.

각 행의 첫 실패 사유는 고정 순서 `symbol_format`(ASCII 숫자6자리) → `currency`(정확 KRW) → `price_invalid`(NaN/Inf/0이하) → `price_below_floor`(<1000) → `price_above_cap`(상한 초과)다. 유효한 행 중 순위 순 앞 N개만 선정한다. 그 이후 유효 행은 탈락 필터 사유 없이 `selected=false`로 남긴다. 빈 입력/0개 선정도 유효 결과이며 `short=true`, actual count를 기록한다(Q10). 가격 경계 1000과 cap은 포함한다. Q12의 영숫자 코드는 자동 허용하지 않는다. `caution_absent`를 추측하거나 투자유의 제외 사실을 행별로 생성하지 않는다.

유형은 raw를 별도 보존하고 exact 값으로 분류한다. STOCK/FOREIGN_STOCK→STOCK, ETF/FOREIGN_ETF→ETF, ETN→ETN, DEPOSITARY_RECEIPT/INFRASTRUCTURE_FUND/REIT/STOCK_WARRANTS→기타, 나머지(빈값·공백·unknown 포함)→미분류. 임의 uppercase/trim으로 원문을 바꾸지 않는다. 유형은 선정 필터가 아니며 유형 미상 때문에 종목을 버리지 않는다.

선정 시각의 최종 결합, U2의 20일/중복 날짜/float64 안전 정수 경계·ratio, 귀속·겹침·판정 보고, 실제 receipt parsing, 재시작 ledger, profile, event reader와 atomic prewire fence, 영속 abort, engine identity는 후속이다. 이 로트로 기존 2.2 전체 또는 측정 준비를 완료로 체크하지 않는다.

## 연구와 검증

기준 `d9827f316`. source 조사: `internal/domain/models.go` RankingItem, 기존 `tools/a140-breakout-signal-shadow/shadow.go` selectUniverse, a169 design §3·§7과 proposal Q1–Q12. Terra의 별도 조사 `/tmp/a169-selection-core-research.md`를 원본 source와 대조했다. repository rg/Go stdlib·기존 내부 타입에서 충분한 해결책이 있어 외부 registry/MCP/web 패키지 탐색은 하지 않았다. math·time·기존 문자열 검사로 충분하며 기존 repository license 범위, 외부 라이선스 추가 없음. grill-me/grilling의 요구·입력 신뢰·경계·실패·수용 조건을 이 문서에 정리했으며 이미 승인된 정책을 다시 묻지 않는다.

신규 함수만이므로 기존 함수 FLM/BTM 수정 대상 없음. 구현 전 영향 목록과 테스트 계획을 기록한다. RED→GREEN, 독립 expected literal fixture, 선택 normal 및 tagged race, gofmt, 범위 gstack 수동 checklist를 수행한다. 전체 tests/gate 금지. 네트워크·계좌·운영·브로커 fixture도 필요 없다.

필수 시험: U0/U1 종류/receipt 불일치; 상위100/N20·30·잘못된N/empty/short; rank 역행·중복·symbol 중복; 모든 필터 사유/다중 실패 순서; 가격 양 경계·NaN/±Inf; 숫자 외 코드; 유형 전체/unknown/raw; 입력과 결과 mutation 독립. 유효 변이: floor 제거, cap 비교 경계 변경, 선정 N 우회, symbol 숫자 검사 제거, 유형 분기 누락. production source에 새 함수 caller가 없음을 AST 또는 한정 source census로 확인한다. default a140 digest 시험이 기존에 독립 실행 가능하면 정확 이름만 추가하고, 없으면 생성한 순수 로트 검증으로 default 실행 전체를 PASS 주장하지 않는다.

독립 리뷰와 Manager 승인 뒤 Terra가 구현한다. 구현 후 Manager 검증·별도 게시 승인 전 push 금지. a169 archive/측정/활성화 금지.

독립 검토 1차 P1 두 건을 정정했다: N은 기존 D3의20/30, 기타 enum은 DEPOSITARY_RECEIPT/STOCK_WARRANTS다. receipt metadata와 rank 지원 제한 P2도 위에 명시했다. 유형 enum literal golden은 저장소 OpenAPI SecurityType 선언과 대조하고 DR/WARRANTS 별칭은 미분류로 고정한다. 검토는 `/tmp/a169-selection-core-plan-review.md`, 재검토 후 Manager가 제한 freeze를 기록한다.

재검토 결론: 두 P1 해소, 새 P0/P1 없음. Manager는 위 미연결 U0/U1·유형 계산 범위만 READY로 승인한다. 전체 1.3은 미완료 그대로다.
