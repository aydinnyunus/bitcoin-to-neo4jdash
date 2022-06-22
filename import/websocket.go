package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type responseJSON struct {
	Op string `json:"op"`
	X  struct {
		LockTime int `json:"lock_time"`
		Ver      int `json:"ver"`
		Size     int `json:"size"`
		Inputs   []struct {
			Sequence int64 `json:"sequence"`
			PrevOut  struct {
				Spent   bool   `json:"spent"`
				TxIndex int    `json:"tx_index"`
				Type    int    `json:"type"`
				Addr    string `json:"addr"`
				Value   int    `json:"value"`
				N       int    `json:"n"`
				Script  string `json:"script"`
			} `json:"prev_out"`
			Script string `json:"script"`
		} `json:"inputs"`
		Time      int    `json:"time"`
		TxIndex   int    `json:"tx_index"`
		VinSz     int    `json:"vin_sz"`
		Hash      string `json:"hash"`
		VoutSz    int    `json:"vout_sz"`
		RelayedBy string `json:"relayed_by"`
		Out       []struct {
			Spent   bool   `json:"spent"`
			TxIndex int    `json:"tx_index"`
			Type    int    `json:"type"`
			Addr    string `json:"addr"`
			Value   int    `json:"value"`
			N       int    `json:"n"`
			Script  string `json:"script"`
		} `json:"out"`
	} `json:"x"`
}

var addr = flag.String("addr", "ws.blockchain.info", "http service address")

func connectWebsocketAllTransactions() {
	flag.Parse()
	count := 0
	x := 0
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/inv"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			var r responseJSON
			err = json.Unmarshal(message, &r)

			if err != nil {

				// if error is not nil
				// print error
				fmt.Println(err)
			}
			log.Printf("recv: %s", r.X.Hash)
			if x%1000 == 0 {
				fmt.Println("Imported " + strconv.Itoa(x)+" transactions")
			}
			x += 1

			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)
			btcToUsd := getBitcoinPrice()
			hash = r.X.Hash
			tm, err := strconv.ParseInt(strconv.Itoa(r.X.Time), 10, 64)
			if err != nil {
				panic(err)
			}
			timestamp = time.Unix(tm, 0)

			for j := range r.X.Inputs {
				if len(r.X.Inputs[j].PrevOut.Addr) == 0 {
					continue
				}
				ignoreAddress = append(ignoreAddress, r.X.Inputs[j].PrevOut.Addr)
				fromAddress = map[int]map[string]string{
					j + count + r1.Intn(100): {"address": r.X.Inputs[j].PrevOut.Addr, "value": strconv.FormatFloat(float64(r.X.Inputs[j].PrevOut.Value)/SatoshiToBitcoin, 'E', -1, 64), "value_usd": strconv.FormatFloat(float64(r.X.Inputs[j].PrevOut.Value)/SatoshiToBitcoin*btcToUsd, 'E', -1, 64)},
				}
			}

			for k := range r.X.Out {
				totalAmount += float64(r.X.Out[k].Value) / SatoshiToBitcoin
				totalUSD = totalAmount * btcToUsd
				toAddress = map[int]map[string]string{
					k + count + r1.Intn(100): {"address": r.X.Out[k].Addr, "value": strconv.FormatFloat(float64(r.X.Out[k].Value)/SatoshiToBitcoin, 'E', -1, 64), "value_usd": strconv.FormatFloat(float64(r.X.Out[k].Value)/SatoshiToBitcoin*btcToUsd, 'E', -1, 64)},
				}

				if !stringInSlice(r.X.Out[k].Addr, ignoreAddress) {
					flowBTC += float64(r.X.Out[k].Value) / SatoshiToBitcoin
				}

				flowUSD = flowBTC * btcToUsd
			}
			_, err = neo4jDatabase(hash, timestamp.Format("2006-01-02"), strconv.FormatFloat(totalUSD, 'E', -1, 64), strconv.FormatFloat(totalAmount, 'E', -1, 64), strconv.FormatFloat(flowBTC, 'E', -1, 64), strconv.FormatFloat(flowUSD, 'E', -1, 64), fromAddress, toAddress)
				if err != nil {
					fmt.Println(err)
					return
				}
			count += 1

		}

	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("{\"op\": \"unconfirmed_sub\"}"))
			if err != nil {
				fmt.Println(t)
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func connectWebsocketSpecificAddress(btcAddress string) {
	flag.Parse()
	count,x := 0,0
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/inv"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			var r responseJSON
			err = json.Unmarshal(message, &r)

			if err != nil {

				// if error is not nil
				// print error
				fmt.Println(err)
			}
			log.Printf("recv: %s", r.X.Hash)
			if x%1000 == 0 {
				fmt.Println("Imported " + strconv.Itoa(x)+" transactions")
			}
			x += 1

			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)
			btcToUsd := getBitcoinPrice()
			hash = r.X.Hash
			tm, err := strconv.ParseInt(strconv.Itoa(r.X.Time), 10, 64)
			if err != nil {
				panic(err)
			}
			timestamp = time.Unix(tm, 0)

			for j := range r.X.Inputs {
				if len(r.X.Inputs[j].PrevOut.Addr) == 0 {
					continue
				}
				ignoreAddress = append(ignoreAddress, r.X.Inputs[j].PrevOut.Addr)
				fromAddress = map[int]map[string]string{
					j + count + r1.Intn(100): {"address": r.X.Inputs[j].PrevOut.Addr, "value": strconv.FormatFloat(float64(r.X.Inputs[j].PrevOut.Value)/SatoshiToBitcoin, 'E', -1, 64), "value_usd": strconv.FormatFloat(float64(r.X.Inputs[j].PrevOut.Value)/SatoshiToBitcoin*btcToUsd, 'E', -1, 64)},
				}
			}

			for k := range r.X.Out {
				totalAmount += float64(r.X.Out[k].Value) / SatoshiToBitcoin
				totalUSD = totalAmount * btcToUsd
				toAddress = map[int]map[string]string{
					k + count + r1.Intn(100): {"address": r.X.Out[k].Addr, "value": strconv.FormatFloat(float64(r.X.Out[k].Value)/SatoshiToBitcoin, 'E', -1, 64), "value_usd": strconv.FormatFloat(float64(r.X.Out[k].Value)/SatoshiToBitcoin*btcToUsd, 'E', -1, 64)},
				}

				if !stringInSlice(r.X.Out[k].Addr, ignoreAddress) {
					flowBTC += float64(r.X.Out[k].Value) / SatoshiToBitcoin
				}

				flowUSD = flowBTC * btcToUsd
			}
			_, err = neo4jDatabase(hash, timestamp.Format("2006-01-02"), strconv.FormatFloat(totalUSD, 'E', -1, 64), strconv.FormatFloat(totalAmount, 'E', -1, 64), strconv.FormatFloat(flowBTC, 'E', -1, 64), strconv.FormatFloat(flowUSD, 'E', -1, 64), fromAddress, toAddress)
				if err != nil {
					fmt.Println(err)
					return
				}
			count += 1

		}

	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			address := "{\"op\": \"addr_sub\",\"addr\": " + btcAddress + "}"
			err := c.WriteMessage(websocket.TextMessage, []byte(address))
			if err != nil {
				fmt.Println(t)
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
