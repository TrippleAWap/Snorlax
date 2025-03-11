use std::env;
use reqwest::Client;
use warp::{Rejection, Reply};

pub async fn handle_avatar_equip(
    avatar_id: String
) -> Result<impl Reply, Rejection> {
    let client = Client::new();
    let url = format!("https://api.vrchat.cloud/api/1/avatars/{}/select", avatar_id);
    let response = client.put(&url)
        .header("User-Agent",   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:132.0) Gecko/20100101 Firefox/132.0")
        .header("Accept",       "*/*")
        .header("Accept-Language","en-CA,en-US;q=0.7,en;q=0.3")
        .header("Referer",      "https://vrchat.com/home?utm-source=hello-login")
        .header("DNT",          "1")
        .header("Priority",     "u=4")
        .header("Cookie", format!("auth={}", env::var("AUTH_TOKEN").unwrap()))
        .send()
        .await
        .map_err(|e| {
            println!("Error: {}", e);
            warp::reject::not_found()
        })?;

    if response.status() != reqwest::StatusCode::OK {
        println!("Error: {}", response.status());
        return Err(warp::reject::not_found());
    }

    Ok(warp::reply::with_status(response.text().await.unwrap(), warp::http::StatusCode::OK))
}