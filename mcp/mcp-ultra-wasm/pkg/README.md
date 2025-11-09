# pkg - Shared Libraries

Biblioteca compartilhável entre múltiplos serviços.

## Propósito

O diretório `pkg` contém código que pode ser importado por aplicações externas.
Use com moderação - se o código não precisa ser compartilhado, mantenha em `internal`.

## Estrutura Sugerida

```
pkg/
├── logger/         # Logger compartilhado
├── version/        # Informações de versão
├── httputil/       # Utilitários HTTP
└── validation/     # Validadores comuns
```

## Exemplos Existentes

Verifique os subdiretórios para implementações específicas.

## Boas Práticas

1. Mantenha APIs estáveis
2. Documente mudanças breaking
3. Evite dependências de `internal`
4. Teste extensivamente (código público)
