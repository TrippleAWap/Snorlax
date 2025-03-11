use std::env;
use reqwest::Client;
    use warp::{Rejection, Reply};

pub async fn handle_login() -> Result<impl Reply, Rejection> {
    let client = Client::new();
    let url = "https://api.vrchat.cloud/api/1/auth/user";
    let response = client.get(url)
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

    let status = response.status();
    let content = match response.text().await {
        Ok(content) => content,
        Err(e) => {
            println!("Error: {}", e);
            return Err(warp::reject::not_found());
        }
    };
    if status != 200 {
        println!("Error: {}", content);
        return Err(warp::reject::not_found());
    }

    Ok(warp::reply::with_status(content, warp::http::StatusCode::OK))
}
