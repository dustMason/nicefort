Nice Fort
---

A multiplayer dungeon crawler that's played over `ssh`.

Try the live dev server (pubkey auth required):

```shell
ssh nicefort.fly.dev
```

### Run Locally

```shell
go build . && ./nicefort
ssh -p 23234 jordan@127.0.0.1
```

### TODO
- better map view
    - don't scroll with each move. fit the level on the screen
    - current player renders as `@`, other players use first initial
- scrolling ui for chat / status messages
- items / inventory
- player stats
  - stat bar on the side
    - health
    - level
    - XP
    - money
    - dungeon level
- handle window resize events
  - make the map view occupy the entire terminal
- monsters
- ticker in `world` to handle moving NPEs
- fighting
- doors
- on disk (or remote) persistence of world state
- multiple levels
- switch from `map[coord]*entity` to `[]*entity` for `World.entity` for better perf
- `look`: highlight items / monsters to see desc
- character creation
- 