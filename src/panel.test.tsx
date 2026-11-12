import { assert, mockHost, mount, test } from 'orca/testing'
import * as panel from './panel'

test('explicit actions reach the worker and display its result', async () => {
  const actions: string[] = []
  const host = mockHost({ requests: { 'avari.run': (input: any) => { actions.push(input.action); return { message: `done ${input.action}` } } } })
  const view = await mount(panel, { host })
  assert.deepEqual(actions, [])
  await view.press({ text: 'Status' })
  await view.press({ text: 'Apply' })
  await view.press({ text: 'Revert' })
  assert.deepEqual(actions, ['status', 'apply', 'revert'])
  assert.deepEqual(view.errors, [])
  assert.deepEqual(view.refusals, [])
  await view.unmount()
})
