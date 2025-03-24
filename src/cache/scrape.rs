use std::cmp::{max, min};
use std::collections::HashMap;
use std::path::{Path, PathBuf};
use std::sync::Arc;
use tokio::sync::Mutex;
use log::info;
use rusqlite::params;
use crate::cache::cache_windows_player::{cache_ids, get_avatar_ids, get_cache_path, get_cached_ids, walk_dir, AVATAR_REGEX};
use crate::cache::db::CONN;
use crate::cache::download_avatars::download_avatars;
use notify::{Config, Event, PollWatcher, RecursiveMode, Watcher};
use crossbeam::channel::{Receiver, Sender};

pub const SCRAPING_THREADS: usize = 10; // maximum number of threads used to scrape files in cache directory.
pub const MIN_PER_THREAD: usize = 25; // minimum number of files a thread is allowed to scrape.

pub async fn scrape() -> Result<HashMap<String, String>, rusqlite::Error> {
    let ids_quick = scrape_local_avatar_data(&get_cache_path()).await.expect("Error scraping local avatar data");
    info!("Scraped {} avatar ids ( quick )", ids_quick.len());
    let mut ids = scrape_avatar_ids().await?;
    info!("Scraped {} avatar ids ( slow )", ids.len());
    // combine ids;
    ids.extend(ids_quick);
    info!("Scraped {} avatar ids", ids.len());
    process_avatars(std::env::var("AUTH_TOKEN").ok(), ids.clone()).await?;
    Ok(ids)
}

async fn scrape_local_avatar_data(path: &str) -> Result<HashMap<String, String>, Box<dyn std::error::Error>> {
    let mut ids = HashMap::new();
    let paths = walk_dir(path, true, 4, |path| {
        path.contains("avtr_")
    }, None).await;
    for path in paths {
        let avatar_id = match AVATAR_REGEX.captures(path.as_str()) {
            Some(caps) => caps.get(0).unwrap().as_str().to_string(),
            None => continue,
        };
        if ids.contains_key(&avatar_id) {
            continue;
        }
        println!("Scraping {}", avatar_id);
        ids.insert(avatar_id.clone(), avatar_id);
    }
    Ok(ids)
}
pub async fn process_avatars(auth_cookie: Option<String>, ids: HashMap<String, String>) -> Result<HashMap<String, String>, rusqlite::Error> {
    if let Some(cookie) = auth_cookie {
        info!("Using auth cookie: {}", cookie);
        let avatar_ids = {
            let mut ret = vec![];
            let conn = CONN.lock().await;
            let mut stmt = conn.prepare("SELECT 1 FROM avatar_cache WHERE key = ? LIMIT 1")?;
            info!("Checking {} cached ids...", ids.len());
            for (_, id) in ids.iter() {
                if stmt.exists(params![id])? {
                    continue;
                }
                ret.push(id.to_string());
            }
            info!("Downloading {} ids...", ret.len());
            ret
        };
        // insert all into cache;

        download_avatars(&avatar_ids, &cookie).await?;
    } else {
        info!("No auth cookie provided, skipping download");
    }
    Ok(ids)
}

async fn scrape_paths() -> Result<Vec<String>, rusqlite::Error> {
    info!("Scraping avatar_ids...");
    let paths = walk_dir(&get_cache_path(), true, 99, |path| {
        path.ends_with("\\__data")
    }, Some(|_, _| {
        None
    })).await;
    info!("Scraped files, found {} paths", paths.len());
    Ok(paths)
}

async fn scrape_avatar_ids_(paths: Vec<String>) -> Result<HashMap<String, String>, rusqlite::Error> {
    let mut ids = HashMap::new();
    let mut new_ids = HashMap::new();
    for path in paths {
        let ids_ = get_avatar_ids(&path, true).await;
        new_ids.insert(path.clone(), ids_.clone().iter().next().unwrap_or(&"".to_string()).to_string());
        ids.insert(path.clone(), ids_.iter().next().unwrap_or(&"".to_string()).to_owned());
    }
    info!("Scraped {} new avatar ids", ids.len());
    info!("Caching {} new ids...", new_ids.len());
    cache_ids(&new_ids).await.expect("Failed to cache ids");
    info!("Cached {} new ids", new_ids.len());
    Ok(ids)
}

async fn scrape_avatar_ids() -> Result<HashMap<String, String>, rusqlite::Error> {
    let paths = scrape_paths().await?;
    let ids: Arc<Mutex<HashMap<String, String>>> = Arc::new(Mutex::new(HashMap::new()));
    let cached_ids = get_cached_ids().await?;
    let mut all_cached_ids: HashMap<String, String> = HashMap::new();
    let mut paths = paths.into_iter().filter(|path| {
        let id = cached_ids.get(path);
        if !id.is_none() {
            all_cached_ids.insert(path.clone(), id.unwrap().0.to_string());
            return false;
        };
        true
    }).collect::<Vec<_>>();
    info!("Scraped {} cached ids", all_cached_ids.len());
    ids.lock().await.extend(all_cached_ids);
    info!("Scraped {} new paths", paths.len());
    info!("Scraping avatar_ids from paths...");
    let mut threads = vec![];
    while paths.len() > 0 {
        let target_paths = min(paths.len(), max(MIN_PER_THREAD, (paths.len() as f32 / SCRAPING_THREADS as f32).ceil() as usize));
        let ids_clone = Arc::clone(&ids);
        let target_paths = paths.drain(..target_paths).collect::<Vec<_>>();
        threads.push(tokio::spawn(async move {
            info!("Scraping {} paths...", target_paths.len());
            let ids_ = scrape_avatar_ids_(target_paths).await.expect("Failed to scrape avatar ids");
            info!("Scraped {} ids from paths", ids_.len());
            let mut ids = ids_clone.lock().await;
            info!("Extending ids with {} ids", ids_.len());
            tokio::spawn(process_avatars(std::env::var("AUTH_TOKEN").ok(), ids_.clone()));
            ids.extend(ids_);
        }));
    }
    info!("Waiting for threads to finish...");
    for thread in threads {
        thread.await.expect("Failed to join thread");
    }
    let ids = ids.lock().await;
    info!("Scraped {} avatar ids", ids.len());
    Ok(ids.clone())
}

pub type WatchResponse = (Sender<PathBuf>, Receiver<PathBuf>, PollWatcher);

/// # Errors
/// Will return `Err` if `PollWatcher::watch` errors
pub fn watch<P: AsRef<Path>>(path: P) -> notify::Result<WatchResponse> {
    let (tx_a, rx_a) = crossbeam::channel::unbounded();
    let tx_b = tx_a.clone();

    let mut watcher = PollWatcher::new(
        move |watch_event: notify::Result<Event>| {
            if let Ok(event) = watch_event {
                for path in event.paths {
                    let _ = tx_b.send(path);
                }
            }
        },
        Config::default(),
    )?;

    watcher.watch(path.as_ref(), RecursiveMode::Recursive)?;

    Ok((tx_a, rx_a, watcher))
}