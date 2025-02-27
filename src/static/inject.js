window.post_message = (json) => {
    console.log("post_message called with message: " + json);
    const jsonParsed = JSON.parse(json);
    window.onPostMessageFuncs?.forEach(func => func(jsonParsed));
}
window.onPostMessage = (func) => {
    window.onPostMessageFuncs ??= [];
    window.onPostMessageFuncs.push(func);
}

console.log(window.port)
const BATCH_SIZE = 2000;
let start = 0;

const createSockets = async () => {
    const ws = new WebSocket("ws://127.0.0.1:" + window.port)
    const fetchAvatars = () => {
        ws.send(JSON.stringify({ "event": "fetch_avatars", "data": { "start": start, "end": start + BATCH_SIZE } }));
        start += BATCH_SIZE;
    }
    ws.onmessage = (event) => {
        const json = JSON.parse(event.data);
        if (!json.event || !json.data)
            return console.log("Expected event and data in message, got: " + JSON.stringify(event.data));
        switch (json.event) {
            case "avatars":
                const grid = document.querySelector("div[id='avatar_grid']");
                if (!grid)
                    return console.log("Could not find avatar grid");
                const newHTML = json.data;
                // prepend new avatars to the grid
                grid.innerHTML = newHTML + grid.innerHTML;
                break;
            default:
                console.log("Unknown event: " + json.event);
        }
    }
    ws.onclose = createSockets;
    while (ws.readyState !== WebSocket.OPEN) {
        await new Promise(r => setTimeout(r, 100))
    }
    console.log("Websocket connected");
    for (let i = 0; i < 10; i++) {
        fetchAvatars();
        await new Promise(r => setTimeout(r, 1000));
    }
}
(async () => {
    while (!document.querySelector("div[id='avatar_grid']")) {
        await new Promise(resolve => setTimeout(resolve, 1000));
    }
    createSockets();
})();