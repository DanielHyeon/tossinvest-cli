# a187 gstack 리뷰 (읽기 전용)

- 대상: `/home/daniel/tossos-work/a187-quote-symbols`, 브랜치 `fix/a187-strict-quote-kr-alphanumeric`, base `6dbb752f6`
- 리뷰 시작 HEAD `d5a2738bf`. **리뷰 도중 HEAD 가 `3e46e8408` 로 움직임**(문서 커밋 2개: `87b643db1`, `3e46e8408` — tasks 4.1 실측 기록). 두 커밋은 `analysis/measurement/l1c-0158S0-2026-11-12.md` 추가와 `tasks.md` 한 줄뿐이고 코드·시험은 `d5a2738bf` 와 같음(`git diff --stat d5a2738bf..HEAD` 2 files). 두 커밋은 내가 직접 읽었음. specialist 중 어느 것이 새 커밋을 봤는지는 red-team 외에는 모름.
- 스킬: gstack `review` 를 Skill 도구로 실행함(수동 체크리스트 대체 아님). gstack 1.98.0.0.

## 판정

**Pre-Landing Review: 7 issues (0 critical, 7 informational) — blocker 0.**
운영 코드 변경(정규식 1개, 래퍼 1개)에서 런타임 오류나 안전 불변식 위반은 찾지 못함. 넓어진 문법이 주문·체결·사이징·Guardian·원장·대사 경로로 가는 새 길은 없음. 남은 것은 스펙 문장·문서 사실 주장·시험 못 박기의 어긋남이고 archive 전에 고칠 것이 1건(major) 있음.

- PR type: mixed (application + docs)
- Scope Check: CLEAN. Intent = 엄격 호가·현재가 리더가 KR 영숫자 6자리를 받음. Delivered = 같음(운영 코드 1파일 16줄, 시험 3파일, 문서·PM). Diff read: full(코드·시험·문서·PM 전부, 생성 산출물 `analysis/*.json`·`*.txt` 는 주장 확인에 필요한 것만). Undeclared behavior changes: none(영숫자 종목의 실제 GET 증가는 proposal Impact 에 적혀 있음).
- Gate edits: none detected (3 listed, 1 read; patterns: RH-1,RH-3,RH-4,RH-12,RH-13,RH-14,RH-15,RH-16; unlisted idioms: none). 읽은 1건은 아래 「거절 사례 2줄 삭제」.
- 완결성: native 패스(체크리스트·specialist 5·red-team·adversarial)는 모두 반환됨. 외부(Codex) 패스는 미실행 — outside coverage 없음. Fix-First·review-log 저장은 지시대로 하지 않음.

## 발견사항

| # | 분류 | file:line | 문제 | 근거 |
|---|---|---|---|---|
| 1 | major | `openspec/changes/a187-…/specs/official-read-compatibility/spec.md:28` · `tools/a087-market-sell-probe/order_test.go:31-33` | 스펙 시나리오 「Order policy is not widened」가 「the a087 request-schema test are evaluated with KR 0101N0」이라 적지만 그런 시험이 없음. archive 하면 이 문장이 정본이 됨. | `rg 0101N0 tools/a087-market-sell-probe/` 0건(rc=1). a087 종목 거절 행은 `AAPL`·`05930`·`0059300` 뿐(order_test.go:31-33). `order.go:49` `krSymbol = ^[0-9]{6}$` 를 영숫자로 넓히는 변이를 잡는 6자리 영숫자 행이 없음. review.md r1 표 4행도 「영숫자 KR 로 a087 스키마를 돌린 시험은 없음」이라 적음. 출처 3개 일치(core·testing·adversarial), confidence 9. 고치는 길 둘: 시나리오를 실제 시험(시장 추론 고정 + NormalizePlace 현 동작 고정, a087 무편집)으로 고쳐 쓰기, 또는 a087 거절 표에 `0101N0` 한 줄 추가. |
| 2 | minor | `internal/official/strict_minute_candles_test.go:196-206` · `a187_kr_alphanumeric_test.go:19,46` | 스펙의 「one shared lexical predicate」「each reader sends exactly one request」가 분봉 쪽에서는 못 박히지 않음. a187 표 두 개는 호가 리더만 돎. | 분봉 수락 시험은 `0101N0` 하나이고 `page.Symbol`·`lastQuery()` 만 보며 `hitCount` 단언이 없음(:204). 분봉 쪽에는 개행·앞자리 영문 사례가 없음. 래퍼가 자기 정규식을 다시 들이는 변이(예: 여러 줄 모드)는 살아남을 것으로 봄 — 변이 9종 중 래퍼 변이는 `return nil` 하나뿐(mutation.py:36). 정적 판단이고 변이는 돌리지 않음. testing, confidence 7. |
| 3 | minor | `internal/official/strict_minute_candles.go:52` | 새 운영 코드 주석 「주문 정책(internal/orderintent)은 숫자 전용 그대로」가 r1 발견 1 로 정정한 사실과 어긋남. 숫자 전용인 것은 시장 추론뿐임. | 같은 change 의 시험 주석 `a187_order_policy_test.go:6` 「주문 검증 전체가 숫자 전용이라는 뜻이 아님」, design D4 「「주문 검증이 숫자 전용이다」는 사실이 아니므로 적지 않음」, 그리고 `TestA187NormalizePlaceExplicitKRAlphanumericUnchanged` 가 `kr`+`0101N0` 통과를 고정함. 주석만 고치면 되지만 a184 브랜치가 이 파일을 `d5a2738bf` 기준으로 복사해 갔으므로(`f7a2d5087`) 다시 가져가야 함. maintainability + adversarial, confidence 8. |
| 4 | minor | `internal/official/strict_quote_reads_test.go:78,93,98` | 지운 행 `{"kr alphanumeric minute-only symbol", "KR", "0101N0", StrictQuoteReasonSymbol}` 이 `wantReason` 을 채운 유일한 행이었음. 남은 7행은 모두 `""` 라 사유 비교 갈래가 실행되지 않음(죽은 단언). | :80-86 넷째 값 전부 `""`, :93 `testCase.wantReason != "" && …`. KR 종목 사유 고정은 `TestA187StrictQuoteReadersStillRefuseOtherKRForms`(11모양 × 2리더, 무조건 `StrictQuoteReasonSymbol`)로 옮겨져 오히려 넓어졌고, 시장 사유는 :354 에 따로 있음. US 종목 사유는 base 에서도 고정돼 있지 않았음. testing + maintainability, confidence 9. |
| 5 | minor | `…/tasks.md:24` · `proposal.md:46` · `docs/pm/portfolio/stories/STORY-TOS-a187.yaml:20` | 프로브 전후 경계를 「main 반영 시각」으로 적었으나 프로브는 기동 때 컴파일됨. 그리고 「원장만으로는 반영 전후를 가를 수 없음」은 코드와 다름. | `openspec/changes/a128-jev-shadow-judgment-probe/run-probe.sh` 마지막 줄 `exec go run ./tools/a128-jev-shadow-probe run …` — 반영이 세션 중이면 그 세션 남은 회차는 옛 문법. design Migration 은 「다음 프로브 기동부터」라 tasks 4.2 와 서로 다름. `collect.go:456` `gaps = append(gaps, "top_of_book: "+reasonOf(err))` → `:562` `CollectionGaps: entry.state.MissingInputs` 라 반영 전 영숫자 종목 행에는 `SYMBOL_INVALID` 문구가 행마다 남음. adversarial, 인용 줄은 내가 확인함. confidence 8. |
| 6 | minor | `…/design.md:58-60` · `proposal.md:47` · `tasks.md:23` | 4.1 실측을 완료로 체크했는데(87b643db1) design Open Questions 와 proposal Impact 는 아직 「모름」으로 남음. 실측은 D5 조건과 다름. | 실측 기록 자체가 편차를 적음: 「한국 정규장 밖」, 「prices: HTTP 200 digest (가림…)」 — task 문장은 「본문 digest 만 기록」인데 현재가 digest 가 없음. 종목 1개·1회. red-team, confidence 8. |
| 7 | minor | `internal/orderintent/a187_order_policy_test.go:9` | 시장 추론을 비공개 술어로만 고정함. `intent.go:101` 의 호출을 더 넓은 인라인 검사로 바꾸는 변이는 통과할 것으로 봄. | `NormalizePlace(Market "us" 또는 생략, Symbol "0101N0")` 가 `us` 로 남는지 보는 시험이 a187 에 없음(있는 것은 명시 `kr` 한 건). 정적 판단. testing, confidence 7. |
| 8 | note | `internal/official/strict_quote_reads.go:46,156` · 실측 기록 | 호가 본문에 종목 필드가 없어 `top.Symbol` 은 요청값임. 실측은 자릿수만 남겨(ask·bid·last 모두 정수 5자리) 호가가 그 종목 것인지는 증명하지 않음. | 현재가 쪽은 `:395` 메아리 검사로 묶임. a143 부터의 한계이고 design Risks 에 있음. a184 가 ask 를 진입 지정가로 쓴다는 것은 adversarial 의 다른 브랜치 인용이며 나는 확인하지 않음. |
| 9 | note | `internal/official/strict_minute_candles.go:53-54` | KR `^[A-Z0-9]{6}$` 과 US `^[A-Z][A-Z0-9.]{0,9}$` 가 처음으로 겹침(`A12345`·`ABCDEF`). 요청에는 시장 인자가 없어 호가에서는 통화 검사만 가름. | 새 수락 시험이 이 두 값을 KR 로 고정함. 실재 KR 코드는 영문자가 5·6번째 자리뿐이라(design D3) 지금은 이론적임. |
| 10 | note | design Risks | a184 실거래 전환에 넘길 「숫자 전용 시장 분류」가 `orderintent` 하나만 적혀 있음. | 같은 가정: `internal/verifylive/pricing.go:359-365`(MarketOf → "US"), `internal/flatten/flatten.go:570`, `internal/client/trading.go:511`, `tools/a087-market-sell-probe/order.go:49`, `internal/official/candle_raw.go:22`. 모두 이 change 가 안 고친 코드임. |
| 11 | note | delta spec :13 | 「No … strategy universe … SHALL change」 문장과 change 의 목적(문법 거절로 빠지던 종목이 후보에 남음)이 읽기에 따라 어긋남. | 유니버스 코드는 안 고침. 문법 거절이 유일한 거름이던 소비처의 사실상 후보는 넓어짐. 같은 문장이 정본 「Strict engine candle contracts are preserved」에도 이미 있음. |
| 12 | note | `analysis/tdd/check-analysis-2.3.txt` | FLM 번들이 동작이 안 바뀐 래퍼에 붙고, 동작이 바뀐 `strictMinuteCheckSymbol`(패키지 변수 값 변경)에는 편집 뒤 BTM 이 없음. | gate ⑤ 는 본문 편집만 셈. 시험과 변이 9종이 실질을 덮음. 게이트 사각으로 기록할지 판단 사항. |
| 13 | note | `tools/a112-l1c-quote-probe/main.go:36` | 도움말 「KR six digits」가 낡음(4.1 실측에 쓴 도구). diff 밖. | `flag.String("symbol", "", "required: the exact broker symbol (KR six digits, US ticker)")` |
| 14 | note | `docs/pm/portfolio/features/FEAT-TOS-014.yaml:9` 등 | a185 브랜치가 같은 줄을 고침(FEAT-TOS-014 stories 끝, `_registry.yaml` 끝, generated 표 3개). 둘째로 반영되는 쪽이 충돌함. | `git diff 6dbb752f6..fix/a185-official-commissions-account-header -- docs/pm` 6 files, 같은 `"STORY-TOS-a143"` 줄. red-team 이 덧붙인 「정본 스펙 꼬리도 겹친다」는 확인 안 됨(a185 에 스펙 delta 가 안 보임). |
| 15 | note | a184 브랜치 `f7a2d5087` | a184 가 a187 의 코드·시험 4파일을 `d5a2738bf` 기준으로 이미 복사해 감. 3·5·7번을 고치면 사본이 낡음. | 커밋 제목 「a187 의 코드·시험 4파일을 가져옴(출처 a187 d5a2738bf…)」 확인. |
| 16 | note | `analysis/mutation/mutation.py:55-62` | `summary()` 가 `-v` 출력의 들여쓴 `_test.go:NN:` 줄을 모두 「실패 원문」으로 찍어 통과 시험의 `t.Log` 가 섞임. 판정에는 영향 없음. | 판정 로직은 직접 읽음: 대상 문자열 1회 단언(:94), 빌드 실패는 INVALID(:106), 대조군 GREEN 선행(:83-88), 원복 SHA 단언(:102). |
| 17 | note (advisory) | `internal/official/strict_minute_candles.go:213` | 한 줄 위임 래퍼를 없애는 후속이 어디에도 적혀 있지 않음. a187 에서 지우라는 뜻 아님(design D2 사유 타당). | simplification, 약 5줄. |
| 18 | note | worktree `openspec/config.yaml` | 추적 안 되는 파일이 worktree 에 있음(`schema: spec-driven`, 2026-11-12 20:09 생성). 리뷰 시작 전부터 있었고 내가 만든 것 아님. | `git status --short` → `?? openspec/config.yaml`. 주 체크아웃에는 없음. |

### 볼 것 네 가지에 대한 답

- (a) 주문 등 High-risk 경로로 가는 새 길: **없음.** 검사기 참조는 `strict_minute_candles.go`·`strict_quote_reads.go:201` 뿐. 두 리더의 시험 외 호출은 7자리(`officialbars/quote.go:113·117`, a128 `collect.go:453`, a140 `source.go:148·154`, a112 `main.go:87·91`)이고 `PollQuoteL1` 운영 호출은 a140 `shadow.go:289` 하나. 이 네 패키지의 시험 외 import 에 trading·orderintent·hybrid·engine 이 없음(a140 의 `internal/journal` 은 엔진 저널 옆 경로를 거절하는 가드 `guard.go:23-56`). a140 유니버스는 `shadow.go:275` 에서 숫자 6자리만 받음. 단 **인증·공유 조회 예산에는 새 실제 GET 이 닿음**(r1 발견 2 와 같음).
- (b) 거절 사례 2줄 삭제: **정당함.** 같은 diff 의 REMOVED 요구사항, 수락 시험으로의 이전(`00593A`·`0101N0` 이 `TestA187StrictQuoteReadersAcceptKRAlphanumeric` 에 있고 요청 수·query·종목·통화를 봄), `numeric-only` 변이 CAUGHT 가 근거. 부작용은 4번(죽은 `wantReason`). 시험이 주장을 재는지는 1·2·7번이 남은 틈.
- (c) 문서와 코드 불일치: 1·3·5·6번.
- (d) 스펙 delta 와 archive: REMOVED 머리글이 정본 `spec.md:50` 과 글자까지 같음. ADDED 머리글은 새 이름. a169 의 같은 capability delta 는 다른 요구사항 둘을 ADDED 할 뿐이라 겹치지 않음. 정본 「Strict engine candle contracts are preserved」의 영숫자 문장과는 오히려 일관돼짐. 충돌은 못 찾았고, 남는 문제는 1번 문장이 정본이 된다는 것.

## 스킬 실행과 건너뛴 단계

- 디스패치: **Testing, Maintainability, Security, Performance, Simplification**(5) + **Red Team**(DIFF_LINES 13,577 > 200) + **Claude adversarial**. 건너뜀: Data Migration·API Contract·Design·State Machines(scope 신호 없음). 게이트된 것 없음.
- 반환: Testing 4건, Maintainability 3건, Security 결함 0(coverage-note 1), Performance `NO FINDINGS`, Simplification advisory 1건, Red Team 4건, Adversarial 8건(Medium 2·Low 6, Blocks 전부 no). CRITICAL 0.
- SPECIALIST REVIEW: 7 findings (0 critical, 7 informational) from 6 specialists. PR Quality Score 6.5/10.
- Security·Adversarial·Red Team 은 스킬 규정대로 시험 파일을 요약 모드(`diff --stat`)로만 봄. 시험 본문은 Testing·Maintainability 와 내가 전부 읽음.
- 질문(AskUserQuestion)은 하나도 내지 않음. 권장 선택지로 진행한 것: 업그레이드 안내(1.98.0.0→1.99.1.0)는 올리지 않음, 플랜 후보 선택은 「바인딩 없음」으로 두고 OpenSpec `tasks.md` 를 의도 출처로 씀, cross-project learnings 는 프로젝트 범위 그대로.
- 건너뛴 단계와 사유:
  - `git fetch origin`(Step 1): 네트워크·git 상태 변경 금지. Base coverage: 로컬 `origin/main` = `6dbb752f6` 기준, 원격 최신 여부는 미확인.
  - Step 2.5 리뷰 봇 댓글, 3.4 버전 큐, Aside/WebSearch 조사, slop scan: 네트워크 또는 해당 없음.
  - Step 4.7: QA evidence 도구·charter 파일은 쓰지 않음(파일 쓰기 금지). 대신 지정된 시험 명령을 직접 돌림. scope·exploratory 섹션을 읽기 전에 첫 시험을 돌린 순서 어긋남이 있었음.
  - Step 4.8 Codex 패스: 미실행. 사전 점검이 외부 유료 호출이라 네트워크 금지에 걸림. 다른 계열 리뷰는 r1(gpt-6-astra)이 이미 있음.
  - Step 5 Fix-First(AUTO-FIX·ASK·test stub 적용), VERSION/CHANGELOG, Step 5.8 `gstack-review-log` 저장, learnings 기록: 읽기 전용 지시로 하지 않음. 발견은 전부 open.
- 파일 쓰기(worktree 밖): gate 스캔 산출물 `~/.gstack/projects/DanielHyeon-tossinvest-cli/gate-diff/6dbb752f63fe+206dcb75b831.jsonl`(도구 부산물), skill-start/end 세션 기록, 스크래치패드의 이 보고서와 `a187-targeted-test.txt`. worktree 는 `?? openspec/config.yaml`(기존) 외 변화 없음, HEAD `3e46e8408`.

## 시험 실행 (HEAD 3e46e8408, 코드 = d5a2738bf)

- `rtk proxy go test -count=1 ./internal/official/ ./internal/orderintent/` → **rc=1**. `internal/official` FAIL 1건: `TestA112MBUSStaticNoProductCallerOrForbiddenPath` — `forbidden resolved A112 M-B0 references: []string{"internal/accountbudgetruntime/v2_token_test.go: selector official.A112MBUSCandle"}`(base 기존 결함으로 보고된 것과 같은 문구, base 재현은 내가 다시 하지 않음). `internal/orderintent` ok.
- `go test -count=1 -v -run 'TestA187|TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest|TestStrictMinuteCandlesAcceptsDomesticAlphanumeric|TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest|TestStrictLastPrice|TestStrictOrderbookTop'` 같은 두 패키지 → rc=0, 최상위 PASS 15, 하위 PASS 85, FAIL 0, SKIP 0. TestA187* 6개 모두 PASS.

## 확인했고 문제없던 것

- 정규식 `^[A-Z0-9]{6}$`: 양끝 고정, ASCII 전용. 분봉 리더 수락 집합은 편집 전후 같음(편집 전 `alnum6 ∪ num6` = 편집 뒤 `alnum6`).
- 주입: query 는 `url.Values` 를 거치고 문자 집합이 대문자·숫자뿐. a128 state 파일명은 내용 해시, evidence SQL 은 바인딩(Security specialist 추적).
- 호출량: a128 은 리더 호출 전에 계수(`collect.go:189`)해 상한 우회 없음. 영숫자 종목당 실제 GET 이 느는 것은 proposal 문장과 맞음(Performance `NO FINDINGS`).
- `internal/orderintent` 무편집(시험 파일 추가뿐).
- 변이 하네스 판정 로직(16번 참조).
- PM 등록: Story 의 feature·change 경로·수락 기준이 change 문서와 맞음.

## 보지 못한 것

- 변이·`check_analysis.py`·`openspec validate`·전체 시험·전체 gate 는 다시 돌리지 않음(기록만 읽음). 2·7번의 「변이가 살아남는다」는 정적 판단.
- base 에서 깨진 정적 가드 5개의 base 재현.
- a184 브랜치 코드(8·15번의 a184 쪽 인용은 specialist 의 것).
- 공식 API 실제 응답(4.1 기록만 읽음). `18064K` 형 코드는 실측에 없음.
- 생성 산출물 전수(AST json·시험 로그 약 13,000줄)는 주장 확인에 필요한 것만 열었음.
- ai-memory·Hindsight recall 은 하지 않음 — 그 채널은 안 봤음.
- 외부 모델(Codex) 패스.
