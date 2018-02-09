package playerHandler

import (
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"fmt"
	"werewolves_server/contracts"
	"werewolves_server/util"
)

// Forwards incoming player messages to redis channel

func handleIncomingMessages(pool *redis.Pool, ws *websocket.Conn, player Player) {
	// Open connection to Redis
	conn := pool.Get()
	defer conn.Close()

	for {
		// Read in a new message as JSON and map it to a OutgoingWsMessage object
		// Todo: more efficient way of reading messages
		var wsMessage contracts.IncomingWsMessage
		err := ws.ReadJSON(&wsMessage)
		if err != nil {
			fmt.Println("error decoding incoming WS message: ", err)
			break
		}

		playerMessage := util.ObjectToString(player.PlayerId, player.GameId, wsMessage)

		_, err = conn.Do("PUBLISH", player.GameId, playerMessage)
		if err != nil {
			fmt.Println(err)
			break
		}


	}
}