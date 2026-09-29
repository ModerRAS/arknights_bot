import sys, json
import numpy as np
from PIL import Image

FINAL = 'src/utils/media/testdata/visual/final'

def delta(a_path, b_path):
    a = np.asarray(Image.open(a_path).convert('RGB'), dtype=np.int32)
    b = np.asarray(Image.open(b_path).convert('RGB'), dtype=np.int32)
    assert a.shape == b.shape, (a.shape, b.shape)
    # replicate harness: sum of |dr|+|dg|+|db| per pixel (alpha=255 both sides cancels)
    d = np.abs(a - b).sum(axis=2)
    return a, b, d

def row_profile(d):
    return d.mean(axis=1)

def analyze(mid):
    old_p = f'{FINAL}/old/{mid}.jpg'
    new_p = f'{FINAL}/new/{mid}.png'
    a, b, d = delta(old_p, new_p)
    H, W = d.shape
    total = d.sum()
    mx = W * H * 3 * 255
    sim = 1 - total / mx
    rp = row_profile(d)
    print(f'== {mid} {W}x{H} raw-sim {sim:.4f}')
    # rows bucketed in 50px bands
    band = 50
    bands = [(rp[i:i+band].mean(), i) for i in range(0, H, band)]
    worst = sorted(bands, reverse=True)[:8]
    print('  worst bands (meanDelta, y):', [(round(v,1), y) for v, y in worst])
    quiet = [v for v,_ in bands if v < 1.0]
    print(f'  bands with meanDelta<1: {len(quiet)}/{len(bands)}')
    return a, b, d

for mid in sys.argv[1:]:
    analyze(mid)
