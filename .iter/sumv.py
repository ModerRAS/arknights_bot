import sys
import numpy as np
from PIL import Image

FINAL = 'src/utils/media/testdata/visual/final'

old = np.asarray(Image.open(f'{FINAL}/old/box-summary.jpg').convert('L'), dtype=np.float32)
new = np.asarray(Image.open(f'{FINAL}/new/box-summary.png').convert('L'), dtype=np.float32)
H, W = old.shape
print('shape', old.shape)

def sep_rows(img, label):
    # rows that are mostly bright (separator lines): count pixels > 180
    cnt = (img > 180).sum(axis=1)
    rows = np.where(cnt > W*0.5)[0]
    # group consecutive
    groups = []
    for r in rows:
        if groups and r - groups[-1][-1] <= 2: groups[-1].append(r)
        else: groups.append([r])
    print(label, 'separator rows:', [(g[0], g[-1]) for g in groups])

sep_rows(old, 'old')
sep_rows(new, 'new')

def text_bands(img, x0, x1, y0, y1, thr=150, minpx=2, label=''):
    prof = (img[y0:y1, x0:x1] > thr).sum(axis=1)
    out = []
    start = None
    for i, v in enumerate(prof):
        if v >= minpx and start is None: start = i
        elif v < minpx and start is not None:
            out.append((y0+start, y0+i-1)); start = None
    if start is not None: out.append((y0+start, y0+len(prof)-1))
    print(label, out)

# column 1 x-range guess: label text starts ~x45 (30+10?) measure text bands in x30..200
text_bands(old, 30, 210, 90, 480, label='old col1 text bands')
text_bands(new, 30, 210, 90, 480, label='new col1 text bands')
# header th row text
text_bands(old, 0, W, 60, 95, label='old th text')
text_bands(new, 0, W, 60, 95, label='new th text')
# title
text_bands(old, 0, 400, 0, 60, label='old title')
text_bands(new, 0, 400, 0, 60, label='new title')
