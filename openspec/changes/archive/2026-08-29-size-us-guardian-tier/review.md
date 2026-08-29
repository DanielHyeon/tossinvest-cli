# Review: size-us-guardian-tier

날짜: 2026-07-30 · 위험 등급: **High-risk**

이 change는 사이징 수치를 **완화 방향**으로 옮긴다. §0.9가 보수 방향만 허용하므로,
이 문서의 본론은 왜 그것이 허용되는 경우인지와 완화의 크기를 무엇이 묶고 있는지다.

## Pre-Edit Gate

```text
change id / task id:  size-us-guardian-tier / 1.1-5.5
대상 심볼 (기존 함수 내부 수정, 기계 검증 7건 — 전부 테스트 함수):
  config.TestTheTiersTranscribeStockOS              (limits_test.go)  — 코퍼스 분리
  config.TestTheCeilingIsTheMaxAcrossRegisteredTiers(limits_test.go)  — USD 기대치 2필드
  config.TestRegisteringTheUSTierMovedExactlyTwoCeilings  (신규)
  config.TestTheUSTierMatchesItsRecordedDerivation        (신규)
  config.TestTheMeasuredInstrumentFitsWithHeadroom        (신규)
  config.TestEveryTierWouldStart                          (신규)
  config.TestTheCeilingIsNotTheInterlock                  (신규)
production 편집:
  internal/config/limits.go — var guardianTiers에 원소 1개 + 파일 헤더 주석.
  함수 본문은 한 줄도 바뀌지 않았다.
기존 동작 파악 근거:
  analysis/function-logic-map.md + 기계 검증 7건
  읽은 파일: internal/config/limits.go, internal/risk/{chain,input}.go,
             internal/console/settings_limits.go,
             openspec/specs/{risk-management,operator-console}/spec.md,
             openspec/changes/verify-execution-capability/measurements.md,
             openspec/changes/console-sets-guardian-limits/specs/operator-console/spec.md,
             stockos packages/trading/stockos_trading/risk_profiles.py
upstream 상속 테스트 영향: no — internal/config·console 블록은 TossOS 전용이다
실패 테스트 선행 작성: yes (RED 5건 관측, 아래 기록)
안전 불변식 §0 위반 여부 검토: 통과 — 조항 3·4·6·7을 아래 표에서 명시적으로 다룬다
```

### RED 관측 (구현 전)

```text
[FAIL] TestTheTiersTranscribeStockOS        — returned 4 tiers, want 5
[FAIL] TestTheCeilingIsTheMaxAcrossRegisteredTiers — ceiling(USD) notional 300
[FAIL] TestRegisteringTheUSTierMovedExactlyTwoCeilings — 300 want 500 / 1000 want 1500
[FAIL] TestTheUSTierMatchesItsRecordedDerivation — us-single-name is not registered
[FAIL] TestTheMeasuredInstrumentFitsWithHeadroom — headroom is 0.0%
```

마지막 줄이 이 change의 존재 이유다. 지금 상한은 실측 계측기 1주가 **여유 0으로**
통과하는 자리에 있다.

## §0 대조

| 조항 | 이 change |
|---|---|
| 1 승인 없는 LIVE side effect 금지 | 무관 — 주문 경로 없음. config 파일에 한 바이트도 쓰지 않는다. 코드 상수 하나가 늘 뿐이다 |
| 2 mutating 자동 실행 금지 | 에이전트가 실행한 것은 테스트뿐이다 |
| 3 토글 OFF = upstream 동작 | 파싱·엔진·인터록 무접촉. 게이트 OFF 경로는 한 갈래도 바뀌지 않는다 |
| 4 손절·비상 청산 즉시성 | **통과.** `internal/risk/chain.go:49` — "a reduction is not judged by the entry chain at all". 한도는 `entryChain`의 rung이고 청산은 그 체인을 타지 않는다. 통화 가드도 `checkOrderSize` 안에 있으므로 USD 프리셋을 적용해도 **기존 KRX 포지션의 청산은 닫히지 않는다** |
| 5 High-risk 경로 | 해당(Guardian·사이징) → full TDD + FLM 7건 + 적대적 리뷰 + gate |
| 6 손절·사이징은 보수 방향만 | **이 change의 본론.** 아래 §0.9 절 |
| 7 운영 토글 flip은 사람 | 게이트 ON/OFF·kill switch 무접촉. 이 change는 콘솔 표면도 안 건드린다 |
| 8 시크릿·개인정보 미저장 | 인용한 것은 공개 시세 관측(M49)뿐이고 계좌 잔고·보유는 인용하지 않았다 |
| 9 주문은 공식 Open API만 | 무관 |
| 10 실계좌 자동 테스트 금지 | 전 테스트가 순수 함수 검사다. 네트워크·파일 접근 없음 |

## §0.9 — 완화를 허용하는 근거와 그것을 묶는 것

**완화임을 인정한다.** USD 상한 두 필드가 올라간다(주문 300→500, 노출
1,000→1,500). 정당화가 아니라 **무엇이 묶고 있는지**를 적는다.

### 묶는 것 넷

1. **위 경계는 콘솔이 이미 승인하는 크기다.** 오늘 이 화면은 클릭 한 번으로
   1,000,000 KRW 주문을 허가한다(등록 KRW 티어의 최대). $500 × 2,000 = 1,000,000
   이므로 환율 2,000 아래에서 새 USD 상한은 그보다 **작은** 주문을 허가한다. USD
   축은 이 change 뒤에도 KRW 축보다 엄격하다.
2. **아래 경계는 실측이다.** M49(2026-07-30) TSLA 1주 $300. 상한을 이보다 낮게
   되돌리면 `TestTheMeasuredInstrumentFitsWithHeadroom`이 막는다.
3. **완화가 샐 수 있는 필드는 전부 못으로 박혔다.** 수량 100주·비율 1%·일일 손실
   $50과 KRW 다섯 필드 전부를 `TestRegisteringTheUSTierMovedExactlyTwoCeilings`가
   이름으로 검사한다. "USD 두 필드"라는 승인 범위가 테스트다.
4. **적용은 여전히 사람의 클릭이고 전후 값이 audit에 남는다.** 이 change는 어떤
   config도 쓰지 않는다.

### 완화하지 않으면 무엇이 나빠지는가

이것이 실제 논거다. 상한은 **콘솔 쓰기 경로 전용**이고 기동 인터록에는 상한 개념이
없다(design D4, `TestTheCeilingIsNotTheInterlock`). 즉 $300 천장 아래에서 US 자동
진입을 하려는 운영자에게 남는 유일한 경로는 `config.json`을 손으로 여는 것이고,
**그 경로에는 상한 검사가 아예 없다.**

낮은 상한이 만든 것은 안전이 아니라 우회 압력이다. 천장을 근거 있는 자리로 옮기는
것이 백스톱을 유지하는 방법이다. 이 결론을 스펙 문장으로도 남겼다
(`risk-management` ADDED "티어 상한의 사정거리").

## 사용자 결정과의 대조

| 결정 | 이 change |
|---|---|
| 1회 주문 상한을 설정할 수 있게 (2026-07-30) | USD 주문 상한 $300→$500. 고급 폼에서 그 아래 임의 값 기입 가능 |
| 실측 근거를 갖춘 US 티어를 별도 change로 (2026-07-30) | 이 change 자체. M49를 식별자로 인용 |
| ① `$500 / $1,500 / $75 · 1%` 선택 | 넷은 그대로, **일일 손실만 $75 → $50** — 아래 A2 |
| 타이핑 확인·추가 승인 마찰 금지 (2026-07-27) | 콘솔 표면 무수정. 새 프리셋도 기존 카드와 같은 확인 1회 |

## 적대적 Eng 리뷰

날짜 2026-07-30 · 방식: "이 숫자가 무엇을 약속하는가, 그 약속은 어디서 깨지는가."

### A1. 라벨이 숫자가 지키지 못하는 약속을 했다 — **P2**

초안 라벨은 "미국 **대형주 1주** 실거래"였다. $500 주문 상한은 미국 대형주 상당수의
1주에 미치지 못한다(NFLX·AVGO·BRK 등). 카드 옆에 다섯 값이 함께 찍히므로 숫자는
보이지만, 라벨이 먼저 읽히고 라벨이 틀렸다.

→ 종목 부류를 이름하지 않는다: "미국 **단일 종목** 실거래". 코드 주석에 이유를 남겼다.

### A2. 일일 손실 $75가 이 change가 세우는 규칙을 통과하지 못했다 — **설계 단계에서 교정**

사용자가 고른 ①의 다섯 값 중 넷은 유도가 섰는데 일일 손실 $75만 서지 않았다.
스펙 문장("각 수치가 이미 승인된 다른 통화의 선호 안쪽에 있을 때만 등재")을 쓰는
과정에서 드러났다 — `100,000 KRW ÷ 75 = 1,333`이므로 환율 1,333/USD를 넘으면 $75는
승인된 일일 손실을 초과한다. 주문($500, 2,000에서 깨짐)·노출($1,500, 6,667에서
깨짐)과 달리 이 필드만 봉투 밖이다.

남는 근거는 "US 계열 내부의 비율 일관성"뿐인데 그것은 §0.9의 **명확한 근거**가
아니다. 그리고 아무것도 사지 못한다: `RiskBudget = MaxDailyLoss`이고 $300 주식·5%
손절이면 예산 허용 수량이 $50에서 3주, $75에서 5주인데 **둘 다 주문 금액 상한이
1주에서 먼저 조인다.**

→ `us-small-live`의 $50 유지. 움직이는 USD 필드가 셋에서 **둘**로 줄었고, 새 티어는
손실 축에서 `us-small-live`보다 오히려 엄격하다(노출 대비 3.33% 대 5%).
사용자에게 이 한 값의 변경을 보고한다.

### A3. "승인된 위험 선호"라는 표현이 논증을 흐렸다 — **P3(문서)**

초안 design은 $500의 위 경계를 "승인된 위험 선호"와 비교했다. 그런데
`risk-management`가 승인한 것은 **사용자 미확정 시의 보수 기본값 집합**이지 최대치가
아니다. 느슨한 이름 위에 세운 §0.9 논증은 다음 리뷰어가 정당하게 깰 수 있다.

→ 비교 대상을 정확히 이름한다: **콘솔의 KRW 주문 상한**(등록 KRW 티어의 최대,
값은 1,000,000 KRW). 주장이 좁아지고 검증 가능해졌다 —
`TestTheUSTierMatchesItsRecordedDerivation`의 `parityBreaksAt` 상수가 그 산술을 건다.

### A4. FLM이 실측 테스트에 없는 능력을 부여했다 — **P3(문서)**

초안 FLM은 `TestTheMeasuredInstrumentFitsWithHeadroom`을 두고 "종목 가격이 올라
헤드룸이 사라지면 다시 빨개진다"고 적었다. 틀렸다 — `measuredShare`는 고정 상수이고
이 테스트는 시세를 읽지 않는다. 시장이 움직여도 조용하다.

→ 실제 역할로 고쳐 적었다: **되돌림 방지**다. 빨개지는 경우는 누군가 상한을 낮출
때뿐이고, 가격 변동에 따른 재검토는 사람 몫이다.

### A5. PM 테스트 fixture가 레지스트리를 손으로 미러링한다 — **P3(도구)**

`_registry.yaml`의 `bootstrap_change_allowlist`에 change id를 넣자
`tools/pm/test_generate_master_tracker.py`의 2건이 깨졌다
(`size-us-guardian-tier: stale bootstrap exception`, `generated tracker stale`).

원인은 `fixture()`가 실제 `_registry.yaml`을 **복사**하면서 change 디렉터리 목록은
**하드코딩된 튜플**로 따로 만든다는 것이다. 두 목록이 같은 사실을 두 번 적고 있으므로
change를 등재할 때마다 반드시 어긋난다. 직전 change도 같은 이유로 자기 id를 그 튜플에
손으로 넣었다(line 43).

→ 선례대로 한 줄 추가했다. **고치지는 않았다** — fixture를 레지스트리에서 유도하도록
바꾸는 것은 공용 도구의 설계 변경이고 이 change의 범위 밖이다. 다만 이 결합은
change마다 되풀이되므로 기록해 둔다: 복사한 레지스트리에서 목록을 읽으면
`test_reverse_link_duplicate_change_and_stale_allowlist_are_rejected`는 여전히
동작한다(그 테스트는 fixture 이후에 `stale-change`를 덧붙인다).

### 닫지 않고 남긴 것

- **$500이 항상 도달 가능하지는 않다** (design D6). 체인은 설정 상한보다 **먼저**
  위험 예산을 본다. `RiskBudget = MaxDailyLoss = $50`이므로 손절폭 10%를 넘으면
  $500 주문은 예산에서 먼저 거부된다. 설계대로지만, 그때 봐야 할 것은 상한이 아니라
  손절폭이라는 것을 적어 둔다.
- **환율 2,000/USD를 넘으면 등가 논증이 깨진다.** 코드에 환율을 넣지 않기로 했으므로
  (스펙 SHALL NOT) 이 재검토는 자동으로 촉발되지 않는다. 사람이 해야 한다.
- **US 자동 진입은 여전히 닫혀 있다.** 인터록 6조항 중 6번 ProtectionReady가 남고,
  그것은 보호주문 change 이전에 충족될 수 없다. 이 change는 그 시점에 상한이 걸림돌이
  되지 않게 미리 치운 것이다.
- **`operator-console` "Guardian 한도 설정 화면"을 MODIFIED로 잡았지만 그 요구는
  아직 base spec에 없다** — 미아카이브 change `console-sets-guardian-limits`의 ADDED
  본문이다. `openspec validate --strict`는 통과했고 두 change는 아카이브 순서와
  무관하게 같은 최종 문장으로 수렴한다. 다만 `console-sets-guardian-limits`가 이
  change보다 **먼저** 아카이브되어야 문장이 한 번만 적용된다.

## 검증 결과

| 항목 | 결과 |
|---|---|
| `go test ./...` | **3878 passed**, 0 failed (57 packages) — 직전 3873 대비 신규 5 |
| `make vet` | 통과 |
| `openspec validate --strict` | valid |
| `check_analysis.py` | evidence complete (7건) |
| 상속 테스트 회귀 | 0 — 콘솔 한도 테스트는 `GuardianTiers()`를 순회하므로 새 티어를 자동으로 덮는다 |
| 인터록 판정 집합 | 불변 — `TestTheCeilingIsNotTheInterlock`이 분리를 직접 보인다 |
