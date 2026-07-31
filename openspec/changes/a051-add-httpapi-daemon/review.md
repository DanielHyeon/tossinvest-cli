# Review: a051-add-httpapi-daemon

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Security decision

no-token private REST/SSE는 read-only로 승인한다. write는 browser session+CSRF+Origin 또는 enrolled mTLS/one-time 60초 signed capability로 제한한다. remote LIVE/gate/kill/protection weakening/activation-manifest routes는 존재하지 않는다.

## Findings and decisions

1. SSE ID는 process epoch+sequence이며 restart/gap/unknown Last-Event-ID는 full snapshot으로 수렴한다.
2. limits는 32 clients, queue 64, heartbeat 15초, queue-full disconnect, body 256KiB, header/read timeout 5초다.
3. idempotency scope는 actor/client+method+resource+canonical body digest+key이며 같은 key의 다른 body는 409다.
4. console origin logic을 shared `internal/networkboundary`로 추출해 strict Origin precedence, opaque-origin refusal와 exact trusted proxy hop을 보존한다.
5. httpapi는 journal/broker writer가 아니며 shared operator view와 narrow commander만 사용한다.

## Verification evidence

- OpenSpec strict validation: pass.
- Real VPN/TLS topology evidence: deployment environment에서 별도 검증 필요.

## Verdict

read-only daemon과 authenticated limited-write contract 구현을 승인한다. 실제 배포는 private bind/TLS/proxy proof와 pending command 0 rollback check를 요구한다.
