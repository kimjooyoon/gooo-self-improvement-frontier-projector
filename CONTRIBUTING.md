# Contributing

Changes are PR-first. A pull request must pass the required GitHub Actions
checks before it can be merged. Do not rewrite or delete failed runs, pull
requests, tags, releases, or evidence artifacts.

The `.gooo` source owns frontier semantics. When semantics change, update the
source and its contract/fixtures in the same pull request. Generated outputs
are produced by CI and are not hand-authored current-state assertions.

The projector has no repository-write, commit, merge, release, or local-test
authority. Operators may perform those product actions separately after
reviewing a proposal.
