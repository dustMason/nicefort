Nice Fort
---

A multiplayer dungeon crawler that's played over `ssh`.

Try the live dev server (pubkey auth required):

```shell
ssh -p 2222 nicefort.fly.dev
```

### Run Locally

```shell
go build . && ./nicefort
ssh -p 23234 jordan@127.0.0.1
```

### TODO
- multiplayer niceties
  - scrolling ui for chat / status messages
  - show location of other players in sidebar (arrow and distance)
- survival gameplay
  - base creation (campfire?)
- can i do mouse support?
- monsters
  - fighting
  - debouncing system for movement/fighting
  - all visible NPCs show in sidebar (name, health)
- better map view
  - don't scroll with each move. fit the level on the screen
  - current player renders as `@`, other players use first initial
- player stats?
- character creation?
- on disk (or remote) persistence of world state
- multiple levels
- `look`: highlight items / monsters to see desc