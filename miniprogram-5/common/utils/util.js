const formatTime = date => {
  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const hour = date.getHours()
  const minute = date.getMinutes()
  const second = date.getSeconds()

  return `${[year, month, day].map(formatNumber).join('/')} ${[hour, minute, second].map(formatNumber).join(':')}`
}

const formatNumber = n => {
  n = n.toString()
  return n[1] ? n : `0${n}`
}

const { API_BASE_URL } = require('../../config/index')

const getToken = () => wx.getStorageSync('token') || ''

const setToken = token => {
  if (!token) return
  wx.setStorageSync('token', token)
}

const clearToken = () => {
  wx.removeStorageSync('token')
}

const ensureToken = () => {
  const existing = getToken()
  if (existing) return Promise.resolve(existing)

  return new Promise((resolve, reject) => {
    wx.login({
      success(res) {
        if (!res.code) {
          reject(new Error('登录失败'))
          return
        }
        wx.request({
          url: `${API_BASE_URL}/travel/login`,
          method: 'POST',
          header: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          data: {
            code: res.code,
          },
          success(loginRes) {
            const token = loginRes && loginRes.data && loginRes.data.token
            if (!token) {
              reject(new Error('登录失败'))
              return
            }
            setToken(token)
            resolve(token)
          },
          fail(err) {
            reject(err)
          },
        })
      },
      fail(err) {
        reject(err)
      },
    })
  })
}

const request = ({ url, method = 'GET', data, header, auth = true }) => {
  const doRequest = token =>
    new Promise((resolve, reject) => {
      const finalHeader = Object.assign({}, header || {})
      if (auth) {
        finalHeader.Authorization = `Bearer ${token}`
      }
      wx.request({
        url: url.startsWith('http') ? url : `${API_BASE_URL}${url}`,
        method,
        data,
        header: finalHeader,
        success(res) {
          resolve(res)
        },
        fail(err) {
          reject(err)
        },
      })
    })

  if (!auth) {
    return doRequest('')
  }
  return ensureToken().then(doRequest)
}

const debugToken = () => {
  const token = getToken()
  if (!token) {
    wx.showToast({ title: 'token 为空', icon: 'none' })
    return
  }
  wx.showModal({ title: 'token', content: token.slice(0, 32) + '...', showCancel: false })
}

module.exports = {
  formatTime,
  API_BASE_URL,
  getToken,
  setToken,
  clearToken,
  ensureToken,
  request,
  debugToken,
}
