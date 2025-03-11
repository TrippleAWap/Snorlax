use warp::{Rejection, Reply};
use warp::http::Response;
use warp::hyper::body::Bytes;
use crate::server::websocket::FILTER_STRING;

pub async fn handle_update_filter(
    filter: Bytes,
) -> Result<impl Reply, Rejection> {
    let filter = std::str::from_utf8(&filter)
       .map_err(|e| {
            println!("Error: {}", e);
            warp::reject::not_found()
        })?;
    println!("Filter: {:?}", filter);
    let mut filter_string = FILTER_STRING.lock().unwrap();
    *filter_string = filter.to_string();
    Ok(Response::new("Filter updated successfully"))
}