use crate::cache::cache_windows_player::get_cache_path;
use crate::cache::db::init_db;
use crate::cache::scrape::scrape;

mod cache_windows_player;
pub(crate) mod db;
mod scrape;
mod download_avatars;

pub async fn run() -> Result<(), Box<dyn std::error::Error>> {
    println!("Cache module running...");
    // resolve CACHE_PATH to absolute path
    println!("Cache path: {}", get_cache_path());
    // avatar id -> JSON data ( minimizing api calls )
    init_db("avatar_cache").expect("Error initializing avatar_cache database");
    // dir path -> avatar id ( removes the process of opening a handle and searching the file )
    init_db("dir_to_id").expect("Error initializing dir_to_id database");
    init_db("avatar_favorites").expect("Error initializing favorite_avatar database");
    println!("Cache loaded successfully!");
    download_dbs().await?;
    loop {

        scrape(std::env::var("AUTH_TOKEN").ok()).await.expect("Error scraping data");
        tokio::time::sleep(std::time::Duration::from_secs(5)).await;
    }
}

