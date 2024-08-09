package load

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"pokemon-battle-simulator/internal/models"
	"regexp"
	"strconv"
	"strings"
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

		// Create a new Pokemon struct
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
var ErrTooManySpellingMistakes = errors.New("too many spelling mistakes in Pokemon name")

func GetPokemonByName(name string, pokemons []LoadPokemon) (*LoadPokemon, error) {
	name = strings.ToLower(name)

	// Exact match check
	for _, p := range pokemons {
		if strings.ToLower(p.Name) == name {
			return &p, nil
		}
	}

	// Regexp for one-word spelling mistake
	pattern := "^" + strings.Join(strings.Split(name, ""), "?.?") + "?$"
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	var matchedPokemon *LoadPokemon
	for _, p := range pokemons {
		if re.MatchString(strings.ToLower(p.Name)) {
			if matchedPokemon != nil {
				// More than one match found, consider it as too many spelling mistakes
				return nil, ErrTooManySpellingMistakes
			}
			matchedPokemon = &p
		}
	}

	if matchedPokemon != nil {
		return matchedPokemon, nil
	}

	return nil, ErrPokemonNotFound
}
