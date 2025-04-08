#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]


use std::env;
use std::env::args;
use std::process::exit;
use crate::server::find_open_port;
use dotenv::dotenv;
use tokio::fs;
use tokio::process::Command;
use crate::webview::login;

mod cache;
mod webview;
mod server;
mod providers;
mod vrchat_api;

#[tokio::main]
async fn main() {
    dotenv().ok();
    env_logger::init();
    if !args().any(|arg| arg == "--launch") {
        match auto_update().await {
            Ok(_) => {}
            Err(e) => {
                eprintln!("Auto-update failed: {}", e);
            }
        }
    }
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

const VERSION_URL: &str = "https://github.com/TrippleAWap/Snorlax/releases/latest";
async fn latest_version() -> Result<String, Box<dyn std::error::Error>> {
    let res = reqwest::get(VERSION_URL).await?;
    if !res.status().is_success() {
        return Err("Failed to get latest version".into())
    }
    let paths = res.url().path().split("/").collect::<Vec<&str>>();
    Ok(paths[paths.len()-1].to_string())
}


async fn auto_update() -> Result<bool, Box<dyn std::error::Error>> {
    #[cfg(debug_assertions)]
    {
        return Err("Auto-update not supported in debug mode".into())
    }
    let latest_version = latest_version().await?;
    let download_url = format!("https://github.com/TrippleAWap/Snorlax/releases/download/{}/update.exe", latest_version);
    let res = reqwest::get(&download_url).await?;
    if !res.status().is_success() {
        return Ok(false)
    }
    let update_exe = res.bytes().await?;
    // write to ./update.exe
    fs::write("./update.exe", update_exe).await?;
    // spawn process;
    let mut cmd = Command::new("./update.exe");
    cmd.arg(args().nth(0).unwrap());
    cmd.spawn().expect("Failed to spawn update process");
    exit(0)
}