# embolden.py: synthesize a faux-bold TTF from NotoSansHans-Regular by offsetting
# glyf outlines along per-vertex averaged edge normals (FreeType FT_Outline_EmboldenXY
# style, simplified). Emulates Skia/Chromium synthetic bold at the FONT level so any
# fontSize gets consistent thickening. Experiment artifact -- not committed as-is.
import sys, math
from fontTools.ttLib import TTFont
from fontTools.ttLib.tables._g_l_y_f import GlyphCoordinates
from fontTools.pens.recordingPen import RecordingPen

def embolden_contour(points, deltas):
    # points: list of (x,y,isOnCurve); naive: shift every point by its averaged normal
    n = len(points)
    out = []
    for i in range(n):
        px, py, _ = points[(i - 1) % n]
        cx, cy, on = points[i]
        nx_, ny_, _ = points[(i + 1) % n]
        ax, ay = cx - px, cy - py
        bx, by = nx_ - cx, ny_ - cy
        la = math.hypot(ax, ay) or 1.0
        lb = math.hypot(bx, by) or 1.0
        # outward normals of incoming/outgoing edges (TT y-up; rotate +90: (-dy, dx))
        nia = (-ay / la, ax / la)
        nib = (-by / lb, bx / lb)
        mx, my = nia[0] + nib[0], nia[1] + nib[1]
        ml = math.hypot(mx, my)
        if ml < 1e-9:
            out.append((cx, cy, on)); continue
        dx, dy = deltas
        # FreeStyle per-axis scaling: ensure at least `strength` displacement per axis
        fx = dx / max(abs(mx), 1e-9) if abs(mx) > 1e-9 else 0
        fy = dy / max(abs(my), 1e-9) if abs(my) > 1e-9 else 0
        f = min(fx, fy) if (fx and fy) else (fx or fy)
        # clamp so tangential points don't overshoot
        f = max(min(f, 2 * dx), -2 * dx)
        out.append((cx + mx * f, cy + my * f, on))
    return out

def process(glyphs, delta):
    for name in glyphs.keys():
        g = glyphs[name]
        if not hasattr(g, 'coordinates') or g.numberOfContours == 0:
            continue
        coords = g.coordinates
        pts = [(c[0], c[1], i) for i, c in enumerate(coords)]
        # walk contours with endpoints
        ends = g.endPtsOfContours
        newpts = list(pts)
        start = 0
        for end in ends:
            contour = pts[start:end + 1]
            emb = embolden_contour(contour, (delta, delta))
            for j, p in enumerate(emb):
                newpts[start + j] = p
            start = end + 1
        g.coordinates = GlyphCoordinates([(round(x), round(y)) for x, y, _ in newpts])

src, dst, strength = sys.argv[1], sys.argv[2], float(sys.argv[3])
f = TTFont(src)
glyf = f['glyf']
process(glyf, strength)
if 'OS/2' in f:
    f['OS/2'].usWeightClass = 600
f.save(dst)
print('saved', dst, 'strength', strength)
