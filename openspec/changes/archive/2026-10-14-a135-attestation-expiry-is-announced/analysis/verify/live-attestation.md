# 운영 attestation · 갱신 타이머 실측 (2026-10-14, 읽기 전용)

잰 순간: `2026-10-13T20:57Z`(KST 10-14 05:57) · 모집단: 호스트 `~/.config/tossctl/capability-attestation.json` 한 파일(컨테이너의
`/var/lib/tossos/config` 로 bind — `compose.yaml` 42~45). 읽은 칸: `format_version` · `issued_at` · `expires_at` · `soak_days` ·
`verified_by` · `endpoints` · `supervised_by[].{endpoint,at,market}`. **`account_ref` · `supervised_by[].account_ref` · `source` 는
읽지도 출력하지도 않음**(불변식 8). `.env` · 세션 · 토큰 무열람.

## attestation 메타데이터

| 칸 | 값 |
|---|---|
| format_version | 1 |
| issued_at | 2026-09-23T01:39:36Z |
| expires_at | **2026-10-23T01:39:36Z** (30일 유효) |
| 남은 시간(잰 순간) | 9일 4시간 42분 → 이 change 의 판정으로 **D-10, healthy**(D-7 경계 = 2026-10-16T01:39:36Z) |
| soak_days | 39 |
| verified_by | tossos-attest.timer |
| endpoints | GET accounts · buying-power · conditional-orders · conditional-orders/{id} · holdings · orders · orders/{id} · prices · sellable-quantity · **POST /api/v1/orders** · **POST /api/v1/orders/{id}/cancel** (11) |
| supervised_by | POST orders `2026-09-09T03:23:14Z` KR · POST cancel `2026-09-09T03:23:15Z` KR |
| 엔진 `RequiredEndpoints()` 여덟 대비 누락 | **0** — 지금 재시작하면 attestation 조항은 통과 |
| 이 change 의 판정이 낼 `supervised_proof_at` | 2026-09-09T03:23:14Z |

## 갱신 타이머

- 설치된 단위(`systemctl --user cat`): `OnBootSec=10min` · `OnUnitActiveSec=6h`. `ExecStart` 는 `soak attest` 실패를
  `|| printf "attest not earned yet"` 로 **성공으로 가림**(저장소 템플릿 `deploy/systemd/tossos-attest.service` 의
  `--record-renewal-status` 판본은 미설치 — a063 4.2 미완). 그래서 `journalctl` 은 매번 `Finished`, renewal-status 파일 없음.
- `~/.local/share/tossos/attest-autorun.log` 끝 20줄: 10-13 11:38 · 17:42 · 23:45 · 10-14 05:49 KST 네 번 모두
  `soak: the capability soak is not complete — 7 cycle(s) inside the window failed the read-completeness check
  (0 quote(s) returned for 1 symbol(s); the CLOSED walk stopped at the 25-page limit …)` → **갱신 거절, 파일 무변경**.
- 결과: 마지막 성공 갱신은 09-23 이후 없음(발급 시각 = 09-23).

## 갱신이 성공하면 POST 를 유지하는가 — **아니오**(2026-10-09 이후)

AST 열거(`analysis/code-context/`, 손으로 읽지 않음):

- `soak.acceptSupervised`(`internal/soak/attest.go:301-343`) B6 `:332` — `age < 0 || age >= validity` 면 `continue`(거절 아님, 조용히 건너뜀).
- `soak.BuildAttestation`(`:220-273`) 의 거절 분기는 B1 `:225`(soak 미완) · B3 `:235`(GET 아닌 soak endpoint) · B4 `:243`
  (acceptSupervised 오류) 셋뿐 — **빈 감독 증거를 거절하는 분기가 없음**.
- 기존 시험 `TestSupervisedEvidenceOlderThanTheValidityIsNotEvidence`(`internal/soak/supervised_test.go:114`)가 바로 이 동작을
  실행으로 고정함: 기간 밖 증거 → `BuildAttestation` 성공 · POST 둘 누락 · `SupervisedBy` 빈 값. 이번 실행 rc 0.
- 유효 기간 = `DefaultCriteria().Validity` 30일(`internal/soak/attest.go:92`), 감독 증거 09-09T03:23Z + 30일 = **2026-10-09T03:23Z**.

그러므로 지금 soak 이 회복되어 타이머가 갱신하면 **GET 만 담은 attestation 이 현재의 좋은 파일을 덮어쓰고**, 그 뒤 재시작은
`ErrEndpointNotAttested`(누락: POST orders · POST cancel)로 거절된다. 이 change 의 감시는 그 덮어쓰기를 1시간 안에 `scope_lost`
critical 로 알린다. 막지는 않는다(갱신 명령 · 기준은 a063 · verify-execution-capability 의 범위).

## 사람이 할 일(이 change 밖)

10-23T01:39Z 전에 둘 다 필요: (1) 감독하 실주문 검증 재실행(POST 증거 갱신 — 사람 승인 실주문), (2) soak 판독 완전성 회복 뒤 갱신.
순서가 (2)→(1) 이면 그 사이 attestation 이 POST 를 잃는다 — (1) 을 먼저.
