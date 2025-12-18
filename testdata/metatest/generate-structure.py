#!/usr/bin/env python3
"""
Generate structure.json from existing test files.
Extracts function names and comments to build test pattern metadata.
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
        # Examples: // GO01: description, // GE15: description
        comment_match = re.match(r'//\s*([A-Z]{2})(\d+[a-z]?):\s*(.+)', line)

        if comment_match:
            pattern_code = comment_match.group(1) + comment_match.group(2)
            description = comment_match.group(3).strip()

            # Next line should be function declaration
            i += 1
            if i < len(lines):
                func_match = re.match(r'func\s+(\w+)\s*\(', lines[i])
                if func_match:
                    func_name = func_match.group(1)
                    functions.append({
                        'name': func_name,
                        'pattern_code': pattern_code,
                        'description': description
                    })

        i += 1

    return functions

def main():
    src_dir = Path('../src')
    targets = [
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
    levels = ['basic', 'advanced', 'evil']

    # Prefix mapping
    prefix_to_target = {
        'GO': 'goroutine',
        'GE': 'errgroup',
        'GW': 'waitgroup',
        'GC': 'goroutinecreator',
        'DD': 'goroutinederive',
        'DA': 'goroutinederiveand',
        'DM': 'goroutinederivemixed',
        'GT': 'gotask',
        'EV': 'gotask',  # evil patterns for gotask
        'CR': 'carrier',
    }

    # Extract all functions grouped by pattern NUMBER (ignoring prefix)
    pattern_groups = defaultdict(lambda: defaultdict(dict))

    for target in targets:
        for level in levels:
            filepath = src_dir / target / f'{level}.go'
            if not filepath.exists():
                continue

            functions = extract_functions_from_file(filepath)
            for func in functions:
                # Extract pattern number (e.g., "01" from "GO01", "GE01", etc.)
                pattern_code = func['pattern_code']
                prefix = pattern_code[:2]  # GO, GE, GW
                number = pattern_code[2:]  # 01, 01b, etc.

                # Group by number only
                pattern_groups[number][target][level] = {
                    'name': func['name'],
                    'description': func['description']
                }

    # Build structure.json
    structure = {
        'targets': targets,
        'tests': {}
    }

    for pattern_number in sorted(pattern_groups.keys()):
        group = pattern_groups[pattern_number]

        # Use first description as canonical
        descriptions = [
            info['description']
            for target, levels in group.items()
            for level, info in levels.items()
        ]
        canonical_desc = descriptions[0] if descriptions else ""

        # Create test ID from pattern number (e.g., "pattern01", "pattern01b")
        test_id = f"pattern{pattern_number}"

        # Extract which targets and functions
        test_targets = list(group.keys())
        functions = {}

        for target in test_targets:
            # Pick first available level for this target
            for level in levels:
                if level in group[target]:
                    func_name = group[target][level]['name']
                    functions[target] = f"{level}:{func_name}"
                    break

        # Generate title from description (first part before dash or period)
        title = canonical_desc.split(' - ')[0].split('.')[0].strip()

        structure['tests'][test_id] = {
            'title': title,
            'description': canonical_desc,
            'targets': test_targets,
            'functions': functions
        }

    # Write structure.json
    with open('structure-generated.json', 'w') as f:
        json.dump(structure, f, indent=2)

    print(f"✅ Generated structure-generated.json with {len(structure['tests'])} test patterns")

if __name__ == '__main__':
    main()
