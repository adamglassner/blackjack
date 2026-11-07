/* Blackjack music: three original, low-key lounge tunes played live by the
   game (electric piano, vibraphone, walking upright bass, brushes and ride
   cymbal). (c) 2026 Ironite */
'use strict';
const MUSIC = (() => {
const parseMel = (s, at) => { const ev = []; let t = at; s.split(/\s+/).filter(x => x && x !== '|').forEach(tok => { const [n, d] = tok.split(':'); if (n !== 'R') ev.push({ n, t, d: +d }); t += +d; }); return ev; };

function arrange(tune) {
  const ev = [];
  const bars = list => list.flatMap(b => { const c = b.split(' '); return c.length === 1 ? [[c[0], 4]] : c.map(x => [x, 4 / c.length]); });
  // rhythm section over a list of chords
  const band = (list, at, opt = {}) => {
    const seq = bars(list); let t = at;
    seq.forEach(([sym, len], k) => {
      const c = IR.chord(sym), next = IR.chord((seq[k + 1] || seq[0])[0]);
      const root = IR.pcAbove(c.root, 36), nroot = IR.pcAbove(next.root, 36);
      // walking bass
      const third = root + c.iv[1], fifth = root + c.iv[2];
      const walk = len === 4 ? [root, Math.random() < .5 ? third : fifth, fifth + (Math.random() < .4 ? 2 : 0), nroot + (Math.random() < .5 ? 1 : -1)]
        : [root, nroot + (nroot > root ? -1 : 1)];
      walk.forEach((m, i) => ev.push({ t: t + i, i: 'bass', m: m > 52 ? m - 12 : m, d: .95, v: i ? .8 : 1 }));
      // rootless electric piano voicing, Charleston rhythm
      const voice = c.iv.slice(1).concat(c.iv.length < 4 ? [c.iv.includes(10) ? 14 : 9] : []).map(i => IR.pcAbove((c.root + i) % 12, 56)).sort((a, b) => a - b);
      if (!opt.sparse) {
        ev.push({ t, i: 'comp', ms: voice, d: 1.2, v: .9 });
        if (len === 4) ev.push({ t: t + 1.5, i: 'comp', ms: voice, d: .6, v: .7 });
      } else ev.push({ t, i: 'comp', ms: voice, d: len * .95, v: .8 });
      // drums: ride "ding, ding-a" + brushes on 2 and 4
      for (let b = 0; b < len; b++) {
        ev.push({ t: t + b, i: 'ride', v: b % 2 ? .75 : 1 });
        if (b % 2 === 1) { ev.push({ t: t + b + .5, i: 'ride', v: .5 }); ev.push({ t: t + b, i: 'brush', v: 1 }); }
        else ev.push({ t: t + b, i: 'sweep', v: .7 });
      }
      t += len;
    });
    return t;
  };
  const lead = (mel, at, inst, oct = 0, v = 1) => parseMel(mel, at).forEach(e => ev.push({ t: e.t, i: inst, m: IR.n2m(e.n) + oct, d: e.d, v }));
  let t = band(tune.intro, 0, { sparse: true });
  for (const [sec, inst, oct] of tune.form) { lead(tune[sec].mel, t, inst, oct); t = band(tune[sec].ch, t); }
  t = band(tune.tag, t, { sparse: true });
  ev.push({ t: t - 4, i: 'comp', ms: tune.last, d: 4, v: .9 });
  ev.sort((a, b) => a.t - b.t);
  return { name: tune.name, bpm: tune.bpm, swing: tune.swing, events: ev, len: t, gap: 3 };
}

const TUNES = [
  { name: 'Velvet Hour', bpm: 88, swing: .32,
    intro: ['Gm7', 'C7', 'Fmaj7', 'Gm7 C7'],
    A: { ch: ['Fmaj7', 'Dm7', 'Gm7', 'C7', 'Am7', 'D7', 'Gm7', 'C7'],
      mel: 'A4:1.5 C5:.5 E5:1 D5:1 | C5:2 A4:1 F4:1 | Bb4:1.5 D5:.5 F5:1 E5:1 | E5:3 R:1 | E5:1.5 G5:.5 E5:1 C5:1 | F#5:1 E5:1 D5:1 C5:1 | Bb4:1 D5:1 G5:1 F5:1 | E5:3 R:1' },
    B: { ch: ['Bbmaj7', 'Bbm6', 'Am7', 'D7', 'Gm7', 'C7', 'Fmaj7', 'Gm7 C7'],
      mel: 'D5:1 F5:1 A5:1.5 G5:.5 | F5:1 Db5:1 Bb4:2 | C5:1 E5:1 G5:1 A5:1 | F#5:3 R:1 | G5:1.5 F5:.5 D5:1 Bb4:1 | E5:1 G5:1 Bb5:1 A5:1 | A5:4 | R:2 C5:1 E5:1' },
    form: [['A', 'keys', 0], ['B', 'keys', 0], ['A', 'vibes', 0], ['B', 'vibes', 0]],
    tag: ['Gm7', 'C7', 'Fmaj7', 'Fmaj7'], last: [57, 60, 64, 69] },
  { name: 'Last Call at the Gold Room', bpm: 100, swing: .34,
    intro: ['Cm7', 'F7', 'Bbmaj7', 'Cm7 F7'],
    A: { ch: ['Bbmaj7', 'G7', 'Cm7', 'F7', 'Dm7', 'G7', 'Cm7 F7', 'Bbmaj7'],
      mel: 'D5:1 F5:.5 D5:.5 Bb4:2 | B4:1 D5:1 F5:1 Ab5:1 | G5:1.5 Eb5:.5 C5:2 | A4:1 C5:1 Eb5:1 D5:1 | F5:1 A5:1 F5:1 D5:1 | B4:2 D5:1 F5:1 | Eb5:1 G5:1 F5:1 Eb5:1 | D5:3 R:1' },
    B: { ch: ['Ebmaj7', 'Ebm6', 'Dm7', 'G7', 'Cm7', 'F7', 'Bbmaj7', 'Cm7 F7'],
      mel: 'G5:1 Bb5:1 G5:1 Eb5:1 | Gb5:2 Eb5:1 C5:1 | F5:1.5 D5:.5 A4:2 | B4:1 D5:1 G5:1 F5:1 | Eb5:1 G5:1 Bb5:1 G5:1 | A5:1 F5:1 Eb5:1 C5:1 | D5:4 | R:2 C5:1 Eb5:1' },
    form: [['A', 'vibes', 0], ['B', 'vibes', 0], ['A', 'keys', 0], ['B', 'vibes', 0]],
    tag: ['Cm7', 'F7', 'Bbmaj7', 'Bbmaj7'], last: [58, 62, 65, 69] },
  { name: 'Midnight Felt', bpm: 70, swing: .22,
    intro: ['Dm9', 'Gm9', 'Em7b5', 'A7'],
    A: { ch: ['Dm9', 'Em7b5 A7', 'Dm9', 'Gm9 C7', 'Fmaj7', 'Bbmaj7', 'Em7b5', 'A7'],
      mel: 'A4:1 D5:1 E5:1 F5:1 | G5:1.5 F5:.5 E5:1 C#5:1 | D5:3 A4:1 | Bb4:1 D5:1 F5:1 E5:1 | C5:2 E5:1 A5:1 | A5:1.5 G5:.5 F5:1 D5:1 | E5:1 G5:1 Bb5:1 G5:1 | E5:3 R:1' },
    B: { ch: ['Gm9', 'C7', 'Fmaj7', 'Bbmaj7', 'Em7b5', 'A7', 'Dm9', 'Em7b5 A7'],
      mel: 'D5:1 F5:1 A5:1 G5:1 | E5:1.5 G5:.5 Bb5:2 | A5:1 G5:1 F5:1 E5:1 | D5:2 F5:2 | G5:1.5 F5:.5 E5:1 D5:1 | C#5:1 E5:1 G5:1 A5:1 | F5:4 | R:2 A4:1 C#5:1' },
    form: [['A', 'vibes', 0], ['B', 'keys', 0], ['A', 'vibes', 0], ['B', 'vibes', 0]],
    tag: ['Gm9', 'A7', 'Dm9', 'Dm9'], last: [53, 57, 60, 64] },
];

const instruments = {
  bass(t, e, d, v) { this.bass(t, e.m, .75 * v, Math.min(d, .9)); },
  comp(t, e, d, v) { e.ms.forEach((m, i) => this.epiano(t + i * .006, m, .42 * v, d)); },
  keys(t, e, d, v) { this.epiano(t, e.m, .95 * v, d * .92); this.epiano(t, e.m + 12, .18 * v, d * .5); },
  vibes(t, e, d, v) { this.vibes(t, e.m, .95 * v, d); },
  ride(t, e, d, v) { this.ride(t, .55 * v); },
  brush(t, e, d, v) { this.brush(t, .9 * v, false); },
  sweep(t, e, d, v) { this.brush(t, .55 * v, true); },
};

function sfx(audio) {
  return {
    win() { audio.sChime([72, 76, 79, 84], .08, .9); },
    blackjack() { audio.sfx(function (t, o) { [72, 76, 79, 84, 88, 91].forEach((m, i) => this.bell(t + i * .07, m, .9, o, 2.2)); this.vibes(t + .45, 79, .6, 1, o); }); },
    push() { audio.sfx(function (t, o) { this.vibes(t, 69, .6, .3, o); this.vibes(t + .14, 69, .5, .3, o); }); },
    lose() { audio.sfx(function (t, o) { this.vibes(t, 64, .55, .2, o); this.vibes(t + .15, 60, .55, .35, o); }); },
    bust() { audio.sfx(function (t, o) { [66, 63, 59].forEach((m, i) => this.vibes(t + i * .12, m, .55, .25, o)); }); },
  };
}

return { tunes: () => TUNES.map(arrange), instruments, sfx };
})();
