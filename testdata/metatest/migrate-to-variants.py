#!/usr/bin/env python3
"""
Migrate structure.json to new schema with good/bad variants.

New schema:
{
  "testKey": {
    "title": "Test title",
    "targets": ["goroutine", "errgroup"],
    "variants": {
      "good": {
        "description": "...",
        "functions": {"goroutine": "goodFuncName", ...},
        "levels": {"goroutine": "basic", ...}
      },
      "bad": {
        "description": "...",
        "functions": {"goroutine": "badFuncName", ...},
        "levels": {"goroutine": "basic", ...}
      }
    }
  }
}
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


def fix_key(key):
    """Convert key to valid function name (remove hyphens, fix casing)."""
    # Remove hyphens and convert to camelCase
    parts = key.split('-')
    if len(parts) == 1:
        return key
    # First part stays as-is, rest get capitalized
    result = parts[0]
    for part in parts[1:]:
        if part:
            result += part[0].upper() + part[1:] if part else ''
    return result


def classify_test(test_id, test):
    """Classify a test as good, bad, limitation, or notChecked."""
    # Look at function names to determine classification
    for func_name in test['functions'].values():
        if func_name.startswith('good') or func_name.endswith('Good'):
            return 'good'
        if func_name.startswith('bad') or func_name.endswith('Bad'):
            return 'bad'
        if func_name.startswith('limitation'):
            return 'limitation'
        if func_name.startswith('notChecked'):
            return 'notChecked'
        if func_name.startswith('evil') and func_name.endswith('Good'):
            return 'good'
        if func_name.startswith('evil') and not func_name.endswith('Good'):
            return 'bad'
        # go80, go81, etc patterns
        if func_name.startswith('go8') and func_name.endswith('Good'):
            return 'good'
        if func_name.startswith('go8') and func_name.endswith('Bad'):
            return 'bad'

    # Check title/description for hints
    title_lower = test['title'].lower()
    desc_lower = test['description'].lower()

    if 'good' in title_lower or 'with ctx' in title_lower or 'uses ctx' in title_lower:
        return 'good'
    if 'bad' in title_lower or 'without ctx' in title_lower or 'no ctx' in title_lower:
        return 'bad'
    if 'limitation' in title_lower:
        return 'limitation'
    if 'not checked' in title_lower:
        return 'notChecked'

    # For derive checkers, analyze the function prefix patterns
    for func_name in test['functions'].values():
        # AND deriver patterns (a01-a99)
        if func_name.startswith('a') and len(func_name) > 2 and func_name[1:3].isdigit():
            # CallsBoth, CallsBothReversed, etc. -> good
            if 'CallsBoth' in func_name or 'BothDerivers' in func_name or 'BothBranches' in func_name or 'BothHaveBothDerivers' in func_name:
                return 'good'
            # CallsOnlyFirst, CallsOnlySecond, CallsNeither -> bad
            if 'CallsOnly' in func_name or 'CallsNeither' in func_name or 'Incomplete' in func_name or 'Partial' in func_name or 'MissingOne' in func_name:
                return 'bad'
            # OwnContextParam -> not checked
            if 'OwnContextParam' in func_name:
                return 'notChecked'
            # Patterns with "OneDeriver" are bad (incomplete derivation)
            if 'OneDeriver' in func_name:
                return 'bad'
            # DifferentOrder patterns are good (just different order, both still called)
            if 'DifferentOrder' in func_name:
                return 'good'
            # Nested patterns with "BothDerivers" are good
            if 'BothDeriver' in func_name:
                return 'good'
            # SplitDerivers across levels (this tests analyzer capability, but both are called)
            if 'SplitDerivers' in func_name:
                return 'good'

        # Mixed deriver patterns (m01-m99)
        if func_name.startswith('m') and len(func_name) > 2 and func_name[1:3].isdigit():
            if 'Satisfies' in func_name or 'DifferentApproaches' in func_name or 'ReversedApproaches' in func_name:
                return 'good'
            if 'Only' in func_name or 'Nothing' in func_name or 'Fails' in func_name or 'Neither' in func_name or 'Incomplete' in func_name:
                return 'bad'
            if 'OwnContextParam' in func_name:
                return 'notChecked'
            # Partial deriver is bad
            if 'Partial' in func_name:
                return 'bad'

        # Single deriver patterns (d01-d99)
        if func_name.startswith('d') and len(func_name) > 2 and func_name[1:3].isdigit():
            if 'CallsDeriver' in func_name or 'BothCallDeriver' in func_name:
                return 'good'
            if 'NoDeriverCall' in func_name or 'Missing' in func_name or 'UsesDifferent' in func_name:
                return 'bad'
            if 'OwnContextParam' in func_name or 'NamedFuncCall' in func_name:
                return 'notChecked'

    # go80-series patterns without clear suffix
    for func_name in test['functions'].values():
        if func_name.startswith('go8') and len(func_name) > 3:
            # go80, go81, etc without Good/Bad suffix - look at test title/description
            if 'shadowing' in func_name.lower():
                # Check if the title indicates good or bad
                if 'ignores' in desc_lower or 'bad' in desc_lower:
                    return 'bad'
                return 'good'
            if 'interleaved' in func_name.lower() or 'middlelayer' in func_name.lower():
                if 'ignores' in desc_lower or 'bad' in desc_lower:
                    return 'bad'
                return 'good'
            # go82TwoLevelArguments - this is a good case (uses ctx)
            if 'twolevel' in func_name.lower():
                return 'good'

    # Mixed deriver patterns that haven't been caught yet
    for func_name in test['functions'].values():
        if func_name.startswith('m') and len(func_name) > 2 and func_name[1:3].isdigit():
            # SplitDerivers, OrAlternative in nested are good (they satisfy the requirement)
            if 'SplitDerivers' in func_name or 'OrAlternative' in func_name:
                return 'good'
            # Reassigned func patterns - check if "should pass" in description
            if 'Reassigned' in func_name:
                if 'should pass' in desc_lower:
                    return 'good'
                return 'bad'

    return 'unknown'


def extract_base_concept(test_id, test):
    """Extract the base concept name without good/bad qualifier."""
    title = test['title']

    # Remove common prefixes/suffixes that indicate good/bad
    patterns_to_remove = [
        r'^(good|bad|limitation)\s*[-:]\s*',
        r'\s*[-:]\s*(good|bad|with ctx|without ctx|uses ctx|no ctx)$',
        r'\s*\(good\)$',
        r'\s*\(bad\)$',
    ]

    for pattern in patterns_to_remove:
        title = re.sub(pattern, '', title, flags=re.IGNORECASE)

    return title.strip()


def find_groupable_tests(tests):
    """Find tests that can be grouped as good/bad variants."""
    # Group by base concept
    concept_groups = defaultdict(list)

    for test_id, test in tests.items():
        classification = classify_test(test_id, test)
        base_concept = extract_base_concept(test_id, test)
        concept_groups[base_concept].append((test_id, test, classification))

    return concept_groups


def normalize_title_for_grouping(title):
    """Normalize title for grouping good/bad pairs."""
    # Remove good/bad indicators
    normalized = title.lower()

    # Remove patterns like "- basic good case", "- basic bad case"
    patterns_to_remove = [
        r'\s*-\s*basic\s*(good|bad)\s*case',
        r'\s*-\s*(good|bad)\s*case',
        r'\s*-\s*(good|bad)$',
        r'\s*\((good|bad)\)$',
        r'\s*-\s*with\s+ctx$',
        r'\s*-\s*without\s+ctx$',
        r'\s*-\s*uses\s+ctx$',
        r'\s*-\s*no\s+ctx$',
        r'\s*-\s*with\s+context\s+usage$',
        r'\s*-\s*without\s+context\s+usage$',
        r'\s*-\s*uses\s+it$',
        r'\s*-\s*does\s+not\s+use\s+it$',
        r'\s*-\s*uses\s+(neither|carrier|context)$',
        r'\s*with\s+deriver$',
        r'\s*without\s+deriver$',
    ]

    for pattern in patterns_to_remove:
        normalized = re.sub(pattern, '', normalized, flags=re.IGNORECASE)

    return normalized.strip()


def find_good_bad_pairs(tests):
    """Find good/bad pairs that should be grouped together."""
    # Group tests by normalized title
    title_groups = defaultdict(list)

    for test_id, test in tests.items():
        classification = classify_test(test_id, test)
        normalized = normalize_title_for_grouping(test['title'])

        # Also consider target set for grouping
        targets_key = tuple(sorted(test['targets']))
        group_key = (normalized, targets_key)

        title_groups[group_key].append({
            'test_id': test_id,
            'test': test,
            'classification': classification
        })

    return title_groups


def migrate_structure(old_structure):
    """Migrate old structure to new variants-based schema."""
    new_tests = {}

    # First, find potential good/bad pairs
    title_groups = find_good_bad_pairs(old_structure['tests'])

    # Track which tests have been processed
    processed = set()

    # Process pairs first
    for group_key, items in title_groups.items():
        if len(items) >= 2:
            # Check if we have both good and bad
            good_items = [i for i in items if i['classification'] == 'good']
            bad_items = [i for i in items if i['classification'] == 'bad']

            if good_items and bad_items:
                # We have a pair! Create a merged entry
                good_item = good_items[0]
                bad_item = bad_items[0]

                # Use the good item's title as base (often more descriptive)
                base_title = good_item['test']['title']

                # Remove good/bad suffix from title
                clean_title = base_title
                for pattern in [r'\s*-\s*(good|bad).*$', r'\s*\((good|bad)\)$']:
                    clean_title = re.sub(pattern, '', clean_title, flags=re.IGNORECASE)

                # Generate key from good item (remove good/bad prefix from key)
                base_key = fix_key(good_item['test_id'])
                # Remove 'good', 'bad' prefix if present
                if base_key.startswith('good'):
                    base_key = base_key[4].lower() + base_key[5:]
                elif base_key.startswith('bad'):
                    base_key = base_key[3].lower() + base_key[4:]

                new_tests[base_key] = {
                    'title': clean_title.strip(),
                    'targets': good_item['test']['targets'],
                    'variants': {
                        'good': {
                            'description': good_item['test']['description'],
                            'functions': good_item['test']['functions'],
                            'levels': good_item['test']['levels']
                        },
                        'bad': {
                            'description': bad_item['test']['description'],
                            'functions': bad_item['test']['functions'],
                            'levels': bad_item['test']['levels']
                        }
                    }
                }

                processed.add(good_item['test_id'])
                processed.add(bad_item['test_id'])

                # If there are more items, process them separately
                for item in good_items[1:] + bad_items[1:]:
                    if item['test_id'] not in processed:
                        pass  # Will be processed in the single-item loop

    # Process remaining tests
    for old_key, test in old_structure['tests'].items():
        if old_key in processed:
            continue

        new_key = fix_key(old_key)
        classification = classify_test(old_key, test)

        # Determine variant type based on classification
        if classification in ('good', 'bad'):
            variant_type = classification
            other_variant = 'bad' if classification == 'good' else 'good'
        elif classification == 'limitation':
            variant_type = 'limitation'
            other_variant = None
        elif classification == 'notChecked':
            variant_type = 'notChecked'
            other_variant = None
        else:
            variant_type = 'unknown'
            other_variant = None

        variant_data = {
            'description': test['description'],
            'functions': test['functions'],
            'levels': test['levels']
        }

        variants = {variant_type: variant_data}
        if other_variant:
            variants[other_variant] = None

        new_tests[new_key] = {
            'title': test['title'],
            'targets': test['targets'],
            'variants': variants
        }

    return {
        'targets': old_structure['targets'],
        'tests': new_tests
    }


def analyze_and_report(old_structure):
    """Analyze the structure and report findings."""
    tests = old_structure['tests']

    # Classify all tests
    classifications = defaultdict(list)
    for test_id, test in tests.items():
        cls = classify_test(test_id, test)
        classifications[cls].append((test_id, test))

    print("=" * 80)
    print("CLASSIFICATION REPORT")
    print("=" * 80)

    for cls, items in sorted(classifications.items()):
        print(f"\n{cls.upper()}: {len(items)} tests")
        if cls == 'unknown':
            for test_id, test in items[:10]:
                funcs = list(test['functions'].values())
                print(f"  - {test_id}: {funcs[0] if funcs else 'N/A'}")
            if len(items) > 10:
                print(f"  ... and {len(items) - 10} more")

    # Check for invalid keys (with hyphens)
    invalid_keys = [k for k in tests.keys() if '-' in k]
    print(f"\n{'=' * 80}")
    print(f"INVALID KEYS (with hyphens): {len(invalid_keys)}")
    print("=" * 80)
    for key in invalid_keys[:20]:
        fixed = fix_key(key)
        print(f"  {key} -> {fixed}")
    if len(invalid_keys) > 20:
        print(f"  ... and {len(invalid_keys) - 20} more")

    # Find good/bad pairs
    title_groups = find_good_bad_pairs(tests)
    pairs_found = []
    for group_key, items in title_groups.items():
        if len(items) >= 2:
            good_items = [i for i in items if i['classification'] == 'good']
            bad_items = [i for i in items if i['classification'] == 'bad']
            if good_items and bad_items:
                pairs_found.append((group_key[0], good_items[0]['test_id'], bad_items[0]['test_id']))

    print(f"\n{'=' * 80}")
    print(f"GOOD/BAD PAIRS FOUND: {len(pairs_found)}")
    print("=" * 80)
    for normalized, good_id, bad_id in sorted(pairs_found)[:30]:
        print(f"  {normalized[:50]}")
        print(f"    good: {good_id}")
        print(f"    bad:  {bad_id}")
    if len(pairs_found) > 30:
        print(f"  ... and {len(pairs_found) - 30} more pairs")

    return classifications


def main():
    import sys

    old_structure = load_structure()

    if '--analyze' in sys.argv:
        analyze_and_report(old_structure)
        return

    if '--dry-run' in sys.argv:
        analyze_and_report(old_structure)
        new_structure = migrate_structure(old_structure)
        print(f"\n{'=' * 80}")
        print("DRY RUN: Would create {len(new_structure['tests'])} test entries")
        print("=" * 80)
        # Show a few examples
        count = 0
        for key, test in new_structure['tests'].items():
            if count >= 5:
                break
            print(f"\n{key}:")
            print(json.dumps(test, indent=2)[:500])
            count += 1
        return

    # Do the migration
    print("Migrating structure.json to new variants schema...")

    new_structure = migrate_structure(old_structure)
    save_structure(new_structure)

    print(f"\nMigrated {len(new_structure['tests'])} tests")

    # Report classification breakdown
    variants_count = defaultdict(int)
    for test in new_structure['tests'].values():
        for variant_type in test['variants'].keys():
            variants_count[variant_type] += 1

    print("\nVariant breakdown:")
    for vtype, count in sorted(variants_count.items()):
        print(f"  {vtype}: {count}")


if __name__ == '__main__':
    main()
