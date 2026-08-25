"""Sweep depot/state layout knobs: set constant, run harness, report score."""
import re, subprocess, sys

P = 'src/ggrender/scene_depot.go'

def setknob(key, val):
    s = open(P, encoding='utf-8').read()
    s2 = re.sub(rf'{key}\s*= [\-\d.]+', f'{key} = {val}', s)
    assert s2 != s or f'{key} = {val}' in s, key
    open(P, 'w', encoding='utf-8').write(s2)

def run():
    r = subprocess.run(['go', 'test', './ggrender/', '-run', 'TestGGPixelParity'],
                       cwd='src', capture_output=True, text=True, timeout=600,
                       encoding='utf-8', errors='replace')
    m = re.search(r'scene depot\s+\S+ scale=\S+ similarity=([\d.]+)', r.stdout + r.stderr)
    return float(m.group(1)) if m else None

if __name__ == '__main__':
    for arg in sys.argv[1:]:
        key, val = arg.split('=')
        setknob(key, val)
        score = run()
        print(f'{key}={val} -> {score}')
