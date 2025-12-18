#!/usr/bin/env python3
"""
Fix structure.json to use levels map instead of single level field.
Detect actual level from test file location.
"""

import json
from pathlib import Path
import re

def find_function_level(target, func_name):
    """Find which level file contains the function."""
    # File mapping
    file_mapping = {
        'goroutinecreator': ['creator'],
        'goroutinederive': ['goroutinederive'],
        'carrier': ['carrier'],
    }

    levels = file_mapping.get(target, ['basic', 'advanced', 'evil'])

    for level in levels:
        file_path = Path(f"../src/{target}/{level}.go")
        if not file_path.exists():
            continue

        with open(file_path, 'r') as f:
            content = f.read()

        # Check if function exists
        pattern = r'^func\s+(?:\([^)]+\)\s+)?' + re.escape(func_name) + r'\s*\('
        if re.search(pattern, content, re.MULTILINE):
            return level

    return None

def main():
    with open('structure.json', 'r') as f:
        structure = json.load(f)

    for test_name, test in structure['tests'].items():
        if 'level' in test:
            # Old format: single level
            # Convert to levels map
            common_level = test['level']
            levels_map = {}

            # Check each target's actual file
            for target in test['targets']:
                func_name = test['functions'][target]
                actual_level = find_function_level(target, func_name)

                if actual_level:
                    levels_map[target] = actual_level
                else:
                    # Fallback to common level
                    levels_map[target] = common_level
                    print(f"⚠️  Could not find {func_name} for {target} in test {test_name}, using {common_level}")

            test['levels'] = levels_map
            del test['level']

    with open('structure.json', 'w') as f:
        json.dump(structure, f, indent=2)

    print(f"✅ Fixed structure.json with levels map")

if __name__ == '__main__':
    main()
