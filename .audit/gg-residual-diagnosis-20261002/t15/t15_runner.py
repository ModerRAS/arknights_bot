#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""task15 第一部分：编码器差异量化的**闸门化**运行器（P1 结构性隔离）。

stage=controls : 跑敏感性对照（白噪声过两条编码器），写 controls.json
stage=measure  : 先 json.load 读 controls.json；controls_pass!=true 或文件缺失 => SystemExit(2)
                 通过才计算 A/B 两路差值并写 result.json

解码器两侧统一用 PIL（libjpeg）—— 解码器被固定住，两路之差才是**编码器**之差。
"""
import json
import os
import sys

import numpy as np
from PIL import Image

ROOT = r"C:/WorkSpace/Golang/arknights_bot-measure-align"
T = os.path.join(ROOT, "tmp", "t15")
INST = os.path.join(T, "instrument.json")
CONTROLS = os.path.join(T, "controls.json")
RESULT = os.path.join(T, "encoder_result.json")
QS = [70, 75, 80, 85, 90, 95, 100]


def jload(p):
    with open(p, encoding="utf-8") as f:
        return json.load(f)


def jdump(p, o):
    with open(p, "w", encoding="utf-8") as f:
        json.dump(o, f, ensure_ascii=False, indent=2)
    with open(p, encoding="utf-8") as f:
        json.load(f)


def rgba(p):
    with Image.open(p) as im:
        return np.asarray(im.convert("RGBA")).astype(np.int32)


def deficit(ref, jpg):
    a, b = rgba(ref), rgba(jpg)
    if a.shape != b.shape:
        raise ValueError("shape mismatch %s %s %s" % (ref, jpg, a.shape))
    h, w = a.shape[:2]
    return int(np.abs(a - b).sum()), h * w * 4 * 255


def main():
    stage = sys.argv[1] if len(sys.argv) > 1 else ""
    inst = jload(INST)
    assert len(QS) > 0, "P7'': quality 集合为空"
    sys.stderr.write("PRE-AGG n_qualities=%d\n" % len(QS))

    if stage == "controls":
        ref = os.path.join(T, "ctrl", "noise.png")
        rows = {}
        for tag, f in (("chromium", "noise_chrom_q70.jpg"), ("go", "noise_go_q70.jpg")):
            d, tot = deficit(ref, os.path.join(T, "ctrl", f))
            rows[tag] = round(1 - d / tot, 5)
            sys.stderr.write("PRE-AGG n_diff_noise_%s=%d\n" % (tag, d))
        thr = inst["sensitivity"]["threshold"]
        observed = min(rows.values())
        ctl = [dict(control="POS_codec_sensitivity", expect="> %s (both routes)" % thr,
                    observed=rows, passed=bool(observed > thr),
                    note="白噪声过两条编码器；量级由 4:2:0 色度子采样主导")]
        passed = all(c["passed"] for c in ctl)
        jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl,
                             controls_pass=passed))
        sys.stderr.write("controls_pass=%s\n" % passed)
        sys.exit(0 if passed else 1)

    if stage == "measure":
        if not os.path.exists(CONTROLS):
            sys.stderr.write("REFUSED: controls.json missing\n")
            sys.exit(2)
        if not jload(CONTROLS)["controls_pass"]:
            sys.stderr.write("REFUSED: controls_pass != true\n")
            sys.exit(2)
        scenes = jload(os.path.join(T, "routes.json"))
        assert len(scenes) > 0, "P7'': 场景集合为空"
        sys.stderr.write("PRE-AGG n_scenes=%d\n" % len(scenes))
        al = {r["scene"]: r["current_sim"]
              for r in jload(os.path.join(ROOT, "tmp", "align", "report.json"))["rows"]}
        rows = {}
        for s in scenes:
            rec = {"A_chromium": {}, "B_go": {}}
            for q in QS:
                da, tot = deficit(os.path.join(T, "png", s + ".png"),
                                  os.path.join(T, "chrom", "%s_q%d.jpg" % (s, q)))
                db, _ = deficit(os.path.join(ROOT, "tmp", "pixel-compare", s, "new.png"),
                                os.path.join(T, "goj", "%s_q%d.jpg" % (s, q)))
                rec["A_chromium"][q] = round(1 - da / tot, 6)
                rec["B_go"][q] = round(1 - db / tot, 6)
            qa = max(QS, key=lambda q: rec["A_chromium"][q])
            qb = max(QS, key=lambda q: rec["B_go"][q])
            gap = round(0.99 - al[s], 5)
            d = round(rec["A_chromium"][qa] - rec["B_go"][qb], 5)
            rows[s] = dict(A_chromium_best=rec["A_chromium"][qa], A_best_q=qa,
                           B_go_best=rec["B_go"][qb], B_best_q=qb,
                           encoder_diff=d, gate_gap=gap,
                           enc_diff_over_gap=round(d / gap, 4) if gap > 0 else None,
                           by_quality=rec,
                           produced_by="src/ggrender/pixel_test.go @ 4584759")
        jdump(RESULT, dict(instrument_provenance=inst["provenance"],
                           policy_bound_verbatim=inst["policy_bound_verbatim"],
                           decoder_fixed="PIL/libjpeg 两侧相同 => 两路之差是编码器之差",
                           best_quality_unanimous=all(v["A_best_q"] == 100 and v["B_best_q"] == 100
                                                      for v in rows.values()),
                           rows=rows))
        sys.exit(0)

    sys.stderr.write("usage: t15_runner.py {controls|measure}\n")
    sys.exit(3)


if __name__ == "__main__":
    main()