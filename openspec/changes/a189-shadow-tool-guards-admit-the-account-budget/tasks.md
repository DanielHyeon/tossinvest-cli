## 1. 계약과 증거 (1단계 — `.go` 무편집)

- [x] 1.1 memory recall(ai-memory·Hindsight)과 번호 재확인, worktree·브랜치 생성. 기록은 `review.md`.
- [x] 1.2 Story `STORY-TOS-a189`(FEAT-TOS-002) 선등록, change 문서 작성, `capture_change_base.py` 로 base 고정, PM tracker 재생성.
- [x] 1.3 CodeGraph(1.6.0) 호출자·`affected --filter '*_test.go'` 원문을 `analysis/code-context/` 에 남김.
- [x] 1.4 D2 열거(`analysis/enumeration/`), D5 영향 원문, 편집 전 AST 5+1개(`analysis/pre-edit-ast/`)를 만듦.
- [x] 1.5 `openspec validate --strict`·PM `--check` 원문을 `analysis/verify-1/` 에 남김.
- [x] 1.6 Manager 결정(2026-11-15): D4 = **A**(마지막 정확 경로 비교에 `internal/accountbudgetruntime/v2_token_test.go` 하나만, 접두사·접미사 조건 무편집; A-wide 기각), 스펙 = **A-2**(초안의 한 문장 개정을 specs/ 로 옮김 — 정본 블록과의 차이가 그 문장 하나뿐인 원문 `analysis/spec-a2-mb0-diff.txt`), R1 = 범위 밖. strict validate 재실행.

## 2. 구현 (RED → GREEN)

- [x] 2.1 RED: 편집 전 다섯 시험이 base 와 같은 문구로 실패함을 이 worktree 에서 원문으로 남김(3개 패키지, 시험 이름 지정 실행). 원문 `analysis/tdd/red-2.1.txt`: FAIL 5·PASS 0, 문구는 base 와 같음.
- [x] 2.2 GREEN: proposal 「What Changes」의 이름만 더함. 다른 함수 본문은 고치지 않음. 기존 문구·금지 목록 무편집. 원문 `analysis/tdd/green-2.2.txt`: PASS 5·FAIL 0. a128 import 실패 문구의 괄호 목록도 proposal 대로 두 패키지를 더해 사실에 맞춤.
- [x] 2.3 편집 뒤 FLM/BTM 번들(함수 5개, 선택지 B 면 그에 맞게)을 만들고 `check_analysis.py --change a189-shadow-tool-guards-admit-the-account-budget` 를 통과시킴. 원문 `analysis/tdd/check-analysis-2.3.txt`(required 5, evidence complete, rc 0).

## 3. 검증

- [x] 3.1 사본 변이(무변이 대조군 GREEN 먼저, 사본에 pid, 한 번에 한 판, 사본 대상 단언 — design D7 목록). 각 변이가 어느 시험의 어느 줄에 잡히는지 원문을 `analysis/mutation/` 에 남김. 1차 16/18(g2·g3 SURVIVED → 픽스처 시험 `a112_mbus_a189_static_test.go` 추가), 2차 18/18, 리뷰 반영 하네스 3차 18/18(기대 시험 이름 단언·HOME/XDG 격리).
- [ ] 3.2 영향 패키지 시험: `internal/official`, `internal/accountbudgetruntime`, `tools/a128-jev-shadow-probe`, `tools/a140-breakout-signal-shadow`, `tools/a112-mb-us-source`(측정 소스 신원 집합에 가드 파일이 있음 — design D4). `go vet`·gofmt. 전체 gate·전체 시험은 사용자가 요청할 때만.
  - 2026-11-15 실행(`analysis/verify/`): 다섯 가드 포함 PASS(최상위 551·서브 포함 961), **FAIL 3**(`internal/accountbudgetruntime` `TestA166WiringCensusBaseline`·`TestA166WiringCensusRejectsBypasses`·그 하위 `new-unclassified-http`), SKIP 2, vet rc 0, gofmt 0. FAIL 3 은 base 7d1e3163a 사본에서 같은 문구(`reviewed function shape changed: tools/a128-jev-shadow-probe/main.go||runCommandWithOptions`)로 실패 — `analysis/verify/base-a166-census.txt`. 이 change 범위(D1) 밖이라 고치지 않음 → Manager 결정 대기로 미체크.
  - 보강 뒤 재실행(`analysis/verify-2/`, HEAD a16dc8292): 같은 결과 — PASS 551·961, FAIL 3(같은 census 셋), SKIP 2, vet rc 0, gofmt 0. census 는 Manager 결정 5 로 사용자 결정 대기.
- [x] 3.3 gstack `review` 1회(다중 적대 리뷰는 not-applicable — 비례 원칙), 처리 결과를 `review.md` 에 기록. 원문 `analysis/gstack-review.md`. Manager 결정 대기 4건(발견 1·9·10, gate G1~G3)은 review.md 「Manager 에게 올림」.
- [ ] 3.4 Manager 독립 재검증.

## 3b. a166 census 고정 모양 (사용자 결정 2026-11-15 으로 범위 포함)

- [x] 3b.1 읽기 전용 조사(`analysis/census-investigation/`): 고정 모양 = `78d777cfe` 시점, 변화는 a179 입력 프로필뿐, 그 키만 갱신한 사본에서 census 시험 셋 PASS.
- [x] 3b.2 `shapes.json` 의 키 `tools/a128-jev-shadow-probe/main.go||runCommandWithOptions` 값 하나만 갱신(Manager 리뷰 승인). 키별 비교 원문 `analysis/census-repair/shapes-key-diff.txt`: 바뀐 키 1개, 무변경 89개, `sites.json` 동일.
- [x] 3b.3 반증(`analysis/census-repair/falsify.txt`): 대조군 GREEN, 고정 함수 안 `http.Get` → 「reviewed function shape changed」, 새 함수 안 `http.Get` → 「callsite inventory count changed: got 117 expected 116」, 둘 다 CAUGHT.
- [x] 3b.4 영향 패키지 재확인(`analysis/verify-3/`, `tools/a112-mb-us-source` 는 Manager 지시로 제외): 4개 패키지 ok, PASS 최상위 485·서브 포함 882, FAIL 0, SKIP 0. `internal/accountbudgetruntime` 최상위 53·서브 포함 77, census 3개 시험 PASS. vet rc 0.

## 3c. 낡은 시험 가드 3건 (사용자 결정 2026-11-16 「1번 — a189 에 전부 포함」, design D11~D13)

- [x] 3c.1 범위 확장을 proposal·design·tasks 에 반영. 원인 커밋과 부모에서 세 시험의 PASS→FAIL 경계 실측(`analysis/boundary/`), #2 렌더 해시 커밋별 측정·`record` AST 대조(`analysis/render-hash/`), 편집 전 AST(`analysis/pre-edit-ast/` 두 파일).
- [x] 3c.2 #1 `TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient` — `(*lazyOrders).Orders` 본문의 공유 resolver 참조 수 단언(D11).
- [x] 3c.3 #2 `a163JudgementRenderHashes["record"]` 갱신과 이유 주석(D12).
- [x] 3c.4 #3 `admitSites` 좁힘(D13)과 새 파일의 fixture 시험.
- [ ] 3c.5 사본 변이(대조군 GREEN 먼저, 기대 시험 이름 단언) — 하네스·원문 `analysis/mutation-2/`. #3 은 좁힘으로 잃는 검출을 열거해 잼.
- [ ] 3c.6 FLM/BTM 번들, `check_analysis.py` required 수·rc 실측.
- [ ] 3c.7 대상 패키지 시험·vet·gofmt·strict validate — 원문 `analysis/verify-4/`.
- [ ] 3c.8 gstack `review`(최종 diff 전체) + 다른 계열 적대 보이스 1, 새 발견 0 까지 — `analysis/gstack-review-2.md`.

## 4. 종결

- [ ] 4.1 PM check, strict validate, archive 와 Story 경로 동기화. 검증된 교훈만 retain.
