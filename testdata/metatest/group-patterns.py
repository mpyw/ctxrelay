#!/usr/bin/env python3
"""
Group patterns that should be unified.
"""

import json

def load_structure():
    with open('structure.json', 'r') as f:
        return json.load(f)

def save_structure(structure):
    with open('structure.json', 'w') as f:
        json.dump(structure, f, indent=2, ensure_ascii=False)
        f.write('\n')

def main():
    structure = load_structure()

    # Group 1: pattern77 + pattern85 (Higher-order with multiple ctx - factory receives ctx1)
    if 'pattern77' in structure['tests'] and 'pattern85' in structure['tests']:
        pattern77 = structure['tests']['pattern77']
        pattern85 = structure['tests']['pattern85']

        # Merge into pattern77
        pattern77['targets'] = ['goroutine', 'errgroup', 'waitgroup']
        pattern77['functions'] = {
            'goroutine': pattern77['functions']['goroutine'],
            'errgroup': pattern85['functions']['errgroup'],
            'waitgroup': pattern85['functions']['waitgroup']
        }
        pattern77['levels'] = {
            'goroutine': pattern77['levels']['goroutine'],
            'errgroup': pattern85['levels']['errgroup'],
            'waitgroup': pattern85['levels']['waitgroup']
        }

        # Remove pattern85
        del structure['tests']['pattern85']
        print("✅ Grouped pattern77 + pattern85 → pattern77")

    # Group 2: pattern78 + pattern86 (Higher-order with multiple ctx - factory receives ctx2)
    if 'pattern78' in structure['tests'] and 'pattern86' in structure['tests']:
        pattern78 = structure['tests']['pattern78']
        pattern86 = structure['tests']['pattern86']

        # Merge into pattern78
        pattern78['targets'] = ['goroutine', 'errgroup', 'waitgroup']
        pattern78['functions'] = {
            'goroutine': pattern78['functions']['goroutine'],
            'errgroup': pattern86['functions']['errgroup'],
            'waitgroup': pattern86['functions']['waitgroup']
        }
        pattern78['levels'] = {
            'goroutine': pattern78['levels']['goroutine'],
            'errgroup': pattern86['levels']['errgroup'],
            'waitgroup': pattern86['levels']['waitgroup']
        }

        # Remove pattern86
        del structure['tests']['pattern86']
        print("✅ Grouped pattern78 + pattern86 → pattern78")

    # Group 3: DA47 + DM47 (Variable reassignment - last assignment with incomplete derivers)
    if 'DA47' in structure['tests'] and 'DM47' in structure['tests']:
        DA47 = structure['tests']['DA47']
        DM47 = structure['tests']['DM47']

        # Merge into DA47
        DA47['targets'] = ['goroutinederiveand', 'goroutinederivemixed']
        DA47['functions'] = {
            'goroutinederiveand': DA47['functions']['goroutinederiveand'],
            'goroutinederivemixed': DM47['functions']['goroutinederivemixed']
        }
        DA47['levels'] = {
            'goroutinederiveand': DA47['levels']['goroutinederiveand'],
            'goroutinederivemixed': DM47['levels']['goroutinederivemixed']
        }

        # Remove DM47
        del structure['tests']['DM47']
        print("✅ Grouped DA47 + DM47 → DA47")

    save_structure(structure)
    print(f"\n✅ Total patterns: {len(structure['tests'])}")

if __name__ == '__main__':
    main()
