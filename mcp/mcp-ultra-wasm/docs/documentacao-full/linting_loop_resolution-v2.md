Eu iria de Opção A (facades) — mas de forma enxuta e faseada, pra te tirar do loop rápido e ainda evitar dívida técnica. Abaixo vai um roteiro “cirúrgico” que resolve os avisos atuais, estabiliza o depguard e te deixa com a base preparada.

Plano recomendado (Opção A, faseado)
Fase 0 — Configuração do depguard (5 min)

Garanta as exceções por path só nos pacotes-facade (pra eles poderem importar o vendor real).

# .golangci.yml  (trecho)
issues:
  exclude-rules:
    - path: pkg/types/
      linters: [depguard]
    - path: pkg/httpx/
      linters: [depguard]
    - path: pkg/redisx/
      linters: [depguard]
    - path: pkg/metrics/
      linters: [depguard]
    - path: pkg/observability/
      linters: [depguard]
    - path: pkg/natsx/
      linters: [depguard]

linters-settings:
  depguard:
    rules:
      main:
        deny:
          - pkg: "github.com/go-chi/chi/v5"                    # use httpx
            desc: "Use pkg/httpx facade instead of direct chi import"
          - pkg: "github.com/go-chi/chi/v5/middleware"         # use httpx
            desc: "Use pkg/httpx facade instead of direct chi import"
          - pkg: "github.com/go-chi/cors"                      # use httpx.CORS
            desc: "Use pkg/httpx.CORS facade"
          - pkg: "github.com/redis/go-redis/v9"                # use redisx
            desc: "Use pkg/redisx facade"
          - pkg: "github.com/prometheus/client_golang/prometheus"          # use metrics
            desc: "Use pkg/metrics facade"
          - pkg: "github.com/prometheus/client_golang/prometheus/promauto" # use metrics
            desc: "Use pkg/metrics facade"
          - pkg: "go.opentelemetry.io/otel"                    # use observability facade
            desc: "Use pkg/observability facade"
          - pkg: "go.opentelemetry.io/otel/*"                  # use observability facade
            desc: "Use pkg/observability facade"
          - pkg: "github.com/nats-io/nats.go"                  # use natsx
            desc: "Use pkg/natsx facade"


Com isso, os facades podem importar as libs-externas; o resto do código é forçado a passar pelos facades.

Fase 1 — Facades mínimos (1–2h)
1) pkg/httpx

Objetivo: substituir chi, cors e chi/middleware fora do facade.

API mínima:

package httpx

import (
  "net/http"
  "github.com/go-chi/chi/v5"
  chimw "github.com/go-chi/chi/v5/middleware"
  "github.com/go-chi/cors"
)

type Router interface {
  Use(mwf ...func(http.Handler) http.Handler)
  Method(method, pattern string, h http.Handler)
  Get(pattern string, h http.HandlerFunc)
  Post(pattern string, h http.HandlerFunc)
  Put(pattern string, h http.HandlerFunc)
  Delete(pattern string, h http.HandlerFunc)
  Mount(pattern string, h http.Handler)
  ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type router struct{ *chi.Mux }

func NewRouter() Router {
  r := chi.NewRouter()
  // expose middlewares commonly used via helpers if quiser
  return &router{r}
}

// CORS helper
func CORS(opts cors.Options) func(http.Handler) http.Handler {
  return cors.New(opts).Handler
}

// Middlewares úteis
func RequestID(next http.Handler) http.Handler { return chimw.RequestID(next) }
func RealIP(next http.Handler) http.Handler    { return chimw.RealIP(next) }
func Recoverer(next http.Handler) http.Handler { return chimw.Recoverer(next) }


No código app: trocar imports de chi, cors etc. por pkg/httpx.

2) pkg/redisx

Crie a API coerente (sem .Result()).

package redisx

import (
  "context"
  "errors"
  "github.com/redis/go-redis/v9"
)

var ErrKeyNotFound = errors.New("redis: key not found")

type Options = redis.Options

type Client struct{ *redis.Client }

func NewClientFromOptions(o *Options) *Client { return &Client{redis.NewClient(o)} }

func (c *Client) Ping(ctx context.Context) error {
  return c.Client.Ping(ctx).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
  val, err := c.Client.Get(ctx, key).Result()
  if err == redis.Nil {
    return "", ErrKeyNotFound
  }
  return val, err
}

func (c *Client) Set(ctx context.Context, key string, val interface{}, ttlSeconds int) error {
  return c.Client.Set(ctx, key, val, time.Duration(ttlSeconds)*time.Second).Err()
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
  return c.Client.Incr(ctx, key).Result()
}

func (c *Client) Exists(ctx context.Context, key string) (int64, error) {
  return c.Client.Exists(ctx, key).Result()
}


No repositório de cache: remover cadeias .Result(), importar redisx e tratar redisx.ErrKeyNotFound.

3) pkg/metrics

Encapsule só o que você usa (Counter, Histogram, Gauge).

package metrics

import (
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promauto"
)

type Counter interface{ Inc(); Add(float64) }
type counter struct{ prometheus.Counter }

func NewCounter(name, help string, labels ...string) Counter {
  c := promauto.NewCounter(prometheus.CounterOpts{Name: name, Help: help})
  return &counter{c}
}


Expanda conforme necessidade (Histogram, Gauge, com labels, etc.), mas comece mínimo.

4) pkg/observability

Você já tem o otelshim. Complete o que faltar para cobrir usos em internal/observability/* e internal/middleware.

Dica: crie wrappers para TracerProvider noop (use trace/noop), atributos, baggage e códigos.

5) pkg/natsx (se necessário)

Se internal/events for legado, mantenha a exceção de path; caso contrário, faça o shim mínimo (conectar, publicar, subscribe) e migre aos poucos.

Fase 2 — Correções de lint atuais (1–2h)

Checklist baseado no seu último make lint:

A) unused-parameter (revive)

Marque parâmetros não usados como _:

internal/cache/distributed.go:getReadThrough(ctx, ...) → (_ context.Context, ...)

Handlers de teste (auth_test.go, etc.): func(w http.ResponseWriter, _ *http.Request) { ... }

TelemetryService.Start(_ context.Context) e collectSystemMetrics(_ context.Context, ...)

metrics/storage.go: extractLabels(_ string, groupBy []string)

B) SA1029 (context.WithValue string key)

Defina uma chave tipada e troque usos:

// internal/middleware/context_keys.go
package middleware
type ctxKey string
const (
  ctxUserIDKey   ctxKey = "user_id"
  ctxUsernameKey ctxKey = "username"
  ctxRolesKey    ctxKey = "user_roles"
)
// uso:
ctx = context.WithValue(ctx, ctxUserIDKey, claims.UserID)


Elimina os 3 SA1029 em internal/middleware/auth.go.

C) goconst em TLS tests

Se já existem const TLS12="1.2"; TLS13="1.3", use-as nos testes:

manager.config.MinVersion = TLS12
manager.config.MaxVersion = TLS13

D) SA1019 (Jaeger / NewNoopTracerProvider)

Troque:

trace.NewNoopTracerProvider() → noop.NewTracerProvider()

Exporters Jaeger: ou migre pra OTLP ou coloque por trás do seu pkg/observability e use build tag para desligar no CI.

Rápido: comente o branch Jaeger e deixe OTLP/noop como padrão no CI.

E) io/ioutil (deprecated)

Troque ioutil.ReadFile → os.ReadFile, ioutil.WriteFile → os.WriteFile, ioutil.NopCloser → io.NopCloser, etc.

F) campo não usado

enhanced_telemetry.go: spanMutex → remova se realmente não é usado.

G) basic_test.go (SA4000)

if true != true → remova ou troque por um teste real.

Fase 3 — Refactors pontuais (≤1h)

internal/middleware/auth.go

Trocar imports diretos de chi, otel/*, zap → httpx, observability, e seu logger-facade (ou a interface mínima que você já criou).

Ajustar contexto (item B acima).

handlers http

router.go, health.go, feature_flag_handlers.go → usar httpx e observability.

AI telemetry

Trocar prometheus/promauto → pkg/metrics.

Validação sugerida
make fmt tidy
golangci-lint run --disable-all -E depguard    # checa só depguard
golangci-lint run                              # tudo
make build
# se -race falhar no Windows sem cgo:
set CGO_ENABLED=0 && go test ./... -count=1

Se precisar de “plano B” (temporário, 5 min)

Desativar depguard só para permitir o fluxo enquanto os facades são implementados:

linters:
  disable:
    - depguard


Eu não recomendo ficar assim: tua base volta ao risco de dependências diretas espalhadas. Mas serve para destravar até terminar a Fase 1.

Minha recomendação

Opção A (faseado). Você já tem parte da fundação (observability). Completar httpx, redisx e metrics em versão mínima e fazer as trocas onde o linter acusa vai te tirar do loop de forma sustentável.

Se pintar pressão de tempo, aplica o Plano B por algumas horas, mas já com a Fase 1 em andamento para reativar o depguard ainda hoje.