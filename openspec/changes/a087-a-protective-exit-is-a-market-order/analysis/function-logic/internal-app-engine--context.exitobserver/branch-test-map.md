# Branch Test Map: `Context.ExitObserver`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil Context | existing wiring test | 기존 green | unavailable |
| B2 | unverified gate | existing wiring test | 기존 green | unavailable |
| B3 | wrong Guardian | existing wiring test | 기존 green | unavailable |
| B4 | names nil | existing wiring test | 기존 green | default |
| B5 | notifier present | existing wiring test | 기존 green | record-only |
| B6 | floor nil | a087 board ownership | 작성 전 | Context board wins |
