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

## Technical Charter

- section 2(B) requires a list of Technical Steering Committee (TSC)
  members here

### Technical Steering Committee

(in alphabetical sort order of surname)

```text
- Jonathon Anderson <jonathon.anderson@ciq.co>, <anderbubble@gmail.com>
- Christian Goll <christian.goll@googlemail.com>
- John Hanks <griznog@gmail.com>
- Gregory Kurtzer <gmkurtzer@gmail.com>
- Jeremy Siadal <jeremy.c.siadal@intel.com>
```


## Pull Requests (PRs)

1. All PRs should summarize the purpose of the PR in the attached
   GitHub conversation.
2. Larger fixes or enhancemens should be discussed with the project
   leader or developers, e.g., on Slack or over Email.
3. Essential bug fix PRs should be sent to both development and
   release branches.
4. Small bug fix and feature enhancement PRs should be sent to
   development only.
5. Follow the existing code style precedent. For Go, you will mostly
   conform to the style and form enforced by the "go fmt" and "golint"
   tools for proper formatting.
6. For any new functionality, please write appropriate go tests that
   will run as part of the Continuous Integration (github workflow
   actions) system.
7. Make sure that the project's default copyright and header have been
   included in any new source files.
8. Make sure your code passes linting, by running `make list` before
   submitting the PR. We use `golangci-lint` as our linter. You may
   need to address linting errors by:
   - Running `gofumpt -w .` to format all `.go` files. We use
     [gofumpt](https://github.com/mvdan/gofumpt) instead of `gofmt` as
     it adds additional formatting rules which are helpful for
     clarity.
   - Leaving a function comment on **every** new exported function and
     package that your PR has introduced. To learn about how to
     properly comment Go code, read [this post on
     golang.org](https://golang.org/doc/effective_go.html#commentary)
9. Make sure you have locally tested using `make test-it` and that all
   tests succeed before submitting the PR.
10. Ask yourself is the code human understandable? This can be
   accomplished via a clear code style as well as documentation and/or
   comments.
11. The pull request will be reviewed by others, and finally merged
   when all requirements are met.
12. The `CHANGELOG.md` must be updated for any of the following
   changes:
   - Renamed commands
   - Deprecated / removed commands
   - Changed defaults / behaviors
   - Backwards incompatible changes
   - New features / functionalities
13. PRs which introduce a new Go dependency to the project via `go
   get` and additions to `go.mod` should explain why the dependency is
   required. Any new dependency should be added to the
   `LICENSE_DEPENDENCIES.md` by running
   `scripts/update-license-dependencies.md`.

## Documentation

There are a few places where documentation for the Warewulf project
lives. The [changelog](CHANGELOG.md) is where PRs should include
documentation if necessary. When a new release is tagged, the [user
guide](https://warewulf.org/docs/) will be updated using the contents
of the `CHANGELOG.md` file as reference.

1. The [changelog](CHANGELOG.md) is a place to document **functional**
   differences between versions of Warewulf. PRs which require
   documentation must update this file. This should be a document
   which can be used to explain what the new features of each version
   of Warewulf are, and should **not** read like a commit log. Once a
   release is tagged (*e.g.  v1.0.0*), a new top level section will be
   made titled **Changes Since vX.Y.Z** (*e.g. Changes Since v1.0.0*)
   where new changes will now be documented, leaving the previous
   section immutable.
2. The [README](README.md) is a place to document critical information
   for new users of Warewulf. It should typically not change, but in
   the case where a change is necessary a PR may update it.
3. The [user guide](https://warewulf.org/docs/) should document
   anything pertinent to the usage of Warewulf.
