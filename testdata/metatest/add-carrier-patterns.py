#!/usr/bin/env python3
"""
Add missing carrier test patterns to structure.json.
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

    # carrier patterns
    carrier_patterns = [
        # No pattern code
        ('badEchoHandler', 'Basic echo handler without context usage', None),
        ('goodEchoHandler', 'Basic echo handler with context usage', None),
        ('badGoroutineInEchoHandler', 'Goroutine in echo handler without context', None),
        ('goodGoroutineInEchoHandler', 'Goroutine in echo handler with context', None),
        # Existing pattern codes
        ('goodMixedCtxAndCarrierUsesCarrier', 'Mixed context and carrier - uses carrier', 'CR01'),
        ('badMixedCtxAndCarrierUsesNeither', 'Mixed context and carrier - uses neither', 'CR02'),
        ('goodCarrierAsSecondParam', 'Carrier as second param - uses it', 'CR03'),
        ('badCarrierAsSecondParam', 'Carrier as second param - does not use it', 'CR04'),
    ]

    for func_name, title, old_code in carrier_patterns:
        pattern_id = f"pattern{next_num:02d}"
        structure['tests'][pattern_id] = {
            'title': title,
            'description': title + (f" (was {old_code})" if old_code else ""),
            'targets': ['carrier'],
            'functions': {
                'carrier': func_name
            },
            'levels': {
                'carrier': 'carrier'
            }
        }
        print(f"✅ Added {pattern_id}: {title}")
        next_num += 1

    save_structure(structure)
    print(f"\n✅ Added {next_num - max_num - 1} new patterns")
    print(f"Total patterns: {len(structure['tests'])}")

if __name__ == '__main__':
    main()
