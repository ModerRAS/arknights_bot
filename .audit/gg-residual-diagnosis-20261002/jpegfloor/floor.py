#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""task9 JPEG 编解码噪声地板。两 stage，P1 闸门 + 验牙齿第 5 次。

主测量只用**我们自己的渲染**做 encode->decode 往返；基线仅参与「自比」交叉核对，
并在产物里显式标注为仪器噪声地板测量、不是渲染输入。
"""

import json
import os
import subprocess
import sys

import numpy as np
from PIL import Image

ROOT = r"C:/WorkSpace/Golang/arknights_bot-measure-align"
JD = os.path.join(ROOT, "tmp", "residual", "jpegfloor")
INST = os.path.join(JD, "instrument.json")
CONTROLS = os.path.join(JD, "controls.json")
RESULT = os.path.join(JD, "floor.json")
GO = os.path.join(JD, "goenc", "jpegrt.exe")
PRODUCED_BY = "src/ggrender/pixel_test.go @ 4584759e18fb9f4d17ba530b3d2cdf1bfb322c47"


def jload(p):
    with open(p, encoding="utf-8") as f:
        return json.load(f)


def jdump(p, o):
    with open(p, "w", encoding="utf-8") as f:
        json.dump(o, f, ensure_ascii=False, indent=2)
    with open(p, encoding="utf-8") as f:
        json.load(f)


def roundtrip(path, qualities):
    out = subprocess.run([GO, path, ",".join(str(q) for q in qualities)],
                         capture_output=True, text=True)
    if out.returncode != 0:
        sys.stderr.write("goenc failed: %s\n" % out.stderr)
        sys.exit(2)
    d = json.loads(out.stdout)
    return {int(k): v for k, v in d["sim"].items()}


def synth_flat(path, colour=(40, 42, 43), size=(400, 300)):
    a = np.zeros((size[1], size[0], 4), np.uint8)
    a[:, :] = (colour[0], colour[1], colour[2], 255)
    Image.fromarray(a, "RGBA").save(path)
    return path


def synth_checker(path, size=(400, 300), cell=1):
    a = np.zeros((size[1], size[0], 4), np.uint8)
    yy, xx = np.mgrid[0:size[1], 0:size[0]]
    m = (((yy // cell) + (xx // cell)) % 2).astype(bool)
    a[m] = (255, 255, 255, 255)
    a[~m] = (0, 0, 0, 255)
    Image.fromarray(a, "RGBA").save(path)
    return path


def synth_white_noise(path, size=(512, 512), seed=20240902):
    rs = np.random.RandomState(seed)
    g = rs.randint(0, 256, size=(size[1], size[0], 3), dtype=np.uint8)
    a = np.dstack([g, np.full((size[1], size[0], 1), 255, np.uint8)])
    Image.fromarray(a, "RGBA").save(path)
    return path


def synth_block_aligned(path, size=(512, 512), period=8):
    yy, xx = np.mgrid[0:size[1], 0:size[0]]
    m = (((yy // period) + (xx // period)) % 2).astype(bool)
    a = np.zeros((size[1], size[0], 4), np.uint8)
    a[m] = (255, 255, 255, 255)
    a[~m] = (0, 0, 0, 255)
    Image.fromarray(a, "RGBA").save(path)
    return path


def synth_flat_plus_line(path, size=(400, 300)):
    a = np.zeros((size[1], size[0], 4), np.uint8)
    a[:, :] = (255, 255, 255, 255)
    a[size[1] // 2, :, :3] = 0          # 1px 黑色横线
    Image.fromarray(a, "RGBA").save(path)
    return path


def main():
    inst = jload(INST)
    qs = inst["ruler"]["qualities"]
    stage = sys.argv[1] if len(sys.argv) > 1 else ""
    assert len(qs) > 0, "P7'': quality 集合为空"
    sys.stderr.write("PRE-AGG n_qualities=%d\n" % len(qs))

    if stage == "controls":
        os.makedirs(os.path.join(JD, "ctrl"), exist_ok=True)
        f = roundtrip(synth_flat(os.path.join(JD, "ctrl", "flat.png")), qs)
        c = roundtrip(synth_checker(os.path.join(JD, "ctrl", "checker.png")), qs)
        assert len(f) == len(qs), "P7'': flat 对照未返回全部 quality"
        assert len(c) == len(qs), "P7'': checker 对照未返回全部 quality"
        fmin = min(f.values())
        cdef = {q: round(1.0 - v, 6) for q, v in c.items()}
        ordered = list(qs)
        lo3 = sorted(cdef[q] for q in ordered[:3])
        hi3 = sorted(cdef[q] for q in ordered[-3:])
        range_ok = (cdef[ordered[0]] - cdef[ordered[-1]]) > 0.01
        trend_ok = min(lo3) > max(hi3)
        spread = cdef[ordered[0]] - cdef[ordered[-1]]
        ctl = [
            dict(control="POS_flat_no_detail", expect="S_rt(all q) >= 0.999",
                 observed=fmin, passed=bool(fmin >= 0.999), blocking=True,
                 detail=dict(sim_by_q=f, note="纯色图往返不应产生残差；若这里报出残差说明量具坏")),
            dict(control="POS_trend", expect="low3 > high3",
                 observed=dict(low3_min=round(min(lo3), 6), high3_max=round(max(hi3), 6)),
                 passed=bool(trend_ok), blocking=True,
                 detail=dict(deficit_by_q=cdef,
                             note="方向正确性；逐档严格单调不作为条件，因为 JPEG 量化表非嵌套")),
            dict(control="POS_dynamic_range", expect="deficit(q_min)-deficit(q_max) > 0.01",
                 observed=round(spread, 6), passed=bool(range_ok), blocking=False,
                 detail=dict(note="**未满足**，如实记录。见 instrument amendment_v5：1px 棋盘并不被 JPEG 严重破坏（8x8 DCT 保留 DC），我的前提错了。该断言不阻塞本轮结论方向 —— 量具低报亏损只会让真实地板更高。")),
        ]
        al = jload(os.path.join(ROOT, "tmp", "align", "report.json"))
        worst = 0.0
        for r in al["rows"]:
            out = subprocess.run([GO, "cmp",
                                  os.path.join(ROOT, "tmp", "pixel-compare", r["scene"], "old.png"),
                                  os.path.join(ROOT, "tmp", "pixel-compare", r["scene"], "new.png")],
                                 capture_output=True, text=True)
            if out.returncode != 0:
                sys.stderr.write("cmp failed: %s\n" % out.stderr)
                sys.exit(2)
            worst = max(worst, abs(json.loads(out.stdout)["sim"] - r["go_sim"]))
        ctl.append(dict(control="CALIBRATION_against_harness",
                        expect="max|ruler - go_sim| < 1e-12 over 16 scenes",
                        observed=worst, passed=bool(worst < 1e-12), blocking=True,
                        detail=dict(n_scenes=len(al["rows"]),
                                    note="口径同一/可复现；不等于敏感性，二者独立")))
        # 敏感性对照：白噪声（JPEG 最坏情况），期望值来源见 instrument
        w = roundtrip(synth_white_noise(os.path.join(JD, "ctrl", "noise.png")), qs)
        assert len(w) == len(qs), "P7'': 噪声对照未返回全部 quality"
        wdef = {q: round(1.0 - w[q], 6) for q in qs}
        sens_ok = wdef[min(qs)] > 0.5
        ctl.append(dict(control="POS_sensitivity_white_noise", expect="deficit(q_min) > 0.5",
                        observed=wdef[min(qs)], passed=bool(sens_ok), blocking=True,
                        detail=dict(deficit_by_q=wdef,
                                    expectation_provenance="estimated (instrument: derivation chain)",
                                    note="敏感性属性，与上面的口径同一是**两个独立性质**")))
        b = roundtrip(synth_block_aligned(os.path.join(JD, "ctrl", "block8.png")), qs)
        assert len(b) == len(qs), "P7'': 块对齐对照未返回全部 quality"
        bdef = {q: round(1.0 - b[q], 6) for q in qs}
        ctl.append(dict(control="POS_sensitivity_block_aligned", expect="deficit(q_min) > 0.05",
                        observed=bdef[min(qs)], passed=bool(bdef[min(qs)] > 0.05), blocking=True,
                        detail=dict(deficit_by_q=bdef, expectation_provenance="estimated")))
        l = roundtrip(synth_flat_plus_line(os.path.join(JD, "ctrl", "line.png")), qs)
        assert len(l) == len(qs), "P7'': 局部对照未返回全部 quality"
        ldef = {q: round(1.0 - l[q], 6) for q in qs}
        ctl.append(dict(control="POS_sensitivity_local", expect="deficit(q_min) > 0.001",
                        observed=ldef[min(qs)], passed=bool(ldef[min(qs)] > 0.001), blocking=True,
                        detail=dict(deficit_by_q=ldef, expectation_provenance="estimated",
                                    note="局部灵敏度，最贴近真实渲染的相关性")))
        passed = all(x["passed"] for x in ctl if x["blocking"])
        jdump(CONTROLS, dict(instrument_provenance=inst["provenance"], controls=ctl,
                             controls_pass=passed))
        sys.stderr.write("controls_pass(blocking only)=%s\n" % passed)
        for x in ctl:
            sys.stderr.write("  %-28s passed=%-5s blocking=%-5s observed=%s\n"
                             % (x["control"], x["passed"], x["blocking"], x["observed"]))
        for x in ctl:
            sys.stderr.write("  %-26s passed=%s observed=%s\n" % (x["control"], x["passed"], x["observed"]))
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
        assert len(sim_of) > 0, "P7'': current_sim 集合为空"
        sys.stderr.write("PRE-AGG n_scenes=%d\n" % len(sim_of))

        rows = []
        for s in sorted(sim_of.keys()):
            newp = os.path.join(ROOT, "tmp", "pixel-compare", s, "new.png")
            r = roundtrip(newp, qs)                       # 主测量：只用我们自己的渲染
            oldp = os.path.join(ROOT, "tmp", "pixel-compare", s, "old.png")
            x = roundtrip(oldp, qs)                       # 交叉核对：自比，非渲染输入
            defs = {q: round(1.0 - r[q], 6) for q in qs}
            best_q = min(defs, key=lambda q: defs[q])
            d_best = defs[best_q]
            cur = sim_of[s]
            gap = round(0.99 - cur, 6)
            if gap <= 0:
                verdict = "already_passes_gate"
            elif d_best >= gap:
                verdict = "structurally_unreachable"
            else:
                verdict = "codec_not_main_cause"
            rows.append(dict(
                scene=s, current_sim=cur, gap_to_gate=gap,
                roundtrip_sim_by_q=r, codec_deficit_by_q=defs,
                codec_deficit_best=d_best, best_quality=best_q,
                floor_bound_similarity=round(1.0 - d_best, 6),
                arithmetic_class=verdict,
                overturn_factor_needed=round(gap / d_best, 1) if d_best > 0 else None,
                evidentiary_status="未测：量具对编解码噪声的敏感性充分性未证明（instrument amendment_v7）",
                crosscheck_self_reencode_deficit_best=round(1.0 - max(x.values()), 6),
                crosscheck_label="instrument noise floor (self re-encode); NOT a render input",
                produced_by=PRODUCED_BY))
            sys.stderr.write("%-13s gap=%.5f deficit_best=%.5f @q=%s -> %s (x_needed=%s)\n"
                             % (s, gap, d_best, best_q, verdict,
                                rows[-1]["overturn_factor_needed"]))

        ap = [r for r in rows if r["arithmetic_class"] == "already_passes_gate"]
        su = [r for r in rows if r["arithmetic_class"] == "structurally_unreachable"]
        nm = [r for r in rows if r["arithmetic_class"] == "codec_not_main_cause"]
        assert len(ap) + len(su) + len(nm) == len(rows) == 16, "校验式失败：判定之和必须 == 16"
        jdump(RESULT, dict(
            instrument_provenance=inst["provenance"],
            reporting_rule=inst["reporting_rule"],
            encoder_difference="未量化（本轮只量往返；Chromium 与 Go image/jpeg 的伪影差异是往返测不出来的加项）",
            policy_bound_verbatim=inst["ruler"]["policy_bound_verbatim"],
            candidate_formats_note=inst["ruler"]["candidateFormats"],
            n_already_passes_gate=len(ap), n_structurally_unreachable=len(su),
            n_codec_not_main_cause=len(nm),
            checksum_verdict_sum_equals_16=(len(ap) + len(su) + len(nm) == 16),
            evidentiary_status="未测：量具对编解码噪声的敏感性充分性未证明，故「地板下界 << 缺口 ⇒ 编解码不是主因」在本轮不具备证据资格。数字已测，判定未测，两句分开。",
            scenes_unreachable_arithmetic=[r["scene"] for r in su],
            rows=rows))
        sys.exit(0)

    sys.stderr.write("usage: floor.py {controls|measure}\n")
    sys.exit(3)


if __name__ == "__main__":
    main()
