# Evidence reconciliation — gate.sh 아카이브 인식 (task 1.11)

Date: 2026-09-10
HEAD: `96a073697ab7806269576504fd5f753c6d95a42f`
Verdict: 조정됨 — 단, hard evidence 가 대상 파일을 **덮지 못한다**는 것을 먼저 적는다

| 주장 | CodeGraph (hard) | 직접 확인 (현재 HEAD) | 해소 |
|---|---|---|---|
| 편집 대상 `tools/gate.sh` 의 호출 관계 | **답할 수 없음** — shell 미색인(`PAIR_DIR` 무결과) | `Makefile:169` 하나. `.github/` 에 참조 없음 | CodeGraph 를 근거로 쓰지 않고 `rg` 전수 열거를 근거로 쓴다 |
| `resolve_referenced_change` 의 호출자 | `check` 1건 | `check_analysis.py:690`, `:715` — 둘 다 `check` 본문 | 모순 아님(함수 단위 대 호출 자리 단위) |
| 해소 규칙을 공유할 수 있는가 | — | Python 함수를 shell 에서 부를 수 없다 | **공유 불가.** 규칙을 옮겨 적고 양쪽을 시험으로 못 박는다 |
| 활성이 아카이브를 가리는 경우 | — | `resolve_referenced_change` 는 `direct` 를 먼저 돌려주고 중복을 아카이브 안에서만 센다(`:249-255`) | gate.sh 는 **더 엄격하게** 간다 — 아래 |

## 조정된 구현 경계

이 편집은 `tools/gate.sh` 의 change 디렉터리 해소만 바꾼다. 5단계가 부르는
`check_analysis.py`, 판정 내용, base 규칙, 요구 함수 집합은 건드리지 않는다.
Go 파일은 바뀌지 않으므로 **수정된 기존 Go 함수는 0개**이고, 5단계가 이 편집에
요구할 산출물도 0개다(`Function Logic Map: not-applicable` — 사유는 이 문단).

production trading code, 주문·손절·사이징·Guardian·원장·대사·인증·체결 경로는
영향이 없다. 바뀌는 것은 사람이 손으로 부르는 완료 게이트의 디렉터리 탐색뿐이다.

## 남은 불일치 — 옮겨 적은 사본이 둘이 된다

Python 의 `resolve_referenced_change` 와 shell 의 새 해소기는 같은 규칙을 갖지만
정본이 둘이다. [[transcribed-code-needs-both-sides-pinned]] 가 말하는 모양이므로
**양쪽 다** 시험으로 못 박는다. 그리고 둘의 규칙이 갈리는 지점 하나는 고르지 않고
적는다: 활성과 아카이브에 같은 id 가 동시에 있을 때 Python 은 활성을 조용히 고르고
(a122 task 3.2.4 가 열어 둔 결함), shell 은 fail-closed 한다. shell 을 Python 에 맞춰
낮추지 않는다 — 3.2.4 가 Python 을 shell 쪽으로 올릴 것이다.

---

# Evidence reconciliation — Python 해소기 fail-closed (task 3.2.4)

Date: 2026-09-10
HEAD: `1f2a2d6d7907484929eb122c6218b8878b1458cd`
Verdict: 조정됨 — 그리고 위 「남은 불일치」가 여기서 닫힌다

| 주장 | CodeGraph (hard) | 기계 열거 (`analysis/python-function-logic/`) | 해소 |
|---|---|---|---|
| `resolve_referenced_change` 의 호출자 | `check` 1건 | `check` 안 **두 자리**: `:690`, `:715` | 모순 아님 — CodeGraph 는 함수 단위, 열거는 호출 자리 단위. 편집 범위는 **자리 단위**로 잡는다 |
| 활성이 아카이브를 가린다 | — | `B1 L239 if direct.is_dir():` → `L240 return direct` (early return) | 확인. 태스크 문구의 `:249-255` 는 raise 두 줄의 대략 범위였고, 정확한 좌표는 B1/L239-240 과 B6/L251-253 이다 |
| 중복을 아카이브 안에서만 센다 | — | `B6 len(matches) > 1` 이 세는 `matches` 는 `B2` 아카이브 순회의 산출 | 확인. 세는 **범위**가 반환 경로보다 좁다 |
| 해소기만 고치면 되는가 | — | `check` `B2 except ValueError:`(691)가 게이트 대상 경로에서 예외를 **삼킨다** | **아니다.** 호출 자리 하나를 같이 고쳐야 한다 — 이것이 이번 열거가 새로 준 사실이다 |
| 편집 반경 | — | `check_analysis.py --change` 의 production 호출자는 `tools/gate.sh:321` 하나. CI 는 안 돈다(`.github/workflows/ci.yml:94`) | 사람이 부르는 완료 게이트로 한정 |

## CodeGraph 가 답하지 못한 것과 그 처리

CodeGraph 는 이 함수를 **색인한다**(1.11 의 shell 과 다르다). 다만 돌려주는 단위가
함수라서 "한 함수 안의 서로 다른 호출 자리"를 가르지 못한다. 그 자리는
`enumerate.py` 의 AST 열거로 갈랐고, 그것이 이 편집의 범위를 하나에서 둘로 늘렸다.
CodeGraph 를 반증한 것이 아니라 **해상도가 다른 두 도구를 겹쳐 읽은 것**이다.

## 조정된 구현 경계

바꾸는 것: (1) `resolve_referenced_change` 의 세는 범위, (2) `check:689-692` 가
그 실패를 받는 방법. 바꾸지 않는 것: 요구 함수 집합, base·landing 규칙, 번들 판정,
아카이브 내 중복 메시지, 빌린 증거 경로의 메시지 형태. Go 파일 변경 0줄.

## 위 「남은 불일치」의 종결

1.11 이 shell 을 fail-closed 로 두고 Python 을 낮추지 않기로 한 그 갈림이,
이 편집으로 **Python 을 shell 쪽으로 올려** 닫힌다. 정본은 여전히 둘이므로
양쪽 시험을 각각 유지한다 — [[transcribed-code-needs-both-sides-pinned]].

---

# task 3.2.3.1 — 착지 선언을 고정할 증거가 0 인 경우 (2026-09-10)

| 주장 | CodeGraph | AST 열거 (`analysis/python-function-logic/`) | 조정 |
|---|---|---|---|
| `resolve_landing` 의 호출자 | `check` **1건** | `check` 안의 호출 자리도 **1건**(`L728`) | 일치. 3.2.4 와 달리 여기서는 갈리지 않는다 |
| 피호출 | `_committed_bytes` · `_is_ancestor` · `_ast_value` | 같음(+`git rev-parse` 직접 호출) | 일치 |
| 착지를 **고르지 못하게** 하는 것이 무엇인가 | — | B1~B10 은 "어느 커밋인가"만 묻는다. 고르지 못하게 하는 것은 `B11 L392` 순회 하나 | 확인 |
| 번들이 0 이면 | — | 순회 0회 → `mismatched` 빈 채 → `B19 L403` 거짓 → `L408 return candidate` **무조건** | 구멍의 기제. 손으로 읽었으면 "거절 여섯 개나 있다"에서 멈췄을 자리다 |
| 편집이 경로를 지우는가 | — | 반환 **3 → 3**(`366 ''` · `369 ''` · `424 candidate`) | 안 지운다. task 2.2("기록 없으면 오늘과 동일")가 구조로 남는다 |
| 편집 반경 | `check_analysis.py --change` 의 production 호출자는 `tools/gate.sh:321` 하나 | — | 사람이 부르는 완료 게이트로 한정 |

## CodeGraphContext 가 답하지 못한 것

`codegraph_context("landing point pinning in check_analysis …")` 는 Go 쪽
`internal/protectionreadiness/policy.go` 의 `pinnedKey` · `pinnedTrustPolicy` 를
돌려줬다. **이 경로와 무관하다** — 의미 검색이 "pinned" 라는 낱말로 다른 도메인에
붙었다. advisory 로 두고 쓰지 않았다. GBrain 은 이 세션에서 연결 실패
(`CONNECTION_CLOSED`)라 조회하지 못했다 — 침묵한 생략이 아니라 못 돈 것이다.

## 열거가 편집 범위를 늘린 자리 — 이번에는 **호출자 안**이 아니라 **호출 순서**다

`check` 는 빌린 증거(`function-logic-reference.txt`)를 `resolve_landing` **뒤에서**
푼다. 그래서 빌린 change 는 착지 판정 시점의 `analysis` 가 지역 디렉터리(번들 0)다.
`resolve_landing` 만 고쳤으면 새 거절이 **a073 의 유일한 수리 경로를 죽인다** —
a073 은 오늘 `AST source hash is stale` 로 빨갛고, 그 빨강을 푸는 것이 착지 선언이다.
변이 M2(순서 되돌리기)가 그 죽음을 시험 2건으로 재현한다.

## 조정된 구현 경계

바꾸는 것: (1) `resolve_landing` 의 고정 판정에 "고정한 번들 수" 하나, (2) `check` 의
빌린 증거 해소를 착지 판정 앞으로. 바꾸지 않는 것: 요구 함수 집합의 계산, base 규칙,
번들 판정(`validate_target`), 기존 거절 메시지 전부, 기록이 없는 change 의 판정.
Go 파일 변경 0줄.

---

# task 3.3 — 실패 출력이 비교 창을 말하지 않는다 (2026-09-10)

| 주장 | CodeGraph / grep | AST 열거 | 조정 |
|---|---|---|---|
| 이 문구를 읽는 도구·CI 가 있나 | 저장소 전수 grep → **없다**. 인용은 전부 산문이고 그것들이 인용하는 것은 **성공** 줄이다 | — | 성공 줄은 손대지 않는다. 실패 줄만 바꾼다 |
| 왜 실패 때는 창이 안 찍히나 | — | `main` `B1 L824 if errors:` → `L827 return 1` 이 `B4 L835 if landing:` 을 건너뛴다 | 기제 확인. 3.2.4 와 **같은 모양의 early return** |
| task 1.9("항상 출력한다")는 참인가 | — | 아니다 — `B4` 는 실패 경로에서 도달 불가 | 기록과 코드가 갈려 있었다. 이 태스크가 코드를 기록 쪽으로 올린다 |
| 실측 | — | a074 324줄 중 창 줄 **0** · a076 은 **21,838자 한 줄**에 창 줄 **0** | 이름만 있고 이유가 없다 |
| 편집이 경로를 더하나 | — | 반환 **2 → 2**, `print` 줄 `[850,856,864,867,869]` 이 `return 1`(865) 을 사이에 두고 앞뒤 | 창 줄 둘이 실패 반환보다 **앞**이라 두 경로 모두에서 찍힌다 |

## 이 태스크의 불변식은 "판정 불변"이다

3.2.3.1 은 출력 차이 0 을 쟀지만, 3.3 은 **출력이 반드시 달라진다**. 그래서 잴 것이
다르다: 126개 id 의 **rc 가 하나도 안 바뀌고**, 창 줄과 메시지 머리를 정규화한 뒤
나머지(요구된 함수 이름·다른 오류 줄)가 **글자 그대로 같아야** 한다. 이것을 안 가르면
"설명을 늘렸다"와 "판정을 바꿨다"가 같은 초록으로 보인다.

## 조정된 구현 경계

바꾸는 것: (1) `main` 이 창 줄을 **실패 반환 앞에서** 찍는다, (2) `missing Function
Logic Map for …` 메시지가 개수와 창을 담는다. 바꾸지 않는 것: 판정 전부, 성공 줄
문구, `missing evidence for modified function …` 줄(316번 반복될 자리라 접미사를
붙이지 않는다 — 창 줄이 한 번 말한다). Go 파일 변경 0줄.

---

# task 1.8 — 빌린 증거의 착지 공유 규칙 (2026-09-10)

| 주장 | CodeGraph / 전수 열거 | 실측 | 조정 |
|---|---|---|---|
| 창의 **시작**에는 공유 규칙이 있다 | AST: `check` `B16 L772 if referenced_base != base:` → `L773 return` | — | 있다. 끝에는 없다 — 이것이 1.8 이 가리키는 비대칭 |
| 빌리는 change 가 저장소에 몇 개인가 | `find openspec -name function-logic-reference.txt` → **1건** (a073 → a072) | — | 새 판정의 blast radius 는 1 id 다 |
| 착지를 선언한 change 는 몇 개인가 | `find openspec -name landed-commit.txt` → **1건** (a099) | — | a099 는 빌리지 않는다. 두 집합이 **안 겹친다** |
| 빌린 번들이 착지를 이미 고정하지 않나 | — | a072 의 `revision: current` 번들 **99개 파일**을 동시에 고정하는 커밋은 base..HEAD **326개 중 2개** (`171adda8`·`fb135d85`) | 고정은 구간을 남긴다. 그 구간 **안에서는** 저자가 고른다 |
| 그 자유가 요구 집합을 가르나 | — | 두 커밋의 required 가 **둘 다 147** | 이 저장소에서 측정된 크기는 **0**. 원리로는 열려 있고 실물로는 안 갈린다 |
| a073 의 수리 경로는 실재하나 | — | 오늘 오류 **336개**(240 미덮음 + 61 stale + 35 hash). 양쪽이 `171adda8` 을 선언한 상태를 읽는 자리에 주입해 `check` 를 끝까지 돌리면 **오류 0개** | 실재한다. 새 규칙은 그것을 죽이지 않고 **순서**를 만든다(a072 가 먼저 적는다) |
| `resolve_landing` 을 부르는 자리 | CodeGraph: 호출자 `check` **1건** / AST: 호출 자리 **1곳**(`L803`) | — | 자리 단위로 확인 — [[caller-count-is-not-fix-site-count]] |
| 새 판정이 옛 판정을 가리나 | AST: 새 `B19 L797` 이 `L803 resolve_landing` **앞** | 변이 M4(해시 대조 삭제) → **2건** FAIL | 가리지 않는다. 그 2건 중 하나가 빌린 증거의 위조 거절이다 |

## CodeGraphContext 가 답하지 못한 것

`borrowed function-logic evidence reference shares the comparison base and landing
point` 로 물으면 무관한 Go 심볼(`officialfx.Evidence` · `candidate.Baseline` ·
`riskbucket.PriceEvidence`)이 돌아온다. 대상이 Python 도구이고 의미 색인은 Go 쪽에
서 있다 — 3.2.3.1·3.3 과 같은 결과다. GBrain 은 이번에도 `CONNECTION_CLOSED` 로
붙지 않는다. 그래서 이 태스크의 근거는 전부 **전수 열거와 실측**이다.

## 결정과 그 근거의 크기

규칙: **빌린 증거를 쓰는 change 의 착지는 빌려주는 change 의 것과 정확히 같아야 하고,
한쪽만 선언한 상태는 통과하지 않는다.**

근거는 자유의 **크기**가 아니라 **누가 고르는가**다. 선언 파일을 고른 근거는
`analysis/landing-point.md` 의 한 문장뿐이다 — "번들이 고정하므로 저자가 고를 수
없다". 빌리는 change 에서는 그 번들이 남의 것이라 그 문장이 끝까지 참이 되지 않는다.
착지는 그것을 고정하는 증거가 사는 자리에 선언하고 빌리는 쪽은 값을 복사한다 —
[[two-judgements-cover-for-each-other]] 의 "규칙 하나는 상태가 사는 자리 하나에".

## 이 규칙이 닫지 **않는** 것

빌리는 change 의 작업이 공유된 착지 **뒤에** 착지하면 그 작업은 요구 집합 밖이고
어떤 판정도 못 본다. 빌리는 change 는 정의상 자기 번들이 0 이라 그것을 고정할 증거를
소유하지 않는다 — 공유 규칙이든 오늘의 고정 규칙이든 이 잔여는 같고, 새 규칙이
만드는 것이 아니다. §5 잔여 5.5.

## 조정된 구현 경계

바꾸는 것: (1) 선언을 읽는 자리를 `_declared_landing` 으로 뽑는다(`None`=파일 없음,
`""`=빈 선언을 가른다), (2) `check` 의 빌린 증거 블록에 창의 끝 대조를 하나 더한다.
바꾸지 않는 것: 요구 함수 집합의 계산, base 규칙, 고정 판정, `validate_target`,
기존 거절 메시지 전부, 빌리지 않는 change 의 판정 전부. Go 파일 변경 0줄.

---

# task 1.12 — 이관 예외의 창 끝은 감사된 값이다 (2026-09-11)

| 주장 | CodeGraph / 전수 열거 | 실측 | 조정 |
|---|---|---|---|
| 착지 기록을 record 키 집합에 넣어야 하나 | `execution_baseline.py:410` 의 `required` 는 **닫힌** 14개 키이고 그중 `source_commit` 이 이미 창의 끝이다 | a063 실물 record: 14 키, `source_commit = c727ad12` | **넣지 않는다.** 같은 값의 두 번째 이름은 재진술이다 |
| 감사된 창의 끝을 누가 읽나 | `validate` 는 `{"effective_base","source","ledger"}` 를 돌려주는데 `adoption["source"]` 를 읽는 곳이 저장소에 **0곳** | — | `resolve_base` 가 문맥에 적고 `check` 가 쓴다 |
| 오늘 대상이 워킹트리인데 왜 답이 맞나 | `validate` 의 **다른** 판정 — source→head 에 `openspec/`·`docs/pm/` 밖 파일이 있으면 `source-to-evidence drift` | 감사된 source 면 required **9**, 워킹트리면 **36** (`c727ad12..HEAD` = 커밋 32 · 파일 579 · **.go 24**) | 맞는 답을 우연으로 얻고 있었다 |
| 3.3 의 안내가 이관과 맞나 | — | 이관 픽스처 실행: `… record \`landed-commit.txt\` to narrow it …` 을 **찍는다** | 감사에 없는 손잡이를 만들라고 시킨다 — 모순이므로 이관 경로에선 안 찍는다 |
| 착지 기록이 이관 감사를 통과하나 | `openspec/` 아래 → drift 통과 · 추적 파일 → untracked 감사 밖 · 닫힌 키 집합에 없음 | 픽스처로 재현: 기록을 넣어도 오늘 rc 는 안 바뀐다 | 거절한다 |
| `resolve_base` 호출 자리 | CodeGraph 호출자 `check` 1건 / AST 호출 자리 `L768`·`L782` **둘** | — | 이 편집이 닿는 것은 `L768`(게이트 대상) 하나 — [[caller-count-is-not-fix-site-count]] |

## 곁가지로 나온 결함 — 3.2.3.1 의 고정 순회가 절대경로를 못 읽는다

이관 픽스처의 번들은 `file` 을 **절대경로**로 적는다. `validate_target` 은
`normalized_source` 로 정규화하는데 3.2.3.1 이 넣은 고정 순회는 `str(value["file"])`
을 그대로 `git show <sha>:<path>` 에 넘긴다. 절대경로면 언제나 실패하므로
**정상 입력이 위조로 몰린다**(픽스처에서 실제로 그 사유로 빨갛게 나왔다).

저장소 전수: 번들 `file` 필드 **상대 3048 · 절대 0** — 실물 영향은 0 이다. 그래도
증상 우회가 아니라 근본 원인이므로 같은 커밋에서 고치고, 변이 N4 와
`test_an_absolute_bundle_path_still_pins_a_landing` 으로 묶었다. 계약은 안 바뀐다 —
spec 은 이미 "착지 지점에서 모든 `revision: current` 묶음의 source hash 가 일치해야
한다"고 적었고, 코드가 그 말을 못 지키고 있었다.

## CodeGraphContext / GBrain

GBrain 은 이번에도 `CONNECTION_CLOSED` 다. 이 태스크의 근거는 전부 전수 열거와
픽스처 실측이다.

## 조정된 구현 경계

바꾸는 것: (1) `resolve_base` 가 감사된 `source` 를 문맥에 적는다, (2) `check` 가
문맥을 **항상** 채우고 이관이면 그 값을 대상으로 쓰며 착지 기록을 거절한다,
(3) `_target_text` 가 무엇이 끝을 고정했는지 말한다, (4) 고정 순회가 번들 경로를
정규화한다. 바꾸지 않는 것: `execution_baseline.py` **한 줄도**, 이관이 아닌
change 의 판정과 출력 전부, 성공 줄 문구, 이관 성공 줄 문구. Go 파일 변경 0줄.

---

# task 4.1 — 아카이브된 a075·a076 회귀 (2026-09-11)

| 주장 | 전수 열거 | 실측 | 조정 |
|---|---|---|---|
| 아카이브 id 로 재검사가 안 된다 (4.1 의 선결 조건) | — | 두 change 모두 base `448dfeb1` 을 해소한다 | **이미 닫혔다** — 3.2.4(`52c761de`) |
| 착지를 주면 요구 집합이 빈다 (proposal) | — | a075: required 319 → **0** | 참 |
| 그러면 통과한다 | — | a075 오류 324 → **1** | 남은 1은 task 1.13 이 "a122 가 덮지 않는다"고 적은 a075 자신의 결함 |
| 다섯 전부에 대해 참인가 | 번들 수: a074 7 · a075 8 · **a076 0** · a077 7 · a079 4 | a076: 착지 선언이 `pinned by no` 로 **거절** | **넷에 대해 참이다.** 그 거절은 옳다 — 3.2.3.1 이 막는 위조가 정확히 그 모양이다 |
| 실패 출력이 정직한가 | — | a076 실행이 `working tree … required 0 function(s)` 를 찍었다 — **둘 다 거짓** | 창은 **잰 것만** 찍는다 |

## 측정 방법을 주입에서 실제 커밋으로 올렸다

앞선 태스크들(a073·a063)은 선언을 **읽는 자리**에 값을 주입해 쟀다. 4.1 은 회귀
픽스처를 만드는 태스크이므로 `git worktree add --detach` 로 격리 트리를 만들고
`landed-commit.txt` 를 **실제로 커밋**했다(`06b280e0`). 그래서 "HEAD 커밋에서 읽는다"
"40자리 hex 여야 한다"까지 진짜 경로를 탔다. 측정 뒤 worktree 는 제거했고 아카이브본
에는 아무것도 안 썼다 — 4.1 은 검증 태스크이고 아카이브본 수리는 별개의 판단이다.

## 못 잰 창을 찍는 결함

`main` 은 `effective_base` 만 있으면 창을 찍었다. 착지 해소가 실패하면 `landing` 과
`required_count` 는 **아예 안 채워지고** 기본값이 찍힌다 — `working tree … required 0`.
대상도 아니고 세지도 않은 값이다. 3.3 이 창을 찍게 만든 이유가 "이름만 있고 이유가
없다"였는데 지어낸 이유는 그보다 나쁘다 — [[generated-evidence-must-be-measured]].

변이 O1(조건 되돌리기)과 O2(창 줄 삭제)가 **서로 다른** 시험을 빨갛게 한다(1건 vs
4건). 조건과 줄이 각각 독립으로 묶였다는 뜻이다 — 하나만 묶으면 나머지 변이가 산다.

오늘 저장소에서 이 줄을 보는 change 는 **0건**이다(착지를 선언한 change 는 a099
하나이고 유효하다). 그래서 126건 A/B 에는 안 보이고, 유닛 시험 하나가 재는 전부다.

## 조정된 구현 경계

바꾸는 것: `main` 의 창 출력 조건 하나(`if base:` → `if base and "landing" in context:`).
바꾸지 않는 것: 판정 전부, 성공 줄, 이관 성공 줄, 창 줄의 문구 자체. Go 파일 변경 0줄.

# task 4.2 — a074 · a077 · a079 회귀 측정

**코드 변경 0.** 순수 측정 task 라 AST 산출물이 필요 없다. 조정할 것은 "무엇을
쟀는가"와 "그 수가 어디서 왔는가" 둘뿐이다.

| 주장 | 근거 | 도구 |
|---|---|---|
| 셋 다 base `448dfeb1` 을 공유한다 | `openspec/changes/<id>/base-commit.txt` 3개 동일 | `cat` |
| 착지 후보가 한 점 `840b3377` 로 모인다 | 번들 소스별 hash 일치 커밋 교집합 (a074 14 · a077 2 · a079 2) | 일회용 열거기 `find_landing.py` |
| 착지 없음 → required 319 · rc=1 | 실행 출력 3건 | `check_analysis.py --change` |
| 착지 있음 → required 0 · rc=0 · 오류 0 | 격리 worktree 에 **커밋된** 선언으로 실행 | `git worktree` + `check_analysis.py --root` |
| 오류가 두 종류다 (남의 함수 / 자기 증거 stale) | 출력 줄을 유형별로 셈 | `grep -c` |
| stale 이 부르는 파일 = 그 change 자기 번들 소스 | 출력 줄의 파일명과 `ast.json` 의 `file` 대조 | `grep` + `json` |
| base 와 착지 사이 Go 파일 변경 0 | `git diff --name-only <base> <landing> -- '*.go'` → 0 | `git` |
| 번들 소스 12/12 가 base 에서 이미 일치 | `git cat-file blob <base>:<file>` 의 sha256 vs `source_sha256` | 일회용 census |
| 활성 13건 중 8건이 같은 상태 | 활성 change 전수 census | 일회용 census |
| 구간 안에서 required 가 0→30 | a074 의 14 후보 각각에 대해 실행 | `check_analysis.py --root` × 14 |

## 손으로 읽지 않은 것

착지 후보 구간은 **열거로** 구했다. "재기준화 커밋이 착지일 것이다"는 손으로 읽으면
맞는 답이지만 **볼 곳을 고른 것**이고, 구간이 14 라는 사실은 그렇게는 안 나온다 —
[[generated-evidence-must-be-measured]]. 구간의 존재가 1.8 의 전제였는데 그때는
효과가 0 이었고(a072), 여기서 처음 결과를 바꾼다 —
[[universal-check-passes-on-an-empty-sample]].

## 통과를 증거로 쓰지 않았다

`required 0` + `rc=0` 은 "검사를 안 했다"와 구분되지 않는다. 변이 다섯(P1·P1'·P1''·
P2·P3)이 전부 빨갛게 만든 뒤에 근거로 썼다 — [[passing-test-is-not-evidence]].
P3 은 a076 과 **같은 문**(`pinned by no revision: current evidence`)으로 죽는다.

## CodeGraph / CodeGraphContext / GBrain

해당 없음 — Go 심볼을 안 만지고 Python 도구 코드를 안 바꾼다. GBrain 은 이 세션에서
`CONNECTION_CLOSED` 로 계속 붙지 않는다.

## 조정된 구현 경계

바꾸는 것: 문서 3개(`tasks.md` 4.2·5.3·5.7, `review.md`, 이 파일). 코드 0줄, Go 0줄,
a074 · a077 · a079 디렉터리 0줄(선언은 5.1 때문에 **하지 않는다**).

# task 4.3 — 스위트·정적 게이트 전체

**코드 변경 0.** 검증 task 라 AST 산출물이 필요 없다.

| 주장 | 근거 | 도구 |
|---|---|---|
| focused 75 OK, a122 가 31 추가 | 옛 파일과 `def test_` 집합 diff (44 → 75) | `git show` + `diff` |
| Go 스위트가 **지금** 초록 | `-count=1` 로 cached 줄 0 확인 | `rtk proxy go test -count=1 ./...` |
| 태그 테스트도 돈다 | `make lint` 가 태그 vet 을, `make test-seams` 가 100 패키지를 | `make` |
| 126 id 중 도구 붕괴 0 | 전수 실행의 rc 분포 | `check_analysis.py --change` × 126 |
| 한 줄로 끝나는 10건의 **사유** | 9건은 `base-commit.txt` 부재를 파일 존재로, 1건은 실행 출력으로 | `test -f` + 실행 |
| 옛 도구는 아카이브 id 에서 못 답한다 | `a2d11fb2` 판을 `tools/logic-map/` 안에서 실행 | `git show` + 실행 |
| a063 의 detached 요구는 a122 이전부터다 | 옛 도구도 같은 줄 | 옛 도구 실행 |

## 계측기가 눈멀었던 것을 잡았다

아카이브 A/B 첫 시도는 옛 도구를 scratchpad 에 복사해 돌렸고 셋 다
`ModuleNotFoundError: No module named 'role_check'` 로 죽었다. **옛 도구의 성질이
아니라 형제 모듈 import 실패**다. 그대로 적었으면 맞는 결론을 틀린 증거로 받치게 된다.
도구를 원래 디렉터리에 두고 **활성 id 양성 대조군**을 먼저 통과시킨 뒤 다시 쟀다 —
[[mutation-must-reach-the-thing-under-test]] 가 말하는 "켜면 YES" 대조군이 정확히
이 자리에서 필요했다.

## 캐시된 초록

`make test` 의 `(cached)` 는 Go 가 입력 키로 검증한 값이라 거짓은 아니지만, VERIFY 가
적어야 하는 것은 "지금 돈 값"이다 — [[missing-tool-reports-clean]].

## 못 한 것 (침묵하지 않는다)

1.12 이관 경로의 **실데이터 확인은 못 했다.** 세 지점(메인 브랜치 / a063 worktree /
HEAD 의 새 detached worktree)이 각각 다른 이유로 막히고, 마지막은 a063 의
`source_commit` 이 HEAD 의 조상이 아니어서다. 세 지점 모두 옛 도구와 출력이 동일해
a122 가 만든 차이는 없지만, 그것은 **차이 없음**이지 **동작 확인**이 아니다. 4.4 에
넘긴다.

## CodeGraph / CodeGraphContext / GBrain

해당 없음 — Go 심볼·도구 코드 변경 0. GBrain 은 이 세션 내내 `CONNECTION_CLOSED`
이고 `make sdd-sync` 도 `gbrain advisory index is missing or stale` 을 WARN 으로
남긴다(hard evidence 인 CodeGraph 는 일치).

## 조정된 구현 경계

바꾸는 것: 문서 3개(`tasks.md` 4.3, `review.md`, 이 파일). 코드 0줄, Go 0줄.

# task 4.4 — 독립 적대 리뷰

| 주장 | 증거 | 도구 |
|---|---|---|
| base 상태 증거 + `landing = base` 가 RED → GREEN 을 만든다 | 픽스처 4케이스(A 기준선 · B 정직증거+선언 거절 · C 위조 통과 · D 같은 위조 선언없이 RED) | `scratchpad/probe_base_landing.py`, 실제 Go 추출기 |
| `revision: base` relabel 이 거절을 통과로 바꾼다 | 2케이스 A/B, `Other()` 가 required 밖으로 | `scratchpad/probe_revision_base.py` |
| 아카이브 이름이 이관 판정을 깬다 | 활성 이름 vs 날짜 붙은 이름의 `AdoptionError` 문장 차이 | `scratchpad/probe_a063_archive.py` |
| a099 의 고정은 정직하다(창 밖 0) · 그래도 후보 21 | 번들 37 → 파일 12 전부 창 안 · base..HEAD 239 중 21 통과 | `scratchpad/a099_pin.py` |
| 가드 3개가 지워져도 스위트가 초록이다 | 뮤테이션 9회, 원복 sha256 동일성 확인 | `scratchpad/mutate.py` |
| 옛/새 도구 구분이 필요한 자리 | — | 4.3 이 이미 쟀다. 4.4 는 새로 안 쟀다 |

## 손으로 읽지 않은 것

가드의 유효성은 **뮤테이션으로** 쟀다. "읽어 보니 맞다"는 근거로 쓰지 않았다 —
M2·M3·M4 가 정확히 그 방식으로는 통과했을 자리다.

## 통과를 증거로 쓰지 않았다

세 위조 시나리오 전부 **대조군을 먼저 세웠다.** C 가 초록인 것만으로는 아무것도
증명하지 못한다 — 같은 거짓 증거가 선언 없이는 빨갛다는 D 가 있어야 "a122 가 바꿨다"가
성립한다. `AForgedLandingPointIsRefusedByName` 자신이 같은 규율을 문서로 적어 두었고
(`test_the_fixture_passes_before_any_landing_record` = 양성 대조군), 이 리뷰는 그 규율을
그 클래스의 **바늘**에 적용했을 때 셋이 무너지는 것을 발견했다.

## 남의 상태를 안 건드렸다

a063 검증은 `execution-baseline.json` 을 임시 디렉터리에 **복사**해서 이름만 재현했다.
a063 을 실제로 옮기거나 그 worktree 를 청소하지 않았다 — 4.3 에서와 같은 이유다.

## 못 한 것 (침묵하지 않는다)

- **1.12 이관 경로는 이번에도 실데이터로 못 돌렸다.** 4.3 이 적은 세 막힘이 그대로다.
  4.4 가 더한 것은 그 사각지대 **안에서** 결함 하나(6.2)를 찾아낸 것이고, 경로 자체를
  돌린 것은 아니다.
- **P0-1 의 수정안을 실측하지 않았다.** 6.1 이 제안하는 "창 안 파일로 고정을 한정"이
  기존 13개 change 를 어떻게 가르는지는 안 쟀다. 그 측정은 6.1 의 일이다.
- **생산 코드를 안 고쳤다.** 리뷰가 잰 HEAD 와 기록이 가리키는 HEAD 를 같게 두려는
  의도적 선택이고, a112 8.5 가 P0 셋을 8.8.x 로 연 선례를 따랐다.

## CodeGraph · GBrain

GBrain 은 이번 세션 내내 `CONNECTION_CLOSED` 다(advisory 라 판정에 안 쓴다).
CodeGraph 는 Python 을 인덱싱하지 않으므로 이 리뷰의 대상(`tools/logic-map/*.py`,
`tools/gate.sh`)에 해당 없음 — 함수 관계는 AST 열거(`analysis/python-function-logic/`)와
직접 읽기로 잡았다.

---

# task 6.1.1 — 측정

4.4 가 "못 한 것"으로 남긴 두 번째 항목(**P0-1 의 수정안을 13개 change 에 실측하지
않았다**)이 이 task 로 닫힌다. 네 축을 전부 걸었다.

| 주장 | 증거 | 도구 |
|---|---|---|
| 후보가 하나인 change 는 a066 뿐 (1/13) | 커밋 × 번들소스 blob 의 sha256 전수 | `git cat-file --batch-check` + `--batch` |
| 후보 수는 번들 수와 무관 (a112 132→39, a091 2→27) | 같은 표 | 같음 |
| base 자신이 후보인 change 8건 | `passing[0] == base` | 같음 |
| A축 필터가 그 8건에서 전부 0 | `git diff --name-only base cand` ∩ 번들 소스 | `git diff` |
| 13건 전부 번들이 자기 base 뒤에 커밋됐다 | `git log -1 -- <ast.json>` + `merge-base --is-ancestor` | `git log`·`merge-base` |
| a089·a095 는 B축에서 후보 0 | 위 두 측정의 교집합 | 같음 |
| a089 `outbox.go` · a095 `notifier.go` 가 **자기 번들 커밋에서 불일치** | `git show a30eb35ae:<path>` 의 sha256 | `git show` |
| C축 대가 (a074 0→35, a094 0→74 …) | 최저/최고 후보에서 `git diff --name-only` 의 `*.go` 수 | `git diff` |

## 지어내지 않은 것

- **규칙을 안 골랐다.** 네 축의 비용만 적었고 어느 것도 코드·spec 에 넣지 않았다.
  결정은 6.1.2 로 열려 있고 사람 몫이다.
- **숫자를 안 반올림했다.** 표의 모든 수는 스크립트 출력 그대로다.
- **A축의 0 은 측정이 아니라 증명이다.** 번들이 양 끝에서 맞으면 그 파일은 창 안에서
  안 바뀐 것이므로 필터는 반드시 0 이다. 측정은 그 증명이 13건에서 깨지지 않음을
  확인한 것뿐이다 — 8건의 0 을 "재 봤더니 우연히 0" 으로 읽으면 안 된다.

## 워킹트리를 안 읽었다

모든 해시는 커밋된 blob 에서 왔다. `resolve_landing` 이 워킹트리를 안 보는 것과 같은
조건으로 재야 결과가 그 함수의 판정과 같은 뜻을 갖는다. 번들 선별 조건도 그 순회와
글자 그대로 같게 썼다(`revision` 기본 `"current"`, `file`·`source_sha256` 둘 다 필요).

## 못 한 것 (침묵하지 않는다)

- **required 함수 수를 안 셌다.** 축 비교에 쓴 것은 창 안 **Go 파일 수**다. 실제 요구
  집합은 `check_analysis` 를 후보마다 돌려야 나오는데, 13건 × 후보 최대 59개라 이번에
  안 돌렸다. 파일 수 0 이면 요구도 0 이라는 관계만 쓴다(4.2 에서 실측된 관계다).
- **위조 재현을 다시 안 했다.** 4.4 의 세 재현을 그대로 근거로 쓴다. 이 task 는 그
  구멍의 **범위**를 잰 것이고 존재를 다시 증명한 것이 아니다.
- **a089·a095 의 옛 번들이 왜 그렇게 됐는지 저자에게 확인 못 했다.** 기전(생성 뒤 같은
  세션에서 Go 를 더 고치고 한 커밋에 담기)은 구조에서 유도한 것이고, 그 커밋이 실제로
  그렇게 만들어졌다는 증언은 없다. 6.5 에 그대로 적었다.

---

# task 6.1.2 — 구현

| 주장 | 증거 | 도구 |
|---|---|---|
| 위조 셋이 새 코드에서 전부 거절된다 | 4.4 가 쓴 스크립트 **그대로** 재실행 | `probe_base_landing.py` · `probe_revision_base.py` · (미끼는 같은 헬퍼로 새로 씀) |
| 대조군이 안 움직였다 (RED 는 RED, GREEN 은 GREEN) | 같은 실행의 A·B·D 케이스 | 같음 |
| 13건이 실제로 받을 값 | `compute_landing` 을 13건에 직접 호출 | `check_analysis` import |
| a074·a077·a079 → `840b3377`, 창 Go 0 | `git diff --name-only base landing` | `git diff` |
| a089·a095 는 거부된다 | 같은 호출의 사유 문자열 | 같음 |
| 기존 기록(a099)에 회귀 없음 | `check_analysis.py --change a099-…` rc=0 required 32 | 실행 |
| 하한 셋이 시험에 걸린다 | M1·M2·M3 변이, 원복 sha256 동일 | 직접 |
| 잔여(증거 먼저·작업 나중)가 열려 있다 | 픽스처 왕복 실측 | `probe_residual.py` |
| 분기 16 → 10, raise 6 → 8 | `enumerate.py` 를 HEAD·worktree 양쪽에 | AST 열거 |

## 순서를 어겼다 (침묵하지 않는다)

`.claude/CLAUDE.md` 는 "기존 함수 내부 로직을 바꾸면 Function Logic Map 을 **먼저**
만든다"고 적는다. 이번에 `resolve_landing` 과 `main` 의 내부를 바꾸면서 **열거를 편집
뒤에 뽑았다.** 산출물(`ast.before-6.1.2.json` = HEAD 커밋의 blob, `ast.worktree.json`)
은 둘 다 같은 열거기가 기계로 만든 것이라 표 자체는 손으로 읽지 않았지만, 순서는
어겼다. 편집 전 판본을 blob 에서 뽑을 수 있어서 결과가 같을 뿐이고, 규칙이 막으려는
"반증 산출물이 그것을 필요로 하는 문서보다 나중에 나오는" 상태가 실제로 있었다.
[[flm-before-claiming-not-before-editing]]

## 지어내지 않은 것

- **"셋 다 막힌다"를 미리 안 썼다.** 6.1.2.4 는 "무엇이 빨개지고 무엇이 안 빨개지는지
  재서 적는다, 셋 다 막힌다고 미리 쓰지 않는다"로 열려 있었고, 3번(`revision: base`
  relabel)이 막히는지는 돌려 보고 알았다.
- **잔여를 덮지 않았다.** 하한이 닫지 못하는 순서를 찾아서 재고 6.6 으로 열었다.
  "위조가 막혔다"로 끝내면 그 다음 사람이 같은 구멍을 다시 찾는다.
- **규칙을 안 지어냈다.** 6.6 의 후보 방향이 1.12 의 "신원으로 판정 금지"와 부딪히는
  지점을 안 가른 채로는 아무것도 안 넣었다.

## 못 한 것

- **`compute_landing` 의 갈래 둘, `record_landing` 의 갈래 둘에 픽스처가 없다.**
  (바닥 없음 · 맞는 커밋 없음 · 빌린 증거 · 이관) 앞의 둘은 실물 a089·a095·a063 으로만
  쟀다. branch-test-map 에 `no` 로 적었다 — 커버리지를 주장하지 않는다.
- **`gate.sh` 를 안 건드렸다.** 6.1.2.6 으로 열었다.
- **독립 적대 리뷰(gstack)를 이 로트에 아직 안 돌렸다.** 4.4 가 리뷰한 HEAD 이후로
  코드가 바뀌었으므로 4.4 의 리뷰는 이 변경을 안 봤다.

---

# task 6.2 · 6.2.1 — 이관 경로의 두 P0

| 주장 | 증거 | 도구 |
|---|---|---|
| 아카이브 뒤 막는 자리가 넷이고 다섯째는 없다 | 메모리 사본에서 가드를 하나씩 풀어 다음 실패를 기록, 넷을 풀면 통과 | `62_chain.py` (생산 코드 무변경) |
| 이관 기록은 옮기기 전 경로를 적는다 | a063 실물 `execution-baseline.json` 의 세 경로 · `draft` 의 `relative_to(root)` | `json.tool` · 코드 |
| `_declared_landing` 의 try 밖 호출은 둘이다 | 호출 자리 × 둘러싼 handler | AST(`ast.walk` + 부모 사전) |
| 편집이 계획대로 갈래를 바꿨다 | before/after 열거 diff | `enumerate.py` — HEAD blob · worktree |
| 수리가 복사를 막던 둘째 벽을 없앴다 | 편집 전 사본(이름만 지움) = 접두사에서 막힘 · 편집 후 사본(신원 지움) = 통과 | `62_copywall.py` |
| 각 편집 자리가 시험에 걸린다 | 변이 9, 한 변이에 한 시험, 원복 sha256 | `62_mutations.py` |
| 활성 a063 판정은 편집 전과 같다 | 기존 이관 시험 초록 · 실물 a063 CLI 출력 동일 | unittest · CLI |
| 실데이터 확인은 **아니다** | 실물 a063 은 `adoption requires detached HEAD` 앞을 못 지남 | CLI |

## 순서

`ast.before-6.2.json` 다섯(HEAD blob) → FLM "편집 계획" 다섯 곳 → RED 시험 다섯 → GREEN →
`ast.after-6.2.json` 대조 → 변이 → 벽 실측. 6.1.2 가 어긴 순서를 이번에는 지켰다. 계획의 한
문장이 틀렸던 것은 편집 전에 호출자를 세다가 찾았고, 틀렸다고 적은 채 고쳤다.

## 못 한 것

- 실물 a063 을 아카이브해 재판정하지 않았다 — 남의 change 상태이고, 이 HEAD 에서는 그 앞
  단계에서 막힌다.
- gstack 독립 리뷰는 아직이다.
