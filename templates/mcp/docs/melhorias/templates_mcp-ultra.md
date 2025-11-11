O Windows PowerShell
Copyright (C) Microsoft Corporation. Todos os direitos reservados.

Instale o PowerShell mais recente para obter novos recursos e aprimoramentos! https://aka.ms/PSWindows

✅ GPT5 Integration carregado
🚀 Carregando profile Vertikon...
  ✓ Go bin adicionado ao PATH
✅ Profile Vertikon carregado!
   Root: E:\vertikon
   Digite 'aliases' para ver comandos disponíveis
   Digite 'Check-GoTools' para verificar ferramentas

PS E:\vertikon\business\SaaS\templates\mcp-ultra-wasm> go mod tidy
PS E:\vertikon\business\SaaS\templates\mcp-ultra-wasm> go build ./...
PS E:\vertikon\business\SaaS\templates\mcp-ultra-wasm> go test ./... -count=1
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/cache [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/cache.test]
internal\cache\distributed_test.go:22:9: cannot use l (variable of type *logger.Logger) as logger.Logger value in return statement
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/compliance [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/compliance.test]
internal\compliance\framework_test.go:111:27: framework.ScanForPII undefined (type *ComplianceFramework has no field or method ScanForPII)
internal\compliance\framework_test.go:133:19: framework.RecordConsent undefined (type *ComplianceFramework has no field or method RecordConsent)
internal\compliance\framework_test.go:137:31: framework.HasConsent undefined (type *ComplianceFramework has no field or method HasConsent)
internal\compliance\framework_test.go:142:30: framework.HasConsent undefined (type *ComplianceFramework has no field or method HasConsent)
internal\compliance\framework_test.go:147:18: framework.WithdrawConsent undefined (type *ComplianceFramework has no field or method WithdrawConsent)
internal\compliance\framework_test.go:151:30: framework.HasConsent undefined (type *ComplianceFramework has no field or method HasConsent)
internal\compliance\framework_test.go:156:30: framework.HasConsent undefined (type *ComplianceFramework has no field or method HasConsent)
internal\compliance\framework_test.go:169:19: framework.RecordDataCreation undefined (type *ComplianceFramework has no field or method RecordDataCreation)
internal\compliance\framework_test.go:176:27: framework.GetRetentionPolicy undefined (type *ComplianceFramework has no field or method GetRetentionPolicy)
internal\compliance\framework_test.go:182:33: framework.ShouldDeleteData undefined (type *ComplianceFramework has no field or method ShouldDeleteData)
internal\compliance\framework_test.go:182:33: too many errors
ok      github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm   0.556s
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/api/grpc/gen/compliance/v1        [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/api/grpc/gen/system/v1    [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/api/grpc/gen/task/v1      [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/automation        [no test files]
ok      github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/events        0.373s
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/router        [no test files]
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features.test]
internal\features\manager_test.go:6:2: "time" imported and not used
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/handlers/http [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/handlers/http.test]
internal\handlers\http\router_test.go:23:76: undefined: services.HealthStatus
internal\handlers\http\router_test.go:25:42: undefined: services.HealthStatus
internal\handlers\http\router_test.go:38:75: undefined: services.HealthChecker
internal\handlers\http\router_test.go:47:70: undefined: domain.CreateTaskRequest
internal\handlers\http\router_test.go:60:85: undefined: domain.UpdateTaskRequest
internal\handlers\http\router_test.go:70:73: undefined: domain.TaskFilters
internal\handlers\http\router_test.go:70:95: undefined: domain.TaskList
internal\handlers\http\router_test.go:72:30: undefined: domain.TaskList
internal\handlers\http\health_test.go:272:27: undefined: fmt
internal\handlers\http\health_test.go:273:14: undefined: fmt
internal\handlers\http\router_test.go:72:30: too many errors
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/middleware [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/middleware.test]
internal\middleware\auth_test.go:95:30: undefined: testhelpers.GetTestAPIKeys
internal\middleware\auth_test.go:284:9: undefined: fmt
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability.test]
internal\observability\telemetry_test.go:60:20: service.GetTracer undefined (type *TelemetryService has no field or method GetTracer)
internal\observability\telemetry_test.go:63:19: service.GetMeter undefined (type *TelemetryService has no field or method GetMeter)
internal\observability\telemetry_test.go:83:20: service.GetTracer undefined (type *TelemetryService has no field or method GetTracer)
internal\observability\telemetry_test.go:96:3: undefined: attribute
internal\observability\telemetry_test.go:97:3: undefined: attribute
internal\observability\telemetry_test.go:102:26: undefined: attribute
internal\observability\telemetry_test.go:118:19: service.GetMeter undefined (type *TelemetryService has no field or method GetMeter)
internal\observability\telemetry_test.go:123:3: undefined: metric
internal\observability\telemetry_test.go:124:3: undefined: metric
internal\observability\telemetry_test.go:129:22: undefined: metric
internal\observability\telemetry_test.go:129:22: too many errors
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/services [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/services.test]
internal\services\task_service_test.go:104:70: undefined: domain.UserFilter
internal\services\task_service_test.go:171:28: cannot use taskRepo (variable of type *mockTaskRepository) as domain.TaskRepository value in argument to NewTaskService: *mockTaskRepository does not implement domain.TaskRepository (wrong type for method List)
                have List(context.Context, domain.TaskFilter) ([]*domain.Task, error)
                want List(context.Context, domain.TaskFilter) ([]*domain.Task, int, error)
internal\services\task_service_test.go:171:48: cannot use eventRepo (variable of type *mockEventRepository) as domain.EventRepository value in argument to NewTaskService: *mockEventRepository does not implement domain.EventRepository (missing method GetByType)
internal\services\task_service_test.go:171:59: cannot use cacheRepo (variable of type *mockCacheRepository) as domain.CacheRepository value in argument to NewTaskService: *mockCacheRepository does not implement domain.CacheRepository (missing method Exists)
internal\services\task_service_test.go:199:31: declared and not used: eventRepo
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/security [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/security.test]
internal\security\enhanced_auth_test.go:22:6: MockOPAService redeclared in this block
        internal\security\auth_test.go:23:6: other declaration of MockOPAService
internal\security\enhanced_auth_test.go:26:26: method MockOPAService.IsAuthorized already declared at internal\security\auth_test.go:27:26
internal\security\enhanced_auth_test.go:36:6: TestNewAuthService redeclared in this block
        internal\security\auth_test.go:42:6: other declaration of TestNewAuthService
internal\security\enhanced_auth_test.go:326:6: TestGetUserFromContext redeclared in this block
        internal\security\auth_test.go:414:6: other declaration of TestGetUserFromContext
internal\security\enhanced_auth_test.go:391:6: TestRequireScope redeclared in this block
        internal\security\auth_test.go:285:6: other declaration of TestRequireScope
internal\security\enhanced_auth_test.go:459:6: TestRequireRole redeclared in this block
        internal\security\auth_test.go:345:6: other declaration of TestRequireRole
internal\security\auth_test.go:52:48: cannot use opa (variable of type *MockOPAService) as *OPAService value in argument to NewAuthService
internal\security\auth_test.go:70:48: cannot use opa (variable of type *MockOPAService) as *OPAService value in argument to NewAuthService
internal\security\auth_test.go:143:48: cannot use opa (variable of type *MockOPAService) as *OPAService value in argument to NewAuthService
internal\security\auth_test.go:166:48: cannot use opa (variable of type *MockOPAService) as *OPAService value in argument to NewAuthService
internal\security\auth_test.go:166:48: too many errors
ok      github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/telemetry     1.585s
ok      github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/wiring        1.128s
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/cache [build failed]
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/compliance [build failed]
--- FAIL: TestNewTLSManager (0.09s)
    logger.go:146: 2025-10-16T07:54:25.459-0300 INFO    TLS is disabled
    --- FAIL: TestNewTLSManager/should_create_manager_with_valid_TLS_config (0.03s)
        tls_test.go:120:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/config/tls_test.go:120
                Error:          Received unexpected error:
                                failed to load TLS configuration: failed to load certificate pair: tls: failed to find any PEM data in key input
                Test:           TestNewTLSManager/should_create_manager_with_valid_TLS_config
--- FAIL: TestTLSManager_GetTLSConfig (0.02s)
    --- FAIL: TestTLSManager_GetTLSConfig/should_return_copy_of_TLS_config (0.02s)
        tls_test.go:306:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/config/tls_test.go:306
                Error:          Received unexpected error:
                                failed to load TLS configuration: failed to load certificate pair: tls: failed to find any PEM data in key input
                Test:           TestTLSManager_GetTLSConfig/should_return_copy_of_TLS_config
--- FAIL: TestTLSManager_Stop (0.02s)
    --- FAIL: TestTLSManager_Stop/should_stop_certificate_watcher (0.02s)
        tls_test.go:334:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/config/tls_test.go:334
                Error:          Received unexpected error:
                                failed to load TLS configuration: failed to load certificate pair: tls: failed to find any PEM data in key input
                Test:           TestTLSManager_Stop/should_stop_certificate_watcher
FAIL
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config   1.428s
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config/secrets   [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/constants        [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/dashboard        [no test files]
--- FAIL: TestTaskComplete (0.00s)
    models_test.go:40:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/domain/models_test.go:40
                Error:          Should be true
                Test:           TestTaskComplete
    models_test.go:41:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/domain/models_test.go:41
                Error:          Should be true
                Test:           TestTaskComplete
    models_test.go:42:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/domain/models_test.go:42
                Error:          Should be true
                Test:           TestTaskComplete
--- FAIL: TestTaskCancel (0.00s)
    models_test.go:53:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/domain/models_test.go:53
                Error:          Should be true
                Test:           TestTaskCancel
    models_test.go:54:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/domain/models_test.go:54
                Error:          Should be true
                Test:           TestTaskCancel
--- FAIL: TestTaskUpdateStatus (0.00s)
    models_test.go:65:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/domain/models_test.go:65
                Error:          Should be true
                Test:           TestTaskUpdateStatus
    models_test.go:66:
                Error Trace:    E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/domain/models_test.go:66
                Error:          Should be true
                Test:           TestTaskUpdateStatus
FAIL
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain   0.396s
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/dr       [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/events   [no test files]
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features [build failed]
ok      github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/handlers 0.684s
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/handlers/http [build failed]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/http     [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/lifecycle        [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/metrics  [no test files]
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/middleware [build failed]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/nats     [no test files]
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability [build failed]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ratelimit        [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/repository/postgres      [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/repository/redis [no test files]
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/security [build failed]
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/services [build failed]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/slo      [no test files]
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/component [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/component.test]
test\component\task_service_test.go:39:3: cannot use suite.taskRepo (variable of type *mocks.MockTaskRepository) as domain.TaskRepository value in argument to services.NewTaskService: *mocks.MockTaskRepository does not implement domain.TaskRepository (wrong type for method Delete)
                have Delete(context.Context, string) error
                want Delete(context.Context, uuid.UUID) error
test\component\task_service_test.go:40:3: cannot use suite.validator (variable of type *mocks.MockValidator) as domain.UserRepository value in argument to services.NewTaskService: *mocks.MockValidator does not implement domain.UserRepository (missing method Create)
test\component\task_service_test.go:42:3: cannot use suite.cacheRepo (variable of type *mocks.MockCacheRepository) as domain.CacheRepository value in argument to services.NewTaskService: *mocks.MockCacheRepository does not implement domain.CacheRepository (wrong type for method Get)
                have Get(context.Context, string) (interface{}, error)
                want Get(context.Context, string) (string, error)
test\component\task_service_test.go:44:3: cannot use suite.eventBus (variable of type *mocks.MockEventBus) as services.EventBus value in argument to services.NewTaskService: *mocks.MockEventBus does not implement services.EventBus (wrong type for method Publish)
                have Publish(context.Context, string, []byte) error
                want Publish(context.Context, *domain.Event) error
test\component\task_service_test.go:65:3: unknown field Metadata in struct literal of type services.CreateTaskRequest
test\component\task_service_test.go:78:20: req.Metadata undefined (type *services.CreateTaskRequest has no field or method Metadata)
test\component\task_service_test.go:97:55: too many arguments in call to suite.service.CreateTask
        have (context.Context, uuid.UUID, *services.CreateTaskRequest)
        want (context.Context, services.CreateTaskRequest)
test\component\task_service_test.go:118:29: undefined: services.ValidationError
test\component\task_service_test.go:127:55: too many arguments in call to suite.service.CreateTask
        have (context.Context, uuid.UUID, *services.CreateTaskRequest)
        want (context.Context, services.CreateTaskRequest)
test\component\task_service_test.go:151:52: too many arguments in call to suite.service.GetTask
        have (context.Context, uuid.UUID, uuid.UUID)
        want (context.Context, uuid.UUID)
test\component\task_service_test.go:151:52: too many errors
# github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/property [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/property.test]
test\property\task_properties_test.go:11:2: "github.com/stretchr/testify/assert" imported and not used
test\property\task_properties_test.go:232:4: declared and not used: originalTitle
panic: a previously registered descriptor with the same fully-qualified name as Desc{fqName: "http_request_duration_seconds", help: "Duration of HTTP requests in seconds", constLabels: {}, variableLabels: {method,path,status}} has different label names or a different help string

goroutine 1 [running]:
github.com/prometheus/client_golang/prometheus.(*Registry).MustRegister(0x7ff7d16cf920, {0xc0000a8000?, 0x0?, 0x0?})
        E:/go-workspace/pkg/mod/github.com/prometheus/client_golang@v1.23.0/prometheus/registry.go:406 +0x65
github.com/prometheus/client_golang/prometheus/promauto.Factory.NewHistogramVec({{0x7ff7d1129820?, 0x7ff7d16cf920?}}, {{0x0, 0x0}, {0x0, 0x0}, {0x7ff7d101592e, 0x1d}, {0x7ff7d101d3b7, 0x24}, ...}, ...)
        E:/go-workspace/pkg/mod/github.com/prometheus/client_golang@v1.23.0/prometheus/promauto/auto.go:362 +0x1cb
github.com/prometheus/client_golang/prometheus/promauto.NewHistogramVec(...)
        E:/go-workspace/pkg/mod/github.com/prometheus/client_golang@v1.23.0/prometheus/promauto/auto.go:235
github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/telemetry.init()
        E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/internal/telemetry/telemetry.go:33 +0x392
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/telemetry        0.426s
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/testhelpers      [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/tracing  [no test files]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/scripts   [no test files]
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/component [build failed]
?       github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/mocks        [no test files]
FAIL    github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/property [build failed]
ok      github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/tests/smoke       0.331s
FAIL
PS E:\vertikon\business\SaaS\templates\mcp-ultra-wasm> golangci-lint run
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/cache [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/cache.test]\ninternal\\cache\\distributed_test.go:22:9: cannot use l (variable of type *logger.Logger) as logger.Logger value in return statement"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/compliance [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/compliance.test]\ninternal\\compliance\\framework_test.go:111:27: framework.ScanForPII undefined (type *ComplianceFramework has no field or method ScanForPII)\ninternal\\compliance\\framework_test.go:133:19: framework.RecordConsent undefined (type *ComplianceFramework has no field or method RecordConsent)\ninternal\\compliance\\framework_test.go:137:31: framework.HasConsent undefined (type *ComplianceFramework has no field or method HasConsent)\ninternal\\compliance\\framework_test.go:142:30: framework.HasConsent undefined (type *ComplianceFramework has no field or method HasConsent)\ninternal\\compliance\\framework_test.go:147:18: framework.WithdrawConsent undefined (type *ComplianceFramework has no field or method WithdrawConsent)\ninternal\\compliance\\framework_test.go:151:30: framework.HasConsent undefined (type *ComplianceFramework has no field or method HasConsent)\ninternal\\compliance\\framework_test.go:156:30: framework.HasConsent undefined (type *ComplianceFramework has no field or method HasConsent)\ninternal\\compliance\\framework_test.go:169:19: framework.RecordDataCreation undefined (type *ComplianceFramework has no field or method RecordDataCreation)\ninternal\\compliance\\framework_test.go:176:27: framework.GetRetentionPolicy undefined (type *ComplianceFramework has no field or method GetRetentionPolicy)\ninternal\\compliance\\framework_test.go:182:33: framework.ShouldDeleteData undefined (type *ComplianceFramework has no field or method ShouldDeleteData)\ninternal\\compliance\\framework_test.go:182:33: too many errors"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features.test]\ninternal\\features\\manager_test.go:6:2: \"time\" imported and not used"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/handlers/http [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/handlers/http.test]\ninternal\\handlers\\http\\router_test.go:23:76: undefined: services.HealthStatus\ninternal\\handlers\\http\\router_test.go:25:42: undefined: services.HealthStatus\ninternal\\handlers\\http\\router_test.go:38:75: undefined: services.HealthChecker\ninternal\\handlers\\http\\router_test.go:47:70: undefined: domain.CreateTaskRequest\ninternal\\handlers\\http\\router_test.go:60:85: undefined: domain.UpdateTaskRequest\ninternal\\handlers\\http\\router_test.go:70:73: undefined: domain.TaskFilters\ninternal\\handlers\\http\\router_test.go:70:95: undefined: domain.TaskList\ninternal\\handlers\\http\\router_test.go:72:30: undefined: domain.TaskList\ninternal\\handlers\\http\\health_test.go:272:27: undefined: fmt\ninternal\\handlers\\http\\health_test.go:273:14: undefined: fmt\ninternal\\handlers\\http\\router_test.go:72:30: too many errors"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/middleware [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/middleware.test]\ninternal\\middleware\\auth_test.go:95:30: undefined: testhelpers.GetTestAPIKeys\ninternal\\middleware\\auth_test.go:284:9: undefined: fmt"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability.test]\ninternal\\observability\\telemetry_test.go:60:20: service.GetTracer undefined (type *TelemetryService has no field or method GetTracer)\ninternal\\observability\\telemetry_test.go:63:19: service.GetMeter undefined (type *TelemetryService has no field or method GetMeter)\ninternal\\observability\\telemetry_test.go:83:20: service.GetTracer undefined (type *TelemetryService has no field or method GetTracer)\ninternal\\observability\\telemetry_test.go:96:3: undefined: attribute\ninternal\\observability\\telemetry_test.go:97:3: undefined: attribute\ninternal\\observability\\telemetry_test.go:102:26: undefined: attribute\ninternal\\observability\\telemetry_test.go:118:19: service.GetMeter undefined (type *TelemetryService has no field or method GetMeter)\ninternal\\observability\\telemetry_test.go:123:3: undefined: metric\ninternal\\observability\\telemetry_test.go:124:3: undefined: metric\ninternal\\observability\\telemetry_test.go:129:22: undefined: metric\ninternal\\observability\\telemetry_test.go:129:22: too many errors"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/security [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/security.test]\ninternal\\security\\enhanced_auth_test.go:22:6: MockOPAService redeclared in this block\n\tinternal\\security\\auth_test.go:23:6: other declaration of MockOPAService\ninternal\\security\\enhanced_auth_test.go:26:26: method MockOPAService.IsAuthorized already declared at internal\\security\\auth_test.go:27:26\ninternal\\security\\enhanced_auth_test.go:36:6: TestNewAuthService redeclared in this block\n\tinternal\\security\\auth_test.go:42:6: other declaration of TestNewAuthService\ninternal\\security\\enhanced_auth_test.go:326:6: TestGetUserFromContext redeclared in this block\n\tinternal\\security\\auth_test.go:414:6: other declaration of TestGetUserFromContext\ninternal\\security\\enhanced_auth_test.go:391:6: TestRequireScope redeclared in this block\n\tinternal\\security\\auth_test.go:285:6: other declaration of TestRequireScope\ninternal\\security\\enhanced_auth_test.go:459:6: TestRequireRole redeclared in this block\n\tinternal\\security\\auth_test.go:345:6: other declaration of TestRequireRole\ninternal\\security\\auth_test.go:52:48: cannot use opa (variable of type *MockOPAService) as *OPAService value in argument to NewAuthService\ninternal\\security\\auth_test.go:70:48: cannot use opa (variable of type *MockOPAService) as *OPAService value in argument to NewAuthService\ninternal\\security\\auth_test.go:143:48: cannot use opa (variable of type *MockOPAService) as *OPAService value in argument to NewAuthService\ninternal\\security\\auth_test.go:166:48: cannot use opa (variable of type *MockOPAService) as *OPAService value in argument to NewAuthService\ninternal\\security\\auth_test.go:166:48: too many errors"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/services [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/services.test]\ninternal\\services\\task_service_test.go:104:70: undefined: domain.UserFilter\ninternal\\services\\task_service_test.go:171:28: cannot use taskRepo (variable of type *mockTaskRepository) as domain.TaskRepository value in argument to NewTaskService: *mockTaskRepository does not implement domain.TaskRepository (wrong type for method List)\n\t\thave List(context.Context, domain.TaskFilter) ([]*domain.Task, error)\n\t\twant List(context.Context, domain.TaskFilter) ([]*domain.Task, int, error)\ninternal\\services\\task_service_test.go:171:48: cannot use eventRepo (variable of type *mockEventRepository) as domain.EventRepository value in argument to NewTaskService: *mockEventRepository does not implement domain.EventRepository (missing method GetByType)\ninternal\\services\\task_service_test.go:171:59: cannot use cacheRepo (variable of type *mockCacheRepository) as domain.CacheRepository value in argument to NewTaskService: *mockCacheRepository does not implement domain.CacheRepository (missing method Exists)\ninternal\\services\\task_service_test.go:199:31: declared and not used: eventRepo"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/compliance_test [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/compliance.test]\ntest\\compliance\\compliance_integration_test.go:369:3: declared and not used: result"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/component [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/component.test]\ntest\\component\\task_service_test.go:39:3: cannot use suite.taskRepo (variable of type *mocks.MockTaskRepository) as domain.TaskRepository value in argument to services.NewTaskService: *mocks.MockTaskRepository does not implement domain.TaskRepository (wrong type for method Delete)\n\t\thave Delete(context.Context, string) error\n\t\twant Delete(context.Context, uuid.UUID) error\ntest\\component\\task_service_test.go:40:3: cannot use suite.validator (variable of type *mocks.MockValidator) as domain.UserRepository value in argument to services.NewTaskService: *mocks.MockValidator does not implement domain.UserRepository (missing method Create)\ntest\\component\\task_service_test.go:42:3: cannot use suite.cacheRepo (variable of type *mocks.MockCacheRepository) as domain.CacheRepository value in argument to services.NewTaskService: *mocks.MockCacheRepository does not implement domain.CacheRepository (wrong type for method Get)\n\t\thave Get(context.Context, string) (interface{}, error)\n\t\twant Get(context.Context, string) (string, error)\ntest\\component\\task_service_test.go:44:3: cannot use suite.eventBus (variable of type *mocks.MockEventBus) as services.EventBus value in argument to services.NewTaskService: *mocks.MockEventBus does not implement services.EventBus (wrong type for method Publish)\n\t\thave Publish(context.Context, string, []byte) error\n\t\twant Publish(context.Context, *domain.Event) error\ntest\\component\\task_service_test.go:65:3: unknown field Metadata in struct literal of type services.CreateTaskRequest\ntest\\component\\task_service_test.go:78:20: req.Metadata undefined (type *services.CreateTaskRequest has no field or method Metadata)\ntest\\component\\task_service_test.go:97:55: too many arguments in call to suite.service.CreateTask\n\thave (context.Context, uuid.UUID, *services.CreateTaskRequest)\n\twant (context.Context, services.CreateTaskRequest)\ntest\\component\\task_service_test.go:118:29: undefined: services.ValidationError\ntest\\component\\task_service_test.go:127:55: too many arguments in call to suite.service.CreateTask\n\thave (context.Context, uuid.UUID, *services.CreateTaskRequest)\n\twant (context.Context, services.CreateTaskRequest)\ntest\\component\\task_service_test.go:151:52: too many arguments in call to suite.service.GetTask\n\thave (context.Context, uuid.UUID, uuid.UUID)\n\twant (context.Context, uuid.UUID)\ntest\\component\\task_service_test.go:151:52: too many errors"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/integration [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/integration.test]\ntest\\integration\\database_integration_test.go:70:19: undefined: testcontainers.NewLogWaitStrategy\ntest\\integration\\database_integration_test.go:120:21: undefined: postgresRepo.RunMigrations\ntest\\integration\\database_integration_test.go:140:23: suite.taskRepo.DB undefined (type *\"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/repository/postgres\".TaskRepository has no field or method DB)\ntest\\integration\\database_integration_test.go:145:28: suite.cacheRepo.Client undefined (type *\"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/repository/redis\".CacheRepository has no field or method Client, but does have unexported field client)\ntest\\integration\\database_integration_test.go:169:22: assignment mismatch: 2 variables but suite.taskRepo.Create returns 1 value\ntest\\integration\\database_integration_test.go:187:22: assignment mismatch: 2 variables but suite.taskRepo.Update returns 1 value\ntest\\integration\\database_integration_test.go:194:24: assignment mismatch: 2 variables but suite.taskRepo.Update returns 1 value\ntest\\integration\\database_integration_test.go:201:3: unknown field UserID in struct literal of type domain.TaskFilter\ntest\\integration\\database_integration_test.go:202:11: cannot use domain.TaskStatusCompleted (constant \"completed\" of string type domain.TaskStatus) as []domain.TaskStatus value in struct literal\ntest\\integration\\database_integration_test.go:207:48: cannot use filter (variable of type *domain.TaskFilter) as domain.TaskFilter value in argument to suite.taskRepo.List\ntest\\integration\\database_integration_test.go:207:48: too many errors"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/observability_test [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/observability.test]\ntest\\observability\\integration_test.go:7:2: \"bytes\" imported and not used\ntest\\observability\\integration_test.go:11:2: \"io\" imported and not used\ntest\\observability\\integration_test.go:103:21: telemetryService.CreateAttribute undefined (type *observability.TelemetryService has no field or method CreateAttribute)\ntest\\observability\\integration_test.go:104:21: telemetryService.CreateAttribute undefined (type *observability.TelemetryService has no field or method CreateAttribute)\ntest\\observability\\integration_test.go:112:21: telemetryService.CreateAttribute undefined (type *observability.TelemetryService has no field or method CreateAttribute)\ntest\\observability\\integration_test.go:130:20: telemetryService.IncrementCounter undefined (type *observability.TelemetryService has no field or method IncrementCounter)"
level=error msg="[linters_context] typechecking error: : # github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/property [github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/property.test]\ntest\\property\\task_properties_test.go:11:2: \"github.com/stretchr/testify/assert\" imported and not used\ntest\\property\\task_properties_test.go:232:4: declared and not used: originalTitle"
internal\handlers\health.go:17:27: Error return value of `(*encoding/json.Encoder).Encode` is not checked (errcheck)
        json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
                                 ^
internal\handlers\health.go:23:27: Error return value of `(*encoding/json.Encoder).Encode` is not checked (errcheck)
        json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
                                 ^
internal\handlers\health.go:29:27: Error return value of `(*encoding/json.Encoder).Encode` is not checked (errcheck)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
                                 ^
internal\handlers\health.go:44:10: Error return value of `w.Write` is not checked (errcheck)
                w.Write([]byte("# Metrics placeholder\n"))
                       ^
internal\http\router.go:19:27: Error return value of `(*encoding/json.Encoder).Encode` is not checked (errcheck)
        json.NewEncoder(w).Encode(resp)
                                 ^
internal\http\router.go:38:27: Error return value of `(*encoding/json.Encoder).Encode` is not checked (errcheck)
        json.NewEncoder(w).Encode(map[string]any{"flag": req.Flag, "value": val})
                                 ^
internal\slo\alerting.go:525:23: Error return value of `resp.Body.Close` is not checked (errcheck)
        defer resp.Body.Close()
                             ^
internal\config\config.go:289:18: Error return value of `file.Close` is not checked (errcheck)
        defer file.Close()
                        ^
main.go:29:19: Error return value of `logger.Sync` is not checked (errcheck)
        defer logger.Sync()
                         ^
internal\repository\postgres\task_repository.go:194:18: Error return value of `rows.Close` is not checked (errcheck)
        defer rows.Close()
                        ^
internal\repository\postgres\task_repository.go:221:18: Error return value of `rows.Close` is not checked (errcheck)
        defer rows.Close()
                        ^
internal\repository\postgres\task_repository.go:248:18: Error return value of `rows.Close` is not checked (errcheck)
        defer rows.Close()
                        ^
internal\repository\postgres\task_repository.go:284:17: Error return value of `json.Unmarshal` is not checked (errcheck)
                json.Unmarshal(tagsJSON, &task.Tags)
                              ^
internal\repository\postgres\task_repository.go:290:17: Error return value of `json.Unmarshal` is not checked (errcheck)
                json.Unmarshal(metadataJSON, &task.Metadata)
                              ^
internal\lifecycle\deployment.go:407:20: Error return value of `da.executeCommand` is not checked (errcheck)
                da.executeCommand(ctx, fmt.Sprintf("kubectl delete deployment mcp-ultra-wasm-canary --namespace=%s", da.config.Namespace), result)
                                 ^
internal\lifecycle\deployment.go:420:19: Error return value of `da.executeCommand` is not checked (errcheck)
        da.executeCommand(ctx, fmt.Sprintf("kubectl delete deployment mcp-ultra-wasm-canary --namespace=%s", da.config.Namespace), result)
                         ^
internal\lifecycle\health.go:476:28: Error return value of `(*encoding/json.Encoder).Encode` is not checked (errcheck)
                json.NewEncoder(w).Encode(report)
                                         ^
internal\lifecycle\health.go:483:11: Error return value of `w.Write` is not checked (errcheck)
                        w.Write([]byte("OK"))
                               ^
internal\lifecycle\health.go:486:11: Error return value of `w.Write` is not checked (errcheck)
                        w.Write([]byte("Not Ready"))
                               ^
internal\lifecycle\health.go:494:11: Error return value of `w.Write` is not checked (errcheck)
                        w.Write([]byte("OK"))
                               ^
internal\lifecycle\health.go:497:11: Error return value of `w.Write` is not checked (errcheck)
                        w.Write([]byte("Unhealthy"))
                               ^
internal\slo\alerting.go:230:1: cognitive complexity 47 of func `(*AlertManager).shouldSilence` is high (> 20) (gocognit)
func (am *AlertManager) shouldSilence(alert AlertEvent) bool {
^
internal\tracing\business.go:772:1: cognitive complexity 23 of func `(*BusinessTransactionTracer).extractCorrelationFields` is high (> 20) (gocognit)
func (btt *BusinessTransactionTracer) extractCorrelationFields(transaction *BusinessTransaction, attributes map[string]interface{}) {
^
internal\slo\alerting.go:653:7: string `warning` has 3 occurrences, but such constant `SeverityWarning` already exists (goconst)
        case "warning":
             ^
internal\slo\alerting.go:651:7: string `critical` has 3 occurrences, but such constant `SLOStatusCritical` already exists (goconst)
        case "critical":
             ^
internal\config\tls.go:147:7: string `1.3` has 5 occurrences, make it a constant (goconst)
        case "1.3":
             ^
internal\config\tls.go:145:7: string `1.2` has 5 occurrences, make it a constant (goconst)
        case "1.2":
             ^
internal\metrics\business.go:758:40: string `resolved` has 3 occurrences, make it a constant (goconst)
                if !exists || existingState.State == "resolved" {
                                                     ^
internal\lifecycle\manager.go:37:10: string `healthy` has 3 occurrences, but such constant `HealthStatusHealthy` already exists (goconst)
                return "healthy"
                       ^
internal\repository\postgres\task_repository.go:109:1: cyclomatic complexity 16 of func `(*TaskRepository).List` is high (> 15) (gocyclo)
func (r *TaskRepository) List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error) {
^
internal\metrics\business.go:717:1: cyclomatic complexity 16 of func `(*BusinessMetricsCollector).evaluateAlertRule` is high (> 15) (gocyclo)
func (bmc *BusinessMetricsCollector) evaluateAlertRule(rule MetricAlertRule) {
^
internal\ai\router\router.go:52:15: G304: Potential file inclusion via variable (gosec)
        if b, err := os.ReadFile(ff); err == nil {
                     ^
internal\ai\router\router.go:55:15: G304: Potential file inclusion via variable (gosec)
        if b, err := os.ReadFile(rules); err == nil {
                     ^
automation\autocommit.go:50:10: G301: Expect directory permissions to be 0750 or less (gosec)
                return os.MkdirAll(path, 0755)
                       ^
automation\autocommit.go:103:13: G306: Expect WriteFile permissions to be 0600 or less (gosec)
                if err := ioutil.WriteFile(gitignorePath, []byte(config.GitIgnore), 0644); err != nil {
                          ^
automation\autocommit.go:117:13: G306: Expect WriteFile permissions to be 0600 or less (gosec)
                if err := ioutil.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
                          ^
automation\autocommit.go:199:15: G304: Potential file inclusion via variable (gosec)
        data, err := ioutil.ReadFile(filename)
                     ^
automation\autocommit.go:219:12: G306: Expect WriteFile permissions to be 0600 or less (gosec)
        if err := ioutil.WriteFile(filename, data, 0644); err != nil {
                  ^
internal\config\secrets\loader.go:108:15: G304: Potential file inclusion via variable (gosec)
        data, err := os.ReadFile(configPath)
                     ^
internal\config\config.go:285:15: G304: Potential file inclusion via variable (gosec)
        file, err := os.Open(filename)
                     ^
internal\config\tls.go:96:16: G402: TLS MinVersion too low. (gosec)
        tlsConfig := &tls.Config{
                Certificates:             []tls.Certificate{cert},
                PreferServerCipherSuites: true,
                CurvePreferences: []tls.CurveID{
                        tls.X25519,
                        tls.CurveP256,
                        tls.CurveP384,
                        tls.CurveP521,
                },
        }
internal\config\tls.go:251:22: G304: Potential file inclusion via variable (gosec)
                clientCert, err := os.ReadFile(certFile)
                                   ^
internal\repository\postgres\task_repository.go:173:11: G202: SQL string concatenation (gosec)
        query := `
                SELECT id, title, description, status, priority, assignee_id, created_by,
                       created_at, updated_at, completed_at, due_date, tags, metadata
                FROM tasks ` + whereClause + `
                ORDER BY created_at DESC
                LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)
internal\lifecycle\deployment.go:537:9: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
        cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
               ^
internal\metrics\storage.go:216:3: redefines-builtin-id: redefinition of the built-in function max (revive)
                max := values[0].Value
                ^
internal\metrics\storage.go:219:5: redefines-builtin-id: redefinition of the built-in function max (revive)
                                max = value.Value
                                ^
internal\metrics\storage.go:225:3: redefines-builtin-id: redefinition of the built-in function min (revive)
                min := values[0].Value
                ^
internal\metrics\storage.go:228:5: redefines-builtin-id: redefinition of the built-in function min (revive)
                                min = value.Value
                                ^
internal\repository\postgres\connection.go:7:2: blank-imports: a blank import should be only in a main or test package, or have a comment justifying it (revive)
        _ "github.com/lib/pq"
        ^
internal\ratelimit\distributed.go:36:2: field `mu` is unused (unused)
        mu       sync.RWMutex
        ^
internal\metrics\storage.go:195:1: calculated cyclomatic complexity for function calculateAggregation is 16, max is 15 (cyclop)
func (mms *MemoryMetricStorage) calculateAggregation(values []MetricValue, aggType AggregationType) float64 {
^
internal\config\secrets\loader.go:10:2: import 'github.com/hashicorp/vault/api' is not allowed from list 'Main' (depguard)
        "github.com/hashicorp/vault/api"
        ^
internal\config\secrets\loader.go:11:2: import 'gopkg.in/yaml.v3' is not allowed from list 'Main' (depguard)
        "gopkg.in/yaml.v3"
        ^
internal\events\nats_bus.go:9:2: import 'github.com/nats-io/nats.go' is not allowed from list 'Main' (depguard)
        "github.com/nats-io/nats.go"
        ^
internal\events\nats_bus.go:10:2: import 'go.uber.org/zap' is not allowed from list 'Main' (depguard)
        "go.uber.org/zap"
        ^
internal\events\nats_bus.go:12:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
        ^
internal\http\router.go:8:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features"
        ^
internal\lifecycle\deployment.go:10:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"
        ^
internal\lifecycle\health.go:11:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"
        ^
internal\lifecycle\manager.go:10:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"
        ^
internal\lifecycle\manager.go:11:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
        ^
internal\lifecycle\operations.go:9:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"
        ^
internal\metrics\business.go:9:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"
        ^
internal\metrics\business.go:10:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
        ^
internal\nats\publisher_error_handler.go:9:2: import 'github.com/nats-io/nats.go' is not allowed from list 'Main' (depguard)
        "github.com/nats-io/nats.go"
        ^
internal\ratelimit\distributed.go:10:2: import 'github.com/redis/go-redis/v9' is not allowed from list 'Main' (depguard)
        "github.com/redis/go-redis/v9"
        ^
internal\ratelimit\distributed.go:12:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"
        ^
internal\ratelimit\distributed.go:13:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
        ^
internal\repository\postgres\connection.go:8:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
        ^
internal\repository\postgres\task_repository.go:11:2: import 'github.com/google/uuid' is not allowed from list 'Main' (depguard)
        "github.com/google/uuid"
        ^
internal\repository\postgres\task_repository.go:12:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
        ^
internal\repository\redis\cache_repository.go:9:2: import 'github.com/redis/go-redis/v9' is not allowed from list 'Main' (depguard)
        "github.com/redis/go-redis/v9"
        ^
internal\repository\redis\connection.go:7:2: import 'github.com/redis/go-redis/v9' is not allowed from list 'Main' (depguard)
        "github.com/redis/go-redis/v9"
        ^
internal\repository\redis\connection.go:8:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
        ^
internal\slo\alerting.go:13:2: import 'go.uber.org/zap' is not allowed from list 'Main' (depguard)
        "go.uber.org/zap"
        ^
internal\slo\monitor.go:9:2: import 'github.com/prometheus/client_golang/api' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/api"
        ^
internal\slo\monitor.go:10:2: import 'github.com/prometheus/client_golang/api/prometheus/v1' is not allowed from list 'Main' (depguard)
        v1 "github.com/prometheus/client_golang/api/prometheus/v1"
        ^
internal\slo\monitor.go:11:2: import 'github.com/prometheus/common/model' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/common/model"
        ^
internal\slo\monitor.go:12:2: import 'go.uber.org/zap' is not allowed from list 'Main' (depguard)
        "go.uber.org/zap"
        ^
internal\tracing\business.go:10:2: import 'go.opentelemetry.io/otel' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel"
        ^
internal\tracing\business.go:11:2: import 'go.opentelemetry.io/otel/attribute' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/attribute"
        ^
internal\tracing\business.go:12:2: import 'go.opentelemetry.io/otel/baggage' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/baggage"
        ^
internal\tracing\business.go:13:2: import 'go.opentelemetry.io/otel/codes' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/codes"
        ^
internal\tracing\business.go:14:2: import 'go.opentelemetry.io/otel/trace' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/trace"
        ^
internal\tracing\business.go:16:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"
        ^
internal\tracing\business.go:17:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
        ^
main.go:13:2: import 'github.com/go-chi/chi/v5' is not allowed from list 'Main' (depguard)
        "github.com/go-chi/chi/v5"
        ^
main.go:14:2: import 'github.com/go-chi/chi/v5/middleware' is not allowed from list 'Main' (depguard)
        "github.com/go-chi/chi/v5/middleware"
        ^
main.go:15:2: import 'github.com/go-chi/cors' is not allowed from list 'Main' (depguard)
        "github.com/go-chi/cors"
        ^
main.go:16:2: import 'github.com/prometheus/client_golang/prometheus/promhttp' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/prometheus/promhttp"
        ^
main.go:17:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"
        ^
main.go:18:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/version' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/version"
        ^
main.go:19:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
        ^
main.go:20:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/handlers' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/handlers"
        ^
internal\ai\telemetry\metrics.go:7:2: import 'github.com/prometheus/client_golang/prometheus' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/prometheus"
        ^
internal\ai\telemetry\metrics.go:8:2: import 'github.com/prometheus/client_golang/prometheus/promauto' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/prometheus/promauto"
        ^
internal\ai\wiring\wiring.go:9:2: import 'github.com/prometheus/client_golang/prometheus' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/prometheus"
        ^
internal\ai\wiring\wiring.go:11:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/router' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/router"
        ^
internal\ai\wiring\wiring.go:12:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/telemetry' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/telemetry"
        ^
internal\ai\wiring\wiring_test.go:9:2: import 'github.com/prometheus/client_golang/prometheus' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/prometheus"
        ^
internal\config\config.go:8:2: import 'github.com/kelseyhightower/envconfig' is not allowed from list 'Main' (depguard)
        "github.com/kelseyhightower/envconfig"
        ^
internal\config\config.go:9:2: import 'gopkg.in/yaml.v3' is not allowed from list 'Main' (depguard)
        "gopkg.in/yaml.v3"
        ^
internal\config\config.go:11:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/security' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/security"
        ^
internal\config\tls.go:10:2: import 'go.uber.org/zap' is not allowed from list 'Main' (depguard)
        "go.uber.org/zap"
        ^
internal\config\tls_test.go:9:2: import 'github.com/stretchr/testify/assert' is not allowed from list 'Main' (depguard)
        "github.com/stretchr/testify/assert"
        ^
internal\config\tls_test.go:10:2: import 'github.com/stretchr/testify/require' is not allowed from list 'Main' (depguard)
        "github.com/stretchr/testify/require"
        ^
internal\config\tls_test.go:11:2: import 'go.uber.org/zap/zaptest' is not allowed from list 'Main' (depguard)
        "go.uber.org/zap/zaptest"
        ^
internal\domain\models.go:6:2: import 'github.com/google/uuid' is not allowed from list 'Main' (depguard)
        "github.com/google/uuid"
        ^
internal\domain\repository.go:6:2: import 'github.com/google/uuid' is not allowed from list 'Main' (depguard)
        "github.com/google/uuid"
        ^
internal\domain\models_test.go:7:2: import 'github.com/google/uuid' is not allowed from list 'Main' (depguard)
        "github.com/google/uuid"
        ^
internal\domain\models_test.go:8:2: import 'github.com/stretchr/testify/assert' is not allowed from list 'Main' (depguard)
        "github.com/stretchr/testify/assert"
        ^
internal\telemetry\metrics.go:8:2: import 'github.com/prometheus/client_golang/prometheus' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/prometheus"
        ^
internal\telemetry\metrics.go:9:2: import 'github.com/prometheus/client_golang/prometheus/promauto' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/prometheus/promauto"
        ^
internal\telemetry\telemetry.go:10:2: import 'github.com/go-chi/chi/v5/middleware' is not allowed from list 'Main' (depguard)
        "github.com/go-chi/chi/v5/middleware"
        ^
internal\telemetry\telemetry.go:11:2: import 'github.com/prometheus/client_golang/prometheus' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/prometheus"
        ^
internal\telemetry\telemetry.go:12:2: import 'github.com/prometheus/client_golang/prometheus/promauto' is not allowed from list 'Main' (depguard)
        "github.com/prometheus/client_golang/prometheus/promauto"
        ^
internal\telemetry\telemetry.go:13:2: import 'go.opentelemetry.io/otel' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel"
        ^
internal\telemetry\telemetry.go:14:2: import 'go.opentelemetry.io/otel/attribute' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/attribute"
        ^
internal\telemetry\telemetry.go:15:2: import 'go.opentelemetry.io/otel/exporters/prometheus' is not allowed from list 'Main' (depguard)
        promexporter "go.opentelemetry.io/otel/exporters/prometheus"
        ^
internal\telemetry\telemetry.go:16:2: import 'go.opentelemetry.io/otel/metric' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/metric"
        ^
internal\telemetry\telemetry.go:17:2: import 'go.opentelemetry.io/otel/sdk/metric' is not allowed from list 'Main' (depguard)
        sdkmetric "go.opentelemetry.io/otel/sdk/metric"
        ^
internal\telemetry\telemetry.go:18:2: import 'go.uber.org/zap' is not allowed from list 'Main' (depguard)
        "go.uber.org/zap"
        ^
internal\telemetry\telemetry.go:20:2: import 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config' is not allowed from list 'Main' (depguard)
        "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
        ^
internal\telemetry\tracing.go:8:2: import 'go.opentelemetry.io/otel' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel"
        ^
internal\telemetry\tracing.go:9:2: import 'go.opentelemetry.io/otel/attribute' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/attribute"
        ^
internal\telemetry\tracing.go:10:2: import 'go.opentelemetry.io/otel/codes' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/codes"
        ^
internal\telemetry\tracing.go:11:2: import 'go.opentelemetry.io/otel/exporters/jaeger' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/exporters/jaeger"
        ^
internal\telemetry\tracing.go:12:2: import 'go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
        ^
internal\telemetry\tracing.go:13:2: import 'go.opentelemetry.io/otel/exporters/stdout/stdouttrace' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
        ^
internal\telemetry\tracing.go:14:2: import 'go.opentelemetry.io/otel/propagation' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/propagation"
        ^
internal\telemetry\tracing.go:15:2: import 'go.opentelemetry.io/otel/sdk/resource' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/sdk/resource"
        ^
internal\telemetry\tracing.go:16:2: import 'go.opentelemetry.io/otel/sdk/trace' is not allowed from list 'Main' (depguard)
        sdktrace "go.opentelemetry.io/otel/sdk/trace"
        ^
internal\telemetry\tracing.go:17:2: import 'go.opentelemetry.io/otel/semconv/v1.26.0' is not allowed from list 'Main' (depguard)
        semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
        ^
internal\telemetry\tracing.go:18:2: import 'go.opentelemetry.io/otel/trace' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/trace"
        ^
internal\telemetry\tracing.go:19:2: import 'go.uber.org/zap' is not allowed from list 'Main' (depguard)
        "go.uber.org/zap"
        ^
internal\telemetry\tracing_test.go:8:2: import 'github.com/stretchr/testify/assert' is not allowed from list 'Main' (depguard)
        "github.com/stretchr/testify/assert"
        ^
internal\telemetry\tracing_test.go:9:2: import 'github.com/stretchr/testify/require' is not allowed from list 'Main' (depguard)
        "github.com/stretchr/testify/require"
        ^
internal\telemetry\tracing_test.go:10:2: import 'go.opentelemetry.io/otel/attribute' is not allowed from list 'Main' (depguard)
        "go.opentelemetry.io/otel/attribute"
        ^
internal\telemetry\tracing_test.go:11:2: import 'go.uber.org/zap/zaptest' is not allowed from list 'Main' (depguard)
        "go.uber.org/zap/zaptest"
        ^
main.go:76:17: Multiplication of durations: `time.Duration(cfg.Server.ReadTimeout) * time.Second` (durationcheck)
                ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
                              ^
main.go:77:17: Multiplication of durations: `time.Duration(cfg.Server.WriteTimeout) * time.Second` (durationcheck)
                WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
                              ^
main.go:78:17: Multiplication of durations: `time.Duration(cfg.Server.IdleTimeout) * time.Second` (durationcheck)
                IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
                              ^
internal\ratelimit\distributed.go:631:5: comparing with == will fail on wrapped errors. Use errors.Is to check for a specific error (errorlint)
        if err == redis.Nil {
           ^
internal\repository\postgres\task_repository.go:276:6: comparing with == will fail on wrapped errors. Use errors.Is to check for a specific error (errorlint)
                if err == sql.ErrNoRows {
                   ^
internal\repository\redis\cache_repository.go:45:5: comparing with == will fail on wrapped errors. Use errors.Is to check for a specific error (errorlint)
        if err == redis.Nil {
           ^
internal\config\secrets\loader.go:153:2: missing cases in switch of type config.SecretsBackendType: config.SecretsBackendEnv (exhaustive)
        switch sl.backendType {
        ^
internal\lifecycle\health.go:412:3: missing cases in switch of type lifecycle.HealthStatus: lifecycle.HealthStatusUnknown (exhaustive)
                switch check.Status {
                ^
internal\lifecycle\health.go:464:3: missing cases in switch of type lifecycle.HealthStatus: lifecycle.HealthStatusUnknown (exhaustive)
                switch report.Status {
                ^
internal\metrics\business.go:667:3: missing cases in switch of type metrics.AggregationType: metrics.AggregationP95, metrics.AggregationP99 (exhaustive)
                switch aggType {
                ^
internal\slo\monitor.go:382:2: missing cases in switch of type model.ValueType: model.ValNone, model.ValMatrix, model.ValString (exhaustive)
        switch result.Type() {
        ^
internal\metrics\business.go:186:6: Function 'DefaultBusinessMetrics' is too long (118 > 100) (funlen)
func DefaultBusinessMetrics() []BusinessMetric {
     ^
internal\slo\config.go:8:6: Function 'DefaultSLOs' is too long (363 > 100) (funlen)
func DefaultSLOs() []*SLO {
     ^
automation\autocommit.go:73:24: hugeParam: config is heavy (152 bytes); consider passing it by pointer (gocritic)
func initializeGitRepo(config Config) error {
                       ^
automation\autocommit.go:134:20: hugeParam: config is heavy (152 bytes); consider passing it by pointer (gocritic)
func commitAndPush(config Config) error {
                   ^
automation\autocommit.go:213:23: hugeParam: config is heavy (152 bytes); consider passing it by pointer (gocritic)
func saveConfigToFile(config Config, filename string) error {
                      ^
internal\config\secrets\loader.go:254:5: emptyStringTest: replace `len(value) == 0` with `value == ""` (gocritic)
        if len(value) == 0 {
           ^
internal\lifecycle\deployment.go:134:30: hugeParam: config is heavy (344 bytes); consider passing it by pointer (gocritic)
func NewDeploymentAutomation(config DeploymentConfig, logger logger.Logger) *DeploymentAutomation {
                             ^
internal\lifecycle\deployment.go:515:66: hugeParam: hook is heavy (104 bytes); consider passing it by pointer (gocritic)
func (da *DeploymentAutomation) executeHook(ctx context.Context, hook DeploymentHook, result *DeploymentResult) error {
                                                                 ^
internal\lifecycle\deployment.go:567:70: hugeParam: hook is heavy (104 bytes); consider passing it by pointer (gocritic)
func (da *DeploymentAutomation) executeHTTPHook(ctx context.Context, hook DeploymentHook, result *DeploymentResult) error {
                                                                     ^
internal\lifecycle\deployment.go:624:56: hugeParam: result is heavy (216 bytes); consider passing it by pointer (gocritic)
func (da *DeploymentAutomation) addDeploymentToHistory(result DeploymentResult) {
                                                       ^
internal\lifecycle\health.go:147:23: hugeParam: config is heavy (128 bytes); consider passing it by pointer (gocritic)
func NewHealthMonitor(config HealthConfig, version string, logger logger.Logger) *HealthMonitor {
                      ^
internal\lifecycle\health.go:432:2: ifElseChain: rewrite if-else to switch statement (gocritic)
        if failures == 0 {
        ^
internal\lifecycle\manager.go:621:2: ifElseChain: rewrite if-else to switch statement (gocritic)
        if errorCount == 0 && healthyCount == totalComponents {
        ^
internal\lifecycle\operations.go:597:2: rangeValCopy: each iteration copies 136 bytes (consider pointers or indexing) (gocritic)
        for i, step := range operation.Steps {
        ^
internal\metrics\business.go:365:2: hugeParam: config is heavy (160 bytes); consider passing it by pointer (gocritic)
        config BusinessMetricsConfig,
        ^
internal\metrics\business.go:480:54: hugeParam: query is heavy (80 bytes); consider passing it by pointer (gocritic)
func (bmc *BusinessMetricsCollector) GetMetricValues(query MetricQuery) ([]MetricValue, error) {
                                                     ^
internal\metrics\business.go:505:59: hugeParam: query is heavy (136 bytes); consider passing it by pointer (gocritic)
func (bmc *BusinessMetricsCollector) GetAggregatedMetrics(query AggregationQuery) ([]AggregatedMetric, error) {
                                                          ^
internal\metrics\business.go:871:70: hugeParam: query is heavy (80 bytes); consider passing it by pointer (gocritic)
func (bmc *BusinessMetricsCollector) matchesQuery(value MetricValue, query MetricQuery) bool {
                                                                     ^
internal\metrics\storage.go:40:60: hugeParam: query is heavy (80 bytes); consider passing it by pointer (gocritic)
func (mms *MemoryMetricStorage) Query(ctx context.Context, query MetricQuery) ([]MetricValue, error) {
                                                           ^
internal\metrics\storage.go:70:64: hugeParam: query is heavy (136 bytes); consider passing it by pointer (gocritic)
func (mms *MemoryMetricStorage) Aggregate(ctx context.Context, query AggregationQuery) ([]AggregatedMetric, error) {
                                                               ^
internal\metrics\storage.go:137:65: hugeParam: query is heavy (80 bytes); consider passing it by pointer (gocritic)
func (mms *MemoryMetricStorage) matchesQuery(value MetricValue, query MetricQuery) bool {
                                                                ^
internal\ratelimit\distributed.go:499:35: emptyStringTest: replace `len(fmt.Sprintf("%v", condition.Value)) > 0` with `fmt.Sprintf("%v", condition.Value) != ""` (gocritic)
                return len(requestValue) > 0 && len(fmt.Sprintf("%v", condition.Value)) > 0
                                                ^
internal\ratelimit\distributed.go:501:10: emptyStringTest: replace `len(requestValue) > 0` with `requestValue != ""` (gocritic)
                return len(requestValue) > 0 && fmt.Sprintf("%v", condition.Value) != ""
                       ^
internal\ratelimit\distributed.go:503:10: emptyStringTest: replace `len(requestValue) > 0` with `requestValue != ""` (gocritic)
                return len(requestValue) > 0 && fmt.Sprintf("%v", condition.Value) != ""
                       ^
internal\ratelimit\distributed.go:222:54: hugeParam: config is heavy (112 bytes); consider passing it by pointer (gocritic)
func NewDistributedRateLimiter(client redis.Cmdable, config Config, logger logger.Logger, telemetry *observability.TelemetryService) (*DistributedRateLimiter, error) {
                                                     ^
internal\ratelimit\distributed.go:276:63: hugeParam: request is heavy (120 bytes); consider passing it by pointer (gocritic)
func (drl *DistributedRateLimiter) Allow(ctx context.Context, request Request) (*Response, error) {
                                                              ^
internal\ratelimit\distributed.go:313:71: hugeParam: request is heavy (120 bytes); consider passing it by pointer (gocritic)
func (drl *DistributedRateLimiter) AllowWithRule(ctx context.Context, request Request, rule Rule) (*Response, error) {
                                                                      ^
internal\ratelimit\distributed.go:435:48: hugeParam: request is heavy (120 bytes); consider passing it by pointer (gocritic)
func (drl *DistributedRateLimiter) generateKey(request Request) string {
                                               ^
internal\ratelimit\distributed.go:446:52: hugeParam: rule is heavy (248 bytes); consider passing it by pointer (gocritic)
func (drl *DistributedRateLimiter) generateRuleKey(rule Rule, request Request) string {
                                                   ^
internal\ratelimit\distributed.go:458:52: hugeParam: request is heavy (120 bytes); consider passing it by pointer (gocritic)
func (drl *DistributedRateLimiter) getRequestField(request Request, field string) string {
                                                   ^
internal\ratelimit\distributed.go:476:79: hugeParam: request is heavy (120 bytes); consider passing it by pointer (gocritic)
func (drl *DistributedRateLimiter) evaluateConditions(conditions []Condition, request Request) bool {
                                                                              ^
internal\ratelimit\distributed.go:490:75: hugeParam: request is heavy (120 bytes); consider passing it by pointer (gocritic)
func (drl *DistributedRateLimiter) evaluateCondition(condition Condition, request Request) bool {
                                                                          ^
internal\ratelimit\distributed.go:509:65: hugeParam: rule is heavy (248 bytes); consider passing it by pointer (gocritic)
func (drl *DistributedRateLimiter) getAdaptiveLimit(key string, rule Rule) int64 {
                                                                ^
internal\ratelimit\distributed.go:518:68: hugeParam: rule is heavy (248 bytes); consider passing it by pointer (gocritic)
func (drl *DistributedRateLimiter) updateAdaptiveState(key string, rule Rule, allowed bool) {
                                                                   ^
internal\ratelimit\distributed.go:711:57: hugeParam: rule is heavy (248 bytes); consider passing it by pointer (gocritic)
func (al *AdaptiveLimiter) getAdaptiveLimit(key string, rule Rule) int64 {
                                                        ^
internal\ratelimit\distributed.go:733:52: hugeParam: rule is heavy (248 bytes); consider passing it by pointer (gocritic)
func (al *AdaptiveLimiter) updateState(key string, rule Rule, allowed bool) {
                                                   ^
internal\repository\postgres\connection.go:12:14: hugeParam: cfg is heavy (112 bytes); consider passing it by pointer (gocritic)
func Connect(cfg config.PostgreSQLConfig) (*sql.DB, error) {
             ^
internal\slo\alerting.go:119:22: hugeParam: config is heavy (128 bytes); consider passing it by pointer (gocritic)
func NewAlertManager(config AlertingConfig, logger *zap.Logger) *AlertManager {
                     ^
internal\slo\alerting.go:156:35: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) SendAlert(alert AlertEvent) {
                                  ^
internal\slo\alerting.go:188:38: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) processAlert(alert AlertEvent) error {
                                     ^
internal\slo\alerting.go:303:39: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) isRateLimited(alert AlertEvent) bool {
                                      ^
internal\slo\alerting.go:336:43: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) storeAlertHistory(alert AlertEvent) {
                                          ^
internal\slo\alerting.go:358:39: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) sendToChannel(alert AlertEvent, channel AlertChannel) error {
                                      ^
internal\slo\alerting.go:384:37: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) sendToSlack(alert AlertEvent, config ChannelConfig) error {
                                    ^
internal\slo\alerting.go:427:39: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) sendToDiscord(alert AlertEvent, config ChannelConfig) error {
                                      ^
internal\slo\alerting.go:465:39: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) sendToWebhook(alert AlertEvent, config ChannelConfig) error {
                                      ^
internal\slo\alerting.go:480:37: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) sendToEmail(alert AlertEvent, config ChannelConfig) error {
                                    ^
internal\slo\alerting.go:488:41: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) sendToPagerDuty(alert AlertEvent, config ChannelConfig) error {
                                        ^
internal\slo\alerting.go:496:39: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) sendToMSTeams(alert AlertEvent, config ChannelConfig) error {
                                      ^
internal\slo\alerting.go:535:41: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) startEscalation(alert AlertEvent) {
                                        ^
internal\slo\alerting.go:560:43: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) executeEscalation(alert AlertEvent, policy EscalationPolicy) {
                                          ^
internal\slo\alerting.go:633:57: hugeParam: alert is heavy (104 bytes); consider passing it by pointer (gocritic)
func (am *AlertManager) renderTemplate(template string, alert AlertEvent) string {
                                                        ^
internal\tracing\business.go:271:35: hugeParam: config is heavy (232 bytes); consider passing it by pointer (gocritic)
func NewBusinessTransactionTracer(config TracingConfig, logger logger.Logger, telemetry *observability.TelemetryService) (*BusinessTransactionTracer, error) {
                                  ^
internal\tracing\business.go:512:2: rangeValCopy: each iteration copies 152 bytes (consider pointers or indexing) (gocritic)
        for i, s := range transaction.Steps {
        ^
basic_test.go:18:5: dupSubExpr: suspicious identical LHS and RHS for `!=` operator (gocritic)
        if true != true {
           ^
internal\ai\events\handlers.go:57:85: hugeParam: e is heavy (128 bytes); consider passing it by pointer (gocritic)
func PublishRouterDecision(ctx context.Context, pub EventPublisher, subject string, e RouterDecision) error {
                                                                                    ^
internal\ai\events\handlers.go:63:82: hugeParam: e is heavy (112 bytes); consider passing it by pointer (gocritic)
func PublishPolicyBlock(ctx context.Context, pub EventPublisher, subject string, e PolicyBlock) error {
                                                                                 ^
internal\ai\events\handlers.go:69:85: hugeParam: e is heavy (128 bytes); consider passing it by pointer (gocritic)
func PublishInferenceError(ctx context.Context, pub EventPublisher, subject string, e InferenceError) error {
                                                                                    ^
internal\ai\events\handlers.go:75:87: hugeParam: e is heavy (120 bytes); consider passing it by pointer (gocritic)
func PublishInferenceSummary(ctx context.Context, pub EventPublisher, subject string, e InferenceSummary) error {
                                                                                      ^
internal\ai\telemetry\metrics.go:90:23: hugeParam: meta is heavy (232 bytes); consider passing it by pointer (gocritic)
func ObserveInference(meta InferenceMeta) {
                      ^
internal\ai\telemetry\metrics.go:112:21: hugeParam: l is heavy (160 bytes); consider passing it by pointer (gocritic)
func IncPolicyBlock(l Labels) {
                    ^
internal\ai\telemetry\metrics.go:119:24: hugeParam: l is heavy (160 bytes); consider passing it by pointer (gocritic)
func IncRouterDecision(l Labels) {
                       ^
internal\ai\wiring\wiring_test.go:16:31: octalLiteral: use new octal literal style, 0o755 (gocritic)
        if err := os.MkdirAll(aiDir, 0755); err != nil {
                                     ^
internal\ai\wiring\wiring_test.go:22:91: octalLiteral: use new octal literal style, 0o644 (gocritic)
        if err := os.WriteFile(filepath.Join(aiDir, "feature_flags.json"), []byte(flagsContent), 0644); err != nil {
                                                                                                 ^
internal\ai\wiring\wiring_test.go:55:35: octalLiteral: use new octal literal style, 0o755 (gocritic)
        if err := os.MkdirAll(configDir, 0755); err != nil {
                                         ^
internal\ai\wiring\wiring_test.go:61:91: octalLiteral: use new octal literal style, 0o644 (gocritic)
        if err := os.WriteFile(filepath.Join(aiDir, "feature_flags.json"), []byte(flagsContent), 0644); err != nil {
                                                                                                 ^
internal\ai\wiring\wiring_test.go:75:97: octalLiteral: use new octal literal style, 0o644 (gocritic)
        if err := os.WriteFile(filepath.Join(configDir, "ai-router.rules.json"), []byte(rulesContent), 0644); err != nil {
                                                                                                       ^
internal\config\config.go:304:7: hugeParam: p is heavy (112 bytes); consider passing it by pointer (gocritic)
func (p PostgreSQLConfig) DSN() string {
      ^
internal\handlers\health_test.go:11:9: httpNoBody: http.NoBody should be preferred to the nil request body (gocritic)
        req := httptest.NewRequest(http.MethodGet, "/livez", nil)
               ^
internal\handlers\health_test.go:28:9: httpNoBody: http.NoBody should be preferred to the nil request body (gocritic)
        req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
               ^
internal\handlers\health_test.go:40:9: httpNoBody: http.NoBody should be preferred to the nil request body (gocritic)
        req := httptest.NewRequest(http.MethodGet, "/health", nil)
               ^
internal\handlers\health_test.go:52:9: httpNoBody: http.NoBody should be preferred to the nil request body (gocritic)
        req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
               ^
internal\handlers\health_test.go:69:9: httpNoBody: http.NoBody should be preferred to the nil request body (gocritic)
        req := httptest.NewRequest(http.MethodGet, "/livez", nil)
               ^
internal\handlers\health_test.go:86:9: httpNoBody: http.NoBody should be preferred to the nil request body (gocritic)
        req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
               ^
internal\telemetry\telemetry.go:85:11: hugeParam: cfg is heavy (272 bytes); consider passing it by pointer (gocritic)
func Init(cfg config.TelemetryConfig) (*Telemetry, error) {
          ^
internal\telemetry\tracing.go:170:3: appendCombine: can combine chain of 2 appends into one (gocritic)
                opts = append(opts, jaeger.WithUsername(config.JaegerUser))
                ^
automation\autocommit.go:16:1: Comment should end in a period (godot)
// Config represents the configuration for the auto-commit tool
^
automation\autocommit.go:31:1: Comment should end in a period (godot)
// DefaultConfig returns a default configuration
^
automation\autocommit.go:46:1: Comment should end in a period (godot)
// ensureDirectory creates directory structure if it doesn't exist
^
automation\autocommit.go:55:1: Comment should end in a period (godot)
// runCommand executes a shell command and returns output
^
automation\autocommit.go:72:1: Comment should end in a period (godot)
// initializeGitRepo initializes a git repository if it doesn't exist
^
automation\autocommit.go:133:1: Comment should end in a period (godot)
// commitAndPush commits changes and pushes to GitHub
^
automation\autocommit.go:190:1: Comment should end in a period (godot)
// loadConfigFromFile loads configuration from JSON file
^
automation\autocommit.go:212:1: Comment should end in a period (godot)
// saveConfigToFile saves configuration to JSON file
^
automation\autocommit.go:227:1: Comment should end in a period (godot)
// interactiveConfig allows user to input configuration interactively
^
internal\config\secrets\loader.go:14:1: Comment should end in a period (godot)
// SecretsBackendType define o tipo de backend de secrets
^
internal\config\secrets\loader.go:23:1: Comment should end in a period (godot)
// SecretsConfig representa a configuração de secrets
^
internal\config\secrets\loader.go:36:1: Comment should end in a period (godot)
// SecretsBackendConfig configura o backend de secrets
^
internal\config\secrets\loader.go:42:1: Comment should end in a period (godot)
// VaultConfig configuração do Vault
^
internal\config\secrets\loader.go:49:1: Comment should end in a period (godot)
// DatabaseSecrets secrets do banco de dados
^
internal\config\secrets\loader.go:59:1: Comment should end in a period (godot)
// NATSSecrets secrets do NATS
^
internal\config\secrets\loader.go:67:1: Comment should end in a period (godot)
// TelemetrySecrets secrets de telemetria
^
internal\config\secrets\loader.go:73:1: Comment should end in a period (godot)
// OTLPSecrets configuração OTLP
^
internal\config\secrets\loader.go:79:1: Comment should end in a period (godot)
// PrometheusSecrets configuração Prometheus
^
internal\config\secrets\loader.go:85:1: Comment should end in a period (godot)
// AuthSecrets secrets de autenticação
^
internal\config\secrets\loader.go:91:1: Comment should end in a period (godot)
// EncryptionSecrets secrets de criptografia
^
internal\config\secrets\loader.go:97:1: Comment should end in a period (godot)
// SecretsLoader carrega secrets de diferentes fontes
^
internal\config\secrets\loader.go:106:1: Comment should end in a period (godot)
// NewSecretsLoader cria um novo loader de secrets
^
internal\config\secrets\loader.go:145:1: Comment should end in a period (godot)
// Load carrega todos os secrets
^
internal\config\secrets\loader.go:164:1: Comment should end in a period (godot)
// initVaultClient inicializa o cliente Vault
^
internal\config\secrets\loader.go:184:1: Comment should end in a period (godot)
// loadFromVault carrega secrets do Vault
^
internal\config\secrets\loader.go:206:1: Comment should end in a period (godot)
// loadFromK8s carrega secrets do Kubernetes
^
internal\config\secrets\loader.go:212:1: Comment should end in a period (godot)
// validateRequiredSecrets valida se todos os secrets obrigatórios estão presentes
^
internal\config\secrets\loader.go:230:1: Comment should end in a period (godot)
// GetDatabaseDSN retorna a DSN do banco de dados de forma segura
^
internal\config\secrets\loader.go:239:1: Comment should end in a period (godot)
// GetNATSConnection retorna a string de conexão NATS
^
internal\config\secrets\loader.go:252:1: Comment should end in a period (godot)
// Redact remove informações sensíveis para logs
^
internal\config\secrets\loader.go:263:1: Comment should end in a period (godot)
// SecureString representa uma string segura que não aparece em logs
^
internal\config\secrets\loader.go:268:1: Comment should end in a period (godot)
// NewSecureString cria uma nova string segura
^
internal\config\secrets\loader.go:273:1: Comment should end in a period (godot)
// String implementa Stringer e redact o valor
^
internal\config\secrets\loader.go:283:1: Comment should end in a period (godot)
// MarshalJSON implementa json.Marshaler
^
internal\constants\test_constants.go:3:1: Comment should end in a period (godot)
// Non-sensitive test constants (not secrets)
^
internal\constants\test_constants.go:5:2: Comment should end in a period (godot)
        // JWT Testing Constants (non-secret)
        ^
internal\constants\test_constants.go:11:2: Comment should end in a period (godot)
        // Database Testing Constants (non-secret)
        ^
internal\constants\test_constants.go:16:1: Comment should end in a period (godot)
// Deprecated: Use GetTestSecret() for runtime-generated secrets instead
^
internal\constants\test_constants.go:29:1: Comment should end in a period (godot)
// TestCredentials provides a structured way to access test credentials
^
internal\constants\test_constants.go:38:1: Comment should end in a period (godot)
// GetTestCredentials returns test credentials for containerized testing
^
internal\constants\test_constants.go:51:1: Comment should end in a period (godot)
// IsTestEnvironment checks if we're in a test environment
^
internal\constants\test_secrets.go:31:1: Comment should end in a period (godot)
// generateRandomSecret creates a cryptographically random string of the specified byte length
^
internal\constants\test_secrets.go:41:1: Comment should end in a period (godot)
// ResetTestSecrets clears the cached secrets (useful for test isolation)
^
internal\dashboard\models.go:7:1: Comment should end in a period (godot)
// SystemOverview represents the overall system status
^
internal\dashboard\models.go:16:1: Comment should end in a period (godot)
// SystemHealth represents overall system health status
^
internal\dashboard\models.go:25:1: Comment should end in a period (godot)
// ComponentStatus represents individual component status
^
internal\dashboard\models.go:37:1: Comment should end in a period (godot)
// OverviewMetrics represents key system metrics
^
internal\dashboard\models.go:50:1: Comment should end in a period (godot)
// Alert represents system alerts
^
internal\dashboard\models.go:65:1: Comment should end in a period (godot)
// AlertType represents different types of alerts
^
internal\dashboard\models.go:76:1: Comment should end in a period (godot)
// AlertSeverity represents alert severity levels
^
internal\dashboard\models.go:86:1: Comment should end in a period (godot)
// AlertStatus represents alert status
^
internal\dashboard\models.go:96:1: Comment should end in a period (godot)
// AlertAction represents available actions for alerts
^
internal\dashboard\models.go:104:1: Comment should end in a period (godot)
// RealtimeMetrics represents real-time system metrics
^
internal\dashboard\models.go:113:1: Comment should end in a period (godot)
// SystemMetrics represents system-level metrics
^
internal\dashboard\models.go:122:1: Comment should end in a period (godot)
// CPUMetrics represents CPU usage metrics
^
internal\dashboard\models.go:132:1: Comment should end in a period (godot)
// MemoryMetrics represents memory usage metrics
^
internal\dashboard\models.go:143:1: Comment should end in a period (godot)
// DiskMetrics represents disk usage metrics
^
internal\dashboard\models.go:154:1: Comment should end in a period (godot)
// NetworkMetrics represents network usage metrics
^
internal\dashboard\models.go:167:1: Comment should end in a period (godot)
// ProcessMetrics represents process-level metrics
^
internal\dashboard\models.go:176:1: Comment should end in a period (godot)
// PerformanceMetrics represents application performance metrics
^
internal\dashboard\models.go:187:1: Comment should end in a period (godot)
// ResponseTimeMetrics represents response time statistics
^
internal\dashboard\models.go:198:1: Comment should end in a period (godot)
// DatabaseMetrics represents database performance metrics
^
internal\dashboard\models.go:209:1: Comment should end in a period (godot)
// CacheMetricsData represents cache performance metrics
^
internal\dashboard\models.go:219:1: Comment should end in a period (godot)
// ErrorMetrics represents error tracking metrics
^
internal\dashboard\models.go:229:1: Comment should end in a period (godot)
// RecentError represents recent error information
^
internal\dashboard\models.go:241:1: Comment should end in a period (godot)
// TrafficMetrics represents traffic and usage metrics
^
internal\dashboard\models.go:253:1: Comment should end in a period (godot)
// TrafficPeak represents peak traffic information
^
internal\dashboard\models.go:261:1: Comment should end in a period (godot)
// BandwidthMetrics represents bandwidth usage
^
internal\dashboard\models.go:270:1: Comment should end in a period (godot)
// ChartData represents time-series data for charts
^
internal\dashboard\models.go:276:1: Comment should end in a period (godot)
// Dataset represents a data series in a chart
^
internal\dashboard\models.go:285:1: Comment should end in a period (godot)
// DashboardWidget represents a dashboard widget configuration
^
internal\dashboard\models.go:296:1: Comment should end in a period (godot)
// WidgetSize represents widget dimensions
^
internal\dashboard\models.go:302:1: Comment should end in a period (godot)
// WidgetPosition represents widget position
^
internal\dashboard\models.go:308:1: Comment should end in a period (godot)
// WebSocketMessage represents messages sent via WebSocket
^
internal\dashboard\models.go:316:1: Comment should end in a period (godot)
// SubscriptionRequest represents WebSocket subscription requests
^
internal\events\nats_bus.go:15:1: Comment should end in a period (godot)
// NATSEventBus implements EventBus using NATS
^
internal\events\nats_bus.go:21:1: Comment should end in a period (godot)
// NewNATSEventBus creates a new NATS event bus
^
internal\events\nats_bus.go:49:1: Comment should end in a period (godot)
// Publish publishes an event to NATS
^
internal\events\nats_bus.go:84:1: Comment should end in a period (godot)
// Subscribe subscribes to events of a specific type
^
internal\events\nats_bus.go:114:1: Comment should end in a period (godot)
// SubscribeQueue subscribes to events with queue group
^
internal\events\nats_bus.go:146:1: Comment should end in a period (godot)
// Close closes the NATS connection
^
internal\events\nats_bus.go:154:1: Comment should end in a period (godot)
// EventHandler defines the interface for event handlers
^
internal\events\nats_bus.go:159:1: Comment should end in a period (godot)
// EventHandlerFunc is an adapter to allow using regular functions as EventHandler
^
internal\events\nats_bus.go:162:1: Comment should end in a period (godot)
// Handle implements EventHandler interface
^
internal\events\nats_bus.go:167:1: Comment should end in a period (godot)
// TaskEventHandler handles task-related events
^
internal\events\nats_bus.go:172:1: Comment should end in a period (godot)
// NewTaskEventHandler creates a new task event handler
^
internal\events\nats_bus.go:177:1: Comment should end in a period (godot)
// Handle handles task events
^
internal\lifecycle\deployment.go:13:1: Comment should end in a period (godot)
// DeploymentStrategy represents deployment strategies
^
internal\lifecycle\deployment.go:23:1: Comment should end in a period (godot)
// DeploymentPhase represents deployment phases
^
internal\lifecycle\deployment.go:36:1: Comment should end in a period (godot)
// DeploymentConfig configures deployment automation
^
internal\lifecycle\deployment.go:85:1: Comment should end in a period (godot)
// DeploymentHook represents a deployment hook
^
internal\lifecycle\deployment.go:97:1: Comment should end in a period (godot)
// RollbackThresholds defines when to trigger auto-rollback
^
internal\lifecycle\deployment.go:105:1: Comment should end in a period (godot)
// DeploymentResult represents the result of a deployment
^
internal\lifecycle\deployment.go:122:1: Comment should end in a period (godot)
// DeploymentAutomation manages automated deployments
^
internal\lifecycle\deployment.go:133:1: Comment should end in a period (godot)
// NewDeploymentAutomation creates a new deployment automation system
^
internal\lifecycle\deployment.go:143:1: Comment should end in a period (godot)
// Deploy executes a deployment using the configured strategy
^
internal\lifecycle\deployment.go:194:1: Comment should end in a period (godot)
// Rollback rolls back to the previous version
^
internal\lifecycle\deployment.go:232:1: Comment should end in a period (godot)
// GetDeploymentHistory returns deployment history
^
internal\lifecycle\deployment.go:239:1: Comment should end in a period (godot)
// GetCurrentDeployment returns the current deployment status
^
internal\lifecycle\health.go:14:1: Comment should end in a period (godot)
// HealthStatus represents the health status of a component
^
internal\lifecycle\health.go:24:1: Comment should end in a period (godot)
// HealthCheck represents a health check result
^
internal\lifecycle\health.go:35:1: Comment should end in a period (godot)
// HealthReport represents the overall health status
^
internal\lifecycle\health.go:46:1: Comment should end in a period (godot)
// HealthSummary provides a summary of health checks
^
internal\lifecycle\health.go:55:1: Comment should end in a period (godot)
// DependencyStatus represents the status of an external dependency
^
internal\lifecycle\health.go:65:1: Comment should end in a period (godot)
// HealthChecker interface for health check implementations
^
internal\lifecycle\health.go:73:1: Comment should end in a period (godot)
// HealthMonitor provides comprehensive health monitoring
^
internal\lifecycle\health.go:94:1: Comment should end in a period (godot)
// HealthConfig configures health monitoring
^
internal\lifecycle\health.go:119:1: Comment should end in a period (godot)
// DependencyChecker checks external dependencies
^
internal\lifecycle\health.go:127:1: Comment should end in a period (godot)
// DefaultHealthConfig returns default health monitoring configuration
^
internal\lifecycle\health.go:146:1: Comment should end in a period (godot)
// NewHealthMonitor creates a new health monitor
^
internal\lifecycle\health.go:159:1: Comment should end in a period (godot)
// RegisterChecker registers a health checker
^
internal\lifecycle\health.go:172:1: Comment should end in a period (godot)
// RegisterDependency registers a dependency checker
^
internal\lifecycle\health.go:185:1: Comment should end in a period (godot)
// Start starts the health monitoring
^
internal\lifecycle\health.go:214:1: Comment should end in a period (godot)
// Stop stops the health monitoring
^
internal\lifecycle\health.go:234:1: Comment should end in a period (godot)
// GetHealth returns the current health status
^
internal\lifecycle\health.go:239:1: Comment should end in a period (godot)
// GetLastReport returns the last health report
^
internal\lifecycle\health.go:253:1: Comment should end in a period (godot)
// IsHealthy returns true if the system is healthy
^
internal\lifecycle\health.go:262:1: Comment should end in a period (godot)
// IsDegraded returns true if the system is degraded
^
internal\lifecycle\health.go:271:1: Comment should end in a period (godot)
// IsUnhealthy returns true if the system is unhealthy
^
internal\lifecycle\health.go:532:1: Comment should end in a period (godot)
// DatabaseHealthChecker checks database connectivity
^
internal\lifecycle\health.go:580:1: Comment should end in a period (godot)
// RedisHealthChecker checks Redis connectivity
^
internal\lifecycle\manager.go:14:1: Comment should end in a period (godot)
// LifecycleState represents the current state of the application
^
internal\lifecycle\manager.go:51:1: Comment should end in a period (godot)
// Component represents a lifecycle-managed component
^
internal\lifecycle\manager.go:62:1: Comment should end in a period (godot)
// LifecycleEvent represents events during lifecycle transitions
^
internal\lifecycle\manager.go:73:1: Comment should end in a period (godot)
// LifecycleManager manages application lifecycle and component orchestration
^
internal\lifecycle\manager.go:107:1: Comment should end in a period (godot)
// ComponentState tracks individual component state
^
internal\lifecycle\manager.go:117:1: Comment should end in a period (godot)
// Config configures the lifecycle manager
^
internal\lifecycle\manager.go:130:1: Comment should end in a period (godot)
// DefaultConfig returns default lifecycle manager configuration
^
internal\lifecycle\manager.go:145:1: Comment should end in a period (godot)
// NewLifecycleManager creates a new lifecycle manager
^
internal\lifecycle\manager.go:179:1: Comment should end in a period (godot)
// RegisterComponent registers a component for lifecycle management
^
internal\lifecycle\manager.go:204:1: Comment should end in a period (godot)
// RegisterEventHandler registers an event handler for lifecycle events
^
internal\lifecycle\manager.go:212:1: Comment should end in a period (godot)
// Start starts all registered components in priority order
^
internal\lifecycle\manager.go:259:1: Comment should end in a period (godot)
// Stop stops all components in reverse priority order
^
internal\lifecycle\manager.go:304:1: Comment should end in a period (godot)
// GetState returns the current lifecycle state
^
internal\lifecycle\manager.go:309:1: Comment should end in a period (godot)
// IsReady returns true if the application is ready to serve requests
^
internal\lifecycle\manager.go:315:1: Comment should end in a period (godot)
// IsHealthy returns true if the application is healthy
^
internal\lifecycle\manager.go:320:1: Comment should end in a period (godot)
// GetComponentStates returns the current state of all components
^
internal\lifecycle\manager.go:332:1: Comment should end in a period (godot)
// GetEventHistory returns recent lifecycle events
^
internal\lifecycle\manager.go:349:1: Comment should end in a period (godot)
// GetMetrics returns lifecycle metrics
^
internal\lifecycle\manager.go:386:1: Comment should end in a period (godot)
// LifecycleMetrics contains lifecycle metrics
^
internal\lifecycle\operations.go:12:1: Comment should end in a period (godot)
// OperationType represents different types of operations
^
internal\lifecycle\operations.go:27:1: Comment should end in a period (godot)
// OperationStatus represents the status of an operation
^
internal\lifecycle\operations.go:38:1: Comment should end in a period (godot)
// Operation represents a system operation
^
internal\lifecycle\operations.go:77:1: Comment should end in a period (godot)
// OperationStep represents a step within an operation
^
internal\lifecycle\operations.go:93:1: Comment should end in a period (godot)
// OperationExecutor defines the interface for operation execution
^
internal\lifecycle\operations.go:100:1: Comment should end in a period (godot)
// OperationsManager manages system operations and procedures
^
internal\lifecycle\operations.go:123:1: Comment should end in a period (godot)
// OperationsConfig configures operations management
^
internal\lifecycle\operations.go:135:1: Comment should end in a period (godot)
// DefaultOperationsConfig returns default operations configuration
^
internal\lifecycle\operations.go:149:1: Comment should end in a period (godot)
// NewOperationsManager creates a new operations manager
^
internal\lifecycle\operations.go:164:1: Comment should end in a period (godot)
// RegisterExecutor registers an operation executor
^
internal\lifecycle\operations.go:173:1: Comment should end in a period (godot)
// Start starts the operations manager
^
internal\lifecycle\operations.go:197:1: Comment should end in a period (godot)
// Stop stops the operations manager
^
internal\lifecycle\operations.go:220:1: Comment should end in a period (godot)
// CreateOperation creates a new operation
^
internal\lifecycle\operations.go:282:1: Comment should end in a period (godot)
// ExecuteOperation executes an operation asynchronously
^
internal\lifecycle\operations.go:306:1: Comment should end in a period (godot)
// CancelOperation cancels a running operation
^
internal\lifecycle\operations.go:340:1: Comment should end in a period (godot)
// GetOperation returns an operation by ID
^
internal\lifecycle\operations.go:355:1: Comment should end in a period (godot)
// ListOperations returns all operations with optional filtering
^
internal\lifecycle\operations.go:372:1: Comment should end in a period (godot)
// GetOperationHistory returns operation history
^
internal\lifecycle\operations.go:389:1: Comment should end in a period (godot)
// OperationFilter for filtering operations
^
internal\lifecycle\operations.go:398:1: Comment should end in a period (godot)
// Matches checks if an operation matches the filter
^
internal\lifecycle\operations.go:584:1: Comment should end in a period (godot)
// MaintenanceExecutor handles maintenance operations
^
internal\metrics\business.go:13:1: Comment should end in a period (godot)
// MetricType represents different types of business metrics
^
internal\metrics\business.go:23:1: Comment should end in a period (godot)
// AggregationType represents how metrics should be aggregated
^
internal\metrics\business.go:36:1: Comment should end in a period (godot)
// BusinessMetric defines a business metric configuration
^
internal\metrics\business.go:52:1: Comment should end in a period (godot)
// BusinessMetricsConfig configures business metrics collection
^
internal\metrics\business.go:75:1: Comment should end in a period (godot)
// MetricAlertRule defines alerting rules for business metrics
^
internal\metrics\business.go:87:1: Comment should end in a period (godot)
// MetricValue represents a metric measurement
^
internal\metrics\business.go:96:1: Comment should end in a period (godot)
// AggregatedMetric represents an aggregated metric value
^
internal\metrics\business.go:104:1: Comment should end in a period (godot)
// BusinessMetricsCollector collects and manages business metrics
^
internal\metrics\business.go:126:1: Comment should end in a period (godot)
// AlertState tracks the state of metric alerts
^
internal\metrics\business.go:138:1: Comment should end in a period (godot)
// MetricStorage interface for metric storage backends
^
internal\metrics\business.go:147:1: Comment should end in a period (godot)
// MetricQuery defines a metric query
^
internal\metrics\business.go:156:1: Comment should end in a period (godot)
// AggregationQuery defines an aggregation query
^
internal\metrics\business.go:164:1: Comment should end in a period (godot)
// DefaultBusinessMetricsConfig returns default configuration
^
internal\metrics\business.go:185:1: Comment should end in a period (godot)
// DefaultBusinessMetrics returns default business metrics
^
internal\metrics\business.go:312:1: Comment should end in a period (godot)
// DefaultAlertRules returns default alert rules
^
internal\metrics\business.go:363:1: Comment should end in a period (godot)
// NewBusinessMetricsCollector creates a new business metrics collector
^
internal\metrics\business.go:411:1: Comment should end in a period (godot)
// RecordCounter records a counter metric
^
internal\metrics\business.go:416:1: Comment should end in a period (godot)
// RecordGauge records a gauge metric
^
internal\metrics\business.go:421:1: Comment should end in a period (godot)
// RecordHistogram records a histogram metric
^
internal\metrics\business.go:426:1: Comment should end in a period (godot)
// RecordSummary records a summary metric
^
internal\metrics\business.go:431:1: Comment should end in a period (godot)
// recordMetric is the internal method to record any metric
^
internal\metrics\business.go:479:1: Comment should end in a period (godot)
// GetMetricValues returns raw metric values
^
internal\metrics\business.go:504:1: Comment should end in a period (godot)
// GetAggregatedMetrics returns aggregated metrics
^
internal\metrics\business.go:529:1: Comment should end in a period (godot)
// GetAlertStates returns current alert states
^
internal\metrics\business.go:542:1: Comment should end in a period (godot)
// GetMetrics returns all configured metrics
^
internal\metrics\business.go:555:1: Comment should end in a period (godot)
// Close gracefully shuts down the collector
^
internal\metrics\business.go:890:1: Comment should end in a period (godot)
// NewMetricStorage creates a new metric storage backend
^
internal\metrics\storage.go:10:1: Comment should end in a period (godot)
// MemoryMetricStorage provides in-memory metric storage
^
internal\metrics\storage.go:16:1: Comment should end in a period (godot)
// NewMemoryMetricStorage creates a new in-memory metric storage
^
internal\metrics\storage.go:23:1: Comment should end in a period (godot)
// Store stores metric values
^
internal\metrics\storage.go:39:1: Comment should end in a period (godot)
// Query queries metric values
^
internal\metrics\storage.go:69:1: Comment should end in a period (godot)
// Aggregate performs aggregations on metric values
^
internal\metrics\storage.go:112:1: Comment should end in a period (godot)
// Delete removes old metric values
^
internal\metrics\storage.go:130:1: Comment should end in a period (godot)
// Close closes the storage (no-op for memory storage)
^
internal\nats\publisher_error_handler.go:12:1: Comment should end in a period (godot)
// Publisher publishes messages to NATS with retry and error handling
^
internal\nats\publisher_error_handler.go:20:1: Comment should end in a period (godot)
// NewPublisher creates a new NATS publisher with error handling
^
internal\nats\publisher_error_handler.go:30:1: Comment should end in a period (godot)
// PublishWithRetry publishes a message with retry logic and error reporting
^
internal\nats\publisher_error_handler.go:64:1: Comment should end in a period (godot)
// sanitizeErr prevents leaking credentials in logs
^
internal\ratelimit\distributed.go:16:1: Comment should end in a period (godot)
// Algorithm represents different rate limiting algorithms
^
internal\ratelimit\distributed.go:28:1: Comment should end in a period (godot)
// DistributedRateLimiter provides distributed rate limiting capabilities
^
internal\ratelimit\distributed.go:46:1: Comment should end in a period (godot)
// Config configures the distributed rate limiter
^
internal\ratelimit\distributed.go:77:1: Comment should end in a period (godot)
// Rule defines a rate limiting rule
^
internal\ratelimit\distributed.go:109:1: Comment should end in a period (godot)
// Condition represents a condition for rule application
^
internal\ratelimit\distributed.go:117:1: Comment should end in a period (godot)
// Request represents a rate limiting request
^
internal\ratelimit\distributed.go:129:1: Comment should end in a period (godot)
// Response represents a rate limiting response
^
internal\ratelimit\distributed.go:149:1: Comment should end in a period (godot)
// Limiter interface for different rate limiting algorithms
^
internal\ratelimit\distributed.go:156:1: Comment should end in a period (godot)
// TokenBucketLimiter implements token bucket algorithm
^
internal\ratelimit\distributed.go:162:1: Comment should end in a period (godot)
// SlidingWindowLimiter implements sliding window algorithm
^
internal\ratelimit\distributed.go:168:1: Comment should end in a period (godot)
// AdaptiveLimiter implements adaptive rate limiting
^
internal\ratelimit\distributed.go:178:1: Comment should end in a period (godot)
// AdaptiveState tracks adaptive rate limiting state
^
internal\ratelimit\distributed.go:190:1: Comment should end in a period (godot)
// LuaScripts contains Lua scripts for atomic operations
^
internal\ratelimit\distributed.go:199:1: Comment should end in a period (godot)
// DefaultConfig returns default rate limiter configuration
^
internal\ratelimit\distributed.go:221:1: Comment should end in a period (godot)
// NewDistributedRateLimiter creates a new distributed rate limiter
^
internal\ratelimit\distributed.go:275:1: Comment should end in a period (godot)
// Allow checks if a request should be allowed
^
internal\ratelimit\distributed.go:312:1: Comment should end in a period (godot)
// AllowWithRule checks if a request should be allowed using a specific rule
^
internal\ratelimit\distributed.go:379:1: Comment should end in a period (godot)
// Reset resets the rate limit for a key
^
internal\ratelimit\distributed.go:390:1: Comment should end in a period (godot)
// GetUsage returns current usage for a key
^
internal\ratelimit\distributed.go:400:1: Comment should end in a period (godot)
// GetStats returns rate limiting statistics
^
internal\ratelimit\distributed.go:413:1: Comment should end in a period (godot)
// Close gracefully shuts down the rate limiter
^
internal\ratelimit\distributed.go:423:1: Comment should end in a period (godot)
// Stats contains rate limiting statistics
^
internal\repository\postgres\connection.go:11:1: Comment should end in a period (godot)
// Connect creates a PostgreSQL database connection
^
internal\repository\postgres\task_repository.go:15:1: Comment should end in a period (godot)
// TaskRepository implements domain.TaskRepository using PostgreSQL
^
internal\repository\postgres\task_repository.go:20:1: Comment should end in a period (godot)
// NewTaskRepository creates a new PostgreSQL task repository
^
internal\repository\postgres\task_repository.go:25:1: Comment should end in a period (godot)
// Create inserts a new task
^
internal\repository\postgres\task_repository.go:48:1: Comment should end in a period (godot)
// GetByID retrieves a task by ID
^
internal\repository\postgres\task_repository.go:60:1: Comment should end in a period (godot)
// Update updates an existing task
^
internal\repository\postgres\task_repository.go:91:1: Comment should end in a period (godot)
// Delete removes a task
^
internal\repository\postgres\task_repository.go:108:1: Comment should end in a period (godot)
// List retrieves tasks with filtering and pagination
^
internal\repository\postgres\task_repository.go:208:1: Comment should end in a period (godot)
// GetByStatus retrieves tasks by status
^
internal\repository\postgres\task_repository.go:235:1: Comment should end in a period (godot)
// GetByAssignee retrieves tasks assigned to a specific user
^
internal\repository\postgres\task_repository.go:262:1: Comment should end in a period (godot)
// scanTask scans a database row into a Task struct
^
internal\repository\redis\cache_repository.go:12:1: Comment should end in a period (godot)
// CacheRepository implements domain.CacheRepository using Redis
^
internal\repository\redis\cache_repository.go:17:1: Comment should end in a period (godot)
// NewCacheRepository creates a new Redis cache repository
^
internal\repository\redis\cache_repository.go:22:1: Comment should end in a period (godot)
// Set stores a value in cache with TTL
^
internal\repository\redis\cache_repository.go:42:1: Comment should end in a period (godot)
// Get retrieves a value from cache
^
internal\repository\redis\cache_repository.go:55:1: Comment should end in a period (godot)
// Delete removes a key from cache
^
internal\repository\redis\cache_repository.go:65:1: Comment should end in a period (godot)
// Exists checks if a key exists in cache
^
internal\repository\redis\cache_repository.go:75:1: Comment should end in a period (godot)
// Increment increments a counter
^
internal\repository\redis\cache_repository.go:85:1: Comment should end in a period (godot)
// SetNX sets a value only if the key doesn't exist (atomic operation)
^
internal\repository\redis\cache_repository.go:105:1: Comment should end in a period (godot)
// GetJSON retrieves and unmarshals a JSON value from cache
^
internal\repository\redis\cache_repository.go:120:1: Comment should end in a period (godot)
// SetWithExpiry sets a value with a specific expiry time
^
internal\repository\redis\cache_repository.go:135:1: Comment should end in a period (godot)
// GetTTL returns the remaining time-to-live of a key
^
internal\repository\redis\cache_repository.go:145:1: Comment should end in a period (godot)
// FlushAll removes all keys (use with caution)
^
internal\repository\redis\connection.go:11:1: Comment should end in a period (godot)
// NewClient creates a new Redis client
^
internal\repository\redis\connection.go:23:1: Comment should end in a period (godot)
// Ping tests Redis connection
^
internal\slo\alerting.go:16:1: Comment should end in a period (godot)
// AlertSeverity represents different alert severity levels
^
internal\slo\alerting.go:25:1: Comment should end in a period (godot)
// AlertChannel represents different alerting channels
^
internal\slo\alerting.go:37:1: Comment should end in a period (godot)
// AlertingConfig holds configuration for the alerting system
^
internal\slo\alerting.go:48:1: Comment should end in a period (godot)
// ChannelConfig holds configuration for specific alert channels
^
internal\slo\alerting.go:59:1: Comment should end in a period (godot)
// TemplateConfig holds message templates for different channels
^
internal\slo\alerting.go:68:1: Comment should end in a period (godot)
// RateLimitConfig configures rate limiting for alerts
^
internal\slo\alerting.go:76:1: Comment should end in a period (godot)
// EscalationPolicy defines how alerts should be escalated
^
internal\slo\alerting.go:84:1: Comment should end in a period (godot)
// EscalationStep defines a single step in an escalation policy
^
internal\slo\alerting.go:91:1: Comment should end in a period (godot)
// SilenceRule defines when alerts should be silenced
^
internal\slo\alerting.go:101:1: Comment should end in a period (godot)
// AlertManager manages SLO-based alerting
^
internal\slo\alerting.go:118:1: Comment should end in a period (godot)
// NewAlertManager creates a new alert manager
^
internal\slo\alerting.go:132:1: Comment should end in a period (godot)
// Start begins the alert processing
^
internal\slo\alerting.go:150:1: Comment should end in a period (godot)
// Stop stops the alert manager
^
internal\slo\alerting.go:155:1: Comment should end in a period (godot)
// SendAlert queues an alert for processing
^
internal\slo\alerting.go:166:1: Comment should end in a period (godot)
// processAlerts processes incoming alerts
^
internal\slo\alerting.go:187:1: Comment should end in a period (godot)
// processAlert processes a single alert
^
internal\slo\alerting.go:229:1: Comment should end in a period (godot)
// shouldSilence checks if an alert should be silenced
^
internal\slo\alerting.go:302:1: Comment should end in a period (godot)
// isRateLimited checks if an alert is rate limited
^
internal\slo\alerting.go:335:1: Comment should end in a period (godot)
// storeAlertHistory stores alert in history
^
internal\slo\alerting.go:349:1: Comment should end in a period (godot)
// getChannelsForSeverity returns channels for a given severity
^
internal\slo\alerting.go:357:1: Comment should end in a period (godot)
// sendToChannel sends an alert to a specific channel
^
internal\slo\alerting.go:383:1: Comment should end in a period (godot)
// sendToSlack sends alert to Slack
^
internal\slo\alerting.go:426:1: Comment should end in a period (godot)
// sendToDiscord sends alert to Discord
^
internal\slo\alerting.go:464:1: Comment should end in a period (godot)
// sendToWebhook sends alert to a generic webhook
^
internal\slo\alerting.go:479:1: Comment should end in a period (godot)
// sendToEmail sends alert via email (placeholder implementation)
^
internal\slo\alerting.go:487:1: Comment should end in a period (godot)
// sendToPagerDuty sends alert to PagerDuty (placeholder implementation)
^
internal\slo\alerting.go:495:1: Comment should end in a period (godot)
// sendToMSTeams sends alert to Microsoft Teams (placeholder implementation)
^
internal\slo\alerting.go:503:1: Comment should end in a period (godot)
// sendHTTPPayload sends a JSON payload via HTTP POST
^
internal\slo\alerting.go:534:1: Comment should end in a period (godot)
// startEscalation starts escalation process for an alert
^
internal\slo\alerting.go:559:1: Comment should end in a period (godot)
// executeEscalation executes an escalation policy
^
internal\slo\alerting.go:587:1: Comment should end in a period (godot)
// cleanup performs periodic cleanup of old data
^
internal\slo\alerting.go:604:1: Comment should end in a period (godot)
// performCleanup cleans up old rate limiter and history data
^
internal\slo\alerting.go:675:1: Comment should end in a period (godot)
// GetAlertHistory returns alert history for an SLO
^
internal\slo\alerting.go:683:1: Comment should end in a period (godot)
// GetAllAlertHistory returns all alert history
^
internal\slo\config.go:7:1: Comment should end in a period (godot)
// DefaultSLOs returns the default SLO configuration for MCP Ultra
^
internal\slo\config.go:387:1: Comment should end in a period (godot)
// GetSLOsByService returns SLOs filtered by service name
^
internal\slo\config.go:398:1: Comment should end in a period (godot)
// GetSLOsByComponent returns SLOs filtered by component name
^
internal\slo\config.go:409:1: Comment should end in a period (godot)
// GetSLOsByType returns SLOs filtered by type
^
internal\slo\config.go:420:1: Comment should end in a period (godot)
// GetCriticalSLOs returns SLOs marked as critical
^
internal\slo\monitor.go:15:1: Comment should end in a period (godot)
// SLOType represents the type of SLO being monitored
^
internal\slo\monitor.go:26:1: Comment should end in a period (godot)
// SLOStatus represents the current status of an SLO
^
internal\slo\monitor.go:36:1: Comment should end in a period (godot)
// SLO represents a Service Level Objective
^
internal\slo\monitor.go:69:1: Comment should end in a period (godot)
// SLOResult represents the result of an SLO evaluation
^
internal\slo\monitor.go:81:1: Comment should end in a period (godot)
// ErrorBudget represents the error budget information
^
internal\slo\monitor.go:90:1: Comment should end in a period (godot)
// BurnRate represents burn rate information
^
internal\slo\monitor.go:100:1: Comment should end in a period (godot)
// CompliancePoint represents a point in time compliance measurement
^
internal\slo\monitor.go:107:1: Comment should end in a period (godot)
// AlertRule represents an alerting rule for an SLO
^
internal\slo\monitor.go:118:1: Comment should end in a period (godot)
// Monitor manages SLO monitoring and evaluation
^
internal\slo\monitor.go:136:1: Comment should end in a period (godot)
// AlertEvent represents an SLO alert event
^
internal\slo\monitor.go:147:1: Comment should end in a period (godot)
// StatusEvent represents an SLO status change event
^
internal\slo\monitor.go:156:1: Comment should end in a period (godot)
// NewMonitor creates a new SLO monitor
^
internal\slo\monitor.go:173:1: Comment should end in a period (godot)
// AddSLO adds an SLO to the monitor
^
internal\slo\monitor.go:210:1: Comment should end in a period (godot)
// RemoveSLO removes an SLO from monitoring
^
internal\slo\monitor.go:220:1: Comment should end in a period (godot)
// GetSLO retrieves an SLO by name
^
internal\slo\monitor.go:229:1: Comment should end in a period (godot)
// GetAllSLOs returns all configured SLOs
^
internal\slo\monitor.go:241:1: Comment should end in a period (godot)
// GetSLOResult retrieves the latest SLO evaluation result
^
internal\slo\monitor.go:250:1: Comment should end in a period (godot)
// GetAllSLOResults returns all SLO evaluation results
^
internal\slo\monitor.go:262:1: Comment should end in a period (godot)
// Start begins SLO monitoring
^
internal\slo\monitor.go:283:1: Comment should end in a period (godot)
// Stop stops SLO monitoring
^
internal\slo\monitor.go:288:1: Comment should end in a period (godot)
// AlertChannel returns the alert event channel
^
internal\slo\monitor.go:293:1: Comment should end in a period (godot)
// StatusChannel returns the status change event channel
^
internal\slo\monitor.go:298:1: Comment should end in a period (godot)
// evaluateAllSLOs evaluates all configured SLOs
^
internal\slo\monitor.go:318:1: Comment should end in a period (godot)
// evaluateSLO evaluates a single SLO
^
internal\slo\monitor.go:371:1: Comment should end in a period (godot)
// queryPrometheus executes a Prometheus query
^
internal\slo\monitor.go:403:1: Comment should end in a period (godot)
// calculateErrorBudget calculates the error budget for an SLO
^
internal\slo\monitor.go:445:1: Comment should end in a period (godot)
// calculateBurnRate calculates the burn rate for an SLO
^
internal\slo\monitor.go:489:1: Comment should end in a period (godot)
// determineStatus determines the SLO status based on current metrics
^
internal\slo\monitor.go:509:1: Comment should end in a period (godot)
// getComplianceHistory retrieves historical compliance data
^
internal\slo\monitor.go:558:1: Comment should end in a period (godot)
// storeResult stores an SLO evaluation result and checks for status changes
^
internal\slo\monitor.go:586:1: Comment should end in a period (godot)
// checkAndGenerateAlerts checks if alerts should be generated for an SLO result
^
internal\testhelpers\helpers.go:8:1: Comment should end in a period (godot)
// GetTestJWTSecret returns a safe test JWT secret
^
internal\testhelpers\helpers.go:13:1: Comment should end in a period (godot)
// GenerateTestSecret generates a random test secret
^
internal\testhelpers\helpers.go:22:1: Comment should end in a period (godot)
// GetTestDatabaseURL returns a test database URL
^
internal\testhelpers\helpers.go:27:1: Comment should end in a period (godot)
// GetTestRedisURL returns a test Redis URL
^
internal\testhelpers\helpers.go:32:1: Comment should end in a period (godot)
// GetTestNATSURL returns a test NATS URL
^
internal\tracing\business.go:20:1: Comment should end in a period (godot)
// BusinessTransactionTracer provides advanced tracing for critical business transactions
^
internal\tracing\business.go:39:1: Comment should end in a period (godot)
// TracingConfig configures business transaction tracing
^
internal\tracing\business.go:76:1: Comment should end in a period (godot)
// AlertThresholds defines alerting thresholds
^
internal\tracing\business.go:85:1: Comment should end in a period (godot)
// BusinessTransaction represents a high-level business transaction
^
internal\tracing\business.go:129:1: Comment should end in a period (godot)
// TransactionType represents different types of business transactions
^
internal\tracing\business.go:145:1: Comment should end in a period (godot)
// TransactionStatus represents transaction status
^
internal\tracing\business.go:157:1: Comment should end in a period (godot)
// TransactionStep represents a step within a business transaction
^
internal\tracing\business.go:173:1: Comment should end in a period (godot)
// TransactionEvent represents an event within a transaction
^
internal\tracing\business.go:183:1: Comment should end in a period (godot)
// TransactionError represents an error within a transaction
^
internal\tracing\business.go:195:1: Comment should end in a period (godot)
// TransactionMetrics contains transaction performance metrics
^
internal\tracing\business.go:208:1: Comment should end in a period (godot)
// TransactionTemplate defines a template for transaction creation
^
internal\tracing\business.go:227:1: Comment should end in a period (godot)
// EventLevel represents the severity level of an event
^
internal\tracing\business.go:238:1: Comment should end in a period (godot)
// DefaultTracingConfig returns default tracing configuration
^
internal\tracing\business.go:270:1: Comment should end in a period (godot)
// NewBusinessTransactionTracer creates a new business transaction tracer
^
internal\tracing\business.go:306:1: Comment should end in a period (godot)
// StartTransaction starts a new business transaction
^
internal\tracing\business.go:389:1: Comment should end in a period (godot)
// EndTransaction ends a business transaction
^
internal\tracing\business.go:455:1: Comment should end in a period (godot)
// StartStep starts a new step within a transaction
^
internal\tracing\business.go:489:1: Comment should end in a period (godot)
// EndStep ends a transaction step
^
internal\tracing\business.go:533:1: Comment should end in a period (godot)
// AddEvent adds an event to a transaction
^
internal\tracing\business.go:568:1: Comment should end in a period (godot)
// AddError adds an error to a transaction
^
internal\tracing\business.go:573:1: Comment should end in a period (godot)
// GetTransaction retrieves a transaction by ID
^
internal\tracing\business.go:588:1: Comment should end in a period (godot)
// ListActiveTransactions returns all currently active transactions
^
internal\tracing\business.go:604:1: Comment should end in a period (godot)
// GetTransactionMetrics returns aggregated metrics for transactions
^
internal\tracing\business.go:650:1: Comment should end in a period (godot)
// RegisterTemplate registers a transaction template
^
internal\tracing\business.go:664:1: Comment should end in a period (godot)
// Close gracefully shuts down the tracer
^
internal\tracing\business.go:674:1: Comment should end in a period (godot)
// TransactionAnalytics contains transaction analytics
^
scripts\generate-secrets.go:15:1: Comment should end in a period (godot)
// generateRandomHex creates a cryptographically secure random hex string
^
basic_test.go:7:1: Comment should end in a period (godot)
// TestBasic is a basic test to ensure the test runner works
^
basic_test.go:14:1: Comment should end in a period (godot)
// TestVersion tests that version constants are not empty
^
internal\ai\events\handlers_test.go:9:1: Comment should end in a period (godot)
// Mock publisher for testing
^
internal\config\config.go:14:1: Comment should end in a period (godot)
// Config represents the application configuration
^
internal\config\config.go:29:1: Comment should end in a period (godot)
// ComplianceConfig holds all compliance-related configuration
^
internal\config\config.go:43:1: Comment should end in a period (godot)
// PIIDetectionConfig configures PII detection and classification
^
internal\config\config.go:52:1: Comment should end in a period (godot)
// ConsentConfig configures consent management
^
internal\config\config.go:60:1: Comment should end in a period (godot)
// DataRetentionConfig configures data retention policies
^
internal\config\config.go:69:1: Comment should end in a period (godot)
// AuditLoggingConfig configures compliance audit logging
^
internal\config\config.go:79:1: Comment should end in a period (godot)
// LGPDConfig specific configuration for Brazilian LGPD compliance
^
internal\config\config.go:88:1: Comment should end in a period (godot)
// GDPRConfig specific configuration for European GDPR compliance
^
internal\config\config.go:98:1: Comment should end in a period (godot)
// AnonymizationConfig configures data anonymization
^
internal\config\config.go:108:1: Comment should end in a period (godot)
// DataRightsConfig configures individual data rights handling
^
internal\config\config.go:117:1: Comment should end in a period (godot)
// ServerConfig holds HTTP server configuration
^
internal\config\config.go:125:1: Comment should end in a period (godot)
// GRPCConfig holds gRPC server configuration
^
internal\config\config.go:135:1: Comment should end in a period (godot)
// KeepaliveConfig holds gRPC keepalive configuration
^
internal\config\config.go:146:1: Comment should end in a period (godot)
// DatabaseConfig holds database connections configuration
^
internal\config\config.go:152:1: Comment should end in a period (godot)
// PostgreSQLConfig holds PostgreSQL configuration
^
internal\config\config.go:165:1: Comment should end in a period (godot)
// RedisConfig holds Redis configuration
^
internal\config\config.go:173:1: Comment should end in a period (godot)
// NATSConfig holds NATS configuration
^
internal\config\config.go:180:1: Comment should end in a period (godot)
// TelemetryConfig holds comprehensive telemetry configuration
^
internal\config\config.go:198:1: Comment should end in a period (godot)
// TracingConfig holds distributed tracing configuration
^
internal\config\config.go:207:1: Comment should end in a period (godot)
// MetricsConfig holds metrics collection configuration
^
internal\config\config.go:216:1: Comment should end in a period (godot)
// ExportersConfig holds exporter configurations
^
internal\config\config.go:228:1: Comment should end in a period (godot)
// JaegerConfig holds Jaeger exporter configuration
^
internal\config\config.go:236:1: Comment should end in a period (godot)
// OTLPConfig holds OTLP exporter configuration
^
internal\config\config.go:244:1: Comment should end in a period (godot)
// ConsoleConfig holds console exporter configuration
^
internal\config\config.go:249:1: Comment should end in a period (godot)
// FeaturesConfig holds feature flags configuration
^
internal\config\config.go:255:1: Comment should end in a period (godot)
// SecurityConfig holds all security-related configuration
^
internal\config\config.go:263:1: Comment should end in a period (godot)
// Load loads configuration from file and environment variables
^
internal\config\config.go:283:1: Comment should end in a period (godot)
// loadFromFile loads configuration from YAML file
^
internal\config\config.go:295:1: Comment should end in a period (godot)
// getEnv returns environment variable value or default
^
internal\config\config.go:303:1: Comment should end in a period (godot)
// DSN returns PostgreSQL connection string
^
internal\config\tls.go:355:1: Comment should end in a period (godot)
// GetTLSConfig returns the current TLS configuration
^
internal\config\tls.go:366:1: Comment should end in a period (godot)
// IsEnabled returns whether TLS is enabled
^
internal\config\tls.go:371:1: Comment should end in a period (godot)
// Stop stops the certificate watcher
^
internal\config\tls.go:380:1: Comment should end in a period (godot)
// ValidateConfig validates the TLS configuration
^
internal\config\tls_test.go:351:1: Comment should end in a period (godot)
// Helper function to create temporary files for testing
^
internal\config\tls_test.go:370:1: Comment should end in a period (godot)
// Test certificate and key (for testing purposes only)
^
internal\domain\models.go:9:1: Comment should end in a period (godot)
// Task represents a task in the system
^
internal\domain\models.go:26:1: Comment should end in a period (godot)
// TaskStatus represents the status of a task
^
internal\domain\models.go:36:1: Comment should end in a period (godot)
// Priority represents task priority
^
internal\domain\models.go:46:1: Comment should end in a period (godot)
// User represents a user in the system
^
internal\domain\models.go:57:1: Comment should end in a period (godot)
// Role represents user role
^
internal\domain\models.go:65:1: Comment should end in a period (godot)
// Event represents a domain event
^
internal\domain\models.go:75:1: Comment should end in a period (godot)
// FeatureFlag represents a feature flag
^
internal\domain\models.go:87:1: Comment should end in a period (godot)
// TaskFilter represents filters for task queries
^
internal\domain\models.go:100:1: Comment should end in a period (godot)
// NewTask creates a new task with default values
^
internal\domain\models.go:116:1: Comment should end in a period (godot)
// Complete marks a task as completed
^
internal\domain\models.go:124:1: Comment should end in a period (godot)
// Cancel marks a task as cancelled
^
internal\domain\models.go:130:1: Comment should end in a period (godot)
// UpdateStatus updates task status
^
internal\domain\models.go:136:1: Comment should end in a period (godot)
// IsValidStatus checks if status transition is valid
^
internal\domain\repository.go:9:1: Comment should end in a period (godot)
// TaskRepository defines the interface for task data access
^
internal\domain\repository.go:20:1: Comment should end in a period (godot)
// UserRepository defines the interface for user data access
^
internal\domain\repository.go:30:1: Comment should end in a period (godot)
// EventRepository defines the interface for event data access
^
internal\domain\repository.go:37:1: Comment should end in a period (godot)
// FeatureFlagRepository defines the interface for feature flag data access
^
internal\domain\repository.go:46:1: Comment should end in a period (godot)
// CacheRepository defines the interface for cache operations
^
internal\telemetry\telemetry.go:24:2: Comment should end in a period (godot)
        // HTTP Metrics
        ^
internal\telemetry\telemetry.go:42:2: Comment should end in a period (godot)
        // Business Metrics
        ^
internal\telemetry\telemetry.go:60:2: Comment should end in a period (godot)
        // System Metrics
        ^
internal\telemetry\telemetry.go:78:1: Comment should end in a period (godot)
// Telemetry holds telemetry configuration and clients
^
internal\telemetry\telemetry.go:84:1: Comment should end in a period (godot)
// Init initializes telemetry system
^
internal\telemetry\telemetry.go:107:1: Comment should end in a period (godot)
// HTTPMetrics middleware for HTTP request metrics
^
internal\telemetry\telemetry.go:127:1: Comment should end in a period (godot)
// RecordTaskCreated records task creation metrics
^
internal\telemetry\telemetry.go:132:1: Comment should end in a period (godot)
// RecordTaskProcessingTime records task processing time
^
internal\telemetry\telemetry.go:137:1: Comment should end in a period (godot)
// RecordDatabaseConnections records database connection metrics
^
internal\telemetry\telemetry.go:142:1: Comment should end in a period (godot)
// RecordCacheOperation records cache operation metrics
^
internal\telemetry\telemetry.go:147:1: Comment should end in a period (godot)
// TaskMetrics handles task-related metrics
^
internal\telemetry\telemetry.go:155:1: Comment should end in a period (godot)
// NewTaskMetrics creates new task metrics
^
internal\telemetry\telemetry.go:190:1: Comment should end in a period (godot)
// RecordTaskCreated records a task creation
^
internal\telemetry\telemetry.go:200:1: Comment should end in a period (godot)
// RecordTaskCompleted records a task completion
^
internal\telemetry\telemetry.go:215:1: Comment should end in a period (godot)
// FeatureFlagMetrics handles feature flag metrics
^
internal\telemetry\telemetry.go:221:1: Comment should end in a period (godot)
// NewFeatureFlagMetrics creates new feature flag metrics
^
internal\telemetry\telemetry.go:237:1: Comment should end in a period (godot)
// RecordEvaluation records a feature flag evaluation
^
internal\telemetry\tracing.go:184:1: Comment should end in a period (godot)
// GetTracer returns a tracer for the given name
^
internal\telemetry\tracing.go:192:1: Comment should end in a period (godot)
// Shutdown gracefully shuts down the tracing provider
^
internal\telemetry\tracing.go:203:1: Comment should end in a period (godot)
// TraceFunction wraps a function with tracing
^
internal\telemetry\tracing.go:217:1: Comment should end in a period (godot)
// TraceFunctionWithResult wraps a function with tracing and returns a result
^
internal\telemetry\tracing.go:233:1: Comment should end in a period (godot)
// AddSpanAttributes adds multiple attributes to the current span
^
internal\telemetry\tracing.go:241:1: Comment should end in a period (godot)
// AddSpanEvent adds an event to the current span
^
internal\telemetry\tracing.go:249:1: Comment should end in a period (godot)
// SetSpanError sets error status on the current span
^
internal\telemetry\tracing.go:258:1: Comment should end in a period (godot)
// GetTraceID returns the trace ID from the current context
^
internal\telemetry\tracing.go:267:1: Comment should end in a period (godot)
// GetSpanID returns the span ID from the current context
^
internal\telemetry\tracing.go:276:1: Comment should end in a period (godot)
// InjectTraceContext injects trace context into a map (for cross-service calls)
^
internal\telemetry\tracing.go:281:1: Comment should end in a period (godot)
// ExtractTraceContext extracts trace context from a map
^
internal\telemetry\tracing.go:286:1: Comment should end in a period (godot)
// mapCarrier implements the TextMapCarrier interface
^
internal\telemetry\tracing.go:307:1: Comment should end in a period (godot)
// noopExporter is a no-op span exporter for disabled tracing
^
internal\telemetry\tracing.go:318:1: Comment should end in a period (godot)
// Span naming conventions
^
internal\telemetry\tracing.go:331:1: Comment should end in a period (godot)
// Common span attributes
^
internal\lifecycle\deployment.go:563:20: S1039: unnecessary use of fmt.Sprintf (gosimple)
        da.addLog(result, fmt.Sprintf("Script executed successfully"))
                          ^
internal\domain\models.go:33:37: `cancelled` is a misspelling of `canceled` (misspell)
        TaskStatusCancelled  TaskStatus = "cancelled"
                                           ^
internal\domain\models_test.go:83:31: `Cancelled` is a misspelling of `Canceled` (misspell)
                        name:          "Pending to Cancelled",
                                                   ^
internal\domain\models_test.go:101:34: `Cancelled` is a misspelling of `Canceled` (misspell)
                        name:          "InProgress to Cancelled",
                                                      ^
internal\domain\models_test.go:113:20: `Cancelled` is a misspelling of `Canceled` (misspell)
                        name:          "Cancelled to InProgress",
                                        ^
internal\domain\models_test.go:135:30: `cancelled` is a misspelling of `canceled` (misspell)
        assert.Equal(t, TaskStatus("cancelled"), TaskStatusCancelled)
                                    ^
automation\autocommit.go:84:1: `if os.IsNotExist(err)` has complex nested blocks (complexity: 6) (nestif)
        if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
^
internal\slo\monitor.go:407:1: `if slo.ErrorBudgetQuery == ""` has complex nested blocks (complexity: 5) (nestif)
        if slo.ErrorBudgetQuery == "" {
^
internal\slo\monitor.go:449:1: `if slo.BurnRateQuery != ""` has complex nested blocks (complexity: 4) (nestif)
        if slo.BurnRateQuery != "" {
^
internal\tracing\business.go:625:1: `if transaction.Duration > 0` has complex nested blocks (complexity: 6) (nestif)
                if transaction.Duration > 0 {
^
internal\tracing\business.go:778:1: `if exists` has complex nested blocks (complexity: 4) (nestif)
                if value, exists := attributes[field]; exists {
^
internal\slo\alerting.go:511:29: should rewrite http.NewRequestWithContext or add (*Request).WithContext (noctx)
        req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
                                   ^
automation\autocommit.go:7:2: SA1019: "io/ioutil" has been deprecated since Go 1.19: As of Go 1.16, the same functionality is now provided by package [io] or package [os], and those implementations should be preferred in new code. See the specific function documentation for details. (staticcheck)
        "io/ioutil"
        ^
internal\telemetry\tracing.go:187:10: SA1019: trace.NewNoopTracerProvider is deprecated: Use [go.opentelemetry.io/otel/trace/noop.NewTracerProvider] instead. (staticcheck)
                return trace.NewNoopTracerProvider().Tracer(name)
                       ^
automation\autocommit.go:56:22: `runCommand` - `command` always receives `"git"` (unparam)
func runCommand(dir, command string, args ...string) (string, error) {
                     ^
internal\config\secrets\loader.go:207:38: `(*SecretsLoader).loadFromK8s` - `ctx` is unused (unparam)
func (sl *SecretsLoader) loadFromK8s(ctx context.Context) (*SecretsConfig, error) {
                                     ^
internal\events\nats_bus.go:194:46: `(*TaskEventHandler).handleTaskCreated` - `ctx` is unused (unparam)
func (h *TaskEventHandler) handleTaskCreated(ctx context.Context, event *domain.Event) error {
                                             ^
internal\events\nats_bus.go:203:46: `(*TaskEventHandler).handleTaskUpdated` - `ctx` is unused (unparam)
func (h *TaskEventHandler) handleTaskUpdated(ctx context.Context, event *domain.Event) error {
                                             ^
internal\events\nats_bus.go:212:48: `(*TaskEventHandler).handleTaskCompleted` - `ctx` is unused (unparam)
func (h *TaskEventHandler) handleTaskCompleted(ctx context.Context, event *domain.Event) error {
                                               ^
internal\events\nats_bus.go:221:46: `(*TaskEventHandler).handleTaskDeleted` - `ctx` is unused (unparam)
func (h *TaskEventHandler) handleTaskDeleted(ctx context.Context, event *domain.Event) error {
                                             ^
internal\ratelimit\distributed.go:526:86: `(*DistributedRateLimiter).recordMetrics` - `key` is unused (unparam)
func (drl *DistributedRateLimiter) recordMetrics(status string, algorithm Algorithm, key string, remaining int64) {
                                                                                     ^
internal\tracing\business.go:735:83: `(*BusinessTransactionTracer).shouldSample` - `attributes` is unused (unparam)
func (btt *BusinessTransactionTracer) shouldSample(template *TransactionTemplate, attributes map[string]interface{}) bool {
                                                                                  ^
PS E:\vertikon\business\SaaS\templates\mcp-ultra-wasm> gofmt -s -l .
templates\ai\go\budgets\budgets.go:4:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:5:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:6:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:7:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:8:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:9:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:11:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:15:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:15:30: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:15:59: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:15:88: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:18:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:18:23: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:18:45: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:22:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:22:23: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:22:45: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:25:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:25:23: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:25:45: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:28:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:29:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:30:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:34:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:35:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:35:10: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:36:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:36:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:37:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:39:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:40:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:40:45: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:41:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:41:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:42:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:44:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:45:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:45:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:46:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:46:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:47:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:51:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:52:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:54:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:55:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:56:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:57:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:58:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:58:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:59:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:59:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:59:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:60:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:60:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:60:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:61:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:61:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:61:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:62:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:62:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:62:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:63:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:63:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:64:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:66:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:67:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:68:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:69:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:69:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:70:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:70:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:70:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:71:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:71:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:71:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:72:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:72:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:72:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:72:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:73:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:73:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:73:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:73:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:73:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:74:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:74:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:74:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:74:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:74:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:75:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:75:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:75:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:75:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:75:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:76:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:76:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:76:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:76:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:76:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:77:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:77:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:77:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:77:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:77:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:78:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:78:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:78:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:78:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:79:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:79:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:79:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:80:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:80:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:81:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:83:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:84:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:85:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:86:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:86:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:87:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:87:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:87:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:88:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:88:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:88:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:89:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:89:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:89:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:89:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:90:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:90:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:90:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:90:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:90:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:91:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:91:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:91:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:91:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:91:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:92:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:92:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:92:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:92:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:92:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:93:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:93:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:93:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:93:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:93:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:94:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:94:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:94:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:94:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:94:9: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:95:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:95:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:95:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:95:7: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:96:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:96:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:96:5: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:97:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:97:3: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:98:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:100:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:104:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:105:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:107:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:108:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:110:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:111:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:113:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:114:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:116:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:120:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:121:1: illegal character U+005C '\'
templates\ai\go\budgets\budgets.go:122:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:4:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:5:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:6:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:7:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:9:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:10:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:14:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:15:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:16:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:17:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:21:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:22:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:23:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:23:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:24:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:26:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:27:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:28:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:28:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:29:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:31:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:32:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:32:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:33:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:33:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:34:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:34:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:35:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:35:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:36:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:40:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:41:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:42:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:43:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:43:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:44:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:46:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:47:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:48:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:52:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:53:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:54:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:55:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:55:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:56:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:58:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:59:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:60:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:64:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:65:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:65:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:66:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:66:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:67:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:67:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:68:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:68:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:69:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:69:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:70:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:70:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:71:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:71:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:72:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:72:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:73:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:73:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:74:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:74:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:75:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:77:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:78:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:79:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:79:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:80:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:82:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:83:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:84:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:88:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:89:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:89:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:90:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:90:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:91:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:91:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:92:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:92:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:93:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:93:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:94:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:94:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:95:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:95:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:96:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:96:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:97:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:99:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:100:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:101:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:101:3: illegal character U+005C '\'
templates\ai\go\events\publisher.go:102:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:104:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:105:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:106:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:110:1: illegal character U+005C '\'
templates\ai\go\events\publisher.go:111:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:4:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:5:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:6:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:7:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:8:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:10:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:14:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:15:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:16:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:17:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:21:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:22:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:26:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:27:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:31:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:32:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:33:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:33:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:34:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:36:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:37:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:38:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:38:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:39:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:41:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:42:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:44:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:45:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:45:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:46:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:46:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:47:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:51:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:52:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:53:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:53:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:54:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:54:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:54:5: illegal character U+005C '\'
templates\ai\go\policies\policies.go:55:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:55:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:55:5: illegal character U+005C '\'
templates\ai\go\policies\policies.go:56:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:56:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:56:5: illegal character U+005C '\'
templates\ai\go\policies\policies.go:57:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:57:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:57:5: illegal character U+005C '\'
templates\ai\go\policies\policies.go:58:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:58:3: illegal character U+005C '\'
templates\ai\go\policies\policies.go:59:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:61:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:65:1: illegal character U+005C '\'
templates\ai\go\policies\policies.go:66:1: illegal character U+005C '\'
templates\ai\go\router\router.go:14:40: expected ';', found Default
templates\ai\go\router\router.go:16:1: expected '}', found 'type'
templates\ai\go\router\router.go:20:39: expected ';', found Use
templates\ai\go\router\router.go:22:1: expected '}', found 'type'
templates\ai\go\router\router.go:32:9: illegal character U+005C '\'
templates\ai\go\router\router.go:33:3: expected '{', found 'return'
templates\ai\go\router\router.go:37:44: illegal character U+005C '\'
templates\ai\go\router\router.go:89:46: illegal character U+005C '\'
templates\ai\go\router\router.go:92:55: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:4:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:5:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:9:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:10:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:10:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:11:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:11:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:11:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:12:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:12:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:12:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:13:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:13:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:14:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:14:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:15:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:17:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:18:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:18:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:19:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:19:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:19:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:20:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:20:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:20:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:21:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:21:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:21:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:22:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:22:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:23:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:23:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:24:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:26:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:27:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:27:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:28:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:28:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:28:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:29:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:29:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:29:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:30:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:30:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:31:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:31:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:32:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:34:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:35:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:35:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:36:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:36:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:36:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:37:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:37:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:37:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:38:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:38:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:39:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:39:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:40:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:42:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:43:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:43:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:44:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:44:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:44:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:45:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:45:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:45:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:46:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:46:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:47:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:47:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:48:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:50:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:51:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:51:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:52:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:52:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:52:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:53:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:53:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:53:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:54:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:54:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:55:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:55:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:56:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:58:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:59:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:59:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:60:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:60:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:60:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:61:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:61:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:61:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:62:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:62:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:63:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:63:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:64:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:66:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:67:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:67:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:68:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:68:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:68:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:69:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:69:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:69:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:70:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:70:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:71:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:71:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:72:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:74:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:75:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:75:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:76:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:76:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:76:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:77:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:77:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:77:5: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:78:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:78:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:79:1: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:79:3: illegal character U+005C '\'
templates\ai\go\telemetry\metrics.go:80:1: illegal character U+005C '\'
PS E:\vertikon\business\SaaS\templates\mcp-ultra-wasm>
