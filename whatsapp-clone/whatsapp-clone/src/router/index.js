import { createRouter, createWebHistory } from "vue-router";
import HomeView from '@/views/HomeView.vue'
import SigninView from '@/views/SigninView.vue'
import RegisterView from '@/views/RegisterView.vue'
import 'bootstrap/dist/css/bootstrap.css';


const routes = [
  {
    path: "/",
    component: HomeView,
  },
  {
    path: "/signin",
    component: SigninView,
  },
  {
    path: "/register",
    component: RegisterView,
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
