# Branch Test Map: `TestEveryStateChangingRouteAlsoGoesThroughTheCSRFGate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at line 416 | Existing static guards; TestA154StaticSessionGuardTransparency and TestA154ComputeOnlyHTTP for new boundary | Existing behavior / compile RED | focused final PASS |
| B2 | switch at line 418 | Existing static guards; TestA154StaticSessionGuardTransparency and TestA154ComputeOnlyHTTP for new boundary | Existing behavior / compile RED | focused final PASS |
| B3 | case at line 419 | Existing static guards; TestA154StaticSessionGuardTransparency and TestA154ComputeOnlyHTTP for new boundary | Existing behavior / compile RED | focused final PASS |
| B4 | case at line 421 | Existing static guards; TestA154StaticSessionGuardTransparency and TestA154ComputeOnlyHTTP for new boundary | Existing behavior / compile RED | focused final PASS |
| B5 | case at line 423 | Existing static guards; TestA154StaticSessionGuardTransparency and TestA154ComputeOnlyHTTP for new boundary | Existing behavior / compile RED | focused final PASS |
| B6 | range at line 427 | Existing static guards; TestA154StaticSessionGuardTransparency and TestA154ComputeOnlyHTTP for new boundary | Existing behavior / compile RED | focused final PASS |
| B7 | if at line 428 | Existing static guards; TestA154StaticSessionGuardTransparency and TestA154ComputeOnlyHTTP for new boundary | Existing behavior / compile RED | focused final PASS |
