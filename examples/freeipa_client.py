#!/usr/bin/env python3
"""Example FreeIPA-compatible JSON-RPC client for Casdoor"""

import requests

class CasdoorFreeIPAClient:
    def __init__(self, base_url, username=None, password=None, organization="built-in"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.organization = organization
        self.request_id = 0
        if username and password:
            self.login(username, password)
    
    def login(self, username, password):
        url = f"{self.base_url}/ipa/session/login_password"
        data = {'username': username, 'password': password, 'organization': self.organization}
        try:
            response = self.session.post(url, data=data)
            return response.json().get('status') == 'ok'
        except Exception as e:
            print(f"Login failed: {e}")
            return False

if __name__ == '__main__':
    print("Example FreeIPA client for Casdoor")
    print("See FREEIPA_INTEGRATION.md for usage details")
