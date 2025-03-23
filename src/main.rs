use std::env;
use crate::server::find_open_port;
use dotenv::dotenv;
use crate::webview::login;

mod cache;
mod webview;
mod server;
mod providers;

#[tokio::main]
async fn main() {
    dotenv().ok();
    env_logger::init();
    if env::var("AUTH_TOKEN").is_err() {
        log::error!("AUTH_TOKEN environment variable not set");
        // start login;
        login::run().await.expect("Error while running login");
    }
    let port = find_open_port(1900, 9999).await.expect("Couldn't open port");
    log::info!("Listening on port {}", port);
    tokio::spawn(async move {
        server::run(port).await.expect("Error while running websockets");
    });
    tokio::spawn(async move {
        cache::run().await.expect("Error while running cache");
    });
    tokio::spawn(async move {
        providers::run().await.expect("Error while running authorization");
    });
    // wait for server to start
    tokio::time::sleep(std::time::Duration::from_millis(50)).await;
    webview::run(port).await.expect("Error while running webview");
}
