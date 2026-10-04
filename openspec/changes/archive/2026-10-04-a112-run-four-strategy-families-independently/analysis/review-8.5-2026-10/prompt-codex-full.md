# codex — a112 8.5 독립 적대 리뷰(대상 A · B1 · B2 전부)

**`~/.codex` 아래 어떤 파일도 읽거나 검색하지 마라 — 어겼다면 출력 맨 위에 적어라.** 출력 첫 줄에 너의 신원(모델 · 버전)을 적어라.
읽기 전용 샌드박스. 작업 디렉터리는 `git archive 556c3c1f` 트리(git 이력 없음) — 전/후 비교는 이 디렉터리의
`openspec/changes/a112-run-four-strategy-families-independently/analysis/review-8.5-2026-10/diff-*.patch` 로 하라(대상 A 는 65444341, B 는 0b441270 의 diff). 실행이 안 되면 코드
인용으로 판정하고 그렇다고 적어라. 아래 공통 브리프(`brief.md`)의 대상 · 질문 · 출력 형식을 따른다.
# a112 8.5 독립 적대 리뷰 — 명시 대상 2건 브리프 (초안, 2026-10-04)

너는 이 변경을 **만들지 않은** 신선한 컨텍스트의 리뷰어다. 저자의 결론을 믿지 말고 코드와 실행으로 반증하라.
판정 형식: P0(실주문 · 손절 · 노출 상승 경로를 깨거나 열 수 있음) / P1(판정 · 수락 집합이 의도와 다름, 또는 증거가 주장을 못 받침) /
P2(진단 · 문서 · 시험 판별력). 각 지적에 **재현 명령 또는 변이(overlay/사본)와 그 결과**를 붙여라 — 재현 없는 지적은 P2 이하로만 적는다.

## 안전 규칙(어기면 리뷰 무효 — 어겼다면 출력 맨 위에 적어라)

- `~/.codex` 아래 어떤 파일도 읽거나 검색하지 마라.
- 읽기 전용. 저장소 파일을 편집하지 마라 — 변이 실험은 `go test -overlay` 또는 `/tmp` 아래 사본(`git archive HEAD | tar -x -C <사본>`)에서만.
- 모든 셸은 `set -euo pipefail` 로 시작하고, git 은 `git -C <절대경로>` 로만. 사본에서 git 을 쓰기 전에 `test "$(git -C <사본> rev-parse --show-toplevel)" = "<사본>"`
  을 단언하라(전례: 리뷰 픽스처의 mkdir 실패 뒤 cwd 가 실제 저장소라 git init/commit 이 돈 사고).
- 네트워크 · 브로커 · LIVE · 토글 · 엔진 lock · `mutating: true` 명령 금지. 시험은 `go test`(태그 `tossos_testseams` 허용)만.
- 끝나면 실제 저장소가 손대지 않았음을 직접 재라: `git -C /mnt/D/Axipient/workspace/TossOS status --short` 가 리뷰 전과 같아야 한다.

## 좌표 · 입력(Manager 보강 2026-10-04)

- **전/후는 커밋 좌표로 고정한다.** 대상 A: `65444341^`(전) vs `65444341`(후). 대상 B: `0b441270^`(전) vs `0b441270`(후). 작업 트리 · HEAD 를 「후」로 쓰지 마라 — 그 뒤
  커밋이 같은 파일을 만졌을 수 있다. 동등성 실험은 두 커밋의 **격리 사본**(`git -C <repo> archive <commit> | tar -x -C <사본>`)에서 같은 입력을 돌려 대조한다.
- diff 는 줄이 잘리지 않게 파일로 준다(`rtk proxy git diff` — rtk 요약은 줄을 자른다): 이 디렉터리의 `diff-A-2.3-65444341.patch` ·
  `diff-B-884b-production-0b441270.patch` · `diff-B-884b-tests-0b441270.patch`.
- 저장소: `/mnt/D/Axipient/workspace/TossOS`(브랜치 `feat/a112-four-family-runtime`). 증거 경로는 `openspec/changes/a112-run-four-strategy-families-independently/` 기준.

## 대상 A — 2.3 (b) breakout 1.2 반사실 기록 (커밋 `65444341`, 생산 diff 5 줄)

- 파일 · 함수: `internal/breakoutlane/machine.go` `evaluateFresh` — 범위 뒤 봉 루프(B4) 안, 입장 갈래 B6(입장 시 `break`) **뒤**의 새 갈래 B7:
  close buffer(`BreakoutCloseQualifies`) · wick 상한 통과이면서 `RVOLPPM >= 1_200_000` 인 봉이 있으면 `p.RVOLAt1200000 = true`.
- 저자 주장: 기록 전용 — 단계 · 전이 · 거절 · 수량 · 제안 · `decisionSeal` 불변(`decisionSeal` 은 provenance RVOL 플래그를 해시하지 않음); 입장 경로(B6)와 그 안의
  1.2 · 2.0 · 2.5 기록 줄은 바이트 그대로; 패키지 밖 소비자 0(breakout 미배선 — 생산 breakout 입력은 `ErrBreakoutEvidenceUnavailable` 로 전부 거절).
- 판정 근거: Manager 판정 (b)(design.md 「1.2 반사실의 의미」 — spec :43 · design :151 원문 인용 + (a) 공허 논증).
- 증거: `analysis/measurements/lot-2.3/`(pre-edit · RED · 변이 CF-01~10 CAUGHT · verify) · `analysis/function-logic/internal-breakoutlane--evaluatefresh/` · 시험
  `internal/breakoutlane/a112_rvol_counterfactual_test.go`.
- 적대 질문:
  1. B7 이 **결정**(phase · refusal · transitions · candidate · final · proposalID · seal)을 바꾸는 입력이 있는가? 쌍둥이 비교(같은 봉, RVOL 만 1.0)를 넘어, 1.2 만 통과하는
     봉이 여러 개 · 입장 봉 앞뒤 · 첫 touch 와 겹침 · 범위 경계에서 시험하라.
  2. `decisionSeal` · `snapshotDigest` · `validDecision` 경로에서 provenance RVOL 플래그가 정말 해시 밖인가? prior 재사용(`Evaluate` 의 같은 digest → prior 반환,
     정정 → 보존)에서 B7 의 플래그가 옛 결정과 새 결정 사이에서 어긋나 판정 불일치를 만들 수 있는가?
  3. 리터럴 `1_200_000` 두 자리(입장 경로 · B7) — 골든 `rvol_counterfactual_ppm[0]` 와 묶는 시험은 없고 변이 핀(CF-03/04 · CF-10)만 있다. 이것이 충분한가?
  4. 생산 소비자 0 이 정말인가(`RVOLAt1200000` · breakout `Provenance()` 의 패키지 밖 사용)?

## 대상 B — 8.8.4 로트 B (커밋 `0b441270`, 생산 diff 146 줄 / 2 파일)

### B1. `internal/strategyrouter/production_family_activation.go` — 4-가족 활성화 적재의 거절 사유

- 바뀐 것: 맨 sentinel 반환 18 곳에 필드명 `%w` 래핑; 복합 결속 셋(설정 결속 · 몸통 결속 · 수명)과 서술자 필드 검사는 **분기 하나 그대로** `len(failedFields(...)) != 0`
  (새 함수 `failedFields`) — 어긋난 필드 전부를 한 메시지에; Load 의 `err != nil || digest != pin` 을 두 갈래로(읽기 결함 = 읽기 함수 오류를 `%w` 사슬에 · 핀 불일치 =
  `manifest_digest`) — 분기 9 → 10; 미선언이 맨 앞인 순서 불변; decode 오류는 해석기가 이유를 붙여 감싸고 Load 가 그대로 돌려줌; 서술자 표 대조 셋을 `known &&` 로 묶음.
- 저자 주장: 판정 · 수락 집합 불변(같은 입력이 같은 sentinel 로 거절 — `errors.Is` 보존, `==` 비교 0); 복합 결속의 결정식은 앞 판 OR 의 항과 같다; 항은 전부 순수 비교 ·
  부작용 없는 검사라 무조건 평가해도 같다(단락 평가 의존 0).
- 증거: `analysis/measurements/lot-8.8.4-B/`(pre-edit 여섯 · RED 62 · renumber.txt · 변이 13/13 · named 22 · verify) · 편집 뒤 번들 일곱 · 시험
  `internal/strategyrouter/a112_activation_error_fields_test.go`(실패 모양 40).
- 적대 질문:
  1. **수락 집합**: 편집 전 거절되던 매니페스트가 지금 수락되는(또는 그 반대) 입력이 있는가? 특히 (i) `known &&` 로 묶은 서술자 대조, (ii) 수명 블록의 항 순서
     (해석 실패 시 영값 시각의 비교), (iii) 설정 결속의 `market` 항(앞 판은 `name == ""` 로만 시장을 봤다), (iv) Load B5/B6 분리 — 편집 전 커밋과 편집 뒤를 같은 입력
     집합으로 돌려 수락/거절 표를 대조하라(예: 변이 · 무작위 필드 조합 생성).
  2. **sentinel 판별**: 엔진(`internal/app/engine/strategy_family_activation.go:132`)은 `errors.Is(err, ErrProductionFamilyActivationUndeclared)` 하나로 「기존 경로」를 정한다.
     미선언이 아닌 실패가 Undeclared 로 감싸지거나(→ 활성화가 선언됐는데 기존 경로가 열림 — **노출 쪽 결함**), 미선언이 다른 sentinel 로 바뀌는 입력이 있는가?
     `%w` 두 개를 쓴 `fmt.Errorf("%w: manifest file %s: %w", ErrUnavailable, name, err)` 에서 `err` 가 우연히 Undeclared 를 감쌀 수 있는가?
  3. 메시지에 비밀 · 계좌 정보가 새는가(필드 **값**을 싣는 자리 — `%q` 시장 · 가족 이름 · lane_id · 시각)?
  4. `failedFields` 가 판정 입력이다(비면 결속 통과) — 이름만 모으는 함수가 판정을 지는 구조의 위험(새 항을 목록에 빠뜨리면 그 항의 거절이 사라짐)을 어떤 시험이 막는가?

### B2. `internal/app/engine/strategy_proposal_authority.go` `strategyProposalAuthorityLoader.collectMarket` — 관문 계산 이동

- 바뀐 것: `gate = loader.familyGateFor(ctx, market, schedule, routes, observedAt)` 를 조정 직전에서 ROUTE_NOT_READY 가드 직후로(계산만). 분기 수 불변(15).
- 저자 주장: 13 닫힘의 실패 kind 순서 · FAMILY_GATE_CLOSED 판정 자리 불변; 옮긴 값은 닫힘 갈래가 싣는 활성화와 조정 관문으로만 쓰인다; 두 자리 사이에서 `schedule` ·
  `routes` · `observedAt` · `ctx` 는 변하지 않는다(값 동등); `familyGateFor` 는 읽기 전용(env · 매니페스트 파일 읽기와 검증 · 레인 목록) — 원장 · 브로커 · 토글 쓰기 0;
  ROUTE_NOT_READY 만 영값을 싣는다.
- 증거: `analysis/function-logic/internal-app-engine--strategyproposalauthorityloader.collectmarket/` · 시험 `internal/app/engine/a112_proposal_closure_carriage_test.go`
  (AST census + 닫힘 여덟 × 관문 세 모양) · 변이 E1~E3.
- 적대 질문:
  1. 닫힌 시장이 이제 **검증된 활성화**를 싣는다(앞 판은 영값). 이 활성화가 닫힌 주기에 무엇을 움직이는가 — 레인 관측의 승격(Desired/Effective ON 표시), lease 상한
     (`FamilyActivation.LeaseCeiling`), dispatch 의 ProtectionReady 하한, 투영? 닫힌 시장(항목 0)이 그것으로 주문 · lease · 승격 기록을 만들 수 있는 경로가 있는가?
  2. `familyGateFor` 의 조기 실행: ctx 취소 · 파일 I/O 지연이 앞 판에서는 FX/열쇠 실패로 일찍 닫히던 주기를 이제 붙잡는가(지연 · 교착)? 읽기 결함이 관문을 「되돌림」
     으로 세워(rolledBack) 앞 판과 다른 닫힘 사유를 만드는 입력이 있는가?
  3. 13 닫힘 중 충돌(B11) · 미해결 선택(B14) 은 census 로만 잰다 — 이 둘에서 관문 활성화를 싣는 것이 의도와 맞는가?

## 범위 밖(지적은 하되 P 등급을 매기지 말 것)

- 활성화 실패 사유 · `SwallowedCycleErrors` 의 운영자 표면(snapshot/projection) — ROADMAP 활성화 선행 행으로 이미 등재.
- 공유 읽기 함수 `readProductionRouteFile` 이 OS 원인을 sentinel 하나로 접는 것 — 같은 행.
- SHADOW(7.3.1) · base 재고정 — 사용자 결정 대기.

## 출력

판정 한 줄(SHIP / HOLD / REJECT) + P0/P1/P2 표(지적 · 재현 · 결과 · 제안) + 시도했는데 살아남은 공격(무엇을 해 봤고 왜 막혔는지 — 「안 봤다」와 구별) + 저장소 무변경 확인
출력. 한국어 또는 영어.
