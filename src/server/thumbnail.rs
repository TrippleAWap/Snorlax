use std::collections::HashMap;
use reqwest::Client;
use warp::{Rejection, Reply};
use warp::http::Response;

pub async fn handle_thumbnail(
    p: HashMap<String, String>
) -> Result<impl Reply, Rejection> {
    // Build the URL where the image is hosted.
    // Fetch the image.
    let client = Client::new();

    let response = client.get(p.get("url").unwrap())
        .header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
        .header("Accept", "image/png,image/*;q=0.8,*/*;q=0.5")
        .header("Accept-Language", "en-US,en;q=0.5")
        .header("Accept-Encoding", "gzip, deflate, br")
        .header("Connection", "keep-alive")
        .header("Upgrade-Insecure-Requests", "1")
        .send()
        .await
        .map_err(|_| warp::reject::reject())?;

    // Ensure that the status code is OK.
    if response.status() != reqwest::StatusCode::OK {
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
        .map_err(|_| warp::reject::reject())?;

    Ok(response)
}