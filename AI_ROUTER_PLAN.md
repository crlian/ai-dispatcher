# Plan: AI CLI Cost-Aware Router

## Objetivo
Crear una herramienta CLI que orqueste múltiples AI coding assistants (Claude Code, Codex, OpenCode) optimizando costos mediante routing inteligente basado en:
- Complejidad de la tarea (analizada por LLM)
- Límites/cuotas disponibles en cada herramienta
- Costo estimado por herramienta

## Arquitectura

### Stack Tecnológico
- **Lenguaje:** Go (compilado, startup ~10ms)
- **Distribución:** npm wrapper (descarga binario precompilado)
- **Testing:** Go testing package
- **Build:** Go build + GitHub Actions para cross-compilation

### Estructura del Proyecto

**Proyecto principal (Go):**
```
ai-router/
├── main.go                       # Entry point CLI
├── cmd/
│   ├── exec.go                   # Comando exec
│   └── status.go                 # Comando status
├── pkg/
│   ├── analyzers/
│   │   └── complexity.go         # Complexity analyzer usando LLM
│   ├── trackers/
│   │   ├── tracker.go            # Interface base
│   │   ├── claude_code.go        # ccusage integration
│   │   ├── codex.go              # @ccusage/codex integration
│   │   └── opencode.go           # @ccusage/opencode integration
│   ├── router/
│   │   ├── calculator.go         # Cost calculator
│   │   └── engine.go             # Decision engine
│   └── delegators/
│       ├── delegator.go          # Interface base
│       ├── claude_code.go        # Ejecuta Claude Code
│       ├── codex.go              # Ejecuta Codex
│       └── opencode.go           # Ejecuta OpenCode
├── go.mod
├── go.sum
├── Makefile                      # Build scripts
└── README.md
```

**Wrapper npm (separado):**
```
ai-router-npm/
├── package.json                  # npm package config
├── install.js                    # Descarga binario en npm install
├── bin/
│   └── ai-router.js              # Wrapper que ejecuta binario
└── README.md
```

## Implementación Detallada

### 1. CLI Main (main.go)
```go
package main

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "ai-router",
        Short: "Cost-aware AI coding assistant router",
        Version: "1.0.0",
    }

    var execCmd = &cobra.Command{
        Use:   "exec [task]",
        Short: "Execute a coding task with optimal AI tool",
        Args:  cobra.MinimumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            task := args[0]
            force, _ := cmd.Flags().GetString("force")
            verbose, _ := cmd.Flags().GetBool("verbose")

            // 1. Analyze complexity
            // 2. Check available tools
            // 3. Calculate costs
            // 4. Route and delegate
        },
    }
    execCmd.Flags().StringP("force", "f", "", "Force specific tool")
    execCmd.Flags().BoolP("verbose", "v", false, "Show detailed routing")

    var statusCmd = &cobra.Command{
        Use:   "status",
        Short: "Show current limits for all tools",
        Run: func(cmd *cobra.Command, args []string) {
            // Show ccusage for all tools
        },
    }

    rootCmd.AddCommand(execCmd, statusCmd)
    rootCmd.Execute()
}
```

### 2. Complexity Analyzer (pkg/analyzers/complexity.go)
**Estrategia:** Usa el LLM más barato disponible para analizar la complejidad

```go
package analyzers

type ComplexityAnalysis struct {
    Level            string  // "simple" | "medium" | "complex"
    EstimatedTokens  int
    Reasoning        string
}

func AnalyzeComplexity(task string) (*ComplexityAnalysis, error) {
    // 1. Encuentra la herramienta con más límite disponible
    // 2. Usa esa herramienta con prompt mínimo para analizar
    // 3. Prompt: "Analyze task complexity (simple/medium/complex) and estimate tokens needed: {task}"
    // 4. Parse response y retorna ComplexityAnalysis
    return &ComplexityAnalysis{
        Level: "medium",
        EstimatedTokens: 1000,
        Reasoning: "Task requires moderate code changes",
    }, nil
}
```

### 3. Usage Trackers (pkg/trackers/)

**Base Interface:**
```go
package trackers

type UsageTracker interface {
    GetAvailablePercentage() (float64, error)
    GetRemainingTime() (int, error) // minutes
    GetTotalCost5hWindow() (float64, error)
    IsAvailable() (bool, error)
    GetToolName() string
    GetToolType() ToolType
}
```

#### 3.1 Métodos de Tracking Descubiertos

Después de investigación exhaustiva, estos son los métodos disponibles para cada herramienta:

##### **Claude Code** - ccusage (Local JSONL Parsing) ✅

**Método:** ccusage tool reading `~/.claude/history.jsonl`

**Comando:**
```bash
ccusage blocks --active --token-limit max --json
```

**Estructura JSON Output:**
```json
{
  "blocks": [
    {
      "isActive": true,
      "costUSD": 2.18,
      "projection": {
        "remainingMinutes": 218,
        "totalCost": 2.18
      },
      "tokenLimitStatus": {
        "limit": 100000,
        "percentUsed": 27.2,
        "status": "ok"
      }
    }
  ]
}
```

**Cálculo de Porcentaje:**
- **IMPORTANTE:** Claude Code UI muestra porcentaje basado en COSTO, no tokens
- Fórmula: `available% = ((costLimit - currentCost) / costLimit) * 100`
- Límites por plan (5-hour window):
  - Pro: $2.00
  - Max5: $4.00
  - Max20: $8.00 (default)

**Implementación:**
```go
type UsageData struct {
    Blocks []BlockData `json:"blocks"`
}

type BlockData struct {
    IsActive   bool               `json:"isActive"`
    CostUSD    float64            `json:"costUSD"`
    Projection ProjectionData     `json:"projection"`
    LimitStatus *TokenLimitStatus `json:"tokenLimitStatus,omitempty"`
}

func (t *ClaudeCodeTracker) GetAvailablePercentage() (float64, error) {
    cmd := exec.Command("ccusage", "blocks", "--active", "--token-limit", "max", "--json")
    output, err := cmd.Output()
    // ... parse JSON

    // Find active block
    for _, block := range data.Blocks {
        if block.IsActive {
            currentCost := block.CostUSD
            available := ((costLimit - currentCost) / costLimit) * 100
            return available, nil
        }
    }
}
```

**Ventajas:**
- ✅ Robusto - lee archivo local
- ✅ Mantenido activamente (v18.0.5+)
- ✅ Actualiza pricing automáticamente
- ✅ Maneja edge cases (modelos nuevos, caché, etc.)

**Alternativas investigadas (NO recomendadas):**
- ❌ OAuth API: **DISABLED** por Anthropic (`"OAuth authentication is currently not supported"`)
- ❌ Web API: Requiere manejo complejo de cookies
- ❌ CLI PTY: Inestable, depende de UI

---

##### **Codex (ChatGPT Plus)** - OAuth API ✅

**Método:** OAuth API usando credentials de `~/.codex/auth.json`

**Endpoint:**
```bash
curl -H "Authorization: Bearer $ACCESS_TOKEN" \
  https://chatgpt.com/backend-api/wham/usage
```

**Estructura JSON Output:**
```json
{
  "plan_type": "plus",
  "rate_limit": {
    "allowed": true,
    "limit_reached": false,
    "primary_window": {
      "used_percent": 3,
      "limit_window_seconds": 18000,
      "reset_after_seconds": 13638,
      "reset_at": 1768942594
    },
    "secondary_window": {
      "used_percent": 11,
      "limit_window_seconds": 604800,
      "reset_after_seconds": 275391,
      "reset_at": 1769204347
    }
  }
}
```

**Cálculo de Porcentaje:**
- OpenAI **ya provee el porcentaje usado** directamente
- `primary_window` = 5-hour session window (18000 segundos)
- `secondary_window` = 7-day rolling window (604800 segundos)
- Fórmula: `available% = 100 - used_percent`

**Implementación:**
```go
type CodexUsageOutput struct {
    PlanType  string `json:"plan_type"`
    RateLimit struct {
        Allowed      bool `json:"allowed"`
        LimitReached bool `json:"limit_reached"`
        PrimaryWindow struct {
            UsedPercent        int `json:"used_percent"`
            LimitWindowSeconds int `json:"limit_window_seconds"`
            ResetAfterSeconds  int `json:"reset_after_seconds"`
            ResetAt            int64 `json:"reset_at"`
        } `json:"primary_window"`
    } `json:"rate_limit"`
}

func (t *CodexTracker) GetAvailablePercentage() (float64, error) {
    // 1. Read access_token from ~/.codex/auth.json
    authData, _ := readCodexAuth()
    accessToken := authData.Tokens.AccessToken

    // 2. Call OAuth API
    req, _ := http.NewRequest("GET", "https://chatgpt.com/backend-api/wham/usage", nil)
    req.Header.Set("Authorization", "Bearer " + accessToken)
    resp, _ := http.DefaultClient.Do(req)

    // 3. Parse response
    var usage CodexUsageOutput
    json.NewDecoder(resp.Body).Decode(&usage)

    // 4. Calculate available percentage
    usedPercent := usage.RateLimit.PrimaryWindow.UsedPercent
    available := 100 - float64(usedPercent)

    return available, nil
}
```

**Ventajas:**
- ✅ API oficial de OpenAI
- ✅ Datos en tiempo real
- ✅ Ya provee porcentajes calculados
- ✅ Funciona con tokens existentes de Codex CLI

**Archivo de credenciales:**
```json
{
  "tokens": {
    "access_token": "eyJhbGci...",
    "refresh_token": "rt_...",
    "account_id": "1af1e167-..."
  },
  "last_refresh": "2026-01-16T21:39:02.785094Z"
}
```

---

##### **OpenCode** - NO IMPLEMENTADO ❌

**Estado:** OpenCode no tiene método de tracking confiable disponible

**Alternativas evaluadas:**
- ❌ `@ccusage/opencode`: No existe como paquete real
- ❌ OAuth API: No documentado/no disponible
- ❌ CLI tracking: No implementado

**Decisión:** Marcar como no soportado hasta que haya método confiable

---

#### 3.2 Comparación de Métodos

| Herramienta | Método | Fuente de Datos | Tiempo Real | Complejidad | Estado |
|-------------|--------|-----------------|-------------|-------------|--------|
| **Claude Code** | ccusage | Local JSONL | Sí (~5s delay) | Baja | ✅ IMPLEMENTED |
| **Codex** | OAuth API | Remote API | Sí (instant) | Baja | 🚧 IN PROGRESS |
| **OpenCode** | N/A | N/A | N/A | N/A | ❌ NOT AVAILABLE |

---

#### 3.3 Implementaciones Actualizadas

**Claude Code Tracker (FINAL):**
```go
type ClaudeCodeTracker struct {
    *BaseTracker
}

func NewClaudeCodeTracker() *ClaudeCodeTracker {
    return &ClaudeCodeTracker{
        BaseTracker: NewBaseTracker(
            "Claude Code",
            ClaudeCodeTool,
            "ccusage",
            []string{"blocks", "--active", "--token-limit", "max", "--json"},
            DefaultCostLimit, // $8.00 for Max20
        ),
    }
}

// GetAvailablePercentage uses cost-based calculation
func (b *BaseTracker) GetAvailablePercentage() (float64, error) {
    data, err := b.FetchData()
    if err != nil {
        return 0, err
    }

    // Find active block
    for _, block := range data.Blocks {
        if block.IsActive {
            currentCost := block.CostUSD
            available := ((b.costLimit - currentCost) / b.costLimit) * 100

            // Clamp to 0-100
            if available < 0 { available = 0 }
            if available > 100 { available = 100 }

            return available, nil
        }
    }

    return 0, fmt.Errorf("no active block found")
}
```

**Codex Tracker (TO IMPLEMENT):**
```go
type CodexTracker struct {
    authPath string
}

func NewCodexTracker() *CodexTracker {
    return &CodexTracker{
        authPath: filepath.Join(os.Getenv("HOME"), ".codex", "auth.json"),
    }
}

func (t *CodexTracker) readAuth() (*CodexAuth, error) {
    data, _ := os.ReadFile(t.authPath)
    var auth CodexAuth
    json.Unmarshal(data, &auth)
    return &auth, nil
}

func (t *CodexTracker) GetAvailablePercentage() (float64, error) {
    auth, _ := t.readAuth()

    req, _ := http.NewRequest("GET", "https://chatgpt.com/backend-api/wham/usage", nil)
    req.Header.Set("Authorization", "Bearer " + auth.Tokens.AccessToken)

    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()

    var usage CodexUsageOutput
    json.NewDecoder(resp.Body).Decode(&usage)

    usedPercent := usage.RateLimit.PrimaryWindow.UsedPercent
    available := 100 - float64(usedPercent)

    return available, nil
}
```

### 4. Cost Calculator (pkg/router/calculator.go)

```go
package router

type CostEstimate struct {
    Tool            string
    EstimatedCost   float64
    WillExceedLimit bool
    Confidence      float64
}

func CalculateCost(
    complexity *ComplexityAnalysis,
    toolLimits map[string]UsageTracker,
) []CostEstimate {
    estimates := []CostEstimate{}

    // Para cada herramienta:
    // 1. Chequea si tiene suficiente disponible
    // 2. Calcula costo estimado basado en tokens
    // 3. Retorna array ordenado por costo (menor a mayor)

    return estimates
}
```

### 5. Decision Engine (pkg/router/engine.go)

```go
package router

type RoutingDecision struct {
    SelectedTool   string
    Reason         string
    Alternatives   []string
    EstimatedCost  float64
}

func MakeDecision(
    complexity *ComplexityAnalysis,
    costEstimates []CostEstimate,
) *RoutingDecision {
    // Lógica de decisión:
    // 1. Filtrar herramientas sin límite
    // 2. Si hay gratis disponible → usar gratis
    // 3. Si no hay gratis → usar más barato con límite
    // 4. Si task es compleja → preferir mejor modelo aunque cueste
    // 5. Retorna decisión con reasoning

    return &RoutingDecision{
        SelectedTool: "opencode",
        Reason: "Free tier available with sufficient capacity",
        Alternatives: []string{"claude-code", "codex"},
        EstimatedCost: 0.0,
    }
}
```

### 6. Delegators (pkg/delegators/)

**Base Interface:**
```go
package delegators

type Delegator interface {
    Execute(task string) (*DelegationResult, error)
}

type DelegationResult struct {
    Success    bool
    Output     string
    Error      string
    TokensUsed int
}
```

**Claude Code Delegator:**
```go
package delegators

import "os/exec"

type ClaudeCodeDelegator struct{}

func (d *ClaudeCodeDelegator) Execute(task string) (*DelegationResult, error) {
    // Ejecuta: claude-code --non-interactive "{task}"
    cmd := exec.Command("claude-code", "--non-interactive", task)
    output, err := cmd.CombinedOutput()

    if err != nil {
        return &DelegationResult{
            Success: false,
            Error: err.Error(),
        }, err
    }

    return &DelegationResult{
        Success: true,
        Output: string(output),
    }, nil
}
```

Similar para Codex y OpenCode.

### 7. npm Wrapper (install.js)

```javascript
#!/usr/bin/env node
const { execSync } = require('child_process');
const https = require('https');
const fs = require('fs');
const path = require('path');

const platform = process.platform;
const arch = process.arch;
const version = require('./package.json').version;

const binaryName = `ai-router-${platform}-${arch}`;
const url = `https://github.com/user/ai-router/releases/download/v${version}/${binaryName}`;
const binPath = path.join(__dirname, 'bin', 'ai-router');

console.log(`Downloading ai-router binary for ${platform}-${arch}...`);

https.get(url, (response) => {
    const file = fs.createWriteStream(binPath);
    response.pipe(file);
    file.on('finish', () => {
        file.close();
        fs.chmodSync(binPath, '755');
        console.log('ai-router installed successfully!');
    });
});
```

## Flujo de Ejecución

```
User: ai-router exec "refactor auth system"
  ↓
1. Complexity Analyzer
   - Usa LLM más barato disponible
   - Analiza: "complex, ~5000 tokens"
  ↓
2. Usage Trackers (parallel)
   - Claude Code: 35% usado, 3.2h restante
   - Codex: 80% usado, 1h restante
   - OpenCode: 20% usado, 4h restante
  ↓
3. Cost Calculator
   - Claude: $0.15 estimado, disponible ✅
   - Codex: $0.10 estimado, límite cercano ⚠️
   - OpenCode: $0.00 (gratis), disponible ✅
  ↓
4. Decision Engine
   - Decisión: OpenCode (gratis + disponible)
   - Razón: "Task compleja pero OpenCode tiene límite amplio"
  ↓
5. Delegator
   - Ejecuta en OpenCode
   - Retorna resultado
  ↓
Output:
  ✅ Tarea completada con OpenCode
  💰 Costo: $0.00 (ahorro: $0.15 vs Claude Code)
  📊 Claude Code: 35% usado, OpenCode: 22% usado
```

## Comandos CLI

```bash
# Ejecutar tarea con routing automático
ai-router exec "add user authentication"

# Forzar herramienta específica
ai-router exec "fix bug" --force claude-code

# Ver estado de límites
ai-router status

# Verbose mode (ver decisión de routing)
ai-router exec "refactor code" --verbose
```

## Archivos Críticos a Crear

**Proyecto Go:**
1. `main.go` - Entry point CLI
2. `go.mod` - Go module definition
3. `cmd/exec.go` - Comando exec
4. `cmd/status.go` - Comando status
5. `pkg/analyzers/complexity.go` - Complexity analyzer
6. `pkg/trackers/*.go` - Usage trackers (4 files: interface + 3 implementations)
7. `pkg/router/calculator.go` - Cost calculator
8. `pkg/router/engine.go` - Decision engine
9. `pkg/delegators/*.go` - Delegators (4 files: interface + 3 implementations)
10. `Makefile` - Build automation

**Wrapper npm:**
1. `package.json` - npm package config
2. `install.js` - Descarga binario en install
3. `bin/ai-router.js` - Wrapper script

## Dependencias

**Go (go.mod):**
```go
module github.com/user/ai-router

go 1.21

require (
    github.com/spf13/cobra v1.8.0    // CLI framework
    github.com/fatih/color v1.16.0   // Terminal colors
)
```

**npm (package.json):**
```json
{
  "name": "ai-router",
  "version": "1.0.0",
  "description": "Cost-aware AI coding assistant router",
  "bin": {
    "ai-router": "./bin/ai-router.js"
  },
  "scripts": {
    "postinstall": "node install.js"
  },
  "files": [
    "bin/",
    "install.js"
  ]
}
```

## Build & Release

**Makefile:**
```makefile
.PHONY: build build-all release

build:
	go build -o bin/ai-router main.go

build-all:
	GOOS=darwin GOARCH=amd64 go build -o bin/ai-router-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o bin/ai-router-darwin-arm64
	GOOS=linux GOARCH=amd64 go build -o bin/ai-router-linux-amd64
	GOOS=windows GOARCH=amd64 go build -o bin/ai-router-windows-amd64.exe

release: build-all
	# Upload to GitHub releases
```

**GitHub Actions (.github/workflows/release.yml):**
```yaml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make build-all
      - uses: softprops/action-gh-release@v1
        with:
          files: bin/*
```

## Verificación

### Tests Unitarios (Go)
```bash
# Test analyzer
go test ./pkg/analyzers -v

# Test trackers
go test ./pkg/trackers -v

# Test cost calculator
go test ./pkg/router -v

# Test delegators
go test ./pkg/delegators -v
```

**Ejemplo test (pkg/trackers/claude_code_test.go):**
```go
package trackers

import "testing"

func TestClaudeCodeTracker_GetAvailablePercentage(t *testing.T) {
    tracker := &ClaudeCodeTracker{}
    percentage, err := tracker.GetAvailablePercentage()

    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }

    if percentage < 0 || percentage > 100 {
        t.Fatalf("Expected percentage between 0-100, got %f", percentage)
    }
}
```

### Test E2E
```bash
# Instalar herramienta (Go binario)
go install github.com/user/ai-router@latest

# O vía npm wrapper
npm install -g ai-router

# Test 1: Routing automático
ai-router exec "simple task" --verbose
# Debe elegir herramienta más barata disponible
# Output esperado:
# ✅ Analyzing task complexity...
# 📊 Checking tool availability...
# 💡 Selected: opencode (reason: free tier available)
# 🚀 Executing task...

# Test 2: Ver estado
ai-router status
# Debe mostrar límites de todas las herramientas
# Output esperado:
# 📊 Tool Status:
# Claude Code: 65% available, 3.2h remaining, $1.23 cost
# Codex: 20% available, 1h remaining, $0.45 cost
# OpenCode: 80% available, 4h remaining, $0.00 cost

# Test 3: Forzar herramienta
ai-router exec "task" --force codex
# Debe usar Codex aunque no sea óptimo

# Test 4: Performance test
time ai-router status
# Debe ejecutar en <50ms
```

## Mejoras Futuras (v2)

1. **GitHub Copilot support** - Agregar tracker para Copilot
2. **Learning mode** - Guardar decisiones y aprender de resultados
3. **MCP Server** - Exponer como MCP para Claude Code
4. **Dashboard web** - Visualizar costos y uso histórico
5. **Hooks integration** - Interceptar llamadas automáticamente

## Estimación de Desarrollo

- **MVP funcional (Go):** ~10-15 horas de trabajo
- **Con tests completos:** ~18-24 horas
- **npm wrapper + CI/CD:** ~3-5 horas
- **Production-ready total:** ~30-35 horas

**Nota:** Go toma ~20% más tiempo que TypeScript inicialmente, pero el binario compilado y velocidad de ejecución lo compensan.

## Ventajas de Go vs TypeScript para este proyecto

✅ **Performance:**
- Startup: ~10ms (Go) vs ~200ms (Node.js)
- Ejecución: Binario compilado nativo
- Memoria: ~5-10MB vs ~50-100MB (Node.js)

✅ **Distribución:**
- Un binario standalone (no requiere runtime)
- Cross-compilation trivial (Mac/Linux/Windows)
- npm wrapper da mejor UX manteniendo performance

✅ **Desarrollo:**
- Type-safe nativo
- Stdlib excelente (exec, json, http)
- Compilación rápida (~5-10s)

## Próximos Pasos

### Fase 1: Setup (2h)
1. Crear repositorio Go
2. Configurar go.mod
3. Setup estructura de directorios
4. Configurar Makefile

### Fase 2: Core Implementation (12h)
5. Implementar trackers (Claude Code, Codex, OpenCode) - 4h
6. Implementar complexity analyzer - 3h
7. Implementar cost calculator - 2h
8. Implementar decision engine - 2h
9. Implementar delegators - 1h

### Fase 3: CLI & Testing (8h)
10. Implementar CLI con cobra - 2h
11. Tests unitarios - 4h
12. Testing end-to-end - 2h

### Fase 4: npm Wrapper & Release (5h)
13. Crear npm wrapper package - 2h
14. GitHub Actions para releases - 2h
15. Publicar en npm - 1h

**Total:** ~30 horas para production-ready

---

## Cambios Importantes desde Plan Original

### ✅ Descubrimientos de Implementación

#### 1. Claude Code: Método de Tracking
- **Planificado:** `ccusage 5h-window --json` (NO EXISTE)
- **Real:** `ccusage blocks --active --token-limit max --json`
- **Cambio crítico:** Estructura JSON completamente diferente con BlockData

#### 2. Claude Code: Cálculo de Porcentaje
- **Planificado:** Basado en tokens usados
- **Real:** **Basado en COSTO (USD)**
- **Razón:** La UI de Claude Code muestra % de costo, no % de tokens
- **Fórmula:** `available% = ((costLimit - currentCost) / costLimit) * 100`
- **Límites reales observados:**
  - Pro: $2.00 / 5h
  - Max5: $4.00 / 5h
  - Max20: $8.00 / 5h (default usado)

#### 3. Claude Code: OAuth API No Disponible
- **Investigado:** OAuth API endpoint (`https://api.anthropic.com/api/oauth/usage`)
- **Resultado:** `"OAuth authentication is currently not supported"`
- **Decisión:** Usar ccusage (local file parsing) como único método confiable

#### 4. Codex: No Tiene ccusage Equivalent
- **Planificado:** `@ccusage/codex` similar a Claude
- **Real:** `@ccusage/codex` existe pero solo tiene comandos daily/monthly/session
- **No tiene:** Comando `blocks --active` para tracking en tiempo real
- **Solución:** Usar OAuth API directamente

#### 5. Codex: OAuth API Funciona
- **Descubierto:** OAuth API de OpenAI SÍ funciona (a diferencia de Anthropic)
- **Endpoint:** `https://chatgpt.com/backend-api/wham/usage`
- **Credenciales:** Lee de `~/.codex/auth.json` (instalado por Codex CLI)
- **Ventaja:** Ya provee porcentajes calculados directamente

#### 6. OpenCode: No Implementado
- **Estado:** No hay método confiable de tracking disponible
- **Decisión:** Posponer hasta que exista tooling adecuado

### 📊 Estado Actual de Implementación

| Componente | Estado | Notas |
|------------|--------|-------|
| Go Module Setup | ✅ DONE | v1.25.6 instalado |
| Makefile | ✅ DONE | Build/test/cross-compile |
| Trackers Interface | ✅ DONE | UsageTracker interface |
| Claude Code Tracker | ✅ DONE | Usando ccusage blocks --active |
| Codex Tracker | 🚧 IN PROGRESS | OAuth API método identificado |
| OpenCode Tracker | ❌ NOT AVAILABLE | Sin método confiable |
| Cost Calculator | ✅ DONE | Con pricing real ($0.030/1k tokens) |
| Decision Engine | ✅ DONE | Lógica de routing implementada |
| CLI (exec) | ✅ DONE | Pipeline completo de 5 pasos |
| CLI (status) | ✅ DONE | Tabla con colores y JSON output |
| Delegators | ✅ DONE | Interfaces + Claude Code impl |
| Unit Tests | ✅ DONE | 80%+ coverage |
| Integration Tests | ✅ DONE | Pipeline completo testeado |
| E2E Tests | ⏸️ PENDING | Requiere herramientas instaladas |

### 🔄 Próximos Pasos Inmediatos

1. **Implementar Codex Tracker con OAuth API**
   - Leer credentials de `~/.codex/auth.json`
   - Llamar endpoint de OpenAI
   - Parsear `primary_window.used_percent`
   - Implementar métodos de interface
   - Agregar tests unitarios

2. **Actualizar Decision Engine**
   - Soportar 2 herramientas (Claude + Codex)
   - Comparar disponibilidad entre ambas
   - Considerar Codex como alternativa válida

3. **Actualizar CLI Status**
   - Mostrar ambas herramientas en tabla
   - Comparación lado a lado

4. **Tests con Ambas Herramientas**
   - Verificar routing entre Claude y Codex
   - Test de failover cuando una no está disponible
   - Test de preferencia por herramienta más barata

### 📝 Lecciones Aprendidas

1. **No asumir existencia de comandos sin verificar**
   - Plan original asumió `5h-window` existía
   - Realidad: comando es `blocks --active`

2. **Verificar método de cálculo en UI**
   - Plan asumió % de tokens
   - Realidad: UI muestra % de costo

3. **OAuth APIs pueden estar deshabilitados**
   - Anthropic deshabilitó OAuth
   - OpenAI mantiene OAuth funcional

4. **Local file parsing puede ser más confiable que APIs**
   - ccusage es robusto y mantenido
   - Menos dependencia de APIs externas

5. **Cada herramienta tiene su propio ecosistema**
   - No hay estándar unificado
   - Necesita investigación individual por herramienta
