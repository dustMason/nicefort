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
- use contoured tiling algo to draw dungeon walls (https://github.com/dustMason/legrid/blob/main/sketch.js#L119-L159)
- items / inventory
- ticker in `world` to handle moving NPEs
- fighting