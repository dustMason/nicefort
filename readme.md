Nice Fort
---

A multiplayer dungeon crawler that's played over `ssh`. Nothing much to see yet, but watch this space.

## Run

```shell
go build . && ./nicefort
```

## Play

```shell
ssh -p 23234 jordan@127.0.0.1
```

## TODO
- chat
- scrolling ui for chat / status messages
- items / inventory
- ticker in `world` to handle moving NPEs
- on disk persistence of world state
- fighting
- switch from `map[coord]*entity` to `[]*entity` for `World.entity` for better perf