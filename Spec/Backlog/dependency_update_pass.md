# Dependency Update Pass - Backlog Item

## Problem Description

The Terraform provider for Anexia Cloud has multiple outdated dependencies that need systematic updating. Current analysis shows 96 dependencies with available updates across both the main project and tools directory. Some updates are security-critical while others provide performance improvements and bug fixes.

## Current State Analysis

### Dependency Categories and Update Status

#### 🔴 **CRITICAL SECURITY UPDATES** (Immediate Action Required)
- **golang.org/x/crypto**: v0.41.0 → v0.45.0
  - **Security Impact**: Fixes CVE-2025-58181 (unbounded memory consumption in SSH) and CVE-2025-47914 (DoS in SSH agent)
  - **Risk**: High (4-version jump) but **CRITICAL** for security
  - **Action**: Update immediately with extensive testing

#### 🟡 **HashiCorp Terraform Plugin Ecosystem** (Medium Risk)
- **terraform-plugin-framework**: v1.15.0 → v1.17.0
- **terraform-plugin-go**: v0.28.0 → v0.29.0
- **terraform-plugin-mux**: v0.20.0 → v0.21.0
- **terraform-plugin-sdk/v2**: v2.37.0 (no update available)
- **terraform-exec**: v0.23.0 → v0.24.0
- **terraform-json**: v0.25.0 → v0.27.2
- **Risk**: Medium - Minor versions should be backward compatible
- **Action**: Update all together in next minor release

#### 🟢 **Anexia SDK** (Low Risk)
- **go.anx.io/go-anxcloud**: v0.9.1-alpha → v0.9.2-alpha
- **Risk**: Low - Alpha version with likely bug fixes
- **Action**: Update after reviewing changelog

#### 🟡 **golang.org/x Ecosystem** (Medium Risk)
- **x/net**: v0.43.0 → v0.47.0 (includes security fixes)
- **x/sys**: v0.35.0 → v0.38.0
- **x/tools**: v0.36.0 → v0.39.0
- **x/crypto**: (see critical above)
- **x/oauth2, x/term, x/text, x/time, x/mod, x/sync**: Various updates
- **Risk**: Medium - Large ecosystem with potential interactions
- **Action**: Update together after security fixes

#### 🟢 **Kubernetes Client** (Low Risk)
- **k8s.io/client-go**: v0.34.0 → v0.34.2
- **k8s.io/api, apimachinery**: Matching patch updates
- **Risk**: Low - Patch versions
- **Action**: Update with other Kubernetes packages

#### 🟡 **GitHub Actions** (Low-Medium Risk)
- **actions/checkout**: v5 → v6.0.1 available
- **actions/setup-go**: v5 → v6.1.0 available
- **goreleaser/goreleaser-action**: v6 (current)
- **crazy-max/ghaction-import-gpg**: v5.0.0 (current)
- **Risk**: Low - Generally backward compatible
- **Action**: Update in maintenance cycle

#### 🟡 **Development Tools** (tools/ directory)
- **golangci-lint**: v1.62.2 → v1.64.8
- **terraform-plugin-docs**: v0.20.1 (check for updates)
- **terrafmt**: v0.5.5 (current)
- **Risk**: Low-Medium - Tool updates
- **Action**: Update as needed for new features

## Dependabot Configuration Status

✅ **Properly Configured** for:
- Go modules (main project)
- Go modules (tools directory)
- Docker images
- GitHub Actions (monthly schedule)

## Proposed Update Strategy

### Phase 1: Critical Security Fixes (Immediate - Hotfix Release)
1. Update `golang.org/x/crypto` to v0.45.0
2. Run full test suite including acceptance tests
3. Release as patch/hotfix version

### Phase 2: HashiCorp Ecosystem (Next Minor Release)
1. Update all HashiCorp Terraform packages together
2. Verify Terraform 1.14+ compatibility (action{} blocks)
3. Run acceptance tests against integration environment
4. Update documentation if needed

### Phase 3: golang.org/x Ecosystem (Following Release)
1. Update all golang.org/x packages
2. Test network, crypto, and system interactions
3. Verify performance improvements

### Phase 4: Remaining Updates (Maintenance)
1. Update Kubernetes packages
2. Update GitHub Actions
3. Update development tools
4. Clean up go.mod/go.sum files

## Testing Requirements

- **Unit Tests**: All existing tests must pass
- **Acceptance Tests**: Run against integration-1 environment with ANEXIA_TOKEN
- **Integration Tests**: Verify API compatibility
- **Performance Tests**: Check for regressions in resource operations
- **Documentation**: Update any version-specific documentation

## Risk Mitigation

1. **Staged Updates**: Update related packages together
2. **Comprehensive Testing**: Full test suite before each release
3. **Rollback Plan**: Keep previous working versions documented
4. **Gradual Rollout**: Start with security fixes, then feature updates

## Success Criteria

- ✅ All security vulnerabilities resolved
- ✅ No breaking changes for users
- ✅ All tests passing
- ✅ Performance maintained or improved
- ✅ Dependencies up-to-date within 6 months
- ✅ Dependabot PRs can be cleanly merged

## Estimated Timeline

- **Phase 1**: 1-2 days (security critical)
- **Phase 2**: 3-5 days (Terraform ecosystem)
- **Phase 3**: 2-3 days (Go ecosystem)
- **Phase 4**: 1-2 days (maintenance)

## Dependencies

- Access to integration-1 testing environment
- ANEXIA_TOKEN for acceptance tests
- Review of anxcloud SDK v0.9.2-alpha changelog
- Coordination with release process</content>
<parameter name="filePath">Spec/Backlog/dependency_update_pass.md