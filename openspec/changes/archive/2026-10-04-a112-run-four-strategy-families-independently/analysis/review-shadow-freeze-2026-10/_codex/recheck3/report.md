모델: GPT-6 기반 Codex · 세부 버전 비공개 · 「~/.codex 미접근」

**FAIL — 재검 대상 P1-5·N1 모두 CLOSED; v3 신규 P0 0건·P1 1건.**

CLOSED는 설계 수준 종결이다. 생산 구현·배포 완료 판정이 아니다.

- 기준 사본: 4d22d72663dcf3a42be928b3ecf3db817ffc6ac5.
- V3: `../openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/design-brief-v3-under-review.md`.
- V3 SHA-256: f7883a47fb1f07dffb32e561e9e9ab00c8597f8cc6f108b97f3b5d65053230fc — 지정값 일치.
- A3: 같은 디렉터리 `amendment-v3-4d22d726.patch`.

| 대상 | 판정 | 좌표 | 근거 |
|---|---|---|---|
| P1-5 정규 직렬화가 작성 출처를 증명한다는 논증 | CLOSED | V3:30–33; A3:10; change design.md:245 | 결정 63은 61(a)·(b)만 적용하고 (c)를 SHADOW 근거로 재채택하지 않는다고 명시한다. 바이트 동일성·배포 핀 권한·생성기 운영 정책을 구분한다. 61(c)의 역사 기록 존치는 재채택이 아니다. |
| N1 shadow 운반 값이 dispatch authority 수신자에 포함됨 | CLOSED | V3:40–47,63–72 | 운반 값을 authority/arbitration 밖 별도 반환값으로 분리한다. v3 모양 스케치에서 authority가 ShadowInput/FamilyShadow/batch를 재귀 보유하지 않으며 dispatchHandoffs·authorityForOwnerScope의 본문 사용도 0. 새 접근자와 메서드 값 경유 사용은 Types/Uses 검사에 잡힌다. |

새 P1 — N2: 폐포 전체 unsafe/reflect 금지는 필수 의존성과 충돌

- 근거: V3:37–39는 strategyshadow가 strategyrouter를 import하도록 하면서 그 의존 폐포에서 unsafe·reflect를 금지하고 ListDeps 네 모드로 핀하도록 한다.
- 현재 `internal/strategyrouter/production_family_activation.go:6`은 encoding/json을 import한다. 로컬 Go 표준 라이브러리 `encoding/json/decode.go`는 reflect와 unsafe를 모두 import한다.
- `internal/testenv/closure.go:81–117`의 ListDeps는 표준 라이브러리를 제거하지 않고 전체 import graph를 만든다. 문서에는 표준 라이브러리 절단점이나 예외가 없다.
- 재현: `TestV3WholeClosureBanConflictsWithRequiredRouter`가 소스 import를 AST로 확인한다. 필수 경로 두 개가 모두 존재한다: `strategyshadow → strategyrouter → encoding/json → reflect`, 같은 경로의 `unsafe`.
- 결과: 명세 그대로 전체 폐포 금지를 구현하면 정상 설계부터 실패한다. 우회 공격이나 실주문 재현이 아니라 새 보안 핀의 충족 불가능성이다. 따라서 P0가 아닌 P1.
- 수정 제안: 프로젝트 소유 패키지에서 unsafe/reflect를 직접 사용하는 것을 금지하고, 검토·동결한 표준 라이브러리 내부 사용은 명시적 절단점으로 둔다. 비공개 필드 우회 접근자와 함수 값 세탁 변이는 계속 거절해야 한다. 네 모드 동일 정책을 명문화한다. strategyrouter를 통째로 무검사 예외 처리하면 안 된다.

요청한 v3 신규 표면 재검

| 표면 | 결과·한계 |
|---|---|
| 비동기 shadow 단계 | V3:83–98의 dispatch 이후 실행, 시장별 단일 비행, 마감, 파도 신선도는 기존 주문 호출 위치와 양립한다. 현재 cycle·두 클로저·bounded wrapper의 Go AST를 생성했다. 구체적인 신규 P0/P1은 입증하지 못했다. 아직 구현되지 않은 mutex·스냅숏 복사·늦은 결과 폐기에 대한 race 통과를 주장하지 않는다. |
| 반환 모양 변경 | V3:65–76은 collectMarket 반환 변경과 Pre-Edit FLM을 명시하고, handoff 원천은 proposals로 유지한다. 별도 반환값 스케치는 타입 검사 통과. 실제 조립 전체의 컴파일·회귀 시험은 이번 설계 검토 범위에서 수행하지 않았다. |
| 중립 wrapper export | V3:51–59는 기존 함수 무편집·한 줄 위임·AST 핀·단일 서술자 표·sentinel 변환을 명시한다. 서술자 함수 AST를 생성했다. export 자체가 FamilyActivation 비공개 필드 작성 권한을 열지는 않는다. 새 주문 능력 경로는 입증되지 않았다. |

스케치 시험

1. `TestV3SeparatedCarrierAndAccessorMutation`: 정상 dispatch 2개 사용 0; 접근자 세탁·메서드 값·중첩 generic 별칭·any 변환 검출. collectMarket은 운반 함수로 사용 양성.
2. `TestV3OpaqueInputRejectsForeignMintAndConversion`: 별도 패키지 ShadowInput 비공개 필드 주조 및 일반 Input 매개변수 전달 모두 예상 컴파일 거절.
3. `TestV3WholeClosureBanConflictsWithRequiredRouter`: 표준 라이브러리를 통한 금지 의존 도달 2개 확인.
4. `TestRecheck3Evidence`: 현재 소스에서 함수 5개 AST 산출.

위 4개 시험 PASS. 3번 PASS는 설계 충돌 재현 성공을 뜻한다. 기존 v2 helper를 복사해 v3 검사로 확장했으며, 스케치는 실제 생산 패키지 전체 census가 아니다. 정적 any 추적 전체나 임의 데이터 흐름의 안전성 증명도 아니다. 양성 대조는 원형을 축소한 타입 검사 모델이다.

실행 명령(사본 루트):

```sh
GOCACHE="$PWD/.gocache" GOTMPDIR="$PWD/.gotmp" GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go test -v ./review-recheck3 ./tools/logic-map -run 'Test(V3|Recheck3)' -count=1
```

`tests.log`에 실제 출력 보존. ListDeps는 내부에서 go list를 실행하므로 이번에는 호출하지 않았고, 필요한 의존 경로를 AST import 증거로 확인했다. 네트워크·브로커·LIVE·토글 실행 없음.

변경·절차 범위

- 사본에만 `review-recheck3/`, `tools/logic-map/recheck3_evidence_test.go`, Go cache/temp 생성. 생산 Go 파일·계약·기존 2라운드 증거 수정 없음.
- 기존 고위험 함수 편집/Pre-Edit: not-applicable — 생산 함수 미편집. 본문 근거는 새 Go AST 산출물에 한정.
- CodeGraph/GBrain/memory/외부 skill 및 make sdd-sync/sdd-check/gate: 미실행. 지정 사본 입력 기반 설계 재검이며, 사용자 시험 제한은 go test뿐이다. 구현 완료 또는 gate 통과로 보고하지 않는다.
- 실저장소 시작·끝 status 및 HEAD 비교는 `repository-check.txt`, `start.status`, `end.status`, `start.head`, `end.head`에 기록한다. 기존 작업 트리는 처음부터 dirty였으며 'clean'이라고 주장하지 않는다.
