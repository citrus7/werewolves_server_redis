package gameHandler

import (
	"encoding/json"
	"fmt"
	"werewolves_server/util"
	"werewolves_server/contracts"
	"github.com/garyburd/redigo/redis"
)

func handleRedisRequestJoin(redisMessage contracts.RedisMessage, game *contracts.Game, conn redis.Conn) {
	// Check if player join allowable
	var playerRequestJoin contracts.PlayerRequestJoin
	err := json.Unmarshal([]byte(redisMessage.Data), &playerRequestJoin)
	if err != nil {
		fmt.Println("malformed playerRequestJoin message: ", err)
		return
	}
	joinGameResponse := contracts.GameInfo{}
	fmt.Println("playerRequestJoin: ", playerRequestJoin)
	if !(game.Password == "" || playerRequestJoin.Password == game.Password) {
		// Password is wrong
		joinGameResponse.Error = "Incorrect password"
		fmt.Println("Incorrect password")
	} else if game.NumPlayers >= game.MaxPlayers {
		// Game is full
		joinGameResponse.Error = "Game is full"
		fmt.Println("Game is full")
	} else {
		fmt.Println("joined successfully")
		// Successfully joined game
		newPlayer := contracts.Player{
			PlayerId: playerRequestJoin.PlayerId,
			Name:     playerRequestJoin.Name,
			Role:     "",
			Alive:    true,
		}
		game.GameLock.Lock()
		game.NumPlayers++
		// First player in the room:
		if (game.NumPlayers == 1) {
			game.OwnerId = playerRequestJoin.PlayerId
			game.OwnerName = playerRequestJoin.Name
		}
		game.Players[newPlayer.PlayerId] = newPlayer
		game.GameLock.Unlock()

		// Create newGame object for response message
		game.GameLock.Lock()
		joinGameResponse = contracts.GameInfo{
			OwnerName:       game.OwnerName,
			OwnerId:         game.OwnerId,
			NumWolves:       game.GameRoleSettings.NumWolves,
			NumVillagers:    game.GameRoleSettings.NumVillagers,
			NumSeers:        game.GameRoleSettings.NumSeers,
			NumWitches:      game.GameRoleSettings.NumWitches,
			NumHunters:      game.GameRoleSettings.NumHunters,
			NumSpellcasters: game.GameRoleSettings.NumSpellcasters,
			NumBodyguards:   game.GameRoleSettings.NumBodyguards,
			NumNarrators:    game.GameRoleSettings.NumNarrators,
		}
		game.GameLock.Unlock()

	}
	fmt.Println("joinGameResponse: ", joinGameResponse)
	// Send response to redis
	req := util.ObjectToString(game.GameId, playerRequestJoin.PlayerId, joinGameResponse)
	_, err = conn.Do("PUBLISH", game.GameId, req)
	if err != nil {
		fmt.Println(err)
	}

	updateNewPlayerMessage := contracts.UpdatePlayerMessage{
		PlayerId: playerRequestJoin.PlayerId,
		Name:     playerRequestJoin.Name,
		Seat:     -1,
		Ready:    false,
		Alive:    true,
	}

	// Resend out all player info
	fmt.Println("Sending out player info")
	for _, p := range game.Players {
		updateMessage := contracts.UpdatePlayerMessage{
			PlayerId: p.PlayerId,
			Name:     p.Name,
			Seat:     p.Seat,
			Ready:    p.Ready,
			Alive:    p.Alive,
		}
		// Notify new player of exisitng users
		if p.PlayerId != playerRequestJoin.PlayerId {
			req := util.ObjectToString(game.GameId, playerRequestJoin.PlayerId, updateMessage)
			_, err = conn.Do("PUBLISH", game.GameId, req)
			// Notify other players of new player
			req = util.ObjectToString(game.GameId, p.PlayerId, updateNewPlayerMessage)
			_, err = conn.Do("PUBLISH", game.GameId, req)
		}

		if err != nil {
			fmt.Println(err)
		}
	}

}
