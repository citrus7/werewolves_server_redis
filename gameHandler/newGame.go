package gameHandler

import (
	"werewolves_server/contracts"
	"sync"
	"github.com/garyburd/redigo/redis"
	"fmt"
	"net/http"
	"werewolves_server/util"
	"encoding/json"
)

func NewGame(w http.ResponseWriter, r *http.Request, pool *redis.Pool) {
	fmt.Println("creating a new game")

	// Decode message
	decoder := json.NewDecoder(r.Body)
	var newGameData contracts.NewGameRequest
	err := decoder.Decode(&newGameData)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Open connection to Redis
	conn := pool.Get()
	if conn.Err() != nil {
		fmt.Println("Todo: connection error handling")
	}
	defer conn.Close()

	// Get new gameId
	exists := 0
	var gameId = ""

	// Make sure that game doesn't exist
	for exists >= 0 {
		gameId = util.GenerateId(5)
		fmt.Println("Attempting to create ", gameId)
		exists, _ = redis.Int(conn.Do("TTL", gameId))
		fmt.Println("exists: ", exists)
	}

	conn.Do("SET", gameId, 1)
	conn.Do("EXPIRE", gameId, 3600)
	fmt.Println("GameID is: ", gameId)

	gameLock := sync.RWMutex{}
	game:= contracts.Game{
		RedisServerChannel: gameId,
		NumPlayers: 0,
		MaxPlayers: 16,
		Players: make(map[string]contracts.Player),
		GameLock: &gameLock,
		GameId: gameId,
		Password: newGameData.Password,
	}

	// Todo: Redis server message handler
	// Should this handler or message handler enqueue messages onto player redis queue?
	done := sync.WaitGroup{}
	done.Add(1)
	go subscribeToGameRedisMessages(&game, pool, &done)
	done.Wait()

	// Main Game Loop
	// Relies on Redis Server RedisMessage Handler to update game object
	/*
	for game.GameState == 0 {

		//playersReady := 0

		for _, player := range (players) {
			if player != nil && player.PlayerId == game.OwnerId  && player.Ready{
				gameState = 1
			}
		}
		/*
		playersReady := 0
		for _, player := range (players) {
			if player != nil && player.Ready{
				playersReady++
			}
		}

		// When half of the players are ready game starts
		if playersReady == game.MaxPlayers {
			fmt.Println("All players ready")
			gameState = 1;
		}
		*/

		// End game if it has not started in an hour
		/*
		duration := time.Since(startTime)
		if duration.Minutes() > 60 {
			message := "Game " + game.GameId + " has expired"
			fmt.Println(message)
			gameState = 2
		}
	}
		*/

	response := contracts.NewGameResponse{
		GameId: gameId,
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}