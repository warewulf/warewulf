# Release procedure

1. Create a tag for the release of the format v4.MINOR.PATCH,
   following the format specified in <MAINTAINING.md>.
2. Verify that the associated release action has concluded.
3. Verify that the expected artifacts were built and successfully
   attached to the matching release. This includes the source tarball
   and RPMs for Suse and Rocky Linux.
4. Update the release notes for the release, expanding on the relevant
   contents from <CHANGELOG.md>.
5. Confirm the correct values for the pre-release and latest release
   flags.
6. Publish the release.
7. Announce the release as a [post][1] to warewulf.org/news, linking
   to the GitHub release.
8. Announce the release on the Warewulf Slack, linking to the
   warewulf.org/news post.

[1]: https://github.com/warewulf/warewulf.org/tree/main/src/posts
