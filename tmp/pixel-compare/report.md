# Pixel Parity Report (gg vs Playwright frozen baseline)

Scenes: 16 (manifest: ..\..\src\ggrender\testdata\visual\baseline\manifest.json)

| scene | WxH | scale | format | similarity | hashOld | hashNew | bbox | passed |
|-------|-----|-------|--------|------------|---------|---------|------|--------|
| base | 1100x1290 | 1.5 | jpeg | 0.00000 | 3a82d70486f9 | 9d0aa7be26c4 | [0 0 0 0] | false |
| box | 700x370 | 1.5 | jpeg | 0.00000 | cf10f5d37fe6 | bb9eef5b01d3 | [0 0 0 0] | false |
| box-detail | 900x640 | 1.5 | jpeg | 0.00000 | 047ad057ac19 | 128ed00c9895 | [0 0 0 0] | false |
| box-summary | 900x532 | 1.5 | jpeg | 0.00000 | 6a4633e04841 | 94c04c6445b2 | [0 0 0 0] | false |
| calendar | 900x470 | 1.5 | jpeg | 0.00000 | 80f31c1cfb13 | 04ed79d5937c | [0 0 0 0] | false |
| card | 1000x600 | 1.0 | jpeg | 0.00000 | 75fea0acb6d2 | 2f30e10d82f2 | [0 0 0 0] | false |
| depot | 1275x234 | 1.5 | jpeg | 0.50739 | d09c01dbd300 | 440aeb8d77a2 | [0 0 1274 233] | false |
| enemy | 984x477 | 1.5 | jpeg | 0.99003 | 7aae46ad6d82 | ae48c97e5fa8 | [0 0 983 476] | true |
| gacha | 1500x1323 | 1.5 | jpeg | 0.98237 | b037cf55e967 | 7adbcf9e9036 | [0 0 1499 1322] | false |
| headhunt | 1049x576 | 1.0 | jpeg | 0.97478 | 879933afa3b9 | 6c90ac4b76b1 | [0 0 1048 575] | false |
| help | 990x2049 | 1.5 | jpeg | 0.94287 | 43bcdee85548 | f3a391406970 | [0 0 989 2048] | false |
| lottery | 800x370 | 1.5 | jpeg | 0.00000 | 9f5588be8860 | 9b9930e374b4 | [0 0 0 0] | false |
| missing | 700x370 | 1.5 | jpeg | 0.00000 | cc58fea684f3 | 6404460f7ac5 | [0 0 0 0] | false |
| operator | 800x700 | 1.5 | jpeg | 0.00000 | 2c2834c7e048 | 4a49f862b7dc | [0 0 0 0] | false |
| recruit | 1350x534 | 1.5 | jpeg | 0.95581 | 4a01136f4b12 | 7f502e971c84 | [0 0 1349 533] | false |
| state | 1092x510 | 1.0 | jpeg | 0.95259 | 2c3d37795bac | 0d27d7cf49d8 | [0 0 1091 509] | false |

## Failed (honest red)

- base size mismatch old 1665x918 new 1100x1290
- box size mismatch old 1050x536 new 700x370
- box-detail size mismatch old 722x279 new 900x640
- box-summary size mismatch old 1350x723 new 900x532
- calendar size mismatch old 2880x1620 new 900x470
- card size mismatch old 1280x720 new 1000x600
- depot 0.50739 <0.99 bbox=[0 0 1274 233]
- gacha 0.98237 <0.99 bbox=[0 0 1499 1322]
- headhunt 0.97478 <0.99 bbox=[0 0 1048 575]
- help 0.94287 <0.99 bbox=[0 0 989 2048]
- lottery size mismatch old 1473x1667 new 800x370
- missing size mismatch old 1050x536 new 700x370
- operator size mismatch old 1800x1200 new 800x700
- recruit 0.95581 <0.99 bbox=[0 0 1349 533]
- state 0.95259 <0.99 bbox=[0 0 1091 509]
