use std::cmp::{max, min};
use std::sync::Arc;
use log::debug;
use reqwest::Client;
use rusqlite::params;
use serde::{Deserialize, Serialize};
use serde_json;
use tokio::sync::Mutex;
use crate::cache::db::CONN;
use crate::cache::scrape::{MIN_PER_THREAD, SCRAPING_THREADS};

#[derive(Serialize, Deserialize, Debug, Default, Clone)]
pub struct UnityPackage {
    asset_version: u32,
    created_at: String,
    id: String,
    performance_rating: String,
    platform: String,
    scan_status: String,
    unity_version: String,
    variant: String,
    impostorizer_version: Option<String>,
}

#[derive(Serialize, Deserialize, Debug, Default, Clone)]
pub struct Avatar {
    author_id: String,
    author_name: String,
    created_at: String,
    description: String,
    featured: bool,
    id: String,
    image_url: String,
    name: String,
    release_status: String,
    styles: Styles,
    tags: Vec<String>,
    thumbnail_image_url: String,
    unity_package_url: String,
    unity_package_url_object: serde_json::Value,
    unity_packages: Vec<UnityPackage>,
    updated_at: String,
    version: u32,
}

#[derive(Serialize, Deserialize, Debug, Default, Clone)]
pub struct Styles {
    primary: Option<String>,
    secondary: Option<String>,
}

async fn download_avatar(avatar_id: &str, tkn: &str) -> Result<Avatar, reqwest::Error> {
    let url = format!("https://vrchat.com/api/1/avatars/{}", avatar_id);
    let client = Client::new();
    let response = client.get(&url)
        .header("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0")
        .header("Cookie", format!("auth={}", tkn))
        .header("Accept", "*/*")
        .send()
        .await?;
    Ok(response.json::<Avatar>().await?)
}

pub async fn download_avatars(avatar_ids: &Vec<String>, tkn: &str) -> Result<Vec<Avatar>, rusqlite::Error> {
    let mut avatar_ids = avatar_ids.clone();
    let data = Arc::new(Mutex::new(vec![]));
    let mut threads = Vec::new();
    while avatar_ids.len() > 0 {
        let target_ids = min(avatar_ids.len(), max(MIN_PER_THREAD, (avatar_ids.len() as f32 / SCRAPING_THREADS as f32).ceil() as usize));
        let paths_clone = avatar_ids.clone();
        let data_clone = Arc::clone(&data);
        let tkn_clone = tkn.to_string();
        threads.push(tokio::spawn(async move {
            debug!("Downloading {} avatars...", target_ids);
            let avatars_data = download_avatars_(&paths_clone[..target_ids].to_vec(), &tkn_clone).await.expect("Failed to scrape avatar ids");
            let mut data = data_clone.lock().await;
            data.extend(avatars_data);
        }));
        avatar_ids.drain(..target_ids);
    }
    for thread in threads {
        thread.await.expect("Failed to join thread");
    }
    let data = data.lock().await;
    let conn = CONN.lock().unwrap();
    let mut stmt = conn.prepare("INSERT OR IGNORE INTO avatar_cache (key, value) VALUES (?,?)")?;
    debug!("Inserting {} avatars into cache...", data.len());
    for i in 0..data.len() {
        let avatar = &data[i];
        let json = serde_json::to_string(avatar).unwrap();
        stmt.execute(params![avatar.id, json])?;
    }
    debug!("Done inserting avatars into cache.");
    Ok(data.to_vec())
}

async fn download_avatars_(avatar_ids: &Vec<String>, tkn: &str) -> Result<Vec<Avatar>, rusqlite::Error> {
    let mut result = Vec::new();
    for avatar_id in avatar_ids {
        match download_avatar(avatar_id, tkn).await {
            Ok(avatar) => {
                result.push(avatar);
            }
            Err(e) => {
                result.push(Avatar::default());
                println!("Error downloading avatar {}: {}", avatar_id, e);
            }
        }
    }
    Ok(result)
}