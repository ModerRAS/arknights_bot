import sys
import numpy as np
from PIL import Image

FINAL = 'src/utils/media/testdata/visual/final'

def load(mid):
    a = np.asarray(Image.open(f'{FINAL}/old/{mid}.jpg').convert('RGB'), dtype=np.int32)
    b = np.asarray(Image.open(f'{FINAL}/new/{mid}.png').convert('RGB'), dtype=np.int32)
    return a, b

for mid in sys.argv[1:]:
    a, b = load(mid)
    d = np.abs(a - b).sum(axis=2)
    H, W = d.shape
    tot = d.sum()
    hist = [
        ('d==0', int((d == 0).sum())),
        ('0<d<=10', int(((d > 0) & (d <= 10)).sum())),
        ('10<d<=60', int(((d > 10) & (d <= 60)).sum())),
        ('60<d<=200', int(((d > 60) & (d <= 200)).sum())),
        ('d>200', int((d > 200).sum())),
    ]
    print(f'== {mid} {W}x{H} total_delta_share={tot/(W*H*765):.4f}')
    for name, n in hist:
        print(f'   {name:>10}: {n/W/H*100:6.2f}%')
    # sample background patch stats (corners)
    for label, sl in [('top-left', (slice(0,30), slice(0,30))), ('bottom-right', (slice(H-30,H), slice(W-30,W)))]:
        dd = d[sl]
        print(f'   bg {label}: meanDelta={dd.mean():.1f}')
