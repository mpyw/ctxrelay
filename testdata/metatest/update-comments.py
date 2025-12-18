#!/usr/bin/env python3
"""
Update test function comments to new format.

Old format: // **Title**
New format: // [GOOD]: Title  or  // [BAD]: Title
"""

import json
import re
import os
from pathlib import Path


def load_structure():
    with open('structure.json', 'r') as f:
        return json.load(f)


def get_func_to_variant_map(structure):
    """Build mapping from (target, funcName) -> (title, variantType)."""
    func_map = {}
    for test_name, test in structure['tests'].items():
        title = test['title']
        for variant_type, variant in test['variants'].items():
            if variant is None:
                continue
            for target in test['targets']:
                func_name = variant['functions'].get(target)
                if func_name:
                    func_map[(target, func_name)] = (title, variant_type)
    return func_map


def update_file_comments(file_path, target, func_map):
    """Update comments in a single Go file."""
    with open(file_path, 'r') as f:
        content = f.read()

    # Pattern to match function definitions with their doc comments
    # This regex captures:
    # 1. The comment block (optional)
    # 2. The function name
    pattern = r'((?://[^\n]*\n)+)?(func\s+(?:\([^)]*\)\s+)?(\w+)\s*\([^)]*\)\s*(?:\([^)]*\)|[^{])*\{)'

    def replace_comment(match):
        comments = match.group(1) or ''
        func_def = match.group(2)
        func_name = match.group(3)

        # Check if this function is in our map
        key = (target, func_name)
        if key not in func_map:
            return comments + func_def

        title, variant_type = func_map[key]
        variant_label = variant_type.upper()

        # Check if first line is already in new format
        if comments and re.match(r'^//\s*\[' + variant_label + r'\]:', comments.strip()):
            return comments + func_def

        # Check if first line is old format **Title**
        old_title_match = re.match(r'^//\s*\*\*([^*]+)\*\*', comments.strip())
        if old_title_match:
            # Replace old format with new format
            new_first_line = f'// [{variant_label}]: {title}\n'
            # Remove the old first line
            comment_lines = comments.strip().split('\n')
            rest_of_comments = '\n'.join(comment_lines[1:])
            if rest_of_comments.strip():
                rest_of_comments = rest_of_comments + '\n'
            return new_first_line + rest_of_comments + func_def
        else:
            # No proper comment, add new format
            new_comment = f'// [{variant_label}]: {title}\n'
            return new_comment + comments + func_def

    updated_content = re.sub(pattern, replace_comment, content)

    if updated_content != content:
        with open(file_path, 'w') as f:
            f.write(updated_content)
        return True
    return False


def get_target_files(base_dir, target):
    """Get all Go files for a target."""
    target_dir = base_dir / 'src' / target
    if not target_dir.exists():
        return []
    return list(target_dir.glob('*.go'))


def main():
    structure = load_structure()
    func_map = get_func_to_variant_map(structure)

    print(f"Loaded {len(func_map)} function mappings")

    base_dir = Path('..').resolve()
    updated_count = 0

    for target in structure['targets']:
        files = get_target_files(base_dir, target)
        for file_path in files:
            if update_file_comments(str(file_path), target, func_map):
                updated_count += 1
                print(f"Updated: {file_path}")

    print(f"\nUpdated {updated_count} files")


if __name__ == '__main__':
    main()
