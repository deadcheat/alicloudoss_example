# AliCloud OSS アップロードサンプル

## 実行方法
- $GOPATH配下に`git clone`する
- `cp config.go_default config.go`して、中の定数を適宜変更する
- `glide i`を実行する
- 以下のコマンドで実行
```
go run main.go config.go -d ./files
```
自分のファイルをアップロードしたい場合は-dオプションに与えるディレクトリを変更してください

