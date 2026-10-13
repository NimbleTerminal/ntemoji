//go:build js && wasm

// Emojiwasm exposes the Go text-conversion library to a browser worker.
package main

import (
	"syscall/js"

	"nimbleterminal.dev/ntemoji"
)

func main() {
	convert := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 || args[0].Type() != js.TypeObject {
			return map[string]any{"error": "Expected a conversion request"}
		}
		req := args[0]
		for _, key := range []string{"mode", "text", "locale", "open", "close"} {
			if req.Get(key).Type() != js.TypeString {
				return map[string]any{"error": "Missing string field: " + key}
			}
		}
		opts := []ntemoji.TextOption{ntemoji.WithTextLocale(req.Get("locale").String()), ntemoji.WithDelimiters(req.Get("open").String(), req.Get("close").String())}
		var text string
		var err error
		switch req.Get("mode").String() {
		case "demoji":
			text, err = ntemoji.DemojiString(req.Get("text").String(), opts...)
		case "remoji":
			text, err = ntemoji.RemojiString(req.Get("text").String(), opts...)
		default:
			return map[string]any{"error": "Unknown conversion mode"}
		}
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"text": text}
	})
	locales := []any{}
	for _, locale := range ntemoji.Locales() {
		locales = append(locales, locale)
	}
	js.Global().Set("ntemoji", map[string]any{"convert": convert, "locales": locales})
	js.Global().Call("ntemojiReady")
	select {}
}
