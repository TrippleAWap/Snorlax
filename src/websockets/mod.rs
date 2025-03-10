use futures_util::{SinkExt, StreamExt};
use std::error::Error;
use reqwest::Client;
use rusqlite::{params, Connection};
use tokio::net::{TcpListener};
use crate::cache::db::CONN;
use warp::{Filter, Rejection, Reply};
use warp::http::Response;

pub async fn find_open_port(start_port: u16, end_port: u16) -> Result<u16, Box<dyn Error>> {
    for port in start_port..=end_port {
        let addr = format!("127.0.0.1:{}", port);
        match TcpListener::bind(&addr).await {
            Ok(_) => {
                return Ok(port);
            }
            Err(_) => {
                continue;
            }
        }
    }
    Err("No open port found in the specified range".into())
}

pub async fn run(port: u16) -> Result<(), Box<dyn Error>> {
    let ws_route = warp::path("ws")
        .and(warp::ws())
        .map(|ws: warp::ws::Ws| {
            ws.on_upgrade(process_ws)
        });
    let thumbnail_route = warp::path!("thumbnail" / String / u32 / u32)
        .and_then(handle_thumbnail);
    let home_route = warp::path("home").map(|| {
        warp::reply::html(include_str!("../static/index.html"))
    });
    let routes = ws_route.or(thumbnail_route).or(home_route);

    let addr = ([127, 0, 0, 1], port);
    println!("Server running on http://{}:{}", addr.0[0], addr.1);

    warp::serve(routes).run(addr).await;
    Ok(())
}
const CARD_HTML : &str = include_str!("../static/card.html");


pub fn query_avatar_cache(conn: &Connection, offset: i64, limit: i64) -> Result<Vec<(i64, String, String)>, Box<dyn Error>> {
    let mut stmt = conn.prepare(
        "SELECT id, key, value
         FROM avatar_cache
         ORDER BY id DESC
         LIMIT ? OFFSET ?",
    )?;

    // Map each row into a tuple.
    let avatar_iter = stmt.query_map(params![limit, offset], |row| {
        Ok((row.get(0)?, row.get(1)?, row.get(2)?))
    })?;

    let mut avatars = Vec::new();
    for avatar in avatar_iter {
        avatars.push(avatar?);
    }

    Ok(avatars)
}

async fn handle_thumbnail(
    key: String,
    id: u32,
    resolution: u32,
) -> Result<impl Reply, Rejection> {
    // Build the URL where the image is hosted.
    let url = format!(
        "https://api.vrchat.cloud/api/1/image/{}/{}/{}",
        key, id, resolution
    );

    // Fetch the image.
    let client = Client::new();

    let response = client.get(&url)
        .header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
        .header("Accept", "image/png,image/*;q=0.8,*/*;q=0.5")
        .header("Accept-Language", "en-US,en;q=0.5")
        .header("Accept-Encoding", "gzip, deflate, br")
        .header("Connection", "keep-alive")
        .header("Upgrade-Insecure-Requests", "1")
        .send()
        .await
        .map_err(|e| {
            println!("Error: {}", e);
            warp::reject::not_found()
        })?;

    // Ensure that the status code is OK.
    if response.status() != reqwest::StatusCode::OK {
        println!("Error: {}", response.status());
        return Err(warp::reject::not_found());
    }

    // Retrieve the response bytes.
    let bytes = response
        .bytes()
        .await
        .map_err(|_| warp::reject::not_found())?;

    let response = Response::builder()
        .header("Content-Type", "image/png")
        .header("Cache-Control", "max-age=31536000")
        .body(bytes)
        .map_err(|e| {
            println!("Error: {}", e);
            warp::reject::not_found()
        })?;

    Ok(response)
}

async fn process_ws(ws: warp::ws::WebSocket)  {
    println!("New connection");
    let (mut write, mut read) = ws.split();
    while let Some(msg) = read.next().await {
        let msg = msg.expect("Error receiving message");
        if msg.is_text() {
            let json_text = msg.to_str().expect("Error converting message to text");
            println!("Received: {}", &json_text);
            let json: serde_json::Value = serde_json::from_str(&json_text).expect("Error parsing JSON");
            let response: serde_json::Value;
            match json.get("event").expect("Error getting event").as_str().expect("Error") {
                "fetch_avatars" => {
                    let data = json.get("data").expect("Error").as_object().expect("Error");
                    let start = data.get("start").expect("Error").as_u64().expect("Error");
                    let end = data.get("end").expect("Error").as_u64().expect("Error");

                    let conn = CONN.lock().unwrap();
                    let avatars = query_avatar_cache(&conn, start as i64, end as i64).expect("Error querying avatar cache");
                    let avatar_data = avatars.iter().map(|(id, _avatar_key, avatar_value)| {
                        // Parse the JSON string from the cached avatar.
                        let avatar: vrchatapi::models::Avatar = serde_json::from_str(avatar_value)
                            .expect("Error parsing avatar JSON");

                        let mut card_html = CARD_HTML.to_string();
                        // Adjust this slice as needed
                        card_html = card_html.replace("{{thumbnail_url}}", &format!("/thumbnail/{}", &avatar.thumbnail_image_url[37..]));
                        card_html = card_html.replace("{{author_name}}", &avatar.author_name);
                        card_html = card_html.replace("{{avatar_name}}", &avatar.name);
                        card_html = card_html.replace("{{avatar_id}}", &id.to_string());
                        card_html
                    }).collect::<Vec<String>>();

                    response = serde_json::json!({
                        "event": "avatars",
                        "data": avatar_data
                    });
                }
                _ => {
                    println!("Unknown event: {}", json.get("event").expect("Error getting event"));
                    response = serde_json::json!({
                        "event": "error",
                        "data": "Unknown event"
                    });
                }
            }
            let response_text = serde_json::to_string(&response).expect("Error converting JSON to text");
            let response_msg = warp::ws::Message::text(response_text);
            write.send(response_msg.clone()).await.expect("Error sending message");
        }
    }
}
