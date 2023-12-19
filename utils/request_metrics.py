#!/usr/bin/env python3

import fileinput
import json
from collections import defaultdict

def main():
    counts = defaultdict(int)
    times = defaultdict(int)
    for line in fileinput.input():
        entry = json.loads(line)
        if entry.get('msg') not in ('call jsonrpc', 'relay http'):
            continue
        key = (entry['chain'], entry['method'], entry['endpoint'])
        counts[key] += 1
        times[key] += entry['timeSpentMS']

    print('chain', 'method', 'endopint', 'avgtime', 'count')
    for key, cnt in sorted(counts.items()):
        tm = times[key]  # total request time
        avg = int(tm/cnt) # average request time
        print(key[0], key[1], key[2], avg, cnt)

if __name__ == '__main__':
    main()

