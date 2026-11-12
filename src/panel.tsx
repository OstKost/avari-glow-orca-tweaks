import { atom, Button, Heading, plugin, read, Row, set, Stack, Text } from 'orca'

const message = atom('Ready. Install Go 1.27.3 or newer before using these controls.')
const busy = atom(false)
async function run(action: string) {
  set(busy, true)
  set(message, 'Building the local utility and running the selected action…')
  try {
    const result = await plugin.request('avari.run', { action }, { timeoutMs: 120000 }) as { message: string }
    set(message, result.message)
  } catch (error) {
    set(message, `Action failed: ${String(error)}`)
  } finally { set(busy, false) }
}
export default function Panel() {
  const working = read(busy)
  return <Stack gap="md">
    <Heading>Avari Glow Tweaks</Heading>
    <Text>Optional macOS styling for tabs, panels, worktrees and the composer.</Text>
    <Text>Apply changes files inside Orca.app and its signature. Restart Orca afterwards. Updates can remove the changes. Use Revert before uninstalling this plugin or updating Orca.</Text>
    <Text>Requires Go 1.27.3+ on PATH and access to Go modules. The first build may download dependencies. The Dock icon is not included.</Text>
    <Row gap="xs"><Button disabled={working} onPress={() => run('status')}>Status</Button><Button disabled={working} onPress={() => run('apply')}>Apply</Button><Button disabled={working} onPress={() => run('revert')}>Revert</Button></Row>
    <Text>{read(message)}</Text>
  </Stack>
}
