# Supporting context — gate.sh 아카이브 인식 (task 1.11)

Date: 2026-09-10

## 같은 결함의 선례

이 저장소는 **같은 모양의 결함을 이미 두 번 고쳤다.** 세 번째가 이번 것이다.

| 자리 | 무엇이 아카이브를 못 봤나 | 고친 것 |
|---|---|---|
| `resolve_referenced_change` (`check_analysis.py:231`) | 빌려주는 change 가 먼저 아카이브되면 증거 포인터가 끊긴다 | `f6965ebb` |
| `check` 의 `change_dir` (`check_analysis.py:590`) | 아카이브된 id 로는 5단계 재검사가 안 된다 | a122 task 3.2.2 |
| `tools/gate.sh` 의 `CHANGE_DIR`(`:116`) · `PAIR_DIR`(`:184`) | 짝이 먼저 아카이브되면 3단계에서 죽는다 | **이번 (task 1.11)** |

앞의 둘은 Python 이라 한 함수를 공유해 고쳤다. 셋째는 shell 이라 그 함수를 부를 수
없다 — 문법만 같고 사본이 하나 더 생긴다. 그 사실이 이번 결정의 핵심이다.

## 실물 증거 (2026-09-09 측정)

`a099-a-claim-excludes-the-second-sender` 는 `deploy-pair.txt` 로
`a098-nobody-sends-what-the-outbox-keeps` 를 선언한다. a098 은 2026-08-29 에
`archive/2026-08-29-a098-nobody-sends-what-the-outbox-keeps` 로 옮겨졌다.

```
make gate CHANGE=a099-a-claim-excludes-the-second-sender
==> 3/11 짝 change 확인 (deploy-pair.txt)
짝: a098-nobody-sends-what-the-outbox-keeps
GATE FAIL: 짝 change 의 tasks.md 가 없습니다:
  openspec/changes/a098-nobody-sends-what-the-outbox-keeps/tasks.md
```

같은 change 의 5단계는 착지 지점을 기록하니 `required 32 / errors 0` 으로 통과한다.
즉 a122 가 고친 5단계에 **3단계가 도달을 막고 있다.**

활성 change 전체의 `deploy-pair.txt` 를 훑어 이 상태에 걸린 것은 a099 하나다.

## TossOS 고유로 남겨야 하는 것

- gate 3단계의 나머지 검사(자기 자신 선언 거부, 면제 줄 정확히 하나, 배포 단위 구성원
  집합 일치)는 아카이브 인식과 무관하며 그대로 유지한다.
- id 형태 검사(`*/*`, 공백, `.`, `..`)는 경로 조작 방어이므로 해소 **전에** 남는다.
