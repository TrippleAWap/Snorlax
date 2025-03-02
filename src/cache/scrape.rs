use std::cmp::{max, min};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::Mutex;
use log::debug;
use crate::cache::cache_windows_player::{cache_ids, get_avatar_ids, get_cache_path, get_cached_ids, walk_dir};
use crate::cache::download_avatars::download_avatars;

pub const SCRAPING_THREADS: usize = 10; // maximum number of threads used to scrape files in cache directory.
pub const MIN_PER_THREAD: usize = 5; // minimum number of files a thread is allowed to scrape.

pub async fn scrape() -> Result<Vec<String>, rusqlite::Error> {
    let ids = scrape_avatar_ids().await?;
    debug!("Scraped {} avatar ids", ids.len());
    debug!("Downloading {} ids...", ids.len());
    download_avatars(&ids, "authcookie_55abeec5-4bd9-4eaf-8db4-c5d680450c06").await?;
    Ok(ids)
}

async fn scrape_files() -> Result<Vec<String>, rusqlite::Error> {
    debug!("Scraping avatar_ids...");
    let paths = walk_dir(&get_cache_path(), true, 99, |path| {
        path.ends_with("\\__data")
    }, Some(|_, _| {
        None
    })).await;
    debug!("Scraped files, found {} paths", paths.len());
    Ok(paths)
}

async fn scrape_avatar_ids_(files: Vec<String>) -> Result<Vec<String>, rusqlite::Error> {
    let mut ids = vec![];
    let mut new_ids = HashMap::new();
    for file in files {
        let ids_ = get_avatar_ids(&file, true).await;
        new_ids.insert(file.clone(), ids_.clone().iter().next().unwrap_or(&"".to_string()).to_string());
        ids.extend(ids_);
    }
    debug!("Scraped {} new avatar ids", ids.len());
    debug!("Caching {} new ids...", new_ids.len());
    cache_ids(&new_ids).expect("Failed to cache ids");
    debug!("Cached {} new ids", new_ids.len());
    Ok(ids)
}

async fn scrape_avatar_ids() -> Result<Vec<String>, rusqlite::Error> {
    let paths = scrape_files().await?;
    let ids = Arc::new(Mutex::new(vec![]));
    let cached_ids = get_cached_ids()?;
    let mut all_cached_ids = vec![];
    let mut paths = paths.into_iter().filter(|path| {
        let id = cached_ids.get(path);
        if !id.is_none() {
            all_cached_ids.push(id.unwrap().to_string());
            return false;
        };
        true
    }).collect::<Vec<_>>();
    debug!("Scraped {} cached ids", all_cached_ids.len());
    ids.lock().await.extend(all_cached_ids);
    debug!("Scraped {} new paths", paths.len());
    debug!("Scraping avatar_ids from paths...");
    let mut threads = vec![];
    while paths.len() > 0 {
        let target_paths = min(paths.len(), max(MIN_PER_THREAD, (paths.len() as f32 / SCRAPING_THREADS as f32).ceil() as usize));
        let paths_clone = paths.clone();
        let ids_clone = Arc::clone(&ids);
        threads.push(tokio::spawn(async move {
            debug!("Scraping {} paths...", target_paths);
            let ids_ = scrape_avatar_ids_(paths_clone[..target_paths].to_vec()).await.expect("Failed to scrape avatar ids");
            let mut ids = ids_clone.lock().await;
            ids.extend(ids_);
        }));
        paths.drain(..target_paths);
    }
    debug!("Waiting for threads to finish...");
    for thread in threads {
        thread.await.expect("Failed to join thread");
    }
    let ids = ids.lock().await;
    debug!("Scraped {} avatar ids", ids.len());
    Ok(ids.to_vec())
}
