use crate::websockets::find_open_port;
use dotenv::dotenv;

mod cache;
mod webview;
mod websockets;
mod authorization;

#[tokio::main]
async fn main() {
    dotenv().ok();
    env_logger::init();
    let port = find_open_port(1900, 9999).await.expect("Couldn't open port");
    log::info!("Listening on port {}", port);
    tokio::spawn(async move {
        authorization::run().await.expect("Error while running authorization");
    });
    tokio::spawn(async move {
        cache::run().await.expect("Error while running cache");
    });
    tokio::spawn(async move {
        websockets::run(port).await.expect("Error while running websockets");
    });
    webview::run(port).await.expect("Error while running webview");
}
