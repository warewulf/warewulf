# Release procedure

Major versions (e.g., v4.6.0) are tagged directly on the main branch. Minor
versions (e.g., v4.6.1) are tagged on a separate release branch.

1. Update `CHANGELOG.md`.

   - Identify the release date by changing "unreleased" to a date with
     `%Y-%m-%d` format, following existing convention in the log.

   - Do any final clean-up. (e.g., remove redundancy, add issue or PR numbers,
     add missing changes)
     
         git log --graph --oneline --decorate origin/main origin/v4.6.x

2. Update `userdocs` references that imply the latest release to refer to the
   new version.

3. Add full release notes to `userdocs/release/` and update
   `userdocs/index.rst`.

4. Add summarized release notes to `.github/releases/`.

5. Cherry-pick updates to a release branch if necessary. (i.e., when not doing a
   new major release)
   
       git cherry-pick -x -m1 --signoff
   
6. Create a signed tag for the release of the format v4.MINOR.PATCH, following
   the format specified in <MAINTAINING.md>.
   
       git tag --sign v4.6.2; git push origin v4.6.2

7. Monitor the release action associated with the pushed tag at
   https://github.com/warewulf/warewulf/actions, and verify the generated draft
   release contains the expected artifacts. This includes the source tarball and
   RPMs for Suse and Rocky Linux.

8. Update the release notes for the release with the summary in
   `.github/releases/`.

9. Confirm the correct values for the pre-release and latest release flags.

10. Publish the release.

11. Announce the release as a [post][1] to [warewulf.org/news][2], linking to
   the GitHub release.

12. Announce the release on the Warewulf Slack, linking to the
    [warewulf.org/news][2] post.

[1]: https://github.com/warewulf/warewulf.org/tree/main/src/posts

[2]: https://warewulf.org/news
