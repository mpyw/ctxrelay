#!/usr/bin/env python3
"""
Add missing test patterns to structure.json with temporary pattern IDs.
"""

import json
from pathlib import Path

def load_structure():
    with open('structure.json', 'r') as f:
        return json.load(f)

def save_structure(structure):
    with open('structure.json', 'w') as f:
        json.dump(structure, f, indent=2, ensure_ascii=False)
        f.write('\n')

def main():
    structure = load_structure()

    # Find highest pattern number
    max_num = 0
    for test_name in structure['tests'].keys():
        if test_name.startswith('pattern'):
            num_str = test_name.replace('pattern', '')
            # Remove suffix like 'b', 'c'
            num_str = ''.join(c for c in num_str if c.isdigit())
            if num_str:
                max_num = max(max_num, int(num_str))

    next_num = max_num + 1
    print(f"Starting from pattern{next_num:02d}")

    # errgroup/waitgroup unified patterns
    unified_patterns = [
        # Existing pattern codes
        ('goodStructFieldWithCtx', 'Struct field with ctx', 'GE18/GW18'),
        ('goodInterfaceMethodWithCtxArg', 'Interface method with ctx argument', 'GE100/GW100'),
        ('badInterfaceMethodWithoutCtxArg', 'Interface method without ctx argument', 'GE100b/GW100b'),
        ('badStructFieldWithoutCtx', 'Function from struct field without ctx', 'GE103/GW103'),
        # No pattern code
        ('badFuncThroughInterfaceWithoutCtx', 'Function through interface without ctx - LIMITATION', None),
        ('evilShadowingInnerHasCtx', 'Shadowing - inner ctx shadows outer', None),
        ('evilShadowingInnerIgnoresCtx', 'Shadowing - inner ignores ctx', None),
        ('evilShadowingTwoLevels', 'Two levels of shadowing', None),
        ('evilShadowingTwoLevelsBad', 'Two levels of shadowing - bad case', None),
        ('evilMiddleLayerIntroducesCtx', 'Middle layer introduces ctx', None),
        ('evilMiddleLayerIntroducesCtxGood', 'Middle layer introduces ctx - good case', None),
        ('evilInterleavedLayers', 'Interleaved layers', None),
        ('evilInterleavedLayersGood', 'Interleaved layers - good case', None),
    ]

    for func_name, title, old_code in unified_patterns:
        pattern_id = f"pattern{next_num:02d}"
        structure['tests'][pattern_id] = {
            'title': title,
            'description': title + (f" (was {old_code})" if old_code else ""),
            'targets': ['errgroup', 'waitgroup'],
            'functions': {
                'errgroup': func_name,
                'waitgroup': func_name
            },
            'levels': {
                'errgroup': 'evil',
                'waitgroup': 'evil'
            }
        }
        print(f"✅ Added {pattern_id}: {title}")
        next_num += 1

    # gotask/evil.go patterns
    gotask_patterns = [
        # Existing pattern codes
        ('badDerivedInNestedClosure', 'Deriver only in nested closure', 'EV20'),
        ('limitationVariadicExpansionVariable', 'LIMITATION: Variable slice expansion', 'EV100'),
        ('limitationDerivedInDeferClosure', 'LIMITATION: defer closure not traversed', 'EV110'),
        ('limitationInterfaceTaskMaker', 'LIMITATION: Interface method returns', 'EV120'),
        # Edge cases
        ('goodEmptyDoAll', 'Edge case: Empty call (less than 2 args)', None),
        ('goodOnlyCtxArg', 'Edge case: Only ctx arg', None),
        ('badMultipleDoAsync', 'Edge case: Multiple DoAsync calls', None),
        ('badDifferentCtxName', 'Edge case: Context with different param name', None),
        ('badContextParamUnusualName', 'Edge case: Context param with unusual name', None),
        ('goodDifferentCtxNames', 'Edge case: Good with different ctx param names', None),
        ('badNestedTaskCreation', 'Edge case: Nested task creation', None),
        # Deriver patterns
        ('goodDerivedUsedInExpression', 'Deriver result used directly in expression', None),
        ('goodDerivedStoredAndUsed', 'Deriver result stored and used', None),
        ('goodDerivedBeforeEarlyReturn', 'Deriver called before early return', None),
        ('goodDerivedOnOneBranch', 'Deriver only on one branch (but detected)', None),
    ]

    for func_name, title, old_code in gotask_patterns:
        pattern_id = f"pattern{next_num:02d}"
        structure['tests'][pattern_id] = {
            'title': title,
            'description': title + (f" (was {old_code})" if old_code else ""),
            'targets': ['gotask'],
            'functions': {
                'gotask': func_name
            },
            'levels': {
                'gotask': 'evil'
            }
        }
        print(f"✅ Added {pattern_id}: {title}")
        next_num += 1

    save_structure(structure)
    print(f"\n✅ Added {next_num - max_num - 1} new patterns")
    print(f"Total patterns: {len(structure['tests'])}")

if __name__ == '__main__':
    main()
