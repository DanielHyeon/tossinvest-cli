# a203 리뷰 기록

## 구현 리뷰 1회차 (2026-11-27, tasks 1.x ~ 2.x — 커밋 `0951a2f48` 대상)

- 비례 원칙: 변이 원장 · 다중 적대 리뷰 **not-applicable** — 주문 경로 밖 오프라인 측정 도구, High-risk 경로 · 게이트 판정 변경 0. 경량 리뷰 1회.
- 보이스 두 개: (1) Codex(`codex exec --sandbox read-only`, 작성자와 다른 계열) · (2) Claude 적대 리뷰어(「attacker · chaos engineer」 프롬프트).
  둘 다 읽기 전용, 네트워크 · 자격 파일 접근 금지로 지시.
- 리뷰 2회차 상한 · design 밖 새 가드 금지(Manager 지시). 그래서 지적 중 design 이 정하지 않은 가드는 **보고만** 함.

| # | 보이스 | 지적 | 처분 |
|---|---|---|---|
| 1 | Codex(CRITICAL) · 적대(INFO 5/10) | 시간대 거절은 클라이언트 메서드 호출 직전 한 번 — `official.Client.send` 안의 토큰 취득 · 401 재시도(최대 2회, `internal/official/client.go`)는 그 뒤라 경계 직전 전송이 몇 초 늦게 거절 창 안에서 나갈 수 있고 간격 · 상한 셈에도 안 들어감 | **보고만**(design §8 「매 GET 직전」과의 차이). 고치려면 전송 계층(RoundTripper) 감싸기 — a166 census 배선 · M2 범위 밖. 운영 완화: 거절 경계 가까이 시작하지 않음(dry-run 이 다음 경계를 찍음) |
| 2 | Codex(CRITICAL) | 마지막 봉에서 난 제안을 `labelProposal(final=true)` 가 그 봉 종가로 「close」 라벨함 | **기각** — a140 과 같은 라벨 함수 · 같은 규칙(design §3-6). 라벨 봉은 진입 분의 봉이라 진입 뒤 가격이고, a140 보고서도 세션 끝 「close」 로 셈 |
| 3 | Codex(CRITICAL) · 적대(INFO 4/10) | `--state-dir` 를 저장소 안으로 주면 캐시(가격) · 종목군 파일이 저장소에 떨어짐 | **보고만** — 기본값은 저장소 밖(`~/.local/state/tossos-a203`), design 은 경로 가드를 정하지 않음 |
| 4 | Codex(CRITICAL) · 적대(INFO 6/10) | 대조 1 불일치가 있어도 `collect --sample main` 을 막지 않음(보고서에 fail-stop 문구만) | **보고만** — design §6 · tasks 3.1 의 fail-stop 은 사람 절차(「4 로 가지 않음」). 코드 게이트는 design 에 없음 |
| 5 | 적대(INFO 6/10) | 수집 중 캐시 파일을 못 읽으면(깨짐 · EACCES) 생산자 거절로 세고 넘어가 이어하기마다 영구 결손 | **수리** — 리더 오류(`READER_ERROR`)는 실행을 멈춤(design §8 「오류는 멈춤」). 시험 `TestA203CollectStopsOnAnUnreadableCacheFile`(수리 전 사본에서 FAIL 확인) |
| 6 | 적대(INFO 6/10) | `--max-gets` 는 실행당이라 이어하기마다 새 13,600 — H3 누계 상한이 강제되지 않음 | 보고만 — design §8 은 「실행당」. 누계는 `runs.jsonl` 합으로 보고서에 찍힘 |
| 7 | 적대(INFO 5/10) | 마지막 페이지가 남은 몫 0 을 실어도 다음 작업이 바로 보냄(의도된 429 를 맞음) | 보고만 — 첫 429 에서 멈추므로 spec 안 |
| 8 | 적대(INFO 5/10) | 일봉 200 개는 수집일에서 거슬러 셈 — 수집이 늦으면 이른 창의 d−k 가 조용히 빠져 M6 결측으로 읽힘 | 보고만 — design §5 가 73 세션 여유를 적음(이번 주말 수집이면 해당 없음) |
| 9 | 적대(INFO 5/10) | 새 `a203Live` 인터페이스의 메서드 집합은 static_test 가 고정하지 않음(금지 이름 시험은 걸림) | 보고만 — 새 가드 |
| 10 | 적대(INFO 3/10) | rename 뒤 디렉터리 fsync 없음 · `runs.jsonl` 끝 줄이 찢기면 report 가 실패 · 성공한 이어하기 뒤 낡은 `checkpoint.json` 이 남음 | 보고만 |

Gate edits(gstack gate-diff): 2 — `shapes.json` dispatch 1건 · `sites.json` a203 자리 4건(둘 다 RH-3?, 결정 M2 범위). 처분 **open** — M2 가 「census 재검토는 구현 뒤 Manager 의 diff 검증」으로 정함.

## 구현 리뷰 2회차 (2026-11-27, 단계 평가 시각 정정만 — Manager 결정 design §3-4)

- 대상: `a203ReplaySymbol` 의 단계 평가 시각(봉 n−1 이 열린 뒤 2 분 5 초) · 마감 컷오프, 새 시험 `TestA203StagesStopAtTheA140CloseCutoff`,
  design §3-4 정정 문구. 보이스: Codex(`codex exec --sandbox read-only`, 다른 계열) 1회. 리뷰 상한 2회 도달.

| # | 지적 | 처분 |
|---|---|---|
| 1 | 시험이 `+5_000`(5 초)를 지워도 통과함(분 경계 마감에서는 결과가 같음) · INVALIDATED/TIMED_OUT 컷오프 사례 없음 | **반영(시험만)** — 감지 평가 시각 = open(k)+2m5s 를 초기 대리 호가 관측 시각으로 단언. 컷오프는 단계와 무관한 같은 `break` 라 단계별 사례는 더하지 않음 |
| 2 | 「a140 과 정확히 같음」은 과장 — 장중 지연 · 빠진 분은 재생이 재현 못 함 | **반영(문서)** — design §3-4 에 「같음 = 제때 도착 일정」 범위와 대조 1 불일치 설명 후보를 적음 |
| 3 | 리뷰어 샌드박스에서 `go test` 가 읽기 전용 파일시스템으로 막힘 | 정보 — 작성자가 같은 시험을 돌림(`analysis/verify/go-test-a140.txt`) |

수리 전 RED: `analysis/verify/red-close-cutoff-before-fix.txt`(15:28 봉에서 ARMED 감지 → FAIL, rc 1).

## 종결 경량 리뷰 (2026-11-28, tasks 5.1 — HEAD `6d383eff4`, diff `e2921fc00..HEAD` 의 `tools` · `internal`)

- 비례 원칙: **변이 원장 not-applicable** · **다중 적대 리뷰 not-applicable** — 주문 경로 밖 오프라인 측정 도구(브로커 조회는 사람 승인 · 시간대 거절 아래 읽기 전용 GET 만),
  High-risk 경로(주문 · 손절 · 익절 · 사이징 · Guardian · 원장 · 대사 · 인증 · 체결) 편집 0 · 게이트 판정 변경 0. Function Logic Map 은 기존 함수 `dispatch` 1개 번들(1.3).
- 보이스: gstack `review` 1회(Opus 팀메이트). 전문가 5(testing · maintainability · security · performance · simplification) + Claude 적대 패스 1, 모두 새 문맥 · 읽기 전용.
  Codex 외부 패스는 이번 회차 생략(경량 1회 · 1·2회차에 Codex 를 이미 씀) — 외부 보이스 커버리지 없음으로 기록.
  스모크: `go test -count=1 -run TestA203 ./tools/a140-breakout-signal-shadow/` rc 0(56.9s).
- Manager 지시: 리뷰 2회차 상한 · 리뷰발 새 가드 금지 · 종결 단계라 Go 편집 0 을 목표(착지 기록 보존). 그래서 아래 처분은 전부 **보고만**이고 코드 · 시험 변경 0.
- 판정 영향 확인: 주 표본은 수집 1회로 끝남(6000/6000 · 정지 0, `analysis/replay/runs.md`), 보고서 결손 P0 · S1 · S2 각 0/3000(`analysis/replay/report.txt`). 그러므로 아래 1 은 이번 FAIL 판정에 영향 없음.

| # | 보이스 | 지적 | 처분 |
|---|---|---|---|
| 1 | 적대(High, blocks) | `report` 는 일봉 결손(`dailyMiss`)만 거절하고 수집 안 된 분봉 페이지(`errA203NotCached`)는 `a203ReplaySymbol` 에서 「bars:」 결손으로 셈(`a203_replay.go:766`, 거절은 `:947`) — 수집이 중간에 멈춘 캐시로 report 를 돌리면 결손 20% 이하에서는 잘린 표본으로 PASS/FAIL 을 찍음 | **보고만**(실재 확인, 잠재 결함). 이번 주 표본은 결손 0 이라 판정 무영향. 같은 도구를 다시 쓰는 Story(a204 등)가 「report 전 미수집 분봉 0 · checkpoint 없음」을 거절 조건으로 넣을 것 — 리뷰발 새 가드 금지 |
| 2 | 적대(Medium) | ARMED 뒤 시도 분의 봉이 없으면 시도를 건너뛰고(`a203_quote_proxy.go:76`) 결손으로 안 셈 | 보고만 — 주석이 정한 설계(「창은 분으로 잼, a140 과 같음」), 보고서 한계 5 |
| 3 | 적대(Medium) | 라벨이 진입 봉 k+1+i 다음 봉부터(`:94`, 진입 분 안의 손절 미검사) | 보고만 — design C6 의 사전등록 규칙, 1회차 2번과 같은 계열 |
| 4 | 적대(Low) | 캐시 키가 커서의 `:` 를 지우고 적중 시 저장된 요청과 대조 안 함(`a203_cache.go:190` · `:123`) | 보고만 — 커서는 도구가 만든 고정 형식 |
| 5 | 적대(Low) | 증거 저장소 하나를 대조 · 주 표본 · 재실행이 공유 — 재수집 봉이 다르면 개정 충돌이 「bars:」 결손으로 보임 | 보고만 — 이번 실행 결손 0 |
| 6 | 적대(Low) | 보고서 「GET 합」이 `runs.jsonl` 전부(sessions · control · main)를 더함(`a203_verdict.go:294`) | 보고만 — 보고서 줄이 종류별 내역을 같이 찍음(calendar · daily · minute · ranking) |
| 7 | testing(INFO) | 세션 목록 대조 거절(`a203_replay.go:434`) · 대조 1 불일치 경로 · collect 의 RateGated 정지 · 여러 fail-closed 가드(lot 파라미터 · 옵션 · 캐시 schema)의 거절 시험 없음 · 재생 원장 행을 읽는 시험 없음 | 보고만 — 시험 보강은 후속(종결 단계 Go 편집 0) |
| 8 | maintainability(INFO) | `a203Prepare` 주석의 `needAll` 이 낡음(`:490`) · 판정 규칙 문자열이 상수가 아닌 리터럴이라 lot 지문이 상수 변경을 못 따라감(`:884`) · `a203Trade` 의 안 읽히는 필드(`a203_quote_proxy.go:51`) | 보고만 |
| 9 | simplification(advisory) | `a203Counts` 가 `joinCounts` 한 줄 감쌈(`a203_verdict.go:226`) | 보고만 |
| 10 | security · performance | NO FINDINGS(security 는 시험 · 픽스처를 요약 모드로 읽음) | — |

Gate edits(gstack gate-diff): 2 listed, 2 read, 2 findings — 0 kept, 0 restored, 2 open — `shapes.json` dispatch 1건 · `sites.json` a203 자리 4건(RH-3?, 결정 M2 범위, 1회차와 같은 두 덩어리).
처분 **open** 유지 — M2 가 census 재검토를 Manager 의 diff 검증으로 정함.
