# Casdoor v2 Migration Guide

## Overview

Starting from v2.87.0, Casdoor now follows Go's semantic import versioning standard. This means the module path has been updated to include the `/v2` suffix as required by Go modules for major versions ≥ v2.

## What Changed

### Module Path Update

The module path in `go.mod` has been updated from:
```
module github.com/casdoor/casdoor
```

to:
```
module github.com/casdoor/casdoor/v2
```

### Import Path Updates

All internal imports have been updated to use the `/v2` suffix. For example:
```go
// Before
import "github.com/casdoor/casdoor/object"

// After  
import "github.com/casdoor/casdoor/v2/object"
```

## Migration Guide for Users

If you are using Casdoor as a library or importing its packages in your Go projects, you need to update your imports.

### Step 1: Update Your Imports

Change all imports from `github.com/casdoor/casdoor` to `github.com/casdoor/casdoor/v2`:

```go
// Before
import (
    "github.com/casdoor/casdoor/object"
    "github.com/casdoor/casdoor/util"
)

// After
import (
    "github.com/casdoor/casdoor/v2/object"
    "github.com/casdoor/casdoor/v2/util"
)
```

### Step 2: Update Your go.mod

Run the following command to update your dependencies:
```bash
go get github.com/casdoor/casdoor/v2@latest
go mod tidy
```

## Why This Change?

According to Go's [module versioning specification](https://go.dev/doc/modules/version-numbers):

> For major version 2 or higher, the major version must be included as a `/vN` at the end of the module path.

This ensures that different major versions of a module can coexist in the same build, and prevents confusion about which version of a module is being used.

### Benefits

- **Compliance**: Follows Go's official semantic versioning rules
- **Clarity**: Makes it clear which major version is being imported
- **Compatibility**: Allows projects to use different major versions side-by-side if needed
- **Tool Support**: Better support from Go tools like `go get`, `go mod`, etc.

## Frequently Asked Questions

### Q: Will my existing code break?

If you're importing Casdoor packages in your Go code, you will need to update your import paths to include `/v2`. The functionality remains the same.

### Q: Can I still use older versions?

Yes, if you need to use versions before v2.0.0, you can continue using the old import path. However, it's recommended to upgrade to v2 for better Go ecosystem compatibility.

### Q: Do I need to change anything if I'm just running Casdoor as a service?

No, if you're running Casdoor as a standalone service (e.g., using Docker or binary releases), no changes are needed. This only affects Go developers importing Casdoor packages.

## References

- [Go Modules: Version Numbers](https://go.dev/doc/modules/version-numbers)
- [Go Modules: Major version suffixes](https://go.dev/doc/modules/major-version)
- [Semantic Versioning](https://semver.org/)
