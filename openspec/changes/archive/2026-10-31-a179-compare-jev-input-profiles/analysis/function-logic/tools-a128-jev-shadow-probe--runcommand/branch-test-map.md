# Branch Test Map: `runCommand`

Relevant package test surfaces: TestProfileFlagsRejectedBeforeCredentials, TestAutomaticProfileOutputRouting, TestRunRefusesToStartWithoutTheAPIKey. These tests cover named scenarios, not every AST branch; rows below identify unproven individual branch coverage rather than asserting it. Existing socket-backed tests could not execute in this sandbox.

| Branch | Source branch | Individual branch coverage |
|---|---|---|
| B1 | if at 92:2: `if err := flags.Parse(args); err != nil {` | Not individually established |
| B2 | if at 97:3: `if f.Name == "out" {` | Not individually established |
| B3 | if at 102:2: `if err != nil {` | Not individually established |
| B4 | if at 107:2: `if err != nil {` | Not individually established |
| B5 | if at 111:2: `if err != nil {` | Not individually established |
| B6 | if at 114:2: `if err := validateRunFlags(*top, *pool, *horizon, *interval); err != nil {` | Not individually established |
| B7 | if at 119:2: `if strings.TrimSpace(*sessionWindow) != "" {` | Not individually established |
| B8 | if at 121:3: `if err != nil {` | Not individually established |
| B9 | if at 128:2: `if err != nil {` | Not individually established |
| B10 | if at 132:2: `if err != nil {` | Not individually established |
| B11 | if at 135:2: `if credentials == nil {` | Not individually established |
| B12 | if at 144:2: `if err != nil {` | Not individually established |
| B13 | if at 155:2: `if err != nil {` | Not individually established |
| B14 | if at 159:2: `if err != nil {` | Not individually established |
| B15 | if at 165:2: `if *profile != "current" {` | Not individually established |
| B16 | if at 167:3: `if _, err := rand.Read(token[:]); err != nil {` | Not individually established |
| B17 | if at 179:2: `if err != nil {` | Not individually established |
| B18 | if at 191:2: `if manualSess != nil {` | Not individually established |
