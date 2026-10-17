# a112 7.3.1 SHADOW — freeze 급 독립 적대 리뷰 브리프 (설계 단계, 코드 0 — 2026-10-04)

너는 이 설계를 **만들지 않은** 신선한 컨텍스트의 리뷰어다. 저자의 결론을 믿지 말고 코드 · 계약 원문 · 실행으로 반증하라. 이 리뷰는 **코드가 아니라 설계와 계약 개정**을
본다 — 구현 전에 잡아야 할 것: 설계대로 지으면 spec 을 어기거나 노출을 열거나 증명할 수 없는 자리.
판정 형식: P0(설계대로 지으면 실주문 · 노출 상승 · 활성화 ON · dispatch 소유가 가능해짐, 또는 spec SHALL/MUST NOT 위반이 구조적으로 남음) / P1(설계 주장이 코드 사실과 다름 ·
핀이 주장을 못 받침 · 계약 문구와 설계가 어긋남) / P2(문서 · 진단 · 시험 판별력 · 비용). 각 지적에 **근거 좌표(파일:줄) 또는 재현(명령 · 사본 실험)과 결과**를 붙여라 —
근거 없는 지적은 P2 이하.

## 안전 규칙(어기면 리뷰 무효 — 어겼다면 출력 맨 위에 적어라)

- `~/.codex` 아래 어떤 파일도 읽거나 검색하지 마라.
- 읽기 전용. 저장소 파일을 편집하지 마라 — 실험은 `/tmp` 아래 사본(`git -C /mnt/D/Axipient/workspace/TossOS archive c7219640 | tar -x -C <사본>`)에서만.
- 모든 셸은 `set -euo pipefail` 로 시작하고, git 은 `git -C <절대경로>` 로만. 사본에서 git 을 쓰기 전 `test "$(git -C <사본> rev-parse --show-toplevel)" = "<사본>"` 단언.
- 네트워크 · 브로커 · LIVE · 토글 · 엔진 lock · `mutating: true` 명령 금지. 시험은 `go test`(태그 `tossos_testseams` 허용)만.
- 시작할 때와 끝날 때 `git -C /mnt/D/Axipient/workspace/TossOS status --short` 와 `git -C /mnt/D/Axipient/workspace/TossOS rev-parse HEAD` 를 재서 같음을 보고하라.

## 좌표 · 입력(고정)

- 저장소 `/mnt/D/Axipient/workspace/TossOS`, 커밋 좌표 **`c7219640`**(amendment 착지 = 리뷰 대상 계약). 코드 사실은 이 커밋에서 읽어라 — 작업 트리가 아니라.
- 설계 브리프(리뷰 대상): 이 디렉터리의 `design-brief-under-review.md` — sha256 `96baede4e07171df5ef898cb00018f44c20edf420e7144a5609b561cb402d762`(§0 측정 · §1 신뢰 앵커 ·
  §2 형식 · §3 MUST NOT 집행 핀 · §4 재시작 · §5 R2 핀 · §6 8.6 · §7 로트 계획 · **§8 Manager 판정 · §9 소비자 감사**).
- 계약 개정(리뷰 대상): 이 디렉터리의 `amendment-c7219640.patch`(spec four-family-strategy-runtime :89/:91 · design 결정 63 문단 + :246) — 잘리지 않은 diff.
- 정본 계약: `openspec/changes/a112-run-four-strategy-families-independently/specs/four-family-strategy-runtime/spec.md` :88-93, `design.md` 결정 61(:238) · 62 · 63 · :289 · :291.
- 관련 코드(c7219640):
  - 레인 관문 `internal/strategyworker/worker.go` `FamilyWorker.Run`(:193) · `Desired/Effective`(:107-113) · `productionWorker`(:158) · `lane.go` `Lane.Run`(:211).
  - 활성화 적재 `internal/strategyrouter/production_family_activation.go` `LoadProductionFamilyActivation`(:429~) · 결속 · 서술자 검증 · 인코더 · 생성기 `tools/a112-family-activation`.
  - 엔진 관문 `internal/app/engine/strategy_family_activation.go` `familyGateFor` · `loadFamilyActivation` · `collectMarket`(strategy_proposal_authority.go) · 레인 런타임
    `strategy_lane_runtime.go` `evaluate` · 투영 `strategy_lane_projection.go` · `internal/strategyprojection/lanes.go` `validateLane`(:274) · `model.go` `Validate`.
  - R2 핀 `internal/app/engine/a112_shadow_absent_restart_test.go` · OpenAPI `docs/api/openapi-v1.json` `StrategyRuntimeLaneRuntime`.
  - 형제 선례: 활성화 인코더 가드 `internal/strategyrouter/production_family_activation_encoder_guard_test.go`, 되돌림 경로 호출 허용 목록 `TestTheRollbackPathOnlyReads`,
    strategyhandoff 타입 census `mint_census_test.go`.

## 리뷰 질문(공통 — 배역은 프롬프트가 정함)

1. **MUST NOT 둘이 구조로 막히는가.** 설계(§3)의 `FamilyWorker.Shadow` 둘째 문 · 불투명 `FamilyShadow` · 봉투 없는 `ShadowCycle` 로, SHADOW 가 desired/effective/activation
   ON 이나 dispatch capability(조정자 Submit · handoff · lease · 게이트웨이)에 닿는 길이 남는가? 레인 런타임 · 투영 · 원장 · 활성화 파일 어디로든. 제안된 핀(타입 census ·
   호출 허용 목록 · 행동 · 변이)이 그 길을 정말 막는지, 막지 못하는 모양(별칭 · 함수 값 · 같은 패키지 안 비공개 경로 · 제네릭)을 찾아라.
2. **spec 문면 충족.** 개정된 :89 의 SHALL(「digest 핀으로 결속된 server-owned shadow manifest 가 있을 때만 pure evaluation 과 counterfactual projection」)과 시나리오 :91-93,
   그리고 상위 SHALL(새 설치 · migration · restart = OFF/OFF/UNOBSERVED)을 설계가 문면대로 만족하는가? 「레인 단위 반사실만」(Manager 판정 ③) 좁힘이 「counterfactual
   projection」 을 충족하는가? 「pure evaluation」 = 이미 있는 제안 위의 레인 관문 반사실(§0)이라는 해석이 spec 과 맞는가?
3. **amendment 자체.** 결정 63 이 결정 61 의 근거를 SHADOW 에 옮기는 논증이 성립하는가? spec 문구 개정이 다른 요구(예: 「signed」 를 전제한 다른 절 · 시나리오 · design 문장)와
   충돌을 남기는가? 핀 env 이름 · 「server-owned」 의 의미(누가 쓰는가)가 설계와 일치하는가?
4. **재시작.** 매 파도 재독 · 상태 무기록(§4)으로 「process-local 자동 복구 금지」 가 성립하는가? 레인 런타임의 관측 메모리 · 투영 저장소 · 원장 · 잠금 복구 세대 어디든
   SHADOW 가 재시작을 넘는 길이 있는가?
5. **코드 사실(§0 · §9).** 「평가는 활성화와 무관하게 이미 돈다」, 「골든은 runtime 어휘를 열거하지 않는다」, 「OpenAPI 가 열거한다」, 「validateLane 이 모든 읽기 경로의 관문」,
   「RPC 디코더는 모르는 필드를 허용」 — 각각 c7219640 에서 맞는가? 틀리면 설계가 무엇을 놓치는가?
6. **공유 vs 사본(Manager 판정 ②).** 활성화 적재기와 공유할 부분 / 타입으로 가를 부분(결속 집합 · 서술자 스키마)의 경계가 옳은가? 사본이 되면 양쪽을 다 못 박아야 하는데
   설계가 그 핀을 계획하는가?
7. **노출 · 배포.** 생산 shadow 핀 0 이면 동작 변화 0 인가? SHADOW 가 8.6(배포 BLOCKED)이나 A100 ProtectionReady 와 얽히는 자리가 있는가?

## 범위 밖(지적은 하되 P 등급을 매기지 말 것)

- 조정자 반사실 선택(Manager 판정 ③ 으로 범위 밖) · 활성화 로트 선행 ROADMAP 행 · 운영자 표면(활성화 실패 사유).

## 출력

판정 한 줄(APPROVE / APPROVE-WITH-CHANGES / REJECT) + P0/P1/P2 표(지적 · 근거/재현 · 결과 · 제안) + 시도했는데 살아남은 공격(무엇을 해 봤고 왜 막혔는지) + 저장소 무변경
확인 출력. 한국어 또는 영어.
