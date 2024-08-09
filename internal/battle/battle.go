package battle

import (
	"errors"
	"pokemon-battle-simulator/internal/models"
	"strings"
)

type BattlePokemon struct {
	models.BasePokemon
	Damage float64
}

type BattleResult struct {
	WinnerName  string
	WonByMargin float64
}

func (bp *BattlePokemon) CalculateDamage(attacker, defender BattlePokemon) float64 {
	againstType1 := defender.Against[strings.ToLower(attacker.Type1)]
	againstType2 := defender.Against[strings.ToLower(attacker.Type2)]

	damage := (float64(attacker.Attack) / 200) * 100
	damage -= ((againstType1 / 4) * 100) + ((againstType2 / 4) * 100)
	return damage
}

func Battle(pokemonA, pokemonB BattlePokemon) (*BattleResult, error) {
	if pokemonA.Name == pokemonB.Name {
		return nil, errors.New("both Pokémon are the same")
	}

	// Round 1: A attacks B
	damageAtoB := pokemonA.CalculateDamage(pokemonA, pokemonB)

	// Round 2: B attacks A
	damageBtoA := pokemonB.CalculateDamage(pokemonB, pokemonA)

	var winnerName string
	var wonByMargin float64

	if damageAtoB > damageBtoA {
		winnerName = pokemonA.Name
		wonByMargin = damageAtoB
	} else if damageBtoA > damageAtoB {
		winnerName = pokemonB.Name
		wonByMargin = damageBtoA
	} else {
		return nil, nil // It's a draw
	}

	return &BattleResult{
		WinnerName:  winnerName,
		WonByMargin: wonByMargin,
	}, nil
}
