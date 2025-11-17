package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"pokedexcli/internal/pokeapi"
	"pokedexcli/internal/pokecache"
	"sort"
	"strings"
	"time"
	"math/rand"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, []string) error
}

type Config struct {
	Next		string
	Previous	string
	Cache 		*pokecache.Cache
	Caught		map[string]pokeapi.Pokemon
}

type locationAreaList struct {
    Next     *string                `json:"next"`
    Previous *string                `json:"previous"`
    Results  []struct{ Name string `json:"name"` } `json:"results"`
}

var commands map[string]cliCommand

var cfg = &Config{}

const baseLocationAreaURL = "https://pokeapi.co/api/v2/location-area?limit=20"

func NewRegistry() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:			"exit",
			description:	"		Exit the Pokedex",
			callback:    	commandExit,
		},
		"help": {
			name:         	"help",
			description:  	"		Displays a help message",
			callback:     	commandHelp,    
		},
		"map": {
			name:			"map",
			description:	"		Displays the names of 20 location areas in the Pokemon world",
			callback:		commandMap,
		},
		"mapb": {
			name:			"mapb",
			description:	"		Displays the previous 20 location areas",
			callback: 		commandMapb,
		},
		"explore": {
			name:			"explore",
			description:	"	Virtually explores an area and finds Pokemon in that area",
			callback:		commandExplore,
		},
		"catch": {
			name:			"catch",
			description:	"		Attempt to catch a Pokemon!",
			callback:		commandCatch,
		},
	}
}


func main() {
	rand.Seed(time.Now().UnixNano())
	cfg := &Config{
		Cache:  pokecache.NewCache(5 * time.Second),
		Caught: make(map[string]pokeapi.Pokemon),
}
	startREPL(cfg)
}

func startREPL(cfg *Config) {
	commands = NewRegistry()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		fields := cleanInput(input)
		if len(fields) == 0 {
			continue
		}
		cliCommand, ok := commands[fields[0]]
		if ok {
			err := cliCommand.callback(cfg, fields[1:])
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command. Use 'help' for a list of commands.")
		}
	}
}

func cleanInput(text string) []string {
	loweredText := strings.ToLower(text)
	finalText := strings.Fields(loweredText)
	return finalText
}

func commandExit(cfg *Config, _[]string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config, _[]string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		cmd := commands[k]
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func commandMap(cfg *Config, _[]string) error {
	url := baseLocationAreaURL
	if cfg.Next != "" {
		url = cfg.Next
	}

	var b []byte
	if cached, ok := cfg.Cache.Get(url); ok {
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
		cfg.Cache.Add(url, body)
		b = body
	}
	
	var out locationAreaList
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	
	for _, r := range out.Results {
		fmt.Println(r.Name)
	}
	if out.Next != nil {
		cfg.Next = *out.Next
	} else {
		cfg.Next = ""
	}

	if out.Previous != nil {
		cfg.Previous = *out.Previous
	} else {
		cfg.Previous = ""
	}

	return nil
}

func commandMapb(cfg *Config, _[]string) error {
	url := baseLocationAreaURL
	if cfg.Previous != "" {
		url = cfg.Previous
	} else {
		fmt.Println("You are on the first page.")
		return nil
	}

	var b []byte 
	if cached, ok := cfg.Cache.Get(url); ok {
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
		cfg.Cache.Add(url, body)
		b = body
	}
	
	var out locationAreaList
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}

	for _, r := range out.Results {
		fmt.Println(r.Name)
	}
	if out.Next != nil {
		cfg.Next = *out.Next
	} else {
		cfg.Next = ""
	}

	if out.Previous != nil {
		cfg.Previous = *out.Previous
	} else {
		cfg.Previous = ""
	}

	return nil
}

func commandExplore(cfg *Config, args []string) error {
	if len(args) < 1 {
		fmt.Println("usage: explore <area_name>")
		return nil
	}

	areaName := args[0]
	fmt.Printf("Exploring %s...\n", areaName)

	if err := pokeapi.GetLocationArea(cfg.Cache, areaName); err != nil {
        return err
    }
    return nil
}

func commandCatch(cfg *Config, args []string) error {
	if len(args) < 1 {
	fmt.Println("usage: catch <pokemon>")
	return nil
	}
	name := args[0]
	var poke pokeapi.Pokemon
	poke, err := pokeapi.GetPokemonData(cfg.Cache, name)
	if err != nil {
		return err
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", name)

	max := 100
	threshold := 60 - poke.BaseExperience/5
	roll := rand.Intn(max)
	success := roll < threshold

	if threshold < 5 {
		threshold = 5
	}

	if threshold > 90 {
		threshold = 90
	}

	if success {
		fmt.Printf("%s was caught!\n", name)
		cfg.Caught[name] = poke
	} else {
		fmt.Printf("%s escaped!\n", name)
	}
	return nil
}