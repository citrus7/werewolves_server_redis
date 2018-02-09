package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"fmt"
	"runtime"
	"github.com/garyburd/redigo/redis"
	"werewolves_server/gameHandler"
	"werewolves_server/playerHandler"
)

func MakeHttpHandler(redisPool *redis.Pool) http.Handler {
	r := mux.NewRouter()

	// Create a new game
	r.HandleFunc("/NewGame", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			gameHandler.NewGame(w, r, redisPool)
		}
	})
	r.HandleFunc("/TestNewGame", func(w http.ResponseWriter, r *http.Request) {

		gameHandler.NewGame(w, r, redisPool)

	})

	// View all available games
	// For debugging purposes only
	// Deprecated
	/*
	r.HandleFunc("/ViewGames", func(w http.ResponseWriter, r *http.Request) {
		gamesLock.RLock()
		for _, game := range *games {
			fmt.Fprintln(w, game.GameId)
		}
		gamesLock.RUnlock()
	})
	*/

	// View number of goroutines
	// For debugging only
	r.HandleFunc("/Performance", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, runtime.NumGoroutine())
	})

	// New Player Handler
	r.HandleFunc("/TestPing", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("sending a ping")
		c := redisPool.Get()
		_, err := c.Do("PUBLISH", "testGameChannel", "test message")
		if err != nil {
			fmt.Println(err)
		}
		c.Close()
		/*
		psc := redis.PubSubConn{Conn: redisPool.Get()}
		psc.Subscribe("testGameChannel")
		psc.Ping("testing123")
		*/

	})


	// New Player Handler
	r.HandleFunc("/JoinGame", func(w http.ResponseWriter, r *http.Request) {
		playerHandler.NewPlayer(w, r, redisPool)
	})

	// Livecheck
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "alive")
	})

	return r
}