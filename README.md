# Avari Glow Tweaks

Optional macOS interface tweaks for [Avari Glow](https://github.com/OstKost/avari-glow-orca-theme): tabs, panels, worktree highlights and composer. This is an independent, experimental plugin, not affiliated with Orca.

## Requirements and operation

Requires Orca 1.4.246+, macOS, Go 1.27.3 or newer on the PATH inherited by Orca, network access for Go dependencies on first build, and write access to `/Applications/Orca.app`. Install the plugin, approve its **full access**, and open its sidebar panel. Installation alone does not patch Orca. **Status**, **Apply**, and **Revert** each compile the included source locally and run the resulting temporary executable. Apply selects `tabs,aurora,panels,composer`; it does not install a Dock icon.

Apply modifies the installed application's CSS, HTML loader and script and breaks the app bundle signature. This may affect integrity checks and updates. Restart Orca after Apply or Revert. Use Revert before uninstalling this plugin or updating Orca; uninstalling the plugin does not undo patches. The utility preserves backups beside modified files and refuses an unsafe loader restore. If Orca is updated before reverting, consult the [manual recovery documentation](https://github.com/OstKost/orca-avari-glow-theme#откат-и-обновление). The status reports disk state, not whether the running renderer loaded it.

The panel runs while Orca is open; changes take effect only on restart. Back up important data before modifying an application. Internal selectors and loader structure are version-dependent, and a new Orca version can require changes. The full project and standalone utility are at https://github.com/OstKost/orca-avari-glow-theme.
