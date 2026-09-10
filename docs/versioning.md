# Versioning Standard

This document defines the versioning standard for the unlimit project following [Semantic Versioning 2.0.0](https://semver.org/).

## Version Format

Version numbers follow the format: `MAJOR.MINOR.PATCH`

Example: `v1.4.2`

## When to Increment

### MAJOR Version (X.0.0)

Increment the major version when you make **incompatible API changes** that break backward compatibility.

**Examples:**
- Removing a public function, method, or interface
- Changing function signatures (parameters, return types)
- Removing or renaming exported struct fields
- Changing the behavior of existing functions in a breaking way
- Removing support for a Go version
- Removing or renaming packages

**Code Example:**
```go
// v1.x.x
func ProcessData(data string) error

// v2.0.0 - Breaking change: added new required parameter
func ProcessData(data string, options *Options) error
```

### MINOR Version (x.Y.0)

Increment the minor version when you **add new functionality** in a backward-compatible manner.

**Examples:**
- Adding new public functions, methods, or interfaces
- Adding new optional parameters (with default values)
- Adding new fields to structs (if struct is not meant to be initialized directly)
- Adding new packages
- Deprecating functionality (but not removing it)
- Performance improvements that don't change API

**Code Example:**
```go
// v1.2.0
func ProcessData(data string) error

// v1.3.0 - New function added, existing API unchanged
func ProcessDataWithOptions(data string, opts Options) error
func ProcessData(data string) error // Still works
```

### PATCH Version (x.y.Z)

Increment the patch version when you make **backward-compatible bug fixes**.

**Examples:**
- Fixing bugs that don't change the API
- Internal refactoring without API changes
- Documentation updates
- Performance improvements without API changes
- Security patches that don't break compatibility
- Fixing incorrect behavior to match documentation

**Code Example:**
```go
// v1.2.3
// Fixed: function now correctly handles empty strings
func ProcessData(data string) error {
    if data == "" {
        return ErrEmptyData // Previously returned nil incorrectly
    }
    // ...
}
```

## Special Cases

### Pre-1.0.0 Versions (0.y.z)

During initial development (version 0.y.z), the API is considered unstable:
- Breaking changes may occur in MINOR versions (0.Y.0)
- The project is not yet production-ready
- Use `v0.y.z` format

### Pre-release Versions

For pre-release versions, append a hyphen and identifier:
- `v1.0.0-alpha`
- `v1.0.0-alpha.1`
- `v1.0.0-beta`
- `v1.0.0-rc.1` (release candidate)

### Build Metadata

Build metadata can be appended with a plus sign:
- `v1.0.0+20191201`
- `v1.0.0+exp.sha.5114f85`

## Decision Flow

Use this flowchart to determine version increment:

```
Did you change the public API?
├─ No → Is it a bug fix?
│  ├─ Yes → PATCH (x.y.Z)
│  └─ No → PATCH (x.y.Z) for docs/internal changes
│
└─ Yes → Does it break backward compatibility?
   ├─ Yes → MAJOR (X.0.0)
   └─ No → Is it new functionality?
      ├─ Yes → MINOR (x.Y.0)
      └─ No → PATCH (x.y.Z)
```

## Tagging Guidelines

### Git Tag Format

```bash
# Create annotated tag
git tag -a v1.4.2 -m "Release version 1.4.2"

# Push tag to remote
git push origin v1.4.2
```

### Release Notes

Each version tag should include release notes describing:
1. **Breaking Changes** (for MAJOR)
2. **New Features** (for MINOR)
3. **Bug Fixes** (for PATCH)
4. **Deprecations**
5. **Security Updates**

### Example Release Notes

```markdown
## v1.4.2 (2025-11-19)

### Bug Fixes
- Fixed memory leak in worker pool when tasks are cancelled
- Corrected error message formatting in validation errors

### Documentation
- Updated observe package examples
- Added versioning standard document
```

## Go Module Compatibility

For Go modules, follow these rules:

### v0 and v1
- Use `github.com/nvnamsss/unlimit` as module path
- Tags: `v0.1.0`, `v1.0.0`, `v1.2.3`

### v2 and Beyond
- Update module path to include `/v2`, `/v3`, etc.
- Module path: `github.com/nvnamsss/unlimit/v2`
- Tags: `v2.0.0`, `v2.1.0`

**Example:**
```go
// go.mod for v2
module github.com/nvnamsss/unlimit/v2

go 1.24
```

## Automation

Consider using these tools for version management:
- `git describe` - Generate version from tags
- GitHub Actions - Automate releases
- Conventional Commits - Automate version bumps from commit messages

## References

- [Semantic Versioning 2.0.0](https://semver.org/)
- [Go Modules Version Numbers](https://go.dev/doc/modules/version-numbers)
- [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github)
