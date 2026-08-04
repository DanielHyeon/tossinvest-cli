# Function Logic Map: `riskbucket.bindProductionRiskInputs`

The function validates a sealed cap-free proposal, binds lane/sector/policy/FX evidence and creates
the five-bucket reserve policy. Its current price branch copies `PriceMinor()` into
`WorstExecutableQuote`, while the monetary calculator multiplies quote major currency by FX into
account-base minor risk. KR scale 0 hides the error; US scale 2 is overstated by 100x and exhausts
every bucket before first-leg admission.

Change: require `PriceProvenance.MajorDecimal()` and use that canonical decimal for monetary risk.
The sealed terms identity and the original minor value remain in the production digest, preserving
exact source-unit lineage without float conversion.

# Branch Test Map

| Market | sealed price | risk quote | Expected |
|---|---:|---:|---|
| KR scale 0 | `100` minor | `100` KRW | unchanged |
| US scale 2 | `10000` minor | `100` USD | no 100x overstatement |
| non-canonical/overflow conversion | any | none | fail closed |
| currency/scale/source mismatch | any | none | existing refusal unchanged |
