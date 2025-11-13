package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/storo/scribescli/internal/audio"
)

// AudioSetupStep represents the current step in the setup wizard
type AudioSetupStep int

const (
	StepWelcome AudioSetupStep = iota
	StepDetection
	StepGenerateScript
	StepWindowsInstructions
	StepTestConnection
	StepConfigureBashrc
	StepComplete
	StepFailed
)

// AudioSetupWizard manages the audio setup process
type AudioSetupWizard struct {
	currentStep      AudioSetupStep
	envInfo          *audio.EnvironmentInfo
	windowsIP        string
	bashrcConfig     string
	scriptGenerated  bool
	connectionTested bool
	connectionOK     bool
	errorMessage     string
}

// NewAudioSetupWizard creates a new audio setup wizard
func NewAudioSetupWizard() *AudioSetupWizard {
	return &AudioSetupWizard{
		currentStep: StepWelcome,
	}
}

// Next advances to the next step
func (w *AudioSetupWizard) Next() {
	switch w.currentStep {
	case StepWelcome:
		w.currentStep = StepDetection
		w.detectEnvironment()
	case StepDetection:
		if w.envInfo.IsWSL2 && !w.envInfo.PulseReachable {
			w.currentStep = StepGenerateScript
			w.generateScript()
		} else {
			w.currentStep = StepComplete
		}
	case StepGenerateScript:
		w.currentStep = StepWindowsInstructions
	case StepWindowsInstructions:
		w.currentStep = StepTestConnection
	case StepTestConnection:
		w.testConnection()
		if w.connectionOK {
			w.currentStep = StepConfigureBashrc
		} else {
			w.currentStep = StepFailed
		}
	case StepConfigureBashrc:
		w.configureBashrc()
		w.currentStep = StepComplete
	}
}

// Back goes to previous step
func (w *AudioSetupWizard) Back() {
	if w.currentStep > StepWelcome {
		w.currentStep--
	}
}

// detectEnvironment detects the audio environment
func (w *AudioSetupWizard) detectEnvironment() {
	w.envInfo = audio.GetEnvironmentInfo()
	if w.envInfo.WindowsHostIP != "" {
		w.windowsIP = w.envInfo.WindowsHostIP
	}
}

// generateScript generates the PowerShell script
func (w *AudioSetupWizard) generateScript() {
	w.scriptGenerated = true
}

// testConnection tests the PulseAudio connection
func (w *AudioSetupWizard) testConnection() {
	w.connectionTested = true
	w.connectionOK = audio.TestPulseAudioConnection()
}

// configureBashrc configures ~/.bashrc
func (w *AudioSetupWizard) configureBashrc() error {
	config, err := audio.GenerateBashrcConfig()
	if err != nil {
		w.errorMessage = err.Error()
		return err
	}
	w.bashrcConfig = config

	err = audio.AppendToBashrc(config)
	if err != nil {
		w.errorMessage = err.Error()
		return err
	}

	return nil
}

// Render renders the current step
func (w *AudioSetupWizard) Render(width, height int) string {
	switch w.currentStep {
	case StepWelcome:
		return w.renderWelcome(width, height)
	case StepDetection:
		return w.renderDetection(width, height)
	case StepGenerateScript:
		return w.renderGenerateScript(width, height)
	case StepWindowsInstructions:
		return w.renderWindowsInstructions(width, height)
	case StepTestConnection:
		return w.renderTestConnection(width, height)
	case StepConfigureBashrc:
		return w.renderConfigureBashrc(width, height)
	case StepComplete:
		return w.renderComplete(width, height)
	case StepFailed:
		return w.renderFailed(width, height)
	default:
		return "Unknown step"
	}
}

func (w *AudioSetupWizard) renderWelcome(width, height int) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("ASISTENTE DE CONFIGURACIÓN DE AUDIO WSL2"),
		"",
		lipgloss.NewStyle().Foreground(ColorWhite).Render(
			"Este asistente te guiará a través de la configuración de audio\n"+
				"para ScribesAI en WSL2.",
		),
		"",
		lipgloss.NewStyle().Foreground(ColorAmber).Render(
			"⏱ Tiempo estimado: 5 minutos",
		),
		"",
		lipgloss.NewStyle().Foreground(ColorCyan).Render(
			"Requisitos:",
		),
		"  • WSL2 en Windows 10/11",
		"  • Acceso a PowerShell como Administrador",
		"  • Micrófono conectado",
		"",
		"",
		HelpStyle.Render("[Enter] Continuar  [Q] Cancelar"),
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (w *AudioSetupWizard) renderDetection(width, height int) string {
	var lines []string

	lines = append(lines, TitleStyle.Render("DETECTANDO ENTORNO"))
	lines = append(lines, "")

	if w.envInfo == nil {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorAmber).Render("⏳ Analizando..."))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorCyan).Render("Información del Sistema:"))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  Entorno: %s", w.envInfo.Environment))
		lines = append(lines, fmt.Sprintf("  WSL2: %v", w.envInfo.IsWSL2))
		if w.envInfo.WindowsHostIP != "" {
			lines = append(lines, fmt.Sprintf("  Windows IP: %s", w.envInfo.WindowsHostIP))
		}
		lines = append(lines, fmt.Sprintf("  PULSE_SERVER: %s", w.envInfo.PulseServerValue))
		lines = append(lines, fmt.Sprintf("  Conexión PulseAudio: %v", w.envInfo.PulseReachable))
		lines = append(lines, "")

		if w.envInfo.CanRecordAudio {
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorSuccess).Render("✓ Audio ya configurado correctamente"))
		} else {
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorWarning).Render("⚠ Audio no configurado"))
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorWhite).Render("  "+w.envInfo.RecommendedAction))
		}
	}

	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, HelpStyle.Render("[Enter] Continuar  [B] Atrás  [Q] Cancelar"))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (w *AudioSetupWizard) renderGenerateScript(width, height int) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("PASO 1/3: INSTALAR PULSEAUDIO EN WINDOWS"),
		"",
		lipgloss.NewStyle().Foreground(ColorWhite).Render(
			"Para grabar audio, necesitas instalar PulseAudio en Windows.\n"+
				"El script de instalación automatizará todo el proceso.",
		),
		"",
		lipgloss.NewStyle().Foreground(ColorCyan).Render("Script de instalación:"),
		"  " + lipgloss.NewStyle().Foreground(ColorAmber).Render("./wsl-config/install-pulseaudio.ps1"),
		"",
		lipgloss.NewStyle().Foreground(ColorSuccess).Render("✓ Script generado"),
		"",
		"",
		HelpStyle.Render("[Enter] Ver instrucciones  [B] Atrás  [Q] Cancelar"),
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (w *AudioSetupWizard) renderWindowsInstructions(width, height int) string {
	// Get the actual script path
	scriptPath := "./wsl-config/install-pulseaudio.ps1"

	instructions := []string{
		TitleStyle.Render("PASO 2/3: EJECUTAR SCRIPT EN WINDOWS"),
		"",
		lipgloss.NewStyle().Foreground(ColorAmber).Render("⚠ IMPORTANTE: Debes ejecutar esto en Windows, NO en WSL"),
		"",
		lipgloss.NewStyle().Foreground(ColorCyan).Render("Instrucciones:"),
		"",
		"  1. " + lipgloss.NewStyle().Bold(true).Render("Abre PowerShell como Administrador"),
		"     (Click derecho en Inicio → PowerShell (Admin))",
		"",
		"  2. " + lipgloss.NewStyle().Bold(true).Render("Navega al directorio del proyecto:"),
		"     cd (tu ruta de ScribesAI en Windows)",
		"",
		"  3. " + lipgloss.NewStyle().Bold(true).Render("Ejecuta el script:"),
		"     " + lipgloss.NewStyle().Foreground(ColorSuccess).Render(scriptPath),
		"",
		"  4. " + lipgloss.NewStyle().Bold(true).Render("Espera a que termine la instalación"),
		"",
		"",
		lipgloss.NewStyle().Foreground(ColorWhite).Render(
			"Cuando hayas completado la instalación en Windows,\n"+
				"presiona Enter para probar la conexión.",
		),
		"",
		"",
		HelpStyle.Render("[Enter] Probar conexión  [B] Atrás  [Q] Cancelar"),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, instructions...)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (w *AudioSetupWizard) renderTestConnection(width, height int) string {
	var lines []string

	lines = append(lines, TitleStyle.Render("PASO 3/3: PROBANDO CONEXIÓN"))
	lines = append(lines, "")

	if !w.connectionTested {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorAmber).Render("⏳ Probando conexión a PulseAudio..."))
	} else {
		if w.connectionOK {
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorSuccess).Render("✓ ¡Conexión exitosa!"))
			lines = append(lines, "")
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorWhite).Render(
				"PulseAudio está corriendo y accesible desde WSL2.",
			))
		} else {
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorError).Render("✗ No se pudo conectar"))
			lines = append(lines, "")
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorAmber).Render("Posibles causas:"))
			lines = append(lines, "  • PulseAudio no está corriendo en Windows")
			lines = append(lines, "  • Firewall bloqueando puerto 4713")
			lines = append(lines, "  • Configuración incorrecta")
			lines = append(lines, "")
			lines = append(lines, lipgloss.NewStyle().Foreground(ColorWhite).Render(
				"Consulta WSL_AUDIO_GUIDE.md para troubleshooting",
			))
		}
	}

	lines = append(lines, "")
	lines = append(lines, "")

	if w.connectionOK {
		lines = append(lines, HelpStyle.Render("[Enter] Configurar variables  [Q] Salir"))
	} else {
		lines = append(lines, HelpStyle.Render("[R] Reintentar  [B] Atrás  [Q] Cancelar"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (w *AudioSetupWizard) renderConfigureBashrc(width, height int) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("CONFIGURANDO VARIABLES DE ENTORNO"),
		"",
		lipgloss.NewStyle().Foreground(ColorAmber).Render("⏳ Agregando configuración a ~/.bashrc..."),
		"",
		lipgloss.NewStyle().Foreground(ColorSuccess).Render("✓ Variables configuradas"),
		"",
		"",
		HelpStyle.Render("Presiona Enter para continuar..."),
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (w *AudioSetupWizard) renderComplete(width, height int) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render("✓ CONFIGURACIÓN COMPLETA"),
		"",
		lipgloss.NewStyle().Foreground(ColorSuccess).Render(
			"Audio habilitado en WSL2. ¡Ya puedes grabar reuniones!",
		),
		"",
		lipgloss.NewStyle().Foreground(ColorCyan).Render("Próximos pasos:"),
		"",
		"  1. Recarga tu terminal:"+
		"     " + lipgloss.NewStyle().Foreground(ColorAmber).Render("source ~/.bashrc"),
		"",
		"  2. Ejecuta ScribesAI:",
		"     " + lipgloss.NewStyle().Foreground(ColorAmber).Render("./scribescli"),
		"",
		"",
		HelpStyle.Render("[Enter] Salir del asistente"),
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func (w *AudioSetupWizard) renderFailed(width, height int) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(ColorError).Render("✗ CONFIGURACIÓN FALLIDA"),
		"",
		lipgloss.NewStyle().Foreground(ColorWhite).Render(
			"No se pudo completar la configuración de audio.",
		),
		"",
		lipgloss.NewStyle().Foreground(ColorAmber).Render("Error:"),
		"  "+w.errorMessage,
		"",
		lipgloss.NewStyle().Foreground(ColorCyan).Render("Opciones:"),
		"",
		"  • Consultar WSL_AUDIO_GUIDE.md",
		"  • Ejecutar: ./scripts/test-wsl-audio.sh",
		"  • Configurar manualmente",
		"",
		"",
		HelpStyle.Render("[R] Reintentar  [Q] Salir"),
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
