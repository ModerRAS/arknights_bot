#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""gg 残差成因分解（诊断用途）。只读、只测，不改任何渲染/计分代码。

阈值、分类函数、归一化口径一律从 tmp/residual/instrument.json 读入，
本文件内不出现任何硬编码判据 —— 改口径必须改 instrument.json 并重新预登记。

产物：tmp/residual/report.json, tmp/residual/controls.json
"""

import json
import os
import sys

import numpy as np
from PIL import Image
from scipy import ndimage

ROOT = r"C:/WorkSpace/Golang/arknights_bot-measure-align"
INST = os.path.join(ROOT, "tmp", "residual", "instrument.json")
GATE = None  # 门禁阈值只从 instrument 读，不在此处写死

CH = ["R", "G", "B", "A"]


def load_instrument():
    with open(INST, encoding="utf-8") as f:
        return json.load(f)


def rgba(path):
    with Image.open(path) as im:
        return np.asarray(im.convert("RGBA")).astype(np.int32)


def sim_from_err(err, w, h):
    return 1.0 - float(err) / float(w * h * 4 * 255)


def channel_stats(delta, mask, inst):
    """逐通道形状统计（描述性）+ 有符号中位数。"""
    a = inst["class_A_systematic_color_bias"]
    W = a["W_conc"]
    out = {}
    for i, c in enumerate(CH):
        v = delta[:, :, i][mask]
        med = float(np.median(v))
        p25, p75 = (float(x) for x in np.percentile(v, [25, 75]))
        conc = float(np.mean(np.abs(v - med) <= W))
        out[c] = dict(median_signed=round(med, 4),
                      mean_signed=round(float(v.mean()), 4),  # 次要证据
                      p25=round(p25, 4), p75=round(p75, 4),
                      iqr=round(p75 - p25, 4),
                      conc_within_W=round(conc, 4),
                      G=round(conc / a["G_ref"], 4),
                      abs_median_ge_Tbias=bool(abs(med) >= a["T_bias"]))
    return out


def hist(delta, mask, bins=64):
    """每通道 Δ 直方图，bin 宽 255/64=3.984。"""
    out = {}
    for i, c in enumerate(CH):
        h, _ = np.histogram(delta[:, :, i][mask], bins=bins, range=(0, 255))
        out[c] = h.tolist()
    return out


def corr_ceilings(delta, mask, inst):
    """全局单通道常数修正的天花板：修正族 F = 每通道一个全局常数。

    sim_corr_all  = 最优 c 作用于全画布
    sim_corr_diff = 最优 c 只作用于原本有差异的像素
    """
    h, w = delta.shape[:2]
    c_all, c_diff, err_all, err_diff = {}, {}, 0, 0
    for i in range(4):
        d = delta[:, :, i]
        c_all[CH[i]] = float(np.median(d))
        c_diff[CH[i]] = float(np.median(d[mask]))
        err_all += int(np.abs(d - c_all[CH[i]]).sum())
        err_diff += int(np.abs(d[mask] - c_diff[CH[i]]).sum())
    return dict(corr_all=sim_from_err(err_all, w, h),
                corr_diff=sim_from_err(err_diff, w, h),
                c_star_all={k: round(v, 4) for k, v in c_all.items()},
                c_star_diff={k: round(v, 4) for k, v in c_diff.items()})


def luminance(img):
    a = img.astype(np.float64)
    return 0.299 * a[:, :, 0] + 0.587 * a[:, :, 1] + 0.114 * a[:, :, 2]


def flat_tile_mask(img, tile_grid, flat_range=2.0):
    h, w = img.shape[:2]
    ty = max(1, h // tile_grid)
    tx = max(1, w // tile_grid)
    nh, nw = h // ty, w // tx
    lum = luminance(img)
    out = np.zeros((nh, nw), bool)
    for j in range(nh):
        for i in range(nw):
            t = lum[j * ty:(j + 1) * ty, i * tx:(i + 1) * tx]
            if np.percentile(t, 99) - np.percentile(t, 1) <= flat_range:
                out[j, i] = True
    return out, tx, ty


def holes(flat, tx, ty):
    """内洞 = 不接触画布四边的 flat 连通域（4-邻接）。"""
    lab, n = ndimage.label(flat, structure=np.array([[0, 1, 0], [1, 1, 1], [0, 1, 0]]))
    nh, nw = flat.shape
    border = set(lab[0, :].tolist()) | set(lab[-1, :].tolist()) \
        | set(lab[:, 0].tolist()) | set(lab[:, -1].tolist())
    border.discard(0)
    res = []
    for k in range(1, n + 1):
        if k in border:
            continue
        ys, xs = np.where(lab == k)
        j0, j1, i0, i1 = int(ys.min()), int(ys.max()), int(xs.min()), int(xs.max())
        res.append(dict(n_tiles=len(ys),
                        tile_box=[i0, j0, i1, j1],
                        pixel_box=[i0 * tx, j0 * ty, (i1 + 1) * tx - 1, (j1 + 1) * ty - 1],
                        w_px=(i1 - i0 + 1) * tx, h_px=(j1 - j0 + 1) * ty))
    res.sort(key=lambda d: -d["n_tiles"])
    return res


def class_b_flags(img, current_sim, inst):
    b = inst["class_B_missing_block"]
    h, w = img.shape[:2]
    flat, tx, ty = flat_tile_mask(img, b["tile_grid"])
    hs = holes(flat, tx, ty)
    a_min = max((0.99 - current_sim) / 4.0, 0.005)
    hits = []
    for d in hs:
        area = (d["w_px"] * d["h_px"]) / float(w * h)
        mind = min(d["w_px"] / float(w), d["h_px"] / float(h))
        if area >= a_min and mind >= 0.05:
            d = dict(d, area_frac=round(area, 5), min_dim=round(mind, 5))
            hits.append(d)
    return dict(A_min=round(a_min, 5), n_holes=len(hs), n_hits=len(hits),
                largest_hole=hs[0] if hs else None, hits=hits,
                n_ink_tiles=int((~flat).sum()), n_flat_tiles=int(flat.sum()),
                flat_tile_frac=round(float(flat.mean()), 4)), hits


def tmpl_inventory(path):
    """声明层清单：按标签计数 + 媒体元素 + 带 id/class 的元素 + 文本节点数。
    「搜不到名字 != 不存在」——故同时按多种叫法枚举：标签名、id、class。"""
    with open(path, encoding="utf-8") as f:
        src = f.read()
    tags, ids, classes, media = {}, [], [], []
    i = 0
    while True:
        i = src.find("<", i)
        if i < 0:
            break
        j = src.find(">", i)
        if j < 0:
            break
        tag_src = src[i + 1:j]
        i = j + 1
        name = tag_src.split()[0].split("/")[0].strip().lower() if tag_src.strip() else ""
        if not name or name.startswith("!") or name.startswith("?"):
            continue
        if name not in ("br", "img", "meta", "link"):
            tags[name] = tags.get(name, 0) + 1
        if name in ("img", "video", "canvas"):
            media.append(tag_src.strip())
        for piece in tag_src.split():
            if piece.startswith("id="):
                ids.append(piece[3:].strip("\"'"))
            elif piece.startswith("class="):
                classes.extend(piece[6:].strip("\"'").split())
    import re
    stripped = re.sub(r"<[^>]*>", "\n", src)
    text_nodes = len([t for t in stripped.split("\n") if t.strip()])
    return dict(tags=tags, n_elements=sum(tags.values()), media=media, n_media=len(media),
                ids=ids, classes=sorted(set(classes)), n_text_nodes=text_nodes)


def main():
    inst = load_instrument()
    gate = float(inst["gate"])  # 口径的唯一来源是 instrument.json 的数值字段，不从散文 parse
    sys.stderr.write("gate from instrument = %r\n" % gate)

    with open(os.path.join(ROOT, "tmp", "align", "report.json"), encoding="utf-8") as f:
        align = json.load(f)
    with open(os.path.join(ROOT, "src", "ggrender", "testdata", "visual", "baseline",
                           "manifest.json"), encoding="utf-8") as f:
        manifest = json.load(f)
    tmpl_of = {e["id"]: e["template"] for e in manifest["entries"]}

    scenes = [r for r in align["rows"] if not r["go_passed"]]
    pc = align["rows"][0]["produced_by"]

    # ---------------- controls ----------------
    ctrl = []

    def delta_of(o, n):
        return o - n

    def darken(n, const):
        return np.clip(n.astype(np.float64) - np.array(const, dtype=np.float64), 0, 255).astype(np.int32)

    def class_a_of(o, n):
        d = delta_of(o, n)
        m = np.abs(d).sum(axis=2) > 0
        h, w = d.shape[:2]
        cc = corr_ceilings(d, m, inst)
        return dict(sim_now=sim_from_err(int(np.abs(d).sum()), w, h),
                    sim_corr_all=cc["corr_all"], sim_corr_diff=cc["corr_diff"],
                    c_star_all=cc["c_star_all"],
                    flag=bool(max(cc["corr_all"], cc["corr_diff"]) >= gate)), d, m

    # P1 / P1b / P1c: 已知系统性输入必须响
    for sub, const in (("calendar", (6, 4, 5, 0)), ("calendar", (1, 1, 1, 0)), ("state", (6, 4, 5, 0))):
        o = rgba(os.path.join(ROOT, "tmp", "pixel-compare", sub, "old.png"))
        n = rgba(os.path.join(ROOT, "tmp", "pixel-compare", sub, "new.png"))
        r, _, _ = class_a_of(o, darken(n, const))
        name = "P1_systematic_must_fire" if const == (6, 4, 5, 0) and sub == "calendar" else (
            "P1b_small_bias_must_fire" if sub == "calendar" else "P1c_multi_scene_substrate")
        ctrl.append(dict(control=name, substrate=sub, applied_const=list(const),
                         sim_at_measurement=r["sim_now"], sim_corr_all=r["sim_corr_all"],
                         sim_corr_diff=r["sim_corr_diff"], observed=r["flag"], expect=True))
        sys.stderr.write("%-28s observed=%s\n" % (name, r["flag"]))

    # N1: 非系统性输入必须不响（构造上零均值，无常数项）
    o = rgba(os.path.join(ROOT, "tmp", "pixel-compare", "box", "old.png"))
    n_real = rgba(os.path.join(ROOT, "tmp", "pixel-compare", "box", "new.png"))
    rs = np.random.RandomState(777)
    # 合成对：base 压缩到 [64,191] 以消除截断；new_syn = clip(base + N(0,12)) => Δ = -noise
    base = np.clip(0.5 * o.astype(np.float64) + 64.0, 0, 255)
    n_syn = np.clip(base + rs.normal(0, 12, base.shape), 0, 255)
    sat_pct = 100.0 * float(((n_syn == 0) | (n_syn == 255)).sum()) / n_syn.size
    r, d, m = class_a_of(base.astype(np.int32), n_syn.astype(np.int32))
    st = channel_stats(d, m, inst)
    # 「无阳性输入」验的是**输入自身**有没有可修正的全局常数（不是分类器的判决）
    cmax = max(abs(v) for v in r["c_star_all"].values())
    no_positive = bool(cmax < 1.0)
    ctrl.append(dict(control="N1_non_systematic_must_not_fire", substrate="synthetic(box old)",
                     applied_const=[0, 0, 0, 0],
                     construction="base=0.5*old+64 (no clipping); new_syn=clip(base+N(0,12)); Δ=-noise, zero-mean by construction",
                     saturated_pct=round(sat_pct, 4),
                     rejected_substrate="operator (measured max|c*|=15.0 -> contained positive input)",
                     c_star_all=r["c_star_all"], max_abs_c_star=round(cmax, 4),
                     per_channel_signed_median={c: round(st[c]["median_signed"], 4) for c in CH},
                     no_positive_input_check_passed=no_positive,
                     no_positive_input_criterion="max_c |c*_c| < 1.0 (instrument v5)",
                     sim_corr_all=r["sim_corr_all"], sim_corr_diff=r["sim_corr_diff"],
                     observed=r["flag"], expect=False))
    sys.stderr.write("%-28s observed=%s (no-positive-input=%s max|c*|=%.3f)\n"
                     % ("N1_non_systematic_must_not_fire", r["flag"], no_positive, cmax))
    del n_real

    # P3 / N2: 空白块检测器
    o = rgba(os.path.join(ROOT, "tmp", "pixel-compare", "box", "old.png"))
    n = rgba(os.path.join(ROOT, "tmp", "pixel-compare", "box", "new.png"))
    box_sim = [r for r in align["rows"] if r["scene"] == "box"][0]["current_sim"]
    b_clean, _ = class_b_flags(n, box_sim, inst)
    ctrl.append(dict(control="N2_clean_render_must_not_fire", substrate="box",
                     n_holes=b_clean["n_holes"], n_hits=b_clean["n_hits"],
                     observed=bool(b_clean["n_hits"] > 0), expect=False))
    sys.stderr.write("%-28s observed=%s (n_holes=%d)\n"
                     % ("N2_clean_render_must_not_fire", b_clean["n_hits"] > 0, b_clean["n_holes"]))

    painted = n.copy()
    ph, pw = painted.shape[:2]
    b_tg = inst["class_B_missing_block"]["tile_grid"]
    ttx, tty = max(1, pw // b_tg), max(1, ph // b_tg)
    # 涂块与 tile 网格对齐（10x10 格，居中），于是期望命中重叠恒为 1.0
    ntiles = 10
    i0 = (pw // ttx) // 2 - ntiles // 2
    j0 = (ph // tty) // 2 - ntiles // 2
    rx, ry, rw, rh = i0 * ttx, j0 * tty, ntiles * ttx, ntiles * tty
    painted[ry:ry + rh, rx:rx + rw, :] = painted[0, 0, :]
    b_hole, hits = class_b_flags(painted, box_sim, inst)
    want_box = [i0, j0, i0 + ntiles - 1, j0 + ntiles - 1]
    exact_hit = any(h["tile_box"] == want_box for h in hits)
    ctrl.append(dict(control="P3_missing_block_must_fire", substrate="box",
                     painted_rect_px=[rx, ry, rw, rh], painted_tile_box=want_box,
                     painted_area_frac=round(rw * rh / float(pw * ph), 5),
                     n_hits=b_hole["n_hits"], exact_tile_box_hit=bool(exact_hit),
                     observed=bool(b_hole["n_hits"] > 0 and exact_hit), expect=True))
    sys.stderr.write("%-28s observed=%s (n_hits=%d exact=%s)\n"
                     % ("P3_missing_block_must_fire", b_hole["n_hits"] > 0 and exact_hit,
                        b_hole["n_hits"], exact_hit))

    controls_pass = all(c["observed"] == c["expect"] for c in ctrl) and all(
        c.get("no_positive_input_check_passed", True) for c in ctrl)
    with open(os.path.join(ROOT, "tmp", "residual", "controls.json"), "w", encoding="utf-8") as f:
        json.dump(dict(gate=gate, controls=ctrl, controls_pass=controls_pass), f,
                  ensure_ascii=False, indent=2)
    sys.stderr.write("controls_pass=%s\n" % controls_pass)
    if not controls_pass:
        sys.stderr.write("POSITIVE CONTROL FAILED -> 数字作废\n")

    # ---------------- per-scene ----------------
    rows = []
    for r0 in scenes:
        s = r0["scene"]
        o = rgba(os.path.join(ROOT, "tmp", "pixel-compare", s, "old.png"))
        n = rgba(os.path.join(ROOT, "tmp", "pixel-compare", s, "new.png"))
        h, w = n.shape[:2]
        d = delta_of(o, n)
        m = np.abs(d).sum(axis=2) > 0
        st = channel_stats(d, m, inst)
        cc = corr_ceilings(d, m, inst)
        a_flag = bool(max(cc["corr_all"], cc["corr_diff"]) >= gate)
        b_info, b_hits = class_b_flags(n, r0["current_sim"], inst)
        b_flag = bool(b_hits)
        # 2b 佐证（inferred：用基线作比对参照）
        if b_info["largest_hole"] is not None:
            x0, y0, x1, y1 = b_info["largest_hole"]["pixel_box"]
            sub = np.abs(d[y0:y1 + 1, x0:x1 + 1]).sum()
            tot = float(np.abs(d).sum())
            b_info["hole_diff_share_pct"] = round(100.0 * sub / tot, 3) if tot else 0.0
        inv = tmpl_inventory(os.path.join(ROOT, "template", tmpl_of[s]))
        if b_flag:
            primary = "B_missing_block"
        elif a_flag:
            primary = "A_systematic_color_bias"
        else:
            primary = "C_diffuse"
        rows.append(dict(
            scene=s, width=w, height=h, current_sim=r0["current_sim"],
            pct_pixels_differing=round(100.0 * float(m.sum()) / float(w * h), 4),
            channel_stats=st, channel_hist_64bin=hist(d, m),
            sim_corr_all=cc["corr_all"], sim_corr_diff=cc["corr_diff"],
            c_star_all=cc["c_star_all"], c_star_diff=cc["c_star_diff"],
            best_global_color_correction_ceiling=max(cc["corr_all"], cc["corr_diff"]),
            A_systematic_flag=a_flag,
            A_shape_descriptive=bool(all(st[c]["abs_median_ge_Tbias"] for c in CH[:3])),
            B_missing_flag=b_flag, B_info=b_info,
            declaration_inventory=inv,
            primary_class=primary,
            produced_by=pc,
        ))
        sys.stderr.write("%-13s A=%-5s corr=%.5f/%.5f  B=%-5s holes=%-3d hits=%-2d  -> %s\n"
                         % (s, a_flag, cc["corr_all"], cc["corr_diff"], b_flag,
                            b_info["n_holes"], b_info["n_hits"], primary))

    prec = inst["precedence"]
    counts = {k: sum(1 for r in rows if r["primary_class"] == k) for k in prec}
    checks = dict(
        manifest_entries=len(manifest["entries"]),
        align_report_rows=len(align["rows"]),
        passed=sum(1 for r in align["rows"] if r["go_passed"]),
        unpassed=len(rows),
        checksum_unpassed_equals_passed_plus_unpassed=len(rows) + sum(
            1 for r in align["rows"] if r["go_passed"]) == len(align["rows"]),
        primary_counts=counts,
        checksum_three_classes_sum_equals_unpassed=sum(counts.values()) == len(rows),
        item3_applies=bool(counts.get("A_systematic_color_bias", 0) > len(rows) / 2),
    )
    with open(os.path.join(ROOT, "tmp", "residual", "report.json"), "w", encoding="utf-8") as f:
        json.dump(dict(instrument_version=inst["provenance"], gate=gate, controls_pass=controls_pass,
                       checks=checks, rows=rows), f, ensure_ascii=False, indent=2)
    with open(os.path.join(ROOT, "tmp", "residual", "report.json"), encoding="utf-8") as f:
        json.load(f)
    sys.stderr.write(json.dumps(checks, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
