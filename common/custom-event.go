// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type stringWriter interface {
	io.Writer
	writeString(string) (int, error)
}

type stringWrapper struct {
	io.Writer
}

func (w stringWrapper) writeString(str string) (int, error) {
	return w.Writer.Write([]byte(str))
}

func checkWriter(writer io.Writer) stringWriter {
	if w, ok := writer.(stringWriter); ok {
		return w
	} else {
		return stringWrapper{writer}
	}
}

// Server-Sent Events
// W3C Working Draft 29 October 2009
// http://www.w3.org/TR/2009/WD-eventsource-20091029/

var contentType = []string{"text/event-stream"}
var noCache = []string{"no-cache"}

var fieldReplacer = strings.NewReplacer(
	"\n", "\\n",
	"\r", "\\r")

var dataReplacer = strings.NewReplacer(
	"\n", "\n",
	"\r", "\\r")

type CustomEvent struct {
	Event string
	Id    string
	Retry uint
	Data  interface{}
}

func encode(writer io.Writer, event CustomEvent) error {
	w := checkWriter(writer)
	return writeData(w, event.Data)
}

func writeData(w stringWriter, data interface{}) error {
	dataStr := fmt.Sprint(data)
	dataReplacer.WriteString(w, dataStr)

	// 确保每个事件后面都有两个换行符
	if strings.HasPrefix(dataStr, "data") {
		// 如果数据行已经结尾有两个换行符，不再添加
		if !strings.HasSuffix(dataStr, "\n\n") {
			// 如果只有一个换行符，则再添加一个
			if strings.HasSuffix(dataStr, "\n") {
				w.writeString("\n")
			} else {
				// 如果没有换行符，添加两个
				w.writeString("\n\n")
			}
		}
	} else {
		// 普通数据行添加两个换行
		w.writeString("\n\n")
	}

	return nil
}

func (r CustomEvent) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	err := encode(w, r)

	// 尝试立即刷新数据
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return err
}

func (r CustomEvent) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	header["Content-Type"] = contentType

	if _, exist := header["Cache-Control"]; !exist {
		header["Cache-Control"] = noCache
	}
}
