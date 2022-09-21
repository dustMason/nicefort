package world

import (
	"strconv"
	"strings"
)

// crafting conditions are flexible: things can be crafted depending on where
// the player is, what time of day it is, whether the player is standing next to
// something specific, etc

type Recipe struct {
	Description string
	Result      *Item
	ID          int
	condition   condition
}

func AvailableRecipes(im map[string]*InventoryItem, e *entity, w *World) []Recipe {
	out := make([]Recipe, 0)
	for _, r := range AllRecipes {
		if r.Check(im, e, w) {
			out = append(out, r)
		}
	}
	return out
}

func FindRecipe(id int) (bool, Recipe) {
	ok := false
	found := Recipe{}
	for _, r := range AllRecipes {
		if r.ID == id {
			ok = true
			found = r
			break
		}
	}
	return ok, found
}

type condition func(map[string]*InventoryItem, *entity, *World) (bool, map[string]int)

func NewSimpleRecipe(result *Item, id int, ing ...InventoryItem) Recipe {
	parts := make([]string, len(ing))
	for i, ii := range ing {
		parts[i] = ii.Item.Name + " x " + strconv.Itoa(ii.Quantity)
	}
	return Recipe{
		Description: strings.Join(parts, ", "),
		condition:   ingredientsCondition(ing...),
		Result:      result,
		ID:          id,
	}
}

func (r *Recipe) Check(inv map[string]*InventoryItem, e *entity, w *World) bool {
	ok, _ := r.condition(inv, e, w)
	return ok
}

func (r *Recipe) Do(inv map[string]*InventoryItem, e *entity, w *World) (bool, map[string]*InventoryItem) {
	ok, cost := r.condition(inv, e, w)
	if !ok {
		return false, inv
	}
	for id, q := range cost {
		inv[id].Quantity -= q
	}
	if _, ok := inv[r.Result.ID]; !ok {
		inv[r.Result.ID] = &InventoryItem{Item: r.Result}
	}
	inv[r.Result.ID].Quantity += 1
	return true, inv
}

// todo make condition func that checks if we are standing next to a cooking fire
// and if we have a watertight cooking vessel

func ingredientsCondition(ingredients ...InventoryItem) condition {
	return func(inventoryMap map[string]*InventoryItem, e *entity, w *World) (bool, map[string]int) {
		out := make(map[string]int)
		for _, i := range ingredients {
			pi, ok := inventoryMap[i.Item.ID]
			if !ok || pi.Quantity < i.Quantity {
				return false, out
			}
			out[i.Item.ID] = i.Quantity
		}
		return true, out
	}
}

var AllRecipes = []Recipe{
	// NewSimpleRecipe(&TestItem3, 1,
	// 	InventoryItem{Item: &TestItem, Quantity: 3},
	// 	InventoryItem{Item: &TestItem2, Quantity: 1},
	// ),
	// NewSimpleRecipe(&TestItem4, 2,
	// 	InventoryItem{Item: &TestItem3, Quantity: 1},
	// 	InventoryItem{Item: &TestItem, Quantity: 10},
	// ),
}
