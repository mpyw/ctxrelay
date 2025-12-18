#!/bin/bash
#
# verify-migration.sh
# Verifies that all tests from test-list.txt are accounted for in structure.json
#

set -e

TEST_LIST="test-list.txt"
STRUCTURE_JSON="structure.json"

echo "=== Migration Verification ==="
echo ""

# Extract all function names from structure.json
registered_funcs=$(jq -r '.tests | keys[]' "$STRUCTURE_JSON" | sort -u)

# Count
registered_count=$(echo "$registered_funcs" | wc -l | tr -d ' ')
echo "Registered tests in structure.json: $registered_count"

# Process test-list.txt
total=0
missing=0
helper=0
registered=0

echo ""
echo "=== Checking test-list.txt entries ==="

while IFS= read -r line; do
    # Extract path and function name using awk
    path=$(echo "$line" | awk -F'---' '{print $1}')
    func=$(echo "$line" | awk -F'---' '{print $2}')
    target=$(echo "$path" | sed 's|testdata/src/||' | cut -d'/' -f1)
    
    ((total++)) || true
    
    # Skip helper functions (not starting with bad/good/evil/limitation/twoContextParams/notChecked)
    if ! echo "$func" | grep -qE '^(bad|good|evil|limitation|twoContextParams|notChecked)'; then
        ((helper++)) || true
        continue
    fi
    
    # Check if function is in structure.json
    if echo "$registered_funcs" | grep -q "^${func}$"; then
        ((registered++)) || true
    else
        echo "MISSING: $target/$func"
        ((missing++)) || true
    fi
done < "$TEST_LIST"

echo ""
echo "=== Summary ==="
echo "Total entries in test-list.txt: $total"
echo "Helper functions (skipped): $helper"
echo "Registered in structure.json: $registered"
echo "Missing from structure.json: $missing"

if [ $missing -eq 0 ]; then
    echo ""
    echo "✅ All test functions are accounted for!"
    exit 0
else
    echo ""
    echo "❌ Some test functions are missing!"
    exit 1
fi
