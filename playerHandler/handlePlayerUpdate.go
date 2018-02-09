package playerHandler

import (
	"werewolves_server/contracts"
	"encoding/json"
	"fmt"
)

func handlePlayerUpdate(redisMessage contracts.RedisMessage, player *Player) {

	var gameInfo contracts.UpdatePlayerMessage
	err := json.Unmarshal([]byte(redisMessage.Data), &gameInfo)
	if err != nil {
		fmt.Println("error decoding GameInfo: ", err)
		return
	}

	responseMessage := contracts.OutgoingWsMessage{}
	responseMessage.PlayerId = player.PlayerId
	responseMessage.GameId = player.GameId
	responseMessage.Code = 205
	responseMessage.Data = gameInfo

	fmt.Println("sending player info out")
	player.WsLock.Lock()
	player.Ws.WriteJSON(responseMessage)
	player.WsLock.Unlock()
}

