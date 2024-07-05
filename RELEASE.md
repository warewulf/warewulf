# Release procedure

Major versions (e.g., v4.5.0) are tagged directly on the main
branch. Minor versions (e.g., v4.5.3) are tagged on a separate release
branch.

1. Update `CHANGELOG.md`.
   - Identify the release date by changing "unreleased" to a date with
     `%Y-%m-%d` format, following existing convention in the log.
   - Do any final clean-up. (e.g., removing redundancy, adding issue
     or PR numbers).
2. Update `userdocs` implicit references to the latest version.
   - `contents/installation.rst`
   - `contributing/development-environment-vagrant.rst`
   - `quickstart/debian12.rst`
   - `quickstart/el.rst`
 references that imply the latest release to refer to the
   new version.
3. Cherry-pick updates from 1 and 2 above to a release branch if necessary.
   (i.e., when not doing a new major release)
4. Create a signed tag for the release of the format v4.MINOR.PATCH,
   following the format specified in <MAINTAINING.md>. (e.g., `git tag
   --sign v4.5.3; git push origin v4.5.3`)
5. Monitor the release action associated with the pushed tag at
   https://github.com/warewulf/warewulf/actions, and verify the
   generated draft release contains the expected artifacts. This
   includes the source tarball and RPMs for Suse and Rocky Linux.
6. Update the release notes for the release, summarizing and expanding
   on the relevant contents from <CHANGELOG.md>.
7. Confirm the correct values for the pre-release and latest release
   flags.
8. Publish the release.
9. Announce the release as a [post][1] to warewulf.org/news, linking
   to the GitHub release.
10. Announce the release on the Warewulf Slack, linking to the
    warewulf.org/news post.

[1]: https://github.com/warewulf/warewulf.org/tree/main/src/posts
