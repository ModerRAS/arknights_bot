#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
missing 场景 · 只读差距分解（一次性测量）

复用已入库的 ink 掩膜定义（直接 import，不另造一套）：
    C:/WorkSpace/Golang/arknights_bot-gg-card-atomic/.audit/content-mask-overlap-20260930/mask_overlap.py
    background = 精确众数 24-bit RGB；ink mask = np.abs(pixel-bg).sum(axis=2) > 30（L1，阈值 30）
    inkOv = Jaccard，分母并集；rowCorr/colCorr = 逐行/逐列 ink 计数的 Pearson

相似度公式与 harness 的 similarityNormalized() 同式（只读引用，未修改任何代码）：
    sim = 1 - sum(|dR|+|dG|+|dB|+|dA|) / (w*h*4*255)

红线：下面每一个数字都是「误差落在哪里」的测量，**没有一个坐标、没有一个参数来自基线图**。
`ceiling_if_background_oracle` 那一项是把背景类像素置为基线值后重算的**天花板上界**，
它回答「这一类修法最多能拿多少分」，不产出任何可接线��参数。

输入：只读 harness 已在 1b433bc 上写出的 tmp/pixel-compare/missing/{old,new}.png。
输出：原始产物写到仓库外 C:/WorkSpace/Golang/_lead2/（即本文件的 original_path）；
2026-09-30 决定归档后，本副本与三个结果 JSON 一起入库到
.audit/missing-decomp-1b433bc/。运行本脚本请从仓库外执行并加 -B，理由见
.audit/content-mask-overlap-20260930/report.json 的 regeneration 节。
"""
import io
import json
import os
import sys

import numpy as np
from PIL import Image

ARCHIVE = r'C:\WorkSpace\Golang\arknights_bot-gg-card-atomic\.audit\content-mask-overlap-20260930'
sys.path.insert(0, ARCHIVE)
from mask_overlap import modal, corr  # noqa: E402  复用入库定义，不另造

ROOT = r'C:\WorkSpace\Golang\arknights_bot-consolidate\tmp\pixel-compare'
SCENE = 'missing'
COLS, ROWS = 12, 8
OUT = r'C:\WorkSpace\Golang\_lead2'


def main():
    o = np.asarray(Image.open(os.path.join(ROOT, SCENE, 'old.png')).convert('RGB'), dtype=np.int64)
    n = np.asarray(Image.open(os.path.join(ROOT, SCENE, 'new.png')).convert('RGB'), dtype=np.int64)
    h, w = o.shape[0], o.shape[1]

    # 与 similarityNormalized() 同式的逐像素误差
    d = np.abs(o - n).sum(axis=2)                      # |dR|+|dG|+|dB|，alpha 恒 255
    total_err = int(d.sum())
    budget = w * h * 4 * 255
    sim = 1.0 - total_err / budget

    # ink 掩膜（入库定义）
    mo, _ = modal(o)
    mn, _ = modal(n)
    mo = np.array([(mo >> 16) & 255, (mo >> 8) & 255, mo & 255])
    mn = np.array([(mn >> 16) & 255, (mn >> 8) & 255, mn & 255])
    eo = np.abs(o - mo).sum(axis=2) > 30
    en = np.abs(n - mn).sum(axis=2) > 30
    ink_union = eo | en
    ink_inter = eo & en
    rowCorr = corr(eo.sum(axis=1).astype(float), en.sum(axis=1).astype(float))
    colCorr = corr(eo.sum(axis=0).astype(float), en.sum(axis=0).astype(float))

    # 误差预算分层：内容像素 vs 平铺背景像素（并集掩膜之外的部分）
    err_ink = int(d[ink_union].sum())
    err_bg = int(d[~ink_union].sum())
    n_ink = int(ink_union.sum())
    n_bg = int((~ink_union).sum())

    # 12x8 瓦片
    tiles = []
    ys = np.linspace(0, h, ROWS + 1).astype(int)
    xs = np.linspace(0, w, COLS + 1).astype(int)
    for ry in range(ROWS):
        for rx in range(COLS):
            sub = d[ys[ry]:ys[ry + 1], xs[rx]:xs[rx + 1]]
            sube = err_ink if True else 0
            tiles.append({
                'row': ry, 'col': rx,
                'y0': int(ys[ry]), 'y1': int(ys[ry + 1]),
                'x0': int(xs[rx]), 'x1': int(xs[rx + 1]),
                'px': int(sub.size),
                'err': int(sub.sum()),
                'err_share_pct': round(100.0 * sub.sum() / total_err, 3),
                'mean_abs_l1': round(float(sub.mean()), 3),
                'ink_pct': round(100.0 * float(eo[ys[ry]:ys[ry + 1], xs[rx]:xs[rx + 1]].mean()), 2),
            })
    tiles.sort(key=lambda t: -t['err'])
    top = tiles[:10]
    top5_share = round(sum(t['err_share_pct'] for t in tiles[:5]), 3)
    top10_share = round(sum(t['err_share_pct'] for t in tiles[:10]), 3)

    # 天花板上界：把「非内容」像素按基线值对齐后重算（oracle，仅作上界）
    o2 = o.copy()
    o2[~ink_union] = n[~ink_union]
    sim_ceiling = 1.0 - int(np.abs(o2 - n).sum(axis=2).sum()) / budget

    res = {
        'scene': SCENE,
        'measurement_point': '1b433bc (consolidate/gg-mainline tip, PR #5 source branch)',
        'inputs': {
            'old': os.path.join(ROOT, SCENE, 'old.png'),
            'new': os.path.join(ROOT, SCENE, 'new.png'),
            'hashOld': 'cc58fea684f3c1a9fc879da163cf50e9438753b3ebcb7ecf3a5ef3b2f8a8ca47',
            'hashNew': '01b88d9597b9e29a3efc1aefe66e0020b8d0b1a5e86e18677bce910d69ce95b7',
        },
        'definition_source': 'mask_overlap.py @ 6f444a1 (imported, not reimplemented)',
        'geometry': {'w': int(w), 'h': int(h), 'scale': 1.5},
        'similarity': {
            'value': round(sim, 10),
            'harness_formula': '1 - sum(|dR|+|dG|+|dB|+|dA|) / (w*h*4*255)',
            'total_err': total_err,
            'budget': budget,
            'err_budget_pct': round(100.0 * total_err / budget, 4),
        },
        'ink_metrics': {
            'inkOv_pct': round(100.0 * ink_inter.sum() / max(1, ink_union.sum()), 2),
            'rowCorr': round(rowCorr, 4),
            'colCorr': round(colCorr, 4),
            'ink_px_old': int(eo.sum()),
            'ink_px_new': int(en.sum()),
            'ink_px_union': n_ink,
            'ink_px_inter': int(ink_inter.sum()),
            'flat_bg_px': n_bg,
            'flat_bg_pct_of_canvas': round(100.0 * n_bg / (w * h), 2),
        },
        'error_budget_split': {
            'content_px_pct_of_canvas': round(100.0 * n_ink / (w * h), 2),
            'content_px_err': err_ink,
            'content_px_err_share_pct': round(100.0 * err_ink / total_err, 3),
            'content_px_mean_abs_l1': round(err_ink / max(1, n_ink), 3),
            'flat_bg_px_pct_of_canvas': round(100.0 * n_bg / (w * h), 2),
            'flat_bg_px_err': err_bg,
            'flat_bg_px_err_share_pct': round(100.0 * err_bg / total_err, 3),
            'flat_bg_px_mean_abs_l1': round(err_bg / max(1, n_bg), 3),
        },
        'tiles': {'grid': '%dx%d' % (COLS, ROWS), 'top10': top,
                  'top5_err_share_pct': top5_share, 'top10_err_share_pct': top10_share,
                  'all': tiles},
        'ceiling_if_background_oracle': {
            'similarity': round(sim_ceiling, 10),
            'delta': round(sim_ceiling - sim, 6),
            'meaning': 'upper bound only: aligning every non-ink (flat background) pixel with the baseline. '
                       'Yields no connectable coordinate or parameter; it only bounds what background-level fixes can buy.',
        },
    }
    os.makedirs(OUT, exist_ok=True)
    p = os.path.join(OUT, 'missing-decomp-1b433bc.json')
    with io.open(p, 'w', encoding='utf-8', newline='\n') as f:
        json.dump(res, f, indent=2, ensure_ascii=False)
    assert json.load(io.open(p, encoding='utf-8')) == res
    print('wrote + read-back OK:', p)
    print('sim %.5f  err_budget %.4f%%  top5 %.1f%%  top10 %.1f%%' % (sim, res['similarity']['err_budget_pct'], top5_share, top10_share))
    print('content px %.1f%% of canvas, err share %.1f%%, mean|dL1| %.1f' % (
        res['error_budget_split']['content_px_pct_of_canvas'],
        res['error_budget_split']['content_px_err_share_pct'],
        res['error_budget_split']['content_px_mean_abs_l1']))
    print('flat bg px %.1f%% of canvas, err share %.1f%%, mean|dL1| %.2f' % (
        res['error_budget_split']['flat_bg_px_pct_of_canvas'],
        res['error_budget_split']['flat_bg_px_err_share_pct'],
        res['error_budget_split']['flat_bg_px_mean_abs_l1']))
    print('ceiling if background oracle: %.5f (%+.5f)' % (sim_ceiling, sim_ceiling - sim))
    print('rowCorr %.3f colCorr %.3f inkOv %.1f%%' % (rowCorr, colCorr, res['ink_metrics']['inkOv_pct']))
    print('--- top 8 tiles (row,col, share%, mean|dL1|, ink%) ---')
    for t in top[:8]:
        print('  r%d c%-2d share %6.2f%%  mean %7.2f  ink %5.1f%%  y[%d:%d] x[%d:%d]' % (
            t['row'], t['col'], t['err_share_pct'], t['mean_abs_l1'], t['ink_pct'], t['y0'], t['y1'], t['x0'], t['x1']))


if __name__ == '__main__':
    main()
