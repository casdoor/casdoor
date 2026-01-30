# Linux Machine Login via Casdoor

This guide explains how to configure Linux systems (RHEL, Ubuntu, Debian, etc.) to authenticate users via Casdoor's LDAP server, similar to FreeIPA.

## Overview

Casdoor provides an LDAP server that supports POSIX attributes, allowing Linux machines to authenticate users through SSSD (System Security Services Daemon). This enables:

- Centralized user authentication via Casdoor
- SSH login using Casdoor credentials
- SSH key-based authentication
- Group-based access control
- Unified identity management

## Architecture

```
Linux Machine (SSSD) <--LDAP--> Casdoor LDAP Server <--> Casdoor Database
                                                            |
                                                            Users & Groups
```

## Prerequisites

### Casdoor Server Configuration

1. **Enable LDAP Server**: Configure Casdoor to run the LDAP server by setting in `conf/app.conf`:
   ```ini
   ldapServerPort = 389
   # For secure LDAP (recommended for production)
   ldapsServerPort = 636
   ldapsCertId = "your-cert-id"
   ```

2. **Configure Users**: Ensure users are created in Casdoor with the necessary attributes.

3. **Optional: Set User Properties** (via API or UI):
   - `loginShell`: Custom shell for user (default: `/bin/bash`)
   - `sshPublicKey`: SSH public key for key-based authentication

### Linux Client Requirements

- RHEL/CentOS 7+, Ubuntu 18.04+, Debian 10+, or similar
- Root access to configure system authentication
- Network connectivity to Casdoor LDAP server

## Linux Client Setup

### Step 1: Install Required Packages

**RHEL/CentOS/Fedora:**
```bash
sudo yum install -y sssd sssd-ldap oddjob-mkhomedir
```

**Ubuntu/Debian:**
```bash
sudo apt-get install -y sssd sssd-ldap libnss-sss libpam-sss oddjob-mkhomedir
```

### Step 2: Configure SSSD

Create or edit `/etc/sssd/sssd.conf`:

```ini
[sssd]
config_file_version = 2
services = nss, pam, ssh
domains = casdoor

[domain/casdoor]
# LDAP connection settings
id_provider = ldap
auth_provider = ldap
chpass_provider = ldap
access_provider = ldap

# Casdoor LDAP server details
ldap_uri = ldap://casdoor.example.com:389
# For LDAPS (recommended):
# ldap_uri = ldaps://casdoor.example.com:636
# ldap_id_use_start_tls = False

# Base DN - replace with your organization name
ldap_search_base = ou=your-org-name

# Bind credentials (optional, if anonymous bind is disabled)
# ldap_default_bind_dn = cn=admin,ou=your-org-name
# ldap_default_authtok_type = password
# ldap_default_authtok = your-bind-password

# User and group settings
ldap_user_object_class = posixAccount
ldap_user_name = uid
ldap_user_uid_number = uidNumber
ldap_user_gid_number = gidNumber
ldap_user_home_directory = homeDirectory
ldap_user_shell = loginShell
ldap_user_gecos = gecos
ldap_user_ssh_public_key = sshPublicKey

ldap_group_object_class = posixGroup
ldap_group_name = cn
ldap_group_gid_number = gidNumber
ldap_group_member = memberUid

# Access control (optional - allow all authenticated users)
ldap_access_order = filter
ldap_access_filter = (objectClass=posixAccount)

# Cache settings
cache_credentials = true
enumerate = false

# TLS settings (if using LDAPS)
# ldap_tls_reqcert = demand
# ldap_tls_cacert = /etc/pki/tls/certs/ca-bundle.crt

[nss]
filter_users = root
filter_groups = root

[pam]
offline_credentials_expiration = 2

[ssh]
ssh_authorizedkeys_command = /usr/bin/sss_ssh_authorizedkeys
ssh_authorizedkeys_command_user = nobody
```

**Important:** Replace `your-org-name` with your actual Casdoor organization name.

### Step 3: Set Correct Permissions

```bash
sudo chmod 600 /etc/sssd/sssd.conf
sudo chown root:root /etc/sssd/sssd.conf
```

### Step 4: Configure NSS (Name Service Switch)

Edit `/etc/nsswitch.conf` and modify the following lines:

```
passwd:     files sss
shadow:     files sss
group:      files sss
```

### Step 5: Configure PAM for Home Directory Creation

Edit `/etc/pam.d/common-session` (Debian/Ubuntu) or `/etc/pam.d/system-auth` (RHEL/CentOS):

Add the following line:
```
session optional pam_mkhomedir.so skel=/etc/skel umask=077
```

Or use authconfig/authselect:

**RHEL/CentOS 7:**
```bash
sudo authconfig --enablesssd --enablesssdauth --enablemkhomedir --update
```

**RHEL/CentOS 8+:**
```bash
sudo authselect select sssd with-mkhomedir --force
```

### Step 6: Configure SSH for Public Key Authentication

Edit `/etc/ssh/sshd_config`:

```
# Enable public key authentication
PubkeyAuthentication yes

# Enable SSSD to provide SSH keys
AuthorizedKeysCommand /usr/bin/sss_ssh_authorizedkeys
AuthorizedKeysCommandUser nobody
```

Restart SSH service:
```bash
sudo systemctl restart sshd
```

### Step 7: Start and Enable SSSD

```bash
sudo systemctl enable sssd
sudo systemctl start sssd
```

## Verification

### Test User Lookup

```bash
# Look up a Casdoor user
id username

# Expected output:
# uid=1234567890(username) gid=1234567890 groups=1234567890

# Get user information
getent passwd username

# Get group information
getent group groupname
```

### Test SSH Login

```bash
# SSH to the local machine
ssh username@localhost

# Or from another machine
ssh username@linux-host.example.com
```

### Test SSH Key Authentication

1. Add an SSH public key to a user in Casdoor:
   - Via API: Set user property `sshPublicKey` to the public key content
   - Via UI: Add to user properties (if UI supports it)

2. Test SSH login without password:
   ```bash
   ssh -i /path/to/private/key username@linux-host.example.com
   ```

### Troubleshooting

**Check SSSD logs:**
```bash
sudo tail -f /var/log/sssd/sssd_casdoor.log
```

**Increase SSSD debug level** in `/etc/sssd/sssd.conf`:
```ini
[domain/casdoor]
debug_level = 9
```

**Clear SSSD cache:**
```bash
sudo sss_cache -E
sudo systemctl restart sssd
```

**Test LDAP connectivity:**
```bash
# Install ldap-utils
sudo apt-get install ldap-utils  # Debian/Ubuntu
sudo yum install openldap-clients  # RHEL/CentOS

# Test LDAP search
ldapsearch -x -H ldap://casdoor.example.com:389 \
  -b "ou=your-org-name" \
  "(objectClass=posixAccount)"
```

## Casdoor User Configuration

### Setting User Properties via API

Use the Casdoor API to set user properties:

```bash
curl -X POST "https://casdoor.example.com/api/update-user" \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "your-org-name",
    "name": "username",
    "properties": {
      "loginShell": "/bin/zsh",
      "sshPublicKey": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ... user@host"
    }
  }'
```

### Default Values

If user properties are not set:
- **loginShell**: Defaults to `/bin/bash`
- **sshPublicKey**: Empty (password authentication only)
- **gecos**: Uses user's display name or username

## Advanced Configuration

### Group-Based Access Control

Restrict SSH access to specific groups by modifying `/etc/sssd/sssd.conf`:

```ini
[domain/casdoor]
ldap_access_order = filter
ldap_access_filter = (memberOf=cn=ssh-users,ou=your-org-name)
```

### Sudo Integration

For sudo support, you would need to implement LDAP-based sudo rules. This requires additional Casdoor development to support the `sudoRole` LDAP schema.

### Static UID/GID Mapping

By default, Casdoor uses a hash function to generate consistent UIDs/GIDs. For custom UID/GID assignment:

1. Set user properties `uidNumber` and `gidNumber` in Casdoor
2. Modify the LDAP server code to read these properties instead of hashing

## Security Considerations

1. **Use LDAPS (TLS)** in production environments
2. **Restrict LDAP bind access** - configure bind DN and password
3. **Use SSH keys** instead of passwords when possible
4. **Enable firewall rules** to restrict LDAP port access
5. **Regular security updates** for SSSD and related packages
6. **Monitor authentication logs** in `/var/log/secure` or `/var/log/auth.log`

## Comparison with FreeIPA

| Feature | FreeIPA | Casdoor |
|---------|---------|---------|
| LDAP Directory | ✅ | ✅ |
| POSIX Attributes | ✅ | ✅ |
| SSH Key Management | ✅ | ✅ |
| Kerberos SSO | ✅ | ❌ (Future) |
| Sudo Rules | ✅ | ❌ (Future) |
| Web UI | ✅ | ✅ |
| OAuth/OIDC/SAML | Limited | ✅ |
| Multi-tenancy | Limited | ✅ |
| Cloud-Native | ❌ | ✅ |

## Future Enhancements

Potential improvements to further enhance Linux integration:

1. **Kerberos Integration**: Add Kerberos KDC for SSO support
2. **Sudo Rules**: Implement `sudoRole` LDAP schema
3. **Password Policies**: LDAP-based password expiration and complexity
4. **Home Directory Management**: Automatic home directory creation on first login
5. **Static UID/GID**: UI for assigning specific numeric IDs to users/groups
6. **HBAC (Host-Based Access Control)**: Control which users can access which hosts

## Support

For issues and questions:
- GitHub Issues: https://github.com/casdoor/casdoor/issues
- Documentation: https://casdoor.org
- Community: https://discord.gg/5rPsrAzK7S

## References

- [SSSD Documentation](https://sssd.io/)
- [LDAP POSIX Schema](https://www.ietf.org/rfc/rfc2307.txt)
- [FreeIPA Documentation](https://www.freeipa.org/page/Documentation)
- [Casdoor Documentation](https://casdoor.org/docs/)
