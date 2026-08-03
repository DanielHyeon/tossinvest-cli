# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil context fallback | cmd/tossctl console tests | existing green | existing green |
| B2 | invalid remote access options | cmd/tossctl console tests | existing green | existing green |
| B3 | KR verification record resolution error | cmd/tossctl console tests | existing green | existing green |
| B4 | US verification record resolution error | cmd/tossctl console tests | existing green | existing green |
| B5 | soak record resolution error | cmd/tossctl console tests | existing green | existing green |
| B6 | attestation path resolution error | cmd/tossctl console tests | existing green | existing green |
| B7 | Open API seam construction error | cmd/tossctl console tests | existing green | existing green |
| B8 | journal path unavailable degrades visibly | dashboard command tests | existing green | existing green |
| B9 | journal path available | performance/optimization tests | existing green | existing green |
| B10 | performance DB unavailable | performance console tests | existing green | existing green |
| B11 | performance DB available | performance console tests | existing green | existing green |
| B12 | optimization commander unavailable | optimization console tests | existing green | existing green |
| B13 | optimization commander available | optimization console tests | existing green | existing green |
| B14 | engine journal directory resolves | engine console tests | existing green | existing green |
| B15 | engine journal directory fails | engine console tests | existing green | existing green |
| B16 | container disables system update | system-update tests | existing green | existing green |
| B17 | non-container update path branch | system-update tests | existing green | existing green |
| B18 | self path unavailable | system-update tests | existing green | existing green |
| B19 | self path available | system-update tests | existing green | existing green |
| B20 | update cache path unavailable | system-update tests | existing green | existing green |
| B21 | update cache path available | system-update tests | existing green | existing green |
| B22 | local updater unavailable | system-update tests | existing green | existing green |
| B23 | local updater available | system-update tests | existing green | existing green |
| B24 | verified updater assembled | system-update tests | existing green | existing green |
| B25 | release downloader assembly error | system-update tests | existing green | existing green |
| B26 | release downloader available | system-update tests | existing green | existing green |
| B27 | engine directory enables lock seam | system-update engine-lock tests | existing green | existing green |
| B28 | engine lock acquisition error | system-update engine-lock tests | existing green | existing green |
| B29 | engine autostart settings available | autostart tests | existing green | existing green |
| B30 | autostart emits operator note | autostart tests | existing green | existing green |
| B31 | engine directory enables policy RPC discovery | position-policy tests | existing green | existing green |
| B32 | policy descriptor exists | position-policy tests | existing green | existing green |
| B33 | descriptor absent/nonfatal branch | position-policy tests | existing green | existing green |
| B34 | policy RPC dial fails | position-policy tests | existing green | existing green |
| B35 | policy RPC dial succeeds | position-policy tests | existing green | existing green |
| B36 | descriptor stat error is explicit | position-policy tests | existing green | existing green |
