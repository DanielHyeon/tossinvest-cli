# Tasks: add-candidate-discovery

각 `[T]`는 RED → GREEN → REFACTOR → VERIFY를 거친다. 이 change는 읽기 전용이므로 실계좌
mutating 단계가 없다 — 검증은 fixture와 주입 clock으로 한다.

## 1. 기반 — 관측과 영속

- [x] 1.1 [T] `Observation`·`Candidate` 타입 + 파생값 표시. 원천이 준 값과 계산값을 구분.
  → `candidate.go`. `Reported`가 원천의 말, 파생은 별도. 모든 decimal은 문자열이고 빈 값은
  0이 아니라 미측정 — `DayHigh`가 0으로 저장되면 `near_high`가 전부 통과한다.
- [x] 1.2 [T] 후보 저장소(SQLite, 원장과 **별개 파일·별개 락**). ext4 강제·fuseblk 거부는
  원장 규칙을 따른다. RED: 재시작 후 이전 프로세스의 관측을 읽지 못함 → GREEN.
  → `store.go`(`candidates.db`), `fsguard.go`. 원장 규칙은 import가 아니라 **복사 + 표류
  테스트**다(`fsguard_drift_test.go`가 `internal/journal/fsguard.go`를 소스로 읽어 대조).
  import하면 6.1 격리가 깨진다.
- [x] 1.3 [T] 저장소가 엔진 원장 락을 잡지 않음을 테스트로 고정 — 스캔이 도는 동안
  `tossctl engine run`의 flock 획득이 막히지 않는다.
  → `TestAScanningStoreDoesNotBlockTheEngineFromStarting`.
- [x] 1.4 [T] 후보 수명(활성/냉각/만료)과 **재진입 시 `first_seen_at` 유지**.
  RED: 냉각 중 재진입이 `first_seen_at`을 갱신 → GREEN(유지). 만료 후는 새 후보.
  → `Promote`/`Cool`. 만료는 sweeper가 아니라 **읽는 시점의 시각**으로 판정한다 — sweeper가
  돌아야 진행하는 상태는 방금 올라온 프로세스에서 거짓말을 하고, 그때가 바로 묻는 때다.
- [x] 1.5 [T] 리뷰(2026-07-28)가 찾은 §1 결함 수정. 전부 "아무것도 실패하지 않는" 종류였다.
  - **저장 시각이 순서 키가 아니었다(P0)**. `RFC3339Nano`는 소수부 뒤 0을 떼므로 TEXT 비교에서
    `…00Z` > `…00.5Z`(`Z`=0x5A > `.`=0x2E)다. `PruneObservations`가 **컷오프보다 새 행을
    지웠고**, `Observations(since)`가 요청한 창 안의 행을 빠뜨렸으며, "오래된 것 먼저"가 아니었다.
    §3·§4가 바로 이 두 질의 위에서 가속도와 veto를 계산한다. → 고정폭 stamp +
    `TestTheStoredInstantOrdersLikeTime`(기존 테스트는 전부 정각이라 이 분기에 닿지 못했다).
  - **아무도 냉각시키지 않은 후보가 영원히 활성**이었다. 냉각은 기록하는 쪽이 살아 있어야 하고
    스캔은 죽는다. → `stateAt`가 `last_seen_at` 기준 암묵 냉각(`DefaultStalenessTTL` 10분,
    최장 백오프 300초의 2배).
  - **과거 시각으로 살아 있는 후보를 만료시킬 수 있었다** — D1 우회의 뒷문. → `Promote`·`Cool`이
    마지막 관측보다 이른 시각을 거부.
  - **만료 후 새 후보가 죽은 후보의 `sources`·완전성·`degraded`를 상속**했다. → 만료 시 초기화.
  - **표류 테스트가 상수만 대조**하고 허용 **집합**은 대조하지 않았다 — 원장이 파일시스템을
    거부 목록으로 옮겨도 상수는 그대로다(테스트 헤더가 막겠다고 쓴 바로 그 방향).
    → 원장 소스에서 `allowedMagics`/`allowedNames`/`AllowedFilesystems`를 파싱해 양방향 대조.
    변이 테스트로 확인: `MagicExt` 제거 시 FAIL.
  - `_txlock=immediate` 누락(두 프로세스에서 `SQLITE_BUSY_SNAPSHOT`), DSN 경로 미이스케이프
    (`#` 포함 디렉터리가 다른 파일을 연다), `nearestExistingDir("")`가 **CWD를 승인**,
    `NoteSources`가 대상 없이 조용히 성공(D4가 감추지 않겠다는 바로 그 사실을 감춘다) +
    원천을 병합하지 않고 덮어씀, `rank>0`인데 `rank_total==0`(D8 백분위가 `+Inf`),
    `nullable`이 공백을 다듬지 않음.
  - `internal/testenv`의 `FixedFSProber` 허용을 **파일 단위 → 선언 단위**로 좁힘 — 정의 파일
    안에서의 호출도 잡는다.

## 2. 원천

- [x] 2.1 [T] 공식 Open API 어댑터 — `Rankings`(`MARKET_TRADING_AMOUNT`,
  `MARKET_TRADING_VOLUME`, `TOP_GAINERS`), `MarketInvestorTrading`, 지표 캔들.
  **공식 원천만으로 후보가 산출됨**을 테스트로 고정.
- [x] 2.2 [T] WTS 어댑터 — 인기 순위, 투자자 순매수 순위, 테마, AI 시그널, 스크리너.
  가산 원천이며 없어도 산출이 멈추지 않는다.
- [x] 2.3 [T] 원천 강등 — 일부 실패 시 남은 원천으로 계속하고 `sources`·`completeness`·
  `degraded`를 후보와 스캔 결과에 기록. RED: WTS 전부 실패 시 후보 0건 → GREEN.
- [x] 2.4 [T] 원천 병합 — 같은 `(market, symbol)`을 올린 여러 원천이 후보 **하나**가 되고
  각 원천이 그 하나를 지지한 것으로 기록된다.
- [x] 2.5 [T] hybrid 우회를 강등으로 기록(D14) — **적용 범위는 `official_prices`뿐**.
  랭킹·캔들·투자자는 hybrid에서 **official-only이고 우회가 없다**([client.go:185-220]) —
  최초 D14는 이것을 확인하지 않고 썼다. RED: 우회로 채운 `prices` 후보가 공식이 응답한 후보와
  구분되지 않는다 → GREEN. 랭킹 관측에 `ViaFallback`이 찍히면 그것은 우회가 아니라 버그다.
- [x] 2.5b [T] 랭킹 429는 **강등이 아니라 결손**으로 기록(D14 결정 2). 우회라는 완충재가 없으므로
  429 한 번에 원천이 통째로 사라진다. RED: 랭킹 실패가 `completeness`에 남지 않는다 → GREEN.
- [x] 2.6 [T] rate 예산 잔량 헤더(`X-RateLimit-Limit`/`Remaining`/`Reset`) 기록(D13 결정 2).
  지금은 어디서도 읽지 않아 429가 나야 예산을 안다. RED: 200 응답에서 잔량이 유실 → GREEN.
  **D15 결정**: `internal/official`에 읽기 전용 헤더 노출을 **가산**한다. proposal Impact의
  "수정 없음"을 그에 맞게 고친다 — 헤더는 429에서 sentinel로 바뀌며 버려지고
  ([execgw/retry.go:183]), 유일한 대안인 RoundTripper는 `internal/execgw`라 §6.1 격리가 막는다.
- [x] 2.7 [T] 원천별 간격 — 시장이 아니라 원천에 붙고, 느린 원천이 빠른 원천을 묶지 않는다.
  기본값 WTS 5초(하한 3초) / 공식 랭킹 15초(하한 5초). → `Schedule`(정책과 due 판정만;
  루프는 §5가 갖는다). `YieldToEngine`이 spec R7의 최상위 우선순위에 처음으로 동사를 준다.
- [x] 2.8 [T] **냉각은 침묵이 아니라 증거로 한다**(§2 구현 중 발견). 후보를 냉각하려면
  **그 후보를 올렸던 원천이 전부 이번 스캔에 응답**해야 한다. 느슨하게 하면 랭킹 429 한 번이
  전 watchlist를 냉각시키고 냉각 시계가 그것을 만료시켜, **일시적 rate limit이 우리가 일찍
  발견했다는 기록 전체를 지운다.** 아무것도 실패하지 않으면서.
  → `coolAbsent` + `TestAScanDoesNotCoolASymbolItDidNotLookFor`,
  `TestOneSurvivingSupporterIsNotEnoughToCool`.
- [x] 2.9 [T] 시장별 panel 구성 — 그 시장을 볼 수 없는 원천은 **panel에서 빠진다**.
  "응답했는데 목록에 없다"가 냉각의 근거이므로, 구조적으로 못 보는 원천이 panel에 있으면
  매 스캔 그 시장 후보를 전부 냉각시킨다. → `candidatesrc.Panel`.

## 3. 시간축 지표

- [x] 3.1 [T] 주입 clock 배선. 테스트가 시간 경과를 만든다.
- [x] 3.2 [T] 거래대금·거래량 변화율 — **경과 시간으로 정규화**. RED: 관측당 차분을 쓰면
  1분 간격과 10분 간격이 같은 값 → GREEN(1분 쪽이 크다).
- [x] 3.3 [T] 가속도를 **구간 속도의 비**로 산출(D9). 두 구간의 실제 초를 함께 저장.
  RED: 합의 비를 쓰면 backoff로 3배 늘어난 창이 가속도 3배를 만든다 → GREEN(1배 부근).
- [x] 3.4 [T] `WARMING_UP` — 직전 구간이 없으면 가속도를 산출하지 않는다. RED: 첫 관측이
  무한/과대 가속도로 임계를 통과 → GREEN(미산출, 임계 통과 아님).
- [x] 3.4b [T] **못 쓰는 분모도 미산출**(D9 정정). `prior_rate == 0`(거래 정지·저유동성),
  직전 구간 실제 초 == 0(한 스캔에서 여러 원천이 같은 순간 보고), 누적값 델타가 음수(세션
  경계·원천 재기동). RED: 셋 다 `+Inf`/부호 반전으로 임계를 통과 → GREEN(미산출).
  미산출 사유를 `WARMING_UP`과 **구분해 기록** — 전자가 많으면 원천을, 후자면 아무것도 안 고친다.
- [x] 3.4c [T] rate·가속도 계열은 `(market, symbol, **source**)`별(D9 정정). RED: 원천 섞어
  차분하면 컴파일도 되고 그럴듯한 숫자도 나온다 → GREEN(원천 필터 읽기 + 원천 간 차분 금지).
- [x] 3.5 [T] shadow 임계 5종(1.3/1.5/1.8/2.0/2.5) 통과 여부를 전부 기록. 판정은 하지 않는다.
- [x] 3.6 [T] 최초 발견가 대비 확장률.
- [x] 3.7 [T] 일중 range position(고점 근접도) — 캔들 기반, 캔들 부재 시 **미측정**으로 표시.
- [x] 3.8 [T] 순위 백분위 변화 + `newly_listed`를 **별개 사실**로 기록(D8). RED: 신규 진입이
  직전 최하위로 채워져 순위 속도를 자동 통과 → GREEN(합성하지 않는다).

- [x] 3.9 [T] 리뷰(2026-07-28)가 찾은 §3 결함 수정. 두 P0 모두 "아무것도 실패하지 않는" 종류였다.
  - **아무도 값을 넣지 않은 가속도가 `Computed()==true`(P0)**. `level.go`는 같은 규칙을 자기
    헤더에 쓰고 테스트로 고정했는데 같은 날 같은 섹션의 `metrics.go`가 정반대였다 — D10을
    강제하려고 만든 섹션 안에서 D10이 깨져 있었다. → 값이 있어야 측정으로 세고 `NOT_MEASURED`
    사유 신설 + 파일 내 전 지표에 zero-value 테스트.
  - **`TallyCrossings`가 그런 후보를 양쪽 집계에서 떨어뜨렸다(P0)** — 입력 1개 중 0개 계상.
    화면의 기본 읽기가 "대부분 미확인"에서 "다 측정했고 조용하다"로 뒤집힌다. → `Total` 도입
    (`Total == Computed + Σ NotComputed` 단언), 중복 임계 1회만, 미등록 임계 무시.
  - **확장률 기준선이 조용히 옮겨가고 `Measured: true`로 답했다(P0)** — D17이 예측한 미측정보다
    한 단계 나쁘다. → D17 구현(`first_price`/`_at`/`_source` + schema v2 ALTER 사다리),
    `MeasureExpansion`이 **저장된** 기준선을 받는다. 슬라이스 폴백은 두지 않았다 — 두면 칼럼을
    안 읽은 호출자에게 결함이 그대로 남는다.
  - **`DistancePct`가 float64라 4.3의 경계를 지킬 수 없었다** — 정확히 2.00%인 두 자리 소수
    가격쌍 20,000개 중 **9,598개(48%)**가 float에서 2.0 미만. `metrics.go`가 이름을 대며 거부한
    산술을 veto 경계를 거는 쪽이 쓰고 있었다. → `big.Rat`.
  - **세 원인이 `WARMING_UP` 하나로 뭉쳤다** — 원천 혼합·읽기 실패·관측 없음. 하필 "아무것도 안
    고쳐도 된다"를 뜻하는 통이다. → `NO_OBSERVATIONS`/`MIXED_SERIES`/`READINGS_ALL_LATER` 분리 +
    거부된 series는 sticky.
  - **`RankChange`의 `window`가 아무것도 제한하지 않았다**(30초·40시간이 같은 값·같은 무사유).
    → `Seconds` 기록 + 나이 상한 10분(`DefaultStalenessTTL`과 같은 유도: 최장 백오프 300초의 2배).
    상한은 백스톱이고 보호는 `Seconds`다.
  - **벽시계 가드가 8개 진입점 중 1개만 봤다** — `Since`/`Until`/`After`/`Tick`/`NewTimer`/
    `NewTicker`/`AfterFunc`/`Sleep`/별칭/dot-import 전부 통과했다. §5의 watch 루프와 429 백오프가
    쓸 철자들이다. → import를 **경로로** 해석하고 이름 집합 대조. 11종 변이 전부 FAIL 확인.
  - 중복 행이 사유 없이 가속도를 바꿨다(동률→오래된 쪽 해소로 무력화), 음수·0 window가
    시장 조건 사유를 받았다(`WINDOW_NOT_POSITIVE` 신설), `formatDecimal`의 무-과장 규칙에
    테스트가 없었다(truncate→round 변이가 패키지 전체를 통과했다), `SourceObservations`가
    `source`를 정규화하지 않았다, 읽을 수 없는 가격 하나가 `extended`를 48시간 껐다,
    같은 순간 동률에서 `LastPrice`가 가장 먼저 삽입된 행으로 해소됐다.

## 4. 추격 위험 veto

- [x] 4.1 [T] `seen_late` — 최초 관측 시점에 이미 임계를 크게 초과.
- [x] 4.2 [T] `extended` — 최초 발견가 대비 확장률 상한 초과.
- [x] 4.2b [T] **너무 늦은 기준선은 `false`가 아니라 `unmeasured`**(D17 판정). 마이그레이션
      백필과 prune은 최초 발견가보다 늦은 기준선을 저장할 수 있고, 늦은 기준선은 확장률을
      **과소평가**해 `extended`를 `false` 쪽으로 민다 — veto를 끄는 방향이다.
      RED: 기준선이 최초 발견보다 한참 늦은 후보가 "안 늘어남"으로 통과 → GREEN(미측정).
      경계는 `DefaultStalenessTTL`(10분).
- [x] 4.3 [T] `near_high` — **일중 고가까지의 거리가 하한 미만**(`distance_pct < 2.0`).
      경계 고정: 1.99% true / 2.00% false / 2.01% false.
      비교는 `<`다. RED: 산문의 "상한 초과"를 그대로 구현해 `>`를 쓰면 **고점에서 먼 후보에
      veto가 걸리고 고점에 붙은 후보가 통과한다** → GREEN(부호 정정). D3이 거부권이라 부른
      판정이 정확히 반대로 도는 자리이므로 경계 세 점을 전부 고정한다.
- [x] 4.4 [T] **3-상태 veto**(D10) — `true`/`false`/`unmeasured`. RED: 캔들을 못 받은 후보의
  `near_high`가 `false`로 저장되어 "veto 통과"로 집계 → GREEN(`unmeasured`는 통과가 아니다).
  이 change에서 **가장 중요한 테스트**다 — 상세조회는 초당 5회·종목당 1회이므로 미측정이
  예외가 아니라 평시다.
- [x] 4.5 [T] veto가 점수와 **분리**됨을 고정. RED: 가속도 최상위가 `near_high`를 상쇄해
  통과 → GREEN(거부권 유지).
- [x] 4.6 [T] veto 후보도 저장·보고된다. 늦게 본 비율을 셀 수 있어야 한다.
- [x] 4.7 [T] **입력 나이 상한** — 오래된 캔들로 만든 `near_high=false`는 `unmeasured`다.
  캔들이 초당 5회라 한 번 얻은 고가는 재사용된다. 3분 된 캔들로 만든 "위험하지 않음"은 측정이
  아니라 기억이고, 그 구분이 사라지면 4.4가 막은 실패가 한 단계 아래에서 재현된다.

- [x] 4.8 [T] §4 구현이 찾은 스펙 구멍 넷을 설계에 반영(D18·D19·D20 + D17 나머지 절반).
  - **세 veto 중 둘에 임계값이 저장소 어디에도 없었다**(D18). 구현자가 지어내기를 거부한 것이
    옳다 — D6은 출처 없는 정책 숫자를 금지한다. 근거가 생길 때까지 `seen_late`·`extended`는
    veto하지 않고 **그림자로만 기록**한다. §3.5가 가속도에 이미 쓰는 규칙이다.
  - **입력 나이 상한의 비대칭이 수명 상한과 반대**였다(D19). 수명 쪽은 거부가 비싸서 600초에
    있고, veto 입력은 **수락이 비싸다** — 오래된 캔들의 "위험하지 않음"이 veto를 끈다.
    tasks.md 4.7이 3분을 이미 기억이라 부르는데 600초는 그것을 수락한다. → 2분, 유도는 후퇴가
    아니라 캔들 전달 빈도. `extended`의 최신가에도 적용(같은 방향의 과소평가).
  - **죽은 후보의 관측이 산 후보를 대신 답할 수 있었다**(D20). 148위로 사라졌다 5위로 돌아온
    종목이 "일찍 잡았다"로 기록된다 — `seen_late`가 정확히 반대로 답한다. → 대칭 ±10분 창.
  - **D17이 가격만 보고 순위를 놓쳤다.** 최초 순위도 prunable한 곳에만 있어 `seen_late`가
    가장 오래 달린 후보에서 무너진다. → `first_rank`/`_total`/`_source` 칼럼(schema v3).

- [x] 4.9 [T] D17 나머지 절반 구현 — `first_rank`·`first_rank_total`·`first_rank_source`
  (schema v3, `first_price`와 같은 수명 규칙). 들어가면 D20의 대칭 창은 백스톱으로 내려간다.
  → `first_rank_at`을 **하나 더** 뒀다(설계는 3칼럼). `first_price_at`이 있는 이유가 그대로
  적용되고, 그것이 없으면 저장된 순위의 정체를 읽는 쪽에서 검사할 수 없다 — issues.md 1.
  `MeasureFirstSighting`이 `MeasureExpansion`처럼 저장값을 인자로 받고, 슬라이스의 행으로
  대체하지 않는다(D17이 가격에서 막은 바로 그 대체). 마이그레이션 백필은 `first_seen_at`
  **이후** 10분 안의 행만 쓴다 — 순위는 틀릴 안전한 방향이 없어서 D17의 무조건 백필을 쓸 수
  없고, 대칭 창으로 백필하면 D20의 두 형태를 저장 시점에 구워 넣는다(issues.md 5).
  RED: 실제 저장소로 재현한 "간격의 148위 행" → `measured=true rank=148/150 seen_late=clear`.
  ※ `Collect`는 아직 `NoteFirstRank`를 부르지 않는다 — `NoteFirstPrice`와 같은 상태이고
  둘 다 §5(5.1)에서 배선한다. issues.md 4.

- [x] 4.11 [T] §4 적대적 리뷰(2026-07-28)가 찾은 18번째 강제 통과 경로 수정.
  - **비양수 임계가 목록 전체를 measured-and-clear로 만들었다(P0)**. `"0"`·`"-0"`·`"0.0"`·
    `"-1"`이 전부 파싱되어 `distance < 0`이 모든 후보에서 false다. §5가 손잡이를
    `strconv.FormatFloat`로 렌더하면 부재한 YAML 키가 `""`가 아니라 `"0"`이 되고,
    `near_high`는 D18상 유일하게 승인된 임계를 가진 veto다. → `THRESHOLD_NOT_POSITIVE`
    (D18 정정). 세 코드 모두에서 `clear`→`unmeasured`로만 바뀐다.
  - **기준선 정체성 가드가 한쪽뿐이었다(P1-1)**. `AssessExtended`는 늦은 기준선만 거부했는데
    `MeasureExpansion`은 기준선을 **뒤로** 옮기고 원 관측은 후보 수명보다 48시간 오래 산다.
    죽은 삶의 30,000이 산 삶의 기준선이 되어 두 배가 된 후보가 `-33.33%`·`clear`로 읽혔다.
    → `|FirstAt − FirstSeenAt|` 대칭 + `BASELINE_TOO_EARLY`.
  - **`VetoCodes`가 exported mutable slice였다(P1-3)**. 네 번째 코드는 거부하지만 코드를
    **빼는** 방향은 무방비다 — `VetoCodes = nil`이면 zero `Chase`가 통과하고, 한 칸을
    덮어쓰면 seen_late RAISED인 `Chase`가 통과한다. → `[3]VetoCode` 값 타입 + 길이·내용 고정.
  - `TestAnAbsentInputAgeLimitIsTheDefaultAndNotNoLimit`가 **가드를 지워도 초록**이었다
    (0/−1h 상한은 모든 나이를 거부하므로 §3 `ZERO_ELAPSED_SECONDS`와 같은 형태) →
    기본값 **안쪽**도 단언. `TestTheVetoCannotSeeAScoreToBeOffsetBy`가 최상위 필드만 봤다 →
    재귀 + 사유를 적은 숫자 allowlist(미사용 항목도 실패). `PercentileExceeds`가
    `Rank > RankTotal`(200/150 → −33%, clear)을 막지 않았다 → 거부.
    D19 커버리지 산술 과장(600후보 **및** 백오프 2단계)을 `veto.go` 주석에서 제거 — 상수는 유지.
- [x] 4.10 [T] `seen_late`·`extended`의 **그림자 밴드** 기록(D18). veto하지 않는다.
  밴드는 측정 도구이므로 출처 없이 정해도 되지만(§3.5 선례), **veto 임계는 실측 후 사람 승인**이다.
  → `band.go`. `seen_late` 50/70/80/90/95 백분위, `extended` 10/20/30/50/100%.
  밴드가 veto가 되는 것을 **구조로** 막는다: ① `ShadowBand`에 `Dangerous`/`Clear`/`Raised`가
  없다(§4 리뷰가 §5에 남긴 한 줄짜리 실수를 쓸 철자가 없다), ② 교차 비교가 `>=`로 veto의 `>`와
  **다르다**(밴드 경계는 히스토그램 칸이지 판정선이 아니다), ③ 밴드 3종이
  `TestTheVetoCannotSeeAScoreToBeOffsetBy`의 score 집합에 들어가 `VetoInputs` 폐포에서 도달
  불가, ④ `TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`가 `VetoState`/`Chase`를
  반환하는 함수 10개를 AST로 걸어 밴드 식별자 언급을 거부한다. 변이 5종 확인(M3~M7).
  밴드의 입력 검사는 veto와 **같다** — 죽은 삶의 기준선·늦은 기준선·오래된 최신가는 전부
  그럴듯한 틀린 숫자를 만들고, 그 숫자가 바로 임계를 도출할 데이터셋에 들어간다. 두 사슬은
  따로 쓰고 `TestABandNamesTheSameMissingInputTheVetoWould`가 fsguard 표류 테스트와 같은
  방식으로 묶는다(변이 M4로 확인).

## 5. 사람이 보는 표면

- [x] 5.0 [T] **issues.md 4의 배선** — `Collect`가 `NoteFirstPrice`·`NoteFirstRank`를 함께 부른다.
  칼럼만 있고 writer가 없으면 `extended`는 영원히 `NO_BASELINE`, `seen_late`는 `NO_FIRST_RANK`이고
  **아무것도 실패하지 않는다**. → `recordFirsts`. panel 순서로 가격/순위를 각각 하나씩 고르고,
  D16의 write 예산 때문에 `Store.Summaries`로 **한 번** 읽어 아직 없는 후보에만 쓴다(순진한
  배선은 심볼당 tick당 IMMEDIATE 트랜잭션 2개다). 읽기는 promote **뒤**에 해야 한다 — `Promote`가
  만료 시 두 칼럼을 지우므로 앞에서 읽으면 새 삶이 죽은 삶의 기준선을 그대로 쓴다(변이 M2로 확인).
  쓰기 실패는 `Rejected`에 이름으로 남고 스캔은 계속한다.
- [x] 5.1 [T] `tossctl candidate scan` — 1회 스캔. `mutating: false`. 출력에 **shadow 임계별
  교차 수**(veto 통과가 아니다 — 말을 갈라 쓴다), **미측정 veto를 가진 후보 수**, 강등·백오프.
  → `cmd/tossctl/candidate.go` + `internal/candidate/watch.go`의 `Cycle`/`Assess`.
  표는 미측정을 통과 **위에** 놓고, `passed 0` 옆에 D18 때문에 구조적으로 0이라는 문장을 붙인다
  (JSON은 `veto.passed_note`). 그 문장이 없으면 다음 사람이 "고장"으로 읽고, D18이 지목한
  **유일하게 틀린 수리**가 `THRESHOLD_ABSENT`를 통과로 세는 것이다. 변이 M17로 확인.
  JSON은 `CycleResult`를 직접 marshal하지 않는다 — `VetoState`의 필드가 비공개라 `{}`로 나가고,
  `{}`를 읽는 소비자는 그 타입이 막으려던 결론을 정확히 낸다.
- [x] 5.2 [T] `tossctl candidate watch` — 반복 스캔. 간격 하한 강제. 보존 정리와 WAL 회수를
  스캔 주기 안에서 집행(D16) — 정리는 옵션이 아니다.
  → `Watch`/`WatchInterval`(기본 15초·하한 3초 = 계약의 가장 빠른 원천 floor). 대기는
  **주입 clock의 `Sleep`**이다 — 벽시계 가드 11종이 이 루프를 위해 만들어졌고, 변이 M13이
  `time.Sleep`을 넣으면 가드가 잡는다. `runCleanup`이 매 주기 prune + `wal_checkpoint(TRUNCATE)`.
- [x] 5.3 [T] 실계좌 검증 runlock이 살아 있으면 `watch` 시작 거부.
  → 저장소를 열기 **전에** 검사한다(예산에 대한 거절이 파일을 남기지 않게). stale lock은 무시 —
  크래시한 검증의 대가는 거절 1회이지 손으로 파일을 지울 때까지의 정지가 아니다.
- [x] 5.3b [T] **엔진이 돌면 간격을 늘린다**(spec R7 정정). 우선순위 표에서 최상위인 엔진에만
  지키는 동사가 없었다. RED: 엔진 실행 중에도 간격이 그대로 → GREEN(물러나고 그 사실을 기록).
  → `Cycle`이 매 주기 `Schedule.YieldToEngine`을 호출하고 `CycleResult.EngineYield`·`Intervals`에
  기록한다. 엔진 감지는 `internal/enginelock`을 import하지 않고 **함수로 주입**한다 — 폐포가
  자물쇠이므로(D7). 변이 M11로 확인.
- [x] 5.3c [T] **여유 공간 하한**(D16). 아래로 내려가면 발굴이 먼저 쓰기를 멈추고 말한다.
  RED: 공간이 차면 원장 쓰기가 ENOSPC → GREEN(발굴이 먼저 멈춘다). 원장과 같은 파일시스템이다.
  → `FSInfo.FreeBytes`/`FreeMeasured` + `Store.FreeSpace()`(매 주기 재측정 — 마운트 판정은 한 번
  끝나지만 남은 공간은 아니다) + `DefaultFreeSpaceFloor` 512MiB. **읽기도 멈춘다** — 저장할 수
  없는 행을 위해 엔진과 공유하는 rate 예산을 쓰는 것이므로(변이 M9). **측정 못 한 공간은 공간이
  아니다** — 프로브 실패도 halt다(변이 M8). 정리를 공간 검사 **앞**에 두어 하한을 넘긴 저장소가
  스스로 돌아올 길을 남긴다.
- [x] 5.4 [T] 429 백오프 + 백오프 사실을 스캔 결과에 기록.
  → `Backoff`/`Retreat`, 사다리 30→60→120→300초(꼭대기에서 고정; `DefaultStalenessTTL`과
  `MaxRankPriorAge`가 마지막 rung의 2배로 유도되므로 사다리가 바뀌면 두 상한이 조용히 움직인다).
  후퇴 중인 원천은 panel에서 **빼지 않고 실패로 남긴다** — 빼면 `coverageAnswered`가 "없어진 원천"으로
  읽어 그 원천만 올린 후보를 냉각시키고, 냉각 시계가 만료시켜 §2.8이 막은 파괴에 도달한다(변이 M10).
  자기 후퇴는 `ErrHeldByBackoff`로 `ErrRateLimited`와 구분한다 — 같이 세면 미문서화 RANKING 한도의
  유일한 측정값이 우리 사다리 길이의 함수가 된다.
- [x] 5.0 [T] §3·§4가 §5에 남긴 배선 — `Collect`가 `NoteFirstPrice`·`NoteFirstRank`를 호출한다.
      RED: 스캔이 가격과 순위를 보고도 저장하지 않아 `extended`가 `NO_BASELINE`,
      `seen_late`가 `NO_FIRST_RANK`로 영원히 남는다 → GREEN. 만료 시 칼럼이 초기화되므로
      읽기는 `Promote` **뒤에** 온다.
- [x] 5.8 [T] 만료된 후보 요약 정리(D11 집행자). 만료 후 `DefaultRawRetention` 경과분을 지운다 —
      원 관측이 사라지면 그 요약은 아무것과도 조인되지 않는다. RED: 집행자가 없어 요약이 무한히
      쌓인다 → GREEN. `watch` 주기 안에서 집행한다(D16과 같은 이유).
      → `Store.PruneExpiredCandidates(at, grace)` + `runCleanup`이 매 주기 호출,
      `CleanupReport.Summaries`와 CLI `retention` 줄에 계상. 만료 시각은 칼럼이 아니라
      `stateAt`의 유도값이므로 **암묵 냉각 분기까지 두 갈래로** 쓴다 — `cooled_at`만 보는 문장은
      스캔이 죽어서 남은 후보(이 집행자가 존재하는 이유 그 자체)에 영영 닿지 않는다.
      grace 0/음수는 **하한이 아니라 기본값** `DefaultRawRetention`이다.
      변이 3종 확인: 0=무제한(M1) / 암묵 분기 제거(M2) / 냉각 TTL 누락(M3) 전부 FAIL.
      ※ 기존 `TestACycleEnforcesRetentionItselfRatherThanLeavingItToACaller`의 fixture는
      "3일 전에 한 번 승격하고 만 후보"였고 그것은 이 sweep의 정당한 대상이다. 주장(원 관측
      prune이 요약을 가져가지 않는다)을 유지하려면 후보가 **살아 있어야** 하므로 fixture를
      staleness 간격으로 계속 승격하는 형태로 고쳤다 — D17이 말한 "자기 원 관측보다 오래 사는
      후보"가 정확히 그 경우다.
- [x] 5.5 [T] 콘솔 **`/signals`** — 후보 목록, 최초 발견 시각, 지표, veto 사유, 원천·완전성.
  **읽기 전용 화면**. 확인 문자열 타이핑 같은 마찰을 넣지 않는다.
  경로 이름은 StockOS의 `/signals`에 맞춘다(사용자 지시 2026-07-28).
  → `internal/console/signals.go` + `templates_signals.go`, `Options.Signals SignalsReader`,
  `cmd/tossctl/console.go`의 `consoleSignalsSeam`. 화면은 **스캔을 일으키지 않는다** —
  `candidate.Assess`만 부르므로 rate 예산 소비가 0이고, 열린 탭이 두 번째 발굴자가 되지 않는다
  (spec R7 · D14: 랭킹 429는 원천 상실). 저장소는 렌더 1회 동안만 열었다 닫는다 —
  `Store.Checkpoint`의 주석이 지목한 "긴 수명 독자"가 바로 이 화면이고, 붙잡고 있으면
  watch 주기의 WAL 회수가 조용히 꺼진다.
  `readOnly` wrapper는 붙이지 않았다: console.go가 그 wrapper는 `/orders` 한 곳에서만
  load-bearing이라고 적어 뒀고, 이 경로는 account verb를 하나도 쓰지 않는다.
- [x] 5.6 [T] **미측정이 통과처럼 보이지 않는다**(새 spec Requirement). veto 사유가 하나도
  없는 행은 "모두 측정했고 안전"과 "한 번도 확인 안 함" 둘 다일 수 있고, 후자가 평시다.
  RED: 세 veto가 모두 `unmeasured`인 후보가 통과 후보와 같게 보인다 → GREEN.
  → 세 veto가 **각자의 칸과 각자의 단어**를 갖는다(위험/측정·안전/미측정+사유 코드).
  "사유가 하나도 없는 행"이라는 상태가 렌더에 존재하지 않는다. 행 판정은 `Chase.Vetoed()` →
  `Chase.Passed()` → 미확인 순이며 `n / 3 사유 측정`을 함께 적는다.
  `Passed`는 D18 때문에 구조적으로 0이고, 숫자 **옆에** 그 문장을 붙인다(칸을 숨기지 않는다).
  집계의 분모는 `후보`(VetoTally)와 `계열`(CrossingTally)로 갈라 적는다(D21).
  변이 확인: naive `!Dangerous()`+"아무것도 안 터졌으니 통과"(M-A, 콘솔 테스트와
  `TestNoConsumerReadsAVetoThroughItsDroppableSecondReturn` 동시 FAIL) / passedNote 제거(M-B) /
  0일 때 칸 숨김(M-C) / 계열↔후보 라벨 교환(M-D) / 사유를 `Reason()` 대신 원시 필드로(M-H').
- [x] 5.7 [T] 강등을 **빠진 원천 이름**으로 말한다. 사유 없는 불리언은 대응할 수 없는 표시다.
  → `SignalsPanel.Missing []candidate.SourceFailure`를 이름·사유·429 여부로 렌더한다.
  **강등인데 이름이 없으면 그 사실을 결함으로 적는다** — 빈 목록은 "복구됐다"로 읽힌다.
  변이 확인: 목록 제거(M-E) / `Unnamed` 무력화(M-F) / 미측정 panel을 "강등 없음"으로(M-G).
  ※ 오늘의 배선에서는 콘솔 프로세스가 스캔을 돌지 않아 `Missing`을 채울 원천이 없다.
  panel은 `seam_unwired` + "이름은 `tossctl candidate scan` 출력에 있다"로 렌더한다.
  D7(console-operator-overview)의 미체결 패널과 같은 형태이며, 근본 해소는 issues.md 9.

### §5 리뷰 반영 (2026-07-28, 별도 컨텍스트 Eng 리뷰)

- [x] 5.9 [T] **P0 — 일정상 아무것도 due하지 않은 turn이 루프를 죽이고 그것을 시장 실패로
      보고했다.** `Cycle`이 빈 panel을 `Collect`에 넘기면 `ErrNoSourceAnswered`가 나오고
      (panel이 비었다는 사실에 대해서는 참이고 시장에 대해서는 거짓이다), `Watch`→`OnError`→
      false로 루프가 끝난다. 끝나면 아무도 승격하지 않고 `stateAt`이 `last_seen_at + 10분`에
      암묵 냉각, 그 30분 뒤 만료 → **약 40분 안에 저장소의 모든 `first_seen_at`이 사라진다**.
      §2.8과 D17이 막으려던 파괴를 §5가 추가한 기능이 일으켰고, 운영자에게는 "no source
      answered"(= 브로커·시장 문제)로 보인다.
      도달 경로 3종 전부 평시다 — ① 엔진 양보(`engineYieldFactor=2`가 15초를 30초로 만드는데
      tick은 `DefaultWatchInterval` 15초 그대로. `--market US`는 `candidatesrc.Panel`이 WTS를
      빼므로 **무조건** 공식 3종 단일 간격) ② `--interval`이 모든 원천의 간격보다 짧을 때
      (flag help가 "floor of 3s"로 권한다) ③ 시각이 한 간격 이상 뒤로 갈 때(NTP·resume;
      `Schedule.due`가 `!at.Before(last.Add(every))`라 전 원천이 동시에 not-due).
      → 빈 panel은 오류가 아니라 `CycleResult.Quiet`로 기록하고 계속한다. 정리·공간 검사·평가는
      그대로 돈다(평가를 건너뛰면 조용한 turn이 "시장이 비었다"로 보고된다).
      → tick과 일정의 관계를 고정한다: `Schedule.UntilNextDue` + `watchWait` =
      **운영자 tick보다 빠르지 않고, 어떤 원천이든 읽을 수 있게 되는 시점보다 빠르지 않다.**
      `UntilNextDue`는 panel 전체의 최솟값이므로 늘어나는 상한이 가장 빠른 원천에 묶인다.
      → `tossctl`의 `OnError`가 `ErrNoSourceAnswered`에서 루프를 유지한다. 기존 주석이
      "Cycle only returns an error for the second"라고 적었는데 사실이 아니었고, panel 전체가
      429 후퇴 중이면 시장 장애 없이 같은 자리에 도달한다.
      테스트 4종 + 변이 3종(빈 panel 분기 제거 / `NotAsked` 제거 / `watchWait`을 tick 고정).
- [x] 5.10 [T] **P1-1 — 아직 부를 때가 아닌 원천이, 물어보지도 않은 시장에 대해 보증했다.**
      `Cycle`은 due한 것만 `Collect`에 넘기고 `Collect`는 넘겨받은 것으로 냉각 자격 panel을
      만들었다. `coverageAnswered`가 not-due 원천을 "없어진 원천"으로 읽고 응답 요구를 그만둔다.
      재현: `official_rankings_top_gainers`(not due) + `wts_popular`(due·응답·목록에서 빠짐)로
      올라온 후보가 **다른 지지 원천을 한 번도 묻지 않은 스캔에 의해 냉각된다.**
      엔진이 켜지는 순간 KR 배선에서 격 tick마다 도달한다(공식 30초 / WTS 10초 대 15초 tick).
      → `heldSource`에 준 "있는데 실패함" 모양을 같은 이유로 확장한다: `CollectOptions.NotAsked`가
      냉각 자격 집합(`heard`)에만 들어가고 `Attempted`·`Missing`·`Degraded`는 건드리지 않는다
      (not-due는 강등이 아니다 — 강등으로 적으면 멀쩡한 원천을 찾으러 간다).
      §2-5는 그대로다: panel에 아예 없는 원천은 여전히 "없어진 원천"이고 냉각을 막지 않는다.
- [x] 5.11 [T] **P1-2 — `/signals`가 원천 0개 panel을 "전 원천 응답"으로 렌더했다.**
      `signalsPanelFrom`이 `Known`만 보고 분기해 `응답 / 시도 0 / 0` 위에 `강등: 없음 — 전 원천
      응답`을 **ok 클래스로** 그렸다. veto보다 한 단계 위이고 표 각주가 바로 그 문장을 믿으라고
      말한다. → `signalsPanelView.NothingAttempted` + 템플릿 3분기. 강등이 아니라는 것도 함께
      적는다. 변이: `NothingAttempted`를 항상 false로 → FAIL.
- [x] 5.12 [T] **P2 — 사다리 리셋의 배선과 0행 200.** `for id := range scan.Budgets {
      off.Recovered(id) }`를 지워도 저장소 전체 스위트가 통과했고, "429가 아닌 실패 전부에서
      리셋"도 통과했다(후퇴 중인 원천이 매 tick 자기 후퇴를 지운다).
      `TestARecoveredSourceStartsTheLadderAgainFromTheBottom`은 단위만 보면서 이름으로 배선을
      주장했다 → `TestBackoffRecoveredBringsTheLadderBackToItsBottomRung`으로 개명하고 배선
      테스트 2종을 추가했다.
      **0행 200은 사다리를 리셋하지 않는다**(결정): `Budgets[id]`가 "응답했으나 행이 없어
      커버리지로 세지 않는다"는 분류 **앞에서** 쓰이고 있었다 — 이 파일이 다른 어디에서도
      증거로 취급하지 않는 판독이 유일하게 회복으로 세어졌다. 빈 목록 반환은 rate limit에 걸린
      서비스가 흔히 하는 load shedding이므로, 서버가 덜 달라고 말하는 순간 폴링을 원상 복구하는
      셈이 된다. → `ScanResult.Vouched`(행을 실은 응답만)가 회복 신호다. 예산 헤더는 D13 결정 2의
      유일한 측정값이므로 응답 전체에 대해 그대로 기록한다. 변이 3종 전부 FAIL.
- [x] 5.13 [T] **P2 — 나머지 표면.**
      ① CLI `passedNote`가 무조건이라 `passed 7 (structurally 0: …)`이 가능했다 → 콘솔이 이미
      쓰던 문장 교체 분기를 이식(`passedUnexpected`). 칸을 비우지는 않는다 — 0이 아니게 되는
      경로가 D18이 지목한 유일하게 틀린 수리이므로 그것이 화면을 더 조용하게 만들면 안 된다.
      ② `scan`/`watch`가 저장소를 **닫지 않았다** — `candidateWiring`이 doc과 달리 `func() {}`를
      돌려줬고, 중단된 `watch`가 원장 옆에 checkpoint 안 된 WAL을 남겼다 → 팩토리가 release를
      함께 돌려준다(fixture는 소유권을 유지한 채 호출 횟수만 센다).
      ③ `/signals`가 15초마다 `context.Background()`로 저장소를 열고 마이그레이션했다 →
      요청 context를 팩토리까지 관통시킨다.
      ④ `consoleSignalsMarket`이 `tallyVerdicts`를 손으로 복제했다(네 번째 밴드를 한쪽에만
      추가하면 다른 쪽에 안 나타난다) → `candidate.TallyVerdicts` 하나로 모으고, 콘솔 배선이
      `Tally*` 생성자를 직접 부르지 않는다는 정적 테스트로 고정.
      ⑤ 정적 소비자 가드의 **실제 경계**: 변이로 확인한 6종 중 4종(괄호·`== false`·`!= true`·
      method value/expression)은 값싸게 넓혔고, 나머지 2종(결과를 변수에 담는 형태, pair를
      method value로 받는 형태 — 후자는 `chase.NearHigh` 필드 읽기와 타입 없이 구분 불가)은
      `isolation_test.go`가 하는 것처럼 **테스트 자신의 doc 주석에 경계로 적었다.** 검출기
      자체의 표 테스트를 추가했다.
      ⑥ `candidateVerifyLockPath` 오류가 runlock 게이트를 조용히 건너뛰었다 → fail-open은
      유지하되(읽기 전용 루프가 데이터 디렉터리 오설정으로 못 도는 것이 더 나쁘다) 이유를
      stderr에 적는다. 말없이 건너뛴 게이트는 통과한 게이트와 구분되지 않는다.

## 6. 격리와 게이트

- [x] 6.1 [T] 의존성 격리 테스트 — `internal/candidate`의 import 그래프에 주문 실행 경로가
  없음을 고정. 후보에서 주문으로 가는 코드가 컴파일되지 않아야 한다.
  → `isolation_test.go`. 변이 6종(직접·간접·`_test.go`·컴파일 안 되는 build tag 뒤·외부 test
  패키지·새 하위 패키지)이 전부 FAIL 후 되돌리면 PASS. 금지 목록은 현재 트리 대비 완전하다.
  **다만 이 보증의 경계는 문장보다 좁다** — 모듈 내부 import가 필요 없는 주문 경로
  (`http.Post(".../api/v1/orders")` 한 줄)는 네 테스트를 전부 통과하고, 배선 **방향**
  (candidatesrc·`cmd/`·미래 레인이 `Candidate`를 읽고 `execgw`를 부르는 것)은 제약하지 않는다.
  D7에 적었다. 실제 자물쇠는 금지 목록이 아니라 **의존 폐포가 `{internal/clock}`임을 고정하는
  테스트**다 — 목록에 있든 없든 새 간선이면 실패한다.
- [ ] 6.2 Function Logic Map + `check_analysis.py`
- [ ] 6.3 PM registry allowlist + fixture 등록
- [x] 6.4 `docs/ROADMAP.md` Phase 3 갱신 — T3.1을 이 change로 분리하고 착수 조건을 고친다
      → 착수 조건 분리 근거를 Phase 3 머리에 명시(읽기 전용이라 주문 경로 신뢰가 선행 조건이
      될 이유가 없고, 시간축 데이터는 늦게 쌓기 시작할수록 레인이 붙을 때 근거가 없다).
- [ ] 6.5 `make sdd-sync && make sdd-check && make gate CHANGE=add-candidate-discovery`
- [ ] 6.6 독립 리뷰(gstack) — 특히 격리 테스트가 실제로 컴파일 차단을 하는지

## 착수 조건 — 충족 (2026-07-28)

세 가지 모두 확정됐다. 값은 design.md "결정된 계약값", 근거는 D8~D14.

1. **임계 기본값** — 공통 순위 하드 컷 없음(D8), 가속도는 구간 속도의 비(D9),
   `near_high` 2.0% 3-상태(D10). 레인별 임계는 기록만 하고 판정은 T3.2가 한다.
2. **스캔 간격** — 시장이 아니라 원천별(D13). WTS 5초, 공식 랭킹 15초 + 실측 기록.
3. **KR·US** — 처음부터 동시 ON, `policy_version`을 시장별로 분리.

## 후속 — 이 change의 범위 밖 (체크박스가 아니다)

게이트는 미완료 체크박스를 0개로 요구한다. 아래는 **하지 않기로 결정한 것**이지 남은 일이
아니므로 체크박스로 두지 않는다. 근거는 design.md에 있고, 하려면 새 change가 필요하다.

- **스캔 기록 영속(schema v4)** — 빠진 원천의 **이름**이 `/signals`에 직접 도달하게 한다(D22).
  지금은 한 사이클의 in-memory `ScanResult.Missing`에만 있어 화면이 CLI로 안내한다.
  **콘솔이 요청마다 `Cycle`을 돌리는 우회는 금지다** — 열린 탭이 두 번째 발굴자가 되어
  D12의 우선순위(엔진 > 검증 > 발굴)를 뒤집고, D14가 말한 "429 한 번에 원천 전손"을 자초한다.
- **`seen_late`·`extended`의 veto 임계** — 저장소 어디에도 값이 없고 지어내지 않는다(D18).
  그림자 밴드가 한 달치 실측을 쌓으면 **사람이 승인**한다. 그때까지 두 veto는 판정하지 않으며,
  `Chase.Passed()`가 구조적으로 0인 것은 고장이 아니다.
