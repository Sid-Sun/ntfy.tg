### ntfy.tg

#### a [telegram bot](https://t.me/ntfytgbot) for subscribing to ntfy.sh topics

#### What is supported

- receiving notifications from ntfy.sh topics using the topic name

#### What is not supported

- publishing messages to topics from telegram

#### How to self-host

- this bot uses "Storage Engine" to persist user data across restarts, you need to host an instance of [this](https://github.com/Sid-Sun/notes-api/tree/mongo) - it needs a mongoDB instance, I recommend using MongoDB Atlas if you don't have an instance running
- the environment variables needed are:
  - `APP_ENV=prod` - used by logger
  - `API_TOKEN` - telegram bot API token
  - `ADMIN_CHAT_ID` -used to notify when subscriber restarts
  - `SE_OBJ_ID` - object ID / note name to use on Storage Engine
  - `SE_OBJ_PASS` - object / note password
  - `SE_URL` - base URL of above hosted API (ex: http://127.0.0.1:3003)
  - `NTFY_DOMAIN` - just the domain name of ntfy.sh instance (ex: ntfy.sh); one ntfy.tg instance can only support one ntfy.sh instance.
    - The server must support websockets with a trusted TLS certificate

##### Why do I need to host "Storage Engine"?

- The bot uses it to store data mapping topic name to user ID, thanks to [this library](https://github.com/fitant/storage-engine-go) I built on top of the [storeage engine API](https://github.com/Sid-Sun/notes-api/tree/mongo), it is very easy to add a backing store for small things like this. If you want to, you can add a proper store into the code base
- To get an idea of how easy this was to add, see diff for `subscription_manager.go` in [this](https://github.com/Sid-Sun/ntfy.tg/commit/840b5dc0f3e0273b6dd629b92febf255c2bfd619#diff-d6c30fc5f1c897d45c4b1f450325bf30dca3e7c953ea9b1ea3e92936d628d302) commit
