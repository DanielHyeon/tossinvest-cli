# SHADOW freeze voice2 — SubagentHandback 원문 (agent a9cea4756f3b985f0, 저장 2026-10-04)

`~/.codex` 아래는 읽지도 검색하지도 않았음. (저장소 안의 `.codex/` 도 열지 않음.)

**판정: APPROVE-WITH-CHANGES**
- P0 없음. P1 넷은 코드 착수 전에 닫아야 함.
- 근거 좌표는 전부 c7219640 사본 `/tmp/rv2-c7219640` 기준.
- 리뷰 대상 브리프 sha256 `96baede4…d762` 를 실측했고 브리프에 적힌 값과 같음.

---

## 0. 계약 절 ↔ 설계·핀 사상 (보이스 ② 배역)

| spec 절 (개정 후) | 받치는 설계 문장 · 핀 | 빈 곳 · 과잉 해석 |
|---|---|---|
| :89 SHALL — 새 설치·migration·restart = OFF/OFF/UNOBSERVED | §4 매 파도 재독 · 상태 무기록, R2 재시작 시험의 shadow 짝, 골든 기본값 | 핀이 유효하게 남은 채 재시작하면 첫 물결에서 SHADOW 가 다시 나옴. 이 문면과의 관계를 정한 시나리오 · 핀이 없음 (P2-5) |
| :89 SHALL — digest 핀 + server-owned 매니페스트가 있을 때만 허용 | §2 (F1: 0400 · 소유자 · 정규 바이트 · 24h · 폐기 · 다섯 결속, 생성기 · 골든), §4 (미선언 · 못 씀 → SHADOW 없음) | 적재 실패 다섯 모드마다 「SHADOW 0」 을 재는 시험이 §3 · §5 표에 없음 (P2-6) |
| :89 「pure evaluation」 | §0 「새 평가 없음 = 레인 관문 반사실」 | 해석일 뿐임. OFF 가족의 제안은 매니페스트 없이도 이미 평가됨 (실측 사례 A). 해석을 문서에 남겨야 함 (P2-7) |
| :89 「counterfactual projection」 | §3 `runtime=SHADOW` + `shadowOutcome`, Manager ③ 레인 한정 | 입력 출처가 정해지지 않았고, 지금 배관대로면 결과가 비게 됨 (P1-1) |
| MUST NOT — desired/effective/activation ON | §3 ①②③, §9 교차 규칙 | 「ON 우선」 이 desired 기준인지 effective 기준인지 모호함 (P2-1) |
| MUST NOT — dispatch capability | §3 ①~④ | `ShadowCycle` 에 봉투가 없다는 핀은 약함. 입력 운반체에 핀이 없음 (P1-1) |
| MUST NOT — process-local 재시작 복구 | §4 | 시나리오가 개정에서 좁아졌음 (P1-2) |
| 시나리오 :91-93 | §4 의 R2 짝 시험 | 「핀 없음」 만 다루고, 「핀은 있는데 못 씀」 은 빠짐 (P1-2) |

## P1

| # | 지적 | 근거 · 재현 | 제안 |
|---|---|---|---|
| P1-1 | **반사실 입력 출처가 정해지지 않았고, 기존 배관을 쓰면 OFF 레인의 반사실은 구조적으로 언제나 NO_INPUT 임.** spec 의 「counterfactual projection」 은 빈 표본으로만 충족됨. | 레인 런타임 입력은 조정 **뒤** 선택 목록임: `strategy_lane_runtime.go:266-279` (`authority.entries`), `strategy_proposal_authority.go:435-457`. 관문에서 멈춘 제안은 `continue` 로 Submit 앞에서 빠짐 (`strategy_market_coordinator.go:91-97`). 닫힌 시장은 entries 0 (`:311-315`). 한 레인에 입력이 여럿이어도 첫 하나만 받음 (`strategy_lane_runtime.go:230-235`).<br>사본 실험 `zz_rv2_shadow_input_probe_test.go` (태그 시험, PASS):<br>• A: 선언된 KR 시장, 지속형만 ON, 005930 에 지속형+역전형 → 역전형 레인 입력 0, `gated=[DORMANT]`<br>• B: 역전형만 제안 → `FAMILY_GATE_CLOSED`, ON 레인 포함 입력 0<br>• C: 미선언 시장 → 조정 승자만 입력 (005930 은 역전형, 지속형의 005930 은 빠짐) | 입력을 관문 **앞** 배치 제안(coordinateMarketProposals 루프)으로 못 박을 것. 그 운반체는 `dispatchHandoffs` · `ResultAuthority` 가 닿지 않는 별도 타입 · 필드에 두고, 같은 census · 허용 목록을 걸 것.<br>참고: `ShadowCycle` 에 봉투가 없다는 census 만으로는 약함. `Input` 과 `Envelope` 가 같은 구조라 `strategycoordinator.Envelope(input)` 한 식으로 변환됨 (사본 실험 `zz_rv2_conv_test.go` 컴파일 · PASS, `worker.go:166-170` 대 `coordinator.go:104-108`). |
| P1-2 | **개정이 시나리오를 「신뢰 앵커만」 범위 밖으로 좁혔음.** 옛 문구 「signed shadow manifest 없이」 = 유효한 매니페스트가 없음. 새 문구 「shadow manifest 핀 없이」 = 핀 env 가 없음뿐. | 패치 :27-30, spec.md:91-92.<br>실제로 흔한 재시작은 compose override 에 핀이 남고 매니페스트는 만료 · 삭제 · 폐기된 경우인데, 이것이 계약 시나리오에서 빠짐. §4 의 핀 계획도 「매니페스트 없이」 로만 적혀 있어 핀 env 를 둘지 모호함. | :91 을 「핀과 일치하고 수명 안 · 미폐기 · 결속 일치인 shadow manifest 없이」 로 고칠 것. 재시작 핀을 둘로 세울 것(핀 없음 / 핀 있음 + 못 씀). |
| P1-3 | **Manager ② 「공유할 수 있는 것은 공유」 를 하려면 High-risk 활성화 적재기를 편집해야 하는데, §7.2 FLM 목록에 그 함수들이 없음.** 공유 sentinel 이나 분류를 리팩터하면 결정 62 가 뒤집힐 수 있음. | 수명 · 폐기 · 결속 · 정규 등식은 지금 활성화 전용 sentinel 과 함께 함수 안에 박혀 있음: `production_family_activation.go:429-495` · `:509-528` · `:540-624` · `:384-390`.<br>관문은 `errors.Is(err, ErrProductionFamilyActivationUndeclared)` 하나로 미선언 = 기존 경로를 가름 (`strategy_family_activation.go:132`).<br>추출 리팩터가 sentinel 신원을 섞으면, 「선언됐는데 못 씀」 이 기존 경로로 읽혀 꺼 둔 가족이 되살아남 (노출 상승 방향).<br>FLM 목록(§7.2)은 collectMarket · evaluate · 투영 · familyGateFor 뿐임. | 둘 중 하나로 정할 것:<br>(a) 활성화 적재기는 무편집, shadow 는 사본 + 양쪽 AST 핀 (옮겨 적은 코드 교훈)<br>(b) 공유 추출이면 그 네 함수 FLM + shadow 전용 sentinel 타입 + 교차 종류 치환 시험(활성화 바이트를 shadow 핀으로, 그 반대도) |
| P1-4 | **개정 뒤에도 tasks 7.3.1 의 완료 기준이 옛 계약 그대로임.** | `tasks.md:563` 「server-owned signed shadow manifest」 (작업 트리의 미커밋 1줄 diff 도 이 줄은 안 고침). 완료 판정과 리뷰가 서명을 요구하게 됨. | :563 을 개정 :89 문구와 맞출 것. Manager ③ 의 레인 한정 좁힘도 함께 기록할 것. |

## P2

1. **ON 우선 규칙과 교차 규칙이 모호함.**
   - 「ON 우선 · SHADOW 는 OFF 레인만」 의 기준이 desired 인지 effective 인지 정해지지 않음.
   - 적재기는 desired ON / effective OFF 를 받아들임(`production_family_activation.go:609-613`). 생성기는 둘을 같게만 만들지만, 손으로 만든 정규 바이트 매니페스트로는 이 조합에 닿을 수 있음.
   - effective 기준으로 shadow 를 걸면 §9 교차 규칙(SHADOW ⇒ OFF/OFF)이 그 스냅숏 전체를 거절함. 그러면 REST · RPC · 콘솔 화면이 모두 사라짐.
   - 제안: 「desired=OFF ∧ effective=OFF 인 레인만 SHADOW」 로 못 박을 것.
2. **spec :38 의 「SHADOW counterfactual 외 dispatch handoff는 0건」 이 SHADOW 반사실을 dispatch handoff 의 예외처럼 읽힘.** :89 MUST NOT 과 충돌하는 문면임. 레인 한정 좁힘 뒤에는 가리키는 대상도 없음. design :84(raw score 반사실 기록)도 같은 처지.
   - 좁힘 자체는 범위 밖이라 등급을 매기지 않음. 문구만 P2.
   - 제안: 「dispatch handoff 는 0건 (SHADOW 반사실 포함)」.
3. **결정 63 의 「바꾸지 않는 것」 목록이 부정확함.**
   - 24h · 0400 · 소유자 · 정규 등식 · 생성기는 SHADOW 에 처음 생기는 약속인데 「불변」 으로 적힘.
   - 결속 집합(Manager ②: 다섯, ProtectionReady 제외), 폐기, 「선언/미선언 모두 SHADOW 없음」, ON 우선 규칙이 design.md 에 없음. 지금은 analysis 브리프에만 있음.
   - `tools/a112-family-shadow` 는 아직 없음 (`ls tools` 로 확인).
4. **출처(provenance) 기록.** 사용자 결정은 「선택지 1 = a112 안에서 구축」 뿐임(`review.md:6500`). S2 와 spec SHALL 개정은 Manager 채택에 거부권 보고 형태임(브리프 §8 ①).
   - 그런데 결정 63 머리글과 커밋 제목은 「사용자 결정」 으로 읽힘. 결정 61 은 「사람이 2번을 골랐다」 였음.
   - 제안: HANDOFF.md:67 처럼 「Manager 판정 · 사용자 거부권 보고됨」 으로 명기할 것.
5. **유효한 핀이 남은 채 재시작하는 경우가 상위 SHALL 과 어떻게 맞는지 적혀 있지 않음.**
   - 제안 시나리오와 핀: 첫 물결 전에는 UNOBSERVED (`validateLane` 의 미관측 갈래가 이미 요구함, lanes.go:274-290), SHADOW 는 이 프로세스의 재독에서만 나옴, 이전 wave · outcome 은 0.
6. **「매니페스트가 있을 때만」 의 양성 핀이 없음.** 적재 실패 다섯 모드마다 「투영에 SHADOW 0」 을 재는 시험이 §3 · §5 표에 없음.
7. **「pure evaluation」 해석.** OFF 가족의 평가는 오늘 매니페스트 없이 이미 돔(사례 A: 역전형 OFF 제안이 평가된 뒤 DORMANT 로 멈춤). 따라서 :89 의 허용은 새 평가를 더하지 않음. 이 해석을 design 에 명시할 것.
8. **OpenAPI 의 `StrategyRuntimeLaneRuntime` 은 `additionalProperties:false` 임** (전 a112 투영 스키마 동일, python 으로 실측).
   - §5 · §9 의 「older readers 는 가산 필드 무시」 는 저장소 안 RPC 디코더에만 맞음(`transport.go:77-90`). 공개 스키마로는 `shadowOutcome` 가산도 enum 확장과 똑같이 비호환임.
   - 「내부 계약 취급」 판단 기록에 이 점도 넣을 것.
9. **어휘 census 의 판별력.** 상수 census 는 명시적 Ident 타입을 가진 const 만 셈(`a112_shadow_absent_restart_test.go:92-107`).
   - `var` 선언이나 `LaneRuntime("SHADOW")` 리터럴은 빠져나감.
   - 실제로 어휘를 묶는 것은 `validateLane` + OpenAPI 임. 그래서 「정확히 {UNOBSERVED, SHADOW}」 주장은 그 둘에 걸어야 함.
10. **낡은 문구 (grep 범위: change 의 proposal · design · tasks · 네 spec · goldens, `openspec/specs`, `docs/`, `internal/**.go`).**
    - 개정과 어긋나는 것:
      - design.md:242 — 「spec 에서 signed 는 SHADOW 에만」, 현재형으로 거짓이 됨
      - `docs/ROADMAP.md:250` — 「서명 shadow 매니페스트 선행」, ③ 골든 어휘 개정 → Manager ④ 와 반대
      - R2 시험 메시지 `:11` · `:119` — 「signed shadow manifest, golden amendment」
    - 결정 61 부터 이미 낡았던 것: design.md:293 「signed four-family activation」, worker.go:183-185 「ed25519 서명이 막는다」, strategy_entry_supervisor.go:535-537 「ed25519 매니페스트」.
    - proposal.md · strategy-engine · market-aware-scheduler · breakout spec · goldens · operations.md · HANDOFF.md: 「shadow」 · 「signed」 0건.
    - 브리프의 design 좌표 :289 · :291 은 c7219640 에서 :291 · :293 임 (개정이 2줄 밀어냄).
11. **spec SHALL 에 env 이름이 들어감.** env 이름을 바꾸려면 spec 개정이 필요해짐.
    - 미선언과 「선언됐는데 못 씀」 의 SHADOW 의미는 spec 에 없고 브리프에만 있음 (결정 62 의 활성화 판본은 design 에 있음).
12. **선례 인용 하나 더.** `internal/candidate/band.go` 의 ShadowBand 가 「판정 생산 함수는 shadow 를 볼 수 없다」 를 시험(`TestNoFunctionThatProducesAVerdictCanSeeAShadowBand`)으로 지킴. FamilyShadow 와 dispatch 함수 관계에 바로 쓸 수 있는 형제 선례임.

## 시도했지만 막힌 공격

- **shadow ↔ 활성화 바이트 치환.** 활성화 디코더가 `DisallowUnknownFields` + 정규 등식 + 스키마 · 도메인을 봄(`:509-528`, `:545-546`). 따라서 shadow 서술자 바이트는 활성화로 통과하지 못함.
  - 단 반대 방향(shadow 적재기 쪽)은 아직 코드가 없어 미검증. 제안 시험으로 남김 (P1-3).
- **재시작 너머 투영 저장소 잔존.** `strategyprojection.Store` 는 메모리 전용임(store.go 전체). 투영 쪽으로는 재시작을 넘는 길이 없음.
- **8.6 · A100 얽힘.** §6 은 「핀 0 = 동작 변화 0, 8.6 불재개」 이고 결정 63 과도 모순 없음.
- **openspec 검증.** 사본에서 `openspec validate … --strict` → valid.
- **기존 핀 정상 동작.** R2 두 핀 + `TestTheRollbackPathOnlyReads` 를 사본에서 태그 시험으로 돌려 ok.

## 저장소 무변경 측정

- 시작: HEAD `c7219640d08408a650941c62cf0df6efff77c036`. status 는 ` M …a112…/tasks.md`, `?? .reticle-setup-crash.log`, `?? …/review-shadow-freeze-2026-10/`, `?? …/shadow-2026-10/`, `?? w4.log`.
- 끝: **HEAD 동일**(`c7219640…`). a112 tasks.md diff 도 그대로 1줄(+1).
- **단, 끝 status 에 a100 파일 27개 수정 + `?? …a100…/analysis/harness/` 가 새로 나타남.**
  - 내 행위가 아님: 저장소에는 status · rev-parse · diff · show · archive(읽기)만 실행했고, 쓰기는 `/tmp/rv2-c7219640` 사본에만 했음.
  - 해당 파일 mtime 은 전부 2026-10-04 23:17:35.24~.25 KST, 같은 1ms 안에 몰림. 병행 세션의 일괄 쓰기로 보이므로 Manager 가 확인할 것.
- 실험 파일은 사본 안에만 있음: `/tmp/rv2-c7219640/internal/app/engine/zz_rv2_shadow_input_probe_test.go`, `/tmp/rv2-c7219640/internal/strategyworker/zz_rv2_conv_test.go`.
