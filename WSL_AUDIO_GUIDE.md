# Guía de Audio en WSL2 - ScribesAI

Esta guía proporciona instrucciones completas para configurar audio en WSL2, permitiendo que ScribesAI grabe reuniones desde tu entorno de desarrollo de Windows Subsystem for Linux.

## Tabla de Contenidos

- [Visión General](#visión-general)
- [Arquitectura](#arquitectura)
- [Inicio Rápido](#inicio-rápido)
- [Configuración Manual](#configuración-manual)
- [Troubleshooting](#troubleshooting)
- [Alternativas y Variaciones](#alternativas-y-variaciones)
- [Seguridad](#seguridad)
- [Rendimiento](#rendimiento)
- [FAQ](#faq)
- [Referencias](#referencias)

---

## Visión General

### ¿Por qué es necesaria esta configuración?

WSL2 no tiene acceso directo al hardware de audio de Windows. Para grabar audio en ScribesAI desde WSL2, necesitamos crear un "puente" usando PulseAudio:

```
Micrófono → Windows → PulseAudio → TCP:4713 → WSL2 → ScribesAI
```

### ¿Qué hace esta solución?

1. **En Windows**: Instala PulseAudio que captura audio del micrófono
2. **Puente de red**: Expone PulseAudio en TCP puerto 4713
3. **En WSL2**: Configura aplicaciones para conectarse al PulseAudio de Windows
4. **ScribesAI**: Usa PortAudio para grabar a través de PulseAudio

### Requisitos

- ✅ Windows 10/11 con WSL2 instalado
- ✅ Distribución Linux en WSL2 (Ubuntu, Debian, etc.)
- ✅ PowerShell con permisos de administrador
- ✅ Micrófono conectado a Windows
- ✅ Acceso a internet (para descargar PulseAudio)

---

## Arquitectura

### Componentes

```
┌─────────────────────────────────────────────────────────────┐
│                         WINDOWS HOST                         │
│                                                              │
│  ┌──────────────┐      ┌──────────────────────────────┐    │
│  │  Micrófono   │──────▶│  PulseAudio Server          │    │
│  │              │      │  - Captura audio             │    │
│  └──────────────┘      │  - Puerto 4713 (TCP)         │    │
│                        │  - ACL: 172.16.0.0/12        │    │
│                        └──────────┬───────────────────┘    │
│                                   │                         │
└───────────────────────────────────┼─────────────────────────┘
                                    │ TCP/IP
                  ┌─────────────────┴────────────────┐
                  │      Virtual Network Switch       │
                  │      (172.16.0.0/12)             │
                  └─────────────────┬────────────────┘
                                    │
┌───────────────────────────────────┼─────────────────────────┐
│                                   │                         │
│                             WSL2 Instance                    │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  PULSE_SERVER=tcp:172.x.x.x                          │  │
│  │  (Configurado en ~/.bashrc)                          │  │
│  └──────────────────┬───────────────────────────────────┘  │
│                     │                                       │
│  ┌──────────────────▼───────────────────────────────────┐  │
│  │  ScribesAI                                           │  │
│  │  ├── PortAudio (cliente)                            │  │
│  │  ├── Graba audio remoto                             │  │
│  │  └── Procesa localmente                             │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Flujo de Datos

1. **Captura**: Windows captura audio del micrófono
2. **PulseAudio**: Procesa audio y escucha en TCP:4713
3. **Red Virtual**: WSL2 se conecta a Windows via IP del host
4. **PortAudio**: Cliente de audio en WSL2 se conecta a PulseAudio
5. **ScribesAI**: Graba y procesa audio

### Puertos y Protocolos

- **Puerto 4713**: PulseAudio native protocol (TCP)
- **Puerto 4714**: EsounD protocol (TCP, legacy)
- **ACL**: `172.16.0.0/12` - Rango de IPs WSL2

---

## Inicio Rápido

### Opción 1: Asistente Interactivo (TUI)

Si ScribesAI ya está instalado:

```bash
./scribescli setup-audio
```

El asistente te guiará paso a paso.

### Opción 2: Script Automático

Desde el directorio del proyecto:

```bash
./scripts/setup-wsl-audio.sh
```

Esto:
1. ✅ Detecta tu entorno WSL2
2. ✅ Instala dependencias en WSL
3. ✅ Genera script PowerShell
4. ✅ Te guía a ejecutar en Windows
5. ✅ Prueba la conexión
6. ✅ Configura ~/.bashrc

### Opción 3: Validación Rápida

Para verificar si ya está configurado:

```bash
./scripts/test-wsl-audio.sh
```

---

## Configuración Manual

Si prefieres configurar manualmente o los scripts fallan:

### Paso 1: Instalar PulseAudio en WSL (Opcional)

Aunque no necesitas el servidor PulseAudio en WSL, las herramientas de cliente son útiles:

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install -y pulseaudio pulseaudio-utils
```

**Fedora:**
```bash
sudo dnf install -y pulseaudio pulseaudio-utils
```

**Arch:**
```bash
sudo pacman -S pulseaudio
```

### Paso 2: Detectar IP de Windows

Desde WSL2, ejecuta:

```bash
ip route show default | awk '{print $3}'
```

O alternativamente:

```bash
grep nameserver /etc/resolv.conf | awk '{print $2}'
```

Guarda esta IP (ej: `172.28.176.1`).

### Paso 3: Instalar PulseAudio en Windows

#### A. Descargar PulseAudio

1. Descarga desde: https://www.freedesktop.org/software/pulseaudio/releases/pulseaudio-1.1.zip
2. Extrae a `C:\Program Files\pulse\`

#### B. Configurar PulseAudio

**Edita** `C:\Program Files\pulse\etc\pulse\default.pa`:

Agrega al inicio del archivo:

```
# ScribesAI - WSL2 Configuration
load-module module-native-protocol-tcp port=4713 auth-ip-acl=172.16.0.0/12
load-module module-esound-protocol-tcp port=4714 auth-ip-acl=172.16.0.0/12
```

**Modifica** la línea de `module-waveout`:

```
# Antes:
# load-module module-waveout sink_name=output source_name=input

# Después:
load-module module-waveout sink_name=output source_name=input record=0
```

**Edita** `C:\Program Files\pulse\etc\pulse\daemon.conf`:

Busca y modifica (o agrega si no existe):

```
exit-idle-time = -1
```

#### C. Configurar Firewall de Windows

Abre PowerShell **como Administrador**:

```powershell
New-NetFirewallRule -DisplayName "PulseAudio for WSL2" `
                    -Direction Inbound `
                    -Program "C:\Program Files\pulse\bin\pulseaudio.exe" `
                    -Action Allow `
                    -Profile Private,Domain `
                    -Protocol TCP `
                    -LocalPort 4713,4714
```

#### D. Iniciar PulseAudio

En PowerShell:

```powershell
& "C:\Program Files\pulse\bin\pulseaudio.exe"
```

Para que inicie automáticamente:

1. Presiona `Win + R`
2. Escribe: `shell:startup`
3. Crea un archivo `pulseaudio-wsl.bat` con:

```batch
@echo off
start /B "" "C:\Program Files\pulse\bin\pulseaudio.exe"
```

### Paso 4: Configurar WSL2

**Edita** `~/.bashrc`:

```bash
nano ~/.bashrc
```

**Agrega** al final:

```bash
# ScribesAI - Audio en WSL2
if grep -qi microsoft /proc/version; then
    # Detectar IP de Windows (host)
    export PULSE_SERVER=tcp:$(ip route show default | awk '{print $3}')
fi
```

**Aplica** la configuración:

```bash
source ~/.bashrc
```

### Paso 5: Verificar

**Verifica** variable de entorno:

```bash
echo $PULSE_SERVER
# Debe mostrar: tcp:172.x.x.x
```

**Prueba** conexión al puerto:

```bash
timeout 2 bash -c "echo > /dev/tcp/$(echo $PULSE_SERVER | cut -d: -f2)/4713" && echo "✓ Conectado" || echo "✗ Error"
```

**Lista** dispositivos de audio:

```bash
pactl list sources short
```

**Ejecuta** ScribesAI:

```bash
./scribescli
```

---

## Troubleshooting

### Error: "No se puede conectar a PulseAudio"

#### Causa 1: PulseAudio no está corriendo

**Verificar** proceso:

```powershell
# En PowerShell (Windows)
Get-Process -Name pulseaudio
```

**Solución**: Inicia PulseAudio:

```powershell
& "C:\Program Files\pulse\bin\pulseaudio.exe"
```

#### Causa 2: Firewall bloqueando

**Verificar** regla de firewall:

```powershell
Get-NetFirewallRule -DisplayName "PulseAudio for WSL2"
```

**Solución**: Recrea la regla (ver Paso 3C arriba).

#### Causa 3: IP incorrecta

**Verificar** que `PULSE_SERVER` apunte a la IP correcta:

```bash
# En WSL2
echo $PULSE_SERVER
ip route show default
```

**Solución**: Recarga bashrc o reconéctate:

```bash
source ~/.bashrc
```

#### Causa 4: Puerto bloqueado

**Probar** conexión TCP:

```bash
nc -zv $(echo $PULSE_SERVER | cut -d: -f2) 4713
```

**Solución**: Verifica que PulseAudio esté escuchando:

```powershell
# En PowerShell
netstat -ano | findstr :4713
```

### Error: "PortAudio initialization failed"

#### Causa 1: Módulo TCP no cargado

**Verificar** configuración de PulseAudio:

```bash
pactl info
```

**Solución**: Verifica que `default.pa` tenga `module-native-protocol-tcp`.

#### Causa 2: Permisos de ACL

**Verificar** ACL en `default.pa`:

```
auth-ip-acl=172.16.0.0/12
```

**Solución**: Asegúrate que la IP de WSL2 esté en ese rango:

```bash
# En WSL2
hostname -I
# Debe estar en 172.16.x.x - 172.31.x.x
```

### Error: "ALSA lib ... Unknown PCM"

**Es normal** en WSL2. ALSA intenta acceder a hardware local que no existe.

**Solución**: No requiere acción. PortAudio usará PulseAudio como backend.

**Opcional** - Silenciar mensajes:

```bash
export ALSA_CONFIG_PATH=/dev/null
```

### Audio con latencia alta

**Causa**: Buffer grande o red lenta.

**Solución**: Optimiza configuración de PulseAudio en `daemon.conf`:

```
default-fragments = 4
default-fragment-size-msec = 10
```

**Reinicia** PulseAudio en Windows.

### Audio entrecortado

**Causa**: CPU saturada o sample rate incompatible.

**Solución**:

1. Verifica sample rate en `daemon.conf`:
   ```
   default-sample-rate = 16000
   ```

2. Reduce carga de CPU en Windows

3. Verifica que ScribesAI use el mismo sample rate (16kHz)

### No aparecen dispositivos de grabación

**Verificar** módulo waveout:

```bash
pactl list modules short | grep waveout
```

**Solución**: Asegúrate que `default.pa` tenga:

```
load-module module-waveout sink_name=output source_name=input record=0
```

### PulseAudio se cierra solo

**Causa**: `exit-idle-time` configurado.

**Solución**: En `daemon.conf`:

```
exit-idle-time = -1
```

**Reinicia** PulseAudio.

---

## Alternativas y Variaciones

### Autenticación con Cookie (Más Segura)

En lugar de ACL abierta, usa cookie de autenticación:

#### Windows: Generar Cookie

```powershell
# En PowerShell
[byte[]]$cookie = 1..256 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }
[System.IO.File]::WriteAllBytes("$env:USERPROFILE\.pulse-cookie", $cookie)
```

#### Windows: Configurar PulseAudio

En `default.pa`, cambia:

```
# Antes:
load-module module-native-protocol-tcp port=4713 auth-ip-acl=172.16.0.0/12

# Después:
load-module module-native-protocol-tcp port=4713 auth-cookie=C:\Users\TuUsuario\.pulse-cookie
```

#### WSL2: Copiar Cookie

```bash
cp /mnt/c/Users/TuUsuario/.pulse-cookie ~/.pulse-cookie
```

#### WSL2: Configurar Variable

En `~/.bashrc`:

```bash
export PULSE_SERVER=tcp:$(ip route show default | awk '{print $3}')
export PULSE_COOKIE=$HOME/.pulse-cookie
```

### Usando socat (Alternativa sin PulseAudio en Windows)

Si tienes problemas con PulseAudio en Windows, puedes usar `socat` para redirigir audio:

**No recomendado** - Más complejo y menos estable.

### PipeWire (Futuro)

PipeWire es el sucesor de PulseAudio. En el futuro, WSL2 podría soportarlo nativamente.

**Estado actual**: Experimental, no recomendado para producción.

---

## Seguridad

### Riesgos

1. **Exposición de red**: PulseAudio escucha en todas las interfaces
2. **Sin autenticación fuerte**: ACL solo filtra por IP
3. **Audio no cifrado**: TCP plano sin TLS

### Mitigaciones

#### 1. Limitar a WSL2

En `default.pa`, usa ACL restrictiva:

```
auth-ip-acl=172.16.0.0/12
```

Esto solo permite conexiones desde WSL2, no desde otras máquinas.

#### 2. Usar Autenticación con Cookie

Ver sección [Autenticación con Cookie](#autenticación-con-cookie-más-segura).

#### 3. Firewall Restringido

Configura regla de firewall solo para red privada:

```powershell
-Profile Private
```

**No uses** `Public` o `Domain` si no es necesario.

#### 4. No Expongas a Internet

**Nunca** permitas conexiones desde fuera de tu red local.

**Verifica** que PulseAudio no esté accesible externamente:

```bash
# Desde otra máquina
nc -zv TU_IP_PUBLICA 4713
# Debe fallar
```

### Mejores Prácticas

1. ✅ Usa cookie de autenticación en entornos multiusuario
2. ✅ Mantén PulseAudio actualizado
3. ✅ Revisa logs de conexión regularmente
4. ✅ Deshabilita cuando no estés usando ScribesAI
5. ✅ Usa VPN si conectas desde red pública

---

## Rendimiento

### Latencia Típica

- **Nativa (Linux/macOS)**: 10-30ms
- **WSL2 con PulseAudio**: 50-150ms

**Nota**: Para reuniones pregrabadas, la latencia no es crítica.

### Consumo de Recursos

**Windows (PulseAudio):**
- CPU: 1-3%
- RAM: 20-50 MB
- Red: ~200 KB/s durante grabación

**WSL2 (ScribesAI):**
- CPU: 5-10% (grabación) + 20-40% (transcripción)
- RAM: 100-300 MB
- Disco: Variable según duración

### Optimizaciones

#### Reducir Latencia

En `daemon.conf`:

```
default-fragments = 2
default-fragment-size-msec = 5
```

**Costo**: Mayor uso de CPU.

#### Reducir Uso de CPU

```
default-sample-rate = 16000
resample-method = trivial
```

**Costo**: Menor calidad de audio (suficiente para voz).

#### Reducir Uso de Red

Usa sample rate más bajo (16kHz en lugar de 44.1kHz):

```
default-sample-rate = 16000
```

**ScribesAI ya usa 16kHz por defecto**.

---

## FAQ

### ¿Funciona con todas las distribuciones de WSL2?

**Sí**, siempre que tengas:
- WSL2 (no WSL1)
- Kernel Linux 4.19+
- Red virtual funcionando

Probado en:
- ✅ Ubuntu 20.04, 22.04, 24.04
- ✅ Debian 11, 12
- ✅ Arch Linux
- ✅ Fedora 38+

### ¿Necesito tener PulseAudio instalado en WSL?

**No necesitas el servidor**, pero las **herramientas de cliente** (pactl, pacmd) son útiles para diagnóstico.

ScribesAI usa PortAudio que se conecta directamente via TCP.

### ¿Funciona con WSL1?

**No**. WSL1 no tiene red virtualizada ni soporte para audio de esta manera.

**Solución**: Actualiza a WSL2:

```powershell
wsl --set-version <distro> 2
```

### ¿Puedo usar otro puerto en lugar de 4713?

**Sí**, pero debes cambiar en:
1. `default.pa`: `port=XXXX`
2. Firewall: `-LocalPort XXXX`
3. Test scripts: Cambiar `4713` por tu puerto

**No recomendado** - 4713 es el puerto estándar de PulseAudio.

### ¿Funciona con Bluetooth?

**Sí**, si Windows reconoce el micrófono Bluetooth, PulseAudio lo capturará.

**Latencia adicional**: +50-200ms típico para Bluetooth.

### ¿Puedo grabar audio del sistema (loopback)?

**Sí**, con módulo loopback:

En `default.pa`:

```
load-module module-loopback
```

**Advertencia**: Puede crear retroalimentación si usas altavoces.

### ¿ScribesAI funciona sin esta configuración en WSL2?

**No**. WSL2 no tiene acceso a hardware de audio directamente.

**Alternativas**:
- Usar en Linux nativo
- Usar en macOS
- Grabar audio en Windows y transferir archivos

### ¿Los archivos de audio se guardan en Windows o WSL?

**WSL2**. Los archivos `.wav` se guardan en `~/scribesai/data/recordings/` dentro de tu sistema de archivos de WSL.

**Acceder desde Windows**: `\\wsl$\<distro>\home\<usuario>\scribesai\data\recordings\`

### ¿Puedo usar múltiples instancias de WSL2?

**Sí**, todas las instancias de WSL2 comparten la misma red virtual y pueden acceder al mismo PulseAudio.

### ¿Qué pasa si Windows se suspende?

PulseAudio se detendrá. Cuando Windows despierte:

1. PulseAudio puede reiniciarse automáticamente (si está en Startup)
2. O reinicia manualmente
3. ScribesAI detectará la desconexión y mostrará error

**Solución**: Reinicia PulseAudio y relanza ScribesAI.

---

## Referencias

### Documentación Oficial

- [PulseAudio Documentation](https://www.freedesktop.org/wiki/Software/PulseAudio/Documentation/)
- [WSL2 Documentation](https://docs.microsoft.com/en-us/windows/wsl/)
- [PortAudio Documentation](http://www.portaudio.com/docs/)

### Foros y Comunidad

- [WSL GitHub Issues - Audio](https://github.com/microsoft/WSL/issues?q=is%3Aissue+audio)
- [PulseAudio Mailing List](https://lists.freedesktop.org/mailman/listinfo/pulseaudio-discuss)
- [Reddit r/bashonubuntuonwindows](https://www.reddit.com/r/bashonubuntuonwindows/)

### Scripts de ScribesAI

- `scripts/setup-wsl-audio.sh` - Configuración automatizada
- `scripts/test-wsl-audio.sh` - Diagnóstico y validación
- `wsl-config/install-pulseaudio.ps1` - Instalador de PulseAudio para Windows

### Herramientas de Diagnóstico

**En WSL2:**
```bash
# Información del servidor PulseAudio
pactl info

# Listar dispositivos de entrada
pactl list sources short

# Ver estadísticas de latencia
pactl stat

# Probar grabación (5 segundos)
parecord --device=$(pactl list sources short | head -1 | awk '{print $1}') --channels=1 --rate=16000 /tmp/test.wav &
sleep 5
killall parecord
aplay /tmp/test.wav
```

**En Windows PowerShell:**
```powershell
# Ver proceso PulseAudio
Get-Process pulseaudio

# Ver puertos escuchando
netstat -ano | findstr :4713

# Ver reglas de firewall
Get-NetFirewallRule -DisplayName "*Pulse*"

# Logs de PulseAudio
# (No hay por defecto, usar --log-level=debug)
```

---

## Soporte

### Scripts de Ayuda

```bash
# Validación completa
./scripts/test-wsl-audio.sh

# Configuración interactiva
./scripts/setup-wsl-audio.sh

# Desde la app
./scribescli --help-audio
```

### Reportar Problemas

Si encuentras problemas, incluye:

1. **Salida de** `test-wsl-audio.sh`
2. **Versión de WSL**: `wsl --version`
3. **Distribución**: `lsb_release -a`
4. **Versión de Windows**: `winver`
5. **Logs de PulseAudio**: Ejecuta con `--verbose`

### Contacto

- **GitHub Issues**: [ScribesAI Issues](https://github.com/tu-repo/scribescli/issues)
- **Email**: support@scribesai.dev

---

## Changelog

### v1.0.0 (2025-01-XX)

- ✅ Soporte inicial para WSL2
- ✅ Scripts de configuración automática
- ✅ Detección inteligente de entorno
- ✅ Asistente TUI interactivo
- ✅ Validación con 8 tests
- ✅ Documentación completa

---

*"I'm sorry, Dave. I'm afraid I can't do that... without proper PulseAudio configuration."* - HAL 9000 (adaptado)

