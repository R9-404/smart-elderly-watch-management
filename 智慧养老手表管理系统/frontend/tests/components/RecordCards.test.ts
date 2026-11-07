import { h } from 'vue'
import { mount } from '@vue/test-utils'
import { expect, it, vi } from 'vitest'
import RecordCards from '@/components/RecordCards.vue'

it('keeps field calculations and action-slot data reactive after prop replacement', async () => {
  const clicked = vi.fn()
  const wrapper = mount(RecordCards, {
    props: {
      records: [{ oldId: 7, name: '田风', room: 'A', phone: '13800000001' }],
      fields: [{ key: 'room', label: '房间' }], label: '长者记录'
    },
    slots: {
      actions: ({ row, $index }: { row: Record<string, unknown>; $index: number }) =>
        h('button', { onClick: () => clicked(row.oldId, $index) }, '处理')
    }
  })
  expect(wrapper.get('dd').text()).toBe('A')
  await wrapper.setProps({
    records: [{ oldId: 8, name: '更新姓名', room: 'B', phone: '13800000002' }],
    fields: [{ key: 'phone', label: '电话' }]
  })
  expect(wrapper.get('h2').text()).toBe('更新姓名')
  expect(wrapper.get('dt').text()).toBe('电话')
  expect(wrapper.get('dd').text()).toBe('13800000002')
  await wrapper.get('button').trigger('click')
  expect(clicked).toHaveBeenCalledExactlyOnceWith(8, 0)
  wrapper.unmount()
})

it('retains the default empty message and reacts to an explicit replacement', async () => {
  const wrapper = mount(RecordCards, { props: { records: [], fields: [], label: '记录' } })
  expect(wrapper.get('[role="status"] strong').text()).toBe('暂无记录')
  await wrapper.setProps({ emptyTitle: '暂无健康记录' })
  expect(wrapper.get('[role="status"] strong').text()).toBe('暂无健康记录')
  wrapper.unmount()
})
