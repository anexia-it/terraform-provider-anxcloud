## Summary

Updates GitHub Actions to latest versions to address Node.js 20 deprecation:

| Action | Old | New |
|--------|-----|-----|
| actions/checkout | v5 | v6 |
| actions/setup-go | v5 | v6 |
| crazy-max/ghaction-import-gpg | v5.0.0 | v7 |
| goreleaser/goreleaser-action | v6 | v7 |

Also fixes GoReleaser deprecation by removing deprecated `archives.format` (zip is default).

## Breaking Changes Review

All actions now use **Node.js 24** runtime (requires runner v2.327.1+). GitHub-hosted runners are compatible.

### Verified Non-Breaking:
- YAML API unchanged (pure version bumps)
- All inputs/outputs unchanged
- Cache formats unchanged
- Checkout sparse/LFS support unchanged

### Potential Concerns:
- **setup-go v6**: Improved toolchain handling may affect Go version resolution for `go-version: stable`

## Testing

- [x] Diff review completed
- [ ] Workflow dry-run test needed (can be done by pushing a test tag)
- [ ] Release workflow verification after merge

## References

- Original issue: Release workflow failure (Run #24068223709)
- Dependabot PRs: #292, #283, #295, #293
