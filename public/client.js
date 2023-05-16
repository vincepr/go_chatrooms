/*
*
*   Setting globals up and connecting handlers
*
*/


// initial selection for chat-room:
var selectedChat = "general"

// holds reference to the Websocketconnection once established:
var wsConn = null;  

// setup the onSubmit callbacks
document.getElementById("form-selection").addEventListener("submit", handleRoomSelection);
document.getElementById("form-message").addEventListener("submit", handleSendMessage);
document.getElementById("login").addEventListener("submit", handleLogin);


/*
*
*   Button/Submit Handlers:
*
*/

// handles the onSubmit for Room Selection
function handleRoomSelection(ev) {
    ev.preventDefault();
    let newChat = document.getElementById("chatroom");
    if (newChat != null && newChat.value != selectedChat){
        selectedChat = newChat.value;
        document.getElementById("chat-header").innerHTML = "Currently in room: "+selectedChat;
        
        let newRoomEvent = new ChangeChatRoomEvent(selectedChat);
        sendEvent("change_room", newRoomEvent);
        document.getElementById("chatmessages").innerHTML = "Changed to Room: "+selectedChat;
        

    }
}

// handles the onSubmit for Message Sending
function handleSendMessage(ev) {
    ev.preventDefault();
    let newMessage = document.getElementById("message");
    if (newMessage != null){
        let userName = "Bob"    // TODO: hardcoded Username currently
        let outEvent = new SendMessageEvent(newMessage.value, userName)
        sendEvent("send_message", outEvent);
    }
}

// handles the login request
function handleLogin(ev) {
    ev.preventDefault();
    let formData = {
        "username": document.getElementById("username").value,
        "password": document.getElementById("password").value,
    }
    // send request to the /login api endpoint
    fetch("login", {
        method: "post",
        body: JSON.stringify(formData),
        mode: "cors",
    }).then((response) => {
        if(response.ok) return response.json();
        else throw 'unauthorized';
    }).then((data) => {
        connectWebsocket(data.otp);
    }).catch((err) => {alert(err)});
}

// handles (after successful login) opening the websocket.
function connectWebsocket(oneTimePassword) {
    // check if browser supports WebSocket
    if (!window["WebSocket"]){
        alert("Not supporting websockets, change your browser.");
        return;
    }
    wsConn = new WebSocket("ws://"+ document.location.host + "/ws?otp="+oneTimePassword);
    setupWsHandlers();
}


/*
*
*   Handlers for the Websocket events:
*
*/
function setupWsHandlers() {
    // gets triggered after connection gets accepted by the server
    wsConn.onopen = () => {
        document.getElementById("connection-header").innerHTML = "Logged in - active Websocket connection.";
    }
    
    // gets triggered after the connection has closed
    wsConn.onclose = (ev) => {
        document.getElementById("connection-header").innerHTML = "Not Logged in - Websocket connection closed.";
    }
    
    // gets triggered after receiving a message von the server
    wsConn.onmessage = (ev) => {
        const eventData = JSON.parse(ev.data);
        const event = Object.assign(new Event, eventData);
        routeEvent(event);
        //console.log("message received: " + ev.data);
    }
    
    // gets triggered on errors
    wsConn.onerror = (ev) => {
        console.log("error with the websocket: "+ ev)
    }
}


/*
*
*   Event class is used to wrap all messages
*   Go will be able to use the `same struct` to Deserialize it
*
*/

// Wrapper other Event Types get wrapped into. (into the payload)
class Event {
    constructor(type, payload) {
        this.type = type;
        this.payload = payload;
    }
}

// Message THIS Client sends to Server -> other Clients
class SendMessageEvent {
    constructor(message, from){
        this.message = message;
        this.from = from;
    }
}

// Message of OTHER Client that gets forwarded trough Server to This Client.
class NewMessageEvent {
    constructor(message, from, sent){
        this.message = message;
        this.from = from;
        this.sent = sent;
    }
}

// Event to switch the active Chatroom this Client takes part in:
class ChangeChatRoomEvent {
    constructor(name) {
        this.name = name;
    }
}

// function to send a message-type Event
// to the Server over Websocket -> Server (-> forwards to other Clients if needed)
function sendEvent(eventName, payload) {
    let newEvent = new Event(eventName, payload);
    wsConn.send(JSON.stringify(newEvent));
}

// RouteEvent is a proxy function that routes incoming events to the correct Handler
function routeEvent(ev) {
    if (ev.type === undefined) {
        alert("no type field was provided");
    }
    switch (ev.type) {
        case "send_message":
            const incMessageEvent = Object.assign(new NewMessageEvent, ev.payload)
            appendChatMessage(incMessageEvent)
            break;
        default:
            alert("unsupported message type");
            break;
    }
}

// We just append Messages to our Text Box
function appendChatMessage(incMessageEvent) {
    let date = new Date(incMessageEvent.sent);
    let formattedMsg = `${date.toLocaleString()}: ${incMessageEvent.message}`
    let $textarea = document.getElementById("chatmessages");
    $textarea.innerHTML += "\n" + formattedMsg;
    $textarea.scrollTop = $textarea.scrollHeight;
}