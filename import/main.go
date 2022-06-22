package main

import (
	"fmt"
	"github.com/aydinnyunus/blockchain"
	"math/rand"
	"strconv"
	"time"
)

var Neo4jUri = getEnv("NEO4J_URI", "neo4j://localhost:7687")
var Neo4jUser = getEnv("NEO4J_USER", "neo4j")
var Neo4jPass = getEnv("NEO4J_PASS", "letmein")
var exchange = []string{"uniswap"}
var SatoshiToBitcoin = float64(100000000)
var exitNodes []string
var ignoreAddress []string
var fromAddress = map[int]map[string]string{}
var toAddress = map[int]map[string]string{}
var hash = ""
var totalAmount = float64(0)
var timestamp = time.Now()
var totalUSD = float64(0)
var flowBTC = float64(0)
var flowUSD = float64(0)

func main() {
	//detectUNISWAP()
	//uni := readRedis(exchange[0])
	count := 0
	walletID := "37oTUqiViE3YySs8xxAtKgTzQgoVuSVbse"
	network := checkWalletNetwork(walletID)
	/* TODO: NETWORKE GÖRE AYRI İSTEKLER AT İF ELSE İLE AYIR BLOCKCHAIN KÜTÜPHANESİNİ DÜZENLE*/
	graph := New()

	/* TODO: MESSAGE BROKER NEODASH'TEN HANGI İSTEĞİN İSTENDİĞİNİ SÖYLİYCEK EĞER O İSTEK WEBSOCKETLE ALAKALIYSA O ÇALIŞACAK*/

	if network == BtcNetwork {

		c, e := blockchain.New()

		resp, e := c.GetAddress(walletID)
		if e != nil {
			time.Sleep(5 * time.Second)
		}
		count = 0
		node0 := graph.AddNode(resp.Address, resp.FinalBalance)

		for {
			//connectWebsocketAllTransactions()
			//connectWebsocketSpecificAddress(walletID)
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)
			for i := range resp.Txs {
				btcToUsd := getBitcoinPrice()
				hash = resp.Txs[i].Hash
				tm, err := strconv.ParseInt(strconv.Itoa(resp.Txs[i].Time), 10, 64)
				if err != nil {
					panic(err)
				}
				timestamp = time.Unix(tm, 0)

				//fmt.Println(resp.Txs[i].Result)

				for j := range resp.Txs[i].Inputs {
					if len(resp.Txs[i].Inputs[j].PrevOut.Addr) == 0 {
						continue
					}
					ignoreAddress = append(ignoreAddress, resp.Txs[i].Inputs[j].PrevOut.Addr)
					fromAddress = map[int]map[string]string{
						j + count + r1.Intn(100): {"address": resp.Txs[i].Inputs[j].PrevOut.Addr, "value": strconv.FormatFloat(float64(resp.Txs[i].Inputs[j].PrevOut.Value)/SatoshiToBitcoin, 'E', -1, 64), "value_usd": strconv.FormatFloat(float64(resp.Txs[i].Inputs[j].PrevOut.Value)/SatoshiToBitcoin*btcToUsd, 'E', -1, 64)},
					}
					//fmt.Println(resp.Txs[i].Inputs[j].PrevOut.Addr)
					node1 := graph.AddNode(resp.Txs[i].Inputs[j].PrevOut.Addr, resp.Txs[i].Inputs[j].PrevOut.Value)
					graph.AddEdge(node0, node1, 1)

					//fmt.Println(resp.Txs[i].Inputs[j].PrevOut.Value)
				}

				for k := range resp.Txs[i].Out {
					//fmt.Println(resp.Txs[i].Out[k].Addr)
					//fmt.Println(resp.Txs[i].Out[k].Value)
					totalAmount += float64(resp.Txs[i].Out[k].Value) / SatoshiToBitcoin
					totalUSD = totalAmount * btcToUsd
					toAddress = map[int]map[string]string{
						k + count + r1.Intn(100): {"address": resp.Txs[i].Out[k].Addr, "value": strconv.FormatFloat(float64(resp.Txs[i].Out[k].Value)/SatoshiToBitcoin, 'E', -1, 64), "value_usd": strconv.FormatFloat(float64(resp.Txs[i].Out[k].Value)/SatoshiToBitcoin*btcToUsd, 'E', -1, 64)},
					}

					if !stringInSlice(resp.Txs[i].Out[k].Addr, ignoreAddress) {
						flowBTC += float64(resp.Txs[i].Out[k].Value) / SatoshiToBitcoin
					}

					flowUSD = flowBTC * btcToUsd
					_, err := neo4jDatabase(hash, timestamp.Format("2006-01-02"), strconv.FormatFloat(totalUSD, 'E', -1, 64), strconv.FormatFloat(totalAmount, 'E', -1, 64), strconv.FormatFloat(flowBTC, 'E', -1, 64), strconv.FormatFloat(flowUSD, 'E', -1, 64), fromAddress, toAddress)
					if err != nil {
						return
					}
					node1 := graph.AddNode(resp.Txs[i].Out[k].Addr, resp.Txs[i].Out[k].Value)
					graph.AddEdge(node0, node1, 1)
				}
			}
			count += 1

			resp, e := c.GetAddress(graph.nodes[count].walletId)
			if e != nil {
				fmt.Print(e)
				time.Sleep(1 * time.Minute)
			}
			node0 = graph.AddNode(resp.Address, resp.FinalBalance)

			if len(graph.nodes[len(graph.nodes)-1].edges) == 0 {
				exitNodes = append(exitNodes, graph.nodes[len(graph.nodes)-1].walletId)
				break
			}

		}
	} else if network == EthNetwork {
		c, e := blockchain.New()

		resp2, e := c.GetETHAddressSummary(walletID, true)
		if e != nil {
			time.Sleep(5 * time.Second)
		}
		count = 0
		balance, _ := strconv.Atoi(resp2.Balance)
		node0 := graph.AddNode(resp2.Hash, balance)

		resp, e := c.GetETHAddress(walletID)
		if e != nil {
			time.Sleep(5 * time.Second)
		}

		for {
			//connectWebsocketAllTransactions()
			//connectWebsocketSpecificAddress(walletID)
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)
			for i := range resp.Transactions {
				btcToUsd := getBitcoinPrice()
				hash = resp.Transactions[i].Hash
				tm, err := strconv.ParseInt(resp.Transactions[i].Timestamp, 10, 64)
				if err != nil {
					panic(err)
				}
				timestamp = time.Unix(tm, 0)

				//fmt.Println(resp.Txs[i].Result)

				if resp.Transactions[i].From == resp.Transactions[i].To {
					ignoreAddress = append(ignoreAddress, resp.Transactions[i].From)
					continue
				}
				value, _ := strconv.ParseFloat(resp.Transactions[i].Value,64)
				fromAddress = map[int]map[string]string{
					i + count + r1.Intn(100): {"address": resp.Transactions[i].From, "value": strconv.FormatFloat(value/SatoshiToBitcoin, 'E', -1, 64), "value_usd": strconv.FormatFloat(value/SatoshiToBitcoin*btcToUsd, 'E', -1, 64)},
				}
				//fmt.Println(resp.Txs[i].Inputs[j].PrevOut.Addr)
				node1 := graph.AddNode(resp.Transactions[i].From, int(value))
				graph.AddEdge(node0, node1, 1)

				//fmt.Println(resp.Txs[i].Inputs[j].PrevOut.Value)


				//fmt.Println(resp.Txs[i].Out[k].Addr)
				//fmt.Println(resp.Txs[i].Out[k].Value)
				totalAmount += value / SatoshiToBitcoin
				totalUSD = totalAmount * btcToUsd
				value,_ = strconv.ParseFloat(resp.Transactions[i].Value,64)
				toAddress = map[int]map[string]string{
					i + count + r1.Intn(100): {"address": resp.Transactions[i].To, "value": strconv.FormatFloat(value/SatoshiToBitcoin, 'E', -1, 64), "value_usd": strconv.FormatFloat(value/SatoshiToBitcoin*btcToUsd, 'E', -1, 64)},
				}

				if !stringInSlice(resp.Transactions[i].From, ignoreAddress) {
					flowBTC += value / SatoshiToBitcoin
				}

				flowUSD = flowBTC * btcToUsd
				_, err = neo4jDatabase(hash, timestamp.Format("2006-01-02"), strconv.FormatFloat(totalUSD, 'E', -1, 64), strconv.FormatFloat(totalAmount, 'E', -1, 64), strconv.FormatFloat(flowBTC, 'E', -1, 64), strconv.FormatFloat(flowUSD, 'E', -1, 64), fromAddress, toAddress)
				if err != nil {
					return
				}
				node1 = graph.AddNode(resp.Transactions[i].From, int(value))
				graph.AddEdge(node0, node1, 1)

			}
			count += 1

			resp, e := c.GetETHAddressSummary(graph.nodes[count].walletId,true)
			if e != nil {
				fmt.Print(e)
				time.Sleep(1 * time.Minute)
			}
			balance, _ = strconv.Atoi(resp.Balance)
			node0 = graph.AddNode(resp.Hash, balance)

			if len(graph.nodes[len(graph.nodes)-1].edges) == 0 {
				exitNodes = append(exitNodes, graph.nodes[len(graph.nodes)-1].walletId)
				break
			}

		}
	}
	/*
		for i, _ := range exitNodes {
			if stringInSlice(exitNodes[i], uni) {
				fmt.Printf("Bu adres UNISWAPE ÇIKIYOR %s\n", exitNodes[i])
			}
		}
	*/
	}



