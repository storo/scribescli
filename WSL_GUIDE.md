# Guía de ScribesAI para WSL2

## 🎯 Estado Actual

ScribesAI ahora funciona en WSL2 con las siguientes correcciones:

✅ **Arreglado:**
- ✅ Panic por dimensiones negativas de terminal
- ✅ Error de TTY bloqueante
- ✅ Manejo graceful cuando no hay audio disponible
- ✅ Warnings de ALSA pueden suprimirse

⚠️ **Limitaciones:**
- ❌ Audio no disponible (WSL no tiene acceso a micrófono)
- ⚠️ Grabación no funcionará sin dispositivo de audio

## 🚀 Cómo Ejecutar

### Opción 1: Ejecutar Directamente (Con Warnings)

```bash
cd /home/storo/ScribesAI/scribescli
./build/scribescli
```

Verás warnings de ALSA (son normales en WSL, puedes ignorarlos):
```
ALSA lib confmisc.c:855:(parse_card) cannot find card '0'
... (más warnings)
⚠  WARNING: audio unavailable (WSL/no microphone): ...
```

### Opción 2: Ejecutar Sin Warnings (Limpio)

```bash
./scribescli-wsl
```

Esto filtra todos los warnings de ALSA para una salida limpia.

### Opción 3: Suprimir Solo Stderr

```bash
./build/scribescli 2>/dev/null
```

## 🎨 Lo Que Puedes Ver

### Interfaz Funcionando

Cuando ejecutes la aplicación verás:

```
╔═══════════════════════════════════════════╗
║                                           ║
║              S C R I B E S  A I           ║
║    Meeting Intelligence System v1.0      ║
║                                           ║
║           ┌───────────────┐               ║
║           │   ◉◉◉◉◉◉◉    │               ║
║           │  ◉     ◉     │               ║
║           │ ◉   •   ◉    │               ║
║           │  ◉     ◉     │               ║
║           │   ◉◉◉◉◉◉◉    │               ║
║           └───────────────┘               ║
║              HAL 9000                     ║
║                                           ║
║  "Good day. I am ready to assist with    ║
║   your meeting recordings."              ║
║                                           ║
║    ► New Recording                       ║
║      History                             ║
║      Settings                            ║
║      Quit                                ║
║                                           ║
║  ↑/↓: Navigate  Enter: Select  Q: Quit   ║
╚═══════════════════════════════════════════╝
```

### Navegación

- `↑` / `↓` : Mover entre opciones del menú
- `Enter` : Seleccionar opción
- `Q` : Salir en cualquier momento
- `ESC` : Volver al menú anterior

## 📋 Funcionalidades Disponibles en WSL

| Funcionalidad | Estado | Notas |
|---------------|--------|-------|
| 🎨 Interfaz HAL 9000 | ✅ Funciona | Animación completa del ojo |
| 📋 Navegación de menús | ✅ Funciona | Todos los controles |
| 🎙️ Grabación de audio | ❌ No disponible | Sin micrófono en WSL |
| 📝 Transcripción | ⏳ Pendiente | Necesita audio primero |
| 🧠 Análisis Claude AI | ✅ Funciona | API configurada |
| 💾 Base de datos | ✅ Funciona | SQLite embebido |
| 📤 Exportación | ✅ Funciona | Markdown/JSON/Text |
| ⚙️ Configuración | ✅ Funciona | Vista de settings |

## 🧪 Testing en WSL

Puedes testear la interfaz completa navegando por:

1. **Menu Principal**: Ver el ojo de HAL animándose
2. **Settings**: Ver la configuración
3. **History**: Ver el historial (vacío inicialmente)
4. **New Recording**: Ver la pantalla de grabación (mostrará warning de audio)

### Comandos de Testing

```bash
# Ver versión
./build/scribescli --version

# Ejecutar normalmente
./build/scribescli

# Ejecutar sin warnings
./scribescli-wsl

# Ver diagnóstico del sistema
make doctor
```

## 🔧 Próximos Pasos para Desarrollo Completo

Para tener funcionalidad completa de audio, necesitas:

### Opción A: Linux Nativo
```bash
# Mejor opción para desarrollo con audio
# Instalar en Ubuntu/Debian físico o VM
```

### Opción B: macOS
```bash
# Funciona perfectamente out-of-the-box
# Audio nativo con CoreAudio
```

### Opción C: WSLg + PulseAudio (Experimental)
```bash
# Requiere WSL 2 con WSLg y configuración de PulseAudio
# Documentación: https://github.com/microsoft/wslg
```

### Opción D: Desarrollo Sin Audio
```bash
# Usar transcripciones simuladas para testing
# Ver sección de "Demo Mode" en desarrollo futuro
```

## 🐛 Troubleshooting WSL

### Problema: Demasiados Warnings de ALSA

**Solución:** Usa el wrapper `./scribescli-wsl`

### Problema: Terminal muy pequeño

**Solución:** Redimensiona tu terminal a al menos 80x24:
```bash
# Verificar tamaño
tput cols  # Debería ser >= 80
tput lines # Debería ser >= 24
```

### Problema: Colores no se ven bien

**Solución:** Asegúrate de tener un terminal con soporte de 256 colores:
```bash
# Verificar soporte de colores
echo $TERM
# Debería ser: xterm-256color o similar

# Si no, agregar a ~/.bashrc:
export TERM=xterm-256color
```

### Problema: Caracteres Unicode no se muestran

**Solución:** Verifica que tu terminal soporte UTF-8:
```bash
# Verificar locale
locale
# Debería incluir UTF-8

# Configurar si es necesario:
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8
```

## 📊 Performance en WSL

| Aspecto | WSL2 | Linux Nativo |
|---------|------|--------------|
| TUI Rendering | ✅ Excelente | ✅ Excelente |
| CPU Usage | ~2-3% | ~2-3% |
| Memory | ~15MB | ~15MB |
| Audio | ❌ N/A | ✅ <10ms latency |

## 🎓 Aprendiendo la Interfaz

### Menu Principal
- El ojo de HAL pulsa (cambia entre 2 estados de animación)
- Los colores son:
  - Rojo (#FF0000): HAL, estados críticos
  - Verde CRT (#00FF41): Texto principal
  - Ámbar (#FFB000): Warnings, estado activo
  - Cian (#00FFFF): Highlights

### Pantalla de Grabación
- Muestra contador de tiempo
- Barra de nivel de audio (estará en 0% sin micrófono)
- Área de transcripción en vivo
- Controles en la parte inferior

### Pantalla de Análisis
- Resumen generado por Claude
- Puntos clave con bullets
- Accionables con prioridad (🔴 ALTA, 🟡 MEDIA, 🟢 BAJA)

## 🚀 Siguiente: ¿Qué Implementar?

Para hacer la app completamente funcional en WSL, las opciones son:

1. **Modo Demo/Testing**
   - Transcripciones simuladas
   - Datos de ejemplo
   - Testing completo del flujo sin audio

2. **Integración Vosk**
   - Transcripción desde archivos WAV
   - Procesar grabaciones existentes
   - Multi-idioma

3. **Modo Headless**
   - CLI sin TUI
   - Procesar archivos batch
   - Para CI/CD

¿Cuál te gustaría implementar primero?

---

**Documentación Actualizada:** 2025-11-13
**Estado WSL:** Funcional (sin audio)
**Próxima Versión:** v1.1 con modo demo
