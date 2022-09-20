Nice Fort
---

A multiplayer dungeon crawler that's played over `ssh`.

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
- ui
  - fix layout bug with inventory table
  - show location of other players in sidebar (arrow and distance)
  - `g` to pick up entity player is standing on instead of moving towards
  - can i do mouse support?
  - render event types with color/style
- survival gameplay
  - tools (axe, shovel, etc)
  - placing crafted items
  - diagonal movement instead of only manhattan
- monsters
  - debouncing system for fighting
  - all visible NPCs show in sidebar (name, health)
- better map view
  - don't scroll with each move. fit the level on the screen
  - current player renders as `@`, other players use first initial
- on disk (or remote) persistence of world state
- `look`: highlight items / monsters to see desc