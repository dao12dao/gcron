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

#### Deployment
  + Build master and run, also can run with nginx.
    + ```bash> go build -o master/main/cron_master master/main/master.go```
    + ```bash> ./cron_master -config config.ini```
  + Build worker and run.
    + ```bash> go build -o worker/main/cron_worker worker/main/worker.go```
    + ```bash> ./cron_worker -config config.ini```
  + Visit the backend web page.
    + ```http://localhost:8080/web```

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


  

