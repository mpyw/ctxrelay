#!/usr/bin/env python3
"""
Remove //vt:helper from functions that are in structure.json (they shouldn't be helpers).
"""

import json
import re
from pathlib import Path

def load_structure():
    with open('structure.json', 'r') as f:
        return json.load(f)

def remove_helpers_from_file(file_path, expected_funcs):
    """Remove //vt:helper from test functions."""
    if not file_path.exists():
        return 0

    with open(file_path, 'r') as f:
        content = f.read()

    removed_count = 0

    # For each expected function, remove //vt:helper if present
    for func_name in expected_funcs:
        # Pattern: match //vt:helper followed by any comments, then func declaration
        pattern = r'//vt:helper\n((?://.*\n)*)(func\s+(?:\([^)]+\)\s+)?' + re.escape(func_name) + r'\s*\()'

        def replace_func(match):
            nonlocal removed_count
            removed_count += 1
            comments = match.group(1)  # Keep other comments
            func_decl = match.group(2)  # func declaration
            return f"{comments}{func_decl}"

        content = re.sub(pattern, replace_func, content)

    if removed_count > 0:
        with open(file_path, 'w') as f:
            f.write(content)

    return removed_count

def main():
    structure = load_structure()

    # File mapping
    file_mapping = {
        'goroutine': ['basic', 'advanced', 'evil'],
        'errgroup': ['basic', 'advanced', 'evil'],
        'waitgroup': ['basic', 'advanced', 'evil'],
        'goroutinecreator': ['creator'],
        'goroutinederive': ['goroutinederive'],
        'goroutinederiveand': ['basic', 'advanced', 'evil'],
        'goroutinederivemixed': ['basic', 'advanced', 'evil'],
        'gotask': ['basic', 'evil'],
        'carrier': ['carrier'],
    }

    total_removed = 0

    for target, levels in file_mapping.items():
        for level in levels:
            file_path = Path(f"../src/{target}/{level}.go")
            if not file_path.exists():
                continue

            # Get all function names for this target+level
            expected_funcs = set()
            for test_name, test in structure['tests'].items():
                if target in test['targets'] and test['levels'].get(target) == level:
                    func_name = test['functions'][target]
                    expected_funcs.add(func_name)

            count = remove_helpers_from_file(file_path, expected_funcs)
            if count > 0:
                print(f"✅ Removed {count} //vt:helper from {file_path}")
                total_removed += count

    print(f"\n✅ Total: Removed {total_removed} incorrect //vt:helper markers")

if __name__ == '__main__':
    main()
