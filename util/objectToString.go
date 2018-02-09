package util

import (
	"werewolves_server/contracts"
	"reflect"
	"encoding/json"
)

func ObjectToString(senderId string, receiverId string, obj interface{}) string {

	objStr, _ := json.Marshal(obj)
	msg, _ := json.Marshal(contracts.RedisMessage{
		SenderId: senderId,
		ReceiverId: receiverId,
		Type: reflect.TypeOf(obj).String(),
		Data: string(objStr),
	})

	return string(msg)
}