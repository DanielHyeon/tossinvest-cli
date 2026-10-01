# 목소리 A 출력(리뷰 대상 80ae96a5) — 하위 에이전트 최종 보고 원문 전사

## 판정: BLOCK — P1 하나(J4 경계가 적재 단계에서 뚫림)

규칙 준수: `~/.codex` · `~/.config/tossctl` · 자격 증명은 읽지도 검색하지도 않았음. 실제 저장소에 쓴 것 0(`git -C … diff --stat` 빈 출력). LIVE 주문 · 토글 · 엔진 기동 0.
사본: `/tmp/claude-1000/a112-rev5222-P4XL` 하나를 만들었고(`.git` 없음 확인), 보조 스크립트 · 출력 파일과 함께 **삭제 완료**. `/tmp/claude-1000/a112-rev5222-ywEg` 는 내 것이 아니라 손대지 않았음.

### 발견 표

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | **P1** | **원장 무결성 결함이 범위 거절 타입으로 분류되어, 같은 주기에 다음 범위가 발급됨(J4 위반).** 범위 거절 타입은 「권한이 **없음**」에서 만들어지는데, 그 "없음"은 위험 적재기가 `LoadProductionRiskSnapshotAuthority` 의 **어떤** `err` 이든 `ready=false` 로 접어서 생김. 접히는 오류: `ErrJournalUsageInvalid`(행 손상), sql Open/Ping/Query 실패, ctx, 매니페스트 digest 불일치. 계좌 적재기 `loader.load` 의 `err` 도 같은 모양. 원인은 기록에서도 사라짐: 사유는 늘 `RISK_AUTHORITY_UNAVAILABLE`, 문구는 "no ready risk authority" | `strategy_risk_authority.go` 범위 반복(`err == nil && …` 만 준비, 나머지는 사유 하나) · `strategy_account_first_leg_authority.go:279-287`(그 부재가 `&strategyScopeRefusal`) · 같은 파일 :278 주석 「원장 · Gateway · 중앙 오류 … 는 범위 거절이 아니다」와 `strategy_owner_scope_authority.go` 머리말 「이 분류 경계가 이 설계의 안전선」은 **거짓**. **실행(사본 E1)**: stub 원장에 005930 symbol 버킷 손상 행 하나(state=GARBAGE, held=-5)를 넣고 `wave` → `deliverKR` 결과: `risk scope 005930 ready=false reason=RISK_AUTHORITY_UNAVAILABLE`, `000660 ready=true`, **placed=[000660]**, `errors.As(*strategyScopeRefusal)=true`, err=`…strategy owner scope refused (acct-risk-loader/KR/005930): no ready risk authority for this owner scope`. 대조군(손상 없음, Done 시험)에서는 첫 파도에 005930 이 발급되고 000660 은 STALE. 변이 원장 X01~X22 중 **적재기 단계 오분류를 겨눈 변이는 0** — `TestOnlyATypedScopeRefusal…` 는 결함을 전달 몸통에만 주입함. 완화 사정: 생산 영향 0(서명 활성화 0, R1 로 위험 권한 미준비). 주기를 넘기면 B 는 편집 전에도 다음 주기에 거래될 수 있었음 — 깨진 것은 「주기 안 정지」 규칙과 기록 | 적재기가 오류를 **분류해서** 실어 나를 것. 범위 국소 거절(scope latch, 정책에 종목 없음 등 — 타입 있는 sentinel)만 `ready=false` 와 사유로 두고, 원장 · 무결성 · ctx · digest 결함은 범위 칸에 fault 로 남김. 1차 레그는 fault 이면 **타입 없는** 오류를 내서 주기를 멈춤(또는 시장 전체를 준비 안 됨으로). E1 을 시험으로 커밋하고 「적재 오류를 범위 거절로」 변이가 CAUGHT 되는지 확인. 주석 두 곳 정정 |
| 2 | P2 | census 를 통과하면서 새 생산자를 만드는 철자가 있음: `As(any) bool` 메서드가 reflection 으로 대상 포인터를 채우는 오류 감싸개(식별자 언급 0) | **실행**: 사본에 `revAsAny` 를 넣고 `dispatchStrategyMarketHandoffs` 의 `return err` 를 그것으로 감싼 결과 → `TestTheScopeRefusalTypeIsMadeOnlyWhereTheCensusSaysItIs` **PASS**. 그 자리에서는 행동 시험(`TestAForgedScope…` placed=[000660], Done)이 FAIL 로 잡음. census 주석의 「새 언급은 무엇이든 이 표를 바꿔야」는 완전하지 않음 | census 에 엔진 생산 코드의 `As(` 메서드 선언 금지(또는 전수표)를 더함 |
| 3 | P2 | worker 승격의 「하나라도」가 승격 근거 범위의 위험 · 계좌 준비를 요구하지 않음. 시장 칸 `r/a.snapshot.Ready` 는 "어느 범위든 준비"라서, A(권한 있음 · 관문 거절) + B(관문 통과 · 권한 없음) 조합이면 거래 가능한 범위 0 인 Effective worker 가 됨. `AuthorityExpiresAt` 은 첫 준비 계좌 범위의 FreshUntil(최소값 아님), 증거 digest 는 계좌 identity 하나뿐(위험은 범위 묶음 — 비대칭) | `strategy_entry_supervisor.go` 승격 반복 · :497-502. 코드로 읽은 것, 실행 안 함. 주문 경로는 범위마다 다시 검사하므로(1차 레그 `forScope` · dispatch :96/:143) 발급 안전에는 영향 없음 — 화면 · R2 latch 반복에만 영향 | 승격 근거 범위에 `forScope` 준비를 요구; 만료는 범위 최소값; 계좌 digest 도 범위 묶음 |
| 4 | T | `key, _ := strategyOwnerKeyOf(...)`(:279)가 keyed 를 버림. 지금은 `authorityForOwnerScope` 가 정규화를 보장해서 도달 불가. 봉인이 바뀌면 정규화 실패가 범위 거절(건너뜀)이 됨 | 코드 | `!keyed` 이면 타입 없는 오류 |
| 5 | T | 두-레그 시험 이름이 "OfOneCycle" 인데 실제로는 **둘째 파도**에서 잼(같은 주기의 둘째 레그는 checkLimits 전에 STALE 로 멈춤). 단언 문구는 느슨한 부분 문자열 | 시험 :343-371 | 이름 · 주석 정정 |

### 필수 항목별 판정

1. **J4 오분류 양방향** — (a) **뚫림(#1)**: 원장 무결성 결함이 범위 거절로 분류되어 같은 주기에 다음 범위가 발급됨을 실행으로 확인. 전달 몸통 · admit · dispatch 의 `%w`/`errors.As`/`errors.Join` 사슬 자체는 정상이고, 위조(identity)는 타입 없이 주기를 멈춤(`TestAForgedScope…` 녹색 + reflection 변이에서도 FAIL 로 잡음). census 우회 철자 하나 있음(#2). (b) 굶김: 범위 거절이 타입을 잃는 자리는 못 찾음. 이름 붙지 않은 굶김도 없음 — 범위 거절이 아닌 오류에서의 굶김(Guardian/journal 의 범위 국소 상한 거절 포함)은 코드 주석 「굶음 결정」과 R3 로 명명돼 있음.
2. **CAS 창** — 보호가 **항상 섬**. `refuseStaleBucketUsage` 는 admission 트랜잭션 안에서 원장을 다시 읽고 범위 자기 버킷마다 대조함. 같은 시장의 두 범위는 market 버킷을 늘 공유하므로 섹터 · 종목이 달라도 STALE 이고, 순서를 바꿔도 대칭임. 첫 레그가 원장에서 거절되면 쓰기 0 이라 둘째는 정상 발급, held 0 이면 소비가 없음. cap 검사(`smallestRecordedBucketLimit`)도 원장 기준이라 우회하는 모양을 못 찾음. 두-레그 시험은 held 를 **직접** 잼: 사본에서 `checkLimits` 의 held 합산을 끄면 FAIL(둘째 파도 placed=000660,005930). 대조군 오류는 `OPEN_EXPOSURE would reach 1602 … already held 801, new 801`. 거절 원인을 시계 진행으로 바꾸면 원래 단언이 **불통과**(원인은 FX 미해결) — 원인을 가리지 않음.
3. **위조 · 틀린 범위 권한** — 경로 없음. 키 정규화는 양쪽이 `NewOwnerKey` 한 함수이고, 다른 세대는 키에 세대가 들어가며 캠페인 CAS 의 `gen+1` 이 하나만 통과시킴. 시장 칸 `bundle` 의 생산 소비자는 0(rg 기준 — dispatch 도 `forScope` 로 바뀜). 위험은 범위 대조(:294-296)가 이중 방어임: X08 변이를 사본에서 재현했고 CAUGHT. 승격의 「하나라도」는 주문 경로 검사를 대체하지 않음(dispatch 가 범위마다 보호 · 진입 관문 재관측). 남는 것은 화면 쪽 P2(#3).
4. **토글 OFF = upstream** — 새로 **통과**하는 입력 0. 활성화 없는 시장은 handoff 하나 · 결과 하나 · 적재 하나이고, 개수 관문을 지운 뒤에도 계좌 `len != 1` 갈래와 handoff Capacity=1 이 남음. 통화 · 위험 세대 · 결과 권한 · 승격 · projection 이 같은 값임을 코드로 대조함. 새로 **거절**하는 입력: lineage 가 `NewOwnerKey` 에 실패하는(세대 0 · 256B 초과) 제안을 계좌 · 위험 적재기가 새로 미준비로 둠. 보수 방향이고, 주문은 이미 6.2 봉인이 막던 입력이라 worker · projection 표시만 달라짐(R5 계열, T). 사본 전체 스위트: 무태그 1025 pass / 태그 1190 pass. 각 2 fail 은 a111 원본 경로 시험이 **내 `-trimpath` 플래그** 때문에 깨진 것이고, 플래그 없이 다시 돌리니 둘 다 PASS(로트와 무관).

### 남은 위험
- #1 의 수리는 High-risk 경로(1차 레그 · 위험 · 계좌 적재기) 편집이라 FLM · 변이 대상임.
- 생산 영향은 R1(스키마 핀)과 서명 활성화 0 이 겹쳐 오늘 0. 활성화 로트 전에 반드시 닫아야 함.

Recommendation: BLOCK — #1(적재 단계 오류를 범위 거절로 접음)을 오류 분류 운반 + E1 회귀 시험 + 오분류 변이 CAUGHT 로 닫고 주석 두 곳을 정정한 뒤 재리뷰. #2 · #3 은 같은 로트에서 고치거나 이름 붙여 이월.
