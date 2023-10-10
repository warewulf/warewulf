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

## Technical Steering Committee

(in alphabetical order by surname)

- [Anderson, Jonathon](janderson@ciq.com) (CIQ)
- [Goll, Christian](cgoll@suse.de) (SUSE)
- [Hanks, John](griznog@gmail.com) (Chan Zuckerburg Biohub)
- [Kurtzer, Greg](gmk@ciq.com) (CIQ)
- [Siadal, Jeremy](jeremy.c.siadal@intel.com) (Intel)

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

## Merging

- All new contributions to the project should first be merged through
  a PR to the `main` branch.

- Patches to a minor branch should be copied ("cherry-picked") from a
  previously-merged PR.

- Each PR must, prior to merge, be reviewed and approved by a
  committer other than the author of the PR.

- Before approving of a PR, an approver must confirm that
  - all lint checks and tests pass with the given PR;
  - new tests to cover the change have been added as appropriate;
  - all commits in the PR have an appropriate DCO “Signed-Off-By”;
  - documentation, including (but not necessarily limited to) the
    CHANGELOG and user guide, have been updated appropriately.

- Any committer may, at his discretion, merge a PR that has the
  requisite approvals and for which all tests are passing. This
  includes, but is not limited to, a reviewer of the PR or the author
  of the PR.

- Committers should consider the current roadmap when choosing to
  approve or merge a given PR. For example, proper and successful
  bugfixes are likely always appropriate to merge. However, new
  features may not be appropriate to merge if they are not included in
  the next milestone. New features which incompatibly alter existing
  interfaces or behavior should not be approved or merged unless
  included in the current milestone.

- Committers which approve or merge PRs which are disruptive to the
  current milestone may have their committer access revoked by the
  TSC.

## Roadmapping

- The project roadmap is formally defined by GitHub milestones
  associated with the primary Warewulf repository.

- The roadmap, and its milestones, are drafted and managed by the TSC
  chair to reflect and codify the priorities of the TSC and the
  community. Definition of the roadmap includes, but is not limited to
  - the name, description, and due date for each milestone;
  - the issues and PRs associated with a given milestone.

- Only the TSC chair may assign issues to or remove issues from
  milestones. The TSC chair has proactive discretion to assign or move
  issues to reflect and codify the priorities of the TSC and the
  community, but may not act in opposition to the direction of the
  TSC.

- Requests for changes to the roadmap that are made in the Warewulf
  community meeting are recorded in the Warewulf community journal.

## Releases

- Warewulf releases follow a `MAJOR.MINOR.PATCH` format.

- Releases are generated by an automated process encoded as a GitHub
  actions workflow.

- All releases must pass the full Warewulf test suite.

- Releases are published by a member of the TSC at the direction of
  the TSC.

- Each release is accompanied by an updated changelog.

### Major releases

- All releases occur within major version “4.”

### Minor releases

- A minor release is denoted by a tag named `v4.MINOR.0`.

- A minor release candidate is denoted by a tag named
  `v4.MINOR.0rcNUMBER`, where `NUMBER` begins at “1” and increments
  for each candidate for the given minor release.

- Minor releases are defined by a previously-planned milestone named
  for the projected release.

- A minor release candidate may be published by the TSC when all
  functional issues and pull requests in the associated milestone are
  closed. (e.g., documentation issues and pull requests may remain.)
  This may also be accomplished by the TSC re-scoping the associated
  milestone (e.g., by moving issues or pull requests to a different
  milestone).

- A minor release may be published by the TSC two weeks after a
  release candidate if no major defects have been found in the release
  candidate and there are no open issues or pull requests.

### Patch releases

- A patch release is denoted by a tax named `v4.MINOR.PATCH` where
  `PATCH` is greater than `0`, and tags a commit on a minor branch.

- The TSC may identify changes in the “main” branch to be ported to a
  minor branch.

- A patch release may be published by the TSC whenever one or more
  changes have been ported to a minor branch.
