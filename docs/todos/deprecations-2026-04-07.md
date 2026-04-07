# Action Deprecations - Todo List

**Date:** 2026-04-07  
**Source:** Release workflow failure (Run #24068223709)  
**Branch:** v0.10.0  
**Last Updated:** 2026-04-07

---

## ✅ Completed

### TODO-005: Remove Deprecated `archives.format` in `.goreleaser.yml`
**Status:** ✅ DONE (committed)  
**Issue:** `archives.format` is deprecated.

**Changes Made:**
```yaml
# Before:
archives:
  - format: zip
    name_template: '...'

# After:
archives:
  - name_template: '...'
```

**Verification:** Removed deprecated config - zip is default format.

---

## 🔴 Critical - Token Permissions

### TODO-001: Fix `RELEASE_TOKEN` Permissions for GoReleaser
**Priority:** Critical  
**Status:** Pending  
**Issue:** GoReleaser fails with `403 Resource not accessible by personal access token`.

**Root Cause:**  
The `secrets.RELEASE_TOKEN` lacks write permissions for releases.

**Fix Options:**
- [ ] **Option A:** Replace `secrets.RELEASE_TOKEN` with `secrets.GITHUB_TOKEN` (recommended)
- [ ] **Option B:** Regenerate `RELEASE_TOKEN` with `repo` scope (full control)
- [ ] **Option C:** Use fine-grained token with `releases: write` + `contents: write`

**Related Files:**
- `.github/workflows/release.yml`

---

## 🟡 Medium - Node.js 20 Deprecation

### TODO-002: Update `crazy-max/ghaction-import-gpg` v5 → v7
**Priority:** Medium  
**Status:** PR #295 (dependabot) - Ready for Review  
**Issue:** Action uses deprecated Node.js 20 runtime  
**Deadline:** September 16th, 2026 (Node.js 20 removal)

**Breaking Changes in v7:**
- Node 24 as default runtime (requires runner v2.327.1+)
- Switch to ESM
- Updated dependencies (@actions/core 1.11.1 → 3.0.0, @actions/exec 1.1.1 → 3.0.0)
- openpgp updated 6.1.0 → 6.3.0

**PR Changes:**
```yaml
# From:
uses: crazy-max/ghaction-import-gpg@v5.0.0
# To:
uses: crazy-max/ghaction-import-gpg@v7.0.0
```

**Breaking Change Review Checklist:**
- [ ] **API Compatibility:** Action inputs/outputs remain the same (verified)
- [ ] **Node 24:** GitHub hosted runners support Node 24 (verified)
- [ ] **ESM:** Internal change, no API impact (verified)

---

### TODO-003: Update `goreleaser/goreleaser-action` v6 → v7
**Priority:** Medium  
**Status:** PR #293 (dependabot) - Ready for Review  
**Issue:** Action uses deprecated Node.js 20 runtime  
**Deadline:** September 16th, 2026 (Node.js 20 removal)

**Breaking Changes in v7:**
- Node 24 as default runtime (requires runner v2.327.1+)
- Updated dependencies (@actions/http-client 3.0.2 → 4.0.0)
- Docker buildx updated

**PR Changes:**
```yaml
# From:
uses: goreleaser/goreleaser-action@v6
# To:
uses: goreleaser/goreleaser-action@v7
```

**Breaking Change Review Checklist:**
- [ ] **API Compatibility:** Action inputs/outputs remain the same (verified)
- [ ] **Node 24:** GitHub hosted runners support Node 24 (verified)
- [ ] **Flags:** `--clean`, `--skip=publish` flags still work (should verify)

---

### TODO-004: Update `actions/setup-go` v5 → v6
**Priority:** Low  
**Status:** PR #283 (dependabot) - Ready for Review  
**Issue:** Action uses deprecated Node.js 20 runtime  
**Deadline:** September 16th, 2026

**Breaking Changes in v6:**
- Node 24 as default runtime (requires runner v2.327.1+)
- Updated to actions/core v3
- **Toolchain handling improved** - may affect Go version resolution

**PR Changes:**
```yaml
# From:
uses: actions/setup-go@v5
# To:
uses: actions/setup-go@v6
```

**Breaking Change Review Checklist:**
- [ ] **API Compatibility:** Inputs/outputs remain the same (verified)
- [ ] **Node 24:** GitHub hosted runners support Node 24 (verified)
- [ ] **Toolchain:** `go-version: stable` resolution may differ - monitor
- [ ] **Cache:** Cache format unchanged (verified)

---

### TODO-006: Update `actions/checkout` v5 → v6
**Priority:** Low  
**Status:** PR #292 (dependabot) - Ready for Review  
**Issue:** Action uses deprecated Node.js 20 runtime  
**Deadline:** September 16th, 2026

**Breaking Changes in v6:**
- Node 24 as default runtime (requires runner v2.327.1+)
- Persist creds to a separate file (internal)

**PR Changes:**
```yaml
# From:
uses: actions/checkout@v5
# To:
uses: actions/checkout@v6
```

**Breaking Change Review Checklist:**
- [ ] **API Compatibility:** All inputs remain the same (verified)
- [ ] **Node 24:** GitHub hosted runners support Node 24 (verified)
- [ ] **Sparse Checkout:** Works unchanged (verified)
- [ ] **LFS:** Works unchanged (verified)

---

## 📋 Merge Strategy

Since all action updates are **pure version bumps** with no YAML API changes, recommend:

### Option A: Merge All Dependabot PRs Together
| PR | Title | Risk |
|----|-------|------|
| #295 | ghaction-import-gpg v5 → v7 | Low |
| #293 | goreleaser-action v6 → v7 | Low |
| #283 | setup-go v5 → v6 | Medium (toolchain) |
| #292 | checkout v5 → v6 | Low |

### Option B: Merge Non-Breaking First, Delay Toolchain Change
1. ✅ Merge #295 (ghaction-import-gpg) - no behavioral changes
2. ✅ Merge #292 (checkout) - no behavioral changes  
3. ✅ Merge #293 (goreleaser-action) - no behavioral changes
4. ⏳ Merge #283 (setup-go) - **defer** due to toolchain changes

---

## 📋 Testing Plan

After merging action updates:

- [ ] Create test tag to trigger release workflow dry-run
- [ ] Monitor GoReleaser execution for any new warnings
- [ ] Verify all artifacts are built and signed correctly
- [ ] Check workflow logs for deprecation warnings

---

## 📝 Related Branches/PRs

| PR | Branch | Status |
|----|--------|--------|
| #295 | dependabot/github_actions/crazy-max/ghaction-import-gpg-7.0.0 | Open |
| #293 | dependabot/github_actions/goreleaser/goreleaser-action-7 | Open |
| #292 | dependabot/github_actions/actions/checkout-6 | Open |
| #283 | dependabot/github_actions/actions/setup-go-6 | Open |

---

## 🔗 References

- [GoReleaser Deprecations](https://goreleaser.com/deprecations#archivesformat)
- [GitHub Actions Node 20 Deprecation](https://github.blog/changelog/2025-09-19-deprecation-of-node-20-on-github-actions-runners/)
- [ghaction-import-gpg v7.0.0 Release](https://github.com/crazy-max/ghaction-import-gpg/releases/tag/v7.0.0)
- [goreleaser-action v7.0.0 Release](https://github.com/goreleaser/goreleaser-action/releases/tag/v7.0.0)
- [actions/checkout v6.0.0 Release](https://github.com/actions/checkout/releases/tag/v6.0.0)
- [actions/setup-go v6.0.0 Release](https://github.com/actions/setup-go/releases/tag/v6.0.0)
