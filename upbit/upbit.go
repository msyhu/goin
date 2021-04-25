package upbit

import (
	"fmt"
)

func Upbit_ws_client() {
	uri := "wss://api.upbit.com/websocket/v1"

	// TODO : json-구조체로 만들기
	subscribe_data := `[
		{"ticket":"test"},
		{
			"type": "ticker",
			"codes":["KRW-BTC"],
			"isOnlyRealtime": True
		},
		{"format":"SIMPLE"}
	]`

	// TODO : websocket 연결해보기

	fmt.Println(subscribe_data)
	fmt.Println(uri)
}
