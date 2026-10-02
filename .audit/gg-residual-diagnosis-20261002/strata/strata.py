#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""task8 C 类分层残差。两 stage，P1 结构性隔离；口径全部从 instrument.json 读入。

stage=controls : 合成阳性/阴性对照
stage=measure  : 先读 controls.json，controls_pass!=true -> SystemExit(2)；结果只写文件
"""

import json
import os
import sys

import numpy as np
from PIL import Image, ImageDraw, ImageFont

ROOT = r"C:/WorkSpace/Golang/arknights_bot-measure-align"
SD = os.path.join(ROOT, "tmp", "residual", "strata")
INST = os.path.join(SD, "instrument.json")
CONTROLS = os.path.join(SD, "controls.json")
RESULT = os.path.join(SD, "strata.json")

CLASS_ORDER = ["flat_bg", "glyph_edge", "image_asset", "sep_decor", "other"]


def jload(p):
    with open(p, encoding="utf-8") as f:
        return json.load(f)


def jdump(p, o):
    with open(p, "w", encoding="utf-8") as f:
        json.dump(o, f, ensure_ascii=False, indent=2)
    with open(p, encoding="utf-8") as f:
        json.load(f)


def rgba(scene, which):
    with Image.open(os.path.join(ROOT, "tmp", "pixel-compare", scene, which + ".png")) as im:
        return np.asarray(im.convert("RGBA")).astype(np.int32)


def lum(a):
    return 0.299 * a[:, :, 0] + 0.587 * a[:, :, 1] + 0.114 * a[:, :, 2]


def _dilate(mask, k):
    """k x k 最大值滤波（布尔），用移位或实现，无依赖。"""
    out = np.zeros_like(mask)
    H, W = mask.shape
    r = k // 2
    for dy in range(-r, r + 1):
        for dx in range(-r, r + 1):
            out |= np.roll(np.roll(mask, dy, axis=0), dx, axis=1)
    return out


def features(L, T_edge, T_flat=2):
    """range3、edge 与 solid9；只用渲染图。"""
    H, W = L.shape
    mx, mn = L.copy(), L.copy()
    for dy in range(3):
        for dx in range(3):
            s = L[dy:dy + H - 2, dx:dx + W - 2]
            mx[1:H - 1, 1:W - 1] = np.maximum(mx[1:H - 1, 1:W - 1], s)
            mn[1:H - 1, 1:W - 1] = np.minimum(mn[1:H - 1, 1:W - 1], s)
    range3 = np.zeros_like(L)
    range3[1:H - 1, 1:W - 1] = mx[1:H - 1, 1:W - 1] - mn[1:H - 1, 1:W - 1]
    edge = range3 >= T_edge
    solid = _dilate(range3 > T_flat, 9).astype(np.float64) / 81.0
    return range3, edge, solid


def stratify(old, new, inst):
    t = inst["thresholds"]
    d = np.abs(old - new).sum(axis=2)
    diff = d > 0
    n_diff = int(diff.sum())                      # P7'': 先取大小，聚合前断言非空
    sys.stderr.write("PRE-AGG n_diff=%d\n" % n_diff)
    assert n_diff > 0, "P7'' violation: 空差异集合上的聚合没有证据资格"
    L = lum(new.astype(np.float64))
    range3, edge, solid = features(L, t["T_edge"], t["T_flat"])
    cls = np.zeros(d.shape, np.int8)               # 0 flat_bg
    cls[~diff] = -1
    sel = diff
    flat = sel & (range3 <= t["T_flat"]);            cls[flat] = 0
    sep = sel & edge & (solid < t["D_sparse"]);        cls[sep] = 3
    img = sel & edge & (solid > t["D_dense"]);         cls[img] = 2
    gly = sel & edge & (solid >= t["D_sparse"]) & (solid <= t["D_dense"]); cls[gly] = 1
    oth = sel & (range3 > t["T_flat"]) & (~edge);   cls[oth] = 4
    counts = {c: int((cls[diff] == i).sum()) for i, c in enumerate(CLASS_ORDER)}
    shares = {c: round(counts[c] / n_diff, 5) for c in CLASS_ORDER}
    dom = max(CLASS_ORDER, key=lambda c: (shares[c], -CLASS_ORDER.index(c)))
    return dict(n_diff_pixels=n_diff, counts=counts, shares=shares,
                dominant=dom, tie=bool(sum(1 for c in CLASS_ORDER if shares[c] == shares[dom]) > 1),
                other_magnitude=dict(
                    p50=int(np.median(d[oth])) if counts["other"] else 0,
                    p90=int(np.percentile(d[oth], 90)) if counts["other"] else 0,
                    max=int(d[oth].max()) if counts["other"] else 0),
                flat_area_share_of_canvas=round(float((range3 <= t["T_flat"]).mean()), 4))


def synth_text(inst, fresh=False):
    """合成：真实字体文本，残差只在字形边缘。fresh=True 用全新实例复验（不复用调参用的那份）。"""
    size = 26 if fresh else 22
    font = ImageFont.truetype(os.path.join(ROOT, "assets", "font", "NotoSansHans-Regular.ttf"), size)
    W, H = 720, 260
    base = Image.new("RGB", (W, H), (40, 42, 43))
    d = ImageDraw.Draw(base)
    if fresh:
        lines = ["DEF 300 RES 42", "生命恢复速度 0.00", "skill SpCost 30 SP", "描述：攻击造成法术伤害"]
    else:
        lines = ["ATK 850 DEF 300 HP 12000", "RES 40 MoveSpeed 0", "ability text line 1", "skill effect row 2"]
    for i, txt in enumerate(lines):
        d.text((24, 24 + i * 52), txt, font=font, fill=(220, 224, 226))
    new = np.asarray(base).astype(np.int32)
    old = new.copy()
    Ln = lum(new.astype(np.float64))
    r3, _, _ = features(Ln, inst["thresholds"]["T_edge"], inst["thresholds"]["T_flat"])
    # v3：扰动全部非平坦像素（字形 + 抗锯齿带），这是「残差全在字形上」的忠实实现
    band = r3 > inst["thresholds"]["T_flat"]
    for c in range(3):
        ch = old[:, :, c]
        ch[band] = np.clip(ch[band] - 1, 0, 255)
        old[:, :, c] = ch
    old4 = np.dstack([old, np.full((H, W, 1), 255, np.int32)])
    new4 = np.dstack([new, np.full((H, W, 1), 255, np.int32)])
    return old4, new4


def synth_flat_bias(inst):
    W, H = 300, 120
    new = np.full((H, W, 3), (37, 39, 40), np.int32)
    old = np.full((H, W, 3), (46, 48, 49), np.int32)
    return (np.dstack([old, np.full((H, W, 1), 255, np.int32)]),
            np.dstack([new, np.full((H, W, 1), 255, np.int32)]))


def synth_texture(inst, fresh=False):
    src = Image.open(os.path.join(ROOT, "assets", "common", "amiya.png")).convert("RGB")
    if fresh:
        w = src.width // 2
        src = src.crop((w // 3, w // 3, w // 3 + w // 2, w // 3 + w // 2))
    a = src.resize((240, 240), Image.LANCZOS)
    b = src.resize((237, 237), Image.LANCZOS).resize((240, 240), Image.NEAREST)
    A = np.asarray(a).astype(np.int32)
    B = np.asarray(b).astype(np.int32)
    return (np.dstack([A, np.full((240, 240, 1), 255, np.int32)]),
            np.dstack([B, np.full((240, 240, 1), 255, np.int32)]))


def main():
    inst = jload(INST)
    stage = sys.argv[1] if len(sys.argv) > 1 else ""

    if stage == "controls":
        r_txt = stratify(*synth_text(inst, fresh=True), inst)
        r_flat = stratify(*synth_flat_bias(inst), inst)
        r_tex = stratify(*synth_texture(inst, fresh=True), inst)
        ctl = [
            dict(control="POS_text_edge_residual", expect="glyph_edge>=0.90",
                 observed=r_txt["shares"]["glyph_edge"], passed=bool(r_txt["shares"]["glyph_edge"] >= 0.90),
                 detail=dict(n_diff=r_txt["n_diff_pixels"], shares=r_txt["shares"])),
            dict(control="NEG_flat_colour_bias", expect="flat_bg>=0.90",
                 observed=r_flat["shares"]["flat_bg"], passed=bool(r_flat["shares"]["flat_bg"] >= 0.90),
                 detail=dict(n_diff=r_flat["n_diff_pixels"], shares=r_flat["shares"])),
            dict(control="NEG_image_resampling", expect="dominant != glyph_edge",
                 observed=r_tex["dominant"], passed=bool(r_tex["dominant"] != "glyph_edge"),
                 detail=dict(n_diff=r_tex["n_diff_pixels"], shares=r_tex["shares"])),
        ]
        passed = all(c["passed"] for c in ctl)
        jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl, controls_pass=passed))
        sys.stderr.write("controls_pass=%s\n" % passed)
        for c in ctl:
            sys.stderr.write("  %-28s expect=%s observed=%s passed=%s\n"
                             % (c["control"], c["expect"], c["observed"], c["passed"]))
        sys.exit(0 if passed else 1)

    if stage == "measure":
        if not os.path.exists(CONTROLS):
            sys.stderr.write("REFUSED: controls.json missing\n")
            sys.exit(2)
        if not jload(CONTROLS)["controls_pass"]:
            sys.stderr.write("REFUSED: controls_pass != true\n")
            sys.exit(2)
        al = jload(os.path.join(ROOT, "tmp", "align", "report.json"))
        sim_of = {r["scene"]: r["current_sim"] for r in al["rows"]}
        rows = []
        for s in inst["scenes"]:
            old, new = rgba(s, "old"), rgba(s, "new")
            r = stratify(old, new, inst)
            r["scene"] = s
            r["current_sim"] = sim_of[s]
            r["produced_by"] = "src/ggrender/pixel_test.go @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47"
            rows.append(r)
            sys.stderr.write("%-13s dominant=%-11s shares=%s\n"
                             % (s, r["dominant"], json.dumps(r["shares"], ensure_ascii=False)))
        dom_count = {c: sum(1 for r in rows if r["dominant"] == c) for c in CLASS_ORDER}
        thr = inst["pre_registered_three_way_choice"]
        top = max(dom_count.items(), key=lambda kv: kv[1])
        consistent = top[1] >= 6 and top[0] in ("glyph_edge", "image_asset")
        choice = ("A_text_rendering" if consistent and top[0] == "glyph_edge" else
                  "B_image_asset" if consistent else "C_dispersed_no_consistent_dominant")
        jdump(RESULT, dict(instrument_provenance=inst["provenance"],
                           dominance_counts=dom_count,
                           dominant_consistency_threshold=6,
                           pre_registered_choice=choice,
                           rows=rows))
        sys.exit(0)

    sys.stderr.write("usage: strata.py {controls|measure}\n")
    sys.exit(3)


if __name__ == "__main__":
    main()
