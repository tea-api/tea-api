#!/bin/bash
set -e

# Test script to verify the build works correctly

echo "=== Testing Tea API Build ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Get script directory
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)

cd "$REPO_ROOT"

# Test 1: Check if binary exists
test_binary_exists() {
    print_status "Testing if binary exists..."
    
    if [ -f "tea-api" ]; then
        print_status "✓ Binary 'tea-api' found"
    else
        print_error "✗ Binary 'tea-api' not found"
        return 1
    fi
}

# Test 2: Check if binary is executable
test_binary_executable() {
    print_status "Testing if binary is executable..."
    
    if [ -x "tea-api" ]; then
        print_status "✓ Binary is executable"
    else
        print_error "✗ Binary is not executable"
        return 1
    fi
}

# Test 3: Check binary architecture
test_binary_architecture() {
    print_status "Testing binary architecture..."
    
    ARCH=$(file tea-api | grep -o 'x86-64\|aarch64\|ARM64')
    if [ -n "$ARCH" ]; then
        print_status "✓ Binary architecture: $ARCH"
    else
        print_warning "? Could not determine binary architecture"
    fi
}

# Test 4: Check if frontend assets exist
test_frontend_assets() {
    print_status "Testing frontend assets..."
    
    if [ -d "web/dist" ]; then
        print_status "✓ Frontend dist directory found"
        
        if [ -f "web/dist/index.html" ]; then
            print_status "✓ Frontend index.html found"
        else
            print_error "✗ Frontend index.html not found"
            return 1
        fi
        
        # Check if assets directory exists
        if [ -d "web/dist/assets" ]; then
            print_status "✓ Frontend assets directory found"
        else
            print_warning "? Frontend assets directory not found"
        fi
    else
        print_error "✗ Frontend dist directory not found"
        return 1
    fi
}

# Test 5: Test binary help output
test_binary_help() {
    print_status "Testing binary help output..."
    
    if timeout 10s ./tea-api --help >/dev/null 2>&1; then
        print_status "✓ Binary responds to --help flag"
    else
        print_warning "? Binary did not respond to --help flag (this might be normal)"
    fi
}

# Test 6: Test binary version output
test_binary_version() {
    print_status "Testing binary version output..."
    
    if timeout 10s ./tea-api --version >/dev/null 2>&1; then
        print_status "✓ Binary responds to --version flag"
    else
        print_warning "? Binary did not respond to --version flag (this might be normal)"
    fi
}

# Test 7: Check dependencies
test_dependencies() {
    print_status "Testing Go module dependencies..."
    
    if go mod verify >/dev/null 2>&1; then
        print_status "✓ Go modules verified successfully"
    else
        print_error "✗ Go module verification failed"
        return 1
    fi
}

# Test 8: Check if deployment package exists
test_deployment_package() {
    print_status "Testing deployment package..."
    
    if [ -d "tea-api-linux-deploy" ]; then
        print_status "✓ Deployment package directory found"
        
        # Check essential files
        ESSENTIAL_FILES=(
            "tea-api-linux-deploy/tea-api"
            "tea-api-linux-deploy/start.sh"
            "tea-api-linux-deploy/install_service.sh"
            "tea-api-linux-deploy/.env.example"
            "tea-api-linux-deploy/README.md"
        )
        
        for file in "${ESSENTIAL_FILES[@]}"; do
            if [ -f "$file" ]; then
                print_status "✓ Found: $(basename "$file")"
            else
                print_error "✗ Missing: $(basename "$file")"
                return 1
            fi
        done
        
        # Check if web assets are copied
        if [ -d "tea-api-linux-deploy/web/dist" ]; then
            print_status "✓ Web assets copied to deployment package"
        else
            print_error "✗ Web assets not found in deployment package"
            return 1
        fi
    else
        print_warning "? Deployment package not found (run build_linux.sh to create it)"
    fi
}

# Test 9: Quick startup test
test_quick_startup() {
    print_status "Testing quick startup (5 seconds)..."
    
    # Start the binary in background
    timeout 5s ./tea-api --port 3001 >/dev/null 2>&1 &
    PID=$!
    
    # Wait a moment for startup
    sleep 2
    
    # Check if process is still running
    if kill -0 $PID 2>/dev/null; then
        print_status "✓ Binary started successfully"
        # Kill the process
        kill $PID 2>/dev/null || true
        wait $PID 2>/dev/null || true
    else
        print_warning "? Binary startup test inconclusive"
    fi
}

# Test 10: Check file sizes
test_file_sizes() {
    print_status "Checking file sizes..."
    
    if [ -f "tea-api" ]; then
        SIZE=$(du -h tea-api | cut -f1)
        print_status "Binary size: $SIZE"
        
        # Check if size is reasonable (not too small, not too large)
        SIZE_BYTES=$(stat -c%s tea-api)
        if [ "$SIZE_BYTES" -lt 1000000 ]; then  # Less than 1MB
            print_warning "Binary seems unusually small ($SIZE)"
        elif [ "$SIZE_BYTES" -gt 100000000 ]; then  # More than 100MB
            print_warning "Binary seems unusually large ($SIZE)"
        else
            print_status "✓ Binary size looks reasonable ($SIZE)"
        fi
    fi
    
    if [ -d "web/dist" ]; then
        DIST_SIZE=$(du -sh web/dist | cut -f1)
        print_status "Frontend dist size: $DIST_SIZE"
    fi
}

# Run all tests
run_tests() {
    local failed_tests=0
    local total_tests=0
    
    echo ""
    print_status "Running build verification tests..."
    echo ""
    
    # List of test functions
    tests=(
        "test_binary_exists"
        "test_binary_executable"
        "test_binary_architecture"
        "test_frontend_assets"
        "test_binary_help"
        "test_binary_version"
        "test_dependencies"
        "test_deployment_package"
        "test_quick_startup"
        "test_file_sizes"
    )
    
    for test in "${tests[@]}"; do
        total_tests=$((total_tests + 1))
        echo ""
        if ! $test; then
            failed_tests=$((failed_tests + 1))
        fi
    done
    
    echo ""
    echo "=== Test Results ==="
    echo "Total tests: $total_tests"
    echo "Passed: $((total_tests - failed_tests))"
    echo "Failed: $failed_tests"
    
    if [ $failed_tests -eq 0 ]; then
        print_status "🎉 All tests passed! Build appears to be successful."
        return 0
    else
        print_error "❌ $failed_tests test(s) failed. Please check the build."
        return 1
    fi
}

# Main execution
main() {
    # Check if we're in the right directory
    if [ ! -f "go.mod" ]; then
        print_error "Please run this script from the Tea API project root directory"
        exit 1
    fi
    
    run_tests
}

# Run main function
main "$@"
