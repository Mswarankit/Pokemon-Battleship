package load

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"pokemon-battle-simulator/internal/models"
	"strconv"
	"strings"

	"github.com/agnivade/levenshtein"
)

type LoadPokemon struct {
	models.BasePokemon
}

func LoadDataset(filePath string) ([]LoadPokemon, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	header := records[0]
	var pokemons []LoadPokemon

	for _, record := range records[1:] {
		attackFloat, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing attack value as float: %w", err)
		}

		attack := int(math.Round(attackFloat))

		against := make(map[string]float64)
		for i, h := range header {
			if strings.HasPrefix(h, "against_") {
				value, err := strconv.ParseFloat(record[i], 32)
				if err != nil {
					return nil, fmt.Errorf("error parsing %s as float: %w", h, err)
				}
				against[h[8:]] = value
			}
		}

		pokemon := LoadPokemon{
			BasePokemon: models.BasePokemon{
				Name:    record[0],
				Type1:   record[1],
				Type2:   record[2],
				Attack:  attack,
				Against: against,
			},
		}
		pokemons = append(pokemons, pokemon)
	}

	return pokemons, nil
}

var ErrPokemonNotFound = errors.New("pokemon not found")

func GetPokemonByName(name string, pokemons []LoadPokemon) (*LoadPokemon, error) {
	name = strings.ToLower(name)

	for _, p := range pokemons {
		if strings.ToLower(p.Name) == name {
			return &p, nil
		}
	}

	var bestMatch *LoadPokemon
	minDistance := len(name)

	for _, p := range pokemons {
		distance := levenshtein.ComputeDistance(name, strings.ToLower(p.Name))

		if distance == 1 {
			return &p, nil
		}

		if distance < minDistance {
			minDistance = distance
			bestMatch = &p
		}
	}

	if minDistance <= 2 {
		return bestMatch, nil
	}

	return nil, ErrPokemonNotFound
}
