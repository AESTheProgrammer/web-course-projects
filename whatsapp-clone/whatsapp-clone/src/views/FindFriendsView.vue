<template>
    <div id="FindFriends" class="pt-[100px] overflow-auto fixed h-[100vh] w-full">
        <div v-for="user in filterByInfo()" :key="user.id">
            <div @click="createNewChat(user)" v-on:click.right="deleteContact($event, user.id)" class="flex w-full p-4 items-center cursor-pointer">
                <img class="rounded-full h-12 w-12 object-cover" :src="contactsInfo[user.id].image || ''">
                <div class="w-full">
                    <div class="flex justify-between items-center">
                        <div class="text-[15px] text-gray-600">{{  user.firstName  }} {{  user.lastName  }}</div>
                    </div>
                    <div class="flex items-center">
                        <div class="text-[15px] text-gray-500">Hi, I'm using WhatsApp!</div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, onMounted, toRefs } from 'vue'
import { useUserStore } from "@/store/user-store";
import { storeToRefs } from "pinia";
import { faUserFriends } from '@fortawesome/free-solid-svg-icons';
const userStore = useUserStore()
const { id, chats, contacts, contactsInfo, showFindFriends } = storeToRefs(userStore)
let users = ref([])
let searchRes = ref([])
const props = defineProps({ searchStr: String })
const { searchStr } = toRefs(props)

const deleteContact = async (e, contact_id) => {
    e.preventDefault();
    await userStore.deleteContact(contact_id)
}

watch([() => contacts.value.length, () => chats.value.length], (len1, len2) => {
    users.value = []
    contacts.value.forEach(user => users.value.push(user))
    chats.value.forEach(chat => {
        let index = users.value.findIndex(user => user.id == chat.people[0] || user.id == chat.people[1])
        if (index != -1) {
            chat.firstname = users.value[index].name
            users.value.splice(index, 1)
        }
    })
})

// watch(() => searchStr.value, (ss) => {
//     filterByInfo(users)
// })

onMounted(async () => {
    contacts.value.forEach(user => users.value.push(user))
    chats.value.forEach(chat => {
        let index = users.value.findIndex(user => user.id == chat.people[0] || user.id == chat.people[1])
        if (index != -1) {
            chat.firstname = users.value[index].name
            users.value.splice(index, 1)
        }
    })
    console.log(users.value)
})


const createNewChat = async (user) => {
    searchStr.value = ""
    showFindFriends.value = false 
    let people = [id.value, user.id]
    let index = users.value.findIndex(u => u.id == user.id)
    if (index != -1) {
        users.value.splice(index, 1)
    }
    await userStore.createNewChat(people, user.name)
}

const filterByInfo = () => {
    let ss = searchStr.value
    if (ss == "" || users == null || users.length == 0)
        return users.value
    console.log(users.value)
    return users.value.filter(user => {
        let userInfo = contactsInfo.value[user.id]
        if (userInfo.firstname.search(ss) != -1 ||
            userInfo.lastname.search(ss) != -1 ||
            userInfo.username.search(ss) != -1 ||
            userInfo.phone.search(ss) != -1
         ) {
            return true 
         }
    });
}

</script>

<style lang="scss" scoped>

</style>