import sys
from PIL import Image

# Dev-time baseline measurement (honest: measurement only, never used in render/scoring path).
# Prints OLD(baseline jpg) vs NEW(candidate png) feature metrics in px space (scale 1.5).
old = Image.open('src/utils/media/testdata/visual/baseline/images/gacha.jpg').convert('RGB')
new = Image.open('.neargate/new/gacha.png').convert('RGB')

def rows(img, pred, x0, y0, x1, y1, minn=1, gap=2):
    px = img.load()
    bands = []
    cur = None
    for y in range(y0, y1):
        cnt = 0
        minx = None; maxx = None
        for x in range(x0, x1):
            if pred(px[x, y]):
                cnt += 1
                if minx is None: minx = x
                maxx = x
        if cnt >= minn:
            if cur is None:
                cur = [y, y, minx, maxx, cnt]
            else:
                if y - cur[1] > gap:
                    bands.append(cur); cur = [y, y, minx, maxx, cnt]
                else:
                    cur[1] = y; cur[2] = min(cur[2], minx); cur[3] = max(cur[3], maxx); cur[4] += cnt
        # keep band open across small gaps
    if cur is not None: bands.append(cur)
    return bands

def bbox(img, pred, x0, y0, x1, y1):
    px = img.load()
    minx, miny, maxx, maxy = 10**9, 10**9, -1, -1
    for y in range(y0, y1):
        for x in range(x0, x1):
            if pred(px[x, y]):
                minx = min(minx, x); maxx = max(maxx, x)
                miny = min(miny, y); maxy = max(maxy, y)
    return (minx, miny, maxx, maxy)

W = lambda c: c[0] > 190 and c[1] > 190 and c[2] > 190
colors = {
    'orange': lambda c: abs(c[0]-244) < 20 and abs(c[1]-110) < 20 and abs(c[2]-30) < 20,
    'yellow': lambda c: abs(c[0]-247) < 20 and abs(c[1]-171) < 20 and abs(c[2]-55) < 20,
    'purple': lambda c: abs(c[0]-161) < 20 and abs(c[1]-53) < 20 and abs(c[2]-246) < 20,
    'gray':   lambda c: abs(c[0]-109) < 12 and abs(c[1]-116) < 12 and abs(c[2]-126) < 12,
    'blue':   lambda c: abs(c[0]-84) < 25 and abs(c[1]-112) < 25 and abs(c[2]-198) < 25,
}

for name, img in (('OLD', old), ('NEW', new)):
    print('====', name)
    # 1. header name text
    b = bbox(img, W, 460, 60, 720, 145)
    print('name bbox', b)
    # 2. header total/period line
    b = bbox(img, W, 360, 150, 1260, 240)
    print('total bbox', b)
    # 3. pie (slice colors only, right of legend zone)
    xs = []; ys = []
    for cn in ('orange', 'yellow', 'purple'):
        b = bbox(img, colors[cn], 300, 280, 483, 610)
        xs += [b[0], b[2]]; ys += [b[1], b[3]]
    print('pie bbox', (min(xs), min(ys), max(xs), max(ys)),
          'center', ((min(xs)+max(xs))/2, (min(ys)+max(ys))/2))
    # 4. legend icons per color (x<300) + legend text bands
    for cn in ('orange', 'yellow', 'purple', 'gray'):
        b = bbox(img, colors[cn], 30, 300, 75, 610)
        print('legend icon', cn, b)
    tb = rows(img, W, 78, 320, 300, 600, minn=3)
    print('legend text bands:', [(r[0], r[1], r[2], r[3]) for r in tb])
    # 5. avg card text rows (card2 x 510..965)
    tb = rows(img, W, 520, 290, 960, 580, minn=3)
    print('avg rows:', [(r[0], r[1], r[2], r[3]) for r in tb])
    # 6. pool bars
    for i in range(2):
        pass
    bb = bbox(img, colors['blue'], 990, 225, 1460, 610)
    print('bars union', bb)
    # per-row blue to split bars
    tb = rows(img, colors['blue'], 990, 225, 1460, 610, minn=5)
    print('bar row bands:', [(r[0], r[1], r[2], r[3]) for r in tb])
    # 7. pool card title + labels
    tb = rows(img, W, 1000, 240, 1100, 600, minn=2)
    print('pool label bands:', [(r[0], r[1], r[2], r[3]) for r in tb])
    # 8. list box titles
    tb = rows(img, W, 60, 660, 700, 740, minn=3)
    print('list1 title bands:', [(r[0], r[1], r[2], r[3]) for r in tb])
    tb = rows(img, W, 800, 660, 1440, 740, minn=3)
    print('list2 title bands:', [(r[0], r[1], r[2], r[3]) for r in tb])
    # 9. list1 text lines (right of first avatar)
    tb = rows(img, W, 195, 740, 350, 920, minn=3)
    print('list1 entry1 text bands:', [(r[0], r[1], r[2], r[3]) for r in tb])
    # 10. footer date
    b = bbox(img, W, 700, 1150, 1300, 1300)
    print('footer date bbox', b)
