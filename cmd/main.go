package main

import (
	"fmt"
	"log"
	"net/http"
	"pokemon-battle-simulator/internal/battle"
	"pokemon-battle-simulator/internal/load"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BattleStatus struct {
	Status string               `json:"status"`
	Result *battle.BattleResult `json:"result,omitempty"`
}

var (
	pokemons []load.LoadPokemon
	battles  map[string]*BattleStatus
	mutex    sync.Mutex
)

func main() {
	var err error
	pokemons, err = load.LoadDataset("/home/debianism/projects/Pokemon-Battleship/internal/load/filtered_pokemon.csv")
	if err != nil {
		panic(err)
	}

	battles = make(map[string]*BattleStatus)

	r := gin.Default()

	// API 1: Listing API with pagination
	r.GET("/pokemon", listPokemon)

	// API 2: Battle API
	r.POST("/battle", startBattle)

	// API 3: Battle Status API
	r.GET("/battle/:id", getBattleStatus)

	r.Run(":8000")
}

func listPokemon(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size"})
		return
	}

	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= len(pokemons) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page number out of range"})
		return
	}

	if endIndex > len(pokemons) {
		endIndex = len(pokemons)
	}

	paginatedPokemons := pokemons[startIndex:endIndex]

	c.JSON(http.StatusOK, gin.H{
		"page":          page,
		"pageSize":      pageSize,
		"totalPokemons": len(pokemons),
		"pokemons":      paginatedPokemons,
	})
}

func startBattle(c *gin.Context) {
	var request struct {
		PokemonA string `json:"pokemonA"`
		PokemonB string `json:"pokemonB"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	battleID := uuid.New().String()

	mutex.Lock()
	battles[battleID] = &BattleStatus{Status: "BATTLE_INPROGRESS"}
	mutex.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered in startBattle goroutine: %v", r)
				mutex.Lock()
				battles[battleID].Status = "BATTLE_FAILED"
				mutex.Unlock()
			}
		}()

		pokemonA, err := load.GetPokemonByName(request.PokemonA, pokemons)
		if err != nil {
			panic(fmt.Sprintf("Error getting Pokemon A: %v", err))
		}

		pokemonB, err := load.GetPokemonByName(request.PokemonB, pokemons)
		if err != nil {
			panic(fmt.Sprintf("Error getting Pokemon B: %v", err))
		}

		if pokemonA == nil || pokemonB == nil {
			panic("One or both Pokemon are nil")
		}

		battlePokemonA := battle.BattlePokemon{BasePokemon: pokemonA.BasePokemon}
		battlePokemonB := battle.BattlePokemon{BasePokemon: pokemonB.BasePokemon}

		result, err := battle.Battle(battlePokemonA, battlePokemonB)
		if err != nil {
			panic(fmt.Sprintf("Error in battle: %v", err))
		}

		mutex.Lock()
		battles[battleID].Status = "BATTLE_COMPLETED"
		battles[battleID].Result = result
		mutex.Unlock()
	}()

	c.JSON(http.StatusOK, gin.H{"battleID": battleID})
}

func getBattleStatus(c *gin.Context) {
	battleID := c.Param("id")

	mutex.Lock()
	status, exists := battles[battleID]
	mutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Battle not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}
