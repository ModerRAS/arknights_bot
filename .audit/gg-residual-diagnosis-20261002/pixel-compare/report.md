# Pixel Parity Report (gg vs Playwright frozen baseline)

Scenes: 16 (manifest: ..\..\src\ggrender\testdata\visual\baseline\manifest.json)

| scene | WxH | scale | format | similarity | hashOld | hashNew | bbox | passed |
|-------|-----|-------|--------|------------|---------|---------|------|--------|
| base | 1665x918 | 1.5 | jpeg | 0.97073 | 3a82d70486f9 | ede23e048a38 | [0 0 1664 917] | false |
| box | 1050x536 | 1.5 | jpeg | 0.82504 | cf10f5d37fe6 | 55d299b08d3a | [0 0 1049 535] | false |
| box-detail | 722x279 | 1.5 | jpeg | 0.87844 | 047ad057ac19 | 92d22b58aade | [0 0 721 278] | false |
| box-summary | 1350x723 | 1.5 | jpeg | 0.88098 | 6a4633e04841 | 45e63a36e647 | [0 0 1349 722] | false |
| calendar | 2880x1620 | 1.5 | jpeg | 0.99432 | 80f31c1cfb13 | e7939ca718d0 | [0 0 2863 1619] | true |
| card | 1280x720 | 1.0 | jpeg | 0.96015 | 75fea0acb6d2 | 96200ec144b5 | [0 0 1279 711] | false |
| depot | 1275x234 | 1.5 | jpeg | 0.92352 | d09c01dbd300 | 38f0cef6e9d5 | [0 0 1247 233] | false |
| enemy | 984x477 | 1.5 | jpeg | 0.91951 | 7aae46ad6d82 | bc7b5af135de | [0 0 983 476] | false |
| gacha | 1500x1323 | 1.5 | jpeg | 0.90251 | b037cf55e967 | e43865641c03 | [0 0 1499 1322] | false |
| headhunt | 1049x576 | 1.0 | jpeg | 0.97021 | 879933afa3b9 | 70f9b91a2e60 | [0 0 1048 575] | false |
| help | 990x2049 | 1.5 | jpeg | 0.83943 | 43bcdee85548 | e984bc7e9ffa | [0 0 989 2048] | false |
| lottery | 1473x1667 | 1.5 | jpeg | 0.98358 | 9f5588be8860 | 37a15ba596de | [0 0 1472 1666] | false |
| missing | 1050x536 | 1.5 | jpeg | 0.87614 | cc58fea684f3 | 01b88d9597b9 | [0 0 1049 535] | false |
| operator | 1800x1200 | 1.5 | jpeg | 0.76866 | 2c2834c7e048 | 9456d8f756b8 | [0 0 1799 1199] | false |
| recruit | 1350x534 | 1.5 | jpeg | 0.88603 | 4a01136f4b12 | c7b432acf01f | [0 0 1349 533] | false |
| state | 1092x510 | 1.0 | jpeg | 0.99093 | 2c3d37795bac | 79f31688bbe0 | [32 32 1079 507] | true |

## Failed (honest red)

- base 0.97073 <0.99 bbox=[0 0 1664 917]
- box 0.82504 <0.99 bbox=[0 0 1049 535]
- box-detail 0.87844 <0.99 bbox=[0 0 721 278]
- box-summary 0.88098 <0.99 bbox=[0 0 1349 722]
- card 0.96015 <0.99 bbox=[0 0 1279 711]
- depot 0.92352 <0.99 bbox=[0 0 1247 233]
- enemy 0.91951 <0.99 bbox=[0 0 983 476]
- gacha 0.90251 <0.99 bbox=[0 0 1499 1322]
- headhunt 0.97021 <0.99 bbox=[0 0 1048 575]
- help 0.83943 <0.99 bbox=[0 0 989 2048]
- lottery 0.98358 <0.99 bbox=[0 0 1472 1666]
- missing 0.87614 <0.99 bbox=[0 0 1049 535]
- operator 0.76866 <0.99 bbox=[0 0 1799 1199]
- recruit 0.88603 <0.99 bbox=[0 0 1349 533]
