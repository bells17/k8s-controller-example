# k8s-controller-example

Kubernetesのコントローラー実装を読むための勉強用のサンプルコントローラーです。

- Kubernetesのソースコードを読む際の慣れを想定しているためkubebuilderは使用していません。
- kubernetes/sample-controllerを参考に作成しています。

## コントローラーについて

各 `Deployment` リソースに `sample-controller: "True"` アノテーションを追加するだけのシンプルなコントローラーです。

## 試し方

### devcontainerで開く

devcontainerに対応させているのでVSCodeなどdevcontainerに対応したエディタで開いてください。
もし難しい場合は下記の設定を行ってください。

- ツールのインストール - kubectl, kind, make
- Docker環境の構築
- Go言語の設定

### Kubernetesクラスターの構築とコントローラーの起動

下記のコマンドを実行してください

```
$ make launch-kind
$ make run
```

### 動作確認

別のターミナルを開いて下記を実行してください。

```
$ make apply-deployment
```

上記を実行することでサンプルのDeploymentをインストールします。

下記のコマンドを実行することで、Deploymentリソースのアノテーションが追加されたことを確認したり、Deploymentに紐づくイベントの発行を行ったことが確認できます。

```
$ make check-deployment
$ make check-event
```

また、コントローラーを実行しているターミナルを見ると処理内容に関するログを見ることができます。

### クリーンアップ

`make run` の実行を停止して下記コマンドを実行します

```
$ make stop-kind
```

