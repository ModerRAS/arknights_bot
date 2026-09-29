import sys
import numpy as np
from PIL import Image, ImageDraw

FINAL = 'src/utils/media/testdata/visual/final'

def grid_crop(mid, x, y, w, h, name, zoom=3, grid=10):
    a = Image.open(f'{FINAL}/old/{mid}.jpg').convert('RGB').crop((x, y, x+w, y+h))
    b = Image.open(f'{FINAL}/new/{mid}.png').convert('RGB').crop((x, y, x+w, y+h))
    Wc, Hc = w*zoom, h*zoom
    canvas = Image.new('RGB', (Wc*2+10, Hc), (255, 0, 255))
    canvas.paste(a.resize((Wc, Hc), Image.NEAREST), (0, 0))
    canvas.paste(b.resize((Wc, Hc), Image.NEAREST), (Wc+10, 0))
    dr = ImageDraw.Draw(canvas)
    for gx in range(0, w, grid):
        col = (255, 255, 0) if gx % 50 == 0 else (120, 120, 0)
        dr.line([(gx*zoom, 0), (gx*zoom, Hc)], fill=col, width=1)
        dr.line([(Wc+10+gx*zoom, 0), (Wc+10+gx*zoom, Hc)], fill=col, width=1)
        if gx % 50 == 0:
            dr.text((gx*zoom+2, 2), str(x+gx), fill=(255,255,0))
            dr.text((Wc+10+gx*zoom+2, 2), str(x+gx), fill=(255,255,0))
    for gy in range(0, h, grid):
        col = (255, 255, 0) if gy % 50 == 0 else (120, 120, 0)
        dr.line([(0, gy*zoom), (Wc*2+10, gy*zoom)], fill=col, width=1)
        if gy % 50 == 0:
            dr.text((2, gy*zoom+2), str(y+gy), fill=(255,255,0))
            dr.text((Wc+12, gy*zoom+2), str(y+gy), fill=(255,255,0))
    canvas.save(f'.iter/{mid}-{name}.png')
    print(f'.iter/{mid}-{name}.png')

mid, x, y, w, h, name = sys.argv[1], int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4]), int(sys.argv[5]), sys.argv[6]
grid_crop(mid, x, y, w, h, name)
