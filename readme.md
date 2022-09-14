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
- ticker in the bubbletea prog to push updates when the world changes
- ticker in `world` to handle moving NPEs
- focused, scrolling view for each player
- procedural dungeon generation
- fighting, weapons, items
- make rendering the `world` string more performant