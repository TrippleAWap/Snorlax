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
window.pages = 0;
const PAGE_SIZE = 100;

let lastFilter = "";
let currentFilter = "";

const fetchAvatarsSpecified = (ws, start, size) => {
    if (lastFilter !== currentFilter) {
        lastFilter = currentFilter;
        // reset page to 0;
        window.page = 0;
    }
    ws.send(JSON.stringify({ "event": "fetch_avatars", "data": { "start": start, "end": start + size } }))
}

const fetchAvatars = async (ws) => {
    return fetchAvatarsSpecified(ws, window.page * PAGE_SIZE, PAGE_SIZE)
}
let fav = false;
const toggleFavorites = async (state) => {
    fav = state;
    await fetch("/api/favorites", {
        method: "POST",
        body: state.toString()
    }).then(r => console.log(r.body));
}

const updatePageNumber = (current) => {
    current = Math.max(0, Math.min(current, window.pages));
    window.page = current;
    const pageNumber = document.querySelector("span[id='page_number']");
    if (pageNumber) {
        pageNumber.textContent = `Page ${current + 1} of ${window.pages + 1}`;
    }
}
const handlePagination = async (button, ws) => {
    if (button.disabled) {
        return;
    }
    const pagination = document.querySelectorAll("div[class='pagination']>a");
    pagination.forEach(b => b.disabled = true);
    const direction = button.dataset.direction;
    const current = window.page;
    const grid = document.querySelector("div[id='avatar_grid']");
    grid.innerHTML = "";
    switch (direction) {
        case "first":
            updatePageNumber(0);
            break;
        case "last":
            updatePageNumber(window.pages);
            break;
        case "prev":
            updatePageNumber(current - 1);
            break;
        case "next":
            updatePageNumber(current + 1);
            break;
    }
    await fetchAvatars(ws);
    pagination.forEach(b => b.disabled = false);
}
const createSockets = async () =>  {
    await toggleFavorites(false)
    const ws = new WebSocket("ws://127.0.0.1:" + window.port + "/ws")

    ws.onmessage = (event) => {
        const json = JSON.parse(event.data);
        if (!json.event || !json.data)
            return console.error("Expected event and data in message, got: " + JSON.stringify(event.data));
        const grid = document.querySelector("div[id='avatar_grid']");
        if (!grid)
            return console.error("Could not find avatar grid");
        switch (json.event) {
            case "avatars":
                const count_display = document.querySelector(`span[id='${fav ? "avatar_count_favorites" : "avatar_count_overview"}']`);
                console.log(`Received ${json.data.total_count} avatars`);
                count_display.textContent = json.data.total_count;
                window.pages = Math.floor(json.data.total_count / PAGE_SIZE);
                const htmlLength = grid.innerHTML.length;
                for (const html of json.data.avatars) {
                    grid.insertAdjacentHTML("afterbegin", html)
                }
                grid.innerHTML = grid.innerHTML.slice(0, grid.innerHTML.length - htmlLength);
                updatePageNumber(window.page)
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
    fetchAvatars(ws);
    setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
            fetchAvatars(ws);
        }
    }, 1000);
    return ws;
}

document.addEventListener('DOMContentLoaded', async () => {
    fetch("/login", { method: "GET" }).then(r => r.json()).then(data => {
        const profile_pic = document.querySelector("img[id='profile_pic']");
        profile_pic.src = data.userIcon;
        console.log(data.userIcon.toString(), data.userIcon)
        if (data.userIcon === "") {
            console.log("No profile picture found")
            profile_pic.parentElement.remove();
        }
        const username = document.querySelector("span[id='username']");
        username.textContent = data.displayName;
        console.log(data);
    }).catch(e => console.error(e));
    const ws = await createSockets().catch(e => console.error(e));
    if (!ws) {
        console.error("Could not connect to websocket");
        return;
    }
    const grid = document.querySelector("div[id='avatar_grid']");
    const searchBar = document.querySelector("input[class='search-input']");
    searchBar.addEventListener("keyup", () => {
        console.log('searching for', searchBar.value);
        currentFilter = searchBar.value;
        fetch("/api/filter", {method: "POST", body: searchBar.value});
        grid.childNodes.forEach((node) => {
            if (!node.textContent.toLowerCase().includes(searchBar.value.toLowerCase())) {
                node.remove()
            }
        })
        fetchAvatars(ws);
    })

    const pagination = document.querySelectorAll("div[class='pagination']>a");
    pagination.forEach(button => {
        button.addEventListener("click", () => {
            handlePagination(button, ws);
        });
    });

    const navbar = document.querySelectorAll("div[id='navbar']>a");
    navbar.forEach(link => {
        link.addEventListener("click", async () => {
            if (link.dataset.disabled === "true") {
                return;
            }
            const grid = document.querySelector("div[id='avatar_grid']");
            grid.innerHTML = "";
            const navbar = document.querySelectorAll("div[id='navbar']>a");
            navbar.forEach(l => l.dataset.disabled = "false");
            link.dataset.disabled = "true";
            switch (link.dataset.title) {
                case "Overview":
                    await toggleFavorites(false);
                    fetchAvatars(ws);
                    break;
                case "Favorites":
                    await toggleFavorites(true);
                    fetchAvatars(ws);
                    break;
                default:
                    console.error("Unknown link: " + link.dataset.title);
            }
            console.log("clicked on", link.dataset.title);
        });
    });
});