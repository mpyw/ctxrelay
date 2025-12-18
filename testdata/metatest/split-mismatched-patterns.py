#!/usr/bin/env python3
"""
Split mismatched unified patterns into separate target-specific patterns.
If a unified pattern has different descriptions or missing functions,
split it into separate patterns (GO##, GE##, GW##).
"""

import json
import re
from pathlib import Path

def extract_original_pattern_code(target, func_name):
    """Extract original pattern code from git history."""
    import subprocess

    target_to_prefix = {
        'goroutine': 'GO',
        'errgroup': 'GE',
        'waitgroup': 'GW'
    }

    prefix = target_to_prefix.get(target)
    if not prefix:
        return None

    # Determine file based on current function location
    for level in ['basic', 'advanced', 'evil']:
        file_path = f"testdata/src/{target}/{level}.go"

        try:
            # Get file from HEAD
            result = subprocess.run(
                ['git', 'show', f'HEAD:{file_path}'],
                capture_output=True,
                text=True,
                check=True
            )
            content = result.stdout

            # Find pattern code before function
            pattern = rf'^//\s*({prefix}\d+[a-z]?):\s*(.+)\nfunc\s+(?:\([^)]+\)\s+)?{re.escape(func_name)}\s*\('
            match = re.search(pattern, content, re.MULTILINE)
            if match:
                return match.group(1), match.group(2).strip()
        except:
            continue

    return None

def main():
    with open('structure.json', 'r') as f:
        structure = json.load(f)

    new_tests = {}

    for test_name, test in structure['tests'].items():
        # Only process unified patterns (pattern##)
        if not test_name.startswith('pattern'):
            new_tests[test_name] = test
            continue

        # Check if all targets have the same description
        descriptions = {}
        pattern_codes = {}

        for target in test['targets']:
            func_name = test['functions'][target]
            result = extract_original_pattern_code(target, func_name)

            if result:
                code, desc = result
                descriptions[target] = desc
                pattern_codes[target] = code

        # If descriptions match, keep unified
        unique_descs = set(descriptions.values())
        if len(unique_descs) <= 1:
            new_tests[test_name] = test
            continue

        # Split into separate patterns
        print(f"⚠️  Splitting {test_name} - different descriptions:")
        for target in test['targets']:
            if target in pattern_codes:
                code = pattern_codes[target]
                desc = descriptions[target]
                print(f"  {code}: {desc}")

                # Create separate pattern
                new_tests[code] = {
                    'title': test['title'],  # Keep original title
                    'description': desc,
                    'targets': [target],
                    'functions': {target: test['functions'][target]},
                    'levels': {target: test['levels'][target]}
                }

    # Write updated structure
    with open('structure.json', 'w') as f:
        json.dump({'targets': structure['targets'], 'tests': new_tests}, f, indent=2)

    print(f"\n✅ Updated structure.json")
    print(f"   Before: {len(structure['tests'])} patterns")
    print(f"   After: {len(new_tests)} patterns")

if __name__ == '__main__':
    main()
