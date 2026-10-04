# a112 6.2 봉인 로트(686b94e4) 적대 리뷰 — 목소리 B(증거 · 하네스 · census)

**판정: BLOCK**. 사유는 P1 둘(strategyflow census 가 빌드 맥락을 잘못 보고 있음, 거짓 증거 진술 하나)이고 둘 다 생산 코드 밖의 문제임. 생산 코드에서는 P0 를 찾지 못함 — 1차 레그 봉인은 내가 찾은 변이를 모두 거절했거나, 살아남은 변이도 결국 거절로 끝났음.

`~/.codex` 는 읽지도 검색하지도 않았음. 실제 저장소에서는 `git -C … show|log|diff` 만 썼음. 사본은 `/tmp/claude-1000/a112-rev62-yp9I` 에 만들었고 끝에 지웠음(확인함). 변이 도우미는 스크래치패드의 `mut.py` 로, 사본에서만 돌고 원복한 뒤 sha 를 단언함.
참고: 리뷰 도중 실제 저장소 작업 트리에 **내가 만들지 않은** 미커밋 변경 9개가 있었음(`a112_first_leg_owner_scope_seal_test.go`, `strategy_account_first_leg_authority.go`, `strategy_first_leg_owner_scope.go`, `seal_census_test.go`, `verify_named_tests.py`, `a112_lot_mutate.py`, a091 파일 3개). HEAD 도 `16cb1a1a` 로 옮겨 가 있었음. 이 리뷰의 대상은 686b94e4 트리뿐임.

무변이 대조군은 GREEN 임(-json 기준 pass 사건 177, fail 0).

## 발견 표

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | **P1** | census 가 「기본 빌드」를 **시험 호스트의 `build.Default`** 로 판정함. 시험 환경은 CGO_ENABLED=1 이고 생산 이미지는 `CGO_ENABLED=0 GOOS=linux`(Dockerfile:9)임. 그래서 `//go:build !cgo` 파일은 「tagged」로 분류돼 ① 공개 표면 대조에서 빠지지만, 생산 이미지에는 컴파일되어 들어감. release.yml 의 windows/darwin 빌드도 같아서 `_windows.go` · `_darwin.go` 파일이 ① 에서 빠짐 | `seal_census_test.go:61-65`(MatchFile), `:88-90`(tagged 는 건너뜀). 사본에 `reseal_nocgo.go` 를 넣음: `//go:build !cgo` + 공개 함수 `ResealForAnyone` 가 `copy(r.proposalSeal[:], proposalResultSeal(r)[:])` 를 함. 결과는 `go test -json ./internal/strategyflow` 가 **pass 135 · fail 0**. `CGO_ENABLED=0 GOOS=linux go list -f '{{.GoFiles}}'` 에 `reseal_nocgo.go` 가 들어 있고 빌드도 성공함 | 「tagged」는 `tossos_testseams` 제약만으로 정할 것. 아니면 생산 맥락들(linux/cgo0 과 release GOOS 전부) 중 하나라도 맞으면 생산 파일로 취급할 것. 그 밖의 build 제약이 붙은 파일은 census 가 거절해야 함 |
| 2 | P2 | ② 봉인 쓰기 census 는 **대입문 모양**과 **호출 모양**만 셈. 다음 넷은 기본 빌드에서 재봉인이 되는데도 로트의 모든 시험이 초록임: 함수 값 별칭(`var mintSeal = sealProposalResult`), 색인 쓰기 `final.proposalSeal[i] = seal[i]`, `copy(final.proposalSeal[:], …)`, 포인터 `p := &final.proposalSeal; *p = …`. 머리말이 못 보는 모양으로 적은 것은 위치 기반 리터럴 하나뿐임. 또 review.md:5531 의 「기본 빌드에서 그것을 쓰는 문은 `Propose` 하나」는 census 로 증명되지 않음 | `seal_census_test.go:169-193`. 넷 모두 `FinalizeProposalQuantity` 에 심었음(c1~c4). 결과는 **SURVIVED**, pass 사건 177, fail 0. 양성 대조(c3): 봉인되지 않던 `FinalizeProposalQuantity` 출력이 `ValidProposal()=false → true` 로 바뀜. 엔진 봉인은 막아 냄: identity 가드가 조립 항목과 대조하고, 조건 identity 에 수량이 들어가므로 거절됨. 그래서 P2 임. strategyflow 에 `unsafe` import 를 막는 장치도 없음(`dependency_test.go` 는 internal 경로만 막음) | go/types 로 셀 것: `proposalSeal` 필드 객체를 가리키는 **모든 참조**(Selections)와 `sealProposalResult` 의 **모든 Uses**(호출만이 아님). `unsafe` import 를 금지할 것. 머리말의 못 보는 모양 목록을 갱신할 것 |
| 3 | **P1** | 거짓 증거 진술임. `TestTheFirstLegSealSelectsByScopeBeforeTheMarketCountGate` 가 「`entries[0]` 로 고르면 identity 거절이 된다」를 잡는다고 적힌 곳이 셋임: 시험 주석(`a112_first_leg_owner_scope_seal_test.go:160-161, :166`), review.md:5555, BTM B4 의 「S04 · S06 이 이 시험을 빨갛게 함」. 그런데 개수 관문(B4)이 identity 가드(B5)보다 **앞에** 있음. 그래서 항목 둘인 쌍에서는 무엇을 고르든 선택만 성공하면 개수 문구로 끝남 | 원장 S06 행에도 이 시험 이름은 없음. S06 을 재현하니 실패는 4건(HoldsTwice · unselected · other_market · SingleProposalAssumption)이고 **이 시험은 PASS** 임. 이 시험의 실제 판별 대상은 따로 잼: M12(첫 항목만 훑는 선택)에서 이 시험과 HoldsTwice 가 CAUGHT, M2(개수 관문 제거)에서 이 시험과 TwoOwnerScopes 가 CAUGHT | 세 곳의 진술을 측정값으로 고칠 것: 판별 대상은 「위치에 기대는 선택(첫 항목만 봄)」과 「개수 관문 존재」. 원장에 M12 · M7 을 추가할 것 |
| 4 | P2 | `verify_named_tests.py` 는 「**실제로 돌았는지**」(:2)를 증명한다고 적었지만 과장임. 증명하는 것: `-count=1` 로 빌드가 성공했고, 그 패키지의 test2json 스트림이 최상위 이름마다 pass 사건을 **보고했다**는 것. 증명하지 못하는 것 셋: (a) 같은 프로세스 안 코드가 framing 을 위조하는 경우, (b) 하위 시험의 skip(부모는 pass), (c) 조기 return | 양성 대조 E1(`init(){os.Exit(0)}`): 평범한 `go test` 는 `ok`, 하네스는 NOT-RUN ×3 · rc 1 — 로트의 주장이 재현됨. E2: init 이 `=== RUN`/`--- PASS:` 줄을 찍고 exit 하면 **passed=3/3 · rc 0**(실행된 시험 0). E3: `\x16` 표식을 붙여도 3/3. E4/E5: 하위 시험 2개에 `t.Skip()`, HoldsTwice 첫 줄에 `return` 을 넣어도 **2/2 PASS**. 하위 시험 사건을 통째로 버리는 곳은 `:36-37` | 문서 주장을 위 범위로 좁힐 것. 이름 붙은 부모의 하위 시험에 skip 이 있으면 실패로 볼 것. 시험 패키지의 `init`/`TestMain` 을 AST census 로 금지하는 것을 고려할 것. 같은 프로세스 안의 위조는 하네스로 닫을 수 없다고 명시할 것 |
| 5 | P2 | RED 로그가 커밋된 시험 파일에서 나온 것이 아님. 로그의 실패 줄은 133/138/155 이고 커밋본은 135/140/157 임. 로그가 적은 트리(「a234d8d7 + 시험 파일 하나」)로는 커밋된 시험 파일이 컴파일되지 않음(새 파일의 `strategyProposalSetDigest` 가 필요함). A-lite 시험의 RED 는 로그에 없음 | 커밋된 시험 파일과 편집 전 생산 파일 둘로 재현함(sha `1d710710…` 는 `source-sha256.txt` 와 일치). 결과: 같은 셋이 FAIL 이고 모두 **기능 부재**임(assert 문구, 조립 실패 아님). 여기에 `TestAnActivatedMarketWhoseListIsNotTheArbitratedSetIsClosed` 도 FAIL(:199, 위조 목록 1개가 Admitted)이었는데 로그에는 없음 | 커밋된 파일로 RED 를 다시 기록하고, A-lite RED 를 포함할 것 |
| 6 | P2 | 소유자 범위 키의 4축 중 3축과 정규화 거절 분기에 시험이 없음. ④ 「타 시장」 시험은 시장 축이 아니라 **종목 축**으로 거절되고 있음 | M4(세대 제거) · M5(계좌 제거) · M6(시장 제거, 양쪽 모두) · M3(항목 정규화 오류를 `continue` 로)가 **전부 SURVIVED**(pass 177). 안전 영향은 없음: identity 가드가 계좌 · 시장 · 종목 · 캠페인 · 계보를 모두 담은 조건 identity 로 대조하므로, 거친 키는 거절 문구가 바뀌거나 유일성 거절이 늘 뿐임 | 같은 종목인데 다른 시장 · 다른 계좌 · 다른 세대인 fixture 에서 선택 실패 문구를 기대하는 시험을 둘 것. 정규화가 안 되는 항목 fixture 도 둘 것. 아니면 머리말 주장을 줄일 것 |
| 7 | P2 | 표면 golden 재생성 환경변수 `STRATEGYFLOW_REGENERATE_SURFACE=1` 이 **같은 실행에서** 쓰고 곧바로 통과함 | 사본에서 `flow.go` 에 `SealForAnyone` 를 넣고 env 를 켠 채 돌리니 `ok` 가 나오고 golden 에 한 줄이 추가됨. Makefile · CI · scripts 에서는 이 변수를 참조하지 않음(grep 0) | 파일을 쓴 뒤 `t.Fatal("regenerated — rerun without env")` 로 멈출 것 |
| 8 | T | ③ review 결속은 모든 review.md(archive 포함)를 `strings.Contains` 로 볼 뿐임. 그래서 저자가 같은 커밋에서 스스로 만족시킬 수 있고, 증명하는 것은 「적혀 있다」이지 「리뷰됐다」가 아님. 동결 셋 밖의 해시 입력(`writeLineageString`/`writeLineageUint64`/`executionTermsIdentity`)은 보지 않음. `found` 맵은 파일 정렬상 마지막 정의가 이김 → 플랫폼별로 나뉜 동명 함수(`a_windows.go` 처럼 정렬이 앞서는 파일)는 가려짐 | `seal_census_test.go:247-297`(코드 읽기, 실행하지 않음) | 한계를 머리말에 명시할 것 |
| 9 | T | 계보 identity 조건(`result.Lineage.Identity != …`)은 행동상 중복임. 조건 identity 가 `lineageIdentity` 를 해시하기 때문(types.go:315). 「identity 대조 셋」은 실제로는 ExecutionTerms 조건으로 걸림 | M1(계보 조건 삭제)을 잡은 것은 **구조 시험 `TestTheFirstLegBackstopComparesTheSealedIdentitiesThemselves` 하나뿐**임 | 구조 핀이 있으니 수용 가능함. 문서에 한 줄 적을 것 |
| 10 | T | 분기 재배열 중 둘이 살아남지만 무해함: M8(identity 를 개수 관문 앞으로)과 M10(선택을 준비 검사 앞으로). 둘 다 어느 순서든 거절함. M7(개수 관문을 선택 앞으로 = 편집 전 순서)은 HoldsTwice 가 유일하게 잡음 — 원장에 없음 | mut.py 실행 결과 | M7 을 원장에 추가할 것 |
| 11 | T | 변이 원장의 「N failing」이 이중 계산됨. 같은 시험이 tagged 명령과 무태그 명령에서 두 번 세어짐(S07 은 같은 이름이 2번, S01 은 「10」). `a112_lot_mutate.py:run_tests` 는 `--- FAIL:` 텍스트를 파싱하는데, 이것은 #4 와 같은 위조 채널임. S06 라벨 「the pre-seal shape」도 정확하지 않음: 편집 전 모양은 S06+M7 임 | 원장 tsv | 서로 다른 시험 수를 셀 것 |

## 필수 항목별 판정

1. **RED 진정성: 조건부 통과.** 실패는 기능 부재가 맞음(재현함). 다만 로그는 커밋본보다 이전 판본의 시험 파일에서 나왔고 A-lite RED 가 빠져 있음(#5). 편집 전에도 PASS 하던 시험 각각을 **유일하게** 빨갛게 하는 변이는 원장에 없음. 그래서 직접 쟀음:
   - 같은 범위 패자: K1(선택 키에 CampaignID 추가) → 위조 축 하위 시험 중 이것만 빨감(구조 시험 · HoldsTwice 와 같이).
   - 게이트된 레인: K2(선택 키에 LaneID 추가) → 이것만 빨감(구조 시험과 같이).
   - 조건 재작성: 봉인 시험 중에서는 S03 이 이것만 빨갛게 함.
   - 「개수 관문 앞 선택」: M12 · M2 · S04 가 잡음(유일하지는 않음). 이 시험에 적힌 판별 대상 주장은 거짓임(#3).
   - 판별력은 있으나 원장이 그것을 보이지 않음.
2. **변이 판별력: 부분 통과.**
   - S02 는 identity 로 고르기(자기 참조)를 제대로 모사함.
   - S06 은 편집 전 모양이 아님(#11).
   - 원장 밖에서 살아남은 변이는 M3 · M4 · M5 · M6 · M8 · M10 임. 안전에는 영향 없고 주장에는 영향 있음(#6 · #10).
   - A-lite 는 잘 고정돼 있음: S08/S09 원장에 있음, M9(계약 digest 역순)는 `TestTheProposalSetDigestMatches…` 가 잡음, A1(개수만 봄) · A2 도 CAUGHT.
3. **`verify_named_tests.py`: 좁은 범위에서만 유효함.**
   - 잡는 것: 조용한 exit(E1 재현), `-count=1` 캐시, 빌드 실패, 최상위 skip, 이름 접두 충돌(`^(…)$` 로 앵커).
   - 못 잡는 것: framing 위조(E2 · E3), 하위 시험 skip, 조기 return(#4).
   - rtk 는 관계없음: subprocess 로 `/usr/local/go/bin/go` 를 직접 부르고, PATH 에 다른 go 는 없음.
4. **strategyflow census 완전성: 실패(#1 P1, #2 P2).**
   - 구조체 필드 · 포인터 경유 대입(`(&r).proposalSeal = …`)은 SelectorExpr 라 잡힘.
   - 뚫리는 모양: 함수 값 별칭 · 색인 쓰기 · `copy` · `*p` · `unsafe`(import 가능), 그리고 `!cgo`/GOOS 파일을 「tagged」로 오판하는 것.
   - ③ 은 문자열 포함만 봄(#8). 재생성 env 는 같은 실행에서 통과함(#7).

**생산 동작.** 편집 후 발급 조건은 편집 전 발급 조건의 부분집합임(준비 + 유일한 범위 + 개수=1 + identity). `NewOwnerKey` 는 세대 0 만 거절하는데, 1차 레그는 `cas+1 ≥ 1` 을 요구하므로 새로 거절되는 정상 입력은 0 으로 봄. 오류는 모두 `StrategyFirstLegAuthorityCollectionFailed` 하나로 매핑되므로(`strategy_first_leg_admission.go:77-79`) 바뀌는 것은 상세 문구뿐임. 활성화가 없으면 A-lite 는 닫혀 있음(S09) → 토글 OFF = upstream 이 유지됨.

Recommendation: 생산 코드는 그대로 두고 착지 전에 다음을 할 것 — (1) census 의 「tagged」 판정을 `tossos_testseams` 제약 기준(또는 생산 빌드 맥락 기준)으로 바꿀 것, (2) ② 를 go/types 의 필드 참조 · 식별자 Uses 기반으로 바꾸고 `unsafe` 를 금지할 것, (3) #3 의 거짓 진술 세 곳을 고치고 M7 · M12 · K1 · K2 를 원장에 추가할 것, (4) RED 를 커밋된 시험 파일로 다시 기록할 것(A-lite 포함), (5) `verify_named_tests.py` 의 주장을 좁히고 하위 시험 skip 을 실패로 볼 것, (6) 재생성 env 는 t.Fatal 로 멈출 것. (1)과 (3)이 끝나면 APPROVE 로 볼 수 있음.
