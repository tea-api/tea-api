#!/bin/bash
set -e

# Quick deployment script for Tea API on Linux
# This script automates the entire setup and deployment process

echo "=== Tea API Linux Quick Deployment ==="

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

# Check if we're on a supported system
check_system() {
    if [ ! -f /etc/os-release ]; then
        print_error "Cannot determine OS. This script supports Debian/Ubuntu systems."
        exit 1
    fi
    
    . /etc/os-release
    
    case "$ID" in
        ubuntu|debian)
            print_status "Detected $PRETTY_NAME"
            ;;
        *)
            print_warning "This script is designed for Ubuntu/Debian. Your system: $PRETTY_NAME"
            read -p "Continue anyway? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                exit 1
            fi
            ;;
    esac
}

# Setup environment
setup_environment() {
    print_header "Setting up environment"
    
    if [ -f "./bin/setup_env_linux.sh" ]; then
        chmod +x ./bin/setup_env_linux.sh
        ./bin/setup_env_linux.sh
    else
        print_error "setup_env_linux.sh not found. Please run this script from the project root."
        exit 1
    fi
}

# Build application
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

# Configure application
configure_application() {
    print_header "Configuring application"
    
    # Check if deployment package exists
    if [ ! -d "tea-api-linux-deploy" ]; then
        print_error "Deployment package not found. Build may have failed."
        exit 1
    fi
    
    cd tea-api-linux-deploy
    
    # Create .env from example if it doesn't exist
    if [ ! -f .env ]; then
        cp .env.example .env
        print_status "Created .env file from example"
    fi
    
    # Interactive configuration
    echo ""
    print_status "Configuration options:"
    echo "1. Use SQLite (default, no additional setup required)"
    echo "2. Use MySQL (requires MySQL server)"
    echo "3. Use PostgreSQL (requires PostgreSQL server)"
    echo "4. Skip database configuration"
    echo ""
    
    read -p "Choose database option (1-4) [1]: " DB_CHOICE
    DB_CHOICE=${DB_CHOICE:-1}
    
    case $DB_CHOICE in
        2)
            read -p "MySQL host [localhost]: " MYSQL_HOST
            MYSQL_HOST=${MYSQL_HOST:-localhost}
            read -p "MySQL port [3306]: " MYSQL_PORT
            MYSQL_PORT=${MYSQL_PORT:-3306}
            read -p "MySQL database name: " MYSQL_DB
            read -p "MySQL username: " MYSQL_USER
            read -s -p "MySQL password: " MYSQL_PASS
            echo ""
            
            # Update .env file
            sed -i "s|SQL_DSN=.*|SQL_DSN=${MYSQL_USER}:${MYSQL_PASS}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DB}|" .env
            print_status "MySQL configuration updated in .env"
            ;;
        3)
            read -p "PostgreSQL host [localhost]: " PG_HOST
            PG_HOST=${PG_HOST:-localhost}
            read -p "PostgreSQL port [5432]: " PG_PORT
            PG_PORT=${PG_PORT:-5432}
            read -p "PostgreSQL database name: " PG_DB
            read -p "PostgreSQL username: " PG_USER
            read -s -p "PostgreSQL password: " PG_PASS
            echo ""
            
            # Update .env file
            sed -i "s|SQL_DSN=.*|SQL_DSN=postgres://${PG_USER}:${PG_PASS}@${PG_HOST}:${PG_PORT}/${PG_DB}|" .env
            print_status "PostgreSQL configuration updated in .env"
            ;;
        1|4|*)
            print_status "Using SQLite database (default)"
            ;;
    esac
    
    # Redis configuration
    echo ""
    read -p "Do you want to configure Redis? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Redis host [localhost]: " REDIS_HOST
        REDIS_HOST=${REDIS_HOST:-localhost}
        read -p "Redis port [6379]: " REDIS_PORT
        REDIS_PORT=${REDIS_PORT:-6379}
        read -p "Redis password (leave empty if none): " REDIS_PASS
        
        if [ -n "$REDIS_PASS" ]; then
            REDIS_URL="redis://:${REDIS_PASS}@${REDIS_HOST}:${REDIS_PORT}"
        else
            REDIS_URL="redis://${REDIS_HOST}:${REDIS_PORT}"
        fi
        
        # Update .env file
        if grep -q "REDIS_CONN_STRING" .env; then
            sed -i "s|# REDIS_CONN_STRING=.*|REDIS_CONN_STRING=${REDIS_URL}|" .env
        else
            echo "REDIS_CONN_STRING=${REDIS_URL}" >> .env
        fi
        print_status "Redis configuration updated in .env"
    fi
    
    # Generate secrets
    echo ""
    read -p "Generate random secrets for production? (Y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        SESSION_SECRET=$(openssl rand -hex 32)
        CRYPTO_SECRET=$(openssl rand -hex 32)
        
        # Update .env file
        if grep -q "SESSION_SECRET" .env; then
            sed -i "s|# SESSION_SECRET=.*|SESSION_SECRET=${SESSION_SECRET}|" .env
        else
            echo "SESSION_SECRET=${SESSION_SECRET}" >> .env
        fi
        
        if grep -q "CRYPTO_SECRET" .env; then
            sed -i "s|# CRYPTO_SECRET=.*|CRYPTO_SECRET=${CRYPTO_SECRET}|" .env
        else
            echo "CRYPTO_SECRET=${CRYPTO_SECRET}" >> .env
        fi
        
        print_status "Random secrets generated and added to .env"
    fi
    
    cd ..
}

# Deploy application
deploy_application() {
    print_header "Deploying application"
    
    cd tea-api-linux-deploy
    
    echo ""
    print_status "Deployment options:"
    echo "1. Install as system service (recommended for production)"
    echo "2. Start manually (for testing)"
    echo "3. Use Docker deployment"
    echo ""
    
    read -p "Choose deployment option (1-3) [1]: " DEPLOY_CHOICE
    DEPLOY_CHOICE=${DEPLOY_CHOICE:-1}
    
    case $DEPLOY_CHOICE in
        1)
            print_status "Installing as system service..."
            sudo ./install_service.sh
            print_status "Service installed successfully!"
            print_status "Use 'sudo systemctl status tea-api' to check status"
            print_status "Use 'sudo journalctl -u tea-api -f' to view logs"
            ;;
        2)
            print_status "Starting manually..."
            print_warning "This will run in the foreground. Press Ctrl+C to stop."
            sleep 2
            ./start.sh
            ;;
        3)
            print_status "Setting up Docker deployment..."
            if command -v docker >/dev/null 2>&1; then
                if command -v docker-compose >/dev/null 2>&1 || docker compose version >/dev/null 2>&1; then
                    print_status "Starting with Docker Compose..."
                    docker compose up -d
                    print_status "Docker deployment started!"
                    print_status "Use 'docker compose logs -f' to view logs"
                else
                    print_error "Docker Compose not found. Please install Docker Compose."
                    exit 1
                fi
            else
                print_error "Docker not found. Please install Docker first."
                exit 1
            fi
            ;;
        *)
            print_error "Invalid choice"
            exit 1
            ;;
    esac
    
    cd ..
}

# Show final information
show_final_info() {
    print_header "Deployment Complete"
    
    echo ""
    print_status "Tea API has been deployed successfully!"
    echo ""
    print_status "Access your application:"
    echo "- Web interface: http://localhost:3000"
    echo "- API endpoint: http://localhost:3000/api"
    echo ""
    print_status "Configuration file: tea-api-linux-deploy/.env"
    print_status "Logs directory: tea-api-linux-deploy/logs"
    echo ""
    print_status "Useful commands:"
    echo "- Check service status: sudo systemctl status tea-api"
    echo "- View logs: sudo journalctl -u tea-api -f"
    echo "- Restart service: sudo systemctl restart tea-api"
    echo "- Stop service: sudo systemctl stop tea-api"
    echo ""
    print_warning "Remember to:"
    echo "1. Configure your firewall to allow port 3000"
    echo "2. Set up SSL/TLS for production use"
    echo "3. Configure your domain name if needed"
    echo "4. Set up regular backups for your data"
}

# Main execution
main() {
    # Check if running from correct directory
    if [ ! -f "go.mod" ] || [ ! -d "web" ]; then
        print_error "Please run this script from the Tea API project root directory"
        exit 1
    fi
    
    check_system
    setup_environment
    
    # Source bashrc to get updated PATH
    if [ -f ~/.bashrc ]; then
        source ~/.bashrc
    fi
    
    build_application
    configure_application
    deploy_application
    show_final_info
}

# Handle interruption
trap 'print_error "Deployment interrupted"; exit 1' INT TERM

# Run main function
main "$@"
