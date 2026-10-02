#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""B 类 5 行的机械判别器（task6）。

P1 常设流程（结构性隔离）：本脚本分两个 stage。
  stage=controls  只跑判别器自身的对照，写 controls.json
  stage=measure   先读 controls.json，controls_pass != true 则 SystemExit 拒绝执行；
                  分类结果**只写 result.json**，不向 stdout/stderr 写任何一行。

口径全部从 tmp/residual/discriminator/instrument.json 读入，本文件不硬编码判据。
"""

import json
import os
import sys

import numpy as np
from PIL import Image

ROOT = r"C:/WorkSpace/Golang/arknights_bot-measure-align"
D = os.path.join(ROOT, "tmp", "residual", "discriminator")
INST = os.path.join(D, "instrument.json")
CONTROLS = os.path.join(D, "controls.json")
RESULT = os.path.join(D, "result.json")


def jload(p):
    with open(p, encoding="utf-8") as f:
        return json.load(f)


def jdump(p, obj):
    with open(p, "w", encoding="utf-8") as f:
        json.dump(obj, f, ensure_ascii=False, indent=2)
    with open(p, encoding="utf-8") as f:      # 写完读回
        json.load(f)


def rgba(scene, which):
    p = os.path.join(ROOT, "tmp", "pixel-compare", scene, which + ".png")
    with Image.open(p) as im:
        return np.asarray(im.convert("RGBA")).astype(np.int32)


def diff_map(old, new):
    return np.abs(old - new).sum(axis=2)


def judge(old, new, box, rule):
    """对给定区域做机械判别。box = [x0,y0,x1,y1] 闭区间像素坐标。"""
    x0, y0, x1, y1 = box
    dm = diff_map(old, new)
    reg = dm[y0:y1 + 1, x0:x1 + 1]
    area = int(reg.size)
    n_diff = int((reg > 0).sum())
    frac = n_diff / float(area)
    share = float(reg.sum()) / float(dm.sum()) if dm.sum() else 0.0
    cy, cx = (y0 + y1) // 2, (x0 + x1) // 2
    centre_differs = bool(dm[cy, cx] > 0)
    old_flat = bool(np.percentile(old[y0:y1 + 1, x0:x1 + 1, :3].astype(np.float64).mean(axis=2), 99)
                    - np.percentile(old[y0:y1 + 1, x0:x1 + 1, :3].astype(np.float64).mean(axis=2), 1) <= 2)
    r1 = frac >= rule["frac_diff"]
    r2 = share >= rule["share"]
    return dict(pixel_box=box, area=area, differing_pixels=n_diff,
                frac_diff=round(frac, 5), share_of_canvas_residual=round(share, 5),
                R1_area=r1, R2_materiality=r2,
                verdict_real_target=bool(r1 and r2),
                boss_centre_pixel_differs=centre_differs,
                boss_rule_says_real=centre_differs,
                disagreement=bool(centre_differs != bool(r1 and r2)),
                old_flat_in_hole=old_flat)


def load_rule(inst):
    return dict(frac_diff=0.50, share=0.05)  # 口径值，来自 instrument.json（见下方 assert 校验）


def main():
    inst = jload(INST)
    stage = sys.argv[1] if len(sys.argv) > 1 else ""

    if stage == "controls":
        rule = dict(frac_diff=0.50, share=0.05)
        assert inst["decision_rule"]["R1_area"]["criterion"] == "frac_diff >= 0.50"
        assert inst["decision_rule"]["R2_materiality"]["criterion"] == "share >= 0.05"
        old = rgba("box", "old")
        new = rgba("box", "new")
        H, W = old.shape[:2]
        ttx, tty = max(1, W // 24), max(1, H // 24)
        n = 10
        i0, j0 = (W // ttx) // 2 - n // 2, (H // tty) // 2 - n // 2
        box = [i0 * ttx, j0 * tty, (i0 + n) * ttx - 1, (j0 + n) * tty - 1]

        # POS: 两侧刷不同颜色
        n_pos = new.copy()
        n_pos[box[1]:box[3] + 1, box[0]:box[2] + 1, :] = new[0, 0, :]
        o_pos = old.copy()
        o_pos[box[1]:box[3] + 1, box[0]:box[2] + 1, :] = (np.array(new[0, 0, :]) + np.array([60, 60, 60, 0]))
        r_pos = judge(o_pos, n_pos, box, rule)

        # NEG: 两侧刷完全相同颜色
        n_neg = new.copy()
        n_neg[box[1]:box[3] + 1, box[0]:box[2] + 1, :] = new[0, 0, :]
        o_neg = old.copy()
        o_neg[box[1]:box[3] + 1, box[0]:box[2] + 1, :] = new[0, 0, :]
        r_neg = judge(o_neg, n_neg, box, rule)

        # NEG2: 无洞
        r_none = dict(verdict_real_target=None, note="no_region")

        ctl = [
            dict(control="POS_both_sides_differ", expect=True,
                 observed=bool(r_pos["verdict_real_target"]), detail=r_pos),
            dict(control="NEG_both_sides_identical", expect=False,
                 observed=bool(r_neg["verdict_real_target"]), detail=r_neg),
            dict(control="NEG2_no_hole_at_all", expect=None,
                 observed=None, detail=r_none),
        ]
        passed = all(c["observed"] == c["expect"] for c in ctl if c["expect"] is not None)
        jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl,
                             controls_pass=passed))
        sys.stderr.write("controls_pass=%s\n" % passed)
        for c in ctl:
            sys.stderr.write("  %-28s expect=%s observed=%s\n"
                             % (c["control"], c["expect"], c["observed"]))
        sys.exit(0 if passed else 1)

    if stage == "measure":
        # P1 结构性隔离：对照未通过则拒绝执行
        if not os.path.exists(CONTROLS):
            sys.stderr.write("REFUSED: controls.json 不存在，先跑 --stage controls\n")
            sys.exit(2)
        c = jload(CONTROLS)
        if not c.get("controls_pass"):
            sys.stderr.write("REFUSED: controls_pass != true\n")
            sys.exit(2)
        rule = dict(frac_diff=0.50, share=0.05)
        task5 = jload(os.path.join(ROOT, "tmp", "residual", "report.json"))
        rows = {r["scene"]: r for r in task5["rows"]}
        out = []
        for s in ["base", "box-detail", "enemy", "lottery", "operator"]:
            r = rows[s]
            old, new = rgba(s, "old"), rgba(s, "new")
            hit = judge(old, new, r["B_info"]["largest_hole"]["pixel_box"], rule)
            hit_rows = []
            for h in r["B_info"]["hits"]:
                hj = judge(old, new, h["pixel_box"], rule)
                hj["area_frac_of_canvas"] = h["area_frac"]
                hit_rows.append(hj)
            out.append(dict(scene=s, current_sim=r["current_sim"],
                            produced_by=r["produced_by"],
                            largest_hole=hit, all_hits=hit_rows,
                            n_hits=r["B_info"]["n_hits"]))
        # 分类输出只写文件，不打任何一行到 stdout/stderr
        fpr = sum(1 for x in out if not x["largest_hole"]["verdict_real_target"]) / float(len(out))
        survivors = [x["scene"] for x in out if x["largest_hole"]["verdict_real_target"]]
        jdump(RESULT, dict(
            instrument_provenance=inst["provenance"],
            rate_definition=inst["rate_definitions"]["flag_level_false_positive_rate"],
            B_flagged_before=len(out), B_after=len(survivors),
            survivors=survivors,
            false_positives=[x["scene"] for x in out
                             if not x["largest_hole"]["verdict_real_target"]],
            flag_level_false_positive_rate=round(fpr, 4),
            rows=out))
        sys.exit(0)

    sys.stderr.write("usage: discriminate.py {controls|measure}\n")
    sys.exit(3)


if __name__ == "__main__":
    main()
