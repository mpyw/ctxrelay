#!/bin/bash
# Extract all test functions from testdata/src to create an inventory

set -euo pipefail

cd "$(dirname "$0")"

output_file="test-list.txt"

# Remove old inventory if exists
rm -f "$output_file"

# Find all .go files excluding stub files (github.com, golang.org)
find ../src -name "*.go" -type f \
  -not -path "*/github.com/*" \
  -not -path "*/golang.org/*" \
  | sort \
  | while IFS= read -r file; do
    # Extract function definitions
    grep -n "^func " "$file" 2>/dev/null | while IFS=: read -r line_num func_def; do
      # Extract function name (everything between "func " and "(")
      func_name=$(echo "$func_def" | sed 's/^func //' | sed 's/(.*//')
      echo "${file}---${func_name}"
    done
  done | sort > "$output_file"

# Print summary
total=$(wc -l < "$output_file")
echo "✅ Extracted $total test functions to $output_file"
echo ""
echo "Summary by file:"
awk -F'---' '{print $1}' "$output_file" | uniq -c | sort -rn
