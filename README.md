# observatory-task

The observatory task receives messages from the data notify server through the message queue for parsing and storage of various filecoin data tasks.

## Feature

1. Support a variety of storage methods, currently std output (for debug) and postgresql have been implemented. For
   databases, we currently use `xorm` and only support postgresql.
2. Task abstraction, only need to implement the `pkg/tasks/Task` interface to write your filecoin parsing task.
3. Support use metric trace task.

## Config Sample

```toml
[task]
name = "message" // taskname 

[storage]
name = "pgsql" // use postgresql 

[storage.params]   // pgsql params
dsn = "xx://xxx" 
max_idle=1
max_open=1

[lotus]
addr = "https://api.node.glif.io/rpc/v1"   // lotus node

[chain_notify]
host = ""       // data notify server host

[chain_notify.mq]   // data notify server message queue 
name = "redis"    // use redis

[chain_notify.mq.params]   // redis params
dsn = "redis://127.0.0.1:6379/0"
queue_name = "message" // must same as task.name
```

## Build And Run

1. `make build`
2. `bin/observatory-task --conf=<config path>`

> When you use database as storage, you do not need to manually create a database, the program will automatically 
synchronize the table structure to the specified database according to the model required by the tasks in the configuration

## License

Dual-licensed under [MIT](https://github.com/Spacescope/observatory-task/blob/main/LICENSE-MIT) + [Apache 2.0](https://github.com/Spacescope/observatory-task/blob/main/LICENSE-APACHE) 

