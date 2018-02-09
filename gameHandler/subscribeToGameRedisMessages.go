package gameHandler

import (
	"github.com/garyburd/redigo/redis"
	"werewolves_server/contracts"
	"fmt"
	"encoding/json"
	"sync"
	"werewolves_server/util"
)

// Read incoming player redis messages and update game accordingly
func subscribeToGameRedisMessages(game *contracts.Game, pool *redis.Pool, initialized *sync.WaitGroup) {

	// Open connection to Redis
	conn := pool.Get()
	PSconn := pool.Get()
	if conn.Err() != nil || PSconn.Err() != nil{
		fmt.Println("Todo: connection error handling")
	}
	defer conn.Close()
	defer PSconn.Close()


	psc := redis.PubSubConn{Conn: PSconn}
	err := psc.Subscribe(game.RedisServerChannel)
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		defer psc.Close()
	}
	fmt.Println("Subscribed to ", game.RedisServerChannel, " and reading messages")
	initialized.Done()

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:

			var redisMessage contracts.RedisMessage
			err = json.Unmarshal(v.Data, &redisMessage)
			if err != nil {
				fmt.Println("error decoding redis message: ", err)
				break
			}

			fmt.Println("server has received a redis message of type: ", redisMessage.Type)

			// Check to make sure that we aren't reading our own messages
			if redisMessage.ReceiverId == game.GameId {
				// Successfully joined game
				if redisMessage.Type == "contracts.PlayerRequestJoin" {
					handleRedisRequestJoin(redisMessage, game, conn)
				} else if redisMessage.Type == "contracts.IncomingWsMessage" {

					// Unmarshal Server RedisMessage
					var serverMessage contracts.IncomingWsMessage
					err := json.Unmarshal([]byte(redisMessage.Data), &serverMessage)
					if err != nil {
						fmt.Println("malformed server message: ", err)
						return
					}

					switch code := serverMessage.Code; code {
					case 104:
						var roleSettings contracts.RoleSettings
						err := json.Unmarshal(serverMessage.Data, &roleSettings)
						if err != nil {
							fmt.Println("malformed server message: ", err)
							return
						}
						fmt.Println("Updated game role data: ",roleSettings)
						game.GameLock.Lock()
						game.GameRoleSettings.NumBodyguards = roleSettings.NumBodyguards
						game.GameRoleSettings.NumSpellcasters = roleSettings.NumSpellcasters
						game.GameRoleSettings.NumHunters = roleSettings.NumHunters
						game.GameRoleSettings.NumWitches = roleSettings.NumWitches
						game.GameRoleSettings.NumSeers = roleSettings.NumSeers
						game.GameRoleSettings.NumVillagers = roleSettings.NumVillagers
						game.GameRoleSettings.NumWolves = roleSettings.NumWolves

						for _, player := range (game.Players) {
							if player.PlayerId != serverMessage.PlayerId {
								adjustmentMsg := contracts.OutgoingWsMessage{
									game.GameId,
									player.PlayerId,
									209,
									roleSettings,
								}
								redisMessage := util.ObjectToString(game.GameId, player.PlayerId, adjustmentMsg)
								_, err = conn.Do("PUBLISH", game.GameId, redisMessage)
								if err != nil {
									fmt.Println(err)
								}
							}
						}
						game.GameLock.Unlock()

					case 102:


					}

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