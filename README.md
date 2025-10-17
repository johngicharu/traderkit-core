# Traderkit Core

# Docs

```md
ack approach

server sends request
controller forwards request to terminal
terminal ack's request immediately
controller forward's ack to server
or check if response arrives within specific time, if it doesn't we retry

terminal has response & creates response file
sends response to controller
controller forwards it to server
server ack's response
controller deletes response file
```
