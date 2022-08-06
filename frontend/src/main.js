import './style.css'
import { GetItems, AddFolowItemHandler, UpdateItems, SetApiKey } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime'


const App = {
    data() {
        return {
            items: [],
            itemHashNames: {}
        }
    },
    async mounted() {
        EventsOn("onError", error => {
            window.M.toast({ html: error })
        })

        EventsOn("onItemFolowRemove", (itemHashName) => {
            const item = document.getElementById(itemHashName)
            item.innerText = 'Відслідковувати'
            item.classList.remove('disabled')
        })

        EventsOn("onItemFolowAdd", itemHashName => {
            const item = document.getElementById(itemHashName)
            item.innerText = 'Відісідковується'
            item.classList.add('disabled')
        })
    },
    methods: {
        async folowItemPrice(event) {
            const item = event.target.parentElement.getElementsByTagName('h5')[0]
            AddFolowItemHandler(item.textContent, this.itemHashNames[item.textContent])
        },
        async updateItemList() {
            const apiKey = document.getElementById("apikey")
            if (apiKey.value){
                const err = await SetApiKey(apiKey.value)
                if (err){
                    window.M.toast({ html: err })
                } else {
                    this.items = await GetItems()
                    this.getCountOfItem()
                
                    const rows = (Math.ceil(this.items.length / 5)) + 1
                    const newItems = []
        

                    for (let i = 0; i < this.items.length; i += rows){ //Split on rows
                        const row = this.items.slice(i, i + rows);
                        newItems.push(row)
                    }
                    this.items = newItems
                    UpdateItems()
                }
            } else {
                window.M.toast({ html: "Апі ключ відсутній" })
            }
        },
        getCountOfItem(){
            this.itemHashNames = {}
            this.items.forEach(item => {
                const itemHashName = item.market_hash_name
                if (!this.itemHashNames[itemHashName]){
                    this.itemHashNames[itemHashName] = [{item_id: item.item_id, id: item.id}]
                } else {
                    const index = this.items.indexOf(item)
                    if (index !== -1) {
                        this.items.splice(index, 1);
                    }
                    this.itemHashNames[itemHashName].push({item_id: item.item_id, id: item.id})
                }
            })
        }
    },
}

Vue.createApp(App).mount('#app')