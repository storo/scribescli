package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WSLAudioConfig holds WSL audio configuration
type WSLAudioConfig struct {
	WindowsHostIP    string
	PulseServer      string
	PulseCookie      string
	BashrcPath       string
	ConfigGenerated  bool
}

// GenerateBashrcConfig generates the bashrc configuration for WSL audio
func GenerateBashrcConfig() (string, error) {
	hostIP, err := GetWindowsHostIP()
	if err != nil {
		return "", fmt.Errorf("failed to get Windows host IP: %w", err)
	}

	config := fmt.Sprintf(`
# ScribesAI - Audio en WSL2
# Agregado automáticamente por ScribesAI setup
if grep -qi microsoft /proc/version; then
    # Detectar IP de Windows (host)
    export PULSE_SERVER=tcp:%s

    # Opcional: Cookie para autenticación más segura
    # export PULSE_COOKIE=/mnt/c/Users/$USER/.pulse-cookie

    # Verificar conexión (silencioso)
    # timeout 1 bash -c "echo > /dev/tcp/%s/4713" 2>/dev/null && \
    #     echo "✓ Audio WSL2: conectado a $PULSE_SERVER"
fi
`, hostIP, hostIP)

	return config, nil
}

// AppendToBashrc appends configuration to ~/.bashrc
func AppendToBashrc(config string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	bashrcPath := filepath.Join(home, ".bashrc")

	// Check if already configured
	data, err := os.ReadFile(bashrcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .bashrc: %w", err)
	}

	// Check if already contains ScribesAI audio config
	if strings.Contains(string(data), "ScribesAI - Audio en WSL2") {
		return fmt.Errorf(".bashrc already contains ScribesAI audio configuration")
	}

	// Append to .bashrc
	f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .bashrc: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + config + "\n"); err != nil {
		return fmt.Errorf("failed to write to .bashrc: %w", err)
	}

	return nil
}

// GetWSLDistroName returns the WSL distribution name
func GetWSLDistroName() string {
	return os.Getenv("WSL_DISTRO_NAME")
}

// GetWindowsUserPath returns the Windows user path from WSL
func GetWindowsUserPath() string {
	// WSL mounts Windows drives under /mnt/
	// Try to find Windows username
	username := os.Getenv("USER")
	return fmt.Sprintf("/mnt/c/Users/%s", username)
}

// GetPulseCookiePath returns the path to the PulseAudio cookie
func GetPulseCookiePath() string {
	windowsUserPath := GetWindowsUserPath()
	return filepath.Join(windowsUserPath, ".pulse-cookie")
}

// CheckFirewallPort checks if a port is accessible (basic check)
func CheckFirewallPort(host string, port int) bool {
	return TestPulseAudioConnection()
}

// GenerateSetupInstructions generates human-readable setup instructions
func GenerateSetupInstructions() string {
	hostIP, _ := GetWindowsHostIP()

	instructions := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════╗
║         CONFIGURACIÓN DE AUDIO PARA WSL2                     ║
╠══════════════════════════════════════════════════════════════╣

ScribesAI detectó que estás en WSL2 sin audio configurado.
Para habilitar grabación de audio, necesitas configurar PulseAudio.

OPCIÓN 1: SETUP AUTOMÁTICO (Recomendado)
─────────────────────────────────────────

  1. Ejecuta el asistente de configuración:
     ./scribescli setup-audio

  2. Sigue las instrucciones en pantalla

  3. Tiempo estimado: 5 minutos

OPCIÓN 2: SETUP MANUAL
──────────────────────

PASO 1 - En Windows (PowerShell como Administrador):

  # Descargar PulseAudio
  Invoke-WebRequest -Uri https://www.freedesktop.org/software/pulseaudio/releases/pulseaudio-1.1.zip -OutFile pulseaudio.zip

  # Extraer a C:\Program Files\pulse
  Expand-Archive -Path pulseaudio.zip -DestinationPath "C:\Program Files\pulse"

PASO 2 - Configurar PulseAudio en Windows:

  Edita: C:\Program Files\pulse\etc\pulse\default.pa

  Agrega al inicio:
    load-module module-native-protocol-tcp port=4713 auth-ip-acl=172.16.0.0/12
    load-module module-esound-protocol-tcp port=4714 auth-ip-acl=172.16.0.0/12
    load-module module-waveout sink_name=output source_name=input record=0

  Edita: C:\Program Files\pulse\etc\pulse\daemon.conf

  Cambia:
    exit-idle-time = -1

PASO 3 - Iniciar PulseAudio:

  cd "C:\Program Files\pulse\bin"
  .\pulseaudio.exe

PASO 4 - Configurar Firewall:

  Permitir pulseaudio.exe en Windows Firewall (puerto 4713)

PASO 5 - En WSL2:

  # Agregar a ~/.bashrc
  export PULSE_SERVER=tcp:%s

  # Recargar
  source ~/.bashrc

PASO 6 - Verificar:

  ./scribescli --test-audio

╚══════════════════════════════════════════════════════════════╝

Para más información: WSL_AUDIO_GUIDE.md
`, hostIP)

	return instructions
}

// WSLSystemInfo returns system information useful for debugging
type WSLSystemInfo struct {
	DistroName      string
	WindowsHostIP   string
	KernelVersion   string
	PulseInstalled  bool
	PulseVersion    string
}

// GetWSLSystemInfo collects WSL system information
func GetWSLSystemInfo() *WSLSystemInfo {
	info := &WSLSystemInfo{
		DistroName:     GetWSLDistroName(),
		PulseInstalled: IsPulseAudioInstalled(),
	}

	if hostIP, err := GetWindowsHostIP(); err == nil {
		info.WindowsHostIP = hostIP
	}

	if data, err := os.ReadFile("/proc/version"); err == nil {
		info.KernelVersion = strings.TrimSpace(string(data))
	}

	if info.PulseInstalled {
		info.PulseVersion = GetPulseAudioVersion()
	}

	return info
}
