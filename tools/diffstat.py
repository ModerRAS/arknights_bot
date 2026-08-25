"""diffstat: region-wise red-density analysis of harness diff.png output.

Reads ONLY tmp/pixel-compare/<scene>/diff.png (sanctioned calibration artifact)
and new.png (our own render). Never touches testdata/baseline images.
Prints a grid of mismatch percentages to guide block-by-block calibration.
"""
import sys
from PIL import Image

def diffstat(scene, cols=6, rows=6):
    base = f"tmp/pixel-compare/{scene}"
    diff = Image.open(f"{base}/diff.png").convert("RGB")
    W, H = diff.size
    px = diff.load()
    total = W * H
    red = 0
    grid = [[0] * cols for _ in range(rows)]
    cw, ch = W / cols, H / rows
    for y in range(H):
        gr = min(rows - 1, int(y / ch))
        for x in range(W):
            r, g, b = px[x, y]
            if r > 200 and g < 60 and b < 60:
                red += 1
                grid[gr][min(cols - 1, int(x / cw))] += 1
    print(f"{scene}: {W}x{H} mismatch {red}/{total} = {red/total*100:.2f}%")
    cw_i, ch_i = int(cw), int(ch)
    for ri, row in enumerate(grid):
        cells = " ".join(f"{c*100/(cw_i*ch_i):5.1f}" for c in row)
        print(f"  y{int(ri*ch):4d}-{int((ri+1)*ch):4d} | {cells}")

if __name__ == "__main__":
    for scene in sys.argv[1:]:
        diffstat(scene)
