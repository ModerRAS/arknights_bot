import sys
import numpy as np
from PIL import Image

FINAL = 'src/utils/media/testdata/visual/final'
OUT = '.iter'

def crop_pair(mid, x, y, w, h, name, zoom=2):
    a = Image.open(f'{FINAL}/old/{mid}.jpg').convert('RGB').crop((x, y, x+w, y+h))
    b = Image.open(f'{FINAL}/new/{mid}.png').convert('RGB').crop((x, y, x+w, y+h))
    a = a.resize((w*zoom, h*zoom), Image.NEAREST)
    b = b.resize((w*zoom, h*zoom), Image.NEAREST)
    canvas = Image.new('RGB', (w*zoom*2+8, h*zoom), (255, 0, 255))
    canvas.paste(a, (0, 0)); canvas.paste(b, (w*zoom+8, 0))
    canvas.save(f'{OUT}/{mid}-{name}.png')
    print(f'{OUT}/{mid}-{name}.png')

# box-detail: header + first row region
crop_pair('box-detail', 0, 0, 722, 140, 'top')
crop_pair('box-detail', 0, 139, 722, 140, 'bottom')

# box-summary: top area and mid rows
crop_pair('box-summary', 0, 0, 1350, 240, 'top')
crop_pair('box-summary', 0, 240, 1350, 250, 'mid')

# help: banner and one group
crop_pair('help', 0, 300, 990, 260, 'group1')
crop_pair('help', 0, 1500, 990, 260, 'lower')
