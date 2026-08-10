#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to get the latest tag
get_latest_tag() {
    git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0"
}

# Function to increment version
increment_version() {
    local version=$1
    local increment_type=$2
    
    IFS='.' read -ra VERSION_PARTS <<< "$version"
    local major=${VERSION_PARTS[0]:-0}
    local minor=${VERSION_PARTS[1]:-0}
    local patch=${VERSION_PARTS[2]:-0}
    
    case $increment_type in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch")
            patch=$((patch + 1))
            ;;
        *)
            print_error "Invalid increment type. Use: major, minor, or patch"
            exit 1
            ;;
    esac
    
    echo "$major.$minor.$patch"
}

# Function to confirm action
confirm_action() {
    local message=$1
    echo -n "$message (yes/no): "
    read -r response
    if [[ "$response" != "yes" ]]; then
        print_warning "Operation cancelled by user"
        exit 0
    fi
}

# Main script
main() {
    # Check if git is available
    if ! command -v git &> /dev/null; then
        print_error "Git is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
    
    # Check if there are uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        print_error "You have uncommitted changes. Please commit or stash them first."
        exit 1
    fi
    
    # Get increment type from argument
    if [ $# -ne 1 ]; then
        print_error "Usage: $0 <major|minor|patch>"
        print_info "Example: $0 minor"
        exit 1
    fi
    
    local increment_type=$1
    
    # Validate increment type
    if [[ ! "$increment_type" =~ ^(major|minor|patch)$ ]]; then
        print_error "Invalid increment type: $increment_type"
        print_error "Valid options: major, minor, patch"
        exit 1
    fi
    
    # Get current latest tag
    local current_tag=$(get_latest_tag)
    print_info "Current latest tag: $current_tag"
    
    # Calculate new tag
    local new_tag=$(increment_version "$current_tag" "$increment_type")
    print_info "New tag will be: $new_tag"
    
    # Show what will happen
    echo
    print_warning "This will:"
    echo "  1. Create a new tag: $new_tag"
    echo "  2. Push the tag to origin"
    echo
    
    # Ask for confirmation
    confirm_action "Do you want to proceed?"
    
    # Create the tag
    print_info "Creating tag: $new_tag"
    git tag -a "$new_tag" -m "Release version $new_tag"
    
    # Push the tag
    print_info "Pushing tag to origin..."
    git push origin "$new_tag"
    
    print_info "Successfully created and pushed tag: $new_tag"
}

# Run main function
main "$@"
