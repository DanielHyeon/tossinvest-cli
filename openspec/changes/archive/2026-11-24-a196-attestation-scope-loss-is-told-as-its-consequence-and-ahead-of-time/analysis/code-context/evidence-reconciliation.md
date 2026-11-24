# Evidence reconciliation (a196, 1단계)

기준 HEAD `064f59fd9`. CodeGraph 1.6.0(worktree 색인)·AST 22(`../cited-ast/`)·grep·overlay 프로브(저장소 무편집 — 아래)를 대조했다.

| 질문 | CodeGraph | HEAD(grep·AST·실측) | 결론 |
|---|---|---|---|
| `attestationWatch.prepare` 의 호출자 | 못 찾음(동명 8건) | grep `w\.prepare()`: attestation_watch.go:232(`checkOnce`) 하나 | 1 — CodeGraph 이름 해소 누락, 결론 불변 |
| 발급자의 만료와 감독 증거 수용이 같은 값인가 | `acceptSupervised` 호출자 1 | AST: BuildAttestation `c.withDefaults`@222 → `acceptSupervised(…, now, c.Validity, …)`@242, `ExpiresAt: now.UTC().Add(c.Validity)`@265, `IssuedAt: now.UTC()`@264 | **같은 변수·같은 `now`** → `ExpiresAt − IssuedAt = Validity` 정확히(JSON RFC3339Nano 왕복 손실 없음) |
| 수용 경계 | — | AST acceptSupervised B6 if@332 `age < 0 \|\| age >= validity` → 제외. 프로브: age = V−1ns 수용, V·V+1ns 제외, 0 수용, −1ns 제외 | 기한 시각 **정각에 발급하면 이미 빠진다** — 「기한 뒤 첫 갱신」은 「기한 이후(포함) 첫 갱신」 |
| CLI 의 이중 필터 | `SucceededEndpoints` 호출자 1 | runSoakAttest `now := time.Now().UTC()`·`supervisedProofs(…, now, criteria.Validity)`·`BuildAttestation(…, criteria, now, …)`; SucceededEndpoints B5 if@92 `age < 0 \|\| age >= maxAge` | 같은 `now`·같은 validity·같은 경계 — 이중 필터는 결과를 바꾸지 않음. `criteria` 는 `loadSoakSummary` 의 `soak.DefaultCriteria()`(soak.go:610, 30일)이고 `--validity > 0` 일 때만 덮이므로(:510-511) 두 함수에 들어가는 validity 는 항상 같은 양수(손으로 읽음) |
| `SupervisedBy[].At` 이 원 증거 시각인가 | — | AST acceptSupervised: `p.Endpoint = endpoint`(L339) 만 덮고 `At` 은 그대로; supervisedProofs 는 매 발급 때 검증 기록에서 다시 만든다(L449-457, 이전 attestation 을 읽지 않음) | 원 증거 시각 그대로, 갱신마다 기록에서 재유도 |
| 파일에 실리는 증거가 endpoint 별 **최신**인가 | — | AST acceptSupervised B7 if@335 `seen[key]` → **입력 순서상 첫 유효 증거**; supervisedProofs 는 KR 기록 → US 기록 순으로 붙이고(B5 range@418) `sort.Slice` 로 endpoint 이름만 정렬(@472). 프로브: KR(20일 전)·US(2일 전) 같은 endpoint → **KR 이 실림** | **두 시장 기록이 다 있으면 파일의 증거는 최신이 아닐 수 있다** → D1 기한이 실제 POST 상실보다 이르다(거짓 경보 방향, 놓침 방향 아님). Manager 결정 필요(review.md Q1) |
| 감독 증거가 실린 통과 파일에서 어느 쪽이 기한인가 | — | 발급자 파일은 POST 두 개를 `SupervisedBy` 로만 얻고(BuildAttestation B5 range@246) 증거는 `age ≥ 0` 만 실리므로 `ProofAt ≤ IssuedAt` → 감독 증거의 기한 ≤ 만료 | 발급자 파일이 Verify 를 통과하면 **항상 감독 증거가 기한**(같을 때 포함). 「만료가 기한」은 `SupervisedBy` 없이 POST 를 가진 파일(손으로 쓴 파일 등)뿐 |
| 중복 방지 신원의 자리 | `stageKey` 호출자 1 | AST: prepare Format@286 + B3 if@287(「보냄」 재설정), stageKey return@305·@307(키), eventFor L370(outbox Key 접미사) | 발급 신원은 키 자체가 아니라 재설정·outbox Key 두 곳 — 둘 다 고쳐야 D5 |
| ntfy 제목 인코딩 | `Ntfy.Publish` | AST call `req.Header.Set`@114 — `Title` 헤더에 UTF-8 원바이트; go1.26.5 httpguts `ValidHeaderFieldValue` 는 CTL 만 거절 | 클라이언트는 한글을 그대로 보낸다. 서버 표시는 저장소 밖(ntfy.go:112 주석 「ASCII-folded by the server」는 미검증) — review.md Q3 |
| import 순환 | — | `go list -deps ./internal/app/engine` · `-test -deps` 에 `internal/soak` 없음, `go list -deps ./internal/soak` · `-test -deps` 에 engine 없음 | 어느 방향 시험 import 도 순환 없음 |
| a166 census | — | `testdata/a166-wiring/shapes.json`(키 90)·`sites.json`(76 파일): `attestation_watch` 0, `internal/soak/attest.go` 0, `cmd/tossctl/soak.go` 2(`buildSoakWiring`·`runSoakRun` — 편집 안 함), `ntfy.go` 1(`Ntfy.Publish|client.Do` — 편집 안 함) | 재기준화 불필요 |

불일치: CodeGraph 의 `prepare` 호출자 누락 1건(grep 으로 해소, 결론 불변). 그 밖 없음.

## overlay 프로브 (저장소 무편집)

`go test -overlay` 로 패키지에 가상 시험 파일을 얹어 쟀다 — 저장소 Go 파일 편집 0(하네스가 끝에 `git status -- internal` 을 단언). 하네스·원문은
`../probe/`(`run.sh`, 프로브 소스는 `.go.txt` 로 커밋 — `.go` 가 아니라 게이트·빌드에 안 잡힘, 결과 `output.txt`, 저장소 루트에서 `bash …/probe/run.sh` 로 재현).
두 시장 순서(KR 먼저)는 supervisedProofs 의 코드 읽기이고, 프로브는 그 순서가 주어졌을 때 acceptSupervised 가 첫 것을 싣는 것만 잰다.

| 픽스처 `a135Attestation(r)` 의 r | D1 기한까지 남은 시간 | daysLeftCeil | 만료 기준(종전) |
|---|---|---|---|
| 7d+1s | 167h0m1s | **7** | 8 |
| 30d | 719h | 30 | 30 |
| 7d | 167h | 7 | 7 |
| 6.5d | 155h | 7 | 7 |
| 3d | 71h | 3 | 3 |
| 2d+1s | 47h0m1s | **2** | 3 |
| 2d | 47h | 2 | 2 |
| 1.5d | 35h | 2 | 2 |
| 1d | 23h | 1 | 1 |
| 1s | −59m59s | 1 | 1 |
