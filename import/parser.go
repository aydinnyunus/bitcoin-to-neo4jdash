package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type exitResponse struct {
	D struct {
		TxTotal         string      `json:"txTotal"`
		BalanceTotal    string      `json:"balanceTotal"`
		ColumnTotal     string      `json:"columnTotal"`
		CurrentPage     string      `json:"currentPage"`
		TotalPage       string      `json:"totalPage"`
		ShowPopUp       bool        `json:"showPopUp"`
		Draw            int         `json:"draw"`
		RecordsTotal    int         `json:"recordsTotal"`
		RecordsFiltered int         `json:"recordsFiltered"`
		CustomRep       interface{} `json:"customRep"`
		RawExportLink   interface{} `json:"rawExportLink"`
		Data            []struct {
			Type     string `json:"__type"`
			Address  string `json:"address"`
			NameTag  string `json:"nameTag"`
			Balance  string `json:"balance"`
			TxnCount string `json:"txnCount"`
		} `json:"data"`
		Error interface{} `json:"error"`
	} `json:"d"`
}

func checkControlled(index int) bool {
	exc := readRedis(exchange[index])
	if len(exc) > 1 {
		return true
	}
	return false
}

func detectUNISWAP() bool {
	var response exitResponse
	client := &http.Client{}
	if !checkControlled(0) {
		n := 25
		c := 3
		for {
			var data = strings.NewReader(`{"dataTableModel":{"draw":` + strconv.Itoa(c) + `,"columns":[{"data":"address","name":"","searchable":true,"orderable":false,"search":{"value":"","regex":false}},{"data":"nameTag","name":"","searchable":true,"orderable":false,"search":{"value":"","regex":false}},{"data":"balance","name":"","searchable":true,"orderable":true,"search":{"value":"","regex":false}},{"data":"txnCount","name":"","searchable":true,"orderable":true,"search":{"value":"","regex":false}}],"order":[{"column":1,"dir":"asc"}],"start":` + strconv.Itoa(n) + `,"length":25,"search":{"value":"","regex":false}},"labelModel":{"label":"` + exchange[0] + `","subCategoryId":"0"}}`)
			req, err := http.NewRequest("POST", "https://etherscan.io/accounts.aspx/GetTableEntriesBySubLabel", data)
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Set("authority", "etherscan.io")
			req.Header.Set("accept", "application/json, text/javascript, */*; q=0.01")
			req.Header.Set("accept-language", "tr-TR,tr;q=0.9,en-US;q=0.8,en;q=0.7")
			req.Header.Set("cache-control", "no-cache")
			req.Header.Set("content-type", "application/json")
			req.Header.Set("cookie", "__stripe_mid=c21ec6ed-510f-41fc-a3e7-c1e14445fc21a0b27d; __cuid=d6a5ed9d0fdd48e198deac5ec7c391e9; amp_fef1e8=319cd799-59f9-4611-ac87-9c0cf51119f8R...1g59knd71.1g59lgnpe.b.2.d; __cflb=02DiuFnsSsHWYH8WqVXbZzkeTrZ6gtmGUbTZqBibRrBXi; ASP.NET_SessionId=a3k1lyzw2kpwvdlmxenxjngh; __cf_bm=AU4BFDaEtGIsguspKqv7nNhpvkKyzQUTHS6NuUvXcok-1655314398-0-AaMI1x8Lt93KpRd0OiRKVhSj56Gz9I6rD1xC0L277cCY5biapLrswpmXKXTcOBtze1rqeeIJdBvwOyIKbiRlrNhhgg353mL90U2TfLaeA77Yw0GwAa3X6NagH2MTrW3Z9w==; __stripe_sid=d97c96ab-64fd-47e0-a8d7-32a6f27bb7da3d4fca")
			req.Header.Set("origin", "https://etherscan.io")
			req.Header.Set("pragma", "no-cache")
			req.Header.Set("referer", "https://etherscan.io/accounts/label/"+exchange[0]+"?subcatid=0&size=25&start="+strconv.Itoa(n-25)+"&col=1&order=asc")
			req.Header.Set("sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="102", "Google Chrome";v="102"`)
			req.Header.Set("sec-ch-ua-mobile", "?0")
			req.Header.Set("sec-ch-ua-platform", `"Linux"`)
			req.Header.Set("sec-fetch-dest", "empty")
			req.Header.Set("sec-fetch-mode", "cors")
			req.Header.Set("sec-fetch-site", "same-origin")
			req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36")
			req.Header.Set("x-requested-with", "XMLHttpRequest")
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					log.Fatal(err)
				}
			}(resp.Body)
			bodyText, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			//fmt.Printf("%s\n", bodyText)
			err = json.Unmarshal(bodyText, &response)
			if err != nil {
				return false
			}
			if len(response.D.Data) == 0 {
				return true
			}
			for i := range response.D.Data {
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(response.D.Data[i].Address))

				if err != nil {
					log.Fatal(err)
				}
				address := doc.Find("a").Text()
				err = addRedis(exchange[0], address)
				if err != nil {
					return false
				}
			}
			c += 1
			n += 25
		}

	}
	return true
}
