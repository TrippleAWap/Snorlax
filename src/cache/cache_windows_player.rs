use std::collections::HashMap;
use std::env;
use std::path::PathBuf;
use lazy_regex::{lazy_regex, Lazy, Regex};
use rusqlite::params;
use tokio::fs::File;
use tokio::io::{AsyncBufReadExt, BufReader};
use crate::cache::db::CONN;

static AVATAR_REGEX: Lazy<Regex> = lazy_regex!(r"avtr_\w{8}-\w{4}-\w{4}-\w{4}-\w{12}");

pub fn get_cache_path() -> String {
    let app_data = env::var("APPDATA").expect("APPDATA environment variable not set");

    // Construct the path to the VRChat directory
    let vrc_path = PathBuf::from(app_data).join("..").join("LocalLow").join("VRChat").join("VRChat");

    // Convert the path to a string and return it
    vrc_path.to_str().unwrap().to_string()
}

#[allow(dead_code)]
pub fn cache_ids(map: &HashMap<String, String>) -> Result<(), rusqlite::Error>{
    let conn = CONN.lock().unwrap();
    let mut stmt = conn.prepare("INSERT INTO dir_to_id (key, value) VALUES (?,?)")?;
    for (key, value) in map {
        stmt.execute(params![key, value])?;
    }
    Ok(())
}

pub fn get_cached_ids() -> Result<HashMap<String, String>, rusqlite::Error> {
    let conn = CONN.lock().unwrap();
    let mut stmt = conn.prepare("SELECT key, value FROM dir_to_id")?;
    let mut rows = stmt.query_map([], |row| {
        Ok((row.get::<usize, String>(0)?, row.get::<usize, String>(1)?))
    })?;

    let mut hashmap: HashMap<String, String> = HashMap::new();
    while let Some(row) = rows.next().transpose()? {
        let (key, value) = row;
        hashmap.insert(key, value);
    }

    Ok(hashmap)
}
/// Returns a list of all avatar ids in the specified file.
/// This allows for scraping of databases and the cache via a single function.
#[allow(dead_code)]
pub async fn get_avatar_ids(path: &str, single: bool) -> Vec<String> {
    let Ok(file) = File::open(path).await else {
        return Vec::new();
    };
    let mut cache_id: &str = "";
    if path.ends_with("\\__data") {
        let split: Vec<&str> = path.split("\\").collect();
        cache_id = split[split.len() - 2];
    }
    if !cache_id.is_empty() {
    }

    let mut reader = BufReader::new(file);
    let mut avatar_ids = Vec::new();
    let mut buf = Vec::new();

    while reader.read_until(b'\n', &mut buf).await.unwrap_or(0) > 0 {
        let line = String::from_utf8_lossy(&buf);
        for mat in AVATAR_REGEX.find_iter(&line) {
            avatar_ids.push(mat.as_str().to_string());
            if single {
                return avatar_ids;
            }
        }
        buf.clear();
    }
    avatar_ids
}

/// Returns all paths to files in the specified directory that match the specified filter function.
/// If recursive is true, the function will also search subdirectories.
/// If depth is greater than 0, the function will only search directories up to the specified depth.
/// If dir_handle is specified, it will be called for each directory encountered, and the results will be added to the final list.
/// The function returns a list of all matching paths.
/// ( caching directory to avatar id mapping should be implemented by using dir_handle )
#[allow(dead_code)]
pub async fn walk_dir(path: &str, recursive: bool, depth: u32, file_handle: fn(&str) -> bool, dir_handle: Option<fn(&str, u32) -> Option<Vec<String>>>) -> Vec<String> {
    let mut results = Vec::new();
    let Ok(mut dir) = tokio::fs::read_dir(path).await else {
        println!("Failed to read directory: {}", path);
        return results;
    };
    while let Ok(Some(entry)) = dir.next_entry().await {
        let file_type = entry.file_type().await.unwrap();
        if file_type.is_dir() {
            if let Some(dir_handle) = dir_handle {
                let result = dir_handle(entry.path().to_str().unwrap(), depth);
                // if result is None, fallback to default behavior
                // if result is Some, append to results and continue
                if let Some(mut sub_results) = result {
                    results.append(&mut sub_results);
                    continue;
                }

            }
            if recursive && depth > 0 {
                results.extend(Box::pin(walk_dir(entry.path().to_str().unwrap(), recursive, depth - 1, file_handle, dir_handle)).await);
            }
        } else if file_type.is_file() {
            if file_handle(entry.path().to_str().unwrap()) {
                results.push(entry.path().to_str().unwrap().to_string());
            }
        }
    }
    results
}
