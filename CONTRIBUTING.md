# Contributing

## Contributor's Agreement

You are under no obligation whatsoever to provide any bug fixes,
patches, or upgrades to the features, functionality or performance of
the source code ("Enhancements") to anyone; however, if you choose to
make your Enhancements available either publicly, or directly to the
project, without imposing a separate written license agreement for
such Enhancements, then you hereby grant the following license: a
non-exclusive, royalty-free perpetual license to install, use, modify,
prepare derivative works, incorporate into other computer software,
distribute, and sublicense such enhancements or derivative works
thereof, in binary and source code form.

## Pull requests

- All commits must be "Signed-off" (i.e., by using `git commit -s`),
  acknowledging that you agree to the [Developer Certificate of
  Origin](DCO.txt).

- All PRs should summarize the purpose of the PR in the attached
  GitHub conversation.

- Larger fixes or enhancemens should be discussed with the TSC or
  developers, e.g., on Slack or over Email.

- PRs should be sent to the `main` branch by default. A committer or
  the TSC may request that certain bug fixes also be submitted to a
  minor release branch.

- Follow existing code style precedent. For Go, you should mostly
  conform to the style and form enforced by the "go fmt" and "golint"
  tools for proper formatting.

- For any new functionality, please write appropriate go tests. These
  run as part of the continuous integration system.

- Make sure that the project's default copyright and header have been
  included in any new source files.

- Make sure your code passes linting, by running `make lint` before
  submitting the PR.

- Make sure you have locally tested using `make test` and that all
  tests succeed before submitting the PR.

- PRs which introduce a new Go dependency to the project should
  explain why the dependency is required. Any new dependency should be
  added to `LICENSE_DEPENDENCIES.md` by running
  `scripts/update-license-dependencies.sh`.

## Documentation

- The [README](README.md) is a place to document critical information
  for new users of Warewulf. It should typically not change, but in
  the case where a change is necessary a PR may update it.

- The [CHANGELOG](CHANGELOG.md) documents **functional** differences
  between versions of Warewulf, and should **not** read like a commit
  log.

  Once a release is tagged (*e.g.  v4.0.0*), a new top level section
  is made titled **Changes Since vX.Y.Z** (*e.g. Changes Since
  v4.0.0*) where new changes are documented, leaving the previous
  section immutable.

  The CHANGELOG must be updated for any of the following changes:
  - Renamed commands or subcommands
  - Deprecated / removed commands or subcommands
  - Changed defaults / behaviors
  - Backwards-incompatible changes
  - New features / functionalities

- The [user guide](https://warewulf.org/docs/) should document
  anything pertinent to the use of Warewulf. Changes to Warewulf
  functionality should simultaneously include pertinent updates to the
  user guide, which is maintained alongside the code under
  `userdocs/`.

## Branches

- Development occurs primarily on the `main` branch. This is the
  branch that most PRs should be submitted against unless otherwise
  directed.

- A minor release is accompanied by a minor branch named `v4.MINOR.x`
  from which patch releases may be generated.

- No other branches are maintained in the primary Warewulf repository.

## Maintaining

Additional policies regarding the maintenance of the Warewulf source
code, including roadmapping, merging, and release policies, is
documented at [MAINTAINING.md](MAINTAINING.md).
