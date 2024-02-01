<template>
    <div class="ml-[420px] w-full">
        <div class="w-full">
        <div id="BG"></div>
            <div class="border-l border-green-500 w-full">
                <div 
                    class="
                        bg-[#F0F0F0] 
                        fixed 
                        z-10 
                        min-w-[calc(100vw-420px)] 
                        flex 
                        justify-between 
                        items-center 
                        px-2 
                        py-2
                    "
                >
                    <div class="flex items-center">
                        <img @click="showChatPartnerInfo"
                            v-if="userDataForChat && userDataForChat.image"
                            class="rounded-full mx-1 w-10 h-10 cursor-pointer object-cover" 
                            :src="userDataForChat.image"
                        >
                        <div 
                        v-if="userDataForChat && userDataForChat.firstname"
                        class="text-gray-900 text-lg ml-1 leading-4 font-semibold">
                            {{ contactName }}
                            <br>
                            <span v-if="isOnline" class="text-xs font-semibold text-green-500">{{ "online"}}</span>
                            <span v-if="!isOnline" class="text-xs font-semibold text-gray-500">{{ "offline" }}</span>
                        </div>
                    </div>

                    <DotsVerticalIcon fillColor="#515151" />
                </div>
            </div>

            <div 
            id="MessagesSection"
            class="
                pt-20 
                pb-8 
                z-[-1]
                h-[calc(100vh-65px)]
                w-[calc(100vw-420px)]
                overflow-auto
                fixed
                touch-auto
            "
            >
                <div v-if="currentChat" class="px-20 text-sm">
                    <div v-for="msg in currentChat.messages" v-on:click.right="deleteMessage($event, msg.idx)" :key="msg.idx">
                        <div v-if="parseInt(msg.sub) === id" class="flex w-[calc(100%-50px)]">
                            <div class="inline-block break-all bg-white p-2 rounded-md my-1 max-w-3xl">
                                {{ msg.message }}
                            </div>
                        </div>

                        <div v-else class="flex justify-end space-x-1 w-[calc(100%-50px)] float-right" >
                            <div class="inline-block break-all bg-green-200 p-2 rounded-md my-1">
                                {{ msg.message }}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- @input="handleInput()" -->
            <div class="w-[calc(100vw-420px)] p-2.5 z-10 bg-[#F0F0F0] fixed bottom-0">
                <div class="flex items-center justify-center">
                    <EmoticonExcitedOutlineIcon :size="27" fillColor="#515151" class="mx-1.5" />
                    <PaperclipIcon :size="27" fillColor="#515151" class="mx-1.5 mr-3" />
                    <input 
                    v-model="message"
                    class="
                        mr-1
                        shadow
                        apperance-none
                        rounded-lg
                        w-full
                        py-3
                        px-4
                        text-gray-700
                        leading-tight
                        focus:outline-none 
                        focus:shadow-outline
                    "
                    @keyup.enter="sendMessage"
                    autocomplete="off"
                    type="text"
                    placeholder="Message"
                    >

                    <button 
                        :disabled="disableBtn"
                        @click="sendMessage" 
                        class="ml-3 p-2 w-12 flex items-center justify-center"
                    >
                        <SendIcon fillColor="#515151" />
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import DotsVerticalIcon from 'vue-material-design-icons/DotsVertical.vue'
import EmoticonExcitedOutlineIcon from 'vue-material-design-icons/EmoticonExcitedOutline.vue'
import PaperclipIcon from 'vue-material-design-icons/Paperclip.vue'
import SendIcon from 'vue-material-design-icons/Send.vue'
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useUserStore } from '../store/user-store'
import { storeToRefs } from 'pinia'
const userStore = useUserStore()
const { chats, isNewContactModalVisible, isPartnerInfoVisible, isUserInfoVisible,
     userDataForChat, currentChat, id, contacts } = storeToRefs(userStore)

let sharedChatInfo
let message = ref('')
let disableBtn = ref(false)
let socket = null
let counter = 0
let isOnline = ref(false)
const contactName = ref("") 
const userStatus = ref("") 


window.onbeforeunload = function() {
    socket.onclose = function () {}; // disable onclose handler first
    socket.close();
};

watch(() => userDataForChat.value, (userData) => {
    socket.close()
    if (userData != null) {
        sharedChatInfo = chats.value.find((chat) => chat.chat_id == userDataForChat.value.chat_id);
        getContactName()
        startConnection()
    }
})

// onMounted(async () => {
//   let people = chat.value.people
//   let contactID = people[0]
//   if (people[0] == id.value) {
//     contactID = people[1]
//   }
//   var contact = contacts.value.find(item => item.id === contactID)
//   contactName.value = contact.name
// })

const deleteMessage = async (e, index) => {
    e.preventDefault();
    return
    await userStore.deleteMessage(index)
}

const startConnection = async () => {
    currentChat.value = {
        messages: [],
    }
    socket = new WebSocket(`ws://localhost:8080/api/chats/${userDataForChat.value.chat_id}`);
    socket.onopen = function () {
        console.log("Status: Connected\n")
    }
    socket.onmessage = async function (e) {
        const parts = e.data.split('|');
        console.log("onmessage:", e.data)
        currentChat.value.messages.push({
            idx: counter,
            message: parts[2],
            date: parts[1],
            sub: parts[0]
        })
        isOnline.value = false
        if (parts[3] == "fo") {
            isOnline.value = true
        }
        counter++
        sharedChatInfo.lastMess = parts[2]
        sharedChatInfo.lastMessDate = await userStore.formatDate(parts[1])
    } 
}

const showChatPartnerInfo = async () => {
    isNewContactModalVisible.value = false
    isUserInfoVisible.value = false
    isPartnerInfoVisible.value = !isPartnerInfoVisible.value
}

onMounted(async () => {
    sharedChatInfo = chats.value.find((chat) => chat.chat_id == userDataForChat.value.chat_id);
    getContactName()
    startConnection()
})

const getContactName = () => {
    let people = sharedChatInfo.people
    let contactID = people[0]
    if (people[0] == id.value) {
        contactID = people[1]
    }
    var contact = contacts.value.find(item => item.id === contactID)
    contactName.value = contact.name
}

onUnmounted(async () => {
    socket.onclose = function () {}; 
    socket.close();
})

const handleInput = async() => {
    setTimeout(function() {
        console.log("sendMessage:", "")
        socket.send("");
    }, 300);
}

const sendMessage = async() => {
    if (message.value === '') return 
    if (sharedChatInfo.lastUser != id.value) {
        sharedChatInfo.notviewed = 0
    }
    sharedChatInfo.notviewed += 1
    sharedChatInfo.lastUser = id.value
    var data = message.value
    console.log("sendMessage:", data)
    socket.send(data);
    message.value = ''
}
</script>

<style>
#BG {
    background: url('/message-bg.png') no-repeat center;
    width: 100%;
    height: 100%;
    position: fixed;
    z-index: -1;
}
</style>