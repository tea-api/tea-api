#!/bin/bash
set -e

# Tea API Quick Update Script
# This is a simplified version that performs a quick update without prompts
# Usage: ./bin/quick_update.sh

echo "=== Tea API Quick Update ==="

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

# Check if we're in the correct directory
if [ ! -f "go.mod" ] || [ ! -d "web" ]; then
    print_error "Please run this script from the Tea API project root directory"
    exit 1
fi

# Check git
if ! command -v git >/dev/null 2>&1; then
    print_error "Git is not installed"
    exit 1
fi

if [ ! -d ".git" ]; then
    print_error "This is not a git repository"
    exit 1
fi

print_status "Starting quick update..."

# Backup configuration
print_header "Backing up configuration"
DEPLOY_DIR="tea-api-linux-deploy"
BACKUP_DIR="config-backup-$(date +%Y%m%d-%H%M%S)"

if [ -d "$DEPLOY_DIR" ]; then
    mkdir -p "$BACKUP_DIR"
    
    # Backup important files
    [ -f "$DEPLOY_DIR/.env" ] && cp "$DEPLOY_DIR/.env" "$BACKUP_DIR/"
    [ -f "$DEPLOY_DIR/tea-api.db" ] && cp "$DEPLOY_DIR/tea-api.db" "$BACKUP_DIR/"
    [ -d "$DEPLOY_DIR/data" ] && cp -r "$DEPLOY_DIR/data" "$BACKUP_DIR/"
    [ -d "$DEPLOY_DIR/logs" ] && cp -r "$DEPLOY_DIR/logs" "$BACKUP_DIR/"
    
    print_status "Configuration backed up to: $BACKUP_DIR"
    echo "BACKUP_DIR=$BACKUP_DIR" > .update_backup_info
fi

# Stop service
print_header "Stopping service"
if systemctl is-active --quiet tea-api 2>/dev/null; then
    print_status "Stopping tea-api service..."
    sudo systemctl stop tea-api
fi

# Pull updates
print_header "Pulling latest code"
CURRENT_BRANCH=$(git branch --show-current)
print_status "Pulling from branch: $CURRENT_BRANCH"

if ! git diff-index --quiet HEAD --; then
    print_warning "Stashing uncommitted changes..."
    git stash push -m "Auto-stash before quick update $(date)"
fi

git pull origin "$CURRENT_BRANCH"
print_status "Code updated successfully"

# Build
print_header "Building application"
chmod +x ./bin/build_linux.sh
./bin/build_linux.sh

# Restore configuration
print_header "Restoring configuration"
if [ -f ".update_backup_info" ]; then
    source .update_backup_info
    
    if [ -d "$BACKUP_DIR" ] && [ -d "$DEPLOY_DIR" ]; then
        [ -f "$BACKUP_DIR/.env" ] && cp "$BACKUP_DIR/.env" "$DEPLOY_DIR/"
        [ -f "$BACKUP_DIR/tea-api.db" ] && cp "$BACKUP_DIR/tea-api.db" "$DEPLOY_DIR/"
        [ -d "$BACKUP_DIR/data" ] && cp -r "$BACKUP_DIR/data" "$DEPLOY_DIR/"
        [ -d "$BACKUP_DIR/logs" ] && cp -r "$BACKUP_DIR/logs"/* "$DEPLOY_DIR/logs/" 2>/dev/null || true
        
        print_status "Configuration restored"
    fi
    
    rm -f .update_backup_info
fi

# Start service
print_header "Starting service"
if [ -f "/etc/systemd/system/tea-api.service" ]; then
    sudo systemctl start tea-api
    sleep 3
    
    if systemctl is-active --quiet tea-api; then
        print_status "Service started successfully"
    else
        print_error "Service failed to start. Check logs: sudo journalctl -u tea-api -f"
        exit 1
    fi
else
    print_warning "Systemd service not found. Start manually: cd tea-api-linux-deploy && ./start.sh"
fi

# Clean up old backups (keep only last 3)
BACKUP_DIRS=$(ls -dt config-backup-* 2>/dev/null | tail -n +4)
if [ -n "$BACKUP_DIRS" ]; then
    echo "$BACKUP_DIRS" | while read -r dir; do
        [ -d "$dir" ] && rm -rf "$dir"
    done
fi

print_header "Update Complete"
print_status "Tea API has been updated successfully!"

if [ -f "VERSION" ]; then
    print_status "Current version: $(cat VERSION)"
fi

if systemctl is-active --quiet tea-api 2>/dev/null; then
    print_status "Service is running - Access at: http://localhost:3000"
else
    print_warning "Service is not running"
fi

print_status "Latest backup: $BACKUP_DIR"
