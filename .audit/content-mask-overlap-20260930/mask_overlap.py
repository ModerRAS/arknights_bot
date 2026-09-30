#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
False-friend scan (content-mask / "ink" overlap) -- VERBATIM re-implementation.

Provenance
----------
origin            : C:/WorkSpace/Golang/_lead4/false-friend-scan-16.json
original_script   : NEVER PERSISTED to disk. Recovered verbatim from the pi
                    session transcript that produced the JSON:
                    C:/Users/ModerRAS/.pi/agent/sessions/--C--WorkSpace-Golang--/
                      2026-09-30T01-03-02-285Z_01a0efd6-1e8c-761a-a158-2582d00bd136.jsonl
                    (lead-4 session, tool call id call_function_6xsn6ccngb3p_1)
provenance        : retroactive-migration
original_path     : C:/WorkSpace/Golang/_lead4/  (dir listing at 2026-09-30 09:2x
                    contained 8 PNGs + 3 JSONs + 1 MD and NO script file)

Definition (exact, as in the original):
  bg(img)   = exact modal 24-bit RGB (np.argmax of unique-value counts; NO quantization)
  ink mask  = np.abs(pixel - bg).sum(axis=2) > 30      # L1 over R+G+B, threshold 30
  inkOv%    = 100 * |A & B| / |A | B|                  # Jaccard (union denominator)
  rowCorr   = Pearson r of per-row ink counts
  colCorr   = Pearson r of per-column ink counts
  flatBg%   = 100 * max(modal count over the two images) / (H * W)
  verdict   = STRUCT-RELATED  if inkOv >= 50 and max(|rowCorr|,|colCorr|) >= 0.5
               UNRELATED/FAKE  if inkOv <  50 and max(|rowCorr|,|colCorr|) <  0.5
               BORDERLINE      otherwise

Inputs: <root>/tmp/pixel-compare/<scene>/{old.png,new.png} -- the frozen
Playwright baseline and the gg render, both already written by the pixel-parity
harness. No baseline byte is ever used as a render input. No scoring code is
touched: this file imports nothing from the repo.
"""
import io
import json
import os
import sys

import numpy as np
from PIL import Image


def modal(img):
    v = (img[..., 0] << 16) | (img[..., 1] << 8) | img[..., 2]
    u, c = np.unique(v, return_counts=True)
    return int(u[c.argmax()]), c.max()


def corr(a, b):
    a = a - a.mean()
    b = b - b.mean()
    d = np.sqrt((a * a).sum() * (b * b).sum())
    return float((a * b).sum() / d) if d > 0 else 0.0


def scan(root, scenes, rep_path=None):
    if rep_path is None:
        rep_path = os.path.join(root, "tmp", "pixel-compare", "report.json")
    rep = {e["scene"]: e for e in json.load(io.open(rep_path, encoding="utf-8"))}
    out = []
    for s in scenes:
        o = np.asarray(Image.open(os.path.join(root, "tmp", "pixel-compare", s, "old.png")).convert("RGB"), dtype=np.int64)
        n = np.asarray(Image.open(os.path.join(root, "tmp", "pixel-compare", s, "new.png")).convert("RGB"), dtype=np.int64)
        mo, co = modal(o)
        mn, cn = modal(n)
        mo = np.array([(mo >> 16) & 255, (mo >> 8) & 255, mo & 255])
        mn = np.array([(mn >> 16) & 255, (mn >> 8) & 255, mn & 255])
        eo = np.abs(o - mo).sum(axis=2) > 30
        en = np.abs(n - mn).sum(axis=2) > 30
        ov = 100 * (eo & en).sum() / max(1, (eo | en).sum())
        rc = corr(eo.sum(axis=1).astype(float), en.sum(axis=1).astype(float))
        cc = corr(eo.sum(axis=0).astype(float), en.sum(axis=0).astype(float))
        fb = 100 * max(co, cn) / o.shape[0] / o.shape[1]
        mx = max(abs(rc), abs(cc))
        v = "STRUCT-RELATED" if (ov >= 50 and mx >= 0.5) else ("UNRELATED/FAKE" if (ov < 50 and mx < 0.5) else "BORDERLINE")
        out.append(dict(scene=s, sim=rep[s]["similarity"], passed=rep[s]["passed"],
                        inkOv=round(ov, 1), rowCorr=round(rc, 3), colCorr=round(cc, 3),
                        flatBgPct=round(fb, 1), verdict=v))
    return out


if __name__ == "__main__":
    root = sys.argv[1]
    scenes = sys.argv[2:]
    rows = scan(root, scenes)
    print("| scene | sim | inkOv% | rowCorr | colCorr | flatBg% | verdict |")
    print("|---|---|---|---|---|---|---|")
    for r in rows:
        print("| %s | %.5f%s | %.1f | %+.3f | %+.3f | %.1f | %s |" % (
            r["scene"], r["sim"], " PASS" if r["passed"] else "", r["inkOv"],
            r["rowCorr"], r["colCorr"], r["flatBgPct"], r["verdict"]))
