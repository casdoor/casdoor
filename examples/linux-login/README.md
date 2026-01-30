# Linux Login Configuration Examples

This directory contains example configuration files and helper scripts for integrating Linux systems with Casdoor for user authentication and SSH access.

## Files

- **sssd.conf** - Example SSSD configuration file for Linux clients
  - Copy to `/etc/sssd/sssd.conf` on your Linux client
  - Customize CASDOOR_HOST and YOUR_ORG_NAME
  - Set permissions: `chmod 600 /etc/sssd/sssd.conf`

- **configure-user.sh** - Bash script to configure user POSIX attributes via API
  - Sets loginShell and sshPublicKey for users
  - Requires curl and python3

- **configure-user.py** - Python script to configure user POSIX attributes via API
  - Alternative to bash script for Python users
  - Requires: `pip install requests`

## Configuring Users for Linux Login

Before users can login to Linux machines, you need to set their POSIX attributes in Casdoor:

### Option 1: Using the Bash Script

```bash
# Set custom shell
./configure-user.sh --username john --shell /bin/zsh --token $CASDOOR_TOKEN

# Add SSH public key
./configure-user.sh --username john --ssh-key ~/.ssh/id_rsa.pub --token $CASDOOR_TOKEN

# Set both
./configure-user.sh --username john \
  --shell /bin/zsh \
  --ssh-key ~/.ssh/id_rsa.pub \
  --token $CASDOOR_TOKEN
```

### Option 2: Using the Python Script

```bash
# Install dependencies
pip install requests

# Configure user
./configure-user.py --username john --shell /bin/zsh --token $CASDOOR_TOKEN

# Add SSH key
./configure-user.py --username john --ssh-key ~/.ssh/id_rsa.pub --token $CASDOOR_TOKEN
```

### Option 3: Using the API Directly

```bash
curl -X POST "https://casdoor.example.com/api/update-user?id=org/username" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "org",
    "name": "username",
    "properties": {
      "loginShell": "/bin/zsh",
      "sshPublicKey": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ... user@host"
    }
  }'
```

### Getting an Access Token

To get an access token for the API:

1. Login to Casdoor via the web UI or API
2. Use OAuth2 client credentials flow
3. Or use the Casdoor SDK for your language

See [Casdoor API documentation](https://casdoor.org/docs/basic/public-api) for details.

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

See [../../LINUX_LOGIN.md](../../LINUX_LOGIN.md) for complete documentation including:
- Detailed configuration options
- Troubleshooting guide
- Security considerations
- Advanced features

## Support

- GitHub Issues: https://github.com/casdoor/casdoor/issues
- Documentation: https://casdoor.org
- Community Discord: https://discord.gg/5rPsrAzK7S
