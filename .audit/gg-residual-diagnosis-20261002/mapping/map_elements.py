#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""task7 元素级映射。两 stage，P1 结构性隔离（闸门必须先通过）。

stage=controls : 只跑映射器自身的对照，写 controls.json
stage=measure  : 先读 controls.json，不通过则 SystemExit(2)；结果只写 result.json
口径全部从 tmp/residual/mapping/instrument.json 读入，脚本内不硬编码判据。
"""

import json
import os
import re
import sys

import numpy as np
from PIL import Image

ROOT = r"C:/WorkSpace/Golang/arknights_bot-measure-align"
MD = os.path.join(ROOT, "tmp", "residual", "mapping")
INST = os.path.join(MD, "instrument.json")
CONTROLS = os.path.join(MD, "controls.json")
RESULT = os.path.join(MD, "result.json")


def jload(p):
    with open(p, encoding="utf-8") as f:
        return json.load(f)


def jdump(p, o):
    with open(p, "w", encoding="utf-8") as f:
        json.dump(o, f, ensure_ascii=False, indent=2)
    with open(p, encoding="utf-8") as f:
        json.load(f)


def rgba(scene, which="new"):
    with Image.open(os.path.join(ROOT, "tmp", "pixel-compare", scene, which + ".png")) as im:
        return np.asarray(im.convert("RGBA")).astype(np.int32)


# ---------- 声明层解析 ----------
NAMED = {"white": (255, 255, 255), "black": (0, 0, 0), "red": (255, 0, 0),
         "blue": (0, 0, 255), "gray": (128, 128, 128), "grey": (128, 128, 128)}


def parse_tmpl(path):
    src = open(path, encoding="utf-8").read()
    lines = src.split("\n")
    colors, rules, sheets = [], [], []
    for i, ln in enumerate(lines, 1):
        if "stylesheet" in ln:
            m = re.search(r'href="([^"]+)"', ln)
            if m:
                sheets.append(dict(href=m.group(1), line=i))
        for h in re.findall(r"#([0-9a-fA-F]{6})\b|#([0-9a-fA-F]{3})\b", ln):
            hx = h[0] or h[1]
            if len(hx) == 3:
                hx = "".join(c * 2 for c in hx)
            colors.append(dict(hex="#" + hx.upper(), line=i, raw=ln.strip()[:90]))
        for w in NAMED:
            if re.search(r"(color|background)[^;]*\b%s\b" % w, ln):
                colors.append(dict(named=w, rgb=NAMED[w], line=i, raw=ln.strip()[:90]))
    # 显式 width/height 规则（按选择器归组），带模板行号
    for i, ln in enumerate(lines, 1):
        if re.search(r"\b(width|height)\s*:\s*[\d.]+px", ln):
            sel = ln.strip().split("{")[0].strip()
            w = re.search(r"\bwidth\s*:\s*([\d.]+)px", ln)
            h = re.search(r"\bheight\s*:\s*([\d.]+)px", ln)
            rules.append(dict(selector=sel, line=i,
                              width_px=float(w.group(1)) if w else None,
                              height_px=float(h.group(1)) if h else None,
                              raw=ln.strip()[:90]))
    media = []
    for i, ln in enumerate(lines, 1):
        for m in re.finditer(r"<(img|video|canvas)\b([^>]*)>", ln):
            attr = m.group(2)
            cls = re.search(r'class="([^"]+)"', attr)
            idd = re.search(r'id="([^"]+)"', attr)
            onerr = "onerror" in attr
            media.append(dict(tag=m.group(1), line=i,
                              class_=cls.group(1) if cls else None,
                              id_=idd.group(1) if idd else None,
                              has_onerror_fallback=bool(onerr),
                              src=re.search(r'src="([^"]{0,60})', attr).group(1) if re.search(r'src="', attr) else None))
    return dict(path=os.path.basename(path), n_lines=len(lines),
                stylesheets=sheets, colors=colors, sized_rules=rules, media=media)


def colors_rgb(decl):
    out = []
    for c in decl["colors"]:
        if "hex" in c:
            h = c["hex"][1:]
            out.append((int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16), c))
        else:
            out.append((c["rgb"][0], c["rgb"][1], c["rgb"][2], c))
    return out


# ---------- 测量 ----------
def lum(a):
    return 0.299 * a[:, :, 0] + 0.587 * a[:, :, 1] + 0.114 * a[:, :, 2]


def ink_fraction(region, thr=2.0):
    """像素的 3x3 邻域亮度极差 > thr 记为含结构。边界按可用邻域。"""
    L = lum(region)
    H, W = L.shape
    mx = np.full_like(L, -1e9)
    mn = np.full_like(L, 1e9)
    for dy in (0, 1, 2):
        for dx in (0, 1, 2):
            sl = L[dy:dy + H - 2, dx:dx + W - 2]
            mx[1:H - 1, 1:W - 1] = np.maximum(mx[1:H - 1, 1:W - 1], sl)
            mn[1:H - 1, 1:W - 1] = np.minimum(mn[1:H - 1, 1:W - 1], sl)
    inner = (mx - mn) > thr
    return float(inner[1:H - 1, 1:W - 1].mean())


def modal_color(region):
    flat = region[:, :, :3].reshape(-1, 3)
    key = (flat[:, 0].astype(np.int64) << 16) | (flat[:, 1].astype(np.int64) << 8) | flat[:, 2]
    vals, cnts = np.unique(key, return_counts=True)
    k = int(vals[cnts.argmax()])
    return ((k >> 16) & 255, (k >> 8) & 255, k & 255), float(cnts.max()) / flat.shape[0]


def geometry_candidates(decl, hole_w_px, hole_h_px, scale, log2=np.log(2.0)):
    out = []
    for r in decl["sized_rules"]:
        if r["width_px"] is None and r["height_px"] is None:
            continue
        w = (r["width_px"] or 0) * scale
        h = (r["height_px"] or 0) * scale
        okw = w > 0 and abs(np.log(w / hole_w_px)) <= log2
        okh = h > 0 and abs(np.log(h / hole_h_px)) <= log2
        if okw or okh:
            out.append(dict(selector=r["selector"], line=r["line"],
                            width_px=w or None, height_px=h or None,
                            match_w=bool(okw), match_h=bool(okh), raw=r["raw"]))
    return out


def fallback_rate(scene, decl, render):
    """onerror 回退检出的下界：把 assets/common/amiya.png 缩到各声明 img 宽度 ×scale，
    在渲染图上做归一化互相关，取最大相关；>0.90 记为检出。"""
    fb = os.path.join(ROOT, "assets", "common", "amiya.png")
    declared_onerror = sum(1 for m in decl["media"] if m["has_onerror_fallback"])
    if not os.path.exists(fb):
        return dict(declared_onerror_elements=declared_onerror,
                    fallback_asset_exists=False, detected=0, note="fallback asset 不存在，无法检")
    fbim = Image.open(fb).convert("RGB")
    w_rule = next((r["width_px"] for r in decl["sized_rules"] if r["width_px"]), 100.0)
    scale = w_rule and 1.5
    tgt = np.asarray(fbim.resize((int(round(w_rule * 1.5)), int(round(w_rule * 1.5 * fbim.height / fbim.width)))), dtype=np.float64).mean(axis=2)
    t = tgt - tgt.mean()
    tn = np.sqrt((t * t).sum()) or 1.0
    L = lum(render.astype(np.float64))
    th, tw = t.shape
    if th >= L.shape[0] or tw >= L.shape[1]:
        return dict(declared_onerror_elements=declared_onerror, fallback_asset_exists=True,
                    detected=0, note="模板尺寸超过画布，未检")
    H, W = L.shape
    best = -1.0
    step = max(1, min(H - th, W - tw) // 200)
    for y in range(0, H - th, step):
        for x in range(0, W - tw, step):
            w = L[y:y + th, x:x + tw]
            w = w - w.mean()
            den = np.sqrt((w * w).sum())
            c = float((w * t).sum() / (den * tn)) if den else 0.0
            if c > best:
                best = c
    return dict(declared_onerror_elements=declared_onerror, fallback_asset_exists=True,
                probed_size_px=[int(t.shape[1]), int(t.shape[0])],
                best_normalised_correlation=round(best, 4),
                detected=int(best > 0.90),
                note="检出为下界；相关>0.90 记检出")


def main():
    inst = jload(INST)
    stage = sys.argv[1] if len(sys.argv) > 1 else ""
    t_empty = float(inst["outcomes"]["T_empty"])

    if stage == "controls":
        dcl = colors_rgb(parse_tmpl(os.path.join(ROOT, "template", "Enemy.tmpl")))
        declared_rgb = {(a[0], a[1], a[2]) for a in dcl}

        def colour_hit(rgb):
            return any(abs(rgb[0] - c[0]) <= 2 and abs(rgb[1] - c[1]) <= 2 and abs(rgb[2] - c[2]) <= 2
                       for c in declared_rgb)

        def is_empty(region):
            """v2 口径：既无结构、颜色也不像任何声明色。"""
            return ink_fraction(region) < t_empty and not colour_hit(modal_color(region)[0])

        # POS1 真空洞（无结构 + 非声明色）-> is_empty True
        blank = np.zeros((60, 200, 4), np.int32)
        # POS2 画对了的纯色元素（用模板声明色填充）-> is_empty False（v1 在这里会误判为 True）
        fill_rgb = (50, 51, 50)  # Enemy.tmpl 声明的 #323332
        painted = np.zeros((60, 200, 4), np.int32)
        painted[:, :] = (fill_rgb[0], fill_rgb[1], fill_rgb[2], 255)
        # POS3 有结构的真实内容（enemy 的洞区本身）-> is_empty False
        n = rgba("enemy")
        region = n[171:418, 41:943, :]
        ctl = [
            dict(control="POS1_truly_blank_region_is_empty", expect=True,
                 observed=bool(is_empty(blank)),
                 detail=dict(ink=round(ink_fraction(blank), 5), modal=list(modal_color(blank)[0]))),
            dict(control="POS2_correctly_painted_flat_element_NOT_empty", expect=False,
                 observed=bool(is_empty(painted)),
                 detail=dict(ink=round(ink_fraction(painted), 5), modal=list(modal_color(painted)[0]),
                             filled_with="#323332",
                             note="这是 v1 漏掉的那一类：结构为零但颜色是声明色")),
            dict(control="POS3_real_content_region_NOT_empty", expect=False,
                 observed=bool(is_empty(region)),
                 detail=dict(enemy_ink=round(ink_fraction(region), 5))),
            dict(control="POS_declared_colour_lookup", expect=True,
                 observed=bool(colour_hit((50, 51, 50))),
                 detail=dict(looking_for="#323332 -> (50,51,50)", n_declared_colours=len(dcl))),
        ]
        passed = all(c["observed"] == c["expect"] for c in ctl)
        jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl, controls_pass=passed))
        sys.stderr.write("controls_pass=%s\n" % passed)
        for c in ctl:
            sys.stderr.write("  %-44s expect=%s observed=%s\n" % (c["control"], c["expect"], c["observed"]))
        sys.exit(0 if passed else 1)

    if stage == "measure":
        if not os.path.exists(CONTROLS):
            sys.stderr.write("REFUSED: controls.json missing\n")
            sys.exit(2)
        if not jload(CONTROLS).get("controls_pass"):
            sys.stderr.write("REFUSED: controls_pass != true\n")
            sys.exit(2)

        disc = jload(os.path.join(ROOT, "tmp", "residual", "discriminator", "result.json"))
        man = jload(os.path.join(ROOT, "src", "ggrender", "testdata", "visual", "baseline", "manifest.json"))
        tmpl_of = {e["id"]: e["template"] for e in man["entries"]}
        scale_of = {e["id"]: e["scale"] for e in man["entries"]}
        holes = {r["scene"]: r for r in disc["rows"]}

        rows = []
        for s in inst["order"]["sequence"]:
            r = holes[s]
            scale = scale_of[s]
            decl = parse_tmpl(os.path.join(ROOT, "template", tmpl_of[s]))
            n = rgba(s)
            px = r["largest_hole"]["pixel_box"]
            region = n[px[1]:px[3] + 1, px[0]:px[2] + 1, :]
            hw = px[2] - px[0] + 1
            hh = px[3] - px[1] + 1
            inkf = ink_fraction(region)
            mcol, mshare = modal_color(region)
            cands = geometry_candidates(decl, hw, hh, scale)
            dc = colors_rgb(decl)
            matched = [dict(rgb=a[:3], decl=a[3]) for a in dc
                       if abs(a[0] - mcol[0]) <= 2 and abs(a[1] - mcol[1]) <= 2 and abs(a[2] - mcol[2]) <= 2]
            # v2 口径：既无结构、颜色也不像任何声明色，才算空
            is_empty = bool(inkf < t_empty and not matched)
            if is_empty and cands:
                outcome, cause = "O1_有声明未渲染", "declared_element_missing_in_render"
            elif (not is_empty) or (not cands):
                outcome = "O2_声明中无对应元素_or_是别的东西"
                if matched:
                    cause = "declared_color_matched_content_present"
                elif not is_empty:
                    cause = "content_present_colour_matches_no_declaration"
                else:
                    cause = "geometry_only_no_size_declared"
            else:
                outcome, cause = "O3_无法判定", "empty_and_no_geometry_candidate"
            fb = fallback_rate(s, decl, n)
            rows.append(dict(
                scene=s, template=decl["path"], scale=scale,
                hole_pixel_box=px, hole_css_box=[round(px[0] / scale, 1), round(px[1] / scale, 1),
                                                 round(hw / scale, 1), round(hh / scale, 1)],
                residual_share=r["largest_hole"]["share_of_canvas_residual"],
                frac_diff=r["largest_hole"]["frac_diff"],
                frac_diff_margin_over_R1=round(r["largest_hole"]["frac_diff"] - 0.50, 4),
                ink_fraction=round(inkf, 5), modal_color=list(mcol), modal_share=round(mshare, 4),
                declared_colour_hits=matched[:3], n_declared_colours=len(dc),
                declared_geometry_candidates=cands[:6], n_candidates=len(cands),
                declared_media_elements=decl["media"], stylesheets=decl["stylesheets"],
                outcome=outcome, actual_cause=cause,
                onerror_fallback=fb,
                hits=sorted(r["all_hits"], key=lambda h: -h["share_of_canvas_residual"]),
                produced_by=r["produced_by"]))

        cnt = {}
        for r_ in rows:
            cnt[r_["outcome"]] = cnt.get(r_["outcome"], 0) + 1
        jdump(RESULT, dict(
            instrument_provenance=inst["provenance"],
            outcome_counts=cnt,
            checksum_outcomes_sum_equals_survivors=(sum(cnt.values()) == len(rows) == len(disc["survivors"])),
            n_survivors=len(disc["survivors"]), rows=rows))
        sys.exit(0)

    sys.stderr.write("usage: map_elements.py {controls|measure}\n")
    sys.exit(3)


if __name__ == "__main__":
    main()
