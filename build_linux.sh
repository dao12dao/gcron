# build: linux

# build master
swag init --dir ./ -g master/main/master.go -o master/docs
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags doc -o build/master/master master/main/master.go
cp -r master/main/* build/master
rm -f build/master/*.go


# build worker
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/worker/worker worker/main/worker.go
cp -r worker/main/* build/worker
rm -f build/worker/*.go