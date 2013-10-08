ssh_chat
========

Set your GOPATH correctly

```
export GOPATH=/path/to/your/go/workspace
```

Go get ssh_chat
```
go get github.com/kdorland/ssh_chat
```

Create a private key for the server
```
cd bin/
ssh-keygen -t dsa -f privkey
```

Start up server
```
./ssh_chat
```
Server is now listening for ssh connections on port 8765.
