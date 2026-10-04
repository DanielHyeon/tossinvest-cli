모델: GPT-6 기반 Codex · 세부 버전 미공개 · 「~/.codex 미접근」

**FAIL — 1차 P0/P1: CLOSED 5, PARTIAL 1. 새 P0 0, 새 P1 1.**

대상: 2817064cc1f78c4d0d6abc04211de906a7c3efc5. 설계 단계 종결 판정이다. CLOSED는 아래 특정 1차 지적에 대한 설계 수정의 수용이며, 생산 SHADOW 구현·전체 시장 주기 시험·배포 승인이 아니다. ID는 codex-output.md 표의 P1 등장 순서로 부여했다.

좌표 약칭:
- V2: `openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/design-brief-v2-under-review.md`
- A2: 같은 디렉터리 `amendment-v2-2817064c.patch`의 파일 자체 줄 번호(패치 대상 파일 줄 번호와 구분).
- D: `openspec/changes/a112-run-four-strategy-families-independently/design.md`

V2 SHA-256: `7f1e840c85453fa0739afaca25db61472b5fce03cbd7f4f37e0506bb69ca37bd` — 지정 값 일치.

| 1차 ID · 지적 | 판정 | V2 · amendment 좌표 | 근거 / 재현 |
|---|---|---|---|
| P0 — 봉투 없는 결과형으로 dispatch 미차단, 함수 값 재결속 | CLOSED | V2 §2:38–43, §5:65–69, §8:91–94; A2:17,30,36 | 결과형만으로 증명하던 논증에 타입 해소·전이 호출 폐포·소스 동결·차등 시험을 더했다. `TestRecheckV2RebindingAndFreeze`: `q.Submit`은 MethodVal, `worker.owns`는 FieldVal로 식별; 소스 변이는 digest 불일치. `TestRecheckV2MixedDifferential`: 실제 worker·조정자에서 ON/OFF 혼합 Submit 큐 깊이 [1,1], OFF WOULD_EMIT=1. 전체 시장 주기·handoff·gateway 스파이는 미구현/미측정. |
| P1-1 — census 한계, 같은 router 패키지 주조 | CLOSED | V2 §2:35–43; A2:17,36은 MUST NOT 유지, 패키지/census 수정 자체는 V2에 있음 | 별도 패키지의 실제 FamilyActivation 비공개 필드 주조는 `go test` 컴파일 거절. Unalias 후 중첩 별칭·generic instance 검출. `any`는 여전히 정적 타입 소거이므로 go/types의 완전성을 주장하지 않는다. V2가 도입한 소스 동결과 독립 재고정 리뷰가 잔여 주조를 담당해야 한다. |
| P1-2 — 관문/중재 뒤 입력의 반사실 대상 유실 | CLOSED | V2 §0:10–15, §4:55–61; A2 직접 수정 없음, :36 계약 유지 | 실제 수집 fixture의 같은 routes/batch에서 관문 전 제안을 채취. 미선언/부분 ON 모두 관문 전 3, 기존 선택 입력 2, OFF reversal 제안 1 보존. 복사 접근자 반환 슬라이스 교체가 원본에 영향 없음. 시장 닫힘 13갈래의 관측 없음은 V2의 명시적 범위 제한; 13갈래 새 운반 동작 전체를 구현했다고 주장하지 않음. |
| P1-3 — 활성화 ON 우선 판단 입력 없음 | CLOSED | V2 §6:73–78; A2:17 | 같은 파도 activation으로 Desired==OFF && Effective==OFF를 호출 경계에서 검사. 실제 worker fixture에서 OFF=true/WOULD_EMIT=1, ON=false/관측 0. 혼합 ON/OFF에서도 OFF 레인 표본 1 확보. |
| P1-4 — Validate가 모든 읽기 경로라는 부정확한 주장 | CLOSED | V2 §0:18–21; A2 직접 수정 없음 | 외부 경계로 주장을 명시적으로 좁혔다. 재실행 `TestReviewReadValidationBoundary`: Context.Read 오류 nil, 별도 Validate는 INVALID 상태 거절. 결함을 고쳤다는 주장이 아니라, 1차가 요구한 정확한 경계 설명을 채택한 것. §9:101도 §0의 명시적 외부 경계 정정을 적용해서 읽어야 함. |
| P1-5 — 정규 직렬화가 생성기 작성 출처를 증명한다는 논증 | PARTIAL | V2 §0:24, §1:28–31 ↔ A2:7,15,17 / D:240,246,248 | V2는 신뢰 앵커를 배포 핀 작성 권한으로 바로잡음. 그러나 amendment :15의 현행 결정 63-v2가 여전히 “결정 61의 셋이 그대로 적용된다”고 참조함. 그 셋 중 (c), A2:7/D:240은 정규 등식 때문에 손으로 쓸 수 없고 생성기가 작성한다는 거짓 논증이다. A2:17/D:248의 생성기 약속도 운영 정책과 기술 증명을 구별하지 않음. V2:31의 “문서·결정 63과 일치”는 아직 성립하지 않음. |

새 P1-N1 — 운반 타입이 dispatch에 닿을 수 없다는 설계와 기존 수신자 구조 충돌.

V2 §4:55–59는 strategyShadowInputs를 strategyMarketArbitration에 넣고 strategyProposalMarketAuthority가 그대로 운반하면서, dispatchHandoffs·entries에는 그 타입이 닿을 수 없다고 한다. 그러나 현재:

- `internal/app/engine/strategy_proposal_authority.go:143`의 strategyProposalMarketAuthority는 `internal/app/engine/strategy_dispatch_handoff.go:37` dispatchHandoffs의 값 수신자다.
- `internal/app/engine/strategy_market_coordinator.go:130` entries도 strategyMarketArbitration 전체를 값 수신자로 받는다.
- 따라서 필드를 추가하는 것만으로 두 메서드 수신자 안에 shadow 운반 타입이 포함된다. 같은 engine 패키지의 비공개 필드는 그 메서드로부터 숨겨지지 않는다. 복사 접근자는 원본 변조를 줄일 뿐 보유/읽기를 차단하지 않는다.

`TestRecheckV2CarrierReachesDispatchReceiver`는 V2의 최소 운반 구조를 go/types로 검사한다. entries만 반환하는 dispatchHandoffs도 `receiver carries strategyShadowInputs=true`. 같은 수신자의 공격 함수가 `a.shadow.Values()[0]`을 읽는 코드도 타입 검사를 통과한다. 실제 함수의 수신자와 분기는 `ast-strategyProposalMarketAuthority.dispatchHandoffs.json`, `ast-strategyMarketArbitration.entries.json`에 기록했다.

이것은 실주문 재현이 아니라 **설계와 안전 census가 동시에 만족되지 않는 P1**이다. receiver를 census에서 빼면 보유 증명을 포기하게 되고, 넣으면 정상 설계 배선부터 실패한다. “entries만 읽는다”는 별도 핀은 현재 읽기를 제한하지만 타입 비도달 주장을 성립시키지는 않는다.

종결 방법: shadow 운반 결과와 dispatch 권한을 형제 값으로 분리하고 dispatch 쪽 메서드에 shadow 없는 권한만 전달한다. 또는 보장 범위를 명시적으로 “보유 금지”에서 “읽기/영향 금지”로 변경하고 수신자 예외·필드 읽기/별칭 검사·정확한 동결 소스 범위를 정의한다. 첫 방법이 현재 V2의 비도달 주장과 맞는다. 원본 설계는 이번 재검에서 수정하지 않았다.

P1-5 종결 방법: 결정 63-v2에 정규 등식은 바이트 동일성만 보장하며 신뢰는 배포 핀 관리 권한에서 오고 생성기 사용은 운영 정책이라는 문장을 넣는다. 결정 61(c)는 역사 기록으로 남기더라도 현재 SHADOW 근거로 다시 채택하지 않는다고 명시한다.

추가 P0 없음. 1차 P2 두 건의 v2 악화 없음; 새 P2는 작성하지 않았다. fault 격리·restart 시험은 V2 계획의 내용이지 이번 표적 실행으로 입증된 결과가 아니다.

실험과 한계:

- `tests-final.log`: 표적 Go 시험 9건 PASS. 재검에서 드러난 설계 충돌을 예상하고 확인하는 시험도 PASS로 표시된다. 테스트 PASS가 설계 PASS를 뜻하지 않는다.
- `cross-package-negative.log`: 별도 패키지에서 market/generation 비공개 필드를 설정하려던 공격이 기대대로 build failed. 실패 원인 두 개를 확인했다.
- 별도 패키지 `review-recheck/strategyshadow`는 불투명 타입·영값의 리뷰 전용 스케치다. Fixture는 시험 생성자이며 실제 manifest 적재기/신뢰 체인을 구현하지 않았다.
- 입력 채취는 기존 loader test seam에서 동일한 routes/batch를 읽었다. 생산 coordinateMarketProposals를 편집하지 않았으며, 새 운반 필드의 완전한 통합 구현도 하지 않았다.
- 차등 척도는 실제 worker/조정자까지다. V2 §8이 요구하는 runProductionStrategyMarketCycle 전체, gateway, handoff 차등 시험은 실행하지 않았다.
- 타입 census는 재검 공격 모양에 한정한 스케치다. 저장소 전체 census·전이 폐포 검사를 완성했다는 주장이 아니다. 소스 digest 변이 검사는 모양에 무관한 동결의 성질만 측정하며 실제 V2 경계 패키지의 새 동결 기준을 승인하지 않는다.
- Pre-Edit FLM: 기존 생산 함수 편집 없음. 시험/스케치만 추가하므로 구현 Pre-Edit는 not-applicable. 근거 함수 13개의 tools/logic-map AST를 `go test`로 생성했다.
- CodeGraph/외부 memory/스킬·make gate/PM/archive는 이번 제한된 표적 재검에서 수행하지 않았다. go test만 허용, 네트워크 금지, 전역 에이전트 디렉터리 금지 조건을 지켰다. 개발 완료·archive·배포 완료 보고 아님.

실행 명령(작업 디렉터리 안에서):

```bash
set -euo pipefail
GOCACHE="$PWD/.gocache" GOTMPDIR="$PWD/.gotmp" GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local TMPDIR="$PWD/.gotmp" go test -tags tossos_testseams ./review-recheck/... ./internal/strategyworker ./internal/app/engine ./tools/logic-map -run 'TestRecheckV2|TestReviewShadowInputTap|TestReviewReadValidationBoundary' -count=1 -v
```

음성 대조는 `review-recheck/negative/attack.go.txt`를 같은 디렉터리 `attack.go`로 놓고 동일 환경변수의 `go test ./review-recheck/negative`를 실행했다. 기대된 컴파일 실패 뒤 `.txt`로 복원했다. 네트워크·브로커·LIVE·운영 토글 호출 없음. 활성화 시험 값은 기존 tossos_testseams의 메모리 fixture만 사용했다.

원본 변경 확인:

- git archive 2817064c의 `internal`, `tools`, `go.mod`, `go.sum` 원본 1,916개를 사본과 바이트 비교: 변경 0 (`source-comparison.txt`).
- 추가: worker 시험 1, engine 시험 2, logic-map 시험 1, review-recheck 스케치·보고서·로그·AST 및 사본 내부 Go cache/temp. 기존 생산 함수 수정 0.
- 실저장소 시작/종료 HEAD·status 결과는 별도 `repository-check.txt` 참조.
