const util = require('../../../../common/utils/util')

Page({

  /**
   * 页面的初始数据
   */
  data: {
    articleId: "",
    article: null,
    comments: [],
    totalComments: 0,
    commentContent: "",
    userID: 0,
    isCollected: false,
    isLiked: false,

  },
  

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    const articleId = options.id || "";
    this.setData({ articleId: articleId });
    this.loadPage();
  },

  loadPage() {
    const that = this
    util.ensureToken()
      .then(() => util.request({ url: '/travel/authorization', method: 'GET' }))
      .then(res => {
        const uid = res && res.data && res.data.data && res.data.data.userID
        that.setData({ userID: Number(uid || 0) })
      })
      .catch(() => {})
      .finally(() => {
        that.getArticleDetail()
        that.getComments()
        that.refreshCollectedState()
      })
  },

  getArticleDetail() {
    const articleId = this.data.articleId
    if (!articleId) return
    util.request({
      url: `/travel/post/show/${articleId}`,
      method: 'GET',
    })
      .then(res => {
        const post = res && res.data && res.data.post
        this.setData({
          article: post || null,
        })
      })
      .catch(() => {
        wx.showToast({
          title: '获取文章详情失败',
          icon: 'none'
        })
      })
  },

  refreshCollectedState() {
    const articleId = this.data.articleId
    if (!articleId) return
    util.request({
      url: '/travel/user/start/list',
      method: 'GET',
    })
      .then(res => {
        const list = (res && res.data && res.data.data) || []
        const hit = list.some(x => String(x.ID || x.id || '') === String(articleId))
        this.setData({ isCollected: hit })
      })
      .catch(() => {})
  },

  getComments() {
    const articleId = this.data.articleId
    if (!articleId) return
    util.request({
      url: `/travel/post/${articleId}/comment/list`,
      method: 'GET',
      data: { pageNum: 1, pageSize: 50 },
    })
      .then(res => {
        this.setData({
          comments: (res && res.data && res.data.data) || [],
          totalComments: Number((res && res.data && res.data.total) || 0),
        })
      })
      .catch(() => {})
  },

  toggleCollect() {
    const articleId = this.data.articleId
    if (!articleId) return
    const isCollected = this.data.isCollected
    const method = isCollected ? 'DELETE' : 'POST'
    const url = isCollected ? `/travel/user/start/remove/${articleId}` : `/travel/user/start/add/${articleId}`
    util.request({ url, method })
      .then((res) => {
        if (!res || res.statusCode !== 200) {
          const msg = (res && res.data && (res.data.error || res.data.msg || res.data.message)) || '操作失败'
          throw new Error(msg)
        }
        this.setData({ isCollected: !isCollected })
        wx.setStorageSync('needRefreshFavorites', '1')
        wx.showToast({
          title: isCollected ? '取消收藏成功' : '收藏成功',
          icon: 'success'
        })
      })
      .catch(() => {
        wx.showToast({
          title: '操作失败',
          icon: 'none'
        })
      })
  },

  toggleLike() {
    const articleId = this.data.articleId
    if (!articleId) return
    const isLiked = this.data.isLiked
    const method = isLiked ? 'DELETE' : 'POST'
    util.request({ url: `/travel/post/like/${articleId}`, method })
      .then((res) => {
        if (!res || res.statusCode !== 200) {
          const msg = (res && res.data && (res.data.msg || res.data.message)) || '操作失败'
          throw new Error(msg)
        }
        const article = this.data.article || {}
        const next = !isLiked
        const likeCount = Number(article.LikeCount || 0) + (next ? 1 : -1)
        this.setData({ isLiked: next, article: Object.assign({}, article, { LikeCount: likeCount < 0 ? 0 : likeCount }) })
        wx.showToast({ title: next ? '点赞成功' : '取消点赞成功', icon: 'success' })
      })
      .catch((err) => {
        const msg = (err && err.message) || '操作失败'
        if (!isLiked && msg.indexOf('已点赞') !== -1) {
          this.setData({ isLiked: true })
          return
        }
        if (isLiked && msg.indexOf('未点赞') !== -1) {
          this.setData({ isLiked: false })
          return
        }
        wx.showToast({ title: msg, icon: 'none' })
      })
  },

  bindCommentInput(e) {
    this.setData({ commentContent: e.detail.value })
  },

  submitComment() {
    const articleId = this.data.articleId
    const content = (this.data.commentContent || '').trim()
    if (!articleId) return
    if (!content) {
      wx.showToast({ title: '请输入评论内容', icon: 'none' })
      return
    }
    util.request({
      url: `/travel/post/${articleId}/comment`,
      method: 'POST',
      data: { content: content },
    })
      .then(() => {
        this.setData({ commentContent: '' })
        this.getComments()
        wx.showToast({ title: '评论成功', icon: 'success' })
      })
      .catch(res => {
        wx.showToast({ title: (res && res.data && res.data.msg) || '评论失败', icon: 'none' })
      })
  },

  deleteComment(e) {
    const commentId = e.currentTarget.dataset.id
    if (!commentId) return
    util.request({
      url: `/travel/post/comment/${commentId}`,
      method: 'DELETE',
    })
      .then(() => {
        this.getComments()
        wx.showToast({ title: '删除成功', icon: 'success' })
      })
      .catch(() => {
        wx.showToast({ title: '删除失败', icon: 'none' })
      })
  },

  toAuthorPosts(e) {
    const userID = e.currentTarget.dataset.userid
    if (!userID) return
    wx.navigateTo({ url: `/packages/home/pages/share/share?userID=${encodeURIComponent(userID)}` })
  },

  previewImage(e) {
    const imageUrl = e.currentTarget.dataset.src
    if (!imageUrl) return
    wx.previewImage({
      current: imageUrl,
      urls: [imageUrl]
    })
  },


})
