# traq-system-bot

[![GitHub release](https://img.shields.io/github/release/traPtitech/traq-system-bot.svg)](https://github.com/traPtitech/traq-system-bot/releases/)

[traQ](https://github.com/traPtitech/traQ) のシステムBOTです。

以下のイベントの通知を、設定したチャンネルに行います。

- ユーザーの加入 (`USER_CREATED`)
- ユーザーの再加入 (`USER_ACTIVATED`)
- チャンネルの作成 (`CHANNEL_CREATED`)
- スタンプの作成 (`STAMP_CREATED`)
- スタンプの更新 (`STAMP_UPDATED`): 更新前後の名前を通知
- スタンプの削除 (`STAMP_DELETED`): 削除前の名前を通知

起動時に全スタンプのIDと名前をメモリに読み込み、スタンプの作成・更新・削除イベントに合わせて更新します。
名前が変わらない更新（画像の変更など）も通知します。

## 設定方法

推奨: [bot-console](https://github.com/traPtitech/traQ-bot-console) を事前に設置すること。

1. (bot-console) 特権を持った（`privileged`）BOTをWebSocket Modeで作成する。
2. (bot-console) 上記のイベント購読設定を行う。
   - 更新・削除の通知を使う場合は、キャッシュを維持するため `STAMP_CREATED` / `STAMP_UPDATED` / `STAMP_DELETED` の全てを購読してください。
     通知の有効・無効は `BOT_SYSTEM_SUBSCRIBING_EVENTS` で指定できます。
3. 本BOTを、下記の環境変数を設定してデプロイする。
    - [本リポジトリが公開するDockerイメージ](https://github.com/traPtitech/traq-system-bot/pkgs/container/traq-system-bot) を使うと楽。
    - 最新バージョンはGitHubのリリースページを確認してください。

### 環境変数

- `BOT_SYSTEM_MESSAGE_CHANNEL_ID`: 通知メッセージを投稿したいチャンネルのUUID。
- `BOT_SYSTEM_SUBSCRIBING_EVENTS`: 通知するイベントのID (カンマ区切り)
  - 設定しないときは全てのイベントを通知する
  - `USER_CREATED` / `USER_ACTIVATED` / `CHANNEL_CREATED` / `STAMP_CREATED` / `STAMP_UPDATED` / `STAMP_DELETED`
- `TRAQ_ORIGIN`: 接続するtraQインスタンスのURL。WebSocketプロトコルを使用。例: `wss://q.trap.jp`
- `BOT_ACCESS_TOKEN`: Botのアクセストークン。
