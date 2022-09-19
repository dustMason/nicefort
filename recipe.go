package main

import (
	"strconv"
	"strings"
)

// crafting conditions are flexible: things can be crafted depending on where
// the player is, what time of day it is, whether the player is standing next to
// something specific, etc

type Recipe struct {
	description string
	condition   condition
	result      *item
	id          int
}

func AvailableRecipes(im map[int]*InventoryItem, e *entity, w *World) []Recipe {
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
		if r.id == id {
			ok = true
			found = r
			break
		}
	}
	return ok, found
}

type condition func(map[int]*InventoryItem, *entity, *World) (bool, map[int]int)

func NewSimpleRecipe(result *item, id int, ing ...InventoryItem) Recipe {
	parts := make([]string, len(ing))
	for i, ii := range ing {
		parts[i] = ii.item.name + " x " + strconv.Itoa(ii.quantity)
	}
	return Recipe{
		description: strings.Join(parts, ", "),
		condition:   ingredientsCondition(ing...),
		result:      result,
		id:          id,
	}
}

func (r *Recipe) Check(inv map[int]*InventoryItem, e *entity, w *World) bool {
	ok, _ := r.condition(inv, e, w)
	return ok
}

func (r *Recipe) Do(inv map[int]*InventoryItem, e *entity, w *World) (bool, map[int]*InventoryItem) {
	ok, cost := r.condition(inv, e, w)
	if !ok {
		return false, inv
	}
	for id, q := range cost {
		inv[id].quantity -= q
	}
	if _, ok := inv[r.result.id]; !ok {
		inv[r.result.id] = &InventoryItem{item: r.result}
	}
	inv[r.result.id].quantity += 1
	return true, inv
}

func ingredientsCondition(ingredients ...InventoryItem) condition {
	return func(inventoryMap map[int]*InventoryItem, e *entity, w *World) (bool, map[int]int) {
		out := make(map[int]int)
		for _, i := range ingredients {
			pi, ok := inventoryMap[i.item.id]
			if !ok || pi.quantity < i.quantity {
				return false, out
			}
			out[i.item.id] = i.quantity
		}
		return true, out
	}
}

var AllRecipes = []Recipe{
	NewSimpleRecipe(&TestItem3, 1,
		InventoryItem{item: &TestItem, quantity: 3},
		InventoryItem{item: &TestItem2, quantity: 1},
	),
	NewSimpleRecipe(&TestItem4, 2,
		InventoryItem{item: &TestItem3, quantity: 1},
		InventoryItem{item: &TestItem, quantity: 10},
	),
}
