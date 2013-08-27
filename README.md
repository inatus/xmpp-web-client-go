xmpp-web-client
===============

Sample xmpp web client implemented with Golang.

![Sample screen shot](https://raw.github.com/inatus/xmpp-web-client-go/master/screen_shot.png)
#Usage#

1. Execute "go get https://github.com/inatus/xmpp-web-client-go.git"
2. Set host and port to $GOPATH/src/github.com/inatus/xmpp-web-client-go/config
3. Run "go run $GOPATH/src/github.com/inatus/xmpp-web-client-go/server.go"
4. Connect to "http://HOST_NAME:PORT_NUMBER" from browser
5. (In case of connecting to Google Hangout,) input User Name (xxxx@gmail.com) and Password of Gmail account.

#To-Do#

* Code is messy.
* Connections are forever alive until the application gets terminated.
* Support SSL.
* Improve design.

#License#

You may use the project under the terms of the MIT License
