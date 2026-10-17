Codex · GPT-6(세부 버전 비공개) · 「~/.codex 미접근」

**FAIL — 4라운드 P1 및 비차단 노트 셋 CLOSED(설계 수준), 신규 P0 0건 · P1 1건.**

입력: 지정 archive 사본, 기준 `4d22d72663dcf3a42be928b3ecf3db817ffc6ac5`. v3.2 SHA-256 `68512bc36b58f0c184115016b377c17d0c3065535d136e1a41aa783c48b09af1` 일치. recheck5-brief, v3.1→v3.2 diff, codex 4라운드 출력·보고서·모델을 확인했다. 아래 v3.2 좌표는 `openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/design-brief-v3.2-under-review.md` 기준이다.

| 항목 | 판정 | v3.2 좌표 | 근거 |
|---|---|---|---|
| record 전 복구 오류 후 이전 SHADOW 잔존 | CLOSED | §5:121–131 | 성공 여부를 추적하는 defer로 epoch 증가와 관측 삭제. 모델에서 wave=7 그대로, epoch=1, 즉시 UNOBSERVED. |
| record 전 panic 후 이전 SHADOW 잔존 | CLOSED | §5:124–131 | defer는 recover하지 않음. 동일 panic 값이 외부까지 전파되면서 관측 삭제. |
| 실패 뒤 늦은 in-flight 재게시 | CLOSED | §5:127–131 | epoch와 wave를 같은 mu 아래 비교/쓰기. 삭제 전 게시와 삭제 후 게시 양쪽 순서, 동시 실행에서도 재출현 없음. |
| forMarket 예외·몸체 핀 | CLOSED | §5:136–140 | 인자 index 5 부분 트리만 허용, 유일한 호출을 forMarket으로 명시, 시장 선택·필드 반환 외 호출 금지. |
| ④ evaluate→record 전달 | CLOSED | §5:141–143 | 인자 전달과 record의 유일 보관 대입을 명시적으로 구분. |
| §4 다른 shadow 식별자 0 범위 | CLOSED | §4:91–94 | 조정 경로의 식만 금지, helper와 반환 운반 허용. |
| 나이 상한·만료 경계 | CLOSED, 경계 알고리즘 | §5.1:159–171 | 74s 및 expiresAt 등호는 거부, 각각 -1ns는 허용. 상한 전제와 연속성 주장은 아래 P1과 구별. |
| defer 오류·panic·dispatch 경로 | 추가 P0/P1 없음 | §5:124–133 | 모델의 baseline/wrapped 성공·pre-error·pre-panic·post-error 궤적과 오류/panic identity 동일. 생산 통합 증거는 아님. |
| CAS 잠금 순서 | 추가 P0/P1 없음 | §5:119–128 | snapshot·CAS·invalidate는 runtime mu만 필요. 레인 잠금 보유 중 진입할 이유 없음. 아래 조건/한계 참고. |
| 정상 주기에서 SHADOW 무중단 보장 | **신규 P1 OPEN** | §5.1:157–168, §5:114–120 | 최신 wave 등식과 record 이후 비동기 게시 사이 공백은 MaxAge로 해소되지 않음. 실행 반례 확인. |

## 신규 P1 — 나이 상한 대조 시험이 최신 wave 규칙과 충돌

- §5.1:168은 “주기 한도 30s를 다 쓴 건강한 주기 연속에서 SHADOW가 끊기지 않음”을 요구한다. 하지만 :157의 최신 evaluate wave 등식, §5:114–120의 record 시 wave 증가 및 cycle 성공 후 비동기 SHADOW 시작을 지키면 정상 주기에도 UNOBSERVED 구간이 생긴다.
- 반례: t=0에 wave 7 SHADOW 게시, manifest 만료는 1시간 뒤. 다음 정상 주기가 5s 뒤 시작해 30s−1ns 뒤 record wave 8. 그 순간부터 이전 관측은 `7 != 8`이므로 UNOBSERVED. cycle/dispatch 성공 후 shadow가 2s−1ns 안에 게시하면 t=37s−2ns에 다시 SHADOW. 실패·만료·철회 없음, 관측 나이는 계속 74s 미만. 마감과 정확히 동률인 성공에 기대지 않도록 -1ns 사용.
- `TestV32HealthyCycleContinuityCounterexample`에서 위 공백을 재현했다. 시험 PASS는 반례 재현 성공이다.
- 원인은 새 상수의 숫자가 아니라 “정상 입력 거부 없음” 검증을 전체 투영의 연속성으로 선언한 점이다. 기존 최신-wave 정책을 신규 P1로 다시 집계한 것이 아니다. v3.2가 새로 추가한 보장이 그 정책과 충돌한다.
- 최소 수정: :168을 “같은 wave의 관측이 가정한 정상 관측 간격 안에서 **나이 조건만으로** 거부되지 않음”으로 한정. record→새 shadow 게시 사이 UNOBSERVED는 신선도 정책의 의도된 결과임을 명시하고 해당 대조 시험을 분리한다. 전체 연속성이 정말 필요하면 별도 계약 결정이 필요하다. 이전 wave 허용을 임의로 추가하면 철회·실패 소거를 약화하므로 권하지 않는다.

## defer·잠금·상한 검토

- AST 12개를 이번 사본에서 새로 생성했다. `strategyLaneRuntime.evaluate`:208–209 복구 오류 → :244–247 panic 재던짐 → :249 record → :253–258 기록 뒤 오류 순서. `record`:326–335 잠금 안 wave 증가. 따라서 (i)~(iii)은 wave 불변인 실패를 사용해야 하며 모델이 이를 확인한다.
- `invokeStrategyCycle`:1079–1091의 외부 recover 앞에서 invalidation defer가 완료되어야 한다. 모델은 nil로 초기화된 named error만 보는 구현을 쓰지 않고 `returnedNil=false`를 두고 주기 정상 반환 뒤 갱신한다. 그렇지 않으면 panic 때 err=nil인 채 남을 수 있다.
- 비차단 문구 노트: :124의 실패 조건과 :125–126의 “defer 본문 = 호출 하나”를 문자 그대로 동시에 핀할 수 없다. 실제 핀은 `if !returnedNil { invalidateShadow(market) }`의 조건부 호출 모양과 성공 표식 갱신 자리를 허용해야 한다. 무조건 호출하면 성공 후 관측도 지워진다. 실패 때만 폐기한다는 동작 계약은 명확하므로 별도 P1로 중복 집계하지 않았다. 모델은 이 해석을 명시적으로 택했다.
- 현재 projection:35–40은 runtime 읽기 잠금 아래 `strategyLaneProjection`을 호출하고, 그 함수 :49가 `Lane.Status`를 읽는다. 레인 Status는 `lane_view.go:26–31`의 레인 잠금이다. 현재 방향은 runtime→lane. 설계의 새 snapshot/CAS/invalidate는 runtime 단독 잠금이므로 반대 방향의 새 간선이 없다. CAS 구간에는 verdict·lane 접근·I/O를 넣지 않아야 한다. 실제 미구현 CAS의 교착/경합 성능 검증은 하지 않았다.
- 5s는 생산 worker의 PollInterval 설정(:394, :439, :511), 30s는 생산 supervisor 설정(:401)과 최대 상수(:30), 2s는 새 설계 가정이다. 식 2×(5+30+2)=74s 및 등호 경계는 맞는다. 37s는 정상 스케줄링·주기 및 shadow 한도 준수 조건의 계산값이며 무조건적인 벽시계 최대치 증명은 아니다. 단일 비행 건너뜀과 최신-wave 부재는 별도 신선도 거부 사유다.
- `invokeBoundedStrategyCycle`:1057–1077의 watchdog는 내부 cycle goroutine 종료와 다르다. goroutine이 아직 반환하지 않으면 그 클로저 defer도 아직 실행되지 않는다. “즉시” 증명 범위는 실제 내부 오류 반환/panic unwind이며, watchdog 반환까지 확대하지 않았다. 파도 정지의 backstop은 별도 나이 상한이다.

## 시험·증거·절차

- 새 `.go` 파일은 사본 루트 `_codex/recheck5/`의 `review_test.go`, `extract_go_ast.go` 둘뿐. 후자는 기존 `tools/logic-map/extract_go_ast.go` 복사본. 생산 소스 변경 없음.
- 최종 시험: 최상위 8개 PASS, race detector 보고 0. 오류/panic 하위 2갈래, CAS 두 순서·다른 시장·새 wave, 100회 동시 게시/폐기, 4종 baseline/wrapper 궤적, 나이/만료 경계, 연속성 반례 포함. 상세 출력 `tests-race.log`.
- 실행 명령(모든 셸 첫 줄에 `set -euo pipefail`):

```bash
set -euo pipefail
GOCACHE="$PWD/.gocache" GOTMPDIR="$PWD/.gotmp" GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -race -count=1 -v ./_codex/recheck5
```

- 첫 시험은 `-race` 없는 동일 패키지 실행으로 AST와 기본 모델을 확인했다. 최종 모델에 차등 궤적을 추가하고 시간 경계를 -1ns로 좁힌 뒤 race 시험을 다시 수행했다.
- `function-branch-map.md`, `ast-*.json`은 현재 코드 근거. `source-comparison.txt`는 관련 생산 소스 5개 및 go.mod/go.sum이 기준 커밋과 바이트 일치함을 기록한다.
- **한계:** Go 최소 모델 검증이다. 실제 SHADOW 생산 구현, 실제 dispatch/latch 원장 통합시험, 전체 upstream 회귀, 생산 핀/변이 시험 통과를 주장하지 않는다. Production 편집·Pre-Edit 변경 게이트는 not-applicable. CodeGraph/GBrain/외부 skill/memory 탐색·make gate/sdd는 금지 경로와 go test 전용 제한 때문에 미실행. 구현 완료/배포 승인 판정 아님.
- 모든 셸 호출은 첫 줄 `set -euo pipefail`. 금지 경로 읽기·검색 없음, 네트워크·브로커·LIVE·토글·엔진 lock 호출 없음. 사본에서 git 실행 없음. 실저장소 git은 절대경로 `git -C` 읽기만 사용.
- 최초 상태/HEAD는 최초 두 읽기 도구 출력에서 전사해 start 파일로 보존했다. 종료 때 각각 다시 측정해 `cmp` exit 0 확인. HEAD `4d22d72663dcf3a42be928b3ecf3db817ffc6ac5` 동일. 실저장소 원래 dirty 상태도 동일:

```text
 M docs/ROADMAP.md
 M openspec/changes/a112-run-four-strategy-families-independently/tasks.md
?? .reticle-setup-crash.log
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/shadow-2026-10/
?? w4.log
```
