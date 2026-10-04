모델: GPT-6 기반 Codex · 세부 버전 미공개 · 「~/.codex 미접근」

REJECT — 현 설계의 구조적 MUST NOT 증명이 성립하지 않는다. 계약의 digest 핀 전환 자체를 거부하는 판정은 아니다.

검토 대상은 c7219640의 계약·코드와 지정 설계 브리프다. 실제 SHADOW 구현은 없다. 아래 공격은 사본의 추가 `_test.go` 안에서 의도된 타입/메서드 형태를 세운 변이 실험이며, 배포된 취약점이나 실주문 발생을 뜻하지 않는다. 호출 허용 목록과 타입 census는 선례의 알고리즘을 재사용했다. 아직 없는 SHADOW 핀 전체가 통과했다고 주장하지 않는다. 특히 무조건 Submit하는 변이는 설계의 행동 스파이를 구현하면 잡힌다. 문제는 그것을 구조적으로 불가능하다고 주장할 수 없다는 점이다.

좌표 약칭:
- B = `openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/design-brief-under-review.md`
- D = `openspec/changes/a112-run-four-strategy-families-independently/design.md`
- S = `openspec/changes/a112-run-four-strategy-families-independently/specs/four-family-strategy-runtime/spec.md`
- E = `review-experiments/`

| 등급 | 지적 | 근거·재현 | 결과 | 제안 |
|---|---|---|---|---|
| P0-1 | 봉투 없는 반환형·불투명 설정형은 SHADOW의 dispatch 접근을 구조적으로 차단하지 않는다. 제안된 이름 기반 호출 검사도 우회된다. | B:51-59; `internal/strategyworker/worker.go:165`의 Input은 Envelope의 세 구성요소를 모두 소유. `internal/strategycoordinator/coordinator.go:203`의 Submit은 활성화를 검사하지 않는다. `internal/strategyworker/review_shadow_sketch_test.go`, `TestReviewShadowStructure` | 실제 FamilyWorker에 스케치 메서드를 추가. 결과는 문자열 하나, desired/effective=OFF/OFF. `worker.owns`라는 로컬 함수 필드에 Submit을 넣어 호출하면 문자열 허용 목록 통과, 실제 coordinator depth=1. 브로커·handoff·lease는 실행하지 않았다. | SHADOW 입력을 dispatch 권한이 제거된 평가용 사실로 좁히고 전용 패키지/함수의 접근 경계를 정의하라. 타입 해석으로 호출 대상을 확인하고 재바인딩·클로저·전역 sink를 검사하라. 불가피하게 강한 입력을 유지한다면, 검토한 경계와 연결부 전체의 소스 동결·재리뷰를 포함하라. 반환형 census만으로 안전성을 종결하지 말라. |
| P1-1 | 인용한 타입 census 선례의 완전성은 이미 철회됐다. 활성화 인코더 가드는 활성화 값의 직접 주조를 막지 못한다. | B:57-59,102; `internal/strategyhandoff/mint_census_test.go:17`, `:83`; `source_freeze_test.go:20-31`; `TestReviewCensusAliasAndErasure`, `TestReviewShadowSamePackageMint` | 복사한 carries 알고리즘: 직접 Envelope=true, 중첩 alias=false, any=false, generic instance alias=false. router 같은 패키지에서 쌍둥이 구조체 변환→any로 실제 FamilyActivation을 만들면 Verified=true, 네 레인 ON, encoder reference=0. 파일 쓰기·운영 활성화는 없었다. | 선례의 source freeze까지 채택하거나 동등한 좁은 경계를 마련하라. activation/shadow 주조 경계 둘 다 검사. `types.Unalias`, 변환·generic·type erasure·out-param·함수 값 변이를 포함하고, census 한계를 명시하라. 인코더 가드는 저작 함수 참조 검사이지 activation 주조 금지 증명이 아니다. |
| P1-2 | SHADOW에 공급할 입력의 수집 지점이 빠졌다. 기존 레인 입력은 순수 제안 전체가 아니다. | B:10-13,90-91; `strategy_proposal_authority.go:355,390-407`; `strategy_market_coordinator.go:91-101,130-142`; `strategy_lane_runtime.go:266`; `strategy_entry_supervisor.go:543-546`; `TestReviewShadowInputTap` | 활성화 미선언에서는 두 종목 각각 승자만 남음: `000660:continuation`, `005930:reversal`. 같은 005930의 continuation은 사라진다. 선언됐으나 닫힌 활성화에서는 FAMILY_GATE_CLOSED, 입력=[]이다. | 실제 관문/중재 전에 batch의 모든 해당 레인 제안을 읽기 전용으로 분기하는 지점을 설계에 고정하라. 실패 조기 반환에서도 shadow 관측을 기록/폐기하는 규칙을 정하라. 중재 반사실을 구현하라는 요구가 아니다. |
| P1-3 | “ON이 우선, SHADOW는 OFF 레인만”을 제안된 Shadow 시그니처로 판단할 수 없다. | B:51-53,118; `worker.go:94-113`; `production_family_activation.go:611-616`; `TestReviewShadowStructure` | worker에는 활성화 상태가 없으며 Shadow 인자에도 FamilyActivation이 없다. 동일한 worker/shadow/input은 별도 활성화가 OFF이든 ON이든 WOULD_EMIT. desired=ON/effective=OFF도 허용되는 상태여서 effective만 검사해서도 안 된다. | 같은 파도의 activation snapshot에서 desired/effective 둘 다 OFF인 자격을 만든 뒤 Shadow에 전달하거나, 호출 경계에서 강제·동결하라. ON/OFF, OFF/OFF, ON/ON 및 전환 경합을 시험하라. 투영 검증 실패로 전체 스냅숏을 버리는 것은 우선순위 구현을 대신하지 않는다. |
| P1-4 | Validate가 “모든 읽기 경로의 관문”이라는 코드 사실은 과장이다. | B:114-118; `strategyprojection/store.go:24-35`; `internal/app/engine/strategy_runtime_projection.go:49-60,73`; `TestReviewReadValidationBoundary` | Store는 저장 시 검증한다. Context.Read는 저장소를 읽은 뒤 동적 lane projection을 덧씌우고 재검증 없이 반환한다. 테스트에서 잘못된 desired 관측을 넣으면 Context.Read는 nil error, 이후 명시적 Validate는 거절. | “외부 RPC/HTTP/콘솔 경계에서 검증”으로 표현을 좁히거나 동적 조립 후 공통 검증을 추가. 내부 Reader 소비자와 최종 overlay까지 추적하라. 현재 외부 전송이 무검증이라는 지적은 아니다. |
| P1-5 | “정규 바이트이므로 생성기만 작성 가능”은 작성자 증명이 아니다. 결정 63이 결정 61의 (c)까지 승계한다. | D:240,246; B:27-32,47; `production_family_activation.go:509-527` | 검사는 JSON 해석·정규 직렬화 등식이다. 동일한 정규 바이트를 다른 프로그램/사람이 만들었는지 구분하지 않는다. 인코더 호출 가드도 일반 JSON 직렬화/파일 작성의 출처를 증명하지 않는다. | S2는 유지 가능하되 신뢰를 배포 핀 관리 권한에 둔다고 정확히 적어라. 생성기 사용은 운영 절차·저장소 정책이며 암호학적 출처 보장이 아니다. `server-owned`의 UID, 작성 주체, 핀 변경 주체를 명시하라. |
| P2-1 | 재시작 무복원의 정상 경로 계획은 타당하나, SHADOW 실패가 기존 durable lane lifecycle로 들어가지 않는 경계가 미정이다. | B:68-71,90-91; `strategy_lane_runtime.go:208,240,252,301`; `strategy_lane_latch.go:149,232` | 재사용 후보 evaluate는 recover→RunBounded→persist를 수행한다. 기존 ON 관측 재시작 시험은 통과했지만 SHADOW timeout/panic/cancel의 무기록은 시험하지 않는다. 아직 코드가 없으므로 원장 쓰기가 발생했다고 단정하지 않는다. | shadow 결과·실패는 durable latch/recovery와 분리. 기존 원장 행을 미리 둔 상태에서 전후 delta=0, manifest 제거/핀 제거/만료/폐기, 실패·취소·지연 및 재시작을 시험하라. |
| P2-2 | 브리프 첫 계약 요약은 아직 signed이며, R2의 시나리오 조건도 amendment와 다르다. | B:3-5,70 vs S:89-93 | amendment는 “핀 없이 재시작”; 브리프는 “매니페스트 없이 재시작”. 두 경우 모두 필요하나 동일 시험은 아니다. | 첫 요약 갱신. (파일 유지·핀 없음), (핀 유지·파일 없음), (둘 다 없음)을 별도 행으로 고정하라. 결정 61의 옛 문장은 결정 63이 명시적으로 뒤집으므로 그 역사 문장 자체는 충돌로 세지 않았다. |

질문별 결론:

1. **구조적 MUST NOT: 불충족.** P0-1, P1-1. 실측은 실제 coordinator Submit까지이며 실주문이나 gateway 도달의 증명은 아니다. 불투명 FamilyShadow는 외부 필드 주조를 제한하지만, 같은 router 패키지의 FamilyActivation 필드와 worker가 이미 받은 봉인 제안까지 격리하지 않는다.
2. **문면: 레인 한정 반사실은 허용 가능한 해석.** S:89는 조정자 반사실 선택을 의무화하지 않는다. 기존 제안 생산이 shadow 핀 없이 돈다는 사실도 기존 생산 경로와 SHADOW의 허용 조건을 구분하면 모순이 아니다. 그러나 P1-2의 입력 유실을 해결해야 “레인 평가”가 실제 입력에 대한 반사실이 된다. 새 설치/재시작 초기값은 OFF/OFF/UNOBSERVED, 그 뒤 유효 매니페스트의 새 평가로만 SHADOW라는 시점을 문서화할 것.
3. **amendment: S2 선택 자체에 차단 사유 없음.** 핀 env 이름은 B:38,101과 S:89/D:246이 일치한다. 조사한 정본 change specs에서 남은 signed-shadow 의무는 발견하지 못했다. D:242는 역사 기록이며 D:246이 명시적으로 대체한다. 다만 (c)의 작성자 증명 주장은 삭제/정정 필요(P1-5). 파일과 핀을 함께 변경할 수 있는 배포 주체를 불신하는 모델은 digest 핀 모델의 보호 범위 밖이다.
4. **재시작: 현재 저장 구조에는 shadow 복원원이 없다.** Store와 observed는 메모리이고 새 Context/Store의 초기 상태 시험은 통과했다. 매 파도 재독은 허용 조건이지만, durable lane lifecycle과 실패 경로 분리까지 자동으로 증명하지는 않는다(P2-1). 실제 SHADOW 로트의 재시작 시험은 아직 없다.
5. **코드 사실 다섯 개:**
   - 활성화와 무관한 제안 적재: 조건부 참. collectMarket:316/329/332/337 선행 권한·FX·설정·키 조건 뒤 :355에서 적재하며 가족 활성화의 ON 여부로 그 적재를 막지 않는다. “아무 선결 조건 없이 평가”는 아니다. 레인 입력까지 전체 제안이 보존되는 것은 거짓(P1-2).
   - 골든 runtime 어휘 미열거: 참. `analysis/goldens/four-family-runtime-v1.json:17-24`의 여덟 기본값이고 별도 runtime enum이 없다. 기본값 무변이면 골든 변경은 필요하지 않다.
   - OpenAPI 열거: 참. `docs/api/openapi-v1.json:2641`: `["UNOBSERVED"]`.
   - 모든 읽기 Validate: 외부 경계는 맞으나 전체 내부 읽기까지는 거짓(P1-4).
   - RPC unknown field 허용: 실제 Client.Read를 네트워크 없는 메모리 RoundTripper로 실행. `shadowOutcome` 추가는 통과, runtime=SHADOW는 검증에서 거절. 따라서 구판 호환성 한계를 정확히 기록해야 한다.
6. **공유 경계: 방향 타당.** 파일 읽기/UID/0400/digest/정규성/시간 비교는 공유할 수 있다. activation 승격 생성과 shadow 자격 생성, 서술자와 ProtectionReady 하한은 공유하지 말아야 한다. B:103-104는 사본일 경우 양쪽 핀을 요구하므로 “그 계획이 없다”는 지적은 하지 않는다. 생산 활성화 검증에는 다섯 일반 결속 외에 `ProtectionReadyMinGeneration != 0`도 있다(:553); 이를 공통 validator에 무조건 남기면 shadow가 닫히고, 활성화에서 빼면 보호 계약이 약해진다. 양쪽 회귀표를 분리할 것.
7. **노출·배포: 8.6 불재개 유지.** D:293과 B:83-85,108은 A100/의존 게이트/사람 승인 이후에만 운영 배포한다. shadow가 ProtectionReady를 결속하지 않는 것은 dispatch 권한 0이 입증된다는 조건에서만 수용 가능하다. production shadow 핀 0은 사용자 제공 전제이며 운영 환경은 조사하지 않았다. 현재 미구현이므로 “구현 뒤 동작 변화 0”을 실증 완료로 부를 수 없다. 무핀 no-op, mixed ON/OFF, 구판 reader를 회귀표에 포함할 것. 같은 이미지 태그를 쓰는 compose.yaml:10,94는 확인했지만 모든 소비자가 실제로 같은 판본이라는 증명은 아니다.

시도했으나 방어가 유지된 공격:

- 영값 shadow를 스케치에 투입: NOT_SHADOWED. 실제 기존 Run에 영값 activation과 유효 제안을 투입: DORMANT. 반환형 분리의 유효한 제한은 인정한다.
- 직접 Envelope 필드: 인용 census가 carries=true로 검출. 실패는 모든 타입이 아니라 alias/erasure 등 경계다.
- 인코더의 일반 import/별칭/dot import/blank+alias 우회: 기존 TestTheEncoderGuardCountsEveryWayToNameThePackage의 네 경우 모두 검출·PASS.
- RPC에 unknown shadowOutcome: 읽기는 유지됨. 알 수 없는 runtime=SHADOW: 기존 의미 검증이 거절. 네트워크 접속 없음.
- 기존 관측 ON 후 새 Context/Store, 같은 로컬 시험 원장: 여덟 OFF/OFF/UNOBSERVED, dispatch lease/latch/recovery 0건 시험 PASS. SHADOW 구현 검증과 혼동하지 않는다.

범위 밖(등급 없음): coordinator의 반사실 승자 선택은 요구하지 않았다. ROADMAP 선행 행과 활성화 운영자 실패사유 표면도 수정하지 않았다.

재현·한계:

- `bash review-experiments/reproduce.sh` (스크립트는 set -euo pipefail로 시작하며 모든 Go 실행에 지정한 GOCACHE/GOTMPDIR/GOFLAGS/GOPROXY/GOSUMDB/GOTOOLCHAIN을 붙임. TMPDIR도 사본 내부.)
- 최종 결과: `review-experiments/final-tests.log`. 실패를 기대하는 공격은 취약 형태가 실제로 도달하는지 단언하므로 테스트 PASS가 안전 판정은 아니다.
- `tools/logic-map`의 analyze를 go test 래퍼로 실행하여 대상 9개 함수의 AST 근거를 `E/ast-*.json`에 저장했다. 전역 memory·CodeGraph/GBrain·gstack·make gate는 사용하지 않았다. 네트워크/전역 설정/실저장소 쓰기 제한 아래 고정 소스와 AST, 오프라인 go test로 독립 설계 리뷰를 수행했다. 구현 완료·배포 승인·SDD gate 통과를 주장하지 않는다.
- worker 실험의 FamilyShadow는 비공개 bool 하나를 가진 로컬 surrogate다. 실제 파일 검증 적재기를 구현한 시험이 아니며 신뢰 앵커를 우회했다고 주장하지 않는다. 유효 shadow 자격을 이미 받은 뒤에도 capability가 차단되는지 시험했다.
- router 첫 실험 빌드는 descriptor의 lane ID가 map key라는 점을 잘못 적어 실패했다. 스케치만 고쳐 최종 재실행 PASS. 원본 생산 파일 수정 없음.
- 지정 브리프 SHA256: `96baede4e07171df5ef898cb00018f44c20edf420e7144a5609b561cb402d762` 일치.
- 실저장소 시작/종료 HEAD·status 비교는 `repository-check.txt`; 사본의 기존 추적 소스와 고정 커밋 비교는 `source-comparison.txt` 참조. 추가물은 실험용 테스트·AST·로그·이 보고서·사본 내 Go 캐시뿐이다.
