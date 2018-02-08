# lmd-image-resizer
- S3にアップされた画像をリサイズして保存し直すファンクション

# 概要
- S3がputObjectイベントを発火
- このファンクションがイベントを解析して、`rule.go` 内のルールに適合すればリサイズしてS3に保存し直す

# 参考
- [Golangサポート開始通知のブログ記事](https://aws.amazon.com/jp/blogs/compute/announcing-go-support-for-aws-lambda/)
- [サンプルコード](https://github.com/aws-samples/lambda-go-samples/blob/master/main.go)

# 検討の余地あり
- ruleの置き場/読み込み方(現状は `rule.go` )
