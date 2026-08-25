"""Toggle overlay element groups via marker comments, run harness, report state score."""
import re, subprocess, sys

P = 'src/ggrender/scene_state.go'

GROUPS = {
    'sections': ['sections'],
    'campaign': ['campaign cluster'],
    'training': ['training room card'],
    'tower': ['tower lower row', 'tower higher row'],
    'topleft': ['avatar 54x54', 'name "Dr', 'checked-in flag', 'ap icon + counter'],
}

def disable(group):
    s = open(P, encoding='utf-8').read()
    lines = s.split('\n')
    out, skip = [], False
    for ln in lines:
        if any(k in ln for k in GROUPS[group]) and not ln.strip().startswith('//'):
            indent = len(ln) - len(ln.lstrip('\t'))
            out.append('\t' * indent + 'if false { // PROBE-DISABLED: ' + ln.strip())
            # find end: naive — wrap single statement lines only
            if not ln.rstrip().endswith('{'):
                out[-1] += ' }'
                continue
            skip = True
            depth = ln.count('{') - ln.count('}')
            continue
        if skip:
            depth += ln.count('{') - ln.count('}')
            if depth == 0:
                skip = False
        out.append(ln)
    open(P, 'w', encoding='utf-8').write('\n'.join(out))

def restore():
    s = open(P, encoding='utf-8').read()
    s = re.sub(r'if false \{ // PROBE-DISABLED: (.*) \}', r'\1', s)
    open(P, 'w', encoding='utf-8').write(s)

def run():
    subprocess.run(['go', 'build', './ggrender/'], cwd='src', capture_output=True)
    r = subprocess.run(['go', 'test', './ggrender/', '-run', 'TestGGPixelParity'],
                       cwd='src', capture_output=True, text=True, timeout=600,
                       encoding='utf-8', errors='replace')
    m = re.search(r'scene state\s+\S+ scale=\S+ similarity=([\d.]+)', r.stdout + r.stderr)
    return float(m.group(1)) if m else None

if __name__ == '__main__':
    if sys.argv[1] == 'restore':
        restore()
        print('restored')
    else:
        disable(sys.argv[1])
        print(f'{sys.argv[1]} disabled -> {run()}')
        restore()
