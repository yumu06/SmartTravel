const util = require('../../common/utils/util')

Page({
	/**
	 * 页面的初始数据
	 */
	data: {
			cover:'',
			viewHeight: 640, // 默认高度
			navIndex: 0,
			searchTypes: ['文章', '用户'],
			searchTypeIndex: 0,
			keyword: '',
			notices: [],
			recommendPosts: [],
			isLoadingRecommend: false,
			tabList: [{
					id: '1',
							title: '旅游分享',
							src:'../../icon/笨游黄页.png',
							url:'/packages/home/pages/share/share'
					},
					{
						id: '2',
							title: '景点推荐',
							src:'../../icon/旅游景点.png',
							url:''
					},
					{
						id: '3',
							title: '我的足迹',
							src:'../../icon/旅游攻略.png',
							url:'/packages/user/pages/favoriteList/favoriteList?type=history'
					}
			],

			articleList: []
	},
	

	// 滑动监听
	moduleSelect(e){
			this.setData({
					navIndex:e.detail.current
			})
	},
	/**
	 * 生命周期函数--监听页面加载
	 */
	onLoad(options) {
		this.loadNotices()
	},

	/**
	 * 生命周期函数--监听页面初次渲染完成
	 */
	onReady() {
			var that = this;
			wx.getSystemInfo({
					success(res) {
							that.setData({
									viewHeight:res.windowHeight
							})
					}
			})
	},

	/**
	 * 生命周期函数--监听页面显示
	 */
	onShow() {
		this.loadRecommendPosts()
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
		Promise.all([this.loadNotices(), this.loadRecommendPosts()]).finally(() => {
			wx.stopPullDownRefresh()
		})
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
	,

	onSearchTypeChange(e) {
		this.setData({ searchTypeIndex: Number(e.detail.value || 0) })
	},

	onKeywordInput(e) {
		this.setData({ keyword: e.detail.value })
	},

	doSearch() {
		const keyword = (this.data.keyword || '').trim()
		if (!keyword) {
			wx.showToast({ title: '请输入关键字', icon: 'none' })
			return
		}
		if (this.data.searchTypes[this.data.searchTypeIndex] === '文章') {
			wx.navigateTo({ url: `/packages/home/pages/share/share?keyword=${encodeURIComponent(keyword)}` })
			return
		}
		util.request({
			url: '/travel/user/search',
			method: 'GET',
			data: { keyword: keyword, limit: 10 },
		})
			.then(res => {
				const list = (res && res.data && res.data.data) || []
				const content = list.length
					? list.map(x => `ID:${x.id}  ${x.nick_name || ''}`).join('\n')
					: '无结果'
				wx.showModal({ title: '用户搜索结果', content: content, showCancel: false })
			})
			.catch(() => {
				wx.showToast({ title: '搜索失败', icon: 'none' })
			})
	},

	loadNotices() {
		return util.request({
			url: '/travel/notice/list',
			method: 'GET',
			auth: false,
			data: { pageNum: 1, pageSize: 5 },
		}).then(res => {
			const list = (res && res.data && res.data.data) || []
			this.setData({ notices: list })
		}).catch(() => {})
	},

	onTapNotice() {
		const notices = this.data.notices || []
		if (!notices.length) return
		const n = notices[0]
		wx.showModal({
			title: n.Title || '公告',
			content: n.Content || '',
			showCancel: false,
		})
	},

	loadRecommendPosts() {
		if (this.data.isLoadingRecommend) return Promise.resolve()
		this.setData({ isLoadingRecommend: true })
		return util.request({
			url: '/travel/post/recommand',
			method: 'GET',
			data: { limit: 5 },
		}).then(res => {
			const list = (res && res.data && res.data.data) || []
			this.setData({ recommendPosts: list })
		}).catch(() => {
			this.setData({ recommendPosts: [] })
		}).finally(() => {
			this.setData({ isLoadingRecommend: false })
		})
	},

	toPostDetail(e) {
		const id = e.currentTarget.dataset.id
		if (!id) return
		wx.navigateTo({ url: `/packages/home/pages/articleDetail/articleDetail?id=${id}` })
	},

	onTapTab(e) {
		const id = e.currentTarget.dataset.id
		if (String(id) === '2') {
			wx.setStorageSync('scenicMode', '1')
			wx.switchTab({ url: '/pages/map/map' })
		}
	}
})
