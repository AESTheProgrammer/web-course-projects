<template>
    <div class="popup" @click.self="changeUserInfoVisibility">
        <div v-if="isUserInfoVisible">
            <UserInfoView />
        </div>
    </div>
    <div class="popup" @click.self="changeUserInfoVisibility">
        <div v-if="isPartnerInfoVisible">
            <ChatPartnerInfoView />
        </div>
    </div>
    <div class="popup" @click.self="closeModal">
        <div v-if="isNewContactModalVisible">
            <AddContactView />
        </div>
    </div>
    <div class="flex">
        <div id="Header" class="fixed w-[420px] z-10">

            <div class="bg-[#F0F0F0] w-full flex justify-between items-center px-3 py-2">
                <img @click="changeUserInfoVisibility" class="rounded-full ml-1 h-10 w-10 cursor-pointer object-cover" :src="userStore.image || ''" alt="" style="border-radius: 50%;">
                <!-- <img @click="changeUserInfoVisibility" class="rounded-full ml-1 w-10 cursor-pointer" src="/unknown.png" alt="">  -->
                <div class="flex items-center justify-center">
                    <font-awesome-icon @click="newContact" icon="user-plus" style="color: #515151;" class="mr-6 cursor-pointer" />
                    <!-- <font-awesome-icon @click="newContact" icon="user-plus" style="color: #515151;" class="mr-6 cursor-pointer" /> -->
                    <AccountGroupIcon fillColor="#515151" class="mr-6" />
                    <DotsVerticalIcon @click="logout" fillColor="#515151" class=" cursor-pointer" />
                </div>
            </div>

            <div id="Search" class="bg-white w-full px-2 border-b shadow-sm">
                <div class="px-1 m-2 bg-[#F0F0F0] flex items-center justify-center rounded-md">
                    <MagnifyIcon fillColor="#515151" :size="18" class="ml-2" />
                    <input  v-model="searchStr"
                    @click="showFindFriends = !showFindFriends"
                    class="
                        ml-5
                        apperance-none
                        w-full
                        bg-[#F0F0F0]
                        py-1.5
                        px-2.5
                        text-gray-700
                        leading-tight
                        focus:outline-none 
                        focus:shadow-outline
                        placeholder:text-sm
                        placeholder:text-gray-500
                    "
                    autocomplete="off"
                    type="text"
                    placeholder="Start a new chat"
                    >
                </div>
            </div>
        </div>

        <div v-if="showFindFriends">
            <FindFriendsView class="pt-28"  :searchStr="searchStr"/>
        </div>
        <div v-else>
            <ChatsView class="mt-[100px]"/>
        </div>

        <div v-if="userDataForChat">
            <MessageView />
        </div>
        <div v-else>
            <div class="ml-[420px] fixed w-[calc(100vw-420px)] h-[100vh] bg-gray-100 text-center">
            <div class="grid h-screen place-items-center">
                <div>
                    <div class="w-full flex items-center justify-center">
                        <img width="375" src="/w-web-not-loaded-chat.png" alt="">
                    </div>
                    <div class="text-[32px] text-gray-500 font-light mt-10">WhatsApp Web</div>
                    <div class="text-[14px] text-gray-600 mt-2">
                        <div>Send and receive messages without keeping your phone online.</div>
                        <div>Use WhatsApp on up to 4 linked devices and 1 phone at the same time.</div>
                    </div>
                </div>
            </div>
        </div>
        </div>
    </div>
</template>

<script setup>
import ChatsView from './ChatsView.vue';
import MessageView from './MessageView.vue';
import FindFriendsView from './FindFriendsView.vue';

import AccountGroupIcon from 'vue-material-design-icons/AccountGroup.vue'
import DotsVerticalIcon from 'vue-material-design-icons/DotsVertical.vue'
import MagnifyIcon from 'vue-material-design-icons/Magnify.vue'
import { ref, onMounted } from 'vue';

import { useUserStore } from '@/store/user-store'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia';
import UserInfoView from './UserInfoView.vue';
import AddContactView from './AddContactView.vue'
import ChatPartnerInfoView from './ChatPartnerInfoView.vue';
const router = useRouter()
const userStore = useUserStore()
const searchStr = ref('')
const { showFindFriends, userDataForChat, isUserInfoVisible, isPartnerInfoVisible, isNewContactModalVisible, id } = storeToRefs(userStore)
isUserInfoVisible.value = false
isNewContactModalVisible.value = false

if(localStorage.getItem('token') == "" ||
    localStorage.getItem('token') == null || userStore.id == -1){
    router.push('/signin')
}
onMounted(async () => {
    await userStore.getUserDetails(id.value)
    try {
        await userStore.getAllContacts()
    } catch (error) {
        console.log(error)
    }
})

const logout = () => {
    let res = confirm('Are you sute you want to logout?')
    if (res) { userStore.logout(); router.push('/signin') }
}

const newContact = () => {
    isNewContactModalVisible.value = !isNewContactModalVisible.value
    isUserInfoVisible.value = false
    isPartnerInfoVisible.value = false
}

const changeUserInfoVisibility = () => { 
    isUserInfoVisible.value = !isUserInfoVisible.value
    isNewContactModalVisible.value = false
    isPartnerInfoVisible.value = false
}
</script>

<style lang="scss" scoped>

</style>