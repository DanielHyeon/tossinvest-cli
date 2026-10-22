## ADDED Requirements

### Requirement: breakout universe variants are measured against the a140 baseline without changing the signal

A universe-variant shadow measurement SHALL evaluate breakout v1 with the same metric version, the same entry and exit
labelling and the same round-trip cost model as the a140 shadow measurement, and SHALL differ from it only in which symbols
are selected. It SHALL record three universes in the same KR session — traded-value top 30, traded-volume top 30 and a
computed volume-increase top 30 — after the common filters (investment-caution excluded, price at least 1,000 KRW and at most
the order cap). It SHALL record each symbol's broker security type as a receipt in the ledger and report every universe split
into stock, ETF and ETN branches as mapped in the design, with other and unclassified types reported as counts only. It SHALL withhold win-rate and mean-R verdicts for any universe and branch cell below 30
labelled samples (MUST NOT judge), and it SHALL NOT place orders, read accounts, read or write signed manifests, or call an
external judgment API.

#### Scenario: same signal, different symbols

- **WHEN** a symbol is selected into any of the three universes on a measured session
- **THEN** its setup, phase transitions and label are byte-identical to what the a140 evaluator produces for that symbol and those bars

#### Scenario: ETF results are separated, not removed

- **WHEN** a selected symbol's recorded security type is `ETF`
- **THEN** it stays in its universe and its results are counted in that universe's ETF branch, never in the stock branch

#### Scenario: small cells are withheld

- **WHEN** a universe and branch cell has fewer than 30 labelled shadow proposals across the measured sessions
- **THEN** the report shows its funnel and outcome counts but withholds the win-rate and mean-R verdict
