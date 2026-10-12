#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
check_baseline_as_render_input.py

Re-runnable audit: does ANY render path in src/ggrender/ read the frozen Playwright
baseline as a RENDER INPUT?

The distinction that makes or breaks this check:
  * comparison basis  = baseline bytes consumed by diff/similarity computation  -> LEGAL
  * render input       = baseline bytes consumed by a draw/Scale/encode path    -> CHEAT

So we cannot simply exclude pixel_test.go (that would miss "someone added a render
input inside the test"), and we cannot simply grep "baseline" (that flags the legal
comparison side, and also flags comments and a local variable named `baseline`).

Method: enumerate every read-in primitive, then for each call site classify the path
argument as LITERAL / FORMAT / PARAM / DATA, and decide the consumer.

Usage:
    python check_baseline_as_render_input.py [--ref HEAD] [--json out.json] [--md out.md]

Exit codes:
    0 = no render path reads the baseline as render input
    1 = VIOLATION found
    2 = ruler self-check failed (cannot trust a "zero hits" result)
"""
import argparse, json, os, re, subprocess, sys, io

PKG = "src/ggrender"

# Read-in primitives. Each must be located before we can classify anything.
# kind: "file" = reads the filesystem (subject to the baseline-path check)
#       "net"  = network fetch, its argument is a URL, NOT a filesystem path
READ_PRIMS = [
    ("os.Open",        r"os\.Open\(",     "file"),
    ("os.ReadFile",    r"os\.ReadFile\(", "file"),
    ("LoadImage",      r"\bLoadImage\(",  "file"),
    ("tryLocal",       r"\btryLocal\(",   "file"),
    ("cardAsset",      r"\bcardAsset\(",  "file"),
    ("AssetPath",      r"\bAssetPath\(",  "file"),
    ("fetch",          r"\bfetch\(",      "net"),
    ("FetchImage",     r"\bFetchImage\(", "net"),
    ("LoadFontFace",   r"LoadFontFace\(", "font"),
]

# Path fragments that would indicate the frozen baseline being reached for.
BASELINE_MARKERS = ["testdata", "visual", "baseline", "images/", "\\baseline", "baselineDir"]

# Functions that are render-side consumers of an image.
RENDER_CONSUMERS = ["DrawImage", "ScaleExact", "ScaleCover", "ScaleExactCR", "ScaleToManifest",
                    "EncodePNG", "DrawRectangle", "DrawRoundedRectangle", "Fill", "DrawImageAnchored"]

GO_KEYWORDS = re.compile(r"^\s*(func|type|var|const)\s")


def run_git(args, cwd):
    p = subprocess.run(["git"] + args, cwd=cwd, capture_output=True, text=True,
                       encoding="utf-8", errors="replace")
    return p.returncode, p.stdout, p.stderr


class RulerFailure(Exception):
    pass


def self_check(cwd, ref):
    """POSITIVE CONTROL: prove the ruler fires before trusting any zero-hit.

    Hard controls are STRUCTURAL (guaranteed in any ref of this package) so the check
    stays usable on old commits; an earlier version hard-coded `ScaleToManifest`, which
    did not exist in 641988f and made the control fire-fail on a dirty commit -- i.e.
    the ruler blocked the very test it existed to enable. Semantic controls are reported
    but not fatal.
    """
    rc, out, err = run_git(["ls-tree", "-r", "--name-only", ref, "--", PKG + "/"], cwd)
    if rc != 0:
        raise RulerFailure(f"ls-tree failed rc={rc}: {err.strip()}")
    files = [l.strip() for l in out.splitlines() if l.strip().endswith(".go")]
    if not files:
        raise RulerFailure("no .go files under %s -- cannot trust any zero-hit" % PKG)

    # HARD control 1: git show + grep both work on this package.
    blob = ""
    for f in files:
        rc2, o2, e2 = run_git(["show", f"{ref}:{f}"], cwd)
        if rc2 != 0:
            raise RulerFailure(f"cannot read {f} at {ref}: {e2.strip()}")
        blob += o2
    if "package ggrender" not in blob:
        raise RulerFailure("HARD control failed: 'package ggrender' not found -- ruler cannot read the package")
    hard1 = {"control": "git show + grep reach the package", "needle": "package ggrender",
             "hits": blob.count("package ggrender"), "fired": True, "hard": True}

    # HARD control 2: the read-primitive regexes CAN match real code here.
    prim_hits = {n: len(re.findall(p, blob)) for n, p, _ in READ_PRIMS}
    if sum(prim_hits.values()) == 0:
        raise RulerFailure(f"HARD control failed: no read-primitive matched -- regexes are broken: {prim_hits}")
    hard2 = {"control": "read-primitive regexes match real code", "needle": " | ".join(READ_PRIMS[i][0] for i in range(len(READ_PRIMS))),
             "hits": sum(prim_hits.values()), "fired": True, "hard": True, "per_primitive": prim_hits}

    # SOFT controls: semantic, may legitimately be absent on older refs.
    soft = []
    for name, needle, where in [("baseline token in pixel_test.go", "baseline", f"{PKG}/pixel_test.go"),
                                ("ScaleToManifest in helpers.go", "ScaleToManifest", f"{PKG}/helpers.go")]:
        rc3, o3, _ = run_git(["show", f"{ref}:{where}"], cwd)
        n = o3.count(needle) if rc3 == 0 else 0
        soft.append({"control": name, "needle": needle, "hits": n,
                     "fired": n > 0, "hard": False,
                     "note": "" if n > 0 else "absent in this ref (informational, not fatal)"})

    return [hard1, hard2] + soft


def list_go_files(cwd, ref):
    rc, out, err = run_git(["ls-tree", "-r", "--name-only", ref, "--", PKG + "/"], cwd)
    if rc != 0:
        raise RulerFailure(f"ls-tree failed rc={rc}: {err.strip()}")
    files = [l.strip() for l in out.splitlines() if l.strip().endswith(".go")]
    if not files:
        raise RulerFailure("no .go files found under %s -- ruler cannot be trusted" % PKG)
    return files


def enclosing_func(lines, idx):
    """Walk backwards to the nearest `func` declaration line."""
    for i in range(idx, -1, -1):
        if lines[i].startswith("func "):
            name = lines[i][5:].split("(")[0].strip()
            return name
    return "<file-scope>"


def classify_arg(arg):
    a = arg.strip()
    if a.startswith('"') and a.endswith('"'):
        return "LITERAL", a.strip('"')
    if a.startswith("fmt.Sprintf"):
        m = re.search(r'"([^"]*)"', a)
        if "%d" in a or "%s" in a or "%v" in a:
            return ("PARAM" if ("c." in a or "d." in a or "o." in a or "icon" in a or "rel" in a)
                    else "FORMAT"), (m.group(1) if m else a)
        return "LITERAL", (m.group(1) if m else a)
    if ".." in a:
        return "TRAVERSAL", a
    if re.search(r"\b(rel|icon|path|fallbackPath|portraitURL|iconSrc)\b", a):
        return "PARAM", a
    if a.startswith("AssetPath(") or a.startswith("filepath.Join(AssetRoot"):
        return "PARAM", a
    return "UNKNOWN", a


def split_args(s):
    """Split a Go call argument list on top-level commas."""
    out, depth, cur, instr = [], 0, "", False
    i = 0
    while i < len(s):
        c = s[i]
        if c == '"':
            instr = not instr
        if not instr:
            if c in "([{":
                depth += 1
            elif c in ")]}":
                depth -= 1
            elif c == "," and depth == 0:
                out.append(cur)
                cur = ""
                i += 1
                continue
        cur += c
        i += 1
    if cur.strip():
        out.append(cur)
    return out


def extract_call_args(line):
    """Extract the argument list of the first call on the line, honouring nesting."""
    m = re.search(r"(\w+)\(", line)
    if not m:
        return ""
    start = m.end() - 1
    depth = 0
    for i in range(start, len(line)):
        if line[i] == "(":
            depth += 1
        elif line[i] == ")":
            depth -= 1
            if depth == 0:
                return line[start + 1:i]
    return line[start + 1:]


def scan_file(cwd, ref, path, data):
    rc, out, err = run_git(["show", f"{ref}:{path}"], cwd)
    if rc != 0:
        raise RulerFailure(f"cannot read {path}: rc={rc} {err.strip()}")
    lines = out.splitlines()
    is_test = path.endswith("_test.go")

    # Resolve package-level path vars so we can PROVE where they point, e.g.
    # `var amiyaPath = filepath.Join(AssetRoot, "common", "amiya.png")`.
    pkg_vars = {}
    for l in lines:
        m = re.match(r"\s*var\s+(\w+)\s*=\s*(.+)$", l)
        if m:
            pkg_vars[m.group(1)] = m.group(2).strip()
        m2 = re.match(r"\s*(\w+)\s*=\s*filepath\.Join\((.+)$", l)
        if m2:
            pkg_vars.setdefault(m2.group(1), m2.group(2).strip())

    def resolve(sym):
        """Follow a symbol to its definition; return a short provenance string."""
        seen = set()
        cur = sym
        for _ in range(6):
            if cur in seen:
                break
            seen.add(cur)
            body = pkg_vars.get(cur)
            if body is None:
                return "literal/unknown-origin"
            if "AssetRoot" in body:
                return "AssetRoot-rooted"
            m = re.match(r"(\w+)$", body.strip())
            if m and m.group(1) in pkg_vars:
                cur = m.group(1)
                continue
            if body.strip().startswith('"') or "Join(" in body:
                return "literal-asset"
            return "unresolved"
        return "unresolved"

    sites = []
    for idx, line in enumerate(lines):
        stripped = line.strip()
        # A `func X(...)` line is a DECLARATION, not a call site. Matching it
        # produced phantom findings (e.g. `func fetch(url string)`).
        is_decl = stripped.startswith("func ") or stripped.startswith("//") \
            or stripped.startswith("type ") or stripped.startswith("var ") \
            or stripped.startswith("const ")
        for pname, pat, pkind in READ_PRIMS:
            if is_decl:
                continue
            if pname == "AssetPath" and path != f"{PKG}/helpers.go":
                # AssetPath is only a joiner; classify at its LoadImage/tryLocal call sites
                if not re.search(r"LoadImage\(\s*AssetPath|tryLocal\(\s*AssetPath", line):
                    continue
            for m in re.finditer(pat, line):
                if m.start() > 0 and (line[m.start() - 1].isalnum() or line[m.start() - 1] == "_"):
                    continue  # part of a longer identifier
                arg = extract_call_args(line[m.start():])
                kind, val = classify_arg(arg)

                # A network primitive's argument is a URL, not a filesystem path.
                # FetchImage(url, fallbackPath): ONLY the 2nd arg is a path.
                if pkind == "net":
                    parts = [p.strip() for p in split_args(arg)]
                    path_arg = parts[1] if len(parts) > 1 and pname == "FetchImage" else ""
                    if not path_arg:
                        kind, val = "NETWORK(url)", (parts[0] if parts else arg)[:120]
                if pkind == "font":
                    kind, val = "FONT-CANDIDATE", "FontCandidates (assets/font + system)"

                # Resolve symbol arguments (amiyaPath, FontCandidates, ...) to a
                # provenance instead of dumping them into an UNKNOWN bucket.
                prov = ""
                for sym in re.findall(r"\b(\w+)\b", val):
                    if sym in pkg_vars or sym == "AssetRoot" or sym == "FontCandidates":
                        prov = resolve(sym) if sym in pkg_vars else ("AssetRoot-rooted"
                                                                    if sym == "AssetRoot"
                                                                    else "font-candidate-list")
                        break
                if kind == "UNKNOWN" and prov in ("AssetRoot-rooted", "literal-asset",
                                                   "font-candidate-list", "literal/unknown-origin"):
                    kind = "RESOLVED-" + prov
                if "AssetRoot" in val:
                    prov = prov or "AssetRoot-rooted"
                sites.append({
                    "file": path, "line_no": idx + 1,
                    "function": enclosing_func(lines, idx),
                    "primitive": pname, "primitive_kind": pkind,
                    "arg_raw": arg.strip()[:120],
                    "arg_kind": kind, "arg_value": val[:120],
                    "provenance": prov,
                    "side": "COMPARISON(test file)" if is_test else "RENDER",
                    "line_text": line.strip()[:160],
                })
    # marker hits (baseline path fragments) per file
    markers = []
    for idx, line in enumerate(lines):
        low = line.lower()
        for mk in BASELINE_MARKERS:
            if mk not in low:
                continue
            # A bare identifier named `baseline` is NOT a path. In scene_state.go
            # `baseline := labelTop + 21` is a typography y-coordinate; flagging it
            # is the grep trap this check exists to avoid. Only treat a marker as
            # path evidence when it sits inside a string literal or a path expression.
            in_string = re.search(r'"[^"]*' + re.escape(mk), low) is not None
            in_path_expr = bool(re.search(r"(filepath\.Join|AssetPath|LoadImage|os\.Open|ReadFile)"
                                          r"[^\n]*" + re.escape(mk), line))
            is_comment = line.strip().startswith("//")
            if not (in_string or in_path_expr):
                continue  # identifier use, not a path reference
            markers.append({
                "file": path, "line_no": idx + 1,
                "function": enclosing_func(lines, idx),
                "marker": mk, "line_text": line.strip()[:160],
                "is_comment": is_comment,
                "evidence": "string-literal" if in_string else "path-expression",
                "side": "COMPARISON(test file)" if is_test else "RENDER",
            })
            break
    return sites, markers, out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--ref", default="HEAD")
    ap.add_argument("--cwd", default=os.getcwd())
    ap.add_argument("--json", default=None)
    ap.add_argument("--md", default=None)
    a = ap.parse_args()
    cwd = a.cwd

    report = {"ref": a.ref, "package": PKG}
    try:
        report["ruler_self_check"] = self_check(cwd, a.ref)
        files = list_go_files(cwd, a.ref)
    except RulerFailure as e:
        report["verdict"] = "RULER_FAILED"
        report["error"] = str(e)
        emit(report, a)
        print("RULER SELF-CHECK FAILED:", e, file=sys.stderr)
        return 2

    all_sites, all_markers, sources = [], [], {}
    for f in files:
        s, m, src = scan_file(cwd, a.ref, f, None)
        all_sites += s
        all_markers += m
        sources[f] = src

    violations, reviewed = [], []
    for s in all_sites:
        if s.get("primitive_kind") in ("net", "font"):
            continue  # URL / font-candidate argument: not a filesystem path
        flag = None
        v = (s["arg_value"] or "").lower()
        if any(mk.lower() in v for mk in ["testdata", "visual/baseline", "baseline", "images/"]):
            flag = "read path names the frozen baseline"
        elif s["arg_kind"] == "TRAVERSAL":
            flag = "path contains '..' (can escape AssetRoot)"
        elif s["arg_kind"] == "UNKNOWN" and not s.get("provenance"):
            flag = "unclassifiable argument -- needs human read"
        if flag:
            if s["side"] == "RENDER":
                violations.append(dict(s, reason=flag))
            else:
                reviewed.append(dict(s, reason=flag,
                                     role="COMPARISON BASIS (legal): read by diff/similarity, "
                                          "never passed to a render call"))

    # In the test file, prove the baseline never reaches a render consumer.
    test_flow = []
    if f"{PKG}/pixel_test.go" in sources:
        tsrc = sources[f"{PKG}/pixel_test.go"]
        render_calls = re.findall(r"RenderGG\w*\([^)]*\)", tsrc)
        # find variables assigned from baseline reads
        baseline_vars = set(re.findall(r"(\w+)\s*,?\s*[^\n]*:=\s*(?:os\.ReadFile|image\.Decode|jpeg\.Decode)", tsrc))
        for rv in sorted(baseline_vars):
            leaked = [c for c in render_calls if rv in c]
            test_flow.append({"baseline_var": rv,
                              "passed_to_render_call": bool(leaked),
                              "render_calls": render_calls})
    else:
        raise RulerFailure("pixel_test.go not found -- cannot verify comparison-side flow")

    render_marker_hits = [m for m in all_markers
                          if m["side"] == "RENDER" and not m["is_comment"]]

    report.update({
        "files_scanned": files,
        "file_count": len(files),
        "read_sites": all_sites,
        "read_site_count": len(all_sites),
        "baseline_marker_hits": all_markers,
        "render_side_noncomment_marker_hits": render_marker_hits,
        "render_side_violations": violations,
        "comparison_side_flagged_for_review": reviewed,
        "comparison_side_flow_check": test_flow,
        "comparison_side_leak": any(t["passed_to_render_call"] for t in test_flow) if test_flow else None,
    })

    if violations or render_marker_hits or (test_flow and any(t["passed_to_render_call"] for t in test_flow)):
        report["verdict"] = "VIOLATION"
    else:
        report["verdict"] = "CLEAN"

    emit(report, a)

    print("=" * 72)
    print(f"ref={a.ref}  files={len(files)}  read-sites={len(all_sites)}")
    print(f"ruler self-check: {len(report['ruler_self_check'])}/{len(report['ruler_self_check'])} positive controls FIRED")
    print(f"render-side violations          : {len(violations)}")
    print(f"render-side non-comment markers : {len(render_marker_hits)}")
    print(f"comparison-side leak into render: {report['comparison_side_leak']}")
    print(f"VERDICT: {report['verdict']}")
    print("=" * 72)
    return 1 if report["verdict"] == "VIOLATION" else 0


def emit(report, a):
    if a.json:
        with io.open(a.json, "w", encoding="utf-8", newline="\n") as f:
            json.dump(report, f, ensure_ascii=False, indent=2)
    if a.md:
        with io.open(a.md, "w", encoding="utf-8", newline="\n") as f:
            f.write(render_md(report))


def render_md(r):
    L = []
    L.append("# 基线是否被当作渲染输入 —— 可重跑检查报告\n")
    L.append(f"- ref: `{r['ref']}`")
    L.append(f"- 扫描文件数: **{r['file_count']}**（`{r['package']}` 下全部 `.go`）")
    L.append(f"- 读入点总数: **{r['read_site_count']}**")
    L.append(f"- **最终判定: `{r['verdict']}`**\n")
    L.append("## 尺子自检（阳性对照）\n")
    L.append("| 对照 | 探针 | 命中 | 响了? | 硬? |")
    L.append("|---|---|---|---|---|")
    for c in r["ruler_self_check"]:
        L.append("| %s | `%s` | %s | %s | %s |" % (
            c["control"], c["needle"], c["hits"],
            "是" if c["fired"] else "否（该 ref 不存在，非致命）",
            "硬" if c.get("hard") else "软"))
    L.append("\n> **硬对照**是结构性的（任何 ref 都必然成立），不成立就中止并退 2，"
             "不允许输出任何判定。**软对照**是语义性的，旧 commit 上可能本就不存在，只记录不致命。\n")
    L.append("## 逐个读入点\n")
    L.append("| 文件 | 函数 | 原语 | 参数形态 | 参数值 | 溯源 | 侧 | 判定 |")
    L.append("|---|---|---|---|---|---|---|---|")
    viol = {id(v) for v in r["render_side_violations"]}
    for s in r["read_sites"]:
        bad = id(s) in viol
        L.append("| `%s` | `%s` | `%s` | %s | `%s` | %s | %s | %s |" % (
            os.path.basename(s["file"]), s["function"], s["primitive"],
            s["arg_kind"], (s["arg_value"] or "")[:40], s.get("provenance") or "-",
            s["side"],
            "**违规**" if bad else ("比较基准" if s["side"].startswith("COMPARISON") else "资产")))
    L.append("\n## 渲染侧出现的 `baseline` 字样（须逐条判为非读入）\n")
    L.append("| 文件 | 函数 | 标记 | 是注释? | 原文 |")
    L.append("|---|---|---|---|---|")
    for m in r["render_side_noncomment_marker_hits"] or []:
        L.append("| `%s` | `%s` | `%s` | %s | `%s` |" % (
            os.path.basename(m["file"]), m["function"], m["marker"],
            "是" if m["is_comment"] else "**否**", m["line_text"][:70]))
    if not r["render_side_noncomment_marker_hits"]:
        L.append("| — | — | — | — | 渲染侧无任何非注释的 baseline 字样 |")
    L.append("\n## 比较侧数据流（证明基线没进渲染）\n")
    for t in r["comparison_side_flow_check"]:
        L.append(f"- 基线派生变量 `{t['baseline_var']}` 是否被传入渲染调用: "
                 f"**{'是（违规）' if t['passed_to_render_call'] else '否'}**")
    return "\n".join(L) + "\n"


if __name__ == "__main__":
    sys.exit(main())
