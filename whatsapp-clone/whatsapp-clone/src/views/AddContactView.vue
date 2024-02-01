<template>
  <div class="popup">
    
    <div class="about-user-wrapper">
      <header class="about-header d-flex justify-end pa-2">
        <font-awesome-icon @click="closeWindow" icon="fa-solid fa-xmark" style="color: #edeeed;" class="cursor-pointer" />
        <!-- <font-awesome-icon icon="rectangle-xmark" /> -->
      </header>
      
      <div class="user-avatar d-flex justify-center flex-column align-center">
        <h2 class="user-info-name pa-5">Add Contact</h2>
        <form @submit.prevent="submitForm">
          <label class="form-label mb-2" style="color: white" for="userid">User ID:</label>
          <input class="form-control mb-4" style="height: 35px; width: 210px; border:1px solid white" type="text" id="Iuserid" v-model="userId" placeholder="Contact ID" required />
          <div v-if="showError" class="mb-5">
            <p class="error">User ID not found.</p>
          </div>
          <!-- <font-awesome-icon @click="addContact" icon="fa-solid fa-magnifying-glass" style="color: #ffffff;" class="cursor-pointer mr4" /> -->
          <label class="form-label mb-2" style="color: white" for="userid">Firstname:</label>
          <input class="form-control mb-10" style="height: 35px; width:210px; border:1px solid white" type="text" id="Ifirstname" v-model="firstname" placeholder="Firstname" required />
        </form>
        <font-awesome-icon @click="addContact" icon="fa-solid fa-plus" size="xl" style="color: #ffffff;" class="cursor-pointer mr4"/>
      </div>
    </div>
  </div> 
</template>

<script setup>
import { ref } from 'vue'
import { useUserStore } from '@/store/user-store'
import { storeToRefs } from 'pinia';
const userStore = useUserStore()
const { isNewContactModalVisible } = storeToRefs(userStore)
const userId = ref('')
const firstname = ref('')
const showError = ref(false)


const closeWindow = () => {
  isNewContactModalVisible.value = false
}

const addContact = async () => {
  try {
    showError.value = await userStore.addNewContact(parseInt(userId.value), firstname.value);
    if (!showError.value) {
      isNewContactModalVisible.value = false;
    }
  } catch (error) {
    console.error('Error adding contact:', error);
  }
};
</script>

<style scoped>
@import 'vuetify/dist/vuetify.min.css'; 
#app {
  background: url('../img/bodyBackground.jpg') no-repeat center center;
  background-size: cover;
  backdrop-filter: blur(5px);
}

input {
  color: white;
  border-radius: 1px;
}

input::placeholder {
  font-weight: bold;
  opacity: 0.5;
  font-size: 13px;
  color: white;
}


.error {
  color: red;
  font-size: 12px;
  font-weight: normal;
}

.user-info .user-info-name {
  color: #fafafa;
  font-size: 16px;
  font-weight: normal;
  
}

.about-user-wrapper {
  border-radius: 10px;
  background-color: #1E1E1E;
  min-height: 45vh;
  align-items: center;
  justify-content: center;
}

.user-info-name {
  color: white;
}

.popup {
  opacity: 95%;
  position: fixed;
  top: 50%;
  left: 50%;
  width: 340px;
  z-index: 99;
  transform: translate(-50%, -50%);
  margin-right: -50%;
  align-items: center;
  justify-content: center;
}

button {
  padding: 10px;
  background-color: #3498db;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

</style>