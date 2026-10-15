# a143 focused verification receipt

Product freeze: `60f1d04166c1b00818010fc777ba0ffa50b0ff77`.

## RED and GREEN

- `/tmp/a143-red.log`: exit 1 before implementation. It recorded both intended failures: newest-first general candles and KR `0101N0` refusal.
- `/tmp/a143-green-official-final.log`: exit 0 for `TestAdaptCandles*`, `TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol`, `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest`, and `TestStrictMinuteCandlesAcceptsTheKoreanMarket`.
- `/tmp/a143-green-officialbars.log`: exit 0 for `TestPollBindsEachSuccessorToTheNextNewerObservedBar` and `TestPollConvertsTheBrokerCloseLabelIntoAnOpenInstant`.

No full suite claim is made here. Raw strict page ordering, cursors, close labels, officialbars conversion, US lexical policy, runtime settings, orders, tokens, journals, and risk paths were not edited.

## Independent mutation evidence

An external detached worktree at `/tmp/a143-mutant` was used; the product worktree was untouched.

- Removing the general local reversal made `TestAdaptCandlesOrdersNewestFirstWireChronologically` fail (exit 1; `/tmp/a143-mutant-reversal.log`). The three-item assertion detects reversed order plus all timestamp and OHLCV values.
- Restoring the prior KR numeric-only predicate made `TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol` fail with `SYMBOL_INVALID` (exit 1; `/tmp/a143-mutant-kr.log`).

## Public snapshots

REST source: `https://openapi.tossinvest.com/openapi-docs/latest/openapi.json`; local stored document is version 1.2.24 with 38 paths. Its `jq -S .` SHA-256, including the trailing newline emitted by jq, is `91f9e02f79df606f9a02db23f391ff993f8e634417a57fd2f5ef9d9d188308bd`.

AsyncAPI source was independently fetched from `https://openapi.tossinvest.com/openapi-docs/latest/asyncapi.json` at `2026-10-15T16:44:26Z`; HTTP status was 200. The response byte-compares equal to `docs/migration/asyncapi.latest.json`, reports version 1.2.2 and 4 channels, and has `jq -S .` SHA-256 `176595a3abbba3c1b84576e3c1b6806b75ec6d507a2efee26a24021a55531a63` with the jq trailing newline. AsyncAPI is snapshot evidence only; automated monitoring and WebSocket client implementation remain unimplemented.

## Final strict-invariant focused pass

Test-only commit e6718bc1 adds the pure seven-character KR refusal case. `/tmp/a143-focused-adapt.exit`, `/tmp/a143-focused-strict.exit`, and `/tmp/a143-focused-officialbars.exit` are each 0. The strict command selects all TestStrictMinuteCandles tests, including the existing two-candle canonical query/page/cursor regression, observer chaining, malformed-body and transport refusals. This establishes the named fixture results cited in the final maps; it is not instrumented branch coverage. Manager independently inspected the raw exit files. Product and snapshot bytes remain unchanged.

## Aggregate-gate regression correction (2026-10-16 KST)

The first aggregate gate is retained as a failure, not a completion receipt: `/tmp/a143-full-final/make-gate-a143.exit` is `2`. Stages 1–6 passed; stage 7 (`make test`) failed and stages 8–11 did not run. The raw log records two independent failures: `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest/kr_symbol_with_letters` sent two requests because strict quote shared the widened minute predicate, and `TestWireBodyFitsTheOpenAPIRequestSchema` retained the old published symbol-description expectation.

The correction restores shared `strictMinuteCheckSymbol` to numeric-only KR validation and gives `StrictMinuteCandles` a minute-only uppercase-alphanumeric wrapper. The original quote-letter refusal remains; an added `0101N0` fixture requires both strict quote readers to return `StrictQuoteReasonSymbol` with zero requests, while the existing minute fixture accepts the exact same request symbol. Only the a087 test assertion now recognizes the verified published description; its executable numeric-only order policy is unchanged.

- `/tmp/a143-full-final/fix-red-quote.exit`: `1`; before the wrapper, the new `0101N0` quote case reached the broker twice.
- `/tmp/a143-full-final/fix-green-minute-final.exit`: `0`; all `TestStrictMinuteCandles*` fixtures passed.
- `/tmp/a143-full-final/fix-green-quote-final.exit`: `0`; the quote refusal fixture passed, including both exact symbol-reason assertions and zero-request check for `0101N0`.
- `/tmp/a143-full-final/fix-green-a087-final.exit`: `0`; the schema-contract fixture accepted the refreshed public description.

`/tmp/a143-full-final/fix-check-analysis-current.exit` records the expected stale-landing refusal after refreshed current AST evidence. The existing landing record must be removed in a commit and recorded again after this correction is committed; no landing value is hand-written.
