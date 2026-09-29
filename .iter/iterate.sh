#!/bin/bash
# hardened iterate loop: render -> validate -> score. Usage: iterate.sh
set -o pipefail
cd /c/WorkSpace/Golang/arknights_bot-satori-w3-help-boxes/src
OUT=utils/media/testdata/visual/final
BEFORE=$(python -c "import json,hashlib;print(hashlib.sha256(open('$OUT/new-render-report.json','rb').read()).hexdigest()[:12])" 2>/dev/null)
/tmp/visual-final.exe -spec-dir ../specs -out $OUT -baseline utils/media/testdata/visual/baseline/manifest.json -resource-manifest utils/media/testdata/visual/baseline/resource-manifest.json -clock-erratum utils/media/testdata/visual/capture-clock-erratum.json -legacy-script cmd/visual-final/legacy-capture.mjs > /tmp/vf-stdout.log 2> /tmp/vf-stderr.log
RC=$?
AFTER=$(python -c "import json,hashlib;print(hashlib.sha256(open('$OUT/new-render-report.json','rb').read()).hexdigest()[:12])" 2>/dev/null)
echo "exit=$RC report-changed=$([ "$BEFORE" != "$AFTER" ] && echo yes || echo NO)"
if [ "$BEFORE" = "$AFTER" ]; then echo "FATAL: report not regenerated"; tail -5 /tmp/vf-stderr.log; exit 1; fi
python - <<'EOF'
import json, sys
d = json.load(open('utils/media/testdata/visual/final/new-render-report.json', encoding='utf-8'))
bad = [r for r in d['results'] if r.get('error')]
if bad or d.get('failed'):
    print('FATAL: render errors:', [(r['id'], r['error'][:80]) for r in bad], 'failed=', d['failed'])
    sys.exit(1)
r = json.load(open('utils/media/testdata/visual/final/report.json', encoding='utf-8'))
for p in r['pages']:
    if p['id'] in ('help', 'box-detail', 'box-summary'):
        print(p['id'], 'raw', round(p['rawSimilarity'], 4), 'aligned', round(p['alignedSimilarity'], 4), 'offset', p['globalOffset'])
EOF
