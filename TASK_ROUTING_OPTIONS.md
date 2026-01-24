# 🎯 Opciones de Arquitectura para Task Routing a Claude Code

> Documento que explica las 5 opciones de integración del router con Claude Code, desde lo más simple hasta lo más complejo.
> Cada opción tiene un propósito diferente y puede combinarse con otras.

---

## 📋 Resumen Rápido

| Opción | Nombre | Complejidad | Interactividad | Automatización | Velocidad | Control |
|--------|--------|-------------|----------------|----------------|-----------|---------|
| 1 | Tmux Session | ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | Manual |
| 2 | MCP Server | ⭐⭐⭐ | ⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Ninguno |
| 3 | Workspace Daemon | ⭐⭐⭐⭐⭐ | ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Ninguno |
| 4 | Watch Mode | ⭐⭐⭐⭐ | ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Ninguno |
| 5 | Approval Gate | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | Manual |

---

## 🔴 OPCIÓN 1: Terminal Multiplexer (Tmux/Zellij)

### ¿Qué es?
Abre una **terminal nueva** donde Claude Code corre interactivamente. Es como abrir una ventana de terminal normal, pero controlada por tu router.

### Analogía
```
Sin Tmux:                      Con Tmux:
Abres terminal manualmente    ai-dispatcher exec "tarea" --interactive
$ claude "mi tarea"      →     ↓
Interactúas manualmente        Se abre terminal automáticamente
                               ↓
                               claude está corriendo
                               ↓
                               Interactúas normalmente
```

### Cómo Funciona
```bash
# Usuario ejecuta:
$ ai-dispatcher exec "refactoriza la autenticación" --interactive

# Tu router hace:
1. Detecta: "Claude Code tiene 80% disponible"
2. Crea sesión tmux con contexto:
   tmux new-session -d -s "task-abc123" \
     -c "/path/to/project" \
     "claude --model opus 'refactoriza la autenticación'"
3. Abre la sesión para el usuario
4. Usuario ve terminal con Claude interactuando
5. Cuando termina, sesión cierra automáticamente
```

### Funcionalidad
✅ **Sesión Interactiva**
- Usuario puede hacer preguntas de seguimiento
- Ve output en tiempo real
- Puede copiar/pegar código
- Interacción fluida como chat normal

✅ **Contexto del Proyecto**
- Claude accede a archivos del proyecto
- Puede ejecutar comandos (npm test, etc.)
- Entiende la estructura del proyecto

✅ **Persistencia (opcional)**
- Si usas tmux, puedes cerrar terminal y reconectar después
- La sesión sigue viva

### Beneficios
- ✨ **Súper fácil de implementar** (~100 líneas de código)
- 🎯 **Transparencia total** - Usuario ve exactamente qué pasa
- 💬 **Interacción natural** - Como un chat normal
- 🔄 **Familiar** - Usuario solo abre una terminal
- 🚀 **No requiere infraestructura extra** - Solo tmux/zellij

### Desventajas
- ❌ Tmux debe estar instalado (es común en dev environments)
- ❌ No es 100% automático - Usuario debe estar presente
- ❌ Difícil capturar output automáticamente
- ❌ No ideal para tareas sin supervisión
- ❌ Difícil integrar múltiples herramientas simultáneamente

### Stack Técnico
```
Go Program (tu router)
    ↓
go-tmux library (github.com/jubnzv/go-tmux)
    ↓
Tmux CLI
    ↓
Terminal nueva con Claude Code ejecutando
```

### Comando Básico
```bash
tmux new-session -d -s "task-$uuid" \
  -c "$projectPath" \
  "claude --model opus --system-prompt '$contextPrompt'"
```

### Cuándo Usar
- ✅ Tareas complejas donde el usuario necesita supervisar
- ✅ Cuando quieres máxima transparencia
- ✅ Para desarrollo/debugging interactivo
- ✅ Cuando el usuario quiere "ver y entender" lo que hace Claude
- ❌ Para automatización sin usuario presente

---

## 🟠 OPCIÓN 2: Model Context Protocol (MCP)

### ¿Qué es?
MCP es un **protocolo estándar** que Anthropic creó para que Claude acceda a recursos sin que tú hagas nada. Es como darle a Claude una "lista de APIs" para usar.

### Analogía
```
Sin MCP (Manual):
- Tú: "Aquí están los archivos del proyecto"
- Pasas 20 archivos uno por uno
- Claude: "¿Puedo ejecutar npm test?"
- Tú: "Ok, ejecuto y te paso el resultado"
- = Tedioso

Con MCP (Automático):
- Tú: "Claude, tienes acceso a todo el proyecto"
- Claude accede automáticamente a:
  ✓ Todos los archivos
  ✓ Ejecuta comandos que quiera
  ✓ Lee resultados directamente
- = Automático, sin tu intervención
```

### Cómo Funciona
```
┌────────────────────────────────────────┐
│        Tu Router (Go Program)          │
│  - Inicia servidor MCP en :9999        │
│  - Expone recursos del proyecto:       │
│    • GET /files (lista archivos)       │
│    • POST /execute (ejecuta comandos)  │
│    • GET /project/status               │
└────────────────────┬───────────────────┘
                     │
              HTTP/WebSocket
                     │
┌────────────────────▼───────────────────┐
│       Claude Code (CLI)                │
│  $ claude --mcp-server localhost:9999  │
│                                        │
│  Claude se conecta y:                  │
│  1. Pide lista de archivos             │
│  2. Lee archivos automáticamente       │
│  3. Ejecuta comandos de test           │
│  4. Resuelve la tarea completa         │
└────────────────────────────────────────┘
```

### Funcionalidad
✅ **Acceso Automático a Proyecto**
- Claude ve todos los archivos sin que pases nada
- Puede buscar archivos automáticamente
- Lee package.json, README, etc. directamente

✅ **Ejecución de Comandos Automática**
- Claude puede: `npm test`, `git status`, `node script.js`
- Sin que tú hagas nada
- Recibe resultados directamente

✅ **Retorna Resultados Estructurados**
- Output en JSON limpio
- Fácil de procesar programáticamente
- Ideal para pipelines automatizadas

### Beneficios
- 🔌 **Estándar oficial de Anthropic** - Es el futuro
- 🤖 **Completamente automático** - Sin intervención manual
- 📊 **Retorna JSON** - Fácil integrar con otros sistemas
- 🔄 **Escalable** - Múltiples herramientas lo soportarán
- ⚡ **Eficiente** - Retorna solo lo necesario
- 🎯 **Futuro-proof** - Codex y OpenCode lo soportarán pronto

### Desventajas
- ❌ Más código para implementar
- ❌ Debes crear un "servidor" que expone recursos
- ❌ No es interactivo (sin preguntas de seguimiento)
- ❌ Requiere entender protocolo MCP
- ❌ Debugging más complicado

### Stack Técnico
```
Go Router
├─ HTTP Server (expone MCP endpoints)
├─ Spawn: claude --mcp-server http://localhost:9999
├─ Captura output (JSON)
└─ Procesa resultados

Protocolo: JSON-RPC 2.0
Transporte: HTTP/WebSocket
```

### Ejemplo de Implementación
```bash
# Tu router hace esto:

# 1. Inicia servidor MCP en background
go run cmd/mcp-server.go --port 9999 --project-path /my/project

# 2. Ejecuta Claude Code apuntando al servidor
claude -p "refactoriza toda la autenticación" \
  --output-format json \
  --mcp-server http://localhost:9999

# 3. Claude accede automáticamente a:
#    - Todos los archivos (/my/project/*)
#    - Ejecuta tests (npm test)
#    - Lee package.json
#    - Resuelve tarea

# 4. Retorna JSON con resultados
# {
#   "status": "completed",
#   "files_modified": ["auth.ts", "routes.ts"],
#   "changes": {...},
#   "tests_passed": true
# }
```

### Cuándo Usar
- ✅ Tareas que deben ser 100% automáticas
- ✅ Integración en pipelines CI/CD
- ✅ Cuando necesitas resultados estructurados
- ✅ Para producción / herramientas en background
- ❌ Para desarrollo interactivo

---

## 🔵 OPCIÓN 3: Workspace Daemon (Arquitectura Empresarial)

### ¿Qué es?
Un **programa que corre siempre en background** coordinando todas tus tareas. Es como contratar a un gerente que está siempre esperando órdenes.

### Analogía
```
Sin Daemon (Actual):
Tarea 1 → Startup (2s) → Ejecuta → Termina
Tarea 2 → Startup (2s) → Ejecuta → Termina
Tarea 3 → Startup (2s) → Ejecuta → Termina
= 6 segundos de overhead

Con Daemon (Propuesto):
Inicia una sola vez:    ai-dispatcher start-daemon
                        ↓
                    Corre siempre en background
                        ↓
Tarea 1 → Instantáneo (reutiliza daemon) → Resultado
Tarea 2 → Instantáneo (reutiliza daemon) → Resultado
Tarea 3 → Instantáneo (reutiliza daemon) → Resultado
= Sin overhead, ultra rápido
```

### Cómo Funciona
```
┌─────────────────────────────────────┐
│   AI-Dispatcher (Daemon Process)    │
│   Corriendo siempre en background   │
│                                     │
│  ├─ Workspace Manager               │
│  │  ├─ File Watcher (detecta cambios)
│  │  ├─ Project Analyzer             │
│  │  └─ Context Cache                │
│  │                                  │
│  ├─ Tool Coordinator                │
│  │  ├─ Claude Code Handler          │
│  │  ├─ Codex Handler                │
│  │  └─ OpenCode Handler             │
│  │                                  │
│  ├─ Task Queue                      │
│  │  ├─ Tarea 1 (ejecutando)        │
│  │  ├─ Tarea 2 (en cola)           │
│  │  └─ Tarea 3 (en cola)           │
│  │                                  │
│  └─ gRPC Server (puerto 50051)      │
│     Escucha órdenes externas        │
└─────────────────────────────────────┘
         ↑
    gRPC/HTTP
         ↑
User ejecuta: ai-dispatcher exec "tarea"
         ↓
Conecta al daemon (instantáneo)
         ↓
Envía tarea
         ↓
Recibe resultado
```

### Funcionalidad
✅ **Gestor Centralizado**
- Un programa coordinando todo
- Múltiples herramientas simultáneamente
- Balancea carga entre herramientas

✅ **Velocidad Extrema**
- Sin latencia de startup
- Múltiples tareas en paralelo
- Respuestas instantáneas

✅ **Inteligencia Integrada**
- Entiende qué herramienta es mejor para cada tarea
- Cache de resultados
- File watcher (detecta cambios automáticamente)

✅ **Orchestration Profesional**
- Patrón: Orchestrator-Worker
- Google pattern (8 multi-agent design patterns)
- Coordinador/Dispatcher architecture

### Beneficios
- ⚡ **Velocidad extrema** - No hay startup
- 🎯 **Coordinación perfecta** - Múltiples herramientas trabajando juntas
- 📈 **Escalable** - Maneja 100s de tareas
- 🔄 **Load Balancing** - Distribuye trabajo automáticamente
- 💾 **Caching inteligente** - No repite análisis
- 👀 **File Watcher** - Detecta cambios en proyecto
- 🏢 **Enterprise-ready** - Patrón usado por Google, Vercel
- 📊 **Observabilidad** - Logs, metrics, tracing

### Desventajas
- ❌ Arquitectura muy compleja
- ❌ Requiere manejo robusto de procesos
- ❌ Debugging más difícil
- ❌ Más líneas de código (~2000+ líneas)
- ❌ Manejo de fallos y recuperación
- ❌ Testing más complicado

### Stack Técnico
```
Go Daemon (main process):
├─ gRPC Server (escucha órdenes)
├─ File Watcher (fsnotify)
├─ Process Manager (exec, lifecycle)
├─ Context Cache (memoria)
└─ Load Balancer (distribuye entre tools)

Cliente (CLI):
├─ gRPC Client
└─ Se conecta al daemon

Protocolo: gRPC/Protocol Buffers
Transporte: Unix Socket o TCP localhost
```

### Ejemplo de Uso
```bash
# Una sola vez, inicia el daemon:
$ ai-dispatcher start-daemon
✓ Daemon iniciado en pid 12345
✓ Escuchando en localhost:50051

# Ahora, cualquier comando es instantáneo:
$ ai-dispatcher exec "refactoriza auth"
✓ Tarea enrutada a Claude Code (80% disponible)
✓ Completado en 12 segundos

$ ai-dispatcher exec "write tests"
✓ Tarea enrutada a Claude Code
✓ Completado en 8 segundos

$ ai-dispatcher exec "document code"
✓ Tarea enrutada a Codex (60% disponible)
✓ Completado en 5 segundos

# Ver status en tiempo real:
$ ai-dispatcher status
Claude Code      80% ⚡ (ejecutando 1 tarea)
Codex            60% ✓ (idle)
OpenCode        100% ✓ (idle)
```

### Cuándo Usar
- ✅ Herramienta en producción/profesional
- ✅ Cuando necesitas máxima performance
- ✅ Múltiples tareas simultáneamente
- ✅ Para CI/CD pipelines
- ✅ Cuando quieres "set it and forget it"
- ❌ Para MVP rápido
- ❌ Si necesitas interactividad alta

---

---

## 🟢 OPCIÓN 4: Watch Mode (Monitoreo Automático)

### ¿Qué es?
Un **programa que observa tu proyecto 24/7** y automáticamente ejecuta tareas cuando detecta ciertos eventos (tests fallan, errores en logs, cambios pendientes).

### Analogía
```
Sin Watch Mode (Manual):
- Tests fallan
- Tú lo ves
- Tú ejecutas: ai-dispatcher exec "Arregla el test"
- Esperas resultado
- Repites manualmente

Con Watch Mode (Automático):
- Tests fallan
- Sistema automáticamente: ai-dispatcher exec "Arregla el test"
- Claude lo arregla sin que hagas nada
- Sistema corre tests para verificar
- Todo automático, como tener un "asistente vigilante"
```

### Cómo Funciona

```
┌─────────────────────────────────────┐
│   AI Dispatcher Watch Mode          │
│   (Daemon ejecutando)               │
│                                     │
│  ├─ File Watcher                    │
│  │  └─ Monitorea: *.js, *.ts        │
│  │                                  │
│  ├─ Test Monitor                    │
│  │  └─ Ejecuta tests cada 30s       │
│  │  └─ Si fallan → Dispara trigger  │
│  │                                  │
│  ├─ Error Log Monitor               │
│  │  └─ Lee app logs                 │
│  │  └─ Si detecta error → Dispara   │
│  │                                  │
│  ├─ Trigger Rules                   │
│  │  ├─ IF tests_fail → exec "arregla tests"
│  │  ├─ IF error_found → exec "analiza y corrige"
│  │  ├─ IF git_changes → exec "revisa cambios"
│  │  └─ IF performance_degraded → exec "optimiza"
│  │                                  │
│  └─ Auto-Execute                    │
│     └─ Ejecuta Claude Code con tarea generada
│        automáticamente             │
└─────────────────────────────────────┘
```

### Funcionalidad

✅ **Monitoreo Continuo**
- Detecta cambios en archivos
- Ejecuta tests periódicamente
- Monitorea logs de errores
- Detecta degradación de performance

✅ **Triggers Configurables**
- Tests fallidos → Automáticamente ejecuta arreglo
- Error detectado → Automáticamente analiza
- Cambios sin pusear → Automáticamente revisa
- Performance baja → Automáticamente optimiza

✅ **Auto-Ejecución**
- Sin intervención manual
- Claude corre automáticamente
- Resultados aplicados automáticamente
- Notificaciones cuando termina

### Beneficios

- 🤖 **Completamente automático** - Ni tocas nada
- 👀 **Vigilancia 24/7** - Siempre monitoreando
- ⚡ **Reacciones instantáneas** - Arregla tan pronto detecta
- 🧪 **Tests siempre pasando** - Si fallan, se auto-arreglan
- 📊 **Logs limpios** - Errores se minimizan automáticamente
- 🎯 **Desarrollo sin interrupciones** - Tú código, sistema lo cuida
- 💾 **Historial completo** - Qué cambios hizo, cuándo, por qué

### Desventajas

- ❌ Consume recursos constantemente (RAM, CPU)
- ❌ Puede hacer cambios que no siempre apruebas
- ❌ Difícil debugear "por qué" hizo X cambio
- ❌ Si Claude "alucinaba", podría arruinar el proyecto
- ❌ Requiere configuración de triggers (no es fire-and-forget)
- ❌ Puede ser paranoico (ejecutar demasiado frecuente)

### Stack Técnico

```
Go Daemon (similar a Opción 3):
├─ File Watcher (fsnotify)
├─ Test Monitor (ejecuta npm test cada X segundos)
├─ Log Parser (lee logs en tiempo real)
├─ Trigger Engine (evalúa condiciones)
├─ MCP Server (para ejecutar Claude)
└─ Notifier (Slack, email, webhook)

Configuración (YAML o JSON):
triggers:
  - name: "test_fix"
    event: "test_failed"
    action: "exec"
    task: "Arregla el test fallido"

  - name: "error_fix"
    event: "error_in_logs"
    action: "exec"
    task: "Analiza este error y corrige"

  - name: "code_review"
    event: "git_changes_detected"
    action: "exec"
    task: "Revisa cambios no pusheados"
```

### Ejemplo de Uso

```bash
# Usuario configura watch mode una sola vez:
$ ai-dispatcher watch-mode --config config.yaml

✓ Watch mode iniciado
  ├─ Monitoreando archivos *.js, *.ts
  ├─ Tests: cada 30 segundos
  ├─ Error logs: en tiempo real
  └─ Listo para detectar eventos

# Usuario hace desarrollo normal, y el sistema automáticamente:

# Escena 1: Tests fallan
$ npm test
FAIL src/auth.test.js

[Sistema detecta automáticamente]
🔄 Ejecutando trigger: test_fix
🤖 Claude Code analizando test fallido...
   - Leyendo test fallido
   - Leyendo código
   - Creando fix
✅ Fix aplicado, tests ahora pasan

# Escena 2: Error en logs
[App logs muestran error de ReferenceError]

[Sistema detecta automáticamente]
🔄 Ejecutando trigger: error_fix
🤖 Claude Code analizando error...
   - Leyendo stack trace
   - Encontrando línea del error
   - Aplicando fix
✅ Error corregido

# Escena 3: Usuario hace cambios
$ git add .
$ # No hace commit, cambios pendientes

[Sistema detecta automáticamente cada 5 min]
🔄 Ejecutando trigger: code_review
🤖 Claude Code revisando cambios...
   - Analizando diff
   - Verificando calidad
   - Sugiriendo mejoras
✅ Revisión completada, cambios OK para pusear
```

### Flujo de Trabajo Real

```bash
# Usuario NUNCA executa ai-dispatcher manualmente

# Inicia watch mode (una sola vez):
$ ai-dispatcher watch-mode --config project-config.yaml
✓ Sistema vigilante iniciado

# Usuario hace desarrollo:
$ vim src/auth/service.js
$ # Hace cambios

[Sistema automáticamente]
├─ Detecta cambio (file watcher)
├─ Ejecuta linting (si configuras)
├─ Ejecuta tests (si configuras)
└─ Si algo falla, Claude lo arregla

[Usuario termina su trabajo]
$ git push
# Todo ya está arreglado automáticamente

# Ver qué hizo el sistema:
$ ai-dispatcher watch-logs
[Historial de lo que Claude ha arreglado hoy]
```

### Cuándo Usar

- ✅ Desarrollo local iterativo (tú escribes, sistema cuida la calidad)
- ✅ CI/CD pipeline (detecta issues antes de merge)
- ✅ Mantenimiento de proyecto (observa y cuida)
- ✅ Cuando quieres "manos libres" total
- ❌ Cuando necesitas control manual (uso Opción 5)
- ❌ Cuando no confías en auto-fixes (demasiado riesgo)

---

## 🟡 OPCIÓN 5: Approval Gate (Supervisión Híbrida)

### ¿Qué es?
Claude propone cambios, pero **TÚ apruebas antes** de aplicarse. Es el equilibrio perfecto: automatización PERO con control humano.

### Analogía
```
Sin Approval Gate (Opción 2/3):
- Ejecutas comando
- Claude hace cambios
- Se aplican automáticamente
- "Espero que salga bien" 😰

Con Approval Gate:
- Ejecutas comando
- Claude PROPONE cambios
- TÚ ves exactamente qué va a hacer
- TÚ apruebaas (o rechazas)
- ENTONCES se aplica
- Control total 👍
```

### Cómo Funciona

```
┌────────────────────────────────────────┐
│   Usuario ejecuta tarea                │
└────────────┬─────────────────────────┘
             │
             ↓
┌────────────────────────────────────────┐
│   Claude analiza y PROPONE cambios    │
│   (pero NO los aplica todavía)         │
└────────────┬─────────────────────────┘
             │
             ↓
┌────────────────────────────────────────┐
│   Sistema muestra propuesta a usuario: │
│                                        │
│   📋 CAMBIOS PROPUESTOS                │
│   ├─ src/auth.js (15 líneas cambios) │
│   ├─ src/routes.js (8 líneas cambios)│
│   └─ src/utils.js (5 líneas cambios) │
│                                        │
│   ¿Apruebas? (y/n/review)              │
└────────────┬─────────────────────────┘
             │
    ┌────────┴────────┬──────────────┐
    ↓                 ↓              ↓
  [y]              [n]            [review]
    │                │              │
    ├─Aplica         ├─Descarta   ├─Muestra diff
    │ cambios        │ cambios    ├─Usuario edita
    ├─Corre tests    │            │ propuesta
    ├─Notifica OK    │            ├─Aprueba cambios
    └─Fin            └─Fin        │ editados
                                  └─Aplica y fin
```

### Funcionalidad

✅ **Propuesta Clara**
- Claude muestra exactamente qué cambiaría
- Archivos a modificar listados
- Cantidad de cambios por archivo
- Preview del código nuevo

✅ **Múltiples Opciones**
- Aprobar (apply all)
- Rechazar (discard all)
- Revisar primero (ver diff detallado)
- Editar propuesta (cambiar manualmente)
- Aprobar parcial (solo algunos archivos)

✅ **Ejecución Condicionada**
- Solo aplica si usuario aprueba explícitamente
- Tests corren automáticamente DESPUÉS de aprobar
- Notificación si tests fallan
- Oportunidad de revertir

✅ **Auditoría Completa**
- Qué cambios propuso Claude
- Cuándo fue aprobado
- Quién aprobó
- Tests antes/después

### Beneficios

- ✋ **Control manual** - TÚ decides qué se aplica
- 👁️ **Visibilidad total** - Ves exactamente qué hace
- 🔍 **Reviewable** - Puedes revisar antes de aplicar
- ✏️ **Editable** - Puedes ajustar la propuesta
- 📋 **Auditable** - Historial completo de aprobaciones
- 🧪 **Tests automáticos** - Verifica que no rompió nada
- 🛡️ **Seguridad** - No aplica cambios sin tu OK
- 🎯 **Confianza** - Perfecta para producción

### Desventajas

- ⏸️ Requiere que estés presente (no fully async)
- ⏱️ Más lento (aprobación manual = latencia)
- 👤 No escalable para 1000s de cambios (mucho click)
- 🧠 Aún requiere tu decisión (no es 100% automático)

### Stack Técnico

```
Go Program:
├─ Execute Claude (no apply yet)
├─ Generate Diff (muestra cambios)
├─ Approval Engine:
│  ├─ CLI interactive prompt
│  ├─ O: Web UI (mejor para equipos)
│  └─ O: Webhook callback
├─ Change Applier (aplica solo si aprobado)
├─ Test Runner (corre tests)
└─ Audit Logger (registra todo)

Frontend (opcional, para equipos):
├─ Web UI mostrando propuesta
├─ Diff viewer
├─ Approve/Reject buttons
└─ Comments/notes
```

### Ejemplo de Uso

```bash
$ ai-dispatcher exec "Refactoriza autenticación" --require-approval

🔍 Analizando proyecto...
🤖 Claude proponiendo cambios...

📋 PROPUESTA GENERADA
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 Cambios propuestos:

  src/auth/service.js
    ├─ Línea 15-30: Nueva función validateJWT()
    ├─ Línea 45-60: Reemplazar middleware antiguo
    └─ Línea 120: Añadir logger
    Total: +47 líneas, -32 líneas

  src/auth/middleware.js
    ├─ Línea 1-20: Importar nuevo middleware
    └─ Línea 35: Usar nuevo validateJWT
    Total: +5 líneas, -8 líneas

  src/config/.env.example
    ├─ Añadir: JWT_SECRET=...
    └─ Añadir: JWT_EXPIRY=...
    Total: +2 líneas

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

¿Qué quieres hacer?
  [y] - Aprobar y aplicar todos los cambios
  [n] - Rechazar todos los cambios
  [r] - Revisar diff detallado
  [e] - Editar propuesta manualmente
  [p] - Aprobar solo algunos archivos
  [c] - Cancelar

Opción: r
```

**Si usuario elige [r] (review):**

```bash
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
src/auth/service.js
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  14 | function authenticateUser(username, password) {
  15 |   // ANTES:
  16 |-  const token = jwt.sign({user: username}, SECRET)
  17 |-  return token
  18 |+  // AHORA:
  19 |+  const token = jwt.sign(
  20 |+    { user: username, role: getUserRole(username) },
  21 |+    JWT_SECRET,
  22 |+    { expiresIn: JWT_EXPIRY }
  23 |+  )
  24 |+  logger.info(`Token generated for ${username}`)
  25 |+  return token
  26 | }
  27 |
  28 |+function validateJWT(token) {
  29 |+  try {
  30 |+    return jwt.verify(token, JWT_SECRET)
  31 |+  } catch (err) {
  32 |+    logger.warn(`Invalid JWT: ${err.message}`)
  33 |+    return null
  34 |+  }
  35 |+}

[Mostrando diff completo interactivo]

¿Apruebas este cambio? [y/n/edit]
```

**Si usuario elige [y] (aprobar):**

```bash
✅ Cambios aprobados
🔄 Aplicando cambios...
   ├─ Modificando src/auth/service.js
   ├─ Modificando src/auth/middleware.js
   └─ Creando backup de archivos anteriores

🧪 Ejecutando tests...
   ├─ npm test
   ├─ ✓ Auth tests: PASSED
   ├─ ✓ Integration tests: PASSED
   └─ ✓ All tests: PASSED

✅ COMPLETADO EXITOSAMENTE
   ├─ Archivos modificados: 2
   ├─ Cambios aplicados: 54 líneas
   ├─ Tests: ✓ TODO OK
   └─ Backup guardado: .backup/auth-2026-01-23-14:32
```

**Si usuario elige [e] (editar):**

```bash
📝 EDITOR ABIERTO
   [Se abre tu editor favorito]
   [Puedes editar los cambios propuestos]
   [Guardas cuando terminas]

✓ Cambios editados
¿Ahora apruebas? [y/n]
```

### Flujo de Trabajo Real

```bash
# DÍA 1: Refactorización crítica
$ ai-dispatcher exec "Refactoriza auth a JWT" --require-approval

[Sistema propone cambios]
[Usuario revisa diff]
[Usuario: "Apruebo, pero cambiar el nombre de esta función"]
[Usuario edita]
[Usuario aprueba cambios editados]
✅ Se aplica con tus cambios incluidos

# DÍA 2: Cambios rápidos con confianza
$ ai-dispatcher exec "Escribe tests para JWT" --require-approval

[Sistema propone tests]
[Usuario: Revisa brevemente, "Looks good"]
[Usuario aprueba]
✅ Tests creados, todo pasa

# DÍA 3: Cambio que no te gusta
$ ai-dispatcher exec "Refactoriza database layer" --require-approval

[Sistema propone cambios]
[Usuario revisa]
[Usuario: "Esto no es lo que quería"]
[Usuario rechaza]
❌ Cambios descartados, proyecto sin modificar

# Ver historial:
$ ai-dispatcher approval-history
Aprobados hoy: 3
Rechazados hoy: 1
Editados hoy: 2
```

### Cuándo Usar

- ✅ **Producción** - No quieres que Claude haga cambios sin supervisión
- ✅ **Equipos** - Múltiples personas revisando cambios
- ✅ **Código crítico** - Auth, payments, security
- ✅ **Cuando no confías 100%** - Quieres safety net
- ✅ **Aprendizaje** - Ver qué propone Claude, luego decidir
- ❌ Desarrollo veloz (es más lento por aprobación)
- ❌ Cuando tienes total confianza (usa Opción 2/3)

---

## 📊 Comparativa Detallada

### Por Criterio

#### **Facilidad de Implementación**
```
Opción 1 (Tmux):     ████░░░░░░ 40% - Relativamente simple
Opción 2 (MCP):      ██████░░░░ 60% - Moderada
Opción 3 (Daemon):   ██████████ 100% - Muy compleja
Opción 4 (Watch):    ████████░░ 80% - Compleja (daemon + triggers)
Opción 5 (Approval): ██████░░░░ 60% - Moderada (UI de aprobación)
```

#### **Velocidad de Ejecución**
```
Opción 1 (Tmux):     ██░░░░░░░░ 20% - Hay startup delay
Opción 2 (MCP):      ████░░░░░░ 40% - Razonable
Opción 3 (Daemon):   ██████████ 100% - Instantáneo
Opción 4 (Watch):    ██████████ 100% - Instantáneo (ya corriendo)
Opción 5 (Approval): ████░░░░░░ 40% - Razonable + aprobación manual
```

#### **Interactividad con Usuario**
```
Opción 1 (Tmux):     ██████████ 100% - Máxima (chat en vivo)
Opción 2 (MCP):      ░░░░░░░░░░ 0%   - Ninguna
Opción 3 (Daemon):   ░░░░░░░░░░ 0%   - Ninguna
Opción 4 (Watch):    ░░░░░░░░░░ 0%   - Ninguna (es automático)
Opción 5 (Approval): ████░░░░░░ 40% - Aprobación manual solamente
```

#### **Automatización**
```
Opción 1 (Tmux):     ██░░░░░░░░ 20% - Difícil automatizar
Opción 2 (MCP):      ████████░░ 80% - Muy automatizable
Opción 3 (Daemon):   ██████████ 100% - Perfectamente automatizable
Opción 4 (Watch):    ██████████ 100% - Totalmente automático (0 intervención)
Opción 5 (Approval): ████░░░░░░ 40% - Automático PERO requiere aprobación
```

#### **Escalabilidad (múltiples tareas)**
```
Opción 1 (Tmux):     ████░░░░░░ 40% - Puedes abrir varias terminales
Opción 2 (MCP):      ██████░░░░ 60% - Soporta múltiples, pero simple
Opción 3 (Daemon):   ██████████ 100% - Diseñada para esto
Opción 4 (Watch):    ██████████ 100% - Monitorea todo simultáneamente
Opción 5 (Approval): ████░░░░░░ 40% - Aprobaciones pueden ser cuello botella
```

#### **Listo para Producción**
```
Opción 1 (Tmux):     ██░░░░░░░░ 20% - Solo para dev interactivo
Opción 2 (MCP):      ████████░░ 80% - Muy listo
Opción 3 (Daemon):   ██████████ 100% - Completamente listo
Opción 4 (Watch):    ████████░░ 80% - Listo, pero requiere confianza
Opción 5 (Approval): ██████████ 100% - Perfecta para producción
```

#### **Control Manual (Seguridad)**
```
Opción 1 (Tmux):     ██████████ 100% - Total control, ves todo
Opción 2 (MCP):      ░░░░░░░░░░ 0%   - Sin control, confía todo a Claude
Opción 3 (Daemon):   ░░░░░░░░░░ 0%   - Sin control, confía todo a Claude
Opción 4 (Watch):    ░░░░░░░░░░ 0%   - Sin control, confía todo a Claude
Opción 5 (Approval): ██████████ 100% - Control total, apruebas cambios
```

---

## 🎯 Recomendación por Caso de Uso

### Caso 1: "Quiero empezar RÁPIDO y ver qué pasa"
**→ Usa OPCIÓN 1 (Tmux)**
- Implementas en 2-3 horas
- Funciona perfectamente para desarrollo
- Usuario ve exactamente qué pasa
- Fácil de debugear
- Máxima transparencia

### Caso 2: "Quiero automatización pura, sin intervención"
**→ Usa OPCIÓN 2 (MCP)**
- Implementas en 1-2 días
- Totalmente automático
- Retorna resultados limpios
- Fácil integrar con otros sistemas
- Estándar de Anthropic
- Sin latencia de startup

### Caso 3: "Quiero una herramienta PROFESIONAL para múltiples tareas"
**→ Usa OPCIÓN 3 (Daemon)**
- Implementas en 3-5 días
- Velocidad extrema
- Múltiples herramientas coordinadas
- Listo para vender/publicar
- Enterprise-ready
- Load balancing automático

### Caso 4: "Quiero que el sistema CUIDE mi proyecto automáticamente"
**→ Usa OPCIÓN 4 (Watch Mode)**
- Implementas en 2-3 días
- Sistema observa 24/7
- Arregla tests fallidos automáticamente
- Detecta errores en logs y los corrige
- Perfecto para desarrollo iterativo
- Requiere confianza en Claude

### Caso 5: "Quiero automatización PERO con control total (producción)"
**→ Usa OPCIÓN 5 (Approval Gate)**
- Implementas en 1-2 días
- Claude propone, TÚ apruebas
- Control total sobre cambios
- Perfecta para código crítico
- Auditoría completa
- Safe para producción

---

## 🚀 Mi Recomendación: Progresión

### **Fase 1 (Semana 1): Opción 1 (Tmux)**
```bash
ai-dispatcher exec "tarea" --interactive
# Usuario interactúa con Claude como chat normal
```
**Objetivo:** Validar concepto, ver si el routing funciona
**Usuario:** Developers que quieren interactividad

### **Fase 2a (Semana 2): Opción 2 (MCP)**
```bash
ai-dispatcher exec "tarea"
# Se ejecuta automático, retorna JSON
```
**Objetivo:** Automatización completa, sin supervisor
**Usuario:** Developers en desarrollo local

### **Fase 2b (Semana 2): Opción 5 (Approval Gate)**
```bash
ai-dispatcher exec "tarea" --require-approval
# Claude propone, usuario aprueba
```
**Objetivo:** Automatización SEGURA para producción
**Usuario:** Equipos, código crítico, pipelines CI/CD

### **Fase 3 (Roadmap): Opción 4 (Watch Mode)**
```bash
ai-dispatcher watch-mode --config config.yaml
# Sistema monitorea y arregla automáticamente
```
**Objetivo:** Vigilancia 24/7, manos libres
**Usuario:** Mantenimiento de proyectos

### **Fase 4 (Roadmap Avanzado): Opción 3 (Daemon)**
```bash
ai-dispatcher start-daemon
ai-dispatcher exec "tarea 1"  # Instantáneo
ai-dispatcher exec "tarea 2"  # Instantáneo
ai-dispatcher exec "tarea 3"  # Instantáneo
```
**Objetivo:** Herramienta profesional, múltiples tareas paralelas
**Usuario:** Empresas, herramientas de venta, SaaS

---

## 💻 Requerimientos Técnicos

### Opción 1 (Tmux)
```
Dependencias:
- tmux instalado (brew install tmux)
- Go library: github.com/jubnzv/go-tmux
- Claude Code CLI instalado

Código aproximado: 100-150 líneas
Files: cmd/tmux.go + pkg/spawner/tmux.go
```

### Opción 2 (MCP)
```
Dependencias:
- Go http package (built-in)
- Claude Code CLI instalado
- Entender MCP spec (leer docs)

Código aproximado: 300-500 líneas
Files: pkg/mcp/server.go + pkg/mcp/handlers.go
```

### Opción 3 (Daemon)
```
Dependencias:
- Go gRPC (google.golang.org/grpc)
- Protocol Buffers (protobuf)
- fsnotify (file watcher)
- Process management libraries

Código aproximado: 1500-2500 líneas
Files: pkg/daemon/* + pkg/grpc/* + cmd/daemon.go
```

### Opción 4 (Watch Mode)
```
Dependencias:
- Go Daemon (similar a Opción 3)
- fsnotify (file watcher)
- Log parser library
- Cron/scheduler library

Código aproximado: 800-1200 líneas
Files: pkg/watch/* + pkg/triggers/* + config files
```

### Opción 5 (Approval Gate)
```
Dependencias:
- Go http package (built-in)
- Diff generator library (github.com/go-diff/diff)
- CLI interactive prompts (charmbracelet/bubbles)
- O: Web UI (simple React/HTML)

Código aproximado: 400-700 líneas
Files: pkg/approval/* + cmd/approval-ui.go
```

---

## 📚 Recursos

### Opción 1 (Tmux)
- [go-tmux Library](https://pkg.go.dev/github.com/jubnzv/go-tmux)
- [Tmux Documentation](https://github.com/tmux/tmux/wiki)
- [iTerm2 Tmux Integration](https://iterm2.com/documentation-tmux-integration.html)

### Opción 2 (MCP)
- [Model Context Protocol Specification](https://modelcontextprotocol.io/specification/2025-11-25)
- [Claude Code MCP Integration](https://claudecode.io/guides/mcp-integration)
- [Anthropic News - MCP Launch](https://www.anthropic.com/news/model-context-protocol)

### Opción 3 (Daemon)
- [Google 8 Multi-Agent Design Patterns](https://www.infoq.com/news/2026/01/multi-agent-design-patterns/)
- [Dagster Daemon Architecture](https://docs.dagster.io/deployment/oss/oss-deployment-architecture)
- [gRPC Documentation](https://grpc.io/docs/what-is-grpc/)

### Opción 4 (Watch Mode)
- [fsnotify - File Watcher](https://github.com/fsnotify/fsnotify)
- [Trigger/Rule Engine Pattern](https://en.wikipedia.org/wiki/Event_stream_processing)
- [Log Monitoring & Parsing](https://www.elastic.co/what-is/elasticsearch)

### Opción 5 (Approval Gate)
- [go-diff Library](https://pkg.go.dev/github.com/go-diff/diff)
- [Charmbracelet Bubbles (TUI)](https://github.com/charmbracelet/bubbles)
- [Approval Testing Pattern](https://en.wikipedia.org/wiki/Approval_testing)

---

## ❓ Preguntas Frecuentes

**P: ¿Puedo cambiar de opción después?**
R: Sí, las opciones son progresivas. Empiezas con 1, pasas a 2 o 5, luego a 3 o 4. No son excluyentes.

**P: ¿Cuál es la diferencia entre Opción 2 y Opción 5?**
R: Opción 2 (MCP) = Automatización total sin intervención. Opción 5 (Approval) = Automatización CON aprobación manual. Usa 5 si quieres control, 2 si confías 100%.

**P: ¿Opción 4 puede romper mi proyecto?**
R: Sí, si Claude "alucina". Por eso se recomienda solo cuando confías mucho o en desarrollo local.

**P: ¿Puedo usar Watch Mode + Approval Gate?**
R: Sí, sistema detecta issues → Propone fixes → TÚ apruebas → Se aplica. Lo mejor de ambos.

**P: ¿Qué pasa con múltiples usuarios?**
R: Opción 3 (Daemon) y Opción 5 (Approval) están diseñadas para esto.

**P: ¿Necesito todas las herramientas (Claude Code, Codex, OpenCode)?**
R: No, empiezas con Claude Code. Las otras son para después.

**P: ¿Qué sucede si Claude Code no está disponible?**
R: Tu router ya tiene lógica para esto - usa el siguiente tool disponible según disponibilidad.

**P: ¿Puedo combinar opciones?**
R: Sí, ejemplo: Opción 1 (--interactive) + Opción 3 (daemon en background), o Opción 4 (Watch) + Opción 5 (Approval).

**P: ¿Cuál recomiendan para producción?**
R: Opción 5 (Approval Gate) porque tienes control total. O Opción 2 (MCP) si confías 100% en tu setup.

---

## 🔄 Flujos de Trabajo Reales

### OPCIÓN 2 (MCP): Flujo de Trabajo Día a Día

#### **Escenario: Refactorizar un servicio de autenticación**

**PASO 1: Usuario lanza la tarea**
```bash
$ cd /my-project
$ ai-dispatcher exec "Refactoriza el servicio de autenticación para usar JWT en lugar de sesiones"
```

**¿QUÉ PASA INTERNAMENTE?**

**Paso 1.1: Router analiza la complejidad**
```
Router: "Esta es una tarea COMPLEJA (refactorización, múltiples archivos)"
      └─ Tokens estimados: ~1500
      └─ Tiempo estimado: 5-10 minutos
```

**Paso 1.2: Router selecciona herramienta**
```
Router: "Verificando disponibilidad..."
      └─ Claude Code: 80% disponible ✓ SELECCIONADO
      └─ Codex: 60% disponible
      └─ OpenCode: 100% disponible

Router: "Selecciono Claude Code (mayor especialidad, suficiente disponibilidad)"
```

**Paso 1.3: Router inicia servidor MCP**
```
Router: "Iniciando servidor MCP en :9999..."
      ├─ Servidor HTTP en http://localhost:9999
      ├─ Exponiendo recursos:
      │  ├─ GET /files → Lista todos los archivos del proyecto
      │  ├─ GET /file/:path → Lee contenido de archivo
      │  ├─ POST /execute → Ejecuta comandos (npm test, git, etc)
      │  ├─ GET /project/structure → Estructura del proyecto
      │  └─ GET /project/dependencies → package.json, etc
      └─ Servidor listo ✓
```

**Paso 1.4: Router ejecuta Claude Code apuntando al MCP**
```bash
Router ejecuta:
$ claude -p \
  "Refactoriza el servicio de autenticación para usar JWT en lugar de sesiones" \
  --output-format json \
  --mcp-server http://localhost:9999

# El "-p" significa "print mode" = no interactivo, retorna JSON
```

---

**¿QUÉ HACE CLAUDE CODE?**

Claude se conecta automáticamente al servidor MCP y:

```
Claude: "Conectándome al MCP server..."
        └─ Conectado a http://localhost:9999 ✓

Claude: "Analizando proyecto..."
        ├─ Pide: GET /files
        ├─ Recibe: Lista de todos los archivos
        ├─ Pide: GET /file/src/auth/service.js
        ├─ Recibe: Contenido del archivo de autenticación
        ├─ Pide: GET /project/structure
        ├─ Recibe: Estructura completa del proyecto
        └─ Ahora entiende la arquitectura ✓

Claude: "Refactorizando a JWT..."
        ├─ Ejecuta: POST /execute → "npm test" (para entender tests)
        ├─ Recibe: Output de tests actuales
        ├─ Lee: src/auth/middleware.js
        ├─ Lee: src/auth/routes.js
        ├─ Lee: src/config/database.js
        └─ Crea plan de refactorización ✓

Claude: "Modificando archivos..."
        ├─ Modifica: src/auth/service.js (nuevo JWT logic)
        ├─ Modifica: src/auth/middleware.js (nuevo middleware)
        ├─ Crea: src/auth/jwt-utils.js (funciones auxiliares)
        ├─ Modifica: .env.example (nuevas variables)
        └─ Cambios completados ✓

Claude: "Verificando cambios..."
        ├─ Ejecuta: POST /execute → "npm test"
        ├─ Recibe: ✓ Todos los tests pasaron
        ├─ Ejecuta: POST /execute → "npm run lint"
        ├─ Recibe: ✓ No hay errores de linting
        └─ Validación completada ✓

Claude: "Generando reporte..."
        └─ Retorna JSON con:
           {
             "status": "completed",
             "files_modified": [
               "src/auth/service.js",
               "src/auth/middleware.js"
             ],
             "files_created": [
               "src/auth/jwt-utils.js"
             ],
             "tests_status": "passed",
             "summary": "Refactorización completada. Cambios: ..."
           }
```

---

**¿QUÉ VE EL USUARIO?**

```bash
$ ai-dispatcher exec "Refactoriza el servicio de autenticación para usar JWT..."

🔍 Step 1/5: Analyzing task complexity...
   Level: complex
   Tokens: ~1500
   Method: llm

⚙️  Step 2/5: Initializing decision engine...

📊 Step 3/5: Checking tool availability...
   ✓ Claude Code: 80% available

🎯 Step 4/5: Making routing decision...
   Selected tool: Claude Code
   Estimated cost: $0.30

🚀 Step 5/5: Executing task...

   🔗 Iniciando servidor MCP en :9999
   🤖 Lanzando Claude Code...
   📡 Claude se conecta al servidor MCP

   ⏳ Analizando proyecto...
   📂 Leyendo archivos...
   🧪 Ejecutando tests...

   ✨ Refactorizando...
   📝 Modificando: src/auth/service.js
   📝 Modificando: src/auth/middleware.js
   ✨ Creando: src/auth/jwt-utils.js

   🧪 Verificando tests...
   ✓ Todos los tests pasaron

   ✅ Tarea completada en 8 minutos 32 segundos

📋 Resultados:
   ├─ Archivos modificados: 2
   ├─ Archivos creados: 1
   ├─ Tests: ✓ PASSED
   ├─ Linting: ✓ PASSED
   └─ Resumen: Refactorización completada exitosamente
```

---

#### **¿Qué pasa si necesitas iteraciones (cambios)?**

```bash
# Usuario ve que falta añadir documentación
$ ai-dispatcher exec "Añade JSDoc a todas las funciones nuevas de autenticación"

# Nuevamente:
🚀 Step 5/5: Executing task...
   🔗 Iniciando servidor MCP en :9999
   🤖 Lanzando Claude Code...
   📡 Claude se conecta y AUTOMÁTICAMENTE:
      ├─ Ve los cambios previos (MCP lee archivos actuales)
      ├─ Entiende la nueva arquitectura
      ├─ Añade JSDoc a todas las funciones nuevas
      └─ Ejecuta tests para verificar

✅ Tarea completada en 2 minutos 15 segundos
   ├─ Archivos modificados: 2
   └─ Resumen: JSDoc añadido a todas las nuevas funciones
```

---

**RESUMEN OPCIÓN 2 (MCP):**
```
Usuario    →  "Quiero refactorizar"
           →  ai-dispatcher exec "tarea"

Router     →  Analiza, selecciona Claude Code
           →  Inicia servidor MCP
           →  Ejecuta Claude apuntando a MCP

Claude     →  Se conecta al MCP
           →  Lee AUTOMÁTICAMENTE archivos
           →  Ejecuta AUTOMÁTICAMENTE comandos
           →  Resuelve tarea
           →  Retorna JSON

Usuario    ← Recibe resultado en JSON
           ← Puede procesar automáticamente
           ← Si necesita cambios, ejecuta otra tarea

✨ TODO AUTOMÁTICO, SIN INTERVENCIÓN
```

---

---

### OPCIÓN 3 (DAEMON): Flujo de Trabajo Día a Día

#### **Escenario: Múltiples tareas en paralelo en un proyecto**

**SETUP INICIAL (Una sola vez)**
```bash
# El usuario inicia el daemon:
$ ai-dispatcher start-daemon

✓ Daemon iniciado (PID: 12847)
✓ Escuchando en localhost:50051
✓ File watcher activo en /Users/cesar/projects/my-app
✓ Cache inicializado
✓ Listo para recibir tareas
```

**El daemon corre siempre en background. Usuario puede cerrar la terminal.**

---

#### **DÍA 1: Refactorización de autenticación**

**TAREA 1.1: Refactorizar autenticación**
```bash
$ ai-dispatcher exec "Refactoriza el servicio de autenticación para usar JWT"
```

**¿QUÉ PASA INTERNAMENTE?**

```
Usuario request
    ↓
gRPC Client se conecta al daemon (instantáneo, no hay startup)
    ↓
Daemon recibe tarea:
├─ Analiza complejidad (instantáneo, usa cache si existe)
├─ Verifica disponibilidad en tiempo real:
│  ├─ Claude Code: 80% disponible
│  ├─ Codex: 60% disponible
│  └─ Selecciona: Claude Code (mejor opción)
├─ Encolaca tarea en TaskQueue
├─ Spawn: "claude --mcp-server localhost:9999"
├─ Espera resultado
└─ Retorna al usuario

Usuario recibe:
✅ Completado en 8 minutos 32 segundos
```

**¿Mientras tanto (el daemon hace esto)?**

```
Daemon (proceso siempre corriendo):
├─ Task Queue:
│  └─ Tarea 1: "Refactorizar auth" → EJECUTANDO
│     ├─ Claude Code spawned
│     ├─ Utilizando MCP server
│     └─ Progreso: 45%
│
├─ File Watcher:
│  └─ Detecta cambios en:
│     ├─ src/auth/service.js (modificado)
│     ├─ src/auth/middleware.js (modificado)
│     └─ src/auth/jwt-utils.js (creado)
│        └─ Actualiza su index de archivos
│
├─ Metrics:
│  └─ Claude Code: 72% disponible (disminuyó por tarea actual)
│
└─ Cache:
   ├─ Precalcula próximo análisis de complejidad
   ├─ Prefetcha archivos relacionados
   └─ Estima duración de siguiente tarea
```

---

**TAREA 1.2: Mientras se ejecuta, usuario lanza otra**
```bash
# Daemon está ocupado con tarea 1, pero acepta más
$ ai-dispatcher exec "Escribe tests para el nuevo JWT middleware"
```

**Daemon (multitarea):**
```
Daemon Task Queue:
├─ Tarea 1: "Refactorizar auth" → EJECUTANDO (60%)
│  └─ Claude Code utilizando 50% de capacidad
│
├─ Tarea 2: "Escribe tests" → WAITING (en cola)
│  └─ Esperando espacio en Claude Code
│  └─ (O podría ir a Codex si tiene disponibilidad)
│
└─ Load Balancer:
   ├─ Claude Code: 72% (tarea 1 en ejecución)
   ├─ Codex: 60% disponible → Pero tests requieren Claude (especialidad)
   └─ OpenCode: 100% disponible → Pero specialidad es otra

   Decisión: Esperar a que Claude Code termine tarea 1,
            luego ejecutar tarea 2

# Usuario ve:
✋ Tarea encolada. Esperando a que se libere Claude Code...
   (Tarea anterior: 75% completada, falta ~2 min)
```

---

**TAREA 1.3: Cuando termina tarea 1**
```bash
# Tarea 1 completa, daemon automáticamente inicia tarea 2

Daemon:
├─ Tarea 1: COMPLETADA ✓
│  ├─ Archivos modificados: 2
│  ├─ Tests: ✓ PASSED
│  └─ Actualiza cache con nuevos archivos
│
├─ Claude Code ahora tiene: 80% disponible
│
└─ Inicia automáticamente Tarea 2:
   ├─ Spawn: "claude --mcp-server localhost:9999"
   ├─ MCP server REUTILIZA índice actual (no re-analiza)
   ├─ Claude ve cambios previos (file watcher informó)
   ├─ Crea tests para el nuevo JWT middleware
   └─ Ejecuta tests para verificar
```

**Usuario recibe notificación:**
```bash
✅ Tarea 1 completada en 8 min 32 seg
   ├─ Archivos modificados: 2
   └─ Tests: ✓ PASSED

⏳ Iniciando Tarea 2...
   Estimado: 5-7 minutos
```

---

#### **DÍA 2: Más tareas en paralelo**

Usuario quiere hacer múltiples cosas simultáneamente:

```bash
# Terminal 1: Lanza refactorización de routes
$ ai-dispatcher exec "Refactoriza los routes para incluir JWT validation"
✋ Encolado. Esperando disponibilidad...

# Terminal 2: Lanza documentación
$ ai-dispatcher exec "Documenta el nuevo sistema de autenticación con diagrama"
✋ Encolado. Esperando disponibilidad...

# Terminal 3: Lanza análisis
$ ai-dispatcher exec "Analiza el código y sugiere mejoras de performance"
✋ Encolado. Esperando disponibilidad...

# Terminal 4: Ve el status en tiempo real
$ ai-dispatcher status
```

**¿QUÉ MUESTRA EL STATUS?**

```bash
📊 Daemon Status

TaskQueue:
├─ Ejecutando:
│  ├─ [1/4] Refactoriza routes (Claude Code) - 35% - ETA 4 min
│  └─ [2/4] Documenta sistema (Codex) - 10% - ETA 3 min
│
├─ En cola:
│  ├─ [3/4] Analiza y sugiere mejoras (OpenCode) - Esperando...
│  └─ [4/4] Escribe changelog (OpenCode) - Esperando...
│
└─ Herramientas:
   ├─ Claude Code: 35% disponible ⚡ (TRABAJANDO)
   ├─ Codex: 40% disponible ⚡ (TRABAJANDO)
   └─ OpenCode: 100% disponible ✓ (IDLE)

Sistema:
├─ Archivos en cache: 127
├─ Uptime: 2h 34m
└─ Tasks completadas hoy: 8
```

---

**¿CÓMO ITERA SOBRE CAMBIOS?**

```bash
# Tarea 1 se completó, pero el usuario quiere mejorar
$ ai-dispatcher exec "Mejora la validación JWT, añade refresh tokens"

Daemon:
├─ Lee cambios previos (file watcher lo sabe)
├─ Entiende contexto completo
├─ Claude Code automáticamente:
│  ├─ Ve el JWT anterior
│  ├─ Ve los tests que existen
│  ├─ Mejora el sistema de validación
│  ├─ Añade refresh token logic
│  └─ Ejecuta todos los tests
└─ Retorna resultado

✅ Completado en 6 minutos 12 segundos
```

---

**RESUMEN OPCIÓN 3 (DAEMON):**

```
Setup inicial:
$ ai-dispatcher start-daemon    (Una sola vez)
  Daemon corre SIEMPRE en background

Día a día:
$ ai-dispatcher exec "tarea 1"  → INSTANTÁNEO (sin startup)
$ ai-dispatcher exec "tarea 2"  → INSTANTÁNEO (sin startup)
$ ai-dispatcher exec "tarea 3"  → INSTANTÁNEO (sin startup)
$ ai-dispatcher status           → Ve todo en tiempo real

Ventajas:
✓ Sin latencia de startup
✓ Múltiples tareas EN PARALELO
✓ Daemon entiende contexto acumulado
✓ Load balancing automático
✓ File watcher detecta cambios
✓ Puedes cerrar terminal, daemon sigue corriendo
✓ gRPC es ULTRA RÁPIDO

Costo:
✗ Daemon consume RAM siempre
✗ Más complejo de programar
✗ Debugging más difícil
✗ Manejo robusto de procesos necesario
```

---

## 📊 Comparativa: Opción 2 vs Opción 3

### Mismo escenario: Refactorizar autenticación

#### **OPCIÓN 2 (MCP): Sin iteración**
```bash
Tiempo total: 8 min 32 seg (refactorización) + 2 min 15 seg (docs) = ~11 min
Proceso:
1. Usuario lanza tarea 1
2. Espera 8 min 32 seg
3. Usuario lanza tarea 2
4. Espera 2 min 15 seg
5. Si necesita cambios, lanza tarea 3...

Total: Lineal, secuencial
Latencia: Hay startup delay cada vez (~2 seg por tarea)
```

#### **OPCIÓN 3 (DAEMON): Con múltiples tareas**
```bash
Tiempo total:
├─ Tarea 1 (refactor): 8 min 32 seg
├─ Tarea 2 (docs): PARALELO, termina en 2 min 15 seg
│  (Mientras tarea 1 ejecuta, lanzas tarea 2 en otra herramienta)
└─ Total: ~8 min 32 seg (no 11 min, porque hay paralelismo)

Proceso:
1. Start daemon (una sola vez)
2. Usuario lanza: exec "tarea 1" → Instantáneo
3. Usuario lanza: exec "tarea 2" → Instantáneo (mientras 1 ejecuta)
4. Usuario lanza: exec "tarea 3" → Instantáneo (mientras 1 y 2 ejecutan)
5. Ver status en tiempo real: status

Total: Paralelo
Latencia: CERO (no hay startup)
Velocidad: 3-5x más rápido con múltiples tareas
```

---

## 🎯 ¿Cuándo Usar Cada Una?

### Opción 2 (MCP): Tareas Secuenciales
```
Caso: "Necesito hacer varias refactorizaciones, una por una"

Flujo:
1. Refactoriza auth → Espera resultado
2. Refactoriza database → Espera resultado
3. Escribe tests → Espera resultado
4. Documenta cambios → Espera resultado

Ideal para: Workflows lineales, debugging, desarrollo
```

### Opción 3 (DAEMON): Tareas Paralelas
```
Caso: "Necesito hacer varias cosas al mismo tiempo"

Flujo:
daemon start
├─ Refactoriza auth (Claude) - Ejecutando
├─ Escribe tests (Codex) - Ejecutando PARALELO
├─ Documenta (OpenCode) - Ejecutando PARALELO
└─ Lanza análisis de performance (Claude) - Esperando

Ideal para: Producción, CI/CD, múltiples herramientas
```

---

## ✅ Siguiente Paso Recomendado

### Si eres **Developer Individual** en Desarrollo Local:
**→ Empieza con: OPCIÓN 1 (Tmux) → OPCIÓN 2 (MCP)**
1. Opción 1: Valida que el routing funciona
2. Opción 2: Automatización pura para iteraciones rápidas

### Si trabajas en **Equipo o Producción**:
**→ Empieza con: OPCIÓN 5 (Approval Gate)**
1. Automatización con control = seguridad
2. Luego puedes agregar Opción 2 para desarrollo
3. Opción 4 (Watch) cuando tengas confianza

### Si quieres lo **Más Completo Posible**:
**→ Roadmap Completo: 1 → 2 → 5 → 4 → 3**
```
Fase 1: Opción 1 (Tmux)         - 1 semana - MVP
Fase 2: Opción 2 (MCP)          - 1 semana - Automatización
Fase 3: Opción 5 (Approval)     - 1 semana - Control seguro
Fase 4: Opción 4 (Watch)        - 2 semanas - Vigilancia automática
Fase 5: Opción 3 (Daemon)       - 2 semanas - Enterprise
Total: ~1 mes para herramienta completa
```

---

## 🎯 TL;DR - Decisión Rápida

| Necesidad | Opción | Tiempo |
|-----------|--------|--------|
| "Quiero ver qué pasa" | 1 (Tmux) | 1 sem |
| "Quiero automatización rápida" | 2 (MCP) | 1 sem |
| "Quiero control en producción" | 5 (Approval) | 1 sem |
| "Quiero que cuide mi proyecto" | 4 (Watch) | 2 sem |
| "Quiero herramienta enterprise" | 3 (Daemon) | 2-3 sem |
| "Quiero TODO" | 1→2→5→4→3 | 1 mes |

---

> 📝 **Última actualización:** 2026-01-23
> 👤 **Autor:** ai-dispatcher arquitectura
> 🔗 **Relacionado:** CLAUDE.md, IMPLEMENTATION_SUMMARY.md
> 📊 **Opciones:** 5 (TMux, MCP, Daemon, Watch Mode, Approval Gate)
