// pages/article/article.js
const app = getApp();
const util = require('../../common/utils/util')


Page({

  /**
   * 页面的初始数据
   */
  data: {
    title:'',
    content:'',
    headImg: []
  },
  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {

  },

  bindTitleInput(e) {
    this.setData({ title: e.detail.value });
  },
  bindContentInput(e) {
    this.setData({ content: e.detail.value });
  },

  //选择图片
  chooseImage:function(){
    wx.chooseImage({
      count: 9, 
      sizeType: ['original', 'compressed'],
      sourceType: ['album', 'camera'],
      success: (res)=>{
        const tempFilePaths = res.tempFilePaths;
        this.setData({
          headImg: this.data.headImg.concat(tempFilePaths)
        })
      },
      fail:()=>{
        wx.showToast({
          title: '上传图片失败',
          icon: 'error'
        })
      }
    })
  },

  //预览图片
  previewImg:function(e){
    const img = e.currentTarget.dataset.img;
    wx.previewImage({
      current: img, // 当前显示图片的http链接
      urls: this.data.headImg, // 需要预览的图片http链接列表
      success: () => {

      },
      fail: () => {
        wx.showToast({ title: '预览图片失败', icon: 'none' })
      }
    })
  },

  removeImage:function(e) {
    const img = e.currentTarget.dataset.img;
    let temArr = this.data.headImg.slice();
    for(let i=0;i<temArr.length;i++){
      if(temArr[i]===img){
        temArr.splice(i,1);
        break;
      }
    };
    this.setData({
      headImg:temArr
    })
  },


// //------------------采取本地存储
//   submitPost() {
//     const article = {
//       id: Date.now(), // 用时间戳模拟唯一ID
//       title: this.data.title,
//       content: this.data.content,
//       headImg: this.data.headImg,
//       avatarUrl: this.data.avatarUrl,
//       nickname: this.data.nickname,
//       time: new Date().toLocaleString()
//     };

//     wx.showToast({
//       title: '发布成功',
//     })
    
//     // 从本地存储获取已有文章列表
//     const articles = wx.getStorageSync('articles') || [];
//     articles.unshift(article); // 添加到列表头部
//     wx.setStorageSync('articles', articles);

//     // 跳转到列表页
//     wx.switchTab({
//       url: '../../packages/home/pages/share/share'
//     });
//   },

  submitPost(){
    const { title, content, headImg } = this.data;
    if (!title || !content ) {
      wx.showToast({
        title: '请填写完整信息',
        icon: 'none'
      });
      return;
    }

    const headImgValue = Array.isArray(headImg) && headImg.length > 0 ? headImg[0] : ''
    util.request({
      url: '/travel/post/create',
      method: 'POST',
      header: { 'Content-Type': 'application/json' },
      data: {
        title: title,
        head_img: headImgValue,
        content: content,
      },
    })
      .then(() => {
        wx.showToast({
          title: '文章发布成功',
          icon: 'success'
        });
        this.setData({ title: '', content: '', headImg: [] })
        wx.switchTab({ url: '/pages/home/home' })
      })
      .catch(() => {
        wx.showToast({
          title: '文章发布失败',
          icon: 'none'
        });
      })
  },


  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady() {

  },


})
