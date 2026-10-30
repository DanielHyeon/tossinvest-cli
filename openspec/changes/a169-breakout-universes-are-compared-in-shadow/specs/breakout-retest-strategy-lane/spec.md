## ADDED Requirements

### Requirement: breakout universe variants are measured against the a140 baseline without changing the signal

A universe-variant shadow measurement SHALL evaluate breakout v1 with the same metric version, the same entry and exit
labelling and the same round-trip cost model as the a140 shadow measurement, and SHALL differ from it only in which symbols
are selected. It SHALL record three universes in the same KR session — traded-value top N and traded-volume top N selected
30 seconds after the regular open, and a computed volume-increase top N selected 30 minutes 30 seconds after the regular open
from the union of both rankings' top 100 — after the common filters (investment-caution excluded, price at least 1,000 KRW and
at most the order cap). N SHALL be a field of the universe policy, separate from the a140 universe size: 30, or 20 when the
rehearsal warns (human decision 2026-10-28), and a changed N SHALL be a different lot. The volume increase SHALL be the
ranking's traded volume divided by the sum of the previous 20 trading days' daily volumes (adjusted daily candles, only bars
stamped before 00:00 KST of the session date), and a candidate with fewer than 20 prior days or a zero baseline SHALL be
excluded with its reason recorded. Each universe SHALL be fixed once per session, recorded with its target and actual selection
time, its pre-filter ranking receipt and every candidate's exclusion reason, and restored from the ledger on restart. A setup
SHALL be attributed to a universe only when its breakout bar closed after that universe's selection time. It SHALL record each
measured symbol's broker security type as a receipt in the ledger and report every universe split into stock, ETF and ETN
branches as mapped in the design, with other and unclassified types reported as counts only. It SHALL withhold win-rate and
mean-R verdicts for any universe and branch cell below 30 labelled samples (MUST NOT judge), and it SHALL NOT place orders,
read accounts, read or write signed manifests, or call an external judgment API. Run without the universe mode, the tool SHALL
keep the a140 behaviour and the a140 lot digest unchanged.

In the universe mode the tool SHALL use only a dedicated Open API key held in its own profile directory (human decision
2026-10-28) and SHALL take the engine's profile directory as an explicit, operator-confirmed argument. Before any HTTP request,
evidence or ledger access it SHALL refuse to start when either directory is missing; when the engine directory does not hold the
engine's lock file, log file and credentials; when, after making both paths absolute and resolving symbolic links, the tool's
directory is or lies inside the engine's profile directory, the default config directory or the default token cache directory;
when the credential environment variables are set; when the credential file is not a regular 0600 file; when the engine's API key
cannot be read or equals the tool's; or when the tool's cached access token, compared after parsing, equals an engine access token
or the token file is a link to one. It SHALL read only the API key from the engine's credential file, never holding the engine's
secret as a credential value, and SHALL NOT print, log or record any key, secret, token or engine log line. It SHALL offer a
profile preflight that runs exactly these checks with no network, evidence or ledger access and then exits. While running it SHALL
follow the engine's log file read-only from its start offset, and on the first rate-limit line, or when the log cannot be read, is
rotated or truncated, or the follower stops, it SHALL cancel all further broker calls, record the session as aborted with the
reason, exclude that session from every lot total and refuse to restart the same session; failing to open the log SHALL refuse
the start. It SHALL refuse an a140
reconstruction size other than 30. The volume-increase universe SHALL be fixed before the deadline 10 seconds ahead of the first
minute poll after its selection time or not at all that session: when its baselines are not complete by the deadline, the
universe-session SHALL be recorded as late, kept out of the measured set and both lot totals, and the poll of the other universes
SHALL NOT be delayed.

#### Scenario: same signal, different symbols

- **WHEN** a symbol is selected into any of the three universes on a measured session
- **THEN** its setup, phase transitions and label are byte-identical to what the a140 evaluator produces for that symbol and those bars

#### Scenario: a breakout before the selection is not attributed

- **WHEN** a symbol enters the volume-increase universe at 09:30:30 and its breakout bar closed at 09:29
- **THEN** that setup is counted as "breakout before selection" for that universe and contributes no label to it, while it stays attributed to any other universe that selected the symbol earlier

#### Scenario: today's daily candle is not in the baseline

- **WHEN** the daily candles include a bar stamped 00:00 KST of the session date
- **THEN** that bar is excluded from the 20-day baseline, while the previous trading day's bar is included whether it is stamped at its own 00:00 KST, its open or its close

#### Scenario: ETF results are separated, not removed

- **WHEN** a selected symbol's recorded security type is `ETF`
- **THEN** it stays in its universe and its results are counted in that universe's ETF branch, never in the stock branch

#### Scenario: small cells are withheld

- **WHEN** a universe and branch cell has fewer than 30 labelled shadow proposals across the measured sessions
- **THEN** the report shows its funnel and outcome counts but withholds the win-rate and mean-R verdict

#### Scenario: the engine's key or profile is refused

- **WHEN** the universe mode is started without a separate profile directory, with the engine's profile directory, with the credential environment variables set, or with a profile whose key equals the engine's key
- **THEN** the tool exits before any broker call and before writing the ledger, naming the refused guard and the path only, with no key, secret or token value in its output

#### Scenario: the preflight proves the separate key without the network

- **WHEN** the profile preflight runs against a profile whose key equals the engine's key, or whose token cache holds the engine's access token re-serialised with different whitespace, field order or expiry
- **THEN** it refuses with the guard number and paths only and makes no HTTP request, and against a profile with a different key and token it reports success, still with no HTTP request, evidence or ledger file

#### Scenario: an engine rate-limit line stops the measurement

- **WHEN** a rate-limit line is appended to the engine's log file after the tool started
- **THEN** the tool makes no further broker call, records the session as aborted with reason `engine_rate_limited` and without the line's content, the report excludes that session from both lot totals, and a restart on the same day is refused before any broker call

#### Scenario: the a140 reconstruction stays at 30

- **WHEN** the universe mode is started with `--top` other than 30
- **THEN** the tool refuses to start before any broker call, while `--top` omitted or 30 starts with the A140 reconstruction at 30 and each measured universe at its policy N

#### Scenario: a late volume-increase baseline does not delay the poll

- **WHEN** the daily candles still missing at the volume-increase selection cannot all be read before 10 seconds ahead of the next minute poll
- **THEN** that session's volume-increase universe is recorded as late with its top-up counts, is not measured and is excluded from both lot totals, and the next minute poll of the other universes runs on time without a gap row

#### Scenario: the default mode reproduces the a140 lot

- **WHEN** the tool's effective policy is built in the default mode from the parameters recorded in the archived a140 ledger
- **THEN** its digest equals the a140 ledger's recorded policy digest, and the universe mode's digest differs from it
