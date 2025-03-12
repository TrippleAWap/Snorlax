use std::cmp::{max, min};
use std::collections::HashMap;
use std::sync::Arc;
use log::{info, warn};
use reqwest::Client;
use rusqlite::params;
use serde_json;
use tokio::sync::Mutex;
use vrchatapi::apis::avatars_api::GetAvatarError;
use vrchatapi::apis::{Error, ResponseContent};
use crate::cache::cache_windows_player::cache_avatars;
use crate::cache::db::CONN;
use crate::cache::scrape::{MIN_PER_THREAD, SCRAPING_THREADS};


async fn download_avatar(avatar_id: &str, tkn: &str) -> Result<vrchatapi::models::Avatar, Error<GetAvatarError>> {
    let url = format!("https://vrchat.com/api/1/avatars/{}", avatar_id);
    let client = Client::new();
    let response = client.get(&url)
        .header("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0")
        .header("Cookie", format!("auth={}", tkn))
        .header("Accept", "*/*")
        .send()
        .await?;
    let status = response.status();
    let content = response.text().await?;

    if !status.is_client_error() && !status.is_server_error() {
        serde_json::from_str(&content).map_err(Error::from)
    } else {
        let entity: Option<GetAvatarError> =
            serde_json::from_str(&content).ok();
        let local_var_error = ResponseContent {
            status,
            content,
            entity,
        };
        Err(Error::ResponseError(local_var_error))
    }
}

pub async fn download_avatars(avatar_ids: &Vec<String>, tkn: &str) -> Result<Vec<vrchatapi::models::Avatar>, rusqlite::Error> {
    let mut avatar_ids = avatar_ids.clone();
    let data = Arc::new(Mutex::new(vec![]));
    let mut threads = Vec::new();
    while avatar_ids.len() > 0 {
        let target_ids = min(avatar_ids.len(), max(MIN_PER_THREAD, (avatar_ids.len() as f32 / SCRAPING_THREADS as f32).ceil() as usize));
        let paths_clone = avatar_ids.clone();
        let data_clone = Arc::clone(&data);
        let tkn_clone = tkn.to_string();
        threads.push(tokio::spawn(async move {
            info!("Downloading {} avatars...", target_ids);
            let avatars_data = download_avatars_(&paths_clone[..target_ids].to_vec(), &tkn_clone).await.expect("Failed to scrape avatar ids");
            let mut data = data_clone.lock().await;
            cache_avatars(&HashMap::from_iter(avatars_data.clone().into_iter().map(|avatar| {
                let json_str = serde_json::to_string(&avatar).unwrap();
                (avatar.id.to_string(), json_str)
            }))).expect("Failed to cache avatars");
            data.extend(avatars_data);
        }));
        avatar_ids.drain(..target_ids);
    }
    for thread in threads {
        thread.await.expect("Failed to join thread");
    }
    let data = data.lock().await;
    Ok(data.to_vec())
}

async fn download_avatars_(avatar_ids: &Vec<String>, tkn: &str) -> Result<Vec<vrchatapi::models::Avatar>, rusqlite::Error> {
    let mut result = Vec::new();
    for avatar_id in avatar_ids {
        match download_avatar(avatar_id, tkn).await {
            Ok(avatar) => {
                result.push(avatar);
            }
            Err(e) => {
                warn!("Failed to download avatar {} with error: {}", avatar_id, e)
            }
        }
    }
    Ok(result)
}