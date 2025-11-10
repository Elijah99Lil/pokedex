package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"pokedexcli/internal/pokecache"
	"sort"
	"strings"
	"time"
)
type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

type Config struct {
	Next		string
	Previous	string
	Cache 		*pokecache.Cache
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
			description:	"	Exit the Pokedex",
			callback:    	commandExit,
		},
		"help": {
			name:         	"help",
			description:  	"	Displays a help message",
			callback:     	commandHelp,    
		},
		"map": {
			name:			"map",
			description:	"	Displays the names of 20 location areas in the Pokemon world",
			callback:		commandMap,
		},
		"mapb": {
			name:			"mapb",
			description:	"	Displays the previous 20 location areas",
			callback: 		commandMapb,
		},
	}
}

func main() {
	cfg.Cache = pokecache.NewCache(5 * time.Second)
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
			err := cliCommand.callback(cfg)
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

func commandExit(cfg *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config) error {
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

func commandMap(cfg *Config) error {
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

func commandMapb(cfg *Config) error {
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