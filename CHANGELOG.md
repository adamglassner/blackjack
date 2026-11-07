# Changelog

All notable changes to Blackjack are recorded here.
Versions follow Semantic Versioning (MAJOR.MINOR.PATCH).

## [2.0.0] - 2026-11-06

### Changed
- **Blackjack is now a native Go game.** It no longer opens in a Microsoft Edge window: it runs in its own window, drawn with Ebitengine, starts faster and needs nothing else installed.
- Same graphics as before: the Gold Room table, the court card deck, the card back and the chips are drawn by the game with the same artwork, and the three lounge tunes and all the sound effects are the same.
- Same rules, buttons, keys and layout: Hit, Stand, Double Down and Split up to four hands, a dealer who stands on all 17s, Blackjack paying 3 to 2, and Single Deck, Double Deck or a 6-Deck Shoe.
- Your balance, stats and settings carry over: they are read from and saved to the same file in %APPDATA%\Ironite\Blackjack.
- One copy runs at a time; opening Blackjack again brings the open window forward.
- The window can be resized freely and the table stays sharp at any size and in full screen (F11).

## [1.0.0] - 2026-10-26

### Added
- First release of Blackjack for Windows.
- Player vs. dealer at The Gold Room: dark green felt, gold trim, a padded rail and soft overhead light.
- Single Deck, Double Deck or 6-Deck Shoe, always shown on the table plaque and in the toolbar, with a live count of cards left in the shoe.
- Hit, Stand, Double Down and Split (up to four hands). Split Aces get one card each.
- Dealer stands on all 17s and checks for Blackjack when showing an Ace or a ten. Blackjack pays 3 to 2 and a regular win pays 1 to 1.
- Both hands and their totals (including soft totals) are shown at all times, with a clear result for every hand after each round.
- Bet with 5, 25, 100, 500 and 1,000 chips. Your bet stays up for the next hand.
- Play-money chips only: no real money, no purchases and no accounts. Start with 1,000 chips, and the house restocks you for free if you run out.
- Your balance and stats persist between sessions, kept in %APPDATA%\Ironite\Blackjack. A bet left on the table when the game closes is returned next time.
- HD deck drawn in-house with royal court cards and a gold-filigree card back, plus hand-drawn casino chips.
- Three original lounge tunes: "Velvet Hour", "Last Call at the Gold Room" and "Midnight Felt", with volume control and mute.
- Sound effects for the shuffle, the deal, chips, and win, push and lose chimes.
- Keyboard shortcuts, a How to Play guide and a Credits & Changelog screen.
