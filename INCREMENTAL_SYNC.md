# Incremental Sync Feature

## Overview

The incremental sync feature optimizes the performance of periodic synchronization when dealing with large user datasets (e.g., 50,000+ users). Instead of fetching and comparing all users on every sync cycle, incremental sync only retrieves users that have been modified since the last successful synchronization.

## How It Works

### Automatic Detection

The system automatically determines whether to use incremental sync based on two conditions:

1. **LastSyncTime exists**: The syncer has a recorded timestamp from a previous successful sync
2. **UpdatedTime column exists**: The external database table has a column mapped to the User's `UpdatedTime` field

When both conditions are met, the sync will only fetch users where `updated_time > last_sync_time`.

### First Sync

On the first sync (or when `LastSyncTime` is empty), the system performs a full sync:
- Fetches all users from the external database
- Compares them with local users
- Records the sync timestamp for future incremental syncs

### Subsequent Syncs

On subsequent syncs, if incremental sync is possible:
- Only fetches users modified since the last sync from the external database
- Compares them with local users
- Updates or adds changed users
- Updates the `LastSyncTime` for the next sync

## Configuration

### Database Setup

To enable incremental sync, your external database table must include a timestamp column that tracks when each user record was last modified. Map this column to Casdoor's `UpdatedTime` field in the syncer configuration.

Example table column mappings:
```json
{
  "tableColumns": [
    {
      "name": "id",
      "casdoorName": "Id",
      "isKey": true
    },
    {
      "name": "updated_at",
      "casdoorName": "UpdatedTime"
    }
    // ... other columns
  ]
}
```

### Fallback to Full Sync

The system automatically falls back to full sync when:
- `LastSyncTime` is not set (first sync or after reset)
- The external table doesn't have an `UpdatedTime` column
- The `UpdatedTime` column is not mapped in the syncer configuration

## Performance Benefits

For large user databases:
- **Reduced Database Load**: Only queries modified records instead of all users
- **Lower CPU Usage**: Processes fewer records on each sync cycle
- **Decreased Memory Consumption**: Smaller datasets in memory during comparison
- **Faster Sync Times**: Proportional to the number of changed users, not total users

### Example Performance Improvement

With 50,000 users where only 100 are modified per hour:
- **Full Sync**: Queries 50,000 users, processes 50,000 comparisons
- **Incremental Sync**: Queries 100 users, processes 100 comparisons
- **Improvement**: ~500x reduction in processing load

## Backward Compatibility

This feature is fully backward compatible:
- Existing syncers without `UpdatedTime` column mapping continue to work with full sync
- No configuration changes required for syncers that don't need incremental sync
- The `LastSyncTime` field is automatically added to the syncer table schema

## Monitoring

The sync process logs whether it's using incremental or full sync:
```
Running syncUsers()..
Using incremental sync (last sync: 2024-01-15T10:30:00Z)
Users: 50000, oUsers: 100
```

Or for full sync:
```
Running syncUsers()..
Using full sync
Users: 50000, oUsers: 50000
```

## Troubleshooting

### Incremental Sync Not Activating

If incremental sync isn't being used:
1. Verify the external table has a timestamp column for updates
2. Check that the column is mapped to `UpdatedTime` in the syncer's `tableColumns`
3. Ensure at least one successful sync has completed (to set `LastSyncTime`)

### Force Full Sync

To force a full sync (e.g., after data migration or corruption):
1. Set the syncer's `LastSyncTime` to empty string via API or database
2. The next sync will perform a full sync and reset the timestamp

## Technical Details

### Database Schema Change

A new field `last_sync_time` (varchar(100)) is added to the `syncer` table to track the timestamp of the last successful synchronization.

### SQL Query Example

When incremental sync is active, the query looks like:
```sql
SELECT * FROM user_table WHERE updated_at > '2024-01-15T10:30:00Z'
```

Instead of:
```sql
SELECT * FROM user_table
```

### Time Format

The `LastSyncTime` uses Casdoor's standard time format from `util.GetCurrentTime()`, which is ISO 8601 compatible.
