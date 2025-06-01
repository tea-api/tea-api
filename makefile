FRONTEND_DIR = ./web
BACKEND_DIR = .

.PHONY: all build-frontend start-backend build-linux build-mac setup-linux setup-mac clean

all: build-frontend start-backend

build-frontend:
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && npm install && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) npm run build

start-backend:
	@echo "Starting backend dev server..."
	@cd $(BACKEND_DIR) && go run main.go &

build-linux:
	@echo "Building for Linux..."
	@./bin/build_linux.sh

build-mac:
	@echo "Building for macOS..."
	@./bin/build_mac.sh

setup-linux:
	@echo "Setting up Linux environment..."
	@./bin/setup_env_linux.sh

setup-mac:
	@echo "Setting up macOS environment..."
	@./bin/setup_env_mac.sh

deploy-linux:
	@echo "Quick Linux deployment..."
	@./deploy_linux.sh

clean:
	@echo "Cleaning build artifacts..."
	@rm -f tea-api tea-api-macos
	@rm -rf tea-api-linux-deploy
	@rm -rf $(FRONTEND_DIR)/dist
	@rm -rf $(FRONTEND_DIR)/node_modules
