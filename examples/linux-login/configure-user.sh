#!/bin/bash
# Helper script to configure Casdoor users for Linux machine login
# This script sets the required POSIX attributes for a user via Casdoor API

set -e

# Configuration - Update these values
CASDOOR_URL="${CASDOOR_URL:-http://localhost:8000}"
CASDOOR_CLIENT_ID="${CASDOOR_CLIENT_ID:-}"
CASDOOR_CLIENT_SECRET="${CASDOOR_CLIENT_SECRET:-}"
ORGANIZATION="${ORGANIZATION:-built-in}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Configure Casdoor user for Linux machine login by setting POSIX attributes.

OPTIONS:
    -h, --help              Show this help message
    -u, --username USER     Username in Casdoor (required)
    -s, --shell SHELL       Login shell (default: /bin/bash)
    -k, --ssh-key FILE      Path to SSH public key file
    -t, --token TOKEN       Casdoor access token (required if not using env var)

ENVIRONMENT VARIABLES:
    CASDOOR_URL            Casdoor server URL (default: http://localhost:8000)
    CASDOOR_TOKEN          Casdoor access token
    ORGANIZATION           Organization name (default: built-in)

EXAMPLES:
    # Set custom shell for user
    $0 --username john --shell /bin/zsh --token \$TOKEN

    # Add SSH public key for user
    $0 --username john --ssh-key ~/.ssh/id_rsa.pub --token \$TOKEN

    # Set both shell and SSH key
    $0 --username john --shell /bin/zsh --ssh-key ~/.ssh/id_rsa.pub --token \$TOKEN

NOTES:
    - You need a valid Casdoor access token with permission to update users
    - SSH public key should be in OpenSSH format (e.g., ssh-rsa AAAA...)
    - The user must already exist in Casdoor
EOF
    exit 1
}

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
USERNAME=""
LOGIN_SHELL="/bin/bash"
SSH_KEY_FILE=""
ACCESS_TOKEN="${CASDOOR_TOKEN:-}"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -u|--username)
            USERNAME="$2"
            shift 2
            ;;
        -s|--shell)
            LOGIN_SHELL="$2"
            shift 2
            ;;
        -k|--ssh-key)
            SSH_KEY_FILE="$2"
            shift 2
            ;;
        -t|--token)
            ACCESS_TOKEN="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate required parameters
if [ -z "$USERNAME" ]; then
    log_error "Username is required"
    usage
fi

if [ -z "$ACCESS_TOKEN" ]; then
    log_error "Access token is required (use --token or set CASDOOR_TOKEN env var)"
    usage
fi

# Read SSH key if provided
SSH_PUBLIC_KEY=""
if [ -n "$SSH_KEY_FILE" ]; then
    if [ ! -f "$SSH_KEY_FILE" ]; then
        log_error "SSH key file not found: $SSH_KEY_FILE"
        exit 1
    fi
    SSH_PUBLIC_KEY=$(cat "$SSH_KEY_FILE")
    log_info "Loaded SSH public key from $SSH_KEY_FILE"
fi

# Fetch current user data
log_info "Fetching current user data for: $USERNAME"
USER_DATA=$(curl -s -X GET \
    "${CASDOOR_URL}/api/get-user?id=${ORGANIZATION}/${USERNAME}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "Content-Type: application/json")

# Check if user exists
if echo "$USER_DATA" | grep -q '"status":"error"'; then
    log_error "Failed to fetch user: $(echo "$USER_DATA" | grep -o '"msg":"[^"]*"')"
    exit 1
fi

# Extract current properties
CURRENT_PROPERTIES=$(echo "$USER_DATA" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'data' in data and data['data']:
        props = data['data'].get('properties', {})
        print(json.dumps(props))
    else:
        print('{}')
except:
    print('{}')
" 2>/dev/null || echo "{}")

log_info "Current user properties: $CURRENT_PROPERTIES"

# Build updated properties using safer JSON handling
UPDATED_PROPERTIES=$(python3 << 'EOF'
import sys, json

# Read inputs safely from stdin
current_props_str = input()
login_shell = input()
ssh_key = input()

props = json.loads(current_props_str)
props['loginShell'] = login_shell

if ssh_key:
    props['sshPublicKey'] = ssh_key

print(json.dumps(props))
EOF
<< INPUTS
$CURRENT_PROPERTIES
$LOGIN_SHELL
$SSH_PUBLIC_KEY
INPUTS
)

log_info "Updating user with new properties..."

# Update user via API
UPDATE_PAYLOAD=$(python3 << 'EOF'
import sys, json

# Read inputs safely
user_data_str = input()
updated_props_str = input()

user_data = json.loads(user_data_str)
if 'data' in user_data:
    user_data = user_data['data']

user_data['properties'] = json.loads(updated_props_str)
print(json.dumps(user_data))
EOF
<< INPUTS
$USER_DATA
$UPDATED_PROPERTIES
INPUTS
)

RESPONSE=$(curl -s -X POST \
    "${CASDOOR_URL}/api/update-user?id=${ORGANIZATION}/${USERNAME}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "$UPDATE_PAYLOAD")

# Check response
if echo "$RESPONSE" | grep -q '"status":"ok"'; then
    log_info "Successfully updated user: $USERNAME"
    echo ""
    log_info "POSIX attributes configured:"
    echo "  - Login Shell: $LOGIN_SHELL"
    if [ -n "$SSH_PUBLIC_KEY" ]; then
        echo "  - SSH Public Key: Configured ($(echo "$SSH_PUBLIC_KEY" | wc -c) bytes)"
    fi
    echo ""
    log_info "User can now login to Linux machines configured with Casdoor LDAP"
else
    log_error "Failed to update user: $(echo "$RESPONSE" | grep -o '"msg":"[^"]*"' || echo "$RESPONSE")"
    exit 1
fi
