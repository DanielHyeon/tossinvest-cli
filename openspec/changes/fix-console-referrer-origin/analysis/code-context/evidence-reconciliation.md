# Evidence Reconciliation

| Question | CodeGraph / current HEAD | CodeGraphContext / memory | Reconciled conclusion |
|---|---|---|---|
| Which code emits the failing policy? | `remoteRuntime.security` and `Console.render` set `Referrer-Policy: no-referrer`; the shared and restart templates repeat it in meta declarations. | Broad console connectivity only; no contrary focused symbol. | Keep all four policy surfaces synchronized at `same-origin`. |
| Is the origin gate itself wrong? | Explicit `Origin: null` is intentionally rejected before CSRF. | No evidence supporting opaque-origin acceptance. | Keep `sameOrigin` and `sameOriginForMutation` unchanged. |
| What proves the browser interaction? | Chrome probe reproduced `Origin: null`; policy override produced canonical origin and reached CSRF. | No prior episode. | Response policy is the root cause and the safe correction point. |
| What is the blast radius? | Wrapper fronts all console routes; renderer and templates cover normal/restart documents; three peer/Host branches and all headers run before handlers. | Console is a high-connectivity module. | Pin remote/rendered headers, both meta declarations, opaque-origin rejection, and branch behavior; no adjacent refactor. |

There is no unresolved evidence conflict. Production editing is limited to the
two `Referrer-Policy` response values and two HTML meta values after RED
coverage is observed.
