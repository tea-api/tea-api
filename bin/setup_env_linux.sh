#!/bin/bash
set -e

# This script sets up the development environment for Tea API on Linux (Debian/Ubuntu).
# It installs required dependencies and prepares environment files.

echo "=== Setting up Tea API development environment on Linux ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [ "$EUID" -eq 0 ]; then
        print_warning "Running as root. This script should be run as a regular user."
        print_warning "Some operations will use sudo when needed."
    fi
}

# Update system packages
update_system() {
    print_status "Updating system packages..."
    sudo apt-get update
}

# Install Node.js
install_nodejs() {
    if command -v node >/dev/null 2>&1; then
        NODE_VERSION=$(node --version)
        print_status "Node.js is already installed: $NODE_VERSION"
        
        # Check if version is recent enough (v16+)
        NODE_MAJOR=$(echo $NODE_VERSION | cut -d'.' -f1 | sed 's/v//')
        if [ "$NODE_MAJOR" -lt 16 ]; then
            print_warning "Node.js version is too old. Installing newer version..."
            install_nodejs_fresh
        fi
    else
        print_status "Installing Node.js..."
        install_nodejs_fresh
    fi
}

install_nodejs_fresh() {
    # Install Node.js 18.x
    curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
    sudo apt-get install -y nodejs
    
    # Verify installation
    if command -v node >/dev/null 2>&1; then
        print_status "Node.js installed successfully: $(node --version)"
        print_status "npm version: $(npm --version)"
    else
        print_error "Failed to install Node.js"
        exit 1
    fi
}

# Install Go
install_go() {
    if command -v go >/dev/null 2>&1; then
        GO_VERSION=$(go version | awk '{print $3}')
        print_status "Go is already installed: $GO_VERSION"
        
        # Check if version is recent enough (1.18+)
        GO_MAJOR=$(echo $GO_VERSION | sed 's/go//' | cut -d'.' -f1)
        GO_MINOR=$(echo $GO_VERSION | sed 's/go//' | cut -d'.' -f2)
        if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 18 ]); then
            print_warning "Go version is too old. Installing newer version..."
            install_go_fresh
        fi
    else
        print_status "Installing Go..."
        install_go_fresh
    fi
}

install_go_fresh() {
    # Determine architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)
            GO_ARCH=amd64
            ;;
        aarch64|arm64)
            GO_ARCH=arm64
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    # Download and install Go
    GO_VERSION="1.23.4"
    GO_TARBALL="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    
    print_status "Downloading Go ${GO_VERSION} for ${GO_ARCH}..."
    wget -q "https://go.dev/dl/${GO_TARBALL}" -O "/tmp/${GO_TARBALL}"
    
    # Remove old installation if exists
    sudo rm -rf /usr/local/go
    
    # Extract new installation
    sudo tar -C /usr/local -xzf "/tmp/${GO_TARBALL}"
    
    # Add to PATH if not already there
    if ! echo "$PATH" | grep -q "/usr/local/go/bin"; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        export PATH=$PATH:/usr/local/go/bin
    fi
    
    # Clean up
    rm "/tmp/${GO_TARBALL}"
    
    # Verify installation
    if command -v go >/dev/null 2>&1; then
        print_status "Go installed successfully: $(go version)"
    else
        print_error "Failed to install Go. Please add /usr/local/go/bin to your PATH and restart your shell."
        print_status "Run: echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc"
        exit 1
    fi
}

# Install additional dependencies
install_dependencies() {
    print_status "Installing additional dependencies..."
    
    # Install build essentials and other tools
    sudo apt-get install -y \
        build-essential \
        curl \
        wget \
        git \
        unzip \
        ca-certificates \
        software-properties-common \
        apt-transport-https \
        gnupg \
        lsb-release
    
    print_status "Additional dependencies installed successfully."
}

# Install Docker (optional)
install_docker() {
    if command -v docker >/dev/null 2>&1; then
        print_status "Docker is already installed: $(docker --version)"
    else
        read -p "Do you want to install Docker? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_status "Installing Docker..."
            
            # Add Docker's official GPG key
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
            
            # Add Docker repository
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
            
            # Install Docker
            sudo apt-get update
            sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
            
            # Add user to docker group
            sudo usermod -aG docker $USER
            
            print_status "Docker installed successfully."
            print_warning "Please log out and log back in for Docker group changes to take effect."
        fi
    fi
}

# Setup project environment
setup_project() {
    print_status "Setting up project environment..."
    
    SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
    REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
    
    # Install frontend dependencies
    if [ -d "$REPO_ROOT/web" ]; then
        print_status "Installing frontend dependencies..."
        cd "$REPO_ROOT/web"
        npm install
        cd "$REPO_ROOT"
    fi
    
    # Download Go modules
    if [ -f "$REPO_ROOT/go.mod" ]; then
        print_status "Downloading Go modules..."
        cd "$REPO_ROOT"
        go mod download
    fi
    
    # Create environment file if it doesn't exist
    ENV_FILE="$REPO_ROOT/.env"
    if [ ! -f "$ENV_FILE" ]; then
        if [ -f "$REPO_ROOT/.env.example" ]; then
            cp "$REPO_ROOT/.env.example" "$ENV_FILE"
            print_status "Created .env from .env.example"
        else
            # Create basic .env file
            cat > "$ENV_FILE" << 'EOF'
# Database configuration
SQL_DSN=./tea-api.db

# Server configuration
TZ=Asia/Shanghai
ERROR_LOG_ENABLED=true
TIKTOKEN_CACHE_DIR=./tiktoken_cache

# Cache configuration
MEMORY_CACHE_ENABLED=true

# Rate limiting
RATE_LIMIT_ENABLED=true
EOF
            print_status "Created basic .env file"
        fi
    else
        print_status ".env file already exists"
    fi
    
    # Create necessary directories
    mkdir -p "$REPO_ROOT/data"
    mkdir -p "$REPO_ROOT/logs"
    mkdir -p "$REPO_ROOT/tiktoken_cache"
    
    print_status "Project environment setup completed."
}

# Make scripts executable
make_scripts_executable() {
    SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
    
    print_status "Making scripts executable..."
    chmod +x "$SCRIPT_DIR"/*.sh
}

# Main execution
main() {
    check_root
    update_system
    install_dependencies
    install_nodejs
    install_go
    install_docker
    setup_project
    make_scripts_executable
    
    echo ""
    print_status "=== Environment setup completed successfully! ==="
    echo ""
    print_status "Next steps:"
    echo "1. If you installed Docker, log out and log back in"
    echo "2. Run 'source ~/.bashrc' to reload your shell environment"
    echo "3. Run './bin/build_linux.sh' to build the application"
    echo "4. Configure your .env file as needed"
    echo ""
    print_status "Installed versions:"
    echo "- Node.js: $(node --version 2>/dev/null || echo 'Not found')"
    echo "- npm: $(npm --version 2>/dev/null || echo 'Not found')"
    echo "- Go: $(go version 2>/dev/null || echo 'Not found')"
    echo "- Docker: $(docker --version 2>/dev/null || echo 'Not installed')"
}

# Run main function
main "$@"
