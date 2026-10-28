/* Blackjack - game logic. Play money only. (c) 2026 Ironite */
'use strict';
(async () => {
const VERSION = '1.0.0';
const $ = IR.$;
const START = 1000, MIN_BET = 5, MAX_BET = 5000, MAX_HANDS = 4;
const save = new IR.Save('blackjack.save', {
  settings: { music: true, sfx: true, vol: 60, decks: 1, pace: 'steady' },
  balance: START, atRisk: 0, lastBet: 25,
  stats: { hands: 0, won: 0, lost: 0, pushed: 0, blackjacks: 0, biggestWin: 0, peak: START, restocks: 0 },
});
await save.load();
const S = save.data.settings, ST = save.data.stats, SD = save.data;
const money = n => (Math.round(n * 100) % 100) ? n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : IR.fmt(n);
const deckName = n => ({ 1: 'Single Deck', 2: 'Double Deck', 6: '6-Deck Shoe' })[n] || n + ' Decks';

// A bet left on the table when the game was closed mid-hand is returned.
let notice = '';
if (SD.atRisk > 0) { SD.balance += SD.atRisk; notice = `Your unfinished hand from last time was returned to you: ${money(SD.atRisk)} chips.`; SD.atRisk = 0; save.write(true); }

/* ---------- stage, art, audio ---------- */
const stage = $('#stage'), W = 1600, H = 844, CW = 112, CH = 157;
let res = 0;
const art = new IR.CardArt(ART.theme, CW, CH, 2);
const cards = new IR.Cards($('#cards'), art);
const fx = new IR.Particles($('#fx'));
IR.Stage.init(stage, s => {
  const r = IR.clamp(Math.ceil(s * (devicePixelRatio || 1) * 2) / 2, 1.5, 3.5);
  if (r > res + .4) { res = r; ART.scene($('#scene'), W, H, res); art.setRes(Math.max(2, res)); }
  fx.resize(W, H, Math.min(2, s * (devicePixelRatio || 1)));
});
const audio = new IR.Audio(S);
audio.setTunes(MUSIC.tunes(), MUSIC.instruments);
const sfx = MUSIC.sfx(audio);
IR.wireAudioControls(audio, save);
IR.on('tune', n => { const e = $('#nowPlaying'); e.textContent = '♫ ' + n; e.classList.add('show'); setTimeout(() => e.classList.remove('show'), 5000); });
$('#ver').textContent = 'v' + VERSION + ' · © 2026 Ironite · play money only';
const pace = () => ({ relaxed: 1.35, steady: 1, quick: .6 })[S.pace] || 1;
const wait = ms => new Promise(r => setTimeout(r, ms * pace() / (IR.speed || 1)));

/* ---------- rules ---------- */
const val = id => { const r = IR.rank(id); return r === 1 ? 11 : Math.min(10, r); };
function total(ids) { let t = 0, aces = 0; for (const id of ids) { const r = IR.rank(id); if (r === 1) { aces++; t += 11; } else t += Math.min(10, r); } while (t > 21 && aces) { t -= 10; aces--; } return { t, soft: aces > 0 }; }
const isBJ = h => h.cards.length === 2 && total(h.cards).t === 21 && !h.split;
const cutFor = d => d === 1 ? 18 : d === 2 ? 32 : Math.round(d * 52 * .25);

/* ---------- state ---------- */
let shoe = [], shoeDecks = 0, discards = [];
let hands = [], dealer = { cards: [], els: [] }, active = 0, phase = 'title', bet = 0, busy = false, holeUp = false;
const SHOE = { x: 1388, y: 62 }, DISC = { x: 166, y: 30 };
const newHand = (b, split = false) => ({ cards: [], els: [], bet: b, doubled: false, done: false, split, splitAces: false, result: null, win: 0 });

/* ---------- layout ---------- */
const handCx = (i, n) => n === 1 ? 800 : [[560, 1040], [400, 800, 1200], [290, 620, 980, 1310]][n - 2][i];
const PY = 392, DY = 112;
function layout(dur = 420) {
  const n = hands.length, step = n === 1 ? 34 : 24;
  hands.forEach((h, i) => {
    const cx = handCx(i, n), w = CW + step * (h.cards.length - 1);
    h.els.forEach((c, k) => { const dbl = h.doubled && k === h.cards.length - 1; cards.move(c, cx - w / 2 + k * step + (dbl ? 18 : 0), PY + (dbl ? 14 : 0), { up: true, r: dbl ? 90 : 0, dur, z: 100 + i * 20 + k }); });
  });
  const dw = CW + 40 * (dealer.cards.length - 1);
  dealer.els.forEach((c, k) => cards.move(c, 800 - dw / 2 + k * 40, DY, { up: k !== 1 || holeUp, r: 0, dur, z: 50 + k }));
  labels();
}
function stackHTML(cx, cy, amount, cls = '') {
  const chips = ART.breakdown(amount).reverse().slice(-12);
  return `<div class="stack ${cls}" style="left:${cx}px;top:${cy}px">${chips.map((v, k) => `<img src="${ART.chip(v, 72)}" style="top:${-k * 5}px">`).join('')}${amount ? `<div class="amt">${money(amount)}</div>` : ''}</div>`;
}
function totText(ids, hideHole) {
  if (hideHole) return `${total([ids[0]]).t}`;
  const { t, soft } = total(ids);
  return t > 21 ? `<small>Bust</small>${t}` : soft && t < 21 ? `<small>Soft</small>${t}` : `${t}`;
}
function labels() {
  let h = '';
  const n = hands.length;
  if (phase === 'bet' || phase === 'title') h += stackHTML(800, 646, bet);
  hands.forEach((hd, i) => {
    const cx = handCx(i, n);
    if (hd.cards.length) h += `<div class="tot ${total(hd.cards).t > 21 ? 'bust' : ''}" style="left:${cx}px;top:${PY + CH + 10}px">${totText(hd.cards)}</div>`;
    h += stackHTML(cx, 646, hd.bet);
    if (phase === 'play' && i === active && n > 1) h += `<div class="mark" style="left:${cx}px;top:${PY - 44}px">Playing</div>`;
    if (hd.result) h += `<div class="res ${hd.result.cls} show" style="left:${cx}px;top:${PY - 46}px">${hd.result.text}${hd.result.amt ? `<b>${hd.result.amt}</b>` : ''}</div>`;
  });
  if (dealer.cards.length) h += `<div class="tot ${holeUp && total(dealer.cards).t > 21 ? 'bust' : ''}" style="left:800px;top:${DY + CH + 9}px"><small>Dealer</small>${totText(dealer.cards, !holeUp)}</div>`;
  $('#labels').innerHTML = h;
}
function hud() {
  $('#balance').textContent = money(SD.balance);
  $('#betNow').textContent = money(phase === 'bet' || phase === 'title' ? bet : hands.reduce((a, h) => a + h.bet, 0));
  $('#sHands').textContent = IR.fmt(ST.hands); $('#sWon').textContent = IR.fmt(ST.won); $('#sBJ').textContent = IR.fmt(ST.blackjacks);
  $('#shoeCount').textContent = shoe.length;
  $('#plDecks').textContent = deckName(shoeDecks || S.decks);
  $('#tBalance').textContent = money(SD.balance);
  controls();
}
const msg = html => $('#msg').innerHTML = html;
function controls() {
  const betting = phase === 'bet', playing = phase === 'play' && !busy;
  $('#betBar').classList.toggle('hidden', !(betting || phase === 'title'));
  $('#playBar').classList.toggle('hidden', betting || phase === 'title');
  const broke = betting && SD.balance < MIN_BET && bet === 0;
  $('#restock').classList.toggle('hidden', !broke);
  $('#chipRow').classList.toggle('hidden', broke); document.querySelector('#betBar .btns').classList.toggle('hidden', broke);
  document.querySelectorAll('.chipBtn').forEach(b => b.disabled = !betting || busy || bet + +b.dataset.v > Math.min(SD.balance, MAX_BET));
  $('#bClear').disabled = !betting || busy || !bet;
  $('#bDeal').disabled = !betting || busy || bet < MIN_BET || bet > SD.balance;
  const h = hands[active];
  $('#bHit').disabled = $('#bStand').disabled = !playing || !h;
  $('#bDouble').disabled = !playing || !canDouble(h);
  $('#bSplit').disabled = !playing || !canSplit(h);
}
const canDouble = h => !!h && !h.done && h.cards.length === 2 && !h.splitAces && SD.balance >= h.bet;
const canSplit = h => !!h && !h.done && h.cards.length === 2 && val(h.cards[0]) === val(h.cards[1]) && hands.length < MAX_HANDS && SD.balance >= h.bet && !h.splitAces;

/* ---------- shoe ---------- */
async function shuffle() {
  shoeDecks = S.decks;
  banner('Shuffling', deckName(shoeDecks)); msg(`Shuffling a fresh ${deckName(shoeDecks).toLowerCase()}...`);
  audio.sShuffle();
  discards.forEach((c, k) => { cards.move(c, SHOE.x, SHOE.y, { up: false, r: -8, dur: 500 + k * 2 }); setTimeout(() => cards.remove(c), 900); });
  discards = [];
  shoe = IR.shuffle(Array.from({ length: 52 * shoeDecks }, (_, i) => i));
  hud(); await wait(1700);
}
async function give(target, up = true, dur = 420) {
  const id = shoe.pop();
  const c = cards.make(id, { x: SHOE.x, y: SHOE.y, r: -8, up: false });
  target.cards.push(id); target.els.push(c);
  audio.sDeal();
  layout(dur); hud();
  await wait(dur + 120);
  return id;
}
function sweep() {
  [...hands.flatMap(h => h.els), ...dealer.els].forEach(c => { const k = discards.length; discards.push(c); cards.move(c, DISC.x + (Math.random() - .5) * 8, DISC.y + 8 - Math.min(40, k * .35), { up: false, r: (Math.random() - .5) * 8, dur: 520, z: 20 + k }); });
  hands = []; dealer = { cards: [], els: [] }; holeUp = false;
}

/* ---------- betting ---------- */
const chipRow = $('#chipRow');
ART.DENOMS.slice().reverse().forEach(v => { const b = IR.el('button', 'chipBtn', chipRow, `<img src="${ART.chip(v, 70)}" alt="${v}">`); b.dataset.v = v; b.title = `Add ${v}`; b.onclick = () => addChip(v); });
function addChip(v) { if (phase !== 'bet' || busy || bet + v > Math.min(SD.balance, MAX_BET)) return; bet += v; audio.sChip(1); labels(); hud(); msg(bet >= MIN_BET ? `Bet <b>${money(bet)}</b>. Press <b>Deal</b> when you're ready.` : 'Place your bet.'); }
$('#bClear').onclick = () => { if (phase !== 'bet' || busy) return; bet = 0; audio.sChip(2); labels(); hud(); msg('Place your bet.'); };
$('#bDeal').onclick = () => dealRound();
$('#bRestock').onclick = () => { SD.balance += START; ST.restocks++; save.write(true); audio.sChip(6); bet = Math.min(SD.lastBet || 25, SD.balance); labels(); hud(); msg(`The house restocked you with <b>${IR.fmt(START)}</b> play chips.`); };

function toBetting() {
  phase = 'bet'; busy = false;
  bet = SD.balance >= MIN_BET ? Math.max(MIN_BET, Math.min(SD.lastBet || 25, SD.balance, MAX_BET)) : 0;
  bet -= bet % 5;
  labels(); hud();
  msg(SD.balance < MIN_BET ? 'You’re out of chips.' : `Same bet, <b>${money(bet)}</b>? Press <b>Deal</b>, or change it with the chips.`);
}

/* ---------- a round ---------- */
async function dealRound() {
  if (phase !== 'bet' || busy || bet < MIN_BET || bet > SD.balance) return;
  busy = true; phase = 'deal'; audio.unlock();
  SD.lastBet = bet; SD.balance -= bet; SD.atRisk = bet; save.write(true);
  audio.sChip(3);
  if (!shoe.length || shoe.length < cutFor(shoeDecks) || shoeDecks !== S.decks) await shuffle();
  hands = [newHand(bet)]; dealer = { cards: [], els: [] }; active = 0; holeUp = false;
  hud(); msg('Dealing...');
  await give(hands[0]); await give(dealer); await give(hands[0]); await give(dealer, false);
  const upV = val(dealer.cards[0]);
  if (upV === 11 || upV === 10) {
    msg(`Dealer shows ${upV === 11 ? 'an Ace' : 'a ten'} and checks for Blackjack...`); await wait(1100);
    if (total(dealer.cards).t === 21) { await revealHole(); return settle(); }
    msg('No Blackjack for the dealer. Play on.'); await wait(500);
  }
  if (isBJ(hands[0])) { hands[0].done = true; await revealHole(); return settle(); }
  phase = 'play'; busy = false; prompt(); hud(); labels();
}
function prompt() {
  const h = hands[active]; if (!h) return;
  const { t, soft } = total(h.cards), up = dealer.cards[0];
  const opts = ['<b>Hit</b>', '<b>Stand</b>'].concat(canDouble(h) ? ['<b>Double</b>'] : []).concat(canSplit(h) ? ['<b>Split</b>'] : []);
  msg(`${hands.length > 1 ? `Hand ${active + 1}: ` : ''}You have ${soft && t < 21 ? 'soft ' : ''}<b>${t}</b> against the dealer's <b>${IR.RANKS[IR.rank(up) - 1]}</b>. ${opts.join(', ').replace(/, ([^,]*)$/, ' or $1')}?`);
}
async function act(fn) { if (phase !== 'play' || busy) return; busy = true; hud(); await fn(hands[active]); }
async function hit(h) {
  await give(h);
  const t = total(h.cards).t;
  if (t > 21) { h.done = true; h.result = { cls: 'bust', text: 'Bust' }; sfx.bust(); labels(); await wait(600); }
  else if (t === 21) h.done = true;
  return next();
}
async function stand(h) { h.done = true; audio.sCard(); return next(); }
async function double(h) {
  if (!canDouble(h)) { busy = false; return hud(); }
  SD.balance -= h.bet; SD.atRisk += h.bet; h.bet *= 2; h.doubled = true; save.write(true); audio.sChip(3); hud();
  msg(`Doubled to <b>${money(h.bet)}</b>. One card...`);
  await wait(300); await give(h);
  h.done = true;
  if (total(h.cards).t > 21) { h.result = { cls: 'bust', text: 'Bust' }; sfx.bust(); labels(); await wait(600); }
  return next();
}
async function split(h) {
  if (!canSplit(h)) { busy = false; return hud(); }
  SD.balance -= h.bet; SD.atRisk += h.bet; save.write(true); audio.sChip(3);
  const nh = newHand(h.bet, true);
  nh.cards.push(h.cards.pop()); nh.els.push(h.els.pop());
  h.split = true;
  const aces = IR.rank(h.cards[0]) === 1;
  if (aces) h.splitAces = nh.splitAces = true;
  hands.splice(active + 1, 0, nh);
  layout(); hud(); msg(aces ? 'Split Aces: one card each.' : 'Split into two hands.'); await wait(500);
  await give(h);
  if (aces) { await give(nh); h.done = nh.done = true; }
  else if (total(h.cards).t === 21) h.done = true;
  return next();
}
async function next() {
  if (!hands[active].done) { busy = false; prompt(); hud(); labels(); return; }
  for (let i = active + 1; i < hands.length; i++) {
    const h = hands[i]; if (h.done) continue;
    active = i; labels(); await wait(250);
    if (h.cards.length === 1) { await give(h); if (total(h.cards).t === 21) { h.done = true; continue; } }
    busy = false; prompt(); hud(); labels(); return;
  }
  await dealerPlay();
}
async function revealHole() { holeUp = true; audio.sFlip(); layout(380); await wait(650); }
async function dealerPlay() {
  phase = 'dealer'; busy = true; hud(); labels();
  msg('Dealer turns over the hole card...');
  await revealHole();
  const live = hands.some(h => h.result?.cls !== 'bust');
  if (live) while (total(dealer.cards).t < 17) { msg(`Dealer has <b>${total(dealer.cards).t}</b> and draws.`); await wait(350); await give(dealer, true, 460); }
  const dt = total(dealer.cards).t;
  msg(!live ? 'All hands bust.' : dt > 21 ? `Dealer busts with <b>${dt}</b>!` : `Dealer stands on <b>${dt}</b>.`);
  await wait(500);
  settle();
}
async function settle() {
  phase = 'settle'; busy = true;
  const dt = total(dealer.cards).t, dBJ = dealer.cards.length === 2 && dt === 21;
  let net = 0, bjHit = false, outcomes = [];
  hands.forEach(h => {
    const pt = total(h.cards).t, bj = isBJ(h) && hands.length === 1;
    let pay = 0, r;
    if (pt > 21) r = { cls: 'bust', text: 'Bust' };
    else if (dBJ) r = bj ? { cls: 'push', text: 'Push' } : { cls: 'lose', text: 'Dealer Blackjack' };
    else if (bj) { r = { cls: 'bj', text: 'Blackjack' }; bjHit = true; }
    else if (dt > 21 || pt > dt) r = { cls: 'win', text: 'Win' };
    else if (pt === dt) r = { cls: 'push', text: 'Push' };
    else r = { cls: 'lose', text: 'Lose' };
    if (r.cls === 'bj') pay = h.bet * 2.5; else if (r.cls === 'win') pay = h.bet * 2; else if (r.cls === 'push') pay = h.bet;
    const d = pay - h.bet; net += d;
    r.amt = d > 0 ? '+' + money(d) : d < 0 ? '−' + money(-d) : '';
    h.result = r; SD.balance += pay; outcomes.push(r.cls);
    ST.hands++;
    if (r.cls === 'win' || r.cls === 'bj') ST.won++; else if (r.cls === 'push') ST.pushed++; else ST.lost++;
    if (r.cls === 'bj') ST.blackjacks++;
  });
  SD.atRisk = 0; ST.peak = Math.max(ST.peak, SD.balance); if (net > 0) ST.biggestWin = Math.max(ST.biggestWin, net);
  save.write(true); labels(); hud();
  const why = dBJ ? 'Dealer has Blackjack' : dt > 21 ? `Dealer busts with ${dt}` : `Dealer has ${dt}`;
  if (bjHit) { banner('Blackjack!', `Paid 3 to 2: +${money(net)} chips`); sfx.blackjack(); fx.burst(800, 360, 150, ['#ffe7a6', '#f6dc8a', '#ffffff', '#c99a3a'], { speed: 760, size: 8 }); }
  else if (net > 0) { banner('You Win', `+${money(net)} chips · ${why}`); sfx.win(); fx.burst(800, 380, 80, ['#ffe7a6', '#ffffff', '#c99a3a'], { speed: 600, size: 7, shape: 'circle' }); }
  else if (net === 0) { banner(outcomes.every(o => o === 'push') ? 'Push' : 'Even', outcomes.every(o => o === 'push') ? 'Your bet is returned' : 'You broke even this hand'); sfx.push(); }
  else { banner(outcomes.every(o => o === 'bust') ? 'Bust' : 'Dealer Wins', `−${money(-net)} chips · ${outcomes.every(o => o === 'bust') ? 'Over 21' : why}`); outcomes.every(o => o === 'bust') ? null : sfx.lose(); }
  msg(`${net > 0 ? `You won <b>${money(net)}</b>` : net < 0 ? `You lost <b>${money(-net)}</b>` : 'No chips changed hands'}. ${why}.`);
  await wait(3200);
  sweep(); audio.sCard();
  await wait(450);
  toBetting();
}
function banner(t, sub) { const b = $('#banner'); b.innerHTML = t + (sub ? `<small>${sub}</small>` : ''); b.classList.add('show'); clearTimeout(banner.t); banner.t = setTimeout(() => b.classList.remove('show'), 1500 * pace() / (IR.speed || 1)); }

$('#bHit').onclick = () => act(hit);
$('#bStand').onclick = () => act(stand);
$('#bDouble').onclick = () => act(double);
$('#bSplit').onclick = () => act(split);
addEventListener('keydown', e => {
  if (!$('#title').classList.contains('hidden') || document.querySelector('.veil.closable:not(.hidden)')) return;
  const k = e.key.toLowerCase();
  if (k === 'h') act(hit); else if (k === 's') act(stand); else if (k === 'd') act(double); else if (k === 'p') act(split);
  else if (k === 'enter' || k === ' ') { e.preventDefault(); if (phase === 'bet') dealRound(); }
  else if (k === 'backspace' && phase === 'bet') $('#bClear').click();
});

/* ---------- menus ---------- */
function chips(id, key, num, after) { const box = $(id); const sync = () => box.querySelectorAll('button').forEach(b => b.classList.toggle('on', String(b.dataset.v) === String(S[key]))); box.querySelectorAll('button').forEach(b => b.onclick = () => { S[key] = num ? +b.dataset.v : b.dataset.v; save.write(); sync(); after && after(); }); sync(); return sync; }
const syncDeckSel = () => { $('#sDecks').value = String(S.decks); };
const syncTitleDecks = chips('#optDecks', 'decks', true, () => { syncDeckSel(); hud(); });
chips('#optPace', 'pace');
$('#sDecks').onchange = async e => {
  S.decks = +e.target.value; save.write(); syncTitleDecks();
  if (phase === 'bet' && !busy) { busy = true; await shuffle(); busy = false; hud(); toBetting(); }
  else msg(`The table switches to a <b>${deckName(S.decks).toLowerCase()}</b> after this hand.`);
};
syncDeckSel();
$('#tChips').innerHTML = ART.DENOMS.slice().reverse().map(v => `<img src="${ART.chip(v, 54)}" alt="">`).join('');
if (notice) { $('#notice').textContent = notice; $('#notice').classList.remove('hidden'); }
$('#bStart').onclick = async () => {
  audio.unlock(); IR.panel.hide('title');
  if (!shoe.length || shoeDecks !== S.decks) { phase = 'deal'; busy = true; hud(); await shuffle(); }
  toBetting();
};
$('#bRules').onclick = $('#bHow').onclick = () => IR.panel.show('rules');
$('#bCredits').onclick = $('#bCredits2').onclick = async () => { $('#changelog').innerHTML = IR.mdChangelog(await IR.loadChangelog()); IR.panel.show('credits'); };
IR.panel.wire();
bet = Math.min(SD.lastBet || 25, Math.max(0, SD.balance));
hud(); labels();
window.__bj = { get hands() { return hands; }, get dealer() { return dealer; }, get phase() { return phase; }, get busy() { return busy; }, get shoe() { return shoe; }, SD, ST, total, set shoeOrder(a) { shoe = a.slice(); shoeDecks = S.decks; } };
})();
