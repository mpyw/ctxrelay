#!/usr/bin/env python3
"""
Improve titles to make them more specific based on descriptions.
"""

import json
import re

def load_structure():
    with open('structure.json', 'r') as f:
        return json.load(f)

def save_structure(structure):
    with open('structure.json', 'w') as f:
        json.dump(structure, f, indent=2, ensure_ascii=False)
        f.write('\n')

def improve_title(test_id, test):
    """Generate improved title from description."""
    desc = test['description']
    current_title = test['title']

    # Remove "(was ...)" suffix from description
    desc = re.sub(r'\s*\(was [^)]+\)', '', desc)

    # For "AND" patterns, use the description after " - "
    if current_title == "AND":
        match = re.match(r'^AND\s*-\s*(.+)', desc)
        if match:
            specific = match.group(1).strip()
            # Capitalize first letter
            specific = specific[0].upper() + specific[1:] if specific else specific
            return f"AND - {specific}"

    # For "Basic" patterns, append the specific case from description
    if current_title == "Basic":
        match = re.match(r'^Basic\s*-\s*(.+)', desc)
        if match:
            specific = match.group(1).strip()
            return f"Basic - {specific}"

    # For "Mixed" patterns
    if current_title == "Mixed":
        match = re.match(r'^Mixed\s*-\s*(.+)', desc)
        if match:
            specific = match.group(1).strip()
            return f"Mixed - {specific}"

    # For function names that need context
    function_names = [
        'DoAllFnsSettled', 'DoAllSettled', 'DoAll', 'DoAllFns',
        'DoRace', 'DoRaceFns', 'Task', 'CancelableTask'
    ]

    for func_name in function_names:
        if current_title == func_name:
            # Extract the specific case from description
            match = re.match(rf'^{re.escape(func_name)}\s*-\s*(.+)', desc)
            if match:
                specific = match.group(1).strip()
                return f"{func_name} - {specific}"

    # For "Nested goroutine" - append the specific case
    if current_title == "Nested goroutine":
        match = re.match(r'^Nested goroutine\s*-\s*(.+)', desc)
        if match:
            specific = match.group(1).strip()
            return f"Nested goroutine - {specific}"

    # For "Three contexts" - append the specific case
    if current_title == "Three contexts":
        match = re.match(r'^Three contexts\s*-\s*(.+)', desc)
        if match:
            specific = match.group(1).strip()
            return f"Three contexts - {specific}"

    # For "Variable reassignment" - append the specific case
    if current_title == "Variable reassignment":
        match = re.match(r'^Variable reassignment\s*-\s*(.+)', desc)
        if match:
            specific = match.group(1).strip()
            # Shorten if too long
            if len(specific) > 60:
                specific = specific[:57] + "..."
            return f"Variable reassignment - {specific}"

    # For "Context passed via argument" - append the specific case
    if current_title == "Context passed via argument":
        match = re.match(r'^Context passed via argument\s*-\s*(.+)', desc)
        if match:
            specific = match.group(1).strip()
            return f"Context passed via argument - {specific}"

    # For "Arbitrary depth go fn" - append the specific case
    if "Arbitrary depth go fn" in current_title:
        if "without ctx" in desc:
            return "Arbitrary depth go fn()()() - without ctx"
        elif "with ctx" in desc:
            return "Arbitrary depth go fn(ctx)()() - with ctx"

    # For "Shadow with non-ctx type" - append the type
    if "Shadow with non-ctx type" in current_title:
        match = re.match(r'^Shadow with non-ctx type \(([^)]+)\)', desc)
        if match:
            type_name = match.group(1).strip()
            return f"Shadow with non-ctx type - {type_name}"

    # For "CancelableTask" - add without/with deriver
    if current_title == "CancelableTask":
        if "without deriver" in desc:
            return "CancelableTask - without deriver"
        elif "with deriver" in desc:
            return "CancelableTask - with deriver"

    # For "Task" - add specific case
    if current_title == "Task":
        match = re.match(r'^Task\.DoAsync\s*(.+)', desc)
        if match:
            specific = match.group(1).strip()
            return f"Task - {specific}"

    # For "Basic - func" goroutinecreator patterns
    if current_title.startswith("Basic - func"):
        if "(waitgroup)" in desc:
            return current_title.replace("ctx", "ctx (waitgroup)")
        elif "(errgroup)" in desc:
            return current_title.replace("ctx", "ctx (errgroup)")

    # For "go fn()" patterns - add more specificity
    if current_title in ["go fn()()", "go fn(ctx)()", "go fn()()()", "go fn(ctx)()()"]:
        if "higher-order without ctx" in desc:
            return "go fn()() - higher-order without ctx"
        elif "higher-order with ctx" in desc:
            return "go fn(ctx)() - higher-order with ctx"
        elif "triple higher-order without ctx" in desc:
            return "go fn()()() - triple higher-order without ctx"
        elif "triple higher-order with ctx" in desc:
            return "go fn(ctx)()() - triple higher-order with ctx"

    # For "Ignore directive" - add specific case
    if current_title == "Ignore directive":
        if "same line" in desc:
            return "Ignore directive - same line"
        elif "previous line" in desc:
            return "Ignore directive - previous line"
        elif "with reason" in desc:
            return "Ignore directive - with reason"

    # For "LIMITATION" - add specific case
    if current_title == "LIMITATION":
        if "channel" in desc.lower():
            return "LIMITATION - Function from channel"
        elif "interface{}" in desc or "type assertion" in desc:
            return "LIMITATION - Function through interface{}"
        elif "deferred" in desc.lower():
            return "LIMITATION - ctx in deferred nested closure"

    # For "Literal with ctx" - add specific case
    if current_title == "Literal with ctx":
        if "via function call" in desc:
            return "Literal with ctx - via function call"
        elif "basic good case" in desc:
            return "Literal with ctx - basic good case"

    # For "Literal with derived ctx" - add specific case
    if current_title == "Literal with derived ctx":
        if "errgroup.WithContext" in desc:
            return "Literal with derived ctx - errgroup.WithContext"
        elif "defer" in desc.lower():
            return "Literal with derived ctx - with defer"

    # For "Literal with ctx in select" - add specific case
    if current_title == "Literal with ctx in select":
        if "defer" in desc.lower():
            return "Literal with ctx in select - with defer"

    # For "Literal without ctx" - add specific case
    if current_title == "Literal without ctx":
        if "variant" in desc:
            return "Literal without ctx - variant"
        elif "TryGo" in desc:
            return "Literal without ctx - TryGo"
        elif "Pointer receiver" in desc:
            return "Literal without ctx - pointer receiver"

    # For "Multiple ctx params" - add specific case
    if current_title == "Multiple ctx params":
        if "reports first" in desc or "bad" in desc:
            return "Multiple ctx params - reports first (bad)"
        elif "uses first" in desc:
            return "Multiple ctx params - uses first (good)"
        elif "uses second" in desc:
            return "Multiple ctx params - uses second (good)"

    # For "Multiple ctx in separate param groups" - add specific case
    if "Multiple ctx in separate param groups" in current_title:
        if "none used" in desc:
            return "Multiple ctx in separate param groups - none used (bad)"
        elif "good" not in current_title:
            return "Multiple ctx in separate param groups - (good)"

    # For "Multiple func args" - add specific case
    if current_title == "Multiple func args":
        if "both bad" in desc:
            return "Multiple func args - both bad"
        elif "first bad" in desc:
            return "Multiple func args - first bad"
        elif "second bad" in desc:
            return "Multiple func args - second bad"
        elif "both good" in desc:
            return "Multiple func args - both good"

    # For "Interleaved layers" - add specific case
    if current_title == "Interleaved layers":
        if "ctx -> no ctx -> ctx" in desc:
            return "Interleaved layers - ctx->no ctx->ctx shadowing"
        elif "goroutine ignores" in desc:
            return "Interleaved layers - goroutine ignores shadowing ctx"

    # For "Middle layer introduces ctx" - add specific case
    if "Middle layer introduces ctx" in current_title:
        if "outer has none" in current_title:
            pass  # Already specific
        elif "goroutine ignores" in desc:
            return "Middle layer introduces ctx - goroutine ignores"

    # For "Higher-order with multiple ctx" - add factory info
    if current_title == "Higher-order with multiple ctx":
        if "ctx1" in desc:
            return "Higher-order with multiple ctx - factory receives ctx1"
        elif "ctx2" in desc:
            return "Higher-order with multiple ctx - factory receives ctx2"

    # For "Higher-order go fn()()" in DA/DM
    if current_title == "Higher-order go fn()()":
        if "first of AND" in desc:
            return "Higher-order go fn()() - only first of AND"

    # No improvement needed
    return current_title

def main():
    structure = load_structure()

    updated_count = 0

    for test_id, test in structure['tests'].items():
        new_title = improve_title(test_id, test)
        if new_title != test['title']:
            old_title = test['title']
            test['title'] = new_title
            print(f"✅ {test_id}: '{old_title}' → '{new_title}'")
            updated_count += 1

    if updated_count > 0:
        save_structure(structure)
        print(f"\n✅ Updated {updated_count} titles")
    else:
        print("\n✅ No titles needed updating")

if __name__ == '__main__':
    main()
