// common/components/login.js
const { API_BASE_URL } = require('../../config/index')

Component({

  /**
   * 组件的属性列表
   */
  properties: {

  },

  options:{
    addGlobalClass: true, // 允许使用全局样式
  },

  /**
   * 组件的初始数据
   */
  data: {

  },

  /**
   * 组件的方法列表
   */
  methods: {
    //登录
    handleLogin() {
      wx.login({
        success: res => {
          if (!res.code) {
            wx.showToast({
              title: '登录失败，请重试',
              icon: 'none',
              duration: 2000,
            })
            return
          }

          wx.request({
            url: `${API_BASE_URL}/travel/login`,
            method: 'POST',
            header: {
              'Content-Type': 'application/x-www-form-urlencoded',
            },
            data: {
              code: res.code,
            },
            success: loginRes => {
              if (loginRes.data && loginRes.data.token) {
                const token = loginRes.data.token
                wx.setStorageSync('token', token)
                this.triggerEvent('loginsuccess', { token })
                return
              }
              wx.showToast({
                title: '登录失败，请重试',
                icon: 'none',
                duration: 2000,
              })
            },
            fail: () => {
              wx.showToast({
                title: '登录失败，请检查网络',
                icon: 'none',
                duration: 2000,
              })
            },
          })
        },
        fail: () => {
          wx.showToast({
            title: '登录失败，请重试',
            icon: 'none',
            duration: 2000,
          })
        },
      })
    },

  }
})
