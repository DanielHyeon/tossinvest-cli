# gstack review r2 — codex 바깥 목소리(원문, 2026-11-24)

- 명령: `codex exec -s read-only -C ~/tossos-work/a198-impl -o <out> "<prompt>" </dev/null`(rc 0), codex-cli 0.165.0. 프롬프트 원문: 같은 디렉터리 `gstack-r2-codex-prompt.txt`. 기준 `e1eaf720c..fa8771ed3`, Go 만.

## 원문(마지막 메시지)

검토 기준: `e1eaf720c..fa8771ed3`, Go 변경과 직접 호출 경로 정적 검토. 테스트 실행·파일 수정·Git 상태 변경 없음.

1. **[P1] 취소 실패 로그가 다른 종목의 손절까지 막음.**
   a198_clear_cancel_log.go:57 의 `logger.Warn`은 동기 쓰기이며 시간 제한이 없습니다. 운영에서는 `errOut`에 연결됩니다(`cmd/tossctl/engine.go:218,680`).
   앞 포지션에서 취소 결정 발행 실패 → `exitloop.go:1710` → 로그 파이프 포화·저장장치 쓰기 지연이 발생하면, 직렬 관측 루프가 뒤 포지션의 손절 판정(`exitloop.go:541`)에 도달하지 못합니다. `recover`는 멈춘 쓰기를 해제하지 않습니다. 기존 청산 로그는 같은 주기의 모든 제출 뒤에 쓰도록 지연했지만, 신규 호출은 그 보호를 우회합니다.
   취소 로그도 판정·제출 이후로 미루고, 차단되는 writer에서 뒤 포지션의 제출 여부를 검증해야 합니다. Severity: High; Blocks: yes

2. **[P2] 로그 쓰기 실패를 성공으로 기억해 장애 복구 후에도 기록을 누락함.**
   a198_clear_cancel_log.go:72 는 `Warn` 반환만으로 중복 억제 표시를 설정합니다. 그러나 `obs.Logger.Warn`은 오류를 반환하지 않으며(`internal/obs/log.go:159,203`), 내부 `slog`도 handler 오류를 버립니다.
   첫 writer 호출이 패닉 대신 `ENOSPC`·`EIO`를 반환하면 줄은 저장되지 않았는데 표시가 남습니다. writer가 복구돼도 같은 주문·상태·사유의 실패는 39행에서 계속 생략됩니다. 첫 쓰기만 오류를 반환하고 이후 성공하는 writer로 재현 가능한 운영 증거 유실입니다.
   쓰기 성공을 확인한 경우에만 표시를 확정하거나 실패 기록을 재시도해야 합니다. Severity: Medium; Blocks: yes

Recommendation: 수정 후 재검토 because 취소 실패 로그의 동기 쓰기가 다른 종목의 손절 제출까지 무기한 막을 수 있습니다.
VERDICT: FINDINGS
