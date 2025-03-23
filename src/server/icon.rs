use warp::{Rejection, Reply, http::Response};

pub async fn handle_icon() -> Result<impl Reply, Rejection> {
    let ico = include_bytes!("../static/icon.ico");

    let response = Response::builder()
        .header("Content-Type", "image/x-icon")
        .header("Cache-Control", "public, max-age=86400")
        .body(ico.to_vec())
        .map_err(|_| warp::reject::not_found())?;

    Ok(response)
}
