# Test Data

This directory contains test fixtures for unit and integration tests.

## TLS Certificates

- `test_cert.pem`: Self-signed X.509 certificate for local testing
- `test_key.pem`: Private key for the test certificate

**⚠️ WARNING**: These are test fixtures only. Never use in production.

### Certificate Details

- **Subject**: CN=test.local, O=Test, C=BR
- **Validity**: 10 years (for test stability)
- **Key Type**: RSA 2048-bit

### Usage Example

```go
import (
    "crypto/tls"
    "path/filepath"
)

func loadTestTLSConfig() (*tls.Config, error) {
    cert, err := tls.LoadX509KeyPair(
        filepath.Join("internal", "testdata", "test_cert.pem"),
        filepath.Join("internal", "testdata", "test_key.pem"),
    )
    if err != nil {
        return nil, err
    }

    return &tls.Config{
        Certificates: []tls.Certificate{cert},
    }, nil
}
```

## Adding New Test Fixtures

When adding new test data:

1. Keep files small and focused
2. Document the purpose clearly
3. Never include real credentials or production data
4. Prefix filenames with `test_` for clarity
