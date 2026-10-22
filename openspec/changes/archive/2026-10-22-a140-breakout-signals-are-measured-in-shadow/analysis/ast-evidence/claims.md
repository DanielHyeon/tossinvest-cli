# a140 분기 주장의 AST 근거

proposal · design 이 근거로 쓰는 **기존 함수 내부 분기**는 손으로 읽지 않고 이 디렉터리의 `tools/logic-map` 산출물로 열거한다
(.claude/CLAUDE.md 필수 진입 5 · docs/WORKFLOW.md 「Function Logic Map」). 이 change 는 기존 함수를 **편집하지 않는다** —
여기 있는 것은 FLM 번들이 아니라 주장의 근거이고, 그래서 `analysis/function-logic/` 이 아니라 이 자리에 둔다.

- 기준: HEAD `050d092f`(a138 docs 커밋 — 아래 세 파일의 Go 는 `a1d66f13` 과 같음, `git diff --stat a1d66f13..050d092f` 에 `.go` 0).
- 생성: `GOFLAGS=-trimpath go run ./tools/logic-map --file <file> --func <func> > <name>.ast.json` (2026-10-15).
- 줄 번호는 AST 의 `at.line` 이다.

| 함수 (파일 · sha256 앞 12) | 분기 / return | 주장에 쓰는 열거 |
|---|---|---|
| `strategyproposal.buildLaneInput` (production.go `b6e54b502e50`) | 21 / 19 | **B1@439 → return@440**: 레인 ID 가 KR·US breakout 이면 다른 조건 없이 `ErrBreakoutEvidenceUnavailable`. 첫 분기라 앞에 우회 갈래가 없다 |
| `strategyproposal.LoadProductionAuthorityBatch` (production.go `b6e54b502e50`) | 21 / 9 | B1@311: 설정 · 대상 결손 → 불가. **B3@321 → @322**: 매니페스트 파일 digest ≠ `config.ManifestDigest` → 불가. B4@325: 서명 검증 실패 → 불가. **B5@329 → @330**: `evidence.db` 를 **읽기 전용**으로 못 열면 불가. B15@380: 스냅숏 재생 실패 = 고장. **B16@384**: `buildLaneInput` 오류 = 정상 부재(`continue`) — 그래서 breakout 범위는 고장이 아니라 조용한 부재가 된다 |
| `engine.strategyProposalAuthorityLoader.collectMarket` (strategy_proposal_authority.go `a34982912840`) | 15 / 15 | **B8@370 → @371**: `load` 오류 또는 `batch.ManifestDigest() != digest` → `AUTHORITY_INVALID`. 이 `digest` 는 353 줄에서 `getenv(TOSSOS_STRATEGY_PROPOSAL_{KR,US}_MANIFEST_SHA256)` 로 읽는다 — 프로세스 환경이므로 새 매니페스트는 재시작 없이 통과하지 못한다. (return 15 중 하나(@321)는 `fail` 클로저 안의 것; 함수 머리 주석의 「14」는 그것을 세지 않은 수) |
| `engine.dispatchStrategyMarketHandoffs` (strategy_market_handoff_delivery.go `44ed84069d52`) | 4 / 5 | B1@38: 캠페인 CAS 읽기 오류 → 그 오류. **B2@41 → @42**: 캠페인이 claimed 이거나 FLAT·CLOSED 가 아니면 조용히 `nil`. 그 뒤 44 줄 `dispatcher.dispatch` 가 유일한 dispatch 호출. B3@46: 일일 손실 측정 불가 기록. B4@49: lease 소진 → `nil` |
| `engine.deliverEachStrategyHandoff` (같은 파일) | 5 / 4 | B3@74: `*strategyScopeRefusal` 이면 모아 두고 다음 범위로. **B4@82 → @84**: 일일 손실 차단 · 측정 불가는 뒤 범위를 전달하지 않고 주기를 끝냄(앞서 모은 범위 거절만 보고). B5@87 → @88: 그 밖의 오류는 첫 오류에서 멈춤 |
| `engine.strategyDispatchCycle.dispatch` (strategy_dispatch_cycle.go `72021ef372fe`) | 22 / 21 | B1@78 ~ B11@148: 읽기 전용 경계(보호 관측 96 줄 · 활성화 하한 · 진입 게이트 관측 · dispatch owner). **151 줄 `cycle.firstLeg.admit`** 이 Guardian 첫 다리 승인(q_final · 캠페인 · 예약 커밋)이고 B12@152 가 그 거절. 그 앞의 갈래는 원장 쓰기가 없다 |
| `engine.strategyFirstLegPlaceIntent` (같은 파일) | 4 / 5 | 분기는 수량 · 가격 경계뿐(B1~B4). `Side: "buy"` 는 리터럴(268 줄) — 이 경로로는 매도 의도를 만들 수 없다 |
| `breakoutlane.validStructuralEvidenceInput` (machine.go `ab03efc63446`) | 4 / 4 | B1@146: 봉 15 개 미만 · 512 초과 · **`ATRMinor == 0`** 등 → 거짓. 마지막 return@162 = `quoteValid(v.Quote) && fxValid(v.FX)` — **스냅숏은 단계와 무관하게 구조적으로 유효한 호가 · FX 봉인을 요구**한다 |
| `breakoutlane.evaluateFresh` (machine.go `ab03efc63446`) | 20 / 9 | B1~B3@40~44: 처음 15 봉으로 저항(최고가) · 레인지 저점. B6@57: 돌파 = close 버퍼 통과 ∧ RVOL ≥ 1.5× ∧ 윗꼬리 ≤ 35%. B12@90: close < 레인지 저점 → INVALIDATED. B15@101: 되돌림 중 저항 아래 close ∧ `VolumeExpanded` → INVALIDATED. **B18@111 → @112**: ARMED 가 아니면 호가를 보지 않고 반환. **B19@114**(`validateQuote`)·117 줄 `size` 는 ARMED 뒤에만 — 호가 신선도 · 스프레드 · 진입 괴리 · 사이징은 ARMED 단계에서만 판정된다 |
| `engine.StrategyEntrySupervisor.runMarket` (strategy_entry_supervisor.go `6f7c999a13c4`) | 17 / 11 | B11@953: 주기 오류 없음 → 다음 주기. **B12@956(refreshOnly)**: 오류를 `recordSwallowedCycleError`(967 줄 호출)로 세고 잠그지 않음 — 중앙 무결성만 B13@974 로 진입 차단. 그 밖(effective 조립)은 B14@980 다음 **984 줄 `latchMarket`** — 주기 오류가 시장 잠금이 된다 |

## 이 열거가 받치는 문장

1. 「breakout 레인은 생산 입력이 하나도 만들어지지 않는다」 — buildLaneInput B1 + LoadProductionAuthorityBatch B16.
2. 「생산 제안 입력은 환경에 핀된 서명 매니페스트이고 주기마다 새 신호를 실을 수 없다」 — collectMarket B8 + LoadProductionAuthorityBatch B3.
3. 「생산은 `evidence.db` 를 읽기만 한다」 — LoadProductionAuthorityBatch B5(쓰기 열기는 생산 호출 0, `../code-context/raw/grep-production-refs.txt`).
4. 「Jev seam 후보 자리(전달 몸통의 CAS 확인 뒤 · dispatch 앞)는 Guardian 승인 · 원장 쓰기 · 브로커 호출 전이고, 매수만 지난다」 —
   dispatchStrategyMarketHandoffs B2 · 44 줄 + dispatch 151 줄 + strategyFirstLegPlaceIntent 268 줄.
5. 「범위 거절(`strategyScopeRefusal`)은 주기 오류로 올라간다 — Jev 거부를 그 타입으로 돌리면 refreshOnly 조립에서는 삼킨 오류 계수,
   effective 조립에서는 시장 잠금이 된다」 — deliverEachStrategyHandoff B3 · 반환 @90·@92 + runMarket B12@956(967 줄) · 984 줄.
   그래서 Jev 거부 · 불가는 a137 의 일일 손실 차단처럼 **오류가 아닌 타입 결과**로 주기를 끝내야 한다(design.md S3 개요).
6. 「ARMED 전 단계는 호가의 **신선도**를 보지 않는다 — 구조적으로 유효한 봉인만 있으면 된다」 — validStructuralEvidenceInput @162 +
   evaluateFresh B18@111 · B19@114. 그래서 도구는 세션 시작에 종목마다 호가 한 번(GET 2)을 봉인해 두고, ARMED 에 닿은 setup 에만
   새 호가를 읽어 같은 봉으로 다시 평가한다(design 「측정 방법」). ARMED 판정 자체는 그 묵은 봉인과 무관하다.
7. 「무효화 수준은 오프닝 레인지 저점이다」 — evaluateFresh B12@90(close < 레인지 저점 → INVALIDATED). 가상 손절 권고(D7)의 근거.
