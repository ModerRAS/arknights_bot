import sys
import numpy as np
from PIL import Image

FINAL = 'src/utils/media/testdata/visual/final'

def load(mid):
    a = np.asarray(Image.open(f'{FINAL}/old/{mid}.jpg').convert('L'), dtype=np.float32)
    b = np.asarray(Image.open(f'{FINAL}/new/{mid}.png').convert('L'), dtype=np.float32)
    return a, b

def bbox(img, x0, x1, y0, y1, thr=140):
    # bbox of pixels brighter than thr within window (device coords)
    win = img[y0:y1, x0:x1]
    ys, xs = np.where(win > thr)
    if len(xs) == 0: return None
    return (x0+xs.min(), x0+xs.max(), y0+ys.min(), y0+ys.max())

mid = sys.argv[1]
old, new = load(mid)
print(mid, old.shape)
# windows guessed from grid crop (device px)
wins = {
  'hdr-ganbu':   (40, 190, 5, 55),
  'hdr-dengji':  (230, 350, 5, 55),
  'avatar-row1': (0, 110, 55, 165),
  'name-row1':   (115, 240, 55, 165),
  'evolve1':     (235, 360, 55, 165),
  'potential1':  (360, 470, 55, 165),
  'skill1':      (470, 640, 55, 165),
  'equip1':      (640, 722, 55, 165),
  'evolve2':     (235, 360, 165, 279),
  'skill2':      (470, 640, 165, 279),
}
for name, (x0, x1, y0, y1) in wins.items():
    bo = bbox(old, x0, x1, y0, y1)
    bn = bbox(new, x0, x1, y0, y1)
    def fmt(b):
        return f'x{b[0]}..{b[1]} y{b[2]}..{b[3]} (w{b[1]-b[0]+1} h{b[3]-b[2]+1})' if b else 'none'
    print(f'{name:>12} old {fmt(bo)}')
    print(f'{"":>12} new {fmt(bn)}')
