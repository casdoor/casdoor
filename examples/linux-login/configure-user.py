#!/usr/bin/env python3
"""
Configure Casdoor users for Linux machine login.

This script sets POSIX attributes (loginShell, sshPublicKey) for a user
via the Casdoor API, enabling Linux machine authentication via LDAP.
"""

import argparse
import json
import sys
from pathlib import Path
from typing import Optional

try:
    import requests
except ImportError:
    print("Error: requests library is required. Install with: pip install requests")
    sys.exit(1)


class CasdoorClient:
    """Client for interacting with Casdoor API."""

    def __init__(self, url: str, token: str, organization: str = "built-in"):
        self.url = url.rstrip("/")
        self.token = token
        self.organization = organization
        self.headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json",
        }

    def get_user(self, username: str) -> dict:
        """Fetch user data from Casdoor."""
        response = requests.get(
            f"{self.url}/api/get-user",
            params={"id": f"{self.organization}/{username}"},
            headers=self.headers,
        )
        response.raise_for_status()
        data = response.json()

        if data.get("status") == "error":
            raise Exception(f"Failed to fetch user: {data.get('msg')}")

        return data.get("data", {})

    def update_user(self, username: str, user_data: dict) -> dict:
        """Update user data in Casdoor."""
        response = requests.post(
            f"{self.url}/api/update-user",
            params={"id": f"{self.organization}/{username}"},
            headers=self.headers,
            json=user_data,
        )
        response.raise_for_status()
        data = response.json()

        if data.get("status") != "ok":
            raise Exception(f"Failed to update user: {data.get('msg')}")

        return data

    def configure_linux_login(
        self,
        username: str,
        login_shell: Optional[str] = None,
        ssh_public_key: Optional[str] = None,
    ) -> None:
        """Configure POSIX attributes for Linux login."""
        # Fetch current user data
        print(f"Fetching user data for: {username}")
        user = self.get_user(username)

        # Get or initialize properties
        properties = user.get("properties") or {}

        # Update properties
        if login_shell:
            properties["loginShell"] = login_shell
            print(f"  Setting loginShell: {login_shell}")

        if ssh_public_key:
            properties["sshPublicKey"] = ssh_public_key.strip()
            print(f"  Setting sshPublicKey: {len(ssh_public_key)} bytes")

        # Update user
        user["properties"] = properties
        print(f"Updating user: {username}")
        self.update_user(username, user)
        print("✓ Successfully configured user for Linux login")


def main():
    parser = argparse.ArgumentParser(
        description="Configure Casdoor users for Linux machine login",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Set custom shell
  %(prog)s -u john -s /bin/zsh -t $TOKEN

  # Add SSH public key
  %(prog)s -u john -k ~/.ssh/id_rsa.pub -t $TOKEN

  # Set both shell and SSH key
  %(prog)s -u john -s /bin/zsh -k ~/.ssh/id_rsa.pub -t $TOKEN

Environment Variables:
  CASDOOR_URL          Casdoor server URL (default: http://localhost:8000)
  CASDOOR_TOKEN        Casdoor access token
  ORGANIZATION         Organization name (default: built-in)
        """,
    )

    parser.add_argument(
        "-u", "--username", required=True, help="Username in Casdoor"
    )
    parser.add_argument(
        "-s",
        "--shell",
        default=None,
        help="Login shell (e.g., /bin/bash, /bin/zsh)",
    )
    parser.add_argument(
        "-k", "--ssh-key", type=Path, help="Path to SSH public key file"
    )
    parser.add_argument(
        "-t",
        "--token",
        help="Casdoor access token (or set CASDOOR_TOKEN env var)",
    )
    parser.add_argument(
        "--url",
        help="Casdoor server URL (default: from CASDOOR_URL env var or http://localhost:8000)",
    )
    parser.add_argument(
        "--org",
        help="Organization name (default: from ORGANIZATION env var or built-in)",
    )

    args = parser.parse_args()

    # Get configuration from args or environment
    import os

    url = args.url or os.getenv("CASDOOR_URL", "http://localhost:8000")
    token = args.token or os.getenv("CASDOOR_TOKEN")
    organization = args.org or os.getenv("ORGANIZATION", "built-in")

    if not token:
        print("Error: Access token is required (use --token or set CASDOOR_TOKEN)")
        sys.exit(1)

    # Read SSH key if provided
    ssh_public_key = None
    if args.ssh_key:
        if not args.ssh_key.exists():
            print(f"Error: SSH key file not found: {args.ssh_key}")
            sys.exit(1)
        ssh_public_key = args.ssh_key.read_text()

    # Configure user
    try:
        client = CasdoorClient(url, token, organization)
        client.configure_linux_login(
            args.username,
            login_shell=args.shell,
            ssh_public_key=ssh_public_key,
        )
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
