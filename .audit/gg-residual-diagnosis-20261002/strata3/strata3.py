#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""task11 颜色区域生长基质 + 分层。两 stage，P1 闸门 + 验牙齿第 7 次。

**PC0 基质验证是前置项**：先证明新基质能区分 实心块 / 空心壳 / 细线，
再谈它在上面的区分能力 —— 在任何分层阈值被使用之前完成。

口径全部从 instrument.json 读入；脚本不硬编码判据。
"""

import json
import os
import sys

import numpy as np
from PIL import Image, ImageDraw, ImageFont
from scipy import ndimage
from scipy.sparse import coo_matrix
from scipy.sparse.csgraph import connected_components

ROOT = r"C:/WorkSpace/Golang/arknights_bot-measure-align"
SD = os.path.join(ROOT, "tmp", "residual", "strata3")
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


# ---------- 基质：颜色区域生长 ----------
def quantise(rgb, q):
    return (rgb // q).astype(np.int32)


def grow_equal(key):
    """**真正的**颜色区域生长：同值 + 4-邻接的连通分量（scipy.sparse.csgraph）。
    为什么不用 ndimage.label：它只对**非零元素**做连通标记，**不按取值相等分割** ——
    用错导致基质退化成「非纯黑掩膜」，PC0 实测整个画布变成 1 个区域。"""
    H, W = key.shape
    idx = np.arange(H * W, dtype=np.int64).reshape(H, W)
    rows, cols = [], []
    m = key[:, :-1] == key[:, 1:]
    if m.any():
        rows.append(idx[:, :-1][m]); cols.append(idx[:, 1:][m])
    m2 = key[:-1, :] == key[1:, :]
    if m2.any():
        rows.append(idx[:-1, :][m2]); cols.append(idx[1:, :][m2])
    if not rows:
        return idx.reshape(H, W) + 1, int(H * W)
    r = np.concatenate(rows); c = np.concatenate(cols)
    g = coo_matrix((np.ones(len(r), np.int8), (r, c)), shape=(H * W, H * W))
    n, lab = connected_components(g, directed=False)
    return lab.reshape(H, W) + 1, int(n)


def regions(rgb, q):
    """颜色区域生长：量化后同格 + 4-邻接连通。返回 (lab, n, cellkey)"""
    cells = quantise(rgb, q)
    key = (cells[:, :, 0] * 4096 + cells[:, :, 1] * 64 + cells[:, :, 2]).astype(np.int64)
    lab, n = grow_equal(key)
    return lab, n, key


def per_label_boundary(lab, n):
    """四方向**跨区域**边界穿越数（每方向独立计一次）。
    必须是『跨到另一个区域』而不是『跨到 0』—— 后者对内部区域给 P=0，
    导致 SW=inf（PC0 实测所有区域 SW 均为 inf）。
    """
    p = np.zeros(n + 1, np.int64)
    for sh in ((1, 0), (-1, 0), (0, 1), (0, -1)):
        shifted = np.roll(np.roll(lab, sh[0], axis=0), sh[1], axis=1)
        cross = lab != shifted                     # 标号发生变化 = 边界
        m = cross & (lab > 0)
        p += np.bincount(lab[m].ravel(), minlength=n + 1)
    return p


def cc_map(key, tile=32):
    """CC = tile x tile 窗口内不同量化色格数（纹理签名）。"""
    H, W = key.shape
    th, tw = max(1, H // tile), max(1, W // tile)
    out = np.zeros((H // th, W // tw), np.int32)
    for j in range(out.shape[0]):
        for i in range(out.shape[1]):
            sub = key[j * th:(j + 1) * th, i * tw:(i + 1) * tw]
            out[j, i] = len(np.unique(sub))
    return out, th, tw


def analyse(rgb, t):
    q = t["substrate"]["Q_step"]["value"]
    lab, n, key = regions(rgb, q)
    per = per_label_boundary(lab, n)
    ccm, cth, ctw = cc_map(key)
    bg_label = int(lab[0, 0])                       # 画布角落所在区域 = 背景区域
    feats = {}
    for k in range(1, n + 1):
        ys, xs = np.where(lab == k)
        A = int(len(ys))
        if A == 0:
            continue
        bw, bh = int(xs.max() - xs.min() + 1), int(ys.max() - ys.min() + 1)
        P = int(per[k])
        SW = 2.0 * A / P if P > 0 else float("inf")
        yc, xc = float(ys.mean()), float(xs.mean())
        if A > 2:
            cov = np.cov(np.vstack([xs - xc, ys - yc]))
            ev = np.sort(np.linalg.eigvalsh(cov))[::-1]
            LIN = float(ev[0] / (ev[0] + ev[1])) if (ev[0] + ev[1]) > 0 else 1.0
        else:
            LIN = 1.0
        FILL = A / float(bw * bh)
        cc = int(ccm[ys.min() // cth, xs.min() // ctw])
        feats[k] = dict(area=A, bbox=[int(xs.min()), int(ys.min()), bw, bh],
                        SW=round(SW, 4), LIN=round(LIN, 4), FILL=round(FILL, 4),
                        CC=cc, longest=max(bw, bh), bg=(k == bg_label))
    return lab, n, feats, bg_label


def classify_feats(feats, t):
    T_sw = t["T_sw"]["value"]
    T_lin = t["T_lin"]["value"]
    C_img = t["C_img"]["value"]
    L_sep_frac = 0.25
    cls = {}
    for k, f in feats.items():
        if f["bg"]:
            cls[k] = "flat_bg"
        elif f["CC"] >= C_img:
            cls[k] = "image_asset"
        elif f["SW"] <= T_sw and f["LIN"] >= T_lin and f["longest"] >= f["_L_sep"]:
            cls[k] = "sep_decor"
        elif f["SW"] <= T_sw and f["LIN"] >= T_lin:
            cls[k] = "glyph_edge"
        else:
            cls[k] = "other"
    return cls


def stratify(old, new, t):
    d = np.abs(old - new).sum(axis=2)
    diff = d > 0
    n_diff = int(diff.sum())
    sys.stderr.write("PRE-AGG n_diff=%d\n" % n_diff)
    assert n_diff > 0, "P7'': 空差异集合上的聚合没有证据资格"
    lab, n, feats, bg = analyse(new[:, :, :3], t)
    L_sep = int(0.25 * new.shape[1])
    for k in feats:
        feats[k]["_L_sep"] = L_sep
    cls = classify_feats(feats, t)
    counts = {c: 0 for c in CLASSES}
    for k, c in cls.items():
        counts[c] += int(((lab == k) & diff).sum())
    shares = {c: round(counts[c] / n_diff, 5) for c in CLASSES}
    dom = max(CLASSES, key=lambda c: (shares[c], -CLASSES.index(c)))
    others = [f for k, f in feats.items() if cls[k] == "other"]
    return dict(n_diff_pixels=n_diff, n_regions=int(n), bg_region=int(bg),
                counts=counts, shares=shares, dominant=dom,
                L_sep=L_sep,
                other_profile=dict(n_other_regions=len(others),
                                   median_SW=round(float(np.median([f["SW"] for f in others])), 3) if others else None,
                                   median_FILL=round(float(np.median([f["FILL"] for f in others])), 3) if others else None,
                                   median_CC=round(float(np.median([f["CC"] for f in others])), 1) if others else None),
                feats={str(k): {kk: vv for kk, vv in f.items() if kk != "_L_sep"} for k, f in list(feats.items())[:40]},
                region_cls={str(k): v for k, v in cls.items()})


# ---------- 合成 ----------
BG = (40, 42, 43)


def font(sz):
    return ImageFont.truetype(os.path.join(ROOT, "assets", "font", "NotoSansHans-Regular.ttf"), sz)


def make_text(W=900, H=320, fsz=22):
    im = Image.new("RGB", (W, H), BG)
    d = ImageDraw.Draw(im)
    for i, txt in enumerate(["ATK 850 DEF 300 HP 12000", "RES 42 MoveSpeed 0.00",
                             "攻击方式 单体 行动方式 阻挡", "技能 冷却 30 SP"]):
        d.text((24, 20 + i * (fsz + 22)), txt, font=font(fsz), fill=(225, 228, 230))
    return im


def perturb_on_regions(rgb, t, labels):
    """构造：只在给定区域上施加扰动（模拟残差落那些元素上）。"""
    lab, n, key = regions(rgb[:, :, :3], t["substrate"]["Q_step"]["value"])
    a = rgb.astype(np.int32)
    mask = np.isin(lab, labels)
    old = a.copy()
    for c in range(3):
        old[:, :, c] = np.where(mask, np.clip(a[:, :, c] + 6, 0, 255), a[:, :, c])
    return (np.dstack([old, np.full(a.shape[:2], 255, np.int32)]),
            np.dstack([a, np.full(a.shape[:2], 255, np.int32)]))


def main():
    inst = jload(INST)
    t = dict(inst["thresholds"])
    t["substrate"] = inst["substrate"]
    stage = sys.argv[1] if len(sys.argv) > 1 else ""

    if stage == "controls":
        ctl = []
        Q = t["substrate"]["Q_step"]["value"]

        # ============ PC0 基质验证（前置项，不使用任何分层阈值）============
        W, H = 400, 300
        solid = Image.new("RGB", (W, H), BG)
        ImageDraw.Draw(solid).rectangle([100, 100, 139, 119], fill=(0, 0, 0))   # 40x20
        thin = Image.new("RGB", (W, H), BG)
        ImageDraw.Draw(thin).line([(20, 150), (379, 150)], fill=(0, 0, 0), width=1)
        _, _, f_solid, _ = analyse(np.asarray(solid), t)
        _, _, f_thin, _ = analyse(np.asarray(thin), t)
        # 找到非背景区域
        s_reg = [f for f in f_solid.values() if not f["bg"] and f["area"] >= 100]
        t_reg = [f for f in f_thin.values() if not f["bg"] and f["area"] >= 100]
        pc0_checks = []
        if s_reg:
            big = max(s_reg, key=lambda f: f["area"])
            cov = big["area"] / (40 * 20.0)
            pc0_checks.append(dict(name="solid_region_is_whole_block", observed=round(cov, 4),
                                   expect=">=0.95", ok=bool(cov >= 0.95)))
            pc0_checks.append(dict(name="solid_SW_ge_8", observed=big["SW"], expect=">=8",
                                   ok=bool(big["SW"] >= 8)))
        else:
            pc0_checks.append(dict(name="solid_region_found", observed=None, expect=">=100px", ok=False))
        if t_reg:
            ln = max(t_reg, key=lambda f: f["longest"])
            pc0_checks.append(dict(name="thin_SW_le_3", observed=ln["SW"], expect="<=3",
                                   ok=bool(ln["SW"] <= 3)))
            pc0_checks.append(dict(name="solid_SW_gt_thin_SW", observed=[big["SW"], ln["SW"]] if s_reg else None,
                                   expect="solid > thin", ok=bool(s_reg and big["SW"] > ln["SW"])))
        else:
            pc0_checks.append(dict(name="thin_region_found", observed=None, expect=">=100px", ok=False))
        pc0_ok = all(c["ok"] for c in pc0_checks)
        ctl.append(dict(control="PC0_substrate_validation_PRECONDITION", passed=pc0_ok,
                        expect="新基质能区分 实心块 / 细线（在任何分层阈值之前）",
                        observed=dict(checks=pc0_checks,
                                      note="solid 40x20 与 1px 细线在同一底色上")))
        sys.stderr.write("PC0 substrate validation passed=%s\n" % pc0_ok)
        if not pc0_ok:
            jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl,
                                 controls_pass=False))
            for c in ctl:
                sys.stderr.write("  %-42s passed=%s\n" % (c["control"], c["passed"]))
            sys.exit(1)

        # ============ 以下才允许使用分层阈值 ============
        # PC5 回归护栏：平坦色偏
        a = np.full((200, 300, 4), 255, np.int32)
        b = a.copy()
        b[:, :, 0:3] = 40
        r5 = stratify(a, b, t)
        ctl.append(dict(control="PC5_flat_bias_REGRESSION", passed=bool(r5["shares"]["flat_bg"] == 1.0),
                        expect="flat_bg share == 1.0 (task10: 1.0)", observed=r5["shares"]["flat_bg"]))

        # PC2 回归护栏：分隔线
        sep = Image.new("RGB", (900, 320), BG)
        ImageDraw.Draw(sep).line([(0, 160), (899, 160)], fill=(0, 0, 0), width=1)
        sa = np.asarray(sep)
        lab2, n2, _ = regions(sa, Q)
        # 只扰动**分隔线所在区域**（instrument 声明的构造：残差落在该元素上）。
        # 上一版扰动了所有非背景区域，而那条线把画布上下切成两块背景，
        # 下半块也被扰动 ⇒ 0.99375 的残差落在非分隔线的背景上（属于构造偏差，不是基质失败）。
        line_row = 160
        line_label = int(lab2[line_row, 450])
        labels2 = [line_label]
        o2, n2i = perturb_on_regions(sa, t, labels2)
        r2 = stratify(o2, n2i, t)
        ctl.append(dict(control="PC2_separator_REGRESSION",
                        passed=bool(r2["shares"]["sep_decor"] >= 0.80),
                        expect="sep_decor share >= 0.80 (task10: 1.0)", observed=r2["shares"]["sep_decor"]))

        # PC4 实心 vs 细线（task10 反着失败那条的正向版）
        blk = Image.new("RGB", (400, 300), BG)
        ImageDraw.Draw(blk).rectangle([100, 100, 139, 119], fill=(0, 0, 0))
        ba = np.asarray(blk)
        labb, nb, _ = regions(ba, Q)
        labelsb = [k for k in range(1, nb + 1) if k != labb[0, 0]]
        ob, nbi = perturb_on_regions(ba, t, labelsb)
        rb = stratify(ob, nbi, t)
        fb = [f for k, f in analyse(ba, t)[2].items() if k in labelsb]
        ft = [f for k, f in analyse(np.asarray(thin), t)[2].items() if not f["bg"] and f["area"] >= 100]
        ok4 = bool(fb and ft and max(f["SW"] for f in fb) > max(f["SW"] for f in ft)
                   and max(f["FILL"] for f in fb) >= 0.95)
        ctl.append(dict(control="PC4_solid_vs_thin",
                        passed=ok4,
                        expect="SW(solid)>SW(thin) and FILL(solid)>=0.95",
                        observed=dict(solid_SW=max([f["SW"] for f in fb]) if fb else None,
                                      thin_SW=max([f["SW"] for f in ft]) if ft else None,
                                      solid_FILL=max([f["FILL"] for f in fb]) if fb else None)))

        # PC1 文本
        ta = np.asarray(make_text())
        labt, nt, _ = regions(ta, Q)
        labels_t = [k for k in range(1, nt + 1) if k != labt[0, 0]]
        ot, nti = perturb_on_regions(ta, t, labels_t)
        r1 = stratify(ot, nti, t)
        stroke_ok = all(v["SW"] <= t["T_sw"]["value"]
                        for v in r1["feats"].values() if v["region_cls"].__class__ is not None
                        and r1["region_cls"][v_key] == "glyph_edge") if False else None
        glyph_feats = [r1["feats"][k] for k in r1["feats"] if r1["region_cls"].get(k) == "glyph_edge"]
        sw_ok1 = all(f["SW"] <= t["T_sw"]["value"] for f in glyph_feats) if glyph_feats else False
        ctl.append(dict(control="PC1_text", passed=bool(sw_ok1 and r1["shares"]["glyph_edge"] >= 0.60),
                        expect="glyph_edge share >= 0.60 and stroke SW <= T_sw",
                        observed=dict(glyph_share=r1["shares"]["glyph_edge"], shares=r1["shares"],
                                      n_glyph_regions=len(glyph_feats))))

        # PC3 图片
        im3 = Image.new("RGB", (900, 320), BG)
        src = Image.open(os.path.join(ROOT, "assets", "common", "amiya.png")).convert("RGB")
        crop = src.crop((0, 0, src.width // 3, src.height // 3)).resize((300, 300))
        im3.paste(crop, (100, 10))
        ia = np.asarray(im3)
        labi, ni, _ = regions(ia, Q)
        # 只扰动图片块区域（排除背景）
        labels_i = [k for k in range(1, ni + 1) if k != labi[0, 0]]
        oi, nii = perturb_on_regions(ia, t, labels_i)
        r3 = stratify(oi, nii, t)
        cc_ok = all(f["CC"] >= t["C_img"]["value"]
                    for f in r3["feats"].values() if r3["region_cls"].get(f.get("_k", "")) == "image_asset") if False else None
        img_feats = [r3["feats"][k] for k in r3["feats"] if r3["region_cls"].get(k) == "image_asset"]
        cc_ok = all(f["CC"] >= t["C_img"]["value"] for f in img_feats) if img_feats else False
        ctl.append(dict(control="PC3_image", passed=bool(cc_ok and r3["shares"]["image_asset"] >= 0.60),
                        expect="image_asset share >= 0.60 and CC >= C_img",
                        observed=dict(image_share=r3["shares"]["image_asset"], shares=r3["shares"],
                                      n_image_regions=len(img_feats),
                                      median_CC=round(float(np.median([f["CC"] for f in img_feats])), 1) if img_feats else None)))

        passed = all(c["passed"] for c in ctl)
        jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl,
                             controls_pass=passed))
        sys.stderr.write("controls_pass=%s\n" % passed)
        for c in ctl:
            sys.stderr.write("  %-42s passed=%-5s expect=%s\n" % (c["control"], c["passed"], c["expect"]))
        sys.exit(0 if passed else 1)

    if stage == "measure":
        if not os.path.exists(CONTROLS):
            sys.stderr.write("REFUSED: controls.json missing\n")
            sys.exit(2)
        if not jload(CONTROLS)["controls_pass"]:
            sys.stderr.write("REFUSED: controls_pass != true\n")
            sys.exit(2)
        al = jload(os.path.join(ROOT, "tmp", "align", "report.json"))
        scenes = ["box", "box-summary", "card", "depot", "gacha", "headhunt",
                  "help", "missing", "recruit"]
        assert len(scenes) > 0, "P7'': 场景集合为空"
        sys.stderr.write("PRE-AGG n_scenes=%d\n" % len(scenes))
        rows = []
        for s in scenes:
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
        jdump(RESULT, dict(
            instrument_provenance=inst["provenance"],
            positioning="本表为**像素类构成**，用作**声明轴分解的交叉核对**；**它不单独构成实现依据**，因为像素类不是可指名的目标。",
            dominance_counts=dom_count,
            consistent_dominant=(top[0] if consistent else None),
            consistency_rule="某类作为主导出现的场景数 >= 6 记为一致主导",
            mean_other_share=round(sum(r["shares"]["other"] for r in rows) / len(rows), 5),
            rows=rows))
        sys.exit(0)

    sys.stderr.write("usage: strata3.py {controls|measure}\n")
    sys.exit(3)


if __name__ == "__main__":
    main()
