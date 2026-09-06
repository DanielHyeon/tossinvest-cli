# a120 adversarial proposal review

Date: 2026-09-06
Reviewer: independent adversarial review
Scope: proposal, design, `sdd-workflow` delta, and tasks before implementation

## Verdict

**Clear for proposal freeze and gstack autoplan.** B1 through B4 are corrected
in the current design, delta, and task wording: S is strictly before H, the
P..E serializer fixes commit/parent/path semantics, the a063 planning-base file
is bound to P, and B2 uses a bounded metadata allowlist plus both ordinary and
`tossos_testseams` Go package input enumerations. The proposal has the
right narrow intent: the only tuple is a063/P/E, the original planning base is
not rewritten, ordinary changes retain their existing comparison, and E-based
Function Logic Map obligations are still complete. The corrections are needed
to make the claimed committed-source and complete-history boundaries mechanical
rather than descriptive.

## Blockers

### Resolved B1 — source snapshot ancestry is strict and non-self-referential

The current design and delta now require `P <= E <= S < H`; this correctly
prevents the record at H from naming its own commit. Implementation must retain
the following concrete checks:

- full 40-hex commits with `P` ancestor of `E`, `E` ancestor of `S`, and `S`
  a **strict** ancestor of detached `H`;
- `S` resolves to the record's exact full tree ID;
- the record is at the fixed path
  `openspec/changes/a063-align-attestation-renewal-profile/execution-baseline.json`;
- the record, ledger, and review inputs resolve to regular, repository-contained
  files at H and have the recorded H-object bytes and SHA-256 values.

Required fixture failures: `S == H`; non-ancestor P/E/S/H; abbreviated ID or
moving ref; wrong S tree; a record outside the fixed a063 path; a symlinked
record, ledger, or review file; and a ledger/review blob changed after its
recorded hash.

### Resolved B2 — generated metadata is bounded and cannot be a build input

The normal SDD commands necessarily create ignored CodeGraph, `.sdd`, and Python
bytecode state, so a blanket ignored-file ban would make the required acceptance
sequence impossible. Use a static, fail-closed generated-metadata allowlist
instead of a general cache directory exemption. Every untracked or ignored path
must be rejected unless it is a regular, non-symlink file under one of these
exact generated locations:

- `.codegraph/`;
- `.sdd/index-state.json`;
- `.sdd/gbrain-home/`;
- `.sdd/history/checkpoints/`, `.sdd/history/refresh-queue/`,
  `.sdd/history/index-refresh.lock`, or `.sdd/history/events/*.jsonl`;
- `auth-helper/**/__pycache__/*.pyc` or `tools/**/__pycache__/*.pyc`.

The last two patterns admit only cache bytecode, not an arbitrary file under an
`__pycache__` directory. The allowlist must reject a path component resolved
through a symlink, executable files, and names ending `.go`, `.c`, `.cc`,
`.cpp`, `.cxx`, `.h`, `.hh`, `.hpp`, `.s`, `.S`, `.syso`, `.o`, `.a`, `.so`,
`.dylib`, or `.dll`.

An extension list alone cannot prove that an otherwise data-looking file is not
selected by `go:embed`. Therefore the validator must also run `go list -deps
-test -json ./...`, fail closed if that command fails, construct repository
input paths from each package directory plus `GoFiles`, `CgoFiles`,
`TestGoFiles`, `XTestGoFiles`, `CFiles`, `CXXFiles`, `MFiles`, `HFiles`,
`SFiles`, `SysoFiles`, and `EmbedFiles`, then reject every untracked or ignored
candidate in that input set regardless of its allowlisted location. Evidence
files are not an untracked-file exception: the record, ledger, and reviews must
already be tracked H objects. This is practical because the only mutable
metadata roots are dot-prefixed tool state or Python caches; the reviewed a063
source paths are outside them, and S..H permits no tracked product path change.

Continue to compare every tracked path from S to H and permit only
`openspec/**` and `docs/pm/**`; separately reject staged and unstaged changes.
Do not follow symlinks while scanning. A tracked Go entry in S must itself be a
regular file; merely matching an S symlink is not acceptable source evidence.

Required fixture failures: untracked and ignored `.go`, `.c`, `.s`, `.h`,
object/archive/shared-library, and embedded asset inputs; an arbitrary file
inside an otherwise allowed root; a symlink anywhere on an allowed path; staged
and unstaged source edits; source deletion/rename; tracked or untracked Go
symlink; and any S..H tracked path outside the two evidence prefixes. Positive
fixtures must prove that actual `.codegraph/codegraph.db`,
`.sdd/index-state.json`, a permitted `.sdd/history` state file, a GBrain state
file, and `tools/sdd/__pycache__/*.pyc` do not block a clean detached H.

### Resolved B3 — P..E history now has graph semantics

The design now defines `git rev-list --topo-order --reverse E --not P`, parent
lists, per-parent NUL-delimited no-rename status records, root treatment, and
net endpoint function inventories. This is sufficient if the implementation
shares that serializer between draft generator and validator.

Required fixture failures: an omitted row despite matching count/digest; duplicate
row; a reverted intermediate path hidden by endpoint diff; a merge-parent path
omission; reordered commits; changed status; and a row labeled `complete`,
`waived`, or pre-edit evidence. The positive fixture must include a merge and a
base-revision deletion.

### Resolved B4 — adoption binds the unchanged a063 planning-base file

The design and delta now require the exact regular a063 `base-commit.txt` bytes
to resolve to allowlisted P before E may be selected. This complements, rather
than replaces, the existing no-recapture behavior.

Required fixture failures: a063 `base-commit.txt` altered to E/HEAD/another
commit, abbreviated or invalid, and a symlink. A valid a063 record with no
exception should still use P normally; a malformed exception must fail closed,
not fall back to P or choose a base from `SDD_BASE_REF`.

## Conditions that are already correctly constrained

- The exception is a single hardcoded a063/P/E tuple, so it is not a general
  recapture mechanism.
- The result is named `execution-baseline adoption exception` and explicitly
  preserves the truth that original pre-edit compliance and the inherited
  history's missing FLMs were not established.
- E-based requirements remain full-function requirements. Extra maps cannot
  compensate for an omitted E..S row, and an E-era deleted function still needs
  base-revision evidence.
- `SDD_BASE_REF` is validation-only: after a valid record it must equal E, and
  it cannot select E by itself. Ordinary changes continue to require their
  persisted P.
- The listed E..S Go path allowlist is an additional anti-scope-laundering
  constraint. The adoption review must compare the generated E..S inventory
  with that list before accepting it.

## Task-level additions needed before proposal freeze

Task 2.1 now names B2's generated-metadata positive controls, forbidden
input/symlink negatives, both enumeration modes, and enumeration failure. Task
2.2 must implement that exact static allowlist and `go list` input rejection,
not a broad `.sdd/`, `.codegraph/`, or `__pycache__/` prefix exemption. Task 2.3
retains ordinary-P and malformed-a063 fail-closed behavior. These are concrete
implementation acceptance checks, not remaining proposal-freeze blockers.
