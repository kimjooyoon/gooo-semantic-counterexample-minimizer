# Repository ownership

This repository is owned by the task that creates
`gooo-semantic-counterexample-minimizer`. Its complete scope is the semantic
`.gooo` declaration, fixed denominator, evaluator, deterministic reducer,
caller-owned output artifacts, CI workflows, pull request, and immutable
release.

The evaluator must not write the repository, commit, merge, mutate a release,
or run local Go validation. Outputs are allowed only under a caller-owned empty
directory outside the repository. GitHub Actions is the authority for format,
vet, tests, build, conformance, post-main evidence, and release verification.
