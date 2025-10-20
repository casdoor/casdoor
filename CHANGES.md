# Feature: Transaction Pooling Mode and Read/Write Database Splitting

## Overview

This feature adds support for two important database connectivity improvements:

1. **Transaction Pooling Mode** - Compatibility with connection poolers in transaction mode
2. **Read/Write Database Splitting** - Separate read and write database connections

## Motivation

### Transaction Pooling Mode

In Kubernetes environments with PostgreSQL, connection poolers like PgBouncer, Pgcat, and Odyssey are commonly used in transaction mode. However, Casdoor's use of prepared statements causes errors:

```
pq: prepared statement does not exist
```

This happens because prepared statements are session-scoped and cannot be reused across different connections in transaction pooling mode.

### Read/Write Database Splitting

In high-availability PostgreSQL setups (e.g., CloudNativePG in Kubernetes), it's common to:
- Route SELECT queries to read replicas in the same availability zone for reduced latency
- Keep INSERT/UPDATE/DELETE on the primary database
- Distribute read load across multiple replicas

Previously, Casdoor had no built-in support for this pattern, requiring users to modify the source code or use complex proxy configurations.

## Solution

### Configuration Options

Two new configuration options in `conf/app.conf`:

```ini
# Enable transaction pooling mode (default: false)
enableTransactionPooling = true

# Optional read-only database connection (default: empty)
readDataSourceName = user=readonly password=pass host=replica.example.com port=5432 sslmode=disable dbname=casdoor
```

### Implementation Details

1. **Transaction Pooling Mode**:
   - When enabled, `GetSession()` and `GetSessionForUser()` use `NewSession()` instead of `Prepare()`
   - This prevents prepared statement caching across transactions
   - Compatible with all transaction-mode connection poolers

2. **Read/Write Splitting**:
   - When `readDataSourceName` is configured, a separate read engine is initialized
   - `GetSession()` and `GetSessionForUser()` automatically use the read engine
   - Direct `ormer.Engine` calls continue to use the write engine (for write operations)
   - If read engine is not configured, all operations use the primary database

## Files Modified

1. **conf/conf.go** - Added configuration helper functions
2. **object/ormer.go** - Extended Ormer struct with read engine support
3. **object/ormer_session.go** - Updated session creation logic
4. **conf/app.conf** - Added configuration options with documentation
5. **object/ormer_test.go** - Created tests for new functionality
6. **conf/app.conf.example** - Created comprehensive example configuration
7. **docs/database-configuration.md** - Created detailed documentation

## Backward Compatibility

✅ **Fully backward compatible** - Both features are optional and disabled by default.

Existing deployments will continue to work without any changes:
- `enableTransactionPooling` defaults to `false` (uses prepared statements as before)
- `readDataSourceName` defaults to empty (uses primary database for all operations)

## Usage Examples

### Example 1: Kubernetes with CNPG

```ini
driverName = postgres
dataSourceName = user=app password=secret host=postgres-rw port=5432 sslmode=disable dbname=casdoor
readDataSourceName = user=app password=secret host=postgres-r port=5432 sslmode=disable dbname=casdoor
enableTransactionPooling = false
```

### Example 2: PgBouncer in Transaction Mode

```ini
driverName = postgres
dataSourceName = user=app password=secret host=pgbouncer port=6432 sslmode=disable dbname=casdoor
enableTransactionPooling = true
```

### Example 3: Combined (PgBouncer + Read Replica)

```ini
driverName = postgres
dataSourceName = user=app password=secret host=pgbouncer-rw port=6432 sslmode=disable dbname=casdoor
readDataSourceName = user=app password=secret host=pgbouncer-ro port=6432 sslmode=disable dbname=casdoor
enableTransactionPooling = true
```

## Testing

- ✅ Code builds successfully
- ✅ Passes `go fmt` and `go vet` checks
- ✅ Unit tests created for new functionality
- ✅ Backward compatibility verified
- ✅ Minimal changes (370 insertions, 7 deletions across 7 files)

## Performance Considerations

1. **Read Replicas**: Be aware of potential replication lag
2. **Connection Pooling**: Monitor connection pool usage
3. **Transaction Mode**: May have slight performance impact due to lack of prepared statement caching
4. **Network Latency**: Read replicas should be in the same availability zone for best performance

## Security Considerations

- Read-only credentials can be used for `readDataSourceName` to enforce least privilege
- Connection strings are properly sanitized through existing Docker replacement logic
- No new security vulnerabilities introduced

## Future Enhancements

Potential future improvements:
- Automatic retry logic for read replica connection failures
- Load balancing across multiple read replicas
- Configurable query routing rules
- Monitoring and metrics for read/write split effectiveness

## References

- Issue: Feature request for transaction mode DB connections and read/write splitting
- Related: PgBouncer, Pgcat, Odyssey transaction pooling
- Related: CloudNativePG (CNPG) read replicas in Kubernetes
