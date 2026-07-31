# CodeGraphContext advisory context

`make sdd-sync` refreshed CodeGraphContext before implementation. Its advisory
context agreed that the console owns HTTP gating while `cmd/tossctl` owns path,
credential, official-client, audit, and process seams.

GBrain recall was attempted for `openapi login console soak credentials` but
the local service was busy. This is advisory-only and does not block hard
evidence from the current HEAD, OpenSpec, CodeGraph, Go AST, and tests.
