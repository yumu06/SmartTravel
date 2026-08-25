// pages/user/user.js
// const app= getApp();

// Page({


//   /**
//    * 生命周期函数--监听页面加载
//    */
//   onLoad() {
//     // //获取全局变量
//     // const userInfo = app.globalData.userInfo;
//     // //更新页面数据
//     // this.setData({
//     //   userInfo: userInfo
//     // });

//     //----------改
//     // 检查全局登录状态
//     const isLoggedIn = app.checkLogin();
//     this.setData({
//       isLoggedIn,
//     });

//     // 如果已登录，加载用户信息
//     if (isLoggedIn) {
//       //获取全局变量
//       const userInfo = app.globalData.userInfo;
//       //更新页面数据
//       this.setData({
//         userInfo: userInfo
//       });
//       //---实际应用如下
//       //this.loadUserInfo();
//     }
//   },

//   // 加载用户信息
//   loadUserInfo() {
//     const token = app.globalData.token;
//     wx.request({
//       url: 'https://<your-domain>/GetUserInfo', // 替换为你的用户信息接口
//       method: 'GET',
//       header: {
//         'Authorization': `Bearer ${token}`,
//       },
//       success: res => {
//         if (res.data && res.data.userInfo) {
//           this.setData({
//             userInfo: res.data.userInfo,
//           });
//         } else {
//           wx.showToast({
//             title: '获取用户信息失败',
//             icon: 'none',
//             duration: 2000,
//           });
//         }
//       },
//       fail: err => {
//         console.error('获取用户信息失败', err);
//         wx.showToast({
//           title: '获取用户信息失败，请重试',
//           icon: 'none',
//           duration: 2000,
//         });
//       },
//     });
//   },


// })

const app= getApp();
const util = require('../../common/utils/util')
Page({
  /**
   * 页面的初始数据
   */
  data: {
    isLoggedIn: false, // 是否已登录
    userInfo: {}
  },

  onTap() {
    wx.navigateTo({
      url: '../../packages/user/pages/fixMessage/fixMessage',
    });
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad() {
    // 检查全局登录状态
    const isLoggedIn = app.checkLogin();
    this.setData({
      isLoggedIn,
    });
    this.setData({ userInfo: app.globalData.userInfo || {} })

    // 如果已登录，加载用户信息
    if (isLoggedIn) {
      this.loadUserInfo()
    }
  },

  handleLogin() {
    wx.login({
      success: res => {
        if (!res.code) {
          wx.showToast({ title: '登录失败', icon: 'none' })
          return
        }

        wx.request({
          url: `${util.API_BASE_URL}/travel/login`,
          method: 'POST',
          header: {
            'Content-Type': 'application/x-www-form-urlencoded'
          },
          data: {
            code: res.code
          },
          success: loginRes => {
            if (loginRes.data && loginRes.data.token) {
              const token = loginRes.data.token
              wx.setStorageSync('token', token)
              app.globalData.isLogin = true
              app.globalData.token = token
              this.setData({ isLoggedIn: true })
              this.loadUserInfo()
              wx.showToast({ title: '登录成功', icon: 'success', duration: 2000 })
              return
            }
            wx.showToast({ title: '登录失败', icon: 'none' })
          },
          fail: () => {
            wx.showToast({ title: '登录失败', icon: 'none' })
          }
        })
      },
      fail: () => {
        wx.showToast({ title: '登录失败', icon: 'none' })
      }
    })
  },

  loadUserInfo() {
    this.setData({ userInfo: app.globalData.userInfo || {} })
    util.request({ url: '/travel/user/info', method: 'GET' })
      .then(res => {
        const info = (res && res.data && res.data.information) || {}
        const userInfo = {
          avatar: app.globalData.userInfo.avatar,
          nickname: info.nick_name || info.NickName || app.globalData.userInfo.nickname,
          motto: info.motto || info.Motto || '',
          gender: info.gender != null ? info.gender : '',
          telephone: info.telephone || info.Telephone || '',
        }
        app.globalData.userInfo = Object.assign({}, app.globalData.userInfo, userInfo)
        this.setData({ userInfo: userInfo })
      })
      .catch(() => {
        this.setData({ userInfo: app.globalData.userInfo || {} })
        wx.showToast({ title: '获取用户信息失败', icon: 'none' })
      })
  },

  logout() {
    util.clearToken()
    app.globalData.isLogin = false
    app.globalData.token = null
    this.setData({ isLoggedIn: false, userInfo: {} })
    wx.showToast({ title: '已退出', icon: 'success' })
  },

  onReady() {
  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    const isLoggedIn = app.checkLogin()
    this.setData({ isLoggedIn })
    this.setData({ userInfo: app.globalData.userInfo || {} })
    if (isLoggedIn) {
      this.loadUserInfo()
    }
  }
});
