import { defineStore } from 'pinia'
import axios from 'axios'
axios.defaults.baseURL = 'http://localhost:8080/'

export const useUserStore = defineStore('user', {
  state: () => ({
    id: -1,
    image: null,
    firstname: '',
    lastname: '',
    username: '',
    phone: '',
    bio: '',
    isUserInfoVisible: false,
    chats: [],
    contactsInfo: {},
    contacts: [],
    userDataForChat: null,
    showFindFriends: false,
    isNewContactModalVisible: false,
    isPartnerInfoVisible: false,
    currentChat: null,
    removeUsersFromFindFriends: [],
  }),
  getters: {
    contactsLen: (state) => state.contacts.length,
  },
  actions: {
    async getUserDetails(id) {
      this.userDataForChat = null
      this.showFindFriends = false
      this.isPartnerInfoVisible= false
      this.currentChat = null
      this.extensionInfo = {}
      this.removeAllData()
      this.id = id
      try {
        let res = await axios.get(`api/users/${id}`, {
          token: localStorage.getItem("token")
        })
        this.bio = res.data.bio
        this.firstname = res.data.firstname
        this.lastname = res.data.lastname
        this.username = res.data.username
        this.phone = res.data.phone
        this.image = "data:image/jpeg;base64," + res.data.image_bytes;
        this.getChats()
      } catch (error) {
        console.log(error)
        if (error.response && error.response.status == 401) {
          this.logout()
        }
      }
    }, 
    
    async getChats() {
      let res = await axios.get(`api/chats`, {
        token: localStorage.getItem("token")
      })
      console.log("getChats: api/chats", res)
      this.chats = []
      if (res.data != null) {
        for (const chat of res.data) {
          console.log(`Chat ID: ${chat.chat_id}}`);
          let data = {
            chat_id: chat.chat_id,
            people: chat.people,
            lastMess: "Hello Armin",
            lastMessDate: "Just Now",
            last_seen: "",
            image: null,
            username: "",
          }
          let ext_data = await this.getExtension(chat.chat_id)
          console.log(ext_data)
          data.lastMessDate = await this.formatDate(ext_data.last_msg_date.slice(0, -1))
          if (chat.people[0] == this.id) {
            await this.getOtherUsersInfo(chat.people[1])
            data.image = this.contactsInfo[chat.people[1]].image
            data.username = this.contactsInfo[chat.people[1]].username
            this.contactsInfo[chat.people[1]].last_seen = await this.formatDate(ext_data.laston1) 
          } else {
            await this.getOtherUsersInfo(chat.people[0])
            data.image = this.contactsInfo[chat.people[0]].image
            data.username = this.contactsInfo[chat.people[0]].username
            this.contactsInfo[chat.people[0]].last_seen = await this.formatDate(ext_data.laston2) 
          }
          data.notviewed = ext_data.notviewed
          data.lastMess = ext_data.last_message
          data.lastUser = ext_data.last_user
          this.chats.push(data)           
        }
      }
    },
    async getExtension(chat_id) {
      let res = await axios.get(`api/extensions/${chat_id}`, {
        token: localStorage.getItem("token")
      })
      console.log(`getExtensions: api/extensions/${chat_id}`, res)
      console.log(res.data)
      if (res.data != null) {
          let data = res.data
          let ext = {
            notviewed: data.notviewed,
            last_user: data.last_user,
            last_message: data.last_message,
            last_msg_date: data.last_msg_date,
            laston1: data.laston1,
            laston2: data.laston2,
          }
          // if (data.last_user == this.id) {
          //   ext.notviewed = 0
          // }
          return ext
      }
      throw new Error("extension for the chat is null!")
    },
    
    async setID(id) {
      this.id = id 
    },
    
    async getOtherUsersInfo(user_id) {
      if (this.contactsInfo[parseInt(user_id)]) {
        return
      }
      try {
        let res = await axios.get(`api/users?keyword=${user_id}`, {
          token: localStorage.getItem("token")
        })
        let data = {
          firstname: res.data.firstname,
          lastname: res.data.lastname,
          username: res.data.username,
          phone: res.data.phone,
          bio: res.data.bio,
          image: "data:image/jpeg;base64," + res.data.image_bytes
        }
        console.log("getOtherUsersInfo: ", data)
        this.contactsInfo[parseInt(user_id)] = data
        console.log(res)
        
      } catch (error) {
        console.log(error)
      }
    },
    
    async updateUserInfo(phone, bio, username, firstname, lastname, imageFile, image) {
      const formData = new FormData();
      formData.append('firstname', firstname);
      formData.append('lastname', lastname);
      formData.append('phone', phone);
      formData.append('username', username);
      formData.append('Bio', bio);
      formData.append('image', imageFile);
      // console.log("updateUserInfo: ", imageFile)
      let res = -1
      await axios.patch(`api/users/${this.id}`, formData)
      .then(response => {
        res = 1
        this.firstname = firstname
        this.lastname = lastname
        this.phone = phone
        this.username = username
        this.bio = bio
        this.image = image
        console.log("updateUserInfo: ", response.status)
      })
      .catch(error => {
        console.log("updateUserInfo: ", error)
        if (error.response && error.response.status == 401) {
          this.logout()
        }
        res = -1
      });
      return res
    },
    
    async getAllContacts() {
      try {
        let res = await axios.get(`api/users/${this.id}/contacts`, {
          token: localStorage.getItem("token")
        })
        console.log(res)
        this.contacts = []
        if (res.data != null) {
          for (const contact of res.data) {
            let data = {
              id: contact.contact_id,
              name: contact.contact_name
            }
            await this.getOtherUsersInfo(contact.contact_id)
            console.log(`GetAllContacts: ${data.id}  ${data.name}`)
            this.contacts.push(data)
          }
        }
      } catch (error) {
        console.log(error)
        if (error.response && error.response.status == 401) {
          this.logout()
        }
      }
    },
    async addNewContact(contact_id, contact_name) {
      try {
        let res = await axios.post(`api/users/${this.id}/contacts`, {
          user_id: this.id,
          contact_id: contact_id,
          contact_name: contact_name,
          token: localStorage.getItem("token")
        })
        console.log(res)
        if (res.data != null) {
          let data = {
            id: contact_id,
            name: contact_name
          }
          await this.getOtherUsersInfo(contact_id)
          console.log(data)
          this.contacts.push(data)
        }
        return false
      } catch (error) {
        console.log(error)
        if (error.response && error.response.status == 401) {
          this.logout()
        }
        return true
      }
    },
    
    async createNewChat(people, name) {
      try {
        console.log("createNewChat()=====================")
        let res = await axios.post(`api/chats`, {
          people: people,
          token: localStorage.getItem("token")
        })
        console.log("createNewChat: POST /api/{user_id}/chat: ", res.data)
        let data = {
          chat_id: res.data,
          people: people,
          lastMess: "Hello Armin",
          lastMessDate: "Just Now",
          image: null,
          username: "",
        }
        let ext_data = await this.getExtension(res.data)
        data.lastMessDate = await this.formatDate(ext_data.last_msg_date.slice(0, -1))
        if (people[0] == this.id) {
          data.image = this.contactsInfo[people[1]].image
          data.username = this.contactsInfo[people[1]].username
          this.contactsInfo[data.people[1]].last_seen = await this.formatDate(ext_data.laston1)
        } else {
          data.image = this.contactsInfo[people[0]].image
          data.username = this.contactsInfo[people[0]].username
          this.contactsInfo[data.people[0]].last_seen = await this.formatDate(ext_data.laston2) 
        }
        data.notviewed = ext_data.notviewed
        data.lastMess = ext_data.last_message
        data.lastUser = ext_data.last_user
        console.log("createNewChat", ext_data)
        this.chats.push(data)           
        this.userDataForChat = {
          chat_id: res.data,
          id1: people[0],
          id2: people[1],
          firstname: name,
          image: null,
        }
        if (people[0] == this.id) {
          this.userDataForChat.image = this.contactsInfo[parseInt(people[1])].image 
        } else {
          this.userDataForChat.image = this.contactsInfo[parseInt(people[0])].image 
        }
      } catch (error) {
        console.log(error)
        if (error.response && error.response.status == 401) {
          this.logout()
        }
      }
    },
    
    async deleteChat(chat_id) {
      try {
        let res = await axios.delete(`api/chats/${chat_id}`, {
          token: localStorage.getItem("token")
        })
        console.log(`deleteChat: DELETE api/chats/${chat_id}`)
        this.chats = this.chats.filter(chat => chat.chat_id !== chat_id);
        if (this.userDataForChat && this.userDataForChat.chat_id == chat_id) {
          this.userDataForChat = null
          this.currentChat = null
        }
      } catch (error) {
        console.log("deleteChat: ", error)
        if (error.response && error.response.status == 401) {
          this.logout()
        }
      }
    },
    
    async deleteContact(contact_id) {
      try {
        let res = await axios.delete(`api/users/${this.id}/contacts/${contact_id}`, {
          token: localStorage.getItem("token")
        })
        console.log(`deleteChat: DELETE api/users/${this.id}/contacts/${contact_id}`)
        this.contacts = this.contacts.filter(contact => contact.id !== contact_id);
        this.chats = this.chats.filter(chat => chat.id2 !== contact_id || chat.id1 !== contact_id);
      } catch (error) {
        console.log("deleteChat: ", error)
        if (error.response && error.response.status == 401) {
          this.logout()
        }
      }
    },
    
    async deleteMessage(chat_id) {
      try {
        let res = await axios.delete(`api/chats/${chat_id}`, {
          token: localStorage.getItem("token")
        })
        console.log(`deleteChat: DELETE api/chats/${chat_id}`)
        this.chats = this.chats.filter(chat => chat.chat_id !== chat_id);
        if (this.userDataForChat && this.userDataForChat.chat_id == chat_id) {
          this.userDataForChat = null
          this.currentChat = null
        }
      } catch (error) {
        console.log("deleteChat: ", error)
        if (error.response && error.response.status == 401) {
          this.logout()
        }
      }
    },
    async formatDate(dateString) {
      console.log(dateString)
      const date = new Date(dateString);
      const currentDate = new Date();
      // Check if the date is today
      if (date.toDateString() === currentDate.toDateString()) {
          let res = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
          console.log(res)
          return res
      }
      // Check if the date is this week
      const oneDay = 24 * 60 * 60 * 1000; // hours*minutes*seconds*milliseconds
      const diffDays = Math.round(Math.abs((currentDate - date) / oneDay));
      if (diffDays < 7 && date.getDay() < currentDate.getDay()) {
          let res = date.toLocaleDateString([], { weekday: 'long' });
          console.log(res)
          return res
      }
      // Check if the date is this year
      if (date.getFullYear() === currentDate.getFullYear()) {
          return date.toLocaleDateString([], { month: '2-digit', day: '2-digit' });
          console.log(res)
          return res
      }
      // Otherwise, return "long time ago"
      return "long time ago";
    },

    removeAllData() {
      this.id = -1
      this.image = null
      this.firstname = ''
      this.lastname = ''
      this.username = ''
      this.phone = ''
      this.bio = ''
      this.isUserInfoVisible = false
      this.chats = []
      this.contactsInfo= {}
      this.contacts = []
      this.userDataForChat = null
      this.showFindFriends = false
      this.isNewContactModalVisible = false
      this.isPartnerInfoVisible = false
      this.currentChat = null
      this.removeUsersFromFindFriends = []

    },   
    
    logout() {
      this.removeAllData()
      localStorage.removeItem('token')
    }
  },
  persist: true
})
