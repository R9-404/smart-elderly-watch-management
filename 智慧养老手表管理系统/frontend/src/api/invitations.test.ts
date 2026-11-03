import { afterEach, expect, it, vi } from 'vitest'
import axios from 'axios'
import { invitations } from './invitations'
import InvitePage from '../views/FamilyInvitations.vue'
import RegisterPage from '../views/Register.vue'
import { flushPromises } from '@vue/test-utils'
afterEach(() => vi.restoreAllMocks())
it('preserves invitation routes, payloads and cookies', async () => {
  const post = vi.spyOn(axios, 'post').mockResolvedValue({ data: { returnCode: '1111' } })
  const cases = [['checkCode', { code: 'abc' }], ['join', { code: 'abc' }], ['list', { page: 2, pageSize: 10, state: 'used' }], ['oldPersonList', {}], ['create', { oldId: 7, createUser: 'operator' }], ['delete', { id: 9 }]] as const
  for (const [name, body] of cases) {
    await (invitations[name] as Function)(body)
    expect(post).toHaveBeenLastCalledWith(`api/v1/inviteCode/${name}`, body, { withCredentials: true, signal: undefined })
  }
})
it('clears creation loading after request failure', async () => {
  vi.spyOn(axios, 'post').mockRejectedValue(new Error('offline'))
  const page = { creating: false, createForm: { oldId: 7, count: 1, expireDays: '' }, $refs: { form: { validate: (callback: Function) => callback(true) } }, $message: vi.fn() }
  InvitePage.methods!.createInviteCode.call(page, 'form')
  expect(page.creating).toBe(true)
  await flushPromises()
  expect(page.creating).toBe(false)
  expect(page.$message).toHaveBeenCalledOnce()
})
it('updates registration code feedback in the component context', async () => {
  vi.spyOn(axios, 'post').mockResolvedValue({ data: { returnCode: '1111', oldPersonName: '长者', remainDays: -1 } })
  const page = { registerForm: { code: ' abc ' }, codeOk: false, codeMessage: '' }
  RegisterPage.methods!.checkCode.call(page)
  await flushPromises()
  expect(page.codeOk).toBe(true)
  expect(page.codeMessage).toContain('长者（长期有效）')
})
