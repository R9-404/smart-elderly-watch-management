import { expect, it } from 'vitest'
import { createDeviceState } from '@/composables/deviceModels'
import { useDeviceWorkbench } from '@/composables/useDeviceWorkbench'

it('settles cancelled workbench workers when a network adapter ignores abort', async () => {
  const state = createDeviceState()
  state.deviceList = [{ deviceId: 1, deviceNo: 'watch-1' }]
  state.alertCounts = { '1': { personCount: 1, rateCount: 0, preCount: 0 } }
  const workbench = useDeviceWorkbench(state, () => new Promise(() => {}))
  const pending = workbench.openWorkbench()
  workbench.disposeWorkbench()
  await expect(pending).resolves.toBeUndefined()
  expect(state.workbench.list).toEqual([])
})
