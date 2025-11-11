package telemetry

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// We register once to avoid duplicate metric panics in tests.
var (
	once sync.Once

	infRequests     *prometheus.CounterVec
	infLatency      *prometheus.HistogramVec
	tokensIn        *prometheus.CounterVec
	tokensOut       *prometheus.CounterVec
	costBRL         *prometheus.CounterVec
	policyBlocks    *prometheus.CounterVec
	routerDecisions *prometheus.CounterVec
	budgetBreaches  *prometheus.CounterVec
)

type Labels struct {
	TenantID string
	MCPID    string
	SDKName  string
	Provider string
	Model    string
	UseCase  string
	Reason   string
	Rule     string
	Severity string
	Scope    string // global|tenant|mcp
}

func Init(reg prometheus.Registerer) {
	once.Do(func() {
		if reg == nil {
			reg = prometheus.DefaultRegisterer
		}
		infRequests = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "ai_inference_requests_total",
			Help: "Total de requisições de inferência IA",
		}, []string{"tenant_id", "mcp_id", "sdk_name", "provider", "model", "use_case"})
		infLatency = promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
			Name:    "ai_inference_latency_ms",
			Help:    "Latência de inferência em milissegundos",
			Buckets: []float64{50, 100, 200, 400, 800, 1600, 3200, 6400, 12800},
		}, []string{"tenant_id", "mcp_id", "sdk_name", "provider", "model", "use_case"})
		tokensIn = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "ai_tokens_in_total",
			Help: "Total de tokens de entrada",
		}, []string{"tenant_id", "mcp_id", "sdk_name"})
		tokensOut = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "ai_tokens_out_total",
			Help: "Total de tokens de saída",
		}, []string{"tenant_id", "mcp_id", "sdk_name"})
		costBRL = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "ai_cost_brl_total",
			Help: "Custo acumulado em BRL",
		}, []string{"tenant_id", "mcp_id", "sdk_name"})
		policyBlocks = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "ai_policy_blocks_total",
			Help: "Total de bloqueios por política/guardrail",
		}, []string{"tenant_id", "mcp_id", "sdk_name", "rule", "severity"})
		routerDecisions = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "ai_router_decisions_total",
			Help: "Total de decisões de roteamento (provider/model)",
		}, []string{"tenant_id", "mcp_id", "sdk_name", "provider", "model", "reason"})
		budgetBreaches = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "ai_budget_breaches_total",
			Help: "Ocorrências de violação de orçamento (global/tenant/mcp)",
		}, []string{"scope"})
	})
}

type InferenceMeta struct {
	Labels    Labels
	TokensIn  int
	TokensOut int
	CostBRL   float64
	Start     time.Time
	End       time.Time
}

func ObserveStart() time.Time { return time.Now() }

func ObserveInference(meta InferenceMeta) {
	if infRequests == nil {
		// Not initialized (AI disabled)
		return
	}
	l := meta.Labels
	infRequests.WithLabelValues(l.TenantID, l.MCPID, l.SDKName, l.Provider, l.Model, l.UseCase).Inc()

	lat := meta.End.Sub(meta.Start).Seconds() * 1000.0
	infLatency.WithLabelValues(l.TenantID, l.MCPID, l.SDKName, l.Provider, l.Model, l.UseCase).Observe(lat)

	if meta.TokensIn > 0 {
		tokensIn.WithLabelValues(l.TenantID, l.MCPID, l.SDKName).Add(float64(meta.TokensIn))
	}
	if meta.TokensOut > 0 {
		tokensOut.WithLabelValues(l.TenantID, l.MCPID, l.SDKName).Add(float64(meta.TokensOut))
	}
	if meta.CostBRL > 0 {
		costBRL.WithLabelValues(l.TenantID, l.MCPID, l.SDKName).Add(meta.CostBRL)
	}
}

func IncPolicyBlock(l Labels) {
	if policyBlocks == nil {
		return
	}
	policyBlocks.WithLabelValues(l.TenantID, l.MCPID, l.SDKName, l.Rule, l.Severity).Inc()
}

func IncRouterDecision(l Labels) {
	if routerDecisions == nil {
		return
	}
	routerDecisions.WithLabelValues(l.TenantID, l.MCPID, l.SDKName, l.Provider, l.Model, l.Reason).Inc()
}

func IncBudgetBreach(scope string) {
	if budgetBreaches == nil {
		return
	}
	budgetBreaches.WithLabelValues(scope).Inc()
}
