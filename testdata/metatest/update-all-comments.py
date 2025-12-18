#!/usr/bin/env python3
"""
Update all test function comments based on structure.json.
Removes existing doc comments (except //vt:helper) and adds new format:
// **Title**
//
// Description
//
// See also:
//   target: funcName
"""

import json
import re
from pathlib import Path

def load_structure():
    with open('structure.json', 'r') as f:
        return json.load(f)

def generate_comment(test, func_name, current_target, all_targets):
    """Generate comment block for a test function."""
    lines = []

    # Title with bold formatting
    lines.append(f"// **{test['title']}**")
    lines.append(f"//")

    # Description
    lines.append(f"// {test['description']}")

    # See also section
    other_targets = [t for t in all_targets if t in test['targets'] and t != current_target]
    if other_targets:
        lines.append(f"//")
        lines.append(f"// See also:")
        for target in other_targets:
            target_func = test['functions'][target]
            lines.append(f"//   {target}: {target_func}")

    return '\n'.join(lines)

def update_file(file_path, structure, target, level):
    """Update all function comments in a file."""
    if not file_path.exists():
        print(f"  ⚠️  File not found: {file_path}")
        return 0

    with open(file_path, 'r') as f:
        content = f.read()

    # Build func_name -> test mapping for this target
    func_to_test = {}
    for test_name, test in structure['tests'].items():
        if target in test['targets'] and test['levels'].get(target) == level:
            func_name = test['functions'][target]
            func_to_test[func_name] = test

    updated_count = 0

    # Process each function
    for func_name, test in func_to_test.items():
        # Generate new comment
        new_comment = generate_comment(test, func_name, target, structure['targets'])

        # Find function declaration and replace doc comments
        # Pattern: optional comments (but NOT //vt:helper), then "func funcName("
        # Keep //vt:helper if present
        pattern = r'((?://vt:helper\n)?)((?://(?!vt:helper).*\n)*)(func ' + re.escape(func_name) + r'\()'

        def replace_comment(match):
            nonlocal updated_count
            updated_count += 1
            helper_line = match.group(1)  # Keep //vt:helper if present
            func_decl = match.group(3)
            return f"{helper_line}{new_comment}\n{func_decl}"

        content = re.sub(pattern, replace_comment, content)

    # Write back
    with open(file_path, 'w') as f:
        f.write(content)

    return updated_count

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

    total_updated = 0

    for target, levels in file_mapping.items():
        for level in levels:
            file_path = Path(f"../src/{target}/{level}.go")
            if not file_path.exists():
                continue

            count = update_file(file_path, structure, target, level)
            if count > 0:
                print(f"✅ Updated {file_path}: {count} functions")
                total_updated += count

    print(f"\n✅ Total: Updated {total_updated} function comments")

if __name__ == '__main__':
    main()
