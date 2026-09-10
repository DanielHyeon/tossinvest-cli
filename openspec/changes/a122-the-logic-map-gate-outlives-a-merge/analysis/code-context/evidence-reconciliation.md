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
