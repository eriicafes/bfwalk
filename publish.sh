#!/bin/sh
set -e

# Extract version from package.json
NEW_VERSION=$(grep -m 1 '"version"' package.json | awk -F '"' '{print $4}')

echo "Publishing version v$NEW_VERSION"

# Check if tag already exists
if git rev-parse "v$NEW_VERSION" >/dev/null 2>&1; then
  echo "Tag v$NEW_VERSION already exists, skipping tag creation"
else
  # Create tag
  git tag -a v$NEW_VERSION -m "Release v$NEW_VERSION"
  echo "✓ Created tag v$NEW_VERSION"

  # Push the tag to remote
  git push origin v$NEW_VERSION
  echo "✓ Pushed tag v$NEW_VERSION to remote"
fi
