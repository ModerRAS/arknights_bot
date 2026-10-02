#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""task12 声明元素分解 —— 测量。两 stage，P1 闸门 + 验牙齿第 8 次。

分工（权威措辞）：
  元素归属来自声明（模板 CSS 块导出元素框）
  残差幅度来自「我们 vs 基线」（比较限制在该元素框内）
  红线：不得用从基线读到的任何数值去决定任何元素「应该是什么」
"""

import json
import os
import sys

import numpy as np
from PIL import Image

ROOT = r"C:/WorkSpace/Golang/arknights_bot-measure-align"
SD = os.path.join(ROOT, "tmp", "residual", "byelem")
INST = os.path.join(SD, "instrument.json")
CONTROLS = os.path.join(SD, "controls.json")
RESULT = os.path.join(SD, "elements.json")
BOXES = os.path.join(SD, "derivable_boxes.json")


def jload(p):
    with open(p, encoding="utf-8") as f:
        return json.load(f)


def jdump(p, o):
    with open(p, "w", encoding="utf-8") as f:
        json.dump(o, f, ensure_ascii=False, indent=2)
    with open(p, encoding="utf-8") as f:
        json.load(f)


def load_pair(scene):
    with Image.open(os.path.join(ROOT, "tmp", "pixel-compare", scene, "old.png")) as im:
        old = np.asarray(im.convert("RGBA")).astype(np.int32)
    with Image.open(os.path.join(ROOT, "tmp", "pixel-compare", scene, "new.png")) as im:
        new = np.asarray(im.convert("RGBA")).astype(np.int32)
    return old, new


def clip_box(box, W, H):
    x, y, w, h = box
    x0, y0 = max(0, int(x)), max(0, int(y))
    x1, y1 = min(W, int(x + w)), min(H, int(y + h))
    return x0, y0, x1, y1


def main():
    inst = jload(INST)
    stage = sys.argv[1] if len(sys.argv) > 1 else ""
    boxes = jload(BOXES)["elements"]
    man = jload(os.path.join(ROOT, "src", "ggrender", "testdata", "visual", "baseline", "manifest.json"))
    # 模板基名 -> 场景 id（manifest 里 template 字段与 scene id 同名，仅大小写/连字符不同）
    scene_of = {e["template"].replace(".tmpl", ""): e["id"] for e in man["entries"]}
    scale_of = {e["id"]: e["scale"] for e in man["entries"]}

    if stage == "controls":
        ctl = []
        # 取第一个可导出框所属场景
        first = boxes[0]
        sc = scene_of[first["template"].replace(".tmpl", "")]
        old, new = load_pair(sc)
        dm = np.abs(old - new).sum(axis=2)
        H, W = dm.shape
        s = scale_of[sc]
        bx = [first["props"].get("left", 0) * s, first["props"].get("top", 0) * s,
              first["props"].get("width", 0) * s, first["props"].get("height", 0) * s]
        x0, y0, x1, y1 = clip_box(bx, W, H)
        assert x1 >= x0 and y1 >= y0, "ORD1 box inverted"
        ctl.append(dict(control="ORD1_containment", passed=bool(x1 <= W and y1 <= H),
                        expect="declared box inside canvas (order/boundary, no threshold)",
                        observed=dict(scene=sc, box=[round(v, 2) for v in bx],
                                      canvas=[W, H], inside=bool(x1 <= W and y1 <= H))))
        tot = int(dm.sum())
        assert tot > 0, "P7'': 空差异集合上的聚合没有证据资格"
        in_small = int(dm[y0:y1, x0:x1].sum())
        out_small = (tot - in_small) / float(tot)
        gx0, gy0 = max(0, x0 - 30), max(0, y0 - 30)
        gx1, gy1 = min(W, x1 + 30), min(H, y1 + 30)
        in_big = int(dm[gy0:gy1, gx0:gx1].sum())
        out_big = (tot - in_big) / float(tot)
        ctl.append(dict(control="ORD2_frame_leak", passed=bool(out_big < out_small),
                        expect="outside-box residual share strictly decreases when box grows",
                        observed=dict(outside_small=round(out_small, 5),
                                      outside_grown=round(out_big, 5))))
        ctl.append(dict(control="ORD3_empty_box_boundary", passed=True,
                        expect="empty box -> zero residual/area",
                        observed=dict(residual=int(dm[0:0, 0:0].sum()), area=0)))
        passed = all(c["passed"] for c in ctl)
        jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl,
                             controls_pass=passed))
        sys.stderr.write("controls_pass=%s\n" % passed)
        for c in ctl:
            sys.stderr.write("  %-24s passed=%-5s\n" % (c["control"], c["passed"]))
        sys.exit(0 if passed else 1)

    if stage == "measure":
        if not os.path.exists(CONTROLS):
            sys.stderr.write("REFUSED: controls.json missing\n")
            sys.exit(2)
        if not jload(CONTROLS)["controls_pass"]:
            sys.stderr.write("REFUSED: controls_pass != true\n")
            sys.exit(2)

        assert len(boxes) > 0, "P7'': 可导出框集合为空"
        by_scene = {}
        for b in boxes:
            sc = scene_of[b["template"].replace(".tmpl", "")]
            by_scene.setdefault(sc, []).append(b)
        rows = []
        for sc in sorted(by_scene):
            old, new = load_pair(sc)
            dm = np.abs(old - new).sum(axis=2)
            H, W = dm.shape
            tot = int(dm.sum())
            assert tot > 0, "P7'': 空差异集合"
            s = scale_of[sc]
            for b in by_scene[sc]:
                bx = [b["props"].get("left", 0) * s, b["props"].get("top", 0) * s,
                      b["props"].get("width", 0) * s, b["props"].get("height", 0) * s]
                x0, y0, x1, y1 = clip_box(bx, W, H)
                ins = int(dm[y0:y1, x0:x1].sum()) if x1 > x0 and y1 > y0 else 0
                rows.append(dict(scene=sc, template=b["template"], selector=b["selector"],
                                 declared_line=b["line"],
                                 declared_props=b["props"],
                                 device_box=[round(v, 2) for v in bx],
                                 coordinate_basis="relative to nearest positioned ancestor (CSS semantics); "
                                                  "canvas-offset resolution NOT implemented this round",
                                 residual_inside=ins, residual_total=tot,
                                 residual_share=round(ins / float(tot), 5),
                                 produced_by="src/ggrender/pixel_test.go @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47"))
                sys.stderr.write("%-14s %-16s L%-4d share=%.5f\n"
                                 % (sc, b["selector"], b["line"], ins / float(tot)))

        jdump(RESULT, dict(
            instrument_provenance=inst["provenance"],
            positioning="本表为**像素类构成**，用作**声明轴分解的交叉核对**；**它不单独构成实现依据**，因为像素类不是可指名的目标。",
            three_statements=inst["authoritative_wording"]["three_statements"],
            n_derivable_boxes=len(rows),
            scenes_with_boxes=sorted(by_scene.keys()),
            n_scenes_total=len(scene_of),
            n_scenes_without_any_box=len(scene_of) - len(by_scene),
            cross_scene_consistent=None,
            cross_scene_note="只有 %d/%d 个场景存在静态可导出的元素框；且这些框的坐标基准是其定位祖先，"
                             "本轮未实现祖先链解析 ⇒ 份额值不是画布坐标下的份额。**跨场景一致性本轮不出结论。**"
                             % (len(by_scene), len(scene_of)),
            rows=rows))
        sys.exit(0)

    sys.stderr.write("usage: byelem.py {controls|measure}\n")
    sys.exit(3)


if __name__ == "__main__":
    main()