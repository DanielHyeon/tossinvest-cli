codex
**P2 1건, P3 2건입니다.** V8·§4 밖의 잘못된 자동 해제를 확정한 P0/P1 반례는 찾지 못했습니다. 아래 P2는 실제 해제 기록의 사실 오류입니다.

1. **P2 — 해제 근거가 증명하지 않은 커밋 순서를 기록함**

   반례 입력:

   | 순서 | 브로커·장부·수렴 |
   |---|---|
   | 주기 k 수집 | 종목 S의 브로커 스냅숏 10, 로컬 장부 10 |
   | 표식 읽기 | 전역 최대 fill id = 100 |
   | **로컬 읽기 전** | 엔진 SELL 5가 커밋됨. fill id = 101, 로컬 5 |
   | 로컬 읽기 | 5를 읽으므로 비교는 broker 10 / local 5 |
   | 수렴 | `Positions` 읽기의 일시적 오류로 중단. `Converged`는 비어 있음 |
   | Observe 이후 | Refresh·Observe는 성공. `101 > 100`이므로 체결 credit 부여 |
   | 주기 k+1 | 더 늦은 스냅숏 5 / 로컬 5, 추가 체결·조정 없음. `FILL_APPLIED` 해제 |

   실제 체결은 **로컬 읽기 전에 이미 반영됐는데**, 해제 evidence에는 `"a fill committed after the local read"`가 영속됩니다.

   근거: [reconcileloop.go:435](/home/daniel/tossos-work/a201-impl/internal/app/engine/reconcileloop.go:435)의 두 읽기는 원자적이지 않습니다. [a201_fill_credit.go:65](/home/daniel/tossos-work/a201-impl/internal/reconcile/a201_fill_credit.go:65)는 id만 비교하고, [mismatch.go:952](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:952)는 더 강한 시간적 사실을 기록합니다.

   **공시 경계:** 부여 자체는 V2의 닫힌 조건과 일치합니다. 그러나 V8도 ‘로컬 읽기 뒤 커밋까지는 증명한다’고 설명하므로, 이 **시간 순서 오기록은 공시되지 않았습니다.** 같은 as-of 원인 이름 모호함과도 다릅니다. 근거 문구를 ‘로컬 읽기 전 표식 이후 커밋’으로 제한해야 합니다.

2. **P3 — 정상 경로의 체결 credit 반박 시험이 반박 누락을 놓칠 수 있음**

   대상: [a201_fill_credit_cycles_test.go:298](/home/daniel/tossos-work/a201-impl/internal/app/engine/a201_fill_credit_cycles_test.go:298).

   현재 시험 입력은 다음과 같습니다.

   - k: broker 0 / local 31 → 로컬 읽기 뒤 SELL 31 커밋 → 수렴 미반영 → 체결 credit.
   - k+1: broker 5 / local 0 → 외부 포지션 ingest가 로컬을 5로 만듦 → Observe는 외부 분쟁으로 대기.
   - k+2: broker 0 / local 5 → 수렴이 0으로 조정하고 **새 조정 credit으로 덮어씀**.
   - k+3: broker 0 / local 0 → `ADJUSTMENT_APPLIED` 해제.

   따라서 [mismatch.go:853](/home/daniel/tossos-work/a201-impl/internal/reconcile/mismatch.go:853)의 정상 경로에서 `disputed[symbol]`에 의한 삭제를 제거해도, 기존 체결 credit은 k+2의 조정으로 덮어써집니다. 이 시험의 기대 결과는 그대로 나올 수 있습니다. **변이 실행 결과가 아닌 정적 추론입니다.**

   반박을 직접 측정하려면 k+2에도 **broker 5 / local 5**를 유지하고, 새 체결·조정 없이 차단이 남는지 확인해야 합니다.

   **공시 경계:** V8·§4 예외가 아니라 검증 공백입니다. 앞선 진행 설명을 정정하면, credit을 덮어쓰는 주체는 ingest가 아니라 **그다음 주기의 convergence**입니다.

3. **P3 — 종목별 판독 실패의 ‘전 종목 부여 취소’ 조건을 시험하지 않음**

   [a201_fill_credit_test.go:104](/home/daniel/tossos-work/a201-impl/internal/reconcile/a201_fill_credit_test.go:104)의 실패 시험은 단일 종목과 닫힌 장부를 사용합니다.

   빠진 입력:

   - k: A·B 모두 수량 불일치, 표식 이후 각 체결 커밋, 수렴은 둘 다 미반영, Observe 성공.
   - A의 watermark 읽기는 성공하여 후보 목록에 추가.
   - B의 watermark 읽기만 실패.
   - k+1: A·B 모두 새 비교에서 일치.

   계약상 k에서 **어느 종목에도 credit을 주면 안 됩니다.** 현재 [a201_fill_credit.go:63](/home/daniel/tossos-work/a201-impl/internal/reconcile/a201_fill_credit.go:63)는 올바르게 `nil`을 반환하지만, 이를 `return out`으로 바꿔 A에 부여하는 결함은 현재 단일 종목 실패 시험으로 구별하지 못합니다.

   **공시 경계:** 판독 실패는 공개된 미부여 조건입니다. 구현 결함이 아니라 해당 조건의 검증 공백입니다.

M14·M15의 최종 `CAUGHT` 기록에서는 다른 가드에 우연히 걸린 오류를 확인하지 못했습니다. M14는 `orderTerminalCandidate` 진입으로 생긴 `quantity_dispute_seen` 사유를 검출하므로 **분기·진단 무변성** 증거입니다. M15는 §4에 공개된 **같은 문자열 as-of의 원인 이름**을 측정합니다. 둘 다 잘못된 해제 방지 전체를 입증하지는 않습니다.

HEAD `d31347ab8`과 지정 Go diff, 설계·spec, 관련 소스·시험 및 기존 변이 증거를 읽었습니다. 부여·Observe·Refresh의 mutex와 두 authority-error 반환 경로를 확인했지만, 동시 실행 인터리빙의 안전성을 실행으로 검증하지는 않았습니다. 파일 변경·빌드·시험·네트워크·운영 명령은 실행하지 않았습니다. `impeccable: N/A — 백엔드 읽기 전용 리뷰`.

---

<!-- 리뷰어(Opus) 주: 위는 codex 최종 답 원문 그대로(전사 7861 줄 중 7740-7797). 실행: `codex exec --sandbox read-only -C /home/daniel/tossos-work/a201-impl "<prompt>" </dev/null`,
OpenAI Codex v0.165.0, model gpt-6-astra, reasoning effort medium, session 01a20b98-9eab-7592-ab7b-6d0e1d85a6bc, exit 0, tokens 117,030.
전사 전체(파일 덤프 포함 500K)와 프롬프트는 세션 스크래치에만 있음 — 
/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/6ba8fcd4-b024-437e-a4c2-ebe621826fe5/scratchpad/codex-r1.txt ·
/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/6ba8fcd4-b024-437e-a4c2-ebe621826fe5/scratchpad/codex-prompt.txt
(프롬프트 요지: 대상 파일 · base..HEAD · 초점 1~9 · 「파일 변경 금지 · 반례 입력을 구체적으로」 · go test 금지). -->
