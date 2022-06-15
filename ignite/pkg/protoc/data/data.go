package data

import (
	"embed"
	"io/fs"
)

//go:嵌入包含/* 包含/**/*
var include embed.FS

// Include 返回一個文件系統，其中包含 protoc 使用的標準 proto 文件。
func Include() fs.FS {
	f, _ := fs.Sub(include, "include")
	return f
}

// Binary 返回平台特定的協議二進製文件。
func Binary() []byte {
	return binary
}
