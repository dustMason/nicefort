package world

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustmason/nicefort/events"
	"github.com/dustmason/nicefort/util"
	"math"
	"math/rand"
	"sort"
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

func (c Coord) Distance(loc Coord) int {
	vx := math.Pow(float64(c.X-loc.X), 2)
	vy := math.Pow(float64(c.Y-loc.Y), 2)
	return int(math.Sqrt(vx + vy))
}

func (c Coord) IsZero() bool {
	return c.X == 0 && c.Y == 0
}

type World struct {
	sync.RWMutex
	W, H       int
	wMap       []location         // the actual map of tiles
	players    map[string]*entity // map of player id => entity that points to that player
	activeNPCs []*entity
	events     *events.EventList
	onEvent    map[string]func(string) // map of player id => chat callback
	days       float64                 // age of the world
	lastTick   time.Time
}

func NewWorld(size int) *World {
	w := &World{
		W:        size,
		H:        size,
		players:  make(map[string]*entity),
		wMap:     GenerateOverworld(size),
		events:   events.NewEventList(100),
		onEvent:  make(map[string]func(string)),
		lastTick: time.Now(),
	}
	go w.runTicker()
	return w
}

type location []*entity

func removeEntity(l location, e *entity) location {
	ret := make(location, 0)
	for _, ent := range l {
		if ent != e {
			ret = append(ret, ent)
		}
	}
	return ret
}

func addEntity(l location, e *entity) location {
	ret := make(location, 0)
	for _, ent := range l {
		ret = append(ret, ent)
	}
	ret = append(ret, e)
	return ret

}

func (w *World) tick(t time.Time) {
	w.days += t.Sub(w.lastTick).Seconds() / secondsPerDay
	w.lastTick = t
	for _, e := range w.activeNPCs {
		e.npc.Tick(t, w, e)
	}
	for _, e := range w.players {
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
	newIndex := w.index(nx, ny)
	if w.InBounds(nx, ny) {
		if ent, ok := w.attackable(nx, ny); ok {
			damage, success, dead, drops := ent.npc.Attacked(e.player.wielding, e, 10)
			// todo need a progress calc to use Activity
			if dead {
				e.player.Event(events.Success, fmt.Sprintf("You killed the %s", ent.npc.Name))
				i := w.index(nx, ny)
				w.wMap[i] = removeEntity(w.wMap[i], ent)
				for _, drop := range drops {
					e.player.Event(events.Success, fmt.Sprintf("It dropped %d x %s", drop.Quantity, drop.Item.Name))
					ni, _ := w.findNearbyAvailableIndex(nx, ny)
					w.wMap[ni] = addEntity(w.wMap[ni], &entity{item: drop.Item, quantity: drop.Quantity})
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
			oldIndex := w.index(oldX, oldY)
			w.wMap[newIndex] = append(w.wMap[newIndex], e)
			w.wMap[oldIndex] = removeEntity(w.wMap[oldIndex], e)
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
	index := w.index(x, y)
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
			w.wMap[index] = removeEntity(w.wMap[index], ee)
		}
		return
	}
}

func (w *World) harvest(player *player, ent *entity, x, y int) {
	dead, success, progress, drops := ent.flora.Harvest(player.wielding)
	player.SetActivity(Activity{description: ent.flora.name, progress: progress})
	for _, drop := range drops {
		player.Event(events.Success, fmt.Sprintf("It yielded %d x %s", drop.Quantity, drop.Item.Name))
		i, _ := w.findNearbyAvailableIndex(x, y)
		w.wMap[i] = addEntity(w.wMap[i], &entity{item: drop.Item, quantity: drop.Quantity})
	}
	if dead {
		player.Event(events.Success, fmt.Sprintf("You harvested the %s", ent.flora.name))
		i := w.index(x, y)
		w.wMap[i] = removeEntity(w.wMap[i], ent)
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
		newIndex := w.index(x, y)
		oldIndex := w.index(e.npc.loc.X, e.npc.loc.Y)
		w.wMap[newIndex] = append(w.wMap[newIndex], e)
		w.wMap[oldIndex] = removeEntity(w.wMap[oldIndex], e)
		e.npc.loc = Coord{x, y}
	}
}

func (w *World) ActivateItem(playerID string, inventoryIndex int) {
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
		// todo one or more items in newInv might be nonPortable. place them
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
	ix := left
	iy := top
	for iy < bottom {
		for ix < right {
			var ent *entity
			if !w.InBounds(ix, iy) {
				b.WriteString(blackSpace)
			} else {
				ic := Coord{ix, iy}
				inView, dist := ply.CanSee(ix, iy)
				memString, inPastView := seen[ic]
				if inPastView && !inView {
					b.WriteString(memString)
				} else if inView {
					loc := w.location(ix, iy)
					ent = loc[len(loc)-1]
					bg := loc[0]
					b.WriteString(
						lipgloss.NewStyle().
							Foreground(lipgloss.Color(ent.ForegroundColor(dist))).
							Background(lipgloss.Color(bg.BackgroundColor(dist))).
							Render(ent.String()),
					)
				} else { // not in past or current view
					b.WriteString(blackSpace)
				}
			}
			ix++
		}
		ix = left
		b.WriteString("\n")
		iy++
	}
	return b.String()
}

func (w *World) RenderForMemory(x, y int) string {
	loc := w.location(x, y)
	// iterate backwards to get topmost memorable entity first
	for i := len(loc) - 1; i >= 0; i-- {
		if loc[i].Memorable() {
			return memStyle.Render(loc[i].String())
		}
	}
	return blackSpace
}

// todo refactor this to return some value type (map of string[string]?) or a struct
func (w *World) RenderPlayerSidebar(id string, name string) string {
	var b strings.Builder
	e := w.getOrCreatePlayer(id, name)
	myLoc := e.player.loc
	b.WriteString(name + "\n\n")
	b.WriteString(fmt.Sprintf("Pack: %.1f / %d\n", e.player.carrying, int(e.player.maxCarry)))
	b.WriteString(fmt.Sprintf("Health: %d / %d\n", e.player.health, e.player.maxHealth))
	b.WriteString(fmt.Sprintf("Hunger: %.3f\n", e.player.hunger))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s\n", e.player.wielding.Name))
	b.WriteString("\n")

	a := e.player.GetActivity()
	if a.description != "" {
		b.WriteString(fmt.Sprintf("%s\n%s\n\n", a.description, a.pBar.ViewAs(a.progress)))
	}

	w.withSortedPlayers(func(ee *entity) {
		if ee == e {
			return
		}
		pLoc := ee.player.loc
		b.WriteString(
			compassIndicator(myLoc.X, myLoc.Y, pLoc.X, pLoc.Y) + " " + ee.player.name + "\n",
		)
	})

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
	e, ok := w.players[playerID]
	if !ok {
		x, y, _ := w.randomAvailableCoord()
		e = NewPlayer(playerID, Coord{x, y})
		w.players[playerID] = e
	}
	e.player.name = playerName
	x, y := e.player.GetLocation()
	if !w.isPlayerAtLocation(e, x, y) {
		// ensure that a `wMap` entry exists. (this might be a reconnecting player)
		i := w.index(x, y)
		w.wMap[i] = append(w.wMap[i], e)
		e.player.See(w)
		w.refreshActiveNPCs()
	}
	return e
}

func (w *World) refreshActiveNPCs() {
	// from each player, grab all NPCs within a boundary and make sure they are all in activeNPCs
	// set any remaining to inactive
	found := make(map[*entity]struct{})
	for _, e := range w.players {
		px, py := e.player.GetLocation()
		x1 := util.ClampedInt(px-NPCActivationRadius, 0, w.W-1)
		y1 := util.ClampedInt(py-NPCActivationRadius, 0, w.H-1)
		x2 := util.ClampedInt(px+NPCActivationRadius, 0, w.W-1)
		y2 := util.ClampedInt(py+NPCActivationRadius, 0, w.H-1)
		ix := x1
		iy := y1
		for iy < y2 {
			for ix < x2 {
				for _, ent := range w.location(ix, iy) {
					if ent.npc != nil {
						found[ent] = struct{}{}
					}
				}
				ix++
			}
			ix = x1
			iy++
		}
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
	e, _ := w.players[playerID]
	return e
}

func (w *World) isPlayerAtLocation(e *entity, x, y int) bool {
	for _, ent := range w.location(x, y) {
		if ent == e {
			return true
		}
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
	for _, e := range w.location(x, y) {
		if !e.SeeThrough() {
			return true
		}
	}
	return false
}

func (w *World) walkable(x, y int) bool {
	for _, e := range w.location(x, y) {
		if !e.Walkable() {
			return false
		}
	}
	return true
}

func (w *World) occupied(x, y int) bool {
	for _, e := range w.location(x, y) {
		if e.Occupied() {
			return true
		}
	}
	return false
}

func (w *World) pickupable(x, y int) (*entity, bool) {
	for _, e := range w.location(x, y) {
		if e.Pickupable() {
			return e, true
		}
	}
	return nil, false
}

func (w *World) attackable(x int, y int) (*entity, bool) {
	for _, e := range w.location(x, y) {
		if e.Attackable() {
			return e, true
		}
	}
	return nil, false
}

func (w *World) harvestable(x int, y int) (*entity, bool) {
	for _, e := range w.location(x, y) {
		if e.Harvestable() {
			return e, true
		}
	}
	return nil, false
}

func (w *World) index(x, y int) int {
	return y*w.W + x
}

func (w *World) location(x, y int) location {
	return w.wMap[w.index(x, y)]
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
	i := w.index(x, y)
	w.wMap[i] = removeEntity(w.wMap[i], e)
}

func (w *World) neighbors(x, y int) []Coord {
	return []Coord{
		{x + 1, y},
		{x - 1, y},
		{x, y - 1},
		{x, y + 1},
	}
}

// findNearbyAvailableIndex does a simple BFS to find an empty location
func (w *World) findNearbyAvailableIndex(x int, y int) (int, error) {
	seen := map[Coord]struct{}{}
	stack := w.neighbors(x, y)
	for len(stack) > 0 {
		c := stack[0]
		stack = stack[1:]
		if !w.InBounds(c.X, c.Y) {
			continue
		}
		if !w.occupied(c.X, c.Y) && w.walkable(c.X, c.Y) {
			return w.index(c.X, c.Y), nil
		}
		_, ok := seen[c]
		if ok {
			continue
		}
		seen[c] = struct{}{}
		stack = append(stack, w.neighbors(c.X, c.Y)...)
	}
	return 0, errors.New("could not find an available coordinate")
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

func (w *World) withSortedPlayers(f func(e *entity)) {
	keys := make([]string, 0, len(w.players))
	for k := range w.players {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		f(w.players[k])
	}
}
