#!/bin/bash

# ScribesAI Installation Script
# This script helps set up the development environment

set -e

echo "════════════════════════════════════════════════════════"
echo "  ScribesAI - Installation & Setup"
echo "  Meeting Intelligence System powered by HAL 9000"
echo "════════════════════════════════════════════════════════"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check if running on Linux or macOS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS_TYPE=Linux;;
    Darwin*)    OS_TYPE=Mac;;
    *)          OS_TYPE="UNKNOWN:${OS}"
esac

echo -e "${CYAN}Detected OS:${NC} $OS_TYPE"
echo ""

# 1. Check Go installation
echo -e "${CYAN}[1/6] Checking Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed${NC}"
    echo "Please install Go 1.24+ from https://golang.org/dl/"
    exit 1
else
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓ Go is installed:${NC} $GO_VERSION"
fi
echo ""

# 2. Check/Install PortAudio
echo -e "${CYAN}[2/6] Checking PortAudio...${NC}"
if pkg-config --exists portaudio-2.0; then
    echo -e "${GREEN}✓ PortAudio is installed${NC}"
else
    echo -e "${YELLOW}! PortAudio not found${NC}"
    echo "Installing PortAudio..."

    if [ "$OS_TYPE" == "Linux" ]; then
        if command -v apt-get &> /dev/null; then
            echo "Using apt-get (Debian/Ubuntu)..."
            sudo apt-get update
            sudo apt-get install -y portaudio19-dev
        elif command -v dnf &> /dev/null; then
            echo "Using dnf (Fedora)..."
            sudo dnf install -y portaudio-devel
        elif command -v pacman &> /dev/null; then
            echo "Using pacman (Arch)..."
            sudo pacman -S portaudio
        else
            echo -e "${RED}✗ Could not detect package manager${NC}"
            echo "Please install portaudio19-dev manually"
            exit 1
        fi
    elif [ "$OS_TYPE" == "Mac" ]; then
        if command -v brew &> /dev/null; then
            echo "Using Homebrew..."
            brew install portaudio
        else
            echo -e "${RED}✗ Homebrew not found${NC}"
            echo "Please install Homebrew from https://brew.sh/"
            exit 1
        fi
    fi

    echo -e "${GREEN}✓ PortAudio installed${NC}"
fi
echo ""

# 3. Setup .env file
echo -e "${CYAN}[3/6] Setting up environment variables...${NC}"
if [ ! -f .env ]; then
    cp .env.example .env
    echo -e "${GREEN}✓ Created .env file${NC}"
    echo -e "${YELLOW}! Please edit .env and add your ANTHROPIC_API_KEY${NC}"
    echo ""
    read -p "Do you want to enter your API key now? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Enter your Anthropic API key: " api_key
        sed -i.bak "s/ANTHROPIC_API_KEY=.*/ANTHROPIC_API_KEY=$api_key/" .env
        rm .env.bak 2>/dev/null || true
        echo -e "${GREEN}✓ API key configured${NC}"
    fi
else
    echo -e "${GREEN}✓ .env file already exists${NC}"
fi
echo ""

# 4. Create data directories
echo -e "${CYAN}[4/6] Creating data directories...${NC}"
mkdir -p data
mkdir -p models
mkdir -p data/exports
touch models/.gitkeep
echo -e "${GREEN}✓ Directories created${NC}"
echo ""

# 5. Download Go dependencies
echo -e "${CYAN}[5/6] Installing Go dependencies...${NC}"
go mod download
echo -e "${GREEN}✓ Dependencies installed${NC}"
echo ""

# 6. Build application
echo -e "${CYAN}[6/6] Building application...${NC}"
if go build -o scribescli ./cmd/scribescli; then
    echo -e "${GREEN}✓ Build successful${NC}"
    echo ""
    echo "════════════════════════════════════════════════════════"
    echo -e "${GREEN}✓ Installation complete!${NC}"
    echo "════════════════════════════════════════════════════════"
    echo ""
    echo "To run ScribesAI:"
    echo -e "${CYAN}  ./scribescli${NC}"
    echo ""
    echo "For help:"
    echo -e "${CYAN}  ./scribescli --help${NC}"
    echo ""
    echo -e "${YELLOW}Note:${NC} Make sure to configure your ANTHROPIC_API_KEY in .env"
    echo ""
else
    echo -e "${RED}✗ Build failed${NC}"
    echo "Please check the error messages above"
    exit 1
fi
