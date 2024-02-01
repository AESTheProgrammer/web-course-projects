<template>
    <div 
    id="Messages" 
    class="pt-1 z-0 overflow-auto fixed h-[calc(100vh-100px)] w-[420px]"
    >
    <div v-for="chat in chats" :key="chat.chat_id">
        <div @click="openChat(chat)" v-on:click.right="deleteChat($event, chat.chat_id)">
            <MessageRowComponent :chat="chat"/>
        </div>
    </div>
</div>
</template>

<script setup>
import { useUserStore } from '@/store/user-store'
import MessageRowComponent from '@/components/MessageRowComponent.vue';
import { storeToRefs } from 'pinia'
import { onMounted } from 'vue';
const userStore = useUserStore()
const { chats, userDataForChat } = storeToRefs(userStore)


const deleteChat = async (e, chat_id) => {
    e.preventDefault();
    await userStore.deleteChat(chat_id)
}

const openChat = async (chat) => {
    chat.notviewed = 0
    userDataForChat.value = {
        chat_id: chat.chat_id,
        id1: chat.people[0],
        id2: chat.people[1],
        firstname: chat.username,
        image: chat.image,
    }
}
</script>

<style lang="scss" scoped>

</style>