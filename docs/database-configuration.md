# Database Configuration Guide

## Transaction Pooling Mode

Casdoor now supports transaction pooling mode for compatibility with connection poolers like PgBouncer, Pgcat, and Odyssey in transaction mode.

### Problem

When using connection poolers in transaction mode, prepared statements cannot be reused across different transactions. This causes errors like:

```
pq: prepared statement does not exist
```

### Solution

Enable transaction pooling mode in your `conf/app.conf`:

```ini
enableTransactionPooling = true
```

When enabled, Casdoor will not cache prepared statements, making it compatible with transaction-mode connection poolers.

## Read/Write Database Splitting

Casdoor supports routing SELECT queries to a separate read-only database connection, enabling efficient read/write splitting.

### Use Cases

- Route SELECT queries to read replicas in the same availability zone
- Reduce latency for read-heavy workloads
- Better integrate with PostgreSQL HA setups (e.g., CNPG in Kubernetes)

### Configuration

Add a read-only database connection in your `conf/app.conf`:

```ini
# For PostgreSQL
readDataSourceName = user=readonly password=pass123 host=replica.example.com port=5432 sslmode=disable dbname=casdoor

# For MySQL
# Note: For MySQL, the dbName will be automatically appended
readDataSourceName = readonly:pass123@tcp(replica.example.com:3306)/
```

### How It Works

When `readDataSourceName` is configured:
- SELECT queries through `GetSession()` and `GetSessionForUser()` use the read engine
- INSERT, UPDATE, DELETE operations continue to use the primary database
- If read connection is not configured, all queries use the primary database

### Example: Kubernetes with CNPG

In a Kubernetes environment with CloudNativePG (CNPG):

```ini
# Primary database for writes
driverName = postgres
dataSourceName = user=app password=secret host=postgres-rw port=5432 sslmode=disable dbname=casdoor

# Read replica in same AZ for reads
readDataSourceName = user=app password=secret host=postgres-r port=5432 sslmode=disable dbname=casdoor

# Enable transaction pooling if using Pgcat or similar
enableTransactionPooling = true
```

### Example: Using PgBouncer

When using PgBouncer in transaction mode:

```ini
# Connect through PgBouncer
driverName = postgres
dataSourceName = user=app password=secret host=pgbouncer port=6432 sslmode=disable dbname=casdoor

# Enable transaction pooling for compatibility
enableTransactionPooling = true

# Optional: Use read-only PgBouncer pool for reads
readDataSourceName = user=app password=secret host=pgbouncer-ro port=6432 sslmode=disable dbname=casdoor
```

## Best Practices

1. **Performance Testing**: Test both configurations to ensure they meet your performance requirements
2. **Replication Lag**: Be aware of potential replication lag when using read replicas
3. **Monitoring**: Monitor connection pool usage and database performance
4. **Gradual Rollout**: Test in staging environment before production deployment

## Backward Compatibility

Both features are optional and disabled by default. Existing deployments will continue to work without any configuration changes:
- `enableTransactionPooling` defaults to `false`
- `readDataSourceName` defaults to empty (not configured)
