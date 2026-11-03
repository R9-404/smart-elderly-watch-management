import { api as requestClient } from '../api/client'
import { apiErrorMessage } from '../api/httpContract'

export const deviceRefreshMethods = {
    loadOldPeople () {
      var _this = this
      requestClient.post('api/v1/inviteCode/oldPersonList', {}).then(function (response) {
        _this.oldPeople = response.data.oldPeople || []
      }).catch(function (error) {
        _this.$message.error(apiErrorMessage(error, '请求未完成，请重试'))
      })
    },
    startTimer () {
      this.stopTimer()
      if (!this.autoRefresh || document.hidden || this.refreshDisposed) return
      this.timer = setTimeout(() => {
        this.timer = null
        Promise.resolve(this.loadDevices()).then(() => { this.startTimer() })
      }, 5000)
    },
    handleVisibilityChange () {
      this.stopTimer()
      if (!document.hidden && this.autoRefresh) Promise.resolve(this.loadDevices()).then(() => { this.startTimer() })
    },
    stopTimer () {
      if (this.timer) {
        clearTimeout(this.timer)
        this.timer = null
      }
    },
}
