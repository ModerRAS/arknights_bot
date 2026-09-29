import sys
import numpy as np
from PIL import Image

FINAL = 'src/utils/media/testdata/visual/final'

def load(mid):
    a = np.asarray(Image.open(f'{FINAL}/old/{mid}.jpg').convert('L'), dtype=np.float32)
    b = np.asarray(Image.open(f'{FINAL}/new/{mid}.png').convert('L'), dtype=np.float32)
    return a, b

def runs(profile, thr):
    out = []
    start = None
    for i, v in enumerate(profile):
        if v >= thr and start is None: start = i
        elif v < thr and start is not None:
            out.append((start, i-1)); start = None
    if start is not None: out.append((start, len(profile)-1))
    return out

def col_clusters(img, y0, y1, thr=130, min_px=3):
    prof = (img[y0:y1] > thr).sum(axis=0)
    return [(a, b) for a, b in runs(prof, min_px)]

def row_clusters(img, x0, x1, y0, y1, thr=130, min_px=3):
    prof = (img[y0:y1, x0:x1] > thr).sum(axis=1)
    return [(y0+a, y0+b) for a, b in runs(prof, min_px)]

mid = sys.argv[1]
old, new = load(mid)
print(f'== {mid} row-band column clusters (bright x-runs)')
for label, y0, y1 in [('hdr', 5, 55), ('row1', 55, 165), ('row2', 165, 279)]:
    co = col_clusters(old, y0, y1)
    cn = col_clusters(new, y0, y1)
    print(f'{label} old: {co}')
    print(f'{label} new: {cn}')
print('== row1 col-band row clusters (bright y-runs)')
for label, x0, x1 in [('col1', 0, 235), ('col2-3', 235, 470), ('col4', 470, 645), ('col5', 645, 722)]:
    ro = row_clusters(old, x0, x1, 55, 279)
    rn = row_clusters(new, x0, x1, 55, 279)
    print(f'{label} old: {ro}')
    print(f'{label} new: {rn}')
