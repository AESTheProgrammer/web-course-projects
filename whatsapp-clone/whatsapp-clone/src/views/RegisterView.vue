<template>
    <layout-div>
        <div class="bg-teal-600 z-[-1] w-full h-full fixed top-0 left-0"></div>
        <div class="bg-[#191919] z-[-1] w-full h-[calc(100vh-350px)] fixed bottom-0 left-0"></div>
        <div class="row justify-content-md-center mt-5">
            <div class="col-4">
                <div class="card">
                    <div class="card-body">
                        <h5 class="card-title mb-4">Register</h5>
                        <form >
                            <div class="mb-3">
                                <label 
                                htmlFor="firstname"
                                class="form-label">Firstname
                            </label>
                            <input 
                            type="text"
                            class="form-control"
                            id="firstname"
                            name="firstname"
                            v-model="firstname"
                            placeholder="John"
                            />
                            <div v-if="validationErrors.firstname" class="flex flex-col">
                                <small  class="text-danger">
                                    {{validationErrors?.firstname[0]}}
                                </small >
                            </div>      
                        </div>
                        <div class="mb-3">
                            <label 
                            htmlFor="lastname"
                            class="form-label">Lastname
                        </label>
                        <input 
                        type="text"
                        class="form-control"
                        id="lastname"
                        name="lastname"
                        v-model="lastname"
                        placeholder="Doe"
                        />
                        <div v-if="validationErrors.lastname" class="flex flex-col">
                            <small  class="text-danger">
                                {{validationErrors?.lastname[0]}}
                            </small >
                        </div>      
                    </div>
                    
                    <div class="mb-3">
                        <label 
                        htmlFor="phone"
                        class="form-label">Phone
                    </label>
                    <input 
                    type="text"
                    class="form-control"
                    id="phone"
                    name="phone"
                    v-model="phone"
                    placeholder="123-45-678"
                    />
                    <div v-if="validationErrors.phone" class="flex flex-col">
                        <small  class="text-danger">
                            {{validationErrors?.phone[0]}}
                        </small >
                    </div>      
                </div>
                
                <div class="mb-3">
                    <label 
                    htmlFor="username"
                    class="form-label">Username
                </label>
                <input 
                type="text"
                class="form-control"
                id="username"
                name="username"
                v-model="username"
                placeholder="elonmusk" 
                />
                <div v-if="validationErrors.username" class="flex flex-col">
                    <small  class="text-danger">
                        {{validationErrors?.username[0]}}
                    </small >
                </div>      
            </div>
            
            <div class="mb-3">
                <label 
                htmlFor="bio"
                class="form-label">Bio
            </label>
            <input 
            type="text"
            class="form-control"
            id="bio"
            name="bio"
            v-model="bio"
            />
            <div v-if="validationErrors.bio" class="flex flex-col">
                <small  class="text-danger">
                    {{validationErrors?.bio[0]}}
                </small >
            </div>      
        </div>
        <div class="mb-3">
            <label 
            htmlFor="password"
            class="form-label">Password
        </label>
        <input 
        type="password"
        class="form-control"
        id="password"
        name="password"
        v-model="password"
        placeholder="•••••••••" 
        />
        <div v-if="validationErrors.password" class="flex flex-col">
            <small  class="text-danger">
                {{validationErrors?.password[0]}}
            </small >
        </div>
    </div>
    <div class="mb-3">
        <label 
        htmlFor="image"
        class="form-label">Photo
    </label>
    <input 
    type="file"
    class="form-control"
    id="image"
    name="image"
    @change="handleImageUpload"
    />
    
    <div v-if="validationErrors.password" class="flex flex-col">
        <small  class="text-danger">
            {{validationErrors?.password[0]}}
        </small >
    </div>
</div>
<div class="d-grid gap-2">
    <button 
    :disabled="isSubmitting"
    @click="registerAction()"
    type="button"
    class="btn btn-primary btn-block">Register Now
</button>
<label for="remember" class="ms-2 text-sm font-medium text-gray-900 dark:text-gray-300">I agree with the <a href="#" class="text-blue-600 hover:underline dark:text-blue-500">terms and conditions</a>.</label>
<p 
class="text-center">Have already an account <router-link to="/">Login here</router-link>
</p>
</div>
</form>
</div>
</div>
</div>
</div>
</layout-div>
</template>

<script>
import axios from 'axios';
import LayoutDiv from '@/components/LayoutDiv.vue';
import { useUserStore } from '@/store/user-store'

export default {
    name: 'RegisterView',
    components: {
        LayoutDiv,
    },
    data() {
        return {
            firstname:'',
            lastname:'',
            username:'',
            phone:'',
            password:'',
            image: null,
            bio: '', //{},
            validationErrors:{},
            isSubmitting:false,
        };
    },
    computed: {
        isFormValid() {
            // this.validationErrors = this.user.firstname && this.user.lastname && this.user.phone &&
            //                         this.user.username && this.user.password && this.user.bio && this.user.image
            return true
        },
    },
    created() {
        if(localStorage.getItem('token') != "" && localStorage.getItem('token') != null){
            this.$router.push('/')
        }
    },
    methods: {
        // generateRandomString() {
        //     const currentDate = new Date().toISOString();
        //     const hash = crypto.createHash('sha256').update(currentDate).digest('hex');
        //     const truncatedHash = hash.substring(0, 10); // Adjust the length as needed
        //     return truncatedHash;
        // },
        registerAction(){
            console.log("registering please wait")
            this.isSubmitting = true
            const formData = new FormData();
            formData.append('firstname', this.firstname);
            formData.append('lastname', this.lastname);
            formData.append('phone', this.phone);
            formData.append('username', this.username);
            formData.append('password', this.password);
            formData.append('Bio', this.bio);
            console.log(this.image)
            if (this.image) {
                formData.append('image', this.image);
            }
            axios.post('api/register', formData)
            .then(async response => {
                // console.log("new token is:", response)
                localStorage.setItem('token', response.data.token)
                const userStore = useUserStore()
                // userStore.getUserDetails(response.data.token, response.data.userid)
                await userStore.setID(response.data.userid)
                this.$router.push('/')
                return response
            })
            .catch(error => {
                this.isSubmitting = false
                console.log(error.response, error)
                if (error.response) {
                    this.validationErrors = error.response.data.errors
                }
                return error
            });
        },
        handleImageUpload(event) {
            this.image = event.target.files[0];
        },
   },
};
</script>
