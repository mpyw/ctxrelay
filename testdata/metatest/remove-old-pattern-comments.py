#!/usr/bin/env python3
"""
Remove old pattern code comments like "// GO01:", "// GE02c:", etc.
These are the comments from before update-all-comments.py was run.
"""

import re
from pathlib import Path

def remove_pattern_comments(file_path):
    """Remove old pattern code comments."""
    if not file_path.exists():
        return 0

    with open(file_path, 'r') as f:
        content = f.read()

    removed_count = 0

    # Pattern: match lines like "// GO01:", "// GE02c:", "// pattern10:", etc.
    # These are on their own line before function declarations
    pattern = r'^//\s*(?:GO|GE|GW|GC|DD|DA|DM|GT|EV|CR|pattern)\d+[a-z]?:\s*.*\n'

    matches = re.findall(pattern, content, re.MULTILINE)
    removed_count = len(matches)

    content = re.sub(pattern, '', content, flags=re.MULTILINE)

    if removed_count > 0:
        with open(file_path, 'w') as f:
            f.write(content)

    return removed_count

def main():
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

            count = remove_pattern_comments(file_path)
            if count > 0:
                print(f"✅ Removed {count} old pattern comments from {file_path}")
                total_removed += count

    print(f"\n✅ Total: Removed {total_removed} old pattern comments")

if __name__ == '__main__':
    main()
