const util = require('../../../../common/utils/util')

Page({

  /**
   * 页面的初始数据
   */
  data: {
    posts: [],
    keyword: "",
    userID: "",
    pageNum: 1,
    pageSize: 20,
    total: 0,
    isLoading: false,
    hasMore: true,
  },

  normalizePosts(list) {
    if (!Array.isArray(list)) return []
    return list.map(item => {
      const id = item && (item.ID || item.Id || item.id)
      const content = (item && (item.Content || item.content)) || ''
      const snippetBase = content.replace(/\s+/g, '').slice(0, 40)
      const snippet = snippetBase.length < content.length ? `${snippetBase}...` : snippetBase
      const coverUrl = (item && (item.HeadImg || item.head_img || '')) || ''
      const createdAt = item && (item.CreatedAt || item.created_at || '')
      const userID = item && (item.UserID || item.user_id || '')
      const title = item && (item.Title || item.title || '')
      return {
        ID: id,
        Title: title,
        Content: content,
        HeadImg: coverUrl,
        CreatedAt: createdAt,
        UserID: userID,
        snippet: snippet,
        coverUrl: coverUrl,
      }
    })
  },

  fetchPosts(reset) {
    if (this.data.isLoading) return
    if (!this.data.hasMore && !reset) return

    const nextPage = reset ? 1 : this.data.pageNum
    this.setData({ isLoading: true })

    const keyword = (this.data.keyword || '').trim()
    const userID = (this.data.userID || '').trim()
    const req = keyword
      ? util.request({
          url: '/travel/post/search',
          method: 'GET',
          data: { keyword: keyword, pageNum: nextPage, pageSize: this.data.pageSize },
        })
      : userID
        ? util.request({
            url: '/travel/post/user',
            method: 'GET',
            data: { userID: userID, pageNum: nextPage, pageSize: this.data.pageSize },
          })
        : util.request({
            url: '/travel/post/page/list',
            method: 'GET',
            data: { pageNum: nextPage, pageList: this.data.pageSize },
          })

    req
      .then(res => {
        const list = keyword || userID
          ? ((res && res.data && res.data.data) || [])
          : (((res && res.data && res.data.date) && res.data.date.commodities) || [])
        const total = keyword || userID
          ? Number((res && res.data && res.data.total) || 0)
          : Number((((res && res.data && res.data.date) && res.data.date.total) || 0))
        const normalized = this.normalizePosts(list)
        const merged = reset ? normalized : this.data.posts.concat(normalized)
        const hasMore = merged.length < total
        this.setData({
          posts: merged,
          total: total,
          hasMore: hasMore,
          pageNum: nextPage + 1,
        })
      })
      .catch(() => {
        wx.showToast({ title: '获取文章列表失败', icon: 'none' })
      })
      .finally(() => {
        this.setData({ isLoading: false })
        wx.stopPullDownRefresh()
      })
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    const keyword = options.keyword ? decodeURIComponent(options.keyword) : ""
    const userID = options.userID ? decodeURIComponent(options.userID) : ""
    if (keyword) {
      wx.setNavigationBarTitle({ title: `搜索：${keyword}` })
    }
    if (userID) {
      wx.setNavigationBarTitle({ title: `用户 ${userID} 的文章` })
    }
    this.setData({ keyword, userID })
    this.fetchPosts(true)
  },

  toDetail(e) {
    const articleId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/packages/home/pages/articleDetail/articleDetail?id=${articleId}`
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
    this.fetchPosts(true)
  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom() {
    this.fetchPosts(false)
  },

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {

  }
})
