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
	"time"
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

	// 왜있는거지?
	done := make(chan struct{})

	buyChannel := make(chan int)  // int형 채널 생성
	sellChannel := make(chan int) // int형 채널 생성
	go buyFunction(buyChannel)
	go sellFunction(sellChannel)

	// for 문 돌면서 계속 websocket listen
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			// TODO : 여기서 시세가 indicator 넘는지 끊임없이 확인해주기. 넘으면 대기하고 있는 buy/sell 고루틴으로 flag 전송
			var jsondata map[string]interface{}
			json.Unmarshal([]byte(message), &jsondata)

			tradePrice, err := strconv.Atoi(fmt.Sprint(jsondata["trade_price"]))
			if err != nil {
				fmt.Printf("%T, %v", tradePrice, tradePrice)
			}
			log.Println("now? ", tradePrice)

			if tradePrice >= 1320 {
				sellChannel <- tradePrice
			} else if tradePrice < 1320 {
				buyChannel <- tradePrice
			}

		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// 계속 반복문 돌면서
	for {
		select {
		// done 이라는 채널에서 값이 나오면 메시지 전송을 중지한다는 뜻?
		case <-done:
			return
		case t := <-ticker.C:
			// Ticker 구조체에 있는 tick 배달되는 채널. 1초에 한번씩 들어온다
			log.Println("시간? " + t.String())
			err := c.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("write:", t, err)
				return
			}
		case <-interrupt:
			// 종료 키보드 입력 들어올 시
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}

			// 이게 안정적으로 종료될 수 있도록 약간의 시간을 지연시키는 역할을 하는 듯?
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
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
