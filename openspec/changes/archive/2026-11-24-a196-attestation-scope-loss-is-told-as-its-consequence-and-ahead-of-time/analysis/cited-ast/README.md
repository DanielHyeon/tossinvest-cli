# 인용 함수 AST (a196 1단계)

proposal·design 이 분기·early return·side effect 를 근거로 쓴 함수들의 `tools/logic-map` 산출물이다. 1단계는 Go 를 고치지 않으므로 전부 **인용만** 하는
base 판본이다. 2단계에서 편집하는 기존 함수는 `../function-logic/<pkg>--<fn>/` 번들을 따로 가진다.

- 기준: worktree `~/tossos-work/a196-impl`, HEAD `064f59fd9`(= `base-commit.txt`), 2026-11-23. 작업 트리의 Go 변경 0.
- 명령: `GOFLAGS=-trimpath go run ./tools/logic-map --file <파일> --func <함수>` → `<패키지-경로>--<함수 소문자>.ast.json`. **22개 전부 rc 0**
  (생성 루프는 실패 시 `FAIL` 을 찍고 파일을 옮기지 않게 짰고, 찍힌 줄 없음). 조사 때 사본(`inv-E/`)은 참고만 하고 쓰지 않았다.
- 요약: `summary.txt`(생성기 `astsum.py.txt` — a193 의 것을 그대로 복사). 좌표: `B<n>(kind@line)`·`call@line`·`return@line`.

## proposal 「코드 근거」 표 대조 (여섯 줄)

| # | proposal 주장 | AST·HEAD | 판정 |
|---|---|---|---|
| 1 | 타입(ntfy 첫 태그) `broker.credential_expiring`, 범위 상실 제목·본문(원인 「…쓰였을 수 있다」, 결과 조건문, 할 일 없음) — eventFor :337-339 | eventFor B1 if@316(critical·알림 켜짐 → `EventCredentialExpiring`), B6 case@336 → 제목 리터럴 L337, 본문 Sprintf@338·Join@339; `Ntfy.Publish` Tags 헤더 첫 원소 = Type(call@117) | 맞음. 「할 일이 없다」 — remedy 는 「갱신에는 … 함께 필요하다」 조건 서술뿐, 명령 이름 없음(손으로 읽음) |
| 2 | 중복 방지 key 가 발급 신원을 담아 6시간 갱신마다 새 행 — stageKey :303-308, prepare if@287, outbox key :370 | prepare Format@286(신원 = IssuedAt\|ExpiresAt) → B3 if@287 「보냄」 재설정; stageKey return@305·@307(키에 신원 없음); eventFor L370 Key 접미사 `w.identity` | 맞음 — 신원은 키 문자열이 아니라 재설정과 outbox Key 두 자리에 있다(문서 표현과 일치) |
| 3 | 날수는 만료만 센다 — `daysLeftCeil(att.ExpiresAt.Sub(now))` :118 | **call daysLeftCeil@112**(L118 은 B12 case@118) | 사실은 맞고 줄 번호만 틀림 → proposal 정정(「AST: call daysLeftCeil@112」) |
| 4 | 감독 증거 시각은 계산되지만 판정에 안 쓰임 — ProofAt :91, binding :134-154 | judgeAttestation call binding@91; ProofAt 의 다른 사용은 eventFor B13 if@363(필드)뿐 | 맞음 |
| 5 | 발급자는 같은 유효 기간 값으로 만료와 감독 증거 수용을 정한다 — :242·:265 | BuildAttestation `c = c.withDefaults()`@222, `acceptSupervised(s.AccountRef, now, c.Validity, supervised)`@242, `IssuedAt: now.UTC()`@264, `ExpiresAt: now.UTC().Add(c.Validity)`@265 | 맞음 — **같은 변수·같은 `now`**. CLI 도 같은 `now`·`criteria.Validity` 로 `SucceededEndpoints` 를 먼저 거른다(supervisedProofs call@438) |
| 6 | a135 는 감독 증거 기한 예측을 일부러 뺐다(D6) | archive a135 design 「D6. 「갱신하면 POST 를 잃는다」 예측은 하지 않음」 — 「`acceptSupervised` 의 규칙 사본이 된다」 | 맞음 |

## design D1~D5 코드 주장 대조

| 결정 | 코드 주장 | AST·HEAD | 판정 |
|---|---|---|---|
| D1 | 기한 = `ProofAt + (ExpiresAt − IssuedAt)` 이 발급자 동작과 같다 | 위 5행 + acceptSupervised B6 if@332 `age < 0 \|\| age >= validity` 제외(정각 제외) | **한 시장 기록이면 같다.** 정각 경계: 기한 시각에 발급하면 이미 빠진다 — 「기한 뒤 첫 갱신」은 「기한 이후(포함)」 |
| D1 | `ProofAt` = endpoint 별 최신 중 가장 이른 것 | binding: range@138 의 endpoint 별 최대 → range@148 의 최소 | binding 은 맞다. 그러나 **파일에는 endpoint 당 증거 하나**(acceptSupervised B7 if@335 `seen`)이고 그것은 **입력 순서상 첫 유효 증거**다 — KR·US 두 기록이 다 있으면 supervisedProofs(range@428, 시장 순서 KR→US, sort.Slice@472 이름만) 때문에 더 오래된 KR 증거가 실릴 수 있다(프로브 실측). 그때 D1 기한은 실제 상실보다 이르다 → **Manager 결정 필요**(review.md Q1) |
| D1 | `ProofAt` 0 이거나 `ExpiresAt ≤ IssuedAt` 이면 기한 없음 | — (새 코드) | 추가 사실: `IssuedAt` 이 0 인 파일이면 `ExpiresAt.Sub(0)` 이 Duration 최대값으로 포화 → 기한이 먼 미래 → 결과적으로 만료만 센다(무해). 결정 무변경 |
| D2 | 시험 패키지 import 순환 | `go list -deps`/`-test -deps` 양방향 무관 | 어느 쪽에도 순환 없음(아래 1.4) |
| D3 | 기한 지남·파일 통과 → imminent | daysLeftCeil: `int(remaining/day)` 0 쪽 잘림 + 나머지 있으면 +1 → −59m59s → 1, −25h → 0 | 결정과 맞음. 날수 값(제목 「D-1」·「D-0」)은 구현 몫 |
| D4 | 감독 증거의 기한은 갱신으로 바뀌지 않는다 | supervisedProofs 가 매 발급마다 검증 기록에서 `At` 을 다시 유도(L449-457), acceptSupervised 가 `At` 을 덮지 않음(L339 는 Endpoint 만) | 한 시장이면 맞다. 두 시장이면 오래된 쪽이 기간을 벗어나는 순간 실린 증거가 바뀌어 기한이 뒤로 뛴다(Q1) |
| D4 | 「만료가 기한이면 종전과 같은 동작」 | 발급자 파일은 POST 를 `SupervisedBy` 로만 얻고(BuildAttestation range@246) `ProofAt ≤ IssuedAt` 이라 감독 증거의 기한 ≤ 만료 | 맞지만 발급자 파일에서는 **사실상 일어나지 않는다**(같을 때만). 정본 기존 Scenario(「만료 8일 전」·「정상 갱신」)와의 관계 → Q2 |
| D5 | 거절 신원 = 단계 + 누락 집합 + KST 날짜, outbox key 도 같은 신원 | prepare B3 if@287 가 신원 변경 시 **모든** 「보냄」을 지우고, eventFor L370 이 Key 에 `w.identity` 를 붙인다 | 고칠 곳 = prepare(신원·재설정) + stageKey(누락 집합) + eventFor(Key). 맞음 |

## AST 가 말하지 못해서 손으로 읽은 것

| 무엇 | 자리 | 왜 AST 밖인가 |
|---|---|---|
| `LiveOnlyEndpoints` = POST 주문·취소 둘, 엔진 `RequiredEndpoints` 에 둘 다 있음 | soak/soak.go:96-101, interlock.go:219-230 | 복합 리터럴 |
| `DefaultCriteria().Validity` = 30일, CLI `--validity` 기본 30일, `>0` 일 때만 덮음 | soak/attest.go:88-95, cmd/tossctl/soak.go:230·:510-511·:610 | 리터럴·대입 |
| JSON 시각 형식(RFC3339Nano 왕복) | attest.go `time.Time` 필드 + `json.MarshalIndent` | 표준 라이브러리 |
| `DefaultRemindAfter` = 1시간, outbox `ClaimSettled` = 전달됨·리마인드 창 안 | obs/notifier.go:60, journal/alert_claim.go:57-59 | 상수·주석 |
| Go http 헤더 값 검사(CTL 만 거절) | go1.26.5 `vendor/golang.org/x/net/http/httpguts/httplex.go:303-311` | 표준 라이브러리 |
| `sort.Slice` 는 n ≤ 12 에서 삽입 정렬(안정) — KR→US 순서가 이름 정렬 뒤에도 남음 | go1.26.5 `sort` pdqsort(`maxInsertion = 12`) | 표준 라이브러리, 계약상 비안정 |
