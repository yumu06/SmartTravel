const util = require('../../../../../../common/utils/util')

Page({

  /**
   * 页面的初始数据
   */
  data: {
    articleId: "",
    article: null,
  },

  saveArticle:function () {
    const articleId = this.data.articleId
    const article = this.data.article || {}
    if (!articleId) return
    if (!article.Title || !article.Content) {
      wx.showToast({ title: '请填写标题和内容', icon: 'none' })
      return
    }
    util.request({
      url: `/travel/post/update/${articleId}`,
      method: 'PATCH',
      data: {
        title: article.Title,
        content: article.Content,
        head_img: article.HeadImg || '',
      },
      header: { 'Content-Type': 'application/json' },
    })
      .then(() => {
        wx.showToast({ title: '保存成功', icon: 'success' })
        wx.navigateBack()
      })
      .catch(() => {
        wx.showToast({ title: '保存失败', icon: 'none' })
      })
  },

  onTitleInput: function(e) {
    if (!this.data.article) {
      this.setData({ article: { Title: '', Content: '', HeadImg: '' } })
    }
    this.setData({ 'article.Title': e.detail.value });
  },
  
  onContentInput: function(e) {
    if (!this.data.article) {
      this.setData({ article: { Title: '', Content: '', HeadImg: '' } })
    }
    this.setData({ 'article.Content': e.detail.value });
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    const articleId = options.id || ""
    this.setData({ articleId })
    if (!articleId) return
    util.request({ url: `/travel/post/show/${articleId}`, method: 'GET' })
      .then(res => {
        const post = res && res.data && res.data.post
        this.setData({ article: post || { Title: '', Content: '', HeadImg: '' } })
      })
      .catch(() => {
        wx.showToast({ title: '加载失败', icon: 'none' })
        this.setData({ article: { Title: '', Content: '', HeadImg: '' } })
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
