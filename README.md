# Metadata-task

The metadata task receives messages from the data notify server through the message queue for parsing and storage of various filecoin data tasks.

These tasks are focused on extracting metadata. And the [aggregate tasks](https://github.com/Spacescope/aggregate-task), depending on the metadata extracted from the Metadata tasks.

Remember, if you modify the models of this repository, please make sure the [models of aggregate-task](https://github.com/Spacescope/aggregate-task/tree/main/pkg/models/dependentmodel) are also changed.

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
* builtin_actor_events
* evm_block_header
* evm_internal_tx
* evm_receipt
* evm_transaction
* evm_contract
* evm_address

## Build And Run

1. `make binary`
2. `bin/meta-task --conf <config path>`

> When you use database as storage, you do not need to manually create a database, the program will automatically 
synchronize the table structure to the specified database according to the model required by the tasks in the configuration


## 

## License

Dual-licensed under [MIT](https://github.com/Spacescope/meta-task/blob/main/LICENSE-MIT) + [Apache 2.0](https://github.com/Spacescope/meta-task/blob/main/LICENSE-APACHE) 


## System Design Doc
https://docs.google.com/document/d/1ZDgiejhK0J4RfCvfug4YuG8csaKgW3Io-cozPw1pmpE/edit?usp=sharing
