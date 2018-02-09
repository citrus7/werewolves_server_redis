package playerHandler

import (
	"fmt"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
	"log"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"werewolves_server/contracts"
	"werewolves_server/util"
)

// Create a new player and assign to correct game
// Go routine lives until end of game
// Listens to Redis channel and forwards relevent messages to player WS
func NewPlayer(w http.ResponseWriter, r *http.Request, pool *redis.Pool) {
	player := Player{}
	player.PlayerId = util.GenerateId(16)
	fmt.Println("PlayerId: ", player.PlayerId)

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// Create Player's Websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	player.Ws = ws
	wsLock := sync.RWMutex{}
	player.WsLock = &wsLock

	// Open connection to Redis
	PSconn := pool.Get()
	conn := pool.Get()
	defer PSconn.Close()
	defer conn.Close()



	// Block until receive player info
	// Todo: timeout?
	var newPlayerMsg contracts.NewPlayerRequestMessage
	err = ws.ReadJSON(&newPlayerMsg)
	if err != nil {
		fmt.Println("error decoding newPlayerMessage")
		return
	}

	player.Name = newPlayerMsg.Name
	player.GameId = newPlayerMsg.GameId

	//Check that room exists
	exists, err := redis.Bool(conn.Do("EXISTS", newPlayerMsg.GameId))
	if err != nil {
		fmt.Println(err)
	}
	if exists {
		fmt.Println("game found")
		// Request to join game
		joinGameRequest := contracts.PlayerRequestJoin{
			PlayerId: player.PlayerId,
			Password: newPlayerMsg.Password,
			Name:     newPlayerMsg.Name,
		}
		req := util.ObjectToString(player.PlayerId, player.GameId, joinGameRequest)

		// Subscribe to redis channel
		psc := redis.PubSubConn{Conn: PSconn}
		err = psc.Subscribe(newPlayerMsg.GameId)
		if err != nil {
			fmt.Println(err)
		} else {
			defer psc.Close()
		}

		fmt.Println("publishing join game request")
		_, err = conn.Do("PUBLISH", newPlayerMsg.GameId, req)
		if err != nil {
			fmt.Println(err)
		}

		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				fmt.Println("Player has received a redis message: ", v.Channel, v.Data)

				var redisMessage contracts.RedisMessage
				err = json.Unmarshal(v.Data, &redisMessage)
				if err != nil {
					fmt.Println("error decoding redis message: ", err)
					break
				}

				// Only forward messages meant for player or all
				if redisMessage.ReceiverId == player.PlayerId || (redisMessage.ReceiverId == "") {
					// Successfully joined game
					fmt.Println(redisMessage.Type)
					if redisMessage.Type == "contracts.GameInfo" {
						// Join game request accepted
						var gameInfo contracts.GameInfo
						err = json.Unmarshal([]byte(redisMessage.Data), &gameInfo)
						if err != nil {
							fmt.Println("error decoding GameInfo: ", err)
							break
						}

						if gameInfo.Error == "" {
							go handleIncomingMessages(pool, ws, player)
							responseMessage := contracts.OutgoingWsMessage{}
							responseMessage.PlayerId = player.PlayerId
							responseMessage.GameId = player.GameId
							responseMessage.Code = 201
							responseMessage.Data = gameInfo
							player.WsLock.Lock()
							ws.WriteJSON(responseMessage)
							player.WsLock.Unlock()
						} else {
							// Todo: proper exit and cleanup
							player.WsLock.Lock()
							// Write error?
							ws.WriteJSON(redisMessage.Data)
							player.WsLock.Unlock()
							break
						}

					} else if redisMessage.Type == "contracts.UpdatePlayerMessage" {
						handlePlayerUpdate(redisMessage, &player)
					} else if redisMessage.Type == "contracts.OutgoingWsMessage" {
						player.WsLock.Lock()
						// Write error?
						var message contracts.OutgoingWsMessage
						err := json.Unmarshal([]byte(redisMessage.Data), &message)
						if err != nil {
							fmt.Println("error decoding OutgoingWsMessage: ", err)
							break
						}
						fmt.Println("writing out message to player")
						ws.WriteJSON(message)
						player.WsLock.Unlock()
					}
				}

			case redis.Subscription:
				fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
			case error:
				fmt.Println(v)
				return
			}
		}
	}

}