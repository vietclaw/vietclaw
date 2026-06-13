import urllib.request
import json

try:
    req = urllib.request.Request(
        'http://127.0.0.1:4000/api/memory/record',
        data=json.dumps({'content': 'Regex compilation in loops (or even hot functions) is extremely slow in Go. Moving them to package-level var = regexp.MustCompile(...) and pre-allocating slices significantly improves CPU throughput and reduces allocations.'}).encode(),
        headers={'Content-Type': 'application/json'}
    )
    res = urllib.request.urlopen(req)
    print(res.read().decode())
except Exception as e:
    print(f"Failed to record memory: {e}")
