#!/bin/bash
set -e

# Tea API Update Script
# This script pulls the latest code, rebuilds the application, and restarts the service
# It preserves existing configuration files

echo "=== Tea API Update Script ==="

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
check_directory() {
    if [ ! -f "go.mod" ] || [ ! -d "web" ]; then
        print_error "Please run this script from the Tea API project root directory"
        exit 1
    fi
}

# Check if git is available and we're in a git repository
check_git() {
    if ! command -v git >/dev/null 2>&1; then
        print_error "Git is not installed. Please install git first."
        exit 1
    fi
    
    if [ ! -d ".git" ]; then
        print_error "This is not a git repository. Cannot pull updates."
        exit 1
    fi
}

# Backup current configuration if deployment exists
backup_config() {
    print_header "Backing up configuration"
    
    DEPLOY_DIR="tea-api-linux-deploy"
    BACKUP_DIR="config-backup-$(date +%Y%m%d-%H%M%S)"
    
    if [ -d "$DEPLOY_DIR" ]; then
        print_status "Found existing deployment directory"
        
        # Create backup directory
        mkdir -p "$BACKUP_DIR"
        
        # Backup configuration files
        if [ -f "$DEPLOY_DIR/.env" ]; then
            cp "$DEPLOY_DIR/.env" "$BACKUP_DIR/"
            print_status "Backed up .env file"
        fi
        
        # Backup database if it exists
        if [ -f "$DEPLOY_DIR/tea-api.db" ]; then
            cp "$DEPLOY_DIR/tea-api.db" "$BACKUP_DIR/"
            print_status "Backed up database file"
        fi
        
        # Backup data directory if it exists
        if [ -d "$DEPLOY_DIR/data" ]; then
            cp -r "$DEPLOY_DIR/data" "$BACKUP_DIR/"
            print_status "Backed up data directory"
        fi
        
        # Backup logs directory if it exists
        if [ -d "$DEPLOY_DIR/logs" ]; then
            cp -r "$DEPLOY_DIR/logs" "$BACKUP_DIR/"
            print_status "Backed up logs directory"
        fi
        
        print_status "Configuration backed up to: $BACKUP_DIR"
        echo "BACKUP_DIR=$BACKUP_DIR" > .update_backup_info
    else
        print_warning "No existing deployment found. This might be the first deployment."
    fi
}

# Pull latest code from git
pull_updates() {
    print_header "Pulling latest code"
    
    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        print_warning "You have uncommitted changes in your working directory."
        read -p "Do you want to stash them and continue? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git stash push -m "Auto-stash before update $(date)"
            print_status "Changes stashed"
        else
            print_error "Update cancelled. Please commit or stash your changes first."
            exit 1
        fi
    fi
    
    # Get current branch
    CURRENT_BRANCH=$(git branch --show-current)
    print_status "Current branch: $CURRENT_BRANCH"
    
    # Pull latest changes
    print_status "Pulling latest changes..."
    if git pull origin "$CURRENT_BRANCH"; then
        print_status "Code updated successfully"
    else
        print_error "Failed to pull updates. Please check your git configuration."
        exit 1
    fi
}

# Stop the service if it's running
stop_service() {
    print_header "Stopping service"
    
    if systemctl is-active --quiet tea-api 2>/dev/null; then
        print_status "Stopping tea-api service..."
        sudo systemctl stop tea-api
        print_status "Service stopped"
    else
        print_status "Service is not running"
    fi
}

# Build the application
build_application() {
    print_header "Building application"
    
    if [ -f "./bin/build_linux.sh" ]; then
        chmod +x ./bin/build_linux.sh
        ./bin/build_linux.sh
    else
        print_error "build_linux.sh not found. Please run this script from the project root."
        exit 1
    fi
}

# Restore configuration files
restore_config() {
    print_header "Restoring configuration"
    
    if [ -f ".update_backup_info" ]; then
        source .update_backup_info
        
        DEPLOY_DIR="tea-api-linux-deploy"
        
        if [ -d "$BACKUP_DIR" ] && [ -d "$DEPLOY_DIR" ]; then
            # Restore .env file
            if [ -f "$BACKUP_DIR/.env" ]; then
                cp "$BACKUP_DIR/.env" "$DEPLOY_DIR/"
                print_status "Restored .env file"
            fi
            
            # Restore database
            if [ -f "$BACKUP_DIR/tea-api.db" ]; then
                cp "$BACKUP_DIR/tea-api.db" "$DEPLOY_DIR/"
                print_status "Restored database file"
            fi
            
            # Restore data directory
            if [ -d "$BACKUP_DIR/data" ]; then
                cp -r "$BACKUP_DIR/data" "$DEPLOY_DIR/"
                print_status "Restored data directory"
            fi
            
            # Restore logs directory (merge with new logs)
            if [ -d "$BACKUP_DIR/logs" ]; then
                cp -r "$BACKUP_DIR/logs"/* "$DEPLOY_DIR/logs/" 2>/dev/null || true
                print_status "Restored logs directory"
            fi
            
            print_status "Configuration restored successfully"
        else
            print_warning "Backup directory not found. Using default configuration."
        fi
        
        # Clean up backup info file
        rm -f .update_backup_info
    else
        print_status "No backup to restore"
    fi
}

# Start the service
start_service() {
    print_header "Starting service"

    DEPLOY_DIR="tea-api-linux-deploy"

    if [ -d "$DEPLOY_DIR" ]; then
        # Check if systemd service exists
        if [ -f "/etc/systemd/system/tea-api.service" ]; then
            print_status "Starting tea-api service..."
            sudo systemctl start tea-api

            # Wait a moment and check if service started successfully
            sleep 3
            if systemctl is-active --quiet tea-api; then
                print_status "Service started successfully"
            else
                print_error "Service failed to start. Check logs with: sudo journalctl -u tea-api -f"
                exit 1
            fi
        else
            print_warning "Systemd service not found. You can start manually with:"
            print_status "cd $DEPLOY_DIR && ./start.sh"
        fi
    else
        print_error "Deployment directory not found. Build may have failed."
        exit 1
    fi
}

# Clean up old backups (keep only last 5)
cleanup_backups() {
    print_header "Cleaning up old backups"

    # Find and remove old backup directories (keep only the 5 most recent)
    BACKUP_DIRS=$(ls -dt config-backup-* 2>/dev/null | tail -n +6)

    if [ -n "$BACKUP_DIRS" ]; then
        print_status "Removing old backup directories..."
        echo "$BACKUP_DIRS" | while read -r dir; do
            if [ -d "$dir" ]; then
                rm -rf "$dir"
                print_status "Removed old backup: $dir"
            fi
        done
    else
        print_status "No old backups to clean up"
    fi
}

# Show update summary
show_summary() {
    print_header "Update Complete"

    echo ""
    print_status "Tea API has been updated successfully!"
    echo ""

    # Show version information if available
    if [ -f "VERSION" ]; then
        VERSION=$(cat VERSION)
        print_status "Current version: $VERSION"
    fi

    # Show service status
    if systemctl is-active --quiet tea-api 2>/dev/null; then
        print_status "Service status: Running"
        print_status "Access your application at: http://localhost:3000"
    else
        print_warning "Service status: Not running"
        print_status "Start manually with: cd tea-api-linux-deploy && ./start.sh"
    fi

    echo ""
    print_status "Useful commands:"
    echo "- Check service status: sudo systemctl status tea-api"
    echo "- View logs: sudo journalctl -u tea-api -f"
    echo "- Restart service: sudo systemctl restart tea-api"
    echo ""

    # Show backup information
    if [ -d "config-backup-"* ] 2>/dev/null; then
        LATEST_BACKUP=$(ls -dt config-backup-* 2>/dev/null | head -n 1)
        print_status "Latest backup: $LATEST_BACKUP"
    fi
}

# Main execution function
main() {
    print_header "Starting Tea API Update"

    # Preliminary checks
    check_directory
    check_git

    # Show current status
    echo ""
    print_status "Current directory: $(pwd)"
    if [ -f "VERSION" ]; then
        print_status "Current version: $(cat VERSION)"
    fi

    # Confirm update
    echo ""
    read -p "Do you want to proceed with the update? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Update cancelled by user"
        exit 0
    fi

    # Execute update steps
    backup_config
    stop_service
    pull_updates
    build_application
    restore_config
    start_service
    cleanup_backups
    show_summary
}

# Handle interruption
trap 'print_error "Update interrupted"; exit 1' INT TERM

# Parse command line arguments
FORCE_UPDATE=false
SKIP_BACKUP=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --force)
            FORCE_UPDATE=true
            shift
            ;;
        --skip-backup)
            SKIP_BACKUP=true
            shift
            ;;
        --help|-h)
            echo "Tea API Update Script"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --force        Force update without confirmation"
            echo "  --skip-backup  Skip configuration backup"
            echo "  --help, -h     Show this help message"
            echo ""
            echo "This script will:"
            echo "1. Backup current configuration"
            echo "2. Pull latest code from git"
            echo "3. Rebuild the application"
            echo "4. Restore configuration"
            echo "5. Restart the service"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Modify main function for force mode
if [ "$FORCE_UPDATE" = true ]; then
    print_warning "Force mode enabled - skipping confirmation"
    check_directory
    check_git

    if [ "$SKIP_BACKUP" = false ]; then
        backup_config
    else
        print_warning "Skipping backup as requested"
    fi

    stop_service
    pull_updates
    build_application
    restore_config
    start_service
    cleanup_backups
    show_summary
else
    # Run normal interactive mode
    main "$@"
fi
