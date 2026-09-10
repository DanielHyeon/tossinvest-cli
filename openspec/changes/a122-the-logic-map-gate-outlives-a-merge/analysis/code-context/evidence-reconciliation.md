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
