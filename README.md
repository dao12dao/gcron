### Distributed Task Scheduling System(DTS)

>
> DTS is written in pure Golang.
> Tasks can be executed in more than one node with lock mode.
> 

#### Project Structure
  - master:The scheduler for create task, query log and send signal, etc.
  - worker:Task executor.
  - common:Common module of master and worker.

#### Dependencies
  + etcd
  + mongodb
  + gin

#### Features
  + Create/Edit Task
    - TaskName: means the name of task, keep it unique.
    - ShellCommand: means what the task will do.
    - CronExpr: https://github.com/gorhill/cronexpr.
  + Delete Task
    - Just Remove Task by TaskName.
  + Kill Task
    - Interrupt the running task.
  + View Log
    - List the task log.

#### Deployment
  + Generate swagger docs for api in master.(Optional)
    + ```bash> swag init --dir ./ -g master/main/master.go -o master/docs```
  + Build master and run, also can run with nginx.
    + ```bash> go build (-tags doc) -o master/main/master master/main/master.go```
    + ```bash> cd master/main && ./master -config config.ini```
  + Build worker and run.
    + ```bash> go build -o worker/main/worker worker/main/worker.go```
    + ```bash> cd worker/main && ./worker -config config.ini```
  + Visit the backend web page.
    + ```http://localhost:8080/web```
  + Visit the api docs page.
    + ```http://localhost:8080/docs``` or ```http://localhost:8080/swagger/index.html```

#### Systemctl(Linux)/Launchctl(MacOS)
> Run at system loaded with systemctl in linux or with launchctl in macos.
  + create config file.
    + Linux: copy any service file in the path ```/etc/systemd```, then rename the file and modify the fields like ```Description```、```ExecStart```、```WorkingDirectory```.
    + MacOS: copy any plist file in the path ```~/Library/LaunchAgents/```, then rename the file and modify the fields like ```Label```、```ProgramArguments```、```WorkingDirectory```.
  + enable the file.
    + Linux: ```bash> systemctl enable xx.service```
    + MacOS: ```bash> launchctl load -w xxx.plist```
  + start to run.
    + Linux: ```bash> systemctl start xx.service```


  

