# Implementation sequence note

The intended RED-first sequence could not be completed before the initial helper
skeleton was added: the frozen design's Git fixture harness did not yet exist.
This is a process deviation, not a claim that the skeleton is correct.

The first durable execution was the pre-existing ordinary checker suite:
`/tmp/a120-targeted-exit.txt` records exit `0`. The initial helper unit suite is
also recorded separately at `/tmp/a120-execution-baseline-tests-exit.txt` with
exit `0`; it proves only its six narrow unit cases.

Before additional helper behavior is added, task 2.1 must add an isolated Git
fixture that demonstrates the current skeleton fails to reject both build-input
drift and an E-to-S ledger inventory mismatch. Those observed failures are the
retrospective RED evidence for the omitted contract behavior. The implementation
is not ready for adversarial review until that fixture becomes green after the
minimal corrections.
