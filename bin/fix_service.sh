#!/bin/bash
set -e

# Script to diagnose and fix Tea API systemd service issues

echo "=== Tea API Service Diagnostic and Fix Tool ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Get current directory and user info
get_info() {
    SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
    REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
    
    # Get the actual user who called sudo
    if [ -n "$SUDO_USER" ]; then
        ACTUAL_USER="$SUDO_USER"
        ACTUAL_HOME=$(eval echo ~$SUDO_USER)
    else
        ACTUAL_USER=$(whoami)
        ACTUAL_HOME="$HOME"
    fi
    
    print_status "Repository root: $REPO_ROOT"
    print_status "Target user: $ACTUAL_USER"
    print_status "User home: $ACTUAL_HOME"
}

# Diagnose current issues
diagnose_issues() {
    print_header "Diagnosing Issues"
    
    # Check if service file exists
    if [ -f "/etc/systemd/system/tea-api.service" ]; then
        print_status "Service file exists"
        
        # Show current service content
        echo "Current service file content:"
        cat /etc/systemd/system/tea-api.service
        echo ""
    else
        print_warning "Service file does not exist"
    fi
    
    # Check if user exists
    if id "$ACTUAL_USER" &>/dev/null; then
        print_status "User '$ACTUAL_USER' exists"
    else
        print_error "User '$ACTUAL_USER' does not exist"
        return 1
    fi
    
    # Check if binary exists
    if [ -f "$REPO_ROOT/tea-api" ]; then
        print_status "Binary exists at $REPO_ROOT/tea-api"
        
        # Check binary permissions
        if [ -x "$REPO_ROOT/tea-api" ]; then
            print_status "Binary is executable"
        else
            print_warning "Binary is not executable"
        fi
        
        # Check binary owner
        BINARY_OWNER=$(stat -c '%U' "$REPO_ROOT/tea-api")
        print_status "Binary owner: $BINARY_OWNER"
        
    else
        print_error "Binary not found at $REPO_ROOT/tea-api"
        return 1
    fi
    
    # Check directory permissions
    if [ -d "$REPO_ROOT" ]; then
        DIR_OWNER=$(stat -c '%U' "$REPO_ROOT")
        DIR_PERMS=$(stat -c '%a' "$REPO_ROOT")
        print_status "Directory owner: $DIR_OWNER"
        print_status "Directory permissions: $DIR_PERMS"
    fi
    
    # Check if logs directory exists
    if [ ! -d "$REPO_ROOT/logs" ]; then
        print_warning "Logs directory does not exist"
    fi
    
    # Check if data directory exists
    if [ ! -d "$REPO_ROOT/data" ]; then
        print_warning "Data directory does not exist"
    fi
}

# Fix permissions
fix_permissions() {
    print_header "Fixing Permissions"
    
    # Create necessary directories
    mkdir -p "$REPO_ROOT/logs"
    mkdir -p "$REPO_ROOT/data"
    mkdir -p "$REPO_ROOT/tiktoken_cache"
    
    # Set correct ownership
    chown -R "$ACTUAL_USER:$ACTUAL_USER" "$REPO_ROOT"
    
    # Set correct permissions
    chmod +x "$REPO_ROOT/tea-api"
    chmod 755 "$REPO_ROOT"
    chmod 755 "$REPO_ROOT/logs"
    chmod 755 "$REPO_ROOT/data"
    chmod 755 "$REPO_ROOT/tiktoken_cache"
    
    print_status "Permissions fixed"
}

# Create correct service file
create_service_file() {
    print_header "Creating Service File"
    
    # Stop and disable existing service if it exists
    if systemctl is-active --quiet tea-api; then
        print_status "Stopping existing service..."
        systemctl stop tea-api
    fi
    
    if systemctl is-enabled --quiet tea-api; then
        print_status "Disabling existing service..."
        systemctl disable tea-api
    fi
    
    # Create new service file
    cat > /etc/systemd/system/tea-api.service << EOF
[Unit]
Description=Tea API Service
After=network.target

[Service]
Type=simple
User=$ACTUAL_USER
Group=$ACTUAL_USER
WorkingDirectory=$REPO_ROOT
ExecStart=$REPO_ROOT/tea-api --port 3000 --log-dir $REPO_ROOT/logs
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$REPO_ROOT

# Environment
Environment=HOME=$ACTUAL_HOME
Environment=USER=$ACTUAL_USER

[Install]
WantedBy=multi-user.target
EOF
    
    print_status "Service file created"
}

# Test the service
test_service() {
    print_header "Testing Service"
    
    # Reload systemd
    systemctl daemon-reload
    
    # Enable service
    systemctl enable tea-api
    print_status "Service enabled"
    
    # Start service
    print_status "Starting service..."
    systemctl start tea-api
    
    # Wait a moment
    sleep 3
    
    # Check status
    if systemctl is-active --quiet tea-api; then
        print_status "✅ Service is running successfully!"
        
        # Show status
        echo ""
        systemctl status tea-api --no-pager
        
        # Show recent logs
        echo ""
        print_status "Recent logs:"
        journalctl -u tea-api --no-pager -n 10
        
    else
        print_error "❌ Service failed to start"
        
        # Show detailed status
        echo ""
        systemctl status tea-api --no-pager
        
        # Show recent logs
        echo ""
        print_error "Recent error logs:"
        journalctl -u tea-api --no-pager -n 20
        
        return 1
    fi
}

# Test binary directly
test_binary_direct() {
    print_header "Testing Binary Directly"
    
    print_status "Testing binary as user $ACTUAL_USER..."
    
    # Test as the actual user
    sudo -u "$ACTUAL_USER" bash -c "cd '$REPO_ROOT' && timeout 5s ./tea-api --help" || {
        print_warning "Binary help test failed or timed out"
    }
    
    # Test if binary can start
    print_status "Testing binary startup..."
    sudo -u "$ACTUAL_USER" bash -c "cd '$REPO_ROOT' && timeout 3s ./tea-api --port 3001 --log-dir ./logs" &
    BINARY_PID=$!
    
    sleep 2
    
    if kill -0 $BINARY_PID 2>/dev/null; then
        print_status "✅ Binary can start successfully"
        kill $BINARY_PID 2>/dev/null || true
        wait $BINARY_PID 2>/dev/null || true
    else
        print_error "❌ Binary failed to start"
        return 1
    fi
}

# Show useful commands
show_commands() {
    print_header "Useful Commands"
    
    echo "Service management:"
    echo "  sudo systemctl status tea-api      # Check service status"
    echo "  sudo systemctl start tea-api       # Start service"
    echo "  sudo systemctl stop tea-api        # Stop service"
    echo "  sudo systemctl restart tea-api     # Restart service"
    echo "  sudo systemctl enable tea-api      # Enable auto-start"
    echo "  sudo systemctl disable tea-api     # Disable auto-start"
    echo ""
    echo "Logs:"
    echo "  sudo journalctl -u tea-api -f      # Follow service logs"
    echo "  sudo journalctl -u tea-api -n 50   # Show last 50 log entries"
    echo "  tail -f $REPO_ROOT/logs/oneapi-*.log  # Follow application logs"
    echo ""
    echo "Manual testing:"
    echo "  cd $REPO_ROOT && ./tea-api --port 3000 --log-dir ./logs"
}

# Main execution
main() {
    check_root
    get_info
    
    echo ""
    diagnose_issues
    
    echo ""
    read -p "Do you want to fix the issues? (Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        fix_permissions
        create_service_file
        
        echo ""
        read -p "Do you want to test the binary directly first? (Y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            if ! test_binary_direct; then
                print_error "Binary test failed. Please check the binary and try again."
                exit 1
            fi
        fi
        
        echo ""
        read -p "Do you want to start the service now? (Y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            test_service
        fi
    fi
    
    echo ""
    show_commands
}

# Run main function
main "$@"
