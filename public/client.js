if (!window["WebSocket"]){
    alert("Not supporting websockets, change your browser.");
}

var selectedChat = "general"
var wsConn = new WebSocket("ws://"+ document.location.host + "/ws");

// setup the onSubmit callbacks
document.getElementById("form-selection").addEventListener("submit", handleRoomSelection)
document.getElementById("form-message").addEventListener("submit", handleSendMessage)


/*
*   Button/Submit Handlers:
*/

// handles the onSubmit for Room Selection
function handleRoomSelection(ev) {
    ev.preventDefault();
    let newChat = document.getElementById("chatroom");
    if (newChat != null && newChat.value != selectedChat){
        console.log("DEBUG: changing to" + newChat.value)
    }
}

// handles the onSubmit for Message Sending
function handleSendMessage(ev) {
    ev.preventDefault();
    let newMessage = document.getElementById("message");
    if (newMessage != null){
        sendEvent("send_message", newMessage.value);
    }
}


/*
*   Handlers for the Websocket events:
*/

// gets triggered after connection gets accepted by the server
// wsConn.onopen = () => {
//     console.log("connected to: "+ url);
// }

// gets triggered after the connection has closed
wsConn.onclose = (ev) => {
    console.log("connection close with: " + ev.code);
}

// gets triggered after receiving a message von the server
wsConn.onmessage = (ev) => {
    console.log(ev);
    const eventData = JSON.parse(evt.data);
    const event = Object.assign(new Event, eventData);
    routeEvent(event);
    console.log("message received: " + ev.data);
}

// gets triggered on errors
wsConn.onerror = (ev) => {
    console.log("error with the websocket: "+ ev)
}


/*
*   Event class is used to wrap all messages
*   Go will be able to use the `same struct` to Deserialize it
*/
class Event {
    constructor(type, payload) {
        this.type = type;
        this.payload = payload;
    }
}

function sendEvent(eventName, payload) {
    const event = new Event(eventName, payload);
    wsConn.send(JSON.stringify(event));
}


// RouteEvent is a proxy function that routes events to the correct Handler
function routeEvent(ev) {
    if (ev.type === undefined) {
        alert("no type field in the event");
    }
    switch (ev.type) {
        case "new_message":
            console.log("new message");
            break;
        default:
            alert("unsupported message type");
            break;
    }
}