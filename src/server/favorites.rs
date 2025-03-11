use warp::{Rejection, Reply};
use warp::http::Response;
use std::sync::Mutex;
use once_cell::sync::Lazy;
use std::collections::HashMap;
pub static FAVORITE_FILTER: Lazy<Mutex<bool>> = Lazy::new(|| Mutex::new(false));
pub async fn handle_update_favorites(
    body: HashMap<String, String>,
) -> Result<impl Reply, Rejection> {
    // TODO: this doesnt work, always errors out.
    let state = match body.get("state") {
        Some(state) => state,
        None => Err(warp::reject::not_found())?,
    };
    
    let mut favorite_filter = FAVORITE_FILTER.lock().unwrap();
    if state == "true" {
        *favorite_filter = true;
    } else {
        *favorite_filter = false;
    };
    Ok(Response::new(format!("Favorites Filter updated to {} successfully", state)))
}