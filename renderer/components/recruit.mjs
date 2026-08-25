import { h } from '../lib/h.mjs';
const fallback='assets/common/amiya.png';
// Geometry tuned against frozen Playwright baseline (Recruit.tmpl HTML table):
// header 28; tags col 182.7 (auto table layout); avatar pitch 103.5x106.7
// (inline whitespace + line-box descent); chips 28px h, margin 4, centered.
const SHADOW='0 3px 1px -2px rgba(0,0,0,.2),0 2px 2px rgba(0,0,0,.14),0 1px 5px rgba(0,0,0,.12)';
export default async function render(props,{image}) { const rows=await Promise.all((props??[]).map(async row=>({row,operators:await Promise.all((row.operators??[]).map(async item=>({item,avatar:await image(item.avatar,fallback),profession:await image(`assets/box/${item.profession}.png`,fallback),rarity:await image(`assets/box/Rarity_${item.rarity}.png`,fallback)})))}))); return h('div',{style:{width:900,height:356,display:'flex',flexDirection:'column',fontFamily:'NotoSansHans',backgroundColor:'#fff',color:'#000'}},
h('div',{style:{height:28,display:'flex',alignItems:'center',boxShadow:SHADOW,fontSize:16,fontWeight:700,lineHeight:'28px'}},
h('div',{style:{width:182.7,display:'flex',alignItems:'center',justifyContent:'center'}},'标签'),
h('div',{style:{flex:1,display:'flex',alignItems:'center',justifyContent:'center'}},'干员')),
rows.map(({row,operators},idx)=>h('div',{style:{display:'flex',boxShadow:SHADOW}},
h('div',{style:{width:182.7,display:'flex',alignItems:'center',justifyContent:'center',paddingTop:1.8}},
(row.tags??[]).map(tag=>h('div',{style:{display:'flex',alignItems:'center',height:28,padding:'0 8px',margin:4,backgroundColor:'#313131',boxShadow:'0 3px 5px gray',color:'#fff',fontSize:16}},tag))),
h('div',{style:{flex:1,display:'flex',flexWrap:'wrap',alignContent:'flex-start',paddingLeft:3.97,paddingTop:idx===0?2.67:2.7}},
operators.map(({item,avatar,profession,rarity})=>h('div',{style:{width:103.5,height:idx===0?106.7:106.0,position:'relative',display:'flex'}},
h('img',{src:avatar,width:100,height:100}),
h('img',{src:profession,width:30,height:30,style:{position:'absolute',top:0,left:0}}),
h('img',{src:rarity,height:20,style:{position:'absolute',top:0,left:30}}))))))); }
