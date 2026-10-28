/* Blackjack artwork: The Gold Room table, classic court deck, casino chips. (c) 2026 Ironite */
'use strict';
const ART = (() => {
const D = IR.draw, RH = IR.RH;

/* arched text helper */
function arcText(x, text, cx, cy, r, mid, font, color, spacing = 1) {
  x.save(); x.font = font; x.fillStyle = color; x.textAlign = 'center'; x.textBaseline = 'middle';
  const widths = [...text].map(ch => x.measureText(ch).width * spacing), total = widths.reduce((a, b) => a + b, 0);
  // along the bottom of the circle, left-to-right means decreasing angle
  let a = mid + total / r / 2;
  [...text].forEach((ch, i) => { const w = widths[i]; a -= w / 2 / r; x.save(); x.translate(cx + Math.cos(a) * r, cy + Math.sin(a) * r); x.rotate(a - Math.PI / 2); x.fillText(ch, 0, 0); x.restore(); a -= w / 2 / r; });
  x.restore();
}

/* ---------------- the table ---------------- */
const RAIL = { cx: 800, cy: -1180, r: 1880 }; // big arc: bottom edge of the felt
function scene(c, W, H, res) {
  c.width = W * res; c.height = H * res; const x = c.getContext('2d'); x.scale(res, res);
  // room
  x.fillStyle = D.lin(x, 0, 0, 0, H, ['#120c09', '#1c120c']); x.fillRect(0, 0, W, H);
  // felt
  x.save(); x.beginPath(); x.arc(RAIL.cx, RAIL.cy, RAIL.r, 0, Math.PI); x.lineTo(0, 0); x.closePath(); x.clip();
  x.fillStyle = D.rad(x, 800, 330, 60, 1000, ['#1f6a49', '#17573b', '#0f3f2a', '#0a2d1e']); x.fillRect(0, 0, W, H);
  // felt nap
  const R = IR.rng(77); for (let i = 0; i < 9000; i++) { x.fillStyle = R() < .5 ? `rgba(0,0,0,${R() * .07})` : `rgba(140,220,170,${R() * .035})`; x.fillRect(R() * W, R() * H, 1.2, 1.2); }
  // printed gold trim along the rail
  x.beginPath(); x.arc(RAIL.cx, RAIL.cy, RAIL.r - 28, .1, Math.PI - .1); x.strokeStyle = 'rgba(226,190,110,.55)'; x.lineWidth = 2.5; x.stroke();
  x.beginPath(); x.arc(RAIL.cx, RAIL.cy, RAIL.r - 36, .1, Math.PI - .1); x.strokeStyle = 'rgba(226,190,110,.3)'; x.lineWidth = 1; x.stroke();
  // printed table rules
  arcText(x, 'BLACKJACK PAYS 3 TO 2', 800, -492, 860, Math.PI / 2, '600 30px Lora, Georgia, serif', 'rgba(232,200,120,.82)', 1.25);
  arcText(x, 'Dealer must stand on all 17s and draw to 16', 800, -492, 824, Math.PI / 2, 'italic 19px Lora, Georgia, serif', 'rgba(232,200,120,.6)', 1.05);
  // monogram under the chip tray
  x.textAlign = 'center'; x.fillStyle = 'rgba(232,200,120,.28)'; x.font = 'italic 600 22px Lora, Georgia, serif'; x.fillText('The Gold Room', 800, 112);
  // betting circle
  ring(x, 800, 646, 54);
  // soft overhead light
  x.fillStyle = D.rad(x, 800, 360, 40, 900, ['rgba(255,230,170,.16)', 'rgba(255,230,170,.05)', 'rgba(0,0,0,.35)']); x.fillRect(0, 0, W, H);
  x.restore();
  // padded leather rail
  x.save();
  x.beginPath(); x.arc(RAIL.cx, RAIL.cy, RAIL.r + 46, 0, Math.PI); x.arc(RAIL.cx, RAIL.cy, RAIL.r, Math.PI, 0, true); x.closePath();
  x.fillStyle = D.rad(x, RAIL.cx, RAIL.cy, RAIL.r, RAIL.r + 46, ['#2a140c', '#5a2c18', '#6e3a20', '#3a1a0e', '#1a0a05']); x.fill();
  x.clip();
  for (let a = .2; a < Math.PI - .2; a += .006) { const px = RAIL.cx + Math.cos(a) * (RAIL.r + 23), py = RAIL.cy + Math.sin(a) * (RAIL.r + 23); D.ell(x, px, py, 1.2, 1.2, 'rgba(255,200,150,.05)'); }
  x.beginPath(); x.arc(RAIL.cx, RAIL.cy, RAIL.r + 16, 0, Math.PI); x.strokeStyle = 'rgba(255,220,180,.12)'; x.lineWidth = 3; x.stroke();
  x.restore();
  x.beginPath(); x.arc(RAIL.cx, RAIL.cy, RAIL.r + 1, .05, Math.PI - .05); x.strokeStyle = '#c9a24e'; x.lineWidth = 2; x.stroke();
  x.beginPath(); x.arc(RAIL.cx, RAIL.cy, RAIL.r + 47, .05, Math.PI - .05); x.strokeStyle = 'rgba(201,162,78,.6)'; x.lineWidth = 2; x.stroke();
  // wood ledge below the rail
  x.save(); x.beginPath(); x.arc(RAIL.cx, RAIL.cy, RAIL.r + 48, 0, Math.PI); x.lineTo(0, H); x.lineTo(W, H); x.closePath();
  x.fillStyle = D.lin(x, 0, 700, 0, H, ['#24140c', '#140a06']); x.fill(); x.restore();
  // dealer's chip tray
  tray(x, 560, 4, 480, 70);
  // shoe and discard tray
  shoe(x, 1330, 34); discard(x, 150, 34);
}
function ring(x, cx, cy, r) {
  D.ell(x, cx, cy, r, r, 'rgba(0,0,0,.12)');
  D.ell(x, cx, cy, r, r, null); x.strokeStyle = 'rgba(232,200,120,.75)'; x.lineWidth = 3; x.stroke();
  D.ell(x, cx, cy, r - 7, r - 7, null); x.strokeStyle = 'rgba(232,200,120,.35)'; x.lineWidth = 1.2; x.stroke();
}
function tray(x, X, Y, w, h) {
  D.rr(x, X - 6, Y - 6, w + 12, h + 12, 10); x.fillStyle = '#1b0e08'; x.fill();
  D.rr(x, X, Y, w, h, 7); x.fillStyle = D.lin(x, 0, Y, 0, Y + h, ['#3a2416', '#120a06']); x.fill();
  const cols = [['#b8241f', '#f2e7d6'], ['#1f7a46', '#f2e7d6'], ['#1d1d1d', '#e8d8b0'], ['#5b2d86', '#f2e7d6'], ['#d9a93a', '#5a3a10']];
  const slots = 10, sw = w / slots;
  for (let i = 0; i < slots; i++) { const [c1, c2] = cols[i % 5]; x.save(); x.beginPath(); x.rect(X + i * sw + 3, Y + 4, sw - 6, h - 8); x.clip();
    for (let k = 0; k < 9; k++) { const yy = Y + 6 + k * 7; D.rr(x, X + i * sw + 5, yy, sw - 10, 6, 2); x.fillStyle = c1; x.fill(); for (let s = 0; s < 3; s++) { x.fillStyle = c2; x.fillRect(X + i * sw + 8 + s * (sw - 16) / 2.5, yy + 1, 4, 4); } }
    x.restore(); D.line(x, X + i * sw, Y + 3, X + i * sw, Y + h - 3, 'rgba(0,0,0,.6)', 2); }
  x.fillStyle = D.lin(x, 0, Y, 0, Y + 20, ['rgba(255,255,255,.12)', 'rgba(255,255,255,0)']); x.fillRect(X, Y, w, 20);
}
function shoe(x, X, Y) {
  x.save(); x.translate(X, Y); x.rotate(-.06);
  D.rr(x, 0, 0, 170, 120, 10); x.fillStyle = D.lin(x, 0, 0, 0, 120, ['#2d2a28', '#0e0d0c']); x.fill();
  D.rr(x, 8, 8, 154, 104, 7); x.fillStyle = D.lin(x, 0, 0, 170, 0, ['#3b2414', '#5c3a20', '#3b2414']); x.fill();
  D.rr(x, 14, 20, 90, 80, 5); x.fillStyle = '#5e1717'; x.fill();
  for (let i = 0; i < 6; i++) D.line(x, 18 + i * 14, 24, 18 + i * 14, 96, 'rgba(255,215,140,.25)', 1);
  D.rr(x, 110, 30, 46, 60, 6); x.fillStyle = '#111'; x.fill(); x.strokeStyle = '#c9a24e'; x.lineWidth = 2; x.stroke();
  x.fillStyle = '#c9a24e'; x.font = '600 11px Lora, serif'; x.textAlign = 'center'; x.fillText('SHOE', 85, 115);
  x.restore();
}
function discard(x, X, Y) {
  x.save(); x.translate(X, Y); x.rotate(.05);
  D.rr(x, 0, 0, 140, 150, 9); x.fillStyle = 'rgba(10,8,6,.55)'; x.fill(); x.strokeStyle = 'rgba(201,162,78,.5)'; x.lineWidth = 2; x.stroke();
  x.fillStyle = 'rgba(201,162,78,.5)'; x.font = '600 11px Lora, serif'; x.textAlign = 'center'; x.fillText('DISCARD', 70, 166);
  x.restore();
}

/* ---------------- chips ---------------- */
const CHIP = { 5: ['#b3221d', '#f4ead8', '#ffffff'], 25: ['#1e7a46', '#f4ead8', '#ffffff'], 100: ['#1b1b1b', '#e5d3a5', '#e5d3a5'], 500: ['#5a2a86', '#f4ead8', '#ffffff'], 1000: ['#d6a536', '#5a3a10', '#3e2808'] };
const chipCache = {};
function chip(v, size = 72, res = 2) {
  const key = v + '|' + size + '|' + res; if (chipCache[key]) return chipCache[key];
  const [c, x] = D.canvas(size * res, size * res); x.scale(res, res); const r = size / 2, [base, ins, txt] = CHIP[v];
  D.ell(x, r, r, r - 1, r - 1, base);
  x.save(); x.beginPath(); x.arc(r, r, r - 1, 0, Math.PI * 2); x.clip();
  for (let i = 0; i < 6; i++) { const a = i / 6 * Math.PI * 2; x.save(); x.translate(r, r); x.rotate(a); x.fillStyle = ins; x.fillRect(-r * .13, -r, r * .26, r * .3); x.restore(); }
  x.restore();
  D.ell(x, r, r, r * .68, r * .68, null); x.setLineDash([r * .09, r * .07]); x.strokeStyle = ins; x.lineWidth = r * .045; x.stroke(); x.setLineDash([]);
  D.ell(x, r, r, r * .6, r * .6, D.rad(x, r * .85, r * .8, r * .05, r * .6, [lighten(base), base]));
  x.fillStyle = txt; x.font = `700 ${v >= 1000 ? r * .42 : r * .5}px Lora, Georgia, serif`; x.textAlign = 'center'; x.textBaseline = 'middle'; x.fillText(v >= 1000 ? '1K' : v, r, r + r * .03);
  D.ell(x, r, r, r - 1.5, r - 1.5, null); x.strokeStyle = 'rgba(0,0,0,.45)'; x.lineWidth = 1.5; x.stroke();
  D.ell(x, r, r, r - 1.5, r - 1.5, D.rad(x, r * .7, r * .55, 0, r, ['rgba(255,255,255,.22)', 'rgba(255,255,255,0)', 'rgba(0,0,0,.18)']));
  return (chipCache[key] = c.toDataURL('image/png'));
}
function lighten(hex) { const n = parseInt(hex.slice(1), 16); const f = c => Math.min(255, Math.round(c + (255 - c) * .22)); return `rgb(${f(n >> 16)},${f((n >> 8) & 255)},${f(n & 255)})`; }
const DENOMS = [1000, 500, 100, 25, 5];
function breakdown(amount) { const out = []; let a = amount; for (const d of DENOMS) while (a >= d && out.length < 40) { out.push(d); a -= d; } return out; }

/* ---------------- the deck ---------------- */
IR.hats.beret = (x, h) => {
  D.ell(x, .52, .64, .21, .09, h.color, -.12);
  D.ell(x, .52, .62, .17, .07, D.rad(x, .47, .6, .01, .2, ['rgba(255,255,255,.2)', 'rgba(0,0,0,0)']), -.12);
  x.beginPath(); x.moveTo(.62, .6); x.quadraticCurveTo(.8, .4, .9, .3); x.quadraticCurveTo(.82, .46, .66, .62); x.closePath(); x.fillStyle = h.feather; x.fill();
  for (let i = 1; i < 9; i++) { const t = i / 9; D.line(x, .64 + t * .25, .6 - t * .29, .7 + t * .25, .6 - t * .26, 'rgba(255,255,255,.5)', .004); }
  x.fillStyle = h.band; x.fillRect(.34, .66, .32, .03); D.ell(x, .4, .675, .016, .016, '#e8c66a');
};
const ROYAL = ['#1f3a5f', '#7d1c22', '#7d1c22', '#1f3a5f'];
function damask(c1, c2) {
  return x => {
    x.fillStyle = D.lin(x, 0, 0, 0, RH, [c1, c2]); x.fillRect(0, 0, 1, RH);
    x.fillStyle = 'rgba(232,200,120,.13)';
    for (let i = 0; i < 6; i++) for (let j = 0; j < 9; j++) { const cx = (i + (j % 2) * .5) / 5, cy = j / 5.4; x.save(); x.translate(cx, cy); x.beginPath(); x.moveTo(0, -.06); x.quadraticCurveTo(.05, 0, 0, .06); x.quadraticCurveTo(-.05, 0, 0, -.06); x.fill(); x.restore(); }
  };
}
const ermine = x => { x.fillStyle = '#f6f1e6'; x.beginPath(); x.moveTo(.08, RH); x.lineTo(.12, 1.2); x.quadraticCurveTo(.3, 1.1, .5, 1.13); x.quadraticCurveTo(.7, 1.1, .88, 1.2); x.lineTo(.92, RH); x.closePath(); x.fill(); for (let i = 0; i < 14; i++) { const px = .15 + (i * .37) % .7, py = 1.2 + (i * .23) % .38; D.ell(x, px, py, .012, .02, '#1a1a1a'); } };
const COURTS = {
  13: (x, s) => IR.portrait(x, { bg: damask(s % 3 ? '#5a1a1e' : '#1d2f4a', s % 3 ? '#2e0c0f' : '#0f1828'), coat: [ROYAL[s], '#121212'], hair: '#8a7a6a', beard: '#9c8c78', beardLen: .04, mustache: '#8d7d6a', brows: '#6e604f', skin: '#e6b892',
    hat: { type: 'crown', color: '#dcb24c', band: ROYAL[s], jewel: s % 3 ? '#2f6bd1' : '#c4302b' },
    body: x => { ermine(x); x.fillStyle = ROYAL[s]; x.fillRect(.36, 1.2, .28, .45); D.ell(x, .5, 1.3, .03, .03, '#e8c66a'); },
    front: x => { D.line(x, .82, RH, .86, 1.0, '#c9a24e', .03); D.ell(x, .86, .98, .04, .04, '#e8c66a'); D.ell(x, .86, .98, .02, .02, '#c4302b'); } }),
  12: (x, s) => IR.portrait(x, { bg: damask(s % 3 ? '#5a1a1e' : '#1d2f4a', s % 3 ? '#2e0c0f' : '#0f1828'), coat: [ROYAL[s], '#141414'], hair: '#3b2416', longHair: true, lips: '#b8322e', lashes: true, blush: true, skin: '#f0c9a6', earring: '#f3e7c9',
    hat: { type: 'tiara', color: '#e3c45e', jewel: s % 3 ? '#c22' : '#2f6bd1' },
    collar: x => { for (let i = 0; i < 9; i++) D.ell(x, .4 + i * .025, 1.08 + Math.sin(i / 8 * Math.PI) * .04, .01, .01, '#f6efe0'); },
    front: x => { D.line(x, .72, RH, .76, 1.18, '#3d6b3a', .012); for (let i = 0; i < 6; i++) { const a = i / 6 * Math.PI * 2; D.ell(x, .76 + Math.cos(a) * .03, 1.15 + Math.sin(a) * .03, .03, .022, '#b3222a', a); } D.ell(x, .76, 1.15, .022, .022, '#d93a3a'); } }),
  11: (x, s) => IR.portrait(x, { bg: damask(s % 3 ? '#5a1a1e' : '#1d2f4a', s % 3 ? '#2e0c0f' : '#0f1828'), coat: [ROYAL[s], '#141414'], hair: '#b07a3a', curls: false, blush: true, skin: '#efc29c',
    hat: { type: 'beret', color: '#1a1a1a', feather: '#f3ead6', band: ROYAL[s] },
    body: x => { x.fillStyle = '#e8c66a'; for (let i = 0; i < 4; i++) D.ell(x, .5, 1.24 + i * .1, .018, .018, '#e8c66a'); D.line(x, .3, 1.12, .7, 1.12, '#e8c66a', .02); },
    collar: x => { D.poly(x, [[.36, 1.06], [.5, 1.16], [.64, 1.06], [.62, 1.12], [.5, 1.2], [.38, 1.12]], '#f6efe0'); } }),
};
const theme = {
  paper: ['#fffdf7', '#f1ebdc'], vignette: 'rgba(90,70,30,.1)', border: '#8a6a32', radius: .07,
  index: { font: "Lora, Georgia, serif", red: '#a51e22', black: '#141414' },
  pip: { red: ['#c8262a', '#7e1014'], black: ['#2b2b2b', '#050505'] },
  frame: ['#3a2a10', '#f6dc8a', '#a8822e', '#f2d37a'],
  court: (x, s, r) => COURTS[r](x, s),
  ace(x, s, cw, ch) {
    const cx = cw / 2, cy = ch / 2;
    if (s === 0) { // ornate ace of spades
      for (let i = 0; i < 24; i++) { const a = i / 24 * Math.PI * 2; D.ell(x, cx + Math.cos(a) * cw * .33, cy + Math.sin(a) * cw * .33, cw * .014, cw * .014, i % 2 ? '#c9a24e' : '#e8c66a'); }
      D.ell(x, cx, cy, cw * .3, cw * .3, null); x.strokeStyle = '#c9a24e'; x.lineWidth = cw * .006; x.stroke();
      IR.drawSuit(x, 0, cx, cy, cw * .42, false, { black: ['#2b2b2b', '#050505'], edge: '#c9a24e' });
      x.fillStyle = '#8a6a32'; x.font = `italic 600 ${cw * .07}px Lora, serif`; x.textAlign = 'center'; x.fillText('The Gold Room', cx, ch * .87);
    } else IR.drawSuit(x, s, cx, cy, cw * .36, false, theme.pip);
  },
  backBorder: '#1c0606',
  back(x, cw, ch) {
    x.fillStyle = D.lin(x, 0, 0, cw, ch, ['#6c1520', '#3e0910']); x.fillRect(0, 0, cw, ch);
    const m = cw * .065;
    x.save(); x.beginPath(); x.rect(m, m, cw - 2 * m, ch - 2 * m); x.clip();
    const st = cw * .09; x.strokeStyle = 'rgba(232,200,120,.38)'; x.lineWidth = cw * .006;
    for (let i = -ch; i < cw + ch; i += st) { D.line(x, i, 0, i + ch, ch, 'rgba(232,200,120,.3)', cw * .005); D.line(x, i, ch, i + ch, 0, 'rgba(232,200,120,.3)', cw * .005); }
    for (let i = 0; i < 14; i++) for (let j = 0; j < 20; j++) D.ell(x, i * st, j * st * .5 + (i % 2) * st * .25, cw * .008, cw * .008, 'rgba(232,200,120,.45)');
    x.restore();
    x.strokeStyle = '#d9b45c'; x.lineWidth = cw * .016; x.strokeRect(m, m, cw - 2 * m, ch - 2 * m);
    x.strokeStyle = 'rgba(217,180,92,.55)'; x.lineWidth = cw * .006; x.strokeRect(m + cw * .028, m + cw * .028, cw - 2 * m - cw * .056, ch - 2 * m - cw * .056);
    const cx = cw / 2, cy = ch / 2, r = cw * .2;
    D.ell(x, cx, cy, r * 1.25, r * 1.25, '#4a0c14'); D.ell(x, cx, cy, r * 1.25, r * 1.25, null); x.strokeStyle = '#d9b45c'; x.lineWidth = cw * .012; x.stroke();
    for (let i = 0; i < 4; i++) { const a = i * Math.PI / 2 - Math.PI / 2; IR.drawSuit(x, i, cx + Math.cos(a) * r * .62, cy + Math.sin(a) * r * .62, r * .55, false, { red: ['#f0d27c', '#b8892e'], black: ['#f0d27c', '#b8892e'] }); }
    D.ell(x, cx, cy, r * .2, r * .2, '#e8c66a');
  },
};

return { scene, theme, chip, breakdown, DENOMS, RAIL };
})();
