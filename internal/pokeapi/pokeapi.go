package pokeapi

import (
	"strings"
	"encoding/json"
	"io"
	"net/http"
	"fmt"
)

type locationArea struct {
	PokemonEncounters	[]struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Cacher interface {
    Get(key string) ([]byte, bool)
    Add(key string, val []byte)
}

const baseLocationAreaDetailURL = "https://pokeapi.co/api/v2/location-area/"

func GetLocationArea(name string) string {
	trimmed := strings.TrimSpace(name)
	fullUrl := baseLocationAreaDetailURL + trimmed
	return fullUrl
}

func GetPokemonData(cache Cacher, name string) error {
	url := GetLocationArea(name)
	var b []byte
	if cached, ok := cache.Get(url); ok {
		b = cached
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		cache.Add(url, body)
		b = body
	}
	
	var out locationArea
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	
	for _, r := range out.PokemonEncounters {
		fmt.Println(r.Pokemon.Name)
	}

	return nil
}
