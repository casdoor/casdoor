# WeCom Syncer

## Overview

The WeCom Syncer feature allows Casdoor to synchronize users from WeCom (WeChat Work) enterprise applications. This feature supports both WeCom Internal and Third-party applications.

## Features

- Automatic user synchronization from WeCom to Casdoor
- Periodic sync intervals (configurable)
- Support for WeCom Internal applications
- Support for WeCom Third-party applications
- User information synchronization (name, email, phone, avatar, etc.)

## API References

The WeCom Syncer uses the following WeCom APIs:

1. **User List API**: Fetches the list of user IDs
   - Endpoint: `https://qyapi.weixin.qq.com/cgi-bin/user/list_id`
   - Documentation: https://developer.work.weixin.qq.com/document/path/96021

2. **User Detail API**: Fetches detailed user information
   - Endpoint: `https://qyapi.weixin.qq.com/cgi-bin/user/get`
   - Documentation: https://developer.work.weixin.qq.com/document/path/90332

## Configuration

### WeCom Syncer Model

The WeCom syncer requires the following configuration:

- **Server Name**: A friendly name for the WeCom server
- **Corp ID**: Your WeCom enterprise ID
- **Corp Secret**: Your WeCom application secret
- **Department ID**: (Optional) Specific department to sync users from
- **Sub Type**: "Internal" or "Third-party"
- **Auto Sync**: Sync interval in minutes (0 to disable auto-sync)

### Database Schema

The WeCom syncer creates a `wecom` table with the following structure:

```go
type WeCom struct {
    Id          string // Unique identifier
    Owner       string // Organization owner
    CreatedTime string // Creation timestamp
    ServerName  string // Server display name
    CorpId      string // WeCom Corp ID
    CorpSecret  string // WeCom Corp Secret
    DepartmentId string // Department ID (optional)
    SubType     string // "Internal" or "Third-party"
    AutoSync    int    // Auto sync interval in minutes
    LastSync    string // Last sync timestamp
}
```

## API Endpoints

The WeCom syncer exposes the following REST API endpoints:

- `GET /api/get-wecom-users?id={wecom-id}` - Get users from WeCom
- `GET /api/get-wecoms?owner={owner}` - Get all WeCom configurations
- `GET /api/get-wecom?id={wecom-id}` - Get a specific WeCom configuration
- `POST /api/add-wecom` - Add a new WeCom configuration
- `POST /api/update-wecom` - Update an existing WeCom configuration
- `POST /api/delete-wecom` - Delete a WeCom configuration
- `POST /api/sync-wecom-users?id={wecom-id}` - Manually sync users

## Usage

### Adding a WeCom Syncer

1. Navigate to the Casdoor admin panel
2. Go to the WeCom Syncer section
3. Click "Add WeCom Syncer"
4. Fill in the required fields:
   - Server Name
   - Corp ID
   - Corp Secret
   - Sub Type (Internal or Third-party)
   - Auto Sync interval (optional)
5. Save the configuration

### Manual Sync

To manually trigger a user sync:

1. Go to the WeCom Syncer list
2. Select the syncer you want to sync
3. Click "Sync Users"
4. Review the sync results

### Auto Sync

If you configure the Auto Sync interval:

1. The syncer will automatically sync users at the specified interval
2. The sync runs in the background
3. Check the Last Sync timestamp to verify sync status

## User Mapping

The WeCom syncer maps WeCom user fields to Casdoor user fields as follows:

| WeCom Field | Casdoor Field |
|-------------|---------------|
| userid      | id            |
| name        | username, displayName |
| email       | email         |
| mobile      | phone         |
| avatar      | avatar        |

## Auto-Synchronization

The WeCom Auto-Synchronizer runs as a background process:

1. Starts automatically when Casdoor starts
2. Runs for each WeCom configuration with a non-zero Auto Sync interval
3. Updates existing users and creates new users
4. Logs sync results (new users, existing users, failed users)

## Error Handling

- Invalid credentials: Returns an error with the WeCom error code and message
- Network errors: Logged and retry on next sync interval
- Failed user syncs: Tracked and reported in the sync response

## Security Considerations

- Corp Secret is stored encrypted in the database
- Corp Secret is masked (shown as "***") in API responses
- Only administrators can configure WeCom syncers
- User synchronization respects organization boundaries

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   - Verify Corp ID and Corp Secret are correct
   - Check if the application has the required permissions

2. **No Users Synced**
   - Verify the Department ID is correct (if specified)
   - Check if users exist in the WeCom application
   - Review the WeCom API documentation for rate limits

3. **Auto Sync Not Working**
   - Verify Auto Sync interval is greater than 0
   - Check the Casdoor server logs for errors
   - Ensure the WeCom syncer service is running

## Implementation Details

### Code Structure

- `idp/wecom_syncer.go` - WeCom API client and syncer logic
- `object/wecom.go` - WeCom model and database operations
- `object/wecom_autosync.go` - Auto-sync background service
- `controllers/wecom.go` - REST API controllers
- `idp/wecom_syncer_test.go` - Unit tests

### Dependencies

- WeCom API (qyapi.weixin.qq.com)
- Existing WeCom IdP implementations (wecom_internal.go, wecom_third_party.go)

## Future Enhancements

Potential improvements for future versions:

- Support for syncing user groups/departments
- Support for bi-directional sync (Casdoor to WeCom)
- Advanced field mapping configuration
- Sync filtering based on user attributes
- Webhook support for real-time sync
- Detailed sync logs and audit trail

## References

- WeCom User Management API: https://developer.work.weixin.qq.com/document/path/90194
- WeCom Third-party Application: https://developer.work.weixin.qq.com/document/path/90594
- Casdoor Documentation: https://casdoor.org
