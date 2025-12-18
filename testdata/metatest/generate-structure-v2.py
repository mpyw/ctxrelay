#!/usr/bin/env python3
"""
Generate structure.json from existing test files.
Extracts function names and comments to build test pattern metadata.

Key logic:
- goroutine/errgroup/waitgroup patterns (GO/GE/GW) are unified by NUMBER
- Other targets have independent patterns (not unified)
"""

import json
import os
import re
from pathlib import Path
from collections import defaultdict

def extract_functions_from_file(filepath):
    """Extract functions and their comments from a Go file."""
    functions = []

    with open(filepath, 'r') as f:
        lines = f.readlines()

    i = 0
    while i < len(lines):
        line = lines[i]

        # Match comment line with pattern code
        # Examples: // GO01: description, // GE15: description, // DM20: description
        comment_match = re.match(r'//\s*([A-Z]{2})(\d+[a-z]?):\s*(.+)', line)

        if comment_match:
            pattern_prefix = comment_match.group(1)
            pattern_number = comment_match.group(2)
            description = comment_match.group(3).strip()
            pattern_code = pattern_prefix + pattern_number

            # Next line should be function declaration
            i += 1
            if i < len(lines):
                func_match = re.match(r'func\s+(\w+)\s*\(', lines[i])
                if func_match:
                    func_name = func_match.group(1)
                    functions.append({
                        'name': func_name,
                        'pattern_prefix': pattern_prefix,
                        'pattern_number': pattern_number,
                        'pattern_code': pattern_code,
                        'description': description
                    })

        i += 1

    return functions

def main():
    src_dir = Path('../src')

    # All targets
    all_targets = [
        'goroutine',
        'errgroup',
        'waitgroup',
        'goroutinecreator',
        'goroutinederive',
        'goroutinederiveand',
        'goroutinederivemixed',
        'gotask',
        'carrier'
    ]

    # Goroutine group targets (unified by pattern number)
    goroutine_group = ['goroutine', 'errgroup', 'waitgroup']

    # Level files vary by target
    # goroutine/errgroup/waitgroup/goroutinederiveand/goroutinederivemixed/gotask: basic.go, advanced.go, evil.go
    # goroutinecreator: creator.go
    # goroutinederive: goroutinederive.go
    # carrier: carrier.go
    level_files = {
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

    # Prefix to target mapping
    prefix_to_target = {
        'GO': 'goroutine',
        'GE': 'errgroup',
        'GW': 'waitgroup',
        'GC': 'goroutinecreator',
        'DD': 'goroutinederive',
        'DA': 'goroutinederiveand',
        'DM': 'goroutinederivemixed',
        'GT': 'gotask',
        'EV': 'gotask',
        'CR': 'carrier',
    }

    # Extract all functions
    all_functions = {}
    for target in all_targets:
        all_functions[target] = {}
        for level in level_files.get(target, []):
            filepath = src_dir / target / f'{level}.go'
            if not filepath.exists():
                continue
            functions = extract_functions_from_file(filepath)
            if functions:
                all_functions[target][level] = functions

    # Build tests structure
    tests = {}

    # 1. Goroutine group patterns (unified by pattern NUMBER)
    goroutine_pattern_groups = defaultdict(lambda: defaultdict(dict))
    for target in goroutine_group:
        for level, functions in all_functions.get(target, {}).items():
            for func in functions:
                pattern_number = func['pattern_number']
                goroutine_pattern_groups[pattern_number][target][level] = func

    for pattern_number in sorted(goroutine_pattern_groups.keys()):
        group = goroutine_pattern_groups[pattern_number]

        # Use first description as canonical
        descriptions = [
            func['description']
            for target, levels in group.items()
            for level, func in levels.items()
        ]
        canonical_desc = descriptions[0] if descriptions else ""

        # Create test ID from pattern number
        test_id = f"pattern{pattern_number}"

        # Extract which targets and functions with their levels
        test_targets = list(group.keys())
        functions_map = {}
        levels_map = {}

        for target in test_targets:
            for level in ['basic', 'advanced', 'evil']:
                if level in group[target]:
                    func_name = group[target][level]['name']
                    functions_map[target] = func_name
                    levels_map[target] = level
                    break

        # Generate title from description
        title = canonical_desc.split(' - ')[0].split('.')[0].strip()

        tests[test_id] = {
            'title': title,
            'description': canonical_desc,
            'targets': test_targets,
            'functions': functions_map,
            'levels': levels_map
        }

    # 2. Other targets - independent patterns (NOT unified)
    other_targets = [t for t in all_targets if t not in goroutine_group]
    for target in other_targets:
        for level, functions in all_functions.get(target, {}).items():
            for func in functions:
                # Use full pattern code as test ID (e.g., "GC01", "DM20")
                test_id = func['pattern_code']

                title = func['description'].split(' - ')[0].split('.')[0].strip()

                tests[test_id] = {
                    'title': title,
                    'description': func['description'],
                    'targets': [target],
                    'functions': {target: func['name']},
                    'levels': {target: level}
                }

    # Build final structure
    structure = {
        'targets': all_targets,
        'tests': tests
    }

    # Write structure.json
    with open('structure-generated.json', 'w') as f:
        json.dump(structure, f, indent=2)

    print(f"✅ Generated structure-generated.json with {len(tests)} test patterns")
    print(f"   Goroutine group: {len([t for t in tests if t.startswith('pattern')])} unified patterns")
    print(f"   Other targets: {len([t for t in tests if not t.startswith('pattern')])} independent patterns")


if __name__ == '__main__':
    main()
