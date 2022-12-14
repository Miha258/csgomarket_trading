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

        EventsOn("onItemFolowRemove", itemHashName => {
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
            document.getElementsByTagName('h5')[0]
            const item = event.target.parentElement.getElementsByTagName('h5')[0]
            const min = parseFloat(item.parentElement.getElementsByTagName('input')[0].value)
            const max = parseFloat(item.parentElement.getElementsByTagName('input')[1].value)
            AddFolowItemHandler(item.textContent, this.itemHashNames[item.textContent], min, max)
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
                    
                    const buttons = document.getElementsByTagName('button')
                    for (let i = 0; i < (buttons.length - 1); i++){
                        const btn = buttons.item(i)
                        btn.innerText = 'Відслідковувати'
                        btn.classList.remove('disabled')
                    }
                    UpdateItems()
                }
            } else {
                window.M.toast({ html: "Апі ключ відсутній" })
            }
        },
        getCountOfItem(){
            this.itemHashNames = {}
            this.items.forEach(item => {
                if (this.itemHashNames[item.market_hash_name]){
                    this.itemHashNames[item.market_hash_name].push({item_id: item.item_id, id: item.id})
                } else {
                    this.itemHashNames[item.market_hash_name] = [{item_id: item.item_id, id: item.id}]
                }
            })
            this.items = this.itemHashNames
        }
    }
}

Vue.createApp(App).mount('#app')