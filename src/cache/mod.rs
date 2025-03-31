use std::collections::HashMap;
use log::debug;
use crate::cache::cache_windows_player::{get_avatar_ids, get_cache_path};
use crate::cache::db::init_db;
use crate::cache::scrape::{process_avatars, scrape, watch};

pub mod cache_windows_player;
pub mod db;
pub mod scrape;
pub mod download_avatars;
mod amplitude_cache;

pub async fn run() -> Result<(), Box<dyn std::error::Error>> {
    println!("Cache module running...");
    // resolve CACHE_PATH to absolute path
    println!("Cache path: {}", get_cache_path());
    // avatar id -> JSON data ( minimizing api calls )
    init_db("avatar_cache").await.expect("Error initializing avatar_cache database");
    // dir path -> avatar id ( removes the process of opening a handle and searching the file )
    init_db("dir_to_id").await.expect("Error initializing dir_to_id database");
    println!("Cache loaded successfully!");
    // TODO: implement in-memory sqlite db for Prismic, etc.
    // download_dbs().await?;
    // this processes all the cache at the beginning because the watcher isn't
    // optimized to handle large amounts of avatars in a short amount of time.
    //
    // the watcher is used to process avatars as they are added to the cache. ( small amount over long period of time )
    let path_to_id = scrape().await.expect("Error scraping data");
    //

    // all code below is old and wont work anymore due to cache encrytion.
    return Ok(());
    // TODO: this is slow and seems to be batched, figure out how to fix it. ( even if its a bit less optimized )
    let watcher = watch(get_cache_path())?;
    let (_tx, rx, _) = watcher;
    while let Ok(path) = rx.recv() {
        if !path.ends_with("__data") {
            continue;
        }
        if let Some(_) = path_to_id.get(path.to_str().unwrap()) {
            continue;
        }
        debug!("Scraping data file: {}", path.display());
        let ids = get_avatar_ids(path.to_str().unwrap(), true).await;
        let id = match ids.first() {
            Some(id) => id.to_string(),
            None => continue,
        };
        let mut map = HashMap::new();
        map.insert(path.to_str().unwrap().to_string(), id);
        tokio::spawn(async {
            process_avatars(std::env::var("AUTH_TOKEN").ok(), map).await.expect("Error processing data file in background");
        });
    };
    // this seems to be quicker but uses more CPU
    // loop {
    //     let _ = scrape().await.expect("Error scraping data");
    //     tokio::time::sleep(std::time::Duration::from_secs(1)).await;
    // }
    Ok(())
}

