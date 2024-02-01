<template>
    <layout-div>
        <div class="bg-teal-600 z-[-1] w-full h-full fixed top-0 left-0"></div>
        <div class="bg-[#191919] z-[-1] w-full h-[calc(100vh-250px)] fixed bottom-0 left-0"></div>
        <div class="row justify-content-md-center mt-5 ">
            <div class="col-4">
                <div class="card">
                    <div class="card-body">
                        <h5 class="card-title mb-4">Sign In</h5>
                        <form>
                            <p v-if="Object.keys(validationErrors).length != 0" class='text-center '><small class='text-danger'>Incorrect Username or Password</small></p>
                            <div class="mb-3">
                                <label 
                                htmlFor="username"
                                class="form-label">
                                Username
                            </label>
                            <input 
                            v-model="username"
                            type="username"
                            class="form-control"
                            style="border-top: 1px; border-right:1px; border-left:1px;"
                            id="username"
                            name="username"
                            placeholder="elonmusk"
                            />
                        </div>
                        <div class="mb-3">
                            <label 
                            htmlFor="password"
                            class="form-label">Password
                        </label>
                        <input 
                        v-model="password"
                        type="password"
                        class="form-control"
                        style="border-top: 1px; border-right:1px; border-left:1px"
                        id="password"
                        name="password"
                        placeholder="•••••••••" 
                        />
                    </div>
                    <div class="d-grid gap-2">
                        <button 
                        :disabled="isSubmitting"
                        @click="loginAction()"
                        type="button"
                        class="btn btn-primary btn-block">Login</button>
                        <p class="text-center">Don't have account? 
                            <router-link to="/register">Register here </router-link>
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
    name: 'SigninView',
    components: {
        LayoutDiv,
    },
    data() {
        return {
            username:'',
            password:'',
            validationErrors:{},
            isSubmitting:false,
        };
    },
    created() {
        const userStore = useUserStore()
        if(localStorage.getItem('token') != "" && localStorage.getItem('token') != null && userStore.id != -1){
            console.log(userStore.id)
            this.$router.push('/')
        }
    },
    methods: {
        loginAction(){
            this.isSubmitting = true
            let payload = {
                username: this.username,
                password: this.password,
            }
            axios.post('api/login', payload)
            .then(async response => {
                localStorage.setItem('token', response.data.token)
                const userStore = useUserStore()
                // userStore.getUserDetails(response.data.token, response.data.userid)
                await userStore.setID(response.data.userid)
                this.$router.push('/')
                return response
            })
            .catch(error => {
                this.isSubmitting = false
                if (error.response && error.response.data.errors != undefined) {
                    this.validationErrors = error.response.data.errors
                }
                if (error.response && error.response.data.error != undefined) {
                    this.validationErrors = error.response.data.error
                }
                return error
            });
        }
    },
};
</script>
