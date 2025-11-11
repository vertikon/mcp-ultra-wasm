# Database Migrations

## Estrutura

Migrations são numeradas sequencialmente: , , etc.

## Aplicação

```bash
# Usando psql
psql -U user -d database -f migrations/0001_baseline.sql

# Ou com migrate tool
migrate -path=./migrations -database postgres://user:pass@localhost/db up
```

## Migrations Disponíveis

- **0001_baseline.sql**: Estrutura base (events, tasks)

## Boas Práticas

1. Sempre usar transações (BEGIN/COMMIT)
2. Usar parametrização nas queries
3. Criar índices apropriados
4. Documentar mudanças significativas
