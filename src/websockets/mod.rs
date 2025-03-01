use futures_util::{SinkExt, StreamExt};
use std::error::Error;
use tokio::net::TcpListener;
use tokio_tungstenite::accept_async;
use tokio_tungstenite::tungstenite::protocol::Message;
use tokio_tungstenite::tungstenite::Utf8Bytes;

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
    let addr = "127.0.0.1:".to_string() + &port.to_string();
    let listener = TcpListener::bind(&addr).await.expect("Failed to bind");
    println!("Listening on: {}", addr);

    while let Ok((stream, _)) = listener.accept().await {
        tokio::spawn(async move {
            if let Err(e) = handle_connection(stream).await {
                eprintln!("Error handling connection: {}", e);
            } else {
                println!("Connection closed");
            }
        });
    }
    Ok(())
}
const CARD_HTML : &str = include_str!("../static/card.html");
async fn handle_connection(stream: tokio::net::TcpStream) -> Result<(), Box<dyn Error>> {
    let ws_stream = accept_async(stream).await.expect("Error during the websocket handshake occurred");
    let (mut write, mut read) = ws_stream.split();

    while let Some(msg) = read.next().await {
        let msg = msg.expect("Error receiving message");
        if msg.is_text() {
            let json_text = msg.into_text().expect("Error converting message to text");
            println!("Received: {}", &json_text);
            let json: serde_json::Value = serde_json::from_str(&json_text).expect("Error parsing JSON");
            let response: serde_json::Value;
            match json.get("event").expect("Error getting event").as_str().expect("Error") {
                "fetch_avatars" => {
                    let data = json.get("data").expect("Error").as_object().expect("Error");
                    let start = data.get("start").expect("Error").as_u64().expect("Error");
                    let end = data.get("end").expect("Error").as_u64().expect("Error");
                    response = serde_json::json!({
                        "event": "avatars",
                        "data": vec![CARD_HTML.to_string(); (end - start) as usize]
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
            let response_msg = Message::Text(Utf8Bytes::from(response_text));

            write.send(response_msg.clone()).await.expect("Error sending message");
        }
    }

    Ok(())
}