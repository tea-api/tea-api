#!/bin/bash
set -e

# This script builds Tea API for Linux (Debian/Ubuntu) for production deployment.
# It compiles the frontend and backend and outputs a Linux binary.

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)

echo "=== Building Tea API for Linux ==="

# Check if required tools are installed
check_dependencies() {
    echo "Checking dependencies..."
    
    # Check Node.js
    if ! command -v node >/dev/null 2>&1; then
        echo "Error: Node.js is not installed. Please install Node.js first."
        echo "Run: curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash - && sudo apt-get install -y nodejs"
        exit 1
    fi
    
    # Check npm
    if ! command -v npm >/dev/null 2>&1; then
        echo "Error: npm is not installed. Please install npm first."
        exit 1
    fi
    
    # Check Go
    if ! command -v go >/dev/null 2>&1; then
        echo "Error: Go is not installed. Please install Go first."
        echo "Run: wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz && sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz"
        echo "Add to ~/.bashrc: export PATH=\$PATH:/usr/local/go/bin"
        exit 1
    fi
    
    echo "All dependencies are installed."
}

# Build frontend
build_frontend() {
    echo "Building frontend..."
    cd "$REPO_ROOT/web"
    
    # Install dependencies if node_modules doesn't exist
    if [ ! -d node_modules ]; then
        echo "Installing frontend dependencies..."
        npm install
    fi
    
    # Build frontend with proper environment variables
    echo "Compiling frontend..."
    DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat "$REPO_ROOT/VERSION") npm run build
    
    if [ ! -d dist ]; then
        echo "Error: Frontend build failed - dist directory not found"
        exit 1
    fi
    
    echo "Frontend build completed successfully."
    cd "$REPO_ROOT"
}

# Build backend
build_backend() {
    echo "Building backend..."
    cd "$REPO_ROOT"
    
    # Determine architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)
            GOARCH=amd64
            ;;
        aarch64|arm64)
            GOARCH=arm64
            ;;
        *)
            echo "Unsupported architecture: $ARCH" >&2
            exit 1
            ;;
    esac
    
    echo "Building for Linux $GOARCH..."
    
    # Download Go modules
    echo "Downloading Go modules..."
    go mod download
    
    # Build with static linking for better compatibility
    echo "Compiling backend..."
    CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH go build \
        -ldflags "-s -w -X 'tea-api/common.Version=$(cat VERSION)' -extldflags '-static'" \
        -o tea-api
    
    if [ ! -f tea-api ]; then
        echo "Error: Backend build failed - tea-api binary not found"
        exit 1
    fi
    
    # Make binary executable
    chmod +x tea-api
    
    echo "Backend build completed successfully."
}

# Create deployment package
create_package() {
    echo "Creating deployment package..."
    
    # Create deployment directory
    DEPLOY_DIR="tea-api-linux-deploy"
    rm -rf "$DEPLOY_DIR"
    mkdir -p "$DEPLOY_DIR"
    
    # Copy binary
    cp tea-api "$DEPLOY_DIR/"
    
    # Copy web assets
    cp -r web/dist "$DEPLOY_DIR/web/"
    
    # Copy configuration files
    cp docker-compose.yml "$DEPLOY_DIR/"
    cp tea-api.service "$DEPLOY_DIR/"
    
    # Create directories
    mkdir -p "$DEPLOY_DIR/data"
    mkdir -p "$DEPLOY_DIR/logs"
    mkdir -p "$DEPLOY_DIR/tiktoken_cache"
    
    # Create example environment file
    cat > "$DEPLOY_DIR/.env.example" << 'EOF'
# Database configuration
SQL_DSN=./tea-api.db
# For MySQL: SQL_DSN=username:password@tcp(localhost:3306)/database_name
# For PostgreSQL: SQL_DSN=postgres://username:password@localhost:5432/database_name

# Redis configuration (optional)
# REDIS_CONN_STRING=redis://localhost:6379

# Server configuration
TZ=Asia/Shanghai
ERROR_LOG_ENABLED=true
TIKTOKEN_CACHE_DIR=./tiktoken_cache

# Security (required for production)
# SESSION_SECRET=your_random_session_secret_here
# CRYPTO_SECRET=your_random_crypto_secret_here

# Multi-node deployment (optional)
# NODE_TYPE=master
# SYNC_FREQUENCY=60
# FRONTEND_BASE_URL=https://your-domain.com

# Cache configuration
MEMORY_CACHE_ENABLED=true

# Rate limiting
RATE_LIMIT_ENABLED=true
EOF
    
    # Create startup script
    cat > "$DEPLOY_DIR/start.sh" << 'EOF'
#!/bin/bash
set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Create necessary directories
mkdir -p data logs tiktoken_cache

# Start the application
echo "Starting Tea API..."
./tea-api --port 3000 --log-dir ./logs
EOF
    
    chmod +x "$DEPLOY_DIR/start.sh"
    
    # Create systemd service installation script
    cat > "$DEPLOY_DIR/install_service.sh" << 'EOF'
#!/bin/bash
set -e

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (use sudo)"
    exit 1
fi

CURRENT_DIR=$(pwd)
SERVICE_USER=${1:-ubuntu}

echo "Installing Tea API as systemd service..."
echo "Service user: $SERVICE_USER"
echo "Installation directory: $CURRENT_DIR"

# Update service file with correct paths and user
sed -i "s|User=ubuntu|User=$SERVICE_USER|g" tea-api.service
sed -i "s|WorkingDirectory=/path/to/tea-api|WorkingDirectory=$CURRENT_DIR|g" tea-api.service
sed -i "s|ExecStart=/path/to/tea-api/tea-api|ExecStart=$CURRENT_DIR/tea-api|g" tea-api.service

# Copy service file
cp tea-api.service /etc/systemd/system/

# Reload systemd and enable service
systemctl daemon-reload
systemctl enable tea-api
systemctl start tea-api

echo "Service installed and started successfully!"
echo "Use 'sudo systemctl status tea-api' to check status"
echo "Use 'sudo systemctl logs -f tea-api' to view logs"
EOF
    
    chmod +x "$DEPLOY_DIR/install_service.sh"
    
    # Create README
    cat > "$DEPLOY_DIR/README.md" << 'EOF'
# Tea API Linux Deployment

## Quick Start

1. Copy `.env.example` to `.env` and configure your settings
2. Run `./start.sh` to start the application manually
3. Or run `sudo ./install_service.sh` to install as a system service

## Files

- `tea-api` - Main application binary
- `web/` - Frontend assets
- `start.sh` - Manual startup script
- `install_service.sh` - System service installation script
- `tea-api.service` - Systemd service configuration
- `docker-compose.yml` - Docker deployment configuration

## Configuration

Edit `.env` file to configure database, Redis, and other settings.

## System Service

To install as a system service:
```bash
sudo ./install_service.sh [username]
```

Default username is 'ubuntu'. The service will start automatically on boot.

## Manual Start

To start manually:
```bash
./start.sh
```

The application will be available at http://localhost:3000
EOF
    
    echo "Deployment package created: $DEPLOY_DIR"
}

# Main execution
main() {
    check_dependencies
    build_frontend
    build_backend
    create_package
    
    echo ""
    echo "=== Build completed successfully! ==="
    echo "Binary: tea-api"
    echo "Deployment package: tea-api-linux-deploy/"
    echo ""
    echo "To deploy:"
    echo "1. Copy the tea-api-linux-deploy directory to your server"
    echo "2. Configure .env file"
    echo "3. Run ./start.sh or install as service with sudo ./install_service.sh"
}

# Run main function
main "$@"
