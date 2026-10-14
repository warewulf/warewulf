# AGENTS.md

Orientation for AI coding agents working in the Warewulf repository. It points
to where things are documented rather than restating them. The documents it
points to are written for every contributor, so when you find one wrong or
missing, fix it there.

Warewulf provisions HPC clusters: a server (`wwctl`, `warewulfd`) network-boots
nodes into a node image plus rendered overlays. For the model, read
`userdocs/getting-started/introduction.rst` and `provisioning.rst`.

## Before you change something

These documents hold the project's conventions and name the helpers to reuse.
Read these in full before you start:

1. `CONTRIBUTING.md`
2. `userdocs/contributing/development-environment.rst` (building, testing,
   and writing tests with `internal/pkg/testenv`)

Then read in full the document for each area your change touches:

3. **Go code**, including tests → `userdocs/contributing/code.rst` (layout,
   the `wwctl` subcommand pattern, logging, configuration, existing helpers,
   generated files).
4. **Overlays or templates** (`overlays/`, `*.ww`) →
   `userdocs/overlays/templates.rst` for the template functions (`abort`,
   `file`, `softlink`, `Include`, …) and the `wwdoc` documentation comments;
   also the sections of `userdocs/overlays/overlays.rst` for the overlays you
   change.
5. **Node, profile or `warewulf.conf` fields** → the matching `userdocs/nodes/`
   or `userdocs/server/configuration.rst` page; also update the man pages in
   `docs/man/man5/`, which are written by hand. A change to the on-disk format
   may also need `internal/pkg/upgrade` (`wwctl upgrade`).
6. **Documentation** → `userdocs/contributing/documentation.rst`.

Assume a similar problem has been solved before, and look for the existing
function before you write a new one.

The user guide is `userdocs/` (Sphinx), published per branch at
<https://warewulf.org/docs/>.

## Checking your work

Run what CI runs (`.github/workflows/testsuite.yml`); `make help` lists the
targets. Warewulf targets Linux: the tests build and run only there
(`userdocs/contributing/development-environment.rst`).

```sh
make fmt
make lint vet staticcheck deadcode
make test
```

The make targets generate `internal/pkg/config/buildconfig.go` and other files
from their `.in` sources. To call `go test` directly, run `make config` first
and set `TZ=UTC`, as `make test` does. Edit the `.in` file, not the generated
one.

## Pull requests

- Sign off every commit (`git commit -s`, per `DCO.txt`).
- Target `main`; maintainers cherry-pick fixes to the `v4.N.x` branches.
- Add a `CHANGELOG.md` entry under `## Unreleased` for a functional change. CI
  checks for one; `CONTRIBUTING.md` lists what needs one.
- Update `userdocs/` in the same PR when behavior changes.

## Context

Commit messages are brief. The reasoning behind a change is usually in its pull
request: the merge commit names it, and `gh pr view <N>` shows it. OpenHPC
(<https://github.com/openhpc/ohpc>) packages Warewulf, so changes to packaging
or to the minimum Go version affect it (`MAINTAINING.md`).

When the documentation and the code disagree, say so, and fix whichever is
wrong in the same PR. When you move or rename something this file
points to, update the pointer here too.
