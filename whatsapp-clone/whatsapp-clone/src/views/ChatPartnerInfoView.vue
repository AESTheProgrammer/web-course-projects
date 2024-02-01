<template>
    <div class="popup">
        <div class="about-user-wrapper">
            <header class="d-flex justify-end pa-1">
                <button class="bg-gray-300 text-gray-800 text-white font-bold py-1 px-1 rounded inline-flex items-center">
                    <font-awesome-icon @click="closeWindow" icon="fa-solid fa-xmark" style="color: #edeeed;" class=" pa-2" />
                </button>
            </header>
            <div class="user-avatar d-flex justify-center flex-column align-center">
                <div>
                    <input class="pointer-events-none" type="file" id="imageInput" accept="image/*" @change="handleImageChange" style="display: none;" readonly/>
                    <label for="imageInput">
                        <img v-if="!imageUrl" src="/unknown.png" alt="Placeholder" class="rounded-full ml-1 h-20 w-20 object-cover " />
                        <img v-else :src="imageUrl" alt="Selected Image" class="rounded-full ml-1 h-20 w-20 object-cover " />
                    </label>
                </div>
                <input type="text" v-model="fullname" class="pointer-events-none user-info-name text-lg mt-3 bg-gray-50 border-none text-white rounded-lg focus:ring-white-500 focus:border-white block text-center w-half p-2.5" style="font-weight: bold; font-size:larger" placeholder="Joe Doe" readonly>
                <p class="last-seen">
                    {{ last_seen }}
                </p>
            </div>
            <div class="contacts d-flex">
                <div class="user-info-wrapper">
                    <div class="d-flex flex-column mb-4">
                        <div class="d-flex flex-row">
                            <input type="text" v-model="temp_phone" class="pointer-events-none info-text bg-gray-50 border-none text-white text-xs rounded-lg focus:ring-white-500 focus:border-white block w-full p-2.5" placeholder="123-45-678" readonly>
                        </div>
                        <span class="text-muted">Mobile</span>
                    </div>
                    
                    <div class="d-flex flex-column mb-4">
                        <div class="d-flex flex-row">
                            <input type="text" v-model="temp_bio" class="pointer-events-none info-text bg-gray-50 border-none text-white text-xs rounded-lg focus:ring-white-500 focus:border-white block w-full p-2.5" placeholder="I'm a doctor" readonly>
                        </div>
                        <span class="text-muted">Bio</span>
                    </div>
                    
                    <div class="d-flex flex-column mb-4">
                        <div class="d-flex flex-row">
                            <span class="info-text">@</span>
                            <input type="text" v-model="temp_username" class="pointer-events-none info-text bg-gray-50 border-none text-white text-xs rounded-lg focus:ring-white-500 focus:border-white block w-full p-2.5" placeholder="elonmusk" readonly>
                        </div>
                        <span class="text-muted">Username</span>
                    </div>
                </div>
                
            </div>
        </div>
        
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/store/user-store'
import { storeToRefs } from 'pinia';
const userStore = useUserStore()
const { isPartnerInfoVisible, userDataForChat, contactsInfo, id } = storeToRefs(userStore)
let fullname = ref("")
let temp_phone = ref("")
let temp_bio = ref("")
let temp_username = ref("")
let last_seen = ref("")
let imageUrl = ref("")

onMounted(async () => {
    let data = null
    let people = [userDataForChat.value.id1, userDataForChat.value.id2]
    if (people[0] == id.value) {
        data = contactsInfo.value[people[1]]
    } else {
        data = contactsInfo.value[people[0]]
    }
    fullname = data.firstname + " " + data.lastname
    temp_phone.value = data.phone
    temp_bio.value = data.bio
    temp_username.value = data.username
    imageUrl.value = data.image
    last_seen.value = data.last_seen
})

const closeWindow = () => {
    isPartnerInfoVisible.value = false
}

</script>

<style scoped>
@import 'vuetify/dist/vuetify.min.css'; 
#app {
    background: url('../img/bodyBackground.jpg') no-repeat center center;
    background-size: cover;
    backdrop-filter: blur(5px);
}

.user-info .user-info-name {
    color: #fafafa;
    font-size: 16px;
    font-weight: normal;
}

.user-info .user-info-wrapper {
    padding: 5px 20px;
}

.user-info-wrapper {
    align-items: center;
    justify-content: space-between;
}

.user-info .last-seen {
    color: #969696;
    margin: 0 !important;
    font-size: 14px;
}

.about-user-wrapper {
    border-radius: 10px;
    background-color: #1E1E1E;
    min-height: 60vh;
}

.text-muted {
    color: #A39698 !important;
    font-size: 12px;
    user-select: none;
    margin-top: 2px;
}

.info-text {
    font-size: 15px;
    color: white;
}

.user-info-name {
    color: white;
}

.last-seen {
    color: #A39698;
}

.select-none {
    user-select: none;
    padding-right: 1px;
}

.contacts {
    padding: 20px;
}

.popup {
    opacity: 95%;
    position: absolute;
    top: 50%;
    left: 50%;
    width: 320px;
    z-index: 99;
    transform: translate(-50%, -50%);
    margin-right: -50%;
    /* align-items: center;
    justify-content: center; */
}

.close-button {
    position: absolute;
    top: 10px;
    right: 10px;
    cursor: pointer;
    background: none;
    border: none;
    font-size: 20px;
    color: #333;
}

</style>