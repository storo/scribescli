#!/bin/bash
# ScribesAI - WSL2 Audio Setup Script
# Guides the user through setting up PulseAudio for audio recording in WSL2

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  ScribesAI - Asistente de Configuración de Audio WSL2${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo ""

# Step 0: Check if running in WSL2
if ! grep -qi microsoft /proc/version && [ -z "$WSL_DISTRO_NAME" ]; then
    echo -e "${RED}✗ Este script solo funciona en WSL2${NC}"
    echo -e "${YELLOW}  Estás en un sistema Linux nativo o macOS.${NC}"
    echo -e "${YELLOW}  No necesitas configurar PulseAudio.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ WSL2 detectado: $WSL_DISTRO_NAME${NC}"
echo ""

# Get Windows host IP
WINDOWS_IP=$(ip route show default | awk '{print $3}')
if [ -z "$WINDOWS_IP" ]; then
    WINDOWS_IP=$(grep nameserver /etc/resolv.conf | awk '{print $2}')
fi

echo -e "${CYAN}Windows Host IP:${NC} $WINDOWS_IP"
echo ""

# Step 1: Check if PulseAudio is already configured
if [ -n "$PULSE_SERVER" ]; then
    echo -e "${GREEN}✓ PULSE_SERVER ya está configurado: $PULSE_SERVER${NC}"
    echo ""
    echo -e "${YELLOW}Probando conexión...${NC}"

    if timeout 2 bash -c "echo > /dev/tcp/${WINDOWS_IP}/4713" 2>/dev/null; then
        echo -e "${GREEN}✓ PulseAudio está accesible!${NC}"
        echo ""
        echo "Tu configuración de audio ya está lista."
        echo "Puedes usar ScribesAI para grabar audio."
        exit 0
    else
        echo -e "${RED}✗ No se puede conectar a PulseAudio${NC}"
        echo -e "${YELLOW}  PulseAudio puede no estar corriendo en Windows${NC}"
        echo ""
        echo "¿Quieres reinstalar PulseAudio?"
        read -p "Continuar con instalación? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
fi

# Step 2: Check if pulseaudio is installed in WSL
echo -e "${CYAN}[1/5] Verificando dependencias en WSL...${NC}"

if ! command -v pulseaudio &> /dev/null; then
    echo -e "${YELLOW}! PulseAudio no está instalado en WSL${NC}"
    echo -e "${YELLOW}  Instalando...${NC}"

    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y pulseaudio
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y pulseaudio
    elif command -v pacman &> /dev/null; then
        sudo pacman -S pulseaudio
    else
        echo -e "${RED}✗ No se pudo instalar PulseAudio automáticamente${NC}"
        echo "  Instálalo manualmente y vuelve a ejecutar este script"
        exit 1
    fi
fi

echo -e "${GREEN}✓ PulseAudio instalado en WSL${NC}"
echo ""

# Step 3: Generate PowerShell script
echo -e "${CYAN}[2/5] Generando script de instalación para Windows...${NC}"

PS_SCRIPT_PATH="/tmp/install-pulseaudio-wsl.ps1"
cp "$PROJECT_ROOT/wsl-config/install-pulseaudio.ps1" "$PS_SCRIPT_PATH"

# Convert to Windows path
WINDOWS_PS_PATH=$(wslpath -w "$PS_SCRIPT_PATH")

echo -e "${GREEN}✓ Script generado${NC}"
echo ""

# Step 4: Instructions for Windows
echo -e "${CYAN}[3/5] Instalación en Windows${NC}"
echo ""
echo -e "${YELLOW}Ahora necesitas ejecutar un script en Windows para instalar PulseAudio.${NC}"
echo ""
echo -e "${CYAN}Sigue estos pasos:${NC}"
echo ""
echo "  1. Abre ${YELLOW}PowerShell como Administrador${NC}"
echo "     (Click derecho en el menú Inicio → PowerShell (Admin))"
echo ""
echo "  2. Copia y pega este comando:"
echo ""
echo -e "${GREEN}     powershell -ExecutionPolicy Bypass -File \"$WINDOWS_PS_PATH\"${NC}"
echo ""
echo "  3. Presiona Enter y espera a que termine"
echo ""
echo "  4. Vuelve a esta ventana cuando termine"
echo ""

read -p "Presiona Enter cuando hayas completado la instalación en Windows..." -r
echo ""

# Step 5: Test connection
echo -e "${CYAN}[4/5] Probando conexión a PulseAudio...${NC}"

if timeout 2 bash -c "echo > /dev/tcp/${WINDOWS_IP}/4713" 2>/dev/null; then
    echo -e "${GREEN}✓ ¡PulseAudio está accesible!${NC}"
else
    echo -e "${RED}✗ No se puede conectar a PulseAudio${NC}"
    echo ""
    echo -e "${YELLOW}Posibles problemas:${NC}"
    echo "  1. PulseAudio no está corriendo (ejecutar: C:\\Program Files\\pulse\\bin\\pulseaudio.exe)"
    echo "  2. Firewall bloqueando puerto 4713"
    echo "  3. Configuración incorrecta"
    echo ""
    echo "Consulta WSL_AUDIO_GUIDE.md para troubleshooting"
    echo ""
    read -p "¿Quieres continuar de todas formas? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Step 6: Configure ~/.bashrc
echo -e "${CYAN}[5/5] Configurando variables de entorno...${NC}"

BASHRC="$HOME/.bashrc"

if grep -q "ScribesAI - Audio en WSL2" "$BASHRC"; then
    echo -e "${YELLOW}! ~/.bashrc ya contiene configuración de ScribesAI${NC}"
    echo "  No se modificará"
else
    cat >> "$BASHRC" << EOF

# ScribesAI - Audio en WSL2
# Agregado automáticamente el $(date)
if grep -qi microsoft /proc/version; then
    # Detectar IP de Windows (host)
    export PULSE_SERVER=tcp:\$(ip route show default | awk '{print \$3}')

    # Opcional: Cookie para autenticación más segura
    # export PULSE_COOKIE=/mnt/c/Users/\$USER/.pulse-cookie
fi
EOF

    echo -e "${GREEN}✓ Variables agregadas a ~/.bashrc${NC}"
fi

# Apply immediately
export PULSE_SERVER="tcp:$WINDOWS_IP"

echo ""
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}          ¡Configuración Completa!${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}✓ PulseAudio configurado${NC}"
echo -e "${GREEN}✓ Variables de entorno establecidas${NC}"
echo ""
echo -e "${YELLOW}Para aplicar la configuración:${NC}"
echo -e "  ${CYAN}source ~/.bashrc${NC}"
echo ""
echo -e "${YELLOW}Para probar el audio:${NC}"
echo -e "  ${CYAN}./scribescli --test-audio${NC}"
echo ""
echo -e "${YELLOW}Para ejecutar ScribesAI:${NC}"
echo -e "  ${CYAN}./scribescli${NC}"
echo ""

# Optional: Reload bashrc
read -p "¿Quieres recargar ~/.bashrc ahora? (Y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
    source "$BASHRC"
    echo -e "${GREEN}✓ ~/.bashrc recargado${NC}"
fi

echo ""
echo "¡Listo para usar! 🎉"
