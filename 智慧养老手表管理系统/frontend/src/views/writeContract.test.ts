import { flushPromises } from '@vue/test-utils'
import { expect, it, vi } from 'vitest'
import axios from 'axios'
import { residentCommandMethods } from './person.command-methods'
import { deviceAdminMethods } from './device.admin-methods'

it.each(['addOldPersons', 'editOldPerson'])('%s sends editable fields and the selected guardian only', async method => {
  const fields = { name: '长者', bothdate: '1950-01-01', sex: '女', phone: '13800000001', room: 'A', illness: '', warn: '', foodLike: '', foodHate: '', drag: '', customerId: 2 }
  const draft = { ...fields, oldId: 7, beTime: 'old', familyName: 'display', deviceId: 3, customerName: 'display' }
  const post = vi.spyOn(axios, 'post').mockResolvedValue({ data: { returnCode: '1111' } })
  const validate = (done: (valid: boolean) => void) => done(true)
  const page = { oldForm: draft, oldEditForm: draft, $refs: { addoldFormRef: { validate }, editoldFormRef: { validate } }, $message: vi.fn(), addCustomerList: vi.fn(), searchOldPersons: vi.fn() }
  try {
    residentCommandMethods[method as keyof typeof residentCommandMethods].call(page)
    await flushPromises()
    expect(post).toHaveBeenCalledTimes(1)
    expect(post.mock.calls[0][1]).toEqual(method === 'editOldPerson' ? { ...fields, oldId: 7 } : fields)
  } finally { vi.restoreAllMocks() }
})

it.each(['resetSecret', 'unbind', 'deleteDevice'])('%s displays the rejected write once', async method => {
  const post = vi.fn().mockRejectedValue({ response: { status: 409, data: { returnCode: '0000', code: 'CONFLICT', message: '设备归属已变化' } } })
  const error = vi.fn(), loadDevices = vi.fn()
  const page = { post, loadDevices, $confirm: vi.fn().mockResolvedValue('confirm'), $message: { error } }
  deviceAdminMethods[method as 'resetSecret' | 'unbind' | 'deleteDevice'].call(page, { deviceId: 9 })
  await flushPromises()
  expect(post).toHaveBeenCalledTimes(1)
  expect(error).toHaveBeenCalledExactlyOnceWith('设备归属已变化')
  expect(loadDevices).not.toHaveBeenCalled()
})
