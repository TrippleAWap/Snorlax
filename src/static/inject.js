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
window.page = 0;
const PAGE_SIZE = 101;

const createSockets = async () => {
    const ws = new WebSocket("ws://127.0.0.1:" + window.port + "/ws")
    const fetchAvatarsSpecified = (start, size) => {
        ws.send(JSON.stringify({ "event": "fetch_avatars", "data": { "start": start, "end": start + size - 1 } }))
    }
    const fetchAvatars = () => {
        fetchAvatarsSpecified(window.page * PAGE_SIZE, PAGE_SIZE)
    }
    ws.onmessage = (event) => {
        const json = JSON.parse(event.data);
        if (!json.event || !json.data)
            return console.error("Expected event and data in message, got: " + JSON.stringify(event.data));
        const grid = document.querySelector("div[id='avatar_grid']");
        if (!grid)
            return console.error("Could not find avatar grid");
        switch (json.event) {
            case "new_avatars":
                const count = json.data.length;
                console.log(`Received ${count} new avatars`);
                if (window.page !== 0) {
                     console.log(`Fetching avatars for page ${window.page}`);
                     fetchAvatarsSpecified(window.page * PAGE_SIZE, PAGE_SIZE)
                }
                const entries = Array.from(grid.children);

                while (entries.length > PAGE_SIZE) {
                    entries.pop().remove();
                }

                json.data.forEach(data => {
                    grid.insertAdjacentHTML("afterbegin", `<div class="avatar">${data}</div>`);
                });

                while (grid.children.length > PAGE_SIZE) {
                    grid.lastChild.remove();
                }
                break;
            case "avatars":
                console.log(`Received ${json.data.length} avatars`);
                const newEntries = Math.min(PAGE_SIZE, json.data.length)
                json.data.slice(0, newEntries).forEach(html => {
                    grid.insertAdjacentHTML("afterbegin", html)
                });
                while (grid.children.length > PAGE_SIZE) {
                    grid.lastChild.remove();
                }
                break;
            default:
                console.error("Unknown event: " + json.event);
        }
    }
    ws.onclose = createSockets;
    while (ws.readyState !== WebSocket.OPEN) {
        await new Promise(r => setTimeout(r, 100))
    }
    console.log("Websocket connected");
    fetchAvatars();
}

document.addEventListener('DOMContentLoaded', () => {
    createSockets().catch(e => console.error(e));
});