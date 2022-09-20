Nice Fort
---

A multiplayer realtime roguelike that's played over `ssh`.

## Concept

You find yourself on a remote island with nothing. Together with your friends, you must survive as long as possible in an ever more hostile environment. Grow and harvest food; craft tools, weapons and materials; fight wild beasts. Make permanent changes to the world around you.

## Play

Try the live dev server (pubkey auth required). Note that you will need a terminal that supports "truecolor" such as iTerm2.

```shell
ssh -p 2222 nicefort.fly.dev
```

### Run Locally

```shell
go build . && ./nicefort
ssh -p 23234 jordan@127.0.0.1
```

### TODO
- survival gameplay
  - tools (axe, shovel, etc)
  - eating / hunger
  - farming
    - collect seeds
    - library of plant entities
    - growth loop
    - harvesting
    - NPCs that eat/trample crops
  - placing crafted items
  - diagonal movement instead of only manhattan?
- monsters
  - debouncing system for fighting
- ui
  - fix layout bug with inventory table and make it look consistent
  - `g` to pick up entity player is standing on instead of moving towards
  - mouse support
  - render event types with color/style
- better map view
  - don't scroll with each move. fit the level on the screen
  - current player renders as `@`, other players use first initial
- on disk (or remote) persistence of world state
- `look`: highlight items / monsters to see desc