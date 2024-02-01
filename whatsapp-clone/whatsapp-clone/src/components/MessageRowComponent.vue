<template>
  <div :class="isActive ? 'bg-gray-200' : ''">
    <div class="flex w-full px-4 py-3 items-center cursor-pointer">
      <img class="rounded-full mr-4 h-12 w-12 object-cover" :src="chat.image || ''" style="border-radius: 100%;">
      <div class="w-full">
        
        <div class="flex justify-between items-center">
          <div class="text-[15px] text-gray-600">
            {{ contactName }}
          </div>
          <div class="text-[12px] text-gray-600">
            {{ chat.lastMessDate }}
          </div>
        </div>
        
        <div class="flex items-center">
          <CheckAllIcon  v-if="isTickVisible(chat)" :size="18" :fillColor="tickColor(chat)" class="mr-1" />
          <div class="text-[15px] break-all w-full text-gray-500 flex items-center justify-between">
            {{ formatLastMessage(chat.lastMess) }}
          </div>
        </div>
      </div>
      
      <div v-if="isMsgCountVisible()" class="mt-4 w-6 h-5 bg-emerald-400 flex items-center justify-center rounded-full text-xs font-semibold text-teal-50	" >
        {{ chat.notviewed }}
      </div>
    </div>
    
    <div class="border-b w-[calc(100%-80px)] float-right"></div>
    
  </div>
</template>

<script setup>
import CheckAllIcon from 'vue-material-design-icons/CheckAll.vue'
import { toRefs, computed, ref, onMounted } from 'vue';
import { useUserStore } from '../store/user-store';
import { storeToRefs } from 'pinia';
const userStore = useUserStore()
const { id, userDataForChat, contacts } = storeToRefs(userStore)
const props = defineProps({ chat: Object })
const { chat } = toRefs(props)
const contactName = ref("") 
// let last_seen = ref("")

onMounted(async () => {
  let people = chat.value.people
  let contactID = people[0]
  if (people[0] == id.value) {
    contactID = people[1]
  }
  var contact = contacts.value.find(item => item.id === contactID)
  contactName.value = contact.name
})

const isActive = computed(() => {
  if (userDataForChat.value != null) {
    if (userDataForChat.value.chat_id === chat.value.chat_id) {
      return true
    }
  }
  return false
})

const tickColor = (chat) => {
  let color = '#FCFCFC'
  console.log("tickcolor")
  console.log(chat.notviewed)
  console.log(chat.lastUser)
  if (chat.lastUser == id.value) {
    if (chat.notviewed == 0) color = '#06D6A0'
    else color = '#B5B5B5'
  }
  return color
}

const isTickVisible = (chat) => {
  console.log(id.value, "==", chat.lastUser)
  return chat.lastUser == id.value
}

const formatLastMessage = (msg) => {
  let sub = msg.substring(0,30)
  if (msg.length > 28) {
    sub += "..."
  }
  return sub
}

const isMsgCountVisible = () => {
  // console.log(chat.value.lastUser, id.value,  )
  if (chat.value.lastUser != id.value) {
    if (chat.value.notviewed != 0) {
      return true
    }
  }
  return false
}

</script>