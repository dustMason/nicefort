package world

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustmason/nicefort/events"
	"github.com/dustmason/nicefort/util"
	"github.com/kelindar/tile"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const NPCActivationRadius = 10 // approx. distance from any player where NPCs take turns
const secondsPerDay = 6.66     // seconds of real-world clock time per in-game day
var blackSpace = environmentTiles[Space][0]
var memColor = "#444444"
var memStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(memColor))

type Coord struct {
	X, Y int
}

func (c Coord) GetXY() (int, int) {
	return c.X, c.Y
}

type World struct {
	sync.RWMutex
	W, H int
	grid *tile.Grid
	// wMap       []location         // the actual map of tiles

	players *util.TombstoneList[*entity]
	npcs    *util.TombstoneList[*entity]
	items   *util.TombstoneList[*entity]
	flora   *util.TombstoneList[*entity]

	playerMap  map[string]*entity // map of player id => entity that points to that player
	activeNPCs []*entity
	events     *events.EventList
	onEvent    map[string]func(string) // map of player id => chat callback
	days       float64                 // age of the world
	lastTick   time.Time
}

func NewWorld(size int) *World {
	g := tile.NewGrid(int16(size), int16(size))
	npcs, items, flora := GenerateOverworld(g)
	w := &World{
		W: size,
		H: size,

		grid:    g,
		players: util.NewTombstoneList[*entity](15),
		npcs:    util.NewTombstoneList[*entity](255, npcs...),
		items:   util.NewTombstoneList[*entity](65535, items...),
		flora:   util.NewTombstoneList[*entity](65535, flora...),

		playerMap: make(map[string]*entity),
		events:    events.NewEventList(100),
		onEvent:   make(map[string]func(string)),
		lastTick:  time.Now(),
	}
	go w.runTicker()
	return w
}

// SizeX SizeY IsPassable and OOB satisfy the dmap interface
func (w *World) SizeX() int {
	return w.W
}

func (w *World) SizeY() int {
	return w.H
}

func (w *World) IsPassable(x int, y int) bool {
	return w.walkable(x, y)
}

func (w *World) OOB(x int, y int) bool {
	return !w.InBounds(x, y)
}

func (w *World) tick(t time.Time) {
	w.days += t.Sub(w.lastTick).Seconds() / secondsPerDay
	w.lastTick = t
	for _, e := range w.activeNPCs {
		e.npc.Tick(t, w, e)
	}
	for _, e := range w.playerMap {
		e.player.Tick(t)
	}
}

func (w *World) runTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for t := range ticker.C {
		w.tick(t)
	}
}

func (w *World) MovePlayer(dx, dy int, playerID string) {
	e := w.getPlayer(playerID)
	nx, ny := e.player.GetLocation()
	nx += dx
	ny += dy
	now := time.Now()
	if !e.player.CanMove(now) {
		return
	}
	w.Lock()
	defer w.Unlock()
	if w.InBounds(nx, ny) {
		if ent, ok := w.attackable(nx, ny); ok {
			damage, success, dead, drops := ent.npc.Attacked(e.player.wielding, e, 10)
			// todo need a progress calc to use Activity
			if dead {
				e.player.Event(events.Success, fmt.Sprintf("You killed the %s", ent.npc.Name))
				w.removeNPC(nx, ny)
				for _, drop := range drops {
					e.player.Event(events.Success, fmt.Sprintf("It dropped %d x %s", drop.Quantity, drop.Item.Name))
					x, y, _ := w.findNearbyAvailableIndex(nx, ny)
					w.addItem(x, y, drop.Item, drop.Quantity)
				}
			} else if !success {
				e.player.Event(events.Warning, fmt.Sprintf("Your %s doesn't do anything to the %s", e.player.wielding.Name, ent.npc.Name))
			} else {
				e.player.Event(events.Success, fmt.Sprintf("You hit the %s for %d", ent.npc.Name, damage))
			}
			return
		}
		if w.walkable(nx, ny) && !w.occupied(nx, ny) {
			oldX, oldY := e.player.GetLocation()
			w.swapPlayer(oldX, oldY, nx, ny, e)
			e.player.SetLocation(nx, ny, now)
			e.player.See(w)
			w.refreshActiveNPCs()
			return
		}
		if ent, ok := w.harvestable(nx, ny); ok {
			w.harvest(e.player, ent, nx, ny)
			return
		}
	}
}

func (w *World) InteractPlayer(playerID string) {
	e := w.getPlayer(playerID)
	now := time.Now()
	if !e.player.CanMove(now) {
		return
	}
	x, y := e.player.GetLocation()
	w.Lock()
	defer w.Unlock()
	if ent, ok := w.harvestable(x, y); ok {
		w.harvest(e.player, ent, x, y)
		return
	}
	if ee, ok := w.pickupable(x, y); ok {
		took := e.player.PickUp(ee.item, ee.quantity)
		ee.quantity -= took
		if ee.quantity == 0 {
			w.removeItem(x, y)
		}
		return
	}
}

func (w *World) harvest(player *player, ent *entity, x, y int) {
	dead, success, progress, drops := ent.flora.Harvest(player.wielding)
	player.SetActivity(Activity{description: ent.flora.name, progress: progress})
	for _, drop := range drops {
		player.Event(events.Success, fmt.Sprintf("It yielded %d x %s", drop.Quantity, drop.Item.Name))
		nx, ny, _ := w.findNearbyAvailableIndex(x, y)
		w.addItem(nx, ny, drop.Item, drop.Quantity)
	}
	if dead {
		player.Event(events.Success, fmt.Sprintf("You harvested the %s", ent.flora.name))
		w.removeFlora(x, y, ent)
	} else if !success {
		// handle the case where the ent is exhausted. "you can't harvest any more with your x"
		player.Event(events.Warning, fmt.Sprintf("Your %s does not work here", player.wielding.Name))
	} else {
		// show progress bar?
	}

}

func (w *World) MoveNPC(x, y int, e *entity) {
	w.Lock()
	defer w.Unlock()
	if w.InBounds(x, y) && w.walkable(x, y) && !w.occupied(x, y) { // todo some NPCs can move over different types of terrain
		w.swapNPC(e.npc.loc.X, e.npc.loc.Y, x, y, e)
		e.npc.loc = Coord{x, y}
	}
}

func (w *World) ActivateItem(playerID string, inventoryIndex int) {
	fmt.Println("activating item index", inventoryIndex)
	e := w.getPlayer(playerID)
	ii := e.player.inventory[inventoryIndex]
	consumed, message := ii.Item.Activate(e, w)
	e.player.Event(events.Info, message)
	if consumed {
		e.player.ConsumeItem(ii.Item)
	}
}

func (w *World) PlayerInventory(playerID string) []*InventoryItem {
	e := w.getPlayer(playerID)
	return e.player.Inventory()
}

func (w *World) AvailableRecipes(playerID string) []Recipe {
	e := w.getPlayer(playerID)
	return AvailableRecipes(e.player.inventoryMap, e, w)
}

func (w *World) DoRecipe(playerID string, r Recipe) bool {
	e := w.getPlayer(playerID)
	ok, newInv := r.Do(e.player.inventoryMap, e, w)
	if ok {
		e.player.ReplaceInventory(newInv)
		return true
	}
	return false
}

func (w *World) DisconnectPlayer(playerID string) {
	e := w.getPlayer(playerID)
	w.disconnectPlayer(e)
	w.Event(events.Warning, fmt.Sprintf("%s left.", e.player.name))
	w.Lock()
	defer w.Unlock()
	delete(w.onEvent, playerID)
	w.refreshActiveNPCs()
}

func (w *World) RenderPlayerEvents(playerID string) string {
	e := w.getPlayer(playerID)
	return e.player.Events()
}

func (w *World) RenderMap(playerID, playerName string, vw, vh int) string {
	vw = vw / 2 // each environmentTile is 2 chars wide
	ply := w.getOrCreatePlayer(playerID, playerName).player
	x, y := ply.GetLocation()
	seen := ply.AllVisited()

	w.RLock()
	defer w.RUnlock()
	var b strings.Builder
	b.Grow(vw*vh*2 + vh) // *2 because double-width chars and +vh because line-breaks
	left := x - vw/2
	right := x + vw/2
	top := y - vh/2
	bottom := y + vh/2

	// get all the tiles within the viewport
	w.grid.Within(
		tile.Point{X: int16(left), Y: int16(top)},
		tile.Point{X: int16(right), Y: int16(bottom)},
		func(pt tile.Point, tile tile.Tile) {
			inView, dist := ply.CanSee(int(pt.X), int(pt.Y))
			memString, inPastView := seen[ic]
			if inPastView && !inView {
				b.WriteString(memString)
			} else if inView {
				// grab the tile from the grid
				var ent *entity
				ents, biome := w.allEntitiesAndBiomeAt(ix, iy)
				if len(ents) > 0 {
					fmt.Println("inview", biome, ents)
					ent = ents[0]
				} else {
					ent = &entity{environment: Grass}
				}
				// fmt.Println("render", ix, iy, ent, biome)
				bg := entityFromBiome(biome)
				b.WriteString(
					lipgloss.NewStyle().
						Foreground(lipgloss.Color(ent.ForegroundColor(dist))).
						Background(lipgloss.Color(bg.BackgroundColor(dist))).
						Render(ent.String()),
				)
			} else { // not in past or current fovView
				b.WriteString(blackSpace)
			}
			if int(pt.X) == right {
				b.WriteString("\n")
			}

		},
	)

	// 		// var ent *entity
	// 		// 	ic := Coord{ix, iy}
	// 		}
	// 		ix++
	// 	}
	// 	ix = left
	// 	b.WriteString("\n")
	// 	iy++
	// }
	return b.String()
}

func entityFromBiome(b Biome) *entity {
	switch b {
	case Ocean, River:
		return &entity{environment: Water}
	case Bog:
		return &entity{environment: Mud}
	case BirchForest, Boreal:
		return &entity{environment: Grass}
	case Rocky:
		return &entity{environment: Pebbles}
	case Mountainous, Glacial:
		return &entity{environment: Rock}
	default:
		return &entity{environment: Grass}
	}
}

func (w *World) RenderForMemory(x, y int) string {
	if t, ok := w.grid.At(int16(x), int16(y)); ok {
		_, b, _, i, f := w.tileToEntities(t)
		if f != nil {
			return f.String()
		} else if i != nil {
			return i.String()
		}
		return entityFromBiome(b).String()
	}
	return blackSpace
}

// todo refactor this to return some value type (map of string[string]?) or a struct
func (w *World) RenderPlayerSidebar(id string, name string) string {
	var b strings.Builder
	e := w.getOrCreatePlayer(id, name)
	myLoc := e.player.loc
	b.WriteString(name + "\n")
	b.WriteString(fmt.Sprintf("Pack: %.1f / %d\n", e.player.carrying, int(e.player.maxCarry)))
	b.WriteString(fmt.Sprintf("Health: %d / %d\n", e.player.health, e.player.maxHealth))
	b.WriteString(fmt.Sprintf("Hunger: %.3f\n", e.player.hunger))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Wielding: %s\n", e.player.wielding.Name))
	b.WriteString("\n")

	a := e.player.GetActivity()
	if a.description != "" {
		b.WriteString(fmt.Sprintf("%s\n%s\n", a.description, a.pBar.ViewAs(a.progress)))
	}

	for _, ee := range w.playerMap {
		if ee == e {
			continue
		}
		pLoc := ee.player.loc
		b.WriteString(
			compassIndicator(myLoc.X, myLoc.Y, pLoc.X, pLoc.Y) + " " + ee.player.name + "\n",
		)
	}

	b.WriteString("\n")
	for _, ee := range w.activeNPCs {
		nLoc := ee.npc.loc
		b.WriteString(
			compassIndicator(myLoc.X, myLoc.Y, nLoc.X, nLoc.Y) + " " + ee.npc.Name + "\n",
		)
	}

	// todo also get a list of the items the player can see and render compass items for those

	return b.String()
}

var seasons = []string{"spring", "summer", "fall", "winter"}

func (w *World) RenderWorldStatus() string {
	// world starts on first day of spring?
	day := int(w.days) % 365
	year := int(w.days / 365)
	i := int(float64(day) / 91.25)
	season := seasons[i]
	return fmt.Sprintf("%s : Year %d, Day %d", season, year, day)
}

func (w *World) getOrCreatePlayer(playerID, playerName string) *entity {
	w.Lock()
	defer w.Unlock()
	e, ok := w.playerMap[playerID]
	if !ok {
		x, y, _ := w.randomAvailableCoord()
		e = NewPlayer(playerID, Coord{x, y}, w.grid)
		w.playerMap[playerID] = e
	}
	e.player.name = playerName
	x, y := e.player.GetLocation()
	if !w.isPlayerAtLocation(e, x, y) {
		// ensure that a `wMap` entry exists. (this might be a reconnecting player)
		w.addPlayer(x, y, e)
		e.player.See(w)
		w.refreshActiveNPCs()
	}
	return e
}

func (w *World) refreshActiveNPCs() {
	// from each player, grab all NPCs within a boundary and make sure they are all in activeNPCs
	// set any remaining to inactive
	found := make(map[*entity]struct{})
	for _, e := range w.playerMap {
		px, py := e.player.GetLocation()
		x1 := util.ClampedInt(px-NPCActivationRadius, 0, w.W-1)
		y1 := util.ClampedInt(py-NPCActivationRadius, 0, w.H-1)
		x2 := util.ClampedInt(px+NPCActivationRadius, 0, w.W-1)
		y2 := util.ClampedInt(py+NPCActivationRadius, 0, w.H-1)
		w.grid.Within(tile.Point{X: int16(x1), Y: int16(y1)}, tile.Point{X: int16(x2), Y: int16(y2)}, func(point tile.Point, t tile.Tile) {
			_, _, n, _, _ := w.tileToEntities(t)
			if n != nil {
				found[n] = struct{}{}
			}
		})
	}
	newActiveNPCs := make([]*entity, 0)
	for e := range found {
		newActiveNPCs = append(newActiveNPCs, e)
	}
	w.activeNPCs = newActiveNPCs
}

func (w *World) getPlayer(playerID string) *entity {
	w.Lock()
	defer w.Unlock()
	e, _ := w.playerMap[playerID]
	return e
}

func (w *World) isPlayerAtLocation(e *entity, x, y int) bool {
	if t, ok := w.grid.At(int16(x), int16(y)); ok {
		p, _, _, _, _ := w.tileToEntities(t)
		return p == e
	}
	return false
}

func (w *World) randomAvailableCoord() (int, int, error) {
	tries := 1000
	for tries > 0 {
		x := rand.Intn(w.W)
		y := rand.Intn(w.H)
		if w.walkable(x, y) && !w.occupied(x, y) {
			return x, y, nil
		}
		tries--
	}
	return 0, 0, errors.New("couldn't place random coord")
}

func (w *World) InBounds(x, y int) bool {
	return x >= 0 && x < w.W && y >= 0 && y < w.H
}

func (w *World) IsOpaque(x, y int) bool {
	ents, _ := w.allEntitiesAndBiomeAt(x, y)
	for _, e := range ents {
		if !e.SeeThrough() {
			return true
		}
	}
	return false
}

func (w *World) walkable(x, y int) bool {
	ents, _ := w.allEntitiesAndBiomeAt(x, y)
	for _, e := range ents {
		if !e.Walkable() {
			return false
		}
	}
	return true
}

// todo this should consider a tile with an item as occupied
func (w *World) occupied(x, y int) bool {
	ents, _ := w.allEntitiesAndBiomeAt(x, y)
	for _, e := range ents {
		if e.Occupied() {
			return true
		}
	}
	return false
}

func (w *World) pickupable(x, y int) (*entity, bool) {
	ents, _ := w.allEntitiesAndBiomeAt(x, y)
	for _, e := range ents {
		if e.Pickupable() {
			return e, true
		}
	}
	return nil, false
}

func (w *World) attackable(x int, y int) (*entity, bool) {
	ents, _ := w.allEntitiesAndBiomeAt(x, y)
	for _, e := range ents {
		if e.Attackable() {
			return e, true
		}
	}
	return nil, false
}

func (w *World) harvestable(x int, y int) (*entity, bool) {
	ents, _ := w.allEntitiesAndBiomeAt(x, y)
	for _, e := range ents {
		if e.Harvestable() {
			return e, true
		}
	}
	return nil, false
}

func (w *World) index(x, y int) int {
	return y*w.W + x
}

func (w *World) coordinates(i int) (int, int) {
	x := i % w.W
	y := i / w.H
	return x, y
}

func (w *World) OnEvent(playerID string, f func(string)) {
	w.onEvent[playerID] = f
}

func (w *World) Event(kind events.Class, message string) {
	w.events.Add(kind, message)
	s := w.events.Render()
	for _, f := range w.onEvent {
		f(s)
	}
}

func (w *World) Chat(kind events.Class, subject, message string) {
	w.events.AddWithSubject(kind, message, subject)
	s := w.events.Render()
	for _, f := range w.onEvent {
		f(s)
	}
}

func (w *World) disconnectPlayer(e *entity) {
	w.Lock()
	defer w.Unlock()
	x, y := e.player.GetLocation()
	w.removePlayer(x, y, e)
}

func (w *World) neighbors(x, y int) []Coord {
	return []Coord{
		Coord{x + 1, y},
		Coord{x - 1, y},
		Coord{x, y - 1},
		Coord{x, y + 1},
	}
}

// findNearbyAvailableIndex does a simple BFS to find an empty location
func (w *World) findNearbyAvailableIndex(x int, y int) (int, int, error) {
	seen := map[Coord]struct{}{}
	stack := w.neighbors(x, y)
	for len(stack) > 0 {
		c := stack[0]
		stack = stack[1:]
		if !w.InBounds(c.X, c.Y) {
			continue
		}
		if !w.occupied(c.X, c.Y) && w.walkable(c.X, c.Y) {
			return c.X, c.Y, nil
		}
		_, ok := seen[c]
		if ok {
			continue
		}
		seen[c] = struct{}{}
		stack = append(stack, w.neighbors(c.X, c.Y)...)
	}
	return 0, 0, errors.New("could not find an available coordinate")
}

func (w *World) removeNPC(x int, y int) {
	if t, ok := w.grid.At(int16(x), int16(y)); ok {
		p, b, oldIndex, i, f := tileToIndexes(t)
		t = indexesToTile(p, b, 0, i, f)
		w.grid.WriteAt(int16(x), int16(y), t)
		w.npcs.Remove(int(oldIndex))
	}
}

func (w *World) addItem(x int, y int, item *Item, quantity int) {
	// todo this ignored error means the world is full of items! handle it
	i, _ := w.items.Append(&entity{item: item, quantity: quantity})
	if t, ok := w.grid.At(int16(x), int16(y)); ok {
		p, b, n, _, f := tileToIndexes(t)
		t = indexesToTile(p, b, n, uint16(i), f)
		w.grid.WriteAt(int16(x), int16(y), t)
	}
}

func (w *World) removeItem(x int, y int) {
	if t, ok := w.grid.At(int16(x), int16(y)); ok {
		p, b, n, oldIndex, f := tileToIndexes(t)
		t = indexesToTile(p, b, n, 0, f)
		w.grid.WriteAt(int16(x), int16(y), t)
		w.items.Remove(int(oldIndex))
	}
}

func (w *World) swapPlayer(px int, py int, nx int, ny int, e *entity) {
	var playerIndex uint8
	// remove player from old spot
	if t, ok := w.grid.At(int16(px), int16(py)); ok {
		pi, b, n, i, f := tileToIndexes(t)
		fmt.Println("got indexes", pi, b, n, i, f)
		playerIndex = pi
		t = indexesToTile(0, b, n, i, f)
		fmt.Println("write tile", t)
		w.grid.WriteAt(int16(px), int16(py), t)
	}
	// set player at new location
	if t, ok := w.grid.At(int16(nx), int16(ny)); ok {
		_, b, n, i, f := tileToIndexes(t)
		t = indexesToTile(playerIndex, b, n, i, f)
		w.grid.WriteAt(int16(px), int16(py), t)
	}
}

func (w *World) swapNPC(px int, py int, nx int, ny int, e *entity) {
	var npcIndex uint8
	// remove npc from old spot
	if t, ok := w.grid.At(int16(px), int16(py)); ok {
		p, b, ni, i, f := tileToIndexes(t)
		npcIndex = ni
		t = indexesToTile(p, b, 0, i, f)
		w.grid.WriteAt(int16(px), int16(py), t)
	}
	// set npc at new location
	if t, ok := w.grid.At(int16(nx), int16(ny)); ok {
		p, b, _, i, f := tileToIndexes(t)
		t = indexesToTile(p, b, npcIndex, i, f)
		w.grid.WriteAt(int16(px), int16(py), t)
	}
}

func (w *World) removeFlora(x int, y int, ent *entity) {
	if t, ok := w.grid.At(int16(x), int16(y)); ok {
		p, b, n, i, oldIndex := tileToIndexes(t)
		t = indexesToTile(p, b, n, i, 0)
		w.grid.WriteAt(int16(x), int16(y), t)
		w.flora.Remove(int(oldIndex))
	}
}

func (w *World) removePlayer(x int, y int, ent *entity) {
	if t, ok := w.grid.At(int16(x), int16(y)); ok {
		oldIndex, b, n, i, f := tileToIndexes(t)
		t = indexesToTile(0, b, n, i, f)
		w.grid.WriteAt(int16(x), int16(y), t)
		w.players.Remove(int(oldIndex))
	}
}

func (w *World) tileToEntities(t tile.Tile) (*entity, Biome, *entity, *entity, *entity) {
	p, b, n, i, f := tileToIndexes(t)
	var ply *entity
	var npc *entity
	var item *entity
	var flora *entity
	if p > 0 {
		ply = w.players.Get(int(p))
	}
	if n > 0 {
		npc = w.npcs.Get(int(n))
	}
	if i > 0 {
		item = w.items.Get(int(i))
	}
	if f > 0 {
		flora = w.flora.Get(int(f))
	}
	return ply, Biome(b), npc, item, flora
}

func (w *World) allEntitiesFromTile(t tile.Tile) []*entity {
	p, _, n, i, f := w.tileToEntities(t)
	return []*entity{p, n, i, f}
}

func (w *World) allEntitiesAndBiomeAt(x, y int) ([]*entity, Biome) {
	var out []*entity
	var biome Biome
	if t, ok := w.grid.At(int16(x), int16(y)); ok {
		_, b, _, _, _ := tileToIndexes(t)
		biome = Biome(b)
		ents := w.allEntitiesFromTile(t)
		for _, ent := range ents {
			if ent != nil {
				out = append(out, ent)
			}
		}
	}
	return out, biome
}

func (w *World) addPlayer(x int, y int, e *entity) {
	// todo this ignored error means the world has too many players! handle it
	playerIndex, _ := w.players.Append(e)
	if t, ok := w.grid.At(int16(x), int16(y)); ok {
		_, b, n, i, f := tileToIndexes(t)
		t = indexesToTile(uint8(playerIndex), b, n, i, f)
		w.grid.WriteAt(int16(x), int16(y), t)
	}
}

// compassIndicator returns a string like "↖ 30"
func compassIndicator(fromX, fromY, toX, toY int) string {
	dx := toX - fromX
	dy := toY - fromY
	deltaX := util.ClampedInt(dx, -1, 1)
	deltaY := util.ClampedInt(dy, -1, 1)
	arrow, ok := util.Arrows[[2]int{deltaX, deltaY}]
	out := ""
	if ok {
		out = arrow + " "
	}
	return out + strconv.Itoa(util.AbsInt(dx)+util.AbsInt(dy))
}
