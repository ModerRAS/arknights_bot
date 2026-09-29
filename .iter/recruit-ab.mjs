import { readFileSync, writeFileSync } from 'node:fs';
const ROOT = 'C:/WorkSpace/Golang/arknights_bot-satori-w3-help-boxes';
// render current (baseline restore) -> A
writeFileSync(ROOT + '/renderer/components/recruit.mjs', readFileSync(ROOT + '/.iter/recruit.bak.mjs', 'utf8'));
