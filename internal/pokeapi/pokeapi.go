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

type Pokemon struct {
	Name			string 	`json:"name"`
	BaseExperience	int 	`json:"base_experience"`
	Height			int		`json:"height"`
	Weight			int		`json:"weight"`
	Stats []struct {
		BaseStat 	int		`json:"base_stat"`
		Stat		struct {
			Name	string 	`json:"name"`
		} 					`json:"stat"`
	} 						`json:"stats"`
	Types []struct {
		Type struct {
			Name 	string 	`json:"name"`
		} 					`json:"type"`
	} 						`json:"types"`
}

type Cacher interface {
    Get(key string) ([]byte, bool)
    Add(key string, val []byte)
}

const baseLocationAreaDetailURL = "https://pokeapi.co/api/v2/location-area/"

const basePokemonDataURL = "https://pokeapi.co/api/v2/pokemon/"

func buildLocationURL(name string) string {
	trimmed := strings.TrimSpace(name)
	fullUrl := baseLocationAreaDetailURL + trimmed
	return fullUrl
}

func buildPokemonURL(name string) string {
	trimmed := strings.TrimSpace(name)
	fullURL := basePokemonDataURL + trimmed
	return fullURL
}

func GetPokemonData(cache Cacher, name string) (Pokemon, error) {
	var zero Pokemon
	url := buildPokemonURL(name)
	var b []byte
	if cached, ok := cache.Get(url); ok {
		b = cached
	} else {
		res, err := http.Get(url)
		if err != nil {
			return zero, err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return zero, err
		}
		cache.Add(url, body)
		b = body
	}
	
	var out Pokemon
	if err := json.Unmarshal(b, &out); err != nil {
		return zero, err
	}
	return out, nil
}


func GetLocationArea(cache Cacher, name string) error {
	url := buildLocationURL(name)
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
	
	for _, r := range out.PokemonEncounters{
		fmt.Println(r.Pokemon.Name)
	}

	return nil
}

