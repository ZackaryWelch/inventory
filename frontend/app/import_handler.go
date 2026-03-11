//go:build js && wasm

package app

import "syscall/js"

// SelectImportFile opens a browser file-picker dialog and processes the selected file.
func (ga *GioApp) SelectImportFile() {
	input := js.Global().Get("document").Call("createElement", "input")
	input.Set("type", "file")
	input.Set("accept", ".csv,.json")

	var changeHandler js.Func
	changeHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		files := input.Get("files")
		if files.Length() == 0 {
			changeHandler.Release()
			return nil
		}

		file := files.Index(0)
		filename := file.Get("name").String()

		reader := js.Global().Get("FileReader").New()

		var loadHandler js.Func
		loadHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			result := reader.Get("result").String()
			ga.handleImportFileContent(result, filename)
			loadHandler.Release()
			changeHandler.Release()
			return nil
		})

		var errorHandler js.Func
		errorHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ga.logger.Error("Failed to read import file")
			errorHandler.Release()
			loadHandler.Release()
			changeHandler.Release()
			return nil
		})

		reader.Set("onload", loadHandler)
		reader.Set("onerror", errorHandler)
		reader.Call("readAsText", file)
		return nil
	})

	input.Call("addEventListener", "change", changeHandler)
	input.Call("click")
}
