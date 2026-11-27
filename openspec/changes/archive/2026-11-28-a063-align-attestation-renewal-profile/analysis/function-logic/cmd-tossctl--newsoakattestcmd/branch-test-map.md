# Branch Test Map: `newSoakAttestCmd`

There are no AST B-ids.

| B-id | Happy-path behavior | Focused test and scope |
|---|---|---|
| B1 | Constructing the command registers `attest` below `soak` as a local, read-only command with a nonempty short description. | `TestSoakCommandsAreRegisteredAndReadOnly` finds `attest` below `soak`, checks its `source` annotation is `local`, checks it is not marked mutating, and requires a nonempty short description. It does not individually assert every flag binding or execute `RunE`. |
