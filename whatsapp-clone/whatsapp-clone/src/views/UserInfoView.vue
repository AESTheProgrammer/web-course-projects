<template>
  <div class="popup">
    <div class="about-user-wrapper">
      <header class="d-flex justify-end pa-1">
        <button class="bg-gray-300 text-gray-800 text-white font-bold py-1 px-1 rounded inline-flex items-center">
          <font-awesome-icon @click="closeWindow" icon="fa-solid fa-xmark" style="color: #edeeed;" class="cursor-pointer pa-2" />
        </button>
      </header>
      <div class="user-avatar d-flex justify-center flex-column align-center">
        <div>
          <input type="file" id="imageInput" accept="image/*" @change="handleImageChange" style="display: none;" />
          <label for="imageInput">
            <img v-if="!imageUrl" src="/unknown.png" alt="Placeholder" class="rounded-full ml-1 h-20 w-20 object-cover cursor-pointer" />
            <img v-else :src="imageUrl" alt="Selected Image" class="rounded-full ml-1 h-20 w-20 object-cover cursor-pointer" />
          </label>
          <font-awesome-icon class="relative bottom-3 right-5" size="xl" :icon="['fas', 'camera']" style="color: #74C0FC;" />
        </div>
        <input type="text" v-model="fullname" class="user-info-name text-lg mt-3 bg-gray-50 border-none text-white rounded-lg focus:ring-white-500 focus:border-white block text-center w-half p-2.5" style="font-weight: bold; font-size:larger" placeholder="Joe Doe" required>
        <p class="last-seen ">
          {{ "just now" }}
        </p>
      </div>
      <div class="contacts d-flex">
        <div class="user-info-wrapper">
          <div class="d-flex flex-column mb-4">
            <div class="d-flex flex-row">
              <input type="text" v-model="temp_phone" class="info-text bg-gray-50 border-none text-white text-xs rounded-lg focus:ring-white-500 focus:border-white block w-full p-2.5" placeholder="123-45-678" required>
            </div>
            <span class="text-muted">Mobile</span>
          </div>
          
          <div class="d-flex flex-column mb-4">
            <div class="d-flex flex-row">
              <input type="text" v-model="temp_bio" class="info-text bg-gray-50 border-none text-white text-xs rounded-lg focus:ring-white-500 focus:border-white block w-full p-2.5" placeholder="I'm a doctor" required>
            </div>
            <span class="text-muted">Bio</span>
          </div>
          
          <div class="d-flex flex-column mb-4">
            <div class="d-flex flex-row">
              <span class="info-text">@</span>
              <input type="text" v-model="temp_username" class="info-text bg-gray-50 border-none text-white text-xs rounded-lg focus:ring-white-500 focus:border-white block w-full p-2.5" placeholder="elonmusk" required>
            </div>
            <span class="text-muted">Username</span>
          </div>
        </div>
        
      </div>
      <div class="d-flex justify-center flex-column align-center">
        <button @click="updateUserInfo" class="relative bg-gray-300 text-gray-800 text-white font-bold py-0 px-0 rounded inline-flex items-center">
          <svg class="fill-current w-5 h-5 mr-2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 448 512"><path fill="#ffffff" d="M48 96V416c0 8.8 7.2 16 16 16H384c8.8 0 16-7.2 16-16V170.5c0-4.2-1.7-8.3-4.7-11.3l33.9-33.9c12 12 18.7 28.3 18.7 45.3V416c0 35.3-28.7 64-64 64H64c-35.3 0-64-28.7-64-64V96C0 60.7 28.7 32 64 32H309.5c17 0 33.3 6.7 45.3 18.7l74.5 74.5-33.9 33.9L320.8 84.7c-.3-.3-.5-.5-.8-.8V184c0 13.3-10.7 24-24 24H104c-13.3 0-24-10.7-24-24V80H64c-8.8 0-16 7.2-16 16zm80-16v80H272V80H128zm32 240a64 64 0 1 1 128 0 64 64 0 1 1 -128 0z"/></svg>
          <span>Save</span>
        </button>
      </div>
    </div>
    
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/store/user-store'
import { storeToRefs } from 'pinia';
const userStore = useUserStore()
const { username, firstname, lastname, phone, bio, isUserInfoVisible, image } = storeToRefs(userStore)
let fullname = ref(firstname.value + " " + lastname.value)
let temp_phone = ref("")
let temp_bio = ref("")
let temp_username = ref("")
let imageUrl = ref("")
let file = ref(null) 

onMounted(async () => {
  temp_phone.value = phone.value
  temp_bio.value = bio.value
  temp_username.value = username.value
  fullname.value = firstname.value + " " + lastname.value
  imageUrl.value = image.value
})

const closeWindow = () => {
  isUserInfoVisible.value = false
}

const handleImageChange = async (event) => {
  file.value = event.target.files[0];
  if (file.value) {
    const reader = new FileReader();
    reader.readAsDataURL(file.value);
    reader.onload = () => {
      imageUrl.value = reader.result;
    };
  }
}

const urltoFile = async (url, filename, mimeType) => {
  if (url.startsWith('data:')) {
    var arr = url.split(','),
    mime = arr[0].match(/:(.*?);/)[1],
    bstr = atob(arr[arr.length - 1]), 
    n = bstr.length, 
    u8arr = new Uint8Array(n);
    while(n--){
      u8arr[n] = bstr.charCodeAt(n);
    }
    var file = new File([u8arr], filename, {type:mime || mimeType});
    return Promise.resolve(file);
  }
  const res = await fetch(url);
  const buf = await res.arrayBuffer();
  return new File([buf], filename, { type: mimeType });
}


const updateUserInfo = async () => {
  const splitIndex = fullname.value.indexOf(' ');
  if (splitIndex !== -1) { 
    const firstname = fullname.value.substring(0, splitIndex); 
    const lastname = fullname.value.substring(splitIndex + 1); 
    let new_file = await urltoFile(imageUrl.value, "image.jpeg",  "image/jpeg")
    let res = await userStore.updateUserInfo(temp_phone.value, temp_bio.value, temp_username.value, firstname, lastname, new_file, imageUrl.value)
    if (res == -1) {
      alert("something went wrong")
    } else {
      isUserInfoVisible.value = false     
    }
  } else {
    alert("Fullname should be atleast two parts.")
    return 
  }
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