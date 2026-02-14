# Read Replica Database Configuration

## Overview

Casdoor supports configuring a separate read-only database connection for SELECT queries. This feature enables:

1. **Read/Write Splitting**: Route SELECT queries to read replicas while keeping write operations on the primary database
2. **Transaction Pooling Compatibility**: Better compatibility with PostgreSQL connection poolers like PgBouncer, Pgcat, or Odyssey in transaction mode
3. **Performance Optimization**: Reduce load on primary database by distributing read queries to replicas

## Configuration

### Basic Setup

Add the `readDataSourceName` configuration option to your `conf/app.conf` file:

```ini
# Primary database connection (for writes and reads when readDataSourceName is not set)
driverName = postgres
dataSourceName = user=casdoor password=secret host=primary.db.example.com port=5432 sslmode=disable dbname=
dbName = casdoor

# Optional: Read-only database connection for SELECT queries
readDataSourceName = user=casdoor password=secret host=replica.db.example.com port=5432 sslmode=disable dbname=
```

### Environment Variable

You can also set the read data source name using an environment variable:

```bash
export readDataSourceName="user=casdoor password=secret host=replica.db.example.com port=5432 sslmode=disable dbname="
```

### Database Types

The read replica feature works with all supported database types:

- **PostgreSQL**: `readDataSourceName = user=casdoor password=secret host=replica.example.com port=5432 sslmode=disable dbname=`
- **MySQL**: `readDataSourceName = casdoor:secret@tcp(replica.example.com:3306)/`
- **MSSQL**: `readDataSourceName = sqlserver://casdoor:secret@replica.example.com:1433?database=`

### Behavior

- **When `readDataSourceName` is configured**: 
  - All SELECT queries (read operations) use the read replica connection
  - All INSERT/UPDATE/DELETE queries (write operations) use the primary connection
  
- **When `readDataSourceName` is NOT configured or empty**:
  - All queries use the primary `dataSourceName` connection
  - Backward compatible with existing configurations

## Use Cases

### PostgreSQL with CNPG (CloudNative PostgreSQL)

In Kubernetes environments with CNPG, you can route reads to a read-only service in the same availability zone:

```ini
driverName = postgres
dataSourceName = user=casdoor password=secret host=casdoor-rw.namespace.svc.cluster.local port=5432 sslmode=disable dbname=
readDataSourceName = user=casdoor password=secret host=casdoor-r.namespace.svc.cluster.local port=5432 sslmode=disable dbname=
dbName = casdoor
```

### PostgreSQL with PgBouncer/Pgcat in Transaction Mode

Transaction pooling mode is more efficient but requires that prepared statements are not reused across transactions. Using read replicas with Casdoor helps distribute the load and works well with transaction pooling:

```ini
driverName = postgres
dataSourceName = user=casdoor password=secret host=pgbouncer-write.example.com port=5432 sslmode=disable dbname=
readDataSourceName = user=casdoor password=secret host=pgbouncer-read.example.com port=5432 sslmode=disable dbname=
dbName = casdoor
```

### MySQL Master-Replica Setup

```ini
driverName = mysql
dataSourceName = casdoor:secret@tcp(mysql-master.example.com:3306)/
readDataSourceName = casdoor:secret@tcp(mysql-replica.example.com:3306)/
dbName = casdoor
```

## Important Notes

1. **Replication Lag**: Be aware that read replicas may have replication lag. This means that data written to the primary may not be immediately available on the replica. For use cases requiring strong consistency, consider using the primary database for both reads and writes.

2. **Connection Pooling**: Both the primary and read replica connections benefit from connection pooling. Configure your database settings appropriately to handle the expected load.

3. **Failover**: If the read replica is unavailable, consider implementing failover logic at the database proxy/load balancer level to redirect reads to the primary.

4. **Session Affinity**: Some operations may require reading immediately after writing. The current implementation routes all reads through the read engine, so ensure your replica lag is minimal or use the primary for both if this is critical.

## Testing

To verify your read replica configuration is working:

1. Enable SQL logging: `showSql = true` in `conf/app.conf`
2. Start Casdoor and observe the connection logs
3. Perform read operations and verify they connect to the read replica
4. Perform write operations and verify they connect to the primary database

## Migration from Single Database

Migrating to read replica configuration is non-breaking:

1. Your existing configuration continues to work without changes
2. Add `readDataSourceName` when ready to enable read/write splitting
3. Remove or comment out `readDataSourceName` to revert to single database mode

## Performance Considerations

- Read replicas can significantly reduce load on the primary database
- Most Casdoor operations are reads (authentication, authorization checks, user queries)
- Typical read/write ratio in IAM systems is 90:10 or higher
- Distributing reads to replicas can improve overall system performance and scalability
