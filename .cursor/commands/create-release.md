# Create Release

Create a new release by determining the next semantic version from the latest tag and commits since then, then pushing a `v*` tag to trigger the release workflow. Releases are always created from the **main** branch.

## Steps

1. **Ensure on main branch**
   - Run: `git branch --show-current`
   - If the current branch is not `main`, tell the user that releases must be created from `main` and stop. Do not proceed.

2. **Fetch latest tags**
   - Run: `git fetch --tags origin` (so the latest release tag is known).

3. **Find the latest release tag**
   - Run: `git tag -l 'v*' --sort=-v:refname | head -1`
   - If no tags exist or the output is empty, the **previous version is 0.0.0** and the first release is **v0.0.1** (skip to step 6 with version 0.0.1).
   - Otherwise parse the tag (e.g. `v1.2.3`) to get the current **major**, **minor**, **patch** (strip the leading `v`).

4. **List commits since that tag**
   - Run: `git log <previous_tag>..HEAD --oneline --no-merges` (use the tag from step 3, e.g. `v1.2.3`).
   - If there are no commits, tell the user there’s nothing to release and stop.

5. **Determine semver bump from conventional commits**
   - Inspect the commit subjects (and bodies if needed) since the previous tag:
     - **Major**: any commit with a breaking change (subject contains `BREAKING CHANGE`, or type is followed by `!`, e.g. `feat!: ...` or `fix!: ...`).
     - **Minor**: at least one `feat` or `feat(scope):` commit and no major bump.
     - **Patch**: only `fix`, `docs`, `chore`, `refactor`, `style`, `test`, `perf`, `ci`, `build` (and no feat, no breaking).
   - Apply a single bump: at most one of major, minor, or patch.
   - Priority: if any commit is breaking → **major** (e.g. 1.2.3 → 2.0.0). Else if any commit is **feat** → **minor** (e.g. 1.2.3 → 1.3.0). Else → **patch** (e.g. 1.2.3 → 1.2.4).
   - Compute **next_version** as `major.minor.patch`.

6. **Present release plan and ask for confirmation**
   - Show the user: previous tag (or “none”), number of commits since then, chosen bump type, and **next_version** (e.g. `v1.3.0`).
   - Ask explicitly: "Create and push tag **vX.Y.Z** to trigger the release? (yes/no)"
   - **Do not** create or push the tag yet. Wait for the user to confirm (e.g. "yes", "y", "confirm").

7. **Tag and push (only after confirmation)**
   - If the user confirms: run `git tag v<next_version>`, then `git push origin v<next_version>`.
   - If the user declines or does not confirm: do not create or push; stop.
   - After a successful push, remind the user that the "Release CLI" workflow will run and that `CLERK_PUBLISHABLE_KEY` must be set in repo secrets. For npm install support, also set `NPM_AUTH_TOKEN` (npm publish is skipped if unset).

## Notes

- Releases are only allowed from **main**. The tag is created at the current `HEAD` on main.
- Tag format must be `v<major>.<minor>.<patch>` so the workflow (`tags: - "v*"`) runs.
- Do not push any commits; only create and push the new tag.
