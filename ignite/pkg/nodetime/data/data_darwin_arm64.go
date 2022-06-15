package data

import _ "embed" // embed 是二進制嵌入所必需的。

//go:embed nodetime-darwin-arm64.tar.gz
var binaryCompressed []byte
