"""Grid-probe state knobs: for each combo set knobs, run harness, print score."""
import re, subprocess, sys, itertools

P = 'src/ggrender/scene_state.go'

def setknobs(kvs):
    s = open(P, encoding='utf-8').read()
    for k, v in kvs.items():
        s = re.sub(rf'{k}\s*= [\-\d.]+', f'{k} = {v}', s)
    open(P, 'w', encoding='utf-8').write(s)

def run():
    r = subprocess.run(['go', 'test', './ggrender/', '-run', 'TestGGPixelParity'],
                       cwd='src', capture_output=True, text=True, timeout=600,
                       encoding='utf-8', errors='replace')
    m = re.search(r'scene state\s+\S+ scale=\S+ similarity=([\d.]+)', r.stdout + r.stderr)
    return float(m.group(1)) if m else None

if __name__ == '__main__':
    # args: k1=v1a,v1b,k2=v2a,v2b -> cartesian
    keys = []; grids = []
    base = {}
    for a in sys.argv[1:]:
        k, vs = a.split('=')
        vals = vs.split(',')
        keys.append(k); grids.append(vals)
        m = re.search(rf'{k}\s*= ([\-\d.]+)', open(P, encoding='utf-8').read())
        base[k] = float(m.group(1))
    best = (None, -1)
    for combo in itertools.product(*grids):
        kv = dict(zip(keys, combo))
        allkv = {**base, **{k: float(v) for k, v in kv.items()}}
        setknobs(allkv)
        sc = run()
        print(f'{kv} -> {sc}', flush=True)
        if sc and sc > best[1]: best = ((kv), sc)
    print('BEST:', best)
