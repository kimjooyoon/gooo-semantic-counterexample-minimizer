# Gooo Semantic Counterexample Minimizer

This repository turns a Gooo semantic counterexample into the smallest
deterministic witness that still preserves its failure digest and origin graph.
The semantic contract is declared in
[`.gooo/counterexample-minimizer.gooo`](.gooo/counterexample-minimizer.gooo).
Go provides parsing, evaluation, deterministic delta-debugging, and rendering;
the `.gooo` activities own the reduction vocabulary and preservation law.

The fixed denominator is nine scenarios:

| scenario | expected state |
|---|---|
| `remove-unreachable-declaration` | `CLOSED` |
| `shrink-expression` | `CLOSED` |
| `reduce-effect-sequence` | `CLOSED` |
| `already-minimal` | `CLOSED` |
| `deterministic-replay` | `CLOSED` |
| `missing-oracle` | `UNKNOWN` / `DIRECT_MISSING` |
| `nondeterministic-oracle` | `UNKNOWN` / `AMBIGUOUS` |
| `failure-lost-during-reduction` | `REFUTED` |
| `origin-provenance-drop` | `REFUTED` |

Resolution precedence is `REFUTED > UNKNOWN > CLOSED`. Every UNKNOWN retains
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.
Improvement remains `UNKNOWN` until a matched before/after integer pair has the
same scenario, source, contract, toolchain, and runner identity.

Run the evaluator with a caller-owned empty output directory:

```text
gooo-semantic-counterexample-minimizer run \
  --meta .gooo/counterexample-minimizer.gooo \
  --contract contracts/denominator-v1.json \
  --source examples/counterexamples/shrink-expression.gooo \
  --out /caller-owned/output
```

The output contains exactly four artifacts:

- `minimization-events.ndjson`
- `minimized-counterexample.gooo`
- `preservation-receipt.json`
- `human-report.md`

The evaluator has zero repository-write, commit, merge, release-mutation, and
local-test authority. GitHub Actions is the validation authority and uses Go
1.27. The release workflow uses the standard `GITHUB_TOKEN` to create a draft,
upload the CI evidence asset, publish once, and verify the public immutable
release API plus annotated tag and asset digests.
