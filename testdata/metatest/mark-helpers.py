#!/usr/bin/env python3
"""
Mark helper functions with //vt:helper directive.
Helper functions are those NOT in structure.json.
"""

import json
import re
from pathlib import Path

def load_structure():
    with open('structure.json', 'r') as f:
        return json.load(f)

def get_expected_functions(structure):
    """Build map of expected functions by target and level."""
    expected = {}
    for test in structure['tests'].values():
        for target in test['targets']:
            level = test['levels'][target]
            func_name = test['functions'][target]

            if target not in expected:
                expected[target] = {}
            if level not in expected[target]:
                expected[target][level] = set()
            expected[target][level].add(func_name)
    return expected

def mark_helpers_in_file(file_path, expected_funcs):
    """Mark helper functions in a file with //vt:helper."""
    if not file_path.exists():
        return 0

    with open(file_path, 'r') as f:
        lines = f.readlines()

    marked_count = 0
    i = 0
    result = []

    while i < len(lines):
        line = lines[i]

        # Check if this is a function declaration (with or without receiver)
        # Pattern: func [receiver] funcName(
        func_match = re.match(r'^func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(', line)

        if func_match:
            func_name = func_match.group(1)

            # Check if this is a helper (not in expected functions)
            is_helper = func_name not in expected_funcs

            # Check if already marked
            already_marked = False
            if i > 0:
                prev_line = lines[i-1]
                if '//vt:helper' in prev_line:
                    already_marked = True

            # Mark if it's a helper and not already marked
            if is_helper and not already_marked:
                # Add //vt:helper comment before function
                result.append('//vt:helper\n')
                marked_count += 1

        result.append(line)
        i += 1

    # Write back
    with open(file_path, 'w') as f:
        f.writelines(result)

    return marked_count

def main():
    structure = load_structure()
    expected_functions = get_expected_functions(structure)

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

    total_marked = 0

    for target, levels in file_mapping.items():
        for level in levels:
            file_path = Path(f"../src/{target}/{level}.go")
            if not file_path.exists():
                continue

            expected_funcs = expected_functions.get(target, {}).get(level, set())
            marked = mark_helpers_in_file(file_path, expected_funcs)

            if marked > 0:
                print(f"✅ Marked {marked} helpers in {file_path}")
                total_marked += marked

    print(f"\n✅ Total: Marked {total_marked} helper functions")

if __name__ == '__main__':
    main()
