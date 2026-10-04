Codex · GPT-6(세부 버전 미제공) · ~/.codex 미접근

HOLD — B2 조기 활성화 읽기가 취소된 수집을 붙잡는다. P0 재현 없음. A·B1의 결정/수락 집합 변경은 발견하지 못했다.

검토 기준은 작업 트리/HEAD가 아닌 네 archive다.

| 대상 | 전 | 후 |
|---|---|---|
| A | d54f4dca0475ae61795cc71d110b54dea1f04784 (`65444341^`) | 654443415b36dca624a5e73f5e4d7aadc5fc9b9e |
| B | f473d8155da16c06e90a228593031e56f83c4712 (`0b441270^`) | 0b441270e90ed2c034a288087e71a061b1b9bae6 |

모든 증거/시험/overlay는 이 디렉터리 아래에 있다. 실제 저장소는 archive/rev-parse/status 읽기에만 사용했다. 사본에서 git은 실행하지 않았다. Go는 `run-test.sh`가 지정 GOCACHE/GOTMPDIR/GOFLAGS/GOPROXY/GOSUMDB/GOTOOLCHAIN을 매번 설정하여 `go test`만 실행한다. `TMPDIR`도 사본의 `.gotmp`로 지정했다. 네트워크/실브로커/LIVE/운영 토글/운영 엔진 lock 실행 없음. 사용한 제안·활성화·gateway는 시험 fixture다.

| 등급 | 지적 | 재현 | 결과 | 제안 |
|---|---|---|---|---|
| P0 | 재현된 지적 없음 | 아래 닫힌 시장/dispatch 공격 | 노출 상승 경로 확인 못함 | 안전성을 모든 입력에 대해 증명했다는 의미는 아님 |
| P1 | B2: 앞에서는 FX 실패로 즉시 끝나던 수집이 이제 ctx를 확인하지 않는 파일 읽기에 종속됨 | `TestReviewB2CancelledAndDelayed`, `TestReviewB2ProductionReadDelay`; B-pre/B-post 동일 시험, 후자는 `*-io-overlay.json` | 전: FX_NOT_READY, 적재/파일 읽기 0. 후: 읽기 진입 후 취소해도 150ms barrier 동안 `collect` 미반환, 읽기 해제 뒤에만 FX_NOT_READY. `collect`는 두 시장 결과를 모두 `<-outcomes`로 기다림 | 닫힘 사유를 먼저 확정하고 진단용 활성화 적재가 반환을 무한정 붙잡지 않도록 경계를 설계. 취소/지연 시험을 정규 회귀로 포함. goroutine을 무제한 남기는 timeout 포장만으로 해결하지 말 것 |
| P2 | B2: 닫힘 사유 순서 불변 주장이 nil getenv에서 깨짐 | `TestReviewB2NilGetenv` | 같은 route-ready/FX-not-ready 입력: 전 FX_NOT_READY → 후 INTERNAL_FAILURE. 양쪽 Ready=false, entries=0. 후판은 getenv 검사 전 `loadFamilyActivation`이 nil 함수를 호출, `collect`가 panic을 회수 | FX 우선순위와 설정 결손 처리를 함께 보존하도록 gate 적재의 nil 전제 검사. AST 순서 외 다중 실패 조합 행동 시험 추가 |
| P2 | B1: 실패한 lane_id 원문·개행이 오류 문자열에 복사됨 | `TestReviewB1DiagnosticPayload` | 전 sentinel만 출력. 후 `descriptors[lane_id=synthetic-account-123\nFORGED_RECORD]: lane_id` — 실제 문자열에 개행 포함. Undeclared=false, 계속 거절 | 미검증 값 대신 descriptor index/필드명을 사용. 값이 꼭 필요하면 길이 제한·인코딩·민감값 정책 적용. 실제 계좌/시크릿 유출이나 운영 로그 도달은 입증하지 않았음 |
| P2 | A: 골든의 첫 counterfactual 값과 실행 리터럴이 자동 결속되지 않음 | `A-golden-drift`는 A-post 사본에서 JSON의 `rvol_counterfactual_ppm[0]`만 1200000→1300000, manifest SHA 갱신. 전체 breakoutlane 시험 실행 | PASS. 실행 코드는 여전히 1200000이고 임계값 시험도 통과. 골든 파일 무결성 검사만으로 의미 일치는 보장되지 않음 | 승인 골든 배열과 실행 경계 시험을 직접 연결. 현행 1.2 의미 구현 오류는 아님 |

P1의 범위: 주입한 stalled-reader를 통해 새 대기 의존성과 취소 비응답을 실행 확인했다. 실제 운영 디스크 교착 또는 손절 지연을 재현한 것은 아니다. `readProductionRouteFile`의 실제 OS 호출은 ctx를 받지 않으며, 제안 수집의 join에도 ctx select가 없다. 읽기 자체가 이미 취소된 ctx로 시작되면 정상적으로 빠르게 닫힌다. 문제는 읽기에 들어간 뒤의 취소다. 새 조기 읽기 경로 때문에 기존 FX/설정 실패 주기도 이 대기 가능성에 포함된다.

시도했지만 결정/수락 집합을 깨지 못한 공격:

- A: seed 11285의 임의 봉열 2,400개 + 명명된 8개(1.2-only 다중 봉, 입장 전/후, 첫 touch와 공존/동일 봉, opening-range 경계, invalidation, timeout). 각 입력에 fresh/reuse/옛 flag prior/correction 네 모드. 9,632개에서 phase/refusal/transitions/candidate/final/proposalID/snapshotDigest/diagnostic/seal 및 1.2 flag 외 provenance가 전후 동일했다. 1.2 flag만 1,441개 달랐다. SHA-256: 양쪽 JSONL `01c175a70edbb05832c579492fd306a01637456f2007d642cc1987e443da6b18`.
- A prior: 같은 digest에서 옛 flag를 가진 유효 prior를 그대로 반환한다. 비종결 정정은 다시 평가하고, 종결 정정은 보존한다. 플래그의 관측 신선도는 달라질 수 있지만 이번 입력군에서 판정 불일치는 없었다. `decisionSeal`은 이 flag를 제외하고, `snapshotDigest`는 원시 RVOL을 포함한다. RVOL 1.0 쌍둥이와 seal을 비교한 실험이 아니라, **같은 입력**을 두 커밋에서 비교했다.
- A 소비자: `cmd/internal/tools` 생산 Go에서 `.Provenance()` 호출 없음, `RVOLAt1200000` 참조는 breakoutlane 정의/쓰기만 확인. `internal/strategyproposal/production.go`의 breakout 입력은 `ErrBreakoutEvidenceUnavailable`로 닫힌다. 소비자 부재는 그 검색 범위의 현재 소스 결론이다.
- A 시험 판별력: 임계값 1200000→1200001, 기록 true→false overlay는 각각 기존 정확히 1.2 시험에서 CAUGHT. 현재 숫자 경계는 막지만, 골든을 새로 승인했을 때의 의미 결속까지 막지는 못한다.
- B1: 37개 필드 변이의 단독/쌍 + seed 884 무작위 조합, 총 2,704개. 각 조합에 정상 핀/미선언 핀/불일치 핀/nil ctx/취소 ctx/비정규 bytes 적용, 16,224행 전후 완전 동일. accepted 29, unavailable 10,650, undeclared 2,704, canceled 2,704, revoked 114, expired 23. SHA-256 양쪽 `8206cb9334d42b4ca8a94f8e7ffa0f5c665d06cf271a40cbab008234ac50bfc6`.
- B1: unknown lane/family/horizon/version/desired/effective, malformed time 및 영값 비교, 설정 market, 만료/폐기와 다른 오류 우선순위를 포함했다. 알려지지 않은 lane은 `!known` 자체로 거절되어 `known &&` 묶음으로 수락되지 않았다. config market 검사는 두 판 모두 filename 공백 검사다.
- B1: missing/directory/symlink/wrong-mode/empty/unknown-JSON-field/bad-JSON 읽기 실패 일곱 모양은 양쪽 모두 Unavailable=true, Undeclared=false, Verified=false. `%w` 추가가 실제 읽기 경로에서 미선언을 만드는 반례는 없었다. 실제 reader의 오류 원천은 route sentinel 또는 OS error이며 activation Undeclared를 반환하는 원천이 없다. 임의로 reader가 Undeclared를 반환하도록 바꾸는 비현실적 변이는 현재 입력 반례로 세지 않았다.
- B1: `protection_ready_min_generation` 조건 누락 및 `failedFields` 빈 목록 강제 변이는 기존 필드 시험에 CAUGHT. 누락 방어는 각 필드의 음성 시험이 담당하며, 앞으로 추가할 미시험 항목까지 자동 보증하지는 않는다.
- B2: FX 미준비 + 검증 활성화. 전 ON 0 → 후 ON 4로 관측 변화. 양쪽 entries=0/emitted=0/dispatch callback=0. callback 자체가 호출되지 않아 그 아래 lease 생성/ProtectionReady 검사/주문 dispatch로 진입하지 못했다. 상한/하한은 독립적으로도 기존 관련 시험 통과.
- B2 B11/B14: 각각 `arbitration.collision || true`, `!resolved || true` overlay로 입력 도달 난점을 분리해 **실행**했다. 양쪽 verified=true, entries=0, ON=4, emitted=0, callback=0. 두 갈래는 앞 판에서도 gate 계산 뒤에 있었으므로 실리는 활성화가 이번 이동 때문에 새로 생긴 것은 아니다. 자연 입력 도달성 증명은 아니다.
- 기존 검증: A-post breakoutlane 전체 PASS. B-post 활성화 전용 기존 시험 20개(하위 사례 포함) PASS. B2 기존 7개(13갈래 AST census, 닫힘 8×관문 3, 닫힌 시장 handoff, 레인 투영, lease/만료 등) PASS. 기존 시험이 초록이어도 위 회귀가 별도 입력에서 재현됨을 확인했다.

소스 좌표(각 `B-post/` 내부): `strategy_proposal_authority.go:325` 조기 gate, `:329` nil getenv 검사, `:269-283` panic 회수/join, `strategy_family_activation.go:185` getenv 호출, `production_owner_unix.go:12` ctx 없는 읽기, `production_family_activation.go:604` raw lane_id 오류.

재현 명령(지정 작업 디렉터리 기준):

```bash
set -euo pipefail
bash review-r2/reproduce.sh
```

개별 지적을 빨리 재현하려면:

```bash
set -euo pipefail
bash review-r2/run-test.sh B-pre -tags tossos_testseams ./internal/app/engine -run '^TestReviewB2(NilGetenv|CancelledAndDelayed)$' -count=1 -v
bash review-r2/run-test.sh B-post -tags tossos_testseams ./internal/app/engine -run '^TestReviewB2(NilGetenv|CancelledAndDelayed)$' -count=1 -v
bash review-r2/run-test.sh B-post -overlay "$PWD/review-r2/B-post-io-overlay.json" -tags tossos_testseams ./internal/app/engine -run '^TestReviewB2ProductionReadDelay$' -count=1 -v
bash review-r2/run-test.sh B-post ./internal/strategyrouter -run '^TestReviewB1DiagnosticPayload$' -count=1 -v
REVIEW_OUT="$PWD/review-r2/A-golden-replay.jsonl" bash review-r2/run-test.sh A-golden-drift ./internal/breakoutlane -count=1
```

증거 파일: `A-{pre,post}.jsonl`, `A-{pre,post}.jsonl.flags`, `B-{pre,post}.txt`, `B-{pre,post}-b2.log`, `B-{pre,post}-io.log`, `B-{pre,post}-b1.log`, `B-{pre,post}-read-sentinels.log`, `B-{pre,post}-{downstream,B11,B14}.log`, `A-golden-drift.log`, `*-verify.log`, 네 mutant `*.log`. 시험 소스는 `review_*_test.go`, 변이는 `*.go`/`*-overlay.json`에 보존했다.

범위 밖은 등급 없음: 운영 snapshot/projection의 activation 오류 표면, 공유 reader의 OS 원인 축약, SHADOW/base 재고정. 새 raw lane_id 오류 문자열 문제는 B1 명시 질문에 해당하므로 P2에 포함했다.

절차 한계: 제공된 OpenSpec/Function Logic Map·AST와 현재 소스를 사용했다. user의 go-test-only/읽기 전용 제약에 따라 CodeGraph 재색인, go run AST 생성, make sdd-sync/sdd-check/gate, PM/archive/commit/push는 실행하지 않았다. 이는 구현 완료 보고가 아니라 지정 범위 독립 리뷰다. 추가 agent/전역 skill/memory 접근 없음.

실제 저장소 무변경 확인: 시작/종료 `git -C /mnt/D/Axipient/workspace/TossOS status --short` 출력이 `cmp`로 바이트 일치했다. 두 파일에 동일하게 아래 세 기존 untracked 항목만 있다. 이 측정은 요청된 status 비교이며, 기존 untracked 파일 내용의 별도 해시 감사를 뜻하지 않는다.

```text
?? .reticle-setup-crash.log
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/review-8.5-2026-10/
?? w4.log
```
