package main

import (
	"fmt"
)


func InventoryEndpoint(secretKey string) string {
	return fmt.Sprintf("https://market.csgo.com/api/v2/my-inventory/?key=%s", secretKey)
}

func ItemsEndpoint(secretKey string) string {
	return fmt.Sprintf("https://market.csgo.com/api/v2/items?key=%s", secretKey)
}

func SearchItemEndpoint(secretKey string, hashName string) string {
	return fmt.Sprintf("https://market.csgo.com/api/v2/search-item-by-hash-name-specific?key=%s&hash_name=%s", secretKey, hashName)
}

func SetItemPriceEndpoint(secretKey string, itemId string, price float64) string {
	return fmt.Sprintf("https://market.csgo.com/api/v2/set-price?key=%s&item_id=%s&price=%f&cur=USD", secretKey, itemId, price * 1000)
}

func PutItemOnSaleEndpoint(secretKey string, itemId string, price float64) string {
	return fmt.Sprintf("https://market.csgo.com/api/v2/add-to-sale?key=%s&id=%s&price=%f&cur=USD", secretKey, itemId, price * 1000)
}

func TestEndpoint(secretKey string) string {
	return fmt.Sprintf("https://market.csgo.com/api/v2/test?key=%s", secretKey)
}



