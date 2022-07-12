### Distributed Task Scheduling System(DTS)

> DTS is written in pure Golang.
> Tasks can be executed in more than one node with lock mode.

#### Project Structure
    - master:The scheduler for create task, query log and send signal, etc.
    - worker:Task executor.
    - common:Common module of master and worker.

#### Dependencies
    + etcd
    + mongodb
    + gin

#### Usage
