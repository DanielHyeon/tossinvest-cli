# Remaining changes inventory (read-only, 2026-10-18)

Source: `openspec list --json`, each active `tasks.md`, proposal/design/review, and per-change `openspec status`. Base: `f156b66eaef9958ef7ef3cbfee5c2c6927fee1a8`. a157 belongs to separate Claude work and is excluded. Existing a128 evidence is untouched. Counts are checked tasks, not proof of implementation or operational safety.

| Change | Checked | Next work and boundary |
|---|---:|---|
| a155 | 0/10 | a149/150/153/154 archived. Offline isolated observation and replay are candidates after Manager freezes revised contract; real observation requires separate session approval. |
| a156 | 0/12 | Depends on a155. Offline fixtures can be developed after approved design; tasks 4.1–4.2 require separate human approval and real observation/live evidence before archive. |
| a063 | 12/17 | Service installation/reload, approved three-day survey and fresh attestation remain; explicit human approval and operational window required. |
| a074 | 54/55 | Post-deployment real cycle event measurement remains. |
| a087 | 15/54 | Market-exit work and tests remain; its approved phase ordering requires KR market-sell and session-boundary measurement before the market-order phase lands. |
| a100 | 30/88 | M-A conditional-order trigger evidence and resulting status table task 0.11 precede downstream integration; depends on live capability/trigger measurement. |
| a105 | 0/13 | a100 land and flat-position residual-order window design precede wiring; human approval needed to enable Wired. |
| a107 | 0/13 | a100 land precedes removal of duplicate protection core and its regression gates. |
| a128 | 13/19 | KR/US observation sessions, calibration and human decision remain; preserve existing evidence and obtain required observation budget approval. |
| a140 | 15/21 | Human call-budget approval, KR observation, report and human decision remain. |
| verify-execution-capability | 8/22 | Credential soak, rate/capability evidence, real-account order and conditional-order observations and human review remain. |
| verify-observes-the-trigger | 38/43 | Explicit approval, exclusion setting, pause of other observation, US one-share trigger and residual reconciliation remain. |

All changes report proposal/spec/tasks artifacts `done`; this means artifact creation, not finished checkboxes. `design=ready` for a105/a107/verify-execution-capability is an optional artifact state, not completion evidence. The a155/a156 historical shared design-review record does not establish current implementation review. No live operations, API collection or orders were performed for this inventory.

Sequence: a155 offline first; then a156 offline work while retaining its live gate. In the separate protection chain, approved real evidence precedes verify-observes-the-trigger and a100; a105/a107 follow a100. The a087 market-order landing retains its explicit market-sell measurement prerequisite. Exact concurrent dependencies require fresh Manager design review.
