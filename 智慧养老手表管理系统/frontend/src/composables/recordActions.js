import { accepted, api } from '../api/client'
import { apiErrorMessage } from '../api/httpContract'

export async function writeRecord(view, path, body, success, refresh, failure = '请求未完成，请重试') {
  try {
    const { data } = await api.post(path, body)
    accepted(data)
    view.$message.success(success)
    refresh()
  } catch (error) {
    view.$message.error(apiErrorMessage(error, failure))
  }
}

export async function deleteRecord(view, path, body, prompt, refresh) {
  try {
    await view.$confirm(prompt, '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
  } catch {
    view.$message.info('已取消删除')
    return
  }
  return writeRecord(view, path, body, '删除信息成功', refresh, '删除失败，请稍后重试')
}
