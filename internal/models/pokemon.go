package models

type BasePokemon struct {
	Name    string
	Type1   string
	Type2   string
	Attack  int
	Against map[string]float64
}
