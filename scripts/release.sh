#!/usr/bin/env bash
#
# release.sh - Automated release script using commitizen
#
# This script:
# 1. Checks for uncommitted changes
# 2. Uses commitizen to determine version bump based on conventional commits
# 3. Generates/updates CHANGELOG.md
# 4. Creates a git tag
# 5. Pushes tag to trigger GitHub Actions release workflow
#
# Requirements:
# - commitizen (pip install commitizen)
# - git
#
# Usage: ./scripts/release.sh [--dry-run]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Flags
DRY_RUN=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [--dry-run]"
            echo ""
            echo "Options:"
            echo "  --dry-run    Show what would be done without making changes"
            echo "  -h, --help   Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}🚀 Starting release process...${NC}"

# Check if commitizen is installed
if ! command -v cz &> /dev/null; then
    echo -e "${RED}❌ Error: commitizen is not installed${NC}"
    echo -e "${YELLOW}Install with: pip install commitizen${NC}"
    exit 1
fi

# Check for uncommitted changes
if [[ -n $(git status -s) ]]; then
    echo -e "${RED}❌ Error: You have uncommitted changes${NC}"
    echo -e "${YELLOW}Please commit or stash your changes before releasing${NC}"
    git status -s
    exit 1
fi

# Check if on main branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    echo -e "${YELLOW}⚠️  Warning: You are not on the main branch (current: $CURRENT_BRANCH)${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}Release cancelled${NC}"
        exit 1
    fi
fi

# Get current version (latest tag or default)
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo -e "${BLUE}📋 Current version: ${GREEN}$CURRENT_VERSION${NC}"

# Show recent commits
echo -e "\n${BLUE}📝 Recent commits since last release:${NC}"
git log $CURRENT_VERSION..HEAD --oneline --no-decorate | head -20

# Get all commits since last tag for changelog
echo -e "\n${BLUE}📊 Analyzing commits...${NC}"
COMMIT_COUNT=$(git rev-list $CURRENT_VERSION..HEAD --count)
echo -e "${GREEN}Found $COMMIT_COUNT commits since last release${NC}"

if [[ $COMMIT_COUNT -eq 0 ]]; then
    echo -e "${YELLOW}⚠️  No new commits to release${NC}"
    exit 0
fi

# Show commit breakdown by type
echo -e "\n${BLUE}Commit breakdown:${NC}"
git log $CURRENT_VERSION..HEAD --pretty=format:"%s" | sed 's/:.*//' | sort | uniq -c | sort -rn

if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "\n${YELLOW}🔍 DRY RUN MODE - No changes will be made${NC}"
    echo -e "\n${BLUE}Would run: cz bump --changelog${NC}"
    cz bump --dry-run --changelog || true
    exit 0
fi

# Confirm release
echo -e "\n${YELLOW}Ready to release. This will:${NC}"
echo "  1. Bump version based on conventional commits"
echo "  2. Update CHANGELOG.md"
echo "  3. Create a git tag"
echo "  4. Push tag to GitHub (triggers release workflow)"
echo ""
read -p "Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Release cancelled${NC}"
    exit 1
fi

# Run commitizen bump
echo -e "\n${BLUE}🔖 Creating release with commitizen...${NC}"
cz bump --changelog --yes

# Get the new version
NEW_VERSION=$(git describe --tags --abbrev=0)
echo -e "\n${GREEN}✅ Version bumped: $CURRENT_VERSION → $NEW_VERSION${NC}"

# Push the tag
echo -e "\n${BLUE}⬆️  Pushing tag to GitHub...${NC}"
git push origin "$NEW_VERSION"

echo -e "\n${GREEN}🎉 Release $NEW_VERSION created successfully!${NC}"
echo -e "${BLUE}📦 GitHub Actions will build and publish the release${NC}"
echo -e "${BLUE}🔗 View release at: https://github.com/S33G/proj/releases/tag/$NEW_VERSION${NC}"
