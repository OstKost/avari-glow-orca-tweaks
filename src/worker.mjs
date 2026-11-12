import { execFile } from 'node:child_process'
import { promisify } from 'node:util'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { mkdtemp, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'

const exec = promisify(execFile)
const toolDir = fileURLToPath(new URL('../tool/', import.meta.url))
let running = false
async function run(action) {
  if (running) return { message: 'Another action is already running.' }
  if (process.platform !== 'darwin') return { message: 'This edition supports macOS only.' }
  if (!['status', 'apply', 'revert'].includes(action)) return { message: 'Unknown action.' }
  running = true
  let buildDir
  try {
    buildDir = await mkdtemp(join(tmpdir(), 'avari-tweaks-'))
    const binary = join(buildDir, 'avari-orca-tweak')
    await exec('go', ['build', '-o', binary, '.'], { cwd: toolDir, timeout: 100000, maxBuffer: 1024 * 1024 })
    const args = action === 'apply' ? ['apply', '--features', 'tabs,aurora,panels,composer'] : [action]
    const { stdout, stderr } = await exec(binary, args, { cwd: toolDir, timeout: 15000, maxBuffer: 1024 * 1024 })
    return { message: (stdout || stderr || (action === 'apply' ? 'Tweaks written. Restart Orca to load them.' : 'Tweaks reverted. Restart Orca.')).trim() }
  } catch (error) {
    return { message: `Action failed: ${error.stderr || error.message}. Go 1.27.3+ must be installed and available to Orca.` }
  } finally {
    if (buildDir) await rm(buildDir, { recursive: true, force: true }).catch(() => {})
    running = false
  }
}
export default function activate(orca) {
  orca.panels.handle('avari.run', async (input) => run(input?.action))
}
