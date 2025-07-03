NEW_VERSION=$(grep -m 1 '"version"' package.json | awk -F '"' '{print $4}')
echo "New version for 'bfwalk' is v$NEW_VERSION"
git tag v$NEW_VERSION -m "v$NEW_VERSION"