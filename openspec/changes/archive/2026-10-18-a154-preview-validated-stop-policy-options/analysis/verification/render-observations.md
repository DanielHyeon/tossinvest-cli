# Isolated synthetic UI observations

Observed the actual dashboard template through a temporary httptest server, not an operational service. The server was shut down after inspection. No live account, market data, configuration, or orders were accessed.

- Desktop: submitted manual KR inputs entry 70000, low 69000, ATR 500, tick 100, grid 60000–80000, budget 10000, share cap 100 and notional cap 7000000. Three displayed stop/quantity/risk results matched 69500/20/10000, 68900/9/9900 and 69200/12/9600.
- Desktop: submitted US inputs entry 100, low 98, ATR 2, tick 0.01, grid 1–200, budget 100, share cap 1000 and notional cap 100000. Results matched 99.3/142/99.4, 97.99/49/98.49 and 97/33/99.
- At viewport 390×844, document width was 375 versus viewport 390; result table scrolled horizontally (314 client width, 581 scroll width). Keyboard Tab moved from entry to low and focus outline was solid 3px.
- A missing market produced aria-invalid=true on the market select and aria-describedby pointed to its error message. Browser warning/error log was empty.
- Screenshots were inspected in the browser tool output but were not saved to files; no screenshot artifact is claimed. Observations precede the final one-line unknown-currency display clarification; final targeted rendering tests cover the current template.

The source-only impeccable scan recorded in impeccable-source-scan.log reports two pre-existing side-border CSS matches. Its output states parsed-page rules did not run; it is not a complete rendered-page audit.

Manager independently inspected the synthetic US result and identified nested main landmarks. The template was corrected; main-landmark-red.log captures the failing regression before the fix and focused-final6.log captures the passing current targeted suite. The browser observations above were made before that structural correction.
