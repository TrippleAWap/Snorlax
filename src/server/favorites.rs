use warp::{Rejection, Reply};
use warp::http::Response;
use tokio::sync::Mutex;
use once_cell::sync::Lazy;
use rusqlite::{params, Error};
use warp::hyper::body::Bytes;
use crate::cache::db::CONN;

pub static FAVORITE_FILTER: Lazy<Mutex<bool>> = Lazy::new(|| Mutex::new(false));
pub async fn handle_update_favorites(
    body: Bytes
) -> Result<impl Reply, Rejection> {
    let state = match std::str::from_utf8(&body) {
        Ok(state) => state,
        Err(_) => Err(warp::reject::not_found())?,
    };
    let mut favorite_filter = FAVORITE_FILTER.lock().await;
    if state == "true" {
        *favorite_filter = true;
    } else {
        *favorite_filter = false;
    };
    Ok(Response::new(format!("Favorites Filter updated to {} successfully", state)))
}
pub async fn handle_favorite_avatar(
    avatar_id: String
) -> Result<impl Reply, Rejection> {
    match favorite_avatar_id(avatar_id.clone()).await {
        Ok(_) => Ok(Response::new(format!("Avatar {} added to favorites", avatar_id))),
        Err(e) => Ok(Response::builder().status(404).body(format!("Avatar {} not found: {:?}", avatar_id, e)).unwrap())
    }
}

pub async fn handle_unfavorite_avatar(
    avatar_id: String
) -> Result<impl Reply, Rejection> {
    match unfavorite_avatar_id(avatar_id.clone()).await {
        Ok(_) => Ok(Response::new(format!("Avatar {} removed from favorites", avatar_id))),
        Err(_) => Ok(Response::builder().status(404).body(format!("Avatar {} not found", avatar_id)).unwrap())
    }
}

pub async fn unfavorite_avatar_id(avatar_id: String) -> Result<(), Error> {
    let conn = CONN.lock().await;

    conn.execute(
        "DELETE FROM avatar_favorites WHERE key = ?",
        params![avatar_id],
    )?;

    Ok(())
}

pub async fn favorite_avatar_id(avatar_id: String) -> Result<(), Error> {
    let conn = CONN.lock().await;

    let avatar_data = conn.query_row(
        "SELECT * FROM avatar_cache WHERE key = ? LIMIT 1",
        params![avatar_id],
        |row| {
            Ok((row.get::<_, i32>(0)?, row.get::<_, String>(1)?, row.get::<_, String>(2)?))
        })?;
    conn.execute(
        "INSERT INTO avatar_favorites (key, value) VALUES (?,?)",
        params![avatar_data.1, avatar_data.2],
    )?;

    Ok(())
}
