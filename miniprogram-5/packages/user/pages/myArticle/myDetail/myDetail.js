const util = require('../../../../../common/utils/util')

Page({

  /**
   * 页面的初始数据
   */
  data: {
    articleId: "",
    article: null,
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    const articleId = options.id || ""
    this.setData({ articleId })
    this.getArticleDetail()
  },

  previewImage(e) {
    const imageUrl = e.currentTarget.dataset.src;
    wx.previewImage({
      current: imageUrl, // 当前显示图片的http链接
      urls: [imageUrl]
    });
  },

  getArticleDetail() {
    const articleId = this.data.articleId
    if (!articleId) return
    util.request({ url: `/travel/post/show/${articleId}`, method: 'GET' })
      .then(res => {
        const post = res && res.data && res.data.post
        this.setData({ article: post || null })
      })
      .catch(() => {
        wx.showToast({ title: '获取文章失败', icon: 'none' })
      })
  },

  showOptions() {
    const that = this;
    wx.showActionSheet({
      itemList: ['编辑', '删除'],
      success(res) {
        if (res.tapIndex === 0) {
          wx.navigateTo({
            url: `./editArcitle/editArticle?id=${that.data.articleId}`,
          })
        } else if (res.tapIndex === 1) {
          that.deleteArticle();
          
        }
      },
      fail(res) {
        console.log("取消选择");
      }
    })
  },

  deleteArticle() {
    // 显示确认对话框
    const articleId = this.data.articleId
    if (!articleId) return
    wx.showModal({
      title: '确认删除',
      content: '确定要删除这篇文章吗？',
      success: (res) => {
        if (res.confirm) {
          util.request({ url: `/travel/post/delete/${articleId}`, method: 'DELETE' })
            .then(() => {
              wx.showToast({ title: '删除成功', icon: 'success' })
              wx.navigateBack()
            })
            .catch(() => {
              wx.showToast({ title: '删除失败', icon: 'none' })
            })
        }
      }
    });
  },

  
})
