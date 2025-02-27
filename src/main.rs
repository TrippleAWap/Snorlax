use crate::websockets::find_open_port;

mod cache;
mod webview;
mod websockets;

#[tokio::main]
async fn main() {
    let port = find_open_port(1900, 9999).await.expect("Couldn't open port");
    println!("Listening on port {}", port);

    tokio::spawn(async move {
        cache::run().await.expect("Error while running cache");
    });
    tokio::spawn(async move {
        websockets::run(port).await.expect("Error while running webview");
    });
    webview::run(port).await.expect("Error while running webview");
}
