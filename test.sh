#!/bin/bash

# test_api.sh

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API URL
API_URL="http://localhost:8080"
TOKEN=""

# Print with color
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}➜ $1${NC}"
}

# Function to check if the API is running
check_api() {
    print_info "Checking if API is running..."
    if curl -s "$API_URL/auth" > /dev/null; then
        print_success "API is running"
        return 0
    else
        print_error "API is not running. Please start the API first"
        exit 1
    fi
}

# Function to test authentication endpoint
test_auth() {
    print_info "Testing /auth endpoint..."
    
    # Test POST method
    response=$(curl -s -X POST "$API_URL/auth")
    if echo "$response" | grep -q "token"; then
        print_success "Authentication successful"
        TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | grep -o '[^"]*$')
    else
        print_error "Authentication failed"
        exit 1
    fi

    # Test wrong method (GET)
    response=$(curl -s -X GET "$API_URL/auth")
    if echo "$response" | grep -q "Method not allowed"; then
        print_success "Wrong method handling working correctly"
    else
        print_error "Wrong method handling failed"
    fi
}

# Function to test generate-deck endpoint
test_generate_deck() {
    print_info "Testing /generate-deck endpoint..."

    # Test without token
    response=$(curl -s -X GET "$API_URL/generate-deck")
    if echo "$response" | grep -q "Authorization header required"; then
        print_success "Authorization check working correctly"
    else
        print_error "Authorization check failed"
    fi

    # Test with invalid token
    response=$(curl -s -X GET "$API_URL/generate-deck" \
        -H "Authorization: Bearer invalid_token")
    if echo "$response" | grep -q "Invalid token"; then
        print_success "Invalid token check working correctly"
    else
        print_error "Invalid token check failed"
    fi

    # Test with valid token
    response=$(curl -s -X GET "$API_URL/generate-deck" \
        -H "Authorization: Bearer $TOKEN")
    if echo "$response" | grep -q "Generate deck endpoint"; then
        print_success "Generate deck endpoint working correctly"
    else
        print_error "Generate deck endpoint failed"
    fi

    # Test wrong method (POST)
    response=$(curl -s -X POST "$API_URL/generate-deck" \
        -H "Authorization: Bearer $TOKEN")
    if echo "$response" | grep -q "Method not allowed"; then
        print_success "Wrong method handling working correctly"
    else
        print_error "Wrong method handling failed"
    fi
}

# Function to test prompt-suggestion endpoint
test_prompt_suggestion() {
    print_info "Testing /prompt-suggestion endpoint..."

    # Test without token
    response=$(curl -s -X POST "$API_URL/prompt-suggestion")
    if echo "$response" | grep -q "Authorization header required"; then
        print_success "Authorization check working correctly"
    else
        print_error "Authorization check failed"
    fi

    # Test with invalid token
    response=$(curl -s -X POST "$API_URL/prompt-suggestion" \
        -H "Authorization: Bearer invalid_token")
    if echo "$response" | grep -q "Invalid token"; then
        print_success "Invalid token check working correctly"
    else
        print_error "Invalid token check failed"
    fi

    # Test with valid token
    response=$(curl -s -X POST "$API_URL/prompt-suggestion" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"prompt":"test prompt"}')
    if echo "$response" | grep -q "Prompt suggestion endpoint"; then
        print_success "Prompt suggestion endpoint working correctly"
    else
        print_error "Prompt suggestion endpoint failed"
    fi

    # Test wrong method (GET)
    response=$(curl -s -X GET "$API_URL/prompt-suggestion" \
        -H "Authorization: Bearer $TOKEN")
    if echo "$response" | grep -q "Method not allowed"; then
        print_success "Wrong method handling working correctly"
    else
        print_error "Wrong method handling failed"
    fi
}

# Main test execution
main() {
    echo "Starting API Tests..."
    echo "===================="
    
    check_api
    
    echo
    test_auth
    
    echo
    test_generate_deck
    
    echo
    test_prompt_suggestion
    
    echo
    print_success "All tests completed!"
}

# Run main function
main
