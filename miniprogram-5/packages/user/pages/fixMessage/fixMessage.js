// pages/user/fixMessage/fixMessage.js
const app= getApp();
const util = require('../../../../common/utils/util')
Page({

  /**
   * 页面的初始数据
   */
  data: {
    userInfo:{

    }
  },

  onChooseAvatar(e) {
    const { avatarUrl } = e.detail
    this.setData({
      "userInfo.avatar": avatarUrl,
    })
  },

  onInputChange(e) {
    const nickName = e.detail.value
    this.setData({
      "userInfo.nickname": nickName,
    })
  },

  onInputChangeMotto(e) {
    const motto = e.detail.value
    this.setData({
      "userInfo.motto": motto,
    })
  },

  onInputChangeGender(e) {
    const gender = e.detail.value
    this.setData({
      "userInfo.gender": gender,
    })
  },

  onInputChangeTelephone(e) {
    const telephone = e.detail.value
    this.setData({
      "userInfo.telephone": telephone,
    })
  },

  updateUserInfo: function() {
    const { userInfo } = this.data;
    if (!Object.values(userInfo).some(value => value)) {
      wx.showToast({
        title: '请至少填写一项信息',
        icon: 'none'
      });
      return;
    }

    app.globalData.userInfo = {
      ...app.globalData.userInfo,
      ...userInfo
    };

    const payload = {
      telephone: userInfo.telephone || '',
      nick_name: userInfo.nickname || '',
      motto: userInfo.motto || '',
      gender: Number(userInfo.gender || 0),
    }

    util.request({
      url: '/travel/user/update',
      method: 'PATCH',
      data: payload,
      header: { 'Content-Type': 'application/json' },
    })
      .then(() => {
        wx.showToast({ title: '更新成功', icon: 'success' })
        wx.navigateBack()
      })
      .catch(() => {
        wx.showToast({ title: '更新失败', icon: 'none' })
      })
  },
  

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad() {
    //获取全局变量
    const userInfo = app.globalData.userInfo;
    //更新页面数据
    this.setData({
      userInfo: userInfo
    })
  },

  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady() {

  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    const userInfo= app.globalData.userInfo;
    this.setData({
      userInfo: userInfo
    })
  },


})
