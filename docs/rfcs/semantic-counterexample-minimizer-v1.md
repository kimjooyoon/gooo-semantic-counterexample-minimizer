# Semantic Counterexample Minimizer v1

## Semantic ownership

The `.gooo` activity declaration owns the six activities, effect vocabulary,
preservation predicate, UNKNOWN tuple, resolution precedence, and fixed nine
cell denominator. Go does not define a second semantic contract; it parses the
declaration and executes the declared evaluator and generator roles.

## Preservation law

A reduction is accepted only when the failure digest, failure anchor, origin,
and provenance remain present. Candidates are enumerated in stable source order
and selected one at a time. Events record exact input and candidate digests,
operation, oracle calls, and outcome in NDJSON order.

The direct-missing oracle is `UNKNOWN` with class `DIRECT_MISSING`. Conflicting
oracle observations are `UNKNOWN` with class `AMBIGUOUS`. A failure lost during
reduction or a dropped origin/provenance graph is `REFUTED`, and refutation has
priority over uncertainty and closure.

## Boundary and evidence

The four caller-owned artifacts are the reduction event stream, minimized
witness, preservation receipt, and human report. The receipt records exact
integer metrics for input/output bytes, input/output nodes, attempts, accepted
and rejected reductions, and oracle calls, together with the preserved failure
digest. Source inventory excludes the root README, `.git`, caller output,
cache, vendor, and toolchain internals.

No aggregate score or estimated percentage is emitted. Improvement is UNKNOWN
until both members of an exact matched before/after integer pair are supplied
for the same scenario, source, contract, toolchain, and runner.
