{
  "name": "{{ replace .Module.Pkg.Name "." "-" }}-js",
  "version": "0.1.0",
  "description": "為 Cosmos 模塊自動生成的 vuex 存儲 {{ .Module.Pkg.Name }}",
  "author": "熊網鏈代碼生成器 <bear.network.root@gmail.com>",
  "homepage": "http://{{ .Module.Pkg.GoImportName }}",
  "license": "Apache-2.0",
  "licenses": [
    {
      "type": "Apache-2.0",
      "url": "http://www.apache.org/licenses/LICENSE-2.0"
    }
  ],
  "main": "index.js",
  "publishConfig": {
    "access": "public"
  }
}