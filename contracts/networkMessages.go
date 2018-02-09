package contracts

import "encoding/json"

// Request to create a new game
type NewGameRequest struct {
	Password string `json:"password"`
}

type NewGameResponse struct {
	GameId string `json:"gameId"`
	Error string `json:"error"`
}

// New player requesting to join game
type NewPlayerRequestMessage struct {
	PlayerId string `json:"playerId"`
	GameId   string `json:"gameId"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Notify client of updated player info
type UpdatePlayerMessage struct {
	PlayerId string `json:"playerId"`
	Name     string `json:"name"`
	Seat	int	`json:"seat"`
	Ready 	bool `json:"ready"`
	Alive 	bool `json:"alive"`
}


// Regular message received from Client
type IncomingWsMessage struct {
	PlayerId string `json:"playerId"`
	GameId   string `json:"gameId"`
	Code     int `json:"code"`
	Data     json.RawMessage `json:"data"`
}


// Regular message sent to client
type OutgoingWsMessage struct {
	PlayerId string `json:"playerId"`
	GameId   string `json:"gameId"`
	Code     int `json:"code"`
	Data     interface{} `json:"data"`
}

type PlayerRequestJoin struct {
	PlayerId string `json:"playerId"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type RedisMessage struct {
	SenderId string `json:"senderId"`
	ReceiverId string `json:"receiverId"`
	Type string `json:"type"`
	Data string `json:"data"`
}

// Game info sent to new players upon successful join
type GameInfo struct {
	OwnerName string `json:"ownerName"`
	OwnerId string `json:"ownerId"`
	NumVillagers int `json:"numVillagers"`
	NumWolves int `json:"numWolves"`
	NumSeers int `json:"numSeers"`
	NumWitches int `json:"numWitches"`
	NumHunters int `json:"numHunters"`
	NumSpellcasters int `json:"numSpellcasters"`
	NumBodyguards int `json:"numBodyguards"`
	NumNarrators int `json:"numNarrators"`
	Error string `json:"error"`
}