package main

import "github.com/egoist/mygo/ui"

// The glyphs of the window's icon: "keyboard" and "keyboard-off" from
// Tabler Icons by Paweł Kuna, under the MIT License
// (https://github.com/tabler/tabler-icons/blob/master/LICENSE).
var (
	keyboardIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2 8a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2zm4 2v.01m4-.01v.01m4-.01v.01m4-.01v.01M6 14v.01M18 14v.01M10 14l4 .01"/></svg>`))

	keyboardOffIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 18H4a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h2m4 0h10a2 2 0 0 1 2 2v8c0 .554-.226 1.056-.59 1.418M6 10v.01m4-.01v.01m4-.01v.01m4-.01v.01M6 14v.01M18 14v.01M10 14h4M3 3l18 18"/></svg>`))
)
