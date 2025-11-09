# 🧪 Estratégia de Testes - {{PROJECT_NAME}}

Estratégia completa de testes implementada no projeto **{{PROJECT_NAME}}**.

---

## 🎯 Pirâmide de Testes

```
                    ┌─────────────────┐
                    │   E2E Tests     │ <- 10%
                    │   (Cypress)     │
                ┌───┴─────────────────┴───┐
                │   Integration Tests     │ <- 20%
                │   (API, DB, Cache)      │
            ┌───┴─────────────────────────┴───┐
            │       Unit Tests                │ <- 70%
            │   (Functions, Classes)          │
            └─────────────────────────────────┘
```

### 📊 Cobertura Target
- **Total**: 95%+ cobertura
- **Unit Tests**: 98%+ cobertura
- **Integration**: 90%+ cobertura
- **E2E**: Critical paths (100%)

---

## 🔬 Tipos de Testes Implementados

### 1️⃣ Unit Tests
Testa componentes isolados

```{{LANGUAGE_LOWER}}
// Exemplo: Teste de função pura
func TestCalculateDiscount(t *testing.T) {
    tests := []struct {
        name     string
        price    float64
        discount float64
        expected float64
    }{
        {"10% discount", 100.0, 0.10, 90.0},
        {"25% discount", 200.0, 0.25, 150.0},
        {"No discount", 100.0, 0.0, 100.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateDiscount(tt.price, tt.discount)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2️⃣ Integration Tests
Testa interação entre componentes

```{{LANGUAGE_LOWER}}
// Exemplo: Teste de integração com DB
func TestUserRepository_Create(t *testing.T) {
    // Setup
    db := setupTestDatabase()
    repo := NewUserRepository(db)

    user := &User{
        Email: "test@example.com",
        Name:  "Test User",
    }

    // Execute
    createdUser, err := repo.Create(user)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, createdUser.ID)
    assert.Equal(t, user.Email, createdUser.Email)

    // Cleanup
    cleanupTestDatabase(db)
}
```

### 3️⃣ API Tests
Testa endpoints HTTP completos

```{{LANGUAGE_LOWER}}
func TestCreateUser_API(t *testing.T) {
    // Setup
    app := setupTestApp()

    payload := `{
        "email": "test@example.com",
        "name": "Test User"
    }`

    req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer " + getTestToken())

    w := httptest.NewRecorder()

    // Execute
    app.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", response["email"])
}
```

### 4️⃣ Performance Tests
Testa performance e carga

```{{LANGUAGE_LOWER}}
func BenchmarkCalculateDiscount(b *testing.B) {
    for i := 0; i < b.N; i++ {
        CalculateDiscount(100.0, 0.15)
    }
}

func TestConcurrentRequests(t *testing.T) {
    app := setupTestApp()

    // 100 requests concorrentes
    var wg sync.WaitGroup
    results := make(chan int, 100)

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()

            req := httptest.NewRequest("GET", "/api/v1/users", nil)
            w := httptest.NewRecorder()
            app.ServeHTTP(w, req)

            results <- w.Code
        }()
    }

    wg.Wait()
    close(results)

    // Verificar que todos retornaram 200
    for code := range results {
        assert.Equal(t, http.StatusOK, code)
    }
}
```

### 5️⃣ Security Tests
Testa vulnerabilidades de segurança

```{{LANGUAGE_LOWER}}
func TestSQLInjectionPrevention(t *testing.T) {
    app := setupTestApp()

    // Tentativa de SQL injection
    maliciousPayload := `'; DROP TABLE users; --`

    req := httptest.NewRequest("GET", "/api/v1/users?name=" + maliciousPayload, nil)
    w := httptest.NewRecorder()

    app.ServeHTTP(w, req)

    // Sistema deve retornar erro 400, não 500
    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUnauthorizedAccess(t *testing.T) {
    app := setupTestApp()

    // Request sem token
    req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
    w := httptest.NewRecorder()

    app.ServeHTTP(w, req)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

### 6️⃣ E2E Tests (Cypress/Playwright)
Testa fluxos completos do usuário

```javascript
// cypress/integration/{{entity}}_flow.spec.js
describe('{{Entity}} Management Flow', () => {
  beforeEach(() => {
    cy.login('test@example.com', 'secure_test_password');
  });

  it('should create, edit and delete {{entity}}', () => {
    // Create
    cy.visit('/{{entities}}');
    cy.get('[data-testid="create-{{entity}}"]').click();
    cy.get('[data-testid="{{entity}}-name"]').type('Test {{Entity}}');
    cy.get('[data-testid="save-{{entity}}"]').click();
    cy.contains('{{Entity}} created successfully');

    // Edit
    cy.get('[data-testid="edit-{{entity}}"]').click();
    cy.get('[data-testid="{{entity}}-name"]').clear().type('Updated {{Entity}}');
    cy.get('[data-testid="save-{{entity}}"]').click();
    cy.contains('{{Entity}} updated successfully');

    // Delete
    cy.get('[data-testid="delete-{{entity}}"]').click();
    cy.get('[data-testid="confirm-delete"]').click();
    cy.contains('{{Entity}} deleted successfully');
  });
});
```

---

## 🛠️ Ferramentas de Teste

### 🔧 Unit & Integration Testing
- **Framework**: {{TEST_FRAMEWORK}}
- **Assertions**: {{ASSERTION_LIBRARY}}
- **Mocking**: {{MOCK_LIBRARY}}
- **Coverage**: {{COVERAGE_TOOL}}

### 🌐 API Testing
- **HTTP Testing**: {{HTTP_TEST_LIBRARY}}
- **Database**: Test containers / In-memory DB
- **Authentication**: JWT test tokens
- **Rate Limiting**: Test with multiple requests

### ⚡ Performance Testing
- **Load Testing**: {{LOAD_TEST_TOOL}}
- **Benchmarking**: Built-in benchmark tools
- **Profiling**: {{PROFILING_TOOL}}
- **Memory**: Memory leak detection

### 🔒 Security Testing
- **SAST**: {{SAST_TOOL}} (Static Analysis)
- **DAST**: {{DAST_TOOL}} (Dynamic Analysis)
- **Dependency**: {{DEPENDENCY_SCAN_TOOL}}
- **Secrets**: {{SECRET_SCAN_TOOL}}

---

## 📊 Test Data Management

### Test Database
```{{LANGUAGE_LOWER}}
// Setup test database para cada teste
func setupTestDatabase() *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        log.Fatal(err)
    }

    // Run migrations
    runMigrations(db)

    // Seed test data
    seedTestData(db)

    return db
}

func seedTestData(db *sql.DB) {
    // Insert fixture data
    users := []User{
        {Email: "admin@test.com", Role: "admin"},
        {Email: "user@test.com", Role: "user"},
    }

    for _, user := range users {
        createUser(db, user)
    }
}
```

### Fixtures & Factories
```{{LANGUAGE_LOWER}}
// User factory para testes
func UserFactory() User {
    return User{
        Email:     fmt.Sprintf("user%d@test.com", rand.Int()),
        Name:      "Test User",
        CreatedAt: time.Now(),
        Active:    true,
    }
}

// Customize factory
func AdminUserFactory() User {
    user := UserFactory()
    user.Role = "admin"
    return user
}
```

---

## 🎯 Test Organization

### Estrutura de Pastas
```
tests/
├── unit/                  # Testes unitários
│   ├── entities/         # Testes de entidades
│   ├── usecases/         # Testes de casos de uso
│   └── utils/            # Testes de utilitários
├── integration/          # Testes de integração
│   ├── repositories/     # Testes de repositórios
│   ├── external/         # Testes de APIs externas
│   └── database/         # Testes de DB
├── e2e/                  # Testes end-to-end
│   ├── api/              # Testes de API completa
│   └── web/              # Testes de interface web
├── performance/          # Testes de performance
├── security/             # Testes de segurança
└── fixtures/             # Dados de teste
    ├── users.json
    ├── {{entities}}.json
    └── config.json
```

### Naming Conventions
```{{LANGUAGE_LOWER}}
// Unit tests
func TestClassName_MethodName_Scenario(t *testing.T) {}

// Integration tests
func TestIntegration_FeatureName_Scenario(t *testing.T) {}

// API tests
func TestAPI_EndpointName_HTTPMethod_Scenario(t *testing.T) {}

// Performance tests
func BenchmarkFeatureName_Scenario(b *testing.B) {}
```

---

## 🚀 CI/CD Integration

### Test Pipeline
```yaml
# .github/workflows/tests.yml
name: Test Suite

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup {{LANGUAGE}}
        uses: {{SETUP_ACTION}}
        with:
          {{language}}-version: '{{VERSION}}'

      - name: Run Unit Tests
        run: {{UNIT_TEST_COMMAND}}

      - name: Upload Coverage
        uses: codecov/codecov-action@v3

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test123
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - name: Run Integration Tests
        run: {{INTEGRATION_TEST_COMMAND}}

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Start Application
        run: {{START_APP_COMMAND}} &
      - name: Run E2E Tests
        run: {{E2E_TEST_COMMAND}}

  performance-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Performance Tests
        run: {{PERFORMANCE_TEST_COMMAND}}

  security-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Security Scan
        run: {{SECURITY_SCAN_COMMAND}}
```

---

## 📈 Test Metrics & Reporting

### Coverage Reports
```bash
# Gerar relatório de cobertura
{{COVERAGE_COMMAND}}

# Relatório HTML
{{COVERAGE_HTML_COMMAND}}

# Verificar threshold de cobertura
{{COVERAGE_CHECK_COMMAND}} --threshold=95
```

### Performance Benchmarks
```bash
# Run benchmarks
{{BENCHMARK_COMMAND}}

# Compare com baseline
{{BENCHMARK_COMPARE_COMMAND}} --baseline=main

# Performance regression check
{{PERFORMANCE_CHECK_COMMAND}} --threshold=10%
```

### Test Results Dashboard
- **Total Tests**: {{TOTAL_TESTS}}
- **Success Rate**: 99.2%
- **Average Duration**: 45s
- **Coverage**: 96.8%
- **Performance**: All benchmarks within 5% baseline

---

## ✅ Quality Gates

### Pull Request Requirements
- [ ] **100%** dos testes passando
- [ ] **95%+** cobertura de código
- [ ] **0** vulnerabilidades críticas
- [ ] **Performance** within 10% baseline
- [ ] **E2E** critical paths passing

### Production Deployment Gates
- [ ] **All test suites** passing
- [ ] **Security scan** clean
- [ ] **Performance tests** under SLA
- [ ] **Integration tests** with prod-like environment
- [ ] **Manual QA** approval

---

## 🔄 Test Maintenance

### Regular Test Review
- **Weekly**: Flaky test analysis
- **Monthly**: Performance baseline update
- **Quarterly**: Test strategy review

### Test Data Cleanup
```{{LANGUAGE_LOWER}}
// Cleanup após cada teste
func cleanup(t *testing.T) {
    // Remove test data
    cleanupTestData()

    // Reset mocks
    resetMocks()

    // Clear caches
    clearCaches()
}
```

### Continuous Improvement
- **Test execution time** monitoring
- **Flaky test** detection and fixing
- **Coverage gaps** identification
- **New test types** evaluation