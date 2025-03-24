use crate::providers::provider::{DatabaseEntry, Provider};
use rusqlite::{Connection, Error, params};
use std::fmt::Debug;
use std::time::{Instant, SystemTime, UNIX_EPOCH};
use tokio::fs;
use tokio::sync::Mutex;

#[derive(Debug)]
pub struct PrismicProvider {
    pub(crate) database_connection: Mutex<Connection>,
    pub(crate) table_name: String,
}
const PRISMIC_DATABASE_URL: &str =
    "https://gist.githubusercontent.com/Mwr247/ef9a06ee1d3209a558b05561f7332d8e/raw/vrcavtrdb.txt";

async fn get_prismic_database() -> Vec<DatabaseEntry> {
    let mut entries = vec![];
    let response = reqwest::get(PRISMIC_DATABASE_URL)
        .await
        .unwrap()
        .text()
        .await
        .unwrap();
    let lines = response.split('\n');
    for line in lines {
        let line = line.chars().rev().collect::<String>();
        let parts: Vec<&str> = line.split('\t').rev().collect();
        if parts.len() == 4 {
            let avatar_id = decode_avatar_id(parts[0]);
            if avatar_id.is_err() {
                continue;
            }
            let avatar_name = parts[1].to_string();
            let author_id = parts[2].to_string();
            let avatar_description = parts[3].to_string();
            let entry = DatabaseEntry {
                avatar_id: avatar_id.unwrap(),
                avatar_name,
                avatar_description,
                author_id,
            };
            // avtr_a1c59c2f-b11e-48ad-9ba7-d7e0fede57d3
            // avtr_0009c2a9-265d-497a-7ba8-6912bb83ee5c
            entries.push(entry);
        }
    }

    entries
}

const CIPHER: &str = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+";
fn decode_avatar_id(crypt: &str) -> Result<String, Box<dyn std::error::Error>> {
    if crypt.len() != 24 {
        return Err(Box::new(std::io::Error::new(
            std::io::ErrorKind::InvalidInput,
            "Invalid input length",
        )));
    }
    let mut decrypt: Vec<char> = vec![' '; 33];
    let index_of_char = match CIPHER.find(crypt.chars().nth(21).unwrap()) {
        Some(v) => v,
        None => {
            return Err(Box::new(std::io::Error::new(
                std::io::ErrorKind::InvalidInput,
                "Invalid input",
            )));
        }
    };
    let new_format = (index_of_char >> 2) & 2;

    for i in 0..11 {
        let idx = i * 3;
        let first = match CIPHER.find(crypt.chars().nth(i * 2).unwrap()) {
            Some(v) => v,
            None => {
                return Err(Box::new(std::io::Error::new(
                    std::io::ErrorKind::InvalidInput,
                    "Invalid input",
                )));
            }
        };

        let third = match CIPHER.find(crypt.chars().nth(i * 2 + 1).unwrap()) {
            Some(v) => v,
            None => {
                return Err(Box::new(std::io::Error::new(
                    std::io::ErrorKind::InvalidInput,
                    "Invalid input",
                )));
            }
        };
        decrypt[idx] = CIPHER.chars().nth((first >> new_format) & 15).unwrap();
        let second = if new_format == 0 {
            (first >> 2) & 12
        } else {
            (first & 3) << 2
        };

        decrypt[idx + 1] = CIPHER.chars().nth(second | ((third >> 4) & 3)).unwrap();
        decrypt[idx + 2] = CIPHER.chars().nth(third & 15).unwrap();
    }

    decrypt.pop();

    decrypt.insert(8, '-');
    decrypt.insert(13, '-');
    decrypt.insert(18, '-');
    decrypt.insert(23, '-');

    let decrypted_str: String = decrypt.into_iter().collect();

    Ok(format!("avtr_{}", decrypted_str))
}

impl Provider for PrismicProvider {
    async fn on_start(&self) -> () {
        let conn = self.database_connection().lock().await;
        conn.execute(
            &format!(
                "CREATE TABLE IF NOT EXISTS {} (
            id INTEGER PRIMARY KEY AUTOINCREMENT,

            avatar_id TEXT UNIQUE NOT NULL,

            avatar_name TEXT NOT NULL,
            avatar_description TEXT NOT NULL,

            author_id TEXT NOT NULL,

            timestamp TEXT DEFAULT CURRENT_TIMESTAMP
        )",
                self.table_name
            ),
            params![],
        )
        .unwrap();
    }
    fn database_connection(&self) -> &Mutex<Connection> {
        &self.database_connection
    }

    async fn get_entries(&self) -> Result<Vec<(DatabaseEntry, String)>, Error> {
        let conn = self.database_connection().lock().await;
        let mut stmt = conn.prepare(&format!("SELECT * FROM {}", self.table_name))?;
        let entries = stmt.query_map(params![], |row| {
            Ok((
                DatabaseEntry {
                    avatar_id: row.get(1)?,
                    avatar_name: row.get(2)?,
                    avatar_description: row.get(3)?,
                    author_id: row.get(4)?,
                },
                row.get(4)?,
            ))
        })?;
        let mut result = Vec::new();
        for entry in entries {
            result.push(entry?);
        }
        Ok(result)
    }

    async fn add_entries(&self, avatars: Vec<DatabaseEntry>) -> Result<(), Error> {
        const CHUNK_SIZE: usize = 250; // SQLite's max is 999 params: 250 * 4 = 1000 params
        let mut conn = self.database_connection().lock().await;
        let start_time = Instant::now();
        let tx = conn.transaction()?;
        {
            let mut stmt = tx.prepare_cached(
                "INSERT OR IGNORE INTO avatars
        (avatar_id, avatar_name, avatar_description, author_id)
        VALUES (?, ?, ?, ?)",
            )?;

            for (i, chunk) in avatars.chunks(CHUNK_SIZE).enumerate() {
                for avatar in chunk {
                    stmt.execute(params![
                        avatar.avatar_id,
                        avatar.avatar_name,
                        avatar.avatar_description,
                        avatar.author_id
                    ])?;
                }

                if i % 100 == 0 {
                    let processed = (i + 1) * CHUNK_SIZE;
                    println!(
                        "Inserted {} of {} ({:.2}%)",
                        processed.min(avatars.len()),
                        avatars.len(),
                        (processed as f32 / avatars.len() as f32) * 100.0
                    );
                }
            }
        }
        tx.commit()?;
        println!(
            "Inserted {} avatars in {:?}",
            avatars.len(),
            start_time.elapsed()
        );
        Ok(())
    }

    async fn spawn_thread(&self) -> () {
        println!("Fetching data from Prismic...");
        // check the last time the database was updated, if it was TODAY, skip updating the database
        let metadata = fs::metadata("prismic.db").await;
        if metadata.is_ok() {
            let metadata = metadata.unwrap();
            let last_modified = metadata.modified().unwrap();
            let last_modified_time = last_modified.duration_since(UNIX_EPOCH).unwrap();
            let last_modified_day = last_modified_time.as_secs() / (60 * 60 * 24);
            let current_day = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() / (60 * 60 * 24);
            println!("Current day: {}", current_day);
            println!("Last modified: {}", last_modified_day);
            // if last_modified_day == current_day {
            //     println!("Database is up-to-date, skipping update.");
            //     return;
            // }
        }
        let db = get_prismic_database().await;
        println!("Adding {} entries to the database...", db.len());
        tokio::time::sleep(std::time::Duration::from_secs(10)).await;
        // get entries;
        let entries = self.get_entries().await.unwrap();
        let unique_entries: Vec<DatabaseEntry> = db
           .into_iter()
           .filter(|entry| !entries.iter().any(|e| e.0.avatar_id == entry.avatar_id))
           .collect();
        println!("Adding {} new entries to the database...", unique_entries.len());
        self.add_entries(unique_entries).await.unwrap();
        println!("Database updated successfully!");
    }
}
