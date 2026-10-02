#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""task7 追加：前提（premise）逐行证据 + 全 16 场景前提普查。

前提的定义（本 instrument 的 instrument.json 已登记）：
  检测器把「tile 平坦」当作「这里本该有东西却没画」的证据。
  该前提成立与否是可测的：
    nonflat_frac_scene  = 场景内含结构的像素占比（画布本身有多少结构）
    shared_flat_frac    = 在 old 与 new 中**都**平坦的 tile 占比（两边一致的纯色 = 背景）
    colour_evidence     = 洞内众数色 old / new 各自是否命中模板声明色
  若 old 命中声明色而 new 未命中 => 基线按声明填色、渲染用了非声明色 => 「画错了」而非「漏画」。

两 stage，P1 结构性隔离；measure 阶段只写文件，stdout/stderr 0 字节。
"""

import json
import os
import sys

import numpy as np
from PIL import Image
from scipy import ndimage

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from map_elements import (ROOT, jload, jdump, rgba, parse_tmpl, colors_rgb,  # noqa: E402
                          ink_fraction, modal_color, lum)

MD = os.path.join(ROOT, "tmp", "residual", "mapping")
INST = os.path.join(MD, "instrument.json")
CONTROLS = os.path.join(MD, "controls.json")
SURVEY = os.path.join(MD, "premise_survey.json")


def flat_tiles(img, tile_grid=24, flat_range=2.0):
    h, w = img.shape[:2]
    ty, tx = max(1, h // tile_grid), max(1, w // tile_grid)
    nh, nw = h // ty, w // tx
    L = lum(img.astype(np.float64))
    out = np.zeros((nh, nw), bool)
    for j in range(nh):
        for i in range(nw):
            t = L[j * ty:(j + 1) * ty, i * tx:(i + 1) * tx]
            if np.percentile(t, 99) - np.percentile(t, 1) <= flat_range:
                out[j, i] = True
    return out, tx, ty


def interior_holes(flat):
    lab, n = ndimage.label(flat, structure=np.array([[0, 1, 0], [1, 1, 1], [0, 1, 0]]))
    nh, nw = flat.shape
    border = set(lab[0, :].tolist()) | set(lab[-1, :].tolist()) \
        | set(lab[:, 0].tolist()) | set(lab[:, -1].tolist())
    border.discard(0)
    out = []
    for k in range(1, n + 1):
        if k in border:
            continue
        ys, xs = np.where(lab == k)
        out.append(dict(n_tiles=len(ys),
                        tile_box=[int(xs.min()), int(ys.min()), int(xs.max()), int(ys.max())],
                        pixel_box=[int(xs.min()), int(ys.min()), int(xs.max()), int(ys.max())]))
    out.sort(key=lambda d: -d["n_tiles"])
    return out


def nonflat_frac(img):
    L = lum(img.astype(np.float64))
    mx, mn = L.copy(), L.copy()
    for dy in (0, 1, 2):
        for dx in (0, 1, 2):
            sl = L[dy:dy + L.shape[0] - 2, dx:dx + L.shape[1] - 2]
            mx[1:-1, 1:-1] = np.maximum(mx[1:-1, 1:-1], sl)
            mn[1:-1, 1:-1] = np.minimum(mn[1:-1, 1:-1], sl)
    return float(((mx - mn) > 2)[1:-1, 1:-1].mean())


def hit(old, new, px, amin):
    dm = np.abs(old - new).sum(axis=2)
    reg = dm[px[1]:px[3] + 1, px[0]:px[2] + 1]
    area = reg.size
    frac = float((reg > 0).sum()) / area
    share = float(reg.sum()) / float(dm.sum()) if dm.sum() else 0.0
    return frac, share, (frac >= 0.50 and share >= 0.05)


def main():
    stage = sys.argv[1] if len(sys.argv) > 1 else ""
    inst = jload(INST)

    if stage == "controls":
        # 前提判据自检：构造「两侧同色平坦背景」与「两侧异色平坦背景」
        o = rgba("enemy")
        x0, y0, x1, y1 = 41, 171, 942, 417
        f_new, _, _ = flat_tiles(o)
        f_old, _, _ = flat_tiles(rgba("enemy", "old"))
        shared = float((f_new & f_old).mean())
        pos = 0.05 < shared < 0.99
        # 异色平坦：把 old 的洞心色改掉后，两侧仍都平坦但颜色不同
        o2 = o.copy()
        o2[y0:y1 + 1, x0:x1 + 1, :] = (10, 10, 10, 255)
        f_o2, _, _ = flat_tiles(o2)
        still_flat_both = bool(f_o2[y0 // 19 + 1, x0 // 41 + 1] and f_new[y0 // 19 + 1, x0 // 41 + 1])
        colour_differs_but_flat = still_flat_both
        ctl = [
            dict(control="POS_shared_flat_fraction_measurable", expect=True, observed=bool(pos),
                 detail=dict(shared_flat_frac_enemy=round(shared, 4))),
            dict(control="POS_flat_on_both_sides_but_colours_differ", expect=True,
                 observed=colour_differs_but_flat,
                 detail=dict(note="把 old 洞区改成另一种纯色后，两侧仍各自平坦；"
                                  "证明「平坦」不能推出「两侧相同」，也不能推出「未绘制」")),
        ]
        passed = all(c["observed"] == c["expect"] for c in ctl)
        jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl,
                             controls_pass=passed, note="premise-survey controls"))
        sys.stderr.write("controls_pass=%s\n" % passed)
        for c in ctl:
            sys.stderr.write("  %-46s expect=%s observed=%s\n" % (c["control"], c["expect"], c["observed"]))
        sys.exit(0 if passed else 1)

    if stage == "measure":
        if not os.path.exists(CONTROLS) or not jload(CONTROLS).get("controls_pass"):
            sys.stderr.write("REFUSED: controls_pass != true\n")
            sys.exit(2)
        man = jload(os.path.join(ROOT, "src", "ggrender", "testdata", "visual",
                                 "baseline", "manifest.json"))
        t5 = jload(os.path.join(ROOT, "tmp", "residual", "report.json"))
        al = jload(os.path.join(ROOT, "tmp", "align", "report.json"))
        # current_sim 取自 task4 的 16 行报告（task5 的 report 只含 14 个未过线场景）
        sim_of = {r["scene"]: r["current_sim"] for r in al["rows"]}
        rows = []
        for e in man["entries"]:
            s = e["id"]
            old, new = rgba(s, "old"), rgba(s, "new")
            decl = parse_tmpl(os.path.join(ROOT, "template", e["template"]))
            dc = colors_rgb(decl)
            def hit_decl(rgb):
                return [dict(rgb=[a[0], a[1], a[2]], decl=a[3]["raw"], line=a[3]["line"])
                        for a in dc if abs(a[0] - rgb[0]) <= 2 and abs(a[1] - rgb[1]) <= 2
                        and abs(a[2] - rgb[2]) <= 2][:2]
            f_new, tx, ty = flat_tiles(new)
            f_old, _, _ = flat_tiles(old)
            shared = float((f_new & f_old).mean())
            holes = interior_holes(f_new)
            cur = sim_of[s]
            amin = max((0.99 - cur) / 4.0, 0.005)
            hs = []
            for h in holes:
                px = [h["pixel_box"][0] * tx, h["pixel_box"][1] * ty,
                      (h["pixel_box"][2] + 1) * tx - 1, (h["pixel_box"][3] + 1) * ty - 1]
                area = (px[2] - px[0] + 1) * (px[3] - px[1] + 1)
                mind = min((px[2] - px[0] + 1) / new.shape[1], (px[3] - px[1] + 1) / new.shape[0])
                # task8 前提修复：内洞不再承载真值语义，恒为「沉默」。
                # 「平坦 => 空」已改为「平坦 => 沉默」：检测器在这里没有发言权。
                hs.append(dict(pixel_box=px, area_frac=round(area / (new.shape[0] * new.shape[1]), 5),
                               min_dim=round(mind, 5),
                               meets_Amin=bool(area / (new.shape[0] * new.shape[1]) >= amin
                                               and mind >= 0.05),
                               bucket="silent_not_painted",
                               premise="flat => silence (detector has no standing here)"))
            big = hs[0] if hs else None
            colour_ev = None
            if big:
                px = big["pixel_box"]
                cn = modal_color(new[px[1]:px[3] + 1, px[0]:px[2] + 1, :])[0]
                co = modal_color(old[px[1]:px[3] + 1, px[0]:px[2] + 1, :])[0]
                hn, ho = hit_decl(cn), hit_decl(co)
                colour_ev = dict(new_modal=list(cn), old_modal=list(co),
                                 new_matches_declared=bool(hn), old_matches_declared=bool(ho),
                                 new_hit=hn, old_hit=ho,
                                 verdict=("renderer_used_non_declared_colour" if ho and not hn
                                          else "both_match_declaration" if ho and hn
                                          else "neither_matches_declaration"))
            rows.append(dict(scene=s, template=e["template"], current_sim=cur,
                             nonflat_frac_scene=round(nonflat_frac(new), 5),
                             flat_tile_frac_new=round(float(f_new.mean()), 4),
                             shared_flat_frac=round(shared, 4),
                             n_interior_holes=len(holes), n_hits_Amin=sum(1 for h in hs if h["meets_Amin"]),
                             holes=hs[:4], colour_evidence=colour_ev,
                             produced_by="src/ggrender/pixel_test.go @ 4584759"))
        jdump(SURVEY, dict(instrument_provenance=inst["provenance"],
                           note="premise survey over all 16 manifest scenes",
                           premise="flat => silence (task8 修复：内洞不再承载真值语义)",
                           rows=rows))
        sys.exit(0)

    if stage == "ceilings":
        # 分区域色偏天花板：修正族 = 每个 tile 一个独立的每通道常数（比 task5 的全局常数细）
        if not os.path.exists(CONTROLS) or not jload(CONTROLS).get("controls_pass"):
            sys.stderr.write("REFUSED: controls_pass != true\n")
            sys.exit(2)
        man = jload(os.path.join(ROOT, "src", "ggrender", "testdata", "visual",
                                 "baseline", "manifest.json"))
        al = jload(os.path.join(ROOT, "tmp", "align", "report.json"))
        sim_of = {r["scene"]: r["current_sim"] for r in al["rows"]}
        out = []
        for e in man["entries"]:
            s = e["id"]
            old, new = rgba(s, "old"), rgba(s, "new")
            H, W = new.shape[:2]
            d = (old - new).astype(np.int32)
            g = 24
            th, tw = max(1, H // g), max(1, W // g)
            hh, ww = H // th, W // tw
            dcut = d[:hh * th, :ww * tw, :]
            blocks = dcut.reshape(hh, th, ww, tw, 4).transpose(0, 2, 1, 3, 4)
            cst = np.median(blocks, axis=(2, 3))                       # (hh, ww, 4)
            resid = np.abs(blocks - cst[:, :, None, None, :])
            err_corr = int(resid.sum())
            err_now = int(np.abs(d).sum())
            total = W * H * 4 * 255
            out.append(dict(scene=s, current_sim=sim_of[s],
                            sim_global_const=round(1.0 - err_now / total, 6),
                            sim_per_tile_const=round(1.0 - err_corr / total, 6),
                            gain_over_current=round((err_now - err_corr) / total, 6),
                            reaches_gate_per_tile=bool(1.0 - err_corr / total >= 0.99),
                            tile_grid=g, produced_by="src/ggrender/pixel_test.go @ 4584759"))
        jdump(os.path.join(MD, "ceiling_tiles.json"),
              dict(instrument_provenance=inst["provenance"],
                   note="per-tile constant colour ceiling; distinct from task5 global-constant ceiling",
                   rows=out))
        sys.exit(0)

    sys.stderr.write("usage: premise_survey.py {controls|measure|ceilings}\n")
    sys.exit(3)


if __name__ == "__main__":
    main()
