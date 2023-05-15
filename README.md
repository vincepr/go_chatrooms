# Websockets in go
Proof of concept to get familiar with websockets in golang. Want to actually use them for something else spanning between Broser-JS -> GoServer <- C#-Application, so a small scale test like this went first.

Goals:
- Chatrooms, with a free chatroom selector
- Communication in the Rooms goes over the Server via websockets.
- maybe test the SSL variant `wss vs ws` of Websockets. 

## info on websockets:
- RFC for Websocket Protocol: https://www.rfc-editor.org/rfc/rfc6455.html#section-11.8
- mdn docs: https://developer.mozilla.org/en-US/docs/Web/API/WebSocket
    - instance methods:
        - Websocket.close()
        - Websocket.send()
    - events:
        - close
        - error
        - message
        - open

### ping and pong
To keep the websocket connection open, sending ping and pongs in fixed intervalls.

In this case the Server is sending the pings. And if no Pong is received back it assumes the connection closed and cleans it up.

## Notes about implementation
- Manager : struct to handle (all) clients connecting over websockets
    - read-write mutex here. For Async access to the ClientMap
    - also a map for different handlers for each supported Event Type
        
- Event Type :  blueprints for data (ex. messages) sent over WS.
    - easier to support different message/data-stream Types sent over websockets
- Client : each connection upgrading to a Websocket gets a Client struct.
    - to avoid concurrent writes/race-conditions we use a go-channel `egress`. Because We cant read and write at the same time etc. (the channel will block and do one at a time)
    - we also use that channel to handle the timeout checks. (again using the same block to aboid raceconditions)
    - Remove Inactive/Disconected connections. Webbrowsers implement Ping-Pong on Ws. So we can just send a Ping in fixed intervalls and check if we get the Pong in time. If not we can assume the connection dead.
    - We set limited support for Max-Message size. To avoid abuse.
- main: here we do some setup
    - we also implement `CheckOrign` to be able to filter for allowed sites connecting and upgrading to Websocket connections. To avoid Cross-Site Requests etc.


## Authentification
- happens before upgrading from http to a websocket connection.
- 2 recommended solutions:
    - a regular http request to authenticate returns a one-time-use password to connect
    - connect a WebSocket but dont accept any messages untill a special Auth-message with credentials gets sent.

