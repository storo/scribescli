#!/bin/bash
# ScribesAI - WSL2 Audio Test Script
# Validates that PulseAudio is properly configured and accessible

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  ScribesAI - Diagnóstico de Audio WSL2${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo ""

# Test 1: Check if WSL2
echo -e "${CYAN}[Test 1/8] Verificando entorno WSL2...${NC}"
if grep -qi microsoft /proc/version || [ -n "$WSL_DISTRO_NAME" ]; then
    echo -e "${GREEN}✓ PASS${NC} - Ejecutando en WSL2: ${WSL_DISTRO_NAME:-Unknown}"
    ((PASS_COUNT++))
else
    echo -e "${RED}✗ FAIL${NC} - No estás en WSL2"
    echo -e "${YELLOW}  Este diagnóstico solo funciona en WSL2${NC}"
    ((FAIL_COUNT++))
    exit 1
fi
echo ""

# Test 2: Check PULSE_SERVER variable
echo -e "${CYAN}[Test 2/8] Verificando variable PULSE_SERVER...${NC}"
if [ -n "$PULSE_SERVER" ]; then
    echo -e "${GREEN}✓ PASS${NC} - PULSE_SERVER=$PULSE_SERVER"
    ((PASS_COUNT++))
else
    echo -e "${RED}✗ FAIL${NC} - PULSE_SERVER no está configurado"
    echo -e "${YELLOW}  Ejecuta: ./scripts/setup-wsl-audio.sh${NC}"
    ((FAIL_COUNT++))
fi
echo ""

# Test 3: Get Windows host IP
echo -e "${CYAN}[Test 3/8] Detectando IP del host Windows...${NC}"
WINDOWS_IP=$(ip route show default | awk '{print $3}')
if [ -z "$WINDOWS_IP" ]; then
    WINDOWS_IP=$(grep nameserver /etc/resolv.conf | awk '{print $2}')
fi

if [ -n "$WINDOWS_IP" ]; then
    echo -e "${GREEN}✓ PASS${NC} - Windows Host IP: $WINDOWS_IP"
    ((PASS_COUNT++))
else
    echo -e "${RED}✗ FAIL${NC} - No se pudo detectar la IP de Windows"
    ((FAIL_COUNT++))
fi
echo ""

# Test 4: Test TCP connection to PulseAudio port
echo -e "${CYAN}[Test 4/8] Probando conexión al puerto PulseAudio (4713)...${NC}"
if timeout 2 bash -c "echo > /dev/tcp/${WINDOWS_IP}/4713" 2>/dev/null; then
    echo -e "${GREEN}✓ PASS${NC} - Puerto 4713 accesible"
    ((PASS_COUNT++))
else
    echo -e "${RED}✗ FAIL${NC} - No se puede conectar al puerto 4713"
    echo -e "${YELLOW}  Posibles causas:${NC}"
    echo -e "${YELLOW}  - PulseAudio no está corriendo en Windows${NC}"
    echo -e "${YELLOW}  - Firewall bloqueando el puerto${NC}"
    echo -e "${YELLOW}  - Configuración incorrecta${NC}"
    ((FAIL_COUNT++))
fi
echo ""

# Test 5: Check if pulseaudio client is installed
echo -e "${CYAN}[Test 5/8] Verificando cliente PulseAudio en WSL...${NC}"
if command -v pulseaudio &> /dev/null; then
    VERSION=$(pulseaudio --version | awk '{print $2}')
    echo -e "${GREEN}✓ PASS${NC} - PulseAudio instalado: v$VERSION"
    ((PASS_COUNT++))
else
    echo -e "${YELLOW}⚠ WARN${NC} - PulseAudio no instalado en WSL (opcional)"
    echo -e "${YELLOW}  Instalar con: sudo apt-get install pulseaudio${NC}"
    ((WARN_COUNT++))
fi
echo ""

# Test 6: Check pactl (PulseAudio control)
echo -e "${CYAN}[Test 6/8] Verificando herramienta pactl...${NC}"
if command -v pactl &> /dev/null; then
    echo -e "${GREEN}✓ PASS${NC} - pactl disponible"
    ((PASS_COUNT++))

    # Try to get server info
    if pactl info &> /dev/null; then
        echo -e "${GREEN}  ✓${NC} Información del servidor obtenida correctamente"
        echo ""
        echo -e "${CYAN}  Información del servidor PulseAudio:${NC}"
        pactl info | grep -E "Server String|Server Name|Server Version|Default Sink|Default Source" | sed 's/^/    /'
    else
        echo -e "${YELLOW}  ⚠${NC} No se pudo obtener información del servidor"
    fi
else
    echo -e "${YELLOW}⚠ WARN${NC} - pactl no disponible (opcional)"
    ((WARN_COUNT++))
fi
echo ""

# Test 7: Test recording devices
echo -e "${CYAN}[Test 7/8] Listando dispositivos de grabación...${NC}"
if command -v pactl &> /dev/null && pactl list sources short &> /dev/null; then
    SOURCE_COUNT=$(pactl list sources short | wc -l)
    if [ "$SOURCE_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓ PASS${NC} - $SOURCE_COUNT dispositivos de grabación disponibles"
        pactl list sources short | sed 's/^/  /'
        ((PASS_COUNT++))
    else
        echo -e "${YELLOW}⚠ WARN${NC} - No hay dispositivos de grabación"
        ((WARN_COUNT++))
    fi
else
    echo -e "${YELLOW}⚠ WARN${NC} - No se pueden listar dispositivos"
    ((WARN_COUNT++))
fi
echo ""

# Test 8: Recommend next steps
echo -e "${CYAN}[Test 8/8] Verificando configuración en ~/.bashrc...${NC}"
if grep -q "ScribesAI - Audio en WSL2" "$HOME/.bashrc"; then
    echo -e "${GREEN}✓ PASS${NC} - Configuración encontrada en ~/.bashrc"
    ((PASS_COUNT++))
else
    echo -e "${YELLOW}⚠ WARN${NC} - No se encontró configuración en ~/.bashrc"
    echo -e "${YELLOW}  Ejecuta: ./scripts/setup-wsl-audio.sh${NC}"
    ((WARN_COUNT++))
fi
echo ""

# Summary
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  Resumen del Diagnóstico${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}✓ Tests Pasados:${NC} $PASS_COUNT"
echo -e "${YELLOW}⚠ Advertencias:${NC} $WARN_COUNT"
echo -e "${RED}✗ Tests Fallidos:${NC} $FAIL_COUNT"
echo ""

# Overall status
if [ $FAIL_COUNT -eq 0 ]; then
    if [ $WARN_COUNT -eq 0 ]; then
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}  ✓ ¡Todo está configurado correctamente!${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo "ScribesAI está listo para grabar audio."
        echo ""
        echo -e "${CYAN}Ejecuta:${NC}"
        echo -e "  ${GREEN}./scribescli${NC}"
    else
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${YELLOW}  ⚠ Configuración funcional con advertencias${NC}"
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo "ScribesAI debería funcionar, pero hay algunas advertencias."
    fi
else
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}  ✗ Audio no configurado correctamente${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${YELLOW}Acciones recomendadas:${NC}"
    echo ""

    if [ -z "$PULSE_SERVER" ]; then
        echo -e "  ${CYAN}1.${NC} Configura PULSE_SERVER:"
        echo -e "     ${GREEN}./scripts/setup-wsl-audio.sh${NC}"
        echo ""
    fi

    if ! timeout 2 bash -c "echo > /dev/tcp/${WINDOWS_IP}/4713" 2>/dev/null; then
        echo -e "  ${CYAN}2.${NC} Verifica que PulseAudio esté corriendo en Windows:"
        echo -e "     Abre PowerShell y ejecuta:"
        echo -e "     ${GREEN}C:\\Program Files\\pulse\\bin\\pulseaudio.exe${NC}"
        echo ""
        echo -e "  ${CYAN}3.${NC} Verifica el firewall de Windows:"
        echo -e "     Permite pulseaudio.exe en puerto 4713"
        echo ""
    fi

    echo -e "  ${CYAN}4.${NC} Consulta la guía completa:"
    echo -e "     ${GREEN}cat WSL_AUDIO_GUIDE.md${NC}"
fi

echo ""
exit $FAIL_COUNT
