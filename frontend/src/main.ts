import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import './assets/base.css'
import '../node_modules/frappe-gantt/dist/frappe-gantt.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
