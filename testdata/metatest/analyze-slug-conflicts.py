#!/usr/bin/env python3
"""
Analyze slug conflicts when converting titles to camelCase.
Detect if conflicts should be grouped or if titles need more specificity.
"""

import json
import re
from collections import defaultdict

def load_structure():
    with open('structure.json', 'r') as f:
        return json.load(f)

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

    # Map: slug -> list of (test_id, test_data)
    slug_to_tests = defaultdict(list)

    for test_id, test in structure['tests'].items():
        slug = title_to_camel_case(test['title'])
        slug_to_tests[slug].append((test_id, test))

    # Analyze conflicts
    conflicts = {slug: tests for slug, tests in slug_to_tests.items() if len(tests) > 1}

    print(f"Total unique slugs: {len(slug_to_tests)}")
    print(f"Conflicts found: {len(conflicts)}\n")

    if not conflicts:
        print("✅ No conflicts! All titles convert to unique slugs.")
        return

    print("=" * 80)
    print("CONFLICT ANALYSIS")
    print("=" * 80)

    for slug, tests in sorted(conflicts.items()):
        print(f"\n### Slug: {slug} ({len(tests)} conflicts)")
        print("-" * 80)

        # Check if all tests have same targets (grouping candidate)
        all_targets = set()
        for test_id, test in tests:
            all_targets.update(test['targets'])

        targets_per_test = [set(test['targets']) for _, test in tests]
        can_group = len(all_targets) == sum(len(t) for t in targets_per_test)

        # Check if descriptions are similar
        descriptions = [test['description'] for _, test in tests]
        unique_descriptions = set(descriptions)

        # Analysis
        if can_group and len(unique_descriptions) == 1:
            print("💡 SHOULD GROUP: Same description, complementary targets")
            print(f"   Combined targets: {sorted(all_targets)}")
        elif len(unique_descriptions) == len(tests):
            print("⚠️  NEEDS BETTER TITLES: Different descriptions, same slug")
        else:
            print("🤔 MIXED: Some similar, some different")

        print(f"\n   Tests:")
        for test_id, test in tests:
            targets_str = ', '.join(test['targets'])
            desc = test['description'][:60] + '...' if len(test['description']) > 60 else test['description']
            print(f"     - {test_id}: [{targets_str}]")
            print(f"       Title: {test['title']}")
            print(f"       Desc:  {desc}")

    print("\n" + "=" * 80)
    print("SUMMARY")
    print("=" * 80)

    group_candidates = 0
    title_improvement_needed = 0
    mixed = 0

    for slug, tests in conflicts.items():
        all_targets = set()
        for _, test in tests:
            all_targets.update(test['targets'])
        targets_per_test = [set(test['targets']) for _, test in tests]
        can_group = len(all_targets) == sum(len(t) for t in targets_per_test)

        descriptions = [test['description'] for _, test in tests]
        unique_descriptions = set(descriptions)

        if can_group and len(unique_descriptions) == 1:
            group_candidates += 1
        elif len(unique_descriptions) == len(tests):
            title_improvement_needed += 1
        else:
            mixed += 1

    print(f"\n💡 Should group: {group_candidates} conflicts")
    print(f"⚠️  Need better titles: {title_improvement_needed} conflicts")
    print(f"🤔 Mixed cases: {mixed} conflicts")

if __name__ == '__main__':
    main()
