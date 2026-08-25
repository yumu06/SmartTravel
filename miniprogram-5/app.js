App({
  onLaunch() {
    this.restoreLoginState()
  },

  restoreLoginState() {
    const token = wx.getStorageSync('token') || ''
    this.globalData.token = token
    this.globalData.isLogin = Boolean(token)
  },

  checkLogin() {
    this.restoreLoginState()
    return this.globalData.isLogin
  },

  globalData: {
    isLogin: false,
    token: '',
    userInfo: {},
  },
})
