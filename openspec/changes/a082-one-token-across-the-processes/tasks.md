# a082 · Tasks

## 1. 근거 고정 (편집 전)

- [x] 1.1 `saveCache`를 부르는 곳이 `exchange()` 하나뿐이라는 것을 현재 HEAD에서
      확인해 기록한다 (24시간 토큰이 1분에 7번 다시 쓰이는 유일한 설명이 401
      재시도라는 논거의 근거).
- [x] 1.2 `token()`이 메모리 캐시를 먼저 보고 디스크를 다시 읽지 않는 것을
      확인한다 (token.go:66-75).
- [x] 1.3 `refresh()`가 무조건 `exchange()`하는 것과, 그것을 부르는 곳이
      `send()`의 401 갈래 하나뿐인 것을 확인한다.
- [x] 1.4 운영 컨테이너에서 `openapi-token.json`을 공유하는 프로세스를 열거한다
      (읽기 전용: `docker exec … ps`). 셋: console, engine, httpapi.
- [x] 1.5 `errors.Is(err, official.ErrAuth)`에 기대는 곳을 전부 열거하고 문자열
      비교로 판정하는 곳이 없음을 확인한다.
- [x] 1.6 `token()`·`refresh()`·`classifyStatus`의 Function Logic Map과 Branch
      Test Map을 **편집 전에** 작성한다. 인증 경로는 High-risk이므로 면제 불가.
      `check_analysis.py` 통과.
- [x] 1.7 기존 요구사항과의 충돌을 확인한다 — `persistent credential and token
      lifecycle`의 "existing official client renewal behavior" 문장.

## 2. RED — 지금은 재현되지 않는 것

- [x] 2.1 manager 둘이 한 `cacheFile`을 공유하는 하네스를 만든다. 프로세스 둘을
      흉내내기에 충분한 이유는 design D6.
- [x] 2.2 브로커가 "마지막에 발급한 토큰만 유효"를 흉내내는 httptest 서버를
      만든다 — 교환할 때마다 새 토큰을 주고 옛 토큰에 401을 준다.
- [x] 2.3 그 하네스에서 두 manager가 번갈아 요청하면 **교환이 무한히 늘어나는
      것**을 관측한다 (수정 전 RED).
- [x] 2.4 그 테스트가 무의미하게 통과할 수 없음을 자체 검사한다 — 요청 횟수가
      허용 교환 수보다 많지 않으면 `t.Fatal`.

## 3. GREEN — 채택-우선

- [x] 3.1 `refresh()`가 교환 전에 디스크를 읽고, 유효하며 `m.cache`와 **다른**
      토큰이 있으면 채택한다 (design D1).
- [x] 3.2 없거나 같으면 교환한다 — 기존 동작.
- [x] 3.3 2.3이 GREEN이 된다: 교환 횟수가 토큰 수명당 1회로 수렴한다.
- [x] ~~3.4 `token()`이 캐시 파일 mtime을 보고 디스크를 다시 읽는다.~~
      **철회 → design D2.** 적대적 리뷰 P0-1·P1-3·P2-7. `token()`은 base 그대로다.
- [x] ~~3.5 3.4가 없으면 회전마다 401을 한 번 먹는 것을 고정한다.~~ **철회.**
      그 401은 `send()`의 재시도가 삼키므로 하루 3회의 낭비이고, 대신 산 것이
      요청마다 교환할 수 있는 갈래였다 (issues I8).
- [x] 3.6 `refresh`가 `refused`를 인자로 받아 추론을 없앤다 (design D1, 리뷰 P1-5).
- [x] 3.7 채택한 토큰이 또 거부당하면 `send()`가 발급해 한 번 더 시도한다
      (design D5, 리뷰 P0-2). 상한 2회.
- [x] 3.8 `saveCache`를 임시 파일 + rename으로 바꾼다 (design D6, 리뷰 P1-4).

## 4. 진단 가능성

- [x] 4.1 `classifyStatus`가 인증 거부에 상태 코드를 실어 보낸다 (design D4).
- [x] 4.2 응답 본문은 싣지 않는다 — 테스트로 고정한다.
- [x] 4.3 `errors.Is(err, ErrAuth)`와 `errors.Is(err, ErrIPNotAllowed)`가 그대로
      성립한다. 분류 규칙(401/403 구분, `ip` 낱말 검사)은 한 글자도 안 바뀐다.
- [x] 4.4 `ShouldFallback`의 판정이 달라지지 않는다.

## 5. 안 바뀌는 것

- [x] 5.1 단일 프로세스 의미론이 글자 그대로 보존된다 — `TestTokenRefresh`가
      **손대지 않고** 통과한다 (design D5).
- [x] 5.2 기존 `internal/official` 테스트 전부가 손대지 않고 통과한다. 갱신이
      필요한 테스트가 나오면 그 이유를 `issues.md`에 남긴다.
- [x] 5.3 프로세스 간 블로킹 대기를 도입하지 않는다 (design D3, 스펙 SHALL NOT).
- [x] 5.4 `ClassAuthFatal`의 Gate 차단 정책을 바꾸지 않는다.

## 6. VERIFY

- [x] 6.1 변이 검증: 채택 갈래를 지우면 2.3이 RED가 되는지 확인하고 되돌린다.
- [x] 6.2 변이 검증: mtime 확인을 지우면 3.5가 RED가 되는지 확인하고 되돌린다.
- [x] 6.3 변이 검증: 상태 코드를 다시 버리면 4.1이 RED가 되는지 확인하고
      되돌린다.
- [x] 6.4 변이 검증: 채택 조건을 "다르면"에서 "항상"으로 바꾸면 5.1이 RED가
      되는지 확인하고 되돌린다 (좁음의 증거).
- [x] 6.4b 변이 검증 추가: `send` 상한 2→1(M5), `saveCache`를 plain
      write로(M6), 형제 전용 갈래 제거(M2 — **통과했으므로 그 갈래를 삭제**).
- [x] 6.4c 독립 리뷰 2가 살려낸 변이를 닫는다: 만료 토큰 채택(N5), 채택 조건을
      토큰이 아닌 만료로 키잉(N6), 소비자 둘의 감싼 sentinel 미검증(P1-6 E1·E2).
      네 변이 전부 RED 확인.
- [x] 6.5 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [ ] 6.6 `make gate CHANGE=a082-one-token-across-the-processes`.
- [ ] 6.7 사람 승인 후 배포 → 컨테이너 실측: `openapi-token.json`의 mtime이
      관측 창 안에 **0회** 바뀌는지 (지금 1분 7회). 그리고 `reconcile.mismatch`가
      더 나오는지.

## 7. 리뷰와 기록

- [x] 7.1 proposal-freeze 리뷰.
- [x] 7.2 Requirement 변경 리뷰 — `persistent credential and token lifecycle`을
      MODIFY하므로 WORKFLOW.md §142가 요구한다.
- [x] 7.3 발견 사항을 `issues.md`에 남긴다 (403 재시도 부재, 배포 후에야 답이
      나오는 질문 포함).
- [x] 7.4 PM story/tracker 동기화.
- [x] 7.5 별도 컨텍스트의 독립 리뷰 — 인증 경로는 High-risk이므로 적대적 Eng
      관점을 포함한다 (WORKFLOW.md 위험 등급 가중).
