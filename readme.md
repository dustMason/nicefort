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
- scrolling ui for chat / status messages
- ui for items / inventory
- crafting
  - ui
  - crafting tree
- player stats?
- character creation?
- better map view
  - don't scroll with each move. fit the level on the screen
  - current player renders as `@`, other players use first initial
- monsters
- ticker in `world` to handle moving NPEs
- fighting
- on disk (or remote) persistence of world state
- multiple levels
- `look`: highlight items / monsters to see desc