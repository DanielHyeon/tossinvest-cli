# a143 preimplementation evidence

Current HEAD `6765858af1beb472268cb699efa1964addb84d60` equals intake HEAD. Relevant current SHA-256 values match intake: `candle_reads.go` `c22b069c…`, `strict_minute_candles.go` `d32181a…`, `officialbars/producer.go` `8d45ca93…`.

Generated AST/FLM/BTM paths:

- `function-logic/internal-official--adaptcandles/`
- `function-logic/internal-official--client.strictminutecandles/`
- `function-logic/internal-official--strictminutechecksymbol/`

No CodeGraph or GBrain action occurred: shared indexing has one assigned writer. No immutable baseline capture, production/test edit, test execution, or commit occurred.

| public local document | version | count | SHA-256 over `jq -S .` output with trailing newline |
|---|---:|---:|---|
| `/tmp/toss-openapi-live-20261015.json` | 1.2.24 | 38 paths | `91f9e02f79df606f9a02db23f391ff993f8e634417a57fd2f5ef9d9d188308bd` |
| `/tmp/toss-asyncapi-live-20261015.json` | 1.2.2 | 4 channels | `176595a3abbba3c1b84576e3c1b6806b75ec6d507a2efee26a24021a55531a63` |

Existing snapshot metadata remains REST-only at 1.2.4. Workflow is untouched; AsyncAPI automatic monitoring is not implemented. `officialbars` is not an edit target: it needs newest-first strict pages and alone converts close labels to open time.
