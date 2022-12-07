# observatory-task

The observatory task receives messages from the data notify server through the message queue for parsing and storage of various filecoin data tasks.

## Feature

1. Support a variety of storage methods, currently std output (for debug) and postgresql have been implemented. For
   databases, we currently use `xorm` and only support postgresql.
2. Task abstraction, only need to implement the `pkg/tasks/Task` interface to write your filecoin parsing task.
3. Support use metric trace task.

## Task Model
* block_header
* block_message
* block_parent
* message
* receipt
* raw_actor
* evm_block_header
* evm_internal_tx
* evm_receipt
* evm_transaction
* evm_contract
* evm_address

## Build And Run

1. `make build`
2. `bin/observatory-task --conf=<config path>`

> When you use database as storage, you do not need to manually create a database, the program will automatically 
synchronize the table structure to the specified database according to the model required by the tasks in the configuration

## License

Dual-licensed under [MIT](https://github.com/Spacescope/observatory-task/blob/main/LICENSE-MIT) + [Apache 2.0](https://github.com/Spacescope/observatory-task/blob/main/LICENSE-APACHE) 


## Doc
https://drive.google.com/drive/u/0/folders/1ptiBCy4lsO78KJqQR3oYv2TXrk3BrH8p
