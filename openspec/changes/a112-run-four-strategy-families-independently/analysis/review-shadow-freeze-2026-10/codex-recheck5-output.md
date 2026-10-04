Codex · GPT-6(세부 버전 비공개) · 「~/.codex 미접근」

**FAIL — 기존 항목 CLOSED, 신규 P0 0건·P1 1건.**

| 항목 | 판정 | v3.2 좌표·근거 |
|---|---|---|
| record 전 오류·panic 후 잔존 | CLOSED | §5:124–131, 즉시 삭제 실측 |
| 늦은 in-flight 재게시 | CLOSED | §5:127–131, epoch CAS 차단 실측 |
| 비차단 노트 셋 | CLOSED | §4:91–94, §5:136–143 |
| defer·CAS 잠금 순서 | 추가 P0/P1 없음 | §5:119–133, 상세 보고서에 검증 한계 명시 |
| 나이·만료 경계 | CLOSED | §5.1:159–171, 등호 거부·−1ns 허용 |
| 건강한 주기의 SHADOW 연속성 | **P1 OPEN** | §5.1:168이 최신-wave 규칙과 충돌 |

**신규 P1:** 정상 주기도 record에서 wave가 증가하면 이전 SHADOW가 무효화됩니다. 새 비동기 관측이 게시되기 전까지 UNOBSERVED가 됩니다. **74초 상한은 이 공백을 막지 못합니다.** “끊기지 않음”을 “나이 조건만으로 거부하지 않음”으로 좁혀야 합니다.

최소 모델 `go test -race` 8개 PASS, 반례 재현 포함. 생산 통합시험은 아닙니다. 모든 셸 첫 줄 규칙 준수. 실저장소 시작·종료 status와 HEAD 동일.

[상세 판정·좌표·시험·한계](/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r5/_codex/recheck5/report.md)