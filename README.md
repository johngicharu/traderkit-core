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


## build

### server

```bash
make build-server
# check ./bin/server as the result
```

### controller

```bash
make build-controller
# check ./bin/controller as the result
```

## dev run (with watch)

make sure you have wgo installed

```bash
brew install wgo
```

```bash
make -j2 dev
```

# Todo

- [ ] functions for terminal init/deployment
- [ ] Add handlers for all account actions
- [ ] Add handlers for controller actions
- [ ] add handlers for trade sync details
- [ ] Complete function that clears response json files
- [ ] mql update the json structures
- [ ] mql ensure we have notice of whether we are acking every message or only trade messages
- [ ] mql file saving for responses (maybe only trade responses though)
- [ ] think through copying logic and implement
- [ ] explore price feed structure and implement
- [ ] admin actions for dashboard
- [ ] user actions for dashboard
- [ ] user metrics for dashboard
- [ ] telegram bot for users
