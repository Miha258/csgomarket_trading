package main

import (
	"strconv"
	"time"

	"github.com/asmcos/requests"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/exp/slices"
)


func (a *App) GetInventoryItems() []interface{} {
	resp, _ := requests.Get(InventoryEndpoint(a.secretKey))

	var json map[string][]interface{}
	resp.Json(&json)

	return json["items"]
}


func (a *App) GetItemsOnSell() []interface{} {
	resp, _ := requests.Get(ItemsEndpoint(a.secretKey))

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


func (a *App) AddFolowItemHandler(hashName string, itemIds []map[string]interface{}) {
	defer a.FolowError(hashName)
	runtime.EventsEmit(a.ctx, "onItemFolowAdd", hashName)
	for _, itemId := range itemIds {
		if (itemId["id"] != nil) {
			minPrice := a.GetMinPrice(hashName)
			item_id, success, err := a.PutItemOnSale(itemId["id"].(string), minPrice)
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
			itemId["item_id"] = item_id
			}
			a.priceHandlers[hashName] = func(hashName string) {
			

			for _, itemId := range itemIds {
				time.Sleep(1 * time.Second)
				if (!a.IsItemOnSale(itemId["item_id"].(string))) { //Is item sold and is item selling 
					runtime.EventsEmit(a.ctx, "onItemFolowRemove", hashName)
					delete(a.priceHandlers, hashName)
				} else {
					minPrice := a.GetMinPrice(hashName)
					if minPrice != a.GetMinPrice(hashName) && minPrice != 0 {
						a.SetItemPrice(itemId["item_id"].(string), minPrice)
					}
				}
			}
			a.wg.Done()
			} 
		}
}


func (a *App) GetMinPrice(hashName string) (minPrice float64) {
	resp, _ := requests.Get(SearchItemEndpoint(a.secretKey, hashName))
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
	requests.Get(SetItemPriceEndpoint(a.secretKey, itemId, price))
}


func (a *App) PutItemOnSale(itemId string, price float64) (item_id string, success bool, error interface{}) {
	resp, _ := requests.Get(PutItemOnSaleEndpoint(a.secretKey, itemId, price))
	var json map[string]interface{}
	resp.Json(&json)
	if (!json["success"].(bool)){
		return "", false, json["error"]
	}
	return json["item_id"].(string), true, nil
}


func (a *App) UpdateItems([]string) {
	for k, _ := range a.priceHandlers {
		runtime.EventsEmit(a.ctx, "onItemFolowAdd", k)
	}
}


func (a *App) SetApiKey(apiKey string) interface{} {
	a.secretKey = apiKey
	resp, _ := requests.Get(TestEndpoint(a.secretKey))
	var json map[string]interface{}
	resp.Json(&json)
	
	if (!json["success"].(bool)){
		return "Неправельний api ключ"
	}
	return nil
}