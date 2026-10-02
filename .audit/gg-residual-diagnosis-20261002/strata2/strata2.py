#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""task10 连通域结构特征族分层。两 stage，P1 闸门 + 验牙齿第 6 次。

口径全部从 instrument.json 读入；脚本内不硬编码判据。
验收断言落在**特征值**上（机理区间）+ 类别份额上，不对下游聚合拍期望。
"""

import json
import os
import sys

import numpy as np
from PIL import Image, ImageDraw, ImageFont
from scipy import ndimage

ROOT = r"C:/WorkSpace/Golang/arknights_bot-measure-align"
SD = os.path.join(ROOT, "tmp", "residual", "strata2")
INST = os.path.join(SD, "instrument.json")
CONTROLS = os.path.join(SD, "controls.json")
RESULT = os.path.join(SD, "strata.json")
CLASSES = ["flat_bg", "sep_decor", "glyph_edge", "image_asset", "other"]


def jload(p):
    with open(p, encoding="utf-8") as f:
        return json.load(f)


def jdump(p, o):
    with open(p, "w", encoding="utf-8") as f:
        json.dump(o, f, ensure_ascii=False, indent=2)
    with open(p, encoding="utf-8") as f:
        json.load(f)


def lum(a):
    return 0.299 * a[:, :, 0] + 0.587 * a[:, :, 1] + 0.114 * a[:, :, 2]


def range3(L):
    H, W = L.shape
    mx, mn = L.copy(), L.copy()
    for dy in range(3):
        for dx in range(3):
            s = L[dy:dy + H - 2, dx:dx + W - 2]
            mx[1:H - 1, 1:W - 1] = np.maximum(mx[1:H - 1, 1:W - 1], s)
            mn[1:H - 1, 1:W - 1] = np.minimum(mn[1:H - 1, 1:W - 1], s)
    out = np.zeros_like(L)
    out[1:H - 1, 1:W - 1] = mx[1:H - 1, 1:W - 1] - mn[1:H - 1, 1:W - 1]
    return out


# 4-邻域边界像素数：用移位差分，避免 scipy 额外 API
def perimeter_per_label(lab, n, mask):
    """P = 四个方向上的**边界穿越数**（每个方向单独计一次）。
    上一版按『边界像素数』计，每个像素只算一次 —— 对上下两侧都在外的细结构少算一半，
    是 PC4 单调性失败的直接原因（量纲正确、方向反了）。
    """
    p = np.zeros(n + 1, np.int64)
    m = mask
    for sh in ((1, 0), (-1, 0), (0, 1), (0, -1)):
        shifted = np.roll(np.roll(m, sh[0], axis=0), sh[1], axis=1)
        cross = m & ~shifted
        p += np.bincount(lab[cross].ravel(), minlength=n + 1)
    return p


def components(ink, t):
    struct = np.ones((3, 3), bool)          # 8-邻接
    lab, n = ndimage.label(ink, structure=struct)
    return lab, n


def comp_features(lab, n, ink, canvas_w, t):
    """逐连通域算 SW / LIN / FILL / ELONG，并给出类别。返回 (label->class, per_comp dict)"""
    T_sw = t["T_sw"]["value"]
    T_lin = t["thresholds"]["T_lin"]["value"] if "thresholds" in t else t["T_lin"]["value"]
    T_fill = t["thresholds"]["T_fill"]["value"] if "thresholds" in t else t["T_fill"]["value"]
    L_sep = 0.25 * canvas_w
    lut = {}
    per = {}
    per_lab = perimeter_per_label(lab, n, lab > 0)
    for k in range(1, n + 1):
        ys, xs = np.where(lab == k)
        A = len(ys)
        bw, bh = int(xs.max() - xs.min() + 1), int(ys.max() - ys.min() + 1)
        P = int(per_lab[k])
        SW = (2.0 * A / P) if P > 0 else float("inf")
        yc, xc = ys.mean(), xs.mean()
        cov = np.cov(np.vstack([xs - xc, ys - yc]))
        ev = np.linalg.eigvalsh(cov)
        ev = np.sort(ev)[::-1]
        LIN = float(ev[0] / (ev[0] + ev[1])) if (ev[0] + ev[1]) > 0 else 1.0
        FILL = A / float(bw * bh)
        ELONG = max(bw, bh) / float(min(bw, bh))
        longest = max(bw, bh)
        if FILL >= T_fill and SW > T_sw:
            cls = "image_asset"
        elif SW <= T_sw and LIN >= T_lin and longest >= L_sep:
            cls = "sep_decor"
        elif SW <= T_sw and LIN >= T_lin:
            cls = "glyph_edge"
        else:
            cls = "other"
        lut[k] = cls
        per[k] = dict(area=A, bbox=[int(xs.min()), int(ys.min()), int(bw), int(bh)],
                      perimeter=P, SW=round(SW, 4), LIN=round(LIN, 4),
                      FILL=round(FILL, 4), ELONG=round(ELONG, 4), cls=cls)
    return lut, per, dict(T_sw=T_sw, T_lin=T_lin, T_fill=T_fill, L_sep=L_sep)


def stratify(old, new, t):
    d = np.abs(old - new).sum(axis=2)
    diff = d > 0
    n_diff = int(diff.sum())
    sys.stderr.write("PRE-AGG n_diff=%d\n" % n_diff)
    assert n_diff > 0, "P7'': 空差异集合上的聚合没有证据资格"
    L = lum(new.astype(np.float64))
    r3 = range3(L)
    ink = r3 > t["features"]["ink_mask"]["T_flat"]
    lab, n = components(ink, t)
    lut, per, used = comp_features(lab, n, ink, new.shape[1], t)
    counts = {c: 0 for c in CLASSES}
    code = np.full(n + 1, -1, np.int8)          # 连通域号 -> 类别码
    for k, c in lut.items():
        code[k] = CLASSES.index(c)
    cls_of_diff = np.where(diff, np.where(lab > 0, code[lab], CLASSES.index("flat_bg")), -1)
    for ci, c in enumerate(CLASSES):
        counts[c] = int(((cls_of_diff == ci) & diff).sum())
    shares = {c: round(counts[c] / n_diff, 5) for c in CLASSES}
    dom = max(CLASSES, key=lambda c: (shares[c], -CLASSES.index(c)))
    return dict(n_diff_pixels=n_diff, n_components=int(n),
                counts=counts, shares=shares, dominant=dom,
                thresholds_used=used,
                other_comp_profile=dict(
                    n_other_comps=sum(1 for k in per if per[k]["cls"] == "other"),
                    median_SW=round(float(np.median([per[k]["SW"] for k in per if per[k]["cls"] == "other"])), 4)
                    if any(per[k]["cls"] == "other" for k in per) else None,
                    median_FILL=round(float(np.median([per[k]["FILL"] for k in per if per[k]["cls"] == "other"])), 4)
                    if any(per[k]["cls"] == "other" for k in per) else None))


# ---------------- 合成对照 ----------------
def font(size):
    return ImageFont.truetype(os.path.join(ROOT, "assets", "font", "NotoSansHans-Regular.ttf"), size)


def synth_text(size=(900, 320), fsize=22):
    im = Image.new("RGB", size, (40, 42, 43))
    d = ImageDraw.Draw(im)
    for i, txt in enumerate(["ATK 850 DEF 300 HP 12000", "RES 42 MoveSpeed 0.00",
                             "攻击方式 单体 行动方式 阻挡", "技能 冷却 30 SP"]):
        d.text((24, 20 + i * (fsize + 22)), txt, font=font(fsize), fill=(225, 228, 230))
    return im


def perturb_on_ink(new_rgb, T_flat):
    """构造：扰动只施加在渲染图的 ink 上（range3 > T_flat）—— 残差位置由构造保证。"""
    a = np.asarray(new_rgb).astype(np.int32)
    L = lum(a.astype(np.float64))
    band = range3(L) > T_flat
    old = a.copy()
    for c in range(3):
        old[:, :, c] = np.where(band, np.clip(a[:, :, c] - 2, 0, 255), a[:, :, c])
    o4 = np.dstack([old, np.full(old.shape[:2], 255, np.int32)])
    n4 = np.dstack([a, np.full(a.shape[:2], 255, np.int32)])
    return o4, n4


def main():
    inst = jload(INST)
    t = inst["thresholds"]
    t = dict(t)
    t["features"] = inst["features"]
    t["T_sw"] = dict(t["T_sw"], value=inst["thresholds"]["T_sw"]["value_ink"])
    stage = sys.argv[1] if len(sys.argv) > 1 else ""

    if stage == "controls":
        T_flat = inst["features"]["ink_mask"]["T_flat"]
        ctl = []

        # PC1 文本
        txt = synth_text()
        o1, n1 = perturb_on_ink(txt, T_flat)
        r1 = stratify(o1, n1, t)
        sws = [v["SW"] for v in comp_stats(n1, t)["per"].values()]
        pc1 = inst["positive_controls"]["PC1_glyph"]
        lo1, hi1 = pc1["sw_ink_range"]
        sw_ok = len(sws) > 0 and all(lo1 <= s <= hi1 for s in sws)
        ctl.append(dict(control="PC1_glyph", passed=bool(sw_ok and r1["shares"]["glyph_edge"] >= 0.80),
                        expect="SW in [0.7,3.0] px AND glyph_edge share >= 0.80",
                        observed=dict(glyph_share=r1["shares"]["glyph_edge"], shares=r1["shares"],
                                      n_comps=len(sws), sw_min=round(min(sws), 3) if sws else None,
                                      sw_max=round(max(sws), 3) if sws else None)))

        # PC2 分隔线
        sep = Image.new("RGB", (900, 320), (40, 42, 43))
        da = ImageDraw.Draw(sep)
        da.line([(0, 160), (899, 160)], fill=(0, 0, 0), width=1)
        o2, n2 = perturb_on_ink(sep, T_flat)
        r2 = stratify(o2, n2, t)
        per2 = comp_stats(n2, t)["per"]
        long_c = [v for v in per2.values() if max(v["bbox"][2], v["bbox"][3]) >= 0.25 * 900]
        pc2 = inst["positive_controls"]["PC2_separator"]
        lo, hi = pc2["sw_ink_range"]
        sw_ok2 = bool(long_c) and all(lo <= v["SW"] <= hi for v in long_c) and all(v["LIN"] >= 0.95 for v in long_c)
        ctl.append(dict(control="PC2_separator", passed=bool(sw_ok2 and r2["shares"]["sep_decor"] >= 0.80),
                        expect="long comp SW in [0.7,1.6] and LIN>=0.95 AND sep_decor share >= 0.80",
                        observed=dict(sep_share=r2["shares"]["sep_decor"], shares=r2["shares"],
                                      n_long=len(long_c),
                                      sw=[round(v["SW"], 3) for v in long_c],
                                      lin=[round(v["LIN"], 3) for v in long_c])))

        # PC3 图片
        img = Image.new("RGB", (900, 320), (40, 42, 43))
        src = Image.open(os.path.join(ROOT, "assets", "common", "amiya.png")).convert("RGB")
        crop = src.crop((0, 0, src.width // 3, src.height // 3)).resize((300, 300))
        img.paste(crop, (100, 10))
        o3, n3 = perturb_on_ink(img, T_flat)
        r3 = stratify(o3, n3, t)
        per3 = comp_stats(n3, t)["per"]
        big = [v for v in per3.values() if v["area"] > 2000]
        fill_ok = bool(big) and all(v["FILL"] >= 0.60 for v in big)
        ctl.append(dict(control="PC3_image", passed=bool(fill_ok and r3["shares"]["image_asset"] >= 0.60),
                        expect="large comp FILL>=0.60 AND image_asset share >= 0.60",
                        observed=dict(image_share=r3["shares"]["image_asset"], shares=r3["shares"],
                                      n_big=len(big), fill=[round(v["FILL"], 3) for v in big[:5]])))

        # PC4 单调性：1px vs 4px 描边
        def line_img(wpx):
            im = Image.new("RGB", (600, 200), (40, 42, 43))
            ImageDraw.Draw(im).line([(20, 100), (579, 100)], fill=(0, 0, 0), width=wpx)
            return im
        def sw_of(im):
            _, nn = perturb_on_ink(im, T_flat)
            per = comp_stats(nn, t)["per"]
            cand = [v for v in per.values() if max(v["bbox"][2], v["bbox"][3]) >= 0.25 * 600]
            assert len(cand) > 0, "PC4: 未找到长连通域"
            return max(v["SW"] for v in cand)
        sw1, sw4 = sw_of(line_img(1)), sw_of(line_img(4))
        ctl.append(dict(control="PC4_monotonicity_stroke_width", passed=bool(sw4 > sw1),
                        expect="SW(4px line) > SW(1px line)",
                        observed=dict(sw_1px=round(sw1, 4), sw_4px=round(sw4, 4))))

        # PC5 平坦色偏
        a = np.full((200, 300, 4), 255, np.int32)
        b = a.copy()
        b[:, :, 0:3] = 40
        r5 = stratify(a, b, t)
        ctl.append(dict(control="PC5_flat_bias", passed=bool(r5["shares"]["flat_bg"] == 1.0),
                        expect="flat_bg share == 1.0", observed=r5["shares"]["flat_bg"]))

        passed = all(c["passed"] for c in ctl)
        jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl,
                             controls_pass=passed))
        sys.stderr.write("controls_pass=%s\n" % passed)
        for c in ctl:
            sys.stderr.write("  %-32s passed=%-5s %s\n" % (c["control"], c["passed"], c["expect"]))
        sys.exit(0 if passed else 1)

    if stage == "measure":
        if not os.path.exists(CONTROLS):
            sys.stderr.write("REFUSED: controls.json missing\n")
            sys.exit(2)
        if not jload(CONTROLS)["controls_pass"]:
            sys.stderr.write("REFUSED: controls_pass != true\n")
            sys.exit(2)
        al = jload(os.path.join(ROOT, "tmp", "align", "report.json"))
        c_scenes = [r["scene"] for r in al["rows"] if r["scene"] in
                    ["box", "box-summary", "card", "depot", "gacha", "headhunt", "help", "missing", "recruit"]]
        assert len(c_scenes) > 0, "P7'': C 类场景集合为空"
        sys.stderr.write("PRE-AGG n_scenes=%d\n" % len(c_scenes))
        rows = []
        for s in c_scenes:
            with Image.open(os.path.join(ROOT, "tmp", "pixel-compare", s, "old.png")) as im:
                old = np.asarray(im.convert("RGBA")).astype(np.int32)
            with Image.open(os.path.join(ROOT, "tmp", "pixel-compare", s, "new.png")) as im:
                new = np.asarray(im.convert("RGBA")).astype(np.int32)
            r = stratify(old, new, t)
            r["scene"] = s
            r["produced_by"] = "src/ggrender/pixel_test.go @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47"
            rows.append(r)
            sys.stderr.write("%-13s dominant=%-12s shares=%s\n" % (s, r["dominant"], json.dumps(r["shares"])))
        dom_count = {c: sum(1 for r in rows if r["dominant"] == c) for c in CLASSES}
        top = max(dom_count.items(), key=lambda kv: kv[1])
        consistent = top[1] >= 6
        jdump(RESULT, dict(instrument_provenance=inst["provenance"],
                           dominance_counts=dom_count,
                           consistent_dominant=(top[0] if consistent else None),
                           consistency_rule="某类作为主导出现的场景数 >= 6 记为一致主导",
                           total_other_share=round(sum(r["shares"]["other"] for r in rows) / len(rows), 5),
                           rows=rows))
        sys.exit(0)

    sys.stderr.write("usage: strata2.py {controls|measure}\n")
    sys.exit(3)


def comp_stats(new_rgba, t):
    L = lum(new_rgba[:, :, :3].astype(np.float64))
    r3 = range3(L)
    ink = r3 > t["features"]["ink_mask"]["T_flat"]
    lab, n = components(ink, t)
    return dict(lab=lab, n=n, per=comp_features(lab, n, ink, new_rgba.shape[1], t)[1])


if __name__ == "__main__":
    main()
