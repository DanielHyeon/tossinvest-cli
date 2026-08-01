# Review: a051-add-httpapi-daemon

- Date: 2026-08-01
- Voices: Security, Test Architecture, Maintainability, API Contract

## Security decision

no-token private REST/SSE는 read-only로 승인한다. a051 daemon write는 local human approval channel의 one-time 최대 60초 signed capability만 인증 수단으로 허용한다. shared mutation guard의 browser session+CSRF+Origin 및 enrolled mTLS 지원은 보존하지만 daemon에는 wiring하지 않는다. remote LIVE/gate/kill/protection weakening/activation-manifest routes는 존재하지 않는다.

## Findings and decisions

1. SSE ID는 process epoch+sequence이며 restart/gap/unknown Last-Event-ID는 full snapshot으로 수렴한다.
2. limits는 32 clients, queue 64, heartbeat 15초, queue-full disconnect, body 256KiB, header/read timeout 5초다.
3. idempotency scope는 actor/client+method+resource+canonical body digest+key이며 같은 key의 다른 body는 409다.
4. console origin logic을 shared `internal/networkboundary`로 추출해 strict Origin precedence, opaque-origin refusal와 exact trusted proxy hop을 보존한다.
5. httpapi는 journal/broker writer가 아니며 shared operator view와 narrow commander만 사용한다.
6. command semantic refusal는 검증된 400/403만 stable error로 저장·재생하고, 불명확한
   오류는 503으로 fail closed한다. audit completion 재시도 command를 재실행하지 않는다.
7. public-key와 security DB 경로는 Unix에서 component별 no-follow, root/service ownership,
   전용 `0700` DB 디렉터리를 강제한다. Windows의 ownership/mode 제한은 운영 잔여 위험으로 남긴다.
8. outer boundary refusal도 router와 동일한 stable JSON/no-store contract를 쓴다.

## Verification evidence

- OpenSpec strict validation: pass.
- Full `go test ./... -count=1`: pass.
- Focused race, `go vet`, Windows amd64 cross-compile: pass.
- `internal/httpapi` coverage 79.4%, `internal/networkboundary` coverage 71.5%.
- API lint 97.4/100, 0 error/warning; breaking self-check 0.
- API scorecard 75.0/C는 승인된 B-waiver다. pagination/filter/search/field-selection/batch는
  no-arbitrary-input과 bounded private read contract을 어기므로 점수를 위해 추가하지 않았다.
- Security reviewer: APPROVE, Critical/High/Medium 0.
- Test/Maintainability reviewer: APPROVE, blocker 0; security/idempotency complexity extraction은 non-blocking follow-up.
- Real VPN/TLS topology evidence: deployment environment에서 별도 검증 필요.

## Verdict

read-only daemon과 capability-only limited-write contract 구현을 승인한다. 실제 배포는 private
bind/TLS/proxy proof와 pending command 0 rollback check를 요구한다.
