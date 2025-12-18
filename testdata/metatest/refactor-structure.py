#!/usr/bin/env python3
"""
Refactor structure.json:
1. Hoist levels to group level if identical across variants
2. Identify functions with legacy naming patterns (a01, m43, d05, etc.)
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


def hoist_levels(structure):
    """Move levels to group level if identical across all variants."""
    hoisted_count = 0

    for test_name, test in structure['tests'].items():
        variants = test.get('variants', {})
        non_null_variants = {k: v for k, v in variants.items() if v is not None}

        if not non_null_variants:
            continue

        # Get levels from each variant
        all_levels = [v.get('levels', {}) for v in non_null_variants.values()]

        # Check if all levels are identical
        if len(all_levels) > 0 and all(levels == all_levels[0] for levels in all_levels):
            # Hoist levels to group level
            test['levels'] = all_levels[0]
            # Remove levels from variants
            for variant in non_null_variants.values():
                if 'levels' in variant:
                    del variant['levels']
            hoisted_count += 1

    return hoisted_count


def find_legacy_names(structure):
    """Find functions with legacy naming patterns like a01, m43, d05, go80, etc."""
    legacy_patterns = [
        (r'^[amd]\d{2}[A-Z]', 'derive'),  # a01And..., m43Mixed..., d05No...
        (r'^go\d{2}[A-Z]', 'goroutine'),  # go80Goroutine..., go81Argument...
    ]

    legacy_funcs = []

    for test_name, test in structure['tests'].items():
        for variant_type, variant in test.get('variants', {}).items():
            if variant is None:
                continue
            for target, func_name in variant.get('functions', {}).items():
                for pattern, category in legacy_patterns:
                    if re.match(pattern, func_name):
                        legacy_funcs.append({
                            'test_name': test_name,
                            'variant_type': variant_type,
                            'target': target,
                            'func_name': func_name,
                            'category': category,
                            'title': test['title']
                        })
                        break

    return legacy_funcs


def suggest_new_name(func_info):
    """Suggest a new name for a legacy-named function."""
    func_name = func_info['func_name']
    title = func_info['title']
    variant_type = func_info['variant_type']
    target = func_info['target']

    # Remove the legacy prefix (a01, m43, d05, go80, etc.)
    new_name = re.sub(r'^[amd]\d{2}', '', func_name)
    new_name = re.sub(r'^go\d{2}', '', new_name)

    # Remove redundant "Bad" or "Good" suffix if it will be in prefix
    if new_name.endswith('Bad'):
        new_name = new_name[:-3]
    if new_name.endswith('Good'):
        new_name = new_name[:-4]

    # Add good/bad/notChecked prefix based on variant type
    if variant_type == 'good':
        if not new_name.lower().startswith('good'):
            new_name = 'good' + new_name
    elif variant_type == 'bad':
        if not new_name.lower().startswith('bad'):
            new_name = 'bad' + new_name
    elif variant_type == 'notChecked':
        if not new_name.lower().startswith('notchecked'):
            new_name = 'notChecked' + new_name

    # Ensure first char after prefix is uppercase
    if new_name.startswith('good') and len(new_name) > 4:
        new_name = 'good' + new_name[4].upper() + new_name[5:]
    elif new_name.startswith('bad') and len(new_name) > 3:
        new_name = 'bad' + new_name[3].upper() + new_name[4:]
    elif new_name.startswith('notChecked') and len(new_name) > 10:
        new_name = 'notChecked' + new_name[10].upper() + new_name[11:]

    return new_name


def rename_function_in_file(file_path, old_name, new_name):
    """Rename a function in a Go file."""
    with open(file_path, 'r') as f:
        content = f.read()

    # Replace function definition
    pattern = r'\bfunc\s+' + re.escape(old_name) + r'\s*\('
    replacement = f'func {new_name}('
    new_content = re.sub(pattern, replacement, content)

    if new_content != content:
        with open(file_path, 'w') as f:
            f.write(new_content)
        return True
    return False


def get_file_for_target_level(target, level):
    """Get the file path for a target and level."""
    from pathlib import Path
    base_dir = Path('..').resolve() / 'src' / target

    if target == 'goroutinecreator':
        return base_dir / 'creator.go'
    elif target == 'goroutinederive':
        return base_dir / 'goroutinederive.go'
    elif target == 'carrier':
        return base_dir / 'carrier.go'
    else:
        return base_dir / f'{level}.go'


def apply_renames(structure, legacy_funcs):
    """Apply function renames to structure.json and Go files."""
    rename_map = {}  # (target, old_name) -> new_name

    for func_info in legacy_funcs:
        old_name = func_info['func_name']
        new_name = suggest_new_name(func_info)

        # Skip if no change needed
        if old_name == new_name:
            continue

        target = func_info['target']
        rename_map[(target, old_name)] = new_name

    # Update structure.json
    for test_name, test in structure['tests'].items():
        for variant_type, variant in test.get('variants', {}).items():
            if variant is None:
                continue
            for target in list(variant.get('functions', {}).keys()):
                old_name = variant['functions'][target]
                key = (target, old_name)
                if key in rename_map:
                    variant['functions'][target] = rename_map[key]

    # Update Go files
    from pathlib import Path
    files_updated = set()

    for (target, old_name), new_name in rename_map.items():
        # Find the level for this function
        for test_name, test in structure['tests'].items():
            for variant_type, variant in test.get('variants', {}).items():
                if variant is None:
                    continue
                if variant['functions'].get(target) == new_name:
                    # Get level - could be at test level or variant level
                    if 'levels' in test:
                        level = test['levels'].get(target)
                    elif 'levels' in variant:
                        level = variant['levels'].get(target)
                    else:
                        continue

                    if level:
                        file_path = get_file_for_target_level(target, level)
                        if file_path.exists():
                            if rename_function_in_file(str(file_path), old_name, new_name):
                                files_updated.add(str(file_path))
                                print(f"  Renamed {old_name} -> {new_name} in {file_path.name}")
                    break

    return len(rename_map), files_updated


def main():
    import sys

    structure = load_structure()

    # Step 1: Hoist levels
    hoisted_count = hoist_levels(structure)
    print(f"Hoisted levels in {hoisted_count} tests")

    # Step 2: Find legacy names
    legacy_funcs = find_legacy_names(structure)

    print(f"\n{'=' * 80}")
    print(f"LEGACY FUNCTION NAMES FOUND: {len(legacy_funcs)}")
    print("=" * 80)

    # Group by target
    by_target = defaultdict(list)
    for func in legacy_funcs:
        by_target[func['target']].append(func)

    for target, funcs in sorted(by_target.items()):
        print(f"\n{target} ({len(funcs)} functions):")
        for func in funcs:
            suggested = suggest_new_name(func)
            print(f"  {func['func_name']} → {suggested}")

    if '--apply' in sys.argv:
        print(f"\n{'=' * 80}")
        print("APPLYING CHANGES")
        print("=" * 80)

        # Apply renames
        rename_count, files_updated = apply_renames(structure, legacy_funcs)
        print(f"\nRenamed {rename_count} functions across {len(files_updated)} files")

        # Save structure.json
        save_structure(structure)
        print(f"Saved structure.json with hoisted levels and renamed functions")
    else:
        print(f"\nRun with --apply to save changes")


if __name__ == '__main__':
    main()
