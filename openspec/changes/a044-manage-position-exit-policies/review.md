# Review: a044-manage-position-exit-policies

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability, UI/UX

## Findings and decisions

1. generation-scoped override/release/re-adopt와 global desired setting을 분리한다.
2. console/httpapi는 journal을 writable로 열지 않는다. durable command seam을 통해 engine이 version/generation/approval을 재검증하고 journal의 유일한 trading-state writer로 남는다.
3. release는 active exit/protection 충돌을 검사하고 3초+checkbox/button을 사용한다. typed phrase, free reason, 숫자·symbol 입력은 없다.
4. stale CAS는 412로 끝내며 자동 retry나 partial write를 금지한다.

## Verification evidence

- OpenSpec strict validation: pass.
- Journal migration v11 reserved; downgrade는 backup restore/reconcile 절차를 따른다.

## Verdict

a042 이후 domain/command seam 구현을 승인한다. 실제 보호 약화나 LIVE 권한 승인은 포함하지 않는다.
