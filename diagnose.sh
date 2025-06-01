#!/bin/bash

# Quick diagnostic script for Tea API service issues
# Run this script to identify the problem

echo "=== Tea API Service Diagnostic ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 1. Check current user
echo "1. Current user information:"
echo "   Current user: $(whoami)"
echo "   User ID: $(id)"
echo "   Home directory: $HOME"
echo ""

# 2. Check if tea-api binary exists
echo "2. Checking tea-api binary:"
if [ -f "./tea-api" ]; then
    print_status "Binary found: ./tea-api"
    echo "   File permissions: $(ls -la tea-api)"
    echo "   File type: $(file tea-api)"
    
    if [ -x "./tea-api" ]; then
        print_status "Binary is executable"
    else
        print_error "Binary is not executable"
    fi
else
    print_error "Binary not found in current directory"
    echo "   Current directory: $(pwd)"
    echo "   Files in directory: $(ls -la)"
fi
echo ""

# 3. Check service file
echo "3. Checking systemd service:"
if [ -f "/etc/systemd/system/tea-api.service" ]; then
    print_status "Service file exists"
    echo "   Service file content:"
    cat /etc/systemd/system/tea-api.service
else
    print_error "Service file not found"
fi
echo ""

# 4. Check service status
echo "4. Service status:"
if command -v systemctl >/dev/null 2>&1; then
    systemctl status tea-api --no-pager || true
else
    print_error "systemctl not available"
fi
echo ""

# 5. Check recent logs
echo "5. Recent service logs:"
if command -v journalctl >/dev/null 2>&1; then
    journalctl -u tea-api --no-pager -n 10 || true
else
    print_error "journalctl not available"
fi
echo ""

# 6. Check directories
echo "6. Checking directories:"
for dir in "logs" "data" "tiktoken_cache"; do
    if [ -d "$dir" ]; then
        print_status "Directory exists: $dir"
        echo "   Permissions: $(ls -ld $dir)"
    else
        print_warning "Directory missing: $dir"
    fi
done
echo ""

# 7. Check environment
echo "7. Environment check:"
echo "   PATH: $PATH"
echo "   Working directory: $(pwd)"
echo "   Disk space: $(df -h . | tail -1)"
echo "   Memory: $(free -h | head -2)"
echo ""

# 8. Test binary directly
echo "8. Testing binary directly:"
if [ -f "./tea-api" ] && [ -x "./tea-api" ]; then
    echo "   Testing --help flag:"
    timeout 5s ./tea-api --help 2>&1 || echo "   Help test failed or timed out"
    
    echo "   Testing version flag:"
    timeout 5s ./tea-api --version 2>&1 || echo "   Version test failed or timed out"
else
    print_error "Cannot test binary - not found or not executable"
fi
echo ""

# 9. Suggested fixes
echo "=== Suggested Fixes ==="
echo ""

if [ ! -f "./tea-api" ]; then
    print_error "ISSUE: Binary not found"
    echo "   FIX: Build the application first:"
    echo "        ./bin/build_linux.sh"
    echo ""
fi

if [ -f "./tea-api" ] && [ ! -x "./tea-api" ]; then
    print_error "ISSUE: Binary not executable"
    echo "   FIX: Make binary executable:"
    echo "        chmod +x ./tea-api"
    echo ""
fi

if [ -f "/etc/systemd/system/tea-api.service" ]; then
    # Check if service file has placeholder paths
    if grep -q "/path/to/tea-api" /etc/systemd/system/tea-api.service; then
        print_error "ISSUE: Service file has placeholder paths"
        echo "   FIX: Update service file with correct paths"
        echo ""
    fi
    
    # Check if user in service file exists
    SERVICE_USER=$(grep "^User=" /etc/systemd/system/tea-api.service | cut -d'=' -f2 | tr -d ' ')
    if [ -n "$SERVICE_USER" ]; then
        if ! id "$SERVICE_USER" &>/dev/null; then
            print_error "ISSUE: Service user '$SERVICE_USER' does not exist"
            echo "   FIX: Change user in service file to existing user"
            echo ""
        fi
    fi
fi

echo "To automatically fix these issues, run:"
echo "   sudo ./bin/fix_service.sh"
echo ""
echo "Or manually fix the service file and run:"
echo "   sudo systemctl daemon-reload"
echo "   sudo systemctl restart tea-api"
