package main

import (
	"os"
	"strings"
)


func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func checkWalletNetwork(walletAddr string) int {
	if len(walletAddr) == 42 && strings.HasPrefix(walletAddr, "0x") {
		return EthNetwork // ETH
	} else if len(walletAddr) > 25 && len(walletAddr) < 36 && checkBTCFormat(walletAddr) {
		return BtcNetwork // BTC
	}
	return -1 // Others
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
