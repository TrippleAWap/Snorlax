use std::env::args;
use std::process::{exit, Command};
use tokio::fs::write;
use warp::{Rejection, Reply};

pub async fn handle_logout() -> Result<impl Reply, Rejection> {
    match write(".env", "").await {
        Ok(_) => {}
        Err(e) => {
            return Ok(warp::reply::with_status(format!("Failed to write to .env: {}", e), warp::http::StatusCode::NOT_FOUND))
        }
    }
    match Command::new(args().next().unwrap())
        .spawn() {
        Ok(child) => child,
        Err(e) => {
            eprintln!("Failed to spawn new_exe: {}", e);
            exit(1);
        }
    };
    exit(0);
}
