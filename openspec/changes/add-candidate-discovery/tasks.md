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
- [ ] 4.10 [T] `seen_late`·`extended`의 **그림자 밴드** 기록(D18). veto하지 않는다.
  밴드는 측정 도구이므로 출처 없이 정해도 되지만(§3.5 선례), **veto 임계는 실측 후 사람 승인**이다.

## 5. 사람이 보는 표면

- [ ] 5.1 [T] `tossctl candidate scan` — 1회 스캔. `mutating: false`. 출력에 **shadow 임계별
  교차 수**(veto 통과가 아니다 — 말을 갈라 쓴다), **미측정 veto를 가진 후보 수**, 강등·백오프.
- [ ] 5.2 [T] `tossctl candidate watch` — 반복 스캔. 간격 하한 강제. 보존 정리와 WAL 회수를
  스캔 주기 안에서 집행(D16) — 정리는 옵션이 아니다.
- [ ] 5.3 [T] 실계좌 검증 runlock이 살아 있으면 `watch` 시작 거부.
- [ ] 5.3b [T] **엔진이 돌면 간격을 늘린다**(spec R7 정정). 우선순위 표에서 최상위인 엔진에만
  지키는 동사가 없었다. RED: 엔진 실행 중에도 간격이 그대로 → GREEN(물러나고 그 사실을 기록).
- [ ] 5.3c [T] **여유 공간 하한**(D16). 아래로 내려가면 발굴이 먼저 쓰기를 멈추고 말한다.
  RED: 공간이 차면 원장 쓰기가 ENOSPC → GREEN(발굴이 먼저 멈춘다). 원장과 같은 파일시스템이다.
- [ ] 5.4 [T] 429 백오프 + 백오프 사실을 스캔 결과에 기록.
- [ ] 5.5 [T] 콘솔 **`/signals`** — 후보 목록, 최초 발견 시각, 지표, veto 사유, 원천·완전성.
  **읽기 전용 화면**. 확인 문자열 타이핑 같은 마찰을 넣지 않는다.
  경로 이름은 StockOS의 `/signals`에 맞춘다(사용자 지시 2026-07-28).
- [ ] 5.6 [T] **미측정이 통과처럼 보이지 않는다**(새 spec Requirement). veto 사유가 하나도
  없는 행은 "모두 측정했고 안전"과 "한 번도 확인 안 함" 둘 다일 수 있고, 후자가 평시다.
  RED: 세 veto가 모두 `unmeasured`인 후보가 통과 후보와 같게 보인다 → GREEN.
- [ ] 5.7 [T] 강등을 **빠진 원천 이름**으로 말한다. 사유 없는 불리언은 대응할 수 없는 표시다.

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
