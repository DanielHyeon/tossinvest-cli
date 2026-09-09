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
