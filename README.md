# ScribesAI 🤖

> Meeting Intelligence System powered by HAL 9000

ScribesAI es una aplicación de terminal (TUI) para grabar reuniones, transcribirlas automáticamente y generar análisis inteligentes con IA, todo con una interfaz inspirada en HAL 9000 de *2001: A Space Odyssey*.

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Go](https://img.shields.io/badge/Go-1.24-00ADD8.svg)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS-lightgrey.svg)

## 💻 Plataformas Soportadas

ScribesAI está diseñado específicamente para sistemas **Linux** y **macOS**:

| Plataforma | Arquitectura | Estado | Notas |
|------------|--------------|--------|-------|
| 🐧 Linux   | x86_64 (amd64) | ✅ Completamente soportado | Ubuntu, Debian, Fedora, Arch |
| 🍎 macOS   | Intel (amd64) | ✅ Completamente soportado | macOS 11+ |
| 🍎 macOS   | Apple Silicon (arm64) | ✅ Completamente soportado | M1/M2/M3 nativo |

**Nota**: Windows no está soportado debido a la complejidad de CGO + PortAudio.

## ✨ Características

- **🎙️ Grabación de Audio**: Captura audio en tiempo real con indicadores de nivel
- **📝 Transcripción Automática**: Transcripción multi-idioma con Vosk (offline)
- **🧠 Análisis con IA**: Resumen, puntos clave y accionables generados por Claude AI
- **🎨 Interfaz HAL 9000**: TUI retro-futurista con animación del ojo icónico
- **💾 Historial Persistente**: Base de datos SQLite para todas las grabaciones
- **📤 Exportación**: Exporta a Markdown, JSON o texto plano
- **🌍 Multi-idioma**: Soporte para español, inglés y más
- **🔧 Cross-platform**: Build nativo para Linux y macOS

## 🎯 Demo

```
╔═══════════════════════════════════════════════════════════════════╗
║  S C R I B E S  A I                    [●] 00:05:23  [■] STOP     ║
╠═══════════════════════════════════════════════════════════════════╣
║                     ┌───────────────┐                              ║
║                     │    ◉◉◉◉◉◉◉   │  🎤 RECORDING               ║
║                     │   ◉     ◉    │                              ║
║                     │  ◉   ●   ◉   │  "Recording in progress...  ║
║                     │   ◉     ◉    │   monitoring all inputs."   ║
║                     │    ◉◉◉◉◉◉◉   │                              ║
║                     └───────────────┘                              ║
║  AUDIO LEVEL: ████████████████░░░░░░░░  75%                       ║
╚═══════════════════════════════════════════════════════════════════╝
```

## 🚀 Instalación

### Método 1: Instalación Automática (Recomendado)

```bash
# Clonar el repositorio
git clone https://github.com/storo/ScribesAI.git
cd ScribesAI/scribescli

# Ejecutar script de instalación
./install.sh

# Configurar tu API key cuando se solicite
# O editar manualmente:
nano .env

# Ejecutar
./scribescli
```

### Método 2: Con Makefile

```bash
cd ScribesAI/scribescli

# Ver todas las opciones disponibles
make help

# Verificar sistema
make doctor

# Setup completo (instala PortAudio si es necesario)
make setup

# Ejecutar
make run
```

### Método 3: Manual

#### Paso 1: Prerrequisitos

**Go 1.24+**
```bash
go version  # Verificar instalación
```

**PortAudio** (requerido para grabación de audio)

<details>
<summary>🐧 Linux (Ubuntu/Debian)</summary>

```bash
sudo apt-get update
sudo apt-get install portaudio19-dev
```
</details>

<details>
<summary>🐧 Linux (Fedora/RHEL)</summary>

```bash
sudo dnf install portaudio-devel
```
</details>

<details>
<summary>🐧 Linux (Arch)</summary>

```bash
sudo pacman -S portaudio
```
</details>

<details>
<summary>🍎 macOS</summary>

```bash
# Requiere Homebrew (https://brew.sh)
brew install portaudio
```
</details>

**Claude API Key**
- Obtener en: https://console.anthropic.com/
- Modelo usado: Claude Sonnet 4.5

#### Paso 2: Instalación

```bash
# Clonar el repositorio
git clone https://github.com/storo/ScribesAI.git
cd ScribesAI/scribescli

# Configurar variables de entorno
cp .env.example .env
nano .env  # Agregar tu ANTHROPIC_API_KEY

# Crear directorios
mkdir -p data models data/exports

# Instalar dependencias Go
go mod download

# Compilar
go build -o scribescli ./cmd/scribescli

# Ejecutar
./scribescli
```

### Instalación Global (Opcional)

```bash
# Instalar en /usr/local/bin
make install

# Ahora puedes ejecutar desde cualquier lugar
scribescli

# Para desinstalar
make uninstall
```

## 🎮 Uso

### Comandos Principales

```bash
# Iniciar la aplicación
./scribescli

# Ver versión
./scribescli --version
```

### Atajos de Teclado

#### Menú Principal
- `↑/↓` o `k/j` - Navegar
- `Enter` - Seleccionar
- `Q` - Salir

#### Durante Grabación
- `R` - Iniciar grabación
- `S` - Detener grabación
- `P` - Pausar/Reanudar
- `B` o `ESC` - Volver al menú
- `Q` - Salir

#### Pantalla de Análisis
- `E` - Exportar
- `T` - Ver transcripción completa
- `B` - Volver al menú

## 📁 Estructura del Proyecto

```
scribescli/
├── cmd/
│   └── scribescli/
│       └── main.go                 # Entry point
├── internal/
│   ├── audio/
│   │   ├── recorder.go             # Grabación de audio
│   │   └── wav.go                  # Manejo de archivos WAV
│   ├── transcription/
│   │   └── vosk.go                 # Integración Vosk (TODO)
│   ├── ai/
│   │   └── claude.go               # Cliente Claude API
│   ├── tui/
│   │   ├── model.go                # Modelo Bubble Tea
│   │   ├── views.go                # Vistas de la aplicación
│   │   ├── styles.go               # Estilos HAL 9000
│   │   └── haleye.go               # Animación del ojo
│   ├── storage/
│   │   └── database.go             # Base de datos SQLite
│   └── export/
│       └── exporter.go             # Exportadores
├── pkg/
│   └── models/
│       └── recording.go            # Modelos de datos
├── data/                           # Grabaciones y DB
├── models/                         # Modelos de transcripción
├── .env                            # Configuración
├── .env.example                    # Ejemplo de configuración
├── go.mod
├── go.sum
└── README.md
```

## ⚙️ Configuración

### Variables de Entorno

Editar el archivo `.env`:

```bash
# Claude API Key (requerido)
ANTHROPIC_API_KEY=sk-ant-api03-...

# Audio Settings
SAMPLE_RATE=16000
CHANNELS=1

# Database Path
DB_PATH=./data/scribescli.db

# Vosk Model Path (opcional)
# VOSK_MODEL_PATH=./models/vosk-model-small-es-0.42
```

### Modelos de Transcripción

ScribesAI utiliza Vosk para transcripción offline. Los modelos se descargan automáticamente en el primer uso, pero puedes descargarlos manualmente:

1. Visitar https://alphacephei.com/vosk/models
2. Descargar modelo (recomendado: `vosk-model-small-es-0.42` para español)
3. Extraer en `./models/`
4. Configurar `VOSK_MODEL_PATH` en `.env`

## 🎨 Personalización

### Colores HAL 9000

Los colores están definidos en `internal/tui/styles.go`:

```go
ColorHALRed     = "#FF0000"  // Rojo icónico de HAL
ColorCRTGreen   = "#00FF41"  // Verde de terminal CRT
ColorAmber      = "#FFB000"  // Ámbar para warnings
ColorCyan       = "#00FFFF"  // Cian para highlights
```

### Mensajes de HAL

Personalizar en `internal/tui/haleye.go` en la función `GetHALQuote()`.

## 📊 Exportación

### Formatos Soportados

1. **Markdown** (.md) - Para documentación y Notion/Obsidian
2. **JSON** (.json) - Para integración con otras herramientas
3. **Texto Plano** (.txt) - Para máxima compatibilidad

### Ejemplo de Exportación

```bash
# Desde la interfaz: Presionar 'E' en la pantalla de análisis
# Los archivos se guardan en ./data/exports/
```

## 🛠️ Desarrollo

### Ejecutar en Modo Desarrollo

```bash
go run ./cmd/scribescli
```

### Ejecutar Tests

```bash
go test ./...
```

### Build para Producción

**Usando Makefile (Recomendado)**

```bash
# Build para tu plataforma actual
make build

# Build solo para Linux
make build-linux

# Build solo para macOS (Intel + Apple Silicon)
make build-mac

# Build para todas las plataformas soportadas
make build-all

# Los binarios se generan en: build/
```

**Build Manual**

```bash
# Linux (x86_64)
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o scribescli-linux ./cmd/scribescli

# macOS (Intel)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o scribescli-mac-intel ./cmd/scribescli

# macOS (Apple Silicon / M1/M2/M3)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o scribescli-mac-arm ./cmd/scribescli
```

**Release con Checksums**

```bash
# Crear release completo con checksums SHA256
make release

# Genera:
# - build/release/scribescli-v1.0.0-linux-amd64.tar.gz
# - build/release/scribescli-v1.0.0-darwin-amd64.tar.gz
# - build/release/scribescli-v1.0.0-darwin-arm64.tar.gz
# - build/release/checksums.txt
```

## 🤝 Contribuir

Las contribuciones son bienvenidas! Por favor:

1. Fork el proyecto
2. Crear una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abrir un Pull Request

## 📝 Roadmap

### Completado

- [x] Grabación de audio básica
- [x] Interfaz TUI con animación HAL
- [x] Integración Claude API
- [x] Base de datos SQLite
- [x] Exportación a múltiples formatos

### En Desarrollo

- [ ] Integración Vosk para transcripción
- [ ] Diarización de speakers
- [ ] Detección automática de idioma

### Planificado

- [ ] Modo streaming para reuniones largas
- [ ] Exportación a PDF
- [ ] Integración con calendarios
- [ ] Síntesis de voz (HAL habla)
- [ ] Plugin system
- [ ] Web UI companion

## 🐛 Troubleshooting

### Verificación del Sistema

```bash
# Ejecutar diagnóstico completo
make doctor

# Esto verifica:
# - Sistema operativo (Linux/macOS)
# - Instalación de Go
# - Instalación de PortAudio
# - Configuración de .env
# - Directorios necesarios
```

### Error: "Package portaudio-2.0 was not found"

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install portaudio19-dev

# Fedora
sudo dnf install portaudio-devel

# Arch
sudo pacman -S portaudio
```

**macOS:**
```bash
# Instalar Homebrew si no lo tienes
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Instalar PortAudio
brew install portaudio
```

### Error: "ANTHROPIC_API_KEY environment variable not set"

```bash
# Crear archivo .env con tu API key
cp .env.example .env
nano .env  # Editar y agregar tu clave

# Verificar que esté correctamente configurada
grep ANTHROPIC_API_KEY .env
```

### Audio no se graba

**Linux:**
```bash
# 1. Verificar dispositivos de audio
arecord -l

# 2. Verificar permisos del usuario
groups $USER | grep audio

# 3. Agregar usuario al grupo audio si es necesario
sudo usermod -a -G audio $USER

# 4. Reiniciar sesión para aplicar cambios
```

**macOS:**
```bash
# 1. Verificar dispositivos de audio
system_profiler SPAudioDataType

# 2. Verificar permisos de micrófono
# System Preferences → Security & Privacy → Privacy → Microphone
# Asegurarse de que Terminal/iTerm tenga acceso

# 3. Si usas macOS Catalina o superior, puede requerir permisos adicionales
```

### Build falla con errores de CGO

**Problema**: `cgo: C compiler "gcc" not found`

**Solución Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install build-essential

# Fedora
sudo dnf groupinstall "Development Tools"

# Arch
sudo pacman -S base-devel
```

**Solución macOS:**
```bash
# Instalar Xcode Command Line Tools
xcode-select --install
```

### Rendimiento lento en macOS

Si experimentas lentitud en macOS con Apple Silicon:

```bash
# Asegurarse de usar la versión ARM nativa
make build-mac

# Ejecutar el binario arm64
./build/scribescli-darwin-arm64
```

### Cross-compilation no funciona

**Nota**: Cross-compilation con CGO (requerido por PortAudio) puede ser complicado.

**Recomendación**:
- Para Linux: Compilar en Linux
- Para macOS: Compilar en macOS

**Alternativa**: Usar Docker para compilación:
```bash
# TODO: Agregar Dockerfile para compilación cross-platform
```

## 📄 Licencia

MIT License - ver archivo LICENSE para más detalles.

## 👏 Agradecimientos

- **Anthropic** - Claude AI API
- **Charm.sh** - Bubble Tea, Lipgloss y Bubbles
- **Alpha Cephei** - Vosk Speech Recognition
- **Stanley Kubrick** - Inspiración HAL 9000

---

*"I'm sorry Dave, I'm afraid I can't stop recording."* - HAL 9000 (probably)

Made with ❤️ by [@storo](https://github.com/storo)
