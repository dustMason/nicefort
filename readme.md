Nice Fort
---

A multiplayer realtime roguelike that's played over `ssh`.

## Concept

You find yourself on a remote island with only a backpack. Together with your friends, you must survive as long as possible in an ever more hostile environment. Grow and harvest food; craft tools, weapons and materials; fight wild beasts. Make permanent changes to the world around you.

Each season lasts about 10 minutes. You must create a NICE FORT capable of withstanding the harsh winter.

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
- worldgen
  - seasonal spawning of NPCs
    - always off-screen (not near player)
    - progressively harder with each year
- survival gameplay
  - tools (axe, shovel, etc)
    - [x] game support
    - list of actual items
  - [x] eating / hunger
  - clay (digging mud near water)
  - carrying water
    - tight basket
    - clay pot
    - animal skin
  - farming
    - collect seeds
    - library of plant entities
      - research medicinal plants
    - growth loop
    - [x] harvesting
    - NPCs that eat/trample crops
  - crafting
    - [x] stone tools
    - clay
      - bricks
      - kiln to make charcoal
    - fire 
      1. friction fire starter (stick, twine, kindling)
      2. charcoal oven (make with clay, feed with wood, produce charcoal)
    - plant items
      - roofing
  - diagonal movement instead of only manhattan?
- animals
  - killing the first animal should be really hard. a big milestone that unlocks a lot of crafting
    - needle (bone) and thread (sinew)
    - skins
    - bone fishing hook
  - debouncing system for fighting
- ui
  - [x] fix layout bug with inventory table and make it look consistent
  - [x] `space` to pick up entity player is standing on instead of moving towards
  - mouse support
  - render event types with color/style
- refactor: instead of entity fields `npc`, `player`, `flora` make `NPC`, `player`, `Flora` each embed `entity`. Then switch statements can handle current `attackable`, `harvestable` scenarios. Make a new type `ItemEntity`.
- better map view
  - don't scroll with each move. fit the level on the screen
  - current player renders as `@`, other players use first initial
- on disk (or remote) persistence of world state
- `look`: highlight items / monsters to see desc