# Casdoor Examples

This directory contains example code and scripts demonstrating how to integrate with Casdoor.

## FreeIPA-Compatible Client

### freeipa_client.py

A simple Python client demonstrating how to interact with Casdoor's FreeIPA-compatible JSON-RPC API.

**Requirements:**
```bash
pip install requests
```

**Usage:**
```python
from freeipa_client import CasdoorFreeIPAClient

# Create client and login
client = CasdoorFreeIPAClient(
    base_url="http://localhost:8000",
    username="admin",
    password="123",
    organization="built-in"
)

# Get user information
result = client.user_show("admin")
print(result)
```

For more details, see [FREEIPA_INTEGRATION.md](../FREEIPA_INTEGRATION.md) in the root directory.

## Contributing

To add more examples:

1. Create a new file in this directory
2. Add appropriate documentation
3. Update this README with a description of your example
4. Submit a pull request

## License

All examples in this directory are licensed under the Apache License 2.0, the same as Casdoor.
