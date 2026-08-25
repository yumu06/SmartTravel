const util = require('../../../../common/utils/util')

Page({

  /**
   * 页面的初始数据
   */
  data: {
    myArticles:[],
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    wx.setNavigationBarTitle({ title: '我的文章' })
    this.getMyArticles();
  },

  getMyArticles: function(){
    util.request({
      url: '/travel/user/postCreate',
      method: 'GET',
    })
      .then(res => {
        const list = (res && res.data && res.data.data) || []
        const normalized = Array.isArray(list) ? list.map(x => {
          const content = String(x.Content || '')
          const snippetBase = content.replace(/\s+/g, '').slice(0, 60)
          const snippet = snippetBase.length < content.length ? `${snippetBase}...` : snippetBase
          return Object.assign({}, x, { Snippet: snippet })
        }) : []
        this.setData({ myArticles: normalized })
      })
      .catch(() => {
        wx.showToast({ title: '获取我的文章失败', icon: 'none' })
      })
  },

  toDetail(e) {
    const articleId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/packages/user/pages/myArticle/myDetail/myDetail?id=${articleId}`
    });
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
    this.getMyArticles()
  },

  /**
   * 生命周期函数--监听页面隐藏
   */
  onHide() {

  },

  /**
   * 生命周期函数--监听页面卸载
   */
  onUnload() {

  },

  /**
   * 页面相关事件处理函数--监听用户下拉动作
   */
  onPullDownRefresh() {
    this.getMyArticles()
    wx.stopPullDownRefresh()
  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom() {

  },

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {

  }
})
