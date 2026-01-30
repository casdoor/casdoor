# Linux Login Configuration Examples

This directory contains example configuration files for integrating Linux systems with Casdoor for user authentication and SSH access.

## Files

- **sssd.conf** - Example SSSD configuration file for Linux clients
  - Copy to `/etc/sssd/sssd.conf` on your Linux client
  - Customize CASDOOR_HOST and YOUR_ORG_NAME
  - Set permissions: `chmod 600 /etc/sssd/sssd.conf`

## Quick Start

### 1. Install SSSD on Linux Client

**RHEL/CentOS/Fedora:**
```bash
sudo yum install -y sssd sssd-ldap oddjob-mkhomedir
```

**Ubuntu/Debian:**
```bash
sudo apt-get install -y sssd sssd-ldap libnss-sss libpam-sss oddjob-mkhomedir
```

### 2. Configure SSSD

```bash
# Copy example configuration
sudo cp sssd.conf /etc/sssd/sssd.conf

# Edit with your Casdoor details
sudo nano /etc/sssd/sssd.conf
# Replace:
#   CASDOOR_HOST -> your Casdoor server (e.g., casdoor.example.com)
#   YOUR_ORG_NAME -> your organization name in Casdoor

# Set correct permissions
sudo chmod 600 /etc/sssd/sssd.conf
sudo chown root:root /etc/sssd/sssd.conf
```

### 3. Configure NSS

Edit `/etc/nsswitch.conf`:
```
passwd:     files sss
shadow:     files sss
group:      files sss
```

### 4. Enable Home Directory Creation

**RHEL/CentOS 8+:**
```bash
sudo authselect select sssd with-mkhomedir --force
```

**Ubuntu/Debian:**
Add to `/etc/pam.d/common-session`:
```
session optional pam_mkhomedir.so skel=/etc/skel umask=077
```

### 5. Configure SSH (Optional - for SSH key auth)

Edit `/etc/ssh/sshd_config`:
```
PubkeyAuthentication yes
AuthorizedKeysCommand /usr/bin/sss_ssh_authorizedkeys
AuthorizedKeysCommandUser nobody
```

Restart SSH:
```bash
sudo systemctl restart sshd
```

### 6. Start SSSD

```bash
sudo systemctl enable sssd
sudo systemctl start sssd
```

### 7. Verify

```bash
# Test user lookup
id your-casdoor-username

# Test SSH
ssh your-casdoor-username@localhost
```

## Full Documentation

See [../LINUX_LOGIN.md](../LINUX_LOGIN.md) for complete documentation including:
- Detailed configuration options
- Troubleshooting guide
- Security considerations
- Advanced features

## Support

- GitHub Issues: https://github.com/casdoor/casdoor/issues
- Documentation: https://casdoor.org
- Community Discord: https://discord.gg/5rPsrAzK7S
