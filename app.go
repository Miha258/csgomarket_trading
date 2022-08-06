package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/asmcos/requests"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/exp/slices"
)


func (a *App) GetInventoryItems() []interface{} {
	resp, _ := requests.Get("https://market.csgo.com/api/v2/my-inventory/?key=" + a.secretKey)

	var json map[string][]interface{}
	resp.Json(&json)

	return json["items"]
}


func (a *App) GetItemsOnSell() []interface{} {
	resp, _ := requests.Get("https://market.csgo.com/api/v2/items?key=" + a.secretKey)

	var json map[string][]interface{}
	resp.Json(&json)
	
	return json["items"]
}


func (a *App) IsItemOnSale(itemId string) bool {
	items := a.GetItemsOnSell()
	itemsIds := make([]string, len(items))
	
	for _, item := range items {
		id := item.(map[string]interface{})["item_id"]
		itemsIds = append(itemsIds, id.(string))
	}
	return slices.Contains(itemsIds, itemId)
}


func (a *App) GetItems() []interface{} {
	items := a.GetInventoryItems()
	items = append(items, a.GetItemsOnSell()...)
	return items
}


func (a *App) AddFolowItemHandler(hashName string, itemIds []string) {
	for _, itemId := range itemIds {
		if (!a.IsItemOnSale(itemId)) { //Is item put on sell
			success, err := a.PutItemOnSale(itemId, a.GetMinPrice(hashName))
			if (!success){
				switch err.(string) {
				case "inventory_not_loaded":
				case "item_not_recieved":
					runtime.EventsEmit(a.ctx, "onError", "Перезагрузіть інвентар на сайті")
					return
				case "item_not_in_inventory":
					runtime.EventsEmit(a.ctx, "onError", "Предмет не знайдено в вашому інвентарі.Або він виставлений на продаж")
					return
				case "item_not_inserted":
					runtime.EventsEmit(a.ctx, "onError", "Помилка при виставлянні на продаж")
					return
				case "bad_request":
					runtime.EventsEmit(a.ctx, "onError", "Немає звязку з сайтом")
					return
					}
				}
			}
		}
		runtime.EventsEmit(a.ctx, "onItemFolowAdd", hashName)
		a.priceHandlers[hashName] = func(hashName string) {
			defer a.FolowError(hashName)
			for _, itemId := range itemIds {
				runtime.LogPrintf(a.ctx, "%b", a.IsItemOnSale(itemId))
				time.Sleep(1 * time.Second)
				if (!a.IsItemOnSale(itemId)) { //Is item sold and is item selling 
					runtime.EventsEmit(a.ctx, "onItemFolowRemove", hashName)
					delete(a.priceHandlers, hashName)
				} else {
					minPrice := a.GetMinPrice(hashName)
					if minPrice != a.GetMinPrice(hashName) && minPrice != 0 {
						a.SetItemPrice(itemId, minPrice)
					}
				}
			}
			a.wg.Done()
		} 
}


func (a *App) GetMinPrice(hashName string) (minPrice float64) {
	defer a.PriceError()

	resp, _ := requests.Get("https://market.csgo.com/api/v2/search-item-by-hash-name-specific?key=" + a.secretKey + "&hash_name=" + hashName)
	var json map[string][]map[string]interface{}
	resp.Json(&json)

	float := json["data"][0]["extra"].(map[string]interface{})["float"]

	if minPrice, err := strconv.ParseFloat(float.(string), 64); err == nil {
		for _, item := range json["data"] {
			extra := item["extra"].(map[string]interface{})
			price := extra["float"].(string)
			if price, err := strconv.ParseFloat(price, 64); err == nil {
				if price < minPrice && price != 0 {
					minPrice = price
				}
			}
		}
		return minPrice
	} 
	panic("Invalid float value")
}


func (a *App) RemoveItemFollow(hashName string) {
	delete(a.priceHandlers, hashName)
	runtime.EventsEmit(a.ctx, "onItemFolowRemove", hashName)
}


func (a *App) SetItemPrice(itemId string, price float64) {
	requests.Get("https://market.csgo.com/api/v2/set-price?key=" + a.secretKey + "&item_id=" + itemId + "&price=" + fmt.Sprintf("%f", price) + "&cur=USD")
}


func (a *App) PutItemOnSale(itemId string, price float64) (success bool, error interface{}) {
	resp, _ := requests.Get("https://market.csgo.com/api/v2/add-to-sale?key=" + a.secretKey + "&id=" + itemId + "&price=" + fmt.Sprintf("%f", price * 1000) + "&cur=USD")
	var json map[string]interface{}
	resp.Json(&json)
	if (!json["success"].(bool)){
		return false, json["error"]
	}
	return true, nil
}






func (a *App) UpdateItems([]string) {
	for k := range a.priceHandlers {
		runtime.EventsEmit(a.ctx, "onItemFolowAdd", k)
	}
}


func (a *App) SetApiKey(apiKey string) interface{} {
	a.secretKey = apiKey
	
	resp, _ := requests.Get("https://market.csgo.com/api/v2/test?key=" + a.secretKey)
	var json map[string]interface{}
	resp.Json(&json)
	
	if (!json["success"].(bool)){
		return "Неправельний api ключ"
	}
	return nil
}