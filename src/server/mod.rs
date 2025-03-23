mod websocket;
mod equip;
mod thumbnail;
mod login;
mod filter;
mod favorites;
mod logout;
mod icon;

use std::collections::HashMap;
use rusqlite::{params, Connection};
use std::error::Error;
use log::info;
use tokio::net::TcpListener;
use warp::Filter;

pub async fn find_open_port(start_port: u16, end_port: u16) -> Result<u16, Box<dyn Error>> {
    for port in start_port..=end_port {
        let addr = format!("127.0.0.1:{}", port);
        match TcpListener::bind(&addr).await {
            Ok(_) => {
                return Ok(port);
            }
            Err(_) => {
                continue;
            }
        }
    }
    Err("No open port found in the specified range".into())
}

pub async fn run(port: u16) -> Result<(), Box<dyn Error>> {
    let ws_route = warp::path("ws")
        .and(warp::ws())
        .map(|ws: warp::ws::Ws| {
            ws.on_upgrade(websocket::process_ws)
        });

    let filter_update_route = warp::path!("api" / "filter")
        .and(warp::post())
        .and(warp::body::bytes())
        .and_then(filter::handle_update_filter);
    let filter_favorites_route = warp::path!("api" / "favorites")
        .and(warp::post())
        .and(warp::body::bytes())
        .and_then(favorites::handle_update_favorites);

    let favorite_route = warp::path!("api" / "favorite" / String)
        .and(warp::post())
        .and_then(favorites::handle_favorite_avatar);
    let unfavorite_route = warp::path!("api" / "favorite" / String)
        .and(warp::delete())
        .and_then(favorites::handle_unfavorite_avatar);

    let thumbnail_route = warp::path!("thumbnail")
        .and(warp::query::<HashMap<String, String>>())
        .and_then(thumbnail::handle_thumbnail);
    
    let equip_route = warp::path!("api" / "avatar" / "equip" / String)
        .and_then(equip::handle_avatar_equip);
    
    let home_route = warp::path("home").map(|| {
        warp::reply::html(include_str!("../static/index.html"))
    });

    let login_route = warp::path!("login")
        .and_then(login::handle_login);
    let logout_route = warp::path!("logout")
        .and_then(logout::handle_logout);

    let icon_route = warp::path!("icon.ico")
        .and_then(icon::handle_icon);

    let routes = ws_route
        .or(thumbnail_route)
        .or(home_route)
        .or(equip_route)
        .or(login_route)
        .or(logout_route)
        .or(filter_update_route)
        .or(filter_favorites_route)
        .or(favorite_route)
        .or(unfavorite_route)
        .or(icon_route);

    let addr = ([127, 0, 0, 1], port);
    println!("Server running on http://127.0.0.1:{}", addr.1);

    warp::serve(routes).run(addr).await;
    Ok(())
}
const CARD_HTML : &str = include_str!("../static/card.html");

pub(crate) fn map_row(row: &rusqlite::Row) -> rusqlite::Result<(i64, String, String)> {
    Ok((row.get(0)?, row.get(1)?, row.get(2)?))
}

pub fn query_avatar_cache(conn: &Connection, filter: Option<String>, favorites_only: bool) -> Result<Vec<(i64, String, String)>, Box<dyn Error>> {
    let base_sql = "SELECT id, key, value FROM ".to_owned() + if favorites_only { "avatar_favorites" } else { "avatar_cache" };
    let sql = if filter.is_some() {
        info!("filter: {}", filter.as_ref().unwrap());
        format!("{} WHERE value LIKE ? ORDER BY id DESC", base_sql)
    } else {
        format!("{} ORDER BY id DESC", base_sql)
    };

    let mut stmt = conn.prepare(&sql)?;

    let avatar_iter = if let Some(filter_str) = filter {
        let like_value = format!("%{}%", filter_str);
        stmt.query_map(params![like_value], map_row)?
    } else {
        stmt.query_map(params![], map_row)?
    };

    let mut avatars = Vec::new();
    for avatar in avatar_iter {
        avatars.push(avatar?);
    }

    Ok(avatars)
}



