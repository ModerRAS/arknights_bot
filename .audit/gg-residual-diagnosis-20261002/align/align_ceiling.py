#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""alignment 配准天花板测量（诊断用途，不改门禁、不改渲染）。

仪器定义（INSTRUMENT）——接班人靠这段反解数字来源：
  metric      S(dx,dy) = 1 - sum(|O[y,x,c] - Nshift[y,x,c]|) / (W*H*4*255)
              与 Go 端 similarityNormalized（src/ggrender/pixel_test.go）同式：
              4 通道分母、逐像素 abs 差、对 Nshift 取**零填充**（越界像素视为 0）。
  shift 定义  Nshift[y,x,c] = N[y-dy, x-dx, c]，y∈[max(0,dy), min(H,H+dy))、
              x∈[max(0,dx), min(W,W+dx))，其余为 0。
              => 报告里的 (dx,dy) 读作「把 new 的内容往 (dx,dy) 方向挪，可对齐 old」。
  search      整数 (dx,dy) ∈ [-R,R]²，step=1，穷举，不降采样、不做金字塔。
  边界规则    argmin 落在 |dx|=R 或 |dy|=R 上 => 该 R 下只是下界，
              自动以 R'=2R 重搜，仍在边界则 boundary=true（不进「此路不通」）。
  tie-break   按 (err, |dx|+|dy|, |dx|, |dy|, dx, dy) 字典序取最小，保证确定性。
  gate        复用 Go 的 0.99 判据做「够不够得到」；本脚本不改 gatePassed。

阳性对照（CONTROLS）：
  C0 ruler      S(0,0) 必须逐场景等于 go_report.json 的 similarity（delta<1e-12）
  C1 real-data  把某场景 new.png 的内容按已知 (a,b) 边缘复制平移，再用同一搜索找回来
  C2 synthetic  随机噪声图按已知 (a,b) 平移后找回

用法： python .audit/align_ceiling.py [--range 32] [--out tmp/align]
"""

import argparse
import hashlib
import json
import os
import subprocess
import sys
import time

import numpy as np
from PIL import Image

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
GATE = 0.99


# ---------- 仪器原语 ----------
def load_rgba(path):
    with Image.open(path) as im:
        return np.asarray(im.convert("RGBA"))


def sim_of(o, n, W, H, err):
    total = W * H * 4 * 255
    if total <= 0:
        return 1.0
    return 1.0 - float(err) / float(total)


def rect_sum(II, x0, y0, x1, y1):
    """积分图像 O(1) 矩形和。II 形状 (H+1, W+1)。"""
    return int(II[y1, x1] - II[y0, x1] - II[y1, x0] + II[y0, x0])


def search_shift(o, n, R, II=None, totalO=None, step=1):
    """在 [-R,R]^2 上穷举。

    同时追两个指标（同一遍扫描，额外成本 O(1)/shift）：
      err_zero   = sum|O - Nshift_zero|  —— 越界像素按 0 计入
      err_noband = err_zero - band_sum(O) —— 扣掉被零填充带带进来的 O 自身质量，
                   等价于只看重叠区 sum|O - Nshift|，用来检验零填充带有没有驱动结果。
    返回 (best_zero, best_noband)，每个为 (err, dx, dy)。
    """
    H, W = o.shape[0], o.shape[1]
    oi = o.astype(np.int16)
    ni = n.astype(np.int16)
    buf = np.zeros((H, W, 4), np.int16)
    bz = bn = None
    for dx in range(-R, R + 1, step):
        x0, x1 = max(0, dx), min(W, W + dx)
        for dy in range(-R, R + 1, step):
            y0, y1 = max(0, dy), min(H, H + dy)
            buf.fill(0)
            buf[y0:y1, x0:x1] = ni[y0 - dy : y1 - dy, x0 - dx : x1 - dx]
            err = int(np.abs(oi - buf).sum())
            band = totalO - rect_sum(II, x0, y0, x1, y1)
            err_nb = err - band
            tail = (abs(dx) + abs(dy), abs(dx), abs(dy), dx, dy)
            kz = (err,) + tail
            kn = (err_nb,) + tail
            if bz is None or kz < bz:
                bz = kz
            if bn is None or kn < bn:
                bn = kn
    return bz, bn


def integral(o):
    """像素总通道质量的积分图像 + 全图和。"""
    p = o.astype(np.int64).sum(axis=2)
    II = np.zeros((p.shape[0] + 1, p.shape[1] + 1), np.int64)
    II[1:, 1:] = np.cumsum(np.cumsum(p, axis=0), axis=1)
    return II, int(II[-1, -1])


def edge_shift(a, dx, dy):
    """把 a 的内容往 (dx,dy) 挪，边缘复制填充。shift_edge(shift_edge(a,d),-d) == a 精确成立。"""
    H, W = a.shape[0], a.shape[1]
    iy = np.clip(np.arange(H) - dy, 0, H - 1)
    ix = np.clip(np.arange(W) - dx, 0, W - 1)
    return a[np.ix_(iy, ix)]


def sha256_file(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


# ---------- 输入：场景清单机械枚举自 manifest.json ----------
def manifest_scenes(repo_root):
    p = os.path.join(repo_root, "src", "ggrender", "testdata", "visual", "baseline", "manifest.json")
    with open(p, encoding="utf-8") as f:
        mf = json.load(f)
    return [e["id"] for e in mf["entries"]]


def go_report(repo_root):
    p = os.path.join(repo_root, "tmp", "pixel-compare", "report.json")
    with open(p, encoding="utf-8") as f:
        rs = json.load(f)
    return {e["scene"]: e for e in rs}


# ---------- 阳性对照 ----------
def run_controls(repo_root, scenes, R, npairs, verbose=True, inject_fault="none"):
    """inject_fault: none | invert | zero —— 故意把搜索弄坏，验证本对照有牙齿。"""
    results = []

    def search(a, b):
        II, totalO = integral(a)
        bz, bn = search_shift(a, b, R, II, totalO)
        if inject_fault == "invert":
            return bz[0], -bz[4], -bz[5], bn
        if inject_fault == "zero":
            return bz[0], 0, 0, bn
        return bz[0], bz[4], bz[5], bn

    m = R - 1  # 真值必须严格落在搜索范围内部，否则对照测的是范围不是搜索
    deltas = [(0, 0), (max(1, m // 2), -(m // 3)), (-(m // 2), max(1, m // 3)), (2, max(1, m - 1)), (-m, -(m - 1))]
    deltas = [(a, b) for (a, b) in deltas if abs(a) <= m and abs(b) <= m]
    rng = np.random.RandomState(12345)
    picks = [scenes[i] for i in rng.choice(len(scenes), size=min(npairs, len(scenes)), replace=False)]
    for s in picks:
        n = load_rgba(os.path.join(repo_root, "tmp", "pixel-compare", s, "new.png"))
        for (a, b) in deltas:
            syn = edge_shift(n, a, b)
            err, ddx, ddy, _bn = search(n, syn)
            # 真值：syn 的内容相对 n 挪了 (a,b)，找回来的位移应是 (-a,-b)
            ok = (ddx, ddy) == (-a, -b)
            H, W = n.shape[0], n.shape[1]
            rec = sim_of(n.astype(np.int16), syn.astype(np.int16), W, H, err)
            results.append(
                dict(kind="C1-real-data", scene=s, known_shift=[a, b], recovered=[ddx, ddy],
                     exact=bool(ok), recovered_sim=rec, R=R, inject_fault=inject_fault,
                     on_boundary=bool(abs(ddx) == R or abs(ddy) == R))
            )
            if verbose:
                print("  C1 %-12s known=(%d,%d) recovered=(%d,%d) exact=%s sim=%.6f"
                      % (s, a, b, ddx, ddy, ok, rec), flush=True)

    # C2 synthetic noise（同一 R，真值同样严格在内部）
    for (a, b) in [(max(1, m // 2), max(1, m // 2)), (-(m // 2), max(1, m // 3))]:
        syn_img = rng.randint(0, 256, size=(220, 300, 4), dtype=np.uint8)
        syn = edge_shift(syn_img, a, b)
        err, ddx, ddy, _bn = search(syn_img, syn)
        ok = (ddx, ddy) == (-a, -b)
        results.append(dict(kind="C2-synthetic-noise", scene=None, known_shift=[a, b],
                            recovered=[ddx, ddy], exact=bool(ok), R=R, inject_fault=inject_fault,
                            on_boundary=bool(abs(ddx) == R or abs(ddy) == R),
                            recovered_sim=sim_of(syn_img.astype(np.int16), syn.astype(np.int16),
                                                 300, 220, err)))
        if verbose:
            print("  C2 synthetic known=(%d,%d) recovered=(%d,%d) exact=%s"
                  % (a, b, ddx, ddy, ok), flush=True)

    passed = all(r["exact"] for r in results)
    return passed, results


# ---------- 残差统计 ----------
def residual_stats(o, n, dx, dy):
    H, W = o.shape[0], o.shape[1]
    oi = o.astype(np.int16)
    ni = n.astype(np.int16)
    buf = np.zeros((H, W, 4), np.int16)
    x0, x1 = max(0, dx), min(W, W + dx)
    y0, y1 = max(0, dy), min(H, H + dy)
    buf[y0:y1, x0:x1] = ni[y0 - dy : y1 - dy, x0 - dx : x1 - dx]
    per_px = np.abs(oi - buf).sum(axis=2).astype(np.int64)  # 0..1020
    total = int(per_px.sum())
    nz = per_px[per_px > 0]
    q = lambda p: int(np.percentile(nz, p)) if nz.size else 0  # noqa: E731
    top = int(np.sort(nz)[-max(1, nz.size // 100):].sum()) if nz.size else 0
    return dict(
        err_total=total,
        differing_pixels=int((per_px > 0).sum()),
        total_pixels=int(per_px.size),
        pct_of_pixels_differing=round(100.0 * float((per_px > 0).sum()) / float(per_px.size), 4),
        diff_px_p50=q(50), diff_px_p90=q(90), diff_px_p99=q(99), diff_px_max=int(per_px.max()),
        top1pct_pixels_share_of_error=round(100.0 * top / float(total), 4) if total else 0.0,
        pixels_diff_gt16=int((per_px > 16).sum()),
        shift_err_amp=buf,
    )


def write_reg_diff(out_dir, o, buf, scale=8):
    d = np.abs(o.astype(np.int16) - buf).sum(axis=2)
    d = np.clip(d.astype(np.float32) * scale, 0, 255).astype(np.uint8)
    Image.fromarray(d, mode="L").save(os.path.join(out_dir, "reg-diff.png"))
    return d.shape


# ---------- main ----------
def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--range", type=int, default=32)
    ap.add_argument("--out", default="tmp/align")
    ap.add_argument("--control-pairs", type=int, default=3)
    ap.add_argument("--scenes", default="", help="逗号分隔，仅调试用；正式跑留空")
    ap.add_argument("--controls-only", action="store_true")
    ap.add_argument("--inject-fault", default="none", choices=["none", "invert", "zero"],
                    help="故意弄坏搜索以证明阳性对照有牙齿；正式跑必须是 none")
    ap.add_argument("--step", type=int, default=1, help="场景扫描步长；对照组恒为 1")
    ap.add_argument("--skip-controls", action="store_true",
                    help="仅用于宽范围粗扫的补充扫描；跳过对照时本报告不得用作主天花板")
    args = ap.parse_args()

    repo_root = ROOT
    out_dir = os.path.join(repo_root, args.out)
    os.makedirs(out_dir, exist_ok=True)
    t_start = time.time()

    scenes = manifest_scenes(repo_root)
    go = go_report(repo_root)
    if args.scenes:
        want = set(args.scenes.split(","))
        scenes = [s for s in scenes if s in want]

    instrument = dict(
        provenance="measured",
        script=os.path.relpath(os.path.abspath(__file__), repo_root),
        metric="S(dx,dy)=1-sum(|O-Nshift|)/(W*H*4*255); 4-channel denominator, per-pixel abs diff, integer shifts",
        shift_convention="Nshift[y,x,c]=N[y-dy,x-dx,c] inside [max(0,dy),min(H,H+dy))x[max(0,dx),min(W,W+dx)); outside = 0 (zero padding)",
        shift_reading="(dx,dy) = how far new's content must move to line up with old",
        search="brute force integer (dx,dy) in [-R,R]^2, step %d, full canvas, no downsampling" % args.step,
        R_initial=args.range,
        boundary_rule="argmin on |dx|=R or |dy|=R -> rerun with R'=2R; boundary=true if still on edge",
        tie_break="lexicographic min of (err, |dx|+|dy|, |dx|, |dy|, dx, dy)",
        gate_reused="0.99 (Go gatePassed); this script does NOT modify gatePassed or similarityNormalized",
        positive_controls="C1 real-data edge-shift of new.png; C2 synthetic noise; C0 S(0,0) must equal go report.json",
        scope="diagnosis only: reported (dx,dy) are measurement results, NOT renderer constants",
    )
    with open(os.path.join(out_dir, "instrument.json"), "w", encoding="utf-8") as f:
        json.dump(instrument, f, ensure_ascii=False, indent=2)

    # C0: 尺子自检 —— S(0,0) vs Go
    c0 = []
    for s in scenes:
        o = load_rgba(os.path.join(repo_root, "tmp", "pixel-compare", s, "old.png"))
        n = load_rgba(os.path.join(repo_root, "tmp", "pixel-compare", s, "new.png"))
        assert o.shape == n.shape, (s, o.shape, n.shape)
        H, W = o.shape[0], o.shape[1]
        err0 = int(np.abs(o.astype(np.int32) - n.astype(np.int32)).sum())
        s00 = sim_of(o, n, W, H, err0)
        c0.append(dict(scene=s, py_sim=s00, go_sim=go[s]["similarity"], delta=abs(s00 - go[s]["similarity"])))
    c0_pass = all(c["delta"] < 1e-12 for c in c0)
    print("C0 ruler self-check: pass=%s max_delta=%.3e" % (c0_pass, max(c["delta"] for c in c0)), flush=True)

    # 阳性对照
    if args.skip_controls:
        c0_pass = ctrl_pass = None
        ctrl = []
        print("controls SKIPPED (supplementary wide sweep only; not a primary ceiling run)", flush=True)
    else:
        print("positive controls (R=%d, inject_fault=%s):" % (args.range, args.inject_fault), flush=True)
        ctrl_pass, ctrl = run_controls(repo_root, scenes, args.range, args.control_pairs,
                                       inject_fault=args.inject_fault)
        print("controls pass=%s" % ctrl_pass, flush=True)

    harness = dict(
        harness_path="src/ggrender/pixel_test.go",
        repo_rel_source="arknights_bot-measure-align (git worktree)",
        branch=subprocess.check_output(["git", "-C", repo_root, "rev-parse", "--abbrev-ref", "HEAD"],
                                       text=True).strip(),
        commit=subprocess.check_output(["git", "-C", repo_root, "rev-parse", "HEAD"], text=True).strip(),
        go_cmd="cd src && go test ./ggrender/ -run TestGGPixelParity -count=1",
        go_log=os.path.join(args.out, "gotest.log"),
        produced_by="gg harness only (RenderGGContext); yoga/skia 树未参与、未读取",
        inject_fault=args.inject_fault,
    )
    with open(os.path.join(out_dir, "controls.json"), "w", encoding="utf-8") as f:
        json.dump(dict(c0_ruler=c0, c0_pass=c0_pass, c1_c2_pass=ctrl_pass,
                       c1_c2=ctrl, harness=harness), f, ensure_ascii=False, indent=2)

    if not (c0_pass and ctrl_pass):
        print("POSITIVE CONTROL FAILED -> ceiling numbers are void, do not report a table.", flush=True)
    if args.controls_only:
        return

    # 逐场景
    rows = []
    for s in scenes:
        o = load_rgba(os.path.join(repo_root, "tmp", "pixel-compare", s, "old.png"))
        n = load_rgba(os.path.join(repo_root, "tmp", "pixel-compare", s, "new.png"))
        H, W = o.shape[0], o.shape[1]
        R = args.range
        t0 = time.time()
        II, totalO = integral(o)
        bz, bn = search_shift(o, n, R, II, totalO, args.step)
        ez, dzx, dzy = bz[0], bz[4], bz[5]
        en, dnx, dny = bn[0], bn[4], bn[5]
        # 边界扩展只作用于主口径（零填充）；band-free 探针按构造会单调奖励「大位移」
        # （去掉带惩罚后把内容移出画布不花钱），它的无界 argmax 不是天花板估计值，
        # 只用来检查零填充带有没有在驱动排名。
        if abs(dzx) == R or abs(dzy) == R:
            R = 2 * R
            bz, bn = search_shift(o, n, R, II, totalO, args.step)
            ez, dzx, dzy = bz[0], bz[4], bz[5]
            en, dnx, dny = bn[0], bn[4], bn[5]
        boundary = abs(dzx) == R or abs(dzy) == R
        probe_boundary = abs(dnx) == R or abs(dny) == R
        s_best = sim_of(o, n, W, H, ez)
        s_best_nb = sim_of(o, n, W, H, en)
        err0 = int(np.abs(o.astype(np.int32) - n.astype(np.int32)).sum())
        s_now = sim_of(o, n, W, H, err0)
        res = residual_stats(o, n, dzx, dzy)
        scene_dir = os.path.join(out_dir, s)
        os.makedirs(scene_dir, exist_ok=True)
        write_reg_diff(scene_dir, o, res.pop("shift_err_amp"))
        explained = 0.0 if err0 == 0 else 100.0 * (1.0 - float(ez) / float(err0))
        row = dict(
            scene=s, width=W, height=H,
            current_sim=s_now, go_sim=go[s]["similarity"], go_passed=go[s]["passed"],
            best_shift=[dzx, dzy], best_sim=s_best, delta_sim=s_best - s_now,
            best_shift_noband=[dnx, dny], best_sim_noband=s_best_nb,
            argmax_agrees_noband=bool([dzx, dzy] == [dnx, dny]),
            probe_boundary=bool(probe_boundary),
            gap_to_gate=GATE - s_best, reaches_gate=bool(s_best >= GATE),
            err0=err0, err_best=ez, err_explained_pct=round(explained, 4),
            R_used=R, boundary=bool(boundary),
            residual=res,
            new_png_sha256=sha256_file(os.path.join(repo_root, "tmp", "pixel-compare", s, "new.png")),
            old_png_sha256=sha256_file(os.path.join(repo_root, "tmp", "pixel-compare", s, "old.png")),
            produced_by=harness["harness_path"] + " @ " + harness["commit"],
            seconds=round(time.time() - t0, 1),
        )
        rows.append(row)
        print("%-13s now=%.5f best=%.5f shift=(%d,%d) R=%d bnd=%s nob=(%d,%d)/%.5f agree=%s d=%.5f gap=%.5f %s"
              % (s, s_now, s_best, dzx, dzy, R, boundary, dnx, dny, s_best_nb,
                 [dzx, dzy] == [dnx, dny], s_best - s_now, GATE - s_best,
                 "REACH" if s_best >= GATE else "no-reach"), flush=True)

    # 校验式
    passed = [r for r in rows if r["go_passed"]]
    failed = [r for r in rows if not r["go_passed"]]
    manifest_ids = set(scenes)
    row_ids = set(r["scene"] for r in rows)
    checks = dict(
        manifest_scene_count=len(manifest_ids),
        row_count=len(rows),
        id_sets_equal=bool(manifest_ids == row_ids),
        passed_count=len(passed),
        failed_count=len(failed),
        checksum_passed_plus_failed_eq_manifest=len(passed) + len(failed) == len(manifest_ids),
        unpassed_reaches_gate=[r["scene"] for r in failed if r["reaches_gate"]],
        unpassed_no_reach=[r["scene"] for r in failed if not r["reaches_gate"]],
        unpassed_boundary=[r["scene"] for r in failed if r["boundary"]],
    )
    with open(os.path.join(out_dir, "report.json"), "w", encoding="utf-8") as f:
        json.dump(dict(instrument=instrument, harness=harness,
                       controls=dict(c0_pass=c0_pass, c1_c2_pass=ctrl_pass),
                       checks=checks, rows=rows, elapsed_sec=round(time.time() - t_start, 1)),
                  f, ensure_ascii=False, indent=2)
    print(json.dumps(checks, ensure_ascii=False, indent=2), flush=True)
    print("elapsed %.1fs" % (time.time() - t_start), flush=True)


if __name__ == "__main__":
    main()
