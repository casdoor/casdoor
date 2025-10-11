# LDAP Custom Attributes Feature - UI Screenshot Documentation

## Feature Location

The LDAP Custom Attributes mapping feature is available in the Casdoor Admin UI:

**Navigation Path**: `Providers` → `LDAP` → `Edit LDAP Provider`

## UI Components

### Custom Attributes Section

In the LDAP Edit page, you'll find a **"Custom attributes"** section that displays a table with the following interface:

```
┌─────────────────────────────────────────────────────────────────┐
│ Custom attributes                                                │
├─────────────────────────────────────────────────────────────────┤
│ [+ Add]                                                          │
│                                                                   │
│ ┌─────────────────────┬─────────────────────┬──────────┐        │
│ │ LDAP attribute name │ User property name  │ Action   │        │
│ ├─────────────────────┼─────────────────────┼──────────┤        │
│ │ department          │ department          │ [Delete] │        │
│ ├─────────────────────┼─────────────────────┼──────────┤        │
│ │ employeeNumber      │ employeeId          │ [Delete] │        │
│ ├─────────────────────┼─────────────────────┼──────────┤        │
│ │ title               │ jobTitle            │ [Delete] │        │
│ ├─────────────────────┼─────────────────────┼──────────┤        │
│ │ physicalDeliveryOf  │ office              │ [Delete] │        │
│ │ ficeName            │                     │          │        │
│ └─────────────────────┴─────────────────────┴──────────┘        │
│                                                                   │
│ [< 1 >]  Showing 1-4 of 4 items                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Component Features

1. **Add Button**: Adds a new row to define a custom attribute mapping
2. **LDAP attribute name**: Input field for the LDAP attribute (e.g., "department", "employeeNumber")
3. **User property name**: Input field for the corresponding Casdoor user property name
4. **Delete Button**: Removes a mapping row
5. **Pagination**: Shows 10 items per page with navigation controls

### Example Configuration

Here's what a typical configuration looks like:

```
LDAP Configuration Form
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Server Name:           My Company LDAP
Host:                  ldap.company.com
Port:                  389
Enable SSL:            ☐
Base DN:               DC=company,DC=com
Filter:                (objectClass=person)
Username:              CN=admin,DC=company,DC=com
Password:              •••••••••
Filter Fields:         [uid] [mail] [mobile]
Default Group:         [Select a group...]

Custom attributes:     
┌────────────────────────────────────────────────────┐
│ [+ Add]                                            │
│                                                    │
│ LDAP attribute → User property                    │
│ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━     │
│ department         → department         [Delete]   │
│ employeeNumber     → employeeId         [Delete]   │
│ title              → jobTitle           [Delete]   │
│ manager            → manager            [Delete]   │
│ extensionAttribute1→ costCenter         [Delete]   │
└────────────────────────────────────────────────────┘

Auto Sync:             60 mins
                       ⚠ The Auto Sync option will sync 
                         all users to specify organization

[Save] [Cancel]
```

## User Workflow

1. **Navigate** to the LDAP configuration page
2. **Click** the "Add" button in the Custom attributes section
3. **Enter** the LDAP attribute name from your directory schema
4. **Enter** the desired property name for the Casdoor user object
5. **Repeat** steps 2-4 for each custom attribute
6. **Click** "Save" to store the configuration
7. **Navigate** to the LDAP Sync page
8. **Click** "Sync" to import/update users with custom attributes

## Result

After synchronization, users will have their custom attributes stored in the `properties` field:

```json
{
  "owner": "organization1",
  "name": "john.doe",
  "displayName": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+1234567890",
  "properties": {
    "department": "Engineering",
    "employeeId": "EMP12345",
    "jobTitle": "Senior Software Engineer",
    "manager": "CN=Jane Smith,OU=Users,DC=company,DC=com",
    "costCenter": "CC-ENG-001"
  }
}
```

## Visual Flow Diagram

```
┌──────────────┐
│ LDAP Server  │
│              │
│ • department │
│ • empNumber  │
│ • title      │
│ • office     │
└──────┬───────┘
       │
       │ LDAP Query with
       │ Custom Attributes
       ▼
┌──────────────────────────────────────┐
│ Casdoor LDAP Sync                    │
│                                      │
│ Maps:                                │
│   department   → properties.dept     │
│   empNumber    → properties.empId    │
│   title        → properties.jobTitle │
│   office       → properties.location │
└──────┬───────────────────────────────┘
       │
       │ Creates/Updates User
       ▼
┌────────────────────────────┐
│ Casdoor User Object        │
│                            │
│ {                          │
│   name: "john.doe",        │
│   email: "john@...",       │
│   properties: {            │
│     dept: "Engineering",   │
│     empId: "12345",        │
│     jobTitle: "Engineer",  │
│     location: "NYC"        │
│   }                        │
│ }                          │
└────────────────────────────┘
```

## Code Reference

The UI is implemented in:
- **Component**: `/web/src/table/AttributesMapperTable.js`
- **Integration**: `/web/src/LdapEditPage.js` (line 287)
- **Backend API**: `/controllers/ldap.go`
- **Data Model**: `/object/ldap.go` and `/object/user.go`

## Testing

To test this feature:

1. Set up an LDAP test server (or use an existing LDAP directory)
2. Configure the LDAP provider in Casdoor with custom attribute mappings
3. Perform a manual sync
4. Query the user API to verify properties are populated
5. Use the properties in your application via User Info endpoint or JWT claims

## Browser Compatibility

The UI is built with React and Ant Design, supporting:
- Chrome/Edge (latest 2 versions)
- Firefox (latest 2 versions)
- Safari (latest 2 versions)

## Accessibility

- All form inputs have proper labels
- Keyboard navigation is supported
- Screen reader compatible
- Follows WCAG 2.1 guidelines
