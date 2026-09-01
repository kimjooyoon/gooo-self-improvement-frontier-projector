# v0.1.x release checklist

1. Merge the implementation through a passing pull request.
2. Confirm the post-merge `main` CI run and preserve its run, job, and artifact
   identities.
3. Create the annotated tag and keep it immutable.
4. Use the release workflow to create a draft release from the existing tag.
5. Publish the draft only after its ID and tag binding are recorded.
6. Read the release back through the GitHub API with the workflow-scoped
   `github.token` and require `draft=false` and `immutable=true`.
7. Preserve release ID, tag object, target commit, every asset ID, size, and
   SHA-256 digest in the final audit.

No failed run, pull request, tag, release, or artifact is deleted or rewritten.
