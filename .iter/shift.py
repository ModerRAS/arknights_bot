import sys
import numpy as np
from PIL import Image

FINAL = 'src/utils/media/testdata/visual/final'

def load(mid):
    a = np.asarray(Image.open(f'{FINAL}/old/{mid}.jpg').convert('L'), dtype=np.float32)
    b = np.asarray(Image.open(f'{FINAL}/new/{mid}.png').convert('L'), dtype=np.float32)
    return a, b

def best_shift(old, new, y0, y1, x0, x1, rng=32):
    # find (dx,dy) minimizing delta of new shifted vs old within window
    o = old[y0:y1, x0:x1]
    best = (1e18, 0, 0)
    for dy in range(-rng, rng+1, 1):
        for dx in range(-rng, rng+1, 1):
            n = new[y0+dy:y1+dy, x0+dx:x1+dx]
            if n.shape != o.shape: continue
            v = np.abs(o - n).mean()
            if v < best[0]: best = (v, dx, dy)
    return best

mid = sys.argv[1]
old, new = load(mid)
H, W = old.shape
# tiles: vertical strips of 120px height, full width; report best shift per tile
tile_h = 60
for y in range(0, H, tile_h):
    y1 = min(y+tile_h, H)
    v, dx, dy = best_shift(old, new, y, y1, 0, W)
    print(f'y={y:>4}..{y1:>4}: shift dx={dx:>3} dy={dy:>3} meandelta={v:7.2f}')
