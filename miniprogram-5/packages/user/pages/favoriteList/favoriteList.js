const util = require('../../../../common/utils/util')

Page({

  /**
   * 页面的初始数据
   */
  data: {
    type: 'post',
    tabs: [
      { key: 'post', label: '文章收藏' },
      { key: 'foot', label: '足迹收藏' },
      { key: 'history', label: '历史足迹' },
    ],
    favoriteArticles: [],
    foots: [],
    total: 0,
    pageNum: 1,
    pageSize: 20,
    isLoading: false,
    hasMore: true,
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {
    const type = options.type || 'post'
    if (type === 'history') {
      this.setData({ type, tabs: [{ key: 'history', label: '个人足迹' }] })
      wx.setNavigationBarTitle({ title: '个人足迹' })
    } else {
      this.setData({
        type,
        tabs: [
          { key: 'post', label: '文章收藏' },
          { key: 'foot', label: '足迹收藏' },
        ],
      })
      wx.setNavigationBarTitle({ title: '收藏' })
    }
    this.refresh(true)
  },

  onShow() {
    const flag = wx.getStorageSync('needRefreshFavorites')
    if (flag) {
      wx.removeStorageSync('needRefreshFavorites')
      this.refresh(true)
      return
    }
    this.refresh(true)
  },

  onTabTap(e) {
    const type = e.currentTarget.dataset.type
    if (!type || type === this.data.type) return
    this.setData({ type })
    this.refresh(true)
  },

  normalizeFoots(list) {
    if (!Array.isArray(list)) return []
    return list.map(x => {
      var destText = ''
      try {
        var arr = JSON.parse(x.Destinations || '[]')
        if (Array.isArray(arr) && arr.length) {
          destText = String(arr[arr.length - 1] || '')
        }
      } catch (e) {}
      var destName = ''
      try {
        var names = JSON.parse(x.DestinationNames || '[]')
        if (Array.isArray(names) && names.length) {
          destName = String(names[names.length - 1] || '')
        }
      } catch (e) {}
      return Object.assign({}, x, {
        DestinationText: destText,
        OriginDisplay: x.OriginName || x.Origin || '',
        DestinationDisplay: destName || destText || '',
      })
    })
  },

  normalizeArticles(list) {
    if (!Array.isArray(list)) return []
    return list.map(x => {
      const content = String(x.Content || '')
      const snippetBase = content.replace(/\s+/g, '').slice(0, 60)
      const snippet = snippetBase.length < content.length ? `${snippetBase}...` : snippetBase
      return Object.assign({}, x, { Snippet: snippet })
    })
  },

  refresh(reset) {
    if (this.data.isLoading) return
    if (!this.data.hasMore && !reset) return

    const type = this.data.type
    const nextPage = reset ? 1 : this.data.pageNum
    this.setData({ isLoading: true })

    if (type === 'post') {
      util.request({ url: '/travel/user/start/list', method: 'GET' })
        .then(res => {
          const list = this.normalizeArticles((res && res.data && res.data.data) || [])
          this.setData({
            favoriteArticles: list,
            isLoading: false,
            hasMore: false,
          })
          wx.stopPullDownRefresh()
        })
        .catch(() => {
          this.setData({ isLoading: false })
          wx.stopPullDownRefresh()
        })
      return
    }

    if (type === 'foot') {
      util.request({ url: '/travel/foot/start/list', method: 'GET' })
        .then(res => {
          const list = this.normalizeFoots((res && res.data && res.data.data) || [])
          this.setData({
            foots: list,
            isLoading: false,
            hasMore: false,
          })
          wx.stopPullDownRefresh()
        })
        .catch(() => {
          this.setData({ isLoading: false })
          wx.stopPullDownRefresh()
        })
      return
    }

    util.request({
      url: '/travel/foot/list',
      method: 'GET',
      data: { pageNum: nextPage, pageSize: this.data.pageSize },
    })
      .then(res => {
        const list = this.normalizeFoots((res && res.data && res.data.data) || [])
        const total = Number((res && res.data && res.data.total) || 0)
        const merged = reset ? list : this.data.foots.concat(list)
        this.setData({
          foots: merged,
          total: total,
          hasMore: merged.length < total,
          pageNum: nextPage + 1,
          isLoading: false,
        })
        wx.stopPullDownRefresh()
      })
      .catch(() => {
        this.setData({ isLoading: false })
        wx.stopPullDownRefresh()
      })
  },

  toDetail(e) {
    const articleId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `/packages/home/pages/articleDetail/articleDetail?id=${articleId}`
    });
  },

  toFootDetail(e) {
    const idx = Number(e.currentTarget.dataset.index || 0)
    const foot = (this.data.foots || [])[idx]
    if (!foot) return
    const footID = foot.ID || foot.Id || foot.id
    if (!footID) return

    const type = this.data.type
    const that = this
    const viewRoute = () => {
      wx.setStorageSync('openFootID', String(footID))
      wx.switchTab({ url: '/pages/map/map' })
    }

    if (type === 'foot') {
      wx.showActionSheet({
        itemList: ['查看路线', '取消收藏'],
        success(res) {
          if (res.tapIndex === 0) {
            viewRoute()
            return
          }
          if (res.tapIndex === 1) {
            util.request({ url: `/travel/foot/start/remove/${footID}`, method: 'DELETE' })
              .then(() => {
                wx.showToast({ title: '已取消收藏', icon: 'success' })
                that.refresh(true)
              })
              .catch(() => wx.showToast({ title: '操作失败', icon: 'none' }))
          }
        }
      })
      return
    }

    if (type === 'history') {
      wx.showActionSheet({
        itemList: ['查看路线', '收藏足迹', '删除足迹'],
        success(res) {
          if (res.tapIndex === 0) {
            viewRoute()
            return
          }
          if (res.tapIndex === 1) {
            util.request({ url: `/travel/foot/start/add/${footID}`, method: 'POST' })
              .then(() => {
                wx.showToast({ title: '收藏成功', icon: 'success' })
              })
              .catch(() => wx.showToast({ title: '操作失败', icon: 'none' }))
            return
          }
          if (res.tapIndex === 2) {
            wx.showModal({
              title: '删除足迹',
              content: '确定删除这条足迹吗？删除后不可恢复',
              confirmText: '删除',
              confirmColor: '#ef4444',
              success(r) {
                if (!r.confirm) return
                util.request({ url: `/travel/foot/delete/${footID}`, method: 'DELETE' })
                  .then(() => {
                    wx.showToast({ title: '已删除', icon: 'success' })
                    that.refresh(true)
                  })
                  .catch(() => wx.showToast({ title: '删除失败', icon: 'none' }))
              }
            })
          }
        }
      })
      return
    }

    viewRoute()
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
    this.refresh(true)
  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom() {
    this.refresh(false)
  },

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {

  }
})
