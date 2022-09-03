package main

import (
	"time"
	"github.com/asmcos/requests"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/exp/slices"
	"math"
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


func (a *App) AddFolowItemHandler(hashName string, itemIds []map[string]interface{}, min float64, max float64) {
	defer a.FolowError(hashName)
	runtime.EventsEmit(a.ctx, "onItemFolowAdd", hashName)
	for _, itemId := range itemIds {
		if (itemId["id"] != nil) {
			minPrice := a.GetMinPrice(hashName)
			var item_id string
			var success bool
			var err interface{}
			
			if (minPrice < min && min != 0){
				item_id, success, err = a.PutItemOnSale(itemId["id"].(string), min)
			} else if (minPrice > max && max != 0){
				item_id, success, err = a.PutItemOnSale(itemId["id"].(string), max)
			} else {
				item_id, success, err = a.PutItemOnSale(itemId["id"].(string), minPrice)
			}
			
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
		}
	a.priceHandlers[hashName] = func(hashName string) {
		minPrice := a.GetMinPrice(hashName)
		for _, itemId := range itemIds {
			itemId := itemId["item_id"].(string)
			if (!a.IsItemOnSale(itemId)) { //Is item sold and is item selling 
				runtime.EventsEmit(a.ctx, "onItemFolowRemove", hashName)
				delete(a.priceHandlers, hashName)
			} else {
				if minPrice != 0 {
					if (minPrice < min && min != 0){
						a.SetItemPrice(itemId, min)
					} else if (minPrice > max && max != 0){
						a.SetItemPrice(itemId, max)
					} else {
						a.SetItemPrice(itemId, minPrice)
					}
				}
			}
		}
		time.Sleep(2000 * time.Millisecond)
		a.wg.Done()
	}
}



func (a *App) GetMinPrice(hashName string) (float64) {
	resp, _ := requests.Get(SearchItemEndpoint(a.secretKey, hashName))
	var json map[string][]map[string]float64
	resp.Json(&json)
	minPrice := json["data"][0]["price"] / 1000
	for _, item := range json["data"] {
		price := item["price"] / 1000
		minPrice = math.Min(price, minPrice)
	}
	return minPrice - 0.001
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


// func (a *App) GetCurrentPrice(itemId string) float64 {
// 	items := a.GetItemsOnSell()
	
// 	for _, item := range items {
// 		id := item.(map[string]interface{})["item_id"]
// 		if (id == itemId){
// 			price := item.(map[string]interface{})["price"].(float64)
// 			return price 
// 		}
// 	}
// 	return 0
// }


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
