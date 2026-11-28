# gstack review 원문 2 (a189 범위 확장 2 포함 최종 diff, 2026-11-16)

- 스킬: gstack `review`(preamble 4e82fa4c839f), SESSION_KIND interactive(구현 팀메이트가 subagent 로 돌림 — AskUserQuestion 대신 Manager 에게 올림), REPO_MODE solo.
- 대상: base `7d1e3163a`(= origin/main, BASE_REFRESH fresh) .. 워킹트리. 범위는 a189 전체(앞 로트 + 범위 확장 2). Go·corpus diff 는 `final-diff-go*.patch`(pass1 466줄 → pass3 609줄), 전체 diff 는 증거 문서 포함 약 19.8k 줄.
- Plan completion audit: not run (no plan is bound to this branch and no docs/designs/ file matches). 의도 출처는 커밋 메시지와 OpenSpec proposal·design·tasks.
- Banned terms: none. Approval classes: none. PR 없음(봇 코멘트 없음). Contributor mode off(PR 작성자 판별 불가).
- Gate scan(pass1): `GATE_SUMMARY: listed=17 eligible=7 inspected=7 unread=0 tagged={RH-3?:1}`. pass3: `listed=18 eligible=8 inspected=8 unread=0 tagged={RH-3?:1}`.
- 보이스: 전문가 subagent(testing·maintainability·security·performance·simplification·state-machines), red team, Claude 적대 패스, Codex 적대 패스(gpt-6-astra, read-only 샌드박스, 다른 계열). Codex 구조 리뷰(DIFF_TOTAL≥200 해당)는 돌리지 않음 — 적대 패스로 갈음(아래 「안 한 것」).
- 회차: 3회(각 회차가 고침을 낳음 — skill 상한 3 에 닿음). **수렴하지 않음**: 3회차 발견을 고친 뒤 4회차 리뷰는 돌리지 않음. 3회차 수리는 사본 변이 51/51 기대대로(`mutation-2/mutation2-run7.txt`)로만 확인.

## 1회차 (HEAD 027f7a04f)

| # | 출처 | 등급 | 발견 | 처리 |
|---|---|---|---|---|
| 1 | testing(5/10)·Claude 적대 F1(Medium, 실측) | INFO | `admitSites` 가 `strings.Trim(…, "\"")` 로 import 경로를 읽어 원시 문자열·이스케이프(`\x68`) 철자의 경계 import 를 못 알아봄 — 둘째 주조 자리가 census 를 통과(fail-open, gofmt 가 정규화하므로 lint 를 건너뛸 때만) | 고침(f5e13ea7d): `strconv.Unquote`, 해독 실패는 경계로 간주. 시험 `TestA189AdmitCensusDecodesTheImportPathLiteral`, 변이 a7·a8·g8 |
| 2 | maintainability(8/10) | INFO | `TestExactlyOneProductionSiteAdmitsIntoTheSeam` 주석이 좁힘 전 범위를 말함 | 고침(f5e13ea7d) |
| 3 | Claude 적대 F2·Codex 2·red team(셋 일치) | INFO(Medium) | orders seam AST 단언이 안 쓰는 참조 `_ = l.shared.resolve` 한 줄로 뚫림(o4 + decoy) — design D11 의 「o4 를 AST 판본이 잡음」 근거가 무너짐 | 고침(f5e13ea7d·e6fbcb3bb): 직접 호출이 broker 를 채우고 broker 에 대한 대입은 그 문뿐이며 단언 대상이 broker. 변이 o5(기대 FAIL 로)·o7·o8·o9·o10 |
| 4 | Codex 1(Medium) | INFO | 엔진 안 `//go:linkname Admit …strategyhandoff.Admit` 몸체 없는 선언 — 맨 이름이 dot import 없는 파일이라 안 셈 | 고침(f5e13ea7d): `TestA189EngineProductionNeverImportsUnsafe`(linkname 은 unsafe import 파일에서만 컴파일). 변이 a9 |
| 5 | Claude 적대 F3(Low) | INFO | D12 「나머지 9개는 경계 세 지점에서 같음」 — 하네스가 exitloop.go 만 렌더해 `noteHeldProposal` 미측정 | 고침(015f9f17d): 엔진 생산 파일 88개 전수 렌더, 1차 원문은 `history-render-at-commits-exitloop-only.txt` |
| 6 | Codex 3(Low) | INFO | `compare_record_ast.py` 는 조건식·인자를 안 봐 「판정 불변의 근거는 이 대조뿐」이 과장 | 고침(015f9f17d): 주석 없는 렌더 원문 diff 를 하네스에 더함 — 차이는 clearCtx 네 줄 추가와 호출 첫 인자 하나. D12 근거를 원문 diff 로 바꿈 |
| 7 | red team(3/10)·Claude 적대 F4(Low) | INFO | 표지 없는 context 의 기본값이 Normal — 보조 등급 고정은 명시 표지만 잼 | 근거 있는 기록: design R5, **Manager 결정 대상** |
| 8 | Claude 적대 F5(Low) | INFO | 이름 단위 금지가 `reflect.MethodByName` 으로 비켜 감(PlaceOrder 에도 있던 기존 한계) | 기록: design R6, **Manager 결정 대상** |
| 9 | (core sweep) | INFO | `TestEveryNameTheEngineTakesFromTheSeamIsClassified`(이 diff 가 안 바꾼 기존 함수) 도 `strings.Trim` | 후속(범위 밖 기존 코드 — 고치면 FLM 번들 하나 추가). 그 시험은 경계 문 분류를 보므로 1번 수리로 주조 자리는 이미 잡힘 |
| — | security·performance·simplification·state-machines(`STATE_MATRIX_CHECK: ok machines=0`) | — | NO FINDINGS | — |

## 2회차 (HEAD 015f9f17d)

| # | 출처 | 등급 | 발견 | 처리 |
|---|---|---|---|---|
| 10 | testing(9/10, 실제 저장소 사본 실측) | INFO | 엔진 밖 패키지의 linkname 은 엔진 unsafe 금지로 안 닫힘 — R4 l3 근거도 다시 열림 | 고침(af30602cc): 모듈 전체 linkname 검사 `TestA189NoPackageLinknamesTheSeamDoors`, 변이 a10 |
| 11 | security(6/10)·red team F2(Medium) | INFO | 가림 세기가 이름 기반 — range·함수 리터럴 인자 가림, 둘째 단언으로 뚫림 | 고침(af30602cc): 단언 대상을 `ast.Ident.Obj` 로 대조, consoleOrdersReader 단언은 본문 전체에서 하나만. 변이 o11·o12·o13 |
| 12 | maintainability ×2 | INFO | 주석(「셋」 → 다섯 수, a189 파일 머리 목록) | 고침(af30602cc) |
| 13 | security(6/10) | INFO | `render_at_commits.sh:55` 의 diff 가 `set -e` 로 차이가 있을 때 멎음 | 고침: `|| true`, 재실행 원문 동일 |
| 14 | security(5/10) | INFO | `mutation2.py` apply 가 중간 실패 시 앞 편집을 디스크에 남김 | 고침: 실패 시 원복 뒤 재던짐 |
| 15 | security(4/10) | INFO | 앞 로트 `census-repair/falsify.py` 가 /tmp 사본·HOME 격리 없음(그 시험은 HOME 을 안 읽음 — 리뷰어 확인) | 기록(잠재만, 앞 로트 증거 — 다시 돌리지 않음) |
| 16 | **red team F1(High, Blocks: yes, 실측)** | CRITICAL | 별칭 `ab "…/internal/accountbudget"` 의 `ab.WithPriority(ctx, ab.Critical)` 이 보조 등급 고정(`types.ExprString(call.Fun) == "accountbudget.WithPriority"`)을 비켜 감 — a128 라벨 가격 경로 사본에서 a128·a140 전체와 a166 census 가 ok | **Manager 결정 대상(고치지 않음)**: design R1 은 2026-11-15 Manager 결정으로 범위 밖이었고 그 결정은 보조 등급 고정(ab7023901)보다 앞섬. design R1 에 「재검토 필요」 로 기록 |
| 17 | red team F2 나머지 | INFO | 단언 뒤 흐름(`reads`·`plain` 다시 대입)과 console.go 밖 `official.New(` | 기록: design R7(편집 전부터 있던 한계 — 후속) |
| 18 | red team F3(Low) | INFO | 다른 패키지의 linkname — 단언 claims 과장 | 10번과 같은 수리 |
| — | performance·simplification·state-machines | — | NO FINDINGS | — |
| — | Codex 2회차 | — | 돌리지 않음(3회차에 돌림) | — |

## 3회차 (HEAD 9f364bb75)

| # | 출처 | 등급 | 발견 | 처리 |
|---|---|---|---|---|
| 19 | testing(9/10, 실측)·red team F1(Medium, Blocks: yes, 실측)·Codex 1(Medium) — 셋 일치 | INFO | linkname 검사가 testdata·node_modules·vendor 를 건너뛰고 디렉터리 심링크를 안 따라감 — testdata 아래 중계 패키지가 import·실행됨 | 고침(d9d6772fc): 점 디렉터리만 건너뜀, 디렉터리 심링크는 실패. 변이 a11·a14, 무해 대조 b3 |
| 20 | red team F2(Medium, Blocks: yes)·security(8/10, push 형태 실측) | INFO | 대상 이름 필터가 엔진 사설 dispatch 대상·경계 패키지 안 push 형태를 놓침 | 고침(d9d6772fc): linkname 지시어 전부 금지(오늘 생산 사용 0). 변이 a12·a13 |
| 21 | red team F4(Low, 실측) | INFO | `&broker` 포인터 쓰기로 broker 를 바꿔도 통과 | 고침(d9d6772fc): `&broker` 를 다시 대입으로 셈. 변이 o14 |
| 22 | red team F5·security(5/10) | INFO | 모듈 걷기가 go 가 무시하는 파일(`.#x.go` 등)·병행 세션의 반쯤 쓴 파일을 파싱하다 실패할 수 있음(fail-closed 소음) | 고침(d9d6772fc): 파싱하지 않고 줄 머리만 봄, 점·밑줄 시작 파일 이름 제외 |
| 23 | maintainability(7/10) | INFO | a163 이유 주석이 「네 줄뿐·인자 동일」 — 실제로는 호출 첫 인자 ctx→clearCtx 도 바뀜 | 고침(d9d6772fc) |
| 24 | **red team F3(High, Blocks: no)** | CRITICAL | dot import 로도 16번과 같은 우회 | **Manager 결정 대상**: 16번과 같은 수리로 닫힘 — design R1 에 함께 기록 |
| — | simplification·state-machines(`ok machines=0`) | — | NO FINDINGS | — |
| — | Codex 3회차 판정 | — | `VERDICT: FINDINGS`(19번 하나), 나머지 1회차 수리 셋은 「source and recorded outputs support」 | — |

## Gate edits (Step 5c — 사람 결정, 여기서는 open)

`Gate edits: 8 listed, 8 read, 3 findings — 0 kept, 0 restored, 3 open`(subagent 실행이라 자동 선택은 처분이 아님 — Manager 에게 올림).

| gate_id | 태그 | 경로 | 추천 | 근거(검증) |
|---|---|---|---|---|
| 49234f885d5c | RH-15(고정값 갱신) | internal/app/engine/a163_judgement_pin_test.go | B) Keep — justified | 원인 커밋 9c4bdf4de 와 그 부모의 실제 시험 PASS→FAIL, rev-list 전수 렌더(바뀐 커밋은 9c4bdf4de 와 그것을 들인 병합 둘), 렌더 원문 diff — verified(`render-hash/`) |
| aad7cf1b05f0 | RH-15(인식기 좁힘) | internal/app/engine/strategy_dispatch_handoff_guard_test.go | B) Keep — justified | 5c0aeb2b8 경계 실측, 진짜 둘째 자리 a1~a8 전부 FAIL, 잃는 검출 l1(R4)·l3(별도 시험이 막음) 명시 — verified(`mutation-2/`) |
| c98da0370da7 | RH-3? | internal/accountbudgetruntime/testdata/a166-wiring/shapes.json | B) Keep — justified | 앞 로트 Manager 리뷰 승인(review.md 「Manager 리뷰」) — verified |
| 0f741f24dd35 외 4 | untagged | console_test.go(강화), 앞 로트 가드 4건 | 판정 대상 아님(강화) / 앞 로트에서 Keep(B) | — |

## 안 한 것

- Codex 구조 리뷰(DIFF_TOTAL≥200 해당): 돌리지 않음. 적대 패스 2회(1·3회차)로 갈음 — 누락 커버리지로 적음.
- Codex 2회차 적대 패스: 돌리지 않음(시작 토큰만 받고 미실행으로 기록).
- 4회차 리뷰: skill 상한(편집 회차 3)에 닿아 돌리지 않음. 3회차 수리는 변이 51/51 로만 확인 — **수렴 미확인**.
- Exploratory QA(Step 4.7): 시험 전용 diff 라 기능 표면이 없음. 대신 대상 패키지 시험·태그 seam 시험·변이 하네스가 판정 근거(`verify-4/`, `mutation-2/`).
- Aside 웹 조사: 쓰지 않음(외부 라이브러리 판단 없음).

## Manager 결정 반영 (2026-11-16)

- 16·24번(별칭·dot import 로 보조 등급 고정 우회): 결정 1 로 고침 — 두 도구 import 시험이 세 패키지의 이름 붙인 import 를 금지. 변이 d1a·d1b·d1c(두 도구) FAIL, 결정 전 판본은 d1a·d1b PASS.
- 4·10·19·20번 수리로 더했던 시험 둘(엔진 `unsafe` 금지·모듈 linkname 금지): 결정 2 로 **삭제**. 해당 발견은 「편집 전부터의 맹점, 범위 밖(R8)」 으로 재분류. 변이 a9~a14 는 「잡지 못함」 으로 원장 갱신(run10).
- 7·8번(R5·R6)·R4 l1: 결정 3 으로 기록만.
- Gate 처분 3건: 결정 5 로 Keep(B) 승인 — `Gate edits: 8 listed, 8 read, 3 findings — 3 kept (Manager 승인: verified 근거), 0 restored, 0 open`.
- 리뷰 루프: 결정 4 로 Manager 가 닫음 — 수렴하지 않았고, 4회차는 돌리지 않음.
