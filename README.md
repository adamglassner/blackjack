# Blackjack — The Gold Room

Twenty-one against the dealer at a casino table with dark green felt and gold
trim, by Ironite. Written in Go with [Ebitengine](https://ebitengine.org). It
ships on its own and inside the Ironite Launcher.

**Play money only.** Chips have no cash value: no real money, no purchases, no accounts.

## How to play

- Click chips to build your bet in the circle, then **Deal**. Get closer to 21 than the dealer without going over. Aces count 1 or 11, face cards count 10.
- **Hit**, **Stand**, **Double Down** on your first two cards, or **Split** a pair into two hands, up to four. Split Aces get one card each.
- The dealer checks for Blackjack when showing an Ace or a ten, draws to 16 and stands on all 17s. Blackjack pays 3 to 2, a win pays 1 to 1, a tie is a push.
- Single Deck, Double Deck or a 6-Deck Shoe, from the toolbar or the title screen; Relaxed, Steady or Quick dealing.
- Keys: `Enter` deal, `H` hit, `S` stand, `D` double, `P` split, `Backspace` clear bet, `M` music, `F11` full screen, `Esc` closes a panel.

You start with 1,000 chips and the house restocks you for free if you run out.
Your balance, stats and settings are saved in
`%APPDATA%\Ironite\Blackjack\save.json` (the same file as 1.x), outside the
install folder. A bet left on the table when the game closes mid-hand is
returned next time.

## Code

| File | What it is |
|---|---|
| `main.go` | Window setup, the frame loop, mouse and keys. |
| `game.go` | The rules and the round: betting, dealing, the dealer's peek, hit, stand, double, split, the dealer's draw and the payouts. A round runs as a script on its own goroutine. |
| `view.go`, `ui.go`, `gfx.go` | The table, toolbar, console, banner and panels; buttons and sliders; drawing helpers. Positions, sizes and colours match the original layout. |
| `art.go`, `canvas.go`, `fonts.go` | The table, the court card deck and the chips, drawn in Go at startup, shape for shape the same as the original art. |
| `anim.go`, `colors.go` | Card animation and easing, confetti, colours. |
| `audio.go`, `save.go`, `assets.go` | Music and sound effects, the save file, embedded assets. |
| `platform_windows.go` | One copy at a time: opening it again brings the window forward. |
| `assets/` | Fonts, the Ironite logo and icon, and the music and sound effects. |
| `tools/export/` | Renders the original synthesized music and sounds (`web/core.js`, `web/music.js`) to Ogg in headless Chromium. |
| `tools/mkres/` | Writes the Windows icon and version info into the .exe. |
| `installer/` | Inno Setup scripts for the standalone `Blackjack-Setup-v<version>.exe`. |

`platform_js.go` and `sound_stub.go` (`-tags nosound`) are for the browser
build used for screenshots and the automated play test (`?auto`).

## Building

You need Go 1.24 or newer and bash (Git Bash works on Windows). No C
compiler is needed.

```bash
./build.sh
```

This writes `dist/Games/Blackjack/Blackjack.exe` along with its README,
CREDITS, CHANGELOG and license files. The version comes from the top entry of
`CHANGELOG.md`.

To build the installer, install Inno Setup 6.3 or newer, run `./build.sh`, then:

```
ISCC.exe installer\BlackjackSetup.iss
```

To re-export the music and sound effects (needs Node with Playwright, and ffmpeg):

```bash
node tools/export/export-audio.mjs wav && tools/export/encode.sh wav
```

## Releasing

Bump the version by adding a new section at the top of `CHANGELOG.md` and push
to `main`. The release workflow builds the game and publishes
`Blackjack-<version>-windows.zip` as a GitHub release, which Ironite Launcher
installs and updates from.

## License

© 2026 Ironite. All rights reserved. Fonts are under the SIL Open Font License (`assets/fonts/OFL.txt`).
