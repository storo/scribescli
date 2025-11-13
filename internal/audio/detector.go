package audio

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

// AudioEnvironment represents the detected audio environment
type AudioEnvironment int

const (
	// NativeLinux represents native Linux with direct audio access
	NativeLinux AudioEnvironment = iota
	// NativeMacOS represents native macOS with CoreAudio
	NativeMacOS
	// WSL2WithAudio represents WSL2 with PulseAudio configured
	WSL2WithAudio
	// WSL2NoAudio represents WSL2 without audio configuration
	WSL2NoAudio
	// Unknown represents an unknown environment
	Unknown
)

// String returns a human-readable description of the environment
func (e AudioEnvironment) String() string {
	switch e {
	case NativeLinux:
		return "Native Linux"
	case NativeMacOS:
		return "Native macOS"
	case WSL2WithAudio:
		return "WSL2 with PulseAudio"
	case WSL2NoAudio:
		return "WSL2 without audio"
	default:
		return "Unknown"
	}
}

// DetectAudioEnvironment detects the current audio environment
func DetectAudioEnvironment() AudioEnvironment {
	// Check if macOS
	if IsMacOS() {
		return NativeMacOS
	}

	// Check if WSL2
	if IsWSL2() {
		if HasPulseAudioConfigured() && TestPulseAudioConnection() {
			return WSL2WithAudio
		}
		return WSL2NoAudio
	}

	// Assume native Linux
	return NativeLinux
}

// IsWSL2 detects if running in WSL2
func IsWSL2() bool {
	// Method 1: Check /proc/version for Microsoft or WSL
	if data, err := os.ReadFile("/proc/version"); err == nil {
		version := strings.ToLower(string(data))
		if strings.Contains(version, "microsoft") || strings.Contains(version, "wsl") {
			return true
		}
	}

	// Method 2: Check for WSL_DISTRO_NAME environment variable
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}

	// Method 3: Check if /proc/sys/fs/binfmt_misc/WSLInterop exists
	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
		return true
	}

	return false
}

// IsMacOS detects if running on macOS
func IsMacOS() bool {
	// Check uname -s
	cmd := exec.Command("uname", "-s")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "Darwin"
}

// HasPulseAudioConfigured checks if PulseAudio is configured for WSL2
func HasPulseAudioConfigured() bool {
	// Check if PULSE_SERVER is set
	pulseServer := os.Getenv("PULSE_SERVER")
	if pulseServer == "" {
		return false
	}

	// Validate format (tcp:IP, unix:/path, or just IP)
	if strings.HasPrefix(pulseServer, "tcp:") ||
	   strings.HasPrefix(pulseServer, "unix:") ||
	   net.ParseIP(pulseServer) != nil {
		return true
	}

	return false
}

// GetPulseAudioServer returns the PulseAudio server address
func GetPulseAudioServer() string {
	server := os.Getenv("PULSE_SERVER")
	if server == "" {
		return ""
	}

	// Remove tcp: prefix if present
	return strings.TrimPrefix(server, "tcp:")
}

// TestPulseAudioConnection tests if PulseAudio server is reachable
func TestPulseAudioConnection() bool {
	pulseServer := os.Getenv("PULSE_SERVER")
	if pulseServer == "" {
		return false
	}

	// Handle Unix socket (WSLg)
	if strings.HasPrefix(pulseServer, "unix:") {
		socketPath := strings.TrimPrefix(pulseServer, "unix:")
		_, err := os.Stat(socketPath)
		return err == nil
	}

	// Handle TCP connection
	server := GetPulseAudioServer()
	if server == "" {
		return false
	}

	// Ensure port is included
	if !strings.Contains(server, ":") {
		server = server + ":4713" // Default PulseAudio TCP port
	}

	// Try to connect with 2 second timeout
	conn, err := net.DialTimeout("tcp", server, 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

// GetWindowsHostIP tries to determine the Windows host IP from WSL2
func GetWindowsHostIP() (string, error) {
	// Method 1: Use default route (most reliable)
	cmd := exec.Command("sh", "-c", "ip route show default | awk '{print $3}'")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		ip := strings.TrimSpace(string(output))
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
	}

	// Method 2: Use nameserver from resolv.conf
	data, err := os.ReadFile("/etc/resolv.conf")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "nameserver") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					ip := fields[1]
					if net.ParseIP(ip) != nil {
						return ip, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("could not determine Windows host IP")
}

// IsPulseAudioInstalled checks if pulseaudio is installed in WSL
func IsPulseAudioInstalled() bool {
	_, err := exec.LookPath("pulseaudio")
	return err == nil
}

// GetPulseAudioVersion returns the installed PulseAudio version
func GetPulseAudioVersion() string {
	cmd := exec.Command("pulseaudio", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Output format: "pulseaudio 15.99.1"
	version := strings.TrimSpace(string(output))
	parts := strings.Fields(version)
	if len(parts) >= 2 {
		return parts[1]
	}

	return "unknown"
}

// EnvironmentInfo contains detailed environment information
type EnvironmentInfo struct {
	Environment        AudioEnvironment
	IsWSL2             bool
	PulseServerSet     bool
	PulseServerValue   string
	PulseReachable     bool
	PulseInstalled     bool
	PulseVersion       string
	WindowsHostIP      string
	CanRecordAudio     bool
	RecommendedAction  string
}

// GetEnvironmentInfo returns detailed information about the audio environment
func GetEnvironmentInfo() *EnvironmentInfo {
	info := &EnvironmentInfo{
		Environment:    DetectAudioEnvironment(),
		IsWSL2:         IsWSL2(),
		PulseInstalled: IsPulseAudioInstalled(),
	}

	if info.IsWSL2 {
		windowsIP, _ := GetWindowsHostIP()
		info.WindowsHostIP = windowsIP

		info.PulseServerSet = HasPulseAudioConfigured()
		info.PulseServerValue = GetPulseAudioServer()
		info.PulseReachable = TestPulseAudioConnection()

		if info.PulseInstalled {
			info.PulseVersion = GetPulseAudioVersion()
		}

		// Determine if audio recording is possible
		info.CanRecordAudio = info.PulseServerSet && info.PulseReachable

		// Provide recommendation
		if !info.PulseServerSet {
			info.RecommendedAction = "Configure PulseAudio environment variables (run setup wizard)"
		} else if !info.PulseReachable {
			info.RecommendedAction = "Start PulseAudio server on Windows and check firewall"
		} else {
			info.RecommendedAction = "Audio is ready to use"
		}
	} else {
		// Native Linux or macOS
		info.CanRecordAudio = true
		info.RecommendedAction = "Audio ready (native platform)"
	}

	return info
}
