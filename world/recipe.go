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

func newRecipe(result *Item, id int, description string, conditions ...condition) Recipe {
	mergedConditions := func(ii map[string]*InventoryItem, e *entity, w *World) (bool, map[string]int) {
		out := make(map[string]int)
		for _, f := range conditions {
			ok, newInv := f(ii, e, w)
			if !ok {
				return false, out
			}
			for s, q := range newInv {
				out[s] += q
			}
		}
		return true, out
	}
	return Recipe{
		Description: description,
		condition:   mergedConditions,
		Result:      result,
		ID:          id,
	}

}

func newSimpleRecipe(result *Item, id int, ing ...InventoryItem) Recipe {
	parts := make([]string, len(ing))
	for i, ii := range ing {
		parts[i] = ii.Item.Name + " x " + strconv.Itoa(ii.Quantity)
	}
	return newRecipe(result, id, strings.Join(parts, ", "), ingredientsCondition(ing...))
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

// todo this should be smarter and should allow a combination of many items that have the right traits
func traitMatchingCondition(trait ItemTraits, quantity int) condition {
	return func(inventoryMap map[string]*InventoryItem, e *entity, w *World) (bool, map[string]int) {
		// look at inventoryMap and find the first item that matches trait, and that we have `quantity` of
		out := make(map[string]int)
		for _, ii := range inventoryMap {
			// does this item have the trait?
			if ii.Item.HasTrait(trait) && ii.Quantity >= quantity {
				out[ii.Item.ID] = quantity
				return true, out
			}
		}
		return false, out
	}
}

var AllRecipes = []Recipe{
	newSimpleRecipe(DriedLeaves, 1, InventoryItem{Item: BogMyrtleLeaves, Quantity: 1}),
	newSimpleRecipe(Twine, 2, InventoryItem{Item: HalberdLeavedWillowSticks, Quantity: 5}),
	newRecipe(
		FireStarterBow,
		3,
		"Twine x 3, Sticks x 3, Kindling x 3",
		ingredientsCondition(InventoryItem{Item: Twine, Quantity: 3}),
		traitMatchingCondition(Kindling, 3),
		traitMatchingCondition(Stick, 3),
	),
	newRecipe(Campfire, 4,
		"A fire starter bow and some fuel (wood).",
		ingredientsCondition(InventoryItem{Item: FireStarterBow, Quantity: 1}),
		traitMatchingCondition(Fuel, 1),
	),
}
