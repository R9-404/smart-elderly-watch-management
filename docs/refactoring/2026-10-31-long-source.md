# Long source decomposition

Baseline: `3bd4375d5aafab32948f145f7535d5ef6670ef67`.

## Scope and structure

The production-source inventory identified eight Go files above 350 lines and
four Vue pages above 300 lines. These are review thresholds, not style limits.
Tests, dependencies and generated build outputs are outside this inventory.

The eight Go files now separate routes, queries, mutations, profile credentials,
telemetry parsing, SOS, notification enqueueing, delivery, scheduling and care
dispatch. The original packages and exported interfaces are retained.

Two long orchestration functions have also been decomposed:

- `rules.Engine.Device`: lock the device, read policy, collect observations and
  baselines, evaluate and persist candidates, upgrade night incidents, commit.
  Input collection and the night upgrade use the same transaction as before.
- `notify.Worker.Drain`: claim and commit a bounded batch, deliver external
  messages, persist results guarded by the original lease token. Provider calls
  still start only after the claim commits. Retry order and concurrency are unchanged.

The person, family and threshold pages delegate create/edit forms to record
components. Pages retain drafts, requests, dialogs and refresh ownership. Form
components preserve validation callbacks and reset behavior through parent refs.
Resident writes and health alert methods follow the existing Options API method
module pattern. The health alert dialog emits the original resolution kind.

Following the redundant-source cleanup request, create/edit fields share three
domain components rather than six duplicated forms. Explicit `editing` props retain
hidden identifiers, create-only elder phone input, read-only bound devices and
disabled threshold device identifiers. Six duplicate component files are removed.
An unreferenced `nullableCoordinate` helper and orphaned/repeated comments are also
removed. Frontend import reachability found no unused production source file;
the images and font resources reviewed still have live references.

OpenAPI regeneration changes only 12 implementation-file descriptions following
the capability route move. HTTP paths, methods, authentication and DTOs are unchanged.
There are no database migrations or dependency updates in this batch.

## Verification

- AST comparison: 82 of the 84 declarations in the eight original Go files remain
  identical, ignoring positions/comments. The two intentional exceptions are
  `Engine.Device` and `Worker.Drain`, whose extracted steps received independent
  behavior review. The earlier migration-only intermediate state matched all 84.
- `go vet` passed for handlers, notification, care and rules packages.
- Seven existing MySQL integration regressions passed for short notification
  deadlines, stale leases, bulk fallback, retry ordering, baseline recovery,
  concurrent offline deduplication and idempotent night upgrades.
- Four existing handler regressions passed for guardian integrity, resident edit,
  signed telemetry replay and atomic SOS rollback.
- Frontend type checking and production build passed. The existing large-bundle
  advisory remains; this source-structure change does not optimize bundle size.
- Existing Chromium forms acceptance: 30 checks passed. Eight focused production
  bundle checks also passed for validation/reset, threshold limits, alert display
  and resolution identity. HTTP responses in these browser checks were mocked.
- OpenAPI consistency and native API boundary checks passed.
- Independent standards and behavior reviews found no blocking issue.

PR #55 completed 26 checks (23 successful, two neutral, one skipped). Two
post-CI MySQL regressions and eight component-path checks passed. It was merged
as `77b1cf487b1c8ccf026620e04296326157c2ac18`; the merge tree matched the
tested head `6675f05c748cc3935cb38d3867095ffd5e13c81f` exactly.

## Second batch: normalization, aggregation and statistics reads

The next batch separates four additional long functions into domain steps:

- `Sample.Normalize`: record validation, metric/unit contract, and measurement
  normalization. Private helpers operate on the public method's local copy;
  partial normalized values returned with errors remain unchanged. A duplicate
  interval-duration guard is removed: its earlier identical check already rejects
  invalid intervals, and the time fields cannot change between those checks.
- `Aggregate`: sample selection remains in the entry point, while interval step
  allocation and total reduction have named helpers. Sorting, accumulation order,
  overflow checks and partial statistics on errors are retained.
- `healthStatistics`: request-range validation and bounded SQL sample loading
  have separate helpers. Authorization precedes cache reads; the revision checks,
  candidate limit, byte-exact source predicates and window filtering are retained.
  The named request type preserves the original JSON field order for cache keys.
- `appendEvidenceFast`: eligibility checks move to the evidence value, with a
  pointer receiver to avoid copying the candidate. The encoding sequence remains
  together, including byte order, canonical fallbacks and borrowed-buffer lifetime.

Ten distinct existing regressions passed across normalization, aggregation,
evidence encoding, MySQL cache identity/consistency and submicrosecond windows.
Affected-package vet and OpenAPI consistency passed. Independent review found no
behavior issue; its suggestion to avoid new Sample value copies was applied and
the resulting local-copy/error semantics were reviewed again.

A bounded baseline/candidate benchmark of three changed hot paths found the same
allocation counts: evidence buffer 1, 1000 interval samples 16, and 1000 daily
samples 26 allocations per operation. These short measurements do not establish
a speed improvement. No frontend work or dependency changes are included here.

PR #56 completed 25 checks (23 successful, one neutral, one skipped). Three
post-CI checks passed for MySQL cache authorization, normalization and byte-exact
evidence. It was merged as `ab0699421684781f23f6ec979b7b04a1504912b9`, with the
same tree as tested head `718b2279c23f2f5c7a4521d4dbb5c73bdbc0c188`.

## Dead-code cleanup

Production-call searches and the complete command dependency graph identified
retired adapters alongside live test infrastructure. This cleanup removes the
production `Evaluate`, `marshalEvidenceFast`, `dispatchOne`, `hitWindow.allow`,
`Enqueue`, `SyncOne` and fragment-cache `valueRows` entry points. Their ordinary
regressions now call the active compact evaluator, buffer encoder, dispatcher,
budget limiter, channel enqueueing and batch projection paths directly.

The independent map-shaped rule oracle, single-event projection implementation
and fragment cache remain only in existing reference/A-B test files where they
provide comparison baselines. No new test files or test suites are introduced.
Cross-package test infrastructure (`db.Postgres`, `dbtest.MySQL`, `health.Save`)
is retained: those consumers are valid dependencies, not dead references.

`health/batch_sql.go` and its retired-cache test file are deleted. The live
`buildValueRows` implementation moves unchanged beside its sole production caller
in `batch_statements.go`. Its literal and maximum-size boundary assertions remain
in `TestBuildValueRows`. The two offline benchmark source lists and file indexes
are updated to remove references to deleted files; historical result manifests
retain their original source hashes.

All production commands build. Eleven distinct focused regressions passed for
rule/evidence equivalence, SQL construction, dual-database recovery/replay,
notification rollback, stale notification leases and sliding-window boundaries.
Affected-package vet, OpenAPI consistency and updated Python/index parsing passed.
Independent review caught omitted SQL-builder boundary assertions; those were
restored and rechecked. No review issue remains open.

Frontend entry reachability still accounts for all 72 production source files;
the reviewed images, fonts and font-check fixtures have live consumers. This pass
does not delete referenced resources merely because they are outside the runtime
entry point. PR #57 completed 25 checks (23 successful, one neutral, one skipped)
and three post-CI smoke checks. It was merged as
`fadd93416d1609a19163fa99a1b6fdab05faa216`, matching the tested head's tree.

## Handler decomposition

Six more initially inventoried functions now separate their domain steps inside
the existing source files:

- Device registration delegates defaults and input validation to its parser.
- Family creation delegates the locked, at-most-two-elders read to a target
  loader. It still inserts the account before that read and rolls back on zero,
  ambiguous or already-bound matches.
- Invitation registration parses and validates one input record, then delegates
  account creation/activation and conditional elder linking to one persistence
  step. The parent retains the original locks, password hashing, invitation
  consumption, failure responses and transaction commit/rollback ownership.
- Invitation generation separates expiration calculation and bounded uniqueness
  retries. Elder existence is still checked before expiration parsing.
- Device listing delegates row projection, coordinate decryption and online
  calculation, using the same captured timestamp for every row.
- Vital-alert detail delegates response presentation after the unchanged lookup.

Seven focused MySQL regressions passed: existing family onboarding, overwrite
rejection, ambiguous elder phone, concurrent invitation registration, invitation
expiry and configured threshold explanation tests, plus one registration/list
contract test added to the existing device validation test file. The latter covers
defaults, owner/profile linkage, stored-secret separation, filtered online counts,
decrypted coordinates and rejected input without extra records. No new source or
test file is added. Handler vet, OpenAPI consistency and independent behavior and
standards review passed. PR #58 was subsequently merged as b97cd16; its completed CI had no failing checks.

## Core transaction steps

Six further functions separate real business steps while retaining the same
transactions, SQL, validation priority and response contract:

- SOS separates request validation, locked authentication/owner revalidation and
  notification enqueueing. The coordinator retains retry acknowledgement before
  nonce reservation, rollback before failure responses and commit before success.
- Binding separates bind-code admission and conditional elder/profile linking.
  Its coordinator retains elder/device lock order and both affected-row checks.
- Legacy transfer separates one-table import and row conversion. The importer
  still closes its cursor on every conversion/write failure and checks its count;
  the parent still commits the read snapshot before the atomic target transaction.
- Alert resolution separates input parsing, the original reason lookup and vital
  handling. Selected-record ownership and the remaining-alert check are unchanged.
- Telemetry separates vital, daily and position/alarm persistence. Every step uses
  one captured owner and the same transaction. Metadata sealing still precedes SQL;
  encryption calls, missing-value handling and observation ordering are unchanged.
- Fence observation separates locked configuration loading, pure state advancement
  and event/care persistence in `fence_observation.go`. Stale observations are still
  rejected before timezone loading. Exit-event notification order is unchanged.

Three focused baseline checks passed before changes. Fifteen distinct existing
regressions passed after decomposition, including real MySQL handler behavior and
PG 0006/0007/0008-to-MySQL migration. The three affected SOS paths were rechecked
after extracting locked authentication. No new tests or assertions were needed.
Handler/database vet, OpenAPI consistency and independent standards/behavior
reviews passed. One orphaned telemetry comment found in review was removed.
PR #63 completed 25 checks (23 successful, one neutral, one skipped). Three
post-CI MySQL smoke tests passed. It was merged as
`468a269ad88eb803155fb46fb16478798f682939`; its tree matched tested head
`e32742de588a21b9caba0bd8bcf12b1395f30fa9`.

## Final batch: state machines, batches and worker coordination

The remaining 12 inventoried functions have distinct steps without new workflow
frameworks or changed contracts:

- Care dispatch separates claim/commit, external delivery and attempt-guarded
  completion. Only an empty queue returns `found=false`; retry/fallback ownership
  and authenticated delivery receipts still win over stale attempts.
- Care transitions separate state admission and incident/event persistence.
  Authorization, the row lock, independent verification, verdict audit and commit
  remain in the handler. Receipt signature/body validation is separate from its
  unchanged idempotent delivery transaction.
- Analytics batches separate bounded event reads from aligned/coalesced projection
  arrays. Every claimed event is still acknowledged only after PG projection
  succeeds. The poller delegates its bounded drain while retaining pool recovery,
  optional configuration, status sampling, cancellation and backoff ownership.
- Health upload separates source validation/normalization from both existing lock
  modes. Sample batch writes delegate the stored-version read after the owner lock;
  revision reservation, history ordering and argument-buffer reuse stay together.
- Login shares its two-factor/session-success step across administrator and family
  accounts. Password comparison order, dummy bcrypt handling, unlinked-family
  rejection and submitted-phone versus stored-phone use are unchanged.
- Factor setup and recovery-code generation have named helpers. Setup still resets
  the initial not-found result; confirmation/disable and session revocation retain
  their original state checks and transaction ordering.
- Database worker startup groups the same seven workers and wait group. The parent
  still drains them before closing the pool. The maintenance CLI separates its
  isolated-database operations while keeping probe's early return and safety gates.
- SQL dialect rewriting lives in `mysql_dialect.go`, with both file indexes updated.
  The original regex vocabulary, quote/comment-aware rewrites and placeholder
  scanner remain unchanged; query-plan caching and bound arguments stay separate.

Seventeen existing core regressions passed once on real MySQL/PostgreSQL where
required: notification ownership/retry, transactional replay and saturated drain,
health versions/history/conflict rollback, TOTP recovery/session revocation, native
login response, care verdict audit and signed receipt concurrency, both health
lock modes, and SQL rewriting/binding. No tests or benchmarks were added. The
changed packages passed vet and unused/tagged-switch analysis; OpenAPI consistency
still reports 107 paths. Final-head CI and minimal post-CI smoke are pending.

## Final inventory and deliberate boundaries

The original eight Go files and four Vue pages have been decomposed. A later SQL
hardening change temporarily made `mysql_translate.go` 355 lines; separating its
syntax-rewrite responsibility resolves that new long file too. The current full
production scan has no Go file above 350 lines, Vue/JS/TS file above 300 lines, or
CSS file above 350 lines. Dependencies, generated artifacts and tests are excluded.

All 30 originally inventoried functions have received responsibility review and
applicable decomposition. Eight remain at least 90 lines; this threshold triggers
review and does not justify fragmenting a coherent algorithm or transaction:

| Function | Current lines | Retained responsibility |
| --- | ---: | --- |
| `deviceBind` | 125 | Ordered elder/device locks, binding write and code-consumption CAS, commit/rollback |
| `inviteJoin` | 124 | Invitation/account locks, hashing, activation, invitation consumption and transaction ownership |
| `deviceSos` | 106 | Retry acknowledgement before nonce reservation, incident/notification transaction and responses |
| `securityFactor` | 104 | Factor state admission, OTP replay protection and commit/session-revocation ordering |
| `compileMySQLQuery` | 97 | One cursor through quoted values/comments, operators and placeholder positions |
| `appendEvidenceFast` | 94 | One borrowed-buffer JSON encoding sequence with its ordered fields and fallback |
| `analytics.Run` | 93 | Pool recovery, status publication, drain scheduling and cancellation/backoff coordination |
| `deviceCreate` | 91 | Registration transaction, secret/owner/profile persistence and final response |

These are documented boundaries, not unfinished row-count reductions. The source
structure task does not establish that the project has no further functional bugs
or that all production release conditions have been accepted.
