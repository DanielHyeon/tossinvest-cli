# Review: a103-the-rollback-pin-is-made-not-remembered

## Function Logic Map: not-applicable

이 change는 기존 Go 함수의 내부를 편집하지 않는다. 산출물은 새 Python 파일 둘
(`tools/deploy/image_pin.py`, `tools/deploy/test_image_pin.py`), Makefile의 새 타겟
하나와 기존 두 타겟의 목록 추가(`sdd-test`, `compileall`), 그리고 문서다.
Go 함수의 분기·early return·side effect를 근거로 삼는 주장도 하지 않는다.
`check_analysis.py`가 요구하는 함수 집합은 비어 있다.

인용한 Go 코드는 `cmd/tossctl/soakautostart.go:78-81` 하나이고, 그것은 그 함수의
**주석**(왜 판단을 인자 받는 함수로 뺐는지)이지 분기 주장이 아니다.

## 1라운드 — 독립 리뷰 (2026-08-13)

`/review`가 main 대비 diff(코드 225줄)에 대해 적대적 리뷰를 돌렸다. **치명 6건 ·
정보 6건.** 리뷰어는 PATH에 스텁 `docker`를 세워 `make image`를 **실제로 실행해서**
지적을 검증했다. 실제 이미지·컨테이너는 건드리지 않았다(리뷰 전후 `tossos:local`
= `66553dba92d8`, 컨테이너 31분 가동 유지로 확인).

### 받아들인 치명 4건 — 전부 초판의 결함이다

| # | 지적 | 내가 확인한 방법 |
|---|---|---|
| C1 | **빌드 후 tag 실패가 성공으로 보고된다.** 110-114줄이 `;`로 이어진 한 셸 줄이고 마지막이 `echo`라 make는 rc=0을 본다. tag가 죽어도 "박았다"를 찍는다 | `sh -c 'id=$(false); docker tag "$id" x; echo 박았다'` → **rc=0** |
| C2 | **가드가 docker 실패를 "잃을 게 없다"로 접는다.** `2>/dev/null \|\| true`가 데몬 정지·미설치·소켓 권한을 전부 빈 목록으로 만든다 | `guard --tags ""` → rc=0, "직전 이미지가 없다" |
| C3 | **같은 커밋 재빌드가 핀을 옮긴다.** `docker tag`는 이름을 만드는 게 아니라 옮긴다. 가드가 방금 "안전하다"고 인정한 그 이름을 빼앗아 직전 이미지를 무명으로 만든다. **이 change가 새로 만든 유실 경로다** | 스텁 시나리오 D |
| C4 | **CHANGE가 셸을 빠져나간다.** `"$(CHANGE)"`가 recipe에 끼워지므로 따옴표가 든 값은 image_pin에 닿기 전에 이미 탈출한다 — 그 뒤 검증은 아무것도 막지 못한다 | 리뷰어가 `/tmp/PWNED_BY_CHANGE` 생성으로 실증 |

### 받아들인 치명 2건 — 내 논거 자체를 친 지적

| # | 지적 | 확인 |
|---|---|---|
| C5 | **`tools/deploy` 테스트가 아무 데서도 안 돈다.** `sdd-test`는 scripts·logic-map·sdd·sdd-history·pm 다섯만 돌고, `compileall`도 같다. 즉 *"판단을 테스트 닿는 자리에 뒀다"*는 이 change의 논거가 성립하지 않았다 — 판단은 옮겼는데 거기도 자동으로 닿는 것이 없었다 | `sed -n '/^sdd-test:/,/^$/p' Makefile` — 목록에 없음 |
| C6 | **문서가 금지한 명령을 여전히 가르친다.** operations.md:75의 설치 절에 `docker compose build`가 그대로 있고, 그게 복붙되는 블록이다 | 해당 줄 확인 |

### 정보 6건 중 반영

- **dangling 판정이 docker가 내지 않는 모양을 검사했다.** `{{join .RepoTags ","}}`는
  태그 없는 이미지에 `<none>:<none>`이 아니라 **빈 문자열**을 낸다(`json=[]`로 실측).
  초판 테스트 `test_a_dangling_image_is_already_orphaned`는 손으로 친 입력에서만
  통과했고, 실제 경로에서는 "이미지 없음 → 안전" 가지로 빠졌다. **통과가 증거가
  아니었던 사례다.**
- docker 태그 128자 상한 미검증 → `pin_name`에 추가.

### 반영하지 않은 것

- **`.Id` vs `.RepoDigests`(operations.md:284)** — 지적은 맞다. 다만 그 줄은 release
  digest-pinned 경로의 것이고 이 change의 범위 밖이라 별도로 다룬다. proposal의
  "범위 밖"에 이미 선언돼 있다.
- **동시 빌드 lock** — 실재하는 경합이지만 이 저장소는 배포자가 한 명이고, 도입하면
  `make image`가 락 관리 책임을 지게 된다. `limits`에 기록하고 넘긴다.
- **Dockerfile revision label** — 거부 메시지가 직전 change/commit을 사람 기억에
  의존하는 문제. 옳은 지적이고 이 change의 정신과도 맞지만 Dockerfile 변경은
  별도 범위다. `limits`에 기록했다.

## 2라운드 — 고친 뒤 실측

### 단위

`tools/deploy/test_image_pin.py` **18/18**. 초판 10건에서 8건을 더했고, 새 8건은
전부 위 지적에 대응한다(UNKNOWN 거부 2, 알 수 없는 상태 fail-closed 1, 빈 RepoTags
1, 핀 충돌 2, 길이 상한 2). `make sdd-test`·`compileall`에 `tools/deploy`를 넣어
이제 자동으로 돈다 — C5의 조건.

### 종단 (스텁 docker, 실제 데몬 미접촉)

| 시나리오 | 기대 | 실측 |
|---|---|---|
| A 데몬 정지 | 거부 | rc=2 "docker 를 볼 수 없다" ✓ |
| B 직전 이미지에 핀 있음 | 통과 + 태그 검증 | rc=0 "박았다" ✓ |
| C 직전 이미지가 local 뿐 | 거부 | rc=2 ✓ |
| D 핀 이름 선점 | 거부 | rc=2 ✓ (C3) |
| E 빌드 후 tag 실패 | **실패** | rc=2, "박았다" 안 찍음 ✓ (C1 — 초판은 rc=0) |
| F 첫 빌드 | 통과 | rc=0 ✓ |
| G CHANGE에 셸 탈출 | 거부 + 파일 미생성 | rc=2, `/tmp/PWNED2` 없음 ✓ (C4) |

F는 1차 측정에서 실패로 나왔는데 **스텁 결함**이었다 — 빌드 후에도 `tossos:local`이
없다고 답하게 짜여 있었다. 스텁을 고쳐 재측정했다. 코드 결함이 아니다.

### 저장소 검사

`make sdd-test` OK(18/18 포함) · `make lint` rc=0 · 실제 이미지 무변(`66553dba92d8`에
`tossos:local`·`tossos:a098-ee8eca83`).

## 남은 위험 — story의 `limits`와 동일

1. 손으로 `docker compose build`를 치는 경로는 여전히 열려 있다. 문서에서 이유를
   없앴을 뿐이다.
2. 거부 메시지가 직전 change/commit을 사람 기억에서 요구한다. 이미지에 그 값이
   없다.
3. 핀은 빌드 후 mutable한 `tossos:local`을 다시 읽어 박는다. 동시 빌드가 그 사이에
   태그를 옮기면 엉뚱한 이미지에 박힌다. 락 없음.

## 판정

**초판은 병합하면 안 되는 물건이었다.** 치명 6건 중 4건이 이 change가 막으려던 바로
그 유실을 다른 형태로 다시 열었고(C1·C2·C3), 하나는 새 유실 경로를 만들었으며(C3),
하나는 change의 존재 이유를 성립시키지 못했다(C5). 리뷰가 실행으로 검증했기 때문에
전부 논쟁 없이 확정됐다.

고친 뒤 7개 종단 시나리오와 18개 단위가 전부 기대대로다.
