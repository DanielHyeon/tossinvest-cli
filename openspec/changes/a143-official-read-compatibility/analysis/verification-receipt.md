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
