#!/usr/bin/env python3
"""
Assign camelCase slugs as test IDs based on titles.
Handle conflicts by adding numeric suffixes.
"""

import json
import re
from collections import defaultdict

def load_structure():
    with open('structure.json', 'r') as f:
        return json.load(f)

def save_structure(structure):
    with open('structure.json', 'w') as f:
        json.dump(structure, f, indent=2, ensure_ascii=False)
        f.write('\n')

def title_to_camel_case(title):
    """Convert title to camelCase slug."""
    # Remove special characters and extra info in parentheses
    title = re.sub(r'\s*\([^)]*\)', '', title)
    title = re.sub(r'[^\w\s-]', '', title)

    # Split into words
    words = title.split()
    if not words:
        return 'unknown'

    # First word lowercase, rest titlecase
    camel = words[0].lower() + ''.join(w.capitalize() for w in words[1:])
    return camel

def main():
    structure = load_structure()

    # Map: slug -> list of (old_test_id, test_data)
    slug_to_tests = defaultdict(list)

    for test_id, test in structure['tests'].items():
        slug = title_to_camel_case(test['title'])
        slug_to_tests[slug].append((test_id, test))

    # Build new structure with camelCase IDs
    new_tests = {}
    slug_counts = defaultdict(int)

    for slug, tests in sorted(slug_to_tests.items()):
        if len(tests) == 1:
            # No conflict
            old_id, test = tests[0]
            new_tests[slug] = test
            print(f"✅ {old_id} → {slug}")
        else:
            # Conflict: add numeric suffix
            for i, (old_id, test) in enumerate(tests):
                if i == 0:
                    new_slug = slug
                else:
                    new_slug = f"{slug}{i+1}"
                new_tests[new_slug] = test
                print(f"⚠️  {old_id} → {new_slug} (conflict)")

    structure['tests'] = new_tests
    save_structure(structure)

    print(f"\n✅ Assigned {len(new_tests)} camelCase slugs")
    conflicts = sum(1 for tests in slug_to_tests.values() if len(tests) > 1)
    print(f"⚠️  {conflicts} conflicts resolved with numeric suffixes")

if __name__ == '__main__':
    main()
