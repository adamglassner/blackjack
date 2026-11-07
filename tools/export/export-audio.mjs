// Renders Blackjack's original synthesized music and sound effects (the web
// version's Web Audio code in web/core.js and web/music.js) to WAV files with
// an OfflineAudioContext in headless Chromium. Then run encode.sh.
//   node tools/export/export-audio.mjs <outdir>
import { chromium } from 'playwright';
import fs from 'fs';
import path from 'path';
import http from 'http';
import { fileURLToPath } from 'url';

const here = path.dirname(fileURLToPath(import.meta.url));
const out = process.argv[2] || 'audio-wav';
fs.mkdirSync(out, { recursive: true });
const RATE = 44100;
const srv = http.createServer((q, r) => { const f = path.join(here, decodeURIComponent(q.url.split('?')[0])); fs.readFile(f, (e, b) => { r.writeHead(e ? 404 : 200); r.end(b); }); }).listen(0);
const port = srv.address().port;

const RENDER = async ([what, RATE]) => {
  const tunes = MUSIC.tunes();
  const dur = what.startsWith('tune') ? (() => { const tn = tunes[+what.slice(4)]; return .3 + tn.len * 60 / tn.bpm + 4; })()
    : { card: .4, deal: .4, flip: .3, shuffle: 2.4, chip1: .5, chip2: .6, chip3: .7, chip6: .9, win: 2.8, blackjack: 3.6, push: 2.2, lose: 2.2, bust: 2.4 }[what.replace(/-\d+$/, '')];
  const ctx = new OfflineAudioContext(2, Math.ceil(dur * RATE), RATE);
  window.AudioContext = function () { return ctx; };
  const a = new IR.Audio({ music: true, sfx: true, vol: 100 });
  a.ensure();
  a.music.gain.value = .95;           // volume is applied by the game
  a.sfxBus.gain.value = .85;
  a.sfx = function (fn) { fn.call(this, .02, this.sfxBus); };
  a.setTunes(tunes, MUSIC.instruments);
  const sfx = MUSIC.sfx(a);
  let name = '';
  if (what.startsWith('tune')) {
    const tn = tunes[+what.slice(4)], spb = 60 / tn.bpm; name = tn.name;
    for (const e of tn.events) {
      let bt = e.t; if (tn.swing) { const fr = bt % 1; if (Math.abs(fr - .5) < .01) bt += tn.swing * .5; }
      a.play(e, .3 + bt * spb + (Math.random() - .5) * .008, spb);
    }
  } else {
    const k = what.replace(/-\d+$/, '');
    ({ card: () => a.sCard(), deal: () => a.sDeal(), flip: () => a.sFlip(), shuffle: () => a.sShuffle(),
       chip1: () => a.sChip(1), chip2: () => a.sChip(2), chip3: () => a.sChip(3), chip6: () => a.sChip(6),
       win: () => sfx.win(), blackjack: () => sfx.blackjack(), push: () => sfx.push(), lose: () => sfx.lose(), bust: () => sfx.bust() })[k]();
  }
  const buf = await ctx.startRendering();
  const L = buf.getChannelData(0), R = buf.getChannelData(1), n = L.length;
  const k = what.startsWith('tune') ? .7 : 1; // headroom: the band peaks just over full scale
  const pcm = new Int16Array(n * 2);
  for (let i = 0; i < n; i++) { pcm[2 * i] = Math.max(-1, Math.min(1, L[i] * k)) * 32767; pcm[2 * i + 1] = Math.max(-1, Math.min(1, R[i] * k)) * 32767; }
  let s = ''; const u8 = new Uint8Array(pcm.buffer);
  for (let i = 0; i < u8.length; i += 32768) s += String.fromCharCode.apply(null, u8.subarray(i, i + 32768));
  return { name, b64: btoa(s) };
};

function wav(pcm) {
  const h = Buffer.alloc(44);
  h.write('RIFF', 0); h.writeUInt32LE(36 + pcm.length, 4); h.write('WAVE', 8); h.write('fmt ', 12);
  h.writeUInt32LE(16, 16); h.writeUInt16LE(1, 20); h.writeUInt16LE(2, 22); h.writeUInt32LE(RATE, 24);
  h.writeUInt32LE(RATE * 4, 28); h.writeUInt16LE(4, 32); h.writeUInt16LE(16, 34); h.write('data', 36); h.writeUInt32LE(pcm.length, 40);
  return Buffer.concat([h, pcm]);
}

const jobs = ['tune0', 'tune1', 'tune2'];
for (const [k, n] of [['card', 4], ['deal', 4], ['flip', 3], ['shuffle', 2], ['chip1', 3], ['chip2', 3], ['chip3', 3], ['chip6', 2],
  ['win', 1], ['blackjack', 1], ['push', 1], ['lose', 1], ['bust', 1]]) for (let i = 0; i < n; i++) jobs.push(`${k}-${i}`);
const b = await chromium.launch();
const p = await b.newPage();
await p.goto(`http://127.0.0.1:${port}/audio.html`);
const names = [];
for (const j of jobs) {
  const r = await p.evaluate(RENDER, [j, RATE]);
  const file = j.startsWith('tune') ? `tune-${j.slice(4)}` : j;
  fs.writeFileSync(path.join(out, file + '.wav'), wav(Buffer.from(r.b64, 'base64')));
  if (r.name) names.push(r.name);
  console.log(file);
}
fs.writeFileSync(path.join(out, 'tunes.json'), JSON.stringify(names, null, 1) + '\n');
await b.close(); srv.close();
