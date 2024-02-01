import { createApp } from "vue";
import "./style.css";
import App from "./App.vue";
import router from "@/router";
import { createPinia } from "pinia";
import piniaPluginPersistedstate from "pinia-plugin-persistedstate";
import { library } from '@fortawesome/fontawesome-svg-core'
import { faUserPlus, faXmark, faMagnifyingGlass, faPlus, faCamera } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(faUserPlus, faXmark, faMagnifyingGlass, faPlus, faCamera);

const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);
const app = createApp(App).component("font-awesome-icon", FontAwesomeIcon);
app.use(pinia);
app.use(router);
app.mount("#app");
