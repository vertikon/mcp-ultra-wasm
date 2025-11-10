# Test Mocks

Mocks locais para testes deste projeto.

## Uso com testify

```go
package mocks

import "github.com/stretchr/testify/mock"

type ExampleService struct {
    mock.Mock
}

func (m *ExampleService) DoSomething(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}
```

## Uso com gomock

```bash
go install github.com/golang/mock/mockgen@latest
mockgen -source=internal/services/example.go -destination=test/mocks/example_mock.go -package=mocks
```

## Gerando mocks automaticamente

Adicione ao arquivo de interface:

```go
//go:generate mockgen -destination=../../test/mocks/example_mock.go -package=mocks . ExampleService
```

Depois rode:

```bash
go generate ./...
```
