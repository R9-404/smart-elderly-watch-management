import { expect, it } from 'vitest'
import { CanceledError, type AxiosResponse } from 'axios'
import { apiErrorMessage, assertApiResponse, isApplicationAPI } from './httpContract'

const response = (data: unknown) => ({ data, status: 200, config: { url: '/api/v1/customer/select' } }) as AxiosResponse

it('accepts the common success envelope and rejects missing or failed results', () => {
  const success = response({ returnCode: '1111', customers: [], total: 0 })
  expect(assertApiResponse(success)).toBe(success)
  expect(() => assertApiResponse(response({ customers: [] }))).toThrow('服务响应格式错误')
  expect(() => assertApiResponse(response({ returnCode: '0000', code: 'CONFLICT', message: '绑定已变化' }))).toThrow('绑定已变化')
})

it('retains server messages, reference identifiers, HTTP definitions and cancellation', () => {
  expect(apiErrorMessage({ response: { status: 500, data: { returnCode: '0000', code: 'INTERNAL_ERROR', message: '保存失败', reference: 'r-1' } } })).toBe('保存失败（编号：r-1）')
  expect(apiErrorMessage({ response: { status: 429 } })).toBe('请求过于频繁，请稍后重试')
  expect(apiErrorMessage(new Error('network'), '连接中断')) .toBe('连接中断')
  expect(apiErrorMessage(new CanceledError())).toBe('请求已取消')
  expect(apiErrorMessage({ name: 'UploadAjaxError', status: 400, message: JSON.stringify({ returnCode: '0000', code: 'INVALID_REQUEST', message: '图片格式不支持' }) })).toBe('图片格式不支持')
  expect(apiErrorMessage(new Error('rejected', { cause: { returnCode: '0000', code: 'INVALID_REQUEST', message: '手机号已使用' } }))).toBe('手机号已使用')
})

it('leaves external responses and binary downloads to their own protocols', () => {
  expect(isApplicationAPI('https://other.example/api/v1/customer/select')).toBe(false)
  const external = { ...response('ok'), config: { url: 'https://other.example/api/v1/a' } } as AxiosResponse
  expect(assertApiResponse(external)).toBe(external)
  const image = { ...response(new Blob()), config: { url: '/api/v1/image/download?name=x', responseType: 'blob' } } as AxiosResponse
  expect(assertApiResponse(image)).toBe(image)
})
