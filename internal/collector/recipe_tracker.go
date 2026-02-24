package collector

import (
	"sync"
	"time"

	"github.com/robinson/gouads"
)

type RecipeTracker struct {
	conn         *gouads.Connection
	recipeFields []string

	// Current recipe context
	currentRecipe *RecipeContext
	mu            sync.RWMutex

	// Recipe boundaries
	recipeStartFlag string // e.g., "Recipe.recipe_exe.recipeexecute.Done"
	recipeStartTime string // "Recipe.recipe_start_time"
	recipeEndTime   string // "Recipe.recipe_end_time"

	// Channels
	recipeStartChan chan *RecipeContext
	recipeEndChan   chan *RecipeContext
}

type RecipeContext struct {
	RecipeID    string
	ProcessJob  string
	SubstrateID string
	Algorithm   string
	StartTime   time.Time
	EndTime     time.Time
	StepIndex   int
	LoopIndex   int
	Status      string
	Fields      map[string]interface{}
}

func NewRecipeTracker(conn *gouads.Connection) *RecipeTracker {
	return &RecipeTracker{
		conn:            conn,
		recipeStartChan: make(chan *RecipeContext, 10),
		recipeEndChan:   make(chan *RecipeContext, 10),
		recipeStartFlag: "Recipe.recipe_exe.recipeexecute.Done",
		recipeStartTime: "Recipe.recipe_start_time",
		recipeEndTime:   "Recipe.recipe_end_time",
	}
}

func (rt *RecipeTracker) Start() {
	go rt.monitorRecipeBoundaries()
}

func (rt *RecipeTracker) monitorRecipeBoundaries() {
	var lastDone bool

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// Check if recipe is done (transition from false to true)
		done, err := rt.conn.ReadBool(rt.recipeStartFlag)
		if err != nil {
			continue
		}

		if done && !lastDone {
			rt.handleRecipeStart()
		} else if !done && lastDone {
			rt.handleRecipeEnd()
		}

		lastDone = done
	}
}

func (rt *RecipeTracker) handleRecipeStart() {
	// Read all recipe fields
	recipeData := make(map[string]interface{})
	for _, field := range rt.recipeFields {
		val, err := rt.conn.ReadValue(field)
		if err == nil {
			recipeData[field] = val
		}
	}

	// Parse start time
	startTimeStr, _ := rt.conn.ReadString(rt.recipeStartTime)
	startTime, _ := time.Parse("2006-01-02 15:04:05", startTimeStr)

	context := &RecipeContext{
		ProcessJob:  interfaceToString(recipeData["Recipe.recipe_exe.Process_Job"]),
		SubstrateID: interfaceToString(recipeData["Recipe.recipe_exe.Substrate_ID"]),
		Algorithm:   interfaceToString(recipeData["Recipe.recipe_exe.EP_Algorithm"]),
		StartTime:   startTime,
		Fields:      recipeData,
	}

	rt.mu.Lock()
	rt.currentRecipe = context
	rt.mu.Unlock()

	select {
	case rt.recipeStartChan <- context:
	default:
	}
}

func (rt *RecipeTracker) handleRecipeEnd() {
	rt.mu.RLock()
	context := rt.currentRecipe
	rt.mu.RUnlock()

	if context == nil {
		return
	}

	endTimeStr, _ := rt.conn.ReadString(rt.recipeEndTime)
	endTime, _ := time.Parse("2006-01-02 15:04:05", endTimeStr)
	context.EndTime = endTime

	context.Status, _ = rt.conn.ReadString("Recipe.Status")

	select {
	case rt.recipeEndChan <- context:
	default:
	}
}

func interfaceToString(i interface{}) string {
	if s, ok := i.(string); ok {
		return s
	}
	return ""
}
