package contracts

import (
	"sync"
)

// Represents a game
type Game struct {
	RedisServerChannel	   string
	RedisClientChannel		string
	GameRoleSettings		RoleSettings

	GameId     string
	Password   string
	NumPlayers int
	MaxPlayers int
	OwnerId string
	OwnerName string


	Players			map[string]Player

	GameState		int


	GameLock   *sync.RWMutex
}

type Player struct {
	PlayerId	string
	Name		string
	Seat		int
	Role		string
	Alive		bool
	Ready		bool
}

type RoleSettings struct {
	NumVillagers int `json:"numVillagers"`
	NumWolves int `json:"numWolves"`
	NumSeers int `json:"numSeers"`
	NumWitches int `json:"numWitches"`
	NumHunters int `json:"numHunters"`
	NumSpellcasters int `json:"numSpellcasters"`
	NumNarrators int `json:"numNarrators"`
	NumBodyguards int `json:"numBodyguards"`
}

type GameState struct {
	NumWolves       int
	NumSeers        int
	NumWitches      int
	NumHunters      int
	NumSpellcasters int
	NumBodyguards   int
	NumNarrators    int
	NumVillagers    int
}