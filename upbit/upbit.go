package upbit

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
)

var addr = flag.String("addr", "api.upbit.com", "upbit websocket")
var msg = "[{\"ticket\":\"test\"},{\"type\":\"ticker\",\"codes\":[\"KRW-XRP\"],\"isOnlyRealtime\":\"true\"}]"

func UpbitWsClient() {

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/websocket/v1"}
	log.Printf("connecting to %s", u.String())

	// websocket 연결 요청
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	} else {
		log.Println(u.String() + " 접속 성공!")
	}

	defer c.Close()

	buyChannel := make(chan int)  // int형 채널 생성
	sellChannel := make(chan int) // int형 채널 생성
	go buyFunction(buyChannel)
	go sellFunction(sellChannel)

	// for 문 돌면서 계속 websocket listen
	go func() {

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var jsondata map[string]interface{}
			json.Unmarshal([]byte(message), &jsondata)

			tradePrice, err := strconv.Atoi(fmt.Sprint(jsondata["trade_price"]))
			if err != nil {
				fmt.Printf("%T, %v", tradePrice, tradePrice)
			}

			// TODO: 실제 매수 매매 전략 넣어보기
			if tradePrice >= 1320 {
				sellChannel <- tradePrice
			} else if tradePrice < 1320 {
				buyChannel <- tradePrice
			}

		}
	}()

	for {
		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("write:", err)
			return
		}
		select {
		case <-interrupt:
			// 종료 키보드 입력 들어올 시
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			closeErr := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if closeErr != nil {
				log.Println("write close:", closeErr)
				return
			}
		}
	}

}

func buyFunction(buyChannel chan int) {
	for {
		nowPrice := <-buyChannel
		log.Println("buy! ", nowPrice)
	}
}

func sellFunction(sellChannel chan int) {
	for {
		nowPrice := <-sellChannel
		log.Println("sell! ", nowPrice)
	}
}
