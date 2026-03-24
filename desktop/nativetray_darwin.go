package desktop

// NativeTray is a no-op placeholder.
// macOS systray is not compatible with Wails v2 because both Wails and systray
// libraries (getlantern/systray, fyne.io/systray, raw CGO NSStatusBar) each
// require owning the Cocoa NSApplication / AppDelegate, causing duplicate-symbol
// linker errors or SIGSEGV crashes at runtime.
//
// Instead, the app uses HideWindowOnClose so that closing the window keeps the
// app alive in the macOS Dock. Click the Dock icon to reopen.
type NativeTray struct {
	app *App
}

func NewNativeTray(_ *App) *NativeTray {
	return nil
}
