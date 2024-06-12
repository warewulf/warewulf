# Release procedure

Major versions (e.g., v4.5.0) are tagged directly on the main
branch. Minor versions (e.g., v4.5.3) are tagged on a separate release
branch.

1. Update `CHANGELOG.md`.
   - Identify the release date by changing "unreleased" to a date with
     `%Y-%m-%d` format, following existing convention in the log.
   - Do any final clean-up. (e.g., removing redundancy, adding issue
     or PR numbers).
   - Cherry-pick this to a release branch if necessary. (i.e., when
     not doing a new major release)
2. Create a signed tag for the release of the format v4.MINOR.PATCH,
   following the format specified in <MAINTAINING.md>. (e.g., `git tag
   --sign v4.5.3; git push origin v4.5.3`)
3. Monitor the release action associated with the pushed tag at
   https://github.com/warewulf/warewulf/actions, and verify the
   generated draft release contains the expected artifacts. This
   includes the source tarball and RPMs for Suse and Rocky Linux.
4. Update the release notes for the release, summarizing and expanding
   on the relevant contents from <CHANGELOG.md>.
5. Confirm the correct values for the pre-release and latest release
   flags.
6. Publish the release.
8. Announce the release as a [post][1] to warewulf.org/news, linking
   to the GitHub release.
9. Announce the release on the Warewulf Slack, linking to the
   warewulf.org/news post.

[1]: https://github.com/warewulf/warewulf.org/tree/main/src/posts
